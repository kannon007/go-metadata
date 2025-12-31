# 架构设计文档

## 概述

Go Metadata 是一个元数据管理系统，遵循 Go 语言标准项目布局（golang-standards/project-layout），采用分层架构设计，包含三个基础技术组件和上层业务服务。

## 架构图

```
┌─────────────────────────────────────────────────────────────────┐
│                        应用入口 (cmd/)                           │
│  ┌─────────────────────┐    ┌─────────────────────┐            │
│  │   cmd/server/       │    │   cmd/cli/          │            │
│  │   API 服务入口       │    │   CLI 工具入口       │            │
│  └─────────────────────┘    └─────────────────────┘            │
└─────────────────────────────────────────────────────────────────┘
                                    │
                                    ▼
┌─────────────────────────────────────────────────────────────────┐
│                      业务服务层 (internal/service/)              │
│  ┌─────────────────────┐    ┌─────────────────────┐            │
│  │  metadata/service   │    │  lineage/service    │            │
│  │  元数据管理服务       │    │  血缘查询服务        │            │
│  └─────────────────────┘    └─────────────────────┘            │
└─────────────────────────────────────────────────────────────────┘
                                    │
                                    ▼
┌─────────────────────────────────────────────────────────────────┐
│                    基础技术组件 (internal/)                       │
│  ┌───────────────┐  ┌───────────────┐  ┌───────────────┐       │
│  │   lineage/    │  │  collector/   │  │    graph/     │       │
│  │   血缘解析     │  │  元数据采集    │  │   图数据库     │       │
│  └───────────────┘  └───────────────┘  └───────────────┘       │
└─────────────────────────────────────────────────────────────────┘
                                    │
                                    ▼
┌─────────────────────────────────────────────────────────────────┐
│                      数据访问层 (internal/)                       │
│  ┌─────────────────────┐    ┌─────────────────────┐            │
│  │   repository/       │    │   store/            │            │
│  │   数据访问接口        │    │   存储实现           │            │
│  └─────────────────────┘    └─────────────────────┘            │
└─────────────────────────────────────────────────────────────────┘
```

## 核心组件

### 1. 血缘解析组件 (internal/lineage/)

负责从 SQL 语句中提取列级别的数据血缘关系。

**主要功能：**
- SQL 解析：使用 ANTLR4 解析多种 SQL 方言
- 血缘提取：分析 SELECT、INSERT、UPDATE 等语句的数据流向
- 元数据管理：维护表结构信息用于血缘分析

**目录结构：**
```
internal/lineage/
├── lineage.go          # 主入口 Analyzer
├── catalog.go          # Catalog 接口定义
├── types.go            # 数据类型定义
├── parser.go           # SQL 解析器
├── builder.go          # AST 构建器
├── extractor.go        # 血缘提取器
├── grammar/            # ANTLR 语法文件
├── parser/             # ANTLR 生成的解析器代码
├── ast/                # AST 节点定义
├── metadata/           # 元数据管理
└── tests/              # 测试用例
```

**核心接口：**
```go
// Analyzer SQL 血缘分析器
type Analyzer struct {
    catalog Catalog
}

func (a *Analyzer) Analyze(sql string) (*LineageResult, error)

// Catalog 元数据目录接口
type Catalog interface {
    GetTable(database, table string) (*TableSchema, bool)
    GetColumns(database, table string) []string
}
```

### 2. 元数据采集组件 (internal/collector/)

负责从各类数据源采集表结构、字段等元数据。

**支持的数据源：**
- MySQL
- PostgreSQL
- Hive

**目录结构：**
```
internal/collector/
├── collector.go        # 接口定义
├── types.go            # 类型定义
├── errors.go           # 错误定义
├── mysql/              # MySQL 采集器
├── postgres/           # PostgreSQL 采集器
└── hive/               # Hive 采集器
```

**核心接口：**
```go
// Collector 元数据采集器接口
type Collector interface {
    Connect(ctx context.Context) error
    Close() error
    ListDatabases(ctx context.Context) ([]string, error)
    ListTables(ctx context.Context, database string) ([]string, error)
    GetTableSchema(ctx context.Context, database, table string) (*TableSchema, error)
    GetAllSchemas(ctx context.Context) ([]*TableSchema, error)
}
```

### 3. 图数据库组件 (internal/graph/)

负责存储和查询元数据血缘关系图。

**支持的图数据库：**
- NebulaGraph
- Neo4j

**目录结构：**
```
internal/graph/
├── graph.go            # 接口定义
├── types.go            # 类型定义
├── errors.go           # 错误定义
├── nebula/             # NebulaGraph 适配器
└── neo4j/              # Neo4j 适配器
```

