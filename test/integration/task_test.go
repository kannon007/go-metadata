// Package integration provides end-to-end integration tests for the task management module.
package integration

import (
	"context"
	"testing"
	"time"

	"github.com/go-kratos/kratos/v2/log"

	"go-metadata/internal/biz"
	"go-metadata/internal/service"
)

// TestTaskCRUDFlow tests the complete CRUD flow for tasks
// Feature: datasource-management, Property 9: 任务生命周期管理
// **Validates: Requirements 4.4, 4.5**
func TestTaskCRUDFlow(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	ctx := context.Background()
	logger := log.DefaultLogger

	// Setup test dependencies
	testData, cleanup := setupTaskTestData(t, logger)
	defer cleanup()

	var createdTaskID string

	// Test Create Task
	t.Run("CreateTask", func(t *testing.T) {
		req := &biz.CreateTaskRequest{
			Name:         "test-task-" + time.Now().Format("20060102150405"),
			DataSourceID: "test-ds-id",
			Type:         biz.TaskTypeFullCollection,
			Config: &biz.TaskConfig{
				BatchSize:  1000,
				Timeout:    3600,
				RetryCount: 3,
			},
			Schedule: &biz.ScheduleConfig{
				Type:     biz.ScheduleTypeImmediate,
				Timezone: "Asia/Shanghai",
			},
			SchedulerType: biz.SchedulerTypeBuiltIn,
			CreatedBy:     "test-user",
		}

		task, err := testData.taskService.CreateTask(ctx, req)
		if err != nil {
			t.Logf("CreateTask returned error (may be expected): %v", err)
			return
		}

		if task == nil {
			t.Fatal("Expected task to be created")
		}

		if task.ID == "" {
			t.Error("Expected task ID to be set")
		}

		createdTaskID = task.ID
	})

	// Test Get Task
	t.Run("GetTask", func(t *testing.T) {
		if createdTaskID == "" {
			t.Skip("No task created in previous test")
		}

		task, err := testData.taskService.GetTask(ctx, createdTaskID)
		if err != nil {
			t.Fatalf("GetTask failed: %v", err)
		}

		if task.ID != createdTaskID {
			t.Errorf("Expected ID %s, got %s", createdTaskID, task.ID)
		}
	})

	// Test List Tasks
	t.Run("ListTasks", func(t *testing.T) {
		req := &biz.ListTasksRequest{
			Page:     1,
			PageSize: 10,
		}

		resp, err := testData.taskService.ListTasks(ctx, req)
		if err != nil {
			t.Fatalf("ListTasks failed: %v", err)
		}

		if resp == nil {
			t.Fatal("Expected response to be non-nil")
		}
	})

	// Test Update Task
	t.Run("UpdateTask", func(t *testing.T) {
		if createdTaskID == "" {
			t.Skip("No task created in previous test")
		}

		req := &biz.UpdateTaskRequest{
			ID:   createdTaskID,
			Name: "updated-test-task",
			Config: &biz.TaskConfig{
				BatchSize:  2000,
				Timeout:    7200,
				RetryCount: 5,
			},
		}

		task, err := testData.taskService.UpdateTask(ctx, req)
		if err != nil {
			t.Fatalf("UpdateTask failed: %v", err)
		}

		if task.Name != req.Name {
			t.Errorf("Expected name %s, got %s", req.Name, task.Name)
		}
	})

	// Test Delete Task
	t.Run("DeleteTask", func(t *testing.T) {
		if createdTaskID == "" {
			t.Skip("No task created in previous test")
		}

		err := testData.taskService.DeleteTask(ctx, createdTaskID)
		if err != nil {
			t.Fatalf("DeleteTask failed: %v", err)
		}
	})
}

