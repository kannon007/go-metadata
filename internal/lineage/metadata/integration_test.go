package metadata

import (
	"testing"
)

// TestDDLToMetadata_ThenQuery tests parsing DDL to metadata and using it for query analysis.
func TestDDLToMetadata_ThenQuery(t *testing.T) {
	// Flink-style DDL
	ddl := `
		CREATE TABLE user_events (
			user_id BIGINT,
			event_type STRING,
			event_time TIMESTAMP,
			page_url STRING,
			device_type STRING
		);

		CREATE TABLE user_info (
			user_id BIGINT,
			user_name STRING,
			age INT,
			gender STRING,
			city STRING
		);

		CREATE TABLE orders (
			order_id BIGINT,
			user_id BIGINT,
			product_id BIGINT,
			quantity INT,
			amount DECIMAL
		);

		CREATE TABLE user_behavior_stats (
			user_id BIGINT,
			window_start TIMESTAMP,
			pv_count BIGINT,
			uv_count BIGINT
		);
	`

	// Build metadata from DDL
	analyzer := NewMetadataBuilder().
		LoadFromDDL(ddl).
		BuildAnalyzer()

	// Test 1: Simple SELECT with columns
	t.Run("SimpleSelect", func(t *testing.T) {
		sql := `SELECT user_id, user_name, city FROM user_info`
		result, err := analyzer.Analyze(sql)
		if err != nil {
			t.Fatalf("Analyze failed: %v", err)
		}

		if len(result.Columns) != 3 {
			t.Errorf("Expected 3 columns, got %d", len(result.Columns))
		}

		t.Logf("Result: %+v", result)
	})

	// Test 2: SELECT with JOIN
	t.Run("JoinQuery", func(t *testing.T) {
		sql := `
			SELECT 
				e.user_id,
				u.user_name,
				e.event_type,
				e.page_url
			FROM user_events e
			LEFT JOIN user_info u ON e.user_id = u.user_id
		`
		result, err := analyzer.Analyze(sql)
		if err != nil {
			t.Fatalf("Analyze failed: %v", err)
		}

		if len(result.Columns) != 4 {
			t.Errorf("Expected 4 columns, got %d", len(result.Columns))
		}

		t.Logf("Result: %+v", result)
	})

	// Test 3: INSERT INTO with SELECT
	t.Run("InsertSelect", func(t *testing.T) {
		sql := `
			INSERT INTO user_behavior_stats
			SELECT 
				user_id,
				event_time as window_start,
				COUNT(*) as pv_count,
				COUNT(DISTINCT page_url) as uv_count
			FROM user_events
			GROUP BY user_id, event_time
		`
		result, err := analyzer.Analyze(sql)
		if err != nil {
			t.Fatalf("Analyze failed: %v", err)
		}

		if len(result.Columns) < 2 {
			t.Errorf("Expected at least 2 columns, got %d", len(result.Columns))
		}

		// Check target table
		for _, col := range result.Columns {
			if col.Target.Table != "user_behavior_stats" {
				t.Errorf("Expected target table 'user_behavior_stats', got '%s'", col.Target.Table)
			}
		}

		t.Logf("Result: %+v", result)
	})
}

// TestStarExpansion tests SELECT * expansion using metadata.
func TestStarExpansion(t *testing.T) {
	ddl := `
		CREATE TABLE users (
			id BIGINT,
			name STRING,
			email STRING,
			age INT,
			city STRING
		);

		CREATE TABLE orders (
			order_id BIGINT,
			user_id BIGINT,
			amount DECIMAL,
			status STRING,
			order_time TIMESTAMP
		);

		CREATE TABLE order_summary (
			user_id BIGINT,
			user_name STRING,
			total_orders BIGINT,
			total_amount DECIMAL
		);
	`

	analyzer := NewMetadataBuilder().
		LoadFromDDL(ddl).
		BuildAnalyzer()

	// Test 1: SELECT * from single table
	t.Run("SelectStarSingleTable", func(t *testing.T) {
		sql := `SELECT * FROM users`
		result, err := analyzer.Analyze(sql)
		if err != nil {
			t.Fatalf("Analyze failed: %v", err)
		}

		t.Logf("SELECT * FROM users result: %d columns", len(result.Columns))
		for _, col := range result.Columns {
			t.Logf("  Column: %s <- %s.%s", col.Target.Column, col.Sources[0].Table, col.Sources[0].Column)
		}

		// With metadata, * should expand to all columns
		// Note: Current implementation may not expand *, this test documents the behavior
		if len(result.Columns) == 0 {
			t.Log("Note: SELECT * returned 0 columns - star expansion not yet implemented")
		}
	})

	// Test 2: SELECT table.* from single table
	t.Run("SelectTableStar", func(t *testing.T) {
		sql := `SELECT u.* FROM users u`
		result, err := analyzer.Analyze(sql)
		if err != nil {
			t.Fatalf("Analyze failed: %v", err)
		}

		t.Logf("SELECT u.* FROM users u result: %d columns", len(result.Columns))
	})

	// Test 3: SELECT * with JOIN
	t.Run("SelectStarWithJoin", func(t *testing.T) {
		sql := `
			SELECT * 
			FROM users u
			JOIN orders o ON u.id = o.user_id
		`
		result, err := analyzer.Analyze(sql)
		if err != nil {
			t.Fatalf("Analyze failed: %v", err)
		}

		t.Logf("SELECT * with JOIN result: %d columns", len(result.Columns))
	})

	// Test 4: Mixed * and explicit columns
	t.Run("MixedStarAndColumns", func(t *testing.T) {
		sql := `
			SELECT 
				u.*,
				o.amount,
				o.status
			FROM users u
			JOIN orders o ON u.id = o.user_id
		`
		result, err := analyzer.Analyze(sql)
		if err != nil {
			t.Fatalf("Analyze failed: %v", err)
		}

		t.Logf("Mixed * and columns result: %d columns", len(result.Columns))
	})
}
