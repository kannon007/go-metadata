package scheduler

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/go-kratos/kratos/v2/log"

	"go-metadata/internal/biz"
	"go-metadata/internal/conf"
	"go-metadata/internal/scheduler/types"
)

// SchedulerManager 调度器管理器，支持动态切换调度器
type SchedulerManager struct {
	adapters       map[types.SchedulerType]types.SchedulerAdapter
	currentType    types.SchedulerType
	currentAdapter types.SchedulerAdapter
	config         *conf.Scheduler
	mu             sync.RWMutex
	log            *log.Helper
	taskRepo       biz.TaskRepo
}

// NewSchedulerManager 创建调度器管理器
func NewSchedulerManager(config *conf.Scheduler, logger log.Logger) *SchedulerManager {
	return &SchedulerManager{
		adapters: make(map[types.SchedulerType]types.SchedulerAdapter),
		config:   config,
		log:      log.NewHelper(logger),
	}
}

// RegisterAdapter 注册调度器适配器
func (m *SchedulerManager) RegisterAdapter(adapter types.SchedulerAdapter) {
	m.mu.Lock()
	defer m.mu.Unlock()

	adapterType := adapter.GetType()
	m.adapters[adapterType] = adapter
	m.log.Infof("Scheduler adapter registered: %s", adapterType)
}

// SetTaskRepo 设置任务仓储
func (m *SchedulerManager) SetTaskRepo(repo biz.TaskRepo) {
	m.taskRepo = repo
}

// Initialize 初始化调度器管理器
func (m *SchedulerManager) Initialize(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// 确定初始调度器类型
	schedulerType := types.SchedulerType(m.config.Type)
	if schedulerType == "" {
		schedulerType = types.SchedulerTypeBuiltIn
	}

	adapter, exists := m.adapters[schedulerType]
	if !exists {
		return fmt.Errorf("scheduler adapter not found: %s", schedulerType)
	}

	// 初始化适配器
	if err := adapter.Initialize(ctx); err != nil {
		return fmt.Errorf("failed to initialize scheduler adapter: %w", err)
	}

	m.currentType = schedulerType
	m.currentAdapter = adapter
	m.log.WithContext(ctx).Infof("Scheduler manager initialized with: %s", schedulerType)
	return nil
}

// Shutdown 关闭调度器管理器
func (m *SchedulerManager) Shutdown(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// 关闭所有适配器
	for adapterType, adapter := range m.adapters {
		if err := adapter.Shutdown(ctx); err != nil {
			m.log.WithContext(ctx).Warnf("Failed to shutdown adapter %s: %v", adapterType, err)
		}
	}

	m.log.WithContext(ctx).Info("Scheduler manager shutdown")
	return nil
}

// GetCurrentType 获取当前调度器类型
func (m *SchedulerManager) GetCurrentType() types.SchedulerType {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.currentType
}

// GetCurrentAdapter 获取当前调度器适配器
func (m *SchedulerManager) GetCurrentAdapter() types.SchedulerAdapter {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.currentAdapter
}

// GetAdapter 获取指定类型的调度器适配器
func (m *SchedulerManager) GetAdapter(schedulerType types.SchedulerType) (types.SchedulerAdapter, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	adapter, exists := m.adapters[schedulerType]
	if !exists {
		return nil, fmt.Errorf("scheduler adapter not found: %s", schedulerType)
	}
	return adapter, nil
}

// ListAdapters 列出所有已注册的调度器适配器
func (m *SchedulerManager) ListAdapters() []types.SchedulerType {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var adapterTypes []types.SchedulerType
	for t := range m.adapters {
		adapterTypes = append(adapterTypes, t)
	}
	return adapterTypes
}

// SwitchScheduler 切换调度器类型
func (m *SchedulerManager) SwitchScheduler(ctx context.Context, newType types.SchedulerType) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.currentType == newType {
		return nil // 已经是目标类型
	}

	newAdapter, exists := m.adapters[newType]
	if !exists {
		return fmt.Errorf("scheduler adapter not found: %s", newType)
	}

	// 初始化新适配器
	if err := newAdapter.Initialize(ctx); err != nil {
		return fmt.Errorf("failed to initialize new scheduler adapter: %w", err)
	}

	// 迁移工作流（如果需要）
	if m.currentAdapter != nil {
		if err := m.migrateWorkflows(ctx, m.currentAdapter, newAdapter); err != nil {
			m.log.WithContext(ctx).Warnf("Failed to migrate workflows: %v", err)
			// 继续切换，不阻塞
		}
	}

	// 关闭旧适配器
	if m.currentAdapter != nil {
		if err := m.currentAdapter.Shutdown(ctx); err != nil {
			m.log.WithContext(ctx).Warnf("Failed to shutdown old adapter: %v", err)
		}
	}

	m.currentType = newType
	m.currentAdapter = newAdapter
	m.log.WithContext(ctx).Infof("Scheduler switched from %s to %s", m.currentType, newType)
	return nil
}

