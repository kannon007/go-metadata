// Package rabbitmq provides a RabbitMQ metadata collector implementation.
package rabbitmq

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
	"go-metadata/internal/collector/matcher"
)

const (
	// SourceName identifies this collector type
	SourceName = "rabbitmq"
	// DefaultPort is the default RabbitMQ Management API port
	DefaultPort = 15672
	// DefaultTimeout is the default connection timeout in seconds
	DefaultTimeout = 30
)

// Collector RabbitMQ 元数据采集器
type Collector struct {
	config     *config.ConnectorConfig
	httpClient *http.Client
	baseURL    string
	username   string
	password   string
}

// NewCollector 创建 RabbitMQ 采集器实例
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

// Connect 建立 RabbitMQ Management API 连接
func (c *Collector) Connect(ctx context.Context) error {
	if c.httpClient != nil {
		return nil // Already connected
	}

	// Parse endpoint
	baseURL, err := c.parseEndpoint()
	if err != nil {
		return collector.NewInvalidConfigError(SourceName, "endpoint", err.Error())
	}

	// Set connection timeout
	timeout := DefaultTimeout
	if c.config.Properties.ConnectionTimeout > 0 {
		timeout = c.config.Properties.ConnectionTimeout
	}

	// Create HTTP client
	c.httpClient = &http.Client{
		Timeout: time.Duration(timeout) * time.Second,
	}

	c.baseURL = baseURL
	c.username = c.config.Credentials.User
	c.password = c.config.Credentials.Password

	// Test connection with health check
	_, err = c.HealthCheck(ctx)
	if err != nil {
		c.httpClient = nil
		return err
	}

	return nil
}

// Close 关闭 RabbitMQ 连接
func (c *Collector) Close() error {
	if c.httpClient != nil {
		c.httpClient = nil
	}
	return nil
}

// HealthCheck 健康检查
func (c *Collector) HealthCheck(ctx context.Context) (*collector.HealthStatus, error) {
	if c.httpClient == nil {
		return &collector.HealthStatus{
			Connected: false,
			Message:   "not connected",
		}, nil
	}

	start := time.Now()

	// Test connection by getting overview
	overview, err := c.getOverview(ctx)
	if err != nil {
		return &collector.HealthStatus{
			Connected: false,
			Latency:   time.Since(start),
			Message:   fmt.Sprintf("connection failed: %v", err),
		}, nil
	}

	return &collector.HealthStatus{
		Connected: true,
		Latency:   time.Since(start),
		Version:   overview.RabbitMQVersion,
		Message:   fmt.Sprintf("connected to RabbitMQ %s", overview.RabbitMQVersion),
	}, nil
}

// DiscoverCatalogs 发现 Catalog（RabbitMQ 中 catalog 等同于 RabbitMQ 实例）
func (c *Collector) DiscoverCatalogs(ctx context.Context) ([]collector.CatalogInfo, error) {
	// Check context before starting operation
	if err := collector.CheckContext(ctx, SourceName, "discover_catalogs"); err != nil {
		return nil, err
	}

	if c.httpClient == nil {
		return nil, collector.NewConnectionClosedError(SourceName, "discover_catalogs")
	}

	overview, err := c.getOverview(ctx)
	if err != nil {
		return nil, collector.NewQueryError(SourceName, "discover_catalogs", err)
	}

	// RabbitMQ typically has one catalog per cluster
	return []collector.CatalogInfo{
		{
			Catalog:     "rabbitmq",
			Type:        SourceName,
			Description: "RabbitMQ Cluster",
			Properties: map[string]string{
				"version":      overview.RabbitMQVersion,
				"cluster_name": overview.ClusterName,
				"node":         overview.Node,
			},
		},
	}, nil
}

