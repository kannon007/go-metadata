package metadata

import (
	"go-metadata/internal/lineage"
)

// MetadataBuilder provides a fluent interface for building metadata context.
type MetadataBuilder struct {
	provider  *MemoryProvider
	ddlParser *DDLParser
}

// NewMetadataBuilder creates a new metadata builder.
func NewMetadataBuilder() *MetadataBuilder {
	return &MetadataBuilder{
		provider:  NewMemoryProvider(),
		ddlParser: NewDDLParser(),
	}
}

// WithDefaultDatabase sets the default database name.
func (b *MetadataBuilder) WithDefaultDatabase(db string) *MetadataBuilder {
	b.provider.SetDefaultDatabase(db)
	return b
}

// AddTable adds a simple table with column names.
func (b *MetadataBuilder) AddTable(database, table string, columns []string) *MetadataBuilder {
	_ = b.provider.AddTable(database, table, columns)
	return b
}

// AddTableSchema adds a full table schema.
func (b *MetadataBuilder) AddTableSchema(schema *TableSchema) *MetadataBuilder {
	_ = b.provider.AddTableSchema(schema)
	return b
}

// LoadFromJSON loads table schemas from JSON bytes.
func (b *MetadataBuilder) LoadFromJSON(data []byte) *MetadataBuilder {
	_ = b.provider.LoadFromJSONBytes(data)
	return b
}

// LoadFromJSONFile loads table schemas from a JSON file.
func (b *MetadataBuilder) LoadFromJSONFile(filename string) *MetadataBuilder {
	_ = b.provider.LoadFromJSON(filename)
	return b
}

// LoadFromDDL parses DDL statements and adds extracted schemas.
func (b *MetadataBuilder) LoadFromDDL(ddl string) *MetadataBuilder {
	schemas, err := b.ddlParser.ParseMultipleDDL(ddl)
	if err != nil {
		return b
	}
	for _, schema := range schemas {
		_ = b.provider.AddTableSchema(schema)
	}
	return b
}

// Build returns the metadata provider.
func (b *MetadataBuilder) Build() *MemoryProvider {
	return b.provider
}

// BuildCatalog returns a lineage.Catalog adapter.
func (b *MetadataBuilder) BuildCatalog() lineage.Catalog {
	return NewCatalogAdapter(b.provider)
}

// BuildAnalyzer creates a lineage.Analyzer with the built metadata.
func (b *MetadataBuilder) BuildAnalyzer() *lineage.Analyzer {
	return lineage.NewAnalyzer(b.BuildCatalog())
}
