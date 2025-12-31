package config

import (
	"strings"
	"testing"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
)

// getTestParameters returns the standard test parameters for property tests.
func getTestParameters() *gopter.TestParameters {
	params := gopter.DefaultTestParameters()
	params.MinSuccessfulTests = 100
	return params
}

// TestConfigurationValidation_Property6 tests that invalid configurations return descriptive errors.
// Feature: metadata-collector, Property 6: Configuration Validation
// **Validates: Requirements 3.5, 10.5**
func TestConfigurationValidation_Property6(t *testing.T) {
	properties := gopter.NewProperties(getTestParameters())

	// Property: Missing type field should return error mentioning "type"
	properties.Property("missing type returns error mentioning 'type'", prop.ForAll(
		func(endpoint string, user string) bool {
			cfg := &ConnectorConfig{
				Type:     "", // Missing type
				Endpoint: endpoint,
				Credentials: Credentials{
					User:     user,
					Password: "password",
				},
			}
			err := cfg.Validate()
			if err == nil {
				return false
			}
			errStr := strings.ToLower(err.Error())
			return strings.Contains(errStr, "type")
		},
		gen.AlphaString().SuchThat(func(s string) bool { return len(s) > 0 }),
		gen.AlphaString(),
	))

	// Property: Missing endpoint field should return error mentioning "endpoint"
	properties.Property("missing endpoint returns error mentioning 'endpoint'", prop.ForAll(
		func(connType string, user string) bool {
			cfg := &ConnectorConfig{
				Type:     connType,
				Endpoint: "", // Missing endpoint
				Credentials: Credentials{
					User:     user,
					Password: "password",
				},
			}
			err := cfg.Validate()
			if err == nil {
				return false
			}
			errStr := strings.ToLower(err.Error())
			return strings.Contains(errStr, "endpoint")
		},
		gen.OneConstOf("mysql", "postgres", "hive"),
		gen.AlphaString(),
	))

	// Property: Valid configuration should pass validation
	properties.Property("valid configuration passes validation", prop.ForAll(
		func(connType string, host string, port int, user string) bool {
			if len(host) == 0 {
				return true // Skip empty hosts
			}
			endpoint := host + ":" + intToString(port)
			cfg := &ConnectorConfig{
				Type:     connType,
				Endpoint: endpoint,
				Credentials: Credentials{
					User:     user,
					Password: "password",
				},
			}
			err := cfg.Validate()
			return err == nil
		},
		gen.OneConstOf("mysql", "postgres", "hive"),
		gen.AlphaString().SuchThat(func(s string) bool { return len(s) > 0 }),
		gen.IntRange(1, 65535),
		gen.AlphaString(),
	))

	// Property: Invalid regex pattern should return error mentioning the pattern
	properties.Property("invalid regex pattern returns descriptive error", prop.ForAll(
		func(connType string, host string, port int) bool {
			endpoint := host + ":" + intToString(port)
			cfg := &ConnectorConfig{
				Type:     connType,
				Endpoint: endpoint,
				Matching: &MatchingConfig{
					PatternType: "regex",
					Tables: &MatchingRule{
						Include: []string{"[invalid(regex"}, // Invalid regex
					},
				},
			}
			err := cfg.Validate()
			if err == nil {
				return false
			}
			errStr := strings.ToLower(err.Error())
			// Should mention the invalid pattern or regex
			return strings.Contains(errStr, "regex") || strings.Contains(errStr, "pattern") || strings.Contains(errStr, "invalid")
		},
		gen.OneConstOf("mysql", "postgres", "hive"),
		gen.AlphaString().SuchThat(func(s string) bool { return len(s) > 0 }),
		gen.IntRange(1, 65535),
	))

	// Property: Invalid pattern type should return error
	properties.Property("invalid pattern type returns error", prop.ForAll(
		func(connType string, host string, port int, patternType string) bool {
			if patternType == "glob" || patternType == "regex" || patternType == "" {
				return true // Skip valid pattern types
			}
			endpoint := host + ":" + intToString(port)
			cfg := &ConnectorConfig{
				Type:     connType,
				Endpoint: endpoint,
				Matching: &MatchingConfig{
					PatternType: patternType,
				},
			}
			err := cfg.Validate()
			if err == nil {
				return false
			}
			errStr := strings.ToLower(err.Error())
			return strings.Contains(errStr, "pattern_type") || strings.Contains(errStr, "pattern")
		},
		gen.OneConstOf("mysql", "postgres", "hive"),
		gen.AlphaString().SuchThat(func(s string) bool { return len(s) > 0 }),
		gen.IntRange(1, 65535),
		gen.AlphaString().SuchThat(func(s string) bool {
			lower := strings.ToLower(s)
			return lower != "glob" && lower != "regex" && len(s) > 0
		}),
	))

	// Property: Negative statistics values should return error
	properties.Property("negative statistics values return error", prop.ForAll(
		func(connType string, host string, port int, maxTime int) bool {
			if maxTime >= 0 {
				return true // Skip non-negative values
			}
			endpoint := host + ":" + intToString(port)
			cfg := &ConnectorConfig{
				Type:     connType,
				Endpoint: endpoint,
				Statistics: &StatisticsConfig{
					Enabled:        true,
					MaxTimeSeconds: maxTime,
				},
			}
			err := cfg.Validate()
			if err == nil {
				return false
			}
			errStr := strings.ToLower(err.Error())
			return strings.Contains(errStr, "max_time") || strings.Contains(errStr, "negative")
		},
		gen.OneConstOf("mysql", "postgres", "hive"),
		gen.AlphaString().SuchThat(func(s string) bool { return len(s) > 0 }),
		gen.IntRange(1, 65535),
		gen.IntRange(-1000, -1),
	))

	properties.TestingRun(t)
}

