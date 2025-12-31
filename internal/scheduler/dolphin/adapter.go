package dolphin

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/uuid"

	"go-metadata/internal/biz"
	"go-metadata/internal/scheduler/types"
)

// DolphinSchedulerAdapter DolphinScheduler适配器
type DolphinSchedulerAdapter struct {
	client      *Client
	projectCode int64
	workflows   map[string]*workflowMapping // 内部ID -> 外部映射
	mu          sync.RWMutex
	log         *log.Helper
	config      *Config
}

// Config DolphinScheduler配置
type Config struct {
	Endpoint    string
	Token       string
	ProjectName string
	Timeout     time.Duration
}

// workflowMapping 工作流映射
type workflowMapping struct {
	InternalID  string
	ExternalCode int64
	ScheduleID  int64
	Workflow    *types.Workflow
}

// NewDolphinSchedulerAdapter 创建DolphinScheduler适配器
func NewDolphinSchedulerAdapter(config *Config, logger log.Logger) *DolphinSchedulerAdapter {
	client := NewClient(&ClientConfig{
		Endpoint: config.Endpoint,
		Token:    config.Token,
		Timeout:  config.Timeout,
	})

	return &DolphinSchedulerAdapter{
		client:    client,
		workflows: make(map[string]*workflowMapping),
		log:       log.NewHelper(logger),
		config:    config,
	}
}

// GetType 获取调度器类型
func (a *DolphinSchedulerAdapter) GetType() types.SchedulerType {
	return types.SchedulerTypeDolphinScheduler
}

// Initialize 初始化调度器
func (a *DolphinSchedulerAdapter) Initialize(ctx context.Context) error {
	// 获取或创建项目
	projects, err := a.client.ListProjects(ctx, 1, 100)
	if err != nil {
		return fmt.Errorf("failed to list projects: %w", err)
	}

	// 查找已存在的项目
	for _, p := range projects {
		if p.Name == a.config.ProjectName {
			a.projectCode = p.Code
			a.log.WithContext(ctx).Infof("Using existing project: %s (code: %d)", p.Name, p.Code)
			return nil
		}
	}

	// 创建新项目
	project, err := a.client.CreateProject(ctx, a.config.ProjectName, "Metadata collection project")
	if err != nil {
		return fmt.Errorf("failed to create project: %w", err)
	}

	a.projectCode = project.Code
	a.log.WithContext(ctx).Infof("Created new project: %s (code: %d)", project.Name, project.Code)
	return nil
}

// Shutdown 关闭调度器
func (a *DolphinSchedulerAdapter) Shutdown(ctx context.Context) error {
	a.log.WithContext(ctx).Info("DolphinScheduler adapter shutdown")
	return nil
}

