// Package sqlserver provides SQL Server metadata queries.
package sqlserver

// GetDatabasesQuery returns the query to get all databases in SQL Server.
func GetDatabasesQuery() string {
	return `
		SELECT 
			database_id,
			name,
			create_date,
			collation_name
		FROM sys.databases
		WHERE name NOT IN ('master', 'tempdb', 'model', 'msdb')
		ORDER BY name`
}

// GetSchemasQuery returns the query to get all schemas in a database.
func GetSchemasQuery() string {
	return `
		SELECT s.name
		FROM [?].sys.schemas s
		WHERE s.name NOT IN ('sys', 'information_schema', 'guest', 'INFORMATION_SCHEMA')
		ORDER BY s.name`
}

// GetTablesQuery returns the query to get all tables and views in a schema.
func GetTablesQuery() string {
	return `
		SELECT t.name
		FROM [?].sys.tables t
		INNER JOIN [?].sys.schemas s ON t.schema_id = s.schema_id
		WHERE s.name = ?
		UNION ALL
		SELECT v.name
		FROM [?].sys.views v
		INNER JOIN [?].sys.schemas s ON v.schema_id = s.schema_id
		WHERE s.name = ?
		ORDER BY name`
}

// GetTableInfoQuery returns the query to get table type and description.
func GetTableInfoQuery() string {
	return `
		SELECT 
			CASE 
				WHEN t.name IS NOT NULL THEN 'TABLE'
				WHEN v.name IS NOT NULL THEN 'VIEW'
				ELSE 'TABLE'
			END as table_type,
			ISNULL(ep.value, '') as description
		FROM [?].sys.schemas s
		LEFT JOIN [?].sys.tables t ON s.schema_id = t.schema_id AND t.name = ?
		LEFT JOIN [?].sys.views v ON s.schema_id = v.schema_id AND v.name = ?
		LEFT JOIN [?].sys.extended_properties ep ON ep.major_id = COALESCE(t.object_id, v.object_id)
			AND ep.minor_id = 0 AND ep.name = 'MS_Description'
		WHERE s.name = ?`
}

// GetColumnsQuery returns the query to get all columns for a table.
func GetColumnsQuery() string {
	return `
		SELECT 
			c.column_id,
			c.name,
			t.name as data_type,
			c.max_length,
			c.precision,
			c.scale,
			CASE WHEN c.is_nullable = 1 THEN 'YES' ELSE 'NO' END as is_nullable,
			ISNULL(dc.definition, '') as column_default,
			ISNULL(ep.value, '') as description
		FROM [?].sys.columns c
		INNER JOIN [?].sys.objects o ON c.object_id = o.object_id
		INNER JOIN [?].sys.schemas s ON o.schema_id = s.schema_id
		INNER JOIN [?].sys.types t ON c.user_type_id = t.user_type_id
		LEFT JOIN [?].sys.default_constraints dc ON c.default_object_id = dc.object_id
		LEFT JOIN [?].sys.extended_properties ep ON ep.major_id = c.object_id 
			AND ep.minor_id = c.column_id AND ep.name = 'MS_Description'
		WHERE s.name = ? AND o.name = ?
		ORDER BY c.column_id`
}

// GetIndexesQuery returns the query to get all indexes for a table.
func GetIndexesQuery() string {
	return `
		SELECT 
			i.name,
			c.name as column_name,
			ic.key_ordinal,
			i.is_unique
		FROM [?].sys.indexes i
		INNER JOIN [?].sys.index_columns ic ON i.object_id = ic.object_id AND i.index_id = ic.index_id
		INNER JOIN [?].sys.columns c ON ic.object_id = c.object_id AND ic.column_id = c.column_id
		INNER JOIN [?].sys.objects o ON i.object_id = o.object_id
		INNER JOIN [?].sys.schemas s ON o.schema_id = s.schema_id
		WHERE s.name = ? AND o.name = ? AND i.name IS NOT NULL
		ORDER BY i.name, ic.key_ordinal`
}

