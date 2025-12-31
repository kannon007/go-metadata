// MinIO/S3 对象存储元数据采集样例
package main

import (
	"context"
	"fmt"
	"log"

	"go-metadata/internal/collector"
	"go-metadata/internal/collector/config"
	"go-metadata/internal/collector/oss/minio"
)

func main() {
	fmt.Println("=== MinIO 对象存储元数据采集样例 ===")

	// 1. 创建采集器配置
	cfg := &config.ConnectorConfig{
		ID:       "minio-example",
		Type:     "minio",
		Category: collector.CategoryObjectStorage,
		Endpoint: "localhost:9000", // MinIO 服务地址
		Credentials: config.Credentials{
			User:     "minioadmin",     // MinIO 访问密钥
			Password: "minioadmin",     // MinIO 秘密密钥
		},
		Properties: config.ConnectionProps{
			ConnectionTimeout: 30,
			Extra: map[string]string{
				"region": "us-east-1", // MinIO 区域
			},
		},
		// 配置匹配规则，只采集特定 bucket
		Matching: &config.MatchingConfig{
			PatternType:   "glob",
			CaseSensitive: false,
			Schemas: &config.MatchingRule{ // bucket 作为 schema
				Include: []string{"data-*", "logs"},
				Exclude: []string{"temp-*"},
			},
			Tables: &config.MatchingRule{ // 对象前缀作为 table
				Include: []string{"*"},
				Exclude: []string{".*"}, // 排除隐藏文件
			},
		},
		// 配置 Schema 推断
		Infer: &config.InferConfig{
			Enabled:    true,
			SampleSize: 50,   // 采样 50 个文件
			MaxDepth:   5,    // 最大嵌套深度
			TypeMerge:  config.TypeMergeMostCommon,
		},
	}

	// 2. 创建 MinIO 采集器
	collector, err := minio.NewCollector(cfg)
	if err != nil {
		log.Fatalf("创建采集器失败: %v", err)
	}

	ctx := context.Background()

	// 3. 连接 MinIO
	fmt.Println("连接 MinIO 服务...")
	if err := collector.Connect(ctx); err != nil {
		log.Fatalf("连接 MinIO 失败: %v", err)
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
		for k, v := range catalog.Properties {
			fmt.Printf("  %s: %s\n", k, v)
		}
	}

	// 6. 列出 Bucket (Schema)
	fmt.Println("\n列出 Bucket...")
	schemas, err := collector.ListSchemas(ctx, "minio")
	if err != nil {
		log.Fatalf("列出 Bucket 失败: %v", err)
	}
	for _, schema := range schemas {
		fmt.Printf("Bucket: %s\n", schema)
	}

	// 7. 列出对象前缀 (Table)
	if len(schemas) > 0 {
		schema := schemas[0]
		fmt.Printf("\n列出 Bucket '%s' 中的对象前缀...\n", schema)
		
		result, err := collector.ListTables(ctx, "minio", schema, &collector.ListOptions{
			PageSize: 10,
		})
		if err != nil {
			log.Fatalf("列出对象前缀失败: %v", err)
		}
		
		fmt.Printf("总前缀数: %d\n", result.TotalCount)
		for _, table := range result.Tables {
			fmt.Printf("前缀: %s\n", table)
		}

		// 8. 获取对象前缀元数据
		if len(result.Tables) > 0 {
			table := result.Tables[0]
			fmt.Printf("\n获取前缀 '%s' 的元数据...\n", table)
			
			metadata, err := collector.FetchTableMetadata(ctx, "minio", schema, table)
			if err != nil {
				log.Fatalf("获取前缀元数据失败: %v", err)
			}
			
			fmt.Printf("类型: %s\n", metadata.Type)
			fmt.Printf("Schema 推断: %t\n", metadata.InferredSchema)
			fmt.Printf("列数: %d\n", len(metadata.Columns))
			
			// 显示属性
			fmt.Println("属性:")
			for k, v := range metadata.Properties {
				fmt.Printf("  %s: %s\n", k, v)
			}
			
			// 显示列信息（推断的 Schema）
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

			// 9. 获取统计信息
			fmt.Printf("\n获取前缀 '%s' 的统计信息...\n", table)
			stats, err := collector.FetchTableStatistics(ctx, "minio", schema, table)
			if err != nil {
				log.Printf("获取统计信息失败: %v", err)
			} else {
				fmt.Printf("对象数: %d\n", stats.RowCount)
				fmt.Printf("总大小: %d bytes (%.2f MB)\n", 
					stats.DataSizeBytes, float64(stats.DataSizeBytes)/1024/1024)
				fmt.Printf("统计时间: %s\n", stats.CollectedAt.Format("2006-01-02 15:04:05"))
			}

			// 10. 获取分区信息（子前缀）
			fmt.Printf("\n获取前缀 '%s' 的分区信息...\n", table)
			partitions, err := collector.FetchPartitions(ctx, "minio", schema, table)
			if err != nil {
				log.Printf("获取分区信息失败: %v", err)
			} else if len(partitions) > 0 {
				fmt.Printf("子前缀数: %d\n", len(partitions))
				for _, partition := range partitions {
					fmt.Printf("  分区: %s (%s) - %s\n", 
						partition.Name, partition.Type, partition.Expression)
				}
			} else {
				fmt.Println("该前缀没有子分区")
			}
		}

		// 11. 使用扩展接口功能
		if minioCollector, ok := collector.(*minio.Collector); ok {
			fmt.Printf("\n=== MinIO 扩展功能 ===\n")
			
			// 列出对象前缀
			fmt.Printf("列出 Bucket '%s' 的所有前缀...\n", schema)
			prefixes, err := minioCollector.ListPrefixes(ctx, schema, "", "/")
			if err != nil {
				log.Printf("列出前缀失败: %v", err)
			} else {
				for _, prefix := range prefixes {
					fmt.Printf("  前缀: %s\n", prefix)
				}
			}

			// 获取 Bucket 策略
			fmt.Printf("\n获取 Bucket '%s' 的策略信息...\n", schema)
			policy, err := minioCollector.GetBucketPolicy(ctx, schema)
			if err != nil {
				log.Printf("获取 Bucket 策略失败: %v", err)
			} else {
				fmt.Printf("版本控制: %t\n", policy.Versioning)
				if policy.Encryption != "" {
					fmt.Printf("加密: %s\n", policy.Encryption)
				}
				if policy.Policy != "" {
					fmt.Printf("策略长度: %d 字符\n", len(policy.Policy))
				}
			}
		}
	}

	fmt.Println("\n=== 采集完成 ===")
}