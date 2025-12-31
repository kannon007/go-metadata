// Package clickhouse provides ClickHouse metadata collector implementation.
package clickhouse

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"
	"strings"
	"time"

	"go-metadata/internal/collector"
	"go-metadata/internal/collector/config"

	_ "github.com/ClickHouse/clickhouse-go/v2"
)

const (
	// SourceName is the identifier for ClickHouse collector
	SourceName = "clickhouse"
)

// Collector implements the collector.Collector interface for ClickHouse.
type Collector struct {
	config *config.ConnectorConfig
	db     *sql.DB
}

// NewCollector creates a new ClickHouse collector instance.
func NewCollector(cfg *config.ConnectorConfig) (*Collector, error) {
	if cfg == nil {
		return nil, collector.NewInvalidConfigErrorWithCategory(collector.CategoryDataWarehouse, SourceName, "config", "config cannot be nil")
	}

	return &Collector{
		config: cfg,
	}, nil
}

// Connect establishes a connection to the ClickHouse database.
func (c *Collector) Connect(ctx context.Context) error {
	if c.db != nil {
		return nil // Already connected
	}

	dsn := c.buildDSN()
	db, err := sql.Open("clickhouse", dsn)
	if err != nil {
		return collector.NewNetworkErrorWithCategory(collector.CategoryDataWarehouse, SourceName, "connect", err)
	}

	// Test the connection
	if err := db.PingContext(ctx); err != nil {
		db.Close()
		return collector.NewNetworkErrorWithCategory(collector.CategoryDataWarehouse, SourceName, "connect", err)
	}

	c.db = db
	return nil
}

// Close closes the database connection.
func (c *Collector) Close() error {
	if c.db != nil {
		err := c.db.Close()
		c.db = nil
		if err != nil {
			return collector.NewNetworkErrorWithCategory(collector.CategoryDataWarehouse, SourceName, "close", err)
		}
	}
	return nil
}

// HealthCheck verifies the database connection is healthy.
func (c *Collector) HealthCheck(ctx context.Context) (*collector.HealthStatus, error) {
	if c.db == nil {
		return &collector.HealthStatus{
			Connected: false,
			Message:   "not connected to ClickHouse database",
		}, nil
	}

	start := time.Now()
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if err := c.db.PingContext(ctx); err != nil {
		return &collector.HealthStatus{
			Connected: false,
			Latency:   time.Since(start),
			Message:   err.Error(),
		}, nil
	}

	// Get version information
	var version string
	err := c.db.QueryRowContext(ctx, "SELECT version()").Scan(&version)
	if err != nil {
		version = "unknown"
	}

	return &collector.HealthStatus{
		Connected: true,
		Latency:   time.Since(start),
		Version:   version,
		Message:   "healthy",
	}, nil
}

// Category returns the data source category.
func (c *Collector) Category() collector.DataSourceCategory {
	return collector.CategoryDataWarehouse
}

// Type returns the collector type identifier.
func (c *Collector) Type() string {
	return SourceName
}

// DiscoverCatalogs discovers available catalogs (ClickHouse databases).
func (c *Collector) DiscoverCatalogs(ctx context.Context) ([]collector.CatalogInfo, error) {
	if c.db == nil {
		return nil, collector.NewConnectionClosedErrorWithCategory(collector.CategoryDataWarehouse, SourceName, "discover_catalogs")
	}

	databases, err := c.ListSchemas(ctx, "")
	if err != nil {
		return nil, err
	}

	var catalogs []collector.CatalogInfo
	for _, db := range databases {
		catalogs = append(catalogs, collector.CatalogInfo{
			Catalog:     db,
			Type:        "clickhouse",
			Description: "ClickHouse Database",
		})
	}

	return catalogs, nil
}

// ListSchemas lists all databases in ClickHouse.
func (c *Collector) ListSchemas(ctx context.Context, catalog string) ([]string, error) {
	if c.db == nil {
		return nil, collector.NewConnectionClosedErrorWithCategory(collector.CategoryDataWarehouse, SourceName, "list_schemas")
	}

	rows, err := c.db.QueryContext(ctx, GetDatabasesQuery())
	if err != nil {
		return nil, collector.NewQueryErrorWithCategory(collector.CategoryDataWarehouse, SourceName, "list_schemas", err)
	}
	defer rows.Close()

	var databases []string
	for rows.Next() {
		var database string
		if err := rows.Scan(&database); err != nil {
			return nil, collector.NewQueryErrorWithCategory(collector.CategoryDataWarehouse, SourceName, "list_schemas", err)
		}
		databases = append(databases, database)
	}

	if err := rows.Err(); err != nil {
		return nil, collector.NewQueryErrorWithCategory(collector.CategoryDataWarehouse, SourceName, "list_schemas", err)
	}

	return databases, nil
}

