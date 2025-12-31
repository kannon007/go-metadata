// Package infer provides schema inference capabilities for schema-less data sources
// such as document databases, key-value stores, and object storage.
package infer

import (
	"context"

	"go-metadata/internal/collector"
)

// TypeMergeStrategy defines how to merge multiple types for the same field.
type TypeMergeStrategy string

const (
	// TypeMergeUnion keeps all observed types (e.g., "string|int")
	TypeMergeUnion TypeMergeStrategy = "union"
	// TypeMergeMostCommon uses the most frequently observed type
	TypeMergeMostCommon TypeMergeStrategy = "most_common"
)

// InferConfig holds configuration for schema inference.
type InferConfig struct {
	// Enabled indicates whether schema inference is enabled
	Enabled bool `json:"enabled" yaml:"enabled"`
	// SampleSize is the number of documents/keys to sample for inference
	SampleSize int `json:"sample_size" yaml:"sample_size"`
	// MaxDepth limits the nesting depth for document inference (0 = unlimited)
	MaxDepth int `json:"max_depth" yaml:"max_depth"`
	// TypeMerge specifies the strategy for merging multiple types
	TypeMerge TypeMergeStrategy `json:"type_merge" yaml:"type_merge"`
}

// DefaultInferConfig returns the default inference configuration.
func DefaultInferConfig() *InferConfig {
	return &InferConfig{
		Enabled:    true,
		SampleSize: 100,
		MaxDepth:   10,
		TypeMerge:  TypeMergeMostCommon,
	}
}

// SchemaInferrer defines the interface for schema inference.
type SchemaInferrer interface {
	// Infer analyzes samples and returns inferred column definitions.
	// The samples parameter accepts different types depending on the inferrer:
	// - DocumentInferrer: []map[string]interface{}
	// - KeyPatternInferrer: []string (keys)
	// - FileSchemaInferrer: io.Reader with format specification
	Infer(ctx context.Context, samples []interface{}) ([]collector.Column, error)

	// SetConfig updates the inference configuration.
	SetConfig(config *InferConfig)

	// GetConfig returns the current inference configuration.
	GetConfig() *InferConfig
}

// InferenceResult holds the result of schema inference with additional metadata.
type InferenceResult struct {
	// Columns contains the inferred column definitions
	Columns []collector.Column `json:"columns"`
	// SampleCount is the number of samples analyzed
	SampleCount int `json:"sample_count"`
	// FieldCoverage maps field names to the percentage of samples containing that field
	FieldCoverage map[string]float64 `json:"field_coverage,omitempty"`
	// TypeDistribution maps field names to type frequency counts
	TypeDistribution map[string]map[string]int `json:"type_distribution,omitempty"`
}

// FieldTypeInfo holds type information for a single field.
type FieldTypeInfo struct {
	// Name is the field name (using dot notation for nested fields)
	Name string
	// Types maps type names to occurrence counts
	Types map[string]int
	// Nullable indicates if the field was absent in some samples
	Nullable bool
	// Depth is the nesting depth of this field
	Depth int
}

// NewFieldTypeInfo creates a new FieldTypeInfo instance.
func NewFieldTypeInfo(name string, depth int) *FieldTypeInfo {
	return &FieldTypeInfo{
		Name:     name,
		Types:    make(map[string]int),
		Nullable: false,
		Depth:    depth,
	}
}

// AddType records an occurrence of a type for this field.
func (f *FieldTypeInfo) AddType(typeName string) {
	f.Types[typeName]++
}

// MostCommonType returns the most frequently observed type.
func (f *FieldTypeInfo) MostCommonType() string {
	var maxCount int
	var maxType string
	for typeName, count := range f.Types {
		if count > maxCount {
			maxCount = count
			maxType = typeName
		}
	}
	return maxType
}

// TotalCount returns the total number of type observations.
func (f *FieldTypeInfo) TotalCount() int {
	total := 0
	for _, count := range f.Types {
		total += count
	}
	return total
}
