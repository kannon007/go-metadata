# Requirements Document

## Introduction

本文档定义元数据采集模块的需求规范。该模块旨在提供一个统一、可扩展的元数据采集框架，支持从多种数据源采集结构化元数据，包括：

1. **关系型数据库**：MySQL、PostgreSQL、Oracle、SQL Server、Doris 等
2. **文档数据库**：MongoDB、Elasticsearch 等
3. **键值存储**：Redis、Etcd 等
4. **消息队列**：Kafka、RabbitMQ、RocketMQ 等
5. **对象存储**：MinIO、S3、OSS 等
6. **大数据数仓**：Hive、ClickHouse、Doris 等

设计采用分层架构，通过统一的 Collector 接口和数据源类型适配器，实现不同类型数据源的元数据采集。

## Glossary

- **Collector**: 元数据采集器，负责从特定数据源采集元数据的组件
- **Connector**: 数据源连接器，负责建立和管理与数据源的连接
- **DataSourceCategory**: 数据源类别，如 RDBMS、DocumentDB、KeyValue、MessageQueue、ObjectStorage、DataWarehouse
- **Catalog**: 数据目录，数据源的顶层命名空间（如数据库实例、Bucket）
- **Schema**: 模式，数据库中的逻辑分组（如 PostgreSQL 的 schema）
- **TableMetadata**: 表/集合/Topic 元数据，统一的元数据模型
- **TableStatistics**: 统计信息，包含行数、数据大小、列基数等
- **CollectionTask**: 采集任务，表示一次元数据采集的执行单元
- **MatchingRule**: 匹配规则，用于过滤要采集的数据库/表
- **SchemaInferrer**: Schema 推断器，用于无模式数据源的字段推断

## Requirements

### Requirement 1: 数据源分类体系

**User Story:** 作为架构师，我希望系统能够按类别管理不同类型的数据源，以便采用合适的采集策略。

#### Acceptance Criteria

1. THE System SHALL define data source categories: RDBMS, DocumentDB, KeyValue, MessageQueue, ObjectStorage, DataWarehouse
2. THE System SHALL provide category-specific base interfaces for each data source type
3. WHEN a new data source is added, THE Developer SHALL implement the appropriate category interface
4. THE System SHALL support category-specific metadata models while maintaining a unified output format
5. THE System SHALL provide category-specific configuration options
6. THE System SHALL allow querying collectors by category

### Requirement 2: 统一采集器接口

**User Story:** 作为开发者，我希望有一个统一的采集器接口，以便新增数据源只需实现该接口即可接入系统。

#### Acceptance Criteria

1. THE Collector_Interface SHALL define methods for connection management (Connect, Close, HealthCheck)
2. THE Collector_Interface SHALL define methods for catalog discovery (DiscoverCatalogs, ListSchemas)
3. THE Collector_Interface SHALL define methods for table metadata retrieval (ListTables, FetchTableMetadata)
4. THE Collector_Interface SHALL define methods for statistics collection (FetchTableStatistics)
5. WHEN a new data source type is added, THE System SHALL allow registration through a factory pattern without modifying existing code
6. THE Collector_Interface SHALL support context-based cancellation for all operations
7. THE Collector_Interface SHALL return the data source category for each collector

### Requirement 3: 标准化元数据模型

**User Story:** 作为数据管理员，我希望不同数据源的元数据采用统一格式，以便进行统一管理和分析。

#### Acceptance Criteria

1. THE TableMetadata model SHALL include catalog, schema, table name, table type, and comment fields
2. THE TableMetadata model SHALL include a columns array with ordinal position, name, type, source type, nullable, default, and comment
3. THE Column model SHALL preserve the original database type in a sourceType field
4. THE TableMetadata model SHALL include optional partitions, indexes, and storage information
5. WHEN metadata is serialized to JSON, THE System SHALL produce output conforming to the standard schema
6. THE TableMetadata model SHALL include a lastRefreshedAt timestamp
7. THE TableMetadata model SHALL include a sourceCategory field indicating the data source category

### Requirement 4: 采集器配置管理

**User Story:** 作为运维人员，我希望能够灵活配置采集器的连接参数和采集选项。

#### Acceptance Criteria

1. THE ConnectorConfig SHALL include connection parameters (endpoint, credentials, timeout)
2. THE ConnectorConfig SHALL support matching rules for database/schema/table filtering
3. THE MatchingRule SHALL support include and exclude patterns using glob or regex
4. THE ConnectorConfig SHALL support collection options (partitions, indexes, comments, statistics)
5. WHEN invalid configuration is provided, THE System SHALL return descriptive validation errors
6. THE ConnectorConfig SHALL support credential encryption for sensitive fields
7. THE ConnectorConfig SHALL include a category field for data source classification

