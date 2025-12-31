// MySQL 基础元数据采集样例
package main

import (
	"context"
	"fmt"
	"log"

	"go-metadata/internal/collector"
	"go-metadata/internal/collector/config"
	"go-metadata/internal/collector/rdbms/mysql"
)

func main() {
	fmt.Println("=== MySQL 元数据采集样例 ===")

	// 1. 创建采集器配置
	cfg := &config.ConnectorConfig{
		ID:       "mysql-example",
		Type:     "mysql",
		Category: collector.CategoryRDBMS,
		Endpoint: "localhost:3306",
		Credentials: config.Credentials{
			User:     "root",
			Password: "password",
		},
		Properties: config.ConnectionProps{
			ConnectionTimeout: 30,
			MaxOpenConns:      10,
			MaxIdleConns:      5,
		},
		// 配置匹配规则，只采集特定数据库
		Matching: &config.MatchingConfig{
			PatternType:   "glob",
			CaseSensitive: false,
			Databases: &config.MatchingRule{
				Include: []string{"mydb", "test_*"},
				Exclude: []string{"information_schema", "performance_schema", "mysql", "sys"},
			},
			Tables: &config.MatchingRule{
				Include: []string{"*"},
				Exclude: []string{"tmp_*"},
			},
		},
		// 配置采集选项
		Collect: &config.CollectOptions{
			Partitions: true,
			Indexes:    true,
			Comments:   true,
			Statistics: true,
		},
	}

	// 2. 创建 MySQL 采集器
	collector, err := mysql.NewCollector(cfg)
	if err != nil {
		log.Fatalf("创建采集器失败: %v", err)
	}

	ctx := context.Background()

	// 3. 连接数据库
	fmt.Println("连接 MySQL 数据库...")
	if err := collector.Connect(ctx); err != nil {
		log.Fatalf("连接数据库失败: %v", err)
	}
	defer collector.Close()

	// 4. 健康检查
	health, err := collector.HealthCheck(ctx)
	if err != nil {
		log.Fatalf("健康检查失败: %v", err)
	}
	fmt.Printf("连接状态: %t, 延迟: %v, 版本: %s\n", 
		health.Connected, health.Latency, health.Version)

	// 5. 发现 Catalog
	fmt.Println("\n发现 Catalog...")
	catalogs, err := collector.DiscoverCatalogs(ctx)
	if err != nil {
		log.Fatalf("发现 Catalog 失败: %v", err)
	}
	for _, catalog := range catalogs {
		fmt.Printf("Catalog: %s (%s)\n", catalog.Catalog, catalog.Description)
	}

	// 6. 列出数据库 (Schema)
	fmt.Println("\n列出数据库...")
	schemas, err := collector.ListSchemas(ctx, "mysql")
	if err != nil {
		log.Fatalf("列出数据库失败: %v", err)
	}
	for _, schema := range schemas {
		fmt.Printf("数据库: %s\n", schema)
	}

	// 7. 列出表
	if len(schemas) > 0 {
		schema := schemas[0]
		fmt.Printf("\n列出数据库 '%s' 中的表...\n", schema)
		
		result, err := collector.ListTables(ctx, "mysql", schema, &collector.ListOptions{
			PageSize: 10,
		})
		if err != nil {
			log.Fatalf("列出表失败: %v", err)
		}
		
		fmt.Printf("总表数: %d\n", result.TotalCount)
		for _, table := range result.Tables {
			fmt.Printf("表: %s\n", table)
		}

		// 8. 获取表元数据
		if len(result.Tables) > 0 {
			table := result.Tables[0]
			fmt.Printf("\n获取表 '%s' 的元数据...\n", table)
			
			metadata, err := collector.FetchTableMetadata(ctx, "mysql", schema, table)
			if err != nil {
				log.Fatalf("获取表元数据失败: %v", err)
			}
			
			fmt.Printf("表类型: %s\n", metadata.Type)
			fmt.Printf("列数: %d\n", len(metadata.Columns))
			fmt.Printf("最后刷新时间: %s\n", metadata.LastRefreshedAt.Format("2006-01-02 15:04:05"))
			
			// 显示列信息
			fmt.Println("列信息:")
			for _, col := range metadata.Columns {
				nullable := "NOT NULL"
				if col.Nullable {
					nullable = "NULL"
				}
				fmt.Printf("  %d. %s %s %s", col.OrdinalPosition, col.Name, col.Type, nullable)
				if col.Comment != "" {
					fmt.Printf(" -- %s", col.Comment)
				}
				fmt.Println()
			}

			// 9. 获取表统计信息
			fmt.Printf("\n获取表 '%s' 的统计信息...\n", table)
			stats, err := collector.FetchTableStatistics(ctx, "mysql", schema, table)
			if err != nil {
				log.Printf("获取统计信息失败: %v", err)
			} else {
				fmt.Printf("行数: %d\n", stats.RowCount)
				fmt.Printf("数据大小: %d bytes\n", stats.DataSizeBytes)
				fmt.Printf("统计时间: %s\n", stats.CollectedAt.Format("2006-01-02 15:04:05"))
			}

			// 10. 获取分区信息
			fmt.Printf("\n获取表 '%s' 的分区信息...\n", table)
			partitions, err := collector.FetchPartitions(ctx, "mysql", schema, table)
			if err != nil {
				log.Printf("获取分区信息失败: %v", err)
			} else if len(partitions) > 0 {
				fmt.Printf("分区数: %d\n", len(partitions))
				for _, partition := range partitions {
					fmt.Printf("  分区: %s (%s)\n", partition.Name, partition.Type)
				}
			} else {
				fmt.Println("该表没有分区")
			}
		}
	}

	fmt.Println("\n=== 采集完成 ===")
}