// CreateWorkflow 创建工作流
func (a *DolphinSchedulerAdapter) CreateWorkflow(ctx context.Context, req *types.CreateWorkflowRequest) (*types.Workflow, error) {
	// 构建任务定义
	taskDef := a.buildTaskDefinition(req)
	taskDefJSON, err := json.Marshal([]interface{}{taskDef})
	if err != nil {
		return nil, fmt.Errorf("failed to marshal task definition: %w", err)
	}

	// 构建任务关系（单任务无依赖）
	taskRelation := map[string]interface{}{
		"name":                  "",
		"preTaskCode":           0,
		"preTaskVersion":        0,
		"postTaskCode":          taskDef["code"],
		"postTaskVersion":       1,
		"conditionType":         "NONE",
		"conditionParams":       map[string]interface{}{},
	}
	taskRelationJSON, err := json.Marshal([]interface{}{taskRelation})
	if err != nil {
		return nil, fmt.Errorf("failed to marshal task relation: %w", err)
	}

	// 构建位置信息
	locations := fmt.Sprintf(`[{"taskCode":%d,"x":100,"y":100}]`, taskDef["code"])

	// 创建工作流定义
	createReq := &CreateProcessDefinitionRequest{
		Name:               req.Name,
		Description:        req.Description,
		Locations:          locations,
		TaskDefinitionJson: string(taskDefJSON),
		TaskRelationJson:   string(taskRelationJSON),
		Timeout:            0,
		GlobalParams:       "[]",
		ExecutionType:      "PARALLEL",
	}

	pd, err := a.client.CreateProcessDefinition(ctx, a.projectCode, createReq)
	if err != nil {
		return nil, fmt.Errorf("failed to create process definition: %w", err)
	}

	// 创建工作流对象
	now := time.Now()
	workflow := &types.Workflow{
		ID:           req.ID,
		Name:         req.Name,
		Description:  req.Description,
		Status:       types.WorkflowStatusInactive,
		Schedule:     req.Schedule,
		Config:       req.Config,
		DataSourceID: req.DataSourceID,
		ExternalID:   strconv.FormatInt(pd.Code, 10),
		Properties:   req.Properties,
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	// 保存映射
	a.mu.Lock()
	a.workflows[req.ID] = &workflowMapping{
		InternalID:   req.ID,
		ExternalCode: pd.Code,
		Workflow:     workflow,
	}
	a.mu.Unlock()

	a.log.WithContext(ctx).Infof("Workflow created in DolphinScheduler: %s -> %d", req.ID, pd.Code)
	return workflow, nil
}

// buildTaskDefinition 构建任务定义
func (a *DolphinSchedulerAdapter) buildTaskDefinition(req *types.CreateWorkflowRequest) map[string]interface{} {
	taskCode := time.Now().UnixNano()

	// 构建任务参数
	taskParams := map[string]interface{}{
		"resourceList":         []interface{}{},
		"localParams":          []interface{}{},
		"rawScript":            a.buildCollectionScript(req),
		"dependence":           map[string]interface{}{},
		"conditionResult":      map[string]interface{}{"successNode": []string{}, "failedNode": []string{}},
		"waitStartTimeout":     map[string]interface{}{},
	}

	return map[string]interface{}{
		"code":              taskCode,
		"name":              req.Name,
		"description":       req.Description,
		"taskType":          "SHELL",
		"taskParams":        taskParams,
		"flag":              "YES",
		"taskPriority":      "MEDIUM",
		"workerGroup":       "default",
		"failRetryTimes":    3,
		"failRetryInterval": 1,
		"timeoutFlag":       "CLOSE",
		"timeout":           0,
		"delayTime":         0,
		"environmentCode":   -1,
	}
}

// buildCollectionScript 构建采集脚本
func (a *DolphinSchedulerAdapter) buildCollectionScript(req *types.CreateWorkflowRequest) string {
	// 构建元数据采集命令
	script := fmt.Sprintf(`#!/bin/bash
# Metadata Collection Task: %s
# DataSource: %s
# Task Type: %s

echo "Starting metadata collection..."
echo "Task ID: %s"
echo "DataSource ID: %s"

# 调用元数据采集API
curl -X POST "http://localhost:8000/api/v1/tasks/%s/execute" \
  -H "Content-Type: application/json" \
  -d '{"trigger": "dolphinscheduler"}'

echo "Metadata collection completed."
`, req.Name, req.DataSourceID, req.TaskType, req.ID, req.DataSourceID, req.ID)

	return script
}


// UpdateWorkflow 更新工作流
func (a *DolphinSchedulerAdapter) UpdateWorkflow(ctx context.Context, req *types.UpdateWorkflowRequest) (*types.Workflow, error) {
	a.mu.RLock()
	mapping, exists := a.workflows[req.ID]
	a.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("workflow %s not found", req.ID)
	}

	// 构建任务定义
	taskDef := map[string]interface{}{
		"code":              time.Now().UnixNano(),
		"name":              req.Name,
		"description":       req.Description,
		"taskType":          "SHELL",
		"taskParams":        map[string]interface{}{},
		"flag":              "YES",
		"taskPriority":      "MEDIUM",
		"workerGroup":       "default",
		"failRetryTimes":    3,
		"failRetryInterval": 1,
		"timeoutFlag":       "CLOSE",
		"timeout":           0,
	}
	taskDefJSON, _ := json.Marshal([]interface{}{taskDef})

	taskRelation := map[string]interface{}{
		"name":            "",
		"preTaskCode":     0,
		"postTaskCode":    taskDef["code"],
		"conditionType":   "NONE",
		"conditionParams": map[string]interface{}{},
	}
	taskRelationJSON, _ := json.Marshal([]interface{}{taskRelation})

	locations := fmt.Sprintf(`[{"taskCode":%d,"x":100,"y":100}]`, taskDef["code"])

	updateReq := &CreateProcessDefinitionRequest{
		Name:               req.Name,
		Description:        req.Description,
		Locations:          locations,
		TaskDefinitionJson: string(taskDefJSON),
		TaskRelationJson:   string(taskRelationJSON),
		Timeout:            0,
		GlobalParams:       "[]",
		ExecutionType:      "PARALLEL",
	}

	_, err := a.client.UpdateProcessDefinition(ctx, a.projectCode, mapping.ExternalCode, updateReq)
	if err != nil {
		return nil, fmt.Errorf("failed to update process definition: %w", err)
	}

	// 更新本地映射
	a.mu.Lock()
	mapping.Workflow.Name = req.Name
	mapping.Workflow.Description = req.Description
	mapping.Workflow.Schedule = req.Schedule
	mapping.Workflow.Config = req.Config
	mapping.Workflow.Properties = req.Properties
	mapping.Workflow.UpdatedAt = time.Now()
	a.mu.Unlock()

	a.log.WithContext(ctx).Infof("Workflow updated in DolphinScheduler: %s", req.ID)
	return mapping.Workflow, nil
}

