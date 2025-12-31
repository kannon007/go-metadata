// @ts-ignore
/* eslint-disable */

declare namespace DataSourceAPI {
  // 数据源类型
  type DataSourceType =
    | 'mysql'
    | 'postgresql'
    | 'oracle'
    | 'sqlserver'
    | 'mongodb'
    | 'redis'
    | 'kafka'
    | 'rabbitmq'
    | 'minio'
    | 'clickhouse'
    | 'doris'
    | 'hive'
    | 'elasticsearch';

  // 数据源状态
  type DataSourceStatus = 'active' | 'inactive' | 'error' | 'testing';

  // 连接配置 - 使用索引签名支持不同数据源类型的配置
  type ConnectionConfig = {
    host?: string;
    port?: number;
    database?: string;
    username?: string;
    password?: string;
    ssl?: boolean;
    ssl_mode?: string;
    timeout?: number;
    max_conns?: number;
    max_idle_conns?: number;
    charset?: string;
    extra?: Record<string, string>;
    // 允许其他数据源特有的配置字段
    [key: string]: any;
  };

  // 数据源
  type DataSource = {
    id: string;
    name: string;
    type: DataSourceType;
    description: string;
    config: ConnectionConfig;
    status: DataSourceStatus;
    tags: string[];
    created_by: string;
    created_at: string;
    updated_at: string;
    last_test_at?: string;
  };

  // 调度配置
  type ScheduleConfig = {
    enabled?: boolean;
    type?: 'interval' | 'cron';
    interval?: number;
    cron?: string;
    timeout?: number;
    retries?: number;
  };

  // 创建数据源请求
  type CreateDataSourceRequest = {
    name: string;
    type: DataSourceType;
    description?: string;
    config: ConnectionConfig;
    tags?: string[];
    schedule?: ScheduleConfig;
  };

  // 更新数据源请求
  type UpdateDataSourceRequest = {
    name: string;
    description?: string;
    config?: ConnectionConfig;
    tags?: string[];
  };

  // 列表请求参数
  type ListDataSourcesParams = {
    current?: number;
    pageSize?: number;
    type?: DataSourceType;
    status?: DataSourceStatus;
    name?: string;
  };

  // 列表响应
  type ListDataSourcesResponse = {
    data: DataSource[];
    total: number;
    success: boolean;
  };

  // 测试连接请求
  type TestConnectionRequest = {
    type: DataSourceType;
    config: ConnectionConfig;
  };

  // 连接测试结果
  type ConnectionTestResult = {
    success: boolean;
    message: string;
    latency: number;
    server_info?: string;
    database_info?: string;
    version?: string;
  };

  // 批量更新状态请求
  type BatchUpdateStatusRequest = {
    ids: string[];
    status: DataSourceStatus;
  };

  // 批量操作结果
  type BatchOperationResult = {
    total: number;
    success: number;
    failed: number;
    errors?: string[];
  };

  // 检查是否可以删除的结果
  type CanDeleteResult = {
    can_delete: boolean;
    reason?: string;
    associated_tasks?: number;
  };

  // 批量导入请求
  type BatchImportRequest = {
    format: 'json' | 'yaml';
    data: string;
  };

  // 批量导出请求
  type BatchExportRequest = {
    ids?: string[];
    format: 'json' | 'yaml';
  };

  // 批量导出响应
  type BatchExportResponse = {
    format: string;
    data: string;
    count: number;
  };
}
