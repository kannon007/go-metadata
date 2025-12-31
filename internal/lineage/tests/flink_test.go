package tests

import (
	"go-metadata/internal/lineage"
	"testing"
)

func setupFlinkCatalog() *MockCatalog {
	catalog := NewMockCatalog()
	catalog.AddTable("default", "events", []string{"user_id", "event_type", "event_time", "amount"})
	catalog.AddTable("default", "orders", []string{"order_id", "user_id", "amount", "order_time"})
	catalog.AddTable("default", "users", []string{"id", "name", "region"})
	return catalog
}

func TestFlink_TumbleWindow(t *testing.T) {
	catalog := setupFlinkCatalog()
	parser := lineage.NewParser(catalog)

	sql := `SELECT user_id,
			TUMBLE_START(event_time, INTERVAL '1' HOUR) as window_start,
			COUNT(*) as event_count
			FROM events
			GROUP BY user_id, TUMBLE(event_time, INTERVAL '1' HOUR)`
	result, err := parser.Parse(sql)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}
	if result == nil {
		t.Fatal("Parse returned nil result")
	}
}

func TestFlink_HopWindow(t *testing.T) {
	catalog := setupFlinkCatalog()
	parser := lineage.NewParser(catalog)

	sql := `SELECT user_id,
			HOP_START(event_time, INTERVAL '5' MINUTE, INTERVAL '1' HOUR) as window_start,
			SUM(amount) as total_amount
			FROM events
			GROUP BY user_id, HOP(event_time, INTERVAL '5' MINUTE, INTERVAL '1' HOUR)`
	result, err := parser.Parse(sql)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}
	if result == nil {
		t.Fatal("Parse returned nil result")
	}
}

func TestFlink_SessionWindow(t *testing.T) {
	catalog := setupFlinkCatalog()
	parser := lineage.NewParser(catalog)

	sql := `SELECT user_id,
			SESSION_START(event_time, INTERVAL '30' MINUTE) as session_start,
			COUNT(*) as event_count
			FROM events
			GROUP BY user_id, SESSION(event_time, INTERVAL '30' MINUTE)`
	result, err := parser.Parse(sql)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}
	if result == nil {
		t.Fatal("Parse returned nil result")
	}
}

func TestFlink_TemporalJoin(t *testing.T) {
	catalog := setupFlinkCatalog()
	parser := lineage.NewParser(catalog)

	sql := `SELECT o.order_id, o.amount, u.name, u.region
			FROM orders o
			JOIN users FOR SYSTEM_TIME AS OF o.order_time u
			ON o.user_id = u.id`
	result, err := parser.Parse(sql)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}
	if result == nil {
		t.Fatal("Parse returned nil result")
	}
}

func TestFlink_IntervalJoin(t *testing.T) {
	catalog := setupFlinkCatalog()
	parser := lineage.NewParser(catalog)

	sql := `SELECT o.order_id, e.event_type
			FROM orders o, events e
			WHERE o.user_id = e.user_id
			AND e.event_time BETWEEN o.order_time - INTERVAL '1' HOUR AND o.order_time`
	result, err := parser.Parse(sql)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}
	if result == nil {
		t.Fatal("Parse returned nil result")
	}
}
