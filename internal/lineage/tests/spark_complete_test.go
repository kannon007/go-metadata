package tests

import (
	"go-metadata/internal/lineage"
	"go-metadata/internal/lineage/metadata"
	"os"
	"path/filepath"
	"testing"
)

// loadSparkCompleteSQL loads the Spark complete_example.sql file
func loadSparkCompleteSQL() (string, error) {
	sqlPath := filepath.Join("..", "testdata", "spark", "complete_example.sql")
	content, err := os.ReadFile(sqlPath)
	if err != nil {
		return "", err
	}
	return string(content), nil
}

// setupSparkCompleteCatalogFromDDL creates a catalog by parsing DDL from spark complete_example.sql
func setupSparkCompleteCatalogFromDDL() (lineage.Catalog, error) {
	sql, err := loadSparkCompleteSQL()
	if err != nil {
		return nil, err
	}

	// Use MetadataBuilder to parse DDL and build catalog
	catalog := metadata.NewMetadataBuilder().
		LoadFromDDL(sql).
		BuildCatalog()

	return catalog, nil
}

// TestSparkComplete_LoadSQL tests loading the spark complete_example.sql file
func TestSparkComplete_LoadSQL(t *testing.T) {
	sql, err := loadSparkCompleteSQL()
	if err != nil {
		t.Fatalf("Failed to read SQL file: %v", err)
	}

	if len(sql) == 0 {
		t.Fatal("SQL file is empty")
	}

	t.Logf("Loaded SQL file with %d bytes", len(sql))
}

// TestSparkComplete_ParseDDL tests parsing DDL from spark complete_example.sql
func TestSparkComplete_ParseDDL(t *testing.T) {
	sql, err := loadSparkCompleteSQL()
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

	// Expected tables from spark complete_example.sql:
	// ODS: user_behavior_log, order_fact
	// DIM: user_dim, product_dim
	// DWS: user_behavior_daily, user_order_daily, product_sales_daily
	// ADS: user_rfm_analysis, sales_ranking
	// Views: tmp_user_behavior_detail, tmp_order_detail, tmp_user_behavior_agg, etc.
	if len(schemas) == 0 {
		t.Log("Warning: No schemas parsed from DDL")
	}
}

// TestSparkComplete_BuildCatalogFromDDL tests building catalog from DDL
func TestSparkComplete_BuildCatalogFromDDL(t *testing.T) {
	catalog, err := setupSparkCompleteCatalogFromDDL()
	if err != nil {
		t.Fatalf("Failed to setup catalog from DDL: %v", err)
	}

	// Verify some tables exist
	tablesToCheck := []struct {
		db    string
		table string
	}{
		{"ods", "user_behavior_log"},
		{"ods", "order_fact"},
		{"dim", "user_dim"},
		{"dim", "product_dim"},
		{"dws", "user_behavior_daily"},
	}
	for _, tc := range tablesToCheck {
		schema, err := catalog.GetTableSchema(tc.db, tc.table)
		if err != nil {
			t.Errorf("Table %s.%s not found: %v", tc.db, tc.table, err)
			continue
		}
		t.Logf("Table %s.%s has columns: %v", tc.db, tc.table, schema.Columns)
	}
}

// TestSparkComplete_ParseStatements tests parsing individual statements from spark complete_example.sql
func TestSparkComplete_ParseStatements(t *testing.T) {
	sqlPath := filepath.Join("..", "testdata", "spark", "complete_example.sql")
	content, err := os.ReadFile(sqlPath)
	if err != nil {
		t.Fatalf("Failed to read SQL file: %v", err)
	}

	statements := splitStatements(string(content))
	t.Logf("Found %d statements", len(statements))

	for i, stmt := range statements {
		firstLine := stmt
		if idx := len(firstLine); idx > 60 {
			firstLine = firstLine[:60] + "..."
		}
		t.Logf("Statement %d: %s", i+1, firstLine)
	}
}

// TestSparkComplete_UserBehaviorDetailView tests the tmp_user_behavior_detail view
func TestSparkComplete_UserBehaviorDetailView(t *testing.T) {
	catalog, err := setupSparkCompleteCatalogFromDDL()
	if err != nil {
		t.Fatalf("Failed to setup catalog from DDL: %v", err)
	}
	analyzer := lineage.NewAnalyzer(catalog)

	// tmp_user_behavior_detail view from spark complete_example.sql
	sql := `SELECT 
		b.user_id,
		u.user_name,
		u.city,
		u.province,
		u.user_level,
		b.session_id,
		b.event_type,
		b.page_url,
		b.device_type,
		b.event_time,
		b.dt
	FROM ods.user_behavior_log b
	LEFT JOIN dim.user_dim u ON b.user_id = u.user_id
	WHERE b.dt = '2024-01-01'`

	result, err := analyzer.Analyze(sql)
	if err != nil {
		t.Fatalf("Analyze failed: %v", err)
	}

	printLineageResult(t, sql, result)
	assertColumnCount(t, result, 11)
}

// TestSparkComplete_OrderDetailView tests the tmp_order_detail view
func TestSparkComplete_OrderDetailView(t *testing.T) {
	catalog, err := setupSparkCompleteCatalogFromDDL()
	if err != nil {
		t.Fatalf("Failed to setup catalog from DDL: %v", err)
	}
	analyzer := lineage.NewAnalyzer(catalog)

	// tmp_order_detail view from spark complete_example.sql
	sql := `SELECT 
		o.order_id,
		o.user_id,
		u.user_name,
		u.city,
		u.user_level,
		u.is_vip,
		o.product_id,
		p.product_name,
		p.category_name,
		p.brand_name,
		o.quantity,
		o.unit_price,
		o.total_amount,
		o.discount_amount,
		o.pay_amount,
		o.order_status,
		o.pay_type,
		o.order_time,
		o.pay_time,
		o.dt
	FROM ods.order_fact o
	LEFT JOIN dim.user_dim u ON o.user_id = u.user_id
	LEFT JOIN dim.product_dim p ON o.product_id = p.product_id
	WHERE o.dt = '2024-01-01' AND o.order_status = 'PAID'`

	result, err := analyzer.Analyze(sql)
	if err != nil {
		t.Fatalf("Analyze failed: %v", err)
	}

	printLineageResult(t, sql, result)
	assertColumnCount(t, result, 20)
}
