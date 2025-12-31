// Package doris provides Doris metadata queries.
package doris

// GetDatabasesQuery returns the query to get all databases in Doris.
func GetDatabasesQuery() string {
	return `
		SHOW DATABASES`
}

// GetTablesQuery returns the query to get all tables in a database.
func GetTablesQuery() string {
	return `
		SHOW TABLES FROM ?`
}

// GetColumnsQuery returns the query to get all columns for a table.
func GetColumnsQuery() string {
	return `
		SELECT 
			ORDINAL_POSITION,
			COLUMN_NAME,
			DATA_TYPE,
			IS_NULLABLE,
			COLUMN_DEFAULT,
			COLUMN_COMMENT
		FROM INFORMATION_SCHEMA.COLUMNS 
		WHERE TABLE_SCHEMA = ? AND TABLE_NAME = ?
		ORDER BY ORDINAL_POSITION`
}

// GetTableStatsQuery returns the query to get table statistics.
func GetTableStatsQuery() string {
	return `
		SELECT 
			TABLE_ROWS,
			DATA_LENGTH
		FROM INFORMATION_SCHEMA.TABLES 
		WHERE TABLE_SCHEMA = ? AND TABLE_NAME = ?`
}

// GetPartitionsQuery returns the query to get partition information for a table.
func GetPartitionsQuery() string {
	return `
		SHOW PARTITIONS FROM ?.?`
}

// GetTableInfoQuery returns the query to get detailed table information.
func GetTableInfoQuery() string {
	return `
		SHOW CREATE TABLE ?.?`
}

// GetIndexesQuery returns the query to get indexes for a table.
func GetIndexesQuery() string {
	return `
		SHOW INDEX FROM ?.?`
}

// GetTableSchemaQuery returns the query to get table schema information.
func GetTableSchemaQuery() string {
	return `
		DESC ?.?`
}

// GetMaterializedViewsQuery returns the query to get materialized views.
func GetMaterializedViewsQuery() string {
	return `
		SHOW MATERIALIZED VIEW FROM ?`
}

// GetViewsQuery returns the query to get views in a database.
func GetViewsQuery() string {
	return `
		SELECT TABLE_NAME
		FROM INFORMATION_SCHEMA.VIEWS 
		WHERE TABLE_SCHEMA = ?
		ORDER BY TABLE_NAME`
}

// GetViewDefinitionQuery returns the query to get view definition.
func GetViewDefinitionQuery() string {
	return `
		SELECT VIEW_DEFINITION
		FROM INFORMATION_SCHEMA.VIEWS 
		WHERE TABLE_SCHEMA = ? AND TABLE_NAME = ?`
}

// GetTableEngineQuery returns the query to get table engine information.
func GetTableEngineQuery() string {
	return `
		SELECT ENGINE
		FROM INFORMATION_SCHEMA.TABLES 
		WHERE TABLE_SCHEMA = ? AND TABLE_NAME = ?`
}

// GetTablePropertiesQuery returns the query to get table properties.
func GetTablePropertiesQuery() string {
	return `
		SHOW TABLE ?.? STATUS`
}

// GetBackendsQuery returns the query to get backend nodes information.
func GetBackendsQuery() string {
	return `
		SHOW BACKENDS`
}

// GetFrontendsQuery returns the query to get frontend nodes information.
func GetFrontendsQuery() string {
	return `
		SHOW FRONTENDS`
}

// GetBrokersQuery returns the query to get broker information.
func GetBrokersQuery() string {
	return `
		SHOW BROKER`
}

// GetLoadJobsQuery returns the query to get load jobs.
func GetLoadJobsQuery() string {
	return `
		SHOW LOAD FROM ?`
}

// GetStreamLoadJobsQuery returns the query to get stream load jobs.
func GetStreamLoadJobsQuery() string {
	return `
		SHOW STREAM LOAD FROM ?`
}

// GetRoutineLoadJobsQuery returns the query to get routine load jobs.
func GetRoutineLoadJobsQuery() string {
	return `
		SHOW ROUTINE LOAD FROM ?`
}

// GetBackupJobsQuery returns the query to get backup jobs.
func GetBackupJobsQuery() string {
	return `
		SHOW BACKUP FROM ?`
}

// GetRestoreJobsQuery returns the query to get restore jobs.
func GetRestoreJobsQuery() string {
	return `
		SHOW RESTORE FROM ?`
}

// GetSnapshotsQuery returns the query to get snapshots.
func GetSnapshotsQuery() string {
	return `
		SHOW SNAPSHOT ON ?`
}

// GetRepositoriesQuery returns the query to get repositories.
func GetRepositoriesQuery() string {
	return `
		SHOW REPOSITORIES`
}

// GetUserQuery returns the query to get user information.
func GetUserQuery() string {
	return `
		SHOW ALL GRANTS`
}

// GetRolesQuery returns the query to get roles.
func GetRolesQuery() string {
	return `
		SHOW ROLES`
}

// GetResourcesQuery returns the query to get resources.
func GetResourcesQuery() string {
	return `
		SHOW RESOURCES`
}

// GetWorkloadGroupsQuery returns the query to get workload groups.
func GetWorkloadGroupsQuery() string {
	return `
		SHOW WORKLOAD GROUPS`
}

// GetVariablesQuery returns the query to get system variables.
func GetVariablesQuery() string {
	return `
		SHOW VARIABLES`
}

// GetStatusQuery returns the query to get system status.
func GetStatusQuery() string {
	return `
		SHOW STATUS`
}

// GetProcessListQuery returns the query to get process list.
func GetProcessListQuery() string {
	return `
		SHOW PROCESSLIST`
}

// GetTabletsQuery returns the query to get tablet information.
func GetTabletsQuery() string {
	return `
		SHOW TABLETS FROM ?.?`
}

// GetPartitionStatsQuery returns the query to get partition statistics.
func GetPartitionStatsQuery() string {
	return `
		SHOW PARTITIONS FROM ?.? WHERE PartitionName = ?`
}

// GetReplicaStatusQuery returns the query to get replica status.
func GetReplicaStatusQuery() string {
	return `
		SHOW REPLICA STATUS FROM ?.?`
}

// GetDataSkewQuery returns the query to get data skew information.
func GetDataSkewQuery() string {
	return `
		SHOW DATA SKEW FROM ?.?`
}

// GetQueryStatsQuery returns the query to get query statistics.
func GetQueryStatsQuery() string {
	return `
		SHOW QUERY STATS`
}

// GetCatalogQuery returns the query to get catalog information.
func GetCatalogQuery() string {
	return `
		SHOW CATALOGS`
}

// GetExternalTableQuery returns the query to get external table information.
func GetExternalTableQuery() string {
	return `
		SHOW TABLES FROM ? FROM ?`
}

// GetSyncJobsQuery returns the query to get sync jobs.
func GetSyncJobsQuery() string {
	return `
		SHOW SYNC JOB FROM ?`
}

// GetPolicyQuery returns the query to get policies.
func GetPolicyQuery() string {
	return `
		SHOW POLICY`
}

// GetEncryptKeysQuery returns the query to get encrypt keys.
func GetEncryptKeysQuery() string {
	return `
		SHOW ENCRYPTKEYS`
}

// GetFunctionsQuery returns the query to get functions.
func GetFunctionsQuery() string {
	return `
		SHOW FUNCTIONS FROM ?`
}

// GetTriggersQuery returns the query to get triggers.
func GetTriggersQuery() string {
	return `
		SHOW TRIGGERS FROM ?`
}

// GetEventsQuery returns the query to get events.
func GetEventsQuery() string {
	return `
		SHOW EVENTS FROM ?`
}