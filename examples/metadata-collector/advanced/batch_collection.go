// 批量元数据采集样例
package main

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"go-metadata/internal/collector"
	"go-metadata/internal/collector/config"
	"go-metadata/internal/collector/oss/minio"
	"go-metadata/internal/collector/mq/rabbitmq"
	"go-metadata/internal/collector/rdbms/mysql"
)

// CollectionResult 采集结果
type CollectionResult struct {
	CollectorID string
	Success     bool
	Error       error
	Duration    time.Duration
	TableCount  int
}

func main() {
	fmt.Println("=== 批量元数据采集样例 ===")

	// 1. 定义多个数据源配置
	configs := []*config.ConnectorConfig{
		{
			ID:       "mysql-prod",
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
					Include: []string{"prod_*"},
					Exclude: []string{"*_temp"},
				},
			},
		},
		{
			ID:       "minio-data",
			Type:     "minio",
			Category: collector.CategoryObjectStorage,
			Endpoint: "localhost:9000",
			Credentials: config.Credentials{
				User:     "minioadmin",
				Password: "minioadmin",
			},
			Properties: config.ConnectionProps{
				ConnectionTimeout: 30,
			},
			Matching: &config.MatchingConfig{
				PatternType: "glob",
				Schemas: &config.MatchingRule{
					Include: []string{"data-*"},
				},
			},
		},
		{
			ID:       "rabbitmq-prod",
			Type:     "rabbitmq",
			Category: collector.CategoryMessageQueue,
			Endpoint: "localhost:15672",
			Credentials: config.Credentials{
				User:     "guest",
				Password: "guest",
			},
			Properties: config.ConnectionProps{
				ConnectionTimeout: 30,
			},
		},
	}

	// 2. 并发采集所有数据源
	ctx := context.Background()
	results := make(chan CollectionResult, len(configs))
	var wg sync.WaitGroup

	fmt.Printf("开始并发采集 %d 个数据源...\n", len(configs))
	startTime := time.Now()

	for _, cfg := range configs {
		wg.Add(1)
		go func(config *config.ConnectorConfig) {
			defer wg.Done()
			result := collectMetadata(ctx, config)
			results <- result
		}(cfg)
	}

	// 等待所有采集完成
	go func() {
		wg.Wait()
		close(results)
	}()

	// 3. 收集和显示结果
	var successCount, failCount int
	var totalTables int

	fmt.Println("\n=== 采集结果 ===")
	for result := range results {
		if result.Success {
			successCount++
			totalTables += result.TableCount
			fmt.Printf("✓ %s: 成功 (耗时: %v, 表数: %d)\n", 
				result.CollectorID, result.Duration, result.TableCount)
		} else {
			failCount++
			fmt.Printf("✗ %s: 失败 (耗时: %v, 错误: %v)\n", 
				result.CollectorID, result.Duration, result.Error)
		}
	}

	totalDuration := time.Since(startTime)
	fmt.Printf("\n=== 汇总 ===\n")
	fmt.Printf("总耗时: %v\n", totalDuration)
	fmt.Printf("成功: %d, 失败: %d\n", successCount, failCount)
	fmt.Printf("总表数: %d\n", totalTables)
	fmt.Printf("平均每秒采集: %.2f 个表\n", float64(totalTables)/totalDuration.Seconds())
}

