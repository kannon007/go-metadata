# Design Document: Metadata Collector Module

## Overview

本设计文档描述元数据采集模块的技术架构和实现方案。该模块采用分层可插拔架构，按数据源类别组织采集器，通过统一接口支持多种数据源的元数据采集。

### 数据源分类

| 类别 | 英文标识 | 数据源示例 | 特点 |
|-----|---------|-----------|------|
| 关系型数据库 | RDBMS | MySQL, PostgreSQL, Oracle, SQL Server | 结构化 Schema，SQL 查询 |
| 数据仓库/MPP | DataWarehouse | Hive, ClickHouse, Doris, StarRocks | 分布式存储，分区表 |
| 文档数据库 | DocumentDB | MongoDB, Elasticsearch | 无固定 Schema，需推断 |
| 键值存储 | KeyValue | Redis, Etcd | Key 模式，数据类型 |
| 消息队列 | MessageQueue | Kafka, RabbitMQ, RocketMQ | Topic/Queue，Schema Registry |
| 对象存储 | ObjectStorage | MinIO, S3, OSS | Bucket，对象前缀，文件格式 |

### 设计原则

- **分类抽象**：按数据源类别定义基础接口，提供类别特定的采集策略
- **统一输出**：所有采集器输出统一的 TableMetadata 模型
- **可插拔架构**：通过工厂模式和注册机制支持动态扩展
- **容错设计**：类型化错误、重试机制、部分失败处理

## Architecture

