// Package sqlserver provides a SQL Server metadata collector implementation.
package sqlserver

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

	_ "github.com/denisenkom/go-mssqldb"
)

const (
	// SourceName identifies this collector type
	SourceName = "sqlserver"
	// DefaultPort is the default SQL Server port
	DefaultPort = 1433
	// DefaultTimeout is the default connection timeout in seconds
	DefaultTimeout = 30
)

// Collector SQL Server 元数据采集器
type Collector struct {
	config *config.ConnectorConfig
	db     *sql.DB
}

// NewCollector 创建 SQL Server 采集器实例
func NewCollector(cfg *config.ConnectorConfig) (collector.Collector, error) {
	if cfg == nil {
		return nil, collector.NewInvalidConfigError(SourceName, "config", "configuration cannot be nil")
	}
	if cfg.Type != "" && cfg.Type != SourceName {
		return nil, collector.NewInvalidConfigError(SourceName, "type", fmt.Sprintf("expected '%s', got '%s'", SourceName, cfg.Type))
	}
	return &Collector{config: cfg}, nil
}

// Connect 建立 SQL Server 连接
func (c *Collector) Connect(ctx context.Context) error {
	if c.db != nil {
		return nil // Already connected
	}

	dsn, err := c.buildDSN()
	if err != nil {
		return collector.NewInvalidConfigError(SourceName, "endpoint", err.Error())
	}

	db, err := sql.Open("sqlserver", dsn)
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

// Close 关闭 SQL Server 连接
func (c *Collector) Close() error {
	if c.db != nil {
		err := c.db.Close()
		c.db = nil
		return err
	}
	return nil
}

// Category 返回数据源类别
func (c *Collector) Category() collector.DataSourceCategory {
	return collector.CategoryRDBMS
}

// Type 返回采集器类型
func (c *Collector) Type() string {
	return SourceName
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

	// Get SQL Server version
	var version string
	err := c.db.QueryRowContext(ctx, "SELECT @@VERSION").Scan(&version)
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

// DiscoverCatalogs 发现 Catalog（SQL Server 中 catalog 等同于数据库）
func (c *Collector) DiscoverCatalogs(ctx context.Context) ([]collector.CatalogInfo, error) {
	if c.db == nil {
		return nil, collector.NewConnectionClosedError(SourceName, "discover_catalogs")
	}

	query := GetDatabasesQuery()
	rows, err := c.db.QueryContext(ctx, query)
	if err != nil {
		return nil, collector.NewQueryError(SourceName, "discover_catalogs", err)
	}
	defer rows.Close()

	var catalogs []collector.CatalogInfo
	for rows.Next() {
		var dbName, collation string
		var dbId int
		var createDate time.Time

		if err := rows.Scan(&dbId, &dbName, &createDate, &collation); err != nil {
			return nil, collector.NewQueryError(SourceName, "discover_catalogs", err)
		}

		catalogs = append(catalogs, collector.CatalogInfo{
			Catalog:     dbName,
			Type:        SourceName,
			Description: "SQL Server Database",
			Properties: map[string]string{
				"database_id": strconv.Itoa(dbId),
				"collation":   collation,
				"created":     createDate.Format(time.RFC3339),
			},
		})
	}

	if err := rows.Err(); err != nil {
		return nil, collector.NewQueryError(SourceName, "discover_catalogs", err)
	}

	return catalogs, nil
}

// ListSchemas 列出 Schema
func (c *Collector) ListSchemas(ctx context.Context, catalog string) ([]string, error) {
	if c.db == nil {
		return nil, collector.NewConnectionClosedError(SourceName, "list_schemas")
	}

	query := GetSchemasQuery()
	rows, err := c.db.QueryContext(ctx, query, catalog)
	if err != nil {
		return nil, collector.NewQueryError(SourceName, "list_schemas", err)
	}
	defer rows.Close()

	var schemas []string
	for rows.Next() {
		var schema string
		if err := rows.Scan(&schema); err != nil {
			return nil, collector.NewQueryError(SourceName, "list_schemas", err)
		}
		schemas = append(schemas, schema)
	}

	if err := rows.Err(); err != nil {
		return nil, collector.NewQueryError(SourceName, "list_schemas", err)
	}

	// Apply matching rules if configured
	if c.config.Matching != nil {
		matcher, err := matcher.NewRuleMatcher(c.config.Matching.Schemas, c.config.Matching.PatternType, c.config.Matching.CaseSensitive)
		if err != nil {
			return nil, collector.NewQueryErrorWithCategory(collector.CategoryRDBMS, SourceName, "list_schemas", err)
		}
		var filteredSchemas []string
		for _, schema := range schemas {
			if matcher.Match(schema) {
				filteredSchemas = append(filteredSchemas, schema)
			}
		}
		schemas = filteredSchemas
	}

	return schemas, nil
}

// ListTables 列出表
func (c *Collector) ListTables(ctx context.Context, catalog, schema string, opts *collector.ListOptions) (*collector.TableListResult, error) {
	if c.db == nil {
		return nil, collector.NewConnectionClosedError(SourceName, "list_tables")
	}

	query := GetTablesQuery()
	rows, err := c.db.QueryContext(ctx, query, catalog, schema)
	if err != nil {
		return nil, collector.NewQueryError(SourceName, "list_tables", err)
	}
	defer rows.Close()

	var tables []string
	for rows.Next() {
		var tableName string
		if err := rows.Scan(&tableName); err != nil {
			return nil, collector.NewQueryError(SourceName, "list_tables", err)
		}
		tables = append(tables, tableName)
	}

	if err := rows.Err(); err != nil {
		return nil, collector.NewQueryError(SourceName, "list_tables", err)
	}

	// Apply matching rules if configured
	if c.config.Matching != nil {
		matcher, err := matcher.NewRuleMatcher(c.config.Matching.Tables, c.config.Matching.PatternType, c.config.Matching.CaseSensitive)
		if err != nil {
			return nil, collector.NewQueryErrorWithCategory(collector.CategoryRDBMS, SourceName, "list_tables", err)
		}
		var filteredTables []string
		for _, table := range tables {
			if matcher.Match(table) {
				filteredTables = append(filteredTables, table)
			}
		}
		tables = filteredTables
	}

	return &collector.TableListResult{
		Tables: tables,
	}, nil
}

// FetchTableMetadata 获取表元数据
func (c *Collector) FetchTableMetadata(ctx context.Context, catalog, schema, table string) (*collector.TableMetadata, error) {
	if c.db == nil {
		return nil, collector.NewConnectionClosedError(SourceName, "fetch_table_metadata")
	}

	// Get table basic info
	tableInfo, err := c.getTableInfo(ctx, catalog, schema, table)
	if err != nil {
		return nil, err
	}

	// Get columns
	columns, err := c.getTableColumns(ctx, catalog, schema, table)
	if err != nil {
		return nil, err
	}

	// Get indexes
	indexes, err := c.getTableIndexes(ctx, catalog, schema, table)
	if err != nil {
		return nil, err
	}

	// Get primary key
	primaryKey, err := c.getTablePrimaryKey(ctx, catalog, schema, table)
	if err != nil {
		return nil, err
	}

	metadata := &collector.TableMetadata{
		SourceCategory:  c.Category(),
		SourceType:      c.Type(),
		Catalog:         catalog,
		Schema:          schema,
		Name:            table,
		Type:            tableInfo.Type,
		Comment:         tableInfo.Comment,
		Columns:         columns,
		Indexes:         indexes,
		PrimaryKey:      primaryKey,
		LastRefreshedAt: time.Now(),
	}

	return metadata, nil
}

// FetchTableStatistics 获取表统计信息
func (c *Collector) FetchTableStatistics(ctx context.Context, catalog, schema, table string) (*collector.TableStatistics, error) {
	if c.db == nil {
		return nil, collector.NewConnectionClosedError(SourceName, "fetch_table_statistics")
	}

	query := GetTableStatsQuery()
	var rowCount sql.NullInt64
	var dataSizeKB sql.NullInt64

	err := c.db.QueryRowContext(ctx, query, catalog, schema, table).
		Scan(&rowCount, &dataSizeKB)
	if err != nil {
		return nil, collector.NewQueryError(SourceName, "fetch_table_statistics", err)
	}

	stats := &collector.TableStatistics{
		CollectedAt: time.Now(),
	}

	if rowCount.Valid {
		stats.RowCount = rowCount.Int64
	}

	if dataSizeKB.Valid {
		// Convert KB to bytes
		stats.DataSizeBytes = dataSizeKB.Int64 * 1024
	}

	return stats, nil
}

// FetchPartitions 获取分区信息
func (c *Collector) FetchPartitions(ctx context.Context, catalog, schema, table string) ([]collector.PartitionInfo, error) {
	if c.db == nil {
		return nil, collector.NewConnectionClosedError(SourceName, "fetch_partitions")
	}

	query := GetPartitionsQuery()
	rows, err := c.db.QueryContext(ctx, query, catalog, schema, table)
	if err != nil {
		return nil, collector.NewQueryError(SourceName, "fetch_partitions", err)
	}
	defer rows.Close()

	var partitions []collector.PartitionInfo
	for rows.Next() {
		var partitionNumber int
		var partitionFunction, boundaryValue string
		var rowCount sql.NullInt64

		if err := rows.Scan(&partitionNumber, &partitionFunction, &boundaryValue, &rowCount); err != nil {
			return nil, collector.NewQueryError(SourceName, "fetch_partitions", err)
		}

		partition := collector.PartitionInfo{
			Name:        fmt.Sprintf("partition_%d", partitionNumber),
			Type:        partitionFunction,
			Expression:  boundaryValue,
			ValuesCount: 1,
		}

		if rowCount.Valid {
			// Note: PartitionInfo doesn't have RowCount field, 
			// this information could be stored in Properties if needed
			// partition.Properties = map[string]string{"row_count": strconv.FormatInt(rowCount.Int64, 10)}
		}

		partitions = append(partitions, partition)
	}

	if err := rows.Err(); err != nil {
		return nil, collector.NewQueryError(SourceName, "fetch_partitions", err)
	}

	return partitions, nil
}

// buildDSN 构建 SQL Server 连接字符串
func (c *Collector) buildDSN() (string, error) {
	if c.config.Endpoint == "" {
		return "", fmt.Errorf("endpoint is required")
	}

	// Parse endpoint
	parts := strings.Split(c.config.Endpoint, ":")
	host := parts[0]
	port := DefaultPort

	if len(parts) > 1 {
		if p, err := strconv.Atoi(parts[1]); err == nil {
			port = p
		}
	}

	// Build connection string
	// Format: sqlserver://username:password@host:port?database=dbname
	dsn := fmt.Sprintf("sqlserver://%s:%s@%s:%d",
		c.config.Credentials.User,
		c.config.Credentials.Password,
		host,
		port,
	)

	// Add database if specified
	database := ""
	if c.config.Properties.Extra != nil {
		database = c.config.Properties.Extra["database"]
	}
	if database != "" {
		dsn += "?database=" + database
	}

	return dsn, nil
}

// wrapConnectionError 包装连接错误
func (c *Collector) wrapConnectionError(err error) error {
	errStr := err.Error()
	if strings.Contains(errStr, "login failed") || strings.Contains(errStr, "authentication") {
		return collector.NewAuthError(SourceName, "connect", err)
	}
	if strings.Contains(errStr, "network") || strings.Contains(errStr, "connection refused") {
		return collector.NewNetworkError(SourceName, "connect", err)
	}
	if strings.Contains(errStr, "timeout") {
		return collector.NewTimeoutError(SourceName, "connect", err)
	}
	return collector.NewNetworkError(SourceName, "connect", err)
}
// Helper types for table information
type tableInfo struct {
	Type    collector.TableType
	Comment string
}

// getTableInfo 获取表基本信息
func (c *Collector) getTableInfo(ctx context.Context, catalog, schema, table string) (*tableInfo, error) {
	query := GetTableInfoQuery()
	var tableType, comment string

	err := c.db.QueryRowContext(ctx, query, catalog, schema, table).
		Scan(&tableType, &comment)
	if err != nil {
		return nil, collector.NewQueryError(SourceName, "get_table_info", err)
	}

	var tType collector.TableType
	switch strings.ToUpper(tableType) {
	case "VIEW":
		tType = collector.TableTypeView
	default:
		tType = collector.TableTypeTable
	}

	return &tableInfo{
		Type:    tType,
		Comment: comment,
	}, nil
}

// getTableColumns 获取表列信息
func (c *Collector) getTableColumns(ctx context.Context, catalog, schema, table string) ([]collector.Column, error) {
	query := GetColumnsQuery()
	rows, err := c.db.QueryContext(ctx, query, catalog, schema, table)
	if err != nil {
		return nil, collector.NewQueryError(SourceName, "get_table_columns", err)
	}
	defer rows.Close()

	var columns []collector.Column
	for rows.Next() {
		var columnName, dataType, isNullable, columnDefault, description string
		var ordinalPosition int
		var maxLength, numericPrecision, numericScale sql.NullInt64

		err := rows.Scan(&ordinalPosition, &columnName, &dataType, &maxLength,
			&numericPrecision, &numericScale, &isNullable, &columnDefault, &description)
		if err != nil {
			return nil, collector.NewQueryError(SourceName, "get_table_columns", err)
		}

		// Build source type with precision/scale
		sourceType := dataType
		if numericPrecision.Valid && numericScale.Valid {
			sourceType = fmt.Sprintf("%s(%d,%d)", dataType, numericPrecision.Int64, numericScale.Int64)
		} else if maxLength.Valid && maxLength.Int64 > 0 {
			sourceType = fmt.Sprintf("%s(%d)", dataType, maxLength.Int64)
		}

		var defaultValue *string
		if columnDefault != "" {
			defaultValue = &columnDefault
		}

		column := collector.Column{
			OrdinalPosition: ordinalPosition,
			Name:            columnName,
			Type:            c.mapSQLServerTypeToSQL(dataType),
			SourceType:      sourceType,
			Nullable:        strings.ToUpper(isNullable) == "YES",
			Default:         defaultValue,
			Comment:         description,
		}

		columns = append(columns, column)
	}

	if err := rows.Err(); err != nil {
		return nil, collector.NewQueryError(SourceName, "get_table_columns", err)
	}

	return columns, nil
}

// getTableIndexes 获取表索引信息
func (c *Collector) getTableIndexes(ctx context.Context, catalog, schema, table string) ([]collector.Index, error) {
	query := GetIndexesQuery()
	rows, err := c.db.QueryContext(ctx, query, catalog, schema, table)
	if err != nil {
		return nil, collector.NewQueryError(SourceName, "get_table_indexes", err)
	}
	defer rows.Close()

	indexMap := make(map[string]*collector.Index)
	for rows.Next() {
		var indexName, columnName string
		var isUnique bool
		var keyOrdinal int

		err := rows.Scan(&indexName, &columnName, &keyOrdinal, &isUnique)
		if err != nil {
			return nil, collector.NewQueryError(SourceName, "get_table_indexes", err)
		}

		if index, exists := indexMap[indexName]; exists {
			index.Columns = append(index.Columns, columnName)
		} else {
			indexMap[indexName] = &collector.Index{
				Name:    indexName,
				Columns: []string{columnName},
				Unique:  isUnique,
			}
		}
	}

	if err := rows.Err(); err != nil {
		return nil, collector.NewQueryError(SourceName, "get_table_indexes", err)
	}

	// Convert map to slice
	var indexes []collector.Index
	for _, index := range indexMap {
		indexes = append(indexes, *index)
	}

	return indexes, nil
}

// getTablePrimaryKey 获取表主键信息
func (c *Collector) getTablePrimaryKey(ctx context.Context, catalog, schema, table string) ([]string, error) {
	query := GetPrimaryKeyQuery()
	rows, err := c.db.QueryContext(ctx, query, catalog, schema, table)
	if err != nil {
		return nil, collector.NewQueryError(SourceName, "get_table_primary_key", err)
	}
	defer rows.Close()

	var primaryKey []string
	for rows.Next() {
		var columnName string
		if err := rows.Scan(&columnName); err != nil {
			return nil, collector.NewQueryError(SourceName, "get_table_primary_key", err)
		}
		primaryKey = append(primaryKey, columnName)
	}

	if err := rows.Err(); err != nil {
		return nil, collector.NewQueryError(SourceName, "get_table_primary_key", err)
	}

	return primaryKey, nil
}

// mapSQLServerTypeToSQL 将 SQL Server 数据类型映射到标准 SQL 类型
func (c *Collector) mapSQLServerTypeToSQL(sqlServerType string) string {
	switch strings.ToLower(sqlServerType) {
	case "varchar", "nvarchar", "char", "nchar", "text", "ntext":
		return "TEXT"
	case "int", "bigint", "smallint", "tinyint":
		return "INTEGER"
	case "decimal", "numeric", "money", "smallmoney":
		return "NUMERIC"
	case "float", "real":
		return "DOUBLE"
	case "bit":
		return "BOOLEAN"
	case "datetime", "datetime2", "smalldatetime", "date", "time", "datetimeoffset":
		return "TIMESTAMP"
	case "binary", "varbinary", "image":
		return "BLOB"
	case "uniqueidentifier":
		return "UUID"
	case "xml":
		return "XML"
	default:
		return "TEXT" // Default fallback
	}
}