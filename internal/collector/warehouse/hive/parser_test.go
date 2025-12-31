// Package hive provides a Hive metadata collector implementation.
package hive

import (
	"testing"

	"go-metadata/internal/collector"
)

// TestParseDescribeFormatted tests parsing of DESCRIBE FORMATTED output
func TestParseDescribeFormatted(t *testing.T) {
	tests := []struct {
		name           string
		rows           [][]string
		catalog        string
		schema         string
		table          string
		wantErr        bool
		wantColumns    int
		wantTableType  collector.TableType
		wantPartitions int
	}{
		{
			name:    "empty output",
			rows:    [][]string{},
			catalog: "hive",
			schema:  "default",
			table:   "test",
			wantErr: true,
		},
		{
			name: "simple table with columns",
			rows: [][]string{
				{"# col_name", "data_type", "comment"},
				{"id", "int", "Primary key"},
				{"name", "string", "User name"},
				{"created_at", "timestamp", "Creation time"},
				{"", "", ""},
				{"# Detailed Table Information", "", ""},
				{"Table Type:", "MANAGED_TABLE", ""},
			},
			catalog:        "hive",
			schema:         "default",
			table:          "users",
			wantErr:        false,
			wantColumns:    3,
			wantTableType:  collector.TableTypeTable,
			wantPartitions: 0,
		},
		{
			name: "external table",
			rows: [][]string{
				{"# col_name", "data_type", "comment"},
				{"id", "bigint", ""},
				{"data", "string", ""},
				{"", "", ""},
				{"# Detailed Table Information", "", ""},
				{"Table Type:", "EXTERNAL_TABLE", ""},
			},
			catalog:       "hive",
			schema:        "default",
			table:         "external_data",
			wantErr:       false,
			wantColumns:   2,
			wantTableType: collector.TableTypeExternalTable,
		},
		{
			name: "view",
			rows: [][]string{
				{"# col_name", "data_type", "comment"},
				{"total", "bigint", ""},
				{"", "", ""},
				{"# Detailed Table Information", "", ""},
				{"Table Type:", "VIRTUAL_VIEW", ""},
			},
			catalog:       "hive",
			schema:        "default",
			table:         "summary_view",
			wantErr:       false,
			wantColumns:   1,
			wantTableType: collector.TableTypeView,
		},
		{
			name: "partitioned table",
			rows: [][]string{
				{"# col_name", "data_type", "comment"},
				{"id", "int", ""},
				{"name", "string", ""},
				{"", "", ""},
				{"# Partition Information", "", ""},
				{"# col_name", "data_type", "comment"},
				{"dt", "string", ""},
				{"region", "string", ""},
				{"", "", ""},
				{"# Detailed Table Information", "", ""},
				{"Table Type:", "MANAGED_TABLE", ""},
			},
			catalog:        "hive",
			schema:         "default",
			table:          "events",
			wantErr:        false,
			wantColumns:    2,
			wantTableType:  collector.TableTypeTable,
			wantPartitions: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			metadata, err := ParseDescribeFormatted(tt.rows, tt.catalog, tt.schema, tt.table)

			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if metadata == nil {
				t.Fatal("expected metadata, got nil")
			}

			if metadata.Catalog != tt.catalog {
				t.Errorf("expected catalog %s, got %s", tt.catalog, metadata.Catalog)
			}

			if metadata.Schema != tt.schema {
				t.Errorf("expected schema %s, got %s", tt.schema, metadata.Schema)
			}

			if metadata.Name != tt.table {
				t.Errorf("expected table %s, got %s", tt.table, metadata.Name)
			}

			if len(metadata.Columns) != tt.wantColumns {
				t.Errorf("expected %d columns, got %d", tt.wantColumns, len(metadata.Columns))
			}

			if metadata.Type != tt.wantTableType {
				t.Errorf("expected table type %s, got %s", tt.wantTableType, metadata.Type)
			}

			if len(metadata.Partitions) != tt.wantPartitions {
				t.Errorf("expected %d partitions, got %d", tt.wantPartitions, len(metadata.Partitions))
			}
		})
	}
}


