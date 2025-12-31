package metadata

import (
	"database/sql"
	"fmt"
)

// DatabaseMetadataFetcher fetches metadata from a database connection.
type DatabaseMetadataFetcher interface {
	// FetchTableSchema fetches the schema for a specific table.
	FetchTableSchema(database, table string) (*TableSchema, error)

	// FetchAllTables fetches all table schemas from the database.
	FetchAllTables(database string) ([]*TableSchema, error)

	// Close closes the database connection.
	Close() error
}

// GenericDatabaseFetcher is a generic implementation using database/sql.
type GenericDatabaseFetcher struct {
	db       *sql.DB
	dialect  string // "mysql", "postgres", "sqlite", etc.
	provider *MemoryProvider
}

// NewGenericDatabaseFetcher creates a new database metadata fetcher.
func NewGenericDatabaseFetcher(db *sql.DB, dialect string) *GenericDatabaseFetcher {
	return &GenericDatabaseFetcher{
		db:       db,
		dialect:  dialect,
		provider: NewMemoryProvider(),
	}
}

// FetchTableSchema fetches the schema for a specific table.
func (f *GenericDatabaseFetcher) FetchTableSchema(database, table string) (*TableSchema, error) {
	switch f.dialect {
	case "mysql":
		return f.fetchMySQLTableSchema(database, table)
	case "postgres", "postgresql":
		return f.fetchPostgresTableSchema(database, table)
	default:
		return nil, fmt.Errorf("unsupported dialect: %s", f.dialect)
	}
}

// FetchAllTables fetches all table schemas from the database.
func (f *GenericDatabaseFetcher) FetchAllTables(database string) ([]*TableSchema, error) {
	switch f.dialect {
	case "mysql":
		return f.fetchMySQLAllTables(database)
	case "postgres", "postgresql":
		return f.fetchPostgresAllTables(database)
	default:
		return nil, fmt.Errorf("unsupported dialect: %s", f.dialect)
	}
}

// Close closes the database connection.
func (f *GenericDatabaseFetcher) Close() error {
	return f.db.Close()
}

// fetchMySQLTableSchema fetches table schema from MySQL.
func (f *GenericDatabaseFetcher) fetchMySQLTableSchema(database, table string) (*TableSchema, error) {
	query := `
		SELECT 
			COLUMN_NAME,
			DATA_TYPE,
			IS_NULLABLE,
			COLUMN_KEY,
			COLUMN_COMMENT,
			COLUMN_DEFAULT
		FROM INFORMATION_SCHEMA.COLUMNS
		WHERE TABLE_SCHEMA = ? AND TABLE_NAME = ?
		ORDER BY ORDINAL_POSITION
	`

	rows, err := f.db.Query(query, database, table)
	if err != nil {
		return nil, fmt.Errorf("failed to query columns: %w", err)
	}
	defer rows.Close()

	schema := &TableSchema{
		Database:  database,
		Table:     table,
		Columns:   make([]ColumnSchema, 0),
		TableType: "TABLE",
	}

	for rows.Next() {
		var col ColumnSchema
		var nullable, columnKey string
		var comment, defaultExpr sql.NullString

		if err := rows.Scan(&col.Name, &col.DataType, &nullable, &columnKey, &comment, &defaultExpr); err != nil {
			return nil, fmt.Errorf("failed to scan column: %w", err)
		}

		col.Nullable = nullable == "YES"
		col.PrimaryKey = columnKey == "PRI"
		if comment.Valid {
			col.Comment = comment.String
		}
		if defaultExpr.Valid {
			col.DefaultExpr = defaultExpr.String
		}

		schema.Columns = append(schema.Columns, col)
	}

	if len(schema.Columns) == 0 {
		return nil, fmt.Errorf("%w: %s.%s", ErrTableNotFound, database, table)
	}

	return schema, nil
}

// fetchMySQLAllTables fetches all table schemas from MySQL.
func (f *GenericDatabaseFetcher) fetchMySQLAllTables(database string) ([]*TableSchema, error) {
	query := `
		SELECT TABLE_NAME, TABLE_TYPE, TABLE_COMMENT
		FROM INFORMATION_SCHEMA.TABLES
		WHERE TABLE_SCHEMA = ?
	`

	rows, err := f.db.Query(query, database)
	if err != nil {
		return nil, fmt.Errorf("failed to query tables: %w", err)
	}
	defer rows.Close()

	var schemas []*TableSchema
	for rows.Next() {
		var tableName, tableType string
		var comment sql.NullString

		if err := rows.Scan(&tableName, &tableType, &comment); err != nil {
			return nil, fmt.Errorf("failed to scan table: %w", err)
		}

		schema, err := f.fetchMySQLTableSchema(database, tableName)
		if err != nil {
			return nil, err
		}

		if tableType == "VIEW" {
			schema.TableType = "VIEW"
		}
		if comment.Valid {
			schema.Comment = comment.String
		}

		schemas = append(schemas, schema)
	}

	return schemas, nil
}

