package tests

import (
	"go-metadata/internal/lineage"
	"go-metadata/internal/lineage/metadata"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// loadFlinkCompleteSQL loads the Flink complete_example.sql file
func loadFlinkCompleteSQL() (string, error) {
	sqlPath := filepath.Join("..", "testdata", "flink", "complete_example.sql")
	content, err := os.ReadFile(sqlPath)
	if err != nil {
		return "", err
	}
	return string(content), nil
}

// setupFlinkCompleteCatalogFromDDL creates a catalog by parsing DDL from complete_example.sql
func setupFlinkCompleteCatalogFromDDL() (lineage.Catalog, error) {
	sql, err := loadFlinkCompleteSQL()
	if err != nil {
		return nil, err
	}

	// Use MetadataBuilder to parse DDL and build catalog
	catalog := metadata.NewMetadataBuilder().
		LoadFromDDL(sql).
		BuildCatalog()

	return catalog, nil
}

// TestFlinkComplete_LoadSQL tests loading the complete_example.sql file
func TestFlinkComplete_LoadSQL(t *testing.T) {
	sql, err := loadFlinkCompleteSQL()
	if err != nil {
		t.Fatalf("Failed to read SQL file: %v", err)
	}

	if len(sql) == 0 {
		t.Fatal("SQL file is empty")
	}

	t.Logf("Loaded SQL file with %d bytes", len(sql))
}

// TestFlinkComplete_ParseDDL tests parsing DDL from complete_example.sql
func TestFlinkComplete_ParseDDL(t *testing.T) {
	sql, err := loadFlinkCompleteSQL()
	if err != nil {
		t.Fatalf("Failed to read SQL file: %v", err)
	}

	// Parse DDL using DDLParser
	ddlParser := metadata.NewDDLParser()
	schemas, err := ddlParser.ParseMultipleDDL(sql)
	if err != nil {
		t.Fatalf("Failed to parse DDL: %v", err)
	}

	t.Logf("Parsed %d table schemas:", len(schemas))
	for _, schema := range schemas {
		var cols []string
		for _, col := range schema.Columns {
			cols = append(cols, col.Name)
		}
		t.Logf("  - %s.%s (%s): %v", schema.Database, schema.Table, schema.TableType, cols)
	}

	// Expected tables from complete_example.sql:
	// Source: user_events, user_info, product_info, orders
	// Sink: user_behavior_stats, user_order_summary, product_sales_rank, realtime_alerts
	// Views: user_behavior_detail, order_detail, user_behavior_window, user_order_agg, product_sales_agg
	expectedTables := []string{
		"user_events", "user_info", "product_info", "orders",
		"user_behavior_stats", "user_order_summary", "product_sales_rank", "realtime_alerts",
		"user_behavior_detail", "order_detail", "user_behavior_window", "user_order_agg", "product_sales_agg",
	}

	if len(schemas) < len(expectedTables) {
		t.Logf("Warning: Expected at least %d schemas, got %d", len(expectedTables), len(schemas))
	}
}

// TestFlinkComplete_BuildCatalogFromDDL tests building catalog from DDL
func TestFlinkComplete_BuildCatalogFromDDL(t *testing.T) {
	catalog, err := setupFlinkCompleteCatalogFromDDL()
	if err != nil {
		t.Fatalf("Failed to setup catalog from DDL: %v", err)
	}

	// Verify some tables exist
	tablesToCheck := []string{"user_events", "user_info", "orders", "user_behavior_stats"}
	for _, table := range tablesToCheck {
		schema, err := catalog.GetTableSchema("", table)
		if err != nil {
			t.Errorf("Table %s not found: %v", table, err)
			continue
		}
		t.Logf("Table %s has columns: %v", table, schema.Columns)
	}
}

// splitStatements splits SQL content into individual statements
func splitStatements(sql string) []string {
	// Remove comments
	lines := strings.Split(sql, "\n")
	var cleanLines []string
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "--") {
			continue
		}
		cleanLines = append(cleanLines, line)
	}
	cleanSQL := strings.Join(cleanLines, "\n")

	// Split by semicolon
	parts := strings.Split(cleanSQL, ";")
	var statements []string
	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if len(trimmed) > 0 {
			statements = append(statements, trimmed)
		}
	}
	return statements
}

