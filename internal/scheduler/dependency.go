package scheduler

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/go-kratos/kratos/v2/log"

	"go-metadata/internal/scheduler/types"
)

// DependencyType 依赖类型
type DependencyType string

const (
	DependencyTypeWorkflow DependencyType = "workflow" // 工作流依赖
	DependencyTypeTime     DependencyType = "time"     // 时间依赖
)

// Dependency 依赖定义
type Dependency struct {
	Type       DependencyType       `json:"type"`
	WorkflowID string               `json:"workflow_id,omitempty"` // 依赖的工作流ID
	Status     types.ExecutionStatus `json:"status,omitempty"`      // 期望的执行状态
	TimeWindow *TimeWindow          `json:"time_window,omitempty"` // 时间窗口
}

// TimeWindow 时间窗口
type TimeWindow struct {
	StartHour   int `json:"start_hour"`   // 开始小时 (0-23)
	EndHour     int `json:"end_hour"`     // 结束小时 (0-23)
	DaysOfWeek  []int `json:"days_of_week"` // 星期几 (0=周日, 1=周一, ..., 6=周六)
}

// DependencyManager 依赖管理器
type DependencyManager struct {
	dependencies map[string][]*Dependency // workflowID -> dependencies
	scheduler    *BuiltinScheduler
	mu           sync.RWMutex
	log          *log.Helper
}

// NewDependencyManager 创建依赖管理器
func NewDependencyManager(scheduler *BuiltinScheduler, logger log.Logger) *DependencyManager {
	return &DependencyManager{
		dependencies: make(map[string][]*Dependency),
		scheduler:    scheduler,
		log:          log.NewHelper(logger),
	}
}

// AddDependency 添加依赖
func (m *DependencyManager) AddDependency(workflowID string, dep *Dependency) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// 验证依赖
	if err := m.validateDependency(dep); err != nil {
		return err
	}

	// 检查循环依赖
	if dep.Type == DependencyTypeWorkflow {
		if m.hasCircularDependency(workflowID, dep.WorkflowID) {
			return fmt.Errorf("circular dependency detected")
		}
	}

	m.dependencies[workflowID] = append(m.dependencies[workflowID], dep)
	m.log.Infof("Dependency added for workflow %s: %+v", workflowID, dep)
	return nil
}

// RemoveDependency 移除依赖
func (m *DependencyManager) RemoveDependency(workflowID string, depIndex int) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	deps, exists := m.dependencies[workflowID]
	if !exists || depIndex >= len(deps) {
		return fmt.Errorf("dependency not found")
	}

	m.dependencies[workflowID] = append(deps[:depIndex], deps[depIndex+1:]...)
	m.log.Infof("Dependency removed for workflow %s at index %d", workflowID, depIndex)
	return nil
}

// GetDependencies 获取工作流的依赖
func (m *DependencyManager) GetDependencies(workflowID string) []*Dependency {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return m.dependencies[workflowID]
}

// ClearDependencies 清除工作流的所有依赖
func (m *DependencyManager) ClearDependencies(workflowID string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	delete(m.dependencies, workflowID)
	m.log.Infof("Dependencies cleared for workflow %s", workflowID)
}

// CheckDependencies 检查依赖是否满足
func (m *DependencyManager) CheckDependencies(ctx context.Context, workflowID string) (bool, string) {
	m.mu.RLock()
	deps := m.dependencies[workflowID]
	m.mu.RUnlock()

	if len(deps) == 0 {
		return true, ""
	}

	for _, dep := range deps {
		satisfied, reason := m.checkDependency(ctx, dep)
		if !satisfied {
			return false, reason
		}
	}

	return true, ""
}

// checkDependency 检查单个依赖
func (m *DependencyManager) checkDependency(ctx context.Context, dep *Dependency) (bool, string) {
	switch dep.Type {
	case DependencyTypeWorkflow:
		return m.checkWorkflowDependency(ctx, dep)
	case DependencyTypeTime:
		return m.checkTimeDependency(dep)
	default:
		return false, fmt.Sprintf("unknown dependency type: %s", dep.Type)
	}
}

