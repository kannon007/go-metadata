// Package postgres provides a PostgreSQL metadata collector implementation.
package postgres

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

	_ "github.com/lib/pq"
)

const (
	// SourceName identifies this collector type
	SourceName = "postgres"
	// DefaultPort is the default PostgreSQL port
	DefaultPort = 5432
	// DefaultTimeout is the default connection timeout in seconds
	DefaultTimeout = 30
)

// Collector PostgreSQL 元数据采集器
type Collector struct {
	config *config.ConnectorConfig
	db     *sql.DB
}

// NewCollector 创建 PostgreSQL 采集器实例
func NewCollector(cfg *config.ConnectorConfig) (collector.Collector, error) {
	if cfg == nil {
		return nil, collector.NewInvalidConfigError(SourceName, "config", "configuration cannot be nil")
	}
	if cfg.Type != "" && cfg.Type != SourceName {
		return nil, collector.NewInvalidConfigError(SourceName, "type", fmt.Sprintf("expected '%s', got '%s'", SourceName, cfg.Type))
	}
	return &Collector{config: cfg}, nil
}

// Connect 建立 PostgreSQL 连接
func (c *Collector) Connect(ctx context.Context) error {
	if c.db != nil {
		return nil // Already connected
	}

	dsn, err := c.buildDSN()
	if err != nil {
		return collector.NewInvalidConfigError(SourceName, "endpoint", err.Error())
	}

	db, err := sql.Open("postgres", dsn)
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

// Close 关闭 PostgreSQL 连接
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

	// Get PostgreSQL version
	var version string
	err := c.db.QueryRowContext(ctx, "SELECT version()").Scan(&version)
	if err != nil {
		return &collector.HealthStatus{
			Connected: true,
			Latency:   time.Since(start),
			Message:   "connected but failed to get version: " + err.Error(),
		}, nil
	}

	return &collector.HealthStatus{
		Connected: true,
		Latency:   time.Since(start),
		Version:   version,
	}, nil
}


// DiscoverCatalogs 发现 Catalog（PostgreSQL 中 catalog 等同于数据库）
func (c *Collector) DiscoverCatalogs(ctx context.Context) ([]collector.CatalogInfo, error) {
	if c.db == nil {
		return nil, collector.NewConnectionClosedError(SourceName, "discover_catalogs")
	}

	// Get PostgreSQL version for catalog info
	var version string
	if err := c.db.QueryRowContext(ctx, "SELECT version()").Scan(&version); err != nil {
		return nil, collector.NewQueryError(SourceName, "discover_catalogs", err)
	}

	// Get current database name
	var currentDB string
	if err := c.db.QueryRowContext(ctx, "SELECT current_database()").Scan(&currentDB); err != nil {
		return nil, collector.NewQueryError(SourceName, "discover_catalogs", err)
	}

	// PostgreSQL connection is bound to a single database
	return []collector.CatalogInfo{
		{
			Catalog:     currentDB,
			Type:        SourceName,
			Description: "PostgreSQL Database",
			Properties: map[string]string{
				"version": version,
			},
		},
	}, nil
}