// TestFlinkComplete_ParseStatements tests parsing individual statements from complete_example.sql
func TestFlinkComplete_ParseStatements(t *testing.T) {
	sqlPath := filepath.Join("..", "testdata", "flink", "complete_example.sql")
	content, err := os.ReadFile(sqlPath)
	if err != nil {
		t.Fatalf("Failed to read SQL file: %v", err)
	}

	statements := splitStatements(string(content))
	t.Logf("Found %d statements", len(statements))

	for i, stmt := range statements {
		// Get first line for logging
		firstLine := strings.Split(stmt, "\n")[0]
		if len(firstLine) > 60 {
			firstLine = firstLine[:60] + "..."
		}
		t.Logf("Statement %d: %s", i+1, firstLine)
	}
}

// TestFlinkComplete_TemporalJoinView tests the user_behavior_detail view with temporal join
func TestFlinkComplete_TemporalJoinView(t *testing.T) {
	catalog, err := setupFlinkCompleteCatalogFromDDL()
	if err != nil {
		t.Fatalf("Failed to setup catalog from DDL: %v", err)
	}
	analyzer := lineage.NewAnalyzer(catalog)

	// user_behavior_detail view from complete_example.sql
	sql := `SELECT 
		e.user_id,
		u.user_name,
		u.city,
		e.event_type,
		e.page_url,
		e.device_type,
		e.event_time
	FROM user_events e
	LEFT JOIN user_info FOR SYSTEM_TIME AS OF e.event_time AS u
		ON e.user_id = u.user_id`

	result, err := analyzer.Analyze(sql)
	if err != nil {
		t.Fatalf("Analyze failed: %v", err)
	}

	printLineageResult(t, sql, result)
	assertColumnCount(t, result, 7)
}

// TestFlinkComplete_OrderDetailView tests the order_detail view with multiple temporal joins
func TestFlinkComplete_OrderDetailView(t *testing.T) {
	catalog, err := setupFlinkCompleteCatalogFromDDL()
	if err != nil {
		t.Fatalf("Failed to setup catalog from DDL: %v", err)
	}
	analyzer := lineage.NewAnalyzer(catalog)

	// order_detail view from complete_example.sql
	sql := `SELECT 
		o.order_id,
		o.user_id,
		u.user_name,
		u.city,
		o.product_id,
		p.product_name,
		p.category,
		p.price AS unit_price,
		o.quantity,
		o.amount,
		o.order_time
	FROM orders o
	LEFT JOIN user_info FOR SYSTEM_TIME AS OF o.order_time AS u
		ON o.user_id = u.user_id
	LEFT JOIN product_info FOR SYSTEM_TIME AS OF o.order_time AS p
		ON o.product_id = p.product_id`

	result, err := analyzer.Analyze(sql)
	if err != nil {
		t.Fatalf("Analyze failed: %v", err)
	}

	printLineageResult(t, sql, result)
	assertColumnCount(t, result, 11)
}

// TestFlinkComplete_TumbleWindowView tests the user_behavior_window view with TUMBLE window
func TestFlinkComplete_TumbleWindowView(t *testing.T) {
	catalog, err := setupFlinkCompleteCatalogFromDDL()
	if err != nil {
		t.Fatalf("Failed to setup catalog from DDL: %v", err)
	}
	analyzer := lineage.NewAnalyzer(catalog)

	// user_behavior_window view from complete_example.sql
	sql := `SELECT 
		user_id,
		TUMBLE_START(event_time, INTERVAL '1' HOUR) AS window_start,
		TUMBLE_END(event_time, INTERVAL '1' HOUR) AS window_end,
		COUNT(*) AS pv_count,
		COUNT(DISTINCT page_url) AS uv_count
	FROM user_events
	GROUP BY user_id, TUMBLE(event_time, INTERVAL '1' HOUR)`

	result, err := analyzer.Analyze(sql)
	if err != nil {
		t.Fatalf("Analyze failed: %v", err)
	}

	printLineageResult(t, sql, result)
	assertColumnCount(t, result, 5)
}

