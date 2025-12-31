// Package hive provides a Hive metadata collector implementation.
package hive

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"go-metadata/internal/collector"
)

// ParseDescribeFormatted parses the output of DESCRIBE FORMATTED command
// and returns a TableMetadata structure.
//
// The DESCRIBE FORMATTED output has the following sections:
// 1. Column information (col_name, data_type, comment)
// 2. Partition Information (if partitioned)
// 3. Detailed Table Information (table properties)
// 4. Storage Information (input/output format, serde, location)
// 5. Table Parameters (TBLPROPERTIES)
func ParseDescribeFormatted(rows [][]string, catalog, schema, table string) (*collector.TableMetadata, error) {
	if len(rows) == 0 {
		return nil, fmt.Errorf("empty DESCRIBE FORMATTED output")
	}

	metadata := &collector.TableMetadata{
		Catalog:    catalog,
		Schema:     schema,
		Name:       table,
		Type:       collector.TableTypeTable,
		Properties: make(map[string]string),
	}

	parser := &describeParser{
		rows:     rows,
		metadata: metadata,
	}

	if err := parser.parse(); err != nil {
		return nil, err
	}

	return metadata, nil
}

// describeParser handles parsing of DESCRIBE FORMATTED output
type describeParser struct {
	rows     [][]string
	metadata *collector.TableMetadata
	idx      int
}

// Section markers in DESCRIBE FORMATTED output
const (
	sectionPartitionInfo    = "# Partition Information"
	sectionDetailedInfo     = "# Detailed Table Information"
	sectionStorageInfo      = "# Storage Information"
	sectionStorageDesc      = "# Storage Desc Params"
	sectionTableParameters  = "Table Parameters:"
	sectionColumnHeader     = "# col_name"
)

func (p *describeParser) parse() error {
	// Parse columns first (before any section marker)
	p.parseColumns()

	// Continue parsing other sections
	for p.idx < len(p.rows) {
		row := p.rows[p.idx]
		if len(row) == 0 {
			p.idx++
			continue
		}

		col0 := strings.TrimSpace(row[0])

		switch {
		case strings.Contains(col0, sectionPartitionInfo):
			p.idx++
			p.parsePartitionInfo()
		case strings.Contains(col0, sectionDetailedInfo):
			p.idx++
			p.parseDetailedInfo()
		case strings.Contains(col0, sectionStorageInfo):
			p.idx++
			p.parseStorageInfo()
		case strings.Contains(col0, sectionStorageDesc):
			p.idx++
			p.parseStorageDescParams()
		default:
			p.idx++
		}
	}

	return nil
}


// parseColumns parses the column information section
func (p *describeParser) parseColumns() {
	ordinalPos := 1

	for p.idx < len(p.rows) {
		row := p.rows[p.idx]
		if len(row) == 0 {
			p.idx++
			continue
		}

		col0 := strings.TrimSpace(row[0])

		// Check for section markers
		if strings.HasPrefix(col0, "#") {
			// Check if this is the column header row
			if strings.Contains(col0, "col_name") {
				p.idx++
				continue
			}
			// Other section marker, stop parsing columns
			return
		}

		// Skip empty or separator rows
		if col0 == "" || col0 == "---" {
			p.idx++
			continue
		}

		// Parse column: col_name, data_type, comment
		colName := col0
		dataType := ""
		comment := ""

		if len(row) > 1 {
			dataType = strings.TrimSpace(row[1])
		}
		if len(row) > 2 {
			comment = strings.TrimSpace(row[2])
		}

		// Skip if this looks like a property row (contains ':')
		if strings.Contains(colName, ":") && !strings.Contains(dataType, "<") {
			p.idx++
			continue
		}

		// Skip if dataType is empty (likely a section marker or property)
		if dataType == "" {
			p.idx++
			continue
		}

		col := collector.Column{
			OrdinalPosition: ordinalPos,
			Name:            colName,
			Type:            normalizeHiveType(dataType),
			SourceType:      dataType,
			Nullable:        true, // Hive columns are nullable by default
			Comment:         comment,
		}

		// Parse type parameters (length, precision, scale)
		parseTypeParams(dataType, &col)

		p.metadata.Columns = append(p.metadata.Columns, col)
		ordinalPos++
		p.idx++
	}
}

