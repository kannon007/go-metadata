// Package postgres provides a PostgreSQL metadata collector implementation.
package postgres

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
				Type:     "postgres",
				Endpoint: "localhost:5432",
				Credentials: config.Credentials{
					User:     "postgres",
					Password: "password",
				},
			},
			wantErr: false,
		},
		{
			name: "empty type (allowed)",
			cfg: &config.ConnectorConfig{
				Type:     "",
				Endpoint: "localhost:5432",
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
			}
		})
	}
}

// TestBuildDSN tests the DSN building logic
func TestBuildDSN(t *testing.T) {
	tests := []struct {
		name        string
		cfg         *config.ConnectorConfig
		wantErr     bool
		wantContain string
	}{
		{
			name: "basic endpoint",
			cfg: &config.ConnectorConfig{
				Endpoint: "localhost:5432",
				Credentials: config.Credentials{
					User:     "postgres",
					Password: "password",
				},
			},
			wantErr:     false,
			wantContain: "host=localhost port=5432 user=postgres password=password",
		},
		{
			name: "endpoint without port",
			cfg: &config.ConnectorConfig{
				Endpoint: "localhost",
				Credentials: config.Credentials{
					User:     "postgres",
					Password: "password",
				},
			},
			wantErr:     false,
			wantContain: "port=5432",
		},
		{
			name: "with database",
			cfg: &config.ConnectorConfig{
				Endpoint: "localhost:5432",
				Credentials: config.Credentials{
					User:     "postgres",
					Password: "password",
				},
				Properties: config.ConnectionProps{
					Extra: map[string]string{
						"database": "testdb",
					},
				},
			},
			wantErr:     false,
			wantContain: "dbname=testdb",
		},
		{
			name: "with custom timeout",
			cfg: &config.ConnectorConfig{
				Endpoint: "localhost:5432",
				Credentials: config.Credentials{
					User:     "postgres",
					Password: "password",
				},
				Properties: config.ConnectionProps{
					ConnectionTimeout: 60,
				},
			},
			wantErr:     false,
			wantContain: "connect_timeout=60",
		},
		{
			name: "with sslmode",
			cfg: &config.ConnectorConfig{
				Endpoint: "localhost:5432",
				Credentials: config.Credentials{
					User:     "postgres",
					Password: "password",
				},
				Properties: config.ConnectionProps{
					Extra: map[string]string{
						"sslmode": "require",
					},
				},
			},
			wantErr:     false,
			wantContain: "sslmode=require",
		},
		{
			name: "empty endpoint",
			cfg: &config.ConnectorConfig{
				Endpoint: "",
				Credentials: config.Credentials{
					User:     "postgres",
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
					User:     "postgres",
					Password: "password",
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Collector{config: tt.cfg}
			dsn, err := c.buildDSN()
			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
					return
				}
				if tt.wantContain != "" && !contains(dsn, tt.wantContain) {
					t.Errorf("DSN %q should contain %q", dsn, tt.wantContain)
				}
			}
		})
	}
}

// TestMapTableType tests the table type mapping
func TestMapTableType(t *testing.T) {
	c := &Collector{}

	tests := []struct {
		pgType string
		want   collector.TableType
	}{
		{"BASE TABLE", collector.TableTypeTable},
		{"VIEW", collector.TableTypeView},
		{"MATERIALIZED VIEW", collector.TableTypeMaterializedView},
		{"FOREIGN TABLE", collector.TableTypeExternalTable},
		{"base table", collector.TableTypeTable},
		{"view", collector.TableTypeView},
		{"UNKNOWN", collector.TableTypeTable},
		{"", collector.TableTypeTable},
	}

	for _, tt := range tests {
		t.Run(tt.pgType, func(t *testing.T) {
			got := c.mapTableType(tt.pgType)
			if got != tt.want {
				t.Errorf("mapTableType(%q) = %v, want %v", tt.pgType, got, tt.want)
			}
		})
	}
}

// TestNormalizeType tests the type normalization
func TestNormalizeType(t *testing.T) {
	c := &Collector{}

	tests := []struct {
		dataType string
		want     string
	}{
		{"integer", "INTEGER"},
		{"smallint", "INTEGER"},
		{"bigint", "INTEGER"},
		{"int2", "INTEGER"},
		{"int4", "INTEGER"},
		{"int8", "INTEGER"},
		{"real", "FLOAT"},
		{"double precision", "FLOAT"},
		{"float4", "FLOAT"},
		{"float8", "FLOAT"},
		{"numeric", "DECIMAL"},
		{"decimal", "DECIMAL"},
		{"character", "STRING"},
		{"character varying", "STRING"},
		{"text", "STRING"},
		{"char", "STRING"},
		{"varchar", "STRING"},
		{"date", "DATE"},
		{"time", "TIME"},
		{"time without time zone", "TIME"},
		{"time with time zone", "TIME"},
		{"timestamp", "TIMESTAMP"},
		{"timestamp without time zone", "TIMESTAMP"},
		{"timestamp with time zone", "TIMESTAMP"},
		{"bytea", "BINARY"},
		{"boolean", "BOOLEAN"},
		{"json", "JSON"},
		{"jsonb", "JSON"},
		{"uuid", "UUID"},
		{"array", "ARRAY"},
		{"unknown_type", "UNKNOWN_TYPE"},
	}

	for _, tt := range tests {
		t.Run(tt.dataType, func(t *testing.T) {
			got := c.normalizeType(tt.dataType)
			if got != tt.want {
				t.Errorf("normalizeType(%q) = %v, want %v", tt.dataType, got, tt.want)
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
			name:     "password authentication failed",
			err:      errors.New("pq: password authentication failed for user"),
			wantCode: collector.ErrCodeAuthError,
		},
		{
			name:     "authentication failed",
			err:      errors.New("authentication failed"),
			wantCode: collector.ErrCodeAuthError,
		},
		{
			name:     "connection refused",
			err:      errors.New("dial tcp: connection refused"),
			wantCode: collector.ErrCodeNetworkError,
		},
		{
			name:     "no such host",
			err:      errors.New("dial tcp: lookup unknown: no such host"),
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
		})
	}
}

// TestCollectorNotConnected tests operations when not connected
func TestCollectorNotConnected(t *testing.T) {
	cfg := &config.ConnectorConfig{
		Type:     "postgres",
		Endpoint: "localhost:5432",
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

	_, err = c.ListSchemas(ctx, "testdb")
	assertConnectionClosedError(t, err, "ListSchemas")

	_, err = c.ListTables(ctx, "testdb", "public", nil)
	assertConnectionClosedError(t, err, "ListTables")

	_, err = c.FetchTableMetadata(ctx, "testdb", "public", "users")
	assertConnectionClosedError(t, err, "FetchTableMetadata")

	_, err = c.FetchTableStatistics(ctx, "testdb", "public", "users")
	assertConnectionClosedError(t, err, "FetchTableStatistics")

	_, err = c.FetchPartitions(ctx, "testdb", "public", "users")
	assertConnectionClosedError(t, err, "FetchPartitions")
}

// TestCloseNotConnected tests Close when not connected
func TestCloseNotConnected(t *testing.T) {
	cfg := &config.ConnectorConfig{
		Type:     "postgres",
		Endpoint: "localhost:5432",
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
		{
			name:   "schema filter",
			tables: []string{"users", "orders", "products"},
			cfg: &config.ConnectorConfig{
				Matching: &config.MatchingConfig{
					PatternType: "glob",
					Schemas: &config.MatchingRule{
						Include: []string{"public"},
					},
				},
			},
			opts:     nil,
			expected: []string{"users", "orders", "products"},
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
}

func contains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