// GetPrimaryKeyQuery returns the query to get primary key columns for a table.
func GetPrimaryKeyQuery() string {
	return `
		SELECT c.name
		FROM [?].sys.key_constraints kc
		INNER JOIN [?].sys.index_columns ic ON kc.parent_object_id = ic.object_id 
			AND kc.unique_index_id = ic.index_id
		INNER JOIN [?].sys.columns c ON ic.object_id = c.object_id AND ic.column_id = c.column_id
		INNER JOIN [?].sys.objects o ON kc.parent_object_id = o.object_id
		INNER JOIN [?].sys.schemas s ON o.schema_id = s.schema_id
		WHERE s.name = ? AND o.name = ? AND kc.type = 'PK'
		ORDER BY ic.key_ordinal`
}

// GetTableStatsQuery returns the query to get table statistics.
func GetTableStatsQuery() string {
	return `
		SELECT 
			SUM(p.rows) as row_count,
			SUM(a.used_pages) * 8 as data_size_kb
		FROM [?].sys.tables t
		INNER JOIN [?].sys.schemas s ON t.schema_id = s.schema_id
		INNER JOIN [?].sys.partitions p ON t.object_id = p.object_id
		INNER JOIN [?].sys.allocation_units a ON p.partition_id = a.container_id
		WHERE s.name = ? AND t.name = ? AND p.index_id IN (0, 1)
		GROUP BY t.object_id`
}

// GetPartitionsQuery returns the query to get partition information for a table.
func GetPartitionsQuery() string {
	return `
		SELECT 
			p.partition_number,
			pf.name as partition_function,
			ISNULL(prv.value, '') as boundary_value,
			p.rows
		FROM [?].sys.partitions p
		INNER JOIN [?].sys.objects o ON p.object_id = o.object_id
		INNER JOIN [?].sys.schemas s ON o.schema_id = s.schema_id
		LEFT JOIN [?].sys.partition_schemes ps ON p.partition_id = ps.data_space_id
		LEFT JOIN [?].sys.partition_functions pf ON ps.function_id = pf.function_id
		LEFT JOIN [?].sys.partition_range_values prv ON pf.function_id = prv.function_id 
			AND p.partition_number = prv.boundary_id
		WHERE s.name = ? AND o.name = ? AND p.partition_number > 1
		ORDER BY p.partition_number`
}

// GetForeignKeysQuery returns the query to get foreign key constraints for a table.
func GetForeignKeysQuery() string {
	return `
		SELECT 
			fk.name as constraint_name,
			c1.name as column_name,
			s2.name as referenced_schema,
			o2.name as referenced_table,
			c2.name as referenced_column
		FROM [?].sys.foreign_keys fk
		INNER JOIN [?].sys.foreign_key_columns fkc ON fk.object_id = fkc.constraint_object_id
		INNER JOIN [?].sys.columns c1 ON fkc.parent_object_id = c1.object_id 
			AND fkc.parent_column_id = c1.column_id
		INNER JOIN [?].sys.columns c2 ON fkc.referenced_object_id = c2.object_id 
			AND fkc.referenced_column_id = c2.column_id
		INNER JOIN [?].sys.objects o1 ON fk.parent_object_id = o1.object_id
		INNER JOIN [?].sys.objects o2 ON fk.referenced_object_id = o2.object_id
		INNER JOIN [?].sys.schemas s1 ON o1.schema_id = s1.schema_id
		INNER JOIN [?].sys.schemas s2 ON o2.schema_id = s2.schema_id
		WHERE s1.name = ? AND o1.name = ?
		ORDER BY fk.name, fkc.constraint_column_id`
}

