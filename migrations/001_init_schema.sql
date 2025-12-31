-- 元数据系统数据库表结构
-- 版本: 1.0
-- 创建时间: 2025-12-25
-- 说明: 包含所有表、视图、索引的创建，支持重复执行

-- ============================================================================
-- 1. 删除已存在的对象（按依赖顺序）
-- ============================================================================

-- 删除视图
DROP VIEW IF EXISTS v_lineage_graph;
DROP VIEW IF EXISTS v_column_info;
DROP VIEW IF EXISTS v_table_info;

-- 删除数据源管理模块表
DROP TABLE IF EXISTS datasource_templates;
DROP TABLE IF EXISTS task_executions;
DROP TABLE IF EXISTS collection_tasks;
DROP TABLE IF EXISTS datasources;

-- 删除元数据核心表
DROP TABLE IF EXISTS audit_logs;
DROP TABLE IF EXISTS system_configs;
DROP TABLE IF EXISTS lineage_edges;
DROP TABLE IF EXISTS lineage_nodes;
DROP TABLE IF EXISTS partitions;
DROP TABLE IF EXISTS column_statistics;
DROP TABLE IF EXISTS columns;
DROP TABLE IF EXISTS tables;
DROP TABLE IF EXISTS `schemas`;
DROP TABLE IF EXISTS catalogs;
DROP TABLE IF EXISTS connectors;

-- ============================================================================
-- 2. 数据源连接器表
-- ============================================================================

