// Package oracle provides Oracle Database metadata queries.
package oracle

// GetAllUsersQuery returns the query to get all users (schemas) in Oracle.
func GetAllUsersQuery() string {
	return `
		SELECT USERNAME 
		FROM ALL_USERS 
		WHERE USERNAME NOT IN (
			'SYS', 'SYSTEM', 'DBSNMP', 'SYSMAN', 'OUTLN', 'MGMT_VIEW',
			'DIP', 'ORACLE_OCM', 'APPQOSSYS', 'WMSYS', 'EXFSYS', 'CTXSYS',
			'ANONYMOUS', 'XDB', 'XS$NULL', 'OJVMSYS', 'LBACSYS', 'APEX_040000',
			'APEX_PUBLIC_USER', 'FLOWS_FILES', 'MDDATA', 'SPATIAL_CSW_ADMIN_USR',
			'SPATIAL_WFS_ADMIN_USR', 'PUBLIC'
		)
		ORDER BY USERNAME`
}

// GetAllTablesQuery returns the query to get all tables in a schema.
func GetAllTablesQuery() string {
	return `
		SELECT TABLE_NAME 
		FROM ALL_TABLES 
		WHERE OWNER = :1
		UNION
		SELECT VIEW_NAME as TABLE_NAME
		FROM ALL_VIEWS 
		WHERE OWNER = :1
		ORDER BY TABLE_NAME`
}

// GetAllTabColumnsQuery returns the query to get all columns for a table.
func GetAllTabColumnsQuery() string {
	return `
		SELECT 
			COLUMN_NAME,
			DATA_TYPE,
			COLUMN_ID,
			DATA_LENGTH,
			DATA_PRECISION,
			DATA_SCALE,
			NULLABLE,
			DATA_DEFAULT,
			NVL(cc.COMMENTS, '') as COMMENTS
		FROM ALL_TAB_COLUMNS c
		LEFT JOIN ALL_COL_COMMENTS cc ON c.OWNER = cc.OWNER 
			AND c.TABLE_NAME = cc.TABLE_NAME 
			AND c.COLUMN_NAME = cc.COLUMN_NAME
		WHERE c.OWNER = :1 AND c.TABLE_NAME = :2
		ORDER BY COLUMN_ID`
}

// GetAllIndexesQuery returns the query to get all indexes for a table.
func GetAllIndexesQuery() string {
	return `
		SELECT 
			i.INDEX_NAME,
			ic.COLUMN_NAME,
			ic.COLUMN_POSITION,
			i.UNIQUENESS
		FROM ALL_INDEXES i
		JOIN ALL_IND_COLUMNS ic ON i.OWNER = ic.INDEX_OWNER AND i.INDEX_NAME = ic.INDEX_NAME
		WHERE i.TABLE_OWNER = :1 AND i.TABLE_NAME = :2
		ORDER BY i.INDEX_NAME, ic.COLUMN_POSITION`
}

// GetAllConstraintsQuery returns the query to get primary key constraints for a table.
func GetAllConstraintsQuery() string {
	return `
		SELECT cc.COLUMN_NAME
		FROM ALL_CONSTRAINTS c
		JOIN ALL_CONS_COLUMNS cc ON c.OWNER = cc.OWNER 
			AND c.CONSTRAINT_NAME = cc.CONSTRAINT_NAME
		WHERE c.OWNER = :1 
			AND c.TABLE_NAME = :2 
			AND c.CONSTRAINT_TYPE = 'P'
		ORDER BY cc.POSITION`
}

// GetPrimaryKeyQuery returns the query to get primary key constraints for a table.
func GetPrimaryKeyQuery() string {
	return `
		SELECT cc.COLUMN_NAME
		FROM ALL_CONSTRAINTS c
		JOIN ALL_CONS_COLUMNS cc ON c.OWNER = cc.OWNER 
			AND c.CONSTRAINT_NAME = cc.CONSTRAINT_NAME
		WHERE c.OWNER = :1 
			AND c.TABLE_NAME = :2 
			AND c.CONSTRAINT_TYPE = 'P'
		ORDER BY cc.POSITION`
}

// GetAllTabPartitionsQuery returns the query to get partition information for a table.
func GetAllTabPartitionsQuery() string {
	return `
		SELECT 
			PARTITION_NAME,
			HIGH_VALUE,
			NUM_ROWS
		FROM ALL_TAB_PARTITIONS
		WHERE TABLE_OWNER = :1 AND TABLE_NAME = :2
		ORDER BY PARTITION_POSITION`
}

// GetTableStatsQuery returns the query to get table statistics.
func GetTableStatsQuery() string {
	return `
		SELECT 
			NUM_ROWS,
			BLOCKS,
			AVG_ROW_LEN,
			LAST_ANALYZED
		FROM ALL_TABLES
		WHERE OWNER = :1 AND TABLE_NAME = :2`
}

// GetTableCommentsQuery returns the query to get table comments.
func GetTableCommentsQuery() string {
	return `
		SELECT NVL(COMMENTS, '') as COMMENTS
		FROM ALL_TAB_COMMENTS
		WHERE OWNER = :1 AND TABLE_NAME = :2`
}

// GetColumnCommentsQuery returns the query to get column comments.
func GetColumnCommentsQuery() string {
	return `
		SELECT 
			COLUMN_NAME,
			NVL(COMMENTS, '') as COMMENTS
		FROM ALL_COL_COMMENTS
		WHERE OWNER = :1 AND TABLE_NAME = :2
		ORDER BY COLUMN_NAME`
}

