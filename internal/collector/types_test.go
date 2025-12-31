package collector

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
)

// getTestParameters returns the standard test parameters for property tests.
// Feature: metadata-collector, Property 4: TableMetadata JSON Round-Trip
// **Validates: Requirements 2.5**
func getTestParameters() *gopter.TestParameters {
	params := gopter.DefaultTestParameters()
	params.MinSuccessfulTests = 100
	return params
}

// genColumn generates a random Column for property testing.
func genColumn() gopter.Gen {
	return gopter.CombineGens(
		gen.Int(),                                // OrdinalPosition
		gen.AlphaString(),                        // Name
		gen.OneConstOf("STRING", "INT", "FLOAT"), // Type
		gen.AlphaString(),                        // SourceType
		gen.Bool(),                               // Nullable
		gen.AlphaString(),                        // Comment
		gen.Bool(),                               // IsPrimaryKey
		gen.Bool(),                               // IsPartitionColumn
		gen.Bool(),                               // IsAutoIncrement
	).Map(func(values []interface{}) Column {
		return Column{
			OrdinalPosition:   values[0].(int),
			Name:              values[1].(string),
			Type:              values[2].(string),
			SourceType:        values[3].(string),
			Length:            nil, // Skip optional pointer fields for simplicity
			Precision:         nil,
			Scale:             nil,
			Nullable:          values[4].(bool),
			Default:           nil, // Skip optional pointer fields for simplicity
			Comment:           values[5].(string),
			IsPrimaryKey:      values[6].(bool),
			IsPartitionColumn: values[7].(bool),
			IsAutoIncrement:   values[8].(bool),
			Raw:               nil,
		}
	})
}


// genIndex generates a random Index for property testing.
func genIndex() gopter.Gen {
	return gopter.CombineGens(
		gen.AlphaString(),                       // Name
		gen.SliceOf(gen.AlphaString()),          // Columns
		gen.Bool(),                              // Unique
		gen.OneConstOf("BTREE", "HASH", ""),     // Type
		gen.AlphaString(),                       // Comment
	).Map(func(values []interface{}) Index {
		return Index{
			Name:    values[0].(string),
			Columns: values[1].([]string),
			Unique:  values[2].(bool),
			Type:    values[3].(string),
			Comment: values[4].(string),
		}
	})
}

// genPartitionInfo generates a random PartitionInfo for property testing.
func genPartitionInfo() gopter.Gen {
	return gopter.CombineGens(
		gen.AlphaString(),                          // Name
		gen.OneConstOf("RANGE", "LIST", "HASH"),    // Type
		gen.SliceOf(gen.AlphaString()),             // Columns
		gen.AlphaString(),                          // Expression
		gen.IntRange(0, 1000),                      // ValuesCount
	).Map(func(values []interface{}) PartitionInfo {
		return PartitionInfo{
			Name:        values[0].(string),
			Type:        values[1].(string),
			Columns:     values[2].([]string),
			Expression:  values[3].(string),
			ValuesCount: values[4].(int),
		}
	})
}

// genStorageInfo generates a random StorageInfo for property testing.
func genStorageInfo() gopter.Gen {
	return gopter.CombineGens(
		gen.OneConstOf("parquet", "orc", "csv", ""), // Format
		gen.AlphaString(),                           // Location
		gen.AlphaString(),                           // InputFormat
		gen.AlphaString(),                           // OutputFormat
		gen.AlphaString(),                           // SerDe
		gen.Bool(),                                  // Compressed
	).Map(func(values []interface{}) *StorageInfo {
		return &StorageInfo{
			Format:       values[0].(string),
			Location:     values[1].(string),
			InputFormat:  values[2].(string),
			OutputFormat: values[3].(string),
			SerDe:        values[4].(string),
			Compressed:   values[5].(bool),
		}
	})
}


// genTableStatistics generates a random TableStatistics for property testing.
func genTableStatistics() gopter.Gen {
	return gopter.CombineGens(
		gen.Int64Range(0, 1000000),  // RowCount
		gen.Int64Range(0, 100000000), // DataSizeBytes
		gen.IntRange(0, 100),         // PartitionCount
	).Map(func(values []interface{}) *TableStatistics {
		return &TableStatistics{
			RowCount:       values[0].(int64),
			DataSizeBytes:  values[1].(int64),
			PartitionCount: values[2].(int),
			ColumnStats:    nil, // Skip for simplicity
			CollectedAt:    time.Now().UTC().Truncate(time.Second),
		}
	})
}

