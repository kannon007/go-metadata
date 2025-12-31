package scheduler

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/uuid"
	"github.com/robfig/cron/v3"

	"go-metadata/internal/biz"
	"go-metadata/internal/scheduler/types"
)

// BuiltinScheduler 内置Go调度器实现
type BuiltinScheduler struct {
	cron       *cron.Cron
	workflows  map[string]*workflowEntry
	executions map[string]*types.WorkflowExecution
	mu         sync.RWMutex
	log        *log.Helper
	executor   types.TaskExecutor
	taskRepo   biz.TaskRepo
	running    bool
}

// workflowEntry 工作流条目
type workflowEntry struct {
	workflow  *types.Workflow
	cronID    cron.EntryID
	stopChan  chan struct{}
	isRunning bool
	lastRun   *time.Time
	nextRun   *time.Time
}

// NewBuiltinScheduler 创建内置调度器
func NewBuiltinScheduler(logger log.Logger) *BuiltinScheduler {
	return &BuiltinScheduler{
		cron:       cron.New(cron.WithSeconds()),
		workflows:  make(map[string]*workflowEntry),
		executions: make(map[string]*types.WorkflowExecution),
		log:        log.NewHelper(logger),
		running:    false,
	}
}

// SetTaskExecutor 设置任务执行器
func (s *BuiltinScheduler) SetTaskExecutor(executor types.TaskExecutor) {
	s.executor = executor
}

// SetTaskRepo 设置任务仓储
func (s *BuiltinScheduler) SetTaskRepo(repo biz.TaskRepo) {
	s.taskRepo = repo
}

// GetType 获取调度器类型
func (s *BuiltinScheduler) GetType() types.SchedulerType {
	return types.SchedulerTypeBuiltIn
}

// Initialize 初始化调度器
func (s *BuiltinScheduler) Initialize(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.running {
		return nil
	}

	s.cron.Start()
	s.running = true
	s.log.WithContext(ctx).Info("Builtin scheduler initialized")
	return nil
}

// Shutdown 关闭调度器
func (s *BuiltinScheduler) Shutdown(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.running {
		return nil
	}

	// 停止cron
	cronCtx := s.cron.Stop()
	select {
	case <-cronCtx.Done():
	case <-ctx.Done():
		return ctx.Err()
	}

	// 关闭所有工作流
	for _, entry := range s.workflows {
		if entry.stopChan != nil {
			close(entry.stopChan)
		}
	}

	s.workflows = make(map[string]*workflowEntry)
	s.running = false
	s.log.WithContext(ctx).Info("Builtin scheduler shutdown")
	return nil
}


