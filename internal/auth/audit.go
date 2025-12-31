package auth

import (
	"context"
	"encoding/json"
	"time"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/uuid"
)

// AuditAction 审计操作类型
type AuditAction string

const (
	// 认证相关操作
	AuditActionLogin         AuditAction = "login"
	AuditActionLogout        AuditAction = "logout"
	AuditActionLoginFailed   AuditAction = "login_failed"
	AuditActionTokenRefresh  AuditAction = "token_refresh"
	AuditActionPasswordChange AuditAction = "password_change"

	// 数据源操作
	AuditActionDataSourceCreate AuditAction = "datasource_create"
	AuditActionDataSourceUpdate AuditAction = "datasource_update"
	AuditActionDataSourceDelete AuditAction = "datasource_delete"
	AuditActionDataSourceTest   AuditAction = "datasource_test"

	// 任务操作
	AuditActionTaskCreate  AuditAction = "task_create"
	AuditActionTaskUpdate  AuditAction = "task_update"
	AuditActionTaskDelete  AuditAction = "task_delete"
	AuditActionTaskStart   AuditAction = "task_start"
	AuditActionTaskStop    AuditAction = "task_stop"
	AuditActionTaskPause   AuditAction = "task_pause"
	AuditActionTaskResume  AuditAction = "task_resume"

	// 系统操作
	AuditActionConfigChange AuditAction = "config_change"
	AuditActionBatchOp      AuditAction = "batch_operation"
	AuditActionExport       AuditAction = "export"
	AuditActionImport       AuditAction = "import"
)

// AuditSeverity 审计严重级别
type AuditSeverity string

const (
	AuditSeverityInfo     AuditSeverity = "info"
	AuditSeverityWarning  AuditSeverity = "warning"
	AuditSeverityCritical AuditSeverity = "critical"
)

// AuditEntry 审计日志条目
type AuditEntry struct {
	ID           string                 `json:"id"`
	Timestamp    time.Time              `json:"timestamp"`
	Action       AuditAction            `json:"action"`
	Severity     AuditSeverity          `json:"severity"`
	EntityType   string                 `json:"entity_type"`
	EntityID     string                 `json:"entity_id"`
	EntityName   string                 `json:"entity_name,omitempty"`
	UserID       string                 `json:"user_id"`
	Username     string                 `json:"username"`
	UserRole     Role                   `json:"user_role"`
	ClientIP     string                 `json:"client_ip"`
	UserAgent    string                 `json:"user_agent"`
	RequestID    string                 `json:"request_id"`
	OldValue     map[string]interface{} `json:"old_value,omitempty"`
	NewValue     map[string]interface{} `json:"new_value,omitempty"`
	Details      map[string]interface{} `json:"details,omitempty"`
	Success      bool                   `json:"success"`
	ErrorMessage string                 `json:"error_message,omitempty"`
}

// AuditLogger 审计日志记录器接口
type AuditLogger interface {
	Log(ctx context.Context, entry *AuditEntry) error
	LogAction(ctx context.Context, action AuditAction, entityType, entityID string, details map[string]interface{}) error
	LogSensitiveAction(ctx context.Context, action AuditAction, entityType, entityID string, oldValue, newValue map[string]interface{}) error
}

// DefaultAuditLogger 默认审计日志记录器
type DefaultAuditLogger struct {
	log       *log.Helper
	encryptor Encryptor
}

// NewDefaultAuditLogger 创建默认审计日志记录器
func NewDefaultAuditLogger(logger log.Logger, encryptor Encryptor) *DefaultAuditLogger {
	if encryptor == nil {
		encryptor = NewNoOpEncryptor()
	}
	return &DefaultAuditLogger{
		log:       log.NewHelper(logger),
		encryptor: encryptor,
	}
}