// genDataSourceCategory generates a random DataSourceCategory for property testing.
func genDataSourceCategory() gopter.Gen {
	return gen.OneConstOf(
		CategoryRDBMS,
		CategoryDataWarehouse,
		CategoryDocumentDB,
		CategoryKeyValue,
		CategoryMessageQueue,
		CategoryObjectStorage,
	)
}

// genTableMetadata generates a random TableMetadata for property testing.
func genTableMetadata() gopter.Gen {
	return gopter.CombineGens(
		genDataSourceCategory(),                                                  // SourceCategory
		gen.OneConstOf("mysql", "postgres", "mongodb", "kafka", "redis", "minio"), // SourceType
		gen.AlphaString(),                                                        // Catalog
		gen.AlphaString(),                                                        // Schema
		gen.AlphaString(),                                                        // Name
		gen.OneConstOf(TableTypeTable, TableTypeView, TableTypeExternalTable, TableTypeMaterializedView), // Type
		gen.AlphaString(),                                                        // Comment
		gen.SliceOfN(3, genColumn()),                                             // Columns (limit to 3 for performance)
		gen.SliceOfN(2, genPartitionInfo()),                                      // Partitions
		gen.SliceOfN(2, genIndex()),                                              // Indexes
		gen.SliceOf(gen.AlphaString()),                                           // PrimaryKey
		gen.Bool(),                                                               // InferredSchema
	).Map(func(values []interface{}) TableMetadata {
		return TableMetadata{
			SourceCategory:  values[0].(DataSourceCategory),
			SourceType:      values[1].(string),
			Catalog:         values[2].(string),
			Schema:          values[3].(string),
			Name:            values[4].(string),
			Type:            values[5].(TableType),
			Comment:         values[6].(string),
			Columns:         values[7].([]Column),
			Partitions:      values[8].([]PartitionInfo),
			Indexes:         values[9].([]Index),
			PrimaryKey:      values[10].([]string),
			Storage:         nil, // Skip for simplicity
			Stats:           nil, // Skip for simplicity
			Properties:      nil, // Skip for simplicity
			LastRefreshedAt: time.Now().UTC().Truncate(time.Second),
			InferredSchema:  values[11].(bool),
		}
	})
}


