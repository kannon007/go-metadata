// Package redis provides a Redis metadata collector implementation.
package redis

import (
	"context"
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"

	"go-metadata/internal/collector"
	"go-metadata/internal/collector/config"
	"go-metadata/internal/collector/matcher"

	"github.com/redis/go-redis/v9"
)

const (
	// SourceName identifies this collector type
	SourceName = "redis"
	// DefaultPort is the default Redis port
	DefaultPort = 6379
	// DefaultTimeout is the default connection timeout in seconds
	DefaultTimeout = 30
	// DefaultDatabase is the default Redis database
	DefaultDatabase = 0
	// DefaultScanCount is the default SCAN count parameter
	DefaultScanCount = 100
)

// Collector Redis 元数据采集器
type Collector struct {
	config *config.ConnectorConfig
	client *redis.Client
}

// NewCollector 创建 Redis 采集器实例
func NewCollector(cfg *config.ConnectorConfig) (collector.Collector, error) {
	if cfg == nil {
		return nil, collector.NewInvalidConfigError(SourceName, "config", "configuration cannot be nil")
	}
	if cfg.Type != "" && cfg.Type != SourceName {
		return nil, collector.NewInvalidConfigError(SourceName, "type", fmt.Sprintf("expected '%s', got '%s'", SourceName, cfg.Type))
	}

	return &Collector{
		config: cfg,
	}, nil
}

// Connect 建立 Redis 连接
func (c *Collector) Connect(ctx context.Context) error {
	if c.client != nil {
		return nil // Already connected
	}

	// Parse endpoint
	host, port, err := c.parseEndpoint()
	if err != nil {
		return collector.NewInvalidConfigError(SourceName, "endpoint", err.Error())
	}

	// Set connection timeout
	timeout := DefaultTimeout
	if c.config.Properties.ConnectionTimeout > 0 {
		timeout = c.config.Properties.ConnectionTimeout
	}

	// Get database number
	database := DefaultDatabase
	if c.config.Properties.Extra != nil {
		if dbStr := c.config.Properties.Extra["database"]; dbStr != "" {
			if db, err := strconv.Atoi(dbStr); err == nil {
				database = db
			}
		}
	}

	// Create Redis client options
	opts := &redis.Options{
		Addr:         net.JoinHostPort(host, strconv.Itoa(port)),
		Password:     c.config.Credentials.Password,
		DB:           database,
		DialTimeout:  time.Duration(timeout) * time.Second,
		ReadTimeout:  time.Duration(timeout) * time.Second,
		WriteTimeout: time.Duration(timeout) * time.Second,
	}

	// Configure connection pool
	if c.config.Properties.MaxOpenConns > 0 {
		opts.PoolSize = c.config.Properties.MaxOpenConns
	}
	if c.config.Properties.MaxIdleConns > 0 {
		opts.MinIdleConns = c.config.Properties.MaxIdleConns
	}
	if c.config.Properties.ConnMaxLifetime > 0 {
		opts.ConnMaxLifetime = time.Duration(c.config.Properties.ConnMaxLifetime) * time.Second
	}

	// Add username if provided
	if c.config.Credentials.User != "" {
		opts.Username = c.config.Credentials.User
	}

	client := redis.NewClient(opts)

	// Test connection with ping
	if err := client.Ping(ctx).Err(); err != nil {
		client.Close()
		return c.wrapConnectionError(err)
	}

	c.client = client
	return nil
}

// Close 关闭 Redis 连接
func (c *Collector) Close() error {
	if c.client != nil {
		err := c.client.Close()
		c.client = nil
		return err
	}
	return nil
}