**核心接口：**
```go
// GraphDB 图数据库接口
type GraphDB interface {
    Connect(ctx context.Context) error
    Close() error
    
    // 节点操作
    CreateNode(ctx context.Context, node *Node) error
    GetNode(ctx context.Context, id string) (*Node, error)
    UpdateNode(ctx context.Context, node *Node) error
    DeleteNode(ctx context.Context, id string) error
    
    // 边操作
    CreateEdge(ctx context.Context, edge *Edge) error
    GetEdge(ctx context.Context, id string) (*Edge, error)
    DeleteEdge(ctx context.Context, id string) error
    
    // 血缘查询
    GetUpstream(ctx context.Context, nodeID string, depth int) ([]*Node, []*Edge, error)
    GetDownstream(ctx context.Context, nodeID string, depth int) ([]*Node, []*Edge, error)
    GetLineage(ctx context.Context, nodeID string, depth int) (*LineageGraph, error)
}
```

## 业务服务层

### 元数据管理服务 (internal/service/metadata/)

负责元数据的同步、存储和查询。

```go
type Service struct {
    collectors map[string]collector.Collector
    graphDB    graph.GraphDB
}

func (s *Service) SyncMetadata(ctx context.Context, source string) error
func (s *Service) GetTableMetadata(ctx context.Context, database, table string) (*TableMetadata, error)
```

### 血缘查询服务 (internal/service/lineage/)

负责 SQL 血缘分析和血缘图查询。

```go
type Service struct {
    analyzer *lineage.Analyzer
    graphDB  graph.GraphDB
}

func (s *Service) AnalyzeSQL(ctx context.Context, sql string) (*LineageResult, error)
func (s *Service) GetColumnLineage(ctx context.Context, database, table, column string, depth int) (*LineageGraph, error)
```

## 数据模型

### 节点类型

| 类型 | 说明 |
|------|------|
| database | 数据库 |
| table | 表 |
| column | 列 |
| job | 作业/任务 |

### 边类型

| 类型 | 说明 |
|------|------|
| contains | 包含关系 (database → table → column) |
| depends_on | 依赖关系 (column → column) |
| produced_by | 产出关系 (table → job) |

## 数据流

```
┌──────────────┐     ┌──────────────┐     ┌──────────────┐
│   数据源      │────▶│  Collector   │────▶│  元数据存储   │
│ MySQL/PG/... │     │  元数据采集    │     │              │
└──────────────┘     └──────────────┘     └──────────────┘
                                                 │
                                                 ▼
┌──────────────┐     ┌──────────────┐     ┌──────────────┐
│   SQL 语句    │────▶│   Lineage    │────▶│   GraphDB    │
│              │     │   血缘解析    │     │   血缘存储    │
└──────────────┘     └──────────────┘     └──────────────┘
                                                 │
                                                 ▼
                                          ┌──────────────┐
                                          │   血缘查询    │
                                          │   API/CLI    │
                                          └──────────────┘
```

## 扩展性设计

### 添加新的数据源采集器

1. 在 `internal/collector/` 下创建新的子包
2. 实现 `Collector` 接口
3. 在配置中注册新的采集器类型

### 添加新的图数据库支持

1. 在 `internal/graph/` 下创建新的子包
2. 实现 `GraphDB` 接口
3. 在配置中注册新的图数据库类型

### 添加新的 SQL 方言支持

1. 更新 `internal/lineage/grammar/` 中的 ANTLR 语法文件
2. 重新生成解析器代码
3. 在 `internal/lineage/` 中添加方言特定的处理逻辑

## 部署架构

### 单机部署

```
┌─────────────────────────────────────┐
│           go-metadata               │
│  ┌─────────────┐  ┌─────────────┐  │
│  │   Server    │  │    CLI      │  │
│  └─────────────┘  └─────────────┘  │
└─────────────────────────────────────┘
         │                  │
         ▼                  ▼
┌─────────────┐      ┌─────────────┐
│   MySQL     │      │  NebulaGraph│
│  (元数据)    │      │  (血缘图)    │
└─────────────┘      └─────────────┘
```

### Docker Compose 部署

参考 `deployments/docker/docker-compose.yaml` 配置文件。

## 技术选型

| 组件 | 技术 | 说明 |
|------|------|------|
| 语言 | Go 1.24+ | 高性能、并发支持 |
| SQL 解析 | ANTLR4 | 强大的语法解析能力 |
| 图数据库 | NebulaGraph/Neo4j | 血缘关系存储 |
| 关系数据库 | MySQL/PostgreSQL | 元数据存储 |
| 容器化 | Docker | 部署和开发环境 |
