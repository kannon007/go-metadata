package config

import (
	"fmt"
	"net/url"
	"regexp"
	"strings"

	"go-metadata/internal/collector"
)

// ValidationError represents a configuration validation error
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("invalid configuration: %s - %s", e.Field, e.Message)
}

// ValidationErrors represents multiple validation errors
type ValidationErrors struct {
	Errors []ValidationError `json:"errors"`
}

func (e *ValidationErrors) Error() string {
	if len(e.Errors) == 0 {
		return "no validation errors"
	}
	if len(e.Errors) == 1 {
		return e.Errors[0].Error()
	}
	var msgs []string
	for _, err := range e.Errors {
		msgs = append(msgs, err.Error())
	}
	return fmt.Sprintf("multiple validation errors: %s", strings.Join(msgs, "; "))
}

// Add adds a validation error
func (e *ValidationErrors) Add(field, message string) {
	e.Errors = append(e.Errors, ValidationError{Field: field, Message: message})
}

// HasErrors returns true if there are validation errors
func (e *ValidationErrors) HasErrors() bool {
	return len(e.Errors) > 0
}

// Validate validates the ConnectorConfig and returns validation errors if any
func (c *ConnectorConfig) Validate() error {
	errs := &ValidationErrors{}

	// Validate required fields
	if strings.TrimSpace(c.Type) == "" {
		errs.Add("type", "type is required")
	}

	if strings.TrimSpace(c.Endpoint) == "" {
		errs.Add("endpoint", "endpoint is required")
	} else if err := ValidateEndpoint(c.Endpoint, c.Type); err != nil {
		errs.Add("endpoint", err.Error())
	}

	// Validate category if provided
	if c.Category != "" {
		if !collector.IsValidCategory(c.Category) {
			errs.Add("category", fmt.Sprintf("invalid category '%s', must be one of: RDBMS, DataWarehouse, DocumentDB, KeyValue, MessageQueue, ObjectStorage", c.Category))
		} else {
			// Validate that type matches category if both are provided
			if strings.TrimSpace(c.Type) != "" {
				expectedCategory := collector.GetCategoryByType(c.Type)
				if expectedCategory != "" && expectedCategory != c.Category {
					errs.Add("category", fmt.Sprintf("category '%s' does not match type '%s' (expected category: '%s')", c.Category, c.Type, expectedCategory))
				}
			}
		}
	}

	// Validate matching config if present
	if c.Matching != nil {
		if err := ValidateMatchingConfig(c.Matching); err != nil {
			if verrs, ok := err.(*ValidationErrors); ok {
				for _, e := range verrs.Errors {
					errs.Add("matching."+e.Field, e.Message)
				}
			} else {
				errs.Add("matching", err.Error())
			}
		}
	}

	// Validate statistics config if present
	if c.Statistics != nil {
		if err := validateStatisticsConfig(c.Statistics); err != nil {
			if verrs, ok := err.(*ValidationErrors); ok {
				for _, e := range verrs.Errors {
					errs.Add("statistics."+e.Field, e.Message)
				}
			} else {
				errs.Add("statistics", err.Error())
			}
		}
	}

	// Validate infer config if present
	if c.Infer != nil {
		if err := validateInferConfig(c.Infer); err != nil {
			if verrs, ok := err.(*ValidationErrors); ok {
				for _, e := range verrs.Errors {
					errs.Add("infer."+e.Field, e.Message)
				}
			} else {
				errs.Add("infer", err.Error())
			}
		}
	}

	if errs.HasErrors() {
		return errs
	}
	return nil
}


// ValidateEndpoint validates the endpoint format based on the connector type
func ValidateEndpoint(endpoint, connectorType string) error {
	if strings.TrimSpace(endpoint) == "" {
		return fmt.Errorf("endpoint cannot be empty")
	}

	switch strings.ToLower(connectorType) {
	case "mysql", "postgres", "postgresql":
		return validateDatabaseEndpoint(endpoint)
	case "hive":
		return validateHiveEndpoint(endpoint)
	default:
		// For unknown types, just check it's not empty (already done above)
		return nil
	}
}

// validateDatabaseEndpoint validates MySQL/PostgreSQL endpoint format
// Accepts formats: host:port, host, or full connection string
func validateDatabaseEndpoint(endpoint string) error {
	// Check if it's a URL-style connection string
	if strings.Contains(endpoint, "://") {
		_, err := url.Parse(endpoint)
		if err != nil {
			return fmt.Errorf("invalid connection string format: %v", err)
		}
		return nil
	}

	// Check host:port format
	if strings.Contains(endpoint, ":") {
		parts := strings.Split(endpoint, ":")
		if len(parts) != 2 {
			return fmt.Errorf("invalid host:port format")
		}
		host := strings.TrimSpace(parts[0])
		port := strings.TrimSpace(parts[1])
		if host == "" {
			return fmt.Errorf("host cannot be empty")
		}
		if port == "" {
			return fmt.Errorf("port cannot be empty")
		}
		// Validate port is numeric
		for _, c := range port {
			if c < '0' || c > '9' {
				return fmt.Errorf("port must be numeric")
			}
		}
		return nil
	}

	// Just a hostname is also valid
	return nil
}

