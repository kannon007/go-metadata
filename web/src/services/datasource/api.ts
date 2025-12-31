// @ts-ignore
/* eslint-disable */
import { request } from '@umijs/max';

/** 获取数据源列表 GET /api/v1/datasources */
export async function listDataSources(
  params: DataSourceAPI.ListDataSourcesParams,
  options?: { [key: string]: any },
) {
  const response = await request<{
    data_sources: DataSourceAPI.DataSource[];
    total: number;
    page: number;
    page_size: number;
  }>('/api/v1/datasources', {
    method: 'GET',
    params: {
      page: params.current,
      page_size: params.pageSize,
      type: params.type,
      status: params.status,
      name: params.name,
    },
    ...(options || {}),
  });

  return {
    data: response.data_sources || [],
    total: response.total || 0,
    success: true,
  };
}

/** 获取单个数据源 GET /api/v1/datasources/:id */
export async function getDataSource(id: string, options?: { [key: string]: any }) {
  return request<DataSourceAPI.DataSource>(`/api/v1/datasources/${id}`, {
    method: 'GET',
    ...(options || {}),
  });
}

/** 创建数据源 POST /api/v1/datasources */
export async function createDataSource(
  data: DataSourceAPI.CreateDataSourceRequest,
  options?: { [key: string]: any },
) {
  return request<DataSourceAPI.DataSource>('/api/v1/datasources', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
    data,
    ...(options || {}),
  });
}

/** 更新数据源 PUT /api/v1/datasources/:id */
export async function updateDataSource(
  id: string,
  data: DataSourceAPI.UpdateDataSourceRequest,
  options?: { [key: string]: any },
) {
  return request<DataSourceAPI.DataSource>(`/api/v1/datasources/${id}`, {
    method: 'PUT',
    headers: {
      'Content-Type': 'application/json',
    },
    data,
    ...(options || {}),
  });
}


/** 删除数据源 DELETE /api/v1/datasources/:id */
export async function deleteDataSource(id: string, options?: { [key: string]: any }) {
  return request<{ success: boolean }>(`/api/v1/datasources/${id}`, {
    method: 'DELETE',
    ...(options || {}),
  });
}

/** 测试连接 POST /api/v1/datasources/test */
export async function testConnection(
  data: DataSourceAPI.TestConnectionRequest,
  options?: { [key: string]: any },
) {
  return request<DataSourceAPI.ConnectionTestResult>('/api/v1/datasources/test', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
    data,
    ...(options || {}),
  });
}

/** 测试已有数据源连接 POST /api/v1/datasources/:id/test */
export async function testDataSourceConnection(id: string, options?: { [key: string]: any }) {
  return request<DataSourceAPI.ConnectionTestResult>(`/api/v1/datasources/${id}/test`, {
    method: 'POST',
    ...(options || {}),
  });
}

/** 检查是否可以删除 GET /api/v1/datasources/:id/can-delete */
export async function canDeleteDataSource(id: string, options?: { [key: string]: any }) {
  return request<DataSourceAPI.CanDeleteResult>(`/api/v1/datasources/${id}/can-delete`, {
    method: 'GET',
    ...(options || {}),
  });
}

/** 批量更新状态 POST /api/v1/datasources/batch/status */
export async function batchUpdateStatus(
  data: DataSourceAPI.BatchUpdateStatusRequest,
  options?: { [key: string]: any },
) {
  return request<DataSourceAPI.BatchOperationResult>('/api/v1/datasources/batch/status', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
    data,
    ...(options || {}),
  });
}

/** 批量导入 POST /api/v1/datasources/import */
export async function batchImport(
  data: DataSourceAPI.BatchImportRequest,
  options?: { [key: string]: any },
) {
  return request<DataSourceAPI.BatchOperationResult>('/api/v1/datasources/import', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
    data,
    ...(options || {}),
  });
}

/** 批量导出 POST /api/v1/datasources/export */
export async function batchExport(
  data: DataSourceAPI.BatchExportRequest,
  options?: { [key: string]: any },
) {
  return request<DataSourceAPI.BatchExportResponse>('/api/v1/datasources/export', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
    data,
    ...(options || {}),
  });
}

/** 批量删除（前端循环调用单个删除） */
export async function batchDelete(ids: string[], options?: { [key: string]: any }) {
  let success = 0;
  let failed = 0;
  const errors: string[] = [];

  for (const id of ids) {
    try {
      await deleteDataSource(id, options);
      success++;
    } catch (e: any) {
      failed++;
      errors.push(e.message || `删除 ${id} 失败`);
    }
  }

  return {
    total: ids.length,
    success,
    failed,
    errors,
  } as DataSourceAPI.BatchOperationResult;
}
