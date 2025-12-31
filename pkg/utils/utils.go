// Package utils provides common utility functions for the metadata system.
package utils

import (
	"crypto/md5"
	"encoding/hex"
	"strings"
)

// GenerateID generates a unique identifier from the given parts.
// It concatenates the parts with a separator and returns an MD5 hash.
func GenerateID(parts ...string) string {
	combined := strings.Join(parts, ":")
	hash := md5.Sum([]byte(combined))
	return hex.EncodeToString(hash[:])
}

// NormalizeIdentifier normalizes a database identifier (table name, column name, etc.)
// by converting to lowercase and trimming whitespace.
func NormalizeIdentifier(identifier string) string {
	return strings.ToLower(strings.TrimSpace(identifier))
}

// SplitQualifiedName splits a qualified name (e.g., "database.table.column")
// into its parts. Returns the parts as a slice.
func SplitQualifiedName(name string) []string {
	if name == "" {
		return nil
	}
	return strings.Split(name, ".")
}

// JoinQualifiedName joins parts into a qualified name with dot separator.
func JoinQualifiedName(parts ...string) string {
	var nonEmpty []string
	for _, p := range parts {
		if p != "" {
			nonEmpty = append(nonEmpty, p)
		}
	}
	return strings.Join(nonEmpty, ".")
}

// StringSliceContains checks if a string slice contains a specific value.
func StringSliceContains(slice []string, value string) bool {
	for _, s := range slice {
		if s == value {
			return true
		}
	}
	return false
}

// UniqueStrings returns a new slice with duplicate strings removed.
// The order of first occurrence is preserved.
func UniqueStrings(slice []string) []string {
	seen := make(map[string]struct{})
	result := make([]string, 0, len(slice))
	for _, s := range slice {
		if _, ok := seen[s]; !ok {
			seen[s] = struct{}{}
			result = append(result, s)
		}
	}
	return result
}

// CoalesceString returns the first non-empty string from the arguments.
func CoalesceString(values ...string) string {
	for _, v := range values {
		if v != "" {
			return v
		}
	}
	return ""
}

// TruncateString truncates a string to the specified maximum length.
// If the string is longer than maxLen, it is truncated and "..." is appended.
func TruncateString(s string, maxLen int) string {
	if maxLen <= 0 {
		return ""
	}
	if len(s) <= maxLen {
		return s
	}
	if maxLen <= 3 {
		return s[:maxLen]
	}
	return s[:maxLen-3] + "..."
}
