// Package hive provides a Hive metadata collector implementation.
package hive

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"
	"strings"
	"time"

	"go-metadata/internal/collector"
	"go-metadata/internal/collector/config"
	"go-metadata/internal/collector/matcher"
)

const (
	// SourceName identifies this collector type
	SourceName = "hive"
	// DefaultPort is the default HiveServer2 port
	DefaultPort = 10000
	// DefaultTimeout is the default connection timeout in seconds
	DefaultTimeout = 30
)

// Collector Hive 元数据采集器
type Collector struct {
	config *config.ConnectorConfig
	db     *sql.DB
}

// NewCollector 创建 Hive 采集器实例
func NewCollector(cfg *config.ConnectorConfig) (collector.Collector, error) {
	if cfg == nil {
		return nil, collector.NewInvalidConfigError(SourceName, "config", "configuration cannot be nil")
	}
	if cfg.Type != "" && cfg.Type != SourceName {
		return nil, collector.NewInvalidConfigError(SourceName, "type", fmt.Sprintf("expected '%s', got '%s'", SourceName, cfg.Type))
	}
	return &Collector{config: cfg}, nil
}

// Connect 建立 Hive 连接 (通过 HiveServer2 Thrift 协议)
// Note: Requires a Hive driver to be registered. Common options include:
// - github.com/beltran/gohive (import _ "github.com/beltran/gohive")
// - github.com/apache/thrift based drivers
func (c *Collector) Connect(ctx context.Context) error {
	if c.db != nil {
		return nil // Already connected
	}

	dsn, err := c.buildDSN()
	if err != nil {
		return collector.NewInvalidConfigError(SourceName, "endpoint", err.Error())
	}

	// Get driver name from config, default to "hive"
	driverName := "hive"
	if c.config.Properties.Extra != nil {
		if driver, ok := c.config.Properties.Extra["driver"]; ok && driver != "" {
			driverName = driver
		}
	}

	db, err := sql.Open(driverName, dsn)
	if err != nil {
		return collector.NewNetworkError(SourceName, "connect", err)
	}

	// Configure connection pool
	if c.config.Properties.MaxOpenConns > 0 {
		db.SetMaxOpenConns(c.config.Properties.MaxOpenConns)
	}
	if c.config.Properties.MaxIdleConns > 0 {
		db.SetMaxIdleConns(c.config.Properties.MaxIdleConns)
	}
	if c.config.Properties.ConnMaxLifetime > 0 {
		db.SetConnMaxLifetime(time.Duration(c.config.Properties.ConnMaxLifetime) * time.Second)
	}

	// Test connection with context
	if err := db.PingContext(ctx); err != nil {
		db.Close()
		return c.wrapConnectionError(err)
	}

	c.db = db
	return nil
}

// Close 关闭 Hive 连接
func (c *Collector) Close() error {
	if c.db != nil {
		err := c.db.Close()
		c.db = nil
		return err
	}
	return nil
}

// HealthCheck 健康检查
func (c *Collector) HealthCheck(ctx context.Context) (*collector.HealthStatus, error) {
	if c.db == nil {
		return &collector.HealthStatus{
			Connected: false,
			Message:   "not connected",
		}, nil
	}

	start := time.Now()

	// Ping to check connection
	if err := c.db.PingContext(ctx); err != nil {
		return &collector.HealthStatus{
			Connected: false,
			Latency:   time.Since(start),
			Message:   err.Error(),
		}, nil
	}

	// Get Hive version using SET command
	var version string
	rows, err := c.db.QueryContext(ctx, "SET hive.server2.thrift.http.path")
	if err == nil {
		defer rows.Close()
		// Try to get version from system properties
		version = "HiveServer2"
	}

	// Alternative: try to get version from a simple query
	if version == "" {
		version = "HiveServer2 (version unknown)"
	}

	return &collector.HealthStatus{
		Connected: true,
		Latency:   time.Since(start),
		Version:   version,
	}, nil
}


// DiscoverCatalogs 发现 Catalog（Hive 中 catalog 通常是单一的）
func (c *Collector) DiscoverCatalogs(ctx context.Context) ([]collector.CatalogInfo, error) {
	if c.db == nil {
		return nil, collector.NewConnectionClosedError(SourceName, "discover_catalogs")
	}

	// Hive typically has one catalog per connection
	return []collector.CatalogInfo{
		{
			Catalog:     "hive",
			Type:        SourceName,
			Description: "Hive Metastore",
			Properties:  map[string]string{},
		},
	}, nil
}

