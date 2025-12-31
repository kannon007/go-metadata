// Package minio provides a MinIO/S3 metadata collector implementation.
package minio

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

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

const (
	// SourceName identifies this collector type
	SourceName = "minio"
	// DefaultRegion is the default region for MinIO
	DefaultRegion = "us-east-1"
	// DefaultTimeout is the default connection timeout in seconds
	DefaultTimeout = 30
)

// Collector MinIO/S3 元数据采集器
type Collector struct {
	config       *config.ConnectorConfig
	client       *minio.Client
	fileInferrer *infer.FileSchemaInferrer
}

// NewCollector 创建 MinIO 采集器实例
func NewCollector(cfg *config.ConnectorConfig) (collector.Collector, error) {
	if cfg == nil {
		return nil, collector.NewInvalidConfigError(SourceName, "config", "configuration cannot be nil")
	}
	if cfg.Type != "" && cfg.Type != SourceName {
		return nil, collector.NewInvalidConfigError(SourceName, "type", fmt.Sprintf("expected '%s', got '%s'", SourceName, cfg.Type))
	}

	// Initialize file schema inferrer
	var inferConfig *infer.InferConfig
	if cfg.Infer != nil {
		inferConfig = &infer.InferConfig{
			Enabled:    cfg.Infer.Enabled,
			SampleSize: cfg.Infer.SampleSize,
			MaxDepth:   cfg.Infer.MaxDepth,
			TypeMerge:  infer.TypeMergeStrategy(cfg.Infer.TypeMerge),
		}
	} else {
		inferConfig = infer.DefaultInferConfig()
	}

	fileInferrer := infer.NewFileSchemaInferrerWithConfig(inferConfig)

	return &Collector{
		config:       cfg,
		fileInferrer: fileInferrer,
	}, nil
}

// Connect 建立 MinIO 连接
func (c *Collector) Connect(ctx context.Context) error {
	if c.client != nil {
		return nil // Already connected
	}

	// Parse endpoint
	endpoint, secure, err := c.parseEndpoint()
	if err != nil {
		return collector.NewInvalidConfigError(SourceName, "endpoint", err.Error())
	}

	// Get region from extra properties
	region := DefaultRegion
	if c.config.Properties.Extra != nil {
		if r := c.config.Properties.Extra["region"]; r != "" {
			region = r
		}
	}

	// Create credentials
	var creds *credentials.Credentials
	if c.config.Credentials.User != "" && c.config.Credentials.Password != "" {
		creds = credentials.NewStaticV4(c.config.Credentials.User, c.config.Credentials.Password, "")
	} else {
		// Use anonymous credentials for public buckets
		creds = credentials.NewStaticV4("", "", "")
	}

	// Create MinIO client
	client, err := minio.New(endpoint, &minio.Options{
		Creds:  creds,
		Secure: secure,
		Region: region,
	})
	if err != nil {
		return c.wrapConnectionError(err)
	}

	// Set timeout if configured
	timeout := DefaultTimeout
	if c.config.Properties.ConnectionTimeout > 0 {
		timeout = c.config.Properties.ConnectionTimeout
	}

	// Test connection by listing buckets with timeout
	ctx, cancel := context.WithTimeout(ctx, time.Duration(timeout)*time.Second)
	defer cancel()

	_, err = client.ListBuckets(ctx)
	if err != nil {
		return c.wrapConnectionError(err)
	}

	c.client = client
	return nil
}

// Close 关闭 MinIO 连接
func (c *Collector) Close() error {
	// MinIO client doesn't require explicit closing
	c.client = nil
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

	// Try to list buckets to verify connectivity
	_, err := c.client.ListBuckets(ctx)
	if err != nil {
		return &collector.HealthStatus{
			Connected: false,
			Latency:   time.Since(start),
			Message:   fmt.Sprintf("failed to list buckets: %v", err),
		}, nil
	}

	return &collector.HealthStatus{
		Connected: true,
		Latency:   time.Since(start),
		Version:   "MinIO/S3 Compatible",
		Message:   "connected successfully",
	}, nil
}

