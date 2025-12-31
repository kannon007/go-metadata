// Package clickhouse provides ClickHouse metadata queries.
package clickhouse

// GetDatabasesQuery returns the query to get all databases in ClickHouse.
func GetDatabasesQuery() string {
	return `
		SELECT name 
		FROM system.databases 
		WHERE name NOT IN ('system', 'information_schema', 'INFORMATION_SCHEMA')
		ORDER BY name`
}

// GetTablesQuery returns the query to get all tables in a database.
func GetTablesQuery() string {
	return `
		SELECT name 
		FROM system.tables 
		WHERE database = ?
		AND engine NOT IN ('View', 'MaterializedView', 'Dictionary')
		ORDER BY name`
}

// GetColumnsQuery returns the query to get all columns for a table.
func GetColumnsQuery() string {
	return `
		SELECT 
			name,
			type,
			default_kind,
			default_expression,
			comment
		FROM system.columns 
		WHERE database = ? AND table = ?
		ORDER BY position`
}

// GetTableStatsQuery returns the query to get table statistics.
func GetTableStatsQuery() string {
	return `
		SELECT 
			sum(rows) as total_rows,
			sum(bytes) as total_bytes
		FROM system.parts 
		WHERE database = ? AND table = ? AND active = 1`
}

// GetPartitionsQuery returns the query to get partition information for a table.
func GetPartitionsQuery() string {
	return `
		SELECT 
			partition_id,
			partition,
			sum(rows) as rows,
			sum(bytes) as bytes
		FROM system.parts 
		WHERE database = ? AND table = ? AND active = 1
		GROUP BY partition_id, partition
		ORDER BY partition_id`
}

// GetViewsQuery returns the query to get all views in a database.
func GetViewsQuery() string {
	return `
		SELECT name 
		FROM system.tables 
		WHERE database = ?
		AND engine IN ('View', 'MaterializedView')
		ORDER BY name`
}

// GetViewDefinitionQuery returns the query to get view definition.
func GetViewDefinitionQuery() string {
	return `
		SELECT as_select
		FROM system.tables 
		WHERE database = ? AND name = ?
		AND engine IN ('View', 'MaterializedView')`
}

// GetEnginesQuery returns the query to get available table engines.
func GetEnginesQuery() string {
	return `
		SELECT name, is_readonly
		FROM system.table_engines
		ORDER BY name`
}

// GetDictionariesQuery returns the query to get all dictionaries in a database.
func GetDictionariesQuery() string {
	return `
		SELECT name 
		FROM system.dictionaries 
		WHERE database = ?
		ORDER BY name`
}

// GetDictionaryInfoQuery returns the query to get dictionary information.
func GetDictionaryInfoQuery() string {
	return `
		SELECT 
			name,
			type,
			key,
			attribute.names,
			attribute.types,
			source,
			lifetime_min,
			lifetime_max,
			layout_type,
			is_injective,
			element_count,
			load_factor,
			creation_time,
			last_exception
		FROM system.dictionaries 
		WHERE database = ? AND name = ?`
}

// GetFunctionsQuery returns the query to get available functions.
func GetFunctionsQuery() string {
	return `
		SELECT name, is_aggregate
		FROM system.functions
		ORDER BY name`
}

// GetSettingsQuery returns the query to get current settings.
func GetSettingsQuery() string {
	return `
		SELECT name, value, changed, description
		FROM system.settings
		WHERE changed = 1
		ORDER BY name`
}

// GetProcessesQuery returns the query to get current processes.
func GetProcessesQuery() string {
	return `
		SELECT 
			query_id,
			user,
			address,
			elapsed,
			rows_read,
			bytes_read,
			memory_usage,
			query
		FROM system.processes
		ORDER BY elapsed DESC`
}

// GetMutationsQuery returns the query to get table mutations.
func GetMutationsQuery() string {
	return `
		SELECT 
			database,
			table,
			mutation_id,
			command,
			create_time,
			block_numbers.partition_id,
			block_numbers.number,
			parts_to_do,
			is_done,
			latest_failed_part,
			latest_fail_time,
			latest_fail_reason
		FROM system.mutations 
		WHERE database = ? AND table = ?
		ORDER BY create_time DESC`
}