### 整体架构图

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                           Collector Module                                   │
├─────────────────────────────────────────────────────────────────────────────┤
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐  ┌─────────────────────┐ │
│  │   Factory   │  │   Config    │  │   Matcher   │  │   Schema Inferrer   │ │
│  │  (registry) │  │ (validation)│  │(glob/regex) │  │  (doc/kv/oss)       │ │
│  └──────┬──────┘  └──────┬──────┘  └──────┬──────┘  └──────────┬──────────┘ │
│         │                │                │                     │           │
│  ┌──────▼────────────────▼────────────────▼─────────────────────▼─────────┐ │
│  │                      Collector Interface                                │ │
│  │  Connect | Close | HealthCheck | Category | DiscoverCatalogs           │ │
│  │  ListSchemas | ListTables | FetchTableMetadata | FetchTableStatistics  │ │
│  └──────┬─────────────────────────────────────────────────────────────────┘ │
│         │                                                                    │
│  ┌──────▼──────────────────────────────────────────────────────────────────┐│
│  │                     Category Base Interfaces                             ││
│  ├──────────────┬──────────────┬──────────────┬──────────────┬─────────────┤│
│  │ RDBMSBase    │ WarehouseBase│ DocumentBase │ KeyValueBase │ MQBase      ││
│  │ OSSBase      │              │              │              │             ││
│  └──────────────┴──────────────┴──────────────┴──────────────┴─────────────┘│
│                                                                              │
├──────────────────────────────────────────────────────────────────────────────┤
│                          Collector Implementations                           │
├──────────────────────────────────────────────────────────────────────────────┤
│  ┌─────────────────────────────────────────────────────────────────────────┐│
│  │ RDBMS: MySQL | PostgreSQL | Oracle | SQLServer                          ││
│  └─────────────────────────────────────────────────────────────────────────┘│
│  ┌─────────────────────────────────────────────────────────────────────────┐│
│  │ DataWarehouse: Hive | ClickHouse | Doris | StarRocks                    ││
│  └─────────────────────────────────────────────────────────────────────────┘│
│  ┌─────────────────────────────────────────────────────────────────────────┐│
│  │ DocumentDB: MongoDB | Elasticsearch                                      ││
│  └─────────────────────────────────────────────────────────────────────────┘│
│  ┌─────────────────────────────────────────────────────────────────────────┐│
│  │ KeyValue: Redis                                                          ││
│  └─────────────────────────────────────────────────────────────────────────┘│
│  ┌─────────────────────────────────────────────────────────────────────────┐│
│  │ MessageQueue: Kafka | RabbitMQ                                           ││
│  └─────────────────────────────────────────────────────────────────────────┘│
│  ┌─────────────────────────────────────────────────────────────────────────┐│
│  │ ObjectStorage: MinIO | S3                                                ││
│  └─────────────────────────────────────────────────────────────────────────┘│
├──────────────────────────────────────────────────────────────────────────────┤
│                        Standard Data Models                                  │
│  TableMetadata | Column | TableStatistics | PartitionInfo | StorageInfo     │
└──────────────────────────────────────────────────────────────────────────────┘
```

### 目录结构

```
internal/collector/
├── collector.go              # Collector 接口定义
├── types.go                  # 标准化数据模型
├── category.go               # 数据源类别定义
├── errors.go                 # 类型化错误定义
├── context.go                # Context 取消处理工具
├── batch.go                  # 批量操作与部分失败处理
├── statistics.go             # 统计信息采集与超时控制
│
├── config/
│   ├── config.go             # ConnectorConfig 配置结构
│   └── validation.go         # 配置验证逻辑
│
├── factory/
│   ├── factory.go            # CollectorFactory 工厂实现
│   ├── registry.go           # 采集器类型注册表（按类别）
│   └── imports.go            # 采集器导入触发注册
│
├── matcher/
│   ├── matcher.go            # Matcher 接口
│   ├── glob.go               # Glob 模式匹配
│   └── regex.go              # Regex 模式匹配
│
├── infer/
│   ├── inferrer.go           # Schema 推断接口
│   ├── document.go           # 文档数据库 Schema 推断
│   ├── keyvalue.go           # 键值存储模式推断
│   └── file.go               # 文件格式 Schema 推断 (Parquet/CSV/JSON)
│
├── retry/
│   └── retry.go              # 重试机制实现
│
├── drivers/
│   └── drivers.go            # 统一驱动导入入口
│
├── rdbms/                    # 关系型数据库
│   ├── base.go               # RDBMS 基础实现
│   ├── mysql/
│   │   ├── mysql.go
│   │   ├── queries.go
│   │   └── register.go
│   ├── postgres/
│   │   ├── postgres.go
│   │   ├── queries.go
│   │   └── register.go
│   ├── oracle/
│   │   ├── oracle.go
│   │   ├── queries.go
│   │   └── register.go
│   └── sqlserver/
│       ├── sqlserver.go
│       ├── queries.go
│       └── register.go
│
├── warehouse/                # 数据仓库/MPP
│   ├── base.go               # DataWarehouse 基础实现
│   ├── hive/
│   │   ├── hive.go
│   │   ├── parser.go
│   │   └── register.go
│   ├── clickhouse/
│   │   ├── clickhouse.go
│   │   ├── queries.go
│   │   └── register.go
│   └── doris/
│       ├── doris.go
│       ├── queries.go
│       └── register.go
│
├── docdb/                    # 文档数据库
│   ├── base.go               # DocumentDB 基础实现
│   ├── mongodb/
│   │   ├── mongodb.go
│   │   ├── schema_infer.go
│   │   └── register.go
│   └── elasticsearch/
│       ├── elasticsearch.go
│       └── register.go
│
├── kv/                       # 键值存储
│   ├── base.go               # KeyValue 基础实现
│   └── redis/
│       ├── redis.go
│       ├── scanner.go
│       └── register.go
│
├── mq/                       # 消息队列
│   ├── base.go               # MessageQueue 基础实现
│   ├── kafka/
│   │   ├── kafka.go
│   │   ├── schema.go
│   │   └── register.go
│   └── rabbitmq/
│       ├── rabbitmq.go
│       └── register.go
│
└── oss/                      # 对象存储
    ├── base.go               # ObjectStorage 基础实现
    ├── minio/
    │   ├── minio.go
    │   └── register.go
    └── s3/
        ├── s3.go
        └── register.go
```

## Components and Interfaces

### Data Source Category

```go
// DataSourceCategory 数据源类别
type DataSourceCategory string

const (
    CategoryRDBMS        DataSourceCategory = "RDBMS"         // 关系型数据库
    CategoryDataWarehouse DataSourceCategory = "DataWarehouse" // 数据仓库/MPP
    CategoryDocumentDB   DataSourceCategory = "DocumentDB"    // 文档数据库
    CategoryKeyValue     DataSourceCategory = "KeyValue"      // 键值存储
    CategoryMessageQueue DataSourceCategory = "MessageQueue"  // 消息队列
    CategoryObjectStorage DataSourceCategory = "ObjectStorage" // 对象存储
)