// GetCheckConstraintsQuery returns the query to get check constraints for a table.
func GetCheckConstraintsQuery() string {
	return `
		SELECT 
			cc.name as constraint_name,
			cc.definition,
			cc.is_disabled
		FROM [?].sys.check_constraints cc
		INNER JOIN [?].sys.objects o ON cc.parent_object_id = o.object_id
		INNER JOIN [?].sys.schemas s ON o.schema_id = s.schema_id
		WHERE s.name = ? AND o.name = ?
		ORDER BY cc.name`
}

// GetTriggersQuery returns the query to get triggers for a table.
func GetTriggersQuery() string {
	return `
		SELECT 
			tr.name,
			tr.type_desc,
			CASE 
				WHEN tr.is_instead_of_trigger = 1 THEN 'INSTEAD OF'
				ELSE 'AFTER'
			END as trigger_type,
			tr.is_disabled
		FROM [?].sys.triggers tr
		INNER JOIN [?].sys.objects o ON tr.parent_id = o.object_id
		INNER JOIN [?].sys.schemas s ON o.schema_id = s.schema_id
		WHERE s.name = ? AND o.name = ?
		ORDER BY tr.name`
}

// GetViewDefinitionQuery returns the query to get view definition.
func GetViewDefinitionQuery() string {
	return `
		SELECT m.definition
		FROM [?].sys.views v
		INNER JOIN [?].sys.schemas s ON v.schema_id = s.schema_id
		INNER JOIN [?].sys.sql_modules m ON v.object_id = m.object_id
		WHERE s.name = ? AND v.name = ?`
}

// GetStoredProceduresQuery returns the query to get stored procedures in a schema.
func GetStoredProceduresQuery() string {
	return `
		SELECT 
			p.name,
			p.create_date,
			p.modify_date,
			ISNULL(ep.value, '') as description
		FROM [?].sys.procedures p
		INNER JOIN [?].sys.schemas s ON p.schema_id = s.schema_id
		LEFT JOIN [?].sys.extended_properties ep ON ep.major_id = p.object_id 
			AND ep.minor_id = 0 AND ep.name = 'MS_Description'
		WHERE s.name = ?
		ORDER BY p.name`
}

// GetFunctionsQuery returns the query to get functions in a schema.
func GetFunctionsQuery() string {
	return `
		SELECT 
			o.name,
			o.create_date,
			o.modify_date,
			o.type_desc,
			ISNULL(ep.value, '') as description
		FROM [?].sys.objects o
		INNER JOIN [?].sys.schemas s ON o.schema_id = s.schema_id
		LEFT JOIN [?].sys.extended_properties ep ON ep.major_id = o.object_id 
			AND ep.minor_id = 0 AND ep.name = 'MS_Description'
		WHERE s.name = ? AND o.type IN ('FN', 'IF', 'TF')
		ORDER BY o.name`
}

// GetUserDefinedTypesQuery returns the query to get user-defined types in a schema.
func GetUserDefinedTypesQuery() string {
	return `
		SELECT 
			t.name,
			st.name as base_type,
			t.max_length,
			t.precision,
			t.scale,
			t.is_nullable
		FROM [?].sys.types t
		INNER JOIN [?].sys.schemas s ON t.schema_id = s.schema_id
		LEFT JOIN [?].sys.types st ON t.system_type_id = st.system_type_id AND st.user_type_id = st.system_type_id
		WHERE s.name = ? AND t.is_user_defined = 1
		ORDER BY t.name`
}

// GetServerInfoQuery returns the query to get SQL Server instance information.
func GetServerInfoQuery() string {
	return `
		SELECT 
			@@SERVERNAME as server_name,
			@@VERSION as version,
			SERVERPROPERTY('ProductVersion') as product_version,
			SERVERPROPERTY('ProductLevel') as product_level,
			SERVERPROPERTY('Edition') as edition,
			SERVERPROPERTY('EngineEdition') as engine_edition,
			SERVERPROPERTY('MachineName') as machine_name,
			SERVERPROPERTY('InstanceName') as instance_name`
}