package tests

import (
	"go-metadata/internal/lineage"
)

// MockCatalog is a mock implementation of Catalog for testing.
type MockCatalog struct {
	schemas map[string]*lineage.TableSchema
}

// NewMockCatalog creates a new mock catalog.
func NewMockCatalog() *MockCatalog {
	return &MockCatalog{
		schemas: make(map[string]*lineage.TableSchema),
	}
}

// AddTable adds a table schema to the mock catalog.
func (m *MockCatalog) AddTable(db, table string, columns []string) {
	key := db + "." + table
	m.schemas[key] = &lineage.TableSchema{
		Database: db,
		Table:    table,
		Columns:  columns,
	}
}

// GetTableSchema returns the schema for the specified table.
func (m *MockCatalog) GetTableSchema(db, table string) (*lineage.TableSchema, error) {
	key := db + "." + table
	if schema, ok := m.schemas[key]; ok {
		return schema, nil
	}
	return nil, lineage.ErrTableNotFound
}