// CategoryInfo 类别信息
type CategoryInfo struct {
    Category    DataSourceCategory `json:"category"`
    DisplayName string             `json:"display_name"`
    Description string             `json:"description"`
    Types       []string           `json:"types"` // 该类别下的数据源类型
}

// GetAllCategories 获取所有类别信息
func GetAllCategories() []CategoryInfo
```

### Collector Interface

```go
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
```

### Category Base Interfaces

```go
// RDBMSCollector 关系型数据库采集器扩展接口
type RDBMSCollector interface {
    Collector
    
    // 获取视图定义
    FetchViewDefinition(ctx context.Context, catalog, schema, view string) (string, error)
    
    // 获取外键关系
    FetchForeignKeys(ctx context.Context, catalog, schema, table string) ([]ForeignKey, error)
    
    // 获取触发器
    FetchTriggers(ctx context.Context, catalog, schema, table string) ([]Trigger, error)
}

// DataWarehouseCollector 数据仓库采集器扩展接口
type DataWarehouseCollector interface {
    Collector
    
    // 获取存储信息
    FetchStorageInfo(ctx context.Context, catalog, schema, table string) (*StorageInfo, error)
    
    // 获取分布键
    FetchDistributionKeys(ctx context.Context, catalog, schema, table string) ([]string, error)
    
    // 获取排序键
    FetchSortKeys(ctx context.Context, catalog, schema, table string) ([]string, error)
}

// DocumentDBCollector 文档数据库采集器扩展接口
type DocumentDBCollector interface {
    Collector
    
    // 设置 Schema 推断配置
    SetInferConfig(config *InferConfig)
    
    // 获取采样文档
    SampleDocuments(ctx context.Context, catalog, collection string, limit int) ([]map[string]interface{}, error)
}

// KeyValueCollector 键值存储采集器扩展接口
type KeyValueCollector interface {
    Collector
    
    // 扫描键模式
    ScanKeyPatterns(ctx context.Context, database int, pattern string, limit int) ([]KeyPattern, error)
    
    // 获取键类型分布
    GetKeyTypeDistribution(ctx context.Context, database int) (map[string]int64, error)
    
    // 获取内存使用
    GetMemoryUsage(ctx context.Context) (*MemoryStats, error)
}

// MessageQueueCollector 消息队列采集器扩展接口
type MessageQueueCollector interface {
    Collector
    
    // 获取消费者组
    ListConsumerGroups(ctx context.Context, topic string) ([]ConsumerGroup, error)
    
    // 获取 Schema（从 Schema Registry）
    FetchSchema(ctx context.Context, topic string) (*MessageSchema, error)
    
    // 获取 Topic 配置
    FetchTopicConfig(ctx context.Context, topic string) (map[string]string, error)
}

// ObjectStorageCollector 对象存储采集器扩展接口
type ObjectStorageCollector interface {
    Collector
    
    // 列出对象前缀（作为 Schema）
    ListPrefixes(ctx context.Context, bucket, prefix string, delimiter string) ([]string, error)
    
    // 推断文件 Schema
    InferFileSchema(ctx context.Context, bucket, key string) ([]Column, error)
    
    // 获取 Bucket 策略
    GetBucketPolicy(ctx context.Context, bucket string) (*BucketPolicy, error)
}
```


### ConnectorConfig

```go
// ConnectorConfig 采集器配置
type ConnectorConfig struct {
    ID          string             `json:"id" yaml:"id"`
    Type        string             `json:"type" yaml:"type"`                 // mysql, postgres, mongodb, kafka, etc.
    Category    DataSourceCategory `json:"category" yaml:"category"`         // RDBMS, DocumentDB, etc.
    Endpoint    string             `json:"endpoint" yaml:"endpoint"`
    Credentials Credentials        `json:"credentials" yaml:"credentials"`
    Properties  ConnectionProps    `json:"properties" yaml:"properties"`
    Matching    *MatchingConfig    `json:"matching,omitempty" yaml:"matching"`
    Collect     *CollectOptions    `json:"collect,omitempty" yaml:"collect"`
    Statistics  *StatisticsConfig  `json:"statistics,omitempty" yaml:"statistics"`
    Infer       *InferConfig       `json:"infer,omitempty" yaml:"infer"`     // 用于无 Schema 数据源
}

