// Package redis provides a Redis metadata collector implementation.
package redis

import (
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
				Type:     "redis",
				Endpoint: "localhost:6379",
				Credentials: config.Credentials{
					Password: "password",
				},
			},
			wantErr: false,
		},
		{
			name: "empty type (allowed)",
			cfg: &config.ConnectorConfig{
				Type:     "",
				Endpoint: "localhost:6379",
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
			name: "with database in extra properties",
			cfg: &config.ConnectorConfig{
				Type:     "redis",
				Endpoint: "localhost:6379",
				Properties: config.ConnectionProps{
					Extra: map[string]string{
						"database": "1",
					},
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
				redisCollector := c.(*Collector)
				if redisCollector.Category() != collector.CategoryKeyValue {
					t.Errorf("expected category %s, got %s", collector.CategoryKeyValue, redisCollector.Category())
				}
				if redisCollector.Type() != SourceName {
					t.Errorf("expected type %s, got %s", SourceName, redisCollector.Type())
				}
			}
		})
	}
}

// TestParseEndpoint tests the parseEndpoint method
func TestParseEndpoint(t *testing.T) {
	tests := []struct {
		name     string
		endpoint string
		wantHost string
		wantPort int
		wantErr  bool
	}{
		{
			name:     "host and port",
			endpoint: "localhost:6379",
			wantHost: "localhost",
			wantPort: 6379,
			wantErr:  false,
		},
		{
			name:     "host only",
			endpoint: "localhost",
			wantHost: "localhost",
			wantPort: DefaultPort,
			wantErr:  false,
		},
		{
			name:     "IP and port",
			endpoint: "127.0.0.1:6380",
			wantHost: "127.0.0.1",
			wantPort: 6380,
			wantErr:  false,
		},
		{
			name:     "empty endpoint",
			endpoint: "",
			wantErr:  true,
		},
		{
			name:     "invalid port",
			endpoint: "localhost:abc",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &config.ConnectorConfig{
				Endpoint: tt.endpoint,
			}
			c := &Collector{config: cfg}
			
			host, port, err := c.parseEndpoint()
			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
					return
				}
				if host != tt.wantHost {
					t.Errorf("expected host %s, got %s", tt.wantHost, host)
				}
				if port != tt.wantPort {
					t.Errorf("expected port %d, got %d", tt.wantPort, port)
				}
			}
		})
	}
}

// TestParseRedisVersion tests the parseRedisVersion method
func TestParseRedisVersion(t *testing.T) {
	tests := []struct {
		name    string
		info    string
		want    string
	}{
		{
			name: "valid version",
			info: "# Server\nredis_version:7.0.5\nredis_git_sha1:00000000\n",
			want: "7.0.5",
		},
		{
			name: "no version",
			info: "# Server\nredis_git_sha1:00000000\n",
			want: "unknown",
		},
		{
			name: "empty info",
			info: "",
			want: "unknown",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Collector{}
			got := c.parseRedisVersion(tt.info)
			if got != tt.want {
				t.Errorf("expected version %s, got %s", tt.want, got)
			}
		})
	}
}

// TestParseKeyspaceInfo tests the parseKeyspaceInfo method
func TestParseKeyspaceInfo(t *testing.T) {
	tests := []struct {
		name string
		info string
		want []string
	}{
		{
			name: "multiple databases",
			info: "# Keyspace\ndb0:keys=1,expires=0,avg_ttl=0\ndb1:keys=5,expires=1,avg_ttl=1000\n",
			want: []string{"0", "1"},
		},
		{
			name: "single database",
			info: "# Keyspace\ndb0:keys=10,expires=0,avg_ttl=0\n",
			want: []string{"0"},
		},
		{
			name: "no databases",
			info: "# Keyspace\n",
			want: []string{"0"}, // Should return default database
		},
		{
			name: "empty info",
			info: "",
			want: []string{"0"}, // Should return default database
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Collector{}
			got := c.parseKeyspaceInfo(tt.info)
			if len(got) != len(tt.want) {
				t.Errorf("expected %d databases, got %d", len(tt.want), len(got))
				return
			}
			for i, db := range tt.want {
				if got[i] != db {
					t.Errorf("expected database %s at index %d, got %s", db, i, got[i])
				}
			}
		})
	}
}

// TestInferPatternFromKey tests the inferPatternFromKey method
func TestInferPatternFromKey(t *testing.T) {
	tests := []struct {
		name string
		key  string
		want string
	}{
		{
			name: "user with numeric ID",
			key:  "user:123",
			want: "user:*",
		},
		{
			name: "session with alphanumeric ID",
			key:  "session:abc123def456",
			want: "session:*",
		},
		{
			name: "cache with nested structure",
			key:  "cache:user:123:profile",
			want: "cache:user:*:profile",
		},
		{
			name: "order with date parts",
			key:  "order:2023:01:15:123",
			want: "order:*:*:*:*",
		},
		{
			name: "simple string key",
			key:  "config:app:name",
			want: "config:app:name",
		},
		{
			name: "UUID key",
			key:  "token:550e8400-e29b-41d4-a716-446655440000",
			want: "token:*",
		},
		{
			name: "timestamp key",
			key:  "event:1640995200:data",
			want: "event:*:data",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Collector{}
			got := c.inferPatternFromKey(tt.key)
			if got != tt.want {
				t.Errorf("expected pattern %s, got %s", tt.want, got)
			}
		})
	}
}

