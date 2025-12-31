# Go Metadata

基于 [Kratos](https://go-kratos.dev/) 微服务框架的元数据管理系统，用于管理数据资产的元数据信息，支持 SQL 血缘解析、多数据源元数据采集和图数据库存储。

## 功能特性

- **SQL 血缘解析**: 从 SQL 语句中提取列级别的数据血缘关系
- **多数据源采集**: 支持 MySQL、PostgreSQL、Hive、ClickHouse 等数据源的元数据采集
- **图数据库存储**: 支持 NebulaGraph、Neo4j 存储血缘关系图
- **多 SQL 方言**: 支持 Flink SQL、Spark SQL、Hive、MySQL、PostgreSQL 等
- **RESTful + gRPC**: 同时支持 HTTP 和 gRPC 协议
- **参数校验**: 基于 protoc-gen-validate 的请求参数校验
- **认证授权**: JWT 认证 + RBAC 权限控制

## 技术栈

- **框架**: [Kratos v2](https://go-kratos.dev/) - 微服务框架
- **API**: Protocol Buffers + gRPC + HTTP
- **依赖注入**: [Wire](https://github.com/google/wire)
- **数据库**: MySQL / PostgreSQL + GORM
- **缓存**: Redis
- **配置**: YAML + Protobuf

## 快速开始

### 环境要求

- Go 1.21+
- Protocol Buffers 编译器 (protoc)
- Docker (可选，用于运行依赖服务)

### 安装开发工具

```bash
make init
```

### 生成代码

```bash
# 生成所有 Proto 代码
make proto

# 生成 Wire 依赖注入
make wire
```

### 构建运行

```bash
# 构建
make build

# 运行
make run-server

# 或直接运行
./build/server -conf ./configs
```

### 使用 Docker

```bash
# 启动所有服务
make docker-compose-up

# 停止服务
make docker-compose-down
```

## 项目结构

```
go-metadata/
├── api/                        # API 定义（Protobuf）
│   ├── errors/                 # 错误码定义
│   └── metadata/v1/            # API Proto 文件（proto 和生成代码放一起）
├── cmd/                        # 可执行程序入口
│   ├── server/                 # API 服务
│   └── cli/                    # CLI 工具
├── configs/                    # 配置文件
├── internal/                   # 私有代码
│   ├── biz/                    # 业务逻辑层
│   ├── conf/                   # 配置结构
│   ├── data/                   # 数据访问层
│   ├── server/                 # HTTP/gRPC 服务器
│   ├── service/                # 服务实现层
│   ├── auth/                   # 认证授权
│   ├── collector/              # 元数据采集器
│   └── lineage/                # 血缘解析
├── pkg/                        # 公共库
├── third_party/                # 第三方 Proto 依赖
├── deployments/                # 部署配置
├── migrations/                 # 数据库迁移
├── docs/                       # 文档
└── web/                        # 前端代码
```

## API 文档

启动服务后，可通过以下方式查看 API 文档：

- OpenAPI 文档: `openapi.yaml`
- 健康检查: `GET /health`

### 主要 API

| 服务 | 路径 | 说明 |
|------|------|------|
| 数据源管理 | `/api/v1/datasources` | 数据源 CRUD、连接测试 |
| 任务管理 | `/api/v1/tasks` | 采集任务管理、执行控制 |
| 模板管理 | `/api/v1/templates` | 数据源模板管理 |

## 开发指南

### 新增 API

1. 在 `api/metadata/v1/` 定义 Proto 文件（包含 HTTP 注解和参数校验）
2. 运行 `make proto-api` 生成代码
3. 运行 `make proto-server` 生成 Service 骨架
4. 实现业务逻辑
5. 运行 `make wire` 更新依赖注入

### 常用命令

```bash
make init           # 初始化开发环境
make proto          # 生成所有 Proto 代码
make proto-api      # 生成 API 代码
make proto-errors   # 生成错误码
make wire           # 生成 Wire 依赖
make build          # 构建项目
make test           # 运行测试
make lint           # 代码检查
```

详细开发规范请参考 [项目开发规范](docs/project-specification.md)。

## 配置说明

配置文件位于 `configs/config.yaml`，主要配置项：

```yaml
server:
  http:
    addr: 0.0.0.0:8000
    enabled: true
  grpc:
    addr: 0.0.0.0:9000
    enabled: true

data:
  database:
    driver: mysql
    source: root:password@tcp(127.0.0.1:3306)/metadata

auth:
  jwt_secret: your-secret-key
  jwt_expire: 24h
```

## 文档

- [项目开发规范](docs/project-specification.md)
- [架构设计](docs/architecture.md)
- [API 文档](docs/api.md)
- [部署指南](docs/deployment.md)

## License

MIT
