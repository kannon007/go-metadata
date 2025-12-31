package types

import (
	"context"
	"time"

	"go-metadata/internal/biz"
)

// SchedulerType 调度器类型
type SchedulerType string

const (
	SchedulerTypeBuiltIn          SchedulerType = "builtin"
	SchedulerTypeDolphinScheduler SchedulerType = "dolphinscheduler"
)

// WorkflowStatus 工作流状态
type WorkflowStatus string

const (
	WorkflowStatusActive   WorkflowStatus = "active"
	WorkflowStatusInactive WorkflowStatus = "inactive"
	WorkflowStatusPaused   WorkflowStatus = "paused"
	WorkflowStatusError    WorkflowStatus = "error"
)

// ExecutionStatus 执行状态
type ExecutionStatus string

const (
	ExecutionStatusPending   ExecutionStatus = "pending"
	ExecutionStatusRunning   ExecutionStatus = "running"
	ExecutionStatusCompleted ExecutionStatus = "completed"
	ExecutionStatusFailed    ExecutionStatus = "failed"
	ExecutionStatusCancelled ExecutionStatus = "cancelled"
)

// CreateWorkflowRequest 创建工作流请求
type CreateWorkflowRequest struct {
	ID           string                 `json:"id"`
	Name         string                 `json:"name"`
	Description  string                 `json:"description"`
	TaskType     biz.TaskType           `json:"task_type"`
	Schedule     *biz.ScheduleConfig    `json:"schedule"`
	Config       *biz.TaskConfig        `json:"config"`
	DataSourceID string                 `json:"datasource_id"`
	Properties   map[string]interface{} `json:"properties"`
}

// UpdateWorkflowRequest 更新工作流请求
type UpdateWorkflowRequest struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Schedule    *biz.ScheduleConfig    `json:"schedule"`
	Config      *biz.TaskConfig        `json:"config"`
	Properties  map[string]interface{} `json:"properties"`
}

// Workflow 工作流
type Workflow struct {
	ID           string                 `json:"id"`
	Name         string                 `json:"name"`
	Description  string                 `json:"description"`
	Status       WorkflowStatus         `json:"status"`
	Schedule     *biz.ScheduleConfig    `json:"schedule"`
	Config       *biz.TaskConfig        `json:"config"`
	DataSourceID string                 `json:"datasource_id"`
	ExternalID   string                 `json:"external_id"`
	Properties   map[string]interface{} `json:"properties"`
	CreatedAt    time.Time              `json:"created_at"`
	UpdatedAt    time.Time              `json:"updated_at"`
}

// WorkflowExecution 工作流执行
type WorkflowExecution struct {
	ID           string                 `json:"id"`
	WorkflowID   string                 `json:"workflow_id"`
	Status       ExecutionStatus        `json:"status"`
	StartTime    time.Time              `json:"start_time"`
	EndTime      *time.Time             `json:"end_time"`
	Duration     int64                  `json:"duration"`
	Result       map[string]interface{} `json:"result"`
	ErrorMessage string                 `json:"error_message"`
	ExternalID   string                 `json:"external_id"`
}

// ExecutionLogs 执行日志
type ExecutionLogs struct {
	ExecutionID string     `json:"execution_id"`
	Logs        []LogEntry `json:"logs"`
}

// LogEntry 日志条目
type LogEntry struct {
	Timestamp time.Time `json:"timestamp"`
	Level     string    `json:"level"`
	Message   string    `json:"message"`
}

// SchedulerAdapter 调度器适配器接口
type SchedulerAdapter interface {
	GetType() SchedulerType
	Initialize(ctx context.Context) error
	Shutdown(ctx context.Context) error
	CreateWorkflow(ctx context.Context, req *CreateWorkflowRequest) (*Workflow, error)
	UpdateWorkflow(ctx context.Context, req *UpdateWorkflowRequest) (*Workflow, error)
	DeleteWorkflow(ctx context.Context, id string) error
	GetWorkflow(ctx context.Context, id string) (*Workflow, error)
	TriggerWorkflow(ctx context.Context, id string, params map[string]interface{}) (*WorkflowExecution, error)
	StopWorkflowExecution(ctx context.Context, executionID string) error
	PauseWorkflow(ctx context.Context, id string) error
	ResumeWorkflow(ctx context.Context, id string) error
	GetWorkflowStatus(ctx context.Context, id string) (*WorkflowStatus, error)
	GetExecutionLogs(ctx context.Context, executionID string) (*ExecutionLogs, error)
}

// TaskExecutor 任务执行器接口
type TaskExecutor interface {
	Execute(ctx context.Context, task *biz.CollectionTask) (*biz.ExecutionResult, error)
}
