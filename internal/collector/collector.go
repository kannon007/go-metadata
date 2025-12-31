// Package collector provides interfaces and implementations for metadata collection
// from various data sources such as MySQL, PostgreSQL, and Hive.
package collector

import "context"

// Collector 元数据采集器统一接口
type Collector interface {
	// 基础信息
	Category() DataSourceCategory
	Type() string

	// 连接管理
	Connect(ctx context.Context) error
	Close() error
	HealthCheck(ctx context.Context) (*HealthStatus, error)

	// Catalog/Schema 发现
	DiscoverCatalogs(ctx context.Context) ([]CatalogInfo, error)
	ListSchemas(ctx context.Context, catalog string) ([]string, error)

	// 表元数据采集
	ListTables(ctx context.Context, catalog, schema string, opts *ListOptions) (*TableListResult, error)
	FetchTableMetadata(ctx context.Context, catalog, schema, table string) (*TableMetadata, error)

	// 统计信息采集
	FetchTableStatistics(ctx context.Context, catalog, schema, table string) (*TableStatistics, error)

	// 分区信息采集
	FetchPartitions(ctx context.Context, catalog, schema, table string) ([]PartitionInfo, error)
}

// ListOptions 列表查询选项
type ListOptions struct {
	PageToken string
	PageSize  int
	Filter    *MatchingRule
}

// MatchingRule 匹配规则
type MatchingRule struct {
	Include []string `json:"include,omitempty" yaml:"include"`
	Exclude []string `json:"exclude,omitempty" yaml:"exclude"`
}


// Config 采集器配置 (deprecated: use config.ConnectorConfig instead)
// This type is kept for backward compatibility with existing collectors.
// It will be removed when all collectors are migrated to the new interface.
type Config struct {
	Host     string
	Port     int
	User     string
	Password string
	Database string
	Options  map[string]string
}

// TableSchema 表结构定义 (deprecated: use TableMetadata instead)
// This type is kept for backward compatibility with existing collectors.
// It will be removed when all collectors are migrated to the new interface.
type TableSchema struct {
	Database   string   `json:"database"`
	Table      string   `json:"table"`
	TableType  string   `json:"table_type"`
	Columns    []Column `json:"columns"`
	PrimaryKey []string `json:"primary_key,omitempty"`
	Indexes    []Index  `json:"indexes,omitempty"`
	Comment    string   `json:"comment,omitempty"`
}

// LegacyCollector is the old collector interface (deprecated: use Collector instead)
// This interface is kept for backward compatibility with existing collectors.
// It will be removed when all collectors are migrated to the new interface.
type LegacyCollector interface {
	Connect(ctx context.Context) error
	Close() error
	ListDatabases(ctx context.Context) ([]string, error)
	ListTables(ctx context.Context, database string) ([]string, error)
	GetTableSchema(ctx context.Context, database, table string) (*TableSchema, error)
	GetAllSchemas(ctx context.Context) ([]*TableSchema, error)
}