// InferConfig Schema 推断配置（用于 DocumentDB, KeyValue, ObjectStorage）
type InferConfig struct {
    Enabled     bool   `json:"enabled" yaml:"enabled"`
    SampleSize  int    `json:"sample_size" yaml:"sample_size"`   // 采样数量
    MaxDepth    int    `json:"max_depth" yaml:"max_depth"`       // 嵌套深度限制
    TypeMerge   string `json:"type_merge" yaml:"type_merge"`     // union, most_common
}
```

### CollectorFactory

```go
// CollectorFactory 采集器工厂（支持按类别管理）
type CollectorFactory struct {
    registry map[DataSourceCategory]map[string]CollectorCreator
    mu       sync.RWMutex
}

// Register 注册采集器类型
func (f *CollectorFactory) Register(category DataSourceCategory, typeName string, creator CollectorCreator) error

// Create 创建采集器实例
func (f *CollectorFactory) Create(config *ConnectorConfig) (Collector, error)

// ListTypes 列出所有已注册的采集器类型
func (f *CollectorFactory) ListTypes() map[DataSourceCategory][]string

// ListByCategory 列出指定类别的采集器类型
func (f *CollectorFactory) ListByCategory(category DataSourceCategory) []string
```

## Data Models

### TableMetadata

```go
// TableMetadata 表元数据（统一模型）
type TableMetadata struct {
    // 基础信息
    SourceCategory  DataSourceCategory `json:"source_category"`
    SourceType      string             `json:"source_type"`      // mysql, mongodb, kafka, etc.
    Catalog         string             `json:"catalog"`
    Schema          string             `json:"schema"`
    Name            string             `json:"name"`
    Type            TableType          `json:"type"`
    Comment         string             `json:"comment,omitempty"`
    
    // 结构信息
    Columns         []Column           `json:"columns"`
    Partitions      []PartitionInfo    `json:"partitions,omitempty"`
    Indexes         []Index            `json:"indexes,omitempty"`
    PrimaryKey      []string           `json:"primary_key,omitempty"`
    
    // 存储信息
    Storage         *StorageInfo       `json:"storage,omitempty"`
    
    // 统计信息
    Stats           *TableStatistics   `json:"stats,omitempty"`
    
    // 扩展属性
    Properties      map[string]string  `json:"properties,omitempty"`
    
    // 元信息
    LastRefreshedAt time.Time          `json:"last_refreshed_at"`
    InferredSchema  bool               `json:"inferred_schema"`  // 是否为推断的 Schema
}

// TableType 表类型
type TableType string

const (
    TableTypeTable           TableType = "TABLE"
    TableTypeView            TableType = "VIEW"
    TableTypeExternalTable   TableType = "EXTERNAL_TABLE"
    TableTypeMaterialized    TableType = "MATERIALIZED_VIEW"
    TableTypeCollection      TableType = "COLLECTION"       // MongoDB
    TableTypeTopic           TableType = "TOPIC"            // Kafka
    TableTypeQueue           TableType = "QUEUE"            // RabbitMQ
    TableTypeBucket          TableType = "BUCKET"           // OSS
    TableTypeKeySpace        TableType = "KEYSPACE"         // Redis
    TableTypeIndex           TableType = "INDEX"            // Elasticsearch
)
```

### Category-Specific Models

```go
// KeyPattern Redis 键模式
type KeyPattern struct {
    Pattern     string `json:"pattern"`
    KeyType     string `json:"key_type"`      // string, hash, list, set, zset
    Count       int64  `json:"count"`
    SampleKeys  []string `json:"sample_keys,omitempty"`
}

// MessageSchema 消息 Schema
type MessageSchema struct {
    Subject     string   `json:"subject"`
    Version     int      `json:"version"`
    SchemaType  string   `json:"schema_type"`  // AVRO, PROTOBUF, JSON
    Schema      string   `json:"schema"`
    Columns     []Column `json:"columns"`      // 解析后的字段
}

