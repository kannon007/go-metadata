package tests

import (
	"go-metadata/internal/lineage"
	"testing"
)

func setupPgSQLCatalog() *MockCatalog {
	catalog := NewMockCatalog()
	catalog.AddTable("public", "users", []string{"id", "name", "email", "status"})
	catalog.AddTable("public", "employees", []string{"id", "name", "department", "salary"})
	catalog.AddTable("public", "orders", []string{"id", "user_id", "amount"})
	return catalog
}

func TestPgSQL_CTE(t *testing.T) {
	catalog := setupPgSQLCatalog()
	parser := lineage.NewParser(catalog)

	sql := `WITH active_users AS (
				SELECT id, name FROM users WHERE status = 'active'
			)
			SELECT id, name FROM active_users`
	result, err := parser.Parse(sql)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}
	if result == nil {
		t.Fatal("Parse returned nil result")
	}
}

func TestPgSQL_RecursiveCTE(t *testing.T) {
	catalog := setupPgSQLCatalog()
	parser := lineage.NewParser(catalog)

	sql := `WITH RECURSIVE subordinates AS (
				SELECT id, name, department FROM employees WHERE id = 1
				UNION ALL
				SELECT e.id, e.name, e.department 
				FROM employees e
				INNER JOIN subordinates s ON e.department = s.department
			)
			SELECT id, name FROM subordinates`
	result, err := parser.Parse(sql)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}
	if result == nil {
		t.Fatal("Parse returned nil result")
	}
}

func TestPgSQL_WindowFunction(t *testing.T) {
	catalog := setupPgSQLCatalog()
	parser := lineage.NewParser(catalog)

	sql := `SELECT id, name,
			ROW_NUMBER() OVER (PARTITION BY department ORDER BY salary DESC) as rank
			FROM employees`
	result, err := parser.Parse(sql)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}
	if result == nil {
		t.Fatal("Parse returned nil result")
	}
}

func TestPgSQL_WindowFunctionLag(t *testing.T) {
	catalog := setupPgSQLCatalog()
	parser := lineage.NewParser(catalog)

	sql := `SELECT id, salary,
			LAG(salary, 1) OVER (ORDER BY id) as prev_salary,
			LEAD(salary, 1) OVER (ORDER BY id) as next_salary
			FROM employees`
	result, err := parser.Parse(sql)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}
	if result == nil {
		t.Fatal("Parse returned nil result")
	}
}

func TestPgSQL_MultipleCTE(t *testing.T) {
	catalog := setupPgSQLCatalog()
	parser := lineage.NewParser(catalog)

	sql := `WITH 
			user_stats AS (
				SELECT user_id, SUM(amount) as total FROM orders GROUP BY user_id
			),
			top_users AS (
				SELECT user_id, total FROM user_stats WHERE total > 1000
			)
			SELECT u.id, u.name, t.total
			FROM users u
			INNER JOIN top_users t ON u.id = t.user_id`
	result, err := parser.Parse(sql)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}
	if result == nil {
		t.Fatal("Parse returned nil result")
	}
}
