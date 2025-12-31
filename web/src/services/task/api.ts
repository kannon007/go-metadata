// @ts-ignore
/* eslint-disable */
import { request } from '@umijs/max';

/** 获取任务列表 GET /api/v1/tasks */
export async function listTasks(
  params: TaskAPI.ListTasksParams,
  options?: { [key: string]: any },
) {
  const response = await request<{
    tasks: TaskAPI.CollectionTask[];
    total: number;
    page: number;
    page_size: number;
  }>('/api/v1/tasks', {
    method: 'GET',
    params: {
      page: params.current,
      page_size: params.pageSize,
      datasource_id: params.datasource_id,
      status: params.status,
      type: params.type,
    },
    ...(options || {}),
  });

  return {
    data: response.tasks || [],
    total: response.total || 0,
    success: true,
  };
}

/** 获取单个任务 GET /api/v1/tasks/:id */
export async function getTask(id: string, options?: { [key: string]: any }) {
  return request<TaskAPI.CollectionTask>(`/api/v1/tasks/${id}`, {
    method: 'GET',
    ...(options || {}),
  });
}

/** 创建任务 POST /api/v1/tasks */
export async function createTask(
  data: TaskAPI.CreateTaskRequest,
  options?: { [key: string]: any },
) {
  return request<TaskAPI.CollectionTask>('/api/v1/tasks', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    data,
    ...(options || {}),
  });
}

/** 更新任务 PUT /api/v1/tasks/:id */
export async function updateTask(
  id: string,
  data: TaskAPI.UpdateTaskRequest,
  options?: { [key: string]: any },
) {
  return request<TaskAPI.CollectionTask>(`/api/v1/tasks/${id}`, {
    method: 'PUT',
    headers: { 'Content-Type': 'application/json' },
    data,
    ...(options || {}),
  });
}

/** 删除任务 DELETE /api/v1/tasks/:id */
export async function deleteTask(id: string, options?: { [key: string]: any }) {
  return request<{ success: boolean }>(`/api/v1/tasks/${id}`, {
    method: 'DELETE',
    ...(options || {}),
  });
}

/** 启动任务 POST /api/v1/tasks/:id/start */
export async function startTask(id: string, options?: { [key: string]: any }) {
  return request<{ success: boolean }>(`/api/v1/tasks/${id}/start`, {
    method: 'POST',
    ...(options || {}),
  });
}

/** 停止任务 POST /api/v1/tasks/:id/stop */
export async function stopTask(id: string, options?: { [key: string]: any }) {
  return request<{ success: boolean }>(`/api/v1/tasks/${id}/stop`, {
    method: 'POST',
    ...(options || {}),
  });
}

/** 暂停任务 POST /api/v1/tasks/:id/pause */
export async function pauseTask(id: string, options?: { [key: string]: any }) {
  return request<{ success: boolean }>(`/api/v1/tasks/${id}/pause`, {
    method: 'POST',
    ...(options || {}),
  });
}

/** 恢复任务 POST /api/v1/tasks/:id/resume */
export async function resumeTask(id: string, options?: { [key: string]: any }) {
  return request<{ success: boolean }>(`/api/v1/tasks/${id}/resume`, {
    method: 'POST',
    ...(options || {}),
  });
}

/** 立即执行任务 POST /api/v1/tasks/:id/execute */
export async function executeTask(id: string, options?: { [key: string]: any }) {
  return request<TaskAPI.TaskExecution>(`/api/v1/tasks/${id}/execute`, {
    method: 'POST',
    ...(options || {}),
  });
}

/** 获取任务状态 GET /api/v1/tasks/:id/status */
export async function getTaskStatus(id: string, options?: { [key: string]: any }) {
  return request<TaskAPI.TaskStatusInfo>(`/api/v1/tasks/${id}/status`, {
    method: 'GET',
    ...(options || {}),
  });
}

/** 获取任务执行历史 GET /api/v1/tasks/:id/executions */
export async function listTaskExecutions(
  id: string,
  params: TaskAPI.ListExecutionsParams,
  options?: { [key: string]: any },
) {
  const response = await request<{
    executions: TaskAPI.TaskExecution[];
    total: number;
    page: number;
    page_size: number;
  }>(`/api/v1/tasks/${id}/executions`, {
    method: 'GET',
    params: {
      page: params.current,
      page_size: params.pageSize,
      status: params.status,
      start_time: params.start_time,
      end_time: params.end_time,
    },
    ...(options || {}),
  });

  return {
    data: response.executions || [],
    total: response.total || 0,
    success: true,
  };
}

/** 获取执行日志 GET /api/v1/executions/:id/logs */
export async function getExecutionLogs(executionId: string, options?: { [key: string]: any }) {
  return request<{ logs: string }>(`/api/v1/executions/${executionId}/logs`, {
    method: 'GET',
    ...(options || {}),
  });
}

/** 批量启动任务 POST /api/v1/tasks/batch/start */
export async function batchStartTasks(ids: string[], options?: { [key: string]: any }) {
  return request<TaskAPI.BatchOperationResult>('/api/v1/tasks/batch/start', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    data: { ids },
    ...(options || {}),
  });
}

/** 批量停止任务 POST /api/v1/tasks/batch/stop */
export async function batchStopTasks(ids: string[], options?: { [key: string]: any }) {
  return request<TaskAPI.BatchOperationResult>('/api/v1/tasks/batch/stop', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    data: { ids },
    ...(options || {}),
  });
}

/** 批量删除任务 POST /api/v1/tasks/batch/delete */
export async function batchDeleteTasks(ids: string[], options?: { [key: string]: any }) {
  return request<TaskAPI.BatchOperationResult>('/api/v1/tasks/batch/delete', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    data: { ids },
    ...(options || {}),
  });
}