// intToString converts an int to string without importing strconv
func intToString(n int) string {
	if n == 0 {
		return "0"
	}
	negative := n < 0
	if negative {
		n = -n
	}
	var digits []byte
	for n > 0 {
		digits = append([]byte{byte('0' + n%10)}, digits...)
		n /= 10
	}
	if negative {
		digits = append([]byte{'-'}, digits...)
	}
	return string(digits)
}

// TestValidateEndpoint tests endpoint validation for various formats
func TestValidateEndpoint(t *testing.T) {
	tests := []struct {
		name      string
		endpoint  string
		connType  string
		wantError bool
	}{
		{"valid mysql host:port", "localhost:3306", "mysql", false},
		{"valid postgres host:port", "localhost:5432", "postgres", false},
		{"valid hive host:port", "localhost:10000", "hive", false},
		{"valid hive thrift url", "thrift://localhost:10000", "hive", false},
		{"valid hostname only", "localhost", "mysql", false},
		{"empty endpoint", "", "mysql", true},
		{"invalid port format", "localhost:abc", "mysql", true},
		{"empty host", ":3306", "mysql", true},
		{"empty port", "localhost:", "mysql", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateEndpoint(tt.endpoint, tt.connType)
			if (err != nil) != tt.wantError {
				t.Errorf("ValidateEndpoint(%q, %q) error = %v, wantError %v",
					tt.endpoint, tt.connType, err, tt.wantError)
			}
		})
	}
}

// TestValidateMatchingConfig tests matching configuration validation
func TestValidateMatchingConfig(t *testing.T) {
	tests := []struct {
		name      string
		cfg       *MatchingConfig
		wantError bool
	}{
		{
			name: "valid glob pattern",
			cfg: &MatchingConfig{
				PatternType: "glob",
				Tables: &MatchingRule{
					Include: []string{"user_*", "order_*"},
				},
			},
			wantError: false,
		},
		{
			name: "valid regex pattern",
			cfg: &MatchingConfig{
				PatternType: "regex",
				Tables: &MatchingRule{
					Include: []string{"^user_.*$", "^order_\\d+$"},
				},
			},
			wantError: false,
		},
		{
			name: "invalid regex pattern",
			cfg: &MatchingConfig{
				PatternType: "regex",
				Tables: &MatchingRule{
					Include: []string{"[invalid(regex"},
				},
			},
			wantError: true,
		},
		{
			name: "invalid pattern type",
			cfg: &MatchingConfig{
				PatternType: "invalid",
			},
			wantError: true,
		},
		{
			name:      "empty config is valid",
			cfg:       &MatchingConfig{},
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateMatchingConfig(tt.cfg)
			if (err != nil) != tt.wantError {
				t.Errorf("ValidateMatchingConfig() error = %v, wantError %v", err, tt.wantError)
			}
		})
	}
}


