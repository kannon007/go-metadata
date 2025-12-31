// Package infer provides schema inference capabilities for schema-less data sources.
package infer

import (
	"context"
	"fmt"
	"reflect"
	"sort"
	"strings"

	"go-metadata/internal/collector"
)

// DocumentInferrer implements schema inference for document databases like MongoDB.
type DocumentInferrer struct {
	config *InferConfig
}

// NewDocumentInferrer creates a new DocumentInferrer with default configuration.
func NewDocumentInferrer() *DocumentInferrer {
	return &DocumentInferrer{
		config: DefaultInferConfig(),
	}
}

// NewDocumentInferrerWithConfig creates a new DocumentInferrer with the specified configuration.
func NewDocumentInferrerWithConfig(config *InferConfig) *DocumentInferrer {
	return &DocumentInferrer{
		config: config,
	}
}

// SetConfig updates the inference configuration.
func (d *DocumentInferrer) SetConfig(config *InferConfig) {
	d.config = config
}

// GetConfig returns the current inference configuration.
func (d *DocumentInferrer) GetConfig() *InferConfig {
	return d.config
}

// Infer analyzes document samples and returns inferred column definitions.
// The samples parameter should be []interface{} where each element is a map[string]interface{}.
func (d *DocumentInferrer) Infer(ctx context.Context, samples []interface{}) ([]collector.Column, error) {
	if !d.config.Enabled {
		return []collector.Column{}, nil
	}

	if len(samples) == 0 {
		return []collector.Column{}, nil
	}

	// Convert samples to documents
	documents := make([]map[string]interface{}, 0, len(samples))
	for i, sample := range samples {
		doc, ok := sample.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("sample %d is not a map[string]interface{}, got %T", i, sample)
		}
		documents = append(documents, doc)
	}

	// Limit sample size if configured
	if d.config.SampleSize > 0 && len(documents) > d.config.SampleSize {
		documents = documents[:d.config.SampleSize]
	}

	// Collect field type information
	fieldTypes := make(map[string]*FieldTypeInfo)
	totalSamples := len(documents)

	for _, doc := range documents {
		// Check for context cancellation
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		d.collectFields("", doc, fieldTypes, 0, totalSamples)
	}

	// Convert field types to columns
	columns := d.fieldsToColumns(fieldTypes, totalSamples)

	return columns, nil
}

// collectFields recursively collects field type information from a document.
func (d *DocumentInferrer) collectFields(prefix string, doc map[string]interface{}, fieldTypes map[string]*FieldTypeInfo, depth int, totalSamples int) {
	// Check depth limit
	if d.config.MaxDepth > 0 && depth >= d.config.MaxDepth {
		return
	}

	for key, value := range doc {
		fieldName := key
		if prefix != "" {
			fieldName = prefix + "." + key
		}

		// Get or create field type info
		if fieldTypes[fieldName] == nil {
			fieldTypes[fieldName] = NewFieldTypeInfo(fieldName, depth)
		}

		typeName := d.getTypeName(value)
		fieldTypes[fieldName].AddType(typeName)

		// Recursively process nested documents
		if nested, ok := value.(map[string]interface{}); ok {
			d.collectFields(fieldName, nested, fieldTypes, depth+1, totalSamples)
		}
	}

	// Mark fields as nullable if they don't appear in all samples
	for _, fieldInfo := range fieldTypes {
		if fieldInfo.TotalCount() < totalSamples {
			fieldInfo.Nullable = true
		}
	}
}

// getTypeName returns the type name for a value.
func (d *DocumentInferrer) getTypeName(value interface{}) string {
	if value == nil {
		return "null"
	}

	switch v := value.(type) {
	case bool:
		return "boolean"
	case int, int8, int16, int32, int64:
		return "integer"
	case uint, uint8, uint16, uint32, uint64:
		return "integer"
	case float32, float64:
		return "number"
	case string:
		return "string"
	case []interface{}:
		return "array"
	case map[string]interface{}:
		return "object"
	default:
		// Use reflection for other types
		return strings.ToLower(reflect.TypeOf(v).Kind().String())
	}
}