// ListTables lists all tables in the specified database.
func (c *Collector) ListTables(ctx context.Context, catalog, schema string, opts *collector.ListOptions) (*collector.TableListResult, error) {
	if c.db == nil {
		return nil, collector.NewConnectionClosedErrorWithCategory(collector.CategoryDataWarehouse, SourceName, "list_tables")
	}

	database := schema
	if database == "" {
		database = catalog
	}

	rows, err := c.db.QueryContext(ctx, GetTablesQuery(), database)
	if err != nil {
		return nil, collector.NewQueryErrorWithCategory(collector.CategoryDataWarehouse, SourceName, "list_tables", err)
	}
	defer rows.Close()

	var tables []string
	for rows.Next() {
		var tableName string
		if err := rows.Scan(&tableName); err != nil {
			return nil, collector.NewQueryErrorWithCategory(collector.CategoryDataWarehouse, SourceName, "list_tables", err)
		}
		tables = append(tables, tableName)
	}

	if err := rows.Err(); err != nil {
		return nil, collector.NewQueryErrorWithCategory(collector.CategoryDataWarehouse, SourceName, "list_tables", err)
	}

	return &collector.TableListResult{
		Tables:     tables,
		TotalCount: len(tables),
	}, nil
}

// FetchTableMetadata retrieves detailed metadata for a specific table.
func (c *Collector) FetchTableMetadata(ctx context.Context, catalog, schema, table string) (*collector.TableMetadata, error) {
	if c.db == nil {
		return nil, collector.NewConnectionClosedErrorWithCategory(collector.CategoryDataWarehouse, SourceName, "fetch_table_metadata")
	}

	database := schema
	if database == "" {
		database = catalog
	}

	metadata := &collector.TableMetadata{
		Catalog:        catalog,
		Schema:         database,
		Name:           table,
		SourceCategory: c.Category(),
		SourceType:     c.Type(),
		Type:           collector.TableTypeTable,
	}

	// Fetch columns
	columns, err := c.fetchColumns(ctx, database, table)
	if err != nil {
		return nil, err
	}
	metadata.Columns = columns

	// Fetch partitions if any
	partitions, err := c.fetchPartitions(ctx, database, table)
	if err != nil {
		return nil, err
	}
	metadata.Partitions = partitions

	return metadata, nil
}

// FetchTableStatistics retrieves table statistics.
func (c *Collector) FetchTableStatistics(ctx context.Context, catalog, schema, table string) (*collector.TableStatistics, error) {
	if c.db == nil {
		return nil, collector.NewConnectionClosedErrorWithCategory(collector.CategoryDataWarehouse, SourceName, "fetch_table_statistics")
	}

	database := schema
	if database == "" {
		database = catalog
	}

	var totalRows, totalBytes sql.NullInt64

	err := c.db.QueryRowContext(ctx, GetTableStatsQuery(), database, table).
		Scan(&totalRows, &totalBytes)
	if err != nil {
		return nil, collector.NewQueryErrorWithCategory(collector.CategoryDataWarehouse, SourceName, "fetch_table_statistics", err)
	}

	stats := &collector.TableStatistics{
		CollectedAt: time.Now(),
	}

	if totalRows.Valid {
		stats.RowCount = totalRows.Int64
	}

	if totalBytes.Valid {
		stats.DataSizeBytes = totalBytes.Int64
	}

	return stats, nil
}

// FetchPartitions retrieves partition information for a table.
func (c *Collector) FetchPartitions(ctx context.Context, catalog, schema, table string) ([]collector.PartitionInfo, error) {
	database := schema
	if database == "" {
		database = catalog
	}
	return c.fetchPartitions(ctx, database, table)
}

// buildDSN constructs the ClickHouse connection string.
func (c *Collector) buildDSN() string {
	// Parse endpoint (expected format: host:port or host)
	endpoint := c.config.Endpoint
	if endpoint == "" {
		endpoint = "localhost:9000"
	}

	host := endpoint
	port := 9000

	if idx := strings.LastIndex(endpoint, ":"); idx != -1 {
		host = endpoint[:idx]
		if p, err := strconv.Atoi(endpoint[idx+1:]); err == nil {
			port = p
		}
	}

	// Get database from extra properties
	database := "default"
	if c.config.Properties.Extra != nil {
		if db := c.config.Properties.Extra["database"]; db != "" {
			database = db
		}
	}

	// ClickHouse connection string format: clickhouse://user:password@host:port/database
	dsn := fmt.Sprintf("clickhouse://%s:%s@%s:%d/%s",
		c.config.Credentials.User,
		c.config.Credentials.Password,
		host,
		port,
		database,
	)

	// Add additional parameters
	params := []string{}
	
	// Connection timeout
	if c.config.Properties.ConnectionTimeout > 0 {
		params = append(params, fmt.Sprintf("dial_timeout=%ds", c.config.Properties.ConnectionTimeout))
	}

	// Add extra parameters
	if c.config.Properties.Extra != nil {
		for key, value := range c.config.Properties.Extra {
			if key != "database" {
				params = append(params, fmt.Sprintf("%s=%s", key, value))
			}
		}
	}

	if len(params) > 0 {
		dsn += "?" + strings.Join(params, "&")
	}

	return dsn
}

