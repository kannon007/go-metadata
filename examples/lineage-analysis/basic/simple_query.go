// 简单查询血缘分析样例
package main

import (
	"encoding/json"
	"fmt"
	"log"

	"go-metadata/internal/lineage"
	"go-metadata/internal/lineage/metadata"
)

func main() {
	fmt.Println("=== 简单查询血缘分析样例 ===")

	// 1. 构建元数据 catalog
	fmt.Println("构建元数据 catalog...")
	catalog := metadata.NewMetadataBuilder().
		WithDefaultDatabase("mydb").
		AddTable("mydb", "users", []string{"id", "name", "email", "created_at"}).
		AddTable("mydb", "orders", []string{"id", "user_id", "amount", "order_date"}).
		AddTable("mydb", "products", []string{"id", "name", "price", "category"}).
		BuildCatalog()

	// 2. 创建血缘分析器
	analyzer := lineage.NewAnalyzer(catalog)

	// 3. 分析简单 SELECT 查询
	fmt.Println("\n=== 分析简单 SELECT 查询 ===")
	simpleSQL := `SELECT id, name, email FROM users WHERE created_at > '2023-01-01'`
	
	result, err := analyzer.Analyze(simpleSQL)
	if err != nil {
		log.Fatalf("分析 SQL 失败: %v", err)
	}

	fmt.Printf("SQL: %s\n", simpleSQL)
	printLineageResult(result)

	// 4. 分析带别名的查询
	fmt.Println("\n=== 分析带别名的查询 ===")
	aliasSQL := `SELECT u.id AS user_id, u.name AS user_name, u.email 
	             FROM users u 
	             WHERE u.created_at > '2023-01-01'`
	
	result, err = analyzer.Analyze(aliasSQL)
	if err != nil {
		log.Fatalf("分析 SQL 失败: %v", err)
	}

	fmt.Printf("SQL: %s\n", aliasSQL)
	printLineageResult(result)

	// 5. 分析聚合查询
	fmt.Println("\n=== 分析聚合查询 ===")
	aggregateSQL := `SELECT COUNT(*) AS total_users, 
	                        MAX(created_at) AS latest_signup,
	                        MIN(created_at) AS earliest_signup
	                 FROM users`
	
	result, err = analyzer.Analyze(aggregateSQL)
	if err != nil {
		log.Fatalf("分析 SQL 失败: %v", err)
	}

	fmt.Printf("SQL: %s\n", aggregateSQL)
	printLineageResult(result)

	// 6. 分析计算列查询
	fmt.Println("\n=== 分析计算列查询 ===")
	calculatedSQL := `SELECT id, 
	                         name,
	                         UPPER(email) AS email_upper,
	                         YEAR(created_at) AS signup_year,
	                         CONCAT(name, ' - ', email) AS display_name
	                  FROM users`
	
	result, err = analyzer.Analyze(calculatedSQL)
	if err != nil {
		log.Fatalf("分析 SQL 失败: %v", err)
	}

	fmt.Printf("SQL: %s\n", calculatedSQL)
	printLineageResult(result)

	// 7. 分析 CASE WHEN 查询
	fmt.Println("\n=== 分析 CASE WHEN 查询 ===")
	caseSQL := `SELECT id,
	                   name,
	                   CASE 
	                       WHEN created_at > '2023-01-01' THEN 'New User'
	                       ELSE 'Old User'
	                   END AS user_type,
	                   CASE 
	                       WHEN email LIKE '%@gmail.com' THEN 'Gmail'
	                       WHEN email LIKE '%@yahoo.com' THEN 'Yahoo'
	                       ELSE 'Other'
	                   END AS email_provider
	            FROM users`
	
	result, err = analyzer.Analyze(caseSQL)
	if err != nil {
		log.Fatalf("分析 SQL 失败: %v", err)
	}

	fmt.Printf("SQL: %s\n", caseSQL)
	printLineageResult(result)

	fmt.Println("\n=== 分析完成 ===")
}

// printLineageResult 打印血缘分析结果
func printLineageResult(result *lineage.LineageResult) {
	fmt.Printf("输出列数: %d\n", len(result.Columns))
	
	for i, col := range result.Columns {
		fmt.Printf("  [%d] 目标列: %s\n", i+1, formatColumnRef(col.Target))
		
		if len(col.Sources) > 0 {
			fmt.Printf("      来源列: ")
			for j, source := range col.Sources {
				if j > 0 {
					fmt.Printf(", ")
				}
				fmt.Printf("%s", formatColumnRef(source))
			}
			fmt.Println()
		}
		
		if len(col.Operators) > 0 {
			fmt.Printf("      转换操作: %v\n", col.Operators)
		}
	}

	// 输出 JSON 格式（可选）
	if jsonBytes, err := json.MarshalIndent(result, "", "  "); err == nil {
		fmt.Printf("\nJSON 格式:\n%s\n", string(jsonBytes))
	}
}

// formatColumnRef 格式化列引用
func formatColumnRef(ref lineage.ColumnRef) string {
	if ref.Database != "" && ref.Table != "" {
		return fmt.Sprintf("%s.%s.%s", ref.Database, ref.Table, ref.Column)
	} else if ref.Table != "" {
		return fmt.Sprintf("%s.%s", ref.Table, ref.Column)
	} else {
		return ref.Column
	}
}

// 辅助函数：创建测试用的元数据
func createTestMetadata() lineage.Catalog {
	return metadata.NewMetadataBuilder().
		WithDefaultDatabase("testdb").
		// 用户表
		AddTableSchema(&metadata.TableSchema{
			Database: "testdb",
			Table:    "users",
			Columns: []metadata.ColumnSchema{
				{Name: "id", DataType: "BIGINT", PrimaryKey: true},
				{Name: "name", DataType: "VARCHAR(100)", Nullable: false},
				{Name: "email", DataType: "VARCHAR(255)", Nullable: false},
				{Name: "created_at", DataType: "TIMESTAMP", Nullable: false},
				{Name: "updated_at", DataType: "TIMESTAMP", Nullable: true},
			},
		}).
		// 订单表
		AddTableSchema(&metadata.TableSchema{
			Database: "testdb",
			Table:    "orders",
			Columns: []metadata.ColumnSchema{
				{Name: "id", DataType: "BIGINT", PrimaryKey: true},
				{Name: "user_id", DataType: "BIGINT", Nullable: false},
				{Name: "amount", DataType: "DECIMAL(10,2)", Nullable: false},
				{Name: "order_date", DataType: "DATE", Nullable: false},
				{Name: "status", DataType: "VARCHAR(50)", Nullable: false},
			},
		}).
		// 产品表
		AddTableSchema(&metadata.TableSchema{
			Database: "testdb",
			Table:    "products",
			Columns: []metadata.ColumnSchema{
				{Name: "id", DataType: "BIGINT", PrimaryKey: true},
				{Name: "name", DataType: "VARCHAR(200)", Nullable: false},
				{Name: "price", DataType: "DECIMAL(10,2)", Nullable: false},
				{Name: "category", DataType: "VARCHAR(100)", Nullable: false},
			},
		}).
		BuildCatalog()
}