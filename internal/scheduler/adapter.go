package scheduler

import (
	"go-metadata/internal/biz"
	"go-metadata/internal/scheduler/types"
)

// Re-export types for convenience
type (
	SchedulerType         = types.SchedulerType
	WorkflowStatus        = types.WorkflowStatus
	ExecutionStatus       = types.ExecutionStatus
	CreateWorkflowRequest = types.CreateWorkflowRequest
	UpdateWorkflowRequest = types.UpdateWorkflowRequest
	Workflow              = types.Workflow
	WorkflowExecution     = types.WorkflowExecution
	ExecutionLogs         = types.ExecutionLogs
	LogEntry              = types.LogEntry
	SchedulerAdapter      = types.SchedulerAdapter
	TaskExecutor          = types.TaskExecutor
)

// Re-export constants
const (
	SchedulerTypeBuiltIn          = types.SchedulerTypeBuiltIn
	SchedulerTypeDolphinScheduler = types.SchedulerTypeDolphinScheduler
	WorkflowStatusActive          = types.WorkflowStatusActive
	WorkflowStatusInactive        = types.WorkflowStatusInactive
	WorkflowStatusPaused          = types.WorkflowStatusPaused
	WorkflowStatusError           = types.WorkflowStatusError
	ExecutionStatusPending        = types.ExecutionStatusPending
	ExecutionStatusRunning        = types.ExecutionStatusRunning
	ExecutionStatusCompleted      = types.ExecutionStatusCompleted
	ExecutionStatusFailed         = types.ExecutionStatusFailed
	ExecutionStatusCancelled      = types.ExecutionStatusCancelled
)

// Ensure interface compliance
var _ biz.TaskScheduler = (*SchedulerManager)(nil)