// TestValidateCategoryField tests category field validation
func TestValidateCategoryField(t *testing.T) {
	tests := []struct {
		name      string
		cfg       *ConnectorConfig
		wantError bool
		errorMsg  string
	}{
		{
			name: "valid category RDBMS with mysql type",
			cfg: &ConnectorConfig{
				Type:     "mysql",
				Category: "RDBMS",
				Endpoint: "localhost:3306",
			},
			wantError: false,
		},
		{
			name: "valid category DataWarehouse with hive type",
			cfg: &ConnectorConfig{
				Type:     "hive",
				Category: "DataWarehouse",
				Endpoint: "localhost:10000",
			},
			wantError: false,
		},
		{
			name: "empty category is valid (optional field)",
			cfg: &ConnectorConfig{
				Type:     "mysql",
				Category: "",
				Endpoint: "localhost:3306",
			},
			wantError: false,
		},
		{
			name: "invalid category value",
			cfg: &ConnectorConfig{
				Type:     "mysql",
				Category: "InvalidCategory",
				Endpoint: "localhost:3306",
			},
			wantError: true,
			errorMsg:  "category",
		},
		{
			name: "category mismatch with type",
			cfg: &ConnectorConfig{
				Type:     "mysql",
				Category: "DocumentDB", // mysql should be RDBMS
				Endpoint: "localhost:3306",
			},
			wantError: true,
			errorMsg:  "category",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.cfg.Validate()
			if (err != nil) != tt.wantError {
				t.Errorf("Validate() error = %v, wantError %v", err, tt.wantError)
			}
			if tt.wantError && err != nil && tt.errorMsg != "" {
				if !strings.Contains(strings.ToLower(err.Error()), strings.ToLower(tt.errorMsg)) {
					t.Errorf("Validate() error = %v, should contain %q", err, tt.errorMsg)
				}
			}
		})
	}
}

// TestValidateInferConfig tests inference configuration validation
func TestValidateInferConfig(t *testing.T) {
	tests := []struct {
		name      string
		cfg       *ConnectorConfig
		wantError bool
		errorMsg  string
	}{
		{
			name: "valid infer config",
			cfg: &ConnectorConfig{
				Type:     "mongodb",
				Endpoint: "localhost:27017",
				Infer: &InferConfig{
					Enabled:    true,
					SampleSize: 100,
					MaxDepth:   10,
					TypeMerge:  TypeMergeMostCommon,
				},
			},
			wantError: false,
		},
		{
			name: "valid infer config with union strategy",
			cfg: &ConnectorConfig{
				Type:     "mongodb",
				Endpoint: "localhost:27017",
				Infer: &InferConfig{
					Enabled:    true,
					SampleSize: 50,
					MaxDepth:   5,
					TypeMerge:  TypeMergeUnion,
				},
			},
			wantError: false,
		},
		{
			name: "nil infer config is valid",
			cfg: &ConnectorConfig{
				Type:     "mysql",
				Endpoint: "localhost:3306",
				Infer:    nil,
			},
			wantError: false,
		},
		{
			name: "negative sample_size",
			cfg: &ConnectorConfig{
				Type:     "mongodb",
				Endpoint: "localhost:27017",
				Infer: &InferConfig{
					Enabled:    true,
					SampleSize: -1,
				},
			},
			wantError: true,
			errorMsg:  "sample_size",
		},
		{
			name: "negative max_depth",
			cfg: &ConnectorConfig{
				Type:     "mongodb",
				Endpoint: "localhost:27017",
				Infer: &InferConfig{
					Enabled:  true,
					MaxDepth: -5,
				},
			},
			wantError: true,
			errorMsg:  "max_depth",
		},
		{
			name: "invalid type_merge strategy",
			cfg: &ConnectorConfig{
				Type:     "mongodb",
				Endpoint: "localhost:27017",
				Infer: &InferConfig{
					Enabled:   true,
					TypeMerge: "invalid_strategy",
				},
			},
			wantError: true,
			errorMsg:  "type_merge",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.cfg.Validate()
			if (err != nil) != tt.wantError {
				t.Errorf("Validate() error = %v, wantError %v", err, tt.wantError)
			}
			if tt.wantError && err != nil && tt.errorMsg != "" {
				if !strings.Contains(strings.ToLower(err.Error()), strings.ToLower(tt.errorMsg)) {
					t.Errorf("Validate() error = %v, should contain %q", err, tt.errorMsg)
				}
			}
		})
	}
}