// ListSchemas 列出 Schema（RabbitMQ 中 schema 等同于 vhost）
func (c *Collector) ListSchemas(ctx context.Context, catalog string) ([]string, error) {
	if c.httpClient == nil {
		return nil, collector.NewConnectionClosedError(SourceName, "list_schemas")
	}

	// Check context before starting operation
	if err := collector.CheckContext(ctx, SourceName, "list_schemas"); err != nil {
		return nil, err
	}

	vhosts, err := c.getVHosts(ctx)
	if err != nil {
		return nil, collector.NewQueryError(SourceName, "list_schemas", err)
	}

	var schemas []string
	for _, vhost := range vhosts {
		schemas = append(schemas, vhost.Name)
	}

	return schemas, nil
}

// ListTables 列出表（RabbitMQ 中表等同于 Queue）
func (c *Collector) ListTables(ctx context.Context, catalog, schema string, opts *collector.ListOptions) (*collector.TableListResult, error) {
	if c.httpClient == nil {
		return nil, collector.NewConnectionClosedError(SourceName, "list_tables")
	}

	// Check context before starting operation
	if err := collector.CheckContext(ctx, SourceName, "list_tables"); err != nil {
		return nil, err
	}

	// Get queues for the specified vhost
	queues, err := c.getQueues(ctx, schema)
	if err != nil {
		if ctx.Err() != nil {
			return nil, collector.WrapContextError(ctx, SourceName, "list_tables")
		}
		return nil, collector.NewQueryError(SourceName, "list_tables", err)
	}

	// Extract queue names
	var queueNames []string
	for _, queue := range queues {
		queueNames = append(queueNames, queue.Name)
	}

	// Apply table matching filter
	queueNames = c.filterTables(queueNames, opts)

	// Apply pagination
	result := &collector.TableListResult{
		TotalCount: len(queueNames),
	}

	if opts != nil && opts.PageSize > 0 {
		startIdx := 0
		if opts.PageToken != "" {
			startIdx, _ = strconv.Atoi(opts.PageToken)
		}

		endIdx := startIdx + opts.PageSize
		if endIdx > len(queueNames) {
			endIdx = len(queueNames)
		}

		if startIdx < len(queueNames) {
			result.Tables = queueNames[startIdx:endIdx]
			if endIdx < len(queueNames) {
				result.NextPageToken = strconv.Itoa(endIdx)
			}
		}
	} else {
		result.Tables = queueNames
	}

	return result, nil
}

// FetchTableMetadata 获取表元数据（RabbitMQ Queue 元数据）
func (c *Collector) FetchTableMetadata(ctx context.Context, catalog, schema, table string) (*collector.TableMetadata, error) {
	if c.httpClient == nil {
		return nil, collector.NewConnectionClosedError(SourceName, "fetch_table_metadata")
	}

	// Check context before starting operation
	if err := collector.CheckContext(ctx, SourceName, "fetch_table_metadata"); err != nil {
		return nil, err
	}

	// Get queue details
	queue, err := c.getQueue(ctx, schema, table)
	if err != nil {
		if ctx.Err() != nil {
			return nil, collector.WrapContextError(ctx, SourceName, "fetch_table_metadata")
		}
		return nil, collector.NewQueryError(SourceName, "fetch_table_metadata", err)
	}

	if queue == nil {
		return nil, collector.NewNotFoundError(SourceName, "fetch_table_metadata", table, nil)
	}

	metadata := &collector.TableMetadata{
		SourceCategory:  collector.CategoryMessageQueue,
		SourceType:      SourceName,
		Catalog:         catalog,
		Schema:          schema,
		Name:            table,
		Type:            collector.TableTypeQueue,
		LastRefreshedAt: time.Now(),
		InferredSchema:  true, // RabbitMQ doesn't have strict schemas
		Properties:      make(map[string]string),
	}

	// Add queue properties
	metadata.Properties["durable"] = fmt.Sprintf("%t", queue.Durable)
	metadata.Properties["auto_delete"] = fmt.Sprintf("%t", queue.AutoDelete)
	metadata.Properties["exclusive"] = fmt.Sprintf("%t", queue.Exclusive)
	metadata.Properties["state"] = queue.State
	metadata.Properties["node"] = queue.Node

	if queue.Arguments != nil {
		for key, value := range queue.Arguments {
			metadata.Properties[fmt.Sprintf("arg_%s", key)] = fmt.Sprintf("%v", value)
		}
	}

	// Create basic message structure columns
	metadata.Columns = []collector.Column{
		{
			OrdinalPosition: 1,
			Name:            "routing_key",
			Type:            "string",
			SourceType:      "string",
			Nullable:        true,
			Comment:         "Message routing key",
		},
		{
			OrdinalPosition: 2,
			Name:            "payload",
			Type:            "bytes",
			SourceType:      "bytes",
			Nullable:        true,
			Comment:         "Message payload",
		},
		{
			OrdinalPosition: 3,
			Name:            "properties",
			Type:            "object",
			SourceType:      "object",
			Nullable:        true,
			Comment:         "Message properties",
		},
		{
			OrdinalPosition: 4,
			Name:            "timestamp",
			Type:            "timestamp",
			SourceType:      "timestamp",
			Nullable:        true,
			Comment:         "Message timestamp",
		},
		{
			OrdinalPosition: 5,
			Name:            "exchange",
			Type:            "string",
			SourceType:      "string",
			Nullable:        true,
			Comment:         "Source exchange",
		},
	}

	return metadata, nil
}

