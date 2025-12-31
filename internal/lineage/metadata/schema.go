// Package metadata provides table schema metadata for SQL lineage analysis.
package metadata

// ColumnSchema represents the schema of a database column.
type ColumnSchema struct {
	Name        string `json:"name"`
	DataType    string `json:"data_type"`
	Nullable    bool   `json:"nullable"`
	PrimaryKey  bool   `json:"primary_key"`
	Comment     string `json:"comment,omitempty"`
	DefaultExpr string `json:"default_expr,omitempty"`
}

// TableSchema represents the schema of a database table.
type TableSchema struct {
	Database   string         `json:"database,omitempty"`
	Schema     string         `json:"schema,omitempty"` // For databases that support schema (e.g., PostgreSQL)
	Table      string         `json:"table"`
	Columns    []ColumnSchema `json:"columns"`
	PrimaryKey []string       `json:"primary_key,omitempty"`
	Comment    string         `json:"comment,omitempty"`
	TableType  string         `json:"table_type,omitempty"` // "TABLE", "VIEW", "EXTERNAL"
}

// GetColumnNames returns the list of column names.
func (t *TableSchema) GetColumnNames() []string {
	names := make([]string, len(t.Columns))
	for i, col := range t.Columns {
		names[i] = col.Name
	}
	return names
}

// GetColumn returns the column schema by name.
func (t *TableSchema) GetColumn(name string) *ColumnSchema {
	for i := range t.Columns {
		if t.Columns[i].Name == name {
			return &t.Columns[i]
		}
	}
	return nil
}

// HasColumn checks if the table has a column with the given name.
func (t *TableSchema) HasColumn(name string) bool {
	return t.GetColumn(name) != nil
}

// DatabaseSchema represents a collection of tables in a database.
type DatabaseSchema struct {
	Database string                  `json:"database"`
	Tables   map[string]*TableSchema `json:"tables"`
}

// NewDatabaseSchema creates a new empty database schema.
func NewDatabaseSchema(database string) *DatabaseSchema {
	return &DatabaseSchema{
		Database: database,
		Tables:   make(map[string]*TableSchema),
	}
}

// AddTable adds a table schema to the database.
func (d *DatabaseSchema) AddTable(table *TableSchema) {
	d.Tables[table.Table] = table
}

// GetTable returns the table schema by name.
func (d *DatabaseSchema) GetTable(name string) *TableSchema {
	return d.Tables[name]
}
