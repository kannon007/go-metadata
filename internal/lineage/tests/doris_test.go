package tests

import (
	"go-metadata/internal/lineage"
	"testing"
)

func setupDorisCatalog() *MockCatalog {
	catalog := NewMockCatalog()
	catalog.AddTable("default", "page_stats", []string{"date", "channel", "pv", "user_bitmap"})
	catalog.AddTable("default", "user_behavior", []string{"user_id", "item_id", "behavior", "dt"})
	catalog.AddTable("default", "orders", []string{"order_id", "user_id", "amount", "dt"})
	return catalog
}

func TestDoris_BitmapAggregation(t *testing.T) {
	catalog := setupDorisCatalog()
	parser := lineage.NewParser(catalog)

	sql := `SELECT date, channel,
			SUM(pv) as total_pv,
			BITMAP_UNION_COUNT(user_bitmap) as uv
			FROM page_stats
			GROUP BY date, channel`
	result, err := parser.Parse(sql)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}
	if result == nil {
		t.Fatal("Parse returned nil result")
	}
}

func TestDoris_HllAggregation(t *testing.T) {
	catalog := setupDorisCatalog()
	parser := lineage.NewParser(catalog)

	sql := `SELECT dt,
			HLL_UNION_AGG(user_id) as approx_users
			FROM user_behavior
			GROUP BY dt`
	result, err := parser.Parse(sql)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}
	if result == nil {
		t.Fatal("Parse returned nil result")
	}
}

func TestDoris_RollupQuery(t *testing.T) {
	catalog := setupDorisCatalog()
	parser := lineage.NewParser(catalog)

	sql := `SELECT dt, user_id, SUM(amount) as total
			FROM orders
			GROUP BY ROLLUP(dt, user_id)`
	result, err := parser.Parse(sql)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}
	if result == nil {
		t.Fatal("Parse returned nil result")
	}
}

func TestDoris_WindowFunnel(t *testing.T) {
	catalog := setupDorisCatalog()
	parser := lineage.NewParser(catalog)

	sql := `SELECT user_id,
			WINDOW_FUNNEL(86400, dt, 
				behavior = 'view',
				behavior = 'cart',
				behavior = 'purchase'
			) as funnel_step
			FROM user_behavior
			GROUP BY user_id`
	result, err := parser.Parse(sql)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}
	if result == nil {
		t.Fatal("Parse returned nil result")
	}
}