// ConsumerGroup 消费者组
type ConsumerGroup struct {
    GroupID     string            `json:"group_id"`
    State       string            `json:"state"`
    Members     int               `json:"members"`
    Lag         map[int32]int64   `json:"lag"`  // partition -> lag
}

// BucketPolicy Bucket 策略
type BucketPolicy struct {
    Bucket      string `json:"bucket"`
    Policy      string `json:"policy"`
    Versioning  bool   `json:"versioning"`
    Encryption  string `json:"encryption,omitempty"`
}

// MemoryStats Redis 内存统计
type MemoryStats struct {
    UsedMemory      int64   `json:"used_memory"`
    UsedMemoryPeak  int64   `json:"used_memory_peak"`
    MaxMemory       int64   `json:"max_memory"`
    FragmentRatio   float64 `json:"fragment_ratio"`
}
```

## Category Implementation Details

### RDBMS (关系型数据库)

支持的数据源：MySQL, PostgreSQL, Oracle, SQL Server

**共同特点**：
- 使用 `database/sql` 标准接口
- 从 `information_schema` 或系统表获取元数据
- 支持连接池
- 结构化 Schema，无需推断

**MySQL 实现要点**：
```go
// 查询来源
// - 数据库列表: information_schema.SCHEMATA
// - 表列表: information_schema.TABLES
// - 列信息: information_schema.COLUMNS
// - 索引信息: information_schema.STATISTICS
// - 主键: information_schema.KEY_COLUMN_USAGE
```

**PostgreSQL 实现要点**：
```go
// 查询来源
// - 数据库列表: pg_database
// - Schema 列表: information_schema.schemata
// - 表列表: information_schema.tables
// - 列信息: information_schema.columns
// - 索引信息: pg_indexes
```

**Oracle 实现要点**：
```go
// 查询来源
// - Schema 列表: ALL_USERS
// - 表列表: ALL_TABLES, ALL_VIEWS
// - 列信息: ALL_TAB_COLUMNS
// - 索引信息: ALL_INDEXES, ALL_IND_COLUMNS
// - 分区信息: ALL_TAB_PARTITIONS
```

**SQL Server 实现要点**：
```go
// 查询来源
// - 数据库列表: sys.databases
// - Schema 列表: sys.schemas
// - 表列表: INFORMATION_SCHEMA.TABLES
// - 列信息: INFORMATION_SCHEMA.COLUMNS
// - 索引信息: sys.indexes, sys.index_columns
```

### DataWarehouse (数据仓库/MPP)

支持的数据源：Hive, ClickHouse, Doris, StarRocks

**共同特点**：
- 分布式存储
- 分区表支持
- 存储格式信息（Parquet, ORC 等）
- 分布键/排序键

**Hive 实现要点**：
```go
// 使用 HiveServer2 Thrift 协议
// - 数据库列表: SHOW DATABASES
// - 表列表: SHOW TABLES
// - 表元数据: DESCRIBE FORMATTED
// - 分区信息: SHOW PARTITIONS
```

**ClickHouse 实现要点**：
```go
// 查询来源
// - 数据库列表: system.databases
// - 表列表: system.tables
// - 列信息: system.columns
// - 分区信息: system.parts
```

**Doris 实现要点**：
```go
// 兼容 MySQL 协议
// - 数据库列表: SHOW DATABASES
// - 表列表: SHOW TABLES
// - 列信息: DESCRIBE table
// - 分区信息: SHOW PARTITIONS FROM table
```

### DocumentDB (文档数据库)

支持的数据源：MongoDB, Elasticsearch

**共同特点**：
- 无固定 Schema
- 需要通过采样推断字段结构
- 支持嵌套文档
- 索引信息

**MongoDB 实现要点**：
```go
// Schema 推断流程
// 1. 使用 $sample 聚合采样文档
// 2. 递归遍历文档字段
// 3. 统计字段类型出现频率
// 4. 选择最常见类型作为字段类型
// 5. 处理嵌套文档（使用点号表示法）
```

**Elasticsearch 实现要点**：
```go
// 元数据来源
// - 索引列表: _cat/indices
// - Mapping: GET /{index}/_mapping
// - 设置: GET /{index}/_settings
// - 别名: GET /_aliases
```

### KeyValue (键值存储)

支持的数据源：Redis

**共同特点**：
- 键模式发现
- 数据类型分布
- 内存使用统计

**Redis 实现要点**：
```go
// 元数据采集
// - 数据库: CONFIG GET databases
// - 键扫描: SCAN cursor MATCH pattern COUNT count
// - 键类型: TYPE key
// - 内存: INFO memory
// - 键空间: INFO keyspace