// DeleteWorkflow 删除工作流
func (a *DolphinSchedulerAdapter) DeleteWorkflow(ctx context.Context, id string) error {
	a.mu.Lock()
	mapping, exists := a.workflows[id]
	if !exists {
		a.mu.Unlock()
		return nil
	}
	delete(a.workflows, id)
	a.mu.Unlock()

	// 先下线调度
	if mapping.ScheduleID > 0 {
		_ = a.client.OfflineSchedule(ctx, a.projectCode, mapping.ScheduleID)
		_ = a.client.DeleteSchedule(ctx, a.projectCode, mapping.ScheduleID)
	}

	// 下线工作流
	_ = a.client.ReleaseProcessDefinition(ctx, a.projectCode, mapping.ExternalCode, "OFFLINE")

	// 删除工作流
	if err := a.client.DeleteProcessDefinition(ctx, a.projectCode, mapping.ExternalCode); err != nil {
		a.log.WithContext(ctx).Warnf("Failed to delete process definition: %v", err)
	}

	a.log.WithContext(ctx).Infof("Workflow deleted from DolphinScheduler: %s", id)
	return nil
}

// GetWorkflow 获取工作流
func (a *DolphinSchedulerAdapter) GetWorkflow(ctx context.Context, id string) (*types.Workflow, error) {
	a.mu.RLock()
	mapping, exists := a.workflows[id]
	a.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("workflow %s not found", id)
	}

	return mapping.Workflow, nil
}

// TriggerWorkflow 触发工作流执行
func (a *DolphinSchedulerAdapter) TriggerWorkflow(ctx context.Context, id string, params map[string]interface{}) (*types.WorkflowExecution, error) {
	a.mu.RLock()
	mapping, exists := a.workflows[id]
	a.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("workflow %s not found", id)
	}

	// 转换参数
	startParams := make(map[string]string)
	for k, v := range params {
		startParams[k] = fmt.Sprintf("%v", v)
	}

	// 启动工作流实例
	req := &StartProcessInstanceRequest{
		ProcessDefinitionCode: mapping.ExternalCode,
		FailureStrategy:       "END",
		WarningType:           "NONE",
		StartParams:           startParams,
		DryRun:                0,
	}

	pi, err := a.client.StartProcessInstance(ctx, a.projectCode, req)
	if err != nil {
		return nil, fmt.Errorf("failed to start process instance: %w", err)
	}

	// 创建执行记录
	execution := &types.WorkflowExecution{
		ID:         uuid.New().String(),
		WorkflowID: id,
		Status:     types.ExecutionStatusRunning,
		StartTime:  pi.StartTime,
		ExternalID: strconv.FormatInt(pi.ID, 10),
		Result:     params,
	}

	a.log.WithContext(ctx).Infof("Workflow triggered in DolphinScheduler: %s -> instance %d", id, pi.ID)
	return execution, nil
}

// StopWorkflowExecution 停止工作流执行
func (a *DolphinSchedulerAdapter) StopWorkflowExecution(ctx context.Context, executionID string) error {
	// 解析执行ID获取实例ID
	instanceID, err := strconv.ParseInt(executionID, 10, 64)
	if err != nil {
		return fmt.Errorf("invalid execution ID: %s", executionID)
	}

	if err := a.client.StopProcessInstance(ctx, a.projectCode, instanceID); err != nil {
		return fmt.Errorf("failed to stop process instance: %w", err)
	}

	a.log.WithContext(ctx).Infof("Workflow execution stopped in DolphinScheduler: %s", executionID)
	return nil
}

