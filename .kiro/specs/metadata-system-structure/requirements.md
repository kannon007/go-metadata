# Requirements Document

## Introduction

本文档定义了元数据管理系统的目录结构设计需求。系统遵循 Go 语言标准项目布局（golang-standards/project-layout），包含两个基础技术组件：血缘解析（Lineage）和元数据采集（Collector）。

## Glossary

- **Metadata_System**: 元数据管理系统，负责管理数据资产的元数据信息
- **Lineage_Component**: 血缘解析组件，从 SQL 语句中提取列级别的数据血缘关系
- **Collector_Component**: 元数据采集组件，从各类数据源采集表结构、字段等元数据
- **Graph_Component**: 图数据库组件，负责存储和查询元数据血缘关系图
- **Internal_Package**: internal 目录，存放私有应用和库代码，不可被外部导入
- **Pkg_Package**: pkg 目录，存放可被外部项目导入的公共库代码
- **Cmd_Directory**: cmd 目录，存放项目的主应用入口

## Requirements

### Requirement 1: 遵循 Go 标准项目布局

**User Story:** 作为 Go 开发者，我希望项目遵循 Go 社区标准布局，以便快速理解项目结构并参与开发。

#### Acceptance Criteria

1. THE Metadata_System SHALL 将可执行程序入口放置在 `cmd/` 目录下，每个应用一个子目录
2. THE Metadata_System SHALL 将私有代码放置在 `internal/` 目录下
3. THE Metadata_System SHALL 将可导出的公共库放置在 `pkg/` 目录下
4. THE Metadata_System SHALL 将 API 定义文件（OpenAPI/Swagger、protobuf）放置在 `api/` 目录下

### Requirement 2: 基础技术组件组织

**User Story:** 作为开发者，我希望血缘解析和元数据采集作为独立的内部包组织，以便模块化开发和测试。

#### Acceptance Criteria

1. THE Lineage_Component SHALL 放置在 `internal/lineage/` 目录下作为内部包
2. THE Collector_Component SHALL 放置在 `internal/collector/` 目录下作为内部包
3. THE Internal_Package SHALL 保持各组件之间的低耦合，组件可独立测试
4. WHEN 组件需要被外部项目使用时，THE Metadata_System SHALL 将其移至 `pkg/` 目录

### Requirement 3: 元数据采集组件结构

**User Story:** 作为开发者，我希望元数据采集组件支持多种数据源，以便从不同类型的数据库采集元数据。

#### Acceptance Criteria

1. THE Collector_Component SHALL 在根目录定义统一的采集器接口（collector.go）
2. THE Collector_Component SHALL 为每种数据源创建独立子包（mysql/、postgres/、hive/）
3. THE Collector_Component SHALL 提供统一的元数据结构定义（types.go）
4. WHEN 采集元数据时，THE Collector_Component SHALL 返回统一的 TableSchema 结构

### Requirement 4: 元数据管理业务代码组织

**User Story:** 作为开发者，我希望元数据管理系统的业务代码有清晰的组织结构，以便开发和维护业务功能。

#### Acceptance Criteria

1. THE Metadata_System SHALL 将业务服务代码放置在 `internal/service/` 目录下
2. THE Internal_Package SHALL 包含 `internal/model/` 存放业务数据模型
3. THE Internal_Package SHALL 包含 `internal/repository/` 存放数据访问层代码
4. THE Internal_Package SHALL 包含 `internal/handler/` 存放 HTTP/gRPC 处理器
5. WHEN 需要存储元数据时，THE Metadata_System SHALL 在 `internal/store/` 中实现存储逻辑

### Requirement 5: 图数据库组件组织

**User Story:** 作为开发者，我希望图数据库组件支持多种图数据库后端，以便存储和查询元数据血缘关系图。

#### Acceptance Criteria

1. THE Metadata_System SHALL 将图数据库组件放置在 `internal/graph/` 目录下
2. THE Internal_Package SHALL 在 `internal/graph/` 根目录定义统一的图数据库接口（graph.go）
3. THE Internal_Package SHALL 在 `internal/graph/nebula/` 实现 NebulaGraph 适配器
4. THE Internal_Package SHALL 在 `internal/graph/neo4j/` 实现 Neo4j 适配器
5. WHEN 切换图数据库后端时，THE Metadata_System SHALL 通过配置选择具体实现，无需修改业务代码

### Requirement 6: 应用入口组织

**User Story:** 作为开发者，我希望不同的可执行程序有独立的入口，以便分别构建和部署。

#### Acceptance Criteria

1. THE Cmd_Directory SHALL 为 API 服务创建 `cmd/server/` 入口
2. THE Cmd_Directory SHALL 为 CLI 工具创建 `cmd/cli/` 入口
3. WHEN 添加新的可执行程序时，THE Cmd_Directory SHALL 创建独立子目录
4. THE Cmd_Directory 中的 main.go SHALL 保持精简，主要负责依赖注入和启动

### Requirement 7: 配置与脚本组织

**User Story:** 作为运维人员，我希望配置和脚本有标准的存放位置，以便部署和运维。

#### Acceptance Criteria

1. THE Metadata_System SHALL 将配置文件模板放置在 `configs/` 目录下
2. THE Metadata_System SHALL 将构建、安装、分析等脚本放置在 `scripts/` 目录下
3. THE Metadata_System SHALL 将部署配置（Docker、K8s）放置在 `deployments/` 目录下
4. IF 需要数据库迁移，THEN THE Metadata_System SHALL 将迁移文件放置在 `migrations/` 目录下

### Requirement 8: 测试组织

**User Story:** 作为开发者，我希望测试代码有清晰的组织方式，以便维护和运行测试。

#### Acceptance Criteria

1. THE Metadata_System SHALL 将单元测试文件（_test.go）与源文件放在同一目录
2. THE Metadata_System SHALL 将集成测试和端到端测试放置在 `test/` 目录下
3. THE Metadata_System SHALL 将测试数据和测试工具放置在 `testdata/` 子目录
4. WHEN 编写表驱动测试时，THE Metadata_System SHALL 遵循 Go 测试惯例

### Requirement 9: 文档组织

**User Story:** 作为开发者，我希望项目文档有统一的存放位置，以便查阅和维护。

#### Acceptance Criteria

1. THE Metadata_System SHALL 将设计文档和用户文档放置在 `docs/` 目录下
2. THE Metadata_System SHALL 在项目根目录保留 README.md 作为项目入口文档
3. THE Metadata_System SHALL 将示例代码放置在 `examples/` 目录下
4. IF 有第三方工具配置，THEN THE Metadata_System SHALL 放置在 `tools/` 目录下
