// Package oracle provides Oracle Database metadata collector implementation.
package oracle

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"
	"strings"
	"time"

	"go-metadata/internal/collector"
	"go-metadata/internal/collector/config"

	_ "github.com/godror/godror"
)

const (
	// SourceName is the identifier for Oracle collector
	SourceName = "oracle"
)

// Collector implements the collector.Collector interface for Oracle Database.
type Collector struct {
	config *config.ConnectorConfig
	db     *sql.DB
}

// NewCollector creates a new Oracle collector instance.
func NewCollector(cfg *config.ConnectorConfig) (*Collector, error) {
	if cfg == nil {
		return nil, collector.NewInvalidConfigErrorWithCategory(collector.CategoryRDBMS, SourceName, "config", "config cannot be nil")
	}

	return &Collector{
		config: cfg,
	}, nil
}

// Connect establishes a connection to the Oracle database.
func (c *Collector) Connect(ctx context.Context) error {
	if c.db != nil {
		return nil // Already connected
	}

	dsn := c.buildDSN()
	db, err := sql.Open("godror", dsn)
	if err != nil {
		return collector.NewNetworkErrorWithCategory(collector.CategoryRDBMS, SourceName, "connect", err)
	}

	// Test the connection
	if err := db.PingContext(ctx); err != nil {
		db.Close()
		return collector.NewNetworkErrorWithCategory(collector.CategoryRDBMS, SourceName, "connect", err)
	}

	c.db = db
	return nil
}

// Disconnect closes the database connection.
func (c *Collector) Disconnect() error {
	return c.Close()
}

// Close closes the database connection.
func (c *Collector) Close() error {
	if c.db != nil {
		err := c.db.Close()
		c.db = nil
		if err != nil {
			return collector.NewNetworkErrorWithCategory(collector.CategoryRDBMS, SourceName, "close", err)
		}
	}
	return nil
}

