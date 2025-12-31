package tests

import (
	"go-metadata/internal/lineage"
	"testing"
)

func setupSQLServerCatalog() *MockCatalog {
	catalog := NewMockCatalog()
	catalog.AddTable("dbo", "users", []string{"id", "name", "email", "created_at"})
	catalog.AddTable("dbo", "orders", []string{"id", "user_id", "amount", "order_date"})
	return catalog
}

func TestSQLServer_SelectTop(t *testing.T) {
	catalog := setupSQLServerCatalog()
	parser := lineage.NewParser(catalog)

	sql := `SELECT TOP 10 id, name, email
			FROM users
			ORDER BY created_at DESC`
	result, err := parser.Parse(sql)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}
	if result == nil {
		t.Fatal("Parse returned nil result")
	}
}

func TestSQLServer_SelectWithOffset(t *testing.T) {
	catalog := setupSQLServerCatalog()
	parser := lineage.NewParser(catalog)

	sql := `SELECT id, name, email
			FROM users
			ORDER BY created_at DESC
			OFFSET 10 ROWS FETCH NEXT 20 ROWS ONLY`
	result, err := parser.Parse(sql)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}
	if result == nil {
		t.Fatal("Parse returned nil result")
	}
}

func TestSQLServer_CrossApply(t *testing.T) {
	catalog := setupSQLServerCatalog()
	parser := lineage.NewParser(catalog)

	sql := `SELECT u.id, u.name, o.amount
			FROM users u
			CROSS APPLY (
				SELECT TOP 1 amount FROM orders WHERE user_id = u.id ORDER BY order_date DESC
			) o`
	result, err := parser.Parse(sql)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}
	if result == nil {
		t.Fatal("Parse returned nil result")
	}
}

func TestSQLServer_Pivot(t *testing.T) {
	catalog := setupSQLServerCatalog()
	parser := lineage.NewParser(catalog)

	sql := `SELECT user_id, [2023], [2024]
			FROM (SELECT user_id, YEAR(order_date) as yr, amount FROM orders) src
			PIVOT (SUM(amount) FOR yr IN ([2023], [2024])) pvt`
	result, err := parser.Parse(sql)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}
	if result == nil {
		t.Fatal("Parse returned nil result")
	}
}
