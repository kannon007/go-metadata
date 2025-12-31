package kafka

import (
	"context"
	"testing"
	"time"

	"go-metadata/internal/collector"
	"go-metadata/internal/collector/config"
)

func TestNewCollector(t *testing.T) {
	tests := []struct {
		name    string
		config  *config.ConnectorConfig
		wantErr bool
	}{
		{
			name:    "nil config",
			config:  nil,
			wantErr: true,
		},
		{
			name: "valid config",
			config: &config.ConnectorConfig{
				Type:     SourceName,
				Endpoint: "localhost:9092",
			},
			wantErr: false,
		},
		{
			name: "wrong type",
			config: &config.ConnectorConfig{
				Type:     "mysql",
				Endpoint: "localhost:9092",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			collector, err := NewCollector(tt.config)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewCollector() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && collector == nil {
				t.Error("NewCollector() returned nil collector")
			}
		})
	}
}

func TestCollector_Category(t *testing.T) {
	cfg := &config.ConnectorConfig{
		Type:     SourceName,
		Endpoint: "localhost:9092",
	}
	
	c, err := NewCollector(cfg)
	if err != nil {
		t.Fatalf("NewCollector() error = %v", err)
	}
	
	if got := c.Category(); got != collector.CategoryMessageQueue {
		t.Errorf("Category() = %v, want %v", got, collector.CategoryMessageQueue)
	}
}

func TestCollector_Type(t *testing.T) {
	cfg := &config.ConnectorConfig{
		Type:     SourceName,
		Endpoint: "localhost:9092",
	}
	
	c, err := NewCollector(cfg)
	if err != nil {
		t.Fatalf("NewCollector() error = %v", err)
	}
	
	if got := c.Type(); got != SourceName {
		t.Errorf("Type() = %v, want %v", got, SourceName)
	}
}

