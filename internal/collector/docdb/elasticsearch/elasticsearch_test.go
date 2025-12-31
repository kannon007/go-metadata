// Package elasticsearch provides an Elasticsearch metadata collector implementation.
package elasticsearch

import (
	"context"
	"errors"
	"testing"

	"go-metadata/internal/collector"
	"go-metadata/internal/collector/config"
	"go-metadata/internal/collector/infer"
)

// TestNewCollector tests the NewCollector function
func TestNewCollector(t *testing.T) {
	tests := []struct {
		name    string
		cfg     *config.ConnectorConfig
		wantErr bool
		errCode collector.ErrorCode
	}{
		{
			name:    "nil config",
			cfg:     nil,
			wantErr: true,
			errCode: collector.ErrCodeInvalidConfig,
		},
		{
			name: "valid config",
			cfg: &config.ConnectorConfig{
				Type:     "elasticsearch",
				Endpoint: "localhost:9200",
				Credentials: config.Credentials{
					User:     "elastic",
					Password: "password",
				},
			},
			wantErr: false,
		},
		{
			name: "empty type (allowed)",
			cfg: &config.ConnectorConfig{
				Type:     "",
				Endpoint: "localhost:9200",
			},
			wantErr: false,
		},
		{
			name: "wrong type",
			cfg: &config.ConnectorConfig{
				Type:     "mysql",
				Endpoint: "localhost:3306",
			},
			wantErr: true,
			errCode: collector.ErrCodeInvalidConfig,
		},
		{
			name: "with inference config",
			cfg: &config.ConnectorConfig{
				Type:     "elasticsearch",
				Endpoint: "localhost:9200",
				Infer: &config.InferConfig{
					Enabled:    true,
					SampleSize: 50,
					MaxDepth:   5,
					TypeMerge:  config.TypeMergeMostCommon,
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, err := NewCollector(tt.cfg)
			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
					return
				}
				var collErr *collector.CollectorError
				if errors.As(err, &collErr) {
					if collErr.Code != tt.errCode {
						t.Errorf("expected error code %s, got %s", tt.errCode, collErr.Code)
					}
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
					return
				}
				if c == nil {
					t.Error("expected collector, got nil")
				}
				
				// Verify collector properties
				esCollector := c.(*Collector)
				if esCollector.Category() != collector.CategoryDocumentDB {
					t.Errorf("expected category %s, got %s", collector.CategoryDocumentDB, esCollector.Category())
				}
				if esCollector.Type() != SourceName {
					t.Errorf("expected type %s, got %s", SourceName, esCollector.Type())
				}
			}
		})
	}
}

// TestBuildAddresses tests the address building logic
func TestBuildAddresses(t *testing.T) {
	tests := []struct {
		name        string
		cfg         *config.ConnectorConfig
		wantErr     bool
		wantContain string
	}{
		{
			name: "basic endpoint",
			cfg: &config.ConnectorConfig{
				Endpoint: "localhost:9200",
			},
			wantErr:     false,
			wantContain: "http://localhost:9200",
		},
		{
			name: "endpoint without port",
			cfg: &config.ConnectorConfig{
				Endpoint: "localhost",
			},
			wantErr:     false,
			wantContain: "http://localhost:9200",
		},
		{
			name: "https endpoint",
			cfg: &config.ConnectorConfig{
				Endpoint: "https://localhost:9200",
			},
			wantErr:     false,
			wantContain: "https://localhost:9200",
		},
		{
			name: "http endpoint",
			cfg: &config.ConnectorConfig{
				Endpoint: "http://localhost:9200",
			},
			wantErr:     false,
			wantContain: "http://localhost:9200",
		},
		{
			name: "endpoint with custom port",
			cfg: &config.ConnectorConfig{
				Endpoint: "localhost:9201",
			},
			wantErr:     false,
			wantContain: "http://localhost:9201",
		},
		{
			name: "empty endpoint",
			cfg: &config.ConnectorConfig{
				Endpoint: "",
			},
			wantErr: true,
		},
		{
			name: "invalid URL",
			cfg: &config.ConnectorConfig{
				Endpoint: "http://[::1:invalid",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Collector{config: tt.cfg}
			addresses, err := c.buildAddresses()
			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
					return
				}
				if len(addresses) == 0 {
					t.Error("expected at least one address")
					return
				}
				if tt.wantContain != "" && !contains(addresses[0], tt.wantContain) {
					t.Errorf("Address %q should contain %q", addresses[0], tt.wantContain)
				}
			}
		})
	}
}