// CreateWorkflow 创建工作流
func (s *BuiltinScheduler) CreateWorkflow(ctx context.Context, req *types.CreateWorkflowRequest) (*types.Workflow, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// 检查工作流是否已存在
	if _, exists := s.workflows[req.ID]; exists {
		return nil, fmt.Errorf("workflow %s already exists", req.ID)
	}

	now := time.Now()
	workflow := &types.Workflow{
		ID:           req.ID,
		Name:         req.Name,
		Description:  req.Description,
		Status:       types.WorkflowStatusInactive,
		Schedule:     req.Schedule,
		Config:       req.Config,
		DataSourceID: req.DataSourceID,
		Properties:   req.Properties,
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	entry := &workflowEntry{
		workflow: workflow,
		stopChan: make(chan struct{}),
	}

	s.workflows[req.ID] = entry
	s.log.WithContext(ctx).Infof("Workflow created: %s", req.ID)
	return workflow, nil
}

// UpdateWorkflow 更新工作流
func (s *BuiltinScheduler) UpdateWorkflow(ctx context.Context, req *types.UpdateWorkflowRequest) (*types.Workflow, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	entry, exists := s.workflows[req.ID]
	if !exists {
		return nil, fmt.Errorf("workflow %s not found", req.ID)
	}

	// 如果工作流正在运行且调度配置改变，需要重新调度
	if entry.cronID != 0 && req.Schedule != nil {
		s.cron.Remove(entry.cronID)
		entry.cronID = 0
	}

	// 更新工作流
	entry.workflow.Name = req.Name
	entry.workflow.Description = req.Description
	entry.workflow.Schedule = req.Schedule
	entry.workflow.Config = req.Config
	entry.workflow.Properties = req.Properties
	entry.workflow.UpdatedAt = time.Now()

	s.log.WithContext(ctx).Infof("Workflow updated: %s", req.ID)
	return entry.workflow, nil
}

// DeleteWorkflow 删除工作流
func (s *BuiltinScheduler) DeleteWorkflow(ctx context.Context, id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	entry, exists := s.workflows[id]
	if !exists {
		return nil // 工作流不存在，视为删除成功
	}

	// 停止cron任务
	if entry.cronID != 0 {
		s.cron.Remove(entry.cronID)
	}

	// 发送停止信号
	if entry.stopChan != nil {
		close(entry.stopChan)
	}

	delete(s.workflows, id)
	s.log.WithContext(ctx).Infof("Workflow deleted: %s", id)
	return nil
}

// GetWorkflow 获取工作流
func (s *BuiltinScheduler) GetWorkflow(ctx context.Context, id string) (*types.Workflow, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	entry, exists := s.workflows[id]
	if !exists {
		return nil, fmt.Errorf("workflow %s not found", id)
	}

	return entry.workflow, nil
}

// TriggerWorkflow 触发工作流执行
func (s *BuiltinScheduler) TriggerWorkflow(ctx context.Context, id string, params map[string]interface{}) (*types.WorkflowExecution, error) {
	s.mu.Lock()
	entry, exists := s.workflows[id]
	if !exists {
		s.mu.Unlock()
		return nil, fmt.Errorf("workflow %s not found", id)
	}

	// 创建执行记录
	now := time.Now()
	execution := &types.WorkflowExecution{
		ID:         uuid.New().String(),
		WorkflowID: id,
		Status:     types.ExecutionStatusPending,
		StartTime:  now,
		Result:     params,
	}

	s.executions[execution.ID] = execution
	entry.isRunning = true
	entry.lastRun = &now
	s.mu.Unlock()

	// 异步执行任务
	go s.executeWorkflow(context.Background(), entry, execution)

	s.log.WithContext(ctx).Infof("Workflow triggered: %s, execution: %s", id, execution.ID)
	return execution, nil
}

// executeWorkflow 执行工作流
func (s *BuiltinScheduler) executeWorkflow(ctx context.Context, entry *workflowEntry, execution *types.WorkflowExecution) {
	s.mu.Lock()
	execution.Status = types.ExecutionStatusRunning
	s.mu.Unlock()

	s.log.Infof("Executing workflow: %s, execution: %s", entry.workflow.ID, execution.ID)

	var result *biz.ExecutionResult
	var execErr error

	// 如果有任务执行器和任务仓储，执行实际任务
	if s.executor != nil && s.taskRepo != nil {
		task, err := s.taskRepo.GetTask(ctx, entry.workflow.ID)
		if err != nil {
			execErr = err
		} else {
			result, execErr = s.executor.Execute(ctx, task)
		}
	}

	// 更新执行状态
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()
	execution.EndTime = &now
	execution.Duration = now.Sub(execution.StartTime).Milliseconds()

	if execErr != nil {
		execution.Status = types.ExecutionStatusFailed
		execution.ErrorMessage = execErr.Error()
	} else {
		execution.Status = types.ExecutionStatusCompleted
		if result != nil {
			execution.Result = map[string]interface{}{
				"tables_processed":  result.TablesProcessed,
				"records_processed": result.RecordsProcessed,
				"errors_count":      result.ErrorsCount,
				"warnings_count":    result.WarningsCount,
			}
		}
	}

	entry.isRunning = false
	s.log.Infof("Workflow execution completed: %s, status: %s", execution.ID, execution.Status)
}


// StopWorkflowExecution 停止工作流执行
func (s *BuiltinScheduler) StopWorkflowExecution(ctx context.Context, executionID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	execution, exists := s.executions[executionID]
	if !exists {
		return fmt.Errorf("execution %s not found", executionID)
	}

	if execution.Status != types.ExecutionStatusRunning && execution.Status != types.ExecutionStatusPending {
		return fmt.Errorf("execution %s is not running", executionID)
	}

	now := time.Now()
	execution.Status = types.ExecutionStatusCancelled
	execution.EndTime = &now
	execution.Duration = now.Sub(execution.StartTime).Milliseconds()

	// 更新工作流状态
	if entry, ok := s.workflows[execution.WorkflowID]; ok {
		entry.isRunning = false
	}

	s.log.WithContext(ctx).Infof("Workflow execution stopped: %s", executionID)
	return nil
}

// PauseWorkflow 暂停工作流
func (s *BuiltinScheduler) PauseWorkflow(ctx context.Context, id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	entry, exists := s.workflows[id]
	if !exists {
		return fmt.Errorf("workflow %s not found", id)
	}

	// 移除cron任务但保留配置
	if entry.cronID != 0 {
		s.cron.Remove(entry.cronID)
		entry.cronID = 0
	}

	entry.workflow.Status = types.WorkflowStatusPaused
	entry.workflow.UpdatedAt = time.Now()

	s.log.WithContext(ctx).Infof("Workflow paused: %s", id)
	return nil
}

// ResumeWorkflow 恢复工作流
func (s *BuiltinScheduler) ResumeWorkflow(ctx context.Context, id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	entry, exists := s.workflows[id]
	if !exists {
		return fmt.Errorf("workflow %s not found", id)
	}

	if entry.workflow.Status != types.WorkflowStatusPaused {
		return fmt.Errorf("workflow %s is not paused", id)
	}

	// 重新设置调度
	if entry.workflow.Schedule != nil {
		if err := s.scheduleWorkflowLocked(ctx, entry); err != nil {
			return err
		}
	}

	entry.workflow.Status = types.WorkflowStatusActive
	entry.workflow.UpdatedAt = time.Now()

	s.log.WithContext(ctx).Infof("Workflow resumed: %s", id)
	return nil
}

// GetWorkflowStatus 获取工作流状态
func (s *BuiltinScheduler) GetWorkflowStatus(ctx context.Context, id string) (*types.WorkflowStatus, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	entry, exists := s.workflows[id]
	if !exists {
		status := types.WorkflowStatusInactive
		return &status, nil
	}

	return &entry.workflow.Status, nil
}

// GetExecutionLogs 获取执行日志
func (s *BuiltinScheduler) GetExecutionLogs(ctx context.Context, executionID string) (*types.ExecutionLogs, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	execution, exists := s.executions[executionID]
	if !exists {
		return nil, fmt.Errorf("execution %s not found", executionID)
	}

	logs := &types.ExecutionLogs{
		ExecutionID: executionID,
		Logs: []types.LogEntry{
			{
				Timestamp: execution.StartTime,
				Level:     "INFO",
				Message:   fmt.Sprintf("Workflow execution started: %s", execution.WorkflowID),
			},
		},
	}

	if execution.EndTime != nil {
		logs.Logs = append(logs.Logs, types.LogEntry{
			Timestamp: *execution.EndTime,
			Level:     "INFO",
			Message:   fmt.Sprintf("Workflow execution completed with status: %s", execution.Status),
		})
	}

	if execution.ErrorMessage != "" {
		logs.Logs = append(logs.Logs, types.LogEntry{
			Timestamp: *execution.EndTime,
			Level:     "ERROR",
			Message:   execution.ErrorMessage,
		})
	}

	return logs, nil
}

// StartWorkflow 启动工作流调度
func (s *BuiltinScheduler) StartWorkflow(ctx context.Context, id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	entry, exists := s.workflows[id]
	if !exists {
		return fmt.Errorf("workflow %s not found", id)
	}

	// 设置调度
	if entry.workflow.Schedule != nil {
		if err := s.scheduleWorkflowLocked(ctx, entry); err != nil {
			return err
		}
	}

	entry.workflow.Status = types.WorkflowStatusActive
	entry.workflow.UpdatedAt = time.Now()

	s.log.WithContext(ctx).Infof("Workflow started: %s", id)
	return nil
}

// scheduleWorkflowLocked 设置工作流调度（需要持有锁）
func (s *BuiltinScheduler) scheduleWorkflowLocked(ctx context.Context, entry *workflowEntry) error {
	schedule := entry.workflow.Schedule
	if schedule == nil {
		return nil
	}

	switch schedule.Type {
	case biz.ScheduleTypeImmediate:
		// 立即执行
		go func() {
			execution := &types.WorkflowExecution{
				ID:         uuid.New().String(),
				WorkflowID: entry.workflow.ID,
				Status:     types.ExecutionStatusPending,
				StartTime:  time.Now(),
			}
			s.mu.Lock()
			s.executions[execution.ID] = execution
			s.mu.Unlock()
			s.executeWorkflow(context.Background(), entry, execution)
		}()
		return nil

	case biz.ScheduleTypeCron:
		if schedule.CronExpr == "" {
			return fmt.Errorf("cron expression is required for cron schedule")
		}
		entryID, err := s.cron.AddFunc(schedule.CronExpr, func() {
			s.triggerWorkflowExecution(entry)
		})
		if err != nil {
			return fmt.Errorf("failed to add cron job: %w", err)
		}
		entry.cronID = entryID
		// 计算下次执行时间
		cronEntry := s.cron.Entry(entryID)
		nextRun := cronEntry.Next
		entry.nextRun = &nextRun
		return nil

	case biz.ScheduleTypeInterval:
		if schedule.Interval <= 0 {
			return fmt.Errorf("interval must be positive")
		}
		// 使用cron表达式实现间隔执行
		cronExpr := fmt.Sprintf("@every %ds", schedule.Interval)
		entryID, err := s.cron.AddFunc(cronExpr, func() {
			s.triggerWorkflowExecution(entry)
		})
		if err != nil {
			return fmt.Errorf("failed to add interval job: %w", err)
		}
		entry.cronID = entryID
		// 计算下次执行时间
		nextRun := time.Now().Add(time.Duration(schedule.Interval) * time.Second)
		entry.nextRun = &nextRun
		return nil

	case biz.ScheduleTypeOnce:
		if schedule.StartTime == nil {
			return fmt.Errorf("start time is required for once schedule")
		}
		// 定时执行一次
		go func() {
			delay := time.Until(*schedule.StartTime)
			if delay > 0 {
				select {
				case <-time.After(delay):
					s.triggerWorkflowExecution(entry)
				case <-entry.stopChan:
					return
				}
			} else {
				// 时间已过，立即执行
				s.triggerWorkflowExecution(entry)
			}
		}()
		entry.nextRun = schedule.StartTime
		return nil
	}

	return nil
}

// triggerWorkflowExecution 触发工作流执行
func (s *BuiltinScheduler) triggerWorkflowExecution(entry *workflowEntry) {
	s.mu.Lock()
	now := time.Now()
	execution := &types.WorkflowExecution{
		ID:         uuid.New().String(),
		WorkflowID: entry.workflow.ID,
		Status:     types.ExecutionStatusPending,
		StartTime:  now,
	}
	s.executions[execution.ID] = execution
	entry.isRunning = true
	entry.lastRun = &now
	s.mu.Unlock()

	s.executeWorkflow(context.Background(), entry, execution)
}

// GetRunningWorkflows 获取所有运行中的工作流
func (s *BuiltinScheduler) GetRunningWorkflows() []string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var running []string
	for id, entry := range s.workflows {
		if entry.isRunning {
			running = append(running, id)
		}
	}
	return running
}