func TestCollector_parseBrokers(t *testing.T) {
	tests := []struct {
		name     string
		endpoint string
		want     []string
		wantErr  bool
	}{
		{
			name:     "empty endpoint",
			endpoint: "",
			wantErr:  true,
		},
		{
			name:     "single broker with port",
			endpoint: "localhost:9092",
			want:     []string{"localhost:9092"},
			wantErr:  false,
		},
		{
			name:     "single broker without port",
			endpoint: "localhost",
			want:     []string{"localhost:9092"},
			wantErr:  false,
		},
		{
			name:     "multiple brokers",
			endpoint: "broker1:9092,broker2:9093,broker3",
			want:     []string{"broker1:9092", "broker2:9093", "broker3:9092"},
			wantErr:  false,
		},
		{
			name:     "invalid port",
			endpoint: "localhost:invalid",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &config.ConnectorConfig{
				Type:     SourceName,
				Endpoint: tt.endpoint,
			}
			
			c := &Collector{config: cfg}
			got, err := c.parseBrokers()
			
			if (err != nil) != tt.wantErr {
				t.Errorf("parseBrokers() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			
			if !tt.wantErr {
				if len(got) != len(tt.want) {
					t.Errorf("parseBrokers() got %d brokers, want %d", len(got), len(tt.want))
					return
				}
				
				for i, broker := range got {
					if broker != tt.want[i] {
						t.Errorf("parseBrokers() broker[%d] = %v, want %v", i, broker, tt.want[i])
					}
				}
			}
		})
	}
}

func TestCollector_HealthCheck_NotConnected(t *testing.T) {
	cfg := &config.ConnectorConfig{
		Type:     SourceName,
		Endpoint: "localhost:9092",
	}
	
	c, err := NewCollector(cfg)
	if err != nil {
		t.Fatalf("NewCollector() error = %v", err)
	}
	
	ctx := context.Background()
	status, err := c.HealthCheck(ctx)
	if err != nil {
		t.Errorf("HealthCheck() error = %v", err)
	}
	
	if status.Connected {
		t.Error("HealthCheck() should return not connected when client is nil")
	}
	
	if status.Message != "not connected" {
		t.Errorf("HealthCheck() message = %v, want 'not connected'", status.Message)
	}
}

func TestCollector_Close(t *testing.T) {
	cfg := &config.ConnectorConfig{
		Type:     SourceName,
		Endpoint: "localhost:9092",
	}
	
	c, err := NewCollector(cfg)
	if err != nil {
		t.Fatalf("NewCollector() error = %v", err)
	}
	
	// Close should not error even when not connected
	if err := c.Close(); err != nil {
		t.Errorf("Close() error = %v", err)
	}
}

func TestCollector_DiscoverCatalogs_NotConnected(t *testing.T) {
	cfg := &config.ConnectorConfig{
		Type:     SourceName,
		Endpoint: "localhost:9092",
	}
	
	c, err := NewCollector(cfg)
	if err != nil {
		t.Fatalf("NewCollector() error = %v", err)
	}
	
	ctx := context.Background()
	_, err = c.DiscoverCatalogs(ctx)
	
	// Should return connection closed error
	if err == nil {
		t.Error("DiscoverCatalogs() should return error when not connected")
	}
	
	if collector.GetErrorCode(err) != collector.ErrCodeConnectionClosed {
		t.Errorf("DiscoverCatalogs() error code = %v, want %v", collector.GetErrorCode(err), collector.ErrCodeConnectionClosed)
	}
}

func TestCollector_ListSchemas_NotConnected(t *testing.T) {
	cfg := &config.ConnectorConfig{
		Type:     SourceName,
		Endpoint: "localhost:9092",
	}
	
	c, err := NewCollector(cfg)
	if err != nil {
		t.Fatalf("NewCollector() error = %v", err)
	}
	
	ctx := context.Background()
	_, err = c.ListSchemas(ctx, "kafka")
	
	// Should return connection closed error
	if err == nil {
		t.Error("ListSchemas() should return error when not connected")
	}
	
	if collector.GetErrorCode(err) != collector.ErrCodeConnectionClosed {
		t.Errorf("ListSchemas() error code = %v, want %v", collector.GetErrorCode(err), collector.ErrCodeConnectionClosed)
	}
}

func TestCollector_ListTables_NotConnected(t *testing.T) {
	cfg := &config.ConnectorConfig{
		Type:     SourceName,
		Endpoint: "localhost:9092",
	}
	
	c, err := NewCollector(cfg)
	if err != nil {
		t.Fatalf("NewCollector() error = %v", err)
	}
	
	ctx := context.Background()
	_, err = c.ListTables(ctx, "kafka", "default", nil)
	
	// Should return connection closed error
	if err == nil {
		t.Error("ListTables() should return error when not connected")
	}
	
	if collector.GetErrorCode(err) != collector.ErrCodeConnectionClosed {
		t.Errorf("ListTables() error code = %v, want %v", collector.GetErrorCode(err), collector.ErrCodeConnectionClosed)
	}
}

func TestCollector_FetchTableMetadata_NotConnected(t *testing.T) {
	cfg := &config.ConnectorConfig{
		Type:     SourceName,
		Endpoint: "localhost:9092",
	}
	
	c, err := NewCollector(cfg)
	if err != nil {
		t.Fatalf("NewCollector() error = %v", err)
	}
	
	ctx := context.Background()
	_, err = c.FetchTableMetadata(ctx, "kafka", "default", "test-topic")
	
	// Should return connection closed error
	if err == nil {
		t.Error("FetchTableMetadata() should return error when not connected")
	}
	
	if collector.GetErrorCode(err) != collector.ErrCodeConnectionClosed {
		t.Errorf("FetchTableMetadata() error code = %v, want %v", collector.GetErrorCode(err), collector.ErrCodeConnectionClosed)
	}
}

func TestCollector_FetchTableStatistics_NotConnected(t *testing.T) {
	cfg := &config.ConnectorConfig{
		Type:     SourceName,
		Endpoint: "localhost:9092",
	}
	
	c, err := NewCollector(cfg)
	if err != nil {
		t.Fatalf("NewCollector() error = %v", err)
	}
	
	ctx := context.Background()
	_, err = c.FetchTableStatistics(ctx, "kafka", "default", "test-topic")
	
	// Should return connection closed error
	if err == nil {
		t.Error("FetchTableStatistics() should return error when not connected")
	}
	
	if collector.GetErrorCode(err) != collector.ErrCodeConnectionClosed {
		t.Errorf("FetchTableStatistics() error code = %v, want %v", collector.GetErrorCode(err), collector.ErrCodeConnectionClosed)
	}
}

func TestCollector_FetchPartitions_NotConnected(t *testing.T) {
	cfg := &config.ConnectorConfig{
		Type:     SourceName,
		Endpoint: "localhost:9092",
	}
	
	c, err := NewCollector(cfg)
	if err != nil {
		t.Fatalf("NewCollector() error = %v", err)
	}
	
	ctx := context.Background()
	_, err = c.FetchPartitions(ctx, "kafka", "default", "test-topic")
	
	// Should return connection closed error
	if err == nil {
		t.Error("FetchPartitions() should return error when not connected")
	}
	
	if collector.GetErrorCode(err) != collector.ErrCodeConnectionClosed {
		t.Errorf("FetchPartitions() error code = %v, want %v", collector.GetErrorCode(err), collector.ErrCodeConnectionClosed)
	}
}

func TestCollector_ContextCancellation(t *testing.T) {
	cfg := &config.ConnectorConfig{
		Type:     SourceName,
		Endpoint: "localhost:9092",
	}
	
	c, err := NewCollector(cfg)
	if err != nil {
		t.Fatalf("NewCollector() error = %v", err)
	}
	
	// Test with cancelled context
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately
	
	_, err = c.DiscoverCatalogs(ctx)
	if err == nil {
		t.Error("DiscoverCatalogs() should return error with cancelled context")
	}
	
	if !collector.IsCancelled(err) {
		t.Errorf("DiscoverCatalogs() should return cancelled error, got %v", err)
	}
}

func TestCollector_ContextTimeout(t *testing.T) {
	cfg := &config.ConnectorConfig{
		Type:     SourceName,
		Endpoint: "localhost:9092",
	}
	
	c, err := NewCollector(cfg)
	if err != nil {
		t.Fatalf("NewCollector() error = %v", err)
	}
	
	// Test with timeout context
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
	defer cancel()
	
	// Wait for timeout
	time.Sleep(1 * time.Millisecond)
	
	_, err = c.DiscoverCatalogs(ctx)
	if err == nil {
		t.Error("DiscoverCatalogs() should return error with timeout context")
	}
	
	if !collector.IsDeadlineExceeded(err) {
		t.Errorf("DiscoverCatalogs() should return deadline exceeded error, got %v", err)
	}
}

func TestCollector_parseSchemaToColumns(t *testing.T) {
	cfg := &config.ConnectorConfig{
		Type:     SourceName,
		Endpoint: "localhost:9092",
	}
	
	c, err := NewCollector(cfg)
	if err != nil {
		t.Fatalf("NewCollector() error = %v", err)
	}
	
	kafkaCollector := c.(*Collector)
	
	tests := []struct {
		name       string
		schema     *Schema
		wantCols   int
		wantKeyCol bool
	}{
		{
			name: "AVRO schema",
			schema: &Schema{
				SchemaType: "AVRO",
				Schema:     `{"type":"record","name":"User","fields":[{"name":"id","type":"int"}]}`,
			},
			wantCols:   5, // key, value, timestamp, partition, offset
			wantKeyCol: true,
		},
		{
			name: "PROTOBUF schema",
			schema: &Schema{
				SchemaType: "PROTOBUF",
				Schema:     "syntax = \"proto3\"; message User { int32 id = 1; }",
			},
			wantCols:   5,
			wantKeyCol: true,
		},
		{
			name: "JSON schema",
			schema: &Schema{
				SchemaType: "JSON",
				Schema:     `{"type":"object","properties":{"id":{"type":"integer"}}}`,
			},
			wantCols:   5,
			wantKeyCol: true,
		},
		{
			name: "unknown schema type",
			schema: &Schema{
				SchemaType: "UNKNOWN",
				Schema:     "some schema",
			},
			wantCols:   5,
			wantKeyCol: true,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			columns, err := kafkaCollector.parseSchemaToColumns(tt.schema)
			if err != nil {
				t.Errorf("parseSchemaToColumns() error = %v", err)
				return
			}
			
			if len(columns) != tt.wantCols {
				t.Errorf("parseSchemaToColumns() got %d columns, want %d", len(columns), tt.wantCols)
			}
			
			if tt.wantKeyCol {
				found := false
				for _, col := range columns {
					if col.Name == "key" {
						found = true
						break
					}
				}
				if !found {
					t.Error("parseSchemaToColumns() should include key column")
				}
			}
		})
	}
}

func TestCollector_filterTables(t *testing.T) {
	cfg := &config.ConnectorConfig{
		Type:     SourceName,
		Endpoint: "localhost:9092",
		Matching: &config.MatchingConfig{
			PatternType:   "glob",
			CaseSensitive: false,
			Tables: &config.MatchingRule{
				Include: []string{"test-*"},
				Exclude: []string{"*-internal"},
			},
		},
	}
	
	c, err := NewCollector(cfg)
	if err != nil {
		t.Fatalf("NewCollector() error = %v", err)
	}
	
	kafkaCollector := c.(*Collector)
	
	tables := []string{
		"test-topic1",
		"test-topic2",
		"test-internal",
		"prod-topic",
		"other-topic",
	}
	
	filtered := kafkaCollector.filterTables(tables, nil)
	
	expected := []string{"test-topic1", "test-topic2"}
	if len(filtered) != len(expected) {
		t.Errorf("filterTables() got %d tables, want %d", len(filtered), len(expected))
	}
	
	for i, table := range filtered {
		if table != expected[i] {
			t.Errorf("filterTables() table[%d] = %v, want %v", i, table, expected[i])
		}
	}
}