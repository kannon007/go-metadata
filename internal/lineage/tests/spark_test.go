package tests

import (
	"go-metadata/internal/lineage"
	"testing"
)

func setupSparkCatalog() *MockCatalog {
	catalog := NewMockCatalog()
	catalog.AddTable("default", "orders", []string{"user_id", "items", "amount"})
	catalog.AddTable("default", "events", []string{"user_id", "event_type", "event_time", "properties"})
	catalog.AddTable("default", "users", []string{"id", "name", "tags"})
	return catalog
}

func TestSpark_Transform(t *testing.T) {
	catalog := setupSparkCatalog()
	parser := lineage.NewParser(catalog)

	sql := `SELECT user_id,
			TRANSFORM(items, x -> x.price * x.quantity) as item_totals
			FROM orders`
	result, err := parser.Parse(sql)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}
	if result == nil {
		t.Fatal("Parse returned nil result")
	}
}

func TestSpark_Aggregate(t *testing.T) {
	catalog := setupSparkCatalog()
	parser := lineage.NewParser(catalog)

	sql := `SELECT user_id,
			AGGREGATE(items, 0, (acc, x) -> acc + x.price) as total
			FROM orders`
	result, err := parser.Parse(sql)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}
	if result == nil {
		t.Fatal("Parse returned nil result")
	}
}

func TestSpark_Filter(t *testing.T) {
	catalog := setupSparkCatalog()
	parser := lineage.NewParser(catalog)

	sql := `SELECT user_id,
			FILTER(items, x -> x.price > 100) as expensive_items
			FROM orders`
	result, err := parser.Parse(sql)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}
	if result == nil {
		t.Fatal("Parse returned nil result")
	}
}

func TestSpark_Explode(t *testing.T) {
	catalog := setupSparkCatalog()
	parser := lineage.NewParser(catalog)

	sql := `SELECT user_id, tag
			FROM users
			LATERAL VIEW EXPLODE(tags) t AS tag`
	result, err := parser.Parse(sql)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}
	if result == nil {
		t.Fatal("Parse returned nil result")
	}
}

func TestSpark_JsonExtract(t *testing.T) {
	catalog := setupSparkCatalog()
	parser := lineage.NewParser(catalog)

	sql := `SELECT user_id,
			GET_JSON_OBJECT(properties, '$.browser') as browser,
			GET_JSON_OBJECT(properties, '$.os') as os
			FROM events`
	result, err := parser.Parse(sql)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}
	if result == nil {
		t.Fatal("Parse returned nil result")
	}
}
