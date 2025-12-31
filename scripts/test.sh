#!/bin/bash
# 测试脚本 / Test Script
# 用于运行元数据管理系统的测试

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 配置
COVERAGE_DIR="coverage"
COVERAGE_FILE="${COVERAGE_DIR}/coverage.out"
COVERAGE_HTML="${COVERAGE_DIR}/coverage.html"
MIN_COVERAGE=${MIN_COVERAGE:-60}

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

print_header() {
    echo -e "${BLUE}=== $1 ===${NC}"
}

# 显示帮助信息
show_help() {
    echo "Usage: $0 [OPTIONS] [TARGET]"
    echo ""
    echo "Options:"
    echo "  -h, --help          显示帮助信息"
    echo "  -v, --verbose       详细输出"
    echo "  -c, --coverage      生成覆盖率报告"
    echo "  --min-coverage N    设置最小覆盖率阈值 (默认: 60)"
    echo "  --race              启用竞态检测"
    echo "  --short             跳过长时间运行的测试"
    echo ""
    echo "Targets:"
    echo "  all                 运行所有测试 (默认)"
    echo "  unit                仅运行单元测试"
    echo "  integration         仅运行集成测试"
    echo "  lineage             仅运行血缘解析测试"
    echo "  collector           仅运行采集器测试"
    echo "  graph               仅运行图数据库测试"
    echo ""
    echo "Examples:"
    echo "  $0                          # 运行所有测试"
    echo "  $0 -c                       # 运行测试并生成覆盖率报告"
    echo "  $0 --race unit              # 运行单元测试并启用竞态检测"
    echo "  $0 --min-coverage 80 -c     # 设置80%最小覆盖率"
}

# 检查 Go 环境
check_go() {
    if ! command -v go &> /dev/null; then
        print_error "Go is not installed. Please install Go first."
        exit 1
    fi
    print_info "Go version: $(go version)"
}

# 清理覆盖率目录
clean_coverage() {
    rm -rf "${COVERAGE_DIR}"
    mkdir -p "${COVERAGE_DIR}"
}

# 运行所有测试
run_all_tests() {
    print_header "Running All Tests"
    
    local flags="-v"
    
    if [ "$RACE" = "true" ]; then
        flags="${flags} -race"
    fi
    
    if [ "$SHORT" = "true" ]; then
        flags="${flags} -short"
    fi
    
    if [ "$COVERAGE" = "true" ]; then
        clean_coverage
        go test ${flags} -coverprofile="${COVERAGE_FILE}" -covermode=atomic ./...
        generate_coverage_report
    else
        go test ${flags} ./...
    fi
}

# 运行单元测试
run_unit_tests() {
    print_header "Running Unit Tests"
    
    local flags="-v"
    
    if [ "$RACE" = "true" ]; then
        flags="${flags} -race"
    fi
    
    if [ "$SHORT" = "true" ]; then
        flags="${flags} -short"
    fi
    
    # 排除集成测试目录
    local packages=$(go list ./... | grep -v '/test/')
    
    if [ "$COVERAGE" = "true" ]; then
        clean_coverage
        go test ${flags} -coverprofile="${COVERAGE_FILE}" -covermode=atomic ${packages}
        generate_coverage_report
    else
        go test ${flags} ${packages}
    fi
}

# 运行集成测试
run_integration_tests() {
    print_header "Running Integration Tests"
    
    local flags="-v"
    
    if [ "$RACE" = "true" ]; then
        flags="${flags} -race"
    fi
    
    # 检查集成测试目录是否存在
    if [ -d "test" ]; then
        go test ${flags} ./test/...
    else
        print_warn "No integration tests found in test/ directory"
    fi
}

# 运行血缘解析测试
run_lineage_tests() {
    print_header "Running Lineage Tests"
    
    local flags="-v"
    
    if [ "$RACE" = "true" ]; then
        flags="${flags} -race"
    fi
    
    if [ "$COVERAGE" = "true" ]; then
        clean_coverage
        go test ${flags} -coverprofile="${COVERAGE_FILE}" -covermode=atomic ./internal/lineage/...
        generate_coverage_report
    else
        go test ${flags} ./internal/lineage/...
    fi
}

