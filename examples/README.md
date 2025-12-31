# Go-Metadata 使用样例

本目录包含了 go-metadata 项目的各种使用样例，展示如何使用元数据采集组件和血缘解析组件。

## 目录结构

```
examples/
├── README.md                    # 本文件
├── metadata-collector/          # 元数据采集组件样例
│   ├── README.md               # 采集器使用说明
│   ├── basic/                  # 基础使用样例
│   │   ├── mysql_basic.go     # MySQL 基础采集
│   │   ├── postgres_basic.go  # PostgreSQL 基础采集（待实现）
│   │   └── mongodb_basic.go   # MongoDB 基础采集（待实现）
│   ├── object-storage/         # 对象存储采集样例
│   │   ├── minio_example.go   # MinIO/S3 采集样例
│   │   └── minio_schema_inference.go # Schema 推断样例（待实现）
│   ├── message-queue/          # 消息队列采集样例
│   │   ├── rabbitmq_example.go # RabbitMQ 采集样例
│   │   └── kafka_example.go   # Kafka 采集样例（待实现）
│   ├── advanced/               # 高级使用样例
│   │   ├── batch_collection.go # 批量采集
│   │   ├── filtered_collection.go # 过滤采集（待实现）
│   │   └── statistics_collection.go # 统计信息采集（待实现）
│   └── config/                 # 配置文件样例
│       ├── mysql_config.yaml  # MySQL 配置样例
│       ├── minio_config.yaml  # MinIO 配置样例
│       └── rabbitmq_config.yaml # RabbitMQ 配置样例
└── lineage-analysis/           # 血缘解析组件样例
    ├── README.md              # 血缘解析使用说明
    ├── basic/                 # 基础使用样例
    │   ├── simple_query.go    # 简单查询血缘分析
    │   ├── join_query.go      # JOIN 查询血缘分析
    │   └── insert_query.go    # INSERT 语句血缘分析（待实现）
    ├── advanced/              # 高级使用样例
    │   ├── complex_query.go   # 复杂查询（CTE、子查询）
    │   ├── window_functions.go # 窗口函数血缘分析（待实现）
    │   └── ddl_parsing.go     # DDL 解析和元数据构建（待实现）
    ├── multi-dialect/         # 多方言支持样例
    │   ├── flink_sql.go       # Flink SQL 血缘分析
    │   ├── spark_sql.go       # Spark SQL 血缘分析（待实现）
    │   └── hive_sql.go        # Hive SQL 血缘分析（待实现）
    ├── integration/           # 集成样例
    │   ├── with_collector.go  # 与元数据采集器集成
    │   └── batch_analysis.go  # 批量血缘分析（待实现）
    └── testdata/              # 测试数据
        ├── sample_schema.json # 样例 Schema
        ├── flink_ddl.sql      # Flink DDL 样例（待实现）
        └── complex_queries.sql # 复杂查询样例（待实现）
```

## 快速开始

### 1. 元数据采集

#### MySQL 基础采集
```bash
cd examples/metadata-collector/basic
go run mysql_basic.go
```

#### MinIO 对象存储采集
```bash
cd examples/metadata-collector/object-storage
go run minio_example.go
```

#### RabbitMQ 消息队列采集
```bash
cd examples/metadata-collector/message-queue
go run rabbitmq_example.go
```

#### 批量采集多个数据源
```bash
cd examples/metadata-collector/advanced
go run batch_collection.go
```

### 2. 血缘解析

#### 简单查询血缘分析
```bash
cd examples/lineage-analysis/basic
go run simple_query.go
```

#### JOIN 查询血缘分析
```bash
cd examples/lineage-analysis/basic
go run join_query.go
```

#### 复杂查询血缘分析
```bash
cd examples/lineage-analysis/advanced
go run complex_query.go
```

#### Flink SQL 血缘分析
```bash
cd examples/lineage-analysis/multi-dialect
go run flink_sql.go
```

#### 与采集器集成
```bash
cd examples/lineage-analysis/integration
go run with_collector.go
```

## 功能特性

### 元数据采集组件

- **多数据源支持**: MySQL, PostgreSQL, MongoDB, MinIO/S3, RabbitMQ, Kafka 等
- **灵活配置**: 支持连接配置、匹配规则、采集选项等
- **Schema 推断**: 对于无 Schema 数据源自动推断结构
- **统计信息**: 采集表和列的统计信息
- **批量采集**: 并发采集多个数据源
- **健康检查**: 监控数据源连接状态

### 血缘解析组件

