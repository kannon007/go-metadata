// Package mongodb provides a MongoDB metadata collector implementation.
package mongodb

import (
	"context"
	"errors"
	"testing"

	"go-metadata/internal/collector"
	"go-metadata/internal/collector/config"
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
				Type:     "mongodb",
				Endpoint: "localhost:27017",
				Credentials: config.Credentials{
					User:     "admin",
					Password: "password",
				},
			},
			wantErr: false,
		},
		{
			name: "empty type (allowed)",
			cfg: &config.ConnectorConfig{
				Type:     "",
				Endpoint: "localhost:27017",
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
				Type:     "mongodb",
				Endpoint: "localhost:27017",
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
				mongoCollector := c.(*Collector)
				if mongoCollector.Category() != collector.CategoryDocumentDB {
					t.Errorf("expected category %s, got %s", collector.CategoryDocumentDB, mongoCollector.Category())
				}
				if mongoCollector.Type() != SourceName {
					t.Errorf("expected type %s, got %s", SourceName, mongoCollector.Type())
				}
			}
		})
	}
}

// TestBuildURI tests the URI building logic
func TestBuildURI(t *testing.T) {
	tests := []struct {
		name        string
		cfg         *config.ConnectorConfig
		wantErr     bool
		wantContain string
	}{
		{
			name: "basic endpoint",
			cfg: &config.ConnectorConfig{
				Endpoint: "localhost:27017",
				Credentials: config.Credentials{
					User:     "admin",
					Password: "password",
				},
			},
			wantErr:     false,
			wantContain: "mongodb://admin:password@localhost:27017",
		},
		{
			name: "endpoint without port",
			cfg: &config.ConnectorConfig{
				Endpoint: "localhost",
				Credentials: config.Credentials{
					User:     "admin",
					Password: "password",
				},
			},
			wantErr:     false,
			wantContain: "mongodb://admin:password@localhost:27017",
		},
		{
			name: "no credentials",
			cfg: &config.ConnectorConfig{
				Endpoint: "localhost:27017",
			},
			wantErr:     false,
			wantContain: "mongodb://localhost:27017",
		},
		{
			name: "with database",
			cfg: &config.ConnectorConfig{
				Endpoint: "localhost:27017",
				Credentials: config.Credentials{
					User:     "admin",
					Password: "password",
				},
				Properties: config.ConnectionProps{
					Extra: map[string]string{
						"database": "testdb",
					},
				},
			},
			wantErr:     false,
			wantContain: "/testdb",
		},
		{
			name: "with extra parameters",
			cfg: &config.ConnectorConfig{
				Endpoint: "localhost:27017",
				Credentials: config.Credentials{
					User:     "admin",
					Password: "password",
				},
				Properties: config.ConnectionProps{
					Extra: map[string]string{
						"authSource": "admin",
						"ssl":        "true",
					},
				},
			},
			wantErr:     false,
			wantContain: "authSource=admin",
		},
		{
			name: "empty endpoint",
			cfg: &config.ConnectorConfig{
				Endpoint: "",
				Credentials: config.Credentials{
					User:     "admin",
					Password: "password",
				},
			},
			wantErr: true,
		},
		{
			name: "invalid port",
			cfg: &config.ConnectorConfig{
				Endpoint: "localhost:invalid",
				Credentials: config.Credentials{
					User:     "admin",
					Password: "password",
				},
			},
			wantErr: true,
		},
		{
			name: "special characters in credentials",
			cfg: &config.ConnectorConfig{
				Endpoint: "localhost:27017",
				Credentials: config.Credentials{
					User:     "user@domain.com",
					Password: "pass:word",
				},
			},
			wantErr:     false,
			wantContain: "user%40domain.com:pass%3Aword@",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Collector{config: tt.cfg}
			uri, err := c.buildURI()
			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
					return
				}
				if tt.wantContain != "" && !contains(uri, tt.wantContain) {
					t.Errorf("URI %q should contain %q", uri, tt.wantContain)
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
		Type:     "mongodb",
		Endpoint: "localhost:27017",
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

	_, err = c.ListSchemas(ctx, "mongodb")
	assertConnectionClosedError(t, err, "ListSchemas")

	_, err = c.ListTables(ctx, "mongodb", "test", nil)
	assertConnectionClosedError(t, err, "ListTables")

	_, err = c.FetchTableMetadata(ctx, "mongodb", "test", "users")
	assertConnectionClosedError(t, err, "FetchTableMetadata")

	_, err = c.FetchTableStatistics(ctx, "mongodb", "test", "users")
	assertConnectionClosedError(t, err, "FetchTableStatistics")

	_, err = c.FetchPartitions(ctx, "mongodb", "test", "users")
	// MongoDB doesn't support partitions, so this should not return a connection error
	if err != nil {
		t.Errorf("FetchPartitions should not return error for MongoDB: %v", err)
	}
}

// TestCloseNotConnected tests Close when not connected
func TestCloseNotConnected(t *testing.T) {
	cfg := &config.ConnectorConfig{
		Type:     "mongodb",
		Endpoint: "localhost:27017",
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

// TestGetQueryTimeout tests the query timeout calculation
func TestGetQueryTimeout(t *testing.T) {
	tests := []struct {
		name     string
		cfg      *config.ConnectorConfig
		expected int64
	}{
		{
			name: "default timeout",
			cfg:  &config.ConnectorConfig{},
			expected: DefaultTimeout * 1000,
		},
		{
			name: "custom timeout",
			cfg: &config.ConnectorConfig{
				Properties: config.ConnectionProps{
					ConnectionTimeout: 60,
				},
			},
			expected: 60 * 1000,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Collector{config: tt.cfg}
			timeout := c.getQueryTimeout()
			if *timeout != tt.expected {
				t.Errorf("expected timeout %d, got %d", tt.expected, *timeout)
			}
		})
	}
}

// TestSetInferConfig tests the schema inference configuration
func TestSetInferConfig(t *testing.T) {
	cfg := &config.ConnectorConfig{
		Type:     "mongodb",
		Endpoint: "localhost:27017",
	}
	c, _ := NewCollector(cfg)
	mongoCollector := c.(*Collector)

	// Test with nil config
	mongoCollector.SetInferConfig(nil)
	// Should not panic

	// Test with valid config
	inferConfig := &config.InferConfig{
		Enabled:    true,
		SampleSize: 200,
		MaxDepth:   15,
		TypeMerge:  config.TypeMergeUnion,
	}
	mongoCollector.SetInferConfig(inferConfig)

	// Verify the configuration was applied
	actualConfig := mongoCollector.inferrer.GetConfig()
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

// TestFetchPartitions tests that partitions return empty for MongoDB
func TestFetchPartitions(t *testing.T) {
	cfg := &config.ConnectorConfig{
		Type:     "mongodb",
		Endpoint: "localhost:27017",
	}
	c, _ := NewCollector(cfg)
	ctx := context.Background()

	// Should return empty partitions without connecting (MongoDB doesn't support partitions)
	partitions, err := c.FetchPartitions(ctx, "mongodb", "test", "users")
	if err != nil {
		t.Errorf("FetchPartitions should not return error: %v", err)
	}
	if len(partitions) != 0 {
		t.Errorf("expected 0 partitions, got %d", len(partitions))
	}
}

// TestConstants tests the package constants
func TestConstants(t *testing.T) {
	if SourceName != "mongodb" {
		t.Errorf("expected SourceName to be 'mongodb', got %s", SourceName)
	}
	if DefaultPort != 27017 {
		t.Errorf("expected DefaultPort to be 27017, got %d", DefaultPort)
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