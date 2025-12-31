#!/bin/bash
# 元数据管理系统部署脚本
# Metadata Management System Deployment Script

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# 默认配置
VERSION=${VERSION:-"latest"}
BUILD_TIME=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
GIT_COMMIT=$(git rev-parse --short HEAD 2>/dev/null || echo "unknown")
DOCKER_COMPOSE_FILE="deployments/docker/docker-compose.yaml"

# 打印带颜色的消息
print_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

print_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# 显示帮助信息
show_help() {
    echo "Usage: $0 [command] [options]"
    echo ""
    echo "Commands:"
    echo "  build       Build Docker images"
    echo "  up          Start all services"
    echo "  down        Stop all services"
    echo "  restart     Restart all services"
    echo "  logs        Show service logs"
    echo "  status      Show service status"
    echo "  clean       Clean up containers and volumes"
    echo "  migrate     Run database migrations"
    echo "  test        Run integration tests"
    echo "  help        Show this help message"
    echo ""
    echo "Options:"
    echo "  --profile   Docker Compose profile (e.g., neo4j, postgres, cache)"
    echo "  --version   Version tag for Docker images"
    echo ""
    echo "Examples:"
    echo "  $0 build --version v1.0.0"
    echo "  $0 up --profile cache"
    echo "  $0 logs metadata-server"
}

# 构建Docker镜像
build_images() {
    print_info "Building Docker images..."
    print_info "Version: $VERSION"
    print_info "Build Time: $BUILD_TIME"
    print_info "Git Commit: $GIT_COMMIT"
    
    docker build \
        --build-arg VERSION="$VERSION" \
        --build-arg BUILD_TIME="$BUILD_TIME" \
        --build-arg GIT_COMMIT="$GIT_COMMIT" \
        -t go-metadata:"$VERSION" \
        -f deployments/docker/Dockerfile \
        .
    
    if [ "$VERSION" != "latest" ]; then
        docker tag go-metadata:"$VERSION" go-metadata:latest
    fi
    
    print_info "Docker images built successfully"
}

# 启动服务
start_services() {
    local profile=$1
    print_info "Starting services..."
    
    if [ -n "$profile" ]; then
        print_info "Using profile: $profile"
        docker-compose -f "$DOCKER_COMPOSE_FILE" --profile "$profile" up -d
    else
        docker-compose -f "$DOCKER_COMPOSE_FILE" up -d
    fi
    
    print_info "Services started successfully"
    print_info "Waiting for services to be healthy..."
    sleep 10
    show_status
}

# 停止服务
stop_services() {
    print_info "Stopping services..."
    docker-compose -f "$DOCKER_COMPOSE_FILE" down
    print_info "Services stopped successfully"
}

# 重启服务
restart_services() {
    print_info "Restarting services..."
    stop_services
    start_services "$1"
}

# 显示日志
show_logs() {
    local service=$1
    if [ -n "$service" ]; then
        docker-compose -f "$DOCKER_COMPOSE_FILE" logs -f "$service"
    else
        docker-compose -f "$DOCKER_COMPOSE_FILE" logs -f
    fi
}

# 显示服务状态
show_status() {
    print_info "Service Status:"
    docker-compose -f "$DOCKER_COMPOSE_FILE" ps
    echo ""
    print_info "Health Check:"
    docker-compose -f "$DOCKER_COMPOSE_FILE" ps --format "table {{.Name}}\t{{.Status}}"
}

# 清理资源
clean_up() {
    print_warn "This will remove all containers and volumes. Are you sure? (y/N)"
    read -r confirm
    if [ "$confirm" = "y" ] || [ "$confirm" = "Y" ]; then
        print_info "Cleaning up..."
        docker-compose -f "$DOCKER_COMPOSE_FILE" down -v --remove-orphans
        docker system prune -f
        print_info "Cleanup completed"
    else
        print_info "Cleanup cancelled"
    fi
}

# 运行数据库迁移
run_migrations() {
    print_info "Running database migrations..."
    
    # 等待数据库就绪
    print_info "Waiting for database to be ready..."
    sleep 5
    
    # 执行迁移脚本
    for migration in migrations/*.sql; do
        if [ -f "$migration" ]; then
            print_info "Applying migration: $migration"
            docker-compose -f "$DOCKER_COMPOSE_FILE" exec -T mysql \
                mysql -u metadata -pmetadata123 metadata < "$migration" || true
        fi
    done
    
    print_info "Migrations completed"
}

# 运行集成测试
run_tests() {
    print_info "Running integration tests..."
    go test -v -tags=integration ./test/integration/...
    print_info "Tests completed"
}

# 主函数
main() {
    local command=$1
    shift || true
    
    local profile=""
    local version=""
    
    # 解析参数
    while [ $# -gt 0 ]; do
        case $1 in
            --profile)
                profile=$2
                shift 2
                ;;
            --version)
                version=$2
                VERSION=$version
                shift 2
                ;;
            *)
                break
                ;;
        esac
    done
    
    case $command in
        build)
            build_images
            ;;
        up)
            start_services "$profile"
            ;;
        down)
            stop_services
            ;;
        restart)
            restart_services "$profile"
            ;;
        logs)
            show_logs "$1"
            ;;
        status)
            show_status
            ;;
        clean)
            clean_up
            ;;
        migrate)
            run_migrations
            ;;
        test)
            run_tests
            ;;
        help|--help|-h)
            show_help
            ;;
        *)
            print_error "Unknown command: $command"
            show_help
            exit 1
            ;;
    esac
}

# 执行主函数
main "$@"
