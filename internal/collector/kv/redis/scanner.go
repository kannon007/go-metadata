// Package redis provides Redis key scanning and pattern inference functionality.
package redis

import (
	"context"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"go-metadata/internal/collector"
	"go-metadata/internal/collector/kv"

	"github.com/redis/go-redis/v9"
)

// scanKeyPatterns scans Redis keys and infers patterns
func (c *Collector) scanKeyPatterns(ctx context.Context, database int, pattern string, limit int) ([]kv.KeyPattern, error) {
	if c.client == nil {
		return nil, collector.NewConnectionClosedError(SourceName, "scan_key_patterns")
	}

	// Check context before starting operation
	if err := collector.CheckContext(ctx, SourceName, "scan_key_patterns"); err != nil {
		return nil, err
	}

	// Switch to the specified database if needed
	client := c.client
	if database != c.client.Options().DB {
		opts := *c.client.Options()
		opts.DB = database
		client = redis.NewClient(&opts)
		defer client.Close()
	}

	// Scan keys using SCAN command
	keys, err := c.scanKeys(ctx, client, pattern, limit)
	if err != nil {
		return nil, err
	}

	// Infer patterns from the scanned keys
	patterns := c.inferKeyPatterns(ctx, client, keys)

	return patterns, nil
}

// scanKeys uses Redis SCAN command to retrieve keys
func (c *Collector) scanKeys(ctx context.Context, client *redis.Client, pattern string, limit int) ([]string, error) {
	var keys []string
	var cursor uint64
	count := DefaultScanCount

	for {
		// Check context during iteration
		if err := collector.CheckContext(ctx, SourceName, "scan_keys"); err != nil {
			return nil, err
		}

		// Execute SCAN command
		result, newCursor, err := client.Scan(ctx, cursor, pattern, int64(count)).Result()
		if err != nil {
			if ctx.Err() != nil {
				return nil, collector.WrapContextError(ctx, SourceName, "scan_keys")
			}
			return nil, collector.NewQueryError(SourceName, "scan_keys", err)
		}

		keys = append(keys, result...)

		// Check if we've reached the limit
		if limit > 0 && len(keys) >= limit {
			keys = keys[:limit]
			break
		}

		// Check if scan is complete
		cursor = newCursor
		if cursor == 0 {
			break
		}
	}

	return keys, nil
}

// inferKeyPatterns analyzes keys and infers common patterns
func (c *Collector) inferKeyPatterns(ctx context.Context, client *redis.Client, keys []string) []kv.KeyPattern {
	if len(keys) == 0 {
		return []kv.KeyPattern{}
	}

	// Group keys by inferred patterns
	patternGroups := make(map[string][]string)
	
	for _, key := range keys {
		// Check context during processing
		if ctx.Err() != nil {
			break
		}

		pattern := c.inferPatternFromKey(key)
		patternGroups[pattern] = append(patternGroups[pattern], key)
	}

	// Convert groups to KeyPattern structs
	var patterns []kv.KeyPattern
	for pattern, groupKeys := range patternGroups {
		// Check context during processing
		if ctx.Err() != nil {
			break
		}

		// Get key type from a sample key
		keyType := c.getKeyType(ctx, client, groupKeys[0])

		// Limit sample keys to avoid large responses
		sampleKeys := groupKeys
		if len(sampleKeys) > 10 {
			sampleKeys = sampleKeys[:10]
		}

		keyPattern := kv.KeyPattern{
			Pattern:    pattern,
			KeyType:    keyType,
			Count:      int64(len(groupKeys)),
			SampleKeys: sampleKeys,
		}

		patterns = append(patterns, keyPattern)
	}

	// Sort patterns by count (descending)
	sort.Slice(patterns, func(i, j int) bool {
		return patterns[i].Count > patterns[j].Count
	})

	return patterns
}