### Requirement 5: 关系型数据库采集器 (RDBMS)

**User Story:** 作为用户，我希望能够从关系型数据库（MySQL、PostgreSQL、Oracle、SQL Server）采集完整的元数据信息。

#### Acceptance Criteria

1. THE RDBMS_Collector SHALL implement a common base for relational databases
2. THE RDBMS_Collector SHALL retrieve database/schema list from system catalogs
3. THE RDBMS_Collector SHALL retrieve table metadata including columns, indexes, primary keys
4. THE RDBMS_Collector SHALL retrieve view definitions
5. THE RDBMS_Collector SHALL support partition information retrieval
6. WHEN a connection error occurs, THE RDBMS_Collector SHALL return a typed error with error code
7. THE System SHALL provide implementations for MySQL, PostgreSQL, Oracle, SQL Server

### Requirement 6: MPP/数据仓库采集器 (DataWarehouse)

**User Story:** 作为用户，我希望能够从 MPP 数据仓库（Hive、ClickHouse、Doris、StarRocks）采集完整的元数据信息。

#### Acceptance Criteria

1. THE DataWarehouse_Collector SHALL implement a common base for analytical databases
2. THE DataWarehouse_Collector SHALL retrieve database list and table metadata
3. THE DataWarehouse_Collector SHALL retrieve partition information for partitioned tables
4. THE DataWarehouse_Collector SHALL retrieve storage format and location information
5. THE DataWarehouse_Collector SHALL support distributed table properties
6. WHEN a connection error occurs, THE DataWarehouse_Collector SHALL return a typed error with error code
7. THE System SHALL provide implementations for Hive, ClickHouse, Doris

### Requirement 7: 文档数据库采集器 (DocumentDB)

**User Story:** 作为用户，我希望能够从文档数据库（MongoDB、Elasticsearch）采集集合和字段元数据信息。

#### Acceptance Criteria

1. THE DocumentDB_Collector SHALL implement a common base for document databases
2. THE DocumentDB_Collector SHALL retrieve database/index list
3. THE DocumentDB_Collector SHALL infer field schema by sampling documents
4. THE DocumentDB_Collector SHALL retrieve index information
5. THE DocumentDB_Collector SHALL support configurable sample size for schema inference
6. WHEN a connection error occurs, THE DocumentDB_Collector SHALL return a typed error with error code
7. THE System SHALL provide implementations for MongoDB, Elasticsearch

### Requirement 8: 键值存储采集器 (KeyValue)

**User Story:** 作为用户，我希望能够从键值存储（Redis）采集数据库和键空间元数据信息。

#### Acceptance Criteria

1. THE KeyValue_Collector SHALL implement a common base for key-value stores
2. THE KeyValue_Collector SHALL retrieve database list (Redis DB index)
3. THE KeyValue_Collector SHALL retrieve key patterns and data types
4. THE KeyValue_Collector SHALL support key sampling for pattern discovery
5. THE KeyValue_Collector SHALL retrieve memory usage statistics
6. WHEN a connection error occurs, THE KeyValue_Collector SHALL return a typed error with error code
7. THE System SHALL provide implementation for Redis

### Requirement 9: 消息队列采集器 (MessageQueue)

**User Story:** 作为用户，我希望能够从消息队列（Kafka、RabbitMQ）采集 Topic/Queue 和 Schema 元数据信息。

#### Acceptance Criteria

1. THE MessageQueue_Collector SHALL implement a common base for message queues
2. THE MessageQueue_Collector SHALL retrieve topic/queue list
3. THE MessageQueue_Collector SHALL retrieve partition/queue configuration
4. THE MessageQueue_Collector SHALL integrate with Schema Registry when available
5. THE MessageQueue_Collector SHALL map topics/queues to TableMetadata model
6. WHEN a connection error occurs, THE MessageQueue_Collector SHALL return a typed error with error code
7. THE System SHALL provide implementations for Kafka, RabbitMQ

### Requirement 10: 对象存储采集器 (ObjectStorage)

**User Story:** 作为用户，我希望能够从对象存储（MinIO、S3）采集 Bucket 和对象元数据信息。

#### Acceptance Criteria

