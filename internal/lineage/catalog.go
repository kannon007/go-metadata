package lineage

// TableSchema represents the schema of a database table.
type TableSchema struct {
	Database string   `json:"database"`
	Table    string   `json:"table"`
	Columns  []string `json:"columns"`
}

// Catalog provides table schema information for lineage resolution.
type Catalog interface {
	// GetTableSchema returns the schema for the specified table.
	// Returns ErrTableNotFound if the table does not exist.
	GetTableSchema(db, table string) (*TableSchema, error)
}