// fetchColumns retrieves column information for a table.
func (c *Collector) fetchColumns(ctx context.Context, database, table string) ([]collector.Column, error) {
	rows, err := c.db.QueryContext(ctx, GetColumnsQuery(), database, table)
	if err != nil {
		return nil, collector.NewQueryErrorWithCategory(collector.CategoryDataWarehouse, SourceName, "fetch_columns", err)
	}
	defer rows.Close()

	var columns []collector.Column
	position := 1
	for rows.Next() {
		var col collector.Column
		var dataType, defaultType, defaultExpression, comment string

		err := rows.Scan(
			&col.Name,
			&dataType,
			&defaultType,
			&defaultExpression,
			&comment,
		)
		if err != nil {
			return nil, collector.NewQueryErrorWithCategory(collector.CategoryDataWarehouse, SourceName, "fetch_columns", err)
		}

		col.OrdinalPosition = position
		col.Type = c.mapClickHouseTypeToSQL(dataType)
		col.SourceType = dataType
		col.Nullable = strings.Contains(strings.ToLower(dataType), "nullable")
		col.Comment = comment

		if defaultExpression != "" && defaultExpression != "NULL" {
			col.Default = &defaultExpression
		}

		columns = append(columns, col)
		position++
	}

	if err := rows.Err(); err != nil {
		return nil, collector.NewQueryErrorWithCategory(collector.CategoryDataWarehouse, SourceName, "fetch_columns", err)
	}

	return columns, nil
}

// fetchPartitions retrieves partition information for a table.
func (c *Collector) fetchPartitions(ctx context.Context, database, table string) ([]collector.PartitionInfo, error) {
	rows, err := c.db.QueryContext(ctx, GetPartitionsQuery(), database, table)
	if err != nil {
		return nil, collector.NewQueryErrorWithCategory(collector.CategoryDataWarehouse, SourceName, "fetch_partitions", err)
	}
	defer rows.Close()

	var partitions []collector.PartitionInfo
	for rows.Next() {
		var partition collector.PartitionInfo
		var partitionId, partitionKey string
		var rowCount, bytes sql.NullInt64

		err := rows.Scan(&partitionId, &partitionKey, &rowCount, &bytes)
		if err != nil {
			return nil, collector.NewQueryErrorWithCategory(collector.CategoryDataWarehouse, SourceName, "fetch_partitions", err)
		}

		partition.Name = partitionId
		partition.Expression = partitionKey
		partition.Type = "RANGE" // ClickHouse typically uses range partitioning

		if rowCount.Valid {
			partition.ValuesCount = int(rowCount.Int64)
		}

		partitions = append(partitions, partition)
	}

	if err := rows.Err(); err != nil {
		return nil, collector.NewQueryErrorWithCategory(collector.CategoryDataWarehouse, SourceName, "fetch_partitions", err)
	}

	return partitions, nil
}

// mapClickHouseTypeToSQL maps ClickHouse data types to standard SQL types.
func (c *Collector) mapClickHouseTypeToSQL(clickhouseType string) string {
	// Remove Nullable wrapper
	dataType := strings.TrimPrefix(clickhouseType, "Nullable(")
	dataType = strings.TrimSuffix(dataType, ")")
	
	// Convert to lowercase for comparison
	lowerType := strings.ToLower(dataType)

	switch {
	case strings.HasPrefix(lowerType, "int8"):
		return "TINYINT"
	case strings.HasPrefix(lowerType, "int16"):
		return "SMALLINT"
	case strings.HasPrefix(lowerType, "int32"):
		return "INTEGER"
	case strings.HasPrefix(lowerType, "int64"):
		return "BIGINT"
	case strings.HasPrefix(lowerType, "uint8"):
		return "TINYINT"
	case strings.HasPrefix(lowerType, "uint16"):
		return "SMALLINT"
	case strings.HasPrefix(lowerType, "uint32"):
		return "INTEGER"
	case strings.HasPrefix(lowerType, "uint64"):
		return "BIGINT"
	case strings.HasPrefix(lowerType, "float32"):
		return "REAL"
	case strings.HasPrefix(lowerType, "float64"):
		return "DOUBLE"
	case strings.HasPrefix(lowerType, "decimal"):
		return "DECIMAL"
	case strings.HasPrefix(lowerType, "string"):
		return "TEXT"
	case strings.HasPrefix(lowerType, "fixedstring"):
		return "CHAR"
	case strings.HasPrefix(lowerType, "datetime"):
		return "TIMESTAMP"
	case strings.HasPrefix(lowerType, "date"):
		return "DATE"
	case strings.HasPrefix(lowerType, "uuid"):
		return "UUID"
	case strings.HasPrefix(lowerType, "array"):
		return "ARRAY"
	case strings.HasPrefix(lowerType, "tuple"):
		return "STRUCT"
	case strings.HasPrefix(lowerType, "enum"):
		return "ENUM"
	case lowerType == "bool":
		return "BOOLEAN"
	default:
		return "TEXT"
	}
}