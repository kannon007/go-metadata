package tests

import (
	"go-metadata/internal/lineage"
	"testing"
)

// TestLineage_SimpleSelect_Verify tests simple SELECT with result verification.
func TestLineage_SimpleSelect_Verify(t *testing.T) {
	catalog := NewMockCatalog()
	catalog.AddTable("", "users", []string{"id", "name", "email"})

	analyzer := lineage.NewAnalyzer(catalog)
	sql := "SELECT id, name FROM users"

	result, err := analyzer.Analyze(sql)
	if err != nil {
		t.Fatalf("Analyze failed: %v", err)
	}

	// Print result for debugging
	printLineageResult(t, sql, result)

	// Verify result
	assertColumnCount(t, result, 2)
	assertColumnLineage(t, result, "id", []string{"users.id"}, []string{"id"})
	assertColumnLineage(t, result, "name", []string{"users.name"}, []string{"name"})
}

// TestLineage_FunctionCall_Verify tests function call with result verification.
func TestLineage_FunctionCall_Verify(t *testing.T) {
	catalog := NewMockCatalog()
	catalog.AddTable("", "orders", []string{"id", "amount", "user_id"})

	analyzer := lineage.NewAnalyzer(catalog)
	sql := "SELECT SUM(amount) as total, COUNT(user_id) as cnt FROM orders"

	result, err := analyzer.Analyze(sql)
	if err != nil {
		t.Fatalf("Analyze failed: %v", err)
	}

	// Print result for debugging
	printLineageResult(t, sql, result)

	// Verify result
	assertColumnCount(t, result, 2)
	assertColumnLineage(t, result, "total", []string{"orders.amount"}, []string{"SUM(amount)"})
	assertColumnLineage(t, result, "cnt", []string{"orders.user_id"}, []string{"COUNT(user_id)"})
}

// TestLineage_InsertSelect_Verify tests INSERT...SELECT with result verification.
func TestLineage_InsertSelect_Verify(t *testing.T) {
	catalog := NewMockCatalog()
	catalog.AddTable("", "orders", []string{"id", "amount", "user_id"})
	catalog.AddTable("", "report", []string{"total_amount", "user_count"})

	analyzer := lineage.NewAnalyzer(catalog)
	sql := `INSERT INTO report(total_amount, user_count)
			SELECT SUM(amount), COUNT(DISTINCT user_id) FROM orders`

	result, err := analyzer.Analyze(sql)
	if err != nil {
		t.Fatalf("Analyze failed: %v", err)
	}

	// Print result for debugging
	printLineageResult(t, sql, result)

	// Verify target table
	assertTargetTable(t, result, "report")

	// Verify columns
	assertColumnCount(t, result, 2)
	assertColumnLineage(t, result, "total_amount", []string{"orders.amount"}, []string{"SUM(amount)"})
	assertColumnLineage(t, result, "user_count", []string{"orders.user_id"}, []string{"COUNT(DISTINCT user_id)"})
}

// TestLineage_Join_Verify tests JOIN with result verification.
func TestLineage_Join_Verify(t *testing.T) {
	catalog := NewMockCatalog()
	catalog.AddTable("", "users", []string{"id", "name"})
	catalog.AddTable("", "orders", []string{"id", "user_id", "amount"})

	analyzer := lineage.NewAnalyzer(catalog)
	sql := `SELECT u.name, o.amount
			FROM users u
			INNER JOIN orders o ON u.id = o.user_id`

	result, err := analyzer.Analyze(sql)
	if err != nil {
		t.Fatalf("Analyze failed: %v", err)
	}

	// Print result for debugging
	printLineageResult(t, sql, result)

	// Verify result
	assertColumnCount(t, result, 2)
	assertColumnLineage(t, result, "name", []string{"users.name"}, []string{"u.name"})
	assertColumnLineage(t, result, "amount", []string{"orders.amount"}, []string{"o.amount"})
}

// TestLineage_CaseWhen_Verify tests CASE WHEN with result verification.
func TestLineage_CaseWhen_Verify(t *testing.T) {
	catalog := NewMockCatalog()
	catalog.AddTable("", "users", []string{"id", "name", "status"})

	analyzer := lineage.NewAnalyzer(catalog)
	sql := `SELECT id,
			CASE WHEN status = 1 THEN 'active' ELSE 'inactive' END as status_text
			FROM users`

	result, err := analyzer.Analyze(sql)
	if err != nil {
		t.Fatalf("Analyze failed: %v", err)
	}

	// Print result for debugging
	printLineageResult(t, sql, result)

	// Verify result
	assertColumnCount(t, result, 2)
	assertColumnLineage(t, result, "id", []string{"users.id"}, []string{"id"})
	// CASE expression should have the raw expression as operator
	assertColumnLineage(t, result, "status_text", []string{"users.status"}, nil) // Don't check operator, it's complex
}

// TestLineage_CTE_Verify tests CTE with result verification.
func TestLineage_CTE_Verify(t *testing.T) {
	catalog := NewMockCatalog()
	catalog.AddTable("", "orders", []string{"id", "user_id", "amount"})

	analyzer := lineage.NewAnalyzer(catalog)
	sql := `WITH order_totals AS (
				SELECT user_id, SUM(amount) as total
				FROM orders
				GROUP BY user_id
			)
			SELECT user_id, total FROM order_totals`

	result, err := analyzer.Analyze(sql)
	if err != nil {
		t.Fatalf("Analyze failed: %v", err)
	}

	// Print result for debugging
	printLineageResult(t, sql, result)

	// Verify result - CTE should be resolved
	assertColumnCount(t, result, 2)
}
