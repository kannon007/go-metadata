// Package infer provides schema inference capabilities for schema-less data sources.
package infer

import (
	"context"
	"fmt"
	"regexp"
	"sort"
	"strings"

	"go-metadata/internal/collector"
)

// KeyPatternInferrer implements schema inference for key-value stores like Redis.
type KeyPatternInferrer struct {
	config *InferConfig
}

// NewKeyPatternInferrer creates a new KeyPatternInferrer with default configuration.
func NewKeyPatternInferrer() *KeyPatternInferrer {
	return &KeyPatternInferrer{
		config: DefaultInferConfig(),
	}
}

// NewKeyPatternInferrerWithConfig creates a new KeyPatternInferrer with the specified configuration.
func NewKeyPatternInferrerWithConfig(config *InferConfig) *KeyPatternInferrer {
	return &KeyPatternInferrer{
		config: config,
	}
}

// SetConfig updates the inference configuration.
func (k *KeyPatternInferrer) SetConfig(config *InferConfig) {
	k.config = config
}

// GetConfig returns the current inference configuration.
func (k *KeyPatternInferrer) GetConfig() *InferConfig {
	return k.config
}

// KeyPattern represents a discovered key pattern.
type KeyPattern struct {
	// Pattern is the inferred pattern (e.g., "user:*", "session:*:data")
	Pattern string
	// Count is the number of keys matching this pattern
	Count int
	// SampleKeys contains example keys that match this pattern
	SampleKeys []string
	// Segments contains the pattern segments
	Segments []string
}

// Infer analyzes key samples and returns inferred column definitions representing key patterns.
// The samples parameter should be []interface{} where each element is a string (key name).
func (k *KeyPatternInferrer) Infer(ctx context.Context, samples []interface{}) ([]collector.Column, error) {
	if !k.config.Enabled {
		return []collector.Column{}, nil
	}

	if len(samples) == 0 {
		return []collector.Column{}, nil
	}

	// Convert samples to keys
	keys := make([]string, 0, len(samples))
	for i, sample := range samples {
		key, ok := sample.(string)
		if !ok {
			return nil, fmt.Errorf("sample %d is not a string, got %T", i, sample)
		}
		keys = append(keys, key)
	}

	// Limit sample size if configured
	if k.config.SampleSize > 0 && len(keys) > k.config.SampleSize {
		keys = keys[:k.config.SampleSize]
	}

	// Discover key patterns
	patterns := k.discoverPatterns(ctx, keys)

	// Convert patterns to columns
	columns := k.patternsToColumns(patterns)

	return columns, nil
}

// discoverPatterns analyzes keys and discovers common patterns.
func (k *KeyPatternInferrer) discoverPatterns(ctx context.Context, keys []string) []*KeyPattern {
	// Group keys by their structure
	patternGroups := make(map[string]*KeyPattern)
	
	for _, key := range keys {
		// Check for context cancellation
		select {
		case <-ctx.Done():
			return nil
		default:
		}

		pattern := k.extractPattern(key)
		
		if existing, exists := patternGroups[pattern]; exists {
			existing.Count++
			// Add sample key if we don't have too many
			if len(existing.SampleKeys) < 5 {
				existing.SampleKeys = append(existing.SampleKeys, key)
			}
		} else {
			segments := k.extractSegments(key)
			patternGroups[pattern] = &KeyPattern{
				Pattern:    pattern,
				Count:      1,
				SampleKeys: []string{key},
				Segments:   segments,
			}
		}
	}

	// Convert map to slice and sort by count (descending)
	patterns := make([]*KeyPattern, 0, len(patternGroups))
	for _, pattern := range patternGroups {
		patterns = append(patterns, pattern)
	}

	sort.Slice(patterns, func(i, j int) bool {
		return patterns[i].Count > patterns[j].Count
	})

	return patterns
}

// extractPattern extracts a pattern from a key by replacing variable parts with wildcards.
func (k *KeyPatternInferrer) extractPattern(key string) string {
	// Common separators in Redis keys
	separators := []string{":", "-", "_", ".", "/"}
	
	// Try each separator
	for _, sep := range separators {
		if strings.Contains(key, sep) {
			parts := strings.Split(key, sep)
			pattern := make([]string, len(parts))
			
			for i, part := range parts {
				if k.isVariablePart(part) {
					pattern[i] = "*"
				} else {
					pattern[i] = part
				}
			}
			
			return strings.Join(pattern, sep)
		}
	}
	
	// No separator found, check if the whole key looks variable
	if k.isVariablePart(key) {
		return "*"
	}
	
	return key
}