// 键模式推断
// 1. 使用 SCAN 采样键
// 2. 分析键命名模式（如 user:*, order:*:detail）
// 3. 统计各模式的键数量和类型分布
```

### MessageQueue (消息队列)

支持的数据源：Kafka, RabbitMQ

**共同特点**：
- Topic/Queue 列表
- 分区/队列配置
- Schema Registry 集成（可选）
- 消费者组信息

**Kafka 实现要点**：
```go
// 使用 sarama 库
// - Topic 列表: AdminClient.ListTopics()
// - Topic 配置: AdminClient.DescribeConfigs()
// - 分区信息: AdminClient.DescribeTopics()
// - 消费者组: AdminClient.ListConsumerGroups()
// - Schema: Schema Registry REST API
```

**RabbitMQ 实现要点**：
```go
// 使用 Management HTTP API
// - 队列列表: GET /api/queues
// - Exchange 列表: GET /api/exchanges
// - Binding: GET /api/bindings
// - 消费者: GET /api/consumers
```

### ObjectStorage (对象存储)

支持的数据源：MinIO, S3

**共同特点**：
- Bucket 列表
- 对象前缀作为 Schema
- 文件格式 Schema 推断
- Bucket 策略

**MinIO/S3 实现要点**：
```go
// 使用 minio-go 或 aws-sdk-go
// - Bucket 列表: ListBuckets()
// - 对象列表: ListObjects() with delimiter
// - Bucket 策略: GetBucketPolicy()
// - 文件 Schema: 下载采样文件，解析 Parquet/CSV/JSON
```


## Schema Inference

### Inferrer Interface

```go
// SchemaInferrer Schema 推断器接口
type SchemaInferrer interface {
    // 推断 Schema
    Infer(ctx context.Context, samples []interface{}) ([]Column, error)
    
    // 设置配置
    SetConfig(config *InferConfig)
}

// DocumentInferrer 文档数据库 Schema 推断器
type DocumentInferrer struct {
    config *InferConfig
}

// KeyPatternInferrer 键模式推断器
type KeyPatternInferrer struct {
    config *InferConfig
}

// FileSchemaInferrer 文件 Schema 推断器
type FileSchemaInferrer struct {
    config *InferConfig
}
```

### Document Schema Inference

```go
// InferDocumentSchema 推断文档 Schema
func (d *DocumentInferrer) Infer(ctx context.Context, samples []interface{}) ([]Column, error) {
    fieldTypes := make(map[string]map[string]int) // field -> type -> count
    
    for _, sample := range samples {
        doc, ok := sample.(map[string]interface{})
        if !ok {
            continue
        }
        d.collectFields("", doc, fieldTypes, 0)
    }
    
    return d.fieldsToColumns(fieldTypes), nil
}

// collectFields 递归收集字段类型
func (d *DocumentInferrer) collectFields(prefix string, doc map[string]interface{}, fieldTypes map[string]map[string]int, depth int) {
    if d.config.MaxDepth > 0 && depth >= d.config.MaxDepth {
        return
    }
    
    for key, value := range doc {
        fieldName := key
        if prefix != "" {
            fieldName = prefix + "." + key
        }
        
        typeName := d.getTypeName(value)
        
        if fieldTypes[fieldName] == nil {
            fieldTypes[fieldName] = make(map[string]int)
        }
        fieldTypes[fieldName][typeName]++
        
        // 递归处理嵌套文档
        if nested, ok := value.(map[string]interface{}); ok {
            d.collectFields(fieldName, nested, fieldTypes, depth+1)
        }
    }
}
```

### File Schema Inference

```go
// InferFileSchema 推断文件 Schema
func (f *FileSchemaInferrer) InferFromFile(ctx context.Context, reader io.Reader, format string) ([]Column, error) {
    switch format {
    case "parquet":
        return f.inferParquet(ctx, reader)
    case "csv":
        return f.inferCSV(ctx, reader)
    case "json":
        return f.inferJSON(ctx, reader)
    default:
        return nil, fmt.Errorf("unsupported format: %s", format)
    }
}

