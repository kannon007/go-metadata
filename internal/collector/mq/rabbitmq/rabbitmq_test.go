package rabbitmq

import (
	"context"
	"testing"

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
				Endpoint: "http://localhost:15672",
			},
			wantErr: false,
		},
		{
			name: "wrong type",
			config: &config.ConnectorConfig{
				Type:     "mysql",
				Endpoint: "http://localhost:15672",
			},
			wantErr: true,
		},
		{
			name: "empty type (should be allowed)",
			config: &config.ConnectorConfig{
				Type:     "",
				Endpoint: "http://localhost:15672",
			},
			wantErr: false,
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
		Endpoint: "http://localhost:15672",
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
		Endpoint: "http://localhost:15672",
	}
	
	c, err := NewCollector(cfg)
	if err != nil {
		t.Fatalf("NewCollector() error = %v", err)
	}
	
	if got := c.Type(); got != SourceName {
		t.Errorf("Type() = %v, want %v", got, SourceName)
	}
}

func TestCollector_parseEndpoint(t *testing.T) {
	tests := []struct {
		name     string
		endpoint string
		want     string
		wantErr  bool
	}{
		{
			name:     "empty endpoint",
			endpoint: "",
			wantErr:  true,
		},
		{
			name:     "full URL with API path",
			endpoint: "http://localhost:15672/api",
			want:     "http://localhost:15672/api",
			wantErr:  false,
		},
		{
			name:     "URL without API path",
			endpoint: "http://localhost:15672",
			want:     "http://localhost:15672/api",
			wantErr:  false,
		},
		{
			name:     "URL without scheme",
			endpoint: "localhost:15672",
			want:     "http://localhost:15672/api",
			wantErr:  false,
		},
		{
			name:     "URL without port",
			endpoint: "http://localhost",
			want:     "http://localhost:15672/api",
			wantErr:  false,
		},
		{
			name:     "hostname only",
			endpoint: "localhost",
			want:     "http://localhost:15672/api",
			wantErr:  false,
		},
		{
			name:     "HTTPS URL",
			endpoint: "https://rabbitmq.example.com:15672",
			want:     "https://rabbitmq.example.com:15672/api",
			wantErr:  false,
		},
		{
			name:     "URL with custom path",
			endpoint: "http://localhost:15672/management",
			want:     "http://localhost:15672/management/api",
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &config.ConnectorConfig{
				Type:     SourceName,
				Endpoint: tt.endpoint,
			}
			
			c := &Collector{config: cfg}
			got, err := c.parseEndpoint()
			
			if (err != nil) != tt.wantErr {
				t.Errorf("parseEndpoint() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			
			if !tt.wantErr && got != tt.want {
				t.Errorf("parseEndpoint() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCollector_HealthCheck_NotConnected(t *testing.T) {
	cfg := &config.ConnectorConfig{
		Type:     SourceName,
		Endpoint: "http://localhost:15672",
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
		Endpoint: "http://localhost:15672",
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
		Endpoint: "http://localhost:15672",
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
	
	if collErr, ok := err.(*collector.CollectorError); ok {
		if collErr.Code != collector.ErrCodeConnectionClosed {
			t.Errorf("DiscoverCatalogs() error code = %v, want %v", collErr.Code, collector.ErrCodeConnectionClosed)
		}
	} else {
		t.Errorf("DiscoverCatalogs() should return CollectorError, got %T", err)
	}
}

func TestCollector_ListSchemas_NotConnected(t *testing.T) {
	cfg := &config.ConnectorConfig{
		Type:     SourceName,
		Endpoint: "http://localhost:15672",
	}
	
	c, err := NewCollector(cfg)
	if err != nil {
		t.Fatalf("NewCollector() error = %v", err)
	}
	
	ctx := context.Background()
	_, err = c.ListSchemas(ctx, "rabbitmq")
	
	// Should return connection closed error
	if err == nil {
		t.Error("ListSchemas() should return error when not connected")
	}
	
	if collErr, ok := err.(*collector.CollectorError); ok {
		if collErr.Code != collector.ErrCodeConnectionClosed {
			t.Errorf("ListSchemas() error code = %v, want %v", collErr.Code, collector.ErrCodeConnectionClosed)
		}
	} else {
		t.Errorf("ListSchemas() should return CollectorError, got %T", err)
	}
}

func TestCollector_ListTables_NotConnected(t *testing.T) {
	cfg := &config.ConnectorConfig{
		Type:     SourceName,
		Endpoint: "http://localhost:15672",
	}
	
	c, err := NewCollector(cfg)
	if err != nil {
		t.Fatalf("NewCollector() error = %v", err)
	}
	
	ctx := context.Background()
	_, err = c.ListTables(ctx, "rabbitmq", "/", nil)
	
	// Should return connection closed error
	if err == nil {
		t.Error("ListTables() should return error when not connected")
	}
	
	if collErr, ok := err.(*collector.CollectorError); ok {
		if collErr.Code != collector.ErrCodeConnectionClosed {
			t.Errorf("ListTables() error code = %v, want %v", collErr.Code, collector.ErrCodeConnectionClosed)
		}
	} else {
		t.Errorf("ListTables() should return CollectorError, got %T", err)
	}
}

func TestCollector_FetchTableMetadata_NotConnected(t *testing.T) {
	cfg := &config.ConnectorConfig{
		Type:     SourceName,
		Endpoint: "http://localhost:15672",
	}
	
	c, err := NewCollector(cfg)
	if err != nil {
		t.Fatalf("NewCollector() error = %v", err)
	}
	
	ctx := context.Background()
	_, err = c.FetchTableMetadata(ctx, "rabbitmq", "/", "test-queue")
	
	// Should return connection closed error
	if err == nil {
		t.Error("FetchTableMetadata() should return error when not connected")
	}
	
	if collErr, ok := err.(*collector.CollectorError); ok {
		if collErr.Code != collector.ErrCodeConnectionClosed {
			t.Errorf("FetchTableMetadata() error code = %v, want %v", collErr.Code, collector.ErrCodeConnectionClosed)
		}
	} else {
		t.Errorf("FetchTableMetadata() should return CollectorError, got %T", err)
	}
}

func TestCollector_FetchTableStatistics_NotConnected(t *testing.T) {
	cfg := &config.ConnectorConfig{
		Type:     SourceName,
		Endpoint: "http://localhost:15672",
	}
	
	c, err := NewCollector(cfg)
	if err != nil {
		t.Fatalf("NewCollector() error = %v", err)
	}
	
	ctx := context.Background()
	_, err = c.FetchTableStatistics(ctx, "rabbitmq", "/", "test-queue")
	
	// Should return connection closed error
	if err == nil {
		t.Error("FetchTableStatistics() should return error when not connected")
	}
	
	if collErr, ok := err.(*collector.CollectorError); ok {
		if collErr.Code != collector.ErrCodeConnectionClosed {
			t.Errorf("FetchTableStatistics() error code = %v, want %v", collErr.Code, collector.ErrCodeConnectionClosed)
		}
	} else {
		t.Errorf("FetchTableStatistics() should return CollectorError, got %T", err)
	}
}

func TestCollector_FetchPartitions(t *testing.T) {
	cfg := &config.ConnectorConfig{
		Type:     SourceName,
		Endpoint: "http://localhost:15672",
	}
	
	c, err := NewCollector(cfg)
	if err != nil {
		t.Fatalf("NewCollector() error = %v", err)
	}
	
	ctx := context.Background()
	partitions, err := c.FetchPartitions(ctx, "rabbitmq", "/", "test-queue")
	
	// Should not error and return empty slice (RabbitMQ doesn't have partitions)
	if err != nil {
		t.Errorf("FetchPartitions() error = %v", err)
	}
	
	if len(partitions) != 0 {
		t.Errorf("FetchPartitions() should return empty slice, got %d partitions", len(partitions))
	}
}

func TestCollector_wrapConnectionError(t *testing.T) {
	cfg := &config.ConnectorConfig{
		Type:     SourceName,
		Endpoint: "http://localhost:15672",
	}
	
	c := &Collector{config: cfg}
	
	tests := []struct {
		name     string
		inputErr string
		wantCode collector.ErrorCode
	}{
		{
			name:     "authentication error",
			inputErr: "authentication failed",
			wantCode: collector.ErrCodeAuthError,
		},
		{
			name:     "401 error",
			inputErr: "status 401",
			wantCode: collector.ErrCodeAuthError,
		},
		{
			name:     "connection refused",
			inputErr: "connection refused",
			wantCode: collector.ErrCodeNetworkError,
		},
		{
			name:     "no such host",
			inputErr: "no such host",
			wantCode: collector.ErrCodeNetworkError,
		},
		{
			name:     "timeout error",
			inputErr: "timeout",
			wantCode: collector.ErrCodeTimeout,
		},
		{
			name:     "deadline exceeded",
			inputErr: "deadline exceeded",
			wantCode: collector.ErrCodeDeadlineExceeded,
		},
		{
			name:     "generic error",
			inputErr: "some other error",
			wantCode: collector.ErrCodeNetworkError,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			inputErr := &testError{msg: tt.inputErr}
			wrappedErr := c.wrapConnectionError(inputErr)
			
			if collErr, ok := wrappedErr.(*collector.CollectorError); ok {
				if collErr.Code != tt.wantCode {
					t.Errorf("wrapConnectionError() error code = %v, want %v", collErr.Code, tt.wantCode)
				}
			} else {
				t.Errorf("wrapConnectionError() should return CollectorError, got %T", wrappedErr)
			}
		})
	}
}

func TestCollector_filterTables(t *testing.T) {
	cfg := &config.ConnectorConfig{
		Type:     SourceName,
		Endpoint: "http://localhost:15672",
		Matching: &config.MatchingConfig{
			Tables: &config.MatchingRule{
				Include: []string{"test-*"},
				Exclude: []string{"test-exclude"},
			},
			PatternType:   "glob",
			CaseSensitive: false,
		},
	}
	
	c := &Collector{config: cfg}
	
	input := []string{"test-queue1", "test-queue2", "test-exclude", "other-queue"}
	expected := []string{"test-queue1", "test-queue2"}
	
	result := c.filterTables(input, nil)
	
	if len(result) != len(expected) {
		t.Errorf("filterTables() returned %d tables, want %d", len(result), len(expected))
		return
	}
	
	for i, table := range result {
		if table != expected[i] {
			t.Errorf("filterTables() table[%d] = %v, want %v", i, table, expected[i])
		}
	}
}

// Helper type for testing error wrapping
type testError struct {
	msg string
}

func (e *testError) Error() string {
	return e.msg
}

// Test RabbitMQ-specific extended methods

func TestCollector_ListExchanges_NotConnected(t *testing.T) {
	cfg := &config.ConnectorConfig{
		Type:     SourceName,
		Endpoint: "http://localhost:15672",
	}
	
	c, err := NewCollector(cfg)
	if err != nil {
		t.Fatalf("NewCollector() error = %v", err)
	}
	
	// Cast to concrete type to access extended methods
	rabbitCollector := c.(*Collector)
	
	ctx := context.Background()
	_, err = rabbitCollector.ListExchanges(ctx, "/")
	
	// Should return connection closed error
	if err == nil {
		t.Error("ListExchanges() should return error when not connected")
	}
	
	if collErr, ok := err.(*collector.CollectorError); ok {
		if collErr.Code != collector.ErrCodeConnectionClosed {
			t.Errorf("ListExchanges() error code = %v, want %v", collErr.Code, collector.ErrCodeConnectionClosed)
		}
	} else {
		t.Errorf("ListExchanges() should return CollectorError, got %T", err)
	}
}

func TestCollector_ListBindings_NotConnected(t *testing.T) {
	cfg := &config.ConnectorConfig{
		Type:     SourceName,
		Endpoint: "http://localhost:15672",
	}
	
	c, err := NewCollector(cfg)
	if err != nil {
		t.Fatalf("NewCollector() error = %v", err)
	}
	
	// Cast to concrete type to access extended methods
	rabbitCollector := c.(*Collector)
	
	ctx := context.Background()
	_, err = rabbitCollector.ListBindings(ctx, "/")
	
	// Should return connection closed error
	if err == nil {
		t.Error("ListBindings() should return error when not connected")
	}
	
	if collErr, ok := err.(*collector.CollectorError); ok {
		if collErr.Code != collector.ErrCodeConnectionClosed {
			t.Errorf("ListBindings() error code = %v, want %v", collErr.Code, collector.ErrCodeConnectionClosed)
		}
	} else {
		t.Errorf("ListBindings() should return CollectorError, got %T", err)
	}
}

func TestCollector_ListConsumers_NotConnected(t *testing.T) {
	cfg := &config.ConnectorConfig{
		Type:     SourceName,
		Endpoint: "http://localhost:15672",
	}
	
	c, err := NewCollector(cfg)
	if err != nil {
		t.Fatalf("NewCollector() error = %v", err)
	}
	
	// Cast to concrete type to access extended methods
	rabbitCollector := c.(*Collector)
	
	ctx := context.Background()
	_, err = rabbitCollector.ListConsumers(ctx, "/")
	
	// Should return connection closed error
	if err == nil {
		t.Error("ListConsumers() should return error when not connected")
	}
	
	if collErr, ok := err.(*collector.CollectorError); ok {
		if collErr.Code != collector.ErrCodeConnectionClosed {
			t.Errorf("ListConsumers() error code = %v, want %v", collErr.Code, collector.ErrCodeConnectionClosed)
		}
	} else {
		t.Errorf("ListConsumers() should return CollectorError, got %T", err)
	}
}

func TestCollector_GetQueueBindings_NotConnected(t *testing.T) {
	cfg := &config.ConnectorConfig{
		Type:     SourceName,
		Endpoint: "http://localhost:15672",
	}
	
	c, err := NewCollector(cfg)
	if err != nil {
		t.Fatalf("NewCollector() error = %v", err)
	}
	
	// Cast to concrete type to access extended methods
	rabbitCollector := c.(*Collector)
	
	ctx := context.Background()
	_, err = rabbitCollector.GetQueueBindings(ctx, "/", "test-queue")
	
	// Should return connection closed error
	if err == nil {
		t.Error("GetQueueBindings() should return error when not connected")
	}
	
	if collErr, ok := err.(*collector.CollectorError); ok {
		if collErr.Code != collector.ErrCodeConnectionClosed {
			t.Errorf("GetQueueBindings() error code = %v, want %v", collErr.Code, collector.ErrCodeConnectionClosed)
		}
	} else {
		t.Errorf("GetQueueBindings() should return CollectorError, got %T", err)
	}
}

func TestCollector_GetExchangeBindings_NotConnected(t *testing.T) {
	cfg := &config.ConnectorConfig{
		Type:     SourceName,
		Endpoint: "http://localhost:15672",
	}
	
	c, err := NewCollector(cfg)
	if err != nil {
		t.Fatalf("NewCollector() error = %v", err)
	}
	
	// Cast to concrete type to access extended methods
	rabbitCollector := c.(*Collector)
	
	ctx := context.Background()
	_, err = rabbitCollector.GetExchangeBindings(ctx, "/", "test-exchange")
	
	// Should return connection closed error
	if err == nil {
		t.Error("GetExchangeBindings() should return error when not connected")
	}
	
	if collErr, ok := err.(*collector.CollectorError); ok {
		if collErr.Code != collector.ErrCodeConnectionClosed {
			t.Errorf("GetExchangeBindings() error code = %v, want %v", collErr.Code, collector.ErrCodeConnectionClosed)
		}
	} else {
		t.Errorf("GetExchangeBindings() should return CollectorError, got %T", err)
	}
}