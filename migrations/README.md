# 元数据系统数据库迁移文件

本目录包含元数据系统的数据库表结构和索引的SQL脚本。

## 文件说明

### 001_create_metadata_tables.sql
创建元数据系统的核心表结构，包括：

**数据源管理**
- `connectors` - 数据源连接器配置表

**元数据核心表**
- `catalogs` - 数据目录表
- `schemas` - 数据库/模式表  
- `tables` - 表/视图表
- `columns` - 列信息表

**统计信息表**
- `column_statistics` - 列统计信息表
- `partitions` - 分区信息表

**血缘关系表**
- `lineage_nodes` - 血缘节点表
- `lineage_edges` - 血缘边表

**作业和任务表**
- `collection_tasks` - 采集任务表
- `lineage_jobs` - 血缘分析作业表

**系统表**
- `system_configs` - 系统配置表
- `audit_logs` - 操作审计日志表

**视图**
- `v_table_info` - 完整表信息视图
- `v_column_info` - 完整列信息视图
- `v_lineage_graph` - 血缘关系视图

### 002_create_indexes.sql
创建性能优化索引，包括：

**性能优化索引**
- 复合索引优化查询性能
- 血缘查询优化索引
- 任务查询优化索引
- 全文搜索索引

## 执行顺序

请按以下顺序执行SQL文件：

1. `001_create_metadata_tables.sql` - 创建表结构
2. `002_create_indexes.sql` - 创建索引

## 数据库要求

- MySQL 8.0+ 或 MariaDB 10.5+
- 支持JSON数据类型
- 建议配置：
  - `innodb_buffer_pool_size` >= 1GB
  - `max_connections` >= 200

## 使用示例

### 创建数据库
```sql
CREATE DATABASE metadata_system CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
USE metadata_system;
```

### 执行迁移
```bash
# 创建表结构
mysql -u root -p metadata_system < 001_create_metadata_tables.sql

# 创建索引
mysql -u root -p metadata_system < 002_create_indexes.sql
```

### 查询示例

**查看所有连接器**
```sql
SELECT id, name, type, status FROM connectors;
```

**查看表信息**
```sql
SELECT * FROM v_table_info WHERE catalog_name = 'mysql_prod';
```

**查看列信息**
```sql
SELECT * FROM v_column_info WHERE table_name = 'orders';
```

**查看血缘关系**
```sql
SELECT * FROM v_lineage_graph WHERE source_display_name LIKE '%orders%';
```

## 扩展说明

### 添加新的数据源类型
1. 在 `connectors` 表的 `type` 字段枚举中添加新类型
2. 根据需要扩展 `properties` JSON字段的结构

### 自定义血缘关系类型
1. 在 `lineage_edges` 表的 `type` 字段枚举中添加新类型
2. 根据需要扩展 `properties` JSON字段

### 性能优化建议
1. 根据实际查询模式添加合适的索引
2. 对大表考虑分区策略
3. 监控慢查询并优化

## 注意事项

1. **数据安全**：`credentials` 字段存储的认证信息需要加密处理
2. **权限控制**：建议创建专用的数据库用户，限制权限范围
3. **备份策略**：制定定期备份计划，特别是元数据和血缘关系数据
4. **监控告警**：监控数据库性能和存储空间使用情况