// TestWrapConnectionError tests error wrapping
func TestWrapConnectionError(t *testing.T) {
	c := &Collector{config: &config.ConnectorConfig{}}

	tests := []struct {
		name     string
		err      error
		wantCode collector.ErrorCode
	}{
		{
			name:     "authentication failed",
			err:      errors.New("authentication failed"),
			wantCode: collector.ErrCodeAuthError,
		},
		{
			name:     "auth failed",
			err:      errors.New("auth failed"),
			wantCode: collector.ErrCodeAuthError,
		},
		{
			name:     "401 unauthorized",
			err:      errors.New("401 Unauthorized"),
			wantCode: collector.ErrCodeAuthError,
		},
		{
			name:     "connection refused",
			err:      errors.New("connection refused"),
			wantCode: collector.ErrCodeNetworkError,
		},
		{
			name:     "no such host",
			err:      errors.New("no such host"),
			wantCode: collector.ErrCodeNetworkError,
		},
		{
			name:     "timeout",
			err:      errors.New("i/o timeout"),
			wantCode: collector.ErrCodeTimeout,
		},
		{
			name:     "deadline exceeded",
			err:      errors.New("context deadline exceeded"),
			wantCode: collector.ErrCodeTimeout,
		},
		{
			name:     "generic error",
			err:      errors.New("some other error"),
			wantCode: collector.ErrCodeNetworkError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			wrapped := c.wrapConnectionError(tt.err)
			var collErr *collector.CollectorError
			if !errors.As(wrapped, &collErr) {
				t.Fatalf("expected CollectorError, got %T", wrapped)
			}
			if collErr.Code != tt.wantCode {
				t.Errorf("expected error code %s, got %s", tt.wantCode, collErr.Code)
			}
			if collErr.Source != SourceName {
				t.Errorf("expected source %s, got %s", SourceName, collErr.Source)
			}
			if collErr.Category != collector.CategoryDocumentDB {
				t.Errorf("expected category %s, got %s", collector.CategoryDocumentDB, collErr.Category)
			}
		})
	}
}

// TestCollectorNotConnected tests operations when not connected
func TestCollectorNotConnected(t *testing.T) {
	cfg := &config.ConnectorConfig{
		Type:     "elasticsearch",
		Endpoint: "localhost:9200",
	}
	c, _ := NewCollector(cfg)
	ctx := context.Background()

	// HealthCheck should return not connected status
	status, err := c.HealthCheck(ctx)
	if err != nil {
		t.Errorf("HealthCheck should not return error: %v", err)
	}
	if status.Connected {
		t.Error("HealthCheck should return not connected")
	}

	// Other operations should return connection closed error
	_, err = c.DiscoverCatalogs(ctx)
	assertConnectionClosedError(t, err, "DiscoverCatalogs")

	_, err = c.ListSchemas(ctx, "elasticsearch")
	assertConnectionClosedError(t, err, "ListSchemas")

	_, err = c.ListTables(ctx, "elasticsearch", "", nil)
	assertConnectionClosedError(t, err, "ListTables")

	_, err = c.FetchTableMetadata(ctx, "elasticsearch", "", "users")
	assertConnectionClosedError(t, err, "FetchTableMetadata")

	_, err = c.FetchTableStatistics(ctx, "elasticsearch", "", "users")
	assertConnectionClosedError(t, err, "FetchTableStatistics")

	_, err = c.FetchPartitions(ctx, "elasticsearch", "", "users")
	// Elasticsearch doesn't support partitions, so this should not return a connection error
	if err != nil {
		t.Errorf("FetchPartitions should not return error for Elasticsearch: %v", err)
	}
}

// TestCloseNotConnected tests Close when not connected
func TestCloseNotConnected(t *testing.T) {
	cfg := &config.ConnectorConfig{
		Type:     "elasticsearch",
		Endpoint: "localhost:9200",
	}
	c, _ := NewCollector(cfg)

	// Close should not return error when not connected
	err := c.Close()
	if err != nil {
		t.Errorf("Close should not return error when not connected: %v", err)
	}
}

