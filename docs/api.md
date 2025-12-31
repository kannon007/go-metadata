# API Documentation

## Overview

元数据管理系统提供 RESTful API 和 gRPC 接口，用于管理数据源、采集任务和元数据。

## Base URL

- REST API: `http://localhost:8080/api/v1`
- gRPC: `localhost:9090`

## Authentication

所有 API 请求需要在 Header 中携带 JWT Token：

```
Authorization: Bearer <token>
```

### 获取 Token

```http
POST /api/v1/login
Content-Type: application/json

{
  "username": "admin",
  "password": "password"
}
```

**Response:**
```json
{
  "access_token": "eyJhbGciOiJIUzI1NiIs...",
  "refresh_token": "eyJhbGciOiJIUzI1NiIs...",
  "token_type": "Bearer",
  "expires_in": 3600
}
```

---

## Data Sources API

### List Data Sources

获取数据源列表。

```http
GET /api/v1/datasources
```

**Query Parameters:**
| Parameter | Type | Description |
|-----------|------|-------------|
| page | int | 页码，默认 1 |
| page_size | int | 每页数量，默认 20，最大 100 |
| type | string | 数据源类型过滤 |
| status | string | 状态过滤 |
| name | string | 名称模糊搜索 |

**Response:**
```json
{
  "data_sources": [
    {
      "id": "ds-123",
      "name": "production-mysql",
      "type": "mysql",
      "description": "Production MySQL database",
      "status": "active",
      "config": {
        "host": "localhost",
        "port": 3306,
        "database": "production"
      },
      "tags": ["production", "mysql"],
      "created_at": "2024-01-01T00:00:00Z",
      "updated_at": "2024-01-01T00:00:00Z"
    }
  ],
  "total": 1,
  "page": 1,
  "page_size": 20
}
```

### Create Data Source

创建新数据源。

```http
POST /api/v1/datasources
Content-Type: application/json

{
  "name": "production-mysql",
  "type": "mysql",
  "description": "Production MySQL database",
  "config": {
    "host": "localhost",
    "port": 3306,
    "database": "production",
    "username": "admin",
    "password": "secret",
    "timeout": 30,
    "max_conns": 10,
    "max_idle_conns": 5
  },
  "tags": ["production", "mysql"]
}
```

**Supported Types:**
- `mysql` - MySQL
- `postgresql` - PostgreSQL
- `oracle` - Oracle
- `sqlserver` - SQL Server
- `mongodb` - MongoDB
- `redis` - Redis
- `kafka` - Kafka
- `rabbitmq` - RabbitMQ
- `minio` - MinIO
- `clickhouse` - ClickHouse
- `doris` - Doris
- `hive` - Hive
- `elasticsearch` - Elasticsearch

**Response:**
```json
{
  "id": "ds-123",
  "name": "production-mysql",
  "type": "mysql",
  "status": "active",
  "created_at": "2024-01-01T00:00:00Z"
}
```

### Get Data Source

获取单个数据源详情。

```http
GET /api/v1/datasources/{id}
```

### Update Data Source

更新数据源配置。

```http
PUT /api/v1/datasources/{id}
Content-Type: application/json

{
  "name": "updated-mysql",
  "description": "Updated description",
  "config": {
    "host": "new-host",
    "port": 3306
  },
  "tags": ["updated"]
}
```

### Delete Data Source

删除数据源。

```http
DELETE /api/v1/datasources/{id}
```

**Note:** 如果数据源有关联的采集任务，删除将被阻止。

### Test Connection

测试数据源连接。

```http
POST /api/v1/datasources/{id}/test
```

**Response:**
```json
{
  "success": true,
  "message": "Connection successful",
  "latency": 50,
  "server_info": "MySQL 8.0.32",
  "version": "8.0.32"
}
```

### Batch Update Status

批量更新数据源状态。

```http
POST /api/v1/datasources/batch/status
Content-Type: application/json

{
  "ids": ["ds-1", "ds-2", "ds-3"],
  "status": "inactive"
}
```

**Response:**
```json
{
  "total": 3,
  "success": 3,
  "failed": 0,
  "errors": []
}
```

---

## Tasks API

### List Tasks

获取采集任务列表。

```http
GET /api/v1/tasks
```

**Query Parameters:**
| Parameter | Type | Description |
|-----------|------|-------------|
| page | int | 页码 |
| page_size | int | 每页数量 |
| datasource_id | string | 数据源 ID 过滤 |
| status | string | 状态过滤 |
| type | string | 任务类型过滤 |

### Create Task

创建采集任务。

