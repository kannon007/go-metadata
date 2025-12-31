// Package infer provides schema inference capabilities for schema-less data sources.
package infer

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"sort"
	"strconv"
	"strings"

	"go-metadata/internal/collector"
)

// FileSchemaInferrer implements schema inference for structured files.
type FileSchemaInferrer struct {
	config *InferConfig
}

// NewFileSchemaInferrer creates a new FileSchemaInferrer with default configuration.
func NewFileSchemaInferrer() *FileSchemaInferrer {
	return &FileSchemaInferrer{
		config: DefaultInferConfig(),
	}
}

// NewFileSchemaInferrerWithConfig creates a new FileSchemaInferrer with the specified configuration.
func NewFileSchemaInferrerWithConfig(config *InferConfig) *FileSchemaInferrer {
	return &FileSchemaInferrer{
		config: config,
	}
}

// SetConfig updates the inference configuration.
func (f *FileSchemaInferrer) SetConfig(config *InferConfig) {
	f.config = config
}

// GetConfig returns the current inference configuration.
func (f *FileSchemaInferrer) GetConfig() *InferConfig {
	return f.config
}

// FileFormat represents supported file formats.
type FileFormat string

const (
	FormatCSV     FileFormat = "csv"
	FormatJSON    FileFormat = "json"
	FormatParquet FileFormat = "parquet"
)

// FileInferenceRequest contains the parameters for file schema inference.
type FileInferenceRequest struct {
	Reader io.Reader
	Format FileFormat
	// CSVOptions contains CSV-specific options
	CSVOptions *CSVOptions
}

// CSVOptions contains options for CSV file inference.
type CSVOptions struct {
	// Delimiter is the field delimiter (default: comma)
	Delimiter rune
	// HasHeader indicates if the first row contains column names
	HasHeader bool
	// SkipRows is the number of rows to skip at the beginning
	SkipRows int
}

// DefaultCSVOptions returns default CSV options.
func DefaultCSVOptions() *CSVOptions {
	return &CSVOptions{
		Delimiter: ',',
		HasHeader: true,
		SkipRows:  0,
	}
}

// Infer analyzes file samples and returns inferred column definitions.
// The samples parameter should be []interface{} where each element is a FileInferenceRequest.
func (f *FileSchemaInferrer) Infer(ctx context.Context, samples []interface{}) ([]collector.Column, error) {
	if !f.config.Enabled {
		return []collector.Column{}, nil
	}

	if len(samples) == 0 {
		return []collector.Column{}, nil
	}

	// For file inference, we typically work with a single file
	// Take the first sample as the primary file
	sample := samples[0]
	request, ok := sample.(*FileInferenceRequest)
	if !ok {
		return nil, fmt.Errorf("sample is not a *FileInferenceRequest, got %T", sample)
	}

	return f.InferFromFile(ctx, request)
}

// InferFromFile infers schema from a single file.
func (f *FileSchemaInferrer) InferFromFile(ctx context.Context, request *FileInferenceRequest) ([]collector.Column, error) {
	switch request.Format {
	case FormatCSV:
		return f.inferCSV(ctx, request.Reader, request.CSVOptions)
	case FormatJSON:
		return f.inferJSON(ctx, request.Reader)
	case FormatParquet:
		return f.inferParquet(ctx, request.Reader)
	default:
		return nil, fmt.Errorf("unsupported file format: %s", request.Format)
	}
}

// inferCSV infers schema from a CSV file.
func (f *FileSchemaInferrer) inferCSV(ctx context.Context, reader io.Reader, options *CSVOptions) ([]collector.Column, error) {
	if options == nil {
		options = DefaultCSVOptions()
	}

	csvReader := csv.NewReader(reader)
	csvReader.Comma = options.Delimiter

	// Skip initial rows if configured
	for i := 0; i < options.SkipRows; i++ {
		_, err := csvReader.Read()
		if err != nil {
			if err == io.EOF {
				return []collector.Column{}, nil
			}
			return nil, fmt.Errorf("error skipping row %d: %w", i, err)
		}
	}

	var headers []string
	var firstDataRow []string

	// Read header row
	if options.HasHeader {
		headerRow, err := csvReader.Read()
		if err != nil {
			if err == io.EOF {
				return []collector.Column{}, nil
			}
			return nil, fmt.Errorf("error reading header row: %w", err)
		}
		headers = headerRow

		// Read first data row for type inference
		firstDataRow, err = csvReader.Read()
		if err != nil {
			if err == io.EOF {
				// Only header, no data - create columns with unknown types
				return f.createColumnsFromHeaders(headers), nil
			}
			return nil, fmt.Errorf("error reading first data row: %w", err)
		}
	} else {
		// No header, read first row as data and generate column names
		firstDataRow, err := csvReader.Read()
		if err != nil {
			if err == io.EOF {
				return []collector.Column{}, nil
			}
			return nil, fmt.Errorf("error reading first data row: %w", err)
		}

		// Generate column names
		headers = make([]string, len(firstDataRow))
		for i := range headers {
			headers[i] = fmt.Sprintf("column_%d", i+1)
		}
	}

	// Collect type information from multiple rows
	fieldTypes := make(map[string]*FieldTypeInfo)
	for _, header := range headers {
		fieldTypes[header] = NewFieldTypeInfo(header, 0)
	}

	// Analyze first data row
	f.analyzeCSVRow(firstDataRow, headers, fieldTypes)

	// Read additional rows for better type inference (up to sample size)
	rowCount := 1
	maxRows := f.config.SampleSize
	if maxRows <= 0 {
		maxRows = 100 // Default sample size for CSV
	}

	for rowCount < maxRows {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		row, err := csvReader.Read()
		if err != nil {
			if err == io.EOF {
				break
			}
			// Skip malformed rows and continue
			continue
		}

		f.analyzeCSVRow(row, headers, fieldTypes)
		rowCount++
	}

	// Convert to columns preserving header order
	return f.fieldsToColumnsWithOrder(fieldTypes, headers, rowCount), nil
}

