# Go Metadata Management System Makefile
# 基于 Kratos 框架的元数据管理系统

# Variables
MODULE := go-metadata
BINARY_SERVER := server
BINARY_CLI := cli
BUILD_DIR := build
GO := go
GOFLAGS := -v
VERSION := $(shell git describe --tags --always 2>/dev/null || echo "v0.0.1")

# Proto 文件
API_PROTO_FILES := $(wildcard api/metadata/v1/*.proto)
INTERNAL_PROTO_FILES := $(wildcard internal/conf/*.proto)
ERROR_PROTO_FILES := $(wildcard api/errors/*.proto)

# Build targets
.PHONY: all build build-server build-cli clean test lint fmt help
.PHONY: init wire generate proto proto-conf proto-api proto-errors proto-server

all: proto generate build

## init: 初始化开发环境，安装必要工具
init:
	@echo "Installing Kratos CLI..."
	$(GO) install github.com/go-kratos/kratos/cmd/kratos/v2@latest
	@echo "Installing protoc plugins..."
	$(GO) install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	$(GO) install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
	$(GO) install github.com/go-kratos/kratos/cmd/protoc-gen-go-http/v2@latest
	$(GO) install github.com/go-kratos/kratos/cmd/protoc-gen-go-errors/v2@latest
	$(GO) install github.com/google/gnostic/cmd/protoc-gen-openapi@latest
	$(GO) install github.com/envoyproxy/protoc-gen-validate@latest
	@echo "Installing Wire..."
	$(GO) install github.com/google/wire/cmd/wire@latest
	@echo "Tidying modules..."
	$(GO) mod tidy
	@echo "Done! All tools installed."

## wire: 生成 Wire 依赖注入代码
wire:
	@echo "Generating wire dependencies..."
	cd cmd/server && wire

## generate: 运行 go generate
generate:
	@echo "Running go generate..."
	$(GO) mod tidy
	$(GO) generate ./...

## proto: 生成所有 protobuf 代码
proto: proto-api proto-conf proto-errors
	@echo "All proto files generated."

## proto-api: 生成 API protobuf 代码（HTTP + gRPC + OpenAPI）
proto-api:
	@echo "Generating API protobuf..."
	protoc --proto_path=./api/metadata/v1 \
		--proto_path=./third_party \
		--go_out=paths=source_relative:./api/metadata/v1 \
		--go-http_out=paths=source_relative:./api/metadata/v1 \
		--go-grpc_out=paths=source_relative:./api/metadata/v1 \
		api/metadata/v1/datasource.proto api/metadata/v1/task.proto
	protoc --proto_path=./api/metadata/v1 \
		--proto_path=./third_party \
		--go_out=paths=source_relative:./api/metadata/v1 \
		--go-http_out=paths=source_relative:./api/metadata/v1 \
		--go-grpc_out=paths=source_relative:./api/metadata/v1 \
		api/metadata/v1/template.proto
	@echo "Generating OpenAPI spec..."
	protoc --proto_path=./api/metadata/v1 \
		--proto_path=./third_party \
		--openapi_out=fq_schema_naming=true,default_response=false:. \
		api/metadata/v1/datasource.proto api/metadata/v1/task.proto api/metadata/v1/template.proto

## proto-conf: 生成配置 protobuf 代码
proto-conf:
	@echo "Generating config protobuf..."
	protoc --proto_path=./internal/conf \
		--proto_path=./third_party \
		--go_out=paths=source_relative:./internal/conf \
		$(INTERNAL_PROTO_FILES)

## proto-errors: 生成错误码 protobuf 代码
proto-errors:
	@echo "Generating errors protobuf..."
	protoc --proto_path=./api/errors \
		--proto_path=./third_party \
		--go_out=paths=source_relative:./api/errors \
		--go-errors_out=paths=source_relative:./api/errors \
		$(ERROR_PROTO_FILES)

## proto-server: 从 proto 生成 service 实现骨架
proto-server:
	@echo "Generating service implementations..."
	kratos proto server api/metadata/v1/datasource.proto -t internal/service
	kratos proto server api/metadata/v1/task.proto -t internal/service
	kratos proto server api/metadata/v1/template.proto -t internal/service

## build: 构建所有二进制文件
build: wire build-server build-cli

## build-server: 构建 API 服务
build-server:
	@echo "Building server..."
	@mkdir -p $(BUILD_DIR)
	$(GO) build $(GOFLAGS) -ldflags "-X main.Version=$(VERSION)" -o $(BUILD_DIR)/$(BINARY_SERVER) ./cmd/server

## build-cli: 构建 CLI 工具
build-cli:
	@echo "Building CLI..."
	@mkdir -p $(BUILD_DIR)
	$(GO) build $(GOFLAGS) -ldflags "-X main.Version=$(VERSION)" -o $(BUILD_DIR)/$(BINARY_CLI) ./cmd/cli

## clean: 清理构建产物
clean:
	@echo "Cleaning..."
	@rm -rf $(BUILD_DIR)
	@rm -f server.exe
	$(GO) clean

## test: 运行所有测试
test:
	@echo "Running tests..."
	$(GO) test -v ./...

## test-coverage: 运行测试并生成覆盖率报告
test-coverage:
	@echo "Running tests with coverage..."
	$(GO) test -v -coverprofile=coverage.out ./...
	$(GO) tool cover -html=coverage.out -o coverage.html

## lint: 运行代码检查
lint:
	@echo "Running linter..."
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run ./...; \
	else \
		echo "golangci-lint not installed, skipping..."; \
	fi

## fmt: 格式化代码
fmt:
	@echo "Formatting code..."
	$(GO) fmt ./...

## mod-tidy: 整理 go modules
mod-tidy:
	@echo "Tidying modules..."
	$(GO) mod tidy

## mod-download: 下载依赖
mod-download:
	@echo "Downloading dependencies..."
	$(GO) mod download

## run-server: 运行 API 服务
run-server: build-server
	@echo "Running server..."
	./$(BUILD_DIR)/$(BINARY_SERVER) -conf ./configs

## run-cli: 运行 CLI 工具
run-cli: build-cli
	@echo "Running CLI..."
	./$(BUILD_DIR)/$(BINARY_CLI)

## docker-build: 构建 Docker 镜像
docker-build:
	@echo "Building Docker image..."
	docker build -t $(MODULE):$(VERSION) -f deployments/docker/Dockerfile .

## docker-compose-up: 使用 docker-compose 启动服务
docker-compose-up:
	@echo "Starting services..."
	docker-compose -f deployments/docker/docker-compose.yaml up -d

## docker-compose-down: 使用 docker-compose 停止服务
docker-compose-down:
	@echo "Stopping services..."
	docker-compose -f deployments/docker/docker-compose.yaml down

## help: 显示帮助信息
help:
	@echo "Usage: make [target]"
	@echo ""
	@echo "Targets:"
	@sed -n 's/^##//p' $(MAKEFILE_LIST) | column -t -s ':' | sed -e 's/^/ /'

.DEFAULT_GOAL := help