// ListSchemas 列出 Schema（Hive 中 schema 等同于 database）
func (c *Collector) ListSchemas(ctx context.Context, catalog string) ([]string, error) {
	if c.db == nil {
		return nil, collector.NewConnectionClosedError(SourceName, "list_schemas")
	}

	// Check context before starting operation
	if err := collector.CheckContext(ctx, SourceName, "list_schemas"); err != nil {
		return nil, err
	}

	rows, err := c.db.QueryContext(ctx, "SHOW DATABASES")
	if err != nil {
		if ctx.Err() != nil {
			return nil, collector.WrapContextError(ctx, SourceName, "list_schemas")
		}
		return nil, collector.NewQueryError(SourceName, "list_schemas", err)
	}
	defer rows.Close()

	var schemas []string
	for rows.Next() {
		// Check context during iteration
		if err := collector.CheckContext(ctx, SourceName, "list_schemas"); err != nil {
			return nil, err
		}

		var schema string
		if err := rows.Scan(&schema); err != nil {
			return nil, collector.NewParseError(SourceName, "list_schemas", err)
		}
		schemas = append(schemas, schema)
	}

	if err := rows.Err(); err != nil {
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
		for _, s := range schemas {
			if ruleMatcher.Match(s) {
				filtered = append(filtered, s)
			}
		}
		schemas = filtered
	}

	return schemas, nil
}