```http
POST /api/v1/tasks
Content-Type: application/json

{
  "name": "daily-collection",
  "datasource_id": "ds-123",
  "type": "full_collection",
  "config": {
    "include_schemas": ["public"],
    "exclude_tables": ["temp_*"],
    "batch_size": 1000,
    "timeout": 3600,
    "retry_count": 3
  },
  "schedule": {
    "type": "cron",
    "cron_expr": "0 0 * * *",
    "timezone": "Asia/Shanghai"
  },
  "scheduler_type": "builtin"
}
```

**Task Types:**
- `full_collection` - 全量采集
- `incremental_collection` - 增量采集
- `schema_only` - 仅采集 Schema
- `data_profile` - 数据画像

**Schedule Types:**
- `immediate` - 立即执行
- `once` - 定时执行一次
- `cron` - Cron 表达式
- `interval` - 间隔执行

**Scheduler Types:**
- `builtin` - 内置调度器
- `dolphinscheduler` - DolphinScheduler

### Get Task

获取任务详情。

```http
GET /api/v1/tasks/{id}
```

### Update Task

更新任务配置。

```http
PUT /api/v1/tasks/{id}
Content-Type: application/json

{
  "name": "updated-task",
  "config": {
    "batch_size": 2000
  },
  "schedule": {
    "type": "interval",
    "interval": 3600
  }
}
```

### Delete Task

删除任务。

```http
DELETE /api/v1/tasks/{id}
```

### Start Task

启动任务。

```http
POST /api/v1/tasks/{id}/start
```

### Stop Task

停止任务。

```http
POST /api/v1/tasks/{id}/stop
```

### Pause Task

暂停任务。

```http
POST /api/v1/tasks/{id}/pause
```

### Resume Task

恢复任务。

```http
POST /api/v1/tasks/{id}/resume
```

### Execute Task Now

立即执行任务。

```http
POST /api/v1/tasks/{id}/execute
```

**Response:**
```json
{
  "execution_id": "exec-123",
  "task_id": "task-123",
  "status": "pending",
  "start_time": "2024-01-01T00:00:00Z"
}
```

### Get Task Executions

获取任务执行历史。

```http
GET /api/v1/tasks/{id}/executions
```

**Query Parameters:**
| Parameter | Type | Description |
|-----------|------|-------------|
| page | int | 页码 |
| page_size | int | 每页数量 |
| status | string | 执行状态过滤 |
| start_time | string | 开始时间过滤 |
| end_time | string | 结束时间过滤 |

---

## Templates API

### List Templates

获取配置模板列表。

```http
GET /api/v1/templates
```

**Query Parameters:**
| Parameter | Type | Description |
|-----------|------|-------------|
| type | string | 数据源类型过滤 |

### Create Template

创建配置模板。

```http
POST /api/v1/templates
Content-Type: application/json

{
  "name": "mysql-standard",
  "type": "mysql",
  "description": "Standard MySQL configuration",
  "config_template": {
    "port": 3306,
    "timeout": 30,
    "max_conns": 10,
    "max_idle_conns": 5,
    "charset": "utf8mb4"
  }
}
```

### Get Template

获取模板详情。

```http
GET /api/v1/templates/{id}
```

### Update Template

更新模板。

```http
PUT /api/v1/templates/{id}
```

### Delete Template

删除模板。

```http
DELETE /api/v1/templates/{id}
```

### Apply Template

应用模板创建数据源。

```http
POST /api/v1/templates/{id}/apply
Content-Type: application/json

{
  "name": "new-datasource",
  "config_overrides": {
    "host": "production-host",
    "database": "production_db"
  }
}
```

---

## Error Responses

所有错误响应遵循统一格式：

```json
{
  "code": "ERROR_CODE",
  "message": "Human readable error message",
  "details": "Additional error details",
  "request_id": "req-123"
}
```

**Common Error Codes:**
| Code | HTTP Status | Description |
|------|-------------|-------------|
| INVALID_REQUEST | 400 | 请求参数无效 |
| UNAUTHORIZED | 401 | 未授权 |
| FORBIDDEN | 403 | 权限不足 |
| NOT_FOUND | 404 | 资源不存在 |
| CONFLICT | 409 | 资源冲突 |
| RATE_LIMITED | 429 | 请求频率超限 |
| INTERNAL_ERROR | 500 | 服务器内部错误 |

---

## Rate Limiting

API 请求受到频率限制：

- 默认限制：每 IP 每秒 100 次请求
- 超限响应：HTTP 429 Too Many Requests

响应头包含限流信息：
```
X-RateLimit-Limit: 100
X-RateLimit-Remaining: 99
X-RateLimit-Reset: 1609459200
```

---

## Pagination

列表接口支持分页：

**Request:**
```
GET /api/v1/datasources?page=1&page_size=20
```

**Response:**
```json
{
  "data": [...],
  "total": 100,
  "page": 1,
  "page_size": 20
}
```

---

## Webhooks (Coming Soon)

支持配置 Webhook 接收事件通知：

- 数据源状态变更
- 任务执行完成
- 连接异常告警
