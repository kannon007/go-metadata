// Package config provides configuration types and validation for metadata collectors.
package config

import (
	"go-metadata/internal/collector"
)

// ConnectorConfig 采集器配置
type ConnectorConfig struct {
	ID          string                     `json:"id" yaml:"id"`
	Type        string                     `json:"type" yaml:"type"`         // mysql, postgres, hive, mongodb, kafka, etc.
	Category    collector.DataSourceCategory `json:"category" yaml:"category"` // RDBMS, DocumentDB, KeyValue, MessageQueue, ObjectStorage, DataWarehouse
	Endpoint    string                     `json:"endpoint" yaml:"endpoint"`
	Credentials Credentials                `json:"credentials" yaml:"credentials"`
	Properties  ConnectionProps            `json:"properties" yaml:"properties"`
	Matching    *MatchingConfig            `json:"matching,omitempty" yaml:"matching"`
	Collect     *CollectOptions            `json:"collect,omitempty" yaml:"collect"`
	Statistics  *StatisticsConfig          `json:"statistics,omitempty" yaml:"statistics"`
	Infer       *InferConfig               `json:"infer,omitempty" yaml:"infer"` // Schema inference config for schema-less data sources
}

// Credentials 凭证信息
type Credentials struct {
	User     string `json:"user" yaml:"user"`
	Password string `json:"password" yaml:"password"`
}

// ConnectionProps 连接属性
type ConnectionProps struct {
	ConnectionTimeout int               `json:"connection_timeout" yaml:"connection_timeout"`
	MaxOpenConns      int               `json:"max_open_conns" yaml:"max_open_conns"`
	MaxIdleConns      int               `json:"max_idle_conns" yaml:"max_idle_conns"`
	ConnMaxLifetime   int               `json:"conn_max_lifetime" yaml:"conn_max_lifetime"`
	Extra             map[string]string `json:"extra,omitempty" yaml:"extra"`
}

// MatchingConfig 匹配规则配置
type MatchingConfig struct {
	PatternType   string        `json:"pattern_type" yaml:"pattern_type"` // glob, regex
	CaseSensitive bool          `json:"case_sensitive" yaml:"case_sensitive"`
	Databases     *MatchingRule `json:"databases,omitempty" yaml:"databases"`
	Schemas       *MatchingRule `json:"schemas,omitempty" yaml:"schemas"`
	Tables        *MatchingRule `json:"tables,omitempty" yaml:"tables"`
}

// MatchingRule 匹配规则
type MatchingRule struct {
	Include []string `json:"include,omitempty" yaml:"include"`
	Exclude []string `json:"exclude,omitempty" yaml:"exclude"`
}

// CollectOptions 采集选项
type CollectOptions struct {
	Partitions bool `json:"partitions" yaml:"partitions"`
	Indexes    bool `json:"indexes" yaml:"indexes"`
	Comments   bool `json:"comments" yaml:"comments"`
	Statistics bool `json:"statistics" yaml:"statistics"`
}

// StatisticsConfig 统计配置
type StatisticsConfig struct {
	Enabled        bool             `json:"enabled" yaml:"enabled"`
	Level          string           `json:"level" yaml:"level"` // table, column, full
	MaxTimeSeconds int              `json:"max_time_seconds" yaml:"max_time_seconds"`
	MaxRows        int64            `json:"max_rows" yaml:"max_rows"`
	ColumnStats    *ColumnStatsOpts `json:"column_stats,omitempty" yaml:"column_stats"`
}

// ColumnStatsOpts 列统计选项
type ColumnStatsOpts struct {
	Enabled       bool     `json:"enabled" yaml:"enabled"`
	IncludeTopN   bool     `json:"include_top_n" yaml:"include_top_n"`
	TopNCount     int      `json:"top_n_count" yaml:"top_n_count"`
	IncludeMinMax bool     `json:"include_min_max" yaml:"include_min_max"`
	IncludeAvg    bool     `json:"include_avg" yaml:"include_avg"`
	Columns       []string `json:"columns,omitempty" yaml:"columns"` // empty means all columns
}


// TypeMergeStrategy defines how to merge multiple types for the same field.
type TypeMergeStrategy string

const (
	// TypeMergeUnion keeps all observed types (e.g., "string|int")
	TypeMergeUnion TypeMergeStrategy = "union"
	// TypeMergeMostCommon uses the most frequently observed type
	TypeMergeMostCommon TypeMergeStrategy = "most_common"
)

// InferConfig holds configuration for schema inference.
// Used for schema-less data sources like DocumentDB, KeyValue, and ObjectStorage.
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
