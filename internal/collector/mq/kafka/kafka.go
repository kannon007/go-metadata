// Package kafka provides a Kafka metadata collector implementation.
package kafka

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

	"github.com/IBM/sarama"
)

const (
	// SourceName identifies this collector type
	SourceName = "kafka"
	// DefaultPort is the default Kafka port
	DefaultPort = 9092
	// DefaultTimeout is the default connection timeout in seconds
	DefaultTimeout = 30
	// DefaultMaxMessageBytes is the default max message size
	DefaultMaxMessageBytes = 1000000
)

// Collector Kafka 元数据采集器
type Collector struct {
	config       *config.ConnectorConfig
	client       sarama.Client
	admin        sarama.ClusterAdmin
	schemaClient *SchemaRegistryClient
}

// NewCollector 创建 Kafka 采集器实例
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

// Connect 建立 Kafka 连接
func (c *Collector) Connect(ctx context.Context) error {
	if c.client != nil {
		return nil // Already connected
	}

	// Parse brokers from endpoint
	brokers, err := c.parseBrokers()
	if err != nil {
		return collector.NewInvalidConfigError(SourceName, "endpoint", err.Error())
	}

	// Create Sarama configuration
	saramaConfig := sarama.NewConfig()
	saramaConfig.Version = sarama.V2_6_0_0 // Use a stable version

	// Set connection timeout
	timeout := DefaultTimeout
	if c.config.Properties.ConnectionTimeout > 0 {
		timeout = c.config.Properties.ConnectionTimeout
	}
	saramaConfig.Net.DialTimeout = time.Duration(timeout) * time.Second
	saramaConfig.Net.ReadTimeout = time.Duration(timeout) * time.Second
	saramaConfig.Net.WriteTimeout = time.Duration(timeout) * time.Second

	// Configure authentication if provided
	if c.config.Credentials.User != "" {
		saramaConfig.Net.SASL.Enable = true
		saramaConfig.Net.SASL.User = c.config.Credentials.User
		saramaConfig.Net.SASL.Password = c.config.Credentials.Password
		saramaConfig.Net.SASL.Mechanism = sarama.SASLTypePlaintext
	}

	// Configure TLS if specified in extra properties
	if c.config.Properties.Extra != nil {
		if tlsEnabled := c.config.Properties.Extra["tls_enabled"]; tlsEnabled == "true" {
			saramaConfig.Net.TLS.Enable = true
		}
	}

	// Set consumer configuration for metadata operations
	saramaConfig.Consumer.Return.Errors = true
	saramaConfig.Producer.Return.Successes = true
	saramaConfig.Producer.Return.Errors = true

	// Configure max message bytes
	maxMessageBytes := DefaultMaxMessageBytes
	if c.config.Properties.Extra != nil {
		if maxMsgStr := c.config.Properties.Extra["max_message_bytes"]; maxMsgStr != "" {
			if maxMsg, err := strconv.Atoi(maxMsgStr); err == nil {
				maxMessageBytes = maxMsg
			}
		}
	}
	saramaConfig.Producer.MaxMessageBytes = maxMessageBytes
	saramaConfig.Consumer.Fetch.Max = int32(maxMessageBytes)

	// Create Kafka client
	client, err := sarama.NewClient(brokers, saramaConfig)
	if err != nil {
		return c.wrapConnectionError(err)
	}

	// Create cluster admin
	admin, err := sarama.NewClusterAdminFromClient(client)
	if err != nil {
		client.Close()
		return c.wrapConnectionError(err)
	}

	c.client = client
	c.admin = admin

	// Initialize Schema Registry client if configured
	if c.config.Properties.Extra != nil {
		if schemaRegistryURL := c.config.Properties.Extra["schema_registry_url"]; schemaRegistryURL != "" {
			schemaClient, err := NewSchemaRegistryClient(schemaRegistryURL, c.config.Credentials.User, c.config.Credentials.Password)
			if err != nil {
				// Schema Registry is optional, log but don't fail
				// In a real implementation, you might want to log this
			} else {
				c.schemaClient = schemaClient
			}
		}
	}

	return nil
}