// migrateWorkflows 迁移工作流
func (m *SchedulerManager) migrateWorkflows(ctx context.Context, from, to types.SchedulerAdapter) error {
	// 获取所有任务并在新调度器中重新创建
	if m.taskRepo == nil {
		return nil
	}

	tasks, err := m.taskRepo.ListTasks(ctx, &biz.ListTasksRequest{
		Page:     1,
		PageSize: 1000,
	})
	if err != nil {
		return fmt.Errorf("failed to list tasks: %w", err)
	}

	for _, task := range tasks.Tasks {
		// 在新调度器中创建工作流
		req := &types.CreateWorkflowRequest{
			ID:           task.ID,
			Name:         task.Name,
			TaskType:     task.Type,
			Schedule:     task.Schedule,
			Config:       task.Config,
			DataSourceID: task.DataSourceID,
		}

		if _, err := to.CreateWorkflow(ctx, req); err != nil {
			m.log.WithContext(ctx).Warnf("Failed to migrate workflow %s: %v", task.ID, err)
			continue
		}

		// 如果任务是活跃状态，启动工作流
		if task.Status == biz.TaskStatusActive {
			if starter, ok := to.(interface{ StartWorkflow(context.Context, string) error }); ok {
				if err := starter.StartWorkflow(ctx, task.ID); err != nil {
					m.log.WithContext(ctx).Warnf("Failed to start migrated workflow %s: %v", task.ID, err)
				}
			}
		}
	}

	return nil
}


// SchedulerConfig 调度器配置
type SchedulerConfig struct {
	Type       types.SchedulerType   `json:"type"`
	Properties map[string]string     `json:"properties"`
}

// GetConfig 获取当前调度器配置
func (m *SchedulerManager) GetConfig() *SchedulerConfig {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return &SchedulerConfig{
		Type:       m.currentType,
		Properties: m.config.Properties,
	}
}

// UpdateConfig 更新调度器配置
func (m *SchedulerManager) UpdateConfig(ctx context.Context, config *SchedulerConfig) error {
	// 如果类型改变，切换调度器
	if config.Type != m.currentType {
		if err := m.SwitchScheduler(ctx, config.Type); err != nil {
			return err
		}
	}

	// 更新配置属性
	m.mu.Lock()
	m.config.Properties = config.Properties
	m.mu.Unlock()

	return nil
}

// 实现 biz.TaskScheduler 接口，作为代理转发到当前适配器

// CreateTask 创建任务
func (m *SchedulerManager) CreateTask(ctx context.Context, task *biz.CollectionTask) error {
	adapter := m.GetCurrentAdapter()
	if adapter == nil {
		return fmt.Errorf("no scheduler adapter available")
	}

	req := &types.CreateWorkflowRequest{
		ID:           task.ID,
		Name:         task.Name,
		TaskType:     task.Type,
		Schedule:     task.Schedule,
		Config:       task.Config,
		DataSourceID: task.DataSourceID,
	}

	_, err := adapter.CreateWorkflow(ctx, req)
	return err
}

// UpdateTask 更新任务
func (m *SchedulerManager) UpdateTask(ctx context.Context, task *biz.CollectionTask) error {
	adapter := m.GetCurrentAdapter()
	if adapter == nil {
		return fmt.Errorf("no scheduler adapter available")
	}

	req := &types.UpdateWorkflowRequest{
		ID:          task.ID,
		Name:        task.Name,
		Schedule:    task.Schedule,
		Config:      task.Config,
	}

	_, err := adapter.UpdateWorkflow(ctx, req)
	return err
}

// DeleteTask 删除任务
func (m *SchedulerManager) DeleteTask(ctx context.Context, id string) error {
	adapter := m.GetCurrentAdapter()
	if adapter == nil {
		return fmt.Errorf("no scheduler adapter available")
	}

	return adapter.DeleteWorkflow(ctx, id)
}