// GetReplicationQueueQuery returns the query to get replication queue.
func GetReplicationQueueQuery() string {
	return `
		SELECT 
			database,
			table,
			replica_name,
			position,
			node_name,
			type,
			create_time,
			required_quorum,
			source_replica,
			new_part_name,
			parts_to_merge,
			is_currently_executing,
			num_tries,
			last_exception,
			last_attempt_time,
			num_postponed,
			postpone_reason,
			last_postpone_time
		FROM system.replication_queue 
		WHERE database = ? AND table = ?
		ORDER BY create_time`
}

// GetMergesQuery returns the query to get current merges.
func GetMergesQuery() string {
	return `
		SELECT 
			database,
			table,
			elapsed,
			progress,
			num_parts,
			result_part_name,
			is_mutation,
			total_size_bytes_compressed,
			total_size_marks,
			bytes_read_uncompressed,
			bytes_written_uncompressed,
			rows_read,
			rows_written,
			columns_written,
			memory_usage,
			thread_id
		FROM system.merges 
		WHERE database = ? AND table = ?
		ORDER BY elapsed DESC`
}

// GetDetachedPartsQuery returns the query to get detached parts.
func GetDetachedPartsQuery() string {
	return `
		SELECT 
			database,
			table,
			partition_id,
			name,
			disk,
			reason,
			min_block_number,
			max_block_number,
			level
		FROM system.detached_parts 
		WHERE database = ? AND table = ?
		ORDER BY partition_id, name`
}

// GetTableConstraintsQuery returns the query to get table constraints.
func GetTableConstraintsQuery() string {
	return `
		SELECT 
			database,
			table,
			name,
			type,
			expression
		FROM system.table_constraints 
		WHERE database = ? AND table = ?
		ORDER BY name`
}

// GetZooKeeperQuery returns the query to get ZooKeeper information.
func GetZooKeeperQuery() string {
	return `
		SELECT 
			name,
			value,
			czxid,
			mzxid,
			ctime,
			mtime,
			version,
			cversion,
			aversion,
			ephemeralOwner,
			dataLength,
			numChildren,
			pzxid,
			path
		FROM system.zookeeper 
		WHERE path = ?
		ORDER BY name`
}

// GetReplicasQuery returns the query to get replica information.
func GetReplicasQuery() string {
	return `
		SELECT 
			database,
			table,
			engine,
			is_leader,
			can_become_leader,
			is_readonly,
			is_session_expired,
			future_parts,
			parts_to_check,
			zookeeper_path,
			replica_name,
			replica_path,
			columns_version,
			queue_size,
			inserts_in_queue,
			merges_in_queue,
			part_mutations_in_queue,
			queue_oldest_time,
			inserts_oldest_time,
			merges_oldest_time,
			part_mutations_oldest_time,
			oldest_part_to_get,
			oldest_part_to_merge_to,
			oldest_part_to_mutate_to,
			log_max_index,
			log_pointer,
			last_queue_update,
			absolute_delay,
			total_replicas,
			active_replicas
		FROM system.replicas 
		WHERE database = ? AND table = ?`
}

// GetClusterQuery returns the query to get cluster information.
func GetClusterQuery() string {
	return `
		SELECT 
			cluster,
			shard_num,
			shard_weight,
			replica_num,
			host_name,
			host_address,
			port,
			is_local,
			user,
			default_database,
			errors_count,
			slowdowns_count,
			estimated_recovery_time
		FROM system.clusters
		ORDER BY cluster, shard_num, replica_num`
}

// GetDistributedDDLQueueQuery returns the query to get distributed DDL queue.
func GetDistributedDDLQueueQuery() string {
	return `
		SELECT 
			entry,
			host_name,
			port,
			status,
			cluster,
			query,
			initiator,
			exception_code,
			exception_text,
			query_create_time,
			query_finish_time
		FROM system.distributed_ddl_queue
		ORDER BY query_create_time DESC`
}

// GetMetricsQuery returns the query to get current metrics.
func GetMetricsQuery() string {
	return `
		SELECT 
			metric,
			value,
			description
		FROM system.metrics
		ORDER BY metric`
}

// GetAsyncMetricsQuery returns the query to get asynchronous metrics.
func GetAsyncMetricsQuery() string {
	return `
		SELECT 
			metric,
			value
		FROM system.asynchronous_metrics
		ORDER BY metric`
}

// GetEventsQuery returns the query to get events.
func GetEventsQuery() string {
	return `
		SELECT 
			event,
			value,
			description
		FROM system.events
		WHERE value > 0
		ORDER BY event`
}