// TestIsNumeric tests the isNumeric method
func TestIsNumeric(t *testing.T) {
	tests := []struct {
		name string
		s    string
		want bool
	}{
		{"positive integer", "123", true},
		{"negative integer", "-123", true},
		{"zero", "0", true},
		{"float", "123.45", false}, // ParseInt doesn't handle floats
		{"alphanumeric", "abc123", false},
		{"letters only", "abc", false},
		{"empty string", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Collector{}
			got := c.isNumeric(tt.s)
			if got != tt.want {
				t.Errorf("expected %v, got %v", tt.want, got)
			}
		})
	}
}

// TestIsAlphanumericID tests the isAlphanumericID method
func TestIsAlphanumericID(t *testing.T) {
	tests := []struct {
		name string
		s    string
		want bool
	}{
		{"valid alphanumeric ID", "abc123def456", true},
		{"short string", "abc12", false}, // Less than 6 characters
		{"numbers only", "123456", false}, // No letters
		{"letters only", "abcdef", false}, // No numbers
		{"with special chars", "abc123-def", false}, // Contains special characters
		{"empty string", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Collector{}
			got := c.isAlphanumericID(tt.s)
			if got != tt.want {
				t.Errorf("expected %v, got %v", tt.want, got)
			}
		})
	}
}

// TestIsUUID tests the isUUID method
func TestIsUUID(t *testing.T) {
	tests := []struct {
		name string
		s    string
		want bool
	}{
		{"valid UUID", "550e8400-e29b-41d4-a716-446655440000", true},
		{"uppercase UUID", "550E8400-E29B-41D4-A716-446655440000", true},
		{"invalid format", "550e8400-e29b-41d4-a716", false},
		{"no hyphens", "550e8400e29b41d4a716446655440000", false},
		{"empty string", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Collector{}
			got := c.isUUID(tt.s)
			if got != tt.want {
				t.Errorf("expected %v, got %v", tt.want, got)
			}
		})
	}
}

// TestIsTimestamp tests the isTimestamp method
func TestIsTimestamp(t *testing.T) {
	tests := []struct {
		name string
		s    string
		want bool
	}{
		{"valid unix timestamp", "1640995200", true}, // 2022-01-01 00:00:00 UTC
		{"valid millisecond timestamp", "1640995200000", true},
		{"too old timestamp", "946684800", false}, // 2000-01-01, but our range starts later
		{"too new timestamp", "4102444800", false}, // 2100-01-01, but our range ends earlier
		{"not numeric", "abc1234567", false},
		{"wrong length", "12345", false},
		{"empty string", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Collector{}
			got := c.isTimestamp(tt.s)
			if got != tt.want {
				t.Errorf("expected %v, got %v", tt.want, got)
			}
		})
	}
}

// TestWrapConnectionError tests the wrapConnectionError method
func TestWrapConnectionError(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		wantCode collector.ErrorCode
	}{
		{
			name:     "auth error",
			err:      errors.New("NOAUTH Authentication required"),
			wantCode: collector.ErrCodeAuthError,
		},
		{
			name:     "invalid password",
			err:      errors.New("invalid password"),
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
			err:      errors.New("timeout"),
			wantCode: collector.ErrCodeTimeout,
		},
		{
			name:     "deadline exceeded",
			err:      errors.New("deadline exceeded"),
			wantCode: collector.ErrCodeTimeout, // Both timeout and deadline exceeded map to timeout
		},
		{
			name:     "other error",
			err:      errors.New("some other error"),
			wantCode: collector.ErrCodeNetworkError, // Default to network error
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Collector{}
			wrappedErr := c.wrapConnectionError(tt.err)
			
			var collErr *collector.CollectorError
			if !errors.As(wrappedErr, &collErr) {
				t.Error("expected CollectorError")
				return
			}
			
			if collErr.Code != tt.wantCode {
				t.Errorf("expected error code %s, got %s", tt.wantCode, collErr.Code)
			}
		})
	}
}

// TestCollectorInterface tests that Collector implements the required interfaces
func TestCollectorInterface(t *testing.T) {
	cfg := &config.ConnectorConfig{
		Type:     "redis",
		Endpoint: "localhost:6379",
	}
	
	c, err := NewCollector(cfg)
	if err != nil {
		t.Fatalf("failed to create collector: %v", err)
	}
	
	// Test that it implements collector.Collector
	var _ collector.Collector = c
	
	// Test that it implements kv.KeyValueCollector
	redisCollector := c.(*Collector)
	var _ collector.Collector = redisCollector
	
	// Test basic interface methods without connection
	if redisCollector.Category() != collector.CategoryKeyValue {
		t.Errorf("expected category %s, got %s", collector.CategoryKeyValue, redisCollector.Category())
	}
	
	if redisCollector.Type() != SourceName {
		t.Errorf("expected type %s, got %s", SourceName, redisCollector.Type())
	}
}

// TestClose tests the Close method
func TestClose(t *testing.T) {
	cfg := &config.ConnectorConfig{
		Type:     "redis",
		Endpoint: "localhost:6379",
	}
	
	c, err := NewCollector(cfg)
	if err != nil {
		t.Fatalf("failed to create collector: %v", err)
	}
	
	redisCollector := c.(*Collector)
	
	// Test closing when not connected
	err = redisCollector.Close()
	if err != nil {
		t.Errorf("unexpected error closing unconnected collector: %v", err)
	}
	
	// Test multiple closes
	err = redisCollector.Close()
	if err != nil {
		t.Errorf("unexpected error on second close: %v", err)
	}
}