CREATE TABLE connectors (
    id VARCHAR(64) PRIMARY KEY COMMENT '连接器唯一标识',
    name VARCHAR(128) NOT NULL COMMENT '连接器名称',
    type VARCHAR(32) NOT NULL COMMENT '连接器类型: mysql,postgres,hive,doris,clickhouse,s3,hdfs',
    endpoint VARCHAR(512) NOT NULL COMMENT '连接端点',
    credentials JSON COMMENT '认证信息(加密存储)',
    properties JSON COMMENT '连接器配置属性',
    status ENUM('active', 'inactive', 'error') DEFAULT 'active' COMMENT '连接器状态',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    created_by VARCHAR(64) COMMENT '创建人',
    
    INDEX idx_type (type),
    INDEX idx_status (status),
    INDEX idx_created_at (created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='数据源连接器配置表';

-- ============================================================================
-- 3. 元数据核心表
-- ============================================================================

-- 数据目录表 (Catalog)
CREATE TABLE catalogs (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    connector_id VARCHAR(64) NOT NULL COMMENT '所属连接器ID',
    name VARCHAR(128) NOT NULL COMMENT '目录名称',
    type VARCHAR(32) NOT NULL COMMENT '目录类型',
    description TEXT COMMENT '目录描述',
    properties JSON COMMENT '目录属性',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    
    UNIQUE KEY uk_connector_name (connector_id, name),
    FOREIGN KEY (connector_id) REFERENCES connectors(id) ON DELETE CASCADE,
    INDEX idx_connector_id (connector_id),
    INDEX idx_type (type)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='数据目录表';

-- 数据库/模式表 (Schema/Database)
CREATE TABLE `schemas` (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    catalog_id BIGINT NOT NULL COMMENT '所属目录ID',
    name VARCHAR(128) NOT NULL COMMENT '模式名称',
    description TEXT COMMENT '模式描述',
    owner VARCHAR(64) COMMENT '所有者',
    properties JSON COMMENT '模式属性',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    
    UNIQUE KEY uk_catalog_name (catalog_id, name),
    FOREIGN KEY (catalog_id) REFERENCES catalogs(id) ON DELETE CASCADE,
    INDEX idx_catalog_id (catalog_id),
    INDEX idx_owner (owner)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='数据库模式表';

-- 表/视图表 (Table/View)
CREATE TABLE tables (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    schema_id BIGINT NOT NULL COMMENT '所属模式ID',
    name VARCHAR(128) NOT NULL COMMENT '表名',
    type ENUM('TABLE', 'VIEW', 'EXTERNAL_TABLE', 'FILE') NOT NULL COMMENT '表类型',
    comment TEXT COMMENT '表注释',
    owner VARCHAR(64) COMMENT '所有者',
    storage_format VARCHAR(32) COMMENT '存储格式: parquet,orc,delta,csv,json',
    storage_location VARCHAR(1024) COMMENT '存储位置',
    storage_compressed BOOLEAN DEFAULT FALSE COMMENT '是否压缩',
    row_count BIGINT COMMENT '行数',
    data_size_bytes BIGINT COMMENT '数据大小(字节)',
    partition_count INT DEFAULT 0 COMMENT '分区数量',
    properties JSON COMMENT '表属性',
    raw_metadata JSON COMMENT '原始元数据',
    last_refreshed_at TIMESTAMP COMMENT '最后刷新时间',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    
    UNIQUE KEY uk_schema_name (schema_id, name),
    FOREIGN KEY (schema_id) REFERENCES `schemas`(id) ON DELETE CASCADE,
    INDEX idx_schema_id (schema_id),
    INDEX idx_type (type),
    INDEX idx_owner (owner),
    INDEX idx_last_refreshed (last_refreshed_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='表和视图表';

-- 列表 (Column)
CREATE TABLE columns (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    table_id BIGINT NOT NULL COMMENT '所属表ID',
    name VARCHAR(128) NOT NULL COMMENT '列名',
    ordinal_position INT NOT NULL COMMENT '列位置',
    data_type VARCHAR(64) NOT NULL COMMENT '标准化数据类型',
    source_type VARCHAR(128) NOT NULL COMMENT '原始数据类型',
    max_length INT COMMENT '最大长度',
    numeric_precision INT COMMENT '数值精度',
    numeric_scale INT COMMENT '数值标度',
    is_nullable BOOLEAN DEFAULT TRUE COMMENT '是否可空',
    default_value TEXT COMMENT '默认值',
    is_primary_key BOOLEAN DEFAULT FALSE COMMENT '是否主键',
    is_partition_column BOOLEAN DEFAULT FALSE COMMENT '是否分区列',
    is_auto_increment BOOLEAN DEFAULT FALSE COMMENT '是否自增',
    comment TEXT COMMENT '列注释',
    properties JSON COMMENT '列属性',
    raw_metadata JSON COMMENT '原始元数据',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    
    UNIQUE KEY uk_table_name (table_id, name),
    FOREIGN KEY (table_id) REFERENCES tables(id) ON DELETE CASCADE,
    INDEX idx_table_id (table_id),
    INDEX idx_ordinal_position (table_id, ordinal_position),
    INDEX idx_is_primary_key (is_primary_key),
    INDEX idx_is_partition_column (is_partition_column)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='列信息表';

-- ============================================================================
-- 4. 统计信息表
-- ============================================================================

CREATE TABLE column_statistics (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    column_id BIGINT NOT NULL COMMENT '列ID',
    distinct_count BIGINT COMMENT '唯一值数量',
    null_count BIGINT COMMENT '空值数量',
    min_value TEXT COMMENT '最小值',
    max_value TEXT COMMENT '最大值',
    avg_value DECIMAL(20,6) COMMENT '平均值',
    stddev_value DECIMAL(20,6) COMMENT '标准差',
    percentile_25 TEXT COMMENT '25分位数',
    percentile_50 TEXT COMMENT '50分位数(中位数)',
    percentile_75 TEXT COMMENT '75分位数',
    top_values JSON COMMENT 'TopN高频值统计',
    collected_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '统计时间',
    
    FOREIGN KEY (column_id) REFERENCES columns(id) ON DELETE CASCADE,
    INDEX idx_column_id (column_id),
    INDEX idx_collected_at (collected_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='列统计信息表';

-- ============================================================================
-- 5. 分区信息表
-- ============================================================================

CREATE TABLE partitions (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    table_id BIGINT NOT NULL COMMENT '所属表ID',
    name VARCHAR(256) NOT NULL COMMENT '分区名称',
    partition_values JSON NOT NULL COMMENT '分区值',
    row_count BIGINT COMMENT '分区行数',
    data_size_bytes BIGINT COMMENT '分区大小',
    location VARCHAR(1024) COMMENT '分区存储位置',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    
    UNIQUE KEY uk_table_name (table_id, name),
    FOREIGN KEY (table_id) REFERENCES tables(id) ON DELETE CASCADE,
    INDEX idx_table_id (table_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='分区信息表';

-- ============================================================================
-- 6. 血缘关系表
-- ============================================================================

CREATE TABLE lineage_nodes (
    id VARCHAR(128) PRIMARY KEY COMMENT '节点唯一标识',
    type ENUM('database', 'table', 'column', 'job', 'file') NOT NULL COMMENT '节点类型',
    catalog_name VARCHAR(128) COMMENT '目录名',
    schema_name VARCHAR(128) COMMENT '模式名',
    table_name VARCHAR(128) COMMENT '表名',
    column_name VARCHAR(128) COMMENT '列名',
    display_name VARCHAR(256) NOT NULL COMMENT '显示名称',
    description TEXT COMMENT '节点描述',
    table_id BIGINT COMMENT '关联的表ID',
    column_id BIGINT COMMENT '关联的列ID',
    properties JSON COMMENT '节点属性',
    tags JSON COMMENT '标签',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    
    INDEX idx_type (type),
    INDEX idx_catalog_schema (catalog_name, schema_name),
    INDEX idx_table_column (table_name, column_name),
    INDEX idx_table_id (table_id),
    INDEX idx_column_id (column_id),
    FOREIGN KEY (table_id) REFERENCES tables(id) ON DELETE SET NULL,
    FOREIGN KEY (column_id) REFERENCES columns(id) ON DELETE SET NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='血缘节点表';

CREATE TABLE lineage_edges (
    id VARCHAR(128) PRIMARY KEY COMMENT '边唯一标识',
    type ENUM('contains', 'depends_on', 'produced_by', 'derived_from') NOT NULL COMMENT '边类型',
    source_node_id VARCHAR(128) NOT NULL COMMENT '源节点ID',
    target_node_id VARCHAR(128) NOT NULL COMMENT '目标节点ID',
    transformation_type VARCHAR(64) COMMENT '转换类型: select,insert,update,delete,join,union',
    sql_statement TEXT COMMENT '相关SQL语句',
    job_name VARCHAR(256) COMMENT '作业名称',
    properties JSON COMMENT '边属性',
    confidence DECIMAL(3,2) DEFAULT 1.00 COMMENT '置信度(0-1)',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    
    INDEX idx_type (type),
    INDEX idx_source_node (source_node_id),
    INDEX idx_target_node (target_node_id),
    INDEX idx_source_target (source_node_id, target_node_id),
    FOREIGN KEY (source_node_id) REFERENCES lineage_nodes(id) ON DELETE CASCADE,
    FOREIGN KEY (target_node_id) REFERENCES lineage_nodes(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='血缘边表';

-- ============================================================================
-- 7. 系统配置和日志表
-- ============================================================================

CREATE TABLE system_configs (
    id INT AUTO_INCREMENT PRIMARY KEY,
    config_key VARCHAR(128) NOT NULL COMMENT '配置键',
    config_value TEXT COMMENT '配置值',
    description TEXT COMMENT '配置描述',
    is_encrypted BOOLEAN DEFAULT FALSE COMMENT '是否加密',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    
    UNIQUE KEY uk_config_key (config_key)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='系统配置表';

CREATE TABLE audit_logs (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    user_id VARCHAR(64) COMMENT '用户ID',
    action VARCHAR(64) NOT NULL COMMENT '操作动作',
    resource_type VARCHAR(64) COMMENT '资源类型',
    resource_id VARCHAR(128) COMMENT '资源ID',
    details JSON COMMENT '操作详情',
    ip_address VARCHAR(45) COMMENT 'IP地址',
    user_agent TEXT COMMENT '用户代理',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    INDEX idx_user_id (user_id),
    INDEX idx_action (action),
    INDEX idx_resource (resource_type, resource_id),
    INDEX idx_created_at (created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='操作审计日志表';

-- ============================================================================
-- 8. 数据源管理模块表
-- ============================================================================

CREATE TABLE datasources (
    id VARCHAR(36) PRIMARY KEY,
    name VARCHAR(255) NOT NULL UNIQUE,
    type VARCHAR(50) NOT NULL,
    description TEXT,
    config TEXT NOT NULL,
    status VARCHAR(20) DEFAULT 'inactive',
    tags TEXT,
    created_by VARCHAR(255),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    last_test_at TIMESTAMP,
    
    INDEX idx_datasources_type (type),
    INDEX idx_datasources_status (status),
    INDEX idx_datasources_created_by (created_by),
    INDEX idx_datasources_created_at (created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='数据源配置表';

CREATE TABLE collection_tasks (
    id VARCHAR(36) PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    datasource_id VARCHAR(36) NOT NULL,
    type VARCHAR(50) NOT NULL,
    config TEXT NOT NULL,
    schedule TEXT,
    status VARCHAR(20) DEFAULT 'inactive',
    scheduler_type VARCHAR(50) NOT NULL,
    external_id VARCHAR(255),
    created_by VARCHAR(255),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    last_executed_at TIMESTAMP,
    next_execute_at TIMESTAMP,
    
    INDEX idx_collection_tasks_datasource_id (datasource_id),
    INDEX idx_collection_tasks_status (status),
    INDEX idx_collection_tasks_type (type),
    INDEX idx_collection_tasks_scheduler_type (scheduler_type),
    INDEX idx_collection_tasks_created_by (created_by),
    INDEX idx_collection_tasks_next_execute (next_execute_at),
    
    FOREIGN KEY (datasource_id) REFERENCES datasources(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='采集任务表';

CREATE TABLE task_executions (
    id VARCHAR(36) PRIMARY KEY,
    task_id VARCHAR(36) NOT NULL,
    status VARCHAR(20) NOT NULL,
    start_time TIMESTAMP NOT NULL,
    end_time TIMESTAMP,
    duration BIGINT,
    result TEXT,
    error_message TEXT,
    logs TEXT,
    external_id VARCHAR(255),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    INDEX idx_task_executions_task_id (task_id),
    INDEX idx_task_executions_status (status),
    INDEX idx_task_executions_start_time (start_time),
    INDEX idx_task_executions_external_id (external_id),
    
    FOREIGN KEY (task_id) REFERENCES collection_tasks(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='任务执行记录表';

CREATE TABLE datasource_templates (
    id VARCHAR(36) PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    type VARCHAR(50) NOT NULL,
    description TEXT,
    config_template TEXT NOT NULL,
    is_system BOOLEAN DEFAULT FALSE,
    created_by VARCHAR(255),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    
    INDEX idx_datasource_templates_type (type),
    INDEX idx_datasource_templates_is_system (is_system),
    INDEX idx_datasource_templates_created_by (created_by)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='数据源模板表';


-- ============================================================================
-- 9. 创建视图
-- ============================================================================

CREATE VIEW v_table_info AS
SELECT 
    t.id as table_id,
    c.name as catalog_name,
    s.name as schema_name,
    t.name as table_name,
    t.type as table_type,
    t.comment as table_comment,
    t.owner,
    t.row_count,
    t.data_size_bytes,
    t.partition_count,
    t.storage_format,
    t.storage_location,
    t.last_refreshed_at,
    conn.name as connector_name,
    conn.type as connector_type
FROM tables t
JOIN `schemas` s ON t.schema_id = s.id
JOIN catalogs c ON s.catalog_id = c.id
JOIN connectors conn ON c.connector_id = conn.id;

CREATE VIEW v_column_info AS
SELECT 
    col.id as column_id,
    c.name as catalog_name,
    s.name as schema_name,
    t.name as table_name,
    col.name as column_name,
    col.ordinal_position,
    col.data_type,
    col.source_type,
    col.is_nullable,
    col.is_primary_key,
    col.is_partition_column,
    col.comment as column_comment,
    t.id as table_id
FROM columns col
JOIN tables t ON col.table_id = t.id
JOIN `schemas` s ON t.schema_id = s.id
JOIN catalogs c ON s.catalog_id = c.id;

CREATE VIEW v_lineage_graph AS
SELECT 
    e.id as edge_id,
    e.type as edge_type,
    sn.id as source_node_id,
    sn.type as source_node_type,
    sn.display_name as source_display_name,
    tn.id as target_node_id,
    tn.type as target_node_type,
    tn.display_name as target_display_name,
    e.transformation_type,
    e.confidence,
    e.created_at
FROM lineage_edges e
JOIN lineage_nodes sn ON e.source_node_id = sn.id
JOIN lineage_nodes tn ON e.target_node_id = tn.id;

-- ============================================================================
-- 10. 性能优化索引
-- ============================================================================

CREATE INDEX idx_tables_schema_type ON tables(schema_id, type);
CREATE INDEX idx_tables_owner_updated ON tables(owner, updated_at);
CREATE INDEX idx_columns_table_position ON columns(table_id, ordinal_position);
CREATE INDEX idx_columns_type_nullable ON columns(data_type, is_nullable);
CREATE INDEX idx_lineage_nodes_catalog_schema_table ON lineage_nodes(catalog_name, schema_name, table_name);
CREATE INDEX idx_lineage_edges_type_confidence ON lineage_edges(type, confidence);
CREATE INDEX idx_collection_tasks_datasource_status ON collection_tasks(datasource_id, status);
CREATE INDEX idx_collection_tasks_created_updated ON collection_tasks(created_at, updated_at);

-- 全文搜索索引
CREATE FULLTEXT INDEX ft_tables_name_comment ON tables(name, comment);
CREATE FULLTEXT INDEX ft_columns_name_comment ON columns(name, comment);
CREATE FULLTEXT INDEX ft_lineage_nodes_display_desc ON lineage_nodes(display_name, description);

-- ============================================================================
-- 11. 插入默认数据
-- ============================================================================

INSERT IGNORE INTO datasource_templates (id, name, type, description, config_template, is_system, created_by) VALUES
('mysql-template-1', 'MySQL Standard', 'mysql', 'Standard MySQL connection template', 
'{"host": "localhost", "port": 3306, "username": "", "password": "", "database": "", "charset": "utf8mb4", "timeout": 30}', 
true, 'system'),
('postgres-template-1', 'PostgreSQL Standard', 'postgresql', 'Standard PostgreSQL connection template', 
'{"host": "localhost", "port": 5432, "username": "", "password": "", "database": "", "sslmode": "disable", "timeout": 30}', 
true, 'system'),
('redis-template-1', 'Redis Standard', 'redis', 'Standard Redis connection template', 
'{"host": "localhost", "port": 6379, "password": "", "db": 0, "timeout": 30}', 
true, 'system'),
('mongodb-template-1', 'MongoDB Standard', 'mongodb', 'Standard MongoDB connection template', 
'{"host": "localhost", "port": 27017, "username": "", "password": "", "database": "", "auth_source": "admin", "timeout": 30}', 
true, 'system');