// TestParseDescribeFormattedWithStorage tests parsing storage information
func TestParseDescribeFormattedWithStorage(t *testing.T) {
	rows := [][]string{
		{"# col_name", "data_type", "comment"},
		{"id", "int", ""},
		{"data", "string", ""},
		{"", "", ""},
		{"# Detailed Table Information", "", ""},
		{"Table Type:", "MANAGED_TABLE", ""},
		{"", "", ""},
		{"# Storage Information", "", ""},
		{"InputFormat:", "org.apache.hadoop.hive.ql.io.parquet.MapredParquetInputFormat", ""},
		{"OutputFormat:", "org.apache.hadoop.hive.ql.io.parquet.MapredParquetOutputFormat", ""},
		{"SerDe Library:", "org.apache.hadoop.hive.ql.io.parquet.serde.ParquetHiveSerDe", ""},
		{"Location:", "hdfs://namenode:8020/user/hive/warehouse/test_db.db/test_table", ""},
		{"Compressed:", "Yes", ""},
	}

	metadata, err := ParseDescribeFormatted(rows, "hive", "test_db", "test_table")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if metadata.Storage == nil {
		t.Fatal("expected storage info, got nil")
	}

	if metadata.Storage.Format != "parquet" {
		t.Errorf("expected format 'parquet', got %s", metadata.Storage.Format)
	}

	if metadata.Storage.InputFormat != "org.apache.hadoop.hive.ql.io.parquet.MapredParquetInputFormat" {
		t.Errorf("unexpected input format: %s", metadata.Storage.InputFormat)
	}

	if metadata.Storage.OutputFormat != "org.apache.hadoop.hive.ql.io.parquet.MapredParquetOutputFormat" {
		t.Errorf("unexpected output format: %s", metadata.Storage.OutputFormat)
	}

	if metadata.Storage.SerDe != "org.apache.hadoop.hive.ql.io.parquet.serde.ParquetHiveSerDe" {
		t.Errorf("unexpected serde: %s", metadata.Storage.SerDe)
	}

	if metadata.Storage.Location != "hdfs://namenode:8020/user/hive/warehouse/test_db.db/test_table" {
		t.Errorf("unexpected location: %s", metadata.Storage.Location)
	}

	if !metadata.Storage.Compressed {
		t.Error("expected compressed to be true")
	}
}