// PauseWorkflow 暂停工作流
func (a *DolphinSchedulerAdapter) PauseWorkflow(ctx context.Context, id string) error {
	a.mu.Lock()
	mapping, exists := a.workflows[id]
	if !exists {
		a.mu.Unlock()
		return fmt.Errorf("workflow %s not found", id)
	}

	// 下线调度
	if mapping.ScheduleID > 0 {
		if err := a.client.OfflineSchedule(ctx, a.projectCode, mapping.ScheduleID); err != nil {
			a.mu.Unlock()
			return fmt.Errorf("failed to offline schedule: %w", err)
		}
	}

	mapping.Workflow.Status = types.WorkflowStatusPaused
	mapping.Workflow.UpdatedAt = time.Now()
	a.mu.Unlock()

	a.log.WithContext(ctx).Infof("Workflow paused in DolphinScheduler: %s", id)
	return nil
}

// ResumeWorkflow 恢复工作流
func (a *DolphinSchedulerAdapter) ResumeWorkflow(ctx context.Context, id string) error {
	a.mu.Lock()
	mapping, exists := a.workflows[id]
	if !exists {
		a.mu.Unlock()
		return fmt.Errorf("workflow %s not found", id)
	}

	// 上线调度
	if mapping.ScheduleID > 0 {
		if err := a.client.OnlineSchedule(ctx, a.projectCode, mapping.ScheduleID); err != nil {
			a.mu.Unlock()
			return fmt.Errorf("failed to online schedule: %w", err)
		}
	}

	mapping.Workflow.Status = types.WorkflowStatusActive
	mapping.Workflow.UpdatedAt = time.Now()
	a.mu.Unlock()

	a.log.WithContext(ctx).Infof("Workflow resumed in DolphinScheduler: %s", id)
	return nil
}

// GetWorkflowStatus 获取工作流状态
func (a *DolphinSchedulerAdapter) GetWorkflowStatus(ctx context.Context, id string) (*types.WorkflowStatus, error) {
	a.mu.RLock()
	mapping, exists := a.workflows[id]
	a.mu.RUnlock()

	if !exists {
		status := types.WorkflowStatusInactive
		return &status, nil
	}

	return &mapping.Workflow.Status, nil
}

// GetExecutionLogs 获取执行日志
func (a *DolphinSchedulerAdapter) GetExecutionLogs(ctx context.Context, executionID string) (*types.ExecutionLogs, error) {
	// 解析执行ID
	instanceID, err := strconv.ParseInt(executionID, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid execution ID: %s", executionID)
	}

	// 获取工作流实例
	pi, err := a.client.GetProcessInstance(ctx, a.projectCode, instanceID)
	if err != nil {
		return nil, fmt.Errorf("failed to get process instance: %w", err)
	}

	logs := &types.ExecutionLogs{
		ExecutionID: executionID,
		Logs: []types.LogEntry{
			{
				Timestamp: pi.StartTime,
				Level:     "INFO",
				Message:   fmt.Sprintf("Process instance started: %s", pi.Name),
			},
			{
				Timestamp: time.Now(),
				Level:     "INFO",
				Message:   fmt.Sprintf("Process instance state: %s", pi.State),
			},
		},
	}

	if pi.EndTime != nil {
		logs.Logs = append(logs.Logs, types.LogEntry{
			Timestamp: *pi.EndTime,
			Level:     "INFO",
			Message:   fmt.Sprintf("Process instance completed with state: %s", pi.State),
		})
	}

	return logs, nil
}

