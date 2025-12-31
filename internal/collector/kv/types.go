// Package kv provides types and interfaces for key-value store metadata collection.
package kv

import "context"

// KeyPattern represents a Redis key pattern with type and count information
type KeyPattern struct {
	Pattern     string   `json:"pattern"`
	KeyType     string   `json:"key_type"`      // string, hash, list, set, zset
	Count       int64    `json:"count"`
	SampleKeys  []string `json:"sample_keys,omitempty"`
}

// MemoryStats represents Redis memory usage statistics
type MemoryStats struct {
	UsedMemory      int64   `json:"used_memory"`
	UsedMemoryPeak  int64   `json:"used_memory_peak"`
	MaxMemory       int64   `json:"max_memory"`
	FragmentRatio   float64 `json:"fragment_ratio"`
}

// KeyValueCollector extends the base Collector interface for key-value stores
type KeyValueCollector interface {
	// ScanKeyPatterns scans for key patterns in a database
	ScanKeyPatterns(ctx context.Context, database int, pattern string, limit int) ([]KeyPattern, error)
	
	// GetKeyTypeDistribution returns the distribution of key types
	GetKeyTypeDistribution(ctx context.Context, database int) (map[string]int64, error)
	
	// GetMemoryUsage returns memory usage statistics
	GetMemoryUsage(ctx context.Context) (*MemoryStats, error)
}