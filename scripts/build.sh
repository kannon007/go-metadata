#!/bin/bash
# 构建脚本 / Build Script
# 用于构建元数据管理系统的可执行文件

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# 项目配置
MODULE="go-metadata"
BUILD_DIR="build"
VERSION=${VERSION:-"dev"}
BUILD_TIME=$(date -u '+%Y-%m-%d_%H:%M:%S')
GIT_COMMIT=${GIT_COMMIT:-$(git rev-parse --short HEAD 2>/dev/null || echo "unknown")}

# 构建标志
LDFLAGS="-X main.Version=${VERSION} -X main.BuildTime=${BUILD_TIME} -X main.GitCommit=${GIT_COMMIT}"

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
    echo "Usage: $0 [OPTIONS] [TARGET]"
    echo ""
    echo "Options:"
    echo "  -h, --help      显示帮助信息"
    echo "  -v, --version   设置版本号 (默认: dev)"
    echo "  -o, --output    设置输出目录 (默认: build)"
    echo "  --os            目标操作系统 (linux, darwin, windows)"
    echo "  --arch          目标架构 (amd64, arm64)"
    echo ""
    echo "Targets:"
    echo "  all             构建所有组件 (默认)"
    echo "  server          仅构建 API 服务器"
    echo "  cli             仅构建 CLI 工具"
    echo "  clean           清理构建产物"
    echo ""
    echo "Examples:"
    echo "  $0                          # 构建所有组件"
    echo "  $0 server                   # 仅构建服务器"
    echo "  $0 --os linux --arch amd64  # 交叉编译"
    echo "  $0 -v 1.0.0 all             # 指定版本号构建"
}

# 检查 Go 环境
check_go() {
    if ! command -v go &> /dev/null; then
        print_error "Go is not installed. Please install Go first."
        exit 1
    fi
    print_info "Go version: $(go version)"
}

# 清理构建目录
clean() {
    print_info "Cleaning build directory..."
    rm -rf "${BUILD_DIR}"
    go clean
    print_info "Clean completed."
}

# 下载依赖
download_deps() {
    print_info "Downloading dependencies..."
    go mod download
    go mod tidy
}

# 构建服务器
build_server() {
    local os=${1:-$(go env GOOS)}
    local arch=${2:-$(go env GOARCH)}
    local output="${BUILD_DIR}/server"
    
    if [ "$os" = "windows" ]; then
        output="${output}.exe"
    fi
    
    print_info "Building server for ${os}/${arch}..."
    
    mkdir -p "${BUILD_DIR}"
    
    GOOS=$os GOARCH=$arch go build \
        -ldflags "${LDFLAGS}" \
        -o "${output}" \
        ./cmd/server
    
    print_info "Server built: ${output}"
}

# 构建 CLI
build_cli() {
    local os=${1:-$(go env GOOS)}
    local arch=${2:-$(go env GOARCH)}
    local output="${BUILD_DIR}/cli"
    
    if [ "$os" = "windows" ]; then
        output="${output}.exe"
    fi
    
    print_info "Building CLI for ${os}/${arch}..."
    
    mkdir -p "${BUILD_DIR}"
    
    GOOS=$os GOARCH=$arch go build \
        -ldflags "${LDFLAGS}" \
        -o "${output}" \
        ./cmd/cli
    
    print_info "CLI built: ${output}"
}

# 构建所有组件
build_all() {
    local os=${1:-$(go env GOOS)}
    local arch=${2:-$(go env GOARCH)}
    
    print_info "Building all components for ${os}/${arch}..."
    build_server "$os" "$arch"
    build_cli "$os" "$arch"
    print_info "All components built successfully!"
}

# 解析命令行参数
TARGET_OS=""
TARGET_ARCH=""
TARGET="all"

while [[ $# -gt 0 ]]; do
    case $1 in
        -h|--help)
            show_help
            exit 0
            ;;
        -v|--version)
            VERSION="$2"
            shift 2
            ;;
        -o|--output)
            BUILD_DIR="$2"
            shift 2
            ;;
        --os)
            TARGET_OS="$2"
            shift 2
            ;;
        --arch)
            TARGET_ARCH="$2"
            shift 2
            ;;
        all|server|cli|clean)
            TARGET="$1"
            shift
            ;;
        *)
            print_error "Unknown option: $1"
            show_help
            exit 1
            ;;
    esac
done

# 主流程
main() {
    print_info "=== Go Metadata Build Script ==="
    print_info "Version: ${VERSION}"
    print_info "Build Time: ${BUILD_TIME}"
    print_info "Git Commit: ${GIT_COMMIT}"
    echo ""
    
    check_go
    
    case $TARGET in
        clean)
            clean
            ;;
        server)
            download_deps
            build_server "$TARGET_OS" "$TARGET_ARCH"
            ;;
        cli)
            download_deps
            build_cli "$TARGET_OS" "$TARGET_ARCH"
            ;;
        all)
            download_deps
            build_all "$TARGET_OS" "$TARGET_ARCH"
            ;;
    esac
    
    print_info "=== Build completed ==="
}

main