// Log 记录审计日志
func (l *DefaultAuditLogger) Log(ctx context.Context, entry *AuditEntry) error {
	// 确保ID存在
	if entry.ID == "" {
		entry.ID = uuid.New().String()
	}

	// 确保时间戳存在
	if entry.Timestamp.IsZero() {
		entry.Timestamp = time.Now()
	}

	// 从上下文获取请求信息
	if entry.RequestID == "" {
		entry.RequestID = RequestIDFromContext(ctx)
	}
	if entry.ClientIP == "" {
		entry.ClientIP = ClientIPFromContext(ctx)
	}
	if entry.UserAgent == "" {
		entry.UserAgent = UserAgentFromContext(ctx)
	}

	// 从上下文获取用户信息
	if user, ok := UserFromContext(ctx); ok {
		if entry.UserID == "" {
			entry.UserID = user.ID
		}
		if entry.Username == "" {
			entry.Username = user.Username
		}
		if entry.UserRole == "" {
			entry.UserRole = user.Role
		}
	}

	// 掩码敏感数据
	if entry.OldValue != nil {
		entry.OldValue = MaskConnectionConfig(entry.OldValue)
	}
	if entry.NewValue != nil {
		entry.NewValue = MaskConnectionConfig(entry.NewValue)
	}

	// 记录到日志
	logData, _ := json.Marshal(entry)
	l.log.WithContext(ctx).Infof("AUDIT: %s", string(logData))

	return nil
}

// LogAction 记录操作审计日志
func (l *DefaultAuditLogger) LogAction(ctx context.Context, action AuditAction, entityType, entityID string, details map[string]interface{}) error {
	entry := &AuditEntry{
		Action:     action,
		Severity:   getSeverityForAction(action),
		EntityType: entityType,
		EntityID:   entityID,
		Details:    details,
		Success:    true,
	}
	return l.Log(ctx, entry)
}

// LogSensitiveAction 记录敏感操作审计日志
func (l *DefaultAuditLogger) LogSensitiveAction(ctx context.Context, action AuditAction, entityType, entityID string, oldValue, newValue map[string]interface{}) error {
	entry := &AuditEntry{
		Action:     action,
		Severity:   AuditSeverityCritical,
		EntityType: entityType,
		EntityID:   entityID,
		OldValue:   oldValue,
		NewValue:   newValue,
		Success:    true,
	}
	return l.Log(ctx, entry)
}

// LogError 记录错误审计日志
func (l *DefaultAuditLogger) LogError(ctx context.Context, action AuditAction, entityType, entityID string, err error) error {
	entry := &AuditEntry{
		Action:       action,
		Severity:     AuditSeverityWarning,
		EntityType:   entityType,
		EntityID:     entityID,
		Success:      false,
		ErrorMessage: err.Error(),
	}
	return l.Log(ctx, entry)
}

// getSeverityForAction 根据操作类型获取严重级别
func getSeverityForAction(action AuditAction) AuditSeverity {
	switch action {
	case AuditActionLogin, AuditActionLogout, AuditActionTokenRefresh:
		return AuditSeverityInfo
	case AuditActionLoginFailed, AuditActionPasswordChange:
		return AuditSeverityWarning
	case AuditActionDataSourceDelete, AuditActionTaskDelete, AuditActionConfigChange:
		return AuditSeverityCritical
	default:
		return AuditSeverityInfo
	}
}

// SensitiveFields 敏感字段列表
var SensitiveFields = []string{
	"password",
	"secret",
	"api_key",
	"access_key",
	"secret_key",
	"token",
	"private_key",
	"credentials",
}

// IsSensitiveField 检查是否是敏感字段
func IsSensitiveField(field string) bool {
	for _, f := range SensitiveFields {
		if f == field {
			return true
		}
	}
	return false
}

// SanitizeForAudit 清理数据用于审计（移除敏感信息）
func SanitizeForAudit(data map[string]interface{}) map[string]interface{} {
	sanitized := make(map[string]interface{})
	for k, v := range data {
		if IsSensitiveField(k) {
			sanitized[k] = "[REDACTED]"
		} else if nested, ok := v.(map[string]interface{}); ok {
			sanitized[k] = SanitizeForAudit(nested)
		} else {
			sanitized[k] = v
		}
	}
	return sanitized
}
