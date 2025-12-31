// Package postgres provides a PostgreSQL metadata collector implementation.
package postgres

// SQL queries for PostgreSQL metadata collection

// queryListDatabases retrieves all database names from pg_database
const queryListDatabases = `
SELECT datname 
FROM pg_database 
WHERE datistemplate = false 
  AND datname NOT IN ('postgres')
ORDER BY datname
`

// queryListSchemas retrieves all schema names from information_schema.schemata
const queryListSchemas = `
SELECT schema_name 
FROM information_schema.schemata 
WHERE schema_name NOT IN ('pg_catalog', 'information_schema', 'pg_toast')
  AND schema_name NOT LIKE 'pg_temp_%'
  AND schema_name NOT LIKE 'pg_toast_temp_%'
ORDER BY schema_name
`

// queryListTables retrieves all table names from a specific schema
const queryListTables = `
SELECT table_name 
FROM information_schema.tables 
WHERE table_schema = $1
  AND table_type IN ('BASE TABLE', 'VIEW')
ORDER BY table_name
`

// queryGetTableInfo retrieves basic table information
const queryGetTableInfo = `
SELECT 
    t.table_type,
    obj_description((quote_ident(t.table_schema) || '.' || quote_ident(t.table_name))::regclass, 'pg_class') as table_comment
FROM information_schema.tables t
WHERE t.table_schema = $1 AND t.table_name = $2
`

// queryGetColumns retrieves column information for a specific table
const queryGetColumns = `
SELECT 
    c.ordinal_position,
    c.column_name,
    c.data_type,
    c.udt_name,
    c.character_maximum_length,
    c.numeric_precision,
    c.numeric_scale,
    c.is_nullable,
    c.column_default,
    col_description((quote_ident(c.table_schema) || '.' || quote_ident(c.table_name))::regclass, c.ordinal_position) as column_comment,
    c.is_identity
FROM information_schema.columns c
WHERE c.table_schema = $1 AND c.table_name = $2
ORDER BY c.ordinal_position
`

// queryGetIndexes retrieves index information from pg_indexes
const queryGetIndexes = `
SELECT 
    i.indexname,
    a.attname as column_name,
    ix.indisunique,
    am.amname as index_type
FROM pg_indexes i
JOIN pg_class c ON c.relname = i.indexname
JOIN pg_index ix ON ix.indexrelid = c.oid
JOIN pg_attribute a ON a.attrelid = ix.indrelid AND a.attnum = ANY(ix.indkey)
JOIN pg_am am ON am.oid = c.relam
WHERE i.schemaname = $1 AND i.tablename = $2
ORDER BY i.indexname, a.attnum
`

// queryGetPrimaryKeys retrieves primary key columns
const queryGetPrimaryKeys = `
SELECT a.attname
FROM pg_index i
JOIN pg_attribute a ON a.attrelid = i.indrelid AND a.attnum = ANY(i.indkey)
JOIN pg_class c ON c.oid = i.indrelid
JOIN pg_namespace n ON n.oid = c.relnamespace
WHERE i.indisprimary
  AND n.nspname = $1
  AND c.relname = $2
ORDER BY array_position(i.indkey, a.attnum)
`

// queryGetTableStats retrieves table statistics from pg_class
const queryGetTableStats = `
SELECT 
    c.reltuples,
    c.relpages
FROM pg_class c
JOIN pg_namespace n ON n.oid = c.relnamespace
WHERE n.nspname = $1 AND c.relname = $2
`

// queryGetPartitions retrieves partition information for partitioned tables
const queryGetPartitions = `
SELECT 
    child.relname as partition_name,
    pg_get_partkeydef(parent.oid) as partition_method,
    pg_get_expr(child.relpartbound, child.oid) as partition_expression
FROM pg_inherits
JOIN pg_class parent ON pg_inherits.inhparent = parent.oid
JOIN pg_class child ON pg_inherits.inhrelid = child.oid
JOIN pg_namespace n ON n.oid = parent.relnamespace
WHERE n.nspname = $1 AND parent.relname = $2
ORDER BY child.relname
`
