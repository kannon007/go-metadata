package matcher

import (
	"strings"
	"testing"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"

	"go-metadata/internal/collector/config"
)

// getTestParameters returns the standard test parameters for property tests.
func getTestParameters() *gopter.TestParameters {
	params := gopter.DefaultTestParameters()
	params.MinSuccessfulTests = 100
	return params
}

// TestPatternMatchingCorrectness tests Property 5: Pattern Matching Correctness
// *For any* pattern (glob or regex), input string, and case sensitivity setting:
// - Glob patterns should match according to glob semantics (*, ?, [])
// - Regex patterns should match according to Go regex semantics
// - When both include and exclude patterns match, exclude takes precedence
// - Case sensitivity setting should be respected in matching
// Feature: metadata-collector, Property 5: Pattern Matching Correctness
// **Validates: Requirements 3.3, 9.1, 9.2, 9.3, 9.4, 9.5**
func TestPatternMatchingCorrectness(t *testing.T) {
	properties := gopter.NewProperties(getTestParameters())

	// Property 5.1: Glob * matches any string
	properties.Property("Glob * pattern matches any string", prop.ForAll(
		func(value string) bool {
			m, err := NewGlobMatcher("*", true)
			if err != nil {
				t.Logf("Error creating glob matcher: %v", err)
				return false
			}
			return m.Match(value)
		},
		gen.AlphaString(),
	))

	// Property 5.2: Glob ? matches exactly one character
	properties.Property("Glob ? matches single character strings", prop.ForAll(
		func(char rune) bool {
			m, err := NewGlobMatcher("?", true)
			if err != nil {
				t.Logf("Error creating glob matcher: %v", err)
				return false
			}
			singleChar := string(char)
			return m.Match(singleChar)
		},
		gen.Rune(),
	))

	// Property 5.3: Glob prefix* matches strings starting with prefix
	properties.Property("Glob prefix* matches strings starting with prefix", prop.ForAll(
		func(prefix, suffix string) bool {
			// Skip empty prefix to avoid trivial case
			if prefix == "" {
				return true
			}
			m, err := NewGlobMatcher(prefix+"*", true)
			if err != nil {
				t.Logf("Error creating glob matcher: %v", err)
				return false
			}
			value := prefix + suffix
			return m.Match(value)
		},
		gen.AlphaString(),
		gen.AlphaString(),
	))

	// Property 5.4: Case insensitive matching
	properties.Property("Case insensitive glob matches regardless of case", prop.ForAll(
		func(value string) bool {
			if value == "" {
				return true
			}
			// Create case-insensitive matcher with lowercase pattern
			m, err := NewGlobMatcher(strings.ToLower(value), false)
			if err != nil {
				t.Logf("Error creating glob matcher: %v", err)
				return false
			}
			// Should match uppercase version
			return m.Match(strings.ToUpper(value))
		},
		gen.AlphaString(),
	))

	// Property 5.5: Case sensitive matching respects case
	properties.Property("Case sensitive glob respects case differences", prop.ForAll(
		func(value string) bool {
			if value == "" || strings.ToLower(value) == strings.ToUpper(value) {
				return true // Skip if no case difference
			}
			m, err := NewGlobMatcher(strings.ToLower(value), true)
			if err != nil {
				t.Logf("Error creating glob matcher: %v", err)
				return false
			}
			// Should NOT match uppercase version when case sensitive
			return !m.Match(strings.ToUpper(value))
		},
		gen.AlphaString().SuchThat(func(s string) bool {
			return strings.ToLower(s) != strings.ToUpper(s)
		}),
	))

	// Property 5.6: Regex matches according to Go regex semantics
	properties.Property("Regex .* matches any string", prop.ForAll(
		func(value string) bool {
			m, err := NewRegexMatcher(".*", true)
			if err != nil {
				t.Logf("Error creating regex matcher: %v", err)
				return false
			}
			return m.Match(value)
		},
		gen.AlphaString(),
	))

	// Property 5.7: Regex case insensitive matching
	properties.Property("Case insensitive regex matches regardless of case", prop.ForAll(
		func(value string) bool {
			if value == "" {
				return true
			}
			// Create case-insensitive matcher with lowercase pattern
			m, err := NewRegexMatcher("^"+strings.ToLower(value)+"$", false)
			if err != nil {
				t.Logf("Error creating regex matcher: %v", err)
				return false
			}
			// Should match uppercase version
			return m.Match(strings.ToUpper(value))
		},
		gen.AlphaString(),
	))

	// Property 5.8: Exclude takes precedence over include
	properties.Property("Exclude takes precedence over include", prop.ForAll(
		func(value string) bool {
			if value == "" {
				return true
			}
			rule := &config.MatchingRule{
				Include: []string{"*"},  // Include all
				Exclude: []string{value}, // Exclude this specific value
			}
			m, err := NewRuleMatcher(rule, "glob", true)
			if err != nil {
				t.Logf("Error creating rule matcher: %v", err)
				return false
			}
			// Should NOT match because exclude takes precedence
			return !m.Match(value)
		},
		gen.AlphaString().SuchThat(func(s string) bool { return s != "" }),
	))

	// Property 5.9: No rules means match all
	properties.Property("No rules means match all", prop.ForAll(
		func(value string) bool {
			m, err := NewRuleMatcher(nil, "glob", true)
			if err != nil {
				t.Logf("Error creating rule matcher: %v", err)
				return false
			}
			return m.Match(value)
		},
		gen.AlphaString(),
	))

	// Property 5.10: Empty include and exclude means match all
	properties.Property("Empty include and exclude means match all", prop.ForAll(
		func(value string) bool {
			rule := &config.MatchingRule{
				Include: []string{},
				Exclude: []string{},
			}
			m, err := NewRuleMatcher(rule, "glob", true)
			if err != nil {
				t.Logf("Error creating rule matcher: %v", err)
				return false
			}
			return m.Match(value)
		},
		gen.AlphaString(),
	))

	properties.TestingRun(t)
}

