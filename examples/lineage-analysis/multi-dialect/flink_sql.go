// Flink SQL 血缘分析样例
package main

import (
	"encoding/json"
	"fmt"
	"log"

	"go-metadata/internal/lineage"
	"go-metadata/internal/lineage/metadata"
)

func main() {
	fmt.Println("=== Flink SQL 血缘分析样例 ===")

	// 1. 从 Flink DDL 构建元数据
	fmt.Println("从 Flink DDL 构建元数据...")
	flinkDDL := `
		-- 用户事件流表
		CREATE TABLE user_events (
		    user_id BIGINT,
		    event_type STRING,
		    event_time TIMESTAMP(3),
		    page_url STRING,
		    session_id STRING,
		    user_agent STRING,
		    ip_address STRING,
		    WATERMARK FOR event_time AS event_time - INTERVAL '5' SECOND
		) WITH (
		    'connector' = 'kafka',
		    'topic' = 'user_events',
		    'properties.bootstrap.servers' = 'localhost:9092',
		    'format' = 'json'
		);

		-- 用户维度表
		CREATE TABLE user_info (
		    user_id BIGINT,
		    user_name STRING,
		    email STRING,
		    city STRING,
		    country STRING,
		    registration_date DATE,
		    PRIMARY KEY (user_id) NOT ENFORCED
		) WITH (
		    'connector' = 'jdbc',
		    'url' = 'jdbc:mysql://localhost:3306/userdb',
		    'table-name' = 'users'
		);

		-- 页面维度表
		CREATE TABLE page_info (
		    page_url STRING,
		    page_title STRING,
		    page_category STRING,
		    page_type STRING,
		    PRIMARY KEY (page_url) NOT ENFORCED
		) WITH (
		    'connector' = 'jdbc',
		    'url' = 'jdbc:mysql://localhost:3306/contentdb',
		    'table-name' = 'pages'
		);

		-- 用户会话汇总表
		CREATE TABLE user_session_summary (
		    user_id BIGINT,
		    session_id STRING,
		    session_start_time TIMESTAMP(3),
		    session_end_time TIMESTAMP(3),
		    page_views INT,
		    unique_pages INT,
		    session_duration_seconds BIGINT,
		    PRIMARY KEY (user_id, session_id) NOT ENFORCED
		) WITH (
		    'connector' = 'jdbc',
		    'url' = 'jdbc:mysql://localhost:3306/analyticsdb',
		    'table-name' = 'session_summary'
		);
	`

	// 使用 DDL 构建分析器
	analyzer := metadata.NewMetadataBuilder().
		LoadFromDDL(flinkDDL).
		BuildAnalyzer()

	// 2. 分析 Flink 流处理查询
	fmt.Println("\n=== 分析 Flink 流处理查询 ===")
	streamSQL := `
		SELECT user_id,
		       event_type,
		       event_time,
		       page_url,
		       session_id,
		       COUNT(*) OVER (
		           PARTITION BY user_id, session_id 
		           ORDER BY event_time 
		           RANGE BETWEEN INTERVAL '1' HOUR PRECEDING AND CURRENT ROW
		       ) AS events_in_session,
		       LAG(event_time, 1) OVER (
		           PARTITION BY user_id 
		           ORDER BY event_time
		       ) AS prev_event_time,
		       event_time - LAG(event_time, 1) OVER (
		           PARTITION BY user_id 
		           ORDER BY event_time
		       ) AS time_since_last_event
		FROM user_events
		WHERE event_type IN ('page_view', 'click', 'scroll')
	`
	
	result, err := analyzer.Analyze(streamSQL)
	if err != nil {
		log.Fatalf("分析 Flink 流处理 SQL 失败: %v", err)
	}

	fmt.Printf("SQL: %s\n", streamSQL)
	printLineageResult(result)

	// 3. 分析 Flink 时间窗口聚合
	fmt.Println("\n=== 分析 Flink 时间窗口聚合 ===")
	windowAggSQL := `
		SELECT user_id,
		       TUMBLE_START(event_time, INTERVAL '1' HOUR) AS window_start,
		       TUMBLE_END(event_time, INTERVAL '1' HOUR) AS window_end,
		       COUNT(*) AS event_count,
		       COUNT(DISTINCT page_url) AS unique_pages,
		       COUNT(DISTINCT session_id) AS unique_sessions,
		       COLLECT(DISTINCT event_type) AS event_types,
		       MIN(event_time) AS first_event_time,
		       MAX(event_time) AS last_event_time
		FROM user_events
		WHERE event_time > CURRENT_TIMESTAMP - INTERVAL '24' HOUR
		GROUP BY user_id, TUMBLE(event_time, INTERVAL '1' HOUR)
		HAVING COUNT(*) > 10
	`
	
	result, err = analyzer.Analyze(windowAggSQL)
	if err != nil {
		log.Fatalf("分析 Flink 窗口聚合 SQL 失败: %v", err)
	}

	fmt.Printf("SQL: %s\n", windowAggSQL)
	printLineageResult(result)

	// 4. 分析 Flink 流表 JOIN
	fmt.Println("\n=== 分析 Flink 流表 JOIN ===")
	streamJoinSQL := `
		SELECT e.user_id,
		       u.user_name,
		       u.email,
		       u.city,
		       u.country,
		       e.event_type,
		       e.event_time,
		       e.page_url,
		       p.page_title,
		       p.page_category,
		       p.page_type,
		       e.session_id,
		       CASE 
		           WHEN p.page_category = 'product' THEN 'Product View'
		           WHEN p.page_category = 'checkout' THEN 'Purchase Intent'
		           WHEN p.page_category = 'support' THEN 'Support Request'
		           ELSE 'General Browsing'
		       END AS interaction_type
		FROM user_events e
		LEFT JOIN user_info FOR SYSTEM_TIME AS OF e.event_time AS u
		    ON e.user_id = u.user_id
		LEFT JOIN page_info FOR SYSTEM_TIME AS OF e.event_time AS p
		    ON e.page_url = p.page_url
		WHERE e.event_time > CURRENT_TIMESTAMP - INTERVAL '1' DAY
		  AND e.event_type = 'page_view'
	`
	
	result, err = analyzer.Analyze(streamJoinSQL)
	if err != nil {
		log.Fatalf("分析 Flink 流表 JOIN SQL 失败: %v", err)
	}

	fmt.Printf("SQL: %s\n", streamJoinSQL)
	printLineageResult(result)

	// 5. 分析 Flink INSERT INTO 语句
	fmt.Println("\n=== 分析 Flink INSERT INTO 语句 ===")
	insertSQL := `
		INSERT INTO user_session_summary
		SELECT user_id,
		       session_id,
		       MIN(event_time) AS session_start_time,
		       MAX(event_time) AS session_end_time,
		       COUNT(*) AS page_views,
		       COUNT(DISTINCT page_url) AS unique_pages,
		       EXTRACT(EPOCH FROM (MAX(event_time) - MIN(event_time))) AS session_duration_seconds
		FROM user_events
		WHERE event_type = 'page_view'
		  AND event_time > CURRENT_TIMESTAMP - INTERVAL '1' HOUR
		GROUP BY user_id, session_id
		HAVING COUNT(*) >= 2  -- 至少2个页面浏览才算有效会话
	`
	
	result, err = analyzer.Analyze(insertSQL)
	if err != nil {
		log.Fatalf("分析 Flink INSERT SQL 失败: %v", err)
	}

	fmt.Printf("SQL: %s\n", insertSQL)
	printLineageResult(result)

	// 6. 分析 Flink CEP (Complex Event Processing) 模式查询
	fmt.Println("\n=== 分析 Flink CEP 模式查询 ===")
	cepSQL := `
		SELECT user_id,
		       session_id,
		       start_event.event_time AS funnel_start_time,
		       end_event.event_time AS funnel_end_time,
		       start_event.page_url AS entry_page,
		       end_event.page_url AS exit_page,
		       EXTRACT(EPOCH FROM (end_event.event_time - start_event.event_time)) AS funnel_duration_seconds
		FROM user_events
		MATCH_RECOGNIZE (
		    PARTITION BY user_id, session_id
		    ORDER BY event_time
		    MEASURES
		        FIRST(A.event_time) AS start_event,
		        LAST(C.event_time) AS end_event
		    PATTERN (A B* C)
		    DEFINE
		        A AS A.page_url LIKE '%/product/%',
		        B AS B.event_type = 'page_view',
		        C AS C.page_url LIKE '%/checkout%'
		) AS T
		WHERE funnel_duration_seconds BETWEEN 60 AND 3600  -- 1分钟到1小时的转化路径
	`
	
	result, err = analyzer.Analyze(cepSQL)
	if err != nil {
		log.Printf("分析 Flink CEP SQL 失败（可能不支持 MATCH_RECOGNIZE）: %v", err)
	} else {
		fmt.Printf("SQL: %s\n", cepSQL)
		printLineageResult(result)
	}

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

	// 分析 Flink 特有的特性
	analyzeFlinkFeatures(result)
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

// analyzeFlinkFeatures 分析 Flink 特有的特性
func analyzeFlinkFeatures(result *lineage.LineageResult) {
	fmt.Println("\nFlink 特性分析:")
	
	// 检查时间相关的操作
	timeOperations := 0
	windowOperations := 0
	watermarkOperations := 0
	
	for _, col := range result.Columns {
		for _, op := range col.Operators {
			if containsTimeFunction(op) {
				timeOperations++
			}
			if containsWindowFunction(op) {
				windowOperations++
			}
			if containsWatermarkFunction(op) {
				watermarkOperations++
			}
		}
	}
	
	fmt.Printf("  时间函数操作: %d\n", timeOperations)
	fmt.Printf("  窗口函数操作: %d\n", windowOperations)
	fmt.Printf("  水印相关操作: %d\n", watermarkOperations)
	
	// 检查流处理特性
	streamFeatures := analyzeStreamFeatures(result)
	if len(streamFeatures) > 0 {
		fmt.Printf("  流处理特性: %v\n", streamFeatures)
	}
	
	// 输出 JSON 格式
	if jsonBytes, err := json.MarshalIndent(result, "", "  "); err == nil {
		fmt.Printf("\nJSON 格式:\n%s\n", string(jsonBytes))
	}
}

// containsTimeFunction 检查是否包含时间函数
func containsTimeFunction(operation string) bool {
	timeFunctions := []string{
		"CURRENT_TIMESTAMP", "NOW()", "event_time", "WATERMARK",
		"INTERVAL", "EXTRACT", "DATE_FORMAT", "TIMESTAMP",
	}
	
	for _, timeFunc := range timeFunctions {
		if contains(operation, timeFunc) {
			return true
		}
	}
	return false
}

// containsWindowFunction 检查是否包含窗口函数
func containsWindowFunction(operation string) bool {
	windowFunctions := []string{
		"TUMBLE", "HOP", "SESSION", "OVER", "ROW_NUMBER",
		"RANK", "DENSE_RANK", "LAG", "LEAD",
	}
	
	for _, windowFunc := range windowFunctions {
		if contains(operation, windowFunc) {
			return true
		}
	}
	return false
}

// containsWatermarkFunction 检查是否包含水印相关函数
func containsWatermarkFunction(operation string) bool {
	watermarkFunctions := []string{
		"WATERMARK", "FOR SYSTEM_TIME AS OF",
	}
	
	for _, watermarkFunc := range watermarkFunctions {
		if contains(operation, watermarkFunc) {
			return true
		}
	}
	return false
}

// analyzeStreamFeatures 分析流处理特性
func analyzeStreamFeatures(result *lineage.LineageResult) []string {
	features := []string{}
	
	// 检查是否有流表特征
	hasStreamTables := false
	hasDimensionTables := false
	
	for _, col := range result.Columns {
		for _, source := range col.Sources {
			if source.Table == "user_events" {
				hasStreamTables = true
			}
			if source.Table == "user_info" || source.Table == "page_info" {
				hasDimensionTables = true
			}
		}
	}
	
	if hasStreamTables {
		features = append(features, "流表处理")
	}
	if hasDimensionTables {
		features = append(features, "维度表关联")
	}
	
	return features
}

// contains 检查字符串是否包含子字符串（忽略大小写）
func contains(s, substr string) bool {
	return len(s) >= len(substr) && 
		   (s == substr || 
		    len(s) > len(substr) && 
		    (s[:len(substr)] == substr || 
		     s[len(s)-len(substr):] == substr ||
		     containsSubstring(s, substr)))
}

// containsSubstring 辅助函数检查子字符串
func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}