// GetViewDefinitionQuery returns the query to get view definition.
func GetViewDefinitionQuery() string {
	return `
		SELECT TEXT
		FROM ALL_VIEWS
		WHERE OWNER = :1 AND VIEW_NAME = :2`
}

// GetTablespacesQuery returns the query to get tablespace information.
func GetTablespacesQuery() string {
	return `
		SELECT 
			TABLESPACE_NAME,
			STATUS,
			CONTENTS,
			LOGGING
		FROM DBA_TABLESPACES
		ORDER BY TABLESPACE_NAME`
}

// GetUserTablespacesQuery returns the query to get user tablespace information.
func GetUserTablespacesQuery() string {
	return `
		SELECT 
			TABLESPACE_NAME,
			STATUS,
			CONTENTS
		FROM USER_TABLESPACES
		ORDER BY TABLESPACE_NAME`
}

// GetSequencesQuery returns the query to get sequences in a schema.
func GetSequencesQuery() string {
	return `
		SELECT 
			SEQUENCE_NAME,
			MIN_VALUE,
			MAX_VALUE,
			INCREMENT_BY,
			CYCLE_FLAG,
			ORDER_FLAG,
			CACHE_SIZE,
			LAST_NUMBER
		FROM ALL_SEQUENCES
		WHERE SEQUENCE_OWNER = :1
		ORDER BY SEQUENCE_NAME`
}

// GetTriggersQuery returns the query to get triggers for a table.
func GetTriggersQuery() string {
	return `
		SELECT 
			TRIGGER_NAME,
			TRIGGER_TYPE,
			TRIGGERING_EVENT,
			STATUS,
			DESCRIPTION
		FROM ALL_TRIGGERS
		WHERE OWNER = :1 AND TABLE_NAME = :2
		ORDER BY TRIGGER_NAME`
}

// GetForeignKeysQuery returns the query to get foreign key constraints for a table.
func GetForeignKeysQuery() string {
	return `
		SELECT 
			c.CONSTRAINT_NAME,
			cc.COLUMN_NAME,
			c.R_OWNER,
			rc.TABLE_NAME as REFERENCED_TABLE,
			rcc.COLUMN_NAME as REFERENCED_COLUMN
		FROM ALL_CONSTRAINTS c
		JOIN ALL_CONS_COLUMNS cc ON c.OWNER = cc.OWNER 
			AND c.CONSTRAINT_NAME = cc.CONSTRAINT_NAME
		JOIN ALL_CONSTRAINTS rc ON c.R_OWNER = rc.OWNER 
			AND c.R_CONSTRAINT_NAME = rc.CONSTRAINT_NAME
		JOIN ALL_CONS_COLUMNS rcc ON rc.OWNER = rcc.OWNER 
			AND rc.CONSTRAINT_NAME = rcc.CONSTRAINT_NAME
			AND cc.POSITION = rcc.POSITION
		WHERE c.OWNER = :1 
			AND c.TABLE_NAME = :2 
			AND c.CONSTRAINT_TYPE = 'R'
		ORDER BY c.CONSTRAINT_NAME, cc.POSITION`
}

// GetCheckConstraintsQuery returns the query to get check constraints for a table.
func GetCheckConstraintsQuery() string {
	return `
		SELECT 
			CONSTRAINT_NAME,
			SEARCH_CONDITION,
			STATUS
		FROM ALL_CONSTRAINTS
		WHERE OWNER = :1 
			AND TABLE_NAME = :2 
			AND CONSTRAINT_TYPE = 'C'
			AND CONSTRAINT_NAME NOT LIKE 'SYS_%'
		ORDER BY CONSTRAINT_NAME`
}

// GetSynonymsQuery returns the query to get synonyms in a schema.
func GetSynonymsQuery() string {
	return `
		SELECT 
			SYNONYM_NAME,
			TABLE_OWNER,
			TABLE_NAME,
			DB_LINK
		FROM ALL_SYNONYMS
		WHERE OWNER = :1
		ORDER BY SYNONYM_NAME`
}

// GetMaterializedViewsQuery returns the query to get materialized views in a schema.
func GetMaterializedViewsQuery() string {
	return `
		SELECT 
			MVIEW_NAME,
			CONTAINER_NAME,
			QUERY,
			REFRESH_MODE,
			REFRESH_METHOD,
			BUILD_MODE,
			FAST_REFRESHABLE,
			LAST_REFRESH_DATE
		FROM ALL_MVIEWS
		WHERE OWNER = :1
		ORDER BY MVIEW_NAME`
}

// GetDatabaseInfoQuery returns the query to get database information.
func GetDatabaseInfoQuery() string {
	return `
		SELECT 
			NAME,
			DBID,
			CREATED,
			LOG_MODE,
			OPEN_MODE,
			DATABASE_ROLE,
			PLATFORM_NAME
		FROM V$DATABASE`
}

// GetInstanceInfoQuery returns the query to get instance information.
func GetInstanceInfoQuery() string {
	return `
		SELECT 
			INSTANCE_NAME,
			HOST_NAME,
			VERSION,
			STARTUP_TIME,
			STATUS,
			DATABASE_STATUS,
			INSTANCE_ROLE
		FROM V$INSTANCE`
}