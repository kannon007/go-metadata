# Implementation Plan: 元数据管理系统目录结构

## Overview

本任务列表将现有项目从 `core/` 结构迁移到 Go 标准项目布局，并创建新的组件骨架代码。

## Tasks

- [x] 1. 创建基础目录结构
  - 创建 `cmd/`、`internal/`、`pkg/`、`api/`、`configs/`、`scripts/`、`deployments/`、`migrations/`、`test/`、`docs/`、`examples/` 目录
  - 创建 `Makefile` 构建文件
  - _Requirements: 1.1, 1.2, 1.3, 1.4, 7.1, 7.2, 7.3, 9.1, 9.3_

- [x] 2. 迁移血缘解析组件
  - [x] 2.1 将 `core/lineage/` 迁移到 `internal/lineage/`
    - 移动所有文件并更新 import 路径
    - _Requirements: 2.1_
  - [x] 2.2 更新 go.mod 中的模块引用
    - 确保所有 import 路径正确
    - _Requirements: 2.1_

- [x] 3. 创建元数据采集组件骨架
  - [x] 3.1 创建采集器接口定义
    - 创建 `internal/collector/collector.go` 定义 `Collector` 接口
    - 创建 `internal/collector/types.go` 定义 `TableSchema`、`Column` 等类型
    - _Requirements: 3.1, 3.3_
  - [x] 3.2 创建 MySQL 采集器骨架
    - 创建 `internal/collector/mysql/mysql.go`
    - _Requirements: 3.2_
  - [x] 3.3 创建 PostgreSQL 采集器骨架
    - 创建 `internal/collector/postgres/postgres.go`
    - _Requirements: 3.2_
  - [x] 3.4 创建 Hive 采集器骨架
    - 创建 `internal/collector/hive/hive.go`
    - _Requirements: 3.2_

- [x] 4. 创建图数据库组件骨架
  - [x] 4.1 创建图数据库接口定义
    - 创建 `internal/graph/graph.go` 定义 `GraphDB` 接口
    - 创建 `internal/graph/types.go` 定义 `Node`、`Edge`、`LineageGraph` 等类型
    - _Requirements: 5.1, 5.2_
  - [x] 4.2 创建 NebulaGraph 适配器骨架
    - 创建 `internal/graph/nebula/nebula.go`
    - _Requirements: 5.3_
  - [x] 4.3 创建 Neo4j 适配器骨架
    - 创建 `internal/graph/neo4j/neo4j.go`
    - _Requirements: 5.4_

- [x] 5. 创建业务层骨架
  - [x] 5.1 创建业务数据模型
    - 创建 `internal/model/datasource.go`
    - 创建 `internal/model/metadata.go`
    - _Requirements: 4.2_
  - [x] 5.2 创建业务服务层骨架
    - 创建 `internal/service/metadata/service.go`
    - 创建 `internal/service/lineage/service.go`
    - _Requirements: 4.1_
  - [x] 5.3 创建数据访问层骨架
    - 创建 `internal/repository/repository.go`
    - _Requirements: 4.3_
  - [x] 5.4 创建存储层骨架
    - 创建 `internal/store/store.go`
    - _Requirements: 4.5_

- [x] 6. 创建应用入口
  - [x] 6.1 创建 API 服务入口
    - 创建 `cmd/server/main.go`
    - _Requirements: 6.1_
  - [x] 6.2 创建 CLI 工具入口
    - 创建 `cmd/cli/main.go`
    - _Requirements: 6.2_

- [x] 7. 创建公共库骨架
  - [x] 7.1 创建工具函数包
    - 创建 `pkg/utils/utils.go`
    - _Requirements: 1.3_
  - [x] 7.2 创建错误定义包
    - 创建 `pkg/errors/errors.go`
    - 创建 `internal/collector/errors.go`
    - 创建 `internal/graph/errors.go`
    - _Requirements: 1.3_

- [x] 8. 创建配置和部署文件
  - [x] 8.1 创建配置文件模板
    - 创建 `configs/config.yaml.example`
    - _Requirements: 7.1_
  - [x] 8.2 创建构建脚本
    - 创建 `scripts/build.sh`
    - 创建 `scripts/test.sh`
    - _Requirements: 7.2_
  - [x] 8.3 创建 Docker 配置
    - 创建 `deployments/docker/Dockerfile`
    - 创建 `deployments/docker/docker-compose.yaml`
    - _Requirements: 7.3_

- [x] 9. Checkpoint - 验证目录结构
  - 确保所有目录和文件已创建
  - 运行 `go build ./...` 验证编译通过
  - 确保所有 import 路径正确

- [x] 10. 更新文档
  - [x] 10.1 更新 README.md
    - 更新项目结构说明
    - 更新快速开始指南
    - _Requirements: 9.2_
  - [x] 10.2 创建架构文档
    - 创建 `docs/architecture.md`
    - _Requirements: 9.1_

- [x] 11. 清理旧目录
  - 删除 `core/` 目录（确认迁移完成后）
  - 删除旧的 `main.go`
  - _Requirements: 2.1_

- [x] 12. Final Checkpoint - 确保所有测试通过
  - 运行 `go test ./...` 确保所有测试通过
  - 确保项目可以正常编译和运行

## Notes

- 任务按顺序执行，确保依赖关系正确
- 迁移血缘解析组件时需要更新所有 import 路径
- 骨架代码只包含接口定义和基本结构，具体实现在后续任务中完成
- 清理旧目录前需确认所有功能正常