// ListTables 列出表
func (c *Collector) ListTables(ctx context.Context, catalog, schema string, opts *collector.ListOptions) (*collector.TableListResult, error) {
	if c.db == nil {
		return nil, collector.NewConnectionClosedError(SourceName, "list_tables")
	}

	// Check context before starting operation
	if err := collector.CheckContext(ctx, SourceName, "list_tables"); err != nil {
		return nil, err
	}

	// Use the schema (database) in the query
	query := fmt.Sprintf("SHOW TABLES IN %s", schema)
	rows, err := c.db.QueryContext(ctx, query)
	if err != nil {
		if ctx.Err() != nil {
			return nil, collector.WrapContextError(ctx, SourceName, "list_tables")
		}
		return nil, collector.NewQueryError(SourceName, "list_tables", err)
	}
	defer rows.Close()

	var allTables []string
	for rows.Next() {
		// Check context during iteration
		if err := collector.CheckContext(ctx, SourceName, "list_tables"); err != nil {
			return nil, err
		}

		var tableName string
		if err := rows.Scan(&tableName); err != nil {
			return nil, collector.NewParseError(SourceName, "list_tables", err)
		}
		allTables = append(allTables, tableName)
	}

	if err := rows.Err(); err != nil {
		if ctx.Err() != nil {
			return nil, collector.WrapContextError(ctx, SourceName, "list_tables")
		}
		return nil, collector.NewQueryError(SourceName, "list_tables", err)
	}

	// Apply table matching filter
	tables := c.filterTables(allTables, opts)

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

// filterTables applies matching rules to filter tables
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


// FetchTableMetadata 获取表元数据 (使用 DESCRIBE FORMATTED)
func (c *Collector) FetchTableMetadata(ctx context.Context, catalog, schema, table string) (*collector.TableMetadata, error) {
	if c.db == nil {
		return nil, collector.NewConnectionClosedError(SourceName, "fetch_table_metadata")
	}

	// Check context before starting operation
	if err := collector.CheckContext(ctx, SourceName, "fetch_table_metadata"); err != nil {
		return nil, err
	}

	// Execute DESCRIBE FORMATTED to get full table metadata
	query := fmt.Sprintf("DESCRIBE FORMATTED %s.%s", schema, table)
	rows, err := c.db.QueryContext(ctx, query)
	if err != nil {
		if ctx.Err() != nil {
			return nil, collector.WrapContextError(ctx, SourceName, "fetch_table_metadata")
		}
		// Check if table not found
		if strings.Contains(err.Error(), "Table not found") || strings.Contains(err.Error(), "does not exist") {
			return nil, collector.NewNotFoundError(SourceName, "fetch_table_metadata", fmt.Sprintf("%s.%s", schema, table), nil)
		}
		return nil, collector.NewQueryError(SourceName, "fetch_table_metadata", err)
	}
	defer rows.Close()

	// Collect all rows from DESCRIBE FORMATTED output
	var describeOutput [][]string
	cols, err := rows.Columns()
	if err != nil {
		return nil, collector.NewParseError(SourceName, "fetch_table_metadata", err)
	}
	numCols := len(cols)

	for rows.Next() {
		// Check context during iteration
		if err := collector.CheckContext(ctx, SourceName, "fetch_table_metadata"); err != nil {
			return nil, err
		}

		values := make([]sql.NullString, numCols)
		valuePtrs := make([]interface{}, numCols)
		for i := range values {
			valuePtrs[i] = &values[i]
		}

		if err := rows.Scan(valuePtrs...); err != nil {
			return nil, collector.NewParseError(SourceName, "fetch_table_metadata", err)
		}

		row := make([]string, numCols)
		for i, v := range values {
			if v.Valid {
				row[i] = strings.TrimSpace(v.String)
			}
		}
		describeOutput = append(describeOutput, row)
	}

	if err := rows.Err(); err != nil {
		if ctx.Err() != nil {
			return nil, collector.WrapContextError(ctx, SourceName, "fetch_table_metadata")
		}
		return nil, collector.NewQueryError(SourceName, "fetch_table_metadata", err)
	}

	// Parse the DESCRIBE FORMATTED output
	metadata, err := ParseDescribeFormatted(describeOutput, catalog, schema, table)
	if err != nil {
		return nil, collector.NewParseError(SourceName, "fetch_table_metadata", err)
	}

	metadata.LastRefreshedAt = time.Now()
	return metadata, nil
}

// FetchTableStatistics 获取表统计信息
func (c *Collector) FetchTableStatistics(ctx context.Context, catalog, schema, table string) (*collector.TableStatistics, error) {
	if c.db == nil {
		return nil, collector.NewConnectionClosedError(SourceName, "fetch_table_statistics")
	}

	// Check context before starting operation
	if err := collector.CheckContext(ctx, SourceName, "fetch_table_statistics"); err != nil {
		return nil, err
	}

	// Try to get statistics from DESCRIBE FORMATTED
	query := fmt.Sprintf("DESCRIBE FORMATTED %s.%s", schema, table)
	rows, err := c.db.QueryContext(ctx, query)
	if err != nil {
		if ctx.Err() != nil {
			return nil, collector.WrapContextError(ctx, SourceName, "fetch_table_statistics")
		}
		if strings.Contains(err.Error(), "Table not found") || strings.Contains(err.Error(), "does not exist") {
			return nil, collector.NewNotFoundError(SourceName, "fetch_table_statistics", fmt.Sprintf("%s.%s", schema, table), nil)
		}
		return nil, collector.NewQueryError(SourceName, "fetch_table_statistics", err)
	}
	defer rows.Close()

	stats := &collector.TableStatistics{
		CollectedAt: time.Now(),
	}

	cols, _ := rows.Columns()
	numCols := len(cols)

	for rows.Next() {
		// Check context during iteration
		if err := collector.CheckContext(ctx, SourceName, "fetch_table_statistics"); err != nil {
			return nil, err
		}

		values := make([]sql.NullString, numCols)
		valuePtrs := make([]interface{}, numCols)
		for i := range values {
			valuePtrs[i] = &values[i]
		}

		if err := rows.Scan(valuePtrs...); err != nil {
			continue
		}

		if numCols >= 2 && values[0].Valid {
			key := strings.TrimSpace(values[0].String)
			value := ""
			if values[1].Valid {
				value = strings.TrimSpace(values[1].String)
			}

			switch {
			case strings.Contains(strings.ToLower(key), "numrows"):
				if n, err := strconv.ParseInt(value, 10, 64); err == nil {
					stats.RowCount = n
				}
			case strings.Contains(strings.ToLower(key), "rawdatasize") || strings.Contains(strings.ToLower(key), "totalsiz"):
				if n, err := strconv.ParseInt(value, 10, 64); err == nil {
					stats.DataSizeBytes = n
				}
			case strings.Contains(strings.ToLower(key), "numpartitions"):
				if n, err := strconv.Atoi(value); err == nil {
					stats.PartitionCount = n
				}
			}
		}
	}

	if err := rows.Err(); err != nil {
		if ctx.Err() != nil {
			return nil, collector.WrapContextError(ctx, SourceName, "fetch_table_statistics")
		}
	}

	return stats, nil
}

// FetchPartitions 获取分区信息 (使用 SHOW PARTITIONS)
func (c *Collector) FetchPartitions(ctx context.Context, catalog, schema, table string) ([]collector.PartitionInfo, error) {
	if c.db == nil {
		return nil, collector.NewConnectionClosedError(SourceName, "fetch_partitions")
	}

	// Check context before starting operation
	if err := collector.CheckContext(ctx, SourceName, "fetch_partitions"); err != nil {
		return nil, err
	}

	// First check if table is partitioned by getting partition columns
	descQuery := fmt.Sprintf("DESCRIBE FORMATTED %s.%s", schema, table)
	descRows, err := c.db.QueryContext(ctx, descQuery)
	if err != nil {
		if ctx.Err() != nil {
			return nil, collector.WrapContextError(ctx, SourceName, "fetch_partitions")
		}
		if strings.Contains(err.Error(), "Table not found") || strings.Contains(err.Error(), "does not exist") {
			return nil, collector.NewNotFoundError(SourceName, "fetch_partitions", fmt.Sprintf("%s.%s", schema, table), nil)
		}
		return nil, collector.NewQueryError(SourceName, "fetch_partitions", err)
	}

	var partitionColumns []string
	cols, _ := descRows.Columns()
	numCols := len(cols)
	inPartitionSection := false

	for descRows.Next() {
		// Check context during iteration
		if err := collector.CheckContext(ctx, SourceName, "fetch_partitions"); err != nil {
			descRows.Close()
			return nil, err
		}

		values := make([]sql.NullString, numCols)
		valuePtrs := make([]interface{}, numCols)
		for i := range values {
			valuePtrs[i] = &values[i]
		}

		if err := descRows.Scan(valuePtrs...); err != nil {
			continue
		}

		if numCols >= 1 && values[0].Valid {
			col0 := strings.TrimSpace(values[0].String)

			if strings.Contains(col0, "# Partition Information") {
				inPartitionSection = true
				continue
			}

			if inPartitionSection {
				if strings.HasPrefix(col0, "#") || col0 == "" {
					if strings.Contains(col0, "# Detailed Table Information") {
						break
					}
					continue
				}
				// This is a partition column
				partitionColumns = append(partitionColumns, col0)
			}
		}
	}
	descRows.Close()

	// If no partition columns found, return empty list
	if len(partitionColumns) == 0 {
		return []collector.PartitionInfo{}, nil
	}

	// Check context before getting partition values
	if err := collector.CheckContext(ctx, SourceName, "fetch_partitions"); err != nil {
		return nil, err
	}

	// Get partition values using SHOW PARTITIONS
	query := fmt.Sprintf("SHOW PARTITIONS %s.%s", schema, table)
	rows, err := c.db.QueryContext(ctx, query)
	if err != nil {
		if ctx.Err() != nil {
			return nil, collector.WrapContextError(ctx, SourceName, "fetch_partitions")
		}
		// Table might not be partitioned
		if strings.Contains(err.Error(), "not a partitioned table") || strings.Contains(err.Error(), "is not partitioned") {
			return []collector.PartitionInfo{}, nil
		}
		return nil, collector.NewQueryError(SourceName, "fetch_partitions", err)
	}
	defer rows.Close()

	partitionCount := 0
	for rows.Next() {
		// Check context during iteration
		if err := collector.CheckContext(ctx, SourceName, "fetch_partitions"); err != nil {
			return nil, err
		}
		partitionCount++
	}

	if err := rows.Err(); err != nil {
		if ctx.Err() != nil {
			return nil, collector.WrapContextError(ctx, SourceName, "fetch_partitions")
		}
		return nil, collector.NewQueryError(SourceName, "fetch_partitions", err)
	}

	// Return partition info with columns and count
	partitions := []collector.PartitionInfo{
		{
			Name:        "partitions",
			Type:        "LIST", // Hive uses list-style partitioning
			Columns:     partitionColumns,
			ValuesCount: partitionCount,
		},
	}

	return partitions, nil
}


// buildDSN constructs the Hive connection string from configuration
func (c *Collector) buildDSN() (string, error) {
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

	// Build connection string for gohive
	// Format: user:password@host:port/database?auth=NONE|KERBEROS|LDAP
	user := c.config.Credentials.User
	if user == "" {
		user = "hive"
	}
	password := c.config.Credentials.Password

	// Get database from extra properties, default to "default"
	database := "default"
	if c.config.Properties.Extra != nil {
		if db, ok := c.config.Properties.Extra["database"]; ok && db != "" {
			database = db
		}
	}

	// Get auth mode from extra properties, default to "NONE"
	authMode := "NONE"
	if c.config.Properties.Extra != nil {
		if auth, ok := c.config.Properties.Extra["auth"]; ok && auth != "" {
			authMode = strings.ToUpper(auth)
		}
	}

	// Build DSN
	var dsn string
	if password != "" {
		dsn = fmt.Sprintf("%s:%s@%s:%d/%s?auth=%s", user, password, host, port, database, authMode)
	} else {
		dsn = fmt.Sprintf("%s@%s:%d/%s?auth=%s", user, host, port, database, authMode)
	}

	// Add extra parameters
	if c.config.Properties.Extra != nil {
		for k, v := range c.config.Properties.Extra {
			if k != "database" && k != "auth" {
				dsn += fmt.Sprintf("&%s=%s", k, v)
			}
		}
	}

	return dsn, nil
}

// wrapConnectionError wraps a connection error with appropriate error type
func (c *Collector) wrapConnectionError(err error) error {
	errStr := err.Error()
	if strings.Contains(errStr, "authentication") || strings.Contains(errStr, "Authentication") ||
		strings.Contains(errStr, "Access denied") || strings.Contains(errStr, "Unauthorized") {
		return collector.NewAuthError(SourceName, "connect", err)
	}
	if strings.Contains(errStr, "connection refused") || strings.Contains(errStr, "no such host") ||
		strings.Contains(errStr, "no route to host") {
		return collector.NewNetworkError(SourceName, "connect", err)
	}
	if strings.Contains(errStr, "timeout") || strings.Contains(errStr, "deadline exceeded") {
		return collector.NewTimeoutError(SourceName, "connect", err)
	}
	return collector.NewNetworkError(SourceName, "connect", err)
}

// mapTableType maps Hive table type to standard TableType
func (c *Collector) mapTableType(hiveType string) collector.TableType {
	switch strings.ToUpper(hiveType) {
	case "VIRTUAL_VIEW", "VIEW":
		return collector.TableTypeView
	case "EXTERNAL_TABLE":
		return collector.TableTypeExternalTable
	case "MATERIALIZED_VIEW":
		return collector.TableTypeMaterializedView
	case "MANAGED_TABLE", "TABLE":
		return collector.TableTypeTable
	default:
		return collector.TableTypeTable
	}
}

// normalizeType normalizes Hive data type to standard type
func (c *Collector) normalizeType(dataType string) string {
	// Remove any parameters from type (e.g., varchar(100) -> varchar)
	baseType := strings.ToLower(dataType)
	if idx := strings.Index(baseType, "("); idx != -1 {
		baseType = baseType[:idx]
	}
	if idx := strings.Index(baseType, "<"); idx != -1 {
		baseType = baseType[:idx]
	}
	baseType = strings.TrimSpace(baseType)

	switch baseType {
	case "tinyint", "smallint", "int", "integer", "bigint":
		return "INTEGER"
	case "float", "double", "double precision":
		return "FLOAT"
	case "decimal", "numeric":
		return "DECIMAL"
	case "string", "varchar", "char":
		return "STRING"
	case "date":
		return "DATE"
	case "timestamp", "timestamp with local time zone":
		return "TIMESTAMP"
	case "binary":
		return "BINARY"
	case "boolean":
		return "BOOLEAN"
	case "array":
		return "ARRAY"
	case "map":
		return "MAP"
	case "struct":
		return "STRUCT"
	case "uniontype":
		return "UNION"
	default:
		return strings.ToUpper(baseType)
	}
}

// Category 返回数据源类别
func (c *Collector) Category() collector.DataSourceCategory {
	return collector.CategoryDataWarehouse
}

// Type 返回数据源类型
func (c *Collector) Type() string {
	return SourceName
}

// Ensure Collector implements collector.Collector interface
var _ collector.Collector = (*Collector)(nil)
