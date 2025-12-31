package tests

import (
	"go-metadata/internal/lineage"
	"testing"
)

func setupMySQLCatalog() *MockCatalog {
	catalog := NewMockCatalog()
	catalog.AddTable("test", "users", []string{"id", "name", "email", "status", "created_at"})
	catalog.AddTable("test", "orders", []string{"id", "user_id", "amount", "status"})
	catalog.AddTable("test", "report", []string{"total_amount", "user_count"})
	return catalog
}

func TestMySQL_SimpleSelect(t *testing.T) {
	catalog := setupMySQLCatalog()
	parser := lineage.NewParser(catalog)

	result, err := parser.Parse("SELECT id, name FROM users")
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}
	if result == nil {
		t.Fatal("Parse returned nil result")
	}
}

func TestMySQL_SelectWithAlias(t *testing.T) {
	catalog := setupMySQLCatalog()
	parser := lineage.NewParser(catalog)

	result, err := parser.Parse("SELECT u.id, u.name FROM users u")
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}
	if result == nil {
		t.Fatal("Parse returned nil result")
	}
}

func TestMySQL_SelectWithJoin(t *testing.T) {
	catalog := setupMySQLCatalog()
	parser := lineage.NewParser(catalog)

	sql := `SELECT u.id, u.name, o.amount
			FROM users u
			INNER JOIN orders o ON u.id = o.user_id`
	result, err := parser.Parse(sql)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}
	if result == nil {
		t.Fatal("Parse returned nil result")
	}
}

func TestMySQL_SelectWithLeftJoin(t *testing.T) {
	catalog := setupMySQLCatalog()
	parser := lineage.NewParser(catalog)

	sql := `SELECT u.id, u.name, o.amount
			FROM users u
			LEFT JOIN orders o ON u.id = o.user_id`
	result, err := parser.Parse(sql)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}
	if result == nil {
		t.Fatal("Parse returned nil result")
	}
}

func TestMySQL_InsertSelect(t *testing.T) {
	catalog := setupMySQLCatalog()
	parser := lineage.NewParser(catalog)

	sql := `INSERT INTO report(total_amount, user_count)
			SELECT SUM(amount), COUNT(DISTINCT user_id) FROM orders`
	result, err := parser.Parse(sql)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}
	if result == nil {
		t.Fatal("Parse returned nil result")
	}
}

func TestMySQL_SelectWithSubquery(t *testing.T) {
	catalog := setupMySQLCatalog()
	parser := lineage.NewParser(catalog)

	sql := `SELECT id, name FROM users 
			WHERE id IN (SELECT user_id FROM orders WHERE amount > 100)`
	result, err := parser.Parse(sql)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}
	if result == nil {
		t.Fatal("Parse returned nil result")
	}
}

func TestMySQL_SelectWithAggregation(t *testing.T) {
	catalog := setupMySQLCatalog()
	parser := lineage.NewParser(catalog)

	sql := `SELECT user_id, SUM(amount) as total, COUNT(*) as cnt
			FROM orders GROUP BY user_id`
	result, err := parser.Parse(sql)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}
	if result == nil {
		t.Fatal("Parse returned nil result")
	}
}

func TestMySQL_SelectWithCaseWhen(t *testing.T) {
	catalog := setupMySQLCatalog()
	parser := lineage.NewParser(catalog)

	sql := `SELECT id, 
			CASE WHEN status = 'active' THEN name ELSE 'unknown' END as display_name
			FROM users`
	result, err := parser.Parse(sql)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}
	if result == nil {
		t.Fatal("Parse returned nil result")
	}
}

func TestMySQL_SelectWithUnion(t *testing.T) {
	catalog := setupMySQLCatalog()
	parser := lineage.NewParser(catalog)

	sql := `SELECT id, name FROM users WHERE status = 'active'
			UNION ALL
			SELECT id, name FROM users WHERE status = 'pending'`
	result, err := parser.Parse(sql)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}
	if result == nil {
		t.Fatal("Parse returned nil result")
	}
}