// FetchTableStatistics 获取表统计信息
func (c *Collector) FetchTableStatistics(ctx context.Context, catalog, schema, table string) (*collector.TableStatistics, error) {
	if c.httpClient == nil {
		return nil, collector.NewConnectionClosedError(SourceName, "fetch_table_statistics")
	}

	// Check context before starting operation
	if err := collector.CheckContext(ctx, SourceName, "fetch_table_statistics"); err != nil {
		return nil, err
	}

	// Get queue details
	queue, err := c.getQueue(ctx, schema, table)
	if err != nil {
		if ctx.Err() != nil {
			return nil, collector.WrapContextError(ctx, SourceName, "fetch_table_statistics")
		}
		return nil, collector.NewQueryError(SourceName, "fetch_table_statistics", err)
	}

	if queue == nil {
		return nil, collector.NewNotFoundError(SourceName, "fetch_table_statistics", table, nil)
	}

	stats := &collector.TableStatistics{
		RowCount:      int64(queue.Messages),
		DataSizeBytes: int64(queue.MessageBytes),
		CollectedAt:   time.Now(),
	}

	return stats, nil
}

// FetchPartitions 获取分区信息（RabbitMQ 没有分区概念，返回空）
func (c *Collector) FetchPartitions(ctx context.Context, catalog, schema, table string) ([]collector.PartitionInfo, error) {
	// RabbitMQ doesn't have partitions like Kafka
	return []collector.PartitionInfo{}, nil
}

// Category 返回数据源类别
func (c *Collector) Category() collector.DataSourceCategory {
	return collector.CategoryMessageQueue
}

// Type 返回数据源类型
func (c *Collector) Type() string {
	return SourceName
}

// parseEndpoint parses the endpoint configuration to extract base URL
func (c *Collector) parseEndpoint() (string, error) {
	endpoint := c.config.Endpoint
	if endpoint == "" {
		return "", fmt.Errorf("endpoint is required")
	}

	// Handle hostname only case
	if !strings.Contains(endpoint, "://") && !strings.Contains(endpoint, "/") {
		// Simple hostname, add scheme and port
		if !strings.Contains(endpoint, ":") {
			endpoint = fmt.Sprintf("http://%s:%d", endpoint, DefaultPort)
		} else {
			endpoint = "http://" + endpoint
		}
	}

	// Parse URL
	u, err := url.Parse(endpoint)
	if err != nil {
		return "", fmt.Errorf("invalid endpoint URL: %v", err)
	}

	// If no scheme provided, assume http
	if u.Scheme == "" {
		u.Scheme = "http"
	}

	// If no port provided, use default
	if u.Port() == "" {
		u.Host = fmt.Sprintf("%s:%d", u.Hostname(), DefaultPort)
	}

	// Ensure path ends with /api
	if !strings.HasSuffix(u.Path, "/api") {
		if u.Path == "" || u.Path == "/" {
			u.Path = "/api"
		} else {
			u.Path = strings.TrimSuffix(u.Path, "/") + "/api"
		}
	}

	return u.String(), nil
}