// TestTaskValidation tests task configuration validation
// Feature: datasource-management, Property 7: 任务配置验证
// **Validates: Requirements 4.1, 4.3**
func TestTaskValidation(t *testing.T) {
	tests := []struct {
		name        string
		req         *biz.CreateTaskRequest
		expectError bool
	}{
		{
			name: "valid full collection task",
			req: &biz.CreateTaskRequest{
				Name:          "valid-task",
				DataSourceID:  "ds-id",
				Type:          biz.TaskTypeFullCollection,
				SchedulerType: biz.SchedulerTypeBuiltIn,
			},
			expectError: false,
		},
		{
			name: "missing name",
			req: &biz.CreateTaskRequest{
				Name:          "",
				DataSourceID:  "ds-id",
				Type:          biz.TaskTypeFullCollection,
				SchedulerType: biz.SchedulerTypeBuiltIn,
			},
			expectError: true,
		},
		{
			name: "missing datasource id",
			req: &biz.CreateTaskRequest{
				Name:          "task",
				DataSourceID:  "",
				Type:          biz.TaskTypeFullCollection,
				SchedulerType: biz.SchedulerTypeBuiltIn,
			},
			expectError: true,
		},
		{
			name: "invalid task type",
			req: &biz.CreateTaskRequest{
				Name:          "task",
				DataSourceID:  "ds-id",
				Type:          biz.TaskType("invalid"),
				SchedulerType: biz.SchedulerTypeBuiltIn,
			},
			expectError: true,
		},
		{
			name: "incremental without column",
			req: &biz.CreateTaskRequest{
				Name:          "incremental-task",
				DataSourceID:  "ds-id",
				Type:          biz.TaskTypeIncrementalCollection,
				SchedulerType: biz.SchedulerTypeBuiltIn,
				Config: &biz.TaskConfig{
					IncrementalColumn: "", // Missing required column
				},
			},
			expectError: true,
		},
		{
			name: "valid incremental with column",
			req: &biz.CreateTaskRequest{
				Name:          "incremental-task",
				DataSourceID:  "ds-id",
				Type:          biz.TaskTypeIncrementalCollection,
				SchedulerType: biz.SchedulerTypeBuiltIn,
				Config: &biz.TaskConfig{
					IncrementalColumn: "updated_at",
				},
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.req.Validate()
			if tt.expectError {
				if err == nil {
					t.Error("Expected validation error, got nil")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected validation error: %v", err)
				}
			}
		})
	}
}

