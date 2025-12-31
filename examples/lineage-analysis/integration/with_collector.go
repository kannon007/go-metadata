// 血缘分析与元数据采集器集成样例
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"go-metadata/internal/collector"
	"go-metadata/internal/collector/config"
	"go-metadata/internal/collector/rdbms/mysql"
	"go-metadata/internal/lineage"
	"go-metadata/internal/lineage/metadata"
)

func main() {
	fmt.Println("=== 血缘分析与元数据采集器集成样例 ===")

	// 1. 使用元数据采集器获取真实的表结构
	fmt.Println("使用元数据采集器获取表结构...")
	catalog, err := buildCatalogFromCollector()
	if err != nil {
		log.Fatalf("构建 catalog 失败: %v", err)
	}

	// 2. 创建血缘分析器
	analyzer := lineage.NewAnalyzer(catalog)

	// 3. 分析基于真实表结构的 SQL
	fmt.Println("\n=== 分析基于真实表结构的 SQL ===")
	realSQL := `
		SELECT u.id AS user_id,
		       u.name AS customer_name,
		       u.email,
		       COUNT(o.id) AS total_orders,
		       SUM(o.amount) AS total_spent,
		       AVG(o.amount) AS avg_order_value,
		       MAX(o.order_date) AS last_order_date,
		       MIN(o.order_date) AS first_order_date,
		       DATEDIFF(MAX(o.order_date), MIN(o.order_date)) AS customer_lifetime_days
		FROM users u
		LEFT JOIN orders o ON u.id = o.user_id AND o.status = 'completed'
		WHERE u.created_at >= '2023-01-01'
		GROUP BY u.id, u.name, u.email
		HAVING total_orders > 0
		ORDER BY total_spent DESC
		LIMIT 100
	`
	
	result, err := analyzer.Analyze(realSQL)
	if err != nil {
		log.Fatalf("分析 SQL 失败: %v", err)
	}

	fmt.Printf("SQL: %s\n", realSQL)
	printLineageResult(result)

	// 4. 批量分析多个 SQL 语句
	fmt.Println("\n=== 批量分析多个 SQL 语句 ===")
	sqlQueries := []string{
		`SELECT id, name, email FROM users WHERE status = 'active'`,
		`SELECT user_id, SUM(amount) as total FROM orders GROUP BY user_id`,
		`INSERT INTO user_summary (user_id, total_orders, total_amount) 
		 SELECT user_id, COUNT(*), SUM(amount) FROM orders GROUP BY user_id`,
		`UPDATE users SET last_login = NOW() WHERE id = 123`,
	}

	batchResults := batchAnalyzeSQL(analyzer, sqlQueries)
	for i, batchResult := range batchResults {
		fmt.Printf("\n--- SQL %d ---\n", i+1)
		fmt.Printf("SQL: %s\n", batchResult.SQL)
		if batchResult.Error != nil {
			fmt.Printf("错误: %v\n", batchResult.Error)
		} else {
			printLineageResult(batchResult.Result)
		}
	}

	// 5. 生成血缘关系报告
	fmt.Println("\n=== 生成血缘关系报告 ===")
	generateLineageReport(batchResults)

	fmt.Println("\n=== 集成样例完成 ===")
}

// buildCatalogFromCollector 使用采集器构建 catalog
func buildCatalogFromCollector() (lineage.Catalog, error) {
	// 创建 MySQL 采集器配置
	cfg := &config.ConnectorConfig{
		ID:       "mysql-lineage",
		Type:     "mysql",
		Category: collector.CategoryRDBMS,
		Endpoint: "localhost:3306",
		Credentials: config.Credentials{
			User:     "root",
			Password: "password",
		},
		Properties: config.ConnectionProps{
			ConnectionTimeout: 30,
		},
		Matching: &config.MatchingConfig{
			PatternType: "glob",
			Databases: &config.MatchingRule{
				Include: []string{"ecommerce", "analytics"},
				Exclude: []string{"information_schema", "performance_schema", "mysql", "sys"},
			},
		},
	}

	// 创建采集器
	coll, err := mysql.NewCollector(cfg)
	if err != nil {
		return nil, fmt.Errorf("创建采集器失败: %v", err)
	}

	ctx := context.Background()

	// 连接数据库
	if err := coll.Connect(ctx); err != nil {
		// 如果连接失败，使用模拟数据
		log.Printf("连接数据库失败，使用模拟数据: %v", err)
		return buildMockCatalog(), nil
	}
	defer coll.Close()

	// 构建元数据
	builder := metadata.NewMetadataBuilder()

	// 发现 catalog
	catalogs, err := coll.DiscoverCatalogs(ctx)
	if err != nil {
		return nil, fmt.Errorf("发现 catalog 失败: %v", err)
	}

	for _, catalogInfo := range catalogs {
		// 列出 schema
		schemas, err := coll.ListSchemas(ctx, catalogInfo.Catalog)
		if err != nil {
			log.Printf("列出 schema 失败: %v", err)
			continue
		}

		for _, schema := range schemas {
			// 列出表
			tableResult, err := coll.ListTables(ctx, catalogInfo.Catalog, schema, &collector.ListOptions{
				PageSize: 100,
			})
			if err != nil {
				log.Printf("列出表失败: %v", err)
				continue
			}

			for _, table := range tableResult.Tables {
				// 获取表元数据
				tableMetadata, err := coll.FetchTableMetadata(ctx, catalogInfo.Catalog, schema, table)
				if err != nil {
					log.Printf("获取表 %s.%s 元数据失败: %v", schema, table, err)
					continue
				}

				// 转换为 lineage 元数据格式
				tableSchema := &metadata.TableSchema{
					Database: schema,
					Table:    table,
					Columns:  make([]metadata.ColumnSchema, len(tableMetadata.Columns)),
				}

				for i, col := range tableMetadata.Columns {
					tableSchema.Columns[i] = metadata.ColumnSchema{
						Name:       col.Name,
						DataType:   col.Type,
						Nullable:   col.Nullable,
						PrimaryKey: col.PrimaryKey,
						Comment:    col.Comment,
					}
				}

				builder.AddTableSchema(tableSchema)
			}
		}
	}

	return builder.BuildCatalog(), nil
}