// DiscoverCatalogs 发现 Catalog（MinIO 中 catalog 等同于 MinIO 实例）
func (c *Collector) DiscoverCatalogs(ctx context.Context) ([]collector.CatalogInfo, error) {
	// Check context before starting operation
	if err := collector.CheckContext(ctx, SourceName, "discover_catalogs"); err != nil {
		return nil, err
	}

	if c.client == nil {
		return nil, collector.NewConnectionClosedError(SourceName, "discover_catalogs")
	}

	// MinIO typically has one catalog per instance
	return []collector.CatalogInfo{
		{
			Catalog:     "minio",
			Type:        SourceName,
			Description: "MinIO Object Storage",
			Properties: map[string]string{
				"endpoint": c.config.Endpoint,
			},
		},
	}, nil
}

// ListSchemas 列出 Schema（MinIO 中 schema 等同于 bucket）
func (c *Collector) ListSchemas(ctx context.Context, catalog string) ([]string, error) {
	if c.client == nil {
		return nil, collector.NewConnectionClosedError(SourceName, "list_schemas")
	}

	// Check context before starting operation
	if err := collector.CheckContext(ctx, SourceName, "list_schemas"); err != nil {
		return nil, err
	}

	// List all buckets
	buckets, err := c.client.ListBuckets(ctx)
	if err != nil {
		if ctx.Err() != nil {
			return nil, collector.WrapContextError(ctx, SourceName, "list_schemas")
		}
		return nil, collector.NewQueryError(SourceName, "list_schemas", err)
	}

	var bucketNames []string
	for _, bucket := range buckets {
		bucketNames = append(bucketNames, bucket.Name)
	}

	// Apply schema matching filter
	bucketNames = c.filterSchemas(bucketNames)

	return bucketNames, nil
}

// ListTables 列出表（MinIO 中表等同于对象前缀）
func (c *Collector) ListTables(ctx context.Context, catalog, schema string, opts *collector.ListOptions) (*collector.TableListResult, error) {
	if c.client == nil {
		return nil, collector.NewConnectionClosedError(SourceName, "list_tables")
	}

	// Check context before starting operation
	if err := collector.CheckContext(ctx, SourceName, "list_tables"); err != nil {
		return nil, err
	}

	// List object prefixes as "tables"
	prefixes, err := c.listPrefixes(ctx, schema, "", "/")
	if err != nil {
		return nil, err
	}

	// Apply table matching filter
	prefixes = c.filterTables(prefixes, opts)

	// Apply pagination
	result := &collector.TableListResult{
		TotalCount: len(prefixes),
	}

	if opts != nil && opts.PageSize > 0 {
		startIdx := 0
		if opts.PageToken != "" {
			startIdx, _ = strconv.Atoi(opts.PageToken)
		}

		endIdx := startIdx + opts.PageSize
		if endIdx > len(prefixes) {
			endIdx = len(prefixes)
		}

		if startIdx < len(prefixes) {
			result.Tables = prefixes[startIdx:endIdx]
			if endIdx < len(prefixes) {
				result.NextPageToken = strconv.Itoa(endIdx)
			}
		}
	} else {
		result.Tables = prefixes
	}

	return result, nil
}