1. THE ObjectStorage_Collector SHALL implement a common base for object storage
2. THE ObjectStorage_Collector SHALL retrieve bucket list
3. THE ObjectStorage_Collector SHALL retrieve object prefixes as schema-like structure
4. THE ObjectStorage_Collector SHALL infer schema from structured files (Parquet, CSV, JSON)
5. THE ObjectStorage_Collector SHALL retrieve bucket policies and configurations
6. WHEN a connection error occurs, THE ObjectStorage_Collector SHALL return a typed error with error code
7. THE System SHALL provide implementations for MinIO, S3-compatible storage

### Requirement 11: 表统计信息采集

**User Story:** 作为数据分析师，我希望能够获取表的统计信息，以便了解数据分布和质量。

#### Acceptance Criteria

1. THE TableStatistics model SHALL include row count and data size
2. THE TableStatistics model SHALL include column-level statistics (distinct count, null count, min, max)
3. WHEN statistics collection is enabled, THE Collector SHALL gather statistics according to configuration
4. THE StatisticsConfig SHALL support performance limits (max time, max rows)
5. WHEN statistics collection exceeds time limit, THE Collector SHALL return partial results with a warning
6. THE Collector SHALL support table-level and column-level statistics granularity

### Requirement 12: 错误处理与重试

**User Story:** 作为系统管理员，我希望采集器能够优雅地处理错误并支持重试机制。

#### Acceptance Criteria

1. THE Collector SHALL define typed errors with error codes (AUTH_ERROR, NETWORK_ERROR, TIMEOUT, NOT_FOUND)
2. WHEN a transient error occurs, THE Collector SHALL support configurable retry with exponential backoff
3. WHEN a permanent error occurs, THE Collector SHALL fail fast and return detailed error information
4. THE Collector SHALL support partial failure handling (continue on individual table errors)
5. WHEN an operation is cancelled via context, THE Collector SHALL clean up resources and return context.Canceled
6. THE Error model SHALL include source, operation, and original error for debugging

### Requirement 13: 匹配规则过滤

**User Story:** 作为用户，我希望能够通过规则过滤要采集的数据库和表，避免采集不需要的元数据。

#### Acceptance Criteria

1. THE MatchingRule SHALL support glob pattern matching (e.g., "db_*", "user_*")
2. THE MatchingRule SHALL support regex pattern matching
3. THE MatchingRule SHALL support both include and exclude lists
4. WHEN both include and exclude match, THE exclude rule SHALL take precedence
5. THE MatchingRule SHALL support case-sensitive and case-insensitive matching
6. WHEN no rules are specified, THE Collector SHALL collect all accessible objects

### Requirement 14: 采集器注册与工厂

**User Story:** 作为开发者，我希望能够通过工厂模式创建采集器实例，支持动态注册新的采集器类型。

#### Acceptance Criteria

1. THE CollectorFactory SHALL maintain a registry of collector types organized by category
2. THE CollectorFactory SHALL create collector instances based on configuration type field
3. WHEN an unknown collector type is requested, THE CollectorFactory SHALL return a descriptive error
4. THE CollectorFactory SHALL support runtime registration of new collector types
5. THE CollectorFactory SHALL validate configuration before creating collector instances
6. THE CollectorFactory SHALL support listing all registered collector types by category

### Requirement 15: 模块目录结构

**User Story:** 作为开发者，我希望有清晰的目录结构，以便理解代码组织和快速定位功能模块。

#### Acceptance Criteria

1. THE collector module SHALL organize code under internal/collector directory
2. THE collector module SHALL have category-specific sub-directories (rdbms/, docdb/, kv/, mq/, oss/, warehouse/)
3. EACH data source collector SHALL have its own sub-package under the category directory
4. THE collector module SHALL have a factory package for collector registration and creation
5. THE collector module SHALL have a config package for configuration types and validation
6. THE collector module SHALL have a matcher package for pattern matching utilities

### Requirement 16: 健康检查与连接池

**User Story:** 作为运维人员，我希望采集器支持健康检查和连接池管理，以确保系统稳定性。

#### Acceptance Criteria

1. THE Collector_Interface SHALL define a HealthCheck method returning connection status
2. THE HealthCheck method SHALL verify connectivity and return latency information
3. THE Collector SHALL support connection pooling with configurable pool size (where applicable)
4. THE Collector SHALL support connection timeout and idle timeout configuration
5. WHEN a connection becomes stale, THE Collector SHALL automatically reconnect
6. THE HealthCheck result SHALL include version information of the connected data source