// inferParquet 从 Parquet 文件推断 Schema
func (f *FileSchemaInferrer) inferParquet(ctx context.Context, reader io.Reader) ([]Column, error) {
    // 使用 parquet-go 读取 Schema
    // Parquet 文件自带 Schema 信息
}

// inferCSV 从 CSV 文件推断 Schema
func (f *FileSchemaInferrer) inferCSV(ctx context.Context, reader io.Reader) ([]Column, error) {
    // 读取 header 作为列名
    // 采样数据行推断类型
}

// inferJSON 从 JSON 文件推断 Schema
func (f *FileSchemaInferrer) inferJSON(ctx context.Context, reader io.Reader) ([]Column, error) {
    // 解析 JSON 对象
    // 使用 DocumentInferrer 推断
}
```

## Error Handling

### 错误类型定义

```go
// ErrorCode 错误码
type ErrorCode string

const (
    ErrCodeAuthError          ErrorCode = "AUTH_ERROR"
    ErrCodeNetworkError       ErrorCode = "NETWORK_ERROR"
    ErrCodeTimeout            ErrorCode = "TIMEOUT"
    ErrCodeNotFound           ErrorCode = "NOT_FOUND"
    ErrCodeUnsupportedFeature ErrorCode = "UNSUPPORTED_FEATURE"
    ErrCodeInvalidConfig      ErrorCode = "INVALID_CONFIG"
    ErrCodeQueryError         ErrorCode = "QUERY_ERROR"
    ErrCodeParseError         ErrorCode = "PARSE_ERROR"
    ErrCodeConnectionClosed   ErrorCode = "CONNECTION_CLOSED"
    ErrCodePermissionDenied   ErrorCode = "PERMISSION_DENIED"
    ErrCodeCancelled          ErrorCode = "CANCELLED"
    ErrCodeDeadlineExceeded   ErrorCode = "DEADLINE_EXCEEDED"
    ErrCodeInferenceError     ErrorCode = "INFERENCE_ERROR"
)