// buildMockCatalog 构建模拟的 catalog（当无法连接真实数据库时使用）
func buildMockCatalog() lineage.Catalog {
	return metadata.NewMetadataBuilder().
		WithDefaultDatabase("ecommerce").
		// 用户表
		AddTableSchema(&metadata.TableSchema{
			Database: "ecommerce",
			Table:    "users",
			Columns: []metadata.ColumnSchema{
				{Name: "id", DataType: "BIGINT", PrimaryKey: true},
				{Name: "name", DataType: "VARCHAR(100)", Nullable: false},
				{Name: "email", DataType: "VARCHAR(255)", Nullable: false},
				{Name: "status", DataType: "VARCHAR(20)", Nullable: false},
				{Name: "created_at", DataType: "TIMESTAMP", Nullable: false},
				{Name: "last_login", DataType: "TIMESTAMP"},
			},
		}).
		// 订单表
		AddTableSchema(&metadata.TableSchema{
			Database: "ecommerce",
			Table:    "orders",
			Columns: []metadata.ColumnSchema{
				{Name: "id", DataType: "BIGINT", PrimaryKey: true},
				{Name: "user_id", DataType: "BIGINT", Nullable: false},
				{Name: "amount", DataType: "DECIMAL(10,2)", Nullable: false},
				{Name: "status", DataType: "VARCHAR(50)", Nullable: false},
				{Name: "order_date", DataType: "DATE", Nullable: false},
			},
		}).
		// 用户汇总表
		AddTableSchema(&metadata.TableSchema{
			Database: "ecommerce",
			Table:    "user_summary",
			Columns: []metadata.ColumnSchema{
				{Name: "user_id", DataType: "BIGINT", PrimaryKey: true},
				{Name: "total_orders", DataType: "INT", Nullable: false},
				{Name: "total_amount", DataType: "DECIMAL(12,2)", Nullable: false},
				{Name: "last_updated", DataType: "TIMESTAMP", Nullable: false},
			},
		}).
		BuildCatalog()
}

// BatchAnalysisResult 批量分析结果
type BatchAnalysisResult struct {
	SQL    string
	Result *lineage.LineageResult
	Error  error
}

// batchAnalyzeSQL 批量分析 SQL 语句
func batchAnalyzeSQL(analyzer *lineage.Analyzer, sqlQueries []string) []BatchAnalysisResult {
	results := make([]BatchAnalysisResult, len(sqlQueries))
	
	for i, sql := range sqlQueries {
		result, err := analyzer.Analyze(sql)
		results[i] = BatchAnalysisResult{
			SQL:    sql,
			Result: result,
			Error:  err,
		}
	}
	
	return results
}