// HealthCheck 健康检查
func (c *Collector) HealthCheck(ctx context.Context) (*collector.HealthStatus, error) {
	if c.client == nil {
		return &collector.HealthStatus{
			Connected: false,
			Message:   "not connected",
		}, nil
	}

	start := time.Now()

	// Ping to check connection
	if err := c.client.Ping(ctx).Err(); err != nil {
		return &collector.HealthStatus{
			Connected: false,
			Latency:   time.Since(start),
			Message:   err.Error(),
		}, nil
	}

	// Get Redis version
	info, err := c.client.Info(ctx, "server").Result()
	if err != nil {
		return &collector.HealthStatus{
			Connected: true,
			Latency:   time.Since(start),
			Message:   "connected but failed to get version: " + err.Error(),
		}, nil
	}

	version := c.parseRedisVersion(info)

	return &collector.HealthStatus{
		Connected: true,
		Latency:   time.Since(start),
		Version:   version,
	}, nil
}

// DiscoverCatalogs 发现 Catalog（Redis 中 catalog 等同于 Redis 实例）
func (c *Collector) DiscoverCatalogs(ctx context.Context) ([]collector.CatalogInfo, error) {
	if c.client == nil {
		return nil, collector.NewConnectionClosedError(SourceName, "discover_catalogs")
	}

	// Get Redis server info
	info, err := c.client.Info(ctx, "server").Result()
	if err != nil {
		return nil, collector.NewQueryError(SourceName, "discover_catalogs", err)
	}

	version := c.parseRedisVersion(info)

	// Redis typically has one catalog per connection
	return []collector.CatalogInfo{
		{
			Catalog:     "redis",
			Type:        SourceName,
			Description: "Redis Server",
			Properties: map[string]string{
				"version": version,
			},
		},
	}, nil
}

// ListSchemas 列出 Schema（Redis 中 schema 等同于 database）
func (c *Collector) ListSchemas(ctx context.Context, catalog string) ([]string, error) {
	if c.client == nil {
		return nil, collector.NewConnectionClosedError(SourceName, "list_schemas")
	}

	// Check context before starting operation
	if err := collector.CheckContext(ctx, SourceName, "list_schemas"); err != nil {
		return nil, err
	}

	// Get keyspace info to find databases with keys
	info, err := c.client.Info(ctx, "keyspace").Result()
	if err != nil {
		if ctx.Err() != nil {
			return nil, collector.WrapContextError(ctx, SourceName, "list_schemas")
		}
		return nil, collector.NewQueryError(SourceName, "list_schemas", err)
	}

	databases := c.parseKeyspaceInfo(info)

	// Apply database matching filter if configured
	if c.config.Matching != nil && c.config.Matching.Databases != nil {
		ruleMatcher, err := matcher.NewRuleMatcher(
			c.config.Matching.Databases,
			c.config.Matching.PatternType,
			c.config.Matching.CaseSensitive,
		)
		if err != nil {
			return nil, collector.NewInvalidConfigError(SourceName, "matching.databases", err.Error())
		}
		var filtered []string
		for _, db := range databases {
			if ruleMatcher.Match(db) {
				filtered = append(filtered, db)
			}
		}
		databases = filtered
	}

	return databases, nil
}

// ListTables 列出表（Redis 中表等同于 key pattern）
func (c *Collector) ListTables(ctx context.Context, catalog, schema string, opts *collector.ListOptions) (*collector.TableListResult, error) {
	if c.client == nil {
		return nil, collector.NewConnectionClosedError(SourceName, "list_tables")
	}

	// Check context before starting operation
	if err := collector.CheckContext(ctx, SourceName, "list_tables"); err != nil {
		return nil, err
	}

	// Parse database number
	database, err := strconv.Atoi(schema)
	if err != nil {
		return nil, collector.NewInvalidConfigError(SourceName, "schema", fmt.Sprintf("invalid database number: %s", schema))
	}

	// Switch to the specified database
	originalDB := c.client.Options().DB
	if database != originalDB {
		// Create a new client for the specific database
		opts := *c.client.Options()
		opts.DB = database
		dbClient := redis.NewClient(&opts)
		defer dbClient.Close()
		c.client = dbClient
	}

	// Scan for key patterns
	patterns, err := c.scanKeyPatterns(ctx, database, "*", 1000)
	if err != nil {
		return nil, err
	}

	// Convert patterns to table names
	var tables []string
	for _, pattern := range patterns {
		tables = append(tables, pattern.Pattern)
	}

	// Apply table matching filter
	tables = c.filterTables(tables, opts)

	// Apply pagination
	result := &collector.TableListResult{
		TotalCount: len(tables),
	}

	if opts != nil && opts.PageSize > 0 {
		startIdx := 0
		if opts.PageToken != "" {
			startIdx, _ = strconv.Atoi(opts.PageToken)
		}

		endIdx := startIdx + opts.PageSize
		if endIdx > len(tables) {
			endIdx = len(tables)
		}

		if startIdx < len(tables) {
			result.Tables = tables[startIdx:endIdx]
			if endIdx < len(tables) {
				result.NextPageToken = strconv.Itoa(endIdx)
			}
		}
	} else {
		result.Tables = tables
	}

	return result, nil
}