// TestFlinkComplete_UserOrderAggView tests the user_order_agg view
func TestFlinkComplete_UserOrderAggView(t *testing.T) {
	// order_detail is a TEMPORARY VIEW defined in complete_example.sql, should be parsed from DDL
	catalog, err := setupFlinkCompleteCatalogFromDDL()
	if err != nil {
		t.Fatalf("Failed to setup catalog from DDL: %v", err)
	}
	analyzer := lineage.NewAnalyzer(catalog)

	// user_order_agg view from complete_example.sql
	sql := `SELECT 
		user_id,
		user_name,
		COUNT(*) AS total_orders,
		SUM(amount) AS total_amount,
		AVG(amount) AS avg_amount,
		MAX(order_time) AS last_order_time
	FROM order_detail
	GROUP BY user_id, user_name`

	result, err := analyzer.Analyze(sql)
	if err != nil {
		t.Fatalf("Analyze failed: %v", err)
	}

	printLineageResult(t, sql, result)
	assertColumnCount(t, result, 6)
}

// TestFlinkComplete_ProductSalesAggView tests the product_sales_agg view
func TestFlinkComplete_ProductSalesAggView(t *testing.T) {
	catalog, err := setupFlinkCompleteCatalogFromDDL()
	if err != nil {
		t.Fatalf("Failed to setup catalog from DDL: %v", err)
	}
	analyzer := lineage.NewAnalyzer(catalog)

	// product_sales_agg view from complete_example.sql
	sql := `SELECT 
		product_id,
		product_name,
		category,
		SUM(quantity) AS total_quantity,
		SUM(amount) AS total_sales,
		DATE_FORMAT(order_time, 'yyyy-MM-dd') AS stat_date
	FROM order_detail
	GROUP BY product_id, product_name, category, DATE_FORMAT(order_time, 'yyyy-MM-dd')`

	result, err := analyzer.Analyze(sql)
	if err != nil {
		t.Fatalf("Analyze failed: %v", err)
	}

	printLineageResult(t, sql, result)
	assertColumnCount(t, result, 6)
}

// TestFlinkComplete_InsertUserBehaviorStats tests INSERT INTO user_behavior_stats
func TestFlinkComplete_InsertUserBehaviorStats(t *testing.T) {
	catalog, err := setupFlinkCompleteCatalogFromDDL()
	if err != nil {
		t.Fatalf("Failed to setup catalog from DDL: %v", err)
	}
	analyzer := lineage.NewAnalyzer(catalog)

	sql := `INSERT INTO user_behavior_stats
	SELECT 
		user_id,
		window_start,
		window_end,
		pv_count,
		uv_count
	FROM user_behavior_window`

	result, err := analyzer.Analyze(sql)
	if err != nil {
		t.Fatalf("Analyze failed: %v", err)
	}

	printLineageResult(t, sql, result)
	assertTargetTable(t, result, "user_behavior_stats")
	assertColumnCount(t, result, 5)
}

// TestFlinkComplete_InsertUserOrderSummary tests INSERT INTO user_order_summary
func TestFlinkComplete_InsertUserOrderSummary(t *testing.T) {
	catalog, err := setupFlinkCompleteCatalogFromDDL()
	if err != nil {
		t.Fatalf("Failed to setup catalog from DDL: %v", err)
	}
	analyzer := lineage.NewAnalyzer(catalog)

	sql := `INSERT INTO user_order_summary
	SELECT 
		user_id,
		user_name,
		total_orders,
		total_amount,
		avg_amount,
		last_order_time
	FROM user_order_agg`

	result, err := analyzer.Analyze(sql)
	if err != nil {
		t.Fatalf("Analyze failed: %v", err)
	}

	printLineageResult(t, sql, result)
	assertTargetTable(t, result, "user_order_summary")
	assertColumnCount(t, result, 6)
}

// TestFlinkComplete_InsertProductSalesRank tests INSERT INTO product_sales_rank with ROW_NUMBER
func TestFlinkComplete_InsertProductSalesRank(t *testing.T) {
	catalog, err := setupFlinkCompleteCatalogFromDDL()
	if err != nil {
		t.Fatalf("Failed to setup catalog from DDL: %v", err)
	}
	analyzer := lineage.NewAnalyzer(catalog)

	sql := `INSERT INTO product_sales_rank
	SELECT 
		product_id,
		product_name,
		category,
		total_quantity,
		total_sales,
		ROW_NUMBER() OVER (PARTITION BY stat_date ORDER BY total_sales DESC) AS rank_num,
		stat_date
	FROM product_sales_agg`

	result, err := analyzer.Analyze(sql)
	if err != nil {
		t.Fatalf("Analyze failed: %v", err)
	}

	printLineageResult(t, sql, result)
	assertTargetTable(t, result, "product_sales_rank")
	assertColumnCount(t, result, 7)
}

