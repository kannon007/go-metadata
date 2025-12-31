package collector

import (
	"errors"
	"fmt"
)

// ErrorCode 错误码
type ErrorCode string

const (
	// ErrCodeAuthError 认证错误
	ErrCodeAuthError ErrorCode = "AUTH_ERROR"
	// ErrCodeNetworkError 网络错误
	ErrCodeNetworkError ErrorCode = "NETWORK_ERROR"
	// ErrCodeTimeout 超时错误
	ErrCodeTimeout ErrorCode = "TIMEOUT"
	// ErrCodeNotFound 资源未找到
	ErrCodeNotFound ErrorCode = "NOT_FOUND"
	// ErrCodeUnsupportedFeature 不支持的功能
	ErrCodeUnsupportedFeature ErrorCode = "UNSUPPORTED_FEATURE"
	// ErrCodeInvalidConfig 无效配置
	ErrCodeInvalidConfig ErrorCode = "INVALID_CONFIG"
	// ErrCodeQueryError 查询错误
	ErrCodeQueryError ErrorCode = "QUERY_ERROR"
	// ErrCodeParseError 解析错误
	ErrCodeParseError ErrorCode = "PARSE_ERROR"
	// ErrCodeConnectionClosed 连接已关闭
	ErrCodeConnectionClosed ErrorCode = "CONNECTION_CLOSED"
	// ErrCodePermissionDenied 权限拒绝
	ErrCodePermissionDenied ErrorCode = "PERMISSION_DENIED"
	// ErrCodeCancelled 操作被取消
	ErrCodeCancelled ErrorCode = "CANCELLED"
	// ErrCodeDeadlineExceeded 超过截止时间
	ErrCodeDeadlineExceeded ErrorCode = "DEADLINE_EXCEEDED"
	// ErrCodeInferenceError Schema 推断错误
	ErrCodeInferenceError ErrorCode = "INFERENCE_ERROR"
)

// CollectorError 采集器错误
type CollectorError struct {
	Code      ErrorCode          `json:"code"`
	Message   string             `json:"message"`
	Category  DataSourceCategory `json:"category"`  // RDBMS, DocumentDB, etc.
	Source    string             `json:"source"`    // mysql, postgres, hive
	Operation string             `json:"operation"` // connect, list_tables, etc.
	Cause     error              `json:"-"`
	Retryable bool               `json:"retryable"`
}

// Error 实现 error 接口
func (e *CollectorError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("[%s] %s: %s (category=%s, source=%s, operation=%s)", e.Code, e.Message, e.Cause.Error(), e.Category, e.Source, e.Operation)
	}
	return fmt.Sprintf("[%s] %s (category=%s, source=%s, operation=%s)", e.Code, e.Message, e.Category, e.Source, e.Operation)
}

// Unwrap 返回原始错误，支持 errors.Unwrap
func (e *CollectorError) Unwrap() error {
	return e.Cause
}

// Is 支持 errors.Is 比较
func (e *CollectorError) Is(target error) bool {
	if target == nil {
		return false
	}
	t, ok := target.(*CollectorError)
	if !ok {
		return false
	}
	return e.Code == t.Code
}


// NewCollectorError 创建通用采集器错误
func NewCollectorError(code ErrorCode, message string, category DataSourceCategory, source, operation string, cause error, retryable bool) *CollectorError {
	return &CollectorError{
		Code:      code,
		Message:   message,
		Category:  category,
		Source:    source,
		Operation: operation,
		Cause:     cause,
		Retryable: retryable,
	}
}

// NewAuthError 创建认证错误
func NewAuthError(source, operation string, cause error) *CollectorError {
	return NewAuthErrorWithCategory(GetCategoryByType(source), source, operation, cause)
}

// NewAuthErrorWithCategory 创建带类别的认证错误
func NewAuthErrorWithCategory(category DataSourceCategory, source, operation string, cause error) *CollectorError {
	return &CollectorError{
		Code:      ErrCodeAuthError,
		Message:   "authentication failed",
		Category:  category,
		Source:    source,
		Operation: operation,
		Cause:     cause,
		Retryable: false,
	}
}

// NewNetworkError 创建网络错误
func NewNetworkError(source, operation string, cause error) *CollectorError {
	return NewNetworkErrorWithCategory(GetCategoryByType(source), source, operation, cause)
}

// NewNetworkErrorWithCategory 创建带类别的网络错误
func NewNetworkErrorWithCategory(category DataSourceCategory, source, operation string, cause error) *CollectorError {
	return &CollectorError{
		Code:      ErrCodeNetworkError,
		Message:   "network error",
		Category:  category,
		Source:    source,
		Operation: operation,
		Cause:     cause,
		Retryable: true,
	}
}