// validateHiveEndpoint validates Hive endpoint format
// Accepts formats: host:port or thrift://host:port
func validateHiveEndpoint(endpoint string) error {
	// Remove thrift:// prefix if present
	cleanEndpoint := endpoint
	if strings.HasPrefix(strings.ToLower(endpoint), "thrift://") {
		cleanEndpoint = endpoint[9:]
	}

	// Check host:port format
	if strings.Contains(cleanEndpoint, ":") {
		parts := strings.Split(cleanEndpoint, ":")
		if len(parts) != 2 {
			return fmt.Errorf("invalid host:port format for Hive endpoint")
		}
		host := strings.TrimSpace(parts[0])
		port := strings.TrimSpace(parts[1])
		if host == "" {
			return fmt.Errorf("host cannot be empty")
		}
		if port == "" {
			return fmt.Errorf("port cannot be empty")
		}
		// Validate port is numeric
		for _, c := range port {
			if c < '0' || c > '9' {
				return fmt.Errorf("port must be numeric")
			}
		}
		return nil
	}

	// Just a hostname is also valid
	return nil
}

// ValidateMatchingConfig validates the matching configuration
func ValidateMatchingConfig(cfg *MatchingConfig) error {
	errs := &ValidationErrors{}

	// Validate pattern type
	if cfg.PatternType != "" {
		pt := strings.ToLower(cfg.PatternType)
		if pt != "glob" && pt != "regex" {
			errs.Add("pattern_type", "pattern_type must be 'glob' or 'regex'")
		}
	}

	// Validate regex patterns if pattern type is regex
	if strings.ToLower(cfg.PatternType) == "regex" {
		if cfg.Databases != nil {
			validateRegexPatterns(cfg.Databases, "databases", errs)
		}
		if cfg.Schemas != nil {
			validateRegexPatterns(cfg.Schemas, "schemas", errs)
		}
		if cfg.Tables != nil {
			validateRegexPatterns(cfg.Tables, "tables", errs)
		}
	}

	if errs.HasErrors() {
		return errs
	}
	return nil
}

// validateRegexPatterns validates that all patterns in a MatchingRule are valid regex
func validateRegexPatterns(rule *MatchingRule, fieldPrefix string, errs *ValidationErrors) {
	for i, pattern := range rule.Include {
		if _, err := regexp.Compile(pattern); err != nil {
			errs.Add(fmt.Sprintf("%s.include[%d]", fieldPrefix, i),
				fmt.Sprintf("invalid regex pattern '%s': %v", pattern, err))
		}
	}
	for i, pattern := range rule.Exclude {
		if _, err := regexp.Compile(pattern); err != nil {
			errs.Add(fmt.Sprintf("%s.exclude[%d]", fieldPrefix, i),
				fmt.Sprintf("invalid regex pattern '%s': %v", pattern, err))
		}
	}
}

// validateStatisticsConfig validates the statistics configuration
func validateStatisticsConfig(cfg *StatisticsConfig) error {
	errs := &ValidationErrors{}

	// Validate level
	if cfg.Level != "" {
		level := strings.ToLower(cfg.Level)
		if level != "table" && level != "column" && level != "full" {
			errs.Add("level", "level must be 'table', 'column', or 'full'")
		}
	}

	// Validate max_time_seconds
	if cfg.MaxTimeSeconds < 0 {
		errs.Add("max_time_seconds", "max_time_seconds cannot be negative")
	}

	// Validate max_rows
	if cfg.MaxRows < 0 {
		errs.Add("max_rows", "max_rows cannot be negative")
	}

	// Validate column stats options
	if cfg.ColumnStats != nil {
		if cfg.ColumnStats.TopNCount < 0 {
			errs.Add("column_stats.top_n_count", "top_n_count cannot be negative")
		}
	}

	if errs.HasErrors() {
		return errs
	}
	return nil
}


// validateInferConfig validates the inference configuration
func validateInferConfig(cfg *InferConfig) error {
	errs := &ValidationErrors{}

	// Validate sample_size
	if cfg.SampleSize < 0 {
		errs.Add("sample_size", "sample_size cannot be negative")
	}

	// Validate max_depth
	if cfg.MaxDepth < 0 {
		errs.Add("max_depth", "max_depth cannot be negative")
	}

	// Validate type_merge strategy
	if cfg.TypeMerge != "" {
		if cfg.TypeMerge != TypeMergeUnion && cfg.TypeMerge != TypeMergeMostCommon {
			errs.Add("type_merge", "type_merge must be 'union' or 'most_common'")
		}
	}

	if errs.HasErrors() {
		return errs
	}
	return nil
}

// ValidateCategory validates that a category is valid
func ValidateCategory(category collector.DataSourceCategory) error {
	if category == "" {
		return nil // Empty category is allowed (optional field)
	}
	if !collector.IsValidCategory(category) {
		return fmt.Errorf("invalid category '%s', must be one of: RDBMS, DataWarehouse, DocumentDB, KeyValue, MessageQueue, ObjectStorage", category)
	}
	return nil
}