# 运行采集器测试
run_collector_tests() {
    print_header "Running Collector Tests"
    
    local flags="-v"
    
    if [ "$RACE" = "true" ]; then
        flags="${flags} -race"
    fi
    
    if [ "$COVERAGE" = "true" ]; then
        clean_coverage
        go test ${flags} -coverprofile="${COVERAGE_FILE}" -covermode=atomic ./internal/collector/...
        generate_coverage_report
    else
        go test ${flags} ./internal/collector/...
    fi
}

# 运行图数据库测试
run_graph_tests() {
    print_header "Running Graph Tests"
    
    local flags="-v"
    
    if [ "$RACE" = "true" ]; then
        flags="${flags} -race"
    fi
    
    if [ "$COVERAGE" = "true" ]; then
        clean_coverage
        go test ${flags} -coverprofile="${COVERAGE_FILE}" -covermode=atomic ./internal/graph/...
        generate_coverage_report
    else
        go test ${flags} ./internal/graph/...
    fi
}

# 生成覆盖率报告
generate_coverage_report() {
    print_header "Generating Coverage Report"
    
    if [ ! -f "${COVERAGE_FILE}" ]; then
        print_error "Coverage file not found: ${COVERAGE_FILE}"
        return 1
    fi
    
    # 生成 HTML 报告
    go tool cover -html="${COVERAGE_FILE}" -o "${COVERAGE_HTML}"
    print_info "HTML coverage report: ${COVERAGE_HTML}"
    
    # 显示覆盖率摘要
    local coverage=$(go tool cover -func="${COVERAGE_FILE}" | grep total | awk '{print $3}' | sed 's/%//')
    
    print_info "Total coverage: ${coverage}%"
    
    # 检查最小覆盖率
    if [ -n "${MIN_COVERAGE}" ]; then
        local coverage_int=${coverage%.*}
        if [ "${coverage_int}" -lt "${MIN_COVERAGE}" ]; then
            print_error "Coverage ${coverage}% is below minimum threshold ${MIN_COVERAGE}%"
            return 1
        else
            print_info "Coverage meets minimum threshold (${MIN_COVERAGE}%)"
        fi
    fi
}

# 运行代码检查
run_lint() {
    print_header "Running Linter"
    
    if command -v golangci-lint &> /dev/null; then
        golangci-lint run ./...
    else
        print_warn "golangci-lint not installed, skipping lint check"
        print_info "Install with: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"
    fi
}

# 运行格式检查
run_fmt_check() {
    print_header "Checking Code Format"
    
    local unformatted=$(gofmt -l .)
    
    if [ -n "${unformatted}" ]; then
        print_error "The following files are not formatted:"
        echo "${unformatted}"
        print_info "Run 'go fmt ./...' to fix"
        return 1
    else
        print_info "All files are properly formatted"
    fi
}

# 解析命令行参数
VERBOSE=false
COVERAGE=false
RACE=false
SHORT=false
TARGET="all"

while [[ $# -gt 0 ]]; do
    case $1 in
        -h|--help)
            show_help
            exit 0
            ;;
        -v|--verbose)
            VERBOSE=true
            shift
            ;;
        -c|--coverage)
            COVERAGE=true
            shift
            ;;
        --min-coverage)
            MIN_COVERAGE="$2"
            shift 2
            ;;
        --race)
            RACE=true
            shift
            ;;
        --short)
            SHORT=true
            shift
            ;;
        all|unit|integration|lineage|collector|graph|lint|fmt)
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
    print_header "Go Metadata Test Script"
    echo ""
    
    check_go
    
    case $TARGET in
        all)
            run_all_tests
            ;;
        unit)
            run_unit_tests
            ;;
        integration)
            run_integration_tests
            ;;
        lineage)
            run_lineage_tests
            ;;
        collector)
            run_collector_tests
            ;;
        graph)
            run_graph_tests
            ;;
        lint)
            run_lint
            ;;
        fmt)
            run_fmt_check
            ;;
    esac
    
    print_header "Test completed"
}

main
