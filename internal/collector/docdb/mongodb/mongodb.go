// Package mongodb provides a MongoDB metadata collector implementation.
package mongodb

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"

	"go-metadata/internal/collector"
	"go-metadata/internal/collector/config"
	"go-metadata/internal/collector/infer"
	"go-metadata/internal/collector/matcher"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	// SourceName identifies this collector type
	SourceName = "mongodb"
	// DefaultPort is the default MongoDB port
	DefaultPort = 27017
	// DefaultTimeout is the default connection timeout in seconds
	DefaultTimeout = 30
	// DefaultSampleSize is the default number of documents to sample for schema inference
	DefaultSampleSize = 100
)

// Collector MongoDB 元数据采集器
type Collector struct {
	config   *config.ConnectorConfig
	client   *mongo.Client
	inferrer *infer.DocumentInferrer
}

// NewCollector 创建 MongoDB 采集器实例
func NewCollector(cfg *config.ConnectorConfig) (collector.Collector, error) {
	if cfg == nil {
		return nil, collector.NewInvalidConfigError(SourceName, "config", "configuration cannot be nil")
	}
	if cfg.Type != "" && cfg.Type != SourceName {
		return nil, collector.NewInvalidConfigError(SourceName, "type", fmt.Sprintf("expected '%s', got '%s'", SourceName, cfg.Type))
	}

	// Initialize schema inferrer with configuration
	var inferrer *infer.DocumentInferrer
	if cfg.Infer != nil {
		inferConfig := &infer.InferConfig{
			Enabled:    cfg.Infer.Enabled,
			SampleSize: cfg.Infer.SampleSize,
			MaxDepth:   cfg.Infer.MaxDepth,
			TypeMerge:  infer.TypeMergeStrategy(cfg.Infer.TypeMerge),
		}
		inferrer = infer.NewDocumentInferrerWithConfig(inferConfig)
	} else {
		inferrer = infer.NewDocumentInferrer()
	}

	return &Collector{
		config:   cfg,
		inferrer: inferrer,
	}, nil
}

// Connect 建立 MongoDB 连接
func (c *Collector) Connect(ctx context.Context) error {
	if c.client != nil {
		return nil // Already connected
	}

	uri, err := c.buildURI()
	if err != nil {
		return collector.NewInvalidConfigError(SourceName, "endpoint", err.Error())
	}

	// Set connection timeout
	timeout := DefaultTimeout
	if c.config.Properties.ConnectionTimeout > 0 {
		timeout = c.config.Properties.ConnectionTimeout
	}

	clientOptions := options.Client().
		ApplyURI(uri).
		SetConnectTimeout(time.Duration(timeout) * time.Second).
		SetServerSelectionTimeout(time.Duration(timeout) * time.Second)

	// Configure connection pool
	if c.config.Properties.MaxOpenConns > 0 {
		clientOptions.SetMaxPoolSize(uint64(c.config.Properties.MaxOpenConns))
	}
	if c.config.Properties.MaxIdleConns > 0 {
		clientOptions.SetMinPoolSize(uint64(c.config.Properties.MaxIdleConns))
	}
	if c.config.Properties.ConnMaxLifetime > 0 {
		clientOptions.SetMaxConnIdleTime(time.Duration(c.config.Properties.ConnMaxLifetime) * time.Second)
	}

	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		return c.wrapConnectionError(err)
	}

	// Test connection with ping
	if err := client.Ping(ctx, nil); err != nil {
		client.Disconnect(ctx)
		return c.wrapConnectionError(err)
	}

	c.client = client
	return nil
}

// Close 关闭 MongoDB 连接
func (c *Collector) Close() error {
	if c.client != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		err := c.client.Disconnect(ctx)
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
	if err := c.client.Ping(ctx, nil); err != nil {
		return &collector.HealthStatus{
			Connected: false,
			Latency:   time.Since(start),
			Message:   err.Error(),
		}, nil
	}

	// Get MongoDB version
	var result bson.M
	err := c.client.Database("admin").RunCommand(ctx, bson.D{{"buildInfo", 1}}).Decode(&result)
	if err != nil {
		return &collector.HealthStatus{
			Connected: true,
			Latency:   time.Since(start),
			Message:   "connected but failed to get version: " + err.Error(),
		}, nil
	}

	version := "unknown"
	if v, ok := result["version"].(string); ok {
		version = v
	}

	return &collector.HealthStatus{
		Connected: true,
		Latency:   time.Since(start),
		Version:   version,
	}, nil
}

