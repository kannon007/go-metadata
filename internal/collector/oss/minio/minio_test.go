package minio

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
				Endpoint: "localhost:9000",
				Credentials: config.Credentials{
					User:     "minioadmin",
					Password: "minioadmin",
				},
			},
			wantErr: false,
		},
		{
			name: "wrong type",
			config: &config.ConnectorConfig{
				Type:     "mysql",
				Endpoint: "localhost:9000",
			},
			wantErr: true,
		},
		{
			name: "valid config with infer settings",
			config: &config.ConnectorConfig{
				Type:     SourceName,
				Endpoint: "localhost:9000",
				Infer: &config.InferConfig{
					Enabled:    true,
					SampleSize: 50,
					MaxDepth:   5,
				},
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
		Endpoint: "localhost:9000",
	}
	
	c, err := NewCollector(cfg)
	if err != nil {
		t.Fatalf("NewCollector() error = %v", err)
	}
	
	if got := c.Category(); got != collector.CategoryObjectStorage {
		t.Errorf("Category() = %v, want %v", got, collector.CategoryObjectStorage)
	}
}

func TestCollector_Type(t *testing.T) {
	cfg := &config.ConnectorConfig{
		Type:     SourceName,
		Endpoint: "localhost:9000",
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
		name         string
		endpoint     string
		wantHost     string
		wantSecure   bool
		wantErr      bool
	}{
		{
			name:     "empty endpoint",
			endpoint: "",
			wantErr:  true,
		},
		{
			name:       "http endpoint with port",
			endpoint:   "http://localhost:9000",
			wantHost:   "localhost:9000",
			wantSecure: false,
			wantErr:    false,
		},
		{
			name:       "https endpoint with port",
			endpoint:   "https://s3.amazonaws.com:443",
			wantHost:   "s3.amazonaws.com:443",
			wantSecure: true,
			wantErr:    false,
		},
		{
			name:       "host without scheme",
			endpoint:   "localhost:9000",
			wantHost:   "localhost:9000",
			wantSecure: false,
			wantErr:    false,
		},
		{
			name:       "host without port (http)",
			endpoint:   "http://localhost",
			wantHost:   "localhost:9000",
			wantSecure: false,
			wantErr:    false,
		},
		{
			name:       "host without port (https)",
			endpoint:   "https://s3.amazonaws.com",
			wantHost:   "s3.amazonaws.com:443",
			wantSecure: true,
			wantErr:    false,
		},
		{
			name:     "invalid URL",
			endpoint: "://invalid",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &config.ConnectorConfig{
				Type:     SourceName,
				Endpoint: tt.endpoint,
			}
			
			c, err := NewCollector(cfg)
			if err != nil {
				t.Fatalf("NewCollector() error = %v", err)
			}
			
			collector := c.(*Collector)
			host, secure, err := collector.parseEndpoint()
			
			if (err != nil) != tt.wantErr {
				t.Errorf("parseEndpoint() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			
			if !tt.wantErr {
				if host != tt.wantHost {
					t.Errorf("parseEndpoint() host = %v, want %v", host, tt.wantHost)
				}
				if secure != tt.wantSecure {
					t.Errorf("parseEndpoint() secure = %v, want %v", secure, tt.wantSecure)
				}
			}
		})
	}
}

func TestCollector_isStructuredFile(t *testing.T) {
	cfg := &config.ConnectorConfig{
		Type:     SourceName,
		Endpoint: "localhost:9000",
	}
	
	c, err := NewCollector(cfg)
	if err != nil {
		t.Fatalf("NewCollector() error = %v", err)
	}
	
	collector := c.(*Collector)
	
	tests := []struct {
		name string
		key  string
		want bool
	}{
		{
			name: "CSV file",
			key:  "data/users.csv",
			want: true,
		},
		{
			name: "JSON file",
			key:  "data/config.json",
			want: true,
		},
		{
			name: "Parquet file",
			key:  "data/events.parquet",
			want: true,
		},
		{
			name: "CSV file uppercase",
			key:  "data/USERS.CSV",
			want: true,
		},
		{
			name: "Text file",
			key:  "data/readme.txt",
			want: false,
		},
		{
			name: "Binary file",
			key:  "data/image.jpg",
			want: false,
		},
		{
			name: "No extension",
			key:  "data/file",
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := collector.isStructuredFile(tt.key); got != tt.want {
				t.Errorf("isStructuredFile() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCollector_HealthCheck_NotConnected(t *testing.T) {
	cfg := &config.ConnectorConfig{
		Type:     SourceName,
		Endpoint: "localhost:9000",
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

func TestCollector_DiscoverCatalogs_NotConnected(t *testing.T) {
	cfg := &config.ConnectorConfig{
		Type:     SourceName,
		Endpoint: "localhost:9000",
	}
	
	c, err := NewCollector(cfg)
	if err != nil {
		t.Fatalf("NewCollector() error = %v", err)
	}
	
	ctx := context.Background()
	_, err = c.DiscoverCatalogs(ctx)
	if err == nil {
		t.Error("DiscoverCatalogs() should return error when not connected")
	}
	
	// Check that it's a connection closed error
	if collErr, ok := err.(*collector.CollectorError); ok {
		if collErr.Code != collector.ErrCodeConnectionClosed {
			t.Errorf("DiscoverCatalogs() error code = %v, want %v", collErr.Code, collector.ErrCodeConnectionClosed)
		}
	} else {
		t.Errorf("DiscoverCatalogs() error type = %T, want *collector.CollectorError", err)
	}
}

func TestCollector_filterSchemas(t *testing.T) {
	tests := []struct {
		name     string
		config   *config.ConnectorConfig
		schemas  []string
		expected []string
	}{
		{
			name: "no filtering",
			config: &config.ConnectorConfig{
				Type:     SourceName,
				Endpoint: "localhost:9000",
			},
			schemas:  []string{"bucket1", "bucket2", "test-bucket"},
			expected: []string{"bucket1", "bucket2", "test-bucket"},
		},
		{
			name: "include filter",
			config: &config.ConnectorConfig{
				Type:     SourceName,
				Endpoint: "localhost:9000",
				Matching: &config.MatchingConfig{
					PatternType:   "glob",
					CaseSensitive: false,
					Schemas: &config.MatchingRule{
						Include: []string{"test-*"},
					},
				},
			},
			schemas:  []string{"bucket1", "bucket2", "test-bucket"},
			expected: []string{"test-bucket"},
		},
		{
			name: "exclude filter",
			config: &config.ConnectorConfig{
				Type:     SourceName,
				Endpoint: "localhost:9000",
				Matching: &config.MatchingConfig{
					PatternType:   "glob",
					CaseSensitive: false,
					Schemas: &config.MatchingRule{
						Exclude: []string{"bucket*"},
					},
				},
			},
			schemas:  []string{"bucket1", "bucket2", "test-bucket"},
			expected: []string{"test-bucket"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, err := NewCollector(tt.config)
			if err != nil {
				t.Fatalf("NewCollector() error = %v", err)
			}
			
			collector := c.(*Collector)
			result := collector.filterSchemas(tt.schemas)
			
			if len(result) != len(tt.expected) {
				t.Errorf("filterSchemas() length = %v, want %v", len(result), len(tt.expected))
				return
			}
			
			for i, expected := range tt.expected {
				if result[i] != expected {
					t.Errorf("filterSchemas()[%d] = %v, want %v", i, result[i], expected)
				}
			}
		})
	}
}

func TestCollector_Close(t *testing.T) {
	cfg := &config.ConnectorConfig{
		Type:     SourceName,
		Endpoint: "localhost:9000",
	}
	
	c, err := NewCollector(cfg)
	if err != nil {
		t.Fatalf("NewCollector() error = %v", err)
	}
	
	// Close should not return error even when not connected
	if err := c.Close(); err != nil {
		t.Errorf("Close() error = %v", err)
	}
}

// Test context cancellation
func TestCollector_ContextCancellation(t *testing.T) {
	cfg := &config.ConnectorConfig{
		Type:     SourceName,
		Endpoint: "localhost:9000",
	}
	
	c, err := NewCollector(cfg)
	if err != nil {
		t.Fatalf("NewCollector() error = %v", err)
	}
	
	// Test with cancelled context
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately
	
	// All operations should return context cancelled error when not connected
	_, err = c.DiscoverCatalogs(ctx)
	if err == nil {
		t.Error("DiscoverCatalogs() should return error with cancelled context")
	}
	
	_, err = c.ListSchemas(ctx, "test")
	if err == nil {
		t.Error("ListSchemas() should return error with cancelled context")
	}
}

// Test timeout context
func TestCollector_ContextTimeout(t *testing.T) {
	cfg := &config.ConnectorConfig{
		Type:     SourceName,
		Endpoint: "localhost:9000",
	}
	
	c, err := NewCollector(cfg)
	if err != nil {
		t.Fatalf("NewCollector() error = %v", err)
	}
	
	// Test with timeout context
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
	defer cancel()
	
	// Wait for timeout
	time.Sleep(2 * time.Millisecond)
	
	// Operations should handle timeout appropriately when not connected
	_, err = c.DiscoverCatalogs(ctx)
	if err == nil {
		t.Error("DiscoverCatalogs() should return error with timed out context")
	}
}