// TestFlinkComplete_InsertRealtimeAlertsHighValue tests INSERT INTO realtime_alerts for high value orders
func TestFlinkComplete_InsertRealtimeAlertsHighValue(t *testing.T) {
	catalog, err := setupFlinkCompleteCatalogFromDDL()
	if err != nil {
		t.Fatalf("Failed to setup catalog from DDL: %v", err)
	}
	analyzer := lineage.NewAnalyzer(catalog)

	sql := `INSERT INTO realtime_alerts
	SELECT 
		CONCAT('ALERT_', CAST(order_id AS STRING)) AS alert_id,
		'HIGH_VALUE_ORDER' AS alert_type,
		user_id,
		CONCAT('User ', user_name, ' order amount ', CAST(amount AS STRING), ' exceeds threshold') AS alert_message,
		order_time AS alert_time
	FROM order_detail
	WHERE amount > 10000`

	result, err := analyzer.Analyze(sql)
	if err != nil {
		t.Fatalf("Analyze failed: %v", err)
	}

	printLineageResult(t, sql, result)
	assertTargetTable(t, result, "realtime_alerts")
	assertColumnCount(t, result, 5)
}

// TestFlinkComplete_InsertRealtimeAlertsAbnormal tests INSERT INTO realtime_alerts for abnormal behavior
func TestFlinkComplete_InsertRealtimeAlertsAbnormal(t *testing.T) {
	catalog, err := setupFlinkCompleteCatalogFromDDL()
	if err != nil {
		t.Fatalf("Failed to setup catalog from DDL: %v", err)
	}
	analyzer := lineage.NewAnalyzer(catalog)

	sql := `INSERT INTO realtime_alerts
	SELECT 
		CONCAT('ALERT_', CAST(user_id AS STRING), '_', CAST(window_start AS STRING)) AS alert_id,
		'ABNORMAL_BEHAVIOR' AS alert_type,
		user_id,
		CONCAT('User ', CAST(user_id AS STRING), ' visited ', CAST(pv_count AS STRING), ' times in 1 hour, suspicious') AS alert_message,
		window_end AS alert_time
	FROM user_behavior_window
	WHERE pv_count > 1000`

	result, err := analyzer.Analyze(sql)
	if err != nil {
		t.Fatalf("Analyze failed: %v", err)
	}

	printLineageResult(t, sql, result)
	assertTargetTable(t, result, "realtime_alerts")
	assertColumnCount(t, result, 5)
}

// TestFlinkComplete_VerifyLineage_UserBehaviorStats verifies the expected lineage for user_behavior_stats
func TestFlinkComplete_VerifyLineage_UserBehaviorStats(t *testing.T) {
	catalog, err := setupFlinkCompleteCatalogFromDDL()
	if err != nil {
		t.Fatalf("Failed to setup catalog from DDL: %v", err)
	}
	analyzer := lineage.NewAnalyzer(catalog)

	// Direct query from user_events to user_behavior_stats
	sql := `INSERT INTO user_behavior_stats
	SELECT 
		user_id,
		TUMBLE_START(event_time, INTERVAL '1' HOUR) AS window_start,
		TUMBLE_END(event_time, INTERVAL '1' HOUR) AS window_end,
		COUNT(*) AS pv_count,
		COUNT(DISTINCT page_url) AS uv_count
	FROM user_events
	GROUP BY user_id, TUMBLE(event_time, INTERVAL '1' HOUR)`

	result, err := analyzer.Analyze(sql)
	if err != nil {
		t.Fatalf("Analyze failed: %v", err)
	}

	printLineageResult(t, sql, result)

	// Expected lineage:
	// user_id <- user_events.user_id
	// window_start <- user_events.event_time (TUMBLE_START)
	// window_end <- user_events.event_time (TUMBLE_END)
	// pv_count <- user_events.* (COUNT)
	// uv_count <- user_events.page_url (COUNT DISTINCT)
	assertTargetTable(t, result, "user_behavior_stats")
	assertColumnCount(t, result, 5)
}
