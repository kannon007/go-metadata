package clickhouse

import (
	"testing"

	"go-metadata/internal/collector"
	"go-metadata/internal/collector/config"
)

func TestNewCollector(t *testing.T) {
	tests := []struct {
		name    string
		config  *config.ConnectorConfig
		wantErr bool
		errType string
	}{
		{
			name:    "nil config",
			config:  nil,
			wantErr: true,
			errType: "INVALID_CONFIG",
		},
		{
			name: "valid config",
			config: &config.ConnectorConfig{
				Type:     SourceName,
				Endpoint: "localhost:9000",
				Credentials: config.Credentials{
					User:     "testuser",
					Password: "testpass",
				},
				Properties: config.ConnectionProps{
					Extra: map[string]string{
						"database": "testdb",
					},
				},
			},
			wantErr: false,
		},
		{
			name: "empty type (should be allowed)",
			config: &config.ConnectorConfig{
				Endpoint: "localhost:9000",
				Credentials: config.Credentials{
					User:     "testuser",
					Password: "testpass",
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, err := NewCollector(tt.config)

			if tt.wantErr {
				if err == nil {
					t.Errorf("NewCollector() expected error, got nil")
					return
				}

				if tt.errType != "" {
					if collErr, ok := err.(*collector.CollectorError); ok {
						if string(collErr.Code) != tt.errType {
							t.Errorf("NewCollector() error type = %v, want %v", collErr.Code, tt.errType)
						}
					} else {
						t.Errorf("NewCollector() error type = %T, want *collector.CollectorError", err)
					}
				}
				return
			}

			if err != nil {
				t.Errorf("NewCollector() unexpected error = %v", err)
				return
			}

			if c == nil {
				t.Errorf("NewCollector() returned nil collector")
				return
			}

			// Test interface compliance
			if c.Category() != collector.CategoryDataWarehouse {
				t.Errorf("Category() = %v, want %v", c.Category(), collector.CategoryDataWarehouse)
			}

			if c.Type() != SourceName {
				t.Errorf("Type() = %v, want %v", c.Type(), SourceName)
			}
		})
	}
}

func TestBuildDSN(t *testing.T) {
	tests := []struct {
		name   string
		config *config.ConnectorConfig
		want   string
	}{
		{
			name: "basic connection",
			config: &config.ConnectorConfig{
				Endpoint: "localhost:9000",
				Credentials: config.Credentials{
					User:     "testuser",
					Password: "testpass",
				},
				Properties: config.ConnectionProps{
					Extra: map[string]string{
						"database": "testdb",
					},
				},
			},
			want: "clickhouse://testuser:testpass@localhost:9000/testdb",
		},
		{
			name: "default port",
			config: &config.ConnectorConfig{
				Endpoint: "localhost",
				Credentials: config.Credentials{
					User:     "testuser",
					Password: "testpass",
				},
				Properties: config.ConnectionProps{
					Extra: map[string]string{
						"database": "testdb",
					},
				},
			},
			want: "clickhouse://testuser:testpass@localhost:9000/testdb",
		},
		{
			name: "default database",
			config: &config.ConnectorConfig{
				Endpoint: "localhost:9000",
				Credentials: config.Credentials{
					User:     "testuser",
					Password: "testpass",
				},
			},
			want: "clickhouse://testuser:testpass@localhost:9000/default",
		},
		{
			name: "with connection timeout",
			config: &config.ConnectorConfig{
				Endpoint: "localhost:9000",
				Credentials: config.Credentials{
					User:     "testuser",
					Password: "testpass",
				},
				Properties: config.ConnectionProps{
					ConnectionTimeout: 30,
					Extra: map[string]string{
						"database": "testdb",
					},
				},
			},
			want: "clickhouse://testuser:testpass@localhost:9000/testdb?dial_timeout=30s",
		},
		{
			name: "empty endpoint uses default",
			config: &config.ConnectorConfig{
				Credentials: config.Credentials{
					User:     "testuser",
					Password: "testpass",
				},
			},
			want: "clickhouse://testuser:testpass@localhost:9000/default",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Collector{config: tt.config}
			got := c.buildDSN()

			if got != tt.want {
				t.Errorf("buildDSN() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMapClickHouseTypeToSQL(t *testing.T) {
	c := &Collector{}

	tests := []struct {
		clickhouseType string
		want           string
	}{
		{"Int8", "TINYINT"},
		{"Int16", "SMALLINT"},
		{"Int32", "INTEGER"},
		{"Int64", "BIGINT"},
		{"UInt8", "TINYINT"},
		{"UInt16", "SMALLINT"},
		{"UInt32", "INTEGER"},
		{"UInt64", "BIGINT"},
		{"Float32", "REAL"},
		{"Float64", "DOUBLE"},
		{"Decimal(10,2)", "DECIMAL"},
		{"String", "TEXT"},
		{"FixedString(10)", "CHAR"},
		{"Date", "DATE"},
		{"DateTime", "TIMESTAMP"},
		{"UUID", "UUID"},
		{"Array(String)", "ARRAY"},
		{"Tuple(String, Int32)", "STRUCT"},
		{"Enum8('a'=1, 'b'=2)", "ENUM"},
		{"Bool", "BOOLEAN"},
		{"Nullable(String)", "TEXT"},
		{"Nullable(Int32)", "INTEGER"},
		{"UnknownType", "TEXT"}, // Default fallback
	}

	for _, tt := range tests {
		t.Run(tt.clickhouseType, func(t *testing.T) {
			got := c.mapClickHouseTypeToSQL(tt.clickhouseType)
			if got != tt.want {
				t.Errorf("mapClickHouseTypeToSQL(%v) = %v, want %v", tt.clickhouseType, got, tt.want)
			}
		})
	}
}

func TestConstants(t *testing.T) {
	if SourceName != "clickhouse" {
		t.Errorf("SourceName = %v, want clickhouse", SourceName)
	}
}