// NewTimeoutError 创建超时错误
func NewTimeoutError(source, operation string, cause error) *CollectorError {
	return NewTimeoutErrorWithCategory(GetCategoryByType(source), source, operation, cause)
}

// NewTimeoutErrorWithCategory 创建带类别的超时错误
func NewTimeoutErrorWithCategory(category DataSourceCategory, source, operation string, cause error) *CollectorError {
	return &CollectorError{
		Code:      ErrCodeTimeout,
		Message:   "operation timed out",
		Category:  category,
		Source:    source,
		Operation: operation,
		Cause:     cause,
		Retryable: true,
	}
}

// NewNotFoundError 创建资源未找到错误
func NewNotFoundError(source, operation, resource string, cause error) *CollectorError {
	return NewNotFoundErrorWithCategory(GetCategoryByType(source), source, operation, resource, cause)
}

// NewNotFoundErrorWithCategory 创建带类别的资源未找到错误
func NewNotFoundErrorWithCategory(category DataSourceCategory, source, operation, resource string, cause error) *CollectorError {
	return &CollectorError{
		Code:      ErrCodeNotFound,
		Message:   fmt.Sprintf("resource not found: %s", resource),
		Category:  category,
		Source:    source,
		Operation: operation,
		Cause:     cause,
		Retryable: false,
	}
}

// NewUnsupportedFeatureError 创建不支持功能错误
func NewUnsupportedFeatureError(source, operation, feature string) *CollectorError {
	return NewUnsupportedFeatureErrorWithCategory(GetCategoryByType(source), source, operation, feature)
}

// NewUnsupportedFeatureErrorWithCategory 创建带类别的不支持功能错误
func NewUnsupportedFeatureErrorWithCategory(category DataSourceCategory, source, operation, feature string) *CollectorError {
	return &CollectorError{
		Code:      ErrCodeUnsupportedFeature,
		Message:   fmt.Sprintf("unsupported feature: %s", feature),
		Category:  category,
		Source:    source,
		Operation: operation,
		Cause:     nil,
		Retryable: false,
	}
}

// NewInvalidConfigError 创建无效配置错误
func NewInvalidConfigError(source, field, reason string) *CollectorError {
	return NewInvalidConfigErrorWithCategory(GetCategoryByType(source), source, field, reason)
}

// NewInvalidConfigErrorWithCategory 创建带类别的无效配置错误
func NewInvalidConfigErrorWithCategory(category DataSourceCategory, source, field, reason string) *CollectorError {
	return &CollectorError{
		Code:      ErrCodeInvalidConfig,
		Message:   fmt.Sprintf("invalid configuration: %s - %s", field, reason),
		Category:  category,
		Source:    source,
		Operation: "validate_config",
		Cause:     nil,
		Retryable: false,
	}
}

// NewQueryError 创建查询错误
func NewQueryError(source, operation string, cause error) *CollectorError {
	return NewQueryErrorWithCategory(GetCategoryByType(source), source, operation, cause)
}

// NewQueryErrorWithCategory 创建带类别的查询错误
func NewQueryErrorWithCategory(category DataSourceCategory, source, operation string, cause error) *CollectorError {
	return &CollectorError{
		Code:      ErrCodeQueryError,
		Message:   "query execution failed",
		Category:  category,
		Source:    source,
		Operation: operation,
		Cause:     cause,
		Retryable: false,
	}
}

// NewParseError 创建解析错误
func NewParseError(source, operation string, cause error) *CollectorError {
	return NewParseErrorWithCategory(GetCategoryByType(source), source, operation, cause)
}

// NewParseErrorWithCategory 创建带类别的解析错误
func NewParseErrorWithCategory(category DataSourceCategory, source, operation string, cause error) *CollectorError {
	return &CollectorError{
		Code:      ErrCodeParseError,
		Message:   "failed to parse response",
		Category:  category,
		Source:    source,
		Operation: operation,
		Cause:     cause,
		Retryable: false,
	}
}

// NewConnectionClosedError 创建连接已关闭错误
func NewConnectionClosedError(source, operation string) *CollectorError {
	return NewConnectionClosedErrorWithCategory(GetCategoryByType(source), source, operation)
}

// NewConnectionClosedErrorWithCategory 创建带类别的连接已关闭错误
func NewConnectionClosedErrorWithCategory(category DataSourceCategory, source, operation string) *CollectorError {
	return &CollectorError{
		Code:      ErrCodeConnectionClosed,
		Message:   "connection is closed",
		Category:  category,
		Source:    source,
		Operation: operation,
		Cause:     nil,
		Retryable: false,
	}
}