// CollectorError 采集器错误
type CollectorError struct {
    Code      ErrorCode          `json:"code"`
    Message   string             `json:"message"`
    Category  DataSourceCategory `json:"category"`
    Source    string             `json:"source"`
    Operation string             `json:"operation"`
    Cause     error              `json:"-"`
    Retryable bool               `json:"retryable"`
}
```

## Testing Strategy

### 测试方法

本模块采用双重测试策略：

1. **单元测试**：验证具体示例和边界情况
   - 配置验证测试
   - 模式匹配测试
   - Schema 推断测试
   - 各采集器的 Mock 测试

2. **属性测试**：验证跨所有输入的通用属性
   - 使用 `github.com/leanovate/gopter` 进行属性测试
   - 每个属性测试至少运行 100 次迭代

## Correctness Properties

*A property is a characteristic or behavior that should hold true across all valid executions of a system.*

### Property 1: Category Registration Consistency

*For any* collector type registered with a category, the collector's Category() method should return the same category it was registered under.

**Validates: Requirements 1.3, 2.7**

### Property 2: Factory Extensibility by Category

*For any* new collector type and its creator function, registering it with the factory under a category should succeed, and subsequently creating a collector with that type should return a valid instance with the correct category.

**Validates: Requirements 1.6, 14.1, 14.4**

### Property 3: TableMetadata JSON Round-Trip

*For any* valid TableMetadata object (including sourceCategory field), serializing it to JSON and deserializing back should produce an equivalent object with all fields preserved.

**Validates: Requirements 3.5, 3.7**

### Property 4: Schema Inference Consistency

*For any* set of sample documents from a DocumentDB, running schema inference multiple times should produce consistent Column definitions with the same field names and types.

**Validates: Requirements 7.3, 7.5**

### Property 5: Pattern Matching Correctness

*For any* pattern (glob or regex), input string, and case sensitivity setting:
- Glob patterns should match according to glob semantics
- Regex patterns should match according to Go regex semantics
- When both include and exclude patterns match, exclude takes precedence

**Validates: Requirements 13.1, 13.2, 13.3, 13.4, 13.5**

### Property 6: Configuration Validation

*For any* ConnectorConfig with missing required fields or invalid values, the validation function should return a descriptive error indicating which field is invalid and why.

**Validates: Requirements 4.5, 14.5**

### Property 7: Error Typing Consistency

*For any* error returned by any collector, the error should be of type `*CollectorError` with a valid error code, category, source identifier, and operation name.

**Validates: Requirements 5.6, 6.6, 7.6, 8.6, 9.6, 10.6**

### Property 8: Context Cancellation Handling

*For any* collector operation and any point during execution, when the context is cancelled, the operation should return `context.Canceled` or `context.DeadlineExceeded` error and release resources.

**Validates: Requirements 2.6, 12.5**

### Property 9: Statistics Timeout Behavior

*For any* statistics collection operation with a configured timeout, the operation should complete within the timeout duration and return partial results if the full collection cannot complete in time.

**Validates: Requirements 11.5**

### Property 10: Retry with Exponential Backoff

*For any* retry configuration and transient error sequence, the retry mechanism should:
- Retry up to MaxRetries times
- Wait at least InitialBackoff before first retry
- Increase wait time by Multiplier factor for each subsequent retry
- Not exceed MaxBackoff wait time

**Validates: Requirements 12.2**

### Property 11: Partial Failure Handling

*For any* batch operation where some individual items fail, the operation should continue processing remaining items and return both successful results and a list of failures.

**Validates: Requirements 12.4**

### Property 12: Key Pattern Inference

*For any* set of Redis keys with common prefixes, the key pattern inferrer should identify the common patterns and group keys accordingly.

**Validates: Requirements 8.4**

### Property 13: Message Queue Topic Mapping

*For any* Kafka topic with partitions, the mapping to TableMetadata should preserve partition count in PartitionInfo array, and the TableType should be TOPIC.

**Validates: Requirements 9.2, 9.3, 9.5**

### Property 14: Object Storage Prefix as Schema

*For any* S3/MinIO bucket with object prefixes, listing prefixes with a delimiter should return distinct prefix paths that can be used as schema-like structure.

**Validates: Requirements 10.3**

### Property 15: File Schema Inference

*For any* Parquet file, the inferred schema should match the file's embedded schema exactly. For CSV/JSON files, the inferred types should be consistent with the sampled data.

**Validates: Requirements 10.4**

## New Collector Implementation Guide

### 实现新数据源采集器的步骤

#### Step 1: 确定数据源类别

根据数据源特点选择合适的类别：
- 关系型数据库 → `rdbms/`
- 数据仓库/MPP → `warehouse/`
- 文档数据库 → `docdb/`
- 键值存储 → `kv/`
- 消息队列 → `mq/`
- 对象存储 → `oss/`

#### Step 2: 创建包目录

```
internal/collector/{category}/{datasource}/
├── {datasource}.go      # 主实现文件
├── queries.go           # 查询语句（如适用）
├── register.go          # 工厂注册
└── {datasource}_test.go # 单元测试
```

#### Step 3: 实现 Collector 接口

```go
package newdatasource

import (
    "context"
    "go-metadata/internal/collector"
    "go-metadata/internal/collector/config"
)

type Collector struct {
    config *config.ConnectorConfig
    // 数据源特定字段
}

func New(cfg *config.ConnectorConfig) (collector.Collector, error) {
    return &Collector{config: cfg}, nil
}

func (c *Collector) Category() collector.DataSourceCategory {
    return collector.CategoryXXX // 返回对应类别
}

func (c *Collector) Type() string {
    return "newdatasource"
}

// 实现其他接口方法...
```

#### Step 4: 注册到工厂

```go
// register.go
package newdatasource

import (
    "go-metadata/internal/collector"
    "go-metadata/internal/collector/config"
    "go-metadata/internal/collector/factory"
)

func init() {
    factory.DefaultFactory.Register(
        collector.CategoryXXX,
        "newdatasource",
        func(cfg *config.ConnectorConfig) (collector.Collector, error) {
            return New(cfg)
        },
    )
}
```

#### Step 5: 更新 drivers 包

```go
// internal/collector/drivers/drivers.go
import (
    _ "go-metadata/internal/collector/{category}/newdatasource"
)
```