// inferPatternFromKey infers a pattern from a single key
func (c *Collector) inferPatternFromKey(key string) string {
	// Common Redis key patterns:
	// 1. user:123 -> user:*
	// 2. session:abc123def -> session:*
	// 3. cache:user:123:profile -> cache:user:*:profile
	// 4. order:2023:01:15:123 -> order:*:*:*:*

	parts := strings.Split(key, ":")
	var patternParts []string

	for _, part := range parts {
		if c.isNumeric(part) {
			patternParts = append(patternParts, "*")
		} else if c.isAlphanumericID(part) {
			patternParts = append(patternParts, "*")
		} else if c.isUUID(part) {
			patternParts = append(patternParts, "*")
		} else if c.isTimestamp(part) {
			patternParts = append(patternParts, "*")
		} else {
			patternParts = append(patternParts, part)
		}
	}

	return strings.Join(patternParts, ":")
}

// isNumeric checks if a string is numeric
func (c *Collector) isNumeric(s string) bool {
	_, err := strconv.ParseInt(s, 10, 64)
	return err == nil
}

// isAlphanumericID checks if a string looks like an alphanumeric ID
func (c *Collector) isAlphanumericID(s string) bool {
	// Match strings that are likely IDs (mix of letters and numbers, length > 6)
	if len(s) < 6 {
		return false
	}
	
	hasLetter := false
	hasDigit := false
	
	for _, r := range s {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') {
			hasLetter = true
		} else if r >= '0' && r <= '9' {
			hasDigit = true
		} else {
			return false // Contains non-alphanumeric characters
		}
	}
	
	return hasLetter && hasDigit
}