// checkWorkflowDependency 检查工作流依赖
func (m *DependencyManager) checkWorkflowDependency(ctx context.Context, dep *Dependency) (bool, string) {
	if dep.WorkflowID == "" {
		return false, "workflow dependency requires workflow_id"
	}

	// 获取依赖工作流的最新执行记录
	executions, err := m.scheduler.ListExecutions(ctx, dep.WorkflowID, 1)
	if err != nil || len(executions) == 0 {
		return false, fmt.Sprintf("no execution found for workflow %s", dep.WorkflowID)
	}

	latestExec := executions[0]

	// 检查执行状态
	expectedStatus := dep.Status
	if expectedStatus == "" {
		expectedStatus = types.ExecutionStatusCompleted
	}

	if latestExec.Status != expectedStatus {
		return false, fmt.Sprintf("workflow %s latest execution status is %s, expected %s",
			dep.WorkflowID, latestExec.Status, expectedStatus)
	}

	// 检查执行时间是否在今天
	if latestExec.EndTime != nil {
		today := time.Now().Truncate(24 * time.Hour)
		if latestExec.EndTime.Before(today) {
			return false, fmt.Sprintf("workflow %s latest execution is not from today", dep.WorkflowID)
		}
	}

	return true, ""
}

// checkTimeDependency 检查时间依赖
func (m *DependencyManager) checkTimeDependency(dep *Dependency) (bool, string) {
	if dep.TimeWindow == nil {
		return false, "time dependency requires time_window"
	}

	now := time.Now()
	currentHour := now.Hour()
	currentDay := int(now.Weekday())

	// 检查时间窗口
	if currentHour < dep.TimeWindow.StartHour || currentHour >= dep.TimeWindow.EndHour {
		return false, fmt.Sprintf("current hour %d is outside time window [%d, %d)",
			currentHour, dep.TimeWindow.StartHour, dep.TimeWindow.EndHour)
	}

	// 检查星期
	if len(dep.TimeWindow.DaysOfWeek) > 0 {
		dayAllowed := false
		for _, day := range dep.TimeWindow.DaysOfWeek {
			if day == currentDay {
				dayAllowed = true
				break
			}
		}
		if !dayAllowed {
			return false, fmt.Sprintf("current day %d is not in allowed days", currentDay)
		}
	}

	return true, ""
}

// validateDependency 验证依赖配置
func (m *DependencyManager) validateDependency(dep *Dependency) error {
	switch dep.Type {
	case DependencyTypeWorkflow:
		if dep.WorkflowID == "" {
			return fmt.Errorf("workflow dependency requires workflow_id")
		}
	case DependencyTypeTime:
		if dep.TimeWindow == nil {
			return fmt.Errorf("time dependency requires time_window")
		}
		if dep.TimeWindow.StartHour < 0 || dep.TimeWindow.StartHour > 23 {
			return fmt.Errorf("invalid start_hour: %d", dep.TimeWindow.StartHour)
		}
		if dep.TimeWindow.EndHour < 0 || dep.TimeWindow.EndHour > 24 {
			return fmt.Errorf("invalid end_hour: %d", dep.TimeWindow.EndHour)
		}
		if dep.TimeWindow.StartHour >= dep.TimeWindow.EndHour {
			return fmt.Errorf("start_hour must be less than end_hour")
		}
	default:
		return fmt.Errorf("unknown dependency type: %s", dep.Type)
	}
	return nil
}

// hasCircularDependency 检查循环依赖
func (m *DependencyManager) hasCircularDependency(workflowID, depWorkflowID string) bool {
	visited := make(map[string]bool)
	return m.dfs(depWorkflowID, workflowID, visited)
}

// dfs 深度优先搜索检查循环依赖
func (m *DependencyManager) dfs(current, target string, visited map[string]bool) bool {
	if current == target {
		return true
	}

	if visited[current] {
		return false
	}
	visited[current] = true

	deps := m.dependencies[current]
	for _, dep := range deps {
		if dep.Type == DependencyTypeWorkflow {
			if m.dfs(dep.WorkflowID, target, visited) {
				return true
			}
		}
	}

	return false
}

// GetDependencyGraph 获取依赖图
func (m *DependencyManager) GetDependencyGraph() map[string][]string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	graph := make(map[string][]string)
	for workflowID, deps := range m.dependencies {
		var depIDs []string
		for _, dep := range deps {
			if dep.Type == DependencyTypeWorkflow {
				depIDs = append(depIDs, dep.WorkflowID)
			}
		}
		graph[workflowID] = depIDs
	}
	return graph
}