// generateLineageReport 生成血缘关系报告
func generateLineageReport(results []BatchAnalysisResult) {
	fmt.Println("血缘关系报告:")
	
	// 统计信息
	totalQueries := len(results)
	successfulQueries := 0
	failedQueries := 0
	totalColumns := 0
	totalSources := 0
	
	// 收集所有涉及的表
	sourceTableSet := make(map[string]bool)
	targetTableSet := make(map[string]bool)
	
	for _, result := range results {
		if result.Error != nil {
			failedQueries++
			continue
		}
		
		successfulQueries++
		totalColumns += len(result.Result.Columns)
		
		for _, col := range result.Result.Columns {
			totalSources += len(col.Sources)
			
			// 收集目标表
			if col.Target.Table != "" {
				targetTableSet[col.Target.Table] = true
			}
			
			// 收集源表
			for _, source := range col.Sources {
				if source.Table != "" {
					sourceTableSet[source.Table] = true
				}
			}
		}
	}
	
	fmt.Printf("  总查询数: %d\n", totalQueries)
	fmt.Printf("  成功分析: %d\n", successfulQueries)
	fmt.Printf("  分析失败: %d\n", failedQueries)
	fmt.Printf("  总输出列数: %d\n", totalColumns)
	fmt.Printf("  总源列数: %d\n", totalSources)
	fmt.Printf("  涉及源表数: %d\n", len(sourceTableSet))
	fmt.Printf("  涉及目标表数: %d\n", len(targetTableSet))
	
	if successfulQueries > 0 {
		fmt.Printf("  平均每查询列数: %.2f\n", float64(totalColumns)/float64(successfulQueries))
		fmt.Printf("  平均每列源数: %.2f\n", float64(totalSources)/float64(totalColumns))
	}
	
	// 列出涉及的表
	fmt.Println("\n  涉及的源表:")
	for table := range sourceTableSet {
		fmt.Printf("    %s\n", table)
	}
	
	fmt.Println("\n  涉及的目标表:")
	for table := range targetTableSet {
		fmt.Printf("    %s\n", table)
	}
	
	// 生成表级血缘关系图
	fmt.Println("\n  表级血缘关系:")
	generateTableLevelLineage(results)
}

// generateTableLevelLineage 生成表级血缘关系
func generateTableLevelLineage(results []BatchAnalysisResult) {
	// 表级血缘关系映射：目标表 -> 源表集合
	tableLevelLineage := make(map[string]map[string]bool)
	
	for _, result := range results {
		if result.Error != nil {
			continue
		}
		
		for _, col := range result.Result.Columns {
			targetTable := col.Target.Table
			if targetTable == "" {
				targetTable = "RESULT_SET" // 查询结果集
			}
			
			if _, exists := tableLevelLineage[targetTable]; !exists {
				tableLevelLineage[targetTable] = make(map[string]bool)
			}
			
			for _, source := range col.Sources {
				if source.Table != "" {
					tableLevelLineage[targetTable][source.Table] = true
				}
			}
		}
	}
	
	// 打印表级血缘关系
	for targetTable, sourceTables := range tableLevelLineage {
		fmt.Printf("    %s <- ", targetTable)
		first := true
		for sourceTable := range sourceTables {
			if !first {
				fmt.Printf(", ")
			}
			fmt.Printf("%s", sourceTable)
			first = false
		}
		fmt.Println()
	}
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

// 高级集成功能

// LineageAnalysisService 血缘分析服务
type LineageAnalysisService struct {
	analyzer *lineage.Analyzer
	catalog  lineage.Catalog
}

// NewLineageAnalysisService 创建血缘分析服务
func NewLineageAnalysisService(catalog lineage.Catalog) *LineageAnalysisService {
	return &LineageAnalysisService{
		analyzer: lineage.NewAnalyzer(catalog),
		catalog:  catalog,
	}
}

// AnalyzeWithMetadata 分析 SQL 并返回增强的元数据
func (s *LineageAnalysisService) AnalyzeWithMetadata(sql string) (*EnhancedLineageResult, error) {
	result, err := s.analyzer.Analyze(sql)
	if err != nil {
		return nil, err
	}
	
	enhanced := &EnhancedLineageResult{
		LineageResult: result,
		SQL:           sql,
		Metadata:      make(map[string]interface{}),
	}
	
	// 添加增强信息
	enhanced.Metadata["analysis_timestamp"] = fmt.Sprintf("%d", 1234567890) // 实际应该用 time.Now()
	enhanced.Metadata["source_table_count"] = len(s.getSourceTables(result))
	enhanced.Metadata["complexity_score"] = s.calculateComplexityScore(result)
	
	return enhanced, nil
}

// EnhancedLineageResult 增强的血缘分析结果
type EnhancedLineageResult struct {
	*lineage.LineageResult
	SQL      string                 `json:"sql"`
	Metadata map[string]interface{} `json:"metadata"`
}

// getSourceTables 获取源表列表
func (s *LineageAnalysisService) getSourceTables(result *lineage.LineageResult) []string {
	tableSet := make(map[string]bool)
	for _, col := range result.Columns {
		for _, source := range col.Sources {
			if source.Table != "" {
				tableSet[source.Table] = true
			}
		}
	}
	
	tables := make([]string, 0, len(tableSet))
	for table := range tableSet {
		tables = append(tables, table)
	}
	return tables
}

// calculateComplexityScore 计算复杂度分数
func (s *LineageAnalysisService) calculateComplexityScore(result *lineage.LineageResult) float64 {
	score := 0.0
	
	// 基于列数
	score += float64(len(result.Columns)) * 1.0
	
	// 基于源列数
	totalSources := 0
	for _, col := range result.Columns {
		totalSources += len(col.Sources)
	}
	score += float64(totalSources) * 0.5
	
	// 基于转换操作数
	totalOperators := 0
	for _, col := range result.Columns {
		totalOperators += len(col.Operators)
	}
	score += float64(totalOperators) * 2.0
	
	return score
}