// collectMetadata 采集单个数据源的元数据
func collectMetadata(ctx context.Context, cfg *config.ConnectorConfig) CollectionResult {
	result := CollectionResult{
		CollectorID: cfg.ID,
	}
	
	startTime := time.Now()
	defer func() {
		result.Duration = time.Since(startTime)
	}()

	// 创建采集器
	var coll collector.Collector
	var err error

	switch cfg.Type {
	case "mysql":
		coll, err = mysql.NewCollector(cfg)
	case "minio":
		coll, err = minio.NewCollector(cfg)
	case "rabbitmq":
		coll, err = rabbitmq.NewCollector(cfg)
	default:
		result.Error = fmt.Errorf("不支持的采集器类型: %s", cfg.Type)
		return result
	}

	if err != nil {
		result.Error = fmt.Errorf("创建采集器失败: %v", err)
		return result
	}

	// 连接数据源
	if err := coll.Connect(ctx); err != nil {
		result.Error = fmt.Errorf("连接失败: %v", err)
		return result
	}
	defer coll.Close()

	// 健康检查
	health, err := coll.HealthCheck(ctx)
	if err != nil || !health.Connected {
		result.Error = fmt.Errorf("健康检查失败: %v", err)
		return result
	}

	// 发现并采集元数据
	catalogs, err := coll.DiscoverCatalogs(ctx)
	if err != nil {
		result.Error = fmt.Errorf("发现 Catalog 失败: %v", err)
		return result
	}

	var totalTables int
	for _, catalog := range catalogs {
		// 列出 Schema
		schemas, err := coll.ListSchemas(ctx, catalog.Catalog)
		if err != nil {
			result.Error = fmt.Errorf("列出 Schema 失败: %v", err)
			return result
		}

		for _, schema := range schemas {
			// 列出表
			tableResult, err := coll.ListTables(ctx, catalog.Catalog, schema, &collector.ListOptions{
				PageSize: 100, // 批量获取
			})
			if err != nil {
				result.Error = fmt.Errorf("列出表失败: %v", err)
				return result
			}

			totalTables += len(tableResult.Tables)

			// 可选：获取每个表的详细元数据
			for _, table := range tableResult.Tables {
				// 这里可以获取表的详细元数据
				_, err := coll.FetchTableMetadata(ctx, catalog.Catalog, schema, table)
				if err != nil {
					// 记录错误但继续处理其他表
					log.Printf("获取表 %s.%s.%s 元数据失败: %v", catalog.Catalog, schema, table, err)
				}
			}
		}
	}

	result.Success = true
	result.TableCount = totalTables
	return result
}

// 高级批量采集功能

// BatchCollector 批量采集器
type BatchCollector struct {
	configs     []*config.ConnectorConfig
	concurrency int
	timeout     time.Duration
}

// NewBatchCollector 创建批量采集器
func NewBatchCollector(configs []*config.ConnectorConfig) *BatchCollector {
	return &BatchCollector{
		configs:     configs,
		concurrency: 5,                // 默认并发数
		timeout:     5 * time.Minute,  // 默认超时时间
	}
}

// SetConcurrency 设置并发数
func (bc *BatchCollector) SetConcurrency(concurrency int) *BatchCollector {
	bc.concurrency = concurrency
	return bc
}

// SetTimeout 设置超时时间
func (bc *BatchCollector) SetTimeout(timeout time.Duration) *BatchCollector {
	bc.timeout = timeout
	return bc
}

// CollectAll 采集所有数据源
func (bc *BatchCollector) CollectAll(ctx context.Context) []CollectionResult {
	// 创建带超时的上下文
	ctx, cancel := context.WithTimeout(ctx, bc.timeout)
	defer cancel()

	// 使用信号量控制并发数
	semaphore := make(chan struct{}, bc.concurrency)
	results := make(chan CollectionResult, len(bc.configs))
	var wg sync.WaitGroup

	for _, cfg := range bc.configs {
		wg.Add(1)
		go func(config *config.ConnectorConfig) {
			defer wg.Done()
			
			// 获取信号量
			semaphore <- struct{}{}
			defer func() { <-semaphore }()
			
			result := collectMetadata(ctx, config)
			results <- result
		}(cfg)
	}

	// 等待所有任务完成
	go func() {
		wg.Wait()
		close(results)
	}()

	// 收集结果
	var allResults []CollectionResult
	for result := range results {
		allResults = append(allResults, result)
	}

	return allResults
}