// DiscoverCatalogs 发现 Catalog（MongoDB 中 catalog 等同于 MongoDB 实例）
func (c *Collector) DiscoverCatalogs(ctx context.Context) ([]collector.CatalogInfo, error) {
	if c.client == nil {
		return nil, collector.NewConnectionClosedError(SourceName, "discover_catalogs")
	}

	// Get MongoDB version and other info for catalog
	var result bson.M
	err := c.client.Database("admin").RunCommand(ctx, bson.D{{"buildInfo", 1}}).Decode(&result)
	if err != nil {
		return nil, collector.NewQueryError(SourceName, "discover_catalogs", err)
	}

	version := "unknown"
	if v, ok := result["version"].(string); ok {
		version = v
	}

	// MongoDB typically has one catalog per connection
	return []collector.CatalogInfo{
		{
			Catalog:     "mongodb",
			Type:        SourceName,
			Description: "MongoDB Server",
			Properties: map[string]string{
				"version": version,
			},
		},
	}, nil
}

// ListSchemas 列出 Schema（MongoDB 中 schema 等同于 database）
func (c *Collector) ListSchemas(ctx context.Context, catalog string) ([]string, error) {
	if c.client == nil {
		return nil, collector.NewConnectionClosedError(SourceName, "list_schemas")
	}

	// Check context before starting operation
	if err := collector.CheckContext(ctx, SourceName, "list_schemas"); err != nil {
		return nil, err
	}

	databases, err := c.client.ListDatabaseNames(ctx, bson.D{})
	if err != nil {
		if ctx.Err() != nil {
			return nil, collector.WrapContextError(ctx, SourceName, "list_schemas")
		}
		return nil, collector.NewQueryError(SourceName, "list_schemas", err)
	}

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

// ListTables 列出表（MongoDB 中表等同于 collection）
func (c *Collector) ListTables(ctx context.Context, catalog, schema string, opts *collector.ListOptions) (*collector.TableListResult, error) {
	if c.client == nil {
		return nil, collector.NewConnectionClosedError(SourceName, "list_tables")
	}

	// Check context before starting operation
	if err := collector.CheckContext(ctx, SourceName, "list_tables"); err != nil {
		return nil, err
	}

	db := c.client.Database(schema)
	collections, err := db.ListCollectionNames(ctx, bson.D{})
	if err != nil {
		if ctx.Err() != nil {
			return nil, collector.WrapContextError(ctx, SourceName, "list_tables")
		}
		return nil, collector.NewQueryError(SourceName, "list_tables", err)
	}

	// Apply table matching filter
	tables := c.filterTables(collections, opts)

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

// filterTables applies matching rules to filter collections
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

// FetchTableMetadata 获取表元数据（含 Schema 推断）
func (c *Collector) FetchTableMetadata(ctx context.Context, catalog, schema, table string) (*collector.TableMetadata, error) {
	if c.client == nil {
		return nil, collector.NewConnectionClosedError(SourceName, "fetch_table_metadata")
	}

	// Check context before starting operation
	if err := collector.CheckContext(ctx, SourceName, "fetch_table_metadata"); err != nil {
		return nil, err
	}

	db := c.client.Database(schema)
	collection := db.Collection(table)

	// Check if collection exists
	collections, err := db.ListCollectionNames(ctx, bson.D{{"name", table}})
	if err != nil {
		if ctx.Err() != nil {
			return nil, collector.WrapContextError(ctx, SourceName, "fetch_table_metadata")
		}
		return nil, collector.NewQueryError(SourceName, "fetch_table_metadata", err)
	}
	if len(collections) == 0 {
		return nil, collector.NewNotFoundError(SourceName, "fetch_table_metadata", fmt.Sprintf("%s.%s", schema, table), nil)
	}

	metadata := &collector.TableMetadata{
		SourceCategory:  collector.CategoryDocumentDB,
		SourceType:      SourceName,
		Catalog:         catalog,
		Schema:          schema,
		Name:            table,
		Type:            collector.TableTypeCollection,
		LastRefreshedAt: time.Now(),
		InferredSchema:  true, // MongoDB schemas are always inferred
	}

	// Infer schema from sample documents if enabled
	if c.inferrer.GetConfig().Enabled {
		// Check context before schema inference
		if err := collector.CheckContext(ctx, SourceName, "fetch_table_metadata"); err != nil {
			return nil, err
		}

		columns, err := c.inferSchema(ctx, collection)
		if err != nil {
			return nil, err
		}
		metadata.Columns = columns
	}

	// Get indexes if configured
	if c.config.Collect == nil || c.config.Collect.Indexes {
		// Check context before fetching indexes
		if err := collector.CheckContext(ctx, SourceName, "fetch_table_metadata"); err != nil {
			return nil, err
		}

		indexes, err := c.fetchIndexes(ctx, collection)
		if err != nil {
			return nil, err
		}
		metadata.Indexes = indexes
	}

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

	db := c.client.Database(schema)
	collection := db.Collection(table)

	// Get document count
	count, err := collection.EstimatedDocumentCount(ctx)
	if err != nil {
		if ctx.Err() != nil {
			return nil, collector.WrapContextError(ctx, SourceName, "fetch_table_statistics")
		}
		return nil, collector.NewQueryError(SourceName, "fetch_table_statistics", err)
	}

	stats := &collector.TableStatistics{
		RowCount:    count,
		CollectedAt: time.Now(),
	}

	return stats, nil
}

// FetchPartitions 获取分区信息（MongoDB 不支持分区）
func (c *Collector) FetchPartitions(ctx context.Context, catalog, schema, table string) ([]collector.PartitionInfo, error) {
	// MongoDB doesn't support partitions in the traditional sense
	return []collector.PartitionInfo{}, nil
}

// buildURI constructs the MongoDB URI from configuration
func (c *Collector) buildURI() (string, error) {
	endpoint := c.config.Endpoint
	if endpoint == "" {
		return "", fmt.Errorf("endpoint is required")
	}

	// Parse endpoint (expected format: host:port or host)
	host := endpoint
	port := DefaultPort

	if idx := strings.LastIndex(endpoint, ":"); idx != -1 {
		host = endpoint[:idx]
		var err error
		port, err = strconv.Atoi(endpoint[idx+1:])
		if err != nil {
			return "", fmt.Errorf("invalid port in endpoint: %s", endpoint)
		}
	}

	// Build URI
	user := c.config.Credentials.User
	password := c.config.Credentials.Password

	var uri string
	if user != "" && password != "" {
		uri = fmt.Sprintf("mongodb://%s:%s@%s:%d", 
			url.QueryEscape(user), url.QueryEscape(password), host, port)
	} else {
		uri = fmt.Sprintf("mongodb://%s:%d", host, port)
	}

	// Add database from extra properties
	if c.config.Properties.Extra != nil {
		if database := c.config.Properties.Extra["database"]; database != "" {
			uri += "/" + database
		}
	}

	// Add extra parameters
	if c.config.Properties.Extra != nil {
		params := url.Values{}
		for k, v := range c.config.Properties.Extra {
			if k != "database" {
				params.Add(k, v)
			}
		}
		if len(params) > 0 {
			uri += "?" + params.Encode()
		}
	}

	return uri, nil
}

// wrapConnectionError wraps a connection error with appropriate error type
func (c *Collector) wrapConnectionError(err error) error {
	errStr := err.Error()
	if strings.Contains(errStr, "authentication failed") || strings.Contains(errStr, "auth failed") {
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

// fetchIndexes retrieves index information for a collection
func (c *Collector) fetchIndexes(ctx context.Context, collection *mongo.Collection) ([]collector.Index, error) {
	// Check context before starting
	if err := collector.CheckContext(ctx, SourceName, "fetch_indexes"); err != nil {
		return nil, err
	}

	cursor, err := collection.Indexes().List(ctx)
	if err != nil {
		if ctx.Err() != nil {
			return nil, collector.WrapContextError(ctx, SourceName, "fetch_indexes")
		}
		return nil, collector.NewQueryError(SourceName, "fetch_indexes", err)
	}
	defer cursor.Close(ctx)

	var indexes []collector.Index
	for cursor.Next(ctx) {
		// Check context during iteration
		if err := collector.CheckContext(ctx, SourceName, "fetch_indexes"); err != nil {
			return nil, err
		}

		var indexDoc bson.M
		if err := cursor.Decode(&indexDoc); err != nil {
			return nil, collector.NewParseError(SourceName, "fetch_indexes", err)
		}

		// Extract index information
		name, _ := indexDoc["name"].(string)
		if name == "" {
			continue
		}

		// Parse key specification
		var columns []string
		if keySpec, ok := indexDoc["key"].(bson.M); ok {
			for field := range keySpec {
				columns = append(columns, field)
			}
		}

		// Check if index is unique
		unique := false
		if uniqueVal, ok := indexDoc["unique"].(bool); ok {
			unique = uniqueVal
		}

		// Get index type (default is btree for MongoDB)
		indexType := "btree"
		if typeVal, ok := indexDoc["type"].(string); ok {
			indexType = typeVal
		}

		index := collector.Index{
			Name:    name,
			Columns: columns,
			Unique:  unique,
			Type:    indexType,
		}

		indexes = append(indexes, index)
	}

	if err := cursor.Err(); err != nil {
		if ctx.Err() != nil {
			return nil, collector.WrapContextError(ctx, SourceName, "fetch_indexes")
		}
		return nil, collector.NewQueryError(SourceName, "fetch_indexes", err)
	}

	return indexes, nil
}

// Category 返回数据源类别
func (c *Collector) Category() collector.DataSourceCategory {
	return collector.CategoryDocumentDB
}

// Type 返回数据源类型
func (c *Collector) Type() string {
	return SourceName
}

// Ensure Collector implements collector.Collector interface
var _ collector.Collector = (*Collector)(nil)