// doRequest performs an HTTP request with authentication
func (c *Collector) doRequest(ctx context.Context, method, path string) (*http.Response, error) {
	url := c.baseURL + path

	req, err := http.NewRequestWithContext(ctx, method, url, nil)
	if err != nil {
		return nil, err
	}

	// Add authentication
	if c.username != "" && c.password != "" {
		req.SetBasicAuth(c.username, c.password)
	}

	// Set headers
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, c.wrapConnectionError(err)
	}

	return resp, nil
}

// wrapConnectionError wraps a connection error with appropriate error type
func (c *Collector) wrapConnectionError(err error) error {
	errStr := err.Error()
	if strings.Contains(errStr, "authentication") || strings.Contains(errStr, "401") {
		return collector.NewAuthError(SourceName, "connect", err)
	}
	if strings.Contains(errStr, "connection refused") || strings.Contains(errStr, "no such host") {
		return collector.NewNetworkError(SourceName, "connect", err)
	}
	if strings.Contains(errStr, "deadline exceeded") {
		return collector.NewDeadlineExceededError(SourceName, "connect", err)
	}
	if strings.Contains(errStr, "timeout") {
		return collector.NewTimeoutError(SourceName, "connect", err)
	}
	return collector.NewNetworkError(SourceName, "connect", err)
}

// filterTables applies matching rules to filter queues
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
// RabbitMQ Management API Models

// Overview represents RabbitMQ cluster overview
type Overview struct {
	RabbitMQVersion string `json:"rabbitmq_version"`
	ClusterName     string `json:"cluster_name"`
	Node            string `json:"node"`
	ErlangVersion   string `json:"erlang_version"`
}

// VHost represents a RabbitMQ virtual host
type VHost struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Tags        []string               `json:"tags"`
	Metadata    map[string]interface{} `json:"metadata"`
}

// Queue represents a RabbitMQ queue
type Queue struct {
	Name         string                 `json:"name"`
	VHost        string                 `json:"vhost"`
	Durable      bool                   `json:"durable"`
	AutoDelete   bool                   `json:"auto_delete"`
	Exclusive    bool                   `json:"exclusive"`
	Arguments    map[string]interface{} `json:"arguments"`
	Node         string                 `json:"node"`
	State        string                 `json:"state"`
	Messages     int                    `json:"messages"`
	MessageBytes int                    `json:"message_bytes"`
	Consumers    int                    `json:"consumers"`
}

// Exchange represents a RabbitMQ exchange
type Exchange struct {
	Name       string                 `json:"name"`
	VHost      string                 `json:"vhost"`
	Type       string                 `json:"type"`
	Durable    bool                   `json:"durable"`
	AutoDelete bool                   `json:"auto_delete"`
	Internal   bool                   `json:"internal"`
	Arguments  map[string]interface{} `json:"arguments"`
}

// Binding represents a RabbitMQ binding
type Binding struct {
	Source          string                 `json:"source"`
	VHost           string                 `json:"vhost"`
	Destination     string                 `json:"destination"`
	DestinationType string                 `json:"destination_type"`
	RoutingKey      string                 `json:"routing_key"`
	Arguments       map[string]interface{} `json:"arguments"`
}

// Consumer represents a RabbitMQ consumer
type Consumer struct {
	ConsumerTag string `json:"consumer_tag"`
	Queue       string `json:"queue"`
	VHost       string `json:"vhost"`
	Channel     string `json:"channel"`
	Node        string `json:"node"`
}

// RabbitMQ Management API Methods

