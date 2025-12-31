// 复杂查询血缘分析样例（CTE、子查询、窗口函数）
package main

import (
	"encoding/json"
	"fmt"
	"log"

	"go-metadata/internal/lineage"
	"go-metadata/internal/lineage/metadata"
)

func main() {
	fmt.Println("=== 复杂查询血缘分析样例 ===")

	// 1. 构建元数据 catalog
	fmt.Println("构建元数据 catalog...")
	catalog := createComplexTestMetadata()

	// 2. 创建血缘分析器
	analyzer := lineage.NewAnalyzer(catalog)

	// 3. 分析 CTE (Common Table Expression) 查询
	fmt.Println("\n=== 分析 CTE 查询 ===")
	cteSQL := `
		WITH user_stats AS (
		    SELECT user_id,
		           COUNT(*) AS order_count,
		           SUM(amount) AS total_amount,
		           AVG(amount) AS avg_amount,
		           MAX(order_date) AS last_order_date
		    FROM orders
		    WHERE status = 'completed'
		    GROUP BY user_id
		),
		user_categories AS (
		    SELECT user_id,
		           CASE 
		               WHEN total_amount > 10000 THEN 'VIP'
		               WHEN total_amount > 5000 THEN 'Premium'
		               WHEN total_amount > 1000 THEN 'Regular'
		               ELSE 'Basic'
		           END AS customer_category,
		           CASE 
		               WHEN order_count > 50 THEN 'Frequent'
		               WHEN order_count > 20 THEN 'Regular'
		               ELSE 'Occasional'
		           END AS purchase_frequency
		    FROM user_stats
		)
		SELECT u.id,
		       u.name,
		       u.email,
		       u.city,
		       us.order_count,
		       us.total_amount,
		       us.avg_amount,
		       us.last_order_date,
		       uc.customer_category,
		       uc.purchase_frequency
		FROM users u
		INNER JOIN user_stats us ON u.id = us.user_id
		INNER JOIN user_categories uc ON u.id = uc.user_id
		ORDER BY us.total_amount DESC
	`
	
	result, err := analyzer.Analyze(cteSQL)
	if err != nil {
		log.Fatalf("分析 CTE SQL 失败: %v", err)
	}

	fmt.Printf("SQL: %s\n", cteSQL)
	printLineageResult(result)

	// 4. 分析嵌套子查询
	fmt.Println("\n=== 分析嵌套子查询 ===")
	subquerySQL := `
		SELECT u.id,
		       u.name,
		       u.email,
		       (SELECT COUNT(*) 
		        FROM orders o 
		        WHERE o.user_id = u.id AND o.status = 'completed') AS completed_orders,
		       (SELECT SUM(amount) 
		        FROM orders o 
		        WHERE o.user_id = u.id AND o.status = 'completed') AS total_spent,
		       (SELECT MAX(order_date) 
		        FROM orders o 
		        WHERE o.user_id = u.id) AS last_order_date,
		       CASE 
		           WHEN u.id IN (
		               SELECT user_id 
		               FROM orders 
		               WHERE amount > 1000 
		               GROUP BY user_id 
		               HAVING COUNT(*) > 5
		           ) THEN 'High Value'
		           ELSE 'Regular'
		       END AS customer_type
		FROM users u
		WHERE u.created_at > (
		    SELECT DATE_SUB(MAX(created_at), INTERVAL 1 YEAR) 
		    FROM users
		)
		ORDER BY total_spent DESC NULLS LAST
	`
	
	result, err = analyzer.Analyze(subquerySQL)
	if err != nil {
		log.Fatalf("分析子查询 SQL 失败: %v", err)
	}

	fmt.Printf("SQL: %s\n", subquerySQL)
	printLineageResult(result)

	// 5. 分析窗口函数查询
	fmt.Println("\n=== 分析窗口函数查询 ===")
	windowSQL := `
		SELECT u.id,
		       u.name,
		       u.city,
		       o.amount,
		       o.order_date,
		       -- 排名函数
		       ROW_NUMBER() OVER (PARTITION BY u.city ORDER BY o.amount DESC) AS city_rank,
		       RANK() OVER (ORDER BY o.amount DESC) AS global_rank,
		       DENSE_RANK() OVER (PARTITION BY u.city ORDER BY o.order_date DESC) AS recent_rank,
		       -- 聚合窗口函数
		       SUM(o.amount) OVER (PARTITION BY u.id ORDER BY o.order_date) AS running_total,
		       AVG(o.amount) OVER (PARTITION BY u.city) AS city_avg_amount,
		       COUNT(*) OVER (PARTITION BY u.id) AS user_order_count,
		       -- 偏移函数
		       LAG(o.amount, 1) OVER (PARTITION BY u.id ORDER BY o.order_date) AS prev_amount,
		       LEAD(o.order_date, 1) OVER (PARTITION BY u.id ORDER BY o.order_date) AS next_order_date,
		       -- 分布函数
		       PERCENT_RANK() OVER (ORDER BY o.amount) AS amount_percentile,
		       NTILE(4) OVER (ORDER BY o.amount) AS amount_quartile
		FROM users u
		INNER JOIN orders o ON u.id = o.user_id
		WHERE o.status = 'completed'
		  AND o.order_date >= '2023-01-01'
		ORDER BY u.city, city_rank
	`
	
	result, err = analyzer.Analyze(windowSQL)
	if err != nil {
		log.Fatalf("分析窗口函数 SQL 失败: %v", err)
	}

	fmt.Printf("SQL: %s\n", windowSQL)
	printLineageResult(result)

	// 6. 分析 UNION 查询
	fmt.Println("\n=== 分析 UNION 查询 ===")
	unionSQL := `
		SELECT 'user' AS entity_type,
		       id AS entity_id,
		       name AS entity_name,
		       email AS contact_info,
		       created_at AS created_date,
		       city AS location
		FROM users
		WHERE created_at > '2023-01-01'
		
		UNION ALL
		
		SELECT 'product' AS entity_type,
		       id AS entity_id,
		       name AS entity_name,
		       CONCAT('Category: ', category) AS contact_info,
		       created_at AS created_date,
		       'N/A' AS location
		FROM products
		WHERE created_at > '2023-01-01'
		
		ORDER BY created_date DESC, entity_type
	`
	
	result, err = analyzer.Analyze(unionSQL)
	if err != nil {
		log.Fatalf("分析 UNION SQL 失败: %v", err)
	}

	fmt.Printf("SQL: %s\n", unionSQL)
	printLineageResult(result)

	// 7. 分析递归 CTE（如果支持）
	fmt.Println("\n=== 分析递归 CTE ===")
	recursiveSQL := `
		WITH RECURSIVE user_hierarchy AS (
		    -- 基础查询：顶级用户
		    SELECT id, name, email, referrer_id, 1 AS level
		    FROM users
		    WHERE referrer_id IS NULL
		    
		    UNION ALL
		    
		    -- 递归查询：下级用户
		    SELECT u.id, u.name, u.email, u.referrer_id, uh.level + 1
		    FROM users u
		    INNER JOIN user_hierarchy uh ON u.referrer_id = uh.id
		    WHERE uh.level < 5  -- 限制递归深度
		)
		SELECT id,
		       name,
		       email,
		       referrer_id,
		       level,
		       CONCAT(REPEAT('  ', level - 1), name) AS hierarchy_display
		FROM user_hierarchy
		ORDER BY level, name
	`
	
	result, err = analyzer.Analyze(recursiveSQL)
	if err != nil {
		log.Printf("分析递归 CTE SQL 失败（可能不支持）: %v", err)
	} else {
		fmt.Printf("SQL: %s\n", recursiveSQL)
		printLineageResult(result)
	}

	fmt.Println("\n=== 分析完成 ===")
}