// fetchPostgresTableSchema fetches table schema from PostgreSQL.
func (f *GenericDatabaseFetcher) fetchPostgresTableSchema(database, table string) (*TableSchema, error) {
	query := `
		SELECT 
			c.column_name,
			c.data_type,
			c.is_nullable,
			COALESCE(tc.constraint_type, '') as constraint_type,
			COALESCE(pgd.description, '') as column_comment,
			c.column_default
		FROM information_schema.columns c
		LEFT JOIN information_schema.key_column_usage kcu 
			ON c.table_name = kcu.table_name 
			AND c.column_name = kcu.column_name
			AND c.table_schema = kcu.table_schema
		LEFT JOIN information_schema.table_constraints tc 
			ON kcu.constraint_name = tc.constraint_name
			AND tc.constraint_type = 'PRIMARY KEY'
		LEFT JOIN pg_catalog.pg_statio_all_tables st 
			ON c.table_schema = st.schemaname 
			AND c.table_name = st.relname
		LEFT JOIN pg_catalog.pg_description pgd 
			ON pgd.objoid = st.relid 
			AND pgd.objsubid = c.ordinal_position
		WHERE c.table_schema = $1 AND c.table_name = $2
		ORDER BY c.ordinal_position
	`

	// For PostgreSQL, database is typically the schema name
	schemaName := database
	if schemaName == "" {
		schemaName = "public"
	}

	rows, err := f.db.Query(query, schemaName, table)
	if err != nil {
		return nil, fmt.Errorf("failed to query columns: %w", err)
	}
	defer rows.Close()

	schema := &TableSchema{
		Schema:    schemaName,
		Table:     table,
		Columns:   make([]ColumnSchema, 0),
		TableType: "TABLE",
	}

	for rows.Next() {
		var col ColumnSchema
		var nullable, constraintType string
		var comment sql.NullString
		var defaultExpr sql.NullString

		if err := rows.Scan(&col.Name, &col.DataType, &nullable, &constraintType, &comment, &defaultExpr); err != nil {
			return nil, fmt.Errorf("failed to scan column: %w", err)
		}

		col.Nullable = nullable == "YES"
		col.PrimaryKey = constraintType == "PRIMARY KEY"
		if comment.Valid {
			col.Comment = comment.String
		}
		if defaultExpr.Valid {
			col.DefaultExpr = defaultExpr.String
		}

		schema.Columns = append(schema.Columns, col)
	}

	if len(schema.Columns) == 0 {
		return nil, fmt.Errorf("%w: %s.%s", ErrTableNotFound, schemaName, table)
	}

	return schema, nil
}

// fetchPostgresAllTables fetches all table schemas from PostgreSQL.
func (f *GenericDatabaseFetcher) fetchPostgresAllTables(database string) ([]*TableSchema, error) {
	schemaName := database
	if schemaName == "" {
		schemaName = "public"
	}

	query := `
		SELECT table_name, table_type
		FROM information_schema.tables
		WHERE table_schema = $1
	`

	rows, err := f.db.Query(query, schemaName)
	if err != nil {
		return nil, fmt.Errorf("failed to query tables: %w", err)
	}
	defer rows.Close()

	var schemas []*TableSchema
	for rows.Next() {
		var tableName, tableType string

		if err := rows.Scan(&tableName, &tableType); err != nil {
			return nil, fmt.Errorf("failed to scan table: %w", err)
		}

		schema, err := f.fetchPostgresTableSchema(schemaName, tableName)
		if err != nil {
			return nil, err
		}

		if tableType == "VIEW" {
			schema.TableType = "VIEW"
		}

		schemas = append(schemas, schema)
	}

	return schemas, nil
}

// LoadToProvider loads fetched schemas into a MemoryProvider.
func (f *GenericDatabaseFetcher) LoadToProvider(database string) (*MemoryProvider, error) {
	schemas, err := f.FetchAllTables(database)
	if err != nil {
		return nil, err
	}

	provider := NewMemoryProvider()
	for _, schema := range schemas {
		if err := provider.AddTableSchema(schema); err != nil {
			return nil, err
		}
	}

	return provider, nil
}
