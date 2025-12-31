# Go Metadata 项目开发规范

基于 [Kratos](https://go-kratos.dev/) 微服务框架的元数据管理系统开发规范文档。

## 目录

- [设计理念](#设计理念)
- [项目结构](#项目结构)
- [开发环境](#开发环境)
- [Proto 文件规范](#proto-文件规范)
- [代码生成](#代码生成)
- [分层架构](#分层架构)
- [错误处理](#错误处理)
- [配置管理](#配置管理)
- [开发流程](#开发流程)
- [常用命令](#常用命令)

---

## 设计理念

本项目遵循 Kratos 框架的核心设计哲学：

- **插件化设计**：框架不绑定特定基础设施，可自由选择注册中心、数据库 ORM、缓存等组件
- **DDD + 简洁架构**：参考领域驱动设计和 Clean Architecture，保持代码的可读性、可测试性和可维护性
- **Protobuf 定义 API**：使用 Protocol Buffers 进行接口定义，同时支持 gRPC 和 HTTP
- **依赖注入**：使用 Wire 进行依赖注入，提高代码的可测试性

---

## 项目结构

```
go-metadata/
├── api/                        # API 定义（Protobuf）
│   ├── errors/                 # 错误码定义（Kratos 错误体系）
│   │   ├── errors.proto        # 业务错误码 proto
│   │   └── errors.pb.go        # 生成的错误码代码
│   └── metadata/v1/            # API Proto 文件（按服务名组织）
│       ├── datasource.proto    # 数据源服务定义
│       ├── datasource.pb.go    # 生成的消息代码
│       ├── datasource_http.pb.go  # 生成的 HTTP 代码
│       ├── datasource_grpc.pb.go  # 生成的 gRPC 代码
│       ├── task.proto          # 任务服务定义
│       └── template.proto      # 模板服务定义
│
├── cmd/                        # 可执行程序入口
│   ├── server/                 # API 服务
│   │   ├── main.go             # 程序入口
│   │   ├── wire.go             # Wire 依赖注入定义
│   │   └── wire_gen.go         # Wire 生成的代码
│   └── cli/                    # CLI 工具
│
├── configs/                    # 配置文件
│   ├── config.yaml             # 运行时配置
│   └── config.yaml.example     # 配置模板
│
├── internal/                   # 私有代码（核心业务逻辑）
│   ├── biz/                    # 业务逻辑层（Domain Layer）
│   ├── conf/                   # 配置结构定义
│   │   └── conf.proto          # 配置 proto 文件
│   ├── data/                   # 数据访问层（Repository）
│   ├── server/                 # HTTP/gRPC 服务器
│   ├── service/                # 服务实现层（Application Layer）
│   ├── auth/                   # 认证授权中间件
│   ├── cache/                  # 缓存层
│   ├── collector/              # 元数据采集器（业务特有）
│   ├── lineage/                # 血缘解析（业务特有）
│   ├── metrics/                # 监控指标
│   └── scheduler/              # 任务调度器
│
├── pkg/                        # 公共库（可被外部引用）
│   ├── errors/                 # 错误工具
│   └── utils/                  # 工具函数
│
├── third_party/                # 第三方 Proto 依赖
│   ├── errors/                 # Kratos 错误定义
│   ├── google/                 # Google API 定义
│   │   ├── api/                # HTTP 注解
│   │   └── protobuf/           # 标准类型
│   └── validate/               # 参数校验
│
├── deployments/                # 部署配置
│   └── docker/                 # Docker 相关
│
├── migrations/                 # 数据库迁移
├── docs/                       # 文档
├── test/                       # 集成测试
├── web/                        # 前端代码
├── Makefile                    # 构建脚本
└── openapi.yaml                # 生成的 OpenAPI 文档
```

---

## 开发环境

### 必需工具

```bash
# 安装 Kratos CLI
go install github.com/go-kratos/kratos/cmd/kratos/v2@latest

# 安装 Protobuf 代码生成工具
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

# 安装 Kratos HTTP 代码生成工具
go install github.com/go-kratos/kratos/cmd/protoc-gen-go-http/v2@latest

# 安装 Kratos 错误码生成工具
go install github.com/go-kratos/kratos/cmd/protoc-gen-go-errors/v2@latest

# 安装参数校验生成工具
go install github.com/envoyproxy/protoc-gen-validate@latest

# 安装 OpenAPI 文档生成工具
go install github.com/google/gnostic/cmd/protoc-gen-openapi@latest

# 安装 Wire 依赖注入工具
go install github.com/google/wire/cmd/wire@latest
```

### 一键初始化

```bash
make init
```

---

## Proto 文件规范

### 目录结构

```
api/
├── metadata/v1/                # API 版本目录（proto 和生成代码放一起）
│   ├── datasource.proto        # 数据源服务
│   ├── datasource.pb.go        # 生成的消息代码
│   ├── datasource_http.pb.go   # 生成的 HTTP 代码
│   ├── datasource_grpc.pb.go   # 生成的 gRPC 代码
│   ├── task.proto              # 任务服务
│   └── template.proto          # 模板服务
└── errors/
    └── errors.proto            # 错误码定义
```

### Proto 文件模板

```protobuf
syntax = "proto3";

package api.v1;

option go_package = "go-metadata/api/metadata/v1;v1";

import "google/api/annotations.proto";
import "google/protobuf/timestamp.proto";
import "google/protobuf/empty.proto";
import "validate/validate.proto";

// 服务定义
service DatasourceService {
  // 创建数据源
  rpc CreateDatasource (CreateDatasourceRequest) returns (Datasource) {
    option (google.api.http) = {
      post: "/api/v1/datasources"
      body: "*"
    };
  }
  
  // 获取数据源
  rpc GetDatasource (GetDatasourceRequest) returns (Datasource) {
    option (google.api.http) = {
      get: "/api/v1/datasources/{id}"
    };
  }
  
  // 列表查询
  rpc ListDatasources (ListDatasourcesRequest) returns (ListDatasourcesResponse) {
    option (google.api.http) = {
      get: "/api/v1/datasources"
    };
  }
  
  // 更新数据源
  rpc UpdateDatasource (UpdateDatasourceRequest) returns (Datasource) {
    option (google.api.http) = {
      put: "/api/v1/datasources/{id}"
      body: "*"
    };
  }
  
  // 删除数据源
  rpc DeleteDatasource (DeleteDatasourceRequest) returns (google.protobuf.Empty) {
    option (google.api.http) = {
      delete: "/api/v1/datasources/{id}"
    };
  }
}

// 请求消息（带参数校验）
message CreateDatasourceRequest {
  string name = 1 [(validate.rules).string = {min_len: 1, max_len: 100}];
  string type = 2 [(validate.rules).string = {min_len: 1}];
  ConnectionConfig config = 3 [(validate.rules).message = {required: true}];
}

// 响应消息
message Datasource {
  string id = 1;
  string name = 2;
  string type = 3;
  // ...
}
```

### 命名规范

| 类型 | 规范 | 示例 |
|------|------|------|
| 服务名 | PascalCase + Service | `DataSourceService` |
| 方法名 | PascalCase，动词开头 | `CreateDataSource`, `GetDataSource` |
| 请求消息 | 方法名 + Request | `CreateDataSourceRequest` |
| 响应消息 | 实体名或方法名 + Response | `DataSource`, `ListDataSourcesResponse` |
| 字段名 | snake_case | `connection_string` |
| 枚举值 | UPPER_SNAKE_CASE | `DATA_SOURCE_TYPE_MYSQL` |

### HTTP 注解规范

| 操作 | HTTP 方法 | 路径模式 | 示例 |
|------|-----------|----------|------|
| 创建 | POST | `/api/v1/{resources}` | `POST /api/v1/datasources` |
| 获取 | GET | `/api/v1/{resources}/{id}` | `GET /api/v1/datasources/{id}` |
| 列表 | GET | `/api/v1/{resources}` | `GET /api/v1/datasources` |
| 更新 | PUT | `/api/v1/{resources}/{id}` | `PUT /api/v1/datasources/{id}` |
| 删除 | DELETE | `/api/v1/{resources}/{id}` | `DELETE /api/v1/datasources/{id}` |
| 操作 | POST | `/api/v1/{resources}/{id}/{action}` | `POST /api/v1/tasks/{id}/start` |

### 参数校验规范

```protobuf
import "validate/validate.proto";

message CreateRequest {
  // 字符串校验
  string name = 1 [(validate.rules).string = {min_len: 1, max_len: 100}];
  
  // 数字校验
  int32 port = 2 [(validate.rules).int32 = {gte: 1, lte: 65535}];
  
  // 枚举校验（必须是已定义的值，且不能是默认值 0）
  DataSourceType type = 3 [(validate.rules).enum = {defined_only: true, not_in: [0]}];
  
  // 必填消息
  ConnectionConfig config = 4 [(validate.rules).message = {required: true}];
  
  // 数组校验
  repeated string ids = 5 [(validate.rules).repeated = {min_items: 1, max_items: 100}];
  
  // 字符串枚举
  string format = 6 [(validate.rules).string = {in: ["json", "yaml", "csv"]}];
}
```

---

## 代码生成

### 生成所有 Proto 代码

```bash
make proto
```

### 生成 API 代码（HTTP + gRPC + OpenAPI）

```bash
make proto-api
```

生成的文件：
- `*.pb.go` - Protobuf 消息定义
- `*_grpc.pb.go` - gRPC 服务代码
- `*_http.pb.go` - HTTP 服务代码
- `openapi.yaml` - OpenAPI 文档

### 生成配置代码

```bash
make proto-conf
```

### 生成错误码

```bash
make proto-errors
```

生成的文件包含错误创建和断言方法：
```go
// 创建错误
errors.ErrorDataSourceNotFound("datasource %s not found", id)

// 断言错误
if errors.IsDataSourceNotFound(err) {
    // 处理未找到的情况
}
```

### 生成 Service 实现骨架

```bash
make proto-server
```

### 生成 Wire 依赖注入

```bash
make wire
```

---

## 分层架构

遵循 DDD 和 Clean Architecture 的分层设计：

```
┌─────────────────────────────────────────────────────────────┐
│                      Transport Layer                        │
│                   (internal/server)                         │
│              HTTP Server / gRPC Server                      │
│         注册生成的 *_http.pb.go / *_grpc.pb.go              │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────┐
│                    Application Layer                        │
│                   (internal/service)                        │
│           实现 proto 定义的服务接口，编排业务逻辑              │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────┐
│                      Domain Layer                           │
│                     (internal/biz)                          │
│              业务实体、业务规则、领域服务                      │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────┐
│                   Infrastructure Layer                      │
│                    (internal/data)                          │
│           数据访问、外部服务调用、缓存等基础设施               │
└─────────────────────────────────────────────────────────────┘
```

### 各层职责

#### internal/server（传输层）

注册生成的 HTTP/gRPC 服务代码：

```go
// internal/server/http.go
func NewHTTPServer(c *conf.Server, ds *service.DataSourceService, logger log.Logger) *http.Server {
    var opts = []http.ServerOption{
        http.Middleware(
            recovery.Recovery(),
            logging.Server(logger),
            validate.Validator(),  // 参数校验中间件
        ),
    }
    srv := http.NewServer(opts...)
    
    // 注册生成的 HTTP 服务
    v1.RegisterDataSourceServiceHTTPServer(srv, ds)
    v1.RegisterTaskServiceHTTPServer(srv, ts)
    
    return srv
}
```

#### internal/service（应用层）

实现 proto 生成的服务接口：

```go
// internal/service/datasource.go
type DataSourceService struct {
    v1.UnimplementedDataSourceServiceServer
    uc *biz.DataSourceUsecase
}

func (s *DataSourceService) CreateDataSource(ctx context.Context, req *v1.CreateDataSourceRequest) (*v1.DataSource, error) {
    ds, err := s.uc.Create(ctx, &biz.DataSource{
        Name: req.Name,
        Type: req.Type,
    })
    if err != nil {
        return nil, err
    }
    return toProtoDataSource(ds), nil
}
```

#### internal/biz（业务层）

定义业务实体和 Repository 接口：

```go
// internal/biz/datasource.go
type DataSource struct {
    ID   string
    Name string
    Type string
}

// Repository 接口定义（依赖倒置）
type DataSourceRepo interface {
    Create(ctx context.Context, ds *DataSource) (*DataSource, error)
    Get(ctx context.Context, id string) (*DataSource, error)
    List(ctx context.Context, page, size int) ([]*DataSource, int64, error)
}

// Usecase 业务用例
type DataSourceUsecase struct {
    repo DataSourceRepo
    log  *log.Helper
}

func (uc *DataSourceUsecase) Create(ctx context.Context, ds *DataSource) (*DataSource, error) {
    // 业务逻辑处理
    return uc.repo.Create(ctx, ds)
}
```

#### internal/data（数据层）

实现 Repository 接口：

```go
// internal/data/datasource.go
type datasourceRepo struct {
    data *Data
    log  *log.Helper
}

func NewDataSourceRepo(data *Data, logger log.Logger) biz.DataSourceRepo {
    return &datasourceRepo{data: data, log: log.NewHelper(logger)}
}

func (r *datasourceRepo) Create(ctx context.Context, ds *biz.DataSource) (*biz.DataSource, error) {
    // 数据库操作
    return ds, nil
}
```

---

## 错误处理

### 错误码定义（Kratos 标准）

```protobuf
// api/errors/errors.proto
syntax = "proto3";

package api.errors;

import "errors/errors.proto";

option go_package = "go-metadata/api/errors;errors";

enum ErrorReason {
  option (errors.default_code) = 500;  // 默认 HTTP 状态码
  
  DATA_SOURCE_NOT_FOUND = 100 [(errors.code) = 404];
  DATA_SOURCE_ALREADY_EXISTS = 101 [(errors.code) = 409];
  TASK_NOT_FOUND = 200 [(errors.code) = 404];
  UNAUTHORIZED = 500 [(errors.code) = 401];
  VALIDATION_FAILED = 600 [(errors.code) = 400];
}
```

### 错误使用

```go
import apierrors "go-metadata/api/errors"

// 创建错误
func (uc *DataSourceUsecase) Get(ctx context.Context, id string) (*DataSource, error) {
    ds, err := uc.repo.Get(ctx, id)
    if err != nil {
        return nil, apierrors.ErrorDataSourceNotFound("datasource %s not found", id)
    }
    return ds, nil
}

// 错误断言
if apierrors.IsDataSourceNotFound(err) {
    // 处理未找到的情况
}
```

### 错误响应格式

```json
{
  "code": 404,
  "reason": "DATA_SOURCE_NOT_FOUND",
  "message": "datasource abc123 not found",
  "metadata": {}
}
```

---

## 配置管理

### 配置 Proto 定义

```protobuf
// internal/conf/conf.proto
syntax = "proto3";

package kratos.api;

option go_package = "go-metadata/internal/conf;conf";

import "google/protobuf/duration.proto";

message Bootstrap {
  Server server = 1;
  Data data = 2;
  Auth auth = 3;
}

message Server {
  message HTTP {
    string network = 1;
    string addr = 2;
    google.protobuf.Duration timeout = 3;
    bool enabled = 4;
  }
  message GRPC {
    string network = 1;
    string addr = 2;
    google.protobuf.Duration timeout = 3;
    bool enabled = 4;
  }
  HTTP http = 1;
  GRPC grpc = 2;
}

message Data {
  message Database {
    string driver = 1;
    string source = 2;
    int32 max_idle_conns = 3;
    int32 max_open_conns = 4;
    google.protobuf.Duration conn_max_lifetime = 5;
  }
  message Redis {
    string network = 1;
    string addr = 2;
    google.protobuf.Duration read_timeout = 3;
    google.protobuf.Duration write_timeout = 4;
  }
  Database database = 1;
  Redis redis = 2;
}
```

### 配置文件示例

```yaml
# configs/config.yaml
server:
  http:
    addr: 0.0.0.0:8000
    timeout: 10s
    enabled: true
  grpc:
    addr: 0.0.0.0:9000
    timeout: 10s
    enabled: true

data:
  database:
    driver: mysql
    source: root:password@tcp(127.0.0.1:3306)/metadata?parseTime=True
    max_idle_conns: 10
    max_open_conns: 100
    conn_max_lifetime: 3600s
  redis:
    addr: 127.0.0.1:6379
    read_timeout: 200ms
    write_timeout: 200ms

auth:
  jwt_secret: your-secret-key
  jwt_expire: 24h
  refresh_expire: 168h
```

---

## 开发流程

### 新增 API 流程

1. **定义 Proto 文件**
   ```bash
   # 在 api/proto/v1/ 下创建或修改 proto 文件
   # 添加 HTTP 注解和参数校验
   ```

2. **生成代码**
   ```bash
   make proto-api
   ```

3. **生成 Service 骨架**（首次）
   ```bash
   kratos proto server api/proto/v1/xxx.proto -t internal/service
   ```

4. **实现业务逻辑**
   - 在 `internal/biz/` 添加业务实体和用例
   - 在 `internal/data/` 实现数据访问
   - 在 `internal/service/` 完善服务实现

5. **注册服务**
   - 在 `internal/server/http.go` 注册 HTTP 服务
   - 在 `internal/server/grpc.go` 注册 gRPC 服务
   - 更新 Wire ProviderSet

6. **生成 Wire**
   ```bash
   make wire
   ```

7. **测试运行**
   ```bash
   make run-server
   ```

### 新增错误码流程

1. 在 `api/errors/errors.proto` 添加错误枚举
2. 运行 `make proto-errors`
3. 在业务代码中使用生成的错误方法

---

## 常用命令

### Makefile 命令

```bash
# 初始化开发环境
make init

# 生成所有 Proto 代码
make proto

# 生成 API Proto（HTTP + gRPC + OpenAPI）
make proto-api

# 生成配置 Proto
make proto-conf

# 生成错误码
make proto-errors

# 生成 Service 实现骨架
make proto-server

# 生成 Wire 依赖注入
make wire

# 构建项目
make build

# 运行服务
make run-server

# 运行测试
make test

# 代码格式化
make fmt

# 代码检查
make lint

# 清理构建产物
make clean

# Docker 构建
make docker-build

# Docker Compose 启动
make docker-compose-up
```

### Kratos CLI 命令

```bash
# 创建新项目
kratos new project-name

# 添加 Proto 文件
kratos proto add api/proto/v1/xxx.proto

# 生成 Proto 客户端代码（HTTP + gRPC）
kratos proto client api/proto/v1/xxx.proto

# 生成 Service 实现
kratos proto server api/proto/v1/xxx.proto -t internal/service

# 查看版本
kratos -v

# 升级工具
kratos upgrade
```

---

## 参考资料

- [Kratos 官方文档](https://go-kratos.dev/zh-cn/docs/)
- [Kratos 设计理念](https://go-kratos.dev/zh-cn/docs/intro/design/)
- [Kratos CLI 使用](https://go-kratos.dev/zh-cn/docs/getting-started/usage/)
- [Kratos Layout](https://github.com/go-kratos/kratos-layout)
- [Kratos Examples](https://github.com/go-kratos/examples)
- [protoc-gen-validate](https://github.com/envoyproxy/protoc-gen-validate)