// HealthCheck verifies the database connection is healthy.
func (c *Collector) HealthCheck(ctx context.Context) (*collector.HealthStatus, error) {
	if c.db == nil {
		return &collector.HealthStatus{
			Connected: false,
			Message:   "not connected to Oracle database",
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

	return &collector.HealthStatus{
		Connected: true,
		Latency:   time.Since(start),
		Message:   "healthy",
	}, nil
}

// Category returns the data source category.
func (c *Collector) Category() collector.DataSourceCategory {
	return collector.CategoryRDBMS
}

// Type returns the collector type identifier.
func (c *Collector) Type() string {
	return SourceName
}

// DiscoverCatalogs discovers available catalogs (Oracle instances).
func (c *Collector) DiscoverCatalogs(ctx context.Context) ([]collector.CatalogInfo, error) {
	if c.db == nil {
		return nil, collector.NewConnectionClosedErrorWithCategory(collector.CategoryRDBMS, SourceName, "discover_catalogs")
	}

	// In Oracle, there's typically one catalog per instance
	// We can get the database name from V$DATABASE
	var dbName string
	err := c.db.QueryRowContext(ctx, "SELECT NAME FROM V$DATABASE").Scan(&dbName)
	if err != nil {
		return nil, collector.NewQueryErrorWithCategory(collector.CategoryRDBMS, SourceName, "discover_catalogs", err)
	}

	return []collector.CatalogInfo{
		{
			Catalog:     dbName,
			Type:        "oracle",
			Description: "Oracle Database Instance",
		},
	}, nil
}

// ListSchemas lists all schemas (users) in the specified catalog.
func (c *Collector) ListSchemas(ctx context.Context, catalog string) ([]string, error) {
	return c.FetchDatabases(ctx)
}

// ListTables lists all tables in the specified schema.
func (c *Collector) ListTables(ctx context.Context, catalog, schema string, opts *collector.ListOptions) (*collector.TableListResult, error) {
	tables, err := c.FetchTables(ctx, schema)
	if err != nil {
		return nil, err
	}

	return &collector.TableListResult{
		Tables:     tables,
		TotalCount: len(tables),
	}, nil
}

// FetchTableStatistics retrieves table statistics.
func (c *Collector) FetchTableStatistics(ctx context.Context, catalog, schema, table string) (*collector.TableStatistics, error) {
	if c.db == nil {
		return nil, collector.NewConnectionClosedErrorWithCategory(collector.CategoryRDBMS, SourceName, "fetch_table_statistics")
	}

	var numRows, blocks, avgRowLen sql.NullInt64
	var lastAnalyzed sql.NullTime

	err := c.db.QueryRowContext(ctx, GetTableStatsQuery(), strings.ToUpper(schema), strings.ToUpper(table)).
		Scan(&numRows, &blocks, &avgRowLen, &lastAnalyzed)
	if err != nil {
		return nil, collector.NewQueryErrorWithCategory(collector.CategoryRDBMS, SourceName, "fetch_table_statistics", err)
	}

	stats := &collector.TableStatistics{
		CollectedAt: time.Now(),
	}

	if numRows.Valid {
		stats.RowCount = numRows.Int64
	}

	if blocks.Valid && avgRowLen.Valid {
		// Estimate data size: blocks * block_size (typically 8KB)
		stats.DataSizeBytes = blocks.Int64 * 8192
	}

	return stats, nil
}

// FetchPartitions retrieves partition information for a table.
func (c *Collector) FetchPartitions(ctx context.Context, catalog, schema, table string) ([]collector.PartitionInfo, error) {
	return c.fetchPartitions(ctx, schema, table)
}

// FetchDatabases retrieves all accessible databases/schemas.
func (c *Collector) FetchDatabases(ctx context.Context) ([]string, error) {
	if c.db == nil {
		return nil, collector.NewConnectionClosedErrorWithCategory(collector.CategoryRDBMS, SourceName, "fetch_databases")
	}

	rows, err := c.db.QueryContext(ctx, GetAllUsersQuery())
	if err != nil {
		return nil, collector.NewQueryErrorWithCategory(collector.CategoryRDBMS, SourceName, "fetch_databases", err)
	}
	defer rows.Close()

	var databases []string
	for rows.Next() {
		var username string
		if err := rows.Scan(&username); err != nil {
			return nil, collector.NewQueryErrorWithCategory(collector.CategoryRDBMS, SourceName, "fetch_databases", err)
		}
		databases = append(databases, username)
	}

	if err := rows.Err(); err != nil {
		return nil, collector.NewQueryErrorWithCategory(collector.CategoryRDBMS, SourceName, "fetch_databases", err)
	}

	return databases, nil
}

// FetchTables retrieves all tables in the specified database/schema.
func (c *Collector) FetchTables(ctx context.Context, database string) ([]string, error) {
	if c.db == nil {
		return nil, collector.NewConnectionClosedErrorWithCategory(collector.CategoryRDBMS, SourceName, "fetch_tables")
	}

	rows, err := c.db.QueryContext(ctx, GetAllTablesQuery(), strings.ToUpper(database))
	if err != nil {
		return nil, collector.NewQueryErrorWithCategory(collector.CategoryRDBMS, SourceName, "fetch_tables", err)
	}
	defer rows.Close()

	var tables []string
	for rows.Next() {
		var tableName string
		if err := rows.Scan(&tableName); err != nil {
			return nil, collector.NewQueryErrorWithCategory(collector.CategoryRDBMS, SourceName, "fetch_tables", err)
		}
		tables = append(tables, tableName)
	}

	if err := rows.Err(); err != nil {
		return nil, collector.NewQueryErrorWithCategory(collector.CategoryRDBMS, SourceName, "fetch_tables", err)
	}

	return tables, nil
}

// FetchTableMetadata retrieves detailed metadata for a specific table.
func (c *Collector) FetchTableMetadata(ctx context.Context, catalog, schema, table string) (*collector.TableMetadata, error) {
	if c.db == nil {
		return nil, collector.NewConnectionClosedErrorWithCategory(collector.CategoryRDBMS, SourceName, "fetch_table_metadata")
	}

	metadata := &collector.TableMetadata{
		Catalog:        catalog,
		Schema:         schema,
		Name:           table,
		SourceCategory: c.Category(),
		SourceType:     c.Type(),
		Type:           collector.TableTypeTable,
	}

	// Fetch columns
	columns, err := c.fetchColumns(ctx, schema, table)
	if err != nil {
		return nil, err
	}
	metadata.Columns = columns

	// Fetch indexes
	indexes, err := c.fetchIndexes(ctx, schema, table)
	if err != nil {
		return nil, err
	}
	metadata.Indexes = indexes

	// Fetch partitions if any
	partitions, err := c.fetchPartitions(ctx, schema, table)
	if err != nil {
		return nil, err
	}
	metadata.Partitions = partitions

	return metadata, nil
}

// buildDSN constructs the Oracle connection string.
func (c *Collector) buildDSN() string {
	// Parse endpoint (expected format: host:port or host)
	endpoint := c.config.Endpoint
	if endpoint == "" {
		endpoint = "localhost:1521"
	}

	host := endpoint
	port := 1521

	if idx := strings.LastIndex(endpoint, ":"); idx != -1 {
		host = endpoint[:idx]
		if p, err := strconv.Atoi(endpoint[idx+1:]); err == nil {
			port = p
		}
	}

	// Get database from extra properties
	database := "XE"
	if c.config.Properties.Extra != nil {
		if db := c.config.Properties.Extra["database"]; db != "" {
			database = db
		}
	}

	// Oracle connection string format: user/password@host:port/service_name
	return fmt.Sprintf("%s/%s@%s:%d/%s",
		c.config.Credentials.User,
		c.config.Credentials.Password,
		host,
		port,
		database,
	)
}

// fetchColumns retrieves column information for a table.
func (c *Collector) fetchColumns(ctx context.Context, schema, table string) ([]collector.Column, error) {
	rows, err := c.db.QueryContext(ctx, GetAllTabColumnsQuery(), strings.ToUpper(schema), strings.ToUpper(table))
	if err != nil {
		return nil, collector.NewQueryErrorWithCategory(collector.CategoryRDBMS, SourceName, "fetch_columns", err)
	}
	defer rows.Close()

	var columns []collector.Column
	for rows.Next() {
		var col collector.Column
		var nullable string
		var dataDefault sql.NullString
		var comments string

		err := rows.Scan(
			&col.Name,
			&col.Type,
			&col.OrdinalPosition,
			&col.Length,
			&col.Precision,
			&col.Scale,
			&nullable,
			&dataDefault,
			&comments,
		)
		if err != nil {
			return nil, collector.NewQueryErrorWithCategory(collector.CategoryRDBMS, SourceName, "fetch_columns", err)
		}

		col.Nullable = (nullable == "Y")
		if dataDefault.Valid && dataDefault.String != "" {
			col.Default = &dataDefault.String
		}
		col.Comment = comments

		columns = append(columns, col)
	}

	if err := rows.Err(); err != nil {
		return nil, collector.NewQueryErrorWithCategory(collector.CategoryRDBMS, SourceName, "fetch_columns", err)
	}

	return columns, nil
}

// fetchIndexes retrieves index information for a table.
func (c *Collector) fetchIndexes(ctx context.Context, schema, table string) ([]collector.Index, error) {
	rows, err := c.db.QueryContext(ctx, GetAllIndexesQuery(), strings.ToUpper(schema), strings.ToUpper(table))
	if err != nil {
		return nil, collector.NewQueryErrorWithCategory(collector.CategoryRDBMS, SourceName, "fetch_indexes", err)
	}
	defer rows.Close()

	indexMap := make(map[string]*collector.Index)
	for rows.Next() {
		var indexName, columnName, uniqueness string
		var columnPosition int

		err := rows.Scan(&indexName, &columnName, &columnPosition, &uniqueness)
		if err != nil {
			return nil, collector.NewQueryErrorWithCategory(collector.CategoryRDBMS, SourceName, "fetch_indexes", err)
		}

		if index, exists := indexMap[indexName]; exists {
			index.Columns = append(index.Columns, columnName)
		} else {
			indexMap[indexName] = &collector.Index{
				Name:    indexName,
				Columns: []string{columnName},
				Unique:  uniqueness == "UNIQUE",
			}
		}
	}

	if err := rows.Err(); err != nil {
		return nil, collector.NewQueryErrorWithCategory(collector.CategoryRDBMS, SourceName, "fetch_indexes", err)
	}

	var indexes []collector.Index
	for _, index := range indexMap {
		indexes = append(indexes, *index)
	}

	return indexes, nil
}

// fetchPartitions retrieves partition information for a table.
func (c *Collector) fetchPartitions(ctx context.Context, schema, table string) ([]collector.PartitionInfo, error) {
	rows, err := c.db.QueryContext(ctx, GetAllTabPartitionsQuery(), strings.ToUpper(schema), strings.ToUpper(table))
	if err != nil {
		return nil, collector.NewQueryErrorWithCategory(collector.CategoryRDBMS, SourceName, "fetch_partitions", err)
	}
	defer rows.Close()

	var partitions []collector.PartitionInfo
	for rows.Next() {
		var partition collector.PartitionInfo
		var highValue string
		var numRows sql.NullInt64

		err := rows.Scan(&partition.Name, &highValue, &numRows)
		if err != nil {
			return nil, collector.NewQueryErrorWithCategory(collector.CategoryRDBMS, SourceName, "fetch_partitions", err)
		}

		partition.Expression = highValue
		if numRows.Valid {
			partition.ValuesCount = int(numRows.Int64)
		}

		partitions = append(partitions, partition)
	}

	if err := rows.Err(); err != nil {
		return nil, collector.NewQueryErrorWithCategory(collector.CategoryRDBMS, SourceName, "fetch_partitions", err)
	}

	return partitions, nil
}