// TestGlobCharacterClass tests glob character class matching [...]
func TestGlobCharacterClass(t *testing.T) {
	tests := []struct {
		pattern string
		value   string
		want    bool
	}{
		{"[abc]", "a", true},
		{"[abc]", "b", true},
		{"[abc]", "c", true},
		{"[abc]", "d", false},
		{"[a-z]", "m", true},
		{"[a-z]", "A", false},
		{"[!abc]", "d", true},
		{"[!abc]", "a", false},
		{"[^abc]", "d", true},
		{"[^abc]", "b", false},
		{"test[0-9]", "test5", true},
		{"test[0-9]", "testa", false},
	}

	for _, tt := range tests {
		t.Run(tt.pattern+"_"+tt.value, func(t *testing.T) {
			m, err := NewGlobMatcher(tt.pattern, true)
			if err != nil {
				t.Fatalf("NewGlobMatcher error: %v", err)
			}
			got := m.Match(tt.value)
			if got != tt.want {
				t.Errorf("GlobMatcher.Match(%q) = %v, want %v", tt.value, got, tt.want)
			}
		})
	}
}

// TestRuleMatcherIncludeExclude tests include/exclude logic
func TestRuleMatcherIncludeExclude(t *testing.T) {
	tests := []struct {
		name    string
		include []string
		exclude []string
		value   string
		want    bool
	}{
		{
			name:    "include only - match",
			include: []string{"db_*"},
			exclude: nil,
			value:   "db_users",
			want:    true,
		},
		{
			name:    "include only - no match",
			include: []string{"db_*"},
			exclude: nil,
			value:   "other_table",
			want:    false,
		},
		{
			name:    "exclude only - match",
			include: nil,
			exclude: []string{"temp_*"},
			value:   "users",
			want:    true,
		},
		{
			name:    "exclude only - excluded",
			include: nil,
			exclude: []string{"temp_*"},
			value:   "temp_data",
			want:    false,
		},
		{
			name:    "include and exclude - exclude wins",
			include: []string{"*"},
			exclude: []string{"temp_*"},
			value:   "temp_data",
			want:    false,
		},
		{
			name:    "include and exclude - not excluded",
			include: []string{"db_*"},
			exclude: []string{"db_temp*"},
			value:   "db_users",
			want:    true,
		},
		{
			name:    "multiple includes - any match",
			include: []string{"db_*", "tbl_*"},
			exclude: nil,
			value:   "tbl_orders",
			want:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rule := &config.MatchingRule{
				Include: tt.include,
				Exclude: tt.exclude,
			}
			m, err := NewRuleMatcher(rule, "glob", true)
			if err != nil {
				t.Fatalf("NewRuleMatcher error: %v", err)
			}
			got := m.Match(tt.value)
			if got != tt.want {
				t.Errorf("RuleMatcher.Match(%q) = %v, want %v", tt.value, got, tt.want)
			}
		})
	}
}

// TestRegexMatcher tests regex pattern matching
func TestRegexMatcher(t *testing.T) {
	tests := []struct {
		pattern       string
		caseSensitive bool
		value         string
		want          bool
	}{
		{"^db_.*", true, "db_users", true},
		{"^db_.*", true, "other_db", false},
		{"^DB_.*", true, "db_users", false},
		{"^DB_.*", false, "db_users", true},
		{"user[0-9]+", true, "user123", true},
		{"user[0-9]+", true, "userabc", false},
	}

	for _, tt := range tests {
		t.Run(tt.pattern+"_"+tt.value, func(t *testing.T) {
			m, err := NewRegexMatcher(tt.pattern, tt.caseSensitive)
			if err != nil {
				t.Fatalf("NewRegexMatcher error: %v", err)
			}
			got := m.Match(tt.value)
			if got != tt.want {
				t.Errorf("RegexMatcher.Match(%q) = %v, want %v", tt.value, got, tt.want)
			}
		})
	}
}

// TestInvalidRegex tests that invalid regex patterns return errors
func TestInvalidRegex(t *testing.T) {
	_, err := NewRegexMatcher("[invalid", true)
	if err == nil {
		t.Error("Expected error for invalid regex pattern, got nil")
	}
}