// fieldsToColumns converts field type information to column definitions.
func (d *DocumentInferrer) fieldsToColumns(fieldTypes map[string]*FieldTypeInfo, totalSamples int) []collector.Column {
	columns := make([]collector.Column, 0, len(fieldTypes))

	// Sort field names for consistent output
	fieldNames := make([]string, 0, len(fieldTypes))
	for name := range fieldTypes {
		fieldNames = append(fieldNames, name)
	}
	sort.Strings(fieldNames)

	ordinalPosition := 1
	for _, fieldName := range fieldNames {
		fieldInfo := fieldTypes[fieldName]
		
		// Determine the final type based on merge strategy
		var finalType string
		var sourceType string
		
		switch d.config.TypeMerge {
		case TypeMergeUnion:
			// Create union type from all observed types
			types := make([]string, 0, len(fieldInfo.Types))
			for typeName := range fieldInfo.Types {
				types = append(types, typeName)
			}
			sort.Strings(types)
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
		sqlType := d.mapToSQLType(finalType)

		column := collector.Column{
			OrdinalPosition: ordinalPosition,
			Name:            fieldName,
			Type:            sqlType,
			SourceType:      sourceType,
			Nullable:        fieldInfo.Nullable,
			Comment:         fmt.Sprintf("Inferred from %d samples, coverage: %.1f%%", 
				fieldInfo.TotalCount(), 
				float64(fieldInfo.TotalCount())/float64(totalSamples)*100),
		}

		columns = append(columns, column)
		ordinalPosition++
	}

	return columns
}

// mapToSQLType maps inferred types to standard SQL types.
func (d *DocumentInferrer) mapToSQLType(inferredType string) string {
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
	case "array":
		return "JSON"
	case "object":
		return "JSON"
	case "null":
		return "TEXT" // Nullable text as fallback
	default:
		return "TEXT" // Safe fallback
	}
}

// InferWithResult returns detailed inference results including metadata.
func (d *DocumentInferrer) InferWithResult(ctx context.Context, samples []interface{}) (*InferenceResult, error) {
	columns, err := d.Infer(ctx, samples)
	if err != nil {
		return nil, err
	}

	// Convert samples to documents for coverage calculation
	documents := make([]map[string]interface{}, 0, len(samples))
	for _, sample := range samples {
		if doc, ok := sample.(map[string]interface{}); ok {
			documents = append(documents, doc)
		}
	}

	// Limit sample size if configured
	if d.config.SampleSize > 0 && len(documents) > d.config.SampleSize {
		documents = documents[:d.config.SampleSize]
	}

	// Calculate field coverage and type distribution
	fieldCoverage := make(map[string]float64)
	typeDistribution := make(map[string]map[string]int)
	
	for _, column := range columns {
		fieldName := column.Name
		coverage := 0
		typeCounts := make(map[string]int)
		
		for _, doc := range documents {
			if d.hasField(fieldName, doc) {
				coverage++
				typeName := d.getFieldType(fieldName, doc)
				if typeName != "" {
					typeCounts[typeName]++
				}
			}
		}
		
		fieldCoverage[fieldName] = float64(coverage) / float64(len(documents)) * 100
		typeDistribution[fieldName] = typeCounts
	}

	return &InferenceResult{
		Columns:          columns,
		SampleCount:      len(documents),
		FieldCoverage:    fieldCoverage,
		TypeDistribution: typeDistribution,
	}, nil
}

// hasField checks if a field exists in a document (supports dot notation).
func (d *DocumentInferrer) hasField(fieldName string, doc map[string]interface{}) bool {
	parts := strings.Split(fieldName, ".")
	current := doc
	
	for i, part := range parts {
		value, exists := current[part]
		if !exists {
			return false
		}
		
		// If this is the last part, we found the field
		if i == len(parts)-1 {
			return true
		}
		
		// Continue traversing nested objects
		if nested, ok := value.(map[string]interface{}); ok {
			current = nested
		} else {
			return false
		}
	}
	
	return false
}

// getFieldType gets the type of a field in a document (supports dot notation).
func (d *DocumentInferrer) getFieldType(fieldName string, doc map[string]interface{}) string {
	parts := strings.Split(fieldName, ".")
	current := doc
	
	for i, part := range parts {
		value, exists := current[part]
		if !exists {
			return ""
		}
		
		// If this is the last part, return its type
		if i == len(parts)-1 {
			return d.getTypeName(value)
		}
		
		// Continue traversing nested objects
		if nested, ok := value.(map[string]interface{}); ok {
			current = nested
		} else {
			return ""
		}
	}
	
	return ""
}