// TestFilterTables tests the table filtering logic
func TestFilterTables(t *testing.T) {
	tests := []struct {
		name     string
		tables   []string
		cfg      *config.ConnectorConfig
		opts     *collector.ListOptions
		expected []string
	}{
		{
			name:     "no filter",
			tables:   []string{"users", "orders", "products"},
			cfg:      &config.ConnectorConfig{},
			opts:     nil,
			expected: []string{"users", "orders", "products"},
		},
		{
			name:   "config include filter",
			tables: []string{"users", "orders", "products", "logs"},
			cfg: &config.ConnectorConfig{
				Matching: &config.MatchingConfig{
					PatternType: "glob",
					Tables: &config.MatchingRule{
						Include: []string{"user*", "order*"},
					},
				},
			},
			opts:     nil,
			expected: []string{"users", "orders"},
		},
		{
			name:   "config exclude filter",
			tables: []string{"users", "orders", "products", "logs"},
			cfg: &config.ConnectorConfig{
				Matching: &config.MatchingConfig{
					PatternType: "glob",
					Tables: &config.MatchingRule{
						Exclude: []string{"logs"},
					},
				},
			},
			opts:     nil,
			expected: []string{"users", "orders", "products"},
		},
		{
			name:   "request filter",
			tables: []string{"users", "orders", "products"},
			cfg:    &config.ConnectorConfig{},
			opts: &collector.ListOptions{
				Filter: &collector.MatchingRule{
					Include: []string{"users"},
				},
			},
			expected: []string{"users"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Collector{config: tt.cfg}
			result := c.filterTables(tt.tables, tt.opts)
			if len(result) != len(tt.expected) {
				t.Errorf("expected %d tables, got %d", len(tt.expected), len(result))
				return
			}
			for i, table := range result {
				if table != tt.expected[i] {
					t.Errorf("expected table[%d] = %s, got %s", i, tt.expected[i], table)
				}
			}
		})
	}
}

// TestConvertElasticsearchType tests the type conversion logic
func TestConvertElasticsearchType(t *testing.T) {
	c := &Collector{}

	tests := []struct {
		esType   string
		expected string
	}{
		{"text", "string"},
		{"keyword", "string"},
		{"long", "bigint"},
		{"integer", "int"},
		{"short", "smallint"},
		{"byte", "tinyint"},
		{"double", "double"},
		{"float", "float"},
		{"half_float", "float"},
		{"scaled_float", "decimal"},
		{"date", "timestamp"},
		{"boolean", "boolean"},
		{"binary", "binary"},
		{"object", "object"},
		{"nested", "object"},
		{"geo_point", "geometry"},
		{"geo_shape", "geometry"},
		{"ip", "string"},
		{"unknown_type", "string"},
	}

	for _, tt := range tests {
		t.Run(tt.esType, func(t *testing.T) {
			result := c.convertElasticsearchType(tt.esType)
			if result != tt.expected {
				t.Errorf("expected %s, got %s", tt.expected, result)
			}
		})
	}
}

// TestExtractFieldsFromProperties tests the field extraction from Elasticsearch properties
func TestExtractFieldsFromProperties(t *testing.T) {
	c := &Collector{
		inferrer: &infer.DocumentInferrer{},
	}
	// Set a default config to avoid nil pointer
	c.inferrer.SetConfig(&infer.InferConfig{MaxDepth: 10})

	properties := map[string]interface{}{
		"name": map[string]interface{}{
			"type": "text",
		},
		"age": map[string]interface{}{
			"type": "integer",
		},
		"address": map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"street": map[string]interface{}{
					"type": "text",
				},
				"city": map[string]interface{}{
					"type": "keyword",
				},
			},
		},
	}

	columns := c.extractFieldsFromProperties(properties, "", 0)

	// Should have at least 4 columns: name, age, address, address.street, address.city
	if len(columns) < 4 {
		t.Errorf("expected at least 4 columns, got %d", len(columns))
	}

	// Check specific fields
	fieldMap := make(map[string]collector.Column)
	for _, col := range columns {
		fieldMap[col.Name] = col
	}

	if col, exists := fieldMap["name"]; exists {
		if col.Type != "string" {
			t.Errorf("expected name type to be string, got %s", col.Type)
		}
		if col.SourceType != "text" {
			t.Errorf("expected name source type to be text, got %s", col.SourceType)
		}
	} else {
		t.Error("expected name field to exist")
	}

	if col, exists := fieldMap["age"]; exists {
		if col.Type != "int" {
			t.Errorf("expected age type to be int, got %s", col.Type)
		}
		if col.SourceType != "integer" {
			t.Errorf("expected age source type to be integer, got %s", col.SourceType)
		}
	} else {
		t.Error("expected age field to exist")
	}

	if col, exists := fieldMap["address.street"]; exists {
		if col.Type != "string" {
			t.Errorf("expected address.street type to be string, got %s", col.Type)
		}
	} else {
		t.Error("expected address.street field to exist")
	}
}

