// Package mysql provides a MySQL metadata collector implementation.
package mysql

// SQL queries for MySQL metadata collection

// queryListDatabases retrieves all database names from information_schema.SCHEMATA
const queryListDatabases = `
SELECT SCHEMA_NAME 
FROM information_schema.SCHEMATA 
WHERE SCHEMA_NAME NOT IN ('information_schema', 'performance_schema', 'mysql', 'sys')
ORDER BY SCHEMA_NAME
`

// queryListTables retrieves all table names from a specific database
const queryListTables = `
SELECT TABLE_NAME 
FROM information_schema.TABLES 
WHERE TABLE_SCHEMA = ?
ORDER BY TABLE_NAME
`

// queryGetTableInfo retrieves basic table information
const queryGetTableInfo = `
SELECT TABLE_TYPE, TABLE_COMMENT
FROM information_schema.TABLES
WHERE TABLE_SCHEMA = ? AND TABLE_NAME = ?
`

// queryGetColumns retrieves column information for a specific table
const queryGetColumns = `
SELECT 
    ORDINAL_POSITION,
    COLUMN_NAME,
    DATA_TYPE,
    COLUMN_TYPE,
    CHARACTER_MAXIMUM_LENGTH,
    NUMERIC_PRECISION,
    NUMERIC_SCALE,
    IS_NULLABLE,
    COLUMN_DEFAULT,
    COLUMN_KEY,
    EXTRA,
    COLUMN_COMMENT
FROM information_schema.COLUMNS
WHERE TABLE_SCHEMA = ? AND TABLE_NAME = ?
ORDER BY ORDINAL_POSITION
`

// queryGetIndexes retrieves index information from information_schema.STATISTICS
const queryGetIndexes = `
SELECT 
    INDEX_NAME,
    COLUMN_NAME,
    NON_UNIQUE,
    INDEX_TYPE,
    INDEX_COMMENT
FROM information_schema.STATISTICS
WHERE TABLE_SCHEMA = ? AND TABLE_NAME = ?
ORDER BY INDEX_NAME, SEQ_IN_INDEX
`

// queryGetPrimaryKeys retrieves primary key columns from information_schema.KEY_COLUMN_USAGE
const queryGetPrimaryKeys = `
SELECT COLUMN_NAME
FROM information_schema.KEY_COLUMN_USAGE
WHERE TABLE_SCHEMA = ? 
    AND TABLE_NAME = ? 
    AND CONSTRAINT_NAME = 'PRIMARY'
ORDER BY ORDINAL_POSITION
`

// queryGetTableStats retrieves table statistics
const queryGetTableStats = `
SELECT 
    TABLE_ROWS,
    DATA_LENGTH,
    INDEX_LENGTH
FROM information_schema.TABLES
WHERE TABLE_SCHEMA = ? AND TABLE_NAME = ?
`

// queryGetPartitions retrieves partition information
const queryGetPartitions = `
SELECT 
    PARTITION_NAME,
    PARTITION_METHOD,
    PARTITION_EXPRESSION,
    TABLE_ROWS
FROM information_schema.PARTITIONS
WHERE TABLE_SCHEMA = ? AND TABLE_NAME = ?
ORDER BY PARTITION_ORDINAL_POSITION
`