// StartTask 启动任务
func (m *SchedulerManager) StartTask(ctx context.Context, id string) error {
	adapter := m.GetCurrentAdapter()
	if adapter == nil {
		return fmt.Errorf("no scheduler adapter available")
	}

	// 尝试调用 StartWorkflow 方法
	if starter, ok := adapter.(interface{ StartWorkflow(context.Context, string) error }); ok {
		return starter.StartWorkflow(ctx, id)
	}

	// 否则触发一次执行
	_, err := adapter.TriggerWorkflow(ctx, id, nil)
	return err
}

// StopTask 停止任务
func (m *SchedulerManager) StopTask(ctx context.Context, id string) error {
	adapter := m.GetCurrentAdapter()
	if adapter == nil {
		return fmt.Errorf("no scheduler adapter available")
	}

	// 尝试调用 StopWorkflow 方法
	if stopper, ok := adapter.(interface{ StopWorkflow(context.Context, string) error }); ok {
		return stopper.StopWorkflow(ctx, id)
	}

	return adapter.PauseWorkflow(ctx, id)
}

// PauseTask 暂停任务
func (m *SchedulerManager) PauseTask(ctx context.Context, id string) error {
	adapter := m.GetCurrentAdapter()
	if adapter == nil {
		return fmt.Errorf("no scheduler adapter available")
	}

	return adapter.PauseWorkflow(ctx, id)
}

// ResumeTask 恢复任务
func (m *SchedulerManager) ResumeTask(ctx context.Context, id string) error {
	adapter := m.GetCurrentAdapter()
	if adapter == nil {
		return fmt.Errorf("no scheduler adapter available")
	}

	return adapter.ResumeWorkflow(ctx, id)
}

// GetTaskStatus 获取任务状态
func (m *SchedulerManager) GetTaskStatus(ctx context.Context, id string) (*biz.TaskStatus, error) {
	adapter := m.GetCurrentAdapter()
	if adapter == nil {
		return nil, fmt.Errorf("no scheduler adapter available")
	}

	workflowStatus, err := adapter.GetWorkflowStatus(ctx, id)
	if err != nil {
		return nil, err
	}

	// 转换工作流状态到任务状态
	var taskStatus biz.TaskStatus
	switch *workflowStatus {
	case types.WorkflowStatusActive:
		taskStatus = biz.TaskStatusActive
	case types.WorkflowStatusInactive:
		taskStatus = biz.TaskStatusInactive
	case types.WorkflowStatusPaused:
		taskStatus = biz.TaskStatusPaused
	case types.WorkflowStatusError:
		taskStatus = biz.TaskStatusFailed
	default:
		taskStatus = biz.TaskStatusInactive
	}

	return &taskStatus, nil
}

// TriggerTask 触发任务执行
func (m *SchedulerManager) TriggerTask(ctx context.Context, id string, params map[string]interface{}) (*types.WorkflowExecution, error) {
	adapter := m.GetCurrentAdapter()
	if adapter == nil {
		return nil, fmt.Errorf("no scheduler adapter available")
	}

	return adapter.TriggerWorkflow(ctx, id, params)
}

// GetExecutionLogs 获取执行日志
func (m *SchedulerManager) GetExecutionLogs(ctx context.Context, executionID string) (*types.ExecutionLogs, error) {
	adapter := m.GetCurrentAdapter()
	if adapter == nil {
		return nil, fmt.Errorf("no scheduler adapter available")
	}

	return adapter.GetExecutionLogs(ctx, executionID)
}

// SchedulerStatus 调度器状态
type SchedulerStatus struct {
	Type            types.SchedulerType `json:"type"`
	Status          string              `json:"status"`
	ActiveWorkflows int                 `json:"active_workflows"`
	LastSwitchTime  *time.Time          `json:"last_switch_time,omitempty"`
}

// GetStatus 获取调度器状态
func (m *SchedulerManager) GetStatus(ctx context.Context) (*SchedulerStatus, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	status := &SchedulerStatus{
		Type:   m.currentType,
		Status: "running",
	}

	// 获取活跃工作流数量
	if m.currentAdapter != nil {
		if getter, ok := m.currentAdapter.(interface{ GetRunningWorkflows() []string }); ok {
			status.ActiveWorkflows = len(getter.GetRunningWorkflows())
		}
	}

	return status, nil
}
