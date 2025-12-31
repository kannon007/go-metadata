package tests

import (
	"go-metadata/internal/lineage"
	"testing"
)

func TestNewParser(t *testing.T) {
	catalog := NewMockCatalog()
	parser := lineage.NewParser(catalog)
	if parser == nil {
		t.Fatal("NewParser returned nil")
	}
}

func TestMockCatalog(t *testing.T) {
	catalog := NewMockCatalog()
	catalog.AddTable("test", "orders", []string{"id", "user_id", "amount"})

	schema, err := catalog.GetTableSchema("test", "orders")
	if err != nil {
		t.Fatalf("GetTableSchema failed: %v", err)
	}
	if len(schema.Columns) != 3 {
		t.Errorf("Expected 3 columns, got %d", len(schema.Columns))
	}
}

func TestCatalogTableNotFound(t *testing.T) {
	catalog := NewMockCatalog()
	_, err := catalog.GetTableSchema("test", "nonexistent")
	if err != lineage.ErrTableNotFound {
		t.Errorf("Expected ErrTableNotFound, got %v", err)
	}
}
