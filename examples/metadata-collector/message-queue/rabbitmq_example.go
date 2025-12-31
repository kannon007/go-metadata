// RabbitMQ 消息队列元数据采集样例
package main

import (
	"context"
	"fmt"
	"log"

	"go-metadata/internal/collector"
	"go-metadata/internal/collector/config"
	"go-metadata/internal/collector/mq/rabbitmq"
)

func main() {
	fmt.Println("=== RabbitMQ 消息队列元数据采集样例 ===")

	// 1. 创建采集器配置
	cfg := &config.ConnectorConfig{
		ID:       "rabbitmq-example",
		Type:     "rabbitmq",
		Category: collector.CategoryMessageQueue,
		Endpoint: "localhost:15672", // RabbitMQ Management API 端口
		Credentials: config.Credentials{
			User:     "guest",
			Password: "guest",
		},
		Properties: config.ConnectionProps{
			ConnectionTimeout: 30,
		},
		// 配置匹配规则
		Matching: &config.MatchingConfig{
			PatternType:   "glob",
			CaseSensitive: false,
			Schemas: &config.MatchingRule{ // vhost 作为 schema
				Include: []string{"*"},
				Exclude: []string{},
			},
			Tables: &config.MatchingRule{ // queue 作为 table
				Include: []string{"*"},
				Exclude: []string{"amq.*"}, // 排除系统队列
			},
		},
	}

	// 2. 创建 RabbitMQ 采集器
	collector, err := rabbitmq.NewCollector(cfg)
	if err != nil {
		log.Fatalf("创建采集器失败: %v", err)
	}

	ctx := context.Background()

	// 3. 连接 RabbitMQ
	fmt.Println("连接 RabbitMQ Management API...")
	if err := collector.Connect(ctx); err != nil {
		log.Fatalf("连接 RabbitMQ 失败: %v", err)
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

	// 6. 列出 VHost (Schema)
	fmt.Println("\n列出 VHost...")
	schemas, err := collector.ListSchemas(ctx, "rabbitmq")
	if err != nil {
		log.Fatalf("列出 VHost 失败: %v", err)
	}
	for _, schema := range schemas {
		fmt.Printf("VHost: %s\n", schema)
	}

	// 7. 列出队列 (Table)
	if len(schemas) > 0 {
		schema := schemas[0]
		fmt.Printf("\n列出 VHost '%s' 中的队列...\n", schema)
		
		result, err := collector.ListTables(ctx, "rabbitmq", schema, &collector.ListOptions{
			PageSize: 10,
		})
		if err != nil {
			log.Fatalf("列出队列失败: %v", err)
		}
		
		fmt.Printf("总队列数: %d\n", result.TotalCount)
		for _, table := range result.Tables {
			fmt.Printf("队列: %s\n", table)
		}

		// 8. 获取队列元数据
		if len(result.Tables) > 0 {
			table := result.Tables[0]
			fmt.Printf("\n获取队列 '%s' 的元数据...\n", table)
			
			metadata, err := collector.FetchTableMetadata(ctx, "rabbitmq", schema, table)
			if err != nil {
				log.Fatalf("获取队列元数据失败: %v", err)
			}
			
			fmt.Printf("类型: %s\n", metadata.Type)
			fmt.Printf("列数: %d\n", len(metadata.Columns))
			
			// 显示队列属性
			fmt.Println("队列属性:")
			for k, v := range metadata.Properties {
				fmt.Printf("  %s: %s\n", k, v)
			}
			
			// 显示消息结构列信息
			fmt.Println("消息结构:")
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

			// 9. 获取队列统计信息
			fmt.Printf("\n获取队列 '%s' 的统计信息...\n", table)
			stats, err := collector.FetchTableStatistics(ctx, "rabbitmq", schema, table)
			if err != nil {
				log.Printf("获取统计信息失败: %v", err)
			} else {
				fmt.Printf("消息数: %d\n", stats.RowCount)
				fmt.Printf("消息总大小: %d bytes (%.2f KB)\n", 
					stats.DataSizeBytes, float64(stats.DataSizeBytes)/1024)
				fmt.Printf("统计时间: %s\n", stats.CollectedAt.Format("2006-01-02 15:04:05"))
			}
		}

		// 10. 使用扩展接口功能
		if rabbitCollector, ok := collector.(*rabbitmq.Collector); ok {
			fmt.Printf("\n=== RabbitMQ 扩展功能 ===\n")
			
			// 列出交换机
			fmt.Printf("列出 VHost '%s' 的交换机...\n", schema)
			exchanges, err := rabbitCollector.ListExchanges(ctx, schema)
			if err != nil {
				log.Printf("列出交换机失败: %v", err)
			} else {
				for _, exchange := range exchanges {
					fmt.Printf("  交换机: %s (类型: %s, 持久化: %t)\n", 
						exchange.Name, exchange.Type, exchange.Durable)
				}
			}

			// 列出绑定关系
			fmt.Printf("\n列出 VHost '%s' 的绑定关系...\n", schema)
			bindings, err := rabbitCollector.ListBindings(ctx, schema)
			if err != nil {
				log.Printf("列出绑定关系失败: %v", err)
			} else {
				fmt.Printf("绑定关系数: %d\n", len(bindings))
				for i, binding := range bindings {
					if i >= 5 { // 只显示前5个
						fmt.Printf("  ... 还有 %d 个绑定关系\n", len(bindings)-5)
						break
					}
					fmt.Printf("  %s -> %s (路由键: %s)\n", 
						binding.Source, binding.Destination, binding.RoutingKey)
				}
			}

			// 列出消费者
			fmt.Printf("\n列出 VHost '%s' 的消费者...\n", schema)
			consumers, err := rabbitCollector.ListConsumers(ctx, schema)
			if err != nil {
				log.Printf("列出消费者失败: %v", err)
			} else {
				fmt.Printf("消费者数: %d\n", len(consumers))
				for _, consumer := range consumers {
					fmt.Printf("  消费者: %s (队列: %s, 节点: %s)\n", 
						consumer.ConsumerTag, consumer.Queue, consumer.Node)
				}
			}

			// 获取特定队列的绑定关系
			if len(result.Tables) > 0 {
				queueName := result.Tables[0]
				fmt.Printf("\n获取队列 '%s' 的绑定关系...\n", queueName)
				queueBindings, err := rabbitCollector.GetQueueBindings(ctx, schema, queueName)
				if err != nil {
					log.Printf("获取队列绑定关系失败: %v", err)
				} else {
					for _, binding := range queueBindings {
						fmt.Printf("  来源交换机: %s, 路由键: %s\n", 
							binding.Source, binding.RoutingKey)
					}
				}
			}
		}
	}

	fmt.Println("\n=== 采集完成 ===")
}