// StartWorkflow 启动工作流调度
func (a *DolphinSchedulerAdapter) StartWorkflow(ctx context.Context, id string) error {
	a.mu.Lock()
	mapping, exists := a.workflows[id]
	if !exists {
		a.mu.Unlock()
		return fmt.Errorf("workflow %s not found", id)
	}

	// 上线工作流定义
	if err := a.client.ReleaseProcessDefinition(ctx, a.projectCode, mapping.ExternalCode, "ONLINE"); err != nil {
		a.mu.Unlock()
		return fmt.Errorf("failed to release process definition: %w", err)
	}

	// 如果有调度配置，创建调度
	if mapping.Workflow.Schedule != nil && mapping.ScheduleID == 0 {
		scheduleJSON := a.buildScheduleJSON(mapping.Workflow.Schedule)
		scheduleReq := &CreateScheduleRequest{
			ProcessDefinitionCode: mapping.ExternalCode,
			Schedule:              scheduleJSON,
			WarningType:           "NONE",
			FailureStrategy:       "END",
		}

		schedule, err := a.client.CreateSchedule(ctx, a.projectCode, scheduleReq)
		if err != nil {
			a.log.WithContext(ctx).Warnf("Failed to create schedule: %v", err)
		} else {
			mapping.ScheduleID = schedule.ID
			// 上线调度
			if err := a.client.OnlineSchedule(ctx, a.projectCode, schedule.ID); err != nil {
				a.log.WithContext(ctx).Warnf("Failed to online schedule: %v", err)
			}
		}
	} else if mapping.ScheduleID > 0 {
		// 上线已有调度
		if err := a.client.OnlineSchedule(ctx, a.projectCode, mapping.ScheduleID); err != nil {
			a.log.WithContext(ctx).Warnf("Failed to online schedule: %v", err)
		}
	}

	mapping.Workflow.Status = types.WorkflowStatusActive
	mapping.Workflow.UpdatedAt = time.Now()
	a.mu.Unlock()

	a.log.WithContext(ctx).Infof("Workflow started in DolphinScheduler: %s", id)
	return nil
}

// buildScheduleJSON 构建调度JSON
func (a *DolphinSchedulerAdapter) buildScheduleJSON(schedule *biz.ScheduleConfig) string {
	now := time.Now()
	startTime := now.Format("2006-01-02 00:00:00")
	endTime := now.AddDate(10, 0, 0).Format("2006-01-02 00:00:00") // 10年后

	if schedule.StartTime != nil {
		startTime = schedule.StartTime.Format("2006-01-02 15:04:05")
	}
	if schedule.EndTime != nil {
		endTime = schedule.EndTime.Format("2006-01-02 15:04:05")
	}

	crontab := "0 0 0 * * ?" // 默认每天0点
	switch schedule.Type {
	case biz.ScheduleTypeCron:
		crontab = schedule.CronExpr
	case biz.ScheduleTypeInterval:
		// 转换间隔为cron表达式
		if schedule.Interval >= 86400 {
			crontab = "0 0 0 * * ?"
		} else if schedule.Interval >= 3600 {
			crontab = fmt.Sprintf("0 0 */%d * * ?", schedule.Interval/3600)
		} else if schedule.Interval >= 60 {
			crontab = fmt.Sprintf("0 */%d * * * ?", schedule.Interval/60)
		} else {
			crontab = fmt.Sprintf("*/%d * * * * ?", schedule.Interval)
		}
	}

	timezone := schedule.Timezone
	if timezone == "" {
		timezone = "Asia/Shanghai"
	}

	return fmt.Sprintf(`{"startTime":"%s","endTime":"%s","crontab":"%s","timezoneId":"%s"}`,
		startTime, endTime, crontab, timezone)
}

// StopWorkflow 停止工作流调度
func (a *DolphinSchedulerAdapter) StopWorkflow(ctx context.Context, id string) error {
	a.mu.Lock()
	mapping, exists := a.workflows[id]
	if !exists {
		a.mu.Unlock()
		return fmt.Errorf("workflow %s not found", id)
	}

	// 下线调度
	if mapping.ScheduleID > 0 {
		if err := a.client.OfflineSchedule(ctx, a.projectCode, mapping.ScheduleID); err != nil {
			a.log.WithContext(ctx).Warnf("Failed to offline schedule: %v", err)
		}
	}

	// 下线工作流定义
	if err := a.client.ReleaseProcessDefinition(ctx, a.projectCode, mapping.ExternalCode, "OFFLINE"); err != nil {
		a.mu.Unlock()
		return fmt.Errorf("failed to offline process definition: %w", err)
	}

	mapping.Workflow.Status = types.WorkflowStatusInactive
	mapping.Workflow.UpdatedAt = time.Now()
	a.mu.Unlock()

	a.log.WithContext(ctx).Infof("Workflow stopped in DolphinScheduler: %s", id)
	return nil
}
