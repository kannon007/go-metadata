// JOIN 查询血缘分析样例
package main

import (
	"encoding/json"
	"fmt"
	"log"

	"go-metadata/internal/lineage"
	"go-metadata/internal/lineage/metadata"
)

func main() {
	fmt.Println("=== JOIN 查询血缘分析样例 ===")

	// 1. 构建元数据 catalog
	fmt.Println("构建元数据 catalog...")
	catalog := metadata.NewMetadataBuilder().
		WithDefaultDatabase("ecommerce").
		// 用户表
		AddTableSchema(&metadata.TableSchema{
			Database: "ecommerce",
			Table:    "users",
			Columns: []metadata.ColumnSchema{
				{Name: "id", DataType: "BIGINT", PrimaryKey: true},
				{Name: "name", DataType: "VARCHAR(100)"},
				{Name: "email", DataType: "VARCHAR(255)"},
				{Name: "city", DataType: "VARCHAR(100)"},
				{Name: "created_at", DataType: "TIMESTAMP"},
			},
		}).
		// 订单表
		AddTableSchema(&metadata.TableSchema{
			Database: "ecommerce",
			Table:    "orders",
			Columns: []metadata.ColumnSchema{
				{Name: "id", DataType: "BIGINT", PrimaryKey: true},
				{Name: "user_id", DataType: "BIGINT"},
				{Name: "product_id", DataType: "BIGINT"},
				{Name: "quantity", DataType: "INT"},
				{Name: "amount", DataType: "DECIMAL(10,2)"},
				{Name: "order_date", DataType: "DATE"},
				{Name: "status", DataType: "VARCHAR(50)"},
			},
		}).
		// 产品表
		AddTableSchema(&metadata.TableSchema{
			Database: "ecommerce",
			Table:    "products",
			Columns: []metadata.ColumnSchema{
				{Name: "id", DataType: "BIGINT", PrimaryKey: true},
				{Name: "name", DataType: "VARCHAR(200)"},
				{Name: "price", DataType: "DECIMAL(10,2)"},
				{Name: "category", DataType: "VARCHAR(100)"},
				{Name: "brand", DataType: "VARCHAR(100)"},
			},
		}).
		// 订单项表
		AddTableSchema(&metadata.TableSchema{
			Database: "ecommerce",
			Table:    "order_items",
			Columns: []metadata.ColumnSchema{
				{Name: "id", DataType: "BIGINT", PrimaryKey: true},
				{Name: "order_id", DataType: "BIGINT"},
				{Name: "product_id", DataType: "BIGINT"},
				{Name: "quantity", DataType: "INT"},
				{Name: "unit_price", DataType: "DECIMAL(10,2)"},
				{Name: "total_price", DataType: "DECIMAL(10,2)"},
			},
		}).
		BuildCatalog()

	// 2. 创建血缘分析器
	analyzer := lineage.NewAnalyzer(catalog)

	// 3. 分析简单 INNER JOIN
	fmt.Println("\n=== 分析简单 INNER JOIN ===")
	innerJoinSQL := `
		SELECT u.id AS user_id,
		       u.name AS user_name,
		       u.email,
		       o.id AS order_id,
		       o.amount,
		       o.order_date
		FROM users u
		INNER JOIN orders o ON u.id = o.user_id
		WHERE o.status = 'completed'
	`
	
	result, err := analyzer.Analyze(innerJoinSQL)
	if err != nil {
		log.Fatalf("分析 SQL 失败: %v", err)
	}

	fmt.Printf("SQL: %s\n", innerJoinSQL)
	printLineageResult(result)

	// 4. 分析 LEFT JOIN 查询
	fmt.Println("\n=== 分析 LEFT JOIN 查询 ===")
	leftJoinSQL := `
		SELECT u.id,
		       u.name,
		       u.city,
		       COALESCE(o.total_orders, 0) AS total_orders,
		       COALESCE(o.total_amount, 0) AS total_spent
		FROM users u
		LEFT JOIN (
		    SELECT user_id,
		           COUNT(*) AS total_orders,
		           SUM(amount) AS total_amount
		    FROM orders
		    WHERE status = 'completed'
		    GROUP BY user_id
		) o ON u.id = o.user_id
	`
	
	result, err = analyzer.Analyze(leftJoinSQL)
	if err != nil {
		log.Fatalf("分析 SQL 失败: %v", err)
	}

	fmt.Printf("SQL: %s\n", leftJoinSQL)
	printLineageResult(result)

	// 5. 分析多表 JOIN
	fmt.Println("\n=== 分析多表 JOIN ===")
	multiJoinSQL := `
		SELECT u.name AS customer_name,
		       u.email,
		       u.city,
		       p.name AS product_name,
		       p.category,
		       p.brand,
		       oi.quantity,
		       oi.unit_price,
		       oi.total_price,
		       o.order_date
		FROM users u
		INNER JOIN orders o ON u.id = o.user_id
		INNER JOIN order_items oi ON o.id = oi.order_id
		INNER JOIN products p ON oi.product_id = p.id
		WHERE o.status = 'completed'
		  AND o.order_date >= '2023-01-01'
	`
	
	result, err = analyzer.Analyze(multiJoinSQL)
	if err != nil {
		log.Fatalf("分析 SQL 失败: %v", err)
	}

	fmt.Printf("SQL: %s\n", multiJoinSQL)
	printLineageResult(result)

	// 6. 分析带聚合的 JOIN
	fmt.Println("\n=== 分析带聚合的 JOIN ===")
	aggregateJoinSQL := `
		SELECT u.city,
		       p.category,
		       COUNT(DISTINCT u.id) AS unique_customers,
		       COUNT(o.id) AS total_orders,
		       SUM(oi.quantity) AS total_quantity,
		       SUM(oi.total_price) AS total_revenue,
		       AVG(oi.unit_price) AS avg_unit_price
		FROM users u
		INNER JOIN orders o ON u.id = o.user_id
		INNER JOIN order_items oi ON o.id = oi.order_id
		INNER JOIN products p ON oi.product_id = p.id
		WHERE o.status = 'completed'
		GROUP BY u.city, p.category
		HAVING SUM(oi.total_price) > 1000
		ORDER BY total_revenue DESC
	`
	
	result, err = analyzer.Analyze(aggregateJoinSQL)
	if err != nil {
		log.Fatalf("分析 SQL 失败: %v", err)
	}

	fmt.Printf("SQL: %s\n", aggregateJoinSQL)
	printLineageResult(result)

	// 7. 分析自连接查询
	fmt.Println("\n=== 分析自连接查询 ===")
	selfJoinSQL := `
		SELECT u1.id AS user_id,
		       u1.name AS user_name,
		       u1.city AS user_city,
		       u2.name AS same_city_user,
		       u2.email AS same_city_email
		FROM users u1
		INNER JOIN users u2 ON u1.city = u2.city AND u1.id != u2.id
		WHERE u1.city = 'New York'
		ORDER BY u1.name, u2.name
	`
	
	result, err = analyzer.Analyze(selfJoinSQL)
	if err != nil {
		log.Fatalf("分析 SQL 失败: %v", err)
	}

	fmt.Printf("SQL: %s\n", selfJoinSQL)
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

	// 生成血缘关系图（文本格式）
	fmt.Println("\n血缘关系图:")
	generateLineageGraph(result)
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

// generateLineageGraph 生成血缘关系图（简单文本格式）
func generateLineageGraph(result *lineage.LineageResult) {
	// 收集所有源表
	sourceTableMap := make(map[string][]string)
	
	for _, col := range result.Columns {
		for _, source := range col.Sources {
			tableName := source.Table
			if tableName == "" {
				continue
			}
			
			if _, exists := sourceTableMap[tableName]; !exists {
				sourceTableMap[tableName] = []string{}
			}
			
			// 避免重复列
			found := false
			for _, existingCol := range sourceTableMap[tableName] {
				if existingCol == source.Column {
					found = true
					break
				}
			}
			if !found {
				sourceTableMap[tableName] = append(sourceTableMap[tableName], source.Column)
			}
		}
	}
	
	// 打印源表和列
	fmt.Println("  源表:")
	for table, columns := range sourceTableMap {
		fmt.Printf("    %s: %v\n", table, columns)
	}
	
	// 打印目标列
	fmt.Println("  目标列:")
	for _, col := range result.Columns {
		targetCol := col.Target.Column
		if col.Target.Table != "" {
			targetCol = col.Target.Table + "." + targetCol
		}
		fmt.Printf("    %s\n", targetCol)
	}
	
	// 输出简化的 JSON 格式
	if jsonBytes, err := json.MarshalIndent(result, "", "  "); err == nil {
		fmt.Printf("\nJSON 格式:\n%s\n", string(jsonBytes))
	}
}