// ListSchemas 列出 Schema（PostgreSQL 有真正的 schema 概念）
func (c *Collector) ListSchemas(ctx context.Context, catalog string) ([]string, error) {
	if c.db == nil {
		return nil, collector.NewConnectionClosedError(SourceName, "list_schemas")
	}

	// Check context before starting operation
	if err := collector.CheckContext(ctx, SourceName, "list_schemas"); err != nil {
		return nil, err
	}

	rows, err := c.db.QueryContext(ctx, queryListSchemas)
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

	// Apply schema matching filter if configured
	if c.config.Matching != nil && c.config.Matching.Schemas != nil {
		ruleMatcher, err := matcher.NewRuleMatcher(
			c.config.Matching.Schemas,
			c.config.Matching.PatternType,
			c.config.Matching.CaseSensitive,
		)
		if err != nil {
			return nil, collector.NewInvalidConfigError(SourceName, "matching.schemas", err.Error())
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

	rows, err := c.db.QueryContext(ctx, queryListTables, schema)
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


// FetchTableMetadata 获取表元数据
func (c *Collector) FetchTableMetadata(ctx context.Context, catalog, schema, table string) (*collector.TableMetadata, error) {
	if c.db == nil {
		return nil, collector.NewConnectionClosedError(SourceName, "fetch_table_metadata")
	}

	// Check context before starting operation
	if err := collector.CheckContext(ctx, SourceName, "fetch_table_metadata"); err != nil {
		return nil, err
	}

	// Get table basic info
	var tableType string
	var comment sql.NullString
	err := c.db.QueryRowContext(ctx, queryGetTableInfo, schema, table).Scan(&tableType, &comment)
	if err != nil {
		if ctx.Err() != nil {
			return nil, collector.WrapContextError(ctx, SourceName, "fetch_table_metadata")
		}
		if err == sql.ErrNoRows {
			return nil, collector.NewNotFoundError(SourceName, "fetch_table_metadata", fmt.Sprintf("%s.%s", schema, table), nil)
		}
		return nil, collector.NewQueryError(SourceName, "fetch_table_metadata", err)
	}

	metadata := &collector.TableMetadata{
		Catalog:         catalog,
		Schema:          schema,
		Name:            table,
		Type:            c.mapTableType(tableType),
		Comment:         comment.String,
		LastRefreshedAt: time.Now(),
	}

	// Check context before fetching columns
	if err := collector.CheckContext(ctx, SourceName, "fetch_table_metadata"); err != nil {
		return nil, err
	}

	// Get columns
	columns, err := c.fetchColumns(ctx, schema, table)
	if err != nil {
		return nil, err
	}
	metadata.Columns = columns

	// Check context before fetching primary keys
	if err := collector.CheckContext(ctx, SourceName, "fetch_table_metadata"); err != nil {
		return nil, err
	}

	// Get primary keys
	primaryKeys, err := c.fetchPrimaryKeys(ctx, schema, table)
	if err != nil {
		return nil, err
	}
	metadata.PrimaryKey = primaryKeys

	// Mark primary key columns
	pkSet := make(map[string]bool)
	for _, pk := range primaryKeys {
		pkSet[pk] = true
	}
	for i := range metadata.Columns {
		if pkSet[metadata.Columns[i].Name] {
			metadata.Columns[i].IsPrimaryKey = true
		}
	}

	// Get indexes if configured
	if c.config.Collect == nil || c.config.Collect.Indexes {
		// Check context before fetching indexes
		if err := collector.CheckContext(ctx, SourceName, "fetch_table_metadata"); err != nil {
			return nil, err
		}

		indexes, err := c.fetchIndexes(ctx, schema, table)
		if err != nil {
			return nil, err
		}
		metadata.Indexes = indexes
	}

	return metadata, nil
}

// fetchColumns retrieves column information for a table
func (c *Collector) fetchColumns(ctx context.Context, schema, table string) ([]collector.Column, error) {
	// Check context before starting
	if err := collector.CheckContext(ctx, SourceName, "fetch_columns"); err != nil {
		return nil, err
	}

	rows, err := c.db.QueryContext(ctx, queryGetColumns, schema, table)
	if err != nil {
		if ctx.Err() != nil {
			return nil, collector.WrapContextError(ctx, SourceName, "fetch_columns")
		}
		return nil, collector.NewQueryError(SourceName, "fetch_columns", err)
	}
	defer rows.Close()

	var columns []collector.Column
	for rows.Next() {
		// Check context during iteration
		if err := collector.CheckContext(ctx, SourceName, "fetch_columns"); err != nil {
			return nil, err
		}

		var (
			ordinalPos                         int
			name, dataType, udtName            string
			charMaxLen, numPrecision, numScale sql.NullInt64
			isNullable                         string
			columnDefault, columnComment       sql.NullString
			isIdentity                         string
		)

		err := rows.Scan(
			&ordinalPos, &name, &dataType, &udtName,
			&charMaxLen, &numPrecision, &numScale,
			&isNullable, &columnDefault, &columnComment, &isIdentity,
		)
		if err != nil {
			return nil, collector.NewParseError(SourceName, "fetch_columns", err)
		}

		col := collector.Column{
			OrdinalPosition: ordinalPos,
			Name:            name,
			Type:            c.normalizeType(dataType),
			SourceType:      udtName,
			Nullable:        isNullable == "YES",
			Comment:         columnComment.String,
			IsAutoIncrement: isIdentity == "YES" || strings.HasPrefix(columnDefault.String, "nextval("),
		}

		if columnDefault.Valid {
			col.Default = &columnDefault.String
		}
		if charMaxLen.Valid {
			length := int(charMaxLen.Int64)
			col.Length = &length
		}
		if numPrecision.Valid {
			precision := int(numPrecision.Int64)
			col.Precision = &precision
		}
		if numScale.Valid {
			scale := int(numScale.Int64)
			col.Scale = &scale
		}

		columns = append(columns, col)
	}

	if err := rows.Err(); err != nil {
		if ctx.Err() != nil {
			return nil, collector.WrapContextError(ctx, SourceName, "fetch_columns")
		}
		return nil, collector.NewQueryError(SourceName, "fetch_columns", err)
	}

	return columns, nil
}

// fetchPrimaryKeys retrieves primary key columns for a table
func (c *Collector) fetchPrimaryKeys(ctx context.Context, schema, table string) ([]string, error) {
	// Check context before starting
	if err := collector.CheckContext(ctx, SourceName, "fetch_primary_keys"); err != nil {
		return nil, err
	}

	rows, err := c.db.QueryContext(ctx, queryGetPrimaryKeys, schema, table)
	if err != nil {
		if ctx.Err() != nil {
			return nil, collector.WrapContextError(ctx, SourceName, "fetch_primary_keys")
		}
		return nil, collector.NewQueryError(SourceName, "fetch_primary_keys", err)
	}
	defer rows.Close()

	var primaryKeys []string
	for rows.Next() {
		var columnName string
		if err := rows.Scan(&columnName); err != nil {
			return nil, collector.NewParseError(SourceName, "fetch_primary_keys", err)
		}
		primaryKeys = append(primaryKeys, columnName)
	}

	if err := rows.Err(); err != nil {
		if ctx.Err() != nil {
			return nil, collector.WrapContextError(ctx, SourceName, "fetch_primary_keys")
		}
		return nil, collector.NewQueryError(SourceName, "fetch_primary_keys", err)
	}

	return primaryKeys, nil
}

// fetchIndexes retrieves index information for a table
func (c *Collector) fetchIndexes(ctx context.Context, schema, table string) ([]collector.Index, error) {
	// Check context before starting
	if err := collector.CheckContext(ctx, SourceName, "fetch_indexes"); err != nil {
		return nil, err
	}

	rows, err := c.db.QueryContext(ctx, queryGetIndexes, schema, table)
	if err != nil {
		if ctx.Err() != nil {
			return nil, collector.WrapContextError(ctx, SourceName, "fetch_indexes")
		}
		return nil, collector.NewQueryError(SourceName, "fetch_indexes", err)
	}
	defer rows.Close()

	indexMap := make(map[string]*collector.Index)
	var indexOrder []string

	for rows.Next() {
		// Check context during iteration
		if err := collector.CheckContext(ctx, SourceName, "fetch_indexes"); err != nil {
			return nil, err
		}

		var (
			indexName, columnName string
			isUnique              bool
			indexType             sql.NullString
		)

		err := rows.Scan(&indexName, &columnName, &isUnique, &indexType)
		if err != nil {
			return nil, collector.NewParseError(SourceName, "fetch_indexes", err)
		}

		if idx, exists := indexMap[indexName]; exists {
			idx.Columns = append(idx.Columns, columnName)
		} else {
			indexMap[indexName] = &collector.Index{
				Name:    indexName,
				Columns: []string{columnName},
				Unique:  isUnique,
				Type:    indexType.String,
			}
			indexOrder = append(indexOrder, indexName)
		}
	}

	if err := rows.Err(); err != nil {
		if ctx.Err() != nil {
			return nil, collector.WrapContextError(ctx, SourceName, "fetch_indexes")
		}
		return nil, collector.NewQueryError(SourceName, "fetch_indexes", err)
	}

	var indexes []collector.Index
	for _, name := range indexOrder {
		indexes = append(indexes, *indexMap[name])
	}

	return indexes, nil
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

	var reltuples, relpages sql.NullFloat64
	err := c.db.QueryRowContext(ctx, queryGetTableStats, schema, table).Scan(&reltuples, &relpages)
	if err != nil {
		if ctx.Err() != nil {
			return nil, collector.WrapContextError(ctx, SourceName, "fetch_table_statistics")
		}
		if err == sql.ErrNoRows {
			return nil, collector.NewNotFoundError(SourceName, "fetch_table_statistics", fmt.Sprintf("%s.%s", schema, table), nil)
		}
		return nil, collector.NewQueryError(SourceName, "fetch_table_statistics", err)
	}

	// PostgreSQL stores row count as float (estimate), convert to int64
	rowCount := int64(0)
	if reltuples.Valid && reltuples.Float64 > 0 {
		rowCount = int64(reltuples.Float64)
	}

	// Calculate approximate data size (pages * 8KB default page size)
	dataSizeBytes := int64(0)
	if relpages.Valid && relpages.Float64 > 0 {
		dataSizeBytes = int64(relpages.Float64) * 8192
	}

	stats := &collector.TableStatistics{
		RowCount:      rowCount,
		DataSizeBytes: dataSizeBytes,
		CollectedAt:   time.Now(),
	}

	return stats, nil
}

// FetchPartitions 获取分区信息
func (c *Collector) FetchPartitions(ctx context.Context, catalog, schema, table string) ([]collector.PartitionInfo, error) {
	if c.db == nil {
		return nil, collector.NewConnectionClosedError(SourceName, "fetch_partitions")
	}

	// Check context before starting operation
	if err := collector.CheckContext(ctx, SourceName, "fetch_partitions"); err != nil {
		return nil, err
	}

	rows, err := c.db.QueryContext(ctx, queryGetPartitions, schema, table)
	if err != nil {
		if ctx.Err() != nil {
			return nil, collector.WrapContextError(ctx, SourceName, "fetch_partitions")
		}
		return nil, collector.NewQueryError(SourceName, "fetch_partitions", err)
	}
	defer rows.Close()

	var partitions []collector.PartitionInfo
	for rows.Next() {
		// Check context during iteration
		if err := collector.CheckContext(ctx, SourceName, "fetch_partitions"); err != nil {
			return nil, err
		}

		var (
			partitionName   string
			partitionMethod sql.NullString
			partitionExpr   sql.NullString
		)

		err := rows.Scan(&partitionName, &partitionMethod, &partitionExpr)
		if err != nil {
			return nil, collector.NewParseError(SourceName, "fetch_partitions", err)
		}

		partitions = append(partitions, collector.PartitionInfo{
			Name:       partitionName,
			Type:       partitionMethod.String,
			Expression: partitionExpr.String,
		})
	}

	if err := rows.Err(); err != nil {
		if ctx.Err() != nil {
			return nil, collector.WrapContextError(ctx, SourceName, "fetch_partitions")
		}
		return nil, collector.NewQueryError(SourceName, "fetch_partitions", err)
	}

	return partitions, nil
}

// buildDSN constructs the PostgreSQL connection string from configuration
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

	// Build connection string
	user := c.config.Credentials.User
	password := c.config.Credentials.Password

	timeout := DefaultTimeout
	if c.config.Properties.ConnectionTimeout > 0 {
		timeout = c.config.Properties.ConnectionTimeout
	}

	// Get database from extra properties, default to "postgres"
	database := "postgres"
	if c.config.Properties.Extra != nil {
		if db, ok := c.config.Properties.Extra["database"]; ok && db != "" {
			database = db
		}
	}

	// Get SSL mode from extra properties, default to "disable"
	sslmode := "disable"
	if c.config.Properties.Extra != nil {
		if ssl, ok := c.config.Properties.Extra["sslmode"]; ok && ssl != "" {
			sslmode = ssl
		}
	}

	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s connect_timeout=%d",
		host, port, user, password, database, sslmode, timeout)

	// Add extra parameters
	if c.config.Properties.Extra != nil {
		for k, v := range c.config.Properties.Extra {
			if k != "database" && k != "sslmode" {
				dsn += fmt.Sprintf(" %s=%s", k, v)
			}
		}
	}

	return dsn, nil
}

// wrapConnectionError wraps a connection error with appropriate error type
func (c *Collector) wrapConnectionError(err error) error {
	errStr := err.Error()
	if strings.Contains(errStr, "password authentication failed") || strings.Contains(errStr, "authentication failed") {
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

// mapTableType maps PostgreSQL table type to standard TableType
func (c *Collector) mapTableType(pgType string) collector.TableType {
	switch strings.ToUpper(pgType) {
	case "VIEW":
		return collector.TableTypeView
	case "BASE TABLE":
		return collector.TableTypeTable
	case "MATERIALIZED VIEW":
		return collector.TableTypeMaterializedView
	case "FOREIGN TABLE":
		return collector.TableTypeExternalTable
	default:
		return collector.TableTypeTable
	}
}

// normalizeType normalizes PostgreSQL data type to standard type
func (c *Collector) normalizeType(dataType string) string {
	switch strings.ToLower(dataType) {
	case "integer", "smallint", "bigint", "int2", "int4", "int8":
		return "INTEGER"
	case "real", "double precision", "float4", "float8":
		return "FLOAT"
	case "numeric", "decimal":
		return "DECIMAL"
	case "character", "character varying", "text", "char", "varchar":
		return "STRING"
	case "date":
		return "DATE"
	case "time", "time without time zone", "time with time zone":
		return "TIME"
	case "timestamp", "timestamp without time zone", "timestamp with time zone":
		return "TIMESTAMP"
	case "bytea":
		return "BINARY"
	case "boolean":
		return "BOOLEAN"
	case "json", "jsonb":
		return "JSON"
	case "uuid":
		return "UUID"
	case "array":
		return "ARRAY"
	default:
		return strings.ToUpper(dataType)
	}
}

// Category 返回数据源类别
func (c *Collector) Category() collector.DataSourceCategory {
	return collector.CategoryRDBMS
}

// Type 返回数据源类型
func (c *Collector) Type() string {
	return SourceName
}

// Ensure Collector implements collector.Collector interface
var _ collector.Collector = (*Collector)(nil)
