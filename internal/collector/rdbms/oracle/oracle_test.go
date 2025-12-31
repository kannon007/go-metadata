package oracle

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
				Endpoint: "localhost:1521",
				Credentials: config.Credentials{
					User:     "testuser",
					Password: "testpass",
				},
				Properties: config.ConnectionProps{
					Extra: map[string]string{
						"database": "TESTDB",
					},
				},
			},
			wantErr: false,
		},
		{
			name: "wrong type",
			config: &config.ConnectorConfig{
				Type:     "mysql",
				Endpoint: "localhost:1521",
			},
			wantErr: false, // Type validation is not enforced in constructor
		},
		{
			name: "empty type (should be allowed)",
			config: &config.ConnectorConfig{
				Endpoint: "localhost:1521",
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
			if c.Category() != collector.CategoryRDBMS {
				t.Errorf("Category() = %v, want %v", c.Category(), collector.CategoryRDBMS)
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
				Endpoint: "localhost:1521",
				Credentials: config.Credentials{
					User:     "testuser",
					Password: "testpass",
				},
				Properties: config.ConnectionProps{
					Extra: map[string]string{
						"database": "TESTDB",
					},
				},
			},
			want: "testuser/testpass@localhost:1521/TESTDB",
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
						"database": "TESTDB",
					},
				},
			},
			want: "testuser/testpass@localhost:1521/TESTDB",
		},
		{
			name: "default service name",
			config: &config.ConnectorConfig{
				Endpoint: "localhost:1521",
				Credentials: config.Credentials{
					User:     "testuser",
					Password: "testpass",
				},
			},
			want: "testuser/testpass@localhost:1521/XE",
		},
		{
			name: "empty endpoint uses default",
			config: &config.ConnectorConfig{
				Credentials: config.Credentials{
					User:     "testuser",
					Password: "testpass",
				},
			},
			want: "testuser/testpass@localhost:1521/XE",
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

func TestConstants(t *testing.T) {
	if SourceName != "oracle" {
		t.Errorf("SourceName = %v, want oracle", SourceName)
	}
}

// Mock error for testing
type mockError struct {
	msg string
}

func (e *mockError) Error() string {
	return e.msg
}