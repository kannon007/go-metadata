# SQL Lineage Parser

SQL 血缘解析库，支持从 SQL 语句中提取列级别的数据血缘关系。

## 功能特性

- **列级血缘分析**: 追踪每个输出列的来源列和转换操作
- **多 SQL 方言支持**: Flink SQL, Spark SQL, Hive, MySQL, PostgreSQL, ClickHouse, Doris 等
- **DDL 自动解析**: 从 CREATE TABLE/VIEW 语句自动提取表结构
- **复杂查询支持**: JOIN, 子查询, CTE, 窗口函数, CASE WHEN 等

## 快速开始

### 基本用法

```go
package main

import (
    "fmt"
    "go-metadata/internal/lineage"
    "go-metadata/internal/lineage/metadata"
)

func main() {
    // 1. 构建元数据 catalog
    catalog := metadata.NewMetadataBuilder().
        AddTable("", "users", []string{"id", "name", "email"}).
        AddTable("", "orders", []string{"id", "user_id", "amount"}).
        BuildCatalog()

    // 2. 创建分析器
    analyzer := lineage.NewAnalyzer(catalog)

    // 3. 分析 SQL
    sql := `SELECT u.id, u.name, SUM(o.amount) AS total
            FROM users u
            JOIN orders o ON u.id = o.user_id
            GROUP BY u.id, u.name`

    result, err := analyzer.Analyze(sql)
    if err != nil {
        panic(err)
    }

    // 4. 输出血缘结果
    for _, col := range result.Columns {
        fmt.Printf("Column: %s\n", col.Target.Column)
        fmt.Printf("  Sources: %v\n", col.Sources)
        fmt.Printf("  Operators: %v\n", col.Operators)
    }
}
```


### 从 DDL 自动解析表结构

```go
// 从 DDL 语句自动提取表结构
ddl := `
    CREATE TABLE user_events (
        user_id BIGINT,
        event_type STRING,
        event_time TIMESTAMP(3),
        WATERMARK FOR event_time AS event_time - INTERVAL '5' SECOND
    ) WITH (
        'connector' = 'kafka',
        'topic' = 'events'
    );

    CREATE TABLE user_info (
        user_id BIGINT,
        user_name STRING,
        PRIMARY KEY (user_id) NOT ENFORCED
    );
`

analyzer := metadata.NewMetadataBuilder().
    LoadFromDDL(ddl).
    BuildAnalyzer()

// 直接分析查询
result, _ := analyzer.Analyze(`
    SELECT e.user_id, u.user_name, COUNT(*) as cnt
    FROM user_events e
    LEFT JOIN user_info u ON e.user_id = u.user_id
    GROUP BY e.user_id, u.user_name
`)
```

### INSERT 语句血缘分析

```go
sql := `INSERT INTO user_summary (user_id, total_amount)
        SELECT user_id, SUM(amount)
        FROM orders
        GROUP BY user_id`

result, _ := analyzer.Analyze(sql)

// result.Columns[0].Target = {Table: "user_summary", Column: "user_id"}
// result.Columns[0].Sources = [{Table: "orders", Column: "user_id"}]
```

## 支持的 SQL 语法

### DML 语句
- `SELECT` - 包括 JOIN, 子查询, CTE, UNION
- `INSERT INTO ... SELECT`
- `UPDATE ... SET`
- `DELETE FROM`
- `MERGE INTO`

### DDL 语句 (用于元数据提取)
- `CREATE TABLE` - 支持各种数据库方言
- `CREATE VIEW` / `CREATE TEMPORARY VIEW`
- `CREATE EXTERNAL TABLE`


### 特殊语法支持

| 引擎 | 支持的语法 |
|------|-----------|
| **Flink** | WATERMARK, WITH 连接器, NOT ENFORCED, FOR SYSTEM_TIME AS OF |
| **Spark/Hive** | PARTITIONED BY, STORED AS, LOCATION, TBLPROPERTIES |
| **PostgreSQL** | GENERATED, IDENTITY, INHERITS, TABLESPACE |
| **MySQL** | ENGINE, CHARSET, COLLATE, AUTO_INCREMENT |
| **ClickHouse** | ENGINE, ORDER BY, TTL |
| **Doris/StarRocks** | DISTRIBUTED BY, PROPERTIES |

## 数据结构

### LineageResult

```go
type LineageResult struct {
    Columns []ColumnLineage  // 列级血缘列表
}

type ColumnLineage struct {
    Target    ColumnRef    // 目标列
    Sources   []ColumnRef  // 来源列
    Operators []string     // 转换操作
}

type ColumnRef struct {
    Database string  // 数据库名
    Table    string  // 表名
    Column   string  // 列名
}
```

### 示例输出

对于 SQL:
```sql
SELECT u.name, SUM(o.amount) AS total
FROM users u JOIN orders o ON u.id = o.user_id
GROUP BY u.name
```

血缘结果:
```json
{
  "columns": [
    {
      "target": {"table": "", "column": "name"},
      "sources": [{"table": "users", "column": "name"}],
      "operators": ["u.name"]
    },
    {
      "target": {"table": "", "column": "total"},
      "sources": [{"table": "orders", "column": "amount"}],
      "operators": ["SUM(o.amount)"]
    }
  ]
}
```


## 元数据管理

### MetadataBuilder

流式 API 构建元数据:

```go
builder := metadata.NewMetadataBuilder().
    WithDefaultDatabase("default").
    AddTable("db", "table1", []string{"col1", "col2"}).
    AddTableSchema(&metadata.TableSchema{...}).
    LoadFromDDL(ddlSQL).
    LoadFromJSON(jsonBytes).
    LoadFromJSONFile("schema.json")

// 获取不同类型的结果
catalog := builder.BuildCatalog()      // lineage.Catalog 接口
provider := builder.Build()            // MemoryProvider
analyzer := builder.BuildAnalyzer()    // lineage.Analyzer
```

### JSON Schema 格式

```json
[
  {
    "database": "mydb",
    "table": "users",
    "table_type": "TABLE",
    "columns": [
      {"name": "id", "data_type": "BIGINT", "primary_key": true},
      {"name": "name", "data_type": "VARCHAR(100)", "nullable": false},
      {"name": "email", "data_type": "VARCHAR(255)"}
    ]
  }
]
```

## 测试

```bash
# 运行所有测试
go test ./...

# 运行特定测试
go test -v -run "TestFlinkComplete" ./tests/
go test -v -run "TestSparkComplete" ./tests/
```


## 目录结构

```
lineage/
├── lineage.go          # 主入口 Analyzer
├── catalog.go          # Catalog 接口定义
├── types.go            # 数据类型定义
├── parser.go           # SQL 解析器
├── builder.go          # AST 构建器
├── extractor.go        # 血缘提取器
├── grammar/            # ANTLR 语法文件
│   ├── SQLLexer.g4
│   ├── SQLParser.g4
│   └── generate.bat
├── parser/             # ANTLR 生成的解析器代码
├── ast/                # AST 节点定义
├── metadata/           # 元数据管理
│   ├── provider.go     # MemoryProvider
│   ├── builder.go      # MetadataBuilder
│   ├── ddl_parser.go   # DDL 解析器
│   └── adapter.go      # Catalog 适配器
├── tests/              # 测试用例
│   ├── flink_complete_test.go
│   ├── spark_complete_test.go
│   └── complex_test.go
└── testdata/           # 测试数据
    ├── flink/complete_example.sql
    └── spark/complete_example.sql
```

## License

MIT