// parsePartitionInfo parses the partition information section
func (p *describeParser) parsePartitionInfo() {
	var partitionColumns []string
	seenHeader := false

	for p.idx < len(p.rows) {
		row := p.rows[p.idx]
		if len(row) == 0 {
			p.idx++
			continue
		}

		col0 := strings.TrimSpace(row[0])

		// Check for next section
		if strings.Contains(col0, "# Detailed") || strings.Contains(col0, "# Storage") {
			break
		}

		// Skip header rows (# col_name, etc.)
		if strings.HasPrefix(col0, "#") {
			seenHeader = true
			p.idx++
			continue
		}

		// Skip empty rows
		if col0 == "" {
			p.idx++
			continue
		}

		// After seeing header, this should be a partition column
		if seenHeader {
			// Check if this looks like a column definition (has data type in second column)
			if len(row) > 1 && strings.TrimSpace(row[1]) != "" {
				partitionColumns = append(partitionColumns, col0)

				// Mark the column as partition column in metadata
				for i := range p.metadata.Columns {
					if p.metadata.Columns[i].Name == col0 {
						p.metadata.Columns[i].IsPartitionColumn = true
						break
					}
				}
			}
		}

		p.idx++
	}

	// Add partition info if we found partition columns
	if len(partitionColumns) > 0 {
		p.metadata.Partitions = append(p.metadata.Partitions, collector.PartitionInfo{
			Name:    "partition",
			Type:    "LIST",
			Columns: partitionColumns,
		})
	}
}

// parseDetailedInfo parses the detailed table information section
func (p *describeParser) parseDetailedInfo() {
	for p.idx < len(p.rows) {
		row := p.rows[p.idx]
		if len(row) == 0 {
			p.idx++
			continue
		}

		col0 := strings.TrimSpace(row[0])

		// Check for next section
		if strings.HasPrefix(col0, "# Storage") {
			return
		}

		// Parse key-value pairs
		key := col0
		value := ""
		if len(row) > 1 {
			value = strings.TrimSpace(row[1])
		}

		// Handle multi-column values
		if len(row) > 2 && row[2] != "" {
			value = value + " " + strings.TrimSpace(row[2])
		}

		// Remove trailing colon from key
		key = strings.TrimSuffix(key, ":")

		switch strings.ToLower(key) {
		case "table type":
			p.metadata.Type = mapHiveTableType(value)
		case "comment":
			p.metadata.Comment = value
		case "table parameters":
			p.idx++
			p.parseTableParameters()
			continue
		default:
			// Store other properties
			if key != "" && value != "" && !strings.HasPrefix(key, "#") {
				p.metadata.Properties[key] = value
			}
		}

		p.idx++
	}
}


// parseTableParameters parses the TBLPROPERTIES section
func (p *describeParser) parseTableParameters() {
	for p.idx < len(p.rows) {
		row := p.rows[p.idx]
		if len(row) == 0 {
			p.idx++
			continue
		}

		col0 := strings.TrimSpace(row[0])

		// Check for next section
		if strings.HasPrefix(col0, "# Storage") || strings.HasPrefix(col0, "# Detailed") {
			return
		}

		// Skip empty or header rows
		if col0 == "" || strings.HasPrefix(col0, "#") {
			p.idx++
			continue
		}

		// Parse property: key, value
		key := col0
		value := ""
		if len(row) > 1 {
			value = strings.TrimSpace(row[1])
		}

		if key != "" && value != "" {
			p.metadata.Properties[key] = value
		}

		p.idx++
	}
}

// parseStorageInfo parses the storage information section
func (p *describeParser) parseStorageInfo() {
	if p.metadata.Storage == nil {
		p.metadata.Storage = &collector.StorageInfo{}
	}

	for p.idx < len(p.rows) {
		row := p.rows[p.idx]
		if len(row) == 0 {
			p.idx++
			continue
		}

		col0 := strings.TrimSpace(row[0])

		// Check for next section
		if strings.HasPrefix(col0, "# Storage Desc") {
			return
		}

		// Skip header rows
		if strings.HasPrefix(col0, "#") {
			p.idx++
			continue
		}

		// Parse key-value pairs
		key := strings.TrimSuffix(col0, ":")
		value := ""
		if len(row) > 1 {
			value = strings.TrimSpace(row[1])
		}

		switch strings.ToLower(key) {
		case "inputformat":
			p.metadata.Storage.InputFormat = value
			p.metadata.Storage.Format = extractStorageFormat(value)
		case "outputformat":
			p.metadata.Storage.OutputFormat = value
		case "serde library", "serdelibrary":
			p.metadata.Storage.SerDe = value
		case "location":
			p.metadata.Storage.Location = value
		case "compressed":
			p.metadata.Storage.Compressed = strings.ToLower(value) == "yes" || strings.ToLower(value) == "true"
		}

		p.idx++
	}
}

