// @ts-ignore
/* eslint-disable */

declare namespace TaskAPI {
  // 任务类型
  type TaskType = 'full_collection' | 'incremental_collection' | 'schema_only' | 'data_profile';

  // 任务状态
  type TaskStatus = 'active' | 'inactive' | 'running' | 'completed' | 'failed' | 'paused';

  // 调度器类型
  type SchedulerType = 'dolphinscheduler' | 'builtin';

  // 调度类型
  type ScheduleType = 'once' | 'cron' | 'interval';

  // 执行状态
  type ExecutionStatus = 'pending' | 'running' | 'completed' | 'failed' | 'cancelled';

  // 任务配置
  type TaskConfig = {
    include_schemas?: string[];
    exclude_schemas?: string[];
    include_tables?: string[];
    exclude_tables?: string[];
    batch_size: number;
    timeout: number;
    retry_count: number;
    retry_interval: number;
    incremental_column?: string;
    incremental_value?: string;
    extra?: Record<string, string>;
  };

  // 调度配置
  type ScheduleConfig = {
    type: ScheduleType;
    cron_expr?: string;
    interval?: number;
    start_time?: string;
    end_time?: string;
    timezone: string;
  };

  // 采集任务
  type CollectionTask = {
    id: string;
    name: string;
    datasource_id: string;
    type: TaskType;
    config: TaskConfig;
    schedule?: ScheduleConfig;
    status: TaskStatus;
    scheduler_type: SchedulerType;
    external_id?: string;
    created_by: string;
    created_at: string;
    updated_at: string;
    last_executed_at?: string;
    next_execute_at?: string;
  };

  // 创建任务请求
  type CreateTaskRequest = {
    name: string;
    datasource_id: string;
    type: TaskType;
    config?: TaskConfig;
    schedule?: ScheduleConfig;
    scheduler_type: SchedulerType;
  };

  // 更新任务请求
  type UpdateTaskRequest = {
    name: string;
    config?: TaskConfig;
    schedule?: ScheduleConfig;
  };

  // 列表请求参数
  type ListTasksParams = {
    current?: number;
    pageSize?: number;
    datasource_id?: string;
    status?: TaskStatus;
    type?: TaskType;
  };

  // 列表响应
  type ListTasksResponse = {
    data: CollectionTask[];
    total: number;
    success: boolean;
  };

  // 执行结果
  type ExecutionResult = {
    tables_processed: number;
    records_processed: number;
    errors_count: number;
    warnings_count: number;
    processing_stats?: Record<string, number>;
  };

  // 任务执行记录
  type TaskExecution = {
    id: string;
    task_id: string;
    status: ExecutionStatus;
    start_time: string;
    end_time?: string;
    duration: number;
    result?: ExecutionResult;
    error_message?: string;
    logs?: string;
    external_id?: string;
    created_at: string;
  };

  // 执行记录列表请求参数
  type ListExecutionsParams = {
    current?: number;
    pageSize?: number;
    task_id?: string;
    status?: ExecutionStatus;
    start_time?: string;
    end_time?: string;
  };

  // 执行记录列表响应
  type ListExecutionsResponse = {
    data: TaskExecution[];
    total: number;
    success: boolean;
  };

  // 任务状态信息
  type TaskStatusInfo = {
    task_id: string;
    status: TaskStatus;
    last_execution?: TaskExecution;
    next_execute_at?: string;
    allowed_actions: string[];
  };

  // 批量操作结果
  type BatchOperationResult = {
    total: number;
    success: number;
    failed: number;
    results?: Record<string, any>;
  };
}