- **列级血缘**: 追踪每个输出列的来源列和转换操作
- **多 SQL 方言**: 支持 Flink SQL, Spark SQL, Hive, MySQL, PostgreSQL 等
- **复杂查询**: 支持 JOIN, 子查询, CTE, 窗口函数等
- **DDL 解析**: 从 CREATE TABLE/VIEW 语句自动提取表结构
- **批量分析**: 批量分析多个 SQL 语句
- **集成能力**: 与元数据采集器无缝集成

## 配置说明

### 元数据采集器配置

每个采集器都使用统一的配置结构：

```yaml
id: "collector-id"
type: "mysql"  # 采集器类型
category: "RDBMS"  # 数据源类别
endpoint: "localhost:3306"
credentials:
  user: "username"
  password: "password"
properties:
  connection_timeout: 30
matching:
  pattern_type: "glob"
  databases:
    include: ["mydb*"]
    exclude: ["test*"]
collect:
  partitions: true
  indexes: true
  statistics: true
```

### 血缘解析元数据

血缘解析需要元数据目录支持：

```go
// 方式1：手动构建
catalog := metadata.NewMetadataBuilder().
    AddTable("db", "table", []string{"col1", "col2"}).
    BuildCatalog()

// 方式2：从 DDL 解析
analyzer := metadata.NewMetadataBuilder().
    LoadFromDDL(ddlSQL).
    BuildAnalyzer()

// 方式3：从采集器获取
catalog := buildCatalogFromCollector()
```

## 支持的数据源

### 元数据采集

| 类别 | 数据源 | 状态 | 样例文件 |
|------|--------|------|----------|
| **RDBMS** | MySQL | ✅ 已实现 | `basic/mysql_basic.go` |
| **RDBMS** | PostgreSQL | 🚧 待实现 | `basic/postgres_basic.go` |
| **DocumentDB** | MongoDB | 🚧 待实现 | `basic/mongodb_basic.go` |
| **ObjectStorage** | MinIO/S3 | ✅ 已实现 | `object-storage/minio_example.go` |
| **MessageQueue** | RabbitMQ | ✅ 已实现 | `message-queue/rabbitmq_example.go` |
| **MessageQueue** | Kafka | 🚧 待实现 | `message-queue/kafka_example.go` |

### 血缘解析

| SQL 方言 | 状态 | 样例文件 |
|----------|------|----------|
| **Flink SQL** | ✅ 已实现 | `multi-dialect/flink_sql.go` |
| **Spark SQL** | 🚧 待实现 | `multi-dialect/spark_sql.go` |
| **Hive SQL** | 🚧 待实现 | `multi-dialect/hive_sql.go` |
| **MySQL** | ✅ 已实现 | `basic/simple_query.go` |
| **PostgreSQL** | 🚧 待实现 | - |

## 使用场景

### 元数据采集

1. **数据资产盘点**: 发现和cataloging企业数据资产
2. **数据治理**: 建立数据字典和元数据管理
3. **合规审计**: 满足数据合规和隐私保护要求
4. **数据迁移**: 评估和规划数据迁移项目
5. **性能监控**: 监控数据源健康状态和性能

### 血缘解析

1. **数据血缘追踪**: 了解数据的来源和流向
2. **影响分析**: 评估表结构变更的影响范围
3. **数据质量**: 追踪数据转换过程中的质量问题
4. **合规审计**: 满足数据治理和合规要求
5. **ETL 优化**: 优化数据处理流程

## 最佳实践

### 元数据采集

1. **配置管理**: 使用配置文件管理采集器配置
2. **匹配规则**: 合理设置匹配规则避免采集不必要的数据
3. **批量采集**: 对于多数据源使用批量采集提高效率
4. **错误处理**: 妥善处理连接失败和采集错误
5. **监控告警**: 建立采集任务的监控和告警机制

### 血缘解析

1. **元数据管理**: 保持元数据的准确性和及时更新
2. **批量处理**: 对于大量 SQL 使用批量分析提高效率
3. **错误处理**: 妥善处理解析失败的 SQL 语句
4. **性能优化**: 对于复杂查询考虑缓存解析结果
5. **方言选择**: 根据实际使用的 SQL 引擎选择合适的方言

## 环境要求

- Go 1.24+
- 相应的数据源服务（MySQL, MinIO, RabbitMQ 等）

## 运行样例

1. 确保相应的数据源服务正在运行
2. 根据实际环境修改配置文件中的连接信息
3. 运行相应的样例程序

```bash
# 运行 MySQL 采集样例
cd examples/metadata-collector/basic
go run mysql_basic.go

# 运行血缘分析样例
cd examples/lineage-analysis/basic
go run simple_query.go
```

## 贡献指南

欢迎贡献更多的使用样例！请参考现有样例的结构和风格：

1. 在相应目录下创建新的样例文件
2. 添加详细的注释说明
3. 更新相应的 README 文件
4. 确保样例可以正常运行

## 许可证

MIT License