// parseStorageDescParams parses storage descriptor parameters
func (p *describeParser) parseStorageDescParams() {
	for p.idx < len(p.rows) {
		row := p.rows[p.idx]
		if len(row) == 0 {
			p.idx++
			continue
		}

		col0 := strings.TrimSpace(row[0])

		// Check for next section
		if strings.HasPrefix(col0, "#") && !strings.Contains(col0, "Storage Desc") {
			return
		}

		// Skip header rows
		if strings.HasPrefix(col0, "#") || col0 == "" {
			p.idx++
			continue
		}

		// Parse property
		key := col0
		value := ""
		if len(row) > 1 {
			value = strings.TrimSpace(row[1])
		}

		// Store as property
		if key != "" && value != "" {
			p.metadata.Properties["storage."+key] = value
		}

		p.idx++
	}
}

// normalizeHiveType normalizes Hive data type to standard type
func normalizeHiveType(dataType string) string {
	// Remove any parameters from type (e.g., varchar(100) -> varchar)
	baseType := strings.ToLower(dataType)
	if idx := strings.Index(baseType, "("); idx != -1 {
		baseType = baseType[:idx]
	}
	if idx := strings.Index(baseType, "<"); idx != -1 {
		baseType = baseType[:idx]
	}
	baseType = strings.TrimSpace(baseType)

	switch baseType {
	case "tinyint", "smallint", "int", "integer", "bigint":
		return "INTEGER"
	case "float", "double", "double precision":
		return "FLOAT"
	case "decimal", "numeric":
		return "DECIMAL"
	case "string", "varchar", "char":
		return "STRING"
	case "date":
		return "DATE"
	case "timestamp", "timestamp with local time zone":
		return "TIMESTAMP"
	case "binary":
		return "BINARY"
	case "boolean":
		return "BOOLEAN"
	case "array":
		return "ARRAY"
	case "map":
		return "MAP"
	case "struct":
		return "STRUCT"
	case "uniontype":
		return "UNION"
	default:
		return strings.ToUpper(baseType)
	}
}


// parseTypeParams extracts length, precision, and scale from type string
func parseTypeParams(dataType string, col *collector.Column) {
	// Match patterns like varchar(100), decimal(10,2), char(50)
	re := regexp.MustCompile(`\((\d+)(?:,\s*(\d+))?\)`)
	matches := re.FindStringSubmatch(dataType)

	if len(matches) >= 2 {
		baseType := strings.ToLower(dataType)

		if strings.HasPrefix(baseType, "varchar") || strings.HasPrefix(baseType, "char") {
			if length, err := strconv.Atoi(matches[1]); err == nil {
				col.Length = &length
			}
		} else if strings.HasPrefix(baseType, "decimal") || strings.HasPrefix(baseType, "numeric") {
			if precision, err := strconv.Atoi(matches[1]); err == nil {
				col.Precision = &precision
			}
			if len(matches) >= 3 && matches[2] != "" {
				if scale, err := strconv.Atoi(matches[2]); err == nil {
					col.Scale = &scale
				}
			}
		}
	}
}

// mapHiveTableType maps Hive table type string to TableType
func mapHiveTableType(hiveType string) collector.TableType {
	switch strings.ToUpper(hiveType) {
	case "VIRTUAL_VIEW", "VIEW":
		return collector.TableTypeView
	case "EXTERNAL_TABLE":
		return collector.TableTypeExternalTable
	case "MATERIALIZED_VIEW":
		return collector.TableTypeMaterializedView
	case "MANAGED_TABLE", "TABLE":
		return collector.TableTypeTable
	default:
		return collector.TableTypeTable
	}
}

// extractStorageFormat extracts the storage format from InputFormat class name
func extractStorageFormat(inputFormat string) string {
	inputFormat = strings.ToLower(inputFormat)

	switch {
	case strings.Contains(inputFormat, "parquet"):
		return "parquet"
	case strings.Contains(inputFormat, "orc"):
		return "orc"
	case strings.Contains(inputFormat, "avro"):
		return "avro"
	case strings.Contains(inputFormat, "rcfile"):
		return "rcfile"
	case strings.Contains(inputFormat, "sequencefile"):
		return "sequencefile"
	case strings.Contains(inputFormat, "textinputformat"):
		return "text"
	case strings.Contains(inputFormat, "json"):
		return "json"
	default:
		return ""
	}
}

// ParsePartitionSpec parses a partition specification string like "dt=2023-01-01/region=us"
func ParsePartitionSpec(spec string) map[string]string {
	result := make(map[string]string)
	parts := strings.Split(spec, "/")

	for _, part := range parts {
		kv := strings.SplitN(part, "=", 2)
		if len(kv) == 2 {
			result[kv[0]] = kv[1]
		}
	}

	return result
}