// getOverview gets cluster overview information
func (c *Collector) getOverview(ctx context.Context) (*Overview, error) {
	resp, err := c.doRequest(ctx, "GET", "/overview")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized {
		return nil, collector.NewAuthError(SourceName, "get_overview", fmt.Errorf("authentication failed"))
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get overview: status %d", resp.StatusCode)
	}

	var overview Overview
	if err := json.NewDecoder(resp.Body).Decode(&overview); err != nil {
		return nil, fmt.Errorf("failed to decode overview response: %v", err)
	}

	return &overview, nil
}

// getVHosts gets list of virtual hosts
func (c *Collector) getVHosts(ctx context.Context) ([]VHost, error) {
	resp, err := c.doRequest(ctx, "GET", "/vhosts")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get vhosts: status %d", resp.StatusCode)
	}

	var vhosts []VHost
	if err := json.NewDecoder(resp.Body).Decode(&vhosts); err != nil {
		return nil, fmt.Errorf("failed to decode vhosts response: %v", err)
	}

	return vhosts, nil
}

// getQueues gets list of queues for a specific vhost
func (c *Collector) getQueues(ctx context.Context, vhost string) ([]Queue, error) {
	path := "/queues"
	if vhost != "" && vhost != "/" {
		path = fmt.Sprintf("/queues/%s", url.PathEscape(vhost))
	}

	resp, err := c.doRequest(ctx, "GET", path)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("vhost not found: %s", vhost)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get queues: status %d", resp.StatusCode)
	}

	var queues []Queue
	if err := json.NewDecoder(resp.Body).Decode(&queues); err != nil {
		return nil, fmt.Errorf("failed to decode queues response: %v", err)
	}

	return queues, nil
}

// getQueue gets details of a specific queue
func (c *Collector) getQueue(ctx context.Context, vhost, queueName string) (*Queue, error) {
	path := fmt.Sprintf("/queues/%s/%s", url.PathEscape(vhost), url.PathEscape(queueName))

	resp, err := c.doRequest(ctx, "GET", path)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, nil // Queue not found
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get queue: status %d", resp.StatusCode)
	}

	var queue Queue
	if err := json.NewDecoder(resp.Body).Decode(&queue); err != nil {
		return nil, fmt.Errorf("failed to decode queue response: %v", err)
	}

	return &queue, nil
}

// getExchanges gets list of exchanges for a specific vhost
func (c *Collector) getExchanges(ctx context.Context, vhost string) ([]Exchange, error) {
	path := "/exchanges"
	if vhost != "" && vhost != "/" {
		path = fmt.Sprintf("/exchanges/%s", url.PathEscape(vhost))
	}

	resp, err := c.doRequest(ctx, "GET", path)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("vhost not found: %s", vhost)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get exchanges: status %d", resp.StatusCode)
	}

	var exchanges []Exchange
	if err := json.NewDecoder(resp.Body).Decode(&exchanges); err != nil {
		return nil, fmt.Errorf("failed to decode exchanges response: %v", err)
	}

	return exchanges, nil
}

// getBindings gets list of bindings for a specific vhost
func (c *Collector) getBindings(ctx context.Context, vhost string) ([]Binding, error) {
	path := "/bindings"
	if vhost != "" && vhost != "/" {
		path = fmt.Sprintf("/bindings/%s", url.PathEscape(vhost))
	}

	resp, err := c.doRequest(ctx, "GET", path)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("vhost not found: %s", vhost)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get bindings: status %d", resp.StatusCode)
	}

	var bindings []Binding
	if err := json.NewDecoder(resp.Body).Decode(&bindings); err != nil {
		return nil, fmt.Errorf("failed to decode bindings response: %v", err)
	}

	return bindings, nil
}

