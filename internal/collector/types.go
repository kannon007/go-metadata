// Package collector provides interfaces and implementations for metadata collection
// from various data sources such as MySQL, PostgreSQL, and Hive.
package collector

import "time"

// TableType 表类型
type TableType string

const (
	TableTypeTable            TableType = "TABLE"
	TableTypeView             TableType = "VIEW"
	TableTypeExternalTable    TableType = "EXTERNAL_TABLE"
	TableTypeMaterializedView TableType = "MATERIALIZED_VIEW"
	TableTypeCollection       TableType = "COLLECTION"       // MongoDB
	TableTypeTopic            TableType = "TOPIC"            // Kafka
	TableTypeQueue            TableType = "QUEUE"            // RabbitMQ
	TableTypeBucket           TableType = "BUCKET"           // OSS
	TableTypeKeySpace         TableType = "KEYSPACE"         // Redis
	TableTypeIndex            TableType = "INDEX"            // Elasticsearch
)

// TableMetadata 表元数据
type TableMetadata struct {
	// 数据源信息
	SourceCategory DataSourceCategory `json:"source_category"`
	SourceType     string             `json:"source_type"` // mysql, mongodb, kafka, etc.

	// 基础信息
	Catalog    string    `json:"catalog"`
	Schema     string    `json:"schema"`
	Name       string    `json:"name"`
	Type       TableType `json:"type"`
	Comment    string    `json:"comment,omitempty"`

	// 结构信息
	Columns    []Column        `json:"columns"`
	Partitions []PartitionInfo `json:"partitions,omitempty"`
	Indexes    []Index         `json:"indexes,omitempty"`
	PrimaryKey []string        `json:"primary_key,omitempty"`

	// 存储信息
	Storage *StorageInfo `json:"storage,omitempty"`

	// 统计信息
	Stats *TableStatistics `json:"stats,omitempty"`

	// 扩展属性
	Properties map[string]string `json:"properties,omitempty"`

	// 元信息
	LastRefreshedAt time.Time `json:"last_refreshed_at"`
	InferredSchema  bool      `json:"inferred_schema"` // 是否为推断的 Schema
}

// Column 列定义
type Column struct {
	OrdinalPosition   int            `json:"ordinal_position"`
	Name              string         `json:"name"`
	Type              string         `json:"type"`
	SourceType        string         `json:"source_type"`
	Length            *int           `json:"length,omitempty"`
	Precision         *int           `json:"precision,omitempty"`
	Scale             *int           `json:"scale,omitempty"`
	Nullable          bool           `json:"nullable"`
	Default           *string        `json:"default,omitempty"`
	Comment           string         `json:"comment,omitempty"`
	IsPrimaryKey      bool           `json:"is_primary_key"`
	IsPartitionColumn bool           `json:"is_partition_column"`
	IsAutoIncrement   bool           `json:"is_auto_increment"`
	Raw               map[string]any `json:"raw,omitempty"`
}


// Index 索引定义
type Index struct {
	Name    string   `json:"name"`
	Columns []string `json:"columns"`
	Unique  bool     `json:"unique"`
	Type    string   `json:"type,omitempty"`
	Comment string   `json:"comment,omitempty"`
}

// PartitionInfo 分区信息
type PartitionInfo struct {
	Name        string   `json:"name"`
	Type        string   `json:"type"`
	Columns     []string `json:"columns"`
	Expression  string   `json:"expression,omitempty"`
	ValuesCount int      `json:"values_count,omitempty"`
}

// StorageInfo 存储信息（主要用于 Hive/数据湖）
type StorageInfo struct {
	Format       string `json:"format,omitempty"`
	Location     string `json:"location,omitempty"`
	InputFormat  string `json:"input_format,omitempty"`
	OutputFormat string `json:"output_format,omitempty"`
	SerDe        string `json:"serde,omitempty"`
	Compressed   bool   `json:"compressed"`
}

// TableStatistics 表统计信息
type TableStatistics struct {
	RowCount       int64         `json:"row_count"`
	DataSizeBytes  int64         `json:"data_size_bytes"`
	PartitionCount int           `json:"partition_count,omitempty"`
	ColumnStats    []ColumnStats `json:"column_stats,omitempty"`
	CollectedAt    time.Time     `json:"collected_at"`
}

// ColumnStats 列统计信息
type ColumnStats struct {
	Name          string     `json:"name"`
	DistinctCount *int64     `json:"distinct_count,omitempty"`
	NullCount     *int64     `json:"null_count,omitempty"`
	Min           any        `json:"min,omitempty"`
	Max           any        `json:"max,omitempty"`
	Avg           *float64   `json:"avg,omitempty"`
	TopN          []TopNItem `json:"top_n,omitempty"`
}

// TopNItem TopN 统计项
type TopNItem struct {
	Value any   `json:"value"`
	Count int64 `json:"count"`
}

// CatalogInfo Catalog 信息
type CatalogInfo struct {
	Catalog     string            `json:"catalog"`
	Type        string            `json:"type"`
	Description string            `json:"description,omitempty"`
	Properties  map[string]string `json:"properties,omitempty"`
}

// HealthStatus 健康检查结果
type HealthStatus struct {
	Connected bool          `json:"connected"`
	Latency   time.Duration `json:"latency"`
	Version   string        `json:"version"`
	Message   string        `json:"message,omitempty"`
}

// TableListResult 表列表结果（支持分页）
type TableListResult struct {
	Tables        []string `json:"tables"`
	NextPageToken string   `json:"next_page_token,omitempty"`
	TotalCount    int      `json:"total_count"`
}

// PartialResult 部分结果（用于批量操作中的部分失败处理）
type PartialResult[T any] struct {
	// Results 成功的结果列表
	Results []T `json:"results"`
	// Failures 失败的项目列表
	Failures []FailureItem `json:"failures,omitempty"`
	// TotalCount 总项目数
	TotalCount int `json:"total_count"`
	// SuccessCount 成功数量
	SuccessCount int `json:"success_count"`
	// FailureCount 失败数量
	FailureCount int `json:"failure_count"`
}

// FailureItem 失败项目信息
type FailureItem struct {
	// Item 失败的项目标识（如表名、数据库名等）
	Item string `json:"item"`
	// Error 错误信息
	Error string `json:"error"`
	// ErrorCode 错误码
	ErrorCode string `json:"error_code,omitempty"`
}

// HasFailures 检查是否有失败项
func (p *PartialResult[T]) HasFailures() bool {
	return len(p.Failures) > 0
}

// IsComplete 检查是否全部成功
func (p *PartialResult[T]) IsComplete() bool {
	return len(p.Failures) == 0
}

// AddResult 添加成功结果
func (p *PartialResult[T]) AddResult(result T) {
	p.Results = append(p.Results, result)
	p.SuccessCount++
	p.TotalCount++
}

// AddFailure 添加失败项
func (p *PartialResult[T]) AddFailure(item string, err error) {
	failure := FailureItem{
		Item:  item,
		Error: err.Error(),
	}
	// Try to extract error code if it's a CollectorError
	if collErr, ok := err.(*CollectorError); ok {
		failure.ErrorCode = string(collErr.Code)
	}
	p.Failures = append(p.Failures, failure)
	p.FailureCount++
	p.TotalCount++
}

// NewPartialResult 创建新的部分结果
func NewPartialResult[T any]() *PartialResult[T] {
	return &PartialResult[T]{
		Results:  make([]T, 0),
		Failures: make([]FailureItem, 0),
	}
}