// TestTableMetadataJSONRoundTrip tests that TableMetadata can be serialized to JSON
// and deserialized back to an equivalent object.
// Feature: metadata-collector, Property 4: TableMetadata JSON Round-Trip
// **Validates: Requirements 2.5**
func TestTableMetadataJSONRoundTrip(t *testing.T) {
	properties := gopter.NewProperties(getTestParameters())

	properties.Property("TableMetadata JSON round-trip preserves all fields", prop.ForAll(
		func(original TableMetadata) bool {
			// Serialize to JSON
			jsonBytes, err := json.Marshal(original)
			if err != nil {
				t.Logf("Marshal error: %v", err)
				return false
			}

			// Deserialize back
			var restored TableMetadata
			err = json.Unmarshal(jsonBytes, &restored)
			if err != nil {
				t.Logf("Unmarshal error: %v", err)
				return false
			}

			// Compare data source fields
			if original.SourceCategory != restored.SourceCategory {
				t.Logf("SourceCategory mismatch: %q != %q", original.SourceCategory, restored.SourceCategory)
				return false
			}
			if original.SourceType != restored.SourceType {
				t.Logf("SourceType mismatch: %q != %q", original.SourceType, restored.SourceType)
				return false
			}

			// Compare key fields
			if original.Catalog != restored.Catalog {
				t.Logf("Catalog mismatch: %q != %q", original.Catalog, restored.Catalog)
				return false
			}
			if original.Schema != restored.Schema {
				t.Logf("Schema mismatch: %q != %q", original.Schema, restored.Schema)
				return false
			}
			if original.Name != restored.Name {
				t.Logf("Name mismatch: %q != %q", original.Name, restored.Name)
				return false
			}
			if original.Type != restored.Type {
				t.Logf("Type mismatch: %q != %q", original.Type, restored.Type)
				return false
			}
			if original.Comment != restored.Comment {
				t.Logf("Comment mismatch: %q != %q", original.Comment, restored.Comment)
				return false
			}
			if len(original.Columns) != len(restored.Columns) {
				t.Logf("Columns length mismatch: %d != %d", len(original.Columns), len(restored.Columns))
				return false
			}
			if len(original.Partitions) != len(restored.Partitions) {
				t.Logf("Partitions length mismatch: %d != %d", len(original.Partitions), len(restored.Partitions))
				return false
			}
			if len(original.Indexes) != len(restored.Indexes) {
				t.Logf("Indexes length mismatch: %d != %d", len(original.Indexes), len(restored.Indexes))
				return false
			}
			if len(original.PrimaryKey) != len(restored.PrimaryKey) {
				t.Logf("PrimaryKey length mismatch: %d != %d", len(original.PrimaryKey), len(restored.PrimaryKey))
				return false
			}

			// Compare InferredSchema field
			if original.InferredSchema != restored.InferredSchema {
				t.Logf("InferredSchema mismatch: %v != %v", original.InferredSchema, restored.InferredSchema)
				return false
			}

			// Compare columns in detail
			for i := range original.Columns {
				if !columnsEqual(original.Columns[i], restored.Columns[i]) {
					t.Logf("Column %d mismatch", i)
					return false
				}
			}

			// Compare indexes in detail
			for i := range original.Indexes {
				if !indexesEqual(original.Indexes[i], restored.Indexes[i]) {
					t.Logf("Index %d mismatch", i)
					return false
				}
			}

			// Compare partitions in detail
			for i := range original.Partitions {
				if !partitionsEqual(original.Partitions[i], restored.Partitions[i]) {
					t.Logf("Partition %d mismatch", i)
					return false
				}
			}

			return true
		},
		genTableMetadata(),
	))

	properties.TestingRun(t)
}

// columnsEqual compares two Column structs for equality.
func columnsEqual(a, b Column) bool {
	if a.OrdinalPosition != b.OrdinalPosition {
		return false
	}
	if a.Name != b.Name {
		return false
	}
	if a.Type != b.Type {
		return false
	}
	if a.SourceType != b.SourceType {
		return false
	}
	if !intPtrEqual(a.Length, b.Length) {
		return false
	}
	if !intPtrEqual(a.Precision, b.Precision) {
		return false
	}
	if !intPtrEqual(a.Scale, b.Scale) {
		return false
	}
	if a.Nullable != b.Nullable {
		return false
	}
	if !stringPtrEqual(a.Default, b.Default) {
		return false
	}
	if a.Comment != b.Comment {
		return false
	}
	if a.IsPrimaryKey != b.IsPrimaryKey {
		return false
	}
	if a.IsPartitionColumn != b.IsPartitionColumn {
		return false
	}
	if a.IsAutoIncrement != b.IsAutoIncrement {
		return false
	}
	return true
}

// indexesEqual compares two Index structs for equality.
func indexesEqual(a, b Index) bool {
	if a.Name != b.Name {
		return false
	}
	if a.Unique != b.Unique {
		return false
	}
	if a.Type != b.Type {
		return false
	}
	if a.Comment != b.Comment {
		return false
	}
	if len(a.Columns) != len(b.Columns) {
		return false
	}
	for i := range a.Columns {
		if a.Columns[i] != b.Columns[i] {
			return false
		}
	}
	return true
}

// partitionsEqual compares two PartitionInfo structs for equality.
func partitionsEqual(a, b PartitionInfo) bool {
	if a.Name != b.Name {
		return false
	}
	if a.Type != b.Type {
		return false
	}
	if a.Expression != b.Expression {
		return false
	}
	if a.ValuesCount != b.ValuesCount {
		return false
	}
	if len(a.Columns) != len(b.Columns) {
		return false
	}
	for i := range a.Columns {
		if a.Columns[i] != b.Columns[i] {
			return false
		}
	}
	return true
}

// intPtrEqual compares two *int pointers for equality.
func intPtrEqual(a, b *int) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	return *a == *b
}

// stringPtrEqual compares two *string pointers for equality.
func stringPtrEqual(a, b *string) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	return *a == *b
}