// extractSegments extracts the segments from a key.
func (k *KeyPatternInferrer) extractSegments(key string) []string {
	// Common separators in Redis keys
	separators := []string{":", "-", "_", ".", "/"}
	
	// Try each separator
	for _, sep := range separators {
		if strings.Contains(key, sep) {
			return strings.Split(key, sep)
		}
	}
	
	// No separator found, return the key as a single segment
	return []string{key}
}

// isVariablePart determines if a key part looks like a variable (ID, timestamp, etc.).
func (k *KeyPatternInferrer) isVariablePart(part string) bool {
	if len(part) == 0 {
		return false
	}
	
	// Check if it's all digits (likely an ID)
	if matched, _ := regexp.MatchString(`^\d+$`, part); matched {
		return true
	}
	
	// Check if it's a UUID pattern
	if matched, _ := regexp.MatchString(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`, part); matched {
		return true
	}
	
	// Check if it's a timestamp pattern (Unix timestamp)
	if matched, _ := regexp.MatchString(`^\d{10,13}$`, part); matched {
		return true
	}
	
	// Check if it's a hash-like string (long alphanumeric)
	if len(part) > 16 && regexp.MustCompile(`^[0-9a-fA-F]+$`).MatchString(part) {
		return true
	}
	
	// Check if it contains mixed case and numbers (likely generated)
	hasUpper := regexp.MustCompile(`[A-Z]`).MatchString(part)
	hasLower := regexp.MustCompile(`[a-z]`).MatchString(part)
	hasDigit := regexp.MustCompile(`\d`).MatchString(part)
	
	if hasUpper && hasLower && hasDigit && len(part) > 8 {
		return true
	}
	
	return false
}

// patternsToColumns converts key patterns to column definitions.
func (k *KeyPatternInferrer) patternsToColumns(patterns []*KeyPattern) []collector.Column {
	columns := make([]collector.Column, 0, len(patterns))
	
	for i, pattern := range patterns {
		// Create a column representing this key pattern
		column := collector.Column{
			OrdinalPosition: i + 1,
			Name:            fmt.Sprintf("pattern_%d", i+1),
			Type:            "TEXT", // Key patterns are always text
			SourceType:      "key_pattern",
			Nullable:        false, // Patterns always exist
			Comment:         fmt.Sprintf("Key pattern: %s (matches %d keys)", pattern.Pattern, pattern.Count),
		}
		
		columns = append(columns, column)
	}
	
	return columns
}

// InferWithResult returns detailed inference results including pattern metadata.
func (k *KeyPatternInferrer) InferWithResult(ctx context.Context, samples []interface{}) (*KeyPatternResult, error) {
	columns, err := k.Infer(ctx, samples)
	if err != nil {
		return nil, err
	}

	// Convert samples to keys
	keys := make([]string, 0, len(samples))
	for _, sample := range samples {
		if key, ok := sample.(string); ok {
			keys = append(keys, key)
		}
	}

	// Limit sample size if configured
	if k.config.SampleSize > 0 && len(keys) > k.config.SampleSize {
		keys = keys[:k.config.SampleSize]
	}

	// Discover patterns
	patterns := k.discoverPatterns(ctx, keys)

	return &KeyPatternResult{
		Columns:     columns,
		SampleCount: len(keys),
		Patterns:    patterns,
	}, nil
}

// KeyPatternResult holds the result of key pattern inference.
type KeyPatternResult struct {
	// Columns contains the inferred column definitions
	Columns []collector.Column `json:"columns"`
	// SampleCount is the number of keys analyzed
	SampleCount int `json:"sample_count"`
	// Patterns contains the discovered key patterns
	Patterns []*KeyPattern `json:"patterns"`
}

// GetPatternByName returns a pattern by its name/pattern string.
func (r *KeyPatternResult) GetPatternByName(pattern string) *KeyPattern {
	for _, p := range r.Patterns {
		if p.Pattern == pattern {
			return p
		}
	}
	return nil
}

// GetTopPatterns returns the top N patterns by count.
func (r *KeyPatternResult) GetTopPatterns(n int) []*KeyPattern {
	if n <= 0 || n >= len(r.Patterns) {
		return r.Patterns
	}
	return r.Patterns[:n]
}

// GetPatternCoverage returns the percentage of keys covered by the top N patterns.
func (r *KeyPatternResult) GetPatternCoverage(n int) float64 {
	if r.SampleCount == 0 {
		return 0.0
	}
	
	topPatterns := r.GetTopPatterns(n)
	totalCovered := 0
	
	for _, pattern := range topPatterns {
		totalCovered += pattern.Count
	}
	
	return float64(totalCovered) / float64(r.SampleCount) * 100
}