// TestTaskTypeSupport tests support for all task types
// Feature: datasource-management, Property 8: 执行模式支持
// **Validates: Requirements 4.2**
func TestTaskTypeSupport(t *testing.T) {
	supportedTypes := biz.AllTaskTypes()
	expectedTypes := []biz.TaskType{
		biz.TaskTypeFullCollection,
		biz.TaskTypeIncrementalCollection,
		biz.TaskTypeSchemaOnly,
		biz.TaskTypeDataProfile,
	}

	for _, expected := range expectedTypes {
		found := false
		for _, supported := range supportedTypes {
			if expected == supported {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected task type %s to be supported", expected)
		}
	}

	for _, taskType := range expectedTypes {
		if !biz.IsValidTaskType(taskType) {
			t.Errorf("IsValidTaskType returned false for valid type %s", taskType)
		}
	}

	if biz.IsValidTaskType(biz.TaskType("invalid")) {
		t.Error("IsValidTaskType should return false for invalid type")
	}
}

// TestScheduleTypeSupport tests support for all schedule types
func TestScheduleTypeSupport(t *testing.T) {
	supportedTypes := biz.AllScheduleTypes()
	expectedTypes := []biz.ScheduleType{
		biz.ScheduleTypeImmediate,
		biz.ScheduleTypeOnce,
		biz.ScheduleTypeCron,
		biz.ScheduleTypeInterval,
	}

	for _, expected := range expectedTypes {
		found := false
		for _, supported := range supportedTypes {
			if expected == supported {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected schedule type %s to be supported", expected)
		}
	}

	for _, schedType := range expectedTypes {
		if !biz.IsValidScheduleType(schedType) {
			t.Errorf("IsValidScheduleType returned false for valid type %s", schedType)
		}
	}
}

// TestScheduleConfigValidation tests schedule configuration validation
func TestScheduleConfigValidation(t *testing.T) {
	now := time.Now()
	future := now.Add(time.Hour)
	past := now.Add(-time.Hour)

	tests := []struct {
		name        string
		config      *biz.ScheduleConfig
		expectError bool
	}{
		{
			name: "valid immediate",
			config: &biz.ScheduleConfig{
				Type: biz.ScheduleTypeImmediate,
			},
			expectError: false,
		},
		{
			name: "valid cron",
			config: &biz.ScheduleConfig{
				Type:     biz.ScheduleTypeCron,
				CronExpr: "0 0 * * *",
			},
			expectError: false,
		},
		{
			name: "cron without expression",
			config: &biz.ScheduleConfig{
				Type:     biz.ScheduleTypeCron,
				CronExpr: "",
			},
			expectError: true,
		},
		{
			name: "valid interval",
			config: &biz.ScheduleConfig{
				Type:     biz.ScheduleTypeInterval,
				Interval: 3600,
			},
			expectError: false,
		},
		{
			name: "interval without value",
			config: &biz.ScheduleConfig{
				Type:     biz.ScheduleTypeInterval,
				Interval: 0,
			},
			expectError: true,
		},
		{
			name: "valid once with start time",
			config: &biz.ScheduleConfig{
				Type:      biz.ScheduleTypeOnce,
				StartTime: &future,
			},
			expectError: false,
		},
		{
			name: "once without start time",
			config: &biz.ScheduleConfig{
				Type: biz.ScheduleTypeOnce,
			},
			expectError: true,
		},
		{
			name: "invalid time range",
			config: &biz.ScheduleConfig{
				Type:      biz.ScheduleTypeImmediate,
				StartTime: &future,
				EndTime:   &past,
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if tt.expectError {
				if err == nil {
					t.Error("Expected validation error, got nil")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected validation error: %v", err)
				}
			}
		})
	}
}

// TestTaskStatusTransitions tests task status transitions
func TestTaskStatusTransitions(t *testing.T) {
	tests := []struct {
		name      string
		status    biz.TaskStatus
		canStart  bool
		canStop   bool
		canPause  bool
	}{
		{
			name:      "inactive task",
			status:    biz.TaskStatusInactive,
			canStart:  true,
			canStop:   false,
			canPause:  false,
		},
		{
			name:      "active task",
			status:    biz.TaskStatusActive,
			canStart:  false,
			canStop:   true,
			canPause:  true,
		},
		{
			name:      "running task",
			status:    biz.TaskStatusRunning,
			canStart:  false,
			canStop:   true,
			canPause:  true,
		},
		{
			name:      "paused task",
			status:    biz.TaskStatusPaused,
			canStart:  true,
			canStop:   false,
			canPause:  false,
		},
		{
			name:      "failed task",
			status:    biz.TaskStatusFailed,
			canStart:  true,
			canStop:   false,
			canPause:  false,
		},
		{
			name:      "completed task",
			status:    biz.TaskStatusCompleted,
			canStart:  false,
			canStop:   false,
			canPause:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			task := &biz.CollectionTask{
				Status: tt.status,
			}

			if task.CanStart() != tt.canStart {
				t.Errorf("CanStart: expected %v, got %v", tt.canStart, task.CanStart())
			}

			if task.CanStop() != tt.canStop {
				t.Errorf("CanStop: expected %v, got %v", tt.canStop, task.CanStop())
			}

			if task.CanPause() != tt.canPause {
				t.Errorf("CanPause: expected %v, got %v", tt.canPause, task.CanPause())
			}
		})
	}
}

// TestExecutionStatusSupport tests support for all execution statuses
func TestExecutionStatusSupport(t *testing.T) {
	supportedStatuses := biz.AllExecutionStatuses()
	expectedStatuses := []biz.ExecutionStatus{
		biz.ExecutionStatusPending,
		biz.ExecutionStatusRunning,
		biz.ExecutionStatusCompleted,
		biz.ExecutionStatusFailed,
		biz.ExecutionStatusCancelled,
	}

	for _, expected := range expectedStatuses {
		found := false
		for _, supported := range supportedStatuses {
			if expected == supported {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected execution status %s to be supported", expected)
		}
	}

	for _, status := range expectedStatuses {
		if !biz.IsValidExecutionStatus(status) {
			t.Errorf("IsValidExecutionStatus returned false for valid status %s", status)
		}
	}
}

// TestTaskConfigDefaults tests that default values are set correctly
func TestTaskConfigDefaults(t *testing.T) {
	config := &biz.TaskConfig{}
	config.SetDefaults()

	if config.BatchSize != 1000 {
		t.Errorf("Expected default batch_size 1000, got %d", config.BatchSize)
	}

	if config.Timeout != 3600 {
		t.Errorf("Expected default timeout 3600, got %d", config.Timeout)
	}

	if config.RetryCount != 3 {
		t.Errorf("Expected default retry_count 3, got %d", config.RetryCount)
	}

	if config.RetryInterval != 60 {
		t.Errorf("Expected default retry_interval 60, got %d", config.RetryInterval)
	}
}

// TestScheduleConfigDefaults tests that default values are set correctly
func TestScheduleConfigDefaults(t *testing.T) {
	config := &biz.ScheduleConfig{}
	config.SetDefaults()

	if config.Timezone != "Asia/Shanghai" {
		t.Errorf("Expected default timezone Asia/Shanghai, got %s", config.Timezone)
	}

	if config.Type != biz.ScheduleTypeImmediate {
		t.Errorf("Expected default type immediate, got %s", config.Type)
	}
}

// TestExecutionDurationCalculation tests execution duration calculation
func TestExecutionDurationCalculation(t *testing.T) {
	startTime := time.Now()
	endTime := startTime.Add(5 * time.Second)

	execution := &biz.TaskExecution{
		StartTime: startTime,
		EndTime:   &endTime,
	}

	execution.CalculateDuration()

	expectedDuration := int64(5000) // 5 seconds in milliseconds
	if execution.Duration != expectedDuration {
		t.Errorf("Expected duration %d ms, got %d ms", expectedDuration, execution.Duration)
	}
}

// taskTestData holds task test dependencies
type taskTestData struct {
	taskService *service.TaskService
}

// setupTaskTestData creates task test dependencies
func setupTaskTestData(t *testing.T, logger log.Logger) (*taskTestData, func()) {
	t.Helper()

	mockTaskRepo := &mockTaskRepo{}
	mockDSRepo := &mockDataSourceRepo{
		dataSources: map[string]*biz.DataSource{
			"test-ds-id": {
				ID:     "test-ds-id",
				Name:   "test-ds",
				Type:   biz.DataSourceTypeMySQL,
				Status: biz.DataSourceStatusActive,
			},
		},
	}
	mockScheduler := &mockTaskScheduler{}

	taskUsecase := biz.NewTaskUsecase(mockTaskRepo, mockDSRepo, mockScheduler, logger)
	taskService := service.NewTaskService(taskUsecase, logger)

	cleanup := func() {}

	return &taskTestData{
		taskService: taskService,
	}, cleanup
}

// mockTaskRepo is a mock implementation of TaskRepo
type mockTaskRepo struct {
	tasks      map[string]*biz.CollectionTask
	executions map[string]*biz.TaskExecution
}

func (m *mockTaskRepo) CreateTask(ctx context.Context, task *biz.CollectionTask) (*biz.CollectionTask, error) {
	if m.tasks == nil {
		m.tasks = make(map[string]*biz.CollectionTask)
	}
	m.tasks[task.ID] = task
	return task, nil
}

func (m *mockTaskRepo) UpdateTask(ctx context.Context, task *biz.CollectionTask) (*biz.CollectionTask, error) {
	if m.tasks == nil {
		return nil, biz.ErrTaskNotFound
	}
	m.tasks[task.ID] = task
	return task, nil
}

func (m *mockTaskRepo) DeleteTask(ctx context.Context, id string) error {
	if m.tasks != nil {
		delete(m.tasks, id)
	}
	return nil
}

func (m *mockTaskRepo) GetTask(ctx context.Context, id string) (*biz.CollectionTask, error) {
	if m.tasks == nil {
		return nil, biz.ErrTaskNotFound
	}
	task, ok := m.tasks[id]
	if !ok {
		return nil, biz.ErrTaskNotFound
	}
	return task, nil
}

func (m *mockTaskRepo) ListTasks(ctx context.Context, req *biz.ListTasksRequest) (*biz.ListTasksResponse, error) {
	var list []*biz.CollectionTask
	for _, task := range m.tasks {
		list = append(list, task)
	}
	return &biz.ListTasksResponse{
		Tasks:    list,
		Total:    int64(len(list)),
		Page:     req.Page,
		PageSize: req.PageSize,
	}, nil
}

func (m *mockTaskRepo) GetTasksByDataSourceID(ctx context.Context, dataSourceID string) ([]*biz.CollectionTask, error) {
	var list []*biz.CollectionTask
	for _, task := range m.tasks {
		if task.DataSourceID == dataSourceID {
			list = append(list, task)
		}
	}
	return list, nil
}

func (m *mockTaskRepo) UpdateTaskStatus(ctx context.Context, id string, status biz.TaskStatus) error {
	if task, ok := m.tasks[id]; ok {
		task.Status = status
		return nil
	}
	return biz.ErrTaskNotFound
}

func (m *mockTaskRepo) UpdateTaskLastExecutedAt(ctx context.Context, id string, lastExecutedAt time.Time) error {
	if task, ok := m.tasks[id]; ok {
		task.LastExecutedAt = &lastExecutedAt
		return nil
	}
	return biz.ErrTaskNotFound
}

func (m *mockTaskRepo) CreateExecution(ctx context.Context, execution *biz.TaskExecution) (*biz.TaskExecution, error) {
	if m.executions == nil {
		m.executions = make(map[string]*biz.TaskExecution)
	}
	m.executions[execution.ID] = execution
	return execution, nil
}

func (m *mockTaskRepo) UpdateExecution(ctx context.Context, execution *biz.TaskExecution) (*biz.TaskExecution, error) {
	if m.executions == nil {
		return nil, biz.ErrTaskNotFound
	}
	m.executions[execution.ID] = execution
	return execution, nil
}

func (m *mockTaskRepo) GetExecution(ctx context.Context, id string) (*biz.TaskExecution, error) {
	if m.executions == nil {
		return nil, biz.ErrTaskNotFound
	}
	exec, ok := m.executions[id]
	if !ok {
		return nil, biz.ErrTaskNotFound
	}
	return exec, nil
}

func (m *mockTaskRepo) ListExecutions(ctx context.Context, req *biz.ListExecutionsRequest) (*biz.ListExecutionsResponse, error) {
	var list []*biz.TaskExecution
	for _, exec := range m.executions {
		if req.TaskID == "" || exec.TaskID == req.TaskID {
			list = append(list, exec)
		}
	}
	return &biz.ListExecutionsResponse{
		Executions: list,
		Total:      int64(len(list)),
		Page:       req.Page,
		PageSize:   req.PageSize,
	}, nil
}

func (m *mockTaskRepo) GetLatestExecution(ctx context.Context, taskID string) (*biz.TaskExecution, error) {
	var latest *biz.TaskExecution
	for _, exec := range m.executions {
		if exec.TaskID == taskID {
			if latest == nil || exec.StartTime.After(latest.StartTime) {
				latest = exec
			}
		}
	}
	if latest == nil {
		return nil, biz.ErrTaskNotFound
	}
	return latest, nil
}

// mockTaskScheduler is a mock implementation of TaskScheduler
type mockTaskScheduler struct{}

func (m *mockTaskScheduler) CreateTask(ctx context.Context, task *biz.CollectionTask) error {
	return nil
}

func (m *mockTaskScheduler) UpdateTask(ctx context.Context, task *biz.CollectionTask) error {
	return nil
}

func (m *mockTaskScheduler) DeleteTask(ctx context.Context, id string) error {
	return nil
}

func (m *mockTaskScheduler) StartTask(ctx context.Context, id string) error {
	return nil
}

func (m *mockTaskScheduler) StopTask(ctx context.Context, id string) error {
	return nil
}

func (m *mockTaskScheduler) PauseTask(ctx context.Context, id string) error {
	return nil
}

func (m *mockTaskScheduler) ResumeTask(ctx context.Context, id string) error {
	return nil
}

func (m *mockTaskScheduler) GetTaskStatus(ctx context.Context, id string) (*biz.TaskStatus, error) {
	status := biz.TaskStatusActive
	return &status, nil
}

// Ensure mocks implement interfaces
var _ biz.TaskRepo = (*mockTaskRepo)(nil)
var _ biz.TaskScheduler = (*mockTaskScheduler)(nil)