// TestSetInferConfig tests the schema inference configuration
func TestSetInferConfig(t *testing.T) {
	cfg := &config.ConnectorConfig{
		Type:     "elasticsearch",
		Endpoint: "localhost:9200",
	}
	c, _ := NewCollector(cfg)
	esCollector := c.(*Collector)

	// Test with nil config
	esCollector.SetInferConfig(nil)
	// Should not panic

	// Test with valid config
	inferConfig := &config.InferConfig{
		Enabled:    true,
		SampleSize: 200,
		MaxDepth:   15,
		TypeMerge:  config.TypeMergeUnion,
	}
	esCollector.SetInferConfig(inferConfig)

	// Verify the configuration was applied
	actualConfig := esCollector.inferrer.GetConfig()
	if actualConfig.Enabled != inferConfig.Enabled {
		t.Errorf("expected Enabled %v, got %v", inferConfig.Enabled, actualConfig.Enabled)
	}
	if actualConfig.SampleSize != inferConfig.SampleSize {
		t.Errorf("expected SampleSize %d, got %d", inferConfig.SampleSize, actualConfig.SampleSize)
	}
	if actualConfig.MaxDepth != inferConfig.MaxDepth {
		t.Errorf("expected MaxDepth %d, got %d", inferConfig.MaxDepth, actualConfig.MaxDepth)
	}
}

// TestFetchPartitions tests that partitions return empty for Elasticsearch
func TestFetchPartitions(t *testing.T) {
	cfg := &config.ConnectorConfig{
		Type:     "elasticsearch",
		Endpoint: "localhost:9200",
	}
	c, _ := NewCollector(cfg)
	ctx := context.Background()

	// Should return empty partitions without connecting (Elasticsearch doesn't support partitions)
	partitions, err := c.FetchPartitions(ctx, "elasticsearch", "", "users")
	if err != nil {
		t.Errorf("FetchPartitions should not return error: %v", err)
	}
	if len(partitions) != 0 {
		t.Errorf("expected 0 partitions, got %d", len(partitions))
	}
}

// TestListSchemas tests that schemas return empty for Elasticsearch
func TestListSchemas(t *testing.T) {
	cfg := &config.ConnectorConfig{
		Type:     "elasticsearch",
		Endpoint: "localhost:9200",
	}
	c, _ := NewCollector(cfg)

	// Should return connection closed error when not connected
	_, err := c.ListSchemas(context.Background(), "elasticsearch")
	assertConnectionClosedError(t, err, "ListSchemas")
}

// TestConstants tests the package constants
func TestConstants(t *testing.T) {
	if SourceName != "elasticsearch" {
		t.Errorf("expected SourceName to be 'elasticsearch', got %s", SourceName)
	}
	if DefaultPort != 9200 {
		t.Errorf("expected DefaultPort to be 9200, got %d", DefaultPort)
	}
	if DefaultTimeout != 30 {
		t.Errorf("expected DefaultTimeout to be 30, got %d", DefaultTimeout)
	}
	if DefaultSampleSize != 100 {
		t.Errorf("expected DefaultSampleSize to be 100, got %d", DefaultSampleSize)
	}
}

// Helper functions

func assertConnectionClosedError(t *testing.T, err error, operation string) {
	t.Helper()
	if err == nil {
		t.Errorf("%s should return error when not connected", operation)
		return
	}
	var collErr *collector.CollectorError
	if !errors.As(err, &collErr) {
		t.Errorf("%s: expected CollectorError, got %T", operation, err)
		return
	}
	if collErr.Code != collector.ErrCodeConnectionClosed {
		t.Errorf("%s: expected error code %s, got %s", operation, collector.ErrCodeConnectionClosed, collErr.Code)
	}
	if collErr.Category != collector.CategoryDocumentDB {
		t.Errorf("%s: expected category %s, got %s", operation, collector.CategoryDocumentDB, collErr.Category)
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}