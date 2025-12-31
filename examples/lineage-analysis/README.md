# 血缘解析使用样例

本目录包含了 SQL 血缘解析组件的各种使用样例，展示如何使用 go-metadata 项目中的血缘解析功能。

## 目录结构

```
lineage-analysis/
├── README.md                    # 本文件
├── basic/                       # 基础使用样例
│   ├── simple_query.go         # 简单查询血缘分析
│   ├── join_query.go           # JOIN 查询血缘分析
│   └── insert_query.go         # INSERT 语句血缘分析
├── advanced/                    # 高级使用样例
│   ├── complex_query.go        # 复杂查询（CTE、子查询）
│   ├── window_functions.go     # 窗口函数血缘分析
│   └── ddl_parsing.go          # DDL 解析和元数据构建
├── multi-dialect/               # 多方言支持样例
│   ├── flink_sql.go            # Flink SQL 血缘分析
│   ├── spark_sql.go            # Spark SQL 血缘分析
│   └── hive_sql.go             # Hive SQL 血缘分析
├── integration/                 # 集成样例
│   ├── with_collector.go       # 与元数据采集器集成
│   └── batch_analysis.go       # 批量血缘分析
└── testdata/                    # 测试数据
    ├── sample_schema.json      # 样例 Schema
    ├── flink_ddl.sql           # Flink DDL 样例
    └── complex_queries.sql     # 复杂查询样例
```

## 快速开始

### 1. 简单查询血缘分析

```bash
cd examples/lineage-analysis/basic
go run simple_query.go
```

### 2. 复杂查询血缘分析

```bash
cd examples/lineage-analysis/advanced
go run complex_query.go
```

### 3. Flink SQL 血缘分析

```bash
cd examples/lineage-analysis/multi-dialect
go run flink_sql.go
```

## 核心概念

### 血缘分析结果

血缘分析的核心输出是 `LineageResult`，包含每个输出列的血缘信息：

```go
type LineageResult struct {
    Columns []ColumnLineage  // 列级血缘列表
}

type ColumnLineage struct {
    Target    ColumnRef    // 目标列
    Sources   []ColumnRef  // 来源列
    Operators []string     // 转换操作
}
```

### 元数据管理

使用 `MetadataBuilder` 构建元数据目录：

```go
// 方式1：手动添加表结构
catalog := metadata.NewMetadataBuilder().
    AddTable("", "users", []string{"id", "name", "email"}).
    AddTable("", "orders", []string{"id", "user_id", "amount"}).
    BuildCatalog()

// 方式2：从 DDL 自动解析
analyzer := metadata.NewMetadataBuilder().
    LoadFromDDL(ddlSQL).
    BuildAnalyzer()

// 方式3：从 JSON 文件加载
catalog := metadata.NewMetadataBuilder().
    LoadFromJSONFile("schema.json").
    BuildCatalog()
```

## 支持的 SQL 语法

### DML 语句
- `SELECT` - 包括 JOIN, 子查询, CTE, UNION
- `INSERT INTO ... SELECT`
- `UPDATE ... SET`
- `DELETE FROM`
- `MERGE INTO`

### DDL 语句（用于元数据提取）
- `CREATE TABLE` - 支持各种数据库方言
- `CREATE VIEW` / `CREATE TEMPORARY VIEW`
- `CREATE EXTERNAL TABLE`

### 特殊语法支持

| 引擎 | 支持的语法 |
|------|-----------|
| **Flink** | WATERMARK, WITH 连接器, NOT ENFORCED |
| **Spark/Hive** | PARTITIONED BY, STORED AS, LOCATION |
| **PostgreSQL** | GENERATED, IDENTITY, INHERITS |
| **MySQL** | ENGINE, CHARSET, AUTO_INCREMENT |
| **ClickHouse** | ENGINE, ORDER BY, TTL |

## 使用场景

1. **数据血缘追踪**: 了解数据的来源和流向
2. **影响分析**: 评估表结构变更的影响范围
3. **数据质量**: 追踪数据转换过程中的质量问题
4. **合规审计**: 满足数据治理和合规要求
5. **ETL 优化**: 优化数据处理流程

## 最佳实践

1. **元数据管理**: 保持元数据的准确性和及时更新
2. **批量处理**: 对于大量 SQL 使用批量分析提高效率
3. **错误处理**: 妥善处理解析失败的 SQL 语句
4. **性能优化**: 对于复杂查询考虑缓存解析结果
5. **方言选择**: 根据实际使用的 SQL 引擎选择合适的方言