// analyzeCSVRow analyzes a single CSV row and updates field type information.
func (f *FileSchemaInferrer) analyzeCSVRow(row []string, headers []string, fieldTypes map[string]*FieldTypeInfo) {
	for i, value := range row {
		if i >= len(headers) {
			break // Skip extra columns
		}

		header := headers[i]
		typeName := f.inferCSVFieldType(value)
		fieldTypes[header].AddType(typeName)
	}
}

// inferCSVFieldType infers the type of a CSV field value.
func (f *FileSchemaInferrer) inferCSVFieldType(value string) string {
	// Trim whitespace
	value = strings.TrimSpace(value)

	// Empty or null values
	if value == "" || strings.ToLower(value) == "null" || strings.ToLower(value) == "na" {
		return "null"
	}

	// Boolean values
	lowerValue := strings.ToLower(value)
	if lowerValue == "true" || lowerValue == "false" || lowerValue == "yes" || lowerValue == "no" || lowerValue == "1" || lowerValue == "0" {
		return "boolean"
	}

	// Integer values
	if _, err := strconv.ParseInt(value, 10, 64); err == nil {
		return "integer"
	}

	// Float values
	if _, err := strconv.ParseFloat(value, 64); err == nil {
		return "number"
	}

	// Default to string
	return "string"
}

// createColumnsFromHeaders creates columns with unknown types from headers only.
func (f *FileSchemaInferrer) createColumnsFromHeaders(headers []string) []collector.Column {
	columns := make([]collector.Column, len(headers))
	for i, header := range headers {
		columns[i] = collector.Column{
			OrdinalPosition: i + 1,
			Name:            header,
			Type:            "TEXT",
			SourceType:      "unknown",
			Nullable:        true,
			Comment:         "No data available for type inference",
		}
	}
	return columns
}

// inferJSON infers schema from a JSON file.
func (f *FileSchemaInferrer) inferJSON(ctx context.Context, reader io.Reader) ([]collector.Column, error) {
	// Read all data
	data, err := io.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("error reading JSON data: %w", err)
	}

	// Try to parse as JSON array first
	var jsonArray []interface{}
	if err := json.Unmarshal(data, &jsonArray); err == nil {
		// It's an array of objects - use document inferrer
		documentInferrer := NewDocumentInferrerWithConfig(f.config)
		return documentInferrer.Infer(ctx, jsonArray)
	}

	// Try to parse as single JSON object
	var jsonObject map[string]interface{}
	if err := json.Unmarshal(data, &jsonObject); err == nil {
		// It's a single object - use document inferrer with single sample
		documentInferrer := NewDocumentInferrerWithConfig(f.config)
		return documentInferrer.Infer(ctx, []interface{}{jsonObject})
	}

	return nil, fmt.Errorf("invalid JSON format: expected array or object")
}

// inferParquet infers schema from a Parquet file.
func (f *FileSchemaInferrer) inferParquet(ctx context.Context, reader io.Reader) ([]collector.Column, error) {
	// Note: This is a placeholder implementation
	// In a real implementation, you would use a Parquet library like:
	// - github.com/apache/arrow/go/parquet
	// - github.com/xitongsys/parquet-go
	
	// For now, return an error indicating Parquet support is not implemented
	return nil, fmt.Errorf("Parquet format inference not implemented - requires parquet library integration")
}