// FetchTableMetadata 获取表元数据（Redis key pattern 元数据）
func (c *Collector) FetchTableMetadata(ctx context.Context, catalog, schema, table string) (*collector.TableMetadata, error) {
	if c.client == nil {
		return nil, collector.NewConnectionClosedError(SourceName, "fetch_table_metadata")
	}

	// Check context before starting operation
	if err := collector.CheckContext(ctx, SourceName, "fetch_table_metadata"); err != nil {
		return nil, err
	}

	// Parse database number
	database, err := strconv.Atoi(schema)
	if err != nil {
		return nil, collector.NewInvalidConfigError(SourceName, "schema", fmt.Sprintf("invalid database number: %s", schema))
	}

	metadata := &collector.TableMetadata{
		SourceCategory:  collector.CategoryKeyValue,
		SourceType:      SourceName,
		Catalog:         catalog,
		Schema:          schema,
		Name:            table,
		Type:            collector.TableTypeKeySpace,
		LastRefreshedAt: time.Now(),
		InferredSchema:  true, // Redis schemas are always inferred
	}

	// Get key pattern information
	patterns, err := c.scanKeyPatterns(ctx, database, table, 100)
	if err != nil {
		return nil, err
	}

	// Convert patterns to columns (representing key structure)
	var columns []collector.Column
	for i, pattern := range patterns {
		column := collector.Column{
			OrdinalPosition: i + 1,
			Name:            pattern.Pattern,
			Type:            pattern.KeyType,
			SourceType:      pattern.KeyType,
			Nullable:        false,
		}
		columns = append(columns, column)
	}

	metadata.Columns = columns

	return metadata, nil
}

// FetchTableStatistics 获取表统计信息
func (c *Collector) FetchTableStatistics(ctx context.Context, catalog, schema, table string) (*collector.TableStatistics, error) {
	if c.client == nil {
		return nil, collector.NewConnectionClosedError(SourceName, "fetch_table_statistics")
	}

	// Check context before starting operation
	if err := collector.CheckContext(ctx, SourceName, "fetch_table_statistics"); err != nil {
		return nil, err
	}

	// Parse database number
	database, err := strconv.Atoi(schema)
	if err != nil {
		return nil, collector.NewInvalidConfigError(SourceName, "schema", fmt.Sprintf("invalid database number: %s", schema))
	}

	// Get key count for the pattern
	count, err := c.countKeysForPattern(ctx, database, table)
	if err != nil {
		return nil, err
	}

	stats := &collector.TableStatistics{
		RowCount:    count,
		CollectedAt: time.Now(),
	}

	return stats, nil
}

// FetchPartitions 获取分区信息（Redis 不支持分区）
func (c *Collector) FetchPartitions(ctx context.Context, catalog, schema, table string) ([]collector.PartitionInfo, error) {
	// Redis doesn't support partitions in the traditional sense
	return []collector.PartitionInfo{}, nil
}

// Category 返回数据源类别
func (c *Collector) Category() collector.DataSourceCategory {
	return collector.CategoryKeyValue
}