// getConsumers gets list of consumers for a specific vhost
func (c *Collector) getConsumers(ctx context.Context, vhost string) ([]Consumer, error) {
	path := "/consumers"
	if vhost != "" && vhost != "/" {
		path = fmt.Sprintf("/consumers/%s", url.PathEscape(vhost))
	}

	resp, err := c.doRequest(ctx, "GET", path)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("vhost not found: %s", vhost)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get consumers: status %d", resp.StatusCode)
	}

	var consumers []Consumer
	if err := json.NewDecoder(resp.Body).Decode(&consumers); err != nil {
		return nil, fmt.Errorf("failed to decode consumers response: %v", err)
	}

	return consumers, nil
}

// Extended methods for RabbitMQ-specific functionality

// ListExchanges lists all exchanges in a vhost
func (c *Collector) ListExchanges(ctx context.Context, vhost string) ([]Exchange, error) {
	if c.httpClient == nil {
		return nil, collector.NewConnectionClosedError(SourceName, "list_exchanges")
	}

	// Check context before starting operation
	if err := collector.CheckContext(ctx, SourceName, "list_exchanges"); err != nil {
		return nil, err
	}

	return c.getExchanges(ctx, vhost)
}

// ListBindings lists all bindings in a vhost
func (c *Collector) ListBindings(ctx context.Context, vhost string) ([]Binding, error) {
	if c.httpClient == nil {
		return nil, collector.NewConnectionClosedError(SourceName, "list_bindings")
	}

	// Check context before starting operation
	if err := collector.CheckContext(ctx, SourceName, "list_bindings"); err != nil {
		return nil, err
	}

	return c.getBindings(ctx, vhost)
}

// ListConsumers lists all consumers in a vhost
func (c *Collector) ListConsumers(ctx context.Context, vhost string) ([]Consumer, error) {
	if c.httpClient == nil {
		return nil, collector.NewConnectionClosedError(SourceName, "list_consumers")
	}

	// Check context before starting operation
	if err := collector.CheckContext(ctx, SourceName, "list_consumers"); err != nil {
		return nil, err
	}

	return c.getConsumers(ctx, vhost)
}

// GetQueueBindings gets bindings for a specific queue
func (c *Collector) GetQueueBindings(ctx context.Context, vhost, queueName string) ([]Binding, error) {
	if c.httpClient == nil {
		return nil, collector.NewConnectionClosedError(SourceName, "get_queue_bindings")
	}

	// Check context before starting operation
	if err := collector.CheckContext(ctx, SourceName, "get_queue_bindings"); err != nil {
		return nil, err
	}

	path := fmt.Sprintf("/queues/%s/%s/bindings", url.PathEscape(vhost), url.PathEscape(queueName))

	resp, err := c.doRequest(ctx, "GET", path)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, collector.NewNotFoundError(SourceName, "get_queue_bindings", queueName, nil)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get queue bindings: status %d", resp.StatusCode)
	}

	var bindings []Binding
	if err := json.NewDecoder(resp.Body).Decode(&bindings); err != nil {
		return nil, fmt.Errorf("failed to decode queue bindings response: %v", err)
	}

	return bindings, nil
}

// GetExchangeBindings gets bindings for a specific exchange
func (c *Collector) GetExchangeBindings(ctx context.Context, vhost, exchangeName string) ([]Binding, error) {
	if c.httpClient == nil {
		return nil, collector.NewConnectionClosedError(SourceName, "get_exchange_bindings")
	}

	// Check context before starting operation
	if err := collector.CheckContext(ctx, SourceName, "get_exchange_bindings"); err != nil {
		return nil, err
	}

	path := fmt.Sprintf("/exchanges/%s/%s/bindings/source", url.PathEscape(vhost), url.PathEscape(exchangeName))

	resp, err := c.doRequest(ctx, "GET", path)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, collector.NewNotFoundError(SourceName, "get_exchange_bindings", exchangeName, nil)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get exchange bindings: status %d", resp.StatusCode)
	}

	var bindings []Binding
	if err := json.NewDecoder(resp.Body).Decode(&bindings); err != nil {
		return nil, fmt.Errorf("failed to decode exchange bindings response: %v", err)
	}

	return bindings, nil
}