// fieldsToColumnsWithOrder converts field type information to column definitions preserving the specified order.
func (f *FileSchemaInferrer) fieldsToColumnsWithOrder(fieldTypes map[string]*FieldTypeInfo, fieldOrder []string, totalSamples int) []collector.Column {
	columns := make([]collector.Column, 0, len(fieldOrder))

	ordinalPosition := 1
	for _, fieldName := range fieldOrder {
		fieldInfo, exists := fieldTypes[fieldName]
		if !exists {
			continue // Skip fields that don't exist in fieldTypes
		}
		
		// Determine the final type based on merge strategy
		var finalType string
		var sourceType string
		
		switch f.config.TypeMerge {
		case TypeMergeUnion:
			// Create union type from all observed types
			types := make([]string, 0, len(fieldInfo.Types))
			for typeName := range fieldInfo.Types {
				types = append(types, typeName)
			}
			finalType = strings.Join(types, "|")
			sourceType = finalType
		case TypeMergeMostCommon:
			// Use the most common type
			finalType = fieldInfo.MostCommonType()
			sourceType = finalType
		default:
			// Default to most common
			finalType = fieldInfo.MostCommonType()
			sourceType = finalType
		}

		// Map to standard SQL types
		sqlType := f.mapToSQLType(finalType)

		// Check if field is nullable (has null values or missing in some rows)
		nullable := fieldInfo.Nullable || fieldInfo.Types["null"] > 0

		column := collector.Column{
			OrdinalPosition: ordinalPosition,
			Name:            fieldName,
			Type:            sqlType,
			SourceType:      sourceType,
			Nullable:        nullable,
			Comment:         fmt.Sprintf("Inferred from %d samples", fieldInfo.TotalCount()),
		}

		columns = append(columns, column)
		ordinalPosition++
	}

	return columns
}

// fieldsToColumns converts field type information to column definitions.
func (f *FileSchemaInferrer) fieldsToColumns(fieldTypes map[string]*FieldTypeInfo, totalSamples int) []collector.Column {
	columns := make([]collector.Column, 0, len(fieldTypes))

	// Get field names and sort them for consistent output
	fieldNames := make([]string, 0, len(fieldTypes))
	for fieldName := range fieldTypes {
		fieldNames = append(fieldNames, fieldName)
	}
	sort.Strings(fieldNames)

	ordinalPosition := 1
	for _, fieldName := range fieldNames {
		fieldInfo := fieldTypes[fieldName]
		
		// Determine the final type based on merge strategy
		var finalType string
		var sourceType string
		
		switch f.config.TypeMerge {
		case TypeMergeUnion:
			// Create union type from all observed types
			types := make([]string, 0, len(fieldInfo.Types))
			for typeName := range fieldInfo.Types {
				types = append(types, typeName)
			}
			finalType = strings.Join(types, "|")
			sourceType = finalType
		case TypeMergeMostCommon:
			// Use the most common type
			finalType = fieldInfo.MostCommonType()
			sourceType = finalType
		default:
			// Default to most common
			finalType = fieldInfo.MostCommonType()
			sourceType = finalType
		}

		// Map to standard SQL types
		sqlType := f.mapToSQLType(finalType)

		// Check if field is nullable (has null values or missing in some rows)
		nullable := fieldInfo.Nullable || fieldInfo.Types["null"] > 0

		column := collector.Column{
			OrdinalPosition: ordinalPosition,
			Name:            fieldName,
			Type:            sqlType,
			SourceType:      sourceType,
			Nullable:        nullable,
			Comment:         fmt.Sprintf("Inferred from %d samples", fieldInfo.TotalCount()),
		}

		columns = append(columns, column)
		ordinalPosition++
	}

	return columns
}

// mapToSQLType maps inferred types to standard SQL types.
func (f *FileSchemaInferrer) mapToSQLType(inferredType string) string {
	// Handle union types
	if strings.Contains(inferredType, "|") {
		// For union types, use TEXT as a safe fallback
		return "TEXT"
	}

	switch inferredType {
	case "boolean":
		return "BOOLEAN"
	case "integer":
		return "BIGINT"
	case "number":
		return "DOUBLE"
	case "string":
		return "TEXT"
	case "null":
		return "TEXT" // Nullable text as fallback
	default:
		return "TEXT" // Safe fallback
	}
}

// FileInferenceResult holds the result of file schema inference.
type FileInferenceResult struct {
	// Columns contains the inferred column definitions
	Columns []collector.Column `json:"columns"`
	// SampleCount is the number of rows/records analyzed
	SampleCount int `json:"sample_count"`
	// Format is the detected/specified file format
	Format FileFormat `json:"format"`
	// HasHeader indicates if the file has a header row (CSV only)
	HasHeader bool `json:"has_header,omitempty"`
}

// InferFromFileWithResult returns detailed inference results.
func (f *FileSchemaInferrer) InferFromFileWithResult(ctx context.Context, request *FileInferenceRequest) (*FileInferenceResult, error) {
	columns, err := f.InferFromFile(ctx, request)
	if err != nil {
		return nil, err
	}

	result := &FileInferenceResult{
		Columns: columns,
		Format:  request.Format,
	}

	if request.Format == FormatCSV && request.CSVOptions != nil {
		result.HasHeader = request.CSVOptions.HasHeader
	}

	return result, nil
}