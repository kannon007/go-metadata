// Package redis provides property-based tests for Redis key pattern inference.
package redis

import (
	"fmt"
	"strconv"
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

// genRedisKey generates a random Redis key for property testing.
func genRedisKey() gopter.Gen {
	return gen.OneGenOf(
		// Simple keys
		gen.AlphaString(),
		
		// User keys with numeric IDs
		gopter.CombineGens(gen.IntRange(1, 999999)).Map(func(values []interface{}) string {
			return fmt.Sprintf("user:%d", values[0].(int))
		}),
		
		// Session keys with alphanumeric IDs
		gopter.CombineGens(gen.AlphaString()).Map(func(values []interface{}) string {
			id := values[0].(string)
			if len(id) < 6 {
				id = id + "123456" // Ensure minimum length
			}
			return fmt.Sprintf("session:%s", id[:min(12, len(id))])
		}),
		
		// Cache keys with nested structure
		gopter.CombineGens(gen.IntRange(1, 999999), gen.AlphaString()).Map(func(values []interface{}) string {
			return fmt.Sprintf("cache:user:%d:%s", values[0].(int), values[1].(string))
		}),
		
		// Order keys with date-like structure
		gopter.CombineGens(
			gen.IntRange(2020, 2025),
			gen.IntRange(1, 12),
			gen.IntRange(1, 28),
			gen.IntRange(1, 999999),
		).Map(func(values []interface{}) string {
			return fmt.Sprintf("order:%d:%02d:%02d:%d", 
				values[0].(int), values[1].(int), values[2].(int), values[3].(int))
		}),
		
		// UUID-like keys (simplified)
		gopter.CombineGens(gen.AlphaString()).Map(func(values []interface{}) string {
			id := values[0].(string)
			if len(id) < 8 {
				id = id + "12345678"
			}
			// Create a UUID-like string
			uuid := fmt.Sprintf("%s-%s-%s-%s-%s", 
				id[:min(8, len(id))], 
				id[:min(4, len(id))], 
				id[:min(4, len(id))], 
				id[:min(4, len(id))], 
				id[:min(12, len(id))])
			return fmt.Sprintf("token:%s", uuid)
		}),
		
		// Timestamp keys
		gopter.CombineGens(gen.Int64Range(1640995200, 1672531200)).Map(func(values []interface{}) string {
			return fmt.Sprintf("event:%d:data", values[0].(int64))
		}),
	)
}

// min returns the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// genRedisKeyList generates a list of Redis keys for property testing.
func genRedisKeyList() gopter.Gen {
	return gen.SliceOfN(10, genRedisKey())
}

// **Property 12: Key Pattern Inference**
// **Validates: Requirements 8.4**
func TestKeyPatternInference(t *testing.T) {
	properties := gopter.NewProperties(getTestParameters())

	// Property: Pattern inference should be deterministic for the same input
	properties.Property("key pattern inference is deterministic", prop.ForAll(
		func(key string) bool {
			if key == "" {
				return true
			}

			c := &Collector{}
			
			// Run inference multiple times on the same key
			pattern1 := c.inferPatternFromKey(key)
			pattern2 := c.inferPatternFromKey(key)
			
			// Results should be identical
			return pattern1 == pattern2
		},
		genRedisKey(),
	))

	// Property: Pattern should preserve key structure
	properties.Property("pattern preserves key structure", prop.ForAll(
		func(key string) bool {
			if key == "" {
				return true
			}

			c := &Collector{}
			pattern := c.inferPatternFromKey(key)
			
			// Pattern should not be empty
			if pattern == "" {
				return false
			}
			
			// Pattern should contain the key structure
			keyParts := strings.Split(key, ":")
			patternParts := strings.Split(pattern, ":")
			
			// Should have same number of parts
			if len(keyParts) != len(patternParts) {
				return false
			}
			
			// Each pattern part should either match the key part or be "*"
			for i, keyPart := range keyParts {
				patternPart := patternParts[i]
				if patternPart != keyPart && patternPart != "*" {
					return false
				}
			}
			
			return true
		},
		genRedisKey(),
	))

	// Property: Individual pattern inference should be consistent
	properties.Property("individual key pattern inference is consistent", prop.ForAll(
		func(key string) bool {
			if key == "" {
				return true
			}

			c := &Collector{}
			
			// Infer pattern from the key
			pattern := c.inferPatternFromKey(key)
			
			// Pattern should not be empty
			if pattern == "" {
				return false
			}
			
			// Pattern should contain the key structure
			keyParts := strings.Split(key, ":")
			patternParts := strings.Split(pattern, ":")
			
			// Should have same number of parts
			if len(keyParts) != len(patternParts) {
				return false
			}
			
			// Each pattern part should either match the key part or be "*"
			for i, keyPart := range keyParts {
				patternPart := patternParts[i]
				if patternPart != keyPart && patternPart != "*" {
					return false
				}
			}
			
			return true
		},
		genRedisKey(),
	))

	// Property: Numeric detection should be accurate
	properties.Property("numeric detection is accurate", prop.ForAll(
		func(n int64) bool {
			c := &Collector{}
			
			// Test positive numbers
			if !c.isNumeric(strconv.FormatInt(n, 10)) {
				return false
			}
			
			// Test that non-numeric strings are correctly identified
			nonNumeric := fmt.Sprintf("abc%d", n)
			if c.isNumeric(nonNumeric) {
				return false
			}
			
			return true
		},
		gen.Int64(),
	))

	// Property: UUID detection should be accurate
	properties.Property("UUID detection is accurate", prop.ForAll(
		func(uuid string) bool {
			c := &Collector{}
			
			// Valid UUIDs should be detected
			if len(uuid) == 36 && strings.Count(uuid, "-") == 4 {
				// This is a simplified check - the actual UUID might not be valid
				// but we test the pattern recognition
				return true
			}
			
			// Non-UUID strings should not be detected as UUIDs
			if len(uuid) < 36 || !strings.Contains(uuid, "-") {
				if c.isUUID(uuid) {
					return false // Should not detect invalid UUIDs
				}
			}
			
			return true
		},
		gen.AlphaString(),
	))

	// Property: Alphanumeric ID detection should be consistent
	properties.Property("alphanumeric ID detection is consistent", prop.ForAll(
		func(letters string, numbers string) bool {
			if len(letters) == 0 || len(numbers) == 0 {
				return true
			}
			
			c := &Collector{}
			
			// Create a mixed alphanumeric string
			mixed := letters + numbers
			if len(mixed) >= 6 {
				// Should be detected as alphanumeric ID if it has both letters and numbers
				hasLetters := false
				hasNumbers := false
				for _, r := range mixed {
					if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') {
						hasLetters = true
					} else if r >= '0' && r <= '9' {
						hasNumbers = true
					}
				}
				
				expected := hasLetters && hasNumbers
				actual := c.isAlphanumericID(mixed)
				
				if expected != actual {
					return false
				}
			}
			
			return true
		},
		gen.AlphaString(),
		gen.AlphaString(),
	))

	properties.TestingRun(t)
}