// createComplexTestMetadata 创建复杂测试用的元数据
func createComplexTestMetadata() lineage.Catalog {
	return metadata.NewMetadataBuilder().
		WithDefaultDatabase("analytics").
		// 用户表
		AddTableSchema(&metadata.TableSchema{
			Database: "analytics",
			Table:    "users",
			Columns: []metadata.ColumnSchema{
				{Name: "id", DataType: "BIGINT", PrimaryKey: true},
				{Name: "name", DataType: "VARCHAR(100)", Nullable: false},
				{Name: "email", DataType: "VARCHAR(255)", Nullable: false},
				{Name: "city", DataType: "VARCHAR(100)"},
				{Name: "referrer_id", DataType: "BIGINT"},
				{Name: "created_at", DataType: "TIMESTAMP", Nullable: false},
				{Name: "updated_at", DataType: "TIMESTAMP"},
			},
		}).
		// 订单表
		AddTableSchema(&metadata.TableSchema{
			Database: "analytics",
			Table:    "orders",
			Columns: []metadata.ColumnSchema{
				{Name: "id", DataType: "BIGINT", PrimaryKey: true},
				{Name: "user_id", DataType: "BIGINT", Nullable: false},
				{Name: "product_id", DataType: "BIGINT"},
				{Name: "amount", DataType: "DECIMAL(10,2)", Nullable: false},
				{Name: "order_date", DataType: "DATE", Nullable: false},
				{Name: "status", DataType: "VARCHAR(50)", Nullable: false},
				{Name: "created_at", DataType: "TIMESTAMP", Nullable: false},
			},
		}).
		// 产品表
		AddTableSchema(&metadata.TableSchema{
			Database: "analytics",
			Table:    "products",
			Columns: []metadata.ColumnSchema{
				{Name: "id", DataType: "BIGINT", PrimaryKey: true},
				{Name: "name", DataType: "VARCHAR(200)", Nullable: false},
				{Name: "price", DataType: "DECIMAL(10,2)", Nullable: false},
				{Name: "category", DataType: "VARCHAR(100)", Nullable: false},
				{Name: "brand", DataType: "VARCHAR(100)"},
				{Name: "created_at", DataType: "TIMESTAMP", Nullable: false},
			},
		}).
		// 订单项表
		AddTableSchema(&metadata.TableSchema{
			Database: "analytics",
			Table:    "order_items",
			Columns: []metadata.ColumnSchema{
				{Name: "id", DataType: "BIGINT", PrimaryKey: true},
				{Name: "order_id", DataType: "BIGINT", Nullable: false},
				{Name: "product_id", DataType: "BIGINT", Nullable: false},
				{Name: "quantity", DataType: "INT", Nullable: false},
				{Name: "unit_price", DataType: "DECIMAL(10,2)", Nullable: false},
				{Name: "total_price", DataType: "DECIMAL(10,2)", Nullable: false},
			},
		}).
		BuildCatalog()
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

	// 分析血缘复杂度
	analyzeLineageComplexity(result)
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

// analyzeLineageComplexity 分析血缘复杂度
func analyzeLineageComplexity(result *lineage.LineageResult) {
	fmt.Println("\n血缘复杂度分析:")
	
	// 统计源表数量
	sourceTableSet := make(map[string]bool)
	totalSources := 0
	maxSourcesPerColumn := 0
	
	for _, col := range result.Columns {
		if len(col.Sources) > maxSourcesPerColumn {
			maxSourcesPerColumn = len(col.Sources)
		}
		totalSources += len(col.Sources)
		
		for _, source := range col.Sources {
			if source.Table != "" {
				sourceTableSet[source.Table] = true
			}
		}
	}
	
	fmt.Printf("  涉及源表数: %d\n", len(sourceTableSet))
	fmt.Printf("  总源列数: %d\n", totalSources)
	fmt.Printf("  平均每列源数: %.2f\n", float64(totalSources)/float64(len(result.Columns)))
	fmt.Printf("  最大单列源数: %d\n", maxSourcesPerColumn)
	
	// 统计转换操作
	totalOperators := 0
	for _, col := range result.Columns {
		totalOperators += len(col.Operators)
	}
	fmt.Printf("  总转换操作数: %d\n", totalOperators)
	fmt.Printf("  平均每列转换数: %.2f\n", float64(totalOperators)/float64(len(result.Columns)))
	
	// 输出简化的 JSON 格式
	if jsonBytes, err := json.MarshalIndent(result, "", "  "); err == nil {
		fmt.Printf("\nJSON 格式:\n%s\n", string(jsonBytes))
	}
}