// isUUID checks if a string looks like a UUID
func (c *Collector) isUUID(s string) bool {
	// Simple UUID pattern check
	uuidPattern := regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`)
	return uuidPattern.MatchString(s)
}

// isTimestamp checks if a string looks like a timestamp
func (c *Collector) isTimestamp(s string) bool {
	// Check for Unix timestamp (10 digits)
	if len(s) == 10 && c.isNumeric(s) {
		timestamp, _ := strconv.ParseInt(s, 10, 64)
		// Check if it's a reasonable timestamp (between 2000 and 2100)
		return timestamp > 946684800 && timestamp < 4102444800
	}
	
	// Check for millisecond timestamp (13 digits)
	if len(s) == 13 && c.isNumeric(s) {
		timestamp, _ := strconv.ParseInt(s, 10, 64)
		// Check if it's a reasonable millisecond timestamp
		return timestamp > 946684800000 && timestamp < 4102444800000
	}
	
	return false
}

// getKeyType gets the Redis type of a key
func (c *Collector) getKeyType(ctx context.Context, client *redis.Client, key string) string {
	// Check context before operation
	if ctx.Err() != nil {
		return "unknown"
	}

	keyType, err := client.Type(ctx, key).Result()
	if err != nil {
		return "unknown"
	}

	return keyType
}

// countKeysForPattern counts keys matching a pattern
func (c *Collector) countKeysForPattern(ctx context.Context, database int, pattern string) (int64, error) {
	if c.client == nil {
		return 0, collector.NewConnectionClosedError(SourceName, "count_keys_for_pattern")
	}

	// Check context before starting operation
	if err := collector.CheckContext(ctx, SourceName, "count_keys_for_pattern"); err != nil {
		return 0, err
	}

	// Switch to the specified database if needed
	client := c.client
	if database != c.client.Options().DB {
		opts := *c.client.Options()
		opts.DB = database
		client = redis.NewClient(&opts)
		defer client.Close()
	}

	// Use SCAN to count keys (Redis doesn't have a direct count for patterns)
	var count int64
	var cursor uint64
	scanCount := DefaultScanCount

	for {
		// Check context during iteration
		if err := collector.CheckContext(ctx, SourceName, "count_keys_for_pattern"); err != nil {
			return 0, err
		}

		// Execute SCAN command
		result, newCursor, err := client.Scan(ctx, cursor, pattern, int64(scanCount)).Result()
		if err != nil {
			if ctx.Err() != nil {
				return 0, collector.WrapContextError(ctx, SourceName, "count_keys_for_pattern")
			}
			return 0, collector.NewQueryError(SourceName, "count_keys_for_pattern", err)
		}

		count += int64(len(result))

		// Check if scan is complete
		cursor = newCursor
		if cursor == 0 {
			break
		}
	}

	return count, nil
}

// ScanKeyPatterns implements the KeyValueCollector interface
func (c *Collector) ScanKeyPatterns(ctx context.Context, database int, pattern string, limit int) ([]kv.KeyPattern, error) {
	return c.scanKeyPatterns(ctx, database, pattern, limit)
}

// GetKeyTypeDistribution implements the KeyValueCollector interface
func (c *Collector) GetKeyTypeDistribution(ctx context.Context, database int) (map[string]int64, error) {
	if c.client == nil {
		return nil, collector.NewConnectionClosedError(SourceName, "get_key_type_distribution")
	}

	// Check context before starting operation
	if err := collector.CheckContext(ctx, SourceName, "get_key_type_distribution"); err != nil {
		return nil, err
	}

	// Switch to the specified database if needed
	client := c.client
	if database != c.client.Options().DB {
		opts := *c.client.Options()
		opts.DB = database
		client = redis.NewClient(&opts)
		defer client.Close()
	}

	// Scan a sample of keys to determine type distribution
	keys, err := c.scanKeys(ctx, client, "*", 1000) // Sample up to 1000 keys
	if err != nil {
		return nil, err
	}

	// Count types
	typeCount := make(map[string]int64)
	for _, key := range keys {
		// Check context during processing
		if err := collector.CheckContext(ctx, SourceName, "get_key_type_distribution"); err != nil {
			return nil, err
		}

		keyType := c.getKeyType(ctx, client, key)
		typeCount[keyType]++
	}

	return typeCount, nil
}

// GetMemoryUsage implements the KeyValueCollector interface
func (c *Collector) GetMemoryUsage(ctx context.Context) (*kv.MemoryStats, error) {
	if c.client == nil {
		return nil, collector.NewConnectionClosedError(SourceName, "get_memory_usage")
	}

	// Check context before starting operation
	if err := collector.CheckContext(ctx, SourceName, "get_memory_usage"); err != nil {
		return nil, err
	}

	// Get memory info from Redis
	info, err := c.client.Info(ctx, "memory").Result()
	if err != nil {
		if ctx.Err() != nil {
			return nil, collector.WrapContextError(ctx, SourceName, "get_memory_usage")
		}
		return nil, collector.NewQueryError(SourceName, "get_memory_usage", err)
	}

	stats := &kv.MemoryStats{}
	
	// Parse memory info
	lines := strings.Split(info, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.Contains(line, ":") {
			parts := strings.SplitN(line, ":", 2)
			if len(parts) == 2 {
				key := strings.TrimSpace(parts[0])
				value := strings.TrimSpace(parts[1])
				
				switch key {
				case "used_memory":
					if val, err := strconv.ParseInt(value, 10, 64); err == nil {
						stats.UsedMemory = val
					}
				case "used_memory_peak":
					if val, err := strconv.ParseInt(value, 10, 64); err == nil {
						stats.UsedMemoryPeak = val
					}
				case "maxmemory":
					if val, err := strconv.ParseInt(value, 10, 64); err == nil {
						stats.MaxMemory = val
					}
				case "mem_fragmentation_ratio":
					if val, err := strconv.ParseFloat(value, 64); err == nil {
						stats.FragmentRatio = val
					}
				}
			}
		}
	}

	return stats, nil
}

// Ensure Collector implements kv.KeyValueCollector interface
var _ kv.KeyValueCollector = (*Collector)(nil)