package metadata

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"
)

// ErrTableNotFound is returned when a table is not found in the catalog.
var ErrTableNotFound = errors.New("table not found")

// Provider provides table schema metadata for lineage analysis.
type Provider interface {
	// GetTableSchema returns the schema for the specified table.
	GetTableSchema(database, table string) (*TableSchema, error)

	// ListTables returns all table names in the specified database.
	ListTables(database string) ([]string, error)

	// AddTableSchema adds or updates a table schema.
	AddTableSchema(schema *TableSchema) error
}

// MemoryProvider is an in-memory metadata provider.
type MemoryProvider struct {
	// databases maps database name to DatabaseSchema
	databases map[string]*DatabaseSchema
	// defaultDatabase is used when database is not specified
	defaultDatabase string
}

// NewMemoryProvider creates a new in-memory metadata provider.
func NewMemoryProvider() *MemoryProvider {
	return &MemoryProvider{
		databases:       make(map[string]*DatabaseSchema),
		defaultDatabase: "default",
	}
}

// SetDefaultDatabase sets the default database name.
func (p *MemoryProvider) SetDefaultDatabase(db string) {
	p.defaultDatabase = db
}

// GetTableSchema returns the schema for the specified table.
func (p *MemoryProvider) GetTableSchema(database, table string) (*TableSchema, error) {
	if database == "" {
		database = p.defaultDatabase
	}

	// Normalize table name (case-insensitive)
	table = strings.ToLower(table)

	db, ok := p.databases[database]
	if !ok {
		// Try to find table in any database
		for _, db := range p.databases {
			if schema := db.GetTable(table); schema != nil {
				return schema, nil
			}
		}
		return nil, fmt.Errorf("%w: %s.%s", ErrTableNotFound, database, table)
	}

	schema := db.GetTable(table)
	if schema == nil {
		return nil, fmt.Errorf("%w: %s.%s", ErrTableNotFound, database, table)
	}
	return schema, nil
}

// ListTables returns all table names in the specified database.
func (p *MemoryProvider) ListTables(database string) ([]string, error) {
	if database == "" {
		database = p.defaultDatabase
	}

	db, ok := p.databases[database]
	if !ok {
		return nil, fmt.Errorf("database not found: %s", database)
	}

	tables := make([]string, 0, len(db.Tables))
	for name := range db.Tables {
		tables = append(tables, name)
	}
	return tables, nil
}

// AddTableSchema adds or updates a table schema.
func (p *MemoryProvider) AddTableSchema(schema *TableSchema) error {
	database := schema.Database
	if database == "" {
		database = p.defaultDatabase
	}

	// Normalize table name
	schema.Table = strings.ToLower(schema.Table)

	db, ok := p.databases[database]
	if !ok {
		db = NewDatabaseSchema(database)
		p.databases[database] = db
	}

	db.AddTable(schema)
	return nil
}

// AddTable is a convenience method to add a simple table with column names only.
func (p *MemoryProvider) AddTable(database, table string, columns []string) error {
	cols := make([]ColumnSchema, len(columns))
	for i, name := range columns {
		cols[i] = ColumnSchema{Name: name}
	}

	return p.AddTableSchema(&TableSchema{
		Database: database,
		Table:    table,
		Columns:  cols,
	})
}

// LoadFromJSON loads table schemas from a JSON file.
func (p *MemoryProvider) LoadFromJSON(filename string) error {
	data, err := os.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	return p.LoadFromJSONBytes(data)
}

// LoadFromJSONBytes loads table schemas from JSON bytes.
func (p *MemoryProvider) LoadFromJSONBytes(data []byte) error {
	var schemas []TableSchema
	if err := json.Unmarshal(data, &schemas); err != nil {
		// Try single schema
		var schema TableSchema
		if err2 := json.Unmarshal(data, &schema); err2 != nil {
			return fmt.Errorf("failed to parse JSON: %w", err)
		}
		schemas = []TableSchema{schema}
	}

	for i := range schemas {
		if err := p.AddTableSchema(&schemas[i]); err != nil {
			return err
		}
	}
	return nil
}

// ExportToJSON exports all table schemas to JSON bytes.
func (p *MemoryProvider) ExportToJSON() ([]byte, error) {
	var schemas []TableSchema
	for _, db := range p.databases {
		for _, table := range db.Tables {
			schemas = append(schemas, *table)
		}
	}
	return json.MarshalIndent(schemas, "", "  ")
}

// Clear removes all table schemas.
func (p *MemoryProvider) Clear() {
	p.databases = make(map[string]*DatabaseSchema)
}