// TestNormalizeHiveType tests type normalization
func TestNormalizeHiveType(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"int", "INTEGER"},
		{"INT", "INTEGER"},
		{"bigint", "INTEGER"},
		{"tinyint", "INTEGER"},
		{"smallint", "INTEGER"},
		{"float", "FLOAT"},
		{"double", "FLOAT"},
		{"double precision", "FLOAT"},
		{"decimal", "DECIMAL"},
		{"decimal(10,2)", "DECIMAL"},
		{"numeric", "DECIMAL"},
		{"string", "STRING"},
		{"varchar", "STRING"},
		{"varchar(100)", "STRING"},
		{"char", "STRING"},
		{"char(50)", "STRING"},
		{"date", "DATE"},
		{"timestamp", "TIMESTAMP"},
		{"timestamp with local time zone", "TIMESTAMP"},
		{"binary", "BINARY"},
		{"boolean", "BOOLEAN"},
		{"array", "ARRAY"},
		{"array<string>", "ARRAY"},
		{"map", "MAP"},
		{"map<string,int>", "MAP"},
		{"struct", "STRUCT"},
		{"struct<name:string,age:int>", "STRUCT"},
		{"uniontype", "UNION"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := normalizeHiveType(tt.input)
			if result != tt.expected {
				t.Errorf("normalizeHiveType(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

// TestParseTypeParams tests type parameter extraction
func TestParseTypeParams(t *testing.T) {
	tests := []struct {
		dataType      string
		wantLength    *int
		wantPrecision *int
		wantScale     *int
	}{
		{
			dataType:   "varchar(100)",
			wantLength: intPtr(100),
		},
		{
			dataType:   "char(50)",
			wantLength: intPtr(50),
		},
		{
			dataType:      "decimal(10,2)",
			wantPrecision: intPtr(10),
			wantScale:     intPtr(2),
		},
		{
			dataType:      "decimal(18)",
			wantPrecision: intPtr(18),
		},
		{
			dataType:      "numeric(15,5)",
			wantPrecision: intPtr(15),
			wantScale:     intPtr(5),
		},
		{
			dataType: "int",
		},
		{
			dataType: "string",
		},
	}

	for _, tt := range tests {
		t.Run(tt.dataType, func(t *testing.T) {
			col := &collector.Column{}
			parseTypeParams(tt.dataType, col)

			if !intPtrEqual(col.Length, tt.wantLength) {
				t.Errorf("length: got %v, want %v", intPtrValue(col.Length), intPtrValue(tt.wantLength))
			}
			if !intPtrEqual(col.Precision, tt.wantPrecision) {
				t.Errorf("precision: got %v, want %v", intPtrValue(col.Precision), intPtrValue(tt.wantPrecision))
			}
			if !intPtrEqual(col.Scale, tt.wantScale) {
				t.Errorf("scale: got %v, want %v", intPtrValue(col.Scale), intPtrValue(tt.wantScale))
			}
		})
	}
}

// TestMapHiveTableType tests table type mapping
func TestMapHiveTableType(t *testing.T) {
	tests := []struct {
		input    string
		expected collector.TableType
	}{
		{"MANAGED_TABLE", collector.TableTypeTable},
		{"TABLE", collector.TableTypeTable},
		{"EXTERNAL_TABLE", collector.TableTypeExternalTable},
		{"VIRTUAL_VIEW", collector.TableTypeView},
		{"VIEW", collector.TableTypeView},
		{"MATERIALIZED_VIEW", collector.TableTypeMaterializedView},
		{"managed_table", collector.TableTypeTable},
		{"external_table", collector.TableTypeExternalTable},
		{"UNKNOWN", collector.TableTypeTable},
		{"", collector.TableTypeTable},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := mapHiveTableType(tt.input)
			if result != tt.expected {
				t.Errorf("mapHiveTableType(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

// TestExtractStorageFormat tests storage format extraction
func TestExtractStorageFormat(t *testing.T) {
	tests := []struct {
		inputFormat string
		expected    string
	}{
		{"org.apache.hadoop.hive.ql.io.parquet.MapredParquetInputFormat", "parquet"},
		{"org.apache.hadoop.hive.ql.io.orc.OrcInputFormat", "orc"},
		{"org.apache.hadoop.hive.ql.io.avro.AvroContainerInputFormat", "avro"},
		{"org.apache.hadoop.hive.ql.io.RCFileInputFormat", "rcfile"},
		{"org.apache.hadoop.mapred.SequenceFileInputFormat", "sequencefile"},
		{"org.apache.hadoop.mapred.TextInputFormat", "text"},
		{"org.apache.hive.hcatalog.data.JsonSerDe", "json"},
		{"org.apache.hadoop.hive.ql.io.SomeUnknownFormat", ""},
	}

	for _, tt := range tests {
		t.Run(tt.inputFormat, func(t *testing.T) {
			result := extractStorageFormat(tt.inputFormat)
			if result != tt.expected {
				t.Errorf("extractStorageFormat(%q) = %q, want %q", tt.inputFormat, result, tt.expected)
			}
		})
	}
}

// TestParsePartitionSpec tests partition specification parsing
func TestParsePartitionSpec(t *testing.T) {
	tests := []struct {
		spec     string
		expected map[string]string
	}{
		{
			spec:     "dt=2023-01-01",
			expected: map[string]string{"dt": "2023-01-01"},
		},
		{
			spec:     "dt=2023-01-01/region=us",
			expected: map[string]string{"dt": "2023-01-01", "region": "us"},
		},
		{
			spec:     "year=2023/month=01/day=15",
			expected: map[string]string{"year": "2023", "month": "01", "day": "15"},
		},
		{
			spec:     "",
			expected: map[string]string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.spec, func(t *testing.T) {
			result := ParsePartitionSpec(tt.spec)
			if len(result) != len(tt.expected) {
				t.Errorf("expected %d entries, got %d", len(tt.expected), len(result))
				return
			}
			for k, v := range tt.expected {
				if result[k] != v {
					t.Errorf("expected %s=%s, got %s=%s", k, v, k, result[k])
				}
			}
		})
	}
}

// TestParseDescribeFormattedWithComment tests parsing table comment
func TestParseDescribeFormattedWithComment(t *testing.T) {
	rows := [][]string{
		{"# col_name", "data_type", "comment"},
		{"id", "int", "Primary key"},
		{"", "", ""},
		{"# Detailed Table Information", "", ""},
		{"Table Type:", "MANAGED_TABLE", ""},
		{"Comment:", "This is a test table", ""},
	}

	metadata, err := ParseDescribeFormatted(rows, "hive", "default", "test")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if metadata.Comment != "This is a test table" {
		t.Errorf("expected comment 'This is a test table', got %q", metadata.Comment)
	}
}

// Helper functions

func intPtr(i int) *int {
	return &i
}

func intPtrEqual(a, b *int) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	return *a == *b
}

func intPtrValue(p *int) string {
	if p == nil {
		return "nil"
	}
	return string(rune(*p + '0'))
}