// GetWorkflowNextRunTime 获取工作流下次执行时间
func (s *BuiltinScheduler) GetWorkflowNextRunTime(id string) *time.Time {
	s.mu.RLock()
	defer s.mu.RUnlock()

	entry, exists := s.workflows[id]
	if !exists {
		return nil
	}

	// 如果有cron任务，从cron获取下次执行时间
	if entry.cronID != 0 {
		cronEntry := s.cron.Entry(entry.cronID)
		return &cronEntry.Next
	}

	return entry.nextRun
}

// GetExecution 获取执行记录
func (s *BuiltinScheduler) GetExecution(ctx context.Context, executionID string) (*types.WorkflowExecution, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	execution, exists := s.executions[executionID]
	if !exists {
		return nil, fmt.Errorf("execution %s not found", executionID)
	}

	return execution, nil
}

// ListExecutions 列出工作流的执行记录
func (s *BuiltinScheduler) ListExecutions(ctx context.Context, workflowID string, limit int) ([]*types.WorkflowExecution, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var executions []*types.WorkflowExecution
	for _, exec := range s.executions {
		if exec.WorkflowID == workflowID {
			executions = append(executions, exec)
		}
	}

	// 按开始时间倒序排序
	for i := 0; i < len(executions)-1; i++ {
		for j := i + 1; j < len(executions); j++ {
			if executions[i].StartTime.Before(executions[j].StartTime) {
				executions[i], executions[j] = executions[j], executions[i]
			}
		}
	}

	if limit > 0 && len(executions) > limit {
		executions = executions[:limit]
	}

	return executions, nil
}
