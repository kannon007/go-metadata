# 元数据采集组件使用样例

本目录包含了各种数据源的元数据采集使用样例，展示如何使用 go-metadata 项目中的采集器组件。

## 目录结构

```
metadata-collector/
├── README.md                    # 本文件
├── basic/                       # 基础使用样例
│   ├── mysql_basic.go          # MySQL 基础采集
│   ├── postgres_basic.go       # PostgreSQL 基础采集
│   └── mongodb_basic.go        # MongoDB 基础采集
├── object-storage/              # 对象存储采集样例
│   ├── minio_example.go        # MinIO/S3 采集样例
│   └── minio_schema_inference.go # MinIO 文件 Schema 推断
├── message-queue/               # 消息队列采集样例
│   ├── rabbitmq_example.go     # RabbitMQ 采集样例
│   └── kafka_example.go        # Kafka 采集样例
├── advanced/                    # 高级使用样例
│   ├── batch_collection.go     # 批量采集
│   ├── filtered_collection.go  # 过滤采集
│   └── statistics_collection.go # 统计信息采集
└── config/                      # 配置文件样例
    ├── mysql_config.yaml       # MySQL 配置样例
    ├── minio_config.yaml       # MinIO 配置样例
    └── rabbitmq_config.yaml    # RabbitMQ 配置样例
```

## 快速开始

### 1. 基础 MySQL 采集

```bash
cd examples/metadata-collector/basic
go run mysql_basic.go
```

### 2. MinIO 对象存储采集

```bash
cd examples/metadata-collector/object-storage
go run minio_example.go
```

### 3. RabbitMQ 消息队列采集

```bash
cd examples/metadata-collector/message-queue
go run rabbitmq_example.go
```

## 配置说明

每个采集器都需要相应的配置文件，配置文件样例位于 `config/` 目录下。

### 通用配置结构

```yaml
id: "my-collector"
type: "mysql"  # 采集器类型
category: "RDBMS"  # 数据源类别
endpoint: "localhost:3306"
credentials:
  user: "root"
  password: "password"
properties:
  connection_timeout: 30
  max_open_conns: 10
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

## 支持的数据源类型

| 类别 | 数据源类型 | 样例文件 |
|------|-----------|----------|
| **RDBMS** | MySQL | `basic/mysql_basic.go` |
| **RDBMS** | PostgreSQL | `basic/postgres_basic.go` |
| **DocumentDB** | MongoDB | `basic/mongodb_basic.go` |
| **ObjectStorage** | MinIO/S3 | `object-storage/minio_example.go` |
| **MessageQueue** | RabbitMQ | `message-queue/rabbitmq_example.go` |
| **MessageQueue** | Kafka | `message-queue/kafka_example.go` |

## 高级功能

- **过滤采集**: 使用匹配规则过滤数据库、表等
- **统计信息**: 采集表和列的统计信息
- **Schema 推断**: 对于无 Schema 数据源自动推断结构
- **批量采集**: 并发采集多个数据源

详细使用方法请参考各个样例文件。