// NewPermissionDeniedError 创建权限拒绝错误
func NewPermissionDeniedError(source, operation string, cause error) *CollectorError {
	return NewPermissionDeniedErrorWithCategory(GetCategoryByType(source), source, operation, cause)
}

// NewPermissionDeniedErrorWithCategory 创建带类别的权限拒绝错误
func NewPermissionDeniedErrorWithCategory(category DataSourceCategory, source, operation string, cause error) *CollectorError {
	return &CollectorError{
		Code:      ErrCodePermissionDenied,
		Message:   "permission denied",
		Category:  category,
		Source:    source,
		Operation: operation,
		Cause:     cause,
		Retryable: false,
	}
}

// IsRetryable 检查错误是否可重试
func IsRetryable(err error) bool {
	var collErr *CollectorError
	if errors.As(err, &collErr) {
		return collErr.Retryable
	}
	return false
}

// GetErrorCode 获取错误码
func GetErrorCode(err error) ErrorCode {
	var collErr *CollectorError
	if errors.As(err, &collErr) {
		return collErr.Code
	}
	return ""
}

// GetErrorCategory 获取错误类别
func GetErrorCategory(err error) DataSourceCategory {
	var collErr *CollectorError
	if errors.As(err, &collErr) {
		return collErr.Category
	}
	return ""
}

// NewCancelledError 创建取消错误
func NewCancelledError(source, operation string, cause error) *CollectorError {
	return NewCancelledErrorWithCategory(GetCategoryByType(source), source, operation, cause)
}

// NewCancelledErrorWithCategory 创建带类别的取消错误
func NewCancelledErrorWithCategory(category DataSourceCategory, source, operation string, cause error) *CollectorError {
	return &CollectorError{
		Code:      ErrCodeCancelled,
		Message:   "operation cancelled",
		Category:  category,
		Source:    source,
		Operation: operation,
		Cause:     cause,
		Retryable: false,
	}
}

// NewDeadlineExceededError 创建截止时间超过错误
func NewDeadlineExceededError(source, operation string, cause error) *CollectorError {
	return NewDeadlineExceededErrorWithCategory(GetCategoryByType(source), source, operation, cause)
}

// NewDeadlineExceededErrorWithCategory 创建带类别的截止时间超过错误
func NewDeadlineExceededErrorWithCategory(category DataSourceCategory, source, operation string, cause error) *CollectorError {
	return &CollectorError{
		Code:      ErrCodeDeadlineExceeded,
		Message:   "deadline exceeded",
		Category:  category,
		Source:    source,
		Operation: operation,
		Cause:     cause,
		Retryable: true,
	}
}

// NewInferenceError 创建 Schema 推断错误
func NewInferenceError(source, operation string, cause error) *CollectorError {
	return NewInferenceErrorWithCategory(GetCategoryByType(source), source, operation, cause)
}

// NewInferenceErrorWithCategory 创建带类别的 Schema 推断错误
func NewInferenceErrorWithCategory(category DataSourceCategory, source, operation string, cause error) *CollectorError {
	return &CollectorError{
		Code:      ErrCodeInferenceError,
		Message:   "schema inference failed",
		Category:  category,
		Source:    source,
		Operation: operation,
		Cause:     cause,
		Retryable: false,
	}
}

// Sentinel errors for backward compatibility
var (
	// ErrConnectionFailed is returned when the collector fails to connect to the data source.
	ErrConnectionFailed = errors.New("failed to connect to data source")

	// ErrDatabaseNotFound is returned when the specified database is not found.
	ErrDatabaseNotFound = errors.New("database not found")

	// ErrTableNotFound is returned when the specified table is not found.
	ErrTableNotFound = errors.New("table not found")

	// ErrPermissionDenied is returned when the collector lacks permission to access the resource.
	ErrPermissionDenied = errors.New("permission denied")

	// ErrQueryFailed is returned when a metadata query fails.
	ErrQueryFailed = errors.New("query failed")

	// ErrUnsupportedDataSource is returned when the data source type is not supported.
	ErrUnsupportedDataSource = errors.New("unsupported data source type")

	// ErrInvalidConfig is returned when the collector configuration is invalid.
	ErrInvalidConfig = errors.New("invalid collector configuration")

	// ErrConnectionClosed is returned when attempting to use a closed connection.
	ErrConnectionClosed = errors.New("connection is closed")
)