// Type 返回数据源类型
func (c *Collector) Type() string {
	return SourceName
}

// parseEndpoint parses the endpoint configuration
func (c *Collector) parseEndpoint() (string, int, error) {
	endpoint := c.config.Endpoint
	if endpoint == "" {
		return "", 0, fmt.Errorf("endpoint is required")
	}

	// Parse endpoint (expected format: host:port or host)
	host := endpoint
	port := DefaultPort

	if idx := strings.LastIndex(endpoint, ":"); idx != -1 {
		host = endpoint[:idx]
		var err error
		port, err = strconv.Atoi(endpoint[idx+1:])
		if err != nil {
			return "", 0, fmt.Errorf("invalid port in endpoint: %s", endpoint)
		}
	}

	return host, port, nil
}

// wrapConnectionError wraps a connection error with appropriate error type
func (c *Collector) wrapConnectionError(err error) error {
	errStr := err.Error()
	if strings.Contains(errStr, "NOAUTH") || strings.Contains(errStr, "invalid password") {
		return collector.NewAuthError(SourceName, "connect", err)
	}
	if strings.Contains(errStr, "connection refused") || strings.Contains(errStr, "no such host") {
		return collector.NewNetworkError(SourceName, "connect", err)
	}
	if strings.Contains(errStr, "timeout") || strings.Contains(errStr, "deadline exceeded") {
		return collector.NewTimeoutError(SourceName, "connect", err)
	}
	return collector.NewNetworkError(SourceName, "connect", err)
}

// parseRedisVersion extracts Redis version from INFO server output
func (c *Collector) parseRedisVersion(info string) string {
	lines := strings.Split(info, "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "redis_version:") {
			return strings.TrimSpace(strings.TrimPrefix(line, "redis_version:"))
		}
	}
	return "unknown"
}

// parseKeyspaceInfo extracts database numbers from INFO keyspace output
func (c *Collector) parseKeyspaceInfo(info string) []string {
	var databases []string
	lines := strings.Split(info, "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "db") {
			// Extract database number from "db0:keys=1,expires=0,avg_ttl=0"
			if idx := strings.Index(line, ":"); idx != -1 {
				dbNum := strings.TrimPrefix(line[:idx], "db")
				databases = append(databases, dbNum)
			}
		}
	}
	
	// If no databases found in keyspace info, return default database
	if len(databases) == 0 {
		databases = append(databases, "0")
	}
	
	return databases
}

// filterTables applies matching rules to filter key patterns
func (c *Collector) filterTables(tables []string, opts *collector.ListOptions) []string {
	// First apply config-level table matching
	if c.config.Matching != nil && c.config.Matching.Tables != nil {
		ruleMatcher, err := matcher.NewRuleMatcher(
			c.config.Matching.Tables,
			c.config.Matching.PatternType,
			c.config.Matching.CaseSensitive,
		)
		if err == nil {
			var filtered []string
			for _, t := range tables {
				if ruleMatcher.Match(t) {
					filtered = append(filtered, t)
				}
			}
			tables = filtered
		}
	}

	// Then apply request-level filter
	if opts != nil && opts.Filter != nil {
		patternType := "glob"
		caseSensitive := false
		if c.config.Matching != nil {
			patternType = c.config.Matching.PatternType
			caseSensitive = c.config.Matching.CaseSensitive
		}

		ruleMatcher, err := matcher.NewRuleMatcher(
			&config.MatchingRule{
				Include: opts.Filter.Include,
				Exclude: opts.Filter.Exclude,
			},
			patternType,
			caseSensitive,
		)
		if err == nil {
			var filtered []string
			for _, t := range tables {
				if ruleMatcher.Match(t) {
					filtered = append(filtered, t)
				}
			}
			tables = filtered
		}
	}

	return tables
}

// Ensure Collector implements collector.Collector interface
var _ collector.Collector = (*Collector)(nil)