// FetchTableMetadata 获取表元数据（对象前缀元数据）
func (c *Collector) FetchTableMetadata(ctx context.Context, catalog, schema, table string) (*collector.TableMetadata, error) {
	if c.client == nil {
		return nil, collector.NewConnectionClosedError(SourceName, "fetch_table_metadata")
	}

	// Check context before starting operation
	if err := collector.CheckContext(ctx, SourceName, "fetch_table_metadata"); err != nil {
		return nil, err
	}

	metadata := &collector.TableMetadata{
		SourceCategory:  collector.CategoryObjectStorage,
		SourceType:      SourceName,
		Catalog:         catalog,
		Schema:          schema, // bucket name
		Name:            table,  // prefix
		Type:            collector.TableTypeBucket,
		LastRefreshedAt: time.Now(),
		InferredSchema:  false,
		Properties:      make(map[string]string),
	}

	// List objects under this prefix to analyze structure
	prefix := table
	if !strings.HasSuffix(prefix, "/") {
		prefix += "/"
	}

	objectCh := c.client.ListObjects(ctx, schema, minio.ListObjectsOptions{
		Prefix:    prefix,
		Recursive: false,
		MaxKeys:   100, // Limit for metadata analysis
	})

	var objects []minio.ObjectInfo
	var totalSize int64
	var objectCount int

	for object := range objectCh {
		if object.Err != nil {
			return nil, collector.NewQueryError(SourceName, "fetch_table_metadata", object.Err)
		}

		objects = append(objects, object)
		totalSize += object.Size
		objectCount++
	}

	// Set basic properties
	metadata.Properties["object_count"] = fmt.Sprintf("%d", objectCount)
	metadata.Properties["total_size"] = fmt.Sprintf("%d", totalSize)

	// Try to infer schema from file objects
	if c.fileInferrer.GetConfig().Enabled && len(objects) > 0 {
		columns, err := c.inferSchemaFromObjects(ctx, schema, objects)
		if err == nil && len(columns) > 0 {
			metadata.Columns = columns
			metadata.InferredSchema = true
		}
	}

	// If no schema inferred, create basic object structure
	if len(metadata.Columns) == 0 {
		metadata.Columns = []collector.Column{
			{
				OrdinalPosition: 1,
				Name:            "object_name",
				Type:            "TEXT",
				SourceType:      "string",
				Nullable:        false,
				Comment:         "Object name/key",
			},
			{
				OrdinalPosition: 2,
				Name:            "size",
				Type:            "BIGINT",
				SourceType:      "int64",
				Nullable:        false,
				Comment:         "Object size in bytes",
			},
			{
				OrdinalPosition: 3,
				Name:            "last_modified",
				Type:            "TIMESTAMP",
				SourceType:      "time.Time",
				Nullable:        false,
				Comment:         "Last modification time",
			},
			{
				OrdinalPosition: 4,
				Name:            "etag",
				Type:            "TEXT",
				SourceType:      "string",
				Nullable:        true,
				Comment:         "Object ETag",
			},
			{
				OrdinalPosition: 5,
				Name:            "content_type",
				Type:            "TEXT",
				SourceType:      "string",
				Nullable:        true,
				Comment:         "Object content type",
			},
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

	// List objects under this prefix to calculate statistics
	prefix := table
	if !strings.HasSuffix(prefix, "/") {
		prefix += "/"
	}

	objectCh := c.client.ListObjects(ctx, schema, minio.ListObjectsOptions{
		Prefix:    prefix,
		Recursive: true, // Recursive for complete statistics
	})

	var totalSize int64
	var objectCount int64

	for object := range objectCh {
		if object.Err != nil {
			return nil, collector.NewQueryError(SourceName, "fetch_table_statistics", object.Err)
		}

		totalSize += object.Size
		objectCount++
	}

	stats := &collector.TableStatistics{
		RowCount:      objectCount,
		DataSizeBytes: totalSize,
		CollectedAt:   time.Now(),
	}

	return stats, nil
}

// FetchPartitions 获取分区信息（对象存储中分区等同于子前缀）
func (c *Collector) FetchPartitions(ctx context.Context, catalog, schema, table string) ([]collector.PartitionInfo, error) {
	if c.client == nil {
		return nil, collector.NewConnectionClosedError(SourceName, "fetch_partitions")
	}

	// Check context before starting operation
	if err := collector.CheckContext(ctx, SourceName, "fetch_partitions"); err != nil {
		return nil, err
	}

	// List sub-prefixes as partitions
	prefix := table
	if !strings.HasSuffix(prefix, "/") {
		prefix += "/"
	}

	subPrefixes, err := c.listPrefixes(ctx, schema, prefix, "/")
	if err != nil {
		return nil, err
	}

	var partitions []collector.PartitionInfo
	for i, subPrefix := range subPrefixes {
		partitionInfo := collector.PartitionInfo{
			Name:       fmt.Sprintf("prefix-%d", i),
			Type:       "object_prefix",
			Expression: fmt.Sprintf("prefix = '%s'", subPrefix),
		}
		partitions = append(partitions, partitionInfo)
	}

	return partitions, nil
}

// Category 返回数据源类别
func (c *Collector) Category() collector.DataSourceCategory {
	return collector.CategoryObjectStorage
}

// Type 返回数据源类型
func (c *Collector) Type() string {
	return SourceName
}

// parseEndpoint parses the endpoint configuration to extract host and security settings
func (c *Collector) parseEndpoint() (string, bool, error) {
	endpoint := c.config.Endpoint
	if endpoint == "" {
		return "", false, fmt.Errorf("endpoint is required")
	}

	// Parse URL to determine if it's HTTPS
	if !strings.Contains(endpoint, "://") {
		// Default to HTTP if no scheme specified
		endpoint = "http://" + endpoint
	}

	u, err := url.Parse(endpoint)
	if err != nil {
		return "", false, fmt.Errorf("invalid endpoint URL: %w", err)
	}

	secure := u.Scheme == "https"
	host := u.Host

	// If no port specified, use default ports
	if !strings.Contains(host, ":") {
		if secure {
			host += ":443"
		} else {
			host += ":9000" // Default MinIO port
		}
	}

	return host, secure, nil
}

// wrapConnectionError wraps a connection error with appropriate error type
func (c *Collector) wrapConnectionError(err error) error {
	errStr := err.Error()
	if strings.Contains(errStr, "access denied") || strings.Contains(errStr, "invalid credentials") {
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

// filterSchemas applies matching rules to filter bucket names
func (c *Collector) filterSchemas(schemas []string) []string {
	if c.config.Matching != nil && c.config.Matching.Schemas != nil {
		ruleMatcher, err := matcher.NewRuleMatcher(
			c.config.Matching.Schemas,
			c.config.Matching.PatternType,
			c.config.Matching.CaseSensitive,
		)
		if err == nil {
			var filtered []string
			for _, s := range schemas {
				if ruleMatcher.Match(s) {
					filtered = append(filtered, s)
				}
			}
			schemas = filtered
		}
	}
	return schemas
}

// filterTables applies matching rules to filter prefix names
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
// listPrefixes lists object prefixes in a bucket (used as "tables")
func (c *Collector) listPrefixes(ctx context.Context, bucket, prefix, delimiter string) ([]string, error) {
	objectCh := c.client.ListObjects(ctx, bucket, minio.ListObjectsOptions{
		Prefix:    prefix,
		Recursive: false,
		MaxKeys:   1000, // Reasonable limit for prefix discovery
	})

	prefixSet := make(map[string]bool)
	var prefixes []string

	for object := range objectCh {
		if object.Err != nil {
			if ctx.Err() != nil {
				return nil, collector.WrapContextError(ctx, SourceName, "list_prefixes")
			}
			return nil, collector.NewQueryError(SourceName, "list_prefixes", object.Err)
		}

		// Extract prefix from object key
		objectKey := object.Key
		if prefix != "" && strings.HasPrefix(objectKey, prefix) {
			objectKey = strings.TrimPrefix(objectKey, prefix)
		}

		// Find the first delimiter to get the prefix
		if delimiterIdx := strings.Index(objectKey, delimiter); delimiterIdx != -1 {
			prefixName := objectKey[:delimiterIdx]
			if prefixName != "" && !prefixSet[prefixName] {
				prefixSet[prefixName] = true
				prefixes = append(prefixes, prefixName)
			}
		} else if objectKey != "" {
			// If no delimiter found, treat the whole key as a prefix
			if !prefixSet[objectKey] {
				prefixSet[objectKey] = true
				prefixes = append(prefixes, objectKey)
			}
		}
	}

	return prefixes, nil
}

// inferSchemaFromObjects attempts to infer schema from object files
func (c *Collector) inferSchemaFromObjects(ctx context.Context, bucket string, objects []minio.ObjectInfo) ([]collector.Column, error) {
	// Look for structured files (CSV, JSON, Parquet)
	for _, obj := range objects {
		if c.isStructuredFile(obj.Key) {
			columns, err := c.inferSchemaFromFile(ctx, bucket, obj.Key)
			if err == nil && len(columns) > 0 {
				return columns, nil
			}
		}
	}

	return nil, fmt.Errorf("no structured files found for schema inference")
}

// isStructuredFile checks if a file is a structured format that can be analyzed
func (c *Collector) isStructuredFile(key string) bool {
	lowerKey := strings.ToLower(key)
	return strings.HasSuffix(lowerKey, ".csv") ||
		strings.HasSuffix(lowerKey, ".json") ||
		strings.HasSuffix(lowerKey, ".parquet")
}

// inferSchemaFromFile infers schema from a specific file
func (c *Collector) inferSchemaFromFile(ctx context.Context, bucket, key string) ([]collector.Column, error) {
	// Get object
	object, err := c.client.GetObject(ctx, bucket, key, minio.GetObjectOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get object %s: %w", key, err)
	}
	defer object.Close()

	// Determine file format
	var format infer.FileFormat
	lowerKey := strings.ToLower(key)
	switch {
	case strings.HasSuffix(lowerKey, ".csv"):
		format = infer.FormatCSV
	case strings.HasSuffix(lowerKey, ".json"):
		format = infer.FormatJSON
	case strings.HasSuffix(lowerKey, ".parquet"):
		format = infer.FormatParquet
	default:
		return nil, fmt.Errorf("unsupported file format for %s", key)
	}

	// Create inference request
	request := &infer.FileInferenceRequest{
		Reader: object,
		Format: format,
	}

	if format == infer.FormatCSV {
		request.CSVOptions = infer.DefaultCSVOptions()
	}

	// Infer schema
	columns, err := c.fileInferrer.InferFromFile(ctx, request)
	if err != nil {
		return nil, fmt.Errorf("failed to infer schema from %s: %w", key, err)
	}

	return columns, nil
}

// ObjectStorageCollector interface methods (extended interface)

// ListPrefixes 列出对象前缀（作为 Schema）
func (c *Collector) ListPrefixes(ctx context.Context, bucket, prefix string, delimiter string) ([]string, error) {
	if c.client == nil {
		return nil, collector.NewConnectionClosedError(SourceName, "list_prefixes")
	}

	return c.listPrefixes(ctx, bucket, prefix, delimiter)
}

// InferFileSchema 推断文件 Schema
func (c *Collector) InferFileSchema(ctx context.Context, bucket, key string) ([]collector.Column, error) {
	if c.client == nil {
		return nil, collector.NewConnectionClosedError(SourceName, "infer_file_schema")
	}

	return c.inferSchemaFromFile(ctx, bucket, key)
}

// GetBucketPolicy 获取 Bucket 策略
func (c *Collector) GetBucketPolicy(ctx context.Context, bucket string) (*BucketPolicy, error) {
	if c.client == nil {
		return nil, collector.NewConnectionClosedError(SourceName, "get_bucket_policy")
	}

	// Check context before starting operation
	if err := collector.CheckContext(ctx, SourceName, "get_bucket_policy"); err != nil {
		return nil, err
	}

	policy := &BucketPolicy{
		Bucket: bucket,
	}

	// Get bucket policy
	policyStr, err := c.client.GetBucketPolicy(ctx, bucket)
	if err != nil {
		// Policy might not exist, which is not an error
		policy.Policy = ""
	} else {
		policy.Policy = policyStr
	}

	// Get bucket versioning
	versioningConfig, err := c.client.GetBucketVersioning(ctx, bucket)
	if err == nil {
		policy.Versioning = versioningConfig.Status == "Enabled"
	}

	// Get bucket encryption
	encryptionConfig, err := c.client.GetBucketEncryption(ctx, bucket)
	if err == nil && encryptionConfig != nil {
		// MinIO encryption configuration is complex, simplify for metadata
		policy.Encryption = "enabled"
	}

	return policy, nil
}

// BucketPolicy represents bucket policy information
type BucketPolicy struct {
	Bucket     string `json:"bucket"`
	Policy     string `json:"policy"`
	Versioning bool   `json:"versioning"`
	Encryption string `json:"encryption,omitempty"`
}