// Package doris provides Doris metadata collector implementation.
package doris

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"
	"strings"
	"time"

	"go-metadata/internal/collector"
	"go-metadata/internal/collector/config"

	_ "github.com/go-sql-driver/mysql"
)

const (
	// SourceName is the identifier for Doris collector
	SourceName = "doris"
)

// Collector implements the collector.Collector interface for Doris.
type Collector struct {
	config *config.ConnectorConfig
	db     *sql.DB
}

// NewCollector creates a new Doris collector instance.
func NewCollector(cfg *config.ConnectorConfig) (*Collector, error) {
	if cfg == nil {
		return nil, collector.NewInvalidConfigErrorWithCategory(collector.CategoryDataWarehouse, SourceName, "config", "config cannot be nil")
	}

	return &Collector{
		config: cfg,
	}, nil
}

// Connect establishes a connection to the Doris database.
func (c *Collector) Connect(ctx context.Context) error {
	if c.db != nil {
		return nil // Already connected
	}

	dsn := c.buildDSN()
	db, err := sql.Open("mysql", dsn)
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
			Message:   "not connected to Doris database",
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
	err := c.db.QueryRowContext(ctx, "SELECT @@version").Scan(&version)
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

// DiscoverCatalogs discovers available catalogs (Doris databases).
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
			Type:        "doris",
			Description: "Doris Database",
		})
	}

	return catalogs, nil
}

// ListSchemas lists all databases in Doris.
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

	var tableRows, dataLength sql.NullInt64

	err := c.db.QueryRowContext(ctx, GetTableStatsQuery(), database, table).
		Scan(&tableRows, &dataLength)
	if err != nil {
		return nil, collector.NewQueryErrorWithCategory(collector.CategoryDataWarehouse, SourceName, "fetch_table_statistics", err)
	}

	stats := &collector.TableStatistics{
		CollectedAt: time.Now(),
	}

	if tableRows.Valid {
		stats.RowCount = tableRows.Int64
	}

	if dataLength.Valid {
		stats.DataSizeBytes = dataLength.Int64
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

// buildDSN constructs the Doris connection string.
func (c *Collector) buildDSN() string {
	// Parse endpoint (expected format: host:port or host)
	endpoint := c.config.Endpoint
	if endpoint == "" {
		endpoint = "localhost:9030"
	}

	host := endpoint
	port := 9030

	if idx := strings.LastIndex(endpoint, ":"); idx != -1 {
		host = endpoint[:idx]
		if p, err := strconv.Atoi(endpoint[idx+1:]); err == nil {
			port = p
		}
	}

	// Get database from extra properties
	database := ""
	if c.config.Properties.Extra != nil {
		database = c.config.Properties.Extra["database"]
	}

	// Doris connection string format (MySQL compatible): user:password@tcp(host:port)/database
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s",
		c.config.Credentials.User,
		c.config.Credentials.Password,
		host,
		port,
		database,
	)

	// Add additional parameters
	params := []string{"parseTime=true"}
	
	// Connection timeout
	if c.config.Properties.ConnectionTimeout > 0 {
		params = append(params, fmt.Sprintf("timeout=%ds", c.config.Properties.ConnectionTimeout))
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
	for rows.Next() {
		var col collector.Column
		var nullable, defaultValue, comment string

		err := rows.Scan(
			&col.OrdinalPosition,
			&col.Name,
			&col.SourceType,
			&nullable,
			&defaultValue,
			&comment,
		)
		if err != nil {
			return nil, collector.NewQueryErrorWithCategory(collector.CategoryDataWarehouse, SourceName, "fetch_columns", err)
		}

		col.Type = c.mapDorisTypeToSQL(col.SourceType)
		col.Nullable = strings.ToUpper(nullable) == "YES"
		col.Comment = comment

		if defaultValue != "" && defaultValue != "NULL" {
			col.Default = &defaultValue
		}

		columns = append(columns, col)
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
		var partitionName, partitionDescription string

		err := rows.Scan(&partitionName, &partitionDescription)
		if err != nil {
			return nil, collector.NewQueryErrorWithCategory(collector.CategoryDataWarehouse, SourceName, "fetch_partitions", err)
		}

		partition.Name = partitionName
		partition.Expression = partitionDescription
		partition.Type = "RANGE" // Doris typically uses range partitioning

		partitions = append(partitions, partition)
	}

	if err := rows.Err(); err != nil {
		return nil, collector.NewQueryErrorWithCategory(collector.CategoryDataWarehouse, SourceName, "fetch_partitions", err)
	}

	return partitions, nil
}

// mapDorisTypeToSQL maps Doris data types to standard SQL types.
func (c *Collector) mapDorisTypeToSQL(dorisType string) string {
	// Convert to lowercase for comparison
	lowerType := strings.ToLower(dorisType)

	switch {
	case strings.HasPrefix(lowerType, "tinyint"):
		return "TINYINT"
	case strings.HasPrefix(lowerType, "smallint"):
		return "SMALLINT"
	case strings.HasPrefix(lowerType, "int"):
		return "INTEGER"
	case strings.HasPrefix(lowerType, "bigint"):
		return "BIGINT"
	case strings.HasPrefix(lowerType, "largeint"):
		return "BIGINT"
	case strings.HasPrefix(lowerType, "float"):
		return "REAL"
	case strings.HasPrefix(lowerType, "double"):
		return "DOUBLE"
	case strings.HasPrefix(lowerType, "decimal"):
		return "DECIMAL"
	case strings.HasPrefix(lowerType, "char"):
		return "CHAR"
	case strings.HasPrefix(lowerType, "varchar"):
		return "VARCHAR"
	case strings.HasPrefix(lowerType, "string"):
		return "TEXT"
	case strings.HasPrefix(lowerType, "text"):
		return "TEXT"
	case strings.HasPrefix(lowerType, "datetime"):
		return "TIMESTAMP"
	case strings.HasPrefix(lowerType, "timestamp"):
		return "TIMESTAMP"
	case strings.HasPrefix(lowerType, "date"):
		return "DATE"
	case lowerType == "boolean":
		return "BOOLEAN"
	case strings.HasPrefix(lowerType, "array"):
		return "ARRAY"
	case strings.HasPrefix(lowerType, "map"):
		return "MAP"
	case strings.HasPrefix(lowerType, "struct"):
		return "STRUCT"
	case strings.HasPrefix(lowerType, "json"):
		return "JSON"
	case strings.HasPrefix(lowerType, "bitmap"):
		return "BLOB"
	case strings.HasPrefix(lowerType, "hll"):
		return "BLOB"
	default:
		return "TEXT"
	}
}