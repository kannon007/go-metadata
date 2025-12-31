// Package elasticsearch provides an Elasticsearch metadata collector implementation.
package elasticsearch

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"go-metadata/internal/collector"
	"go-metadata/internal/collector/config"
	"go-metadata/internal/collector/infer"
	"go-metadata/internal/collector/matcher"

	"github.com/elastic/go-elasticsearch/v8"
)

const (
	// SourceName identifies this collector type
	SourceName = "elasticsearch"
	// DefaultPort is the default Elasticsearch port
	DefaultPort = 9200
	// DefaultTimeout is the default connection timeout in seconds
	DefaultTimeout = 30
	// DefaultSampleSize is the default number of documents to sample for schema inference
	DefaultSampleSize = 100
)

// Collector Elasticsearch 元数据采集器
type Collector struct {
	config   *config.ConnectorConfig
	client   *elasticsearch.Client
	inferrer *infer.DocumentInferrer
}

// NewCollector 创建 Elasticsearch 采集器实例
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

// Connect 建立 Elasticsearch 连接
func (c *Collector) Connect(ctx context.Context) error {
	if c.client != nil {
		return nil // Already connected
	}

	addresses, err := c.buildAddresses()
	if err != nil {
		return collector.NewInvalidConfigError(SourceName, "endpoint", err.Error())
	}

	// Set connection timeout
	timeout := DefaultTimeout
	if c.config.Properties.ConnectionTimeout > 0 {
		timeout = c.config.Properties.ConnectionTimeout
	}

	// Build Elasticsearch configuration
	cfg := elasticsearch.Config{
		Addresses: addresses,
		Transport: &http.Transport{
			ResponseHeaderTimeout: time.Duration(timeout) * time.Second,
		},
	}

	// Add authentication if provided
	if c.config.Credentials.User != "" && c.config.Credentials.Password != "" {
		cfg.Username = c.config.Credentials.User
		cfg.Password = c.config.Credentials.Password
	}

	// Add API key if provided in extra properties
	if c.config.Properties.Extra != nil {
		if apiKey := c.config.Properties.Extra["api_key"]; apiKey != "" {
			cfg.APIKey = apiKey
		}
		if cloudID := c.config.Properties.Extra["cloud_id"]; cloudID != "" {
			cfg.CloudID = cloudID
		}
	}

	client, err := elasticsearch.NewClient(cfg)
	if err != nil {
		return c.wrapConnectionError(err)
	}

	// Test connection with cluster info
	res, err := client.Info(client.Info.WithContext(ctx))
	if err != nil {
		return c.wrapConnectionError(err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return c.wrapConnectionError(fmt.Errorf("elasticsearch connection failed: %s", res.Status()))
	}

	c.client = client
	return nil
}

// Close 关闭 Elasticsearch 连接
func (c *Collector) Close() error {
	if c.client != nil {
		c.client = nil
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

	// Get cluster health
	res, err := c.client.Cluster.Health(
		c.client.Cluster.Health.WithContext(ctx),
		c.client.Cluster.Health.WithLevel("cluster"),
	)
	if err != nil {
		return &collector.HealthStatus{
			Connected: false,
			Latency:   time.Since(start),
			Message:   err.Error(),
		}, nil
	}
	defer res.Body.Close()

	if res.IsError() {
		return &collector.HealthStatus{
			Connected: false,
			Latency:   time.Since(start),
			Message:   fmt.Sprintf("cluster health check failed: %s", res.Status()),
		}, nil
	}

	// Get cluster info for version
	infoRes, err := c.client.Info(c.client.Info.WithContext(ctx))
	if err != nil {
		return &collector.HealthStatus{
			Connected: true,
			Latency:   time.Since(start),
			Message:   "connected but failed to get version: " + err.Error(),
		}, nil
	}
	defer infoRes.Body.Close()

	version := "unknown"
	if !infoRes.IsError() {
		var info map[string]interface{}
		if err := json.NewDecoder(infoRes.Body).Decode(&info); err == nil {
			if versionInfo, ok := info["version"].(map[string]interface{}); ok {
				if v, ok := versionInfo["number"].(string); ok {
					version = v
				}
			}
		}
	}

	return &collector.HealthStatus{
		Connected: true,
		Latency:   time.Since(start),
		Version:   version,
	}, nil
}

// DiscoverCatalogs 发现 Catalog（Elasticsearch 中 catalog 等同于 Elasticsearch 集群）
func (c *Collector) DiscoverCatalogs(ctx context.Context) ([]collector.CatalogInfo, error) {
	if c.client == nil {
		return nil, collector.NewConnectionClosedError(SourceName, "discover_catalogs")
	}

	// Get cluster info
	res, err := c.client.Info(c.client.Info.WithContext(ctx))
	if err != nil {
		return nil, collector.NewQueryError(SourceName, "discover_catalogs", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return nil, collector.NewQueryError(SourceName, "discover_catalogs", fmt.Errorf("failed to get cluster info: %s", res.Status()))
	}

	var info map[string]interface{}
	if err := json.NewDecoder(res.Body).Decode(&info); err != nil {
		return nil, collector.NewParseError(SourceName, "discover_catalogs", err)
	}

	clusterName := "elasticsearch"
	version := "unknown"

	if name, ok := info["cluster_name"].(string); ok {
		clusterName = name
	}
	if versionInfo, ok := info["version"].(map[string]interface{}); ok {
		if v, ok := versionInfo["number"].(string); ok {
			version = v
		}
	}

	return []collector.CatalogInfo{
		{
			Catalog:     clusterName,
			Type:        SourceName,
			Description: "Elasticsearch Cluster",
			Properties: map[string]string{
				"version": version,
			},
		},
	}, nil
}

// ListSchemas 列出 Schema（Elasticsearch 中没有 schema 概念，返回空列表）
func (c *Collector) ListSchemas(ctx context.Context, catalog string) ([]string, error) {
	if c.client == nil {
		return nil, collector.NewConnectionClosedError(SourceName, "list_schemas")
	}

	// Elasticsearch doesn't have schemas, return empty list
	return []string{}, nil
}

// ListTables 列出表（Elasticsearch 中表等同于 index）
func (c *Collector) ListTables(ctx context.Context, catalog, schema string, opts *collector.ListOptions) (*collector.TableListResult, error) {
	if c.client == nil {
		return nil, collector.NewConnectionClosedError(SourceName, "list_tables")
	}

	// Check context before starting operation
	if err := collector.CheckContext(ctx, SourceName, "list_tables"); err != nil {
		return nil, err
	}

	// Get indices using _cat/indices API
	res, err := c.client.Cat.Indices(
		c.client.Cat.Indices.WithContext(ctx),
		c.client.Cat.Indices.WithFormat("json"),
		c.client.Cat.Indices.WithH("index"),
	)
	if err != nil {
		if ctx.Err() != nil {
			return nil, collector.WrapContextError(ctx, SourceName, "list_tables")
		}
		return nil, collector.NewQueryError(SourceName, "list_tables", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return nil, collector.NewQueryError(SourceName, "list_tables", fmt.Errorf("failed to list indices: %s", res.Status()))
	}

	var indices []map[string]interface{}
	if err := json.NewDecoder(res.Body).Decode(&indices); err != nil {
		return nil, collector.NewParseError(SourceName, "list_tables", err)
	}

	var indexNames []string
	for _, index := range indices {
		if name, ok := index["index"].(string); ok {
			// Skip system indices (starting with .)
			if !strings.HasPrefix(name, ".") {
				indexNames = append(indexNames, name)
			}
		}
	}

	// Apply table matching filter
	tables := c.filterTables(indexNames, opts)

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

// filterTables applies matching rules to filter indices
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

	// Check if index exists
	res, err := c.client.Indices.Exists([]string{table}, c.client.Indices.Exists.WithContext(ctx))
	if err != nil {
		if ctx.Err() != nil {
			return nil, collector.WrapContextError(ctx, SourceName, "fetch_table_metadata")
		}
		return nil, collector.NewQueryError(SourceName, "fetch_table_metadata", err)
	}
	defer res.Body.Close()

	if res.StatusCode == 404 {
		return nil, collector.NewNotFoundError(SourceName, "fetch_table_metadata", table, nil)
	}
	if res.IsError() {
		return nil, collector.NewQueryError(SourceName, "fetch_table_metadata", fmt.Errorf("failed to check index existence: %s", res.Status()))
	}

	metadata := &collector.TableMetadata{
		SourceCategory:  collector.CategoryDocumentDB,
		SourceType:      SourceName,
		Catalog:         catalog,
		Schema:          schema,
		Name:            table,
		Type:            collector.TableTypeIndex,
		LastRefreshedAt: time.Now(),
		InferredSchema:  true, // Elasticsearch schemas are always inferred
	}

	// Get mapping to extract field information
	mappingRes, err := c.client.Indices.GetMapping(
		c.client.Indices.GetMapping.WithContext(ctx),
		c.client.Indices.GetMapping.WithIndex(table),
	)
	if err != nil {
		if ctx.Err() != nil {
			return nil, collector.WrapContextError(ctx, SourceName, "fetch_table_metadata")
		}
		return nil, collector.NewQueryError(SourceName, "fetch_table_metadata", err)
	}
	defer mappingRes.Body.Close()

	if mappingRes.IsError() {
		return nil, collector.NewQueryError(SourceName, "fetch_table_metadata", fmt.Errorf("failed to get mapping: %s", mappingRes.Status()))
	}

	var mappingData map[string]interface{}
	if err := json.NewDecoder(mappingRes.Body).Decode(&mappingData); err != nil {
		return nil, collector.NewParseError(SourceName, "fetch_table_metadata", err)
	}

	// Extract columns from mapping
	columns := c.extractColumnsFromMapping(mappingData, table)

	// If schema inference is enabled and no columns found from mapping, try to infer from sample documents
	if c.inferrer.GetConfig().Enabled && len(columns) == 0 {
		// Check context before schema inference
		if err := collector.CheckContext(ctx, SourceName, "fetch_table_metadata"); err != nil {
			return nil, err
		}

		inferredColumns, err := c.inferSchema(ctx, table)
		if err != nil {
			return nil, err
		}
		columns = inferredColumns
	}

	metadata.Columns = columns

	// Get index settings if configured
	if c.config.Collect == nil || c.config.Collect.Indexes {
		// Check context before fetching settings
		if err := collector.CheckContext(ctx, SourceName, "fetch_table_metadata"); err != nil {
			return nil, err
		}

		settings, err := c.fetchIndexSettings(ctx, table)
		if err != nil {
			return nil, err
		}
		if settings != nil {
			metadata.Properties = settings
		}
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

	// Get document count using _count API
	res, err := c.client.Count(
		c.client.Count.WithContext(ctx),
		c.client.Count.WithIndex(table),
	)
	if err != nil {
		if ctx.Err() != nil {
			return nil, collector.WrapContextError(ctx, SourceName, "fetch_table_statistics")
		}
		return nil, collector.NewQueryError(SourceName, "fetch_table_statistics", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return nil, collector.NewQueryError(SourceName, "fetch_table_statistics", fmt.Errorf("failed to get document count: %s", res.Status()))
	}

	var countResult map[string]interface{}
	if err := json.NewDecoder(res.Body).Decode(&countResult); err != nil {
		return nil, collector.NewParseError(SourceName, "fetch_table_statistics", err)
	}

	count := int64(0)
	if c, ok := countResult["count"].(float64); ok {
		count = int64(c)
	}

	stats := &collector.TableStatistics{
		RowCount:    count,
		CollectedAt: time.Now(),
	}

	return stats, nil
}

// FetchPartitions 获取分区信息（Elasticsearch 不支持传统分区）
func (c *Collector) FetchPartitions(ctx context.Context, catalog, schema, table string) ([]collector.PartitionInfo, error) {
	// Elasticsearch doesn't support partitions in the traditional sense
	return []collector.PartitionInfo{}, nil
}

// buildAddresses constructs the Elasticsearch addresses from configuration
func (c *Collector) buildAddresses() ([]string, error) {
	endpoint := c.config.Endpoint
	if endpoint == "" {
		return nil, fmt.Errorf("endpoint is required")
	}

	// Parse endpoint (expected format: host:port or host or http://host:port)
	if !strings.HasPrefix(endpoint, "http://") && !strings.HasPrefix(endpoint, "https://") {
		// Add default scheme
		endpoint = "http://" + endpoint
	}

	u, err := url.Parse(endpoint)
	if err != nil {
		return nil, fmt.Errorf("invalid endpoint URL: %s", endpoint)
	}

	// Add default port if not specified
	if u.Port() == "" {
		u.Host = u.Host + ":" + strconv.Itoa(DefaultPort)
	}

	return []string{u.String()}, nil
}

// wrapConnectionError wraps a connection error with appropriate error type
func (c *Collector) wrapConnectionError(err error) error {
	errStr := err.Error()
	if strings.Contains(errStr, "authentication failed") || strings.Contains(errStr, "auth failed") || strings.Contains(errStr, "401") {
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

// Category 返回数据源类别
func (c *Collector) Category() collector.DataSourceCategory {
	return collector.CategoryDocumentDB
}

// Type 返回数据源类型
func (c *Collector) Type() string {
	return SourceName
}

// extractColumnsFromMapping extracts column information from Elasticsearch mapping
func (c *Collector) extractColumnsFromMapping(mappingData map[string]interface{}, indexName string) []collector.Column {
	var columns []collector.Column

	// Navigate to the mapping properties
	if indexData, ok := mappingData[indexName].(map[string]interface{}); ok {
		if mappings, ok := indexData["mappings"].(map[string]interface{}); ok {
			if properties, ok := mappings["properties"].(map[string]interface{}); ok {
				columns = c.extractFieldsFromProperties(properties, "", 0)
			}
		}
	}

	return columns
}

// extractFieldsFromProperties recursively extracts fields from Elasticsearch properties
func (c *Collector) extractFieldsFromProperties(properties map[string]interface{}, prefix string, depth int) []collector.Column {
	var columns []collector.Column
	position := 1

	// Limit depth to prevent infinite recursion
	maxDepth := c.inferrer.GetConfig().MaxDepth
	if maxDepth > 0 && depth >= maxDepth {
		return columns
	}

	for fieldName, fieldDef := range properties {
		if fieldDefMap, ok := fieldDef.(map[string]interface{}); ok {
			fullFieldName := fieldName
			if prefix != "" {
				fullFieldName = prefix + "." + fieldName
			}

			// Get field type
			fieldType := "text" // default
			if t, ok := fieldDefMap["type"].(string); ok {
				fieldType = t
			}

			// Convert Elasticsearch type to standard type
			standardType := c.convertElasticsearchType(fieldType)

			column := collector.Column{
				OrdinalPosition: position,
				Name:            fullFieldName,
				Type:            standardType,
				SourceType:      fieldType,
				Nullable:        true, // Elasticsearch fields are generally nullable
			}

			columns = append(columns, column)
			position++

			// Handle nested objects
			if fieldType == "object" || fieldType == "nested" {
				if nestedProps, ok := fieldDefMap["properties"].(map[string]interface{}); ok {
					nestedColumns := c.extractFieldsFromProperties(nestedProps, fullFieldName, depth+1)
					for i, nestedCol := range nestedColumns {
						nestedCol.OrdinalPosition = position + i
					}
					columns = append(columns, nestedColumns...)
					position += len(nestedColumns)
				}
			}
		}
	}

	return columns
}

// convertElasticsearchType converts Elasticsearch field types to standard types
func (c *Collector) convertElasticsearchType(esType string) string {
	switch esType {
	case "text", "keyword":
		return "string"
	case "long":
		return "bigint"
	case "integer":
		return "int"
	case "short":
		return "smallint"
	case "byte":
		return "tinyint"
	case "double":
		return "double"
	case "float":
		return "float"
	case "half_float":
		return "float"
	case "scaled_float":
		return "decimal"
	case "date":
		return "timestamp"
	case "boolean":
		return "boolean"
	case "binary":
		return "binary"
	case "object", "nested":
		return "object"
	case "geo_point":
		return "geometry"
	case "geo_shape":
		return "geometry"
	case "ip":
		return "string"
	default:
		return "string"
	}
}

// inferSchema infers schema from sample documents in an Elasticsearch index
func (c *Collector) inferSchema(ctx context.Context, indexName string) ([]collector.Column, error) {
	// Check context before starting
	if err := collector.CheckContext(ctx, SourceName, "infer_schema"); err != nil {
		return nil, err
	}

	// Get sample size from configuration
	sampleSize := c.inferrer.GetConfig().SampleSize
	if sampleSize <= 0 {
		sampleSize = DefaultSampleSize
	}

	// Sample documents using search API
	samples, err := c.sampleDocuments(ctx, indexName, sampleSize)
	if err != nil {
		return nil, err
	}

	if len(samples) == 0 {
		// Return empty schema for empty index
		return []collector.Column{}, nil
	}

	// Convert samples to interface{} slice for inferrer
	interfaceSamples := make([]interface{}, len(samples))
	for i, sample := range samples {
		interfaceSamples[i] = sample
	}

	// Use DocumentInferrer to infer schema
	columns, err := c.inferrer.Infer(ctx, interfaceSamples)
	if err != nil {
		return nil, collector.NewInferenceError(SourceName, "infer_schema", err)
	}

	return columns, nil
}

// sampleDocuments samples documents from an Elasticsearch index
func (c *Collector) sampleDocuments(ctx context.Context, indexName string, sampleSize int) ([]map[string]interface{}, error) {
	// Check context before starting
	if err := collector.CheckContext(ctx, SourceName, "sample_documents"); err != nil {
		return nil, err
	}

	// Build search query to sample documents
	query := map[string]interface{}{
		"size": sampleSize,
		"query": map[string]interface{}{
			"match_all": map[string]interface{}{},
		},
	}

	queryBytes, err := json.Marshal(query)
	if err != nil {
		return nil, collector.NewParseError(SourceName, "sample_documents", err)
	}

	// Execute search
	res, err := c.client.Search(
		c.client.Search.WithContext(ctx),
		c.client.Search.WithIndex(indexName),
		c.client.Search.WithBody(strings.NewReader(string(queryBytes))),
	)
	if err != nil {
		if ctx.Err() != nil {
			return nil, collector.WrapContextError(ctx, SourceName, "sample_documents")
		}
		return nil, collector.NewQueryError(SourceName, "sample_documents", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return nil, collector.NewQueryError(SourceName, "sample_documents", fmt.Errorf("search failed: %s", res.Status()))
	}

	var searchResult map[string]interface{}
	if err := json.NewDecoder(res.Body).Decode(&searchResult); err != nil {
		return nil, collector.NewParseError(SourceName, "sample_documents", err)
	}

	var samples []map[string]interface{}
	if hits, ok := searchResult["hits"].(map[string]interface{}); ok {
		if hitsList, ok := hits["hits"].([]interface{}); ok {
			for _, hit := range hitsList {
				// Check context during iteration
				if err := collector.CheckContext(ctx, SourceName, "sample_documents"); err != nil {
					return nil, err
				}

				if hitMap, ok := hit.(map[string]interface{}); ok {
					if source, ok := hitMap["_source"].(map[string]interface{}); ok {
						samples = append(samples, source)
					}
				}
			}
		}
	}

	return samples, nil
}

// fetchIndexSettings retrieves index settings
func (c *Collector) fetchIndexSettings(ctx context.Context, indexName string) (map[string]string, error) {
	// Check context before starting
	if err := collector.CheckContext(ctx, SourceName, "fetch_index_settings"); err != nil {
		return nil, err
	}

	res, err := c.client.Indices.GetSettings(
		c.client.Indices.GetSettings.WithContext(ctx),
		c.client.Indices.GetSettings.WithIndex(indexName),
	)
	if err != nil {
		if ctx.Err() != nil {
			return nil, collector.WrapContextError(ctx, SourceName, "fetch_index_settings")
		}
		return nil, collector.NewQueryError(SourceName, "fetch_index_settings", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return nil, collector.NewQueryError(SourceName, "fetch_index_settings", fmt.Errorf("failed to get settings: %s", res.Status()))
	}

	var settingsData map[string]interface{}
	if err := json.NewDecoder(res.Body).Decode(&settingsData); err != nil {
		return nil, collector.NewParseError(SourceName, "fetch_index_settings", err)
	}

	settings := make(map[string]string)

	// Extract relevant settings
	if indexData, ok := settingsData[indexName].(map[string]interface{}); ok {
		if settingsMap, ok := indexData["settings"].(map[string]interface{}); ok {
			if indexSettings, ok := settingsMap["index"].(map[string]interface{}); ok {
				// Extract common settings
				if shards, ok := indexSettings["number_of_shards"].(string); ok {
					settings["number_of_shards"] = shards
				}
				if replicas, ok := indexSettings["number_of_replicas"].(string); ok {
					settings["number_of_replicas"] = replicas
				}
				if creationDate, ok := indexSettings["creation_date"].(string); ok {
					settings["creation_date"] = creationDate
				}
				if uuid, ok := indexSettings["uuid"].(string); ok {
					settings["uuid"] = uuid
				}
			}
		}
	}

	return settings, nil
}

// SampleDocuments provides public access to document sampling for testing
// This method is used by the DocumentDBCollector interface
func (c *Collector) SampleDocuments(ctx context.Context, catalog, collection string, limit int) ([]map[string]interface{}, error) {
	if c.client == nil {
		return nil, collector.NewConnectionClosedError(SourceName, "sample_documents")
	}

	return c.sampleDocuments(ctx, collection, limit)
}

// SetInferConfig updates the schema inference configuration
// This method is used by the DocumentDBCollector interface
func (c *Collector) SetInferConfig(config *config.InferConfig) {
	if config == nil {
		return
	}

	inferConfig := &infer.InferConfig{
		Enabled:    config.Enabled,
		SampleSize: config.SampleSize,
		MaxDepth:   config.MaxDepth,
		TypeMerge:  infer.TypeMergeStrategy(config.TypeMerge),
	}
	c.inferrer.SetConfig(inferConfig)
}

// Ensure Collector implements collector.Collector interface
var _ collector.Collector = (*Collector)(nil)