package doris

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
				Endpoint: "localhost:9030",
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
				Endpoint: "localhost:9030",
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
				Endpoint: "localhost:9030",
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
			want: "testuser:testpass@tcp(localhost:9030)/testdb?parseTime=true",
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
			want: "testuser:testpass@tcp(localhost:9030)/testdb?parseTime=true",
		},
		{
			name: "no database",
			config: &config.ConnectorConfig{
				Endpoint: "localhost:9030",
				Credentials: config.Credentials{
					User:     "testuser",
					Password: "testpass",
				},
			},
			want: "testuser:testpass@tcp(localhost:9030)/?parseTime=true",
		},
		{
			name: "with connection timeout",
			config: &config.ConnectorConfig{
				Endpoint: "localhost:9030",
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
			want: "testuser:testpass@tcp(localhost:9030)/testdb?parseTime=true&timeout=30s",
		},
		{
			name: "empty endpoint uses default",
			config: &config.ConnectorConfig{
				Credentials: config.Credentials{
					User:     "testuser",
					Password: "testpass",
				},
			},
			want: "testuser:testpass@tcp(localhost:9030)/?parseTime=true",
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

func TestMapDorisTypeToSQL(t *testing.T) {
	c := &Collector{}

	tests := []struct {
		dorisType string
		want      string
	}{
		{"TINYINT", "TINYINT"},
		{"SMALLINT", "SMALLINT"},
		{"INT", "INTEGER"},
		{"BIGINT", "BIGINT"},
		{"LARGEINT", "BIGINT"},
		{"FLOAT", "REAL"},
		{"DOUBLE", "DOUBLE"},
		{"DECIMAL(10,2)", "DECIMAL"},
		{"CHAR(10)", "CHAR"},
		{"VARCHAR(255)", "VARCHAR"},
		{"STRING", "TEXT"},
		{"TEXT", "TEXT"},
		{"DATE", "DATE"},
		{"DATETIME", "TIMESTAMP"},
		{"TIMESTAMP", "TIMESTAMP"},
		{"BOOLEAN", "BOOLEAN"},
		{"ARRAY<STRING>", "ARRAY"},
		{"MAP<STRING,INT>", "MAP"},
		{"STRUCT<name:STRING,age:INT>", "STRUCT"},
		{"JSON", "JSON"},
		{"BITMAP", "BLOB"},
		{"HLL", "BLOB"},
		{"UnknownType", "TEXT"}, // Default fallback
	}

	for _, tt := range tests {
		t.Run(tt.dorisType, func(t *testing.T) {
			got := c.mapDorisTypeToSQL(tt.dorisType)
			if got != tt.want {
				t.Errorf("mapDorisTypeToSQL(%v) = %v, want %v", tt.dorisType, got, tt.want)
			}
		})
	}
}

func TestConstants(t *testing.T) {
	if SourceName != "doris" {
		t.Errorf("SourceName = %v, want doris", SourceName)
	}
}