// Close 关闭 Kafka 连接
func (c *Collector) Close() error {
	var errs []error

	if c.admin != nil {
		if err := c.admin.Close(); err != nil {
			errs = append(errs, err)
		}
		c.admin = nil
	}

	if c.client != nil {
		if err := c.client.Close(); err != nil {
			errs = append(errs, err)
		}
		c.client = nil
	}

	if c.schemaClient != nil {
		c.schemaClient = nil
	}

	if len(errs) > 0 {
		return fmt.Errorf("errors closing Kafka connections: %v", errs)
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

	// Check if client is closed
	if c.client.Closed() {
		return &collector.HealthStatus{
			Connected: false,
			Latency:   time.Since(start),
			Message:   "client is closed",
		}, nil
	}

	// Try to get broker list to verify connectivity
	brokers := c.client.Brokers()
	if len(brokers) == 0 {
		return &collector.HealthStatus{
			Connected: false,
			Latency:   time.Since(start),
			Message:   "no brokers available",
		}, nil
	}

	// Try to refresh metadata to test connectivity
	if err := c.client.RefreshMetadata(); err != nil {
		return &collector.HealthStatus{
			Connected: false,
			Latency:   time.Since(start),
			Message:   fmt.Sprintf("failed to refresh metadata: %v", err),
		}, nil
	}

	// Get Kafka version from broker
	version := "unknown"
	if len(brokers) > 0 {
		// Get version from first available broker
		for _, broker := range brokers {
			connected, _ := broker.Connected()
			if connected {
				// Sarama doesn't directly expose broker version, so we'll use the client version
				version = c.client.Config().Version.String()
				break
			}
		}
	}

	return &collector.HealthStatus{
		Connected: true,
		Latency:   time.Since(start),
		Version:   version,
		Message:   fmt.Sprintf("connected to %d brokers", len(brokers)),
	}, nil
}

// DiscoverCatalogs 发现 Catalog（Kafka 中 catalog 等同于 Kafka 集群）
func (c *Collector) DiscoverCatalogs(ctx context.Context) ([]collector.CatalogInfo, error) {
	// Check context before starting operation
	if err := collector.CheckContext(ctx, SourceName, "discover_catalogs"); err != nil {
		return nil, err
	}

	if c.client == nil {
		return nil, collector.NewConnectionClosedError(SourceName, "discover_catalogs")
	}

	brokers := c.client.Brokers()
	if len(brokers) == 0 {
		return nil, collector.NewNetworkError(SourceName, "discover_catalogs", fmt.Errorf("no brokers available"))
	}

	// Kafka typically has one catalog per cluster
	return []collector.CatalogInfo{
		{
			Catalog:     "kafka",
			Type:        SourceName,
			Description: "Kafka Cluster",
			Properties: map[string]string{
				"brokers": fmt.Sprintf("%d", len(brokers)),
				"version": c.client.Config().Version.String(),
			},
		},
	}, nil
}

// ListSchemas 列出 Schema（Kafka 中 schema 等同于 namespace，这里使用默认值）
func (c *Collector) ListSchemas(ctx context.Context, catalog string) ([]string, error) {
	if c.client == nil {
		return nil, collector.NewConnectionClosedError(SourceName, "list_schemas")
	}

	// Check context before starting operation
	if err := collector.CheckContext(ctx, SourceName, "list_schemas"); err != nil {
		return nil, err
	}

	// Kafka doesn't have explicit schemas like databases, so we return a default namespace
	return []string{"default"}, nil
}

// ListTables 列出表（Kafka 中表等同于 Topic）
func (c *Collector) ListTables(ctx context.Context, catalog, schema string, opts *collector.ListOptions) (*collector.TableListResult, error) {
	if c.client == nil {
		return nil, collector.NewConnectionClosedError(SourceName, "list_tables")
	}

	// Check context before starting operation
	if err := collector.CheckContext(ctx, SourceName, "list_tables"); err != nil {
		return nil, err
	}

	// Get topic metadata
	topics, err := c.admin.ListTopics()
	if err != nil {
		if ctx.Err() != nil {
			return nil, collector.WrapContextError(ctx, SourceName, "list_tables")
		}
		return nil, collector.NewQueryError(SourceName, "list_tables", err)
	}

	// Convert topics map to slice
	var topicNames []string
	for topicName := range topics {
		topicNames = append(topicNames, topicName)
	}

	// Apply table matching filter
	topicNames = c.filterTables(topicNames, opts)

	// Apply pagination
	result := &collector.TableListResult{
		TotalCount: len(topicNames),
	}

	if opts != nil && opts.PageSize > 0 {
		startIdx := 0
		if opts.PageToken != "" {
			startIdx, _ = strconv.Atoi(opts.PageToken)
		}

		endIdx := startIdx + opts.PageSize
		if endIdx > len(topicNames) {
			endIdx = len(topicNames)
		}

		if startIdx < len(topicNames) {
			result.Tables = topicNames[startIdx:endIdx]
			if endIdx < len(topicNames) {
				result.NextPageToken = strconv.Itoa(endIdx)
			}
		}
	} else {
		result.Tables = topicNames
	}

	return result, nil
}

// FetchTableMetadata 获取表元数据（Kafka Topic 元数据）
func (c *Collector) FetchTableMetadata(ctx context.Context, catalog, schema, table string) (*collector.TableMetadata, error) {
	if c.client == nil {
		return nil, collector.NewConnectionClosedError(SourceName, "fetch_table_metadata")
	}

	// Check context before starting operation
	if err := collector.CheckContext(ctx, SourceName, "fetch_table_metadata"); err != nil {
		return nil, err
	}

	// Get topic metadata
	topicMetadata, err := c.admin.DescribeTopics([]string{table})
	if err != nil {
		if ctx.Err() != nil {
			return nil, collector.WrapContextError(ctx, SourceName, "fetch_table_metadata")
		}
		return nil, collector.NewQueryError(SourceName, "fetch_table_metadata", err)
	}

	if len(topicMetadata) == 0 {
		return nil, collector.NewNotFoundError(SourceName, "fetch_table_metadata", table, nil)
	}

	var topicDetail *sarama.TopicMetadata
	for _, topic := range topicMetadata {
		if topic.Name == table {
			topicDetail = topic
			break
		}
	}

	if topicDetail == nil {
		return nil, collector.NewNotFoundError(SourceName, "fetch_table_metadata", table, nil)
	}

	metadata := &collector.TableMetadata{
		SourceCategory:  collector.CategoryMessageQueue,
		SourceType:      SourceName,
		Catalog:         catalog,
		Schema:          schema,
		Name:            table,
		Type:            collector.TableTypeTopic,
		LastRefreshedAt: time.Now(),
		InferredSchema:  false, // Will be set to true if schema is inferred
		Properties:      make(map[string]string),
	}

	// Add partition information
	var partitions []collector.PartitionInfo
	for _, partition := range topicDetail.Partitions {
		partitionInfo := collector.PartitionInfo{
			Name:        fmt.Sprintf("partition-%d", partition.ID),
			Type:        "kafka_partition",
			Expression:  fmt.Sprintf("partition_id = %d", partition.ID),
			ValuesCount: 1,
		}
		partitions = append(partitions, partitionInfo)
	}
	metadata.Partitions = partitions

	// Add topic properties
	metadata.Properties["partition_count"] = fmt.Sprintf("%d", len(topicDetail.Partitions))
	metadata.Properties["replication_factor"] = fmt.Sprintf("%d", len(topicDetail.Partitions[0].Replicas))

	// Try to get topic configuration
	configResource := sarama.ConfigResource{
		Type: sarama.TopicResource,
		Name: table,
	}
	
	configs, err := c.admin.DescribeConfig(configResource)
	if err == nil && len(configs) > 0 {
		for _, config := range configs {
			if config.Value != "" {
				metadata.Properties[config.Name] = config.Value
			}
		}
	}

	// Try to get schema from Schema Registry if available
	if c.schemaClient != nil {
		if schema, err := c.schemaClient.GetLatestSchema(table + "-value"); err == nil {
			metadata.InferredSchema = false
			columns, err := c.parseSchemaToColumns(schema)
			if err == nil {
				metadata.Columns = columns
			}
		} else {
			// If no schema found in registry, we could potentially sample messages
			// For now, we'll create a basic column structure
			metadata.InferredSchema = true
			metadata.Columns = []collector.Column{
				{
					OrdinalPosition: 1,
					Name:            "key",
					Type:            "bytes",
					SourceType:      "bytes",
					Nullable:        true,
					Comment:         "Message key",
				},
				{
					OrdinalPosition: 2,
					Name:            "value",
					Type:            "bytes",
					SourceType:      "bytes",
					Nullable:        true,
					Comment:         "Message value",
				},
				{
					OrdinalPosition: 3,
					Name:            "timestamp",
					Type:            "timestamp",
					SourceType:      "timestamp",
					Nullable:        false,
					Comment:         "Message timestamp",
				},
				{
					OrdinalPosition: 4,
					Name:            "partition",
					Type:            "int",
					SourceType:      "int32",
					Nullable:        false,
					Comment:         "Partition number",
				},
				{
					OrdinalPosition: 5,
					Name:            "offset",
					Type:            "long",
					SourceType:      "int64",
					Nullable:        false,
					Comment:         "Message offset",
				},
			}
		}
	} else {
		// No schema registry, create basic message structure
		metadata.InferredSchema = true
		metadata.Columns = []collector.Column{
			{
				OrdinalPosition: 1,
				Name:            "key",
				Type:            "bytes",
				SourceType:      "bytes",
				Nullable:        true,
				Comment:         "Message key",
			},
			{
				OrdinalPosition: 2,
				Name:            "value",
				Type:            "bytes",
				SourceType:      "bytes",
				Nullable:        true,
				Comment:         "Message value",
			},
			{
				OrdinalPosition: 3,
				Name:            "timestamp",
				Type:            "timestamp",
				SourceType:      "timestamp",
				Nullable:        false,
				Comment:         "Message timestamp",
			},
			{
				OrdinalPosition: 4,
				Name:            "partition",
				Type:            "int",
				SourceType:      "int32",
				Nullable:        false,
				Comment:         "Partition number",
			},
			{
				OrdinalPosition: 5,
				Name:            "offset",
				Type:            "long",
				SourceType:      "int64",
				Nullable:        false,
				Comment:         "Message offset",
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

	// Get topic metadata for partition information
	topicMetadata, err := c.admin.DescribeTopics([]string{table})
	if err != nil {
		if ctx.Err() != nil {
			return nil, collector.WrapContextError(ctx, SourceName, "fetch_table_statistics")
		}
		return nil, collector.NewQueryError(SourceName, "fetch_table_statistics", err)
	}

	if len(topicMetadata) == 0 {
		return nil, collector.NewNotFoundError(SourceName, "fetch_table_statistics", table, nil)
	}

	var topicDetail *sarama.TopicMetadata
	for _, topic := range topicMetadata {
		if topic.Name == table {
			topicDetail = topic
			break
		}
	}

	if topicDetail == nil {
		return nil, collector.NewNotFoundError(SourceName, "fetch_table_statistics", table, nil)
	}

	// Calculate total message count across all partitions
	var totalMessages int64
	var totalSize int64

	// Create a consumer to get partition offsets
	consumer, err := sarama.NewConsumerFromClient(c.client)
	if err != nil {
		return nil, collector.NewQueryError(SourceName, "fetch_table_statistics", err)
	}
	defer consumer.Close()

	for _, partition := range topicDetail.Partitions {
		// Get the latest offset for this partition
		latestOffset, err := c.client.GetOffset(table, partition.ID, sarama.OffsetNewest)
		if err != nil {
			continue // Skip this partition if we can't get offset
		}

		// Get the earliest offset for this partition
		earliestOffset, err := c.client.GetOffset(table, partition.ID, sarama.OffsetOldest)
		if err != nil {
			continue // Skip this partition if we can't get offset
		}

		// Calculate message count for this partition
		partitionMessages := latestOffset - earliestOffset
		totalMessages += partitionMessages

		// Note: Getting actual size would require consuming messages, which is expensive
		// For now, we'll estimate based on message count
		totalSize += partitionMessages * 1024 // Rough estimate of 1KB per message
	}

	stats := &collector.TableStatistics{
		RowCount:       totalMessages,
		DataSizeBytes:  totalSize,
		PartitionCount: len(topicDetail.Partitions),
		CollectedAt:    time.Now(),
	}

	return stats, nil
}

// FetchPartitions 获取分区信息
func (c *Collector) FetchPartitions(ctx context.Context, catalog, schema, table string) ([]collector.PartitionInfo, error) {
	if c.client == nil {
		return nil, collector.NewConnectionClosedError(SourceName, "fetch_partitions")
	}

	// Check context before starting operation
	if err := collector.CheckContext(ctx, SourceName, "fetch_partitions"); err != nil {
		return nil, err
	}

	// Get topic metadata
	topicMetadata, err := c.admin.DescribeTopics([]string{table})
	if err != nil {
		if ctx.Err() != nil {
			return nil, collector.WrapContextError(ctx, SourceName, "fetch_partitions")
		}
		return nil, collector.NewQueryError(SourceName, "fetch_partitions", err)
	}

	if len(topicMetadata) == 0 {
		return nil, collector.NewNotFoundError(SourceName, "fetch_partitions", table, nil)
	}

	var topicDetail *sarama.TopicMetadata
	for _, topic := range topicMetadata {
		if topic.Name == table {
			topicDetail = topic
			break
		}
	}

	if topicDetail == nil {
		return nil, collector.NewNotFoundError(SourceName, "fetch_partitions", table, nil)
	}

	var partitions []collector.PartitionInfo
	for _, partition := range topicDetail.Partitions {
		partitionInfo := collector.PartitionInfo{
			Name:        fmt.Sprintf("partition-%d", partition.ID),
			Type:        "kafka_partition",
			Columns:     []string{"partition_id"},
			Expression:  fmt.Sprintf("partition_id = %d", partition.ID),
			ValuesCount: 1,
		}
		partitions = append(partitions, partitionInfo)
	}

	return partitions, nil
}

// Category 返回数据源类别
func (c *Collector) Category() collector.DataSourceCategory {
	return collector.CategoryMessageQueue
}

// Type 返回数据源类型
func (c *Collector) Type() string {
	return SourceName
}

// parseBrokers parses the endpoint configuration to extract broker addresses
func (c *Collector) parseBrokers() ([]string, error) {
	endpoint := c.config.Endpoint
	if endpoint == "" {
		return nil, fmt.Errorf("endpoint is required")
	}

	// Split by comma to support multiple brokers
	brokerStrs := strings.Split(endpoint, ",")
	var brokers []string

	for _, brokerStr := range brokerStrs {
		brokerStr = strings.TrimSpace(brokerStr)
		if brokerStr == "" {
			continue
		}

		// Parse broker address (expected format: host:port or host)
		host := brokerStr
		port := DefaultPort

		if idx := strings.LastIndex(brokerStr, ":"); idx != -1 {
			host = brokerStr[:idx]
			var err error
			port, err = strconv.Atoi(brokerStr[idx+1:])
			if err != nil {
				return nil, fmt.Errorf("invalid port in broker address: %s", brokerStr)
			}
		}

		brokers = append(brokers, net.JoinHostPort(host, strconv.Itoa(port)))
	}

	if len(brokers) == 0 {
		return nil, fmt.Errorf("no valid brokers found in endpoint: %s", endpoint)
	}

	return brokers, nil
}

// wrapConnectionError wraps a connection error with appropriate error type
func (c *Collector) wrapConnectionError(err error) error {
	errStr := err.Error()
	if strings.Contains(errStr, "SASL") || strings.Contains(errStr, "authentication") {
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

// filterTables applies matching rules to filter topics
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

// parseSchemaToColumns converts a schema registry schema to columns
func (c *Collector) parseSchemaToColumns(schema *Schema) ([]collector.Column, error) {
	// This is a placeholder implementation
	// In a real implementation, you would parse Avro/Protobuf/JSON schemas
	// and convert them to column definitions
	
	var columns []collector.Column
	
	// Add basic message metadata columns
	columns = append(columns, collector.Column{
		OrdinalPosition: 1,
		Name:            "key",
		Type:            "bytes",
		SourceType:      "bytes",
		Nullable:        true,
		Comment:         "Message key",
	})
	
	// Parse schema fields based on schema type
	switch schema.SchemaType {
	case "AVRO":
		// Parse Avro schema - this would require an Avro schema parser
		// For now, add a generic value column
		columns = append(columns, collector.Column{
			OrdinalPosition: 2,
			Name:            "value",
			Type:            "record",
			SourceType:      "avro",
			Nullable:        true,
			Comment:         "Avro message value",
		})
	case "PROTOBUF":
		// Parse Protobuf schema - this would require a Protobuf schema parser
		columns = append(columns, collector.Column{
			OrdinalPosition: 2,
			Name:            "value",
			Type:            "message",
			SourceType:      "protobuf",
			Nullable:        true,
			Comment:         "Protobuf message value",
		})
	case "JSON":
		// Parse JSON schema - this would require a JSON schema parser
		columns = append(columns, collector.Column{
			OrdinalPosition: 2,
			Name:            "value",
			Type:            "object",
			SourceType:      "json",
			Nullable:        true,
			Comment:         "JSON message value",
		})
	default:
		columns = append(columns, collector.Column{
			OrdinalPosition: 2,
			Name:            "value",
			Type:            "bytes",
			SourceType:      "bytes",
			Nullable:        true,
			Comment:         "Message value",
		})
	}
	
	// Add standard Kafka message metadata
	columns = append(columns, 
		collector.Column{
			OrdinalPosition: 3,
			Name:            "timestamp",
			Type:            "timestamp",
			SourceType:      "timestamp",
			Nullable:        false,
			Comment:         "Message timestamp",
		},
		collector.Column{
			OrdinalPosition: 4,
			Name:            "partition",
			Type:            "int",
			SourceType:      "int32",
			Nullable:        false,
			Comment:         "Partition number",
		},
		collector.Column{
			OrdinalPosition: 5,
			Name:            "offset",
			Type:            "long",
			SourceType:      "int64",
			Nullable:        false,
			Comment:         "Message offset",
		},
	)
	
	return columns, nil
}

// ListConsumerGroups 获取消费者组列表
func (c *Collector) ListConsumerGroups(ctx context.Context, topic string) ([]ConsumerGroup, error) {
	if c.client == nil {
		return nil, collector.NewConnectionClosedError(SourceName, "list_consumer_groups")
	}

	// Check context before starting operation
	if err := collector.CheckContext(ctx, SourceName, "list_consumer_groups"); err != nil {
		return nil, err
	}

	// Get all consumer groups
	groups, err := c.admin.ListConsumerGroups()
	if err != nil {
		if ctx.Err() != nil {
			return nil, collector.WrapContextError(ctx, SourceName, "list_consumer_groups")
		}
		return nil, collector.NewQueryError(SourceName, "list_consumer_groups", err)
	}

	var consumerGroups []ConsumerGroup
	for groupID := range groups {
		// Get group details
		groupDetails, err := c.admin.DescribeConsumerGroups([]string{groupID})
		if err != nil {
			continue // Skip groups we can't describe
		}

		if len(groupDetails) == 0 {
			continue
		}

		var groupDetail *sarama.GroupDescription
		for _, desc := range groupDetails {
			if desc.GroupId == groupID {
				groupDetail = desc
				break
			}
		}

		if groupDetail == nil {
			continue
		}

		consumerGroup := ConsumerGroup{
			GroupID: groupID,
			State:   groupDetail.State,
			Members: len(groupDetail.Members),
			Lag:     make(map[int32]int64),
		}

		// If topic is specified, filter by topic and get lag information
		if topic != "" {
			// Get consumer group offsets for the specific topic
			coordinator, err := c.client.Coordinator(groupID)
			if err != nil {
				continue
			}

			request := &sarama.OffsetFetchRequest{
				Version:       1,
				ConsumerGroup: groupID,
			}

			// Get topic partitions
			partitions, err := c.client.Partitions(topic)
			if err != nil {
				continue
			}

			for _, partition := range partitions {
				request.AddPartition(topic, partition)
			}

			response, err := coordinator.FetchOffset(request)
			if err != nil {
				continue
			}

			// Calculate lag for each partition
			for partition, block := range response.Blocks[topic] {
				if block.Err == sarama.ErrNoError {
					// Get latest offset
					latestOffset, err := c.client.GetOffset(topic, partition, sarama.OffsetNewest)
					if err != nil {
						continue
					}
					
					lag := latestOffset - block.Offset
					if lag < 0 {
						lag = 0
					}
					consumerGroup.Lag[partition] = lag
				}
			}

			// Only include groups that consume from this topic
			if len(consumerGroup.Lag) > 0 {
				consumerGroups = append(consumerGroups, consumerGroup)
			}
		} else {
			// Include all groups if no topic filter
			consumerGroups = append(consumerGroups, consumerGroup)
		}
	}

	return consumerGroups, nil
}

// FetchSchema 获取 Topic 的 Schema（从 Schema Registry）
func (c *Collector) FetchSchema(ctx context.Context, topic string) (*MessageSchema, error) {
	if c.schemaClient == nil {
		return nil, collector.NewUnsupportedFeatureError(SourceName, "fetch_schema", "Schema Registry not configured")
	}

	// Check context before starting operation
	if err := collector.CheckContext(ctx, SourceName, "fetch_schema"); err != nil {
		return nil, err
	}

	// Get schemas for the topic
	keySchema, valueSchema, err := c.schemaClient.GetTopicSchemas(topic)
	if err != nil {
		return nil, collector.NewQueryError(SourceName, "fetch_schema", err)
	}

	// Prefer value schema, fall back to key schema
	var schema *Schema
	if valueSchema != nil {
		schema = valueSchema
	} else if keySchema != nil {
		schema = keySchema
	} else {
		return nil, collector.NewNotFoundError(SourceName, "fetch_schema", fmt.Sprintf("schema for topic %s", topic), nil)
	}

	messageSchema := &MessageSchema{
		Subject:    schema.Subject,
		Version:    schema.Version,
		SchemaType: schema.SchemaType,
		Schema:     schema.Schema,
	}

	// Parse schema to extract columns
	columns, err := c.parseSchemaToColumns(schema)
	if err != nil {
		return nil, collector.NewParseError(SourceName, "fetch_schema", err)
	}
	messageSchema.Columns = columns

	return messageSchema, nil
}

// FetchTopicConfig 获取 Topic 配置
func (c *Collector) FetchTopicConfig(ctx context.Context, topic string) (map[string]string, error) {
	if c.client == nil {
		return nil, collector.NewConnectionClosedError(SourceName, "fetch_topic_config")
	}

	// Check context before starting operation
	if err := collector.CheckContext(ctx, SourceName, "fetch_topic_config"); err != nil {
		return nil, err
	}

	// Get topic configuration
	configResource := sarama.ConfigResource{
		Type: sarama.TopicResource,
		Name: topic,
	}
	
	configs, err := c.admin.DescribeConfig(configResource)
	if err != nil {
		if ctx.Err() != nil {
			return nil, collector.WrapContextError(ctx, SourceName, "fetch_topic_config")
		}
		return nil, collector.NewQueryError(SourceName, "fetch_topic_config", err)
	}

	result := make(map[string]string)
	for _, config := range configs {
		// Only include non-default configurations
		if config.Value != "" {
			result[config.Name] = config.Value
		}
	}

	return result, nil
}

// GetTopicPartitionInfo 获取 Topic 分区详细信息
func (c *Collector) GetTopicPartitionInfo(ctx context.Context, topic string) ([]TopicPartitionInfo, error) {
	if c.client == nil {
		return nil, collector.NewConnectionClosedError(SourceName, "get_topic_partition_info")
	}

	// Check context before starting operation
	if err := collector.CheckContext(ctx, SourceName, "get_topic_partition_info"); err != nil {
		return nil, err
	}

	// Get topic metadata
	topicMetadata, err := c.admin.DescribeTopics([]string{topic})
	if err != nil {
		if ctx.Err() != nil {
			return nil, collector.WrapContextError(ctx, SourceName, "get_topic_partition_info")
		}
		return nil, collector.NewQueryError(SourceName, "get_topic_partition_info", err)
	}

	if len(topicMetadata) == 0 {
		return nil, collector.NewNotFoundError(SourceName, "get_topic_partition_info", topic, nil)
	}

	var topicDetail *sarama.TopicMetadata
	for _, topicMeta := range topicMetadata {
		if topicMeta.Name == topic {
			topicDetail = topicMeta
			break
		}
	}

	if topicDetail == nil {
		return nil, collector.NewNotFoundError(SourceName, "get_topic_partition_info", topic, nil)
	}

	var partitionInfos []TopicPartitionInfo
	for _, partition := range topicDetail.Partitions {
		// Get partition offsets
		earliestOffset, err := c.client.GetOffset(topic, partition.ID, sarama.OffsetOldest)
		if err != nil {
			earliestOffset = -1
		}

		latestOffset, err := c.client.GetOffset(topic, partition.ID, sarama.OffsetNewest)
		if err != nil {
			latestOffset = -1
		}

		messageCount := int64(-1)
		if earliestOffset >= 0 && latestOffset >= 0 {
			messageCount = latestOffset - earliestOffset
		}

		partitionInfo := TopicPartitionInfo{
			ID:            partition.ID,
			Leader:        partition.Leader,
			Replicas:      partition.Replicas,
			ISR:           partition.Isr,
			EarliestOffset: earliestOffset,
			LatestOffset:  latestOffset,
			MessageCount:  messageCount,
		}

		partitionInfos = append(partitionInfos, partitionInfo)
	}

	return partitionInfos, nil
}

// GetBrokerInfo 获取 Broker 信息
func (c *Collector) GetBrokerInfo(ctx context.Context) ([]BrokerInfo, error) {
	if c.client == nil {
		return nil, collector.NewConnectionClosedError(SourceName, "get_broker_info")
	}

	// Check context before starting operation
	if err := collector.CheckContext(ctx, SourceName, "get_broker_info"); err != nil {
		return nil, err
	}

	brokers := c.client.Brokers()
	var brokerInfos []BrokerInfo

	for _, broker := range brokers {
		brokerInfo := BrokerInfo{
			ID:        broker.ID(),
			Address:   broker.Addr(),
			Connected: func() bool { connected, _ := broker.Connected(); return connected }(),
		}

		// Try to get additional broker metadata if connected
		connected, _ := broker.Connected()
		if connected {
			// Get broker configuration (this might require admin privileges)
			configResource := sarama.ConfigResource{
				Type: sarama.BrokerResource,
				Name: strconv.Itoa(int(broker.ID())),
			}
			
			configs, err := c.admin.DescribeConfig(configResource)
			if err == nil && len(configs) > 0 {
				brokerInfo.Config = make(map[string]string)
				for _, config := range configs {
					if config.Value != "" {
						brokerInfo.Config[config.Name] = config.Value
					}
				}
			}
		}

		brokerInfos = append(brokerInfos, brokerInfo)
	}

	return brokerInfos, nil
}

// ConsumerGroup represents a Kafka consumer group
type ConsumerGroup struct {
	GroupID string            `json:"group_id"`
	State   string            `json:"state"`
	Members int               `json:"members"`
	Lag     map[int32]int64   `json:"lag"` // partition -> lag
}

// MessageSchema represents a message schema from Schema Registry
type MessageSchema struct {
	Subject    string             `json:"subject"`
	Version    int                `json:"version"`
	SchemaType string             `json:"schema_type"` // AVRO, PROTOBUF, JSON
	Schema     string             `json:"schema"`
	Columns    []collector.Column `json:"columns"` // Parsed fields
}

// TopicPartitionInfo represents detailed partition information
type TopicPartitionInfo struct {
	ID             int32   `json:"id"`
	Leader         int32   `json:"leader"`
	Replicas       []int32 `json:"replicas"`
	ISR            []int32 `json:"isr"` // In-Sync Replicas
	EarliestOffset int64   `json:"earliest_offset"`
	LatestOffset   int64   `json:"latest_offset"`
	MessageCount   int64   `json:"message_count"`
}

// BrokerInfo represents Kafka broker information
type BrokerInfo struct {
	ID        int32             `json:"id"`
	Address   string            `json:"address"`
	Connected bool              `json:"connected"`
	Config    map[string]string `json:"config,omitempty"`
}

// Ensure Collector implements collector.Collector interface
var _ collector.Collector = (*Collector)(nil)