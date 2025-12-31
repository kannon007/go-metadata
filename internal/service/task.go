package service

import (
	"context"

	v1 "go-metadata/api/metadata/v1"
	"go-metadata/internal/biz"

	"github.com/go-kratos/kratos/v2/log"
	"google.golang.org/protobuf/types/known/emptypb"
)

// TaskService implements the TaskService gRPC/HTTP service.
type TaskService struct {
	v1.UnimplementedTaskServiceServer

	uc  *biz.TaskUsecase
	log *log.Helper
}

// NewTaskService creates a new TaskService.
func NewTaskService(uc *biz.TaskUsecase, logger log.Logger) *TaskService {
	return &TaskService{
		uc:  uc,
		log: log.NewHelper(logger),
	}
}

// CreateTask creates a new task.
func (s *TaskService) CreateTask(ctx context.Context, req *v1.CreateTaskRequest) (*v1.CollectionTask, error) {
	task, err := s.uc.Create(ctx, &biz.CollectionTask{
		Name:          req.Name,
		DataSourceID:  req.DatasourceId,
		Type:          req.Type.String(),
		Config:        toBizTaskConfig(req.Config),
		Schedule:      toBizScheduleConfig(req.Schedule),
		SchedulerType: req.SchedulerType.String(),
	})
	if err != nil {
		return nil, err
	}
	return toProtoTask(task), nil
}

// UpdateTask updates an existing task.
func (s *TaskService) UpdateTask(ctx context.Context, req *v1.UpdateTaskRequest) (*v1.CollectionTask, error) {
	task, err := s.uc.Update(ctx, &biz.CollectionTask{
		ID:       req.Id,
		Name:     req.Name,
		Config:   toBizTaskConfig(req.Config),
		Schedule: toBizScheduleConfig(req.Schedule),
	})
	if err != nil {
		return nil, err
	}
	return toProtoTask(task), nil
}

// DeleteTask deletes a task.
func (s *TaskService) DeleteTask(ctx context.Context, req *v1.DeleteTaskRequest) (*emptypb.Empty, error) {
	if err := s.uc.Delete(ctx, req.Id); err != nil {
		return nil, err
	}
	return &emptypb.Empty{}, nil
}

// GetTask gets a task by ID.
func (s *TaskService) GetTask(ctx context.Context, req *v1.GetTaskRequest) (*v1.CollectionTask, error) {
	task, err := s.uc.Get(ctx, req.Id)
	if err != nil {
		return nil, err
	}
	return toProtoTask(task), nil
}

// ListTasks lists tasks.
func (s *TaskService) ListTasks(ctx context.Context, req *v1.ListTasksRequest) (*v1.ListTasksResponse, error) {
	list, total, err := s.uc.List(ctx, int(req.Page), int(req.PageSize))
	if err != nil {
		return nil, err
	}
	return &v1.ListTasksResponse{
		Tasks:    toProtoTasks(list),
		Total:    total,
		Page:     req.Page,
		PageSize: req.PageSize,
	}, nil
}

// StartTask starts a task.
func (s *TaskService) StartTask(ctx context.Context, req *v1.StartTaskRequest) (*emptypb.Empty, error) {
	if err := s.uc.Start(ctx, req.Id); err != nil {
		return nil, err
	}
	return &emptypb.Empty{}, nil
}

// StopTask stops a task.
func (s *TaskService) StopTask(ctx context.Context, req *v1.StopTaskRequest) (*emptypb.Empty, error) {
	if err := s.uc.Stop(ctx, req.Id); err != nil {
		return nil, err
	}
	return &emptypb.Empty{}, nil
}

// PauseTask pauses a task.
func (s *TaskService) PauseTask(ctx context.Context, req *v1.PauseTaskRequest) (*emptypb.Empty, error) {
	if err := s.uc.Pause(ctx, req.Id); err != nil {
		return nil, err
	}
	return &emptypb.Empty{}, nil
}

// ResumeTask resumes a task.
func (s *TaskService) ResumeTask(ctx context.Context, req *v1.ResumeTaskRequest) (*emptypb.Empty, error) {
	if err := s.uc.Resume(ctx, req.Id); err != nil {
		return nil, err
	}
	return &emptypb.Empty{}, nil
}

// ExecuteTaskNow executes a task immediately.
func (s *TaskService) ExecuteTaskNow(ctx context.Context, req *v1.ExecuteTaskNowRequest) (*v1.TaskExecution, error) {
	exec, err := s.uc.ExecuteNow(ctx, req.Id)
	if err != nil {
		return nil, err
	}
	return toProtoExecution(exec), nil
}

// GetTaskStatus gets the status of a task.
func (s *TaskService) GetTaskStatus(ctx context.Context, req *v1.GetTaskStatusRequest) (*v1.TaskStatusInfo, error) {
	status, err := s.uc.GetStatus(ctx, req.Id)
	if err != nil {
		return nil, err
	}
	return &v1.TaskStatusInfo{
		TaskId:         status.TaskID,
		Status:         v1.TaskStatus(v1.TaskStatus_value[status.Status]),
		LastExecution:  toProtoExecution(status.LastExecution),
		AllowedActions: status.AllowedActions,
	}, nil
}

// GetTaskExecution gets a task execution by ID.
func (s *TaskService) GetTaskExecution(ctx context.Context, req *v1.GetTaskExecutionRequest) (*v1.TaskExecution, error) {
	exec, err := s.uc.GetExecution(ctx, req.Id)
	if err != nil {
		return nil, err
	}
	return toProtoExecution(exec), nil
}

// ListTaskExecutions lists task executions.
func (s *TaskService) ListTaskExecutions(ctx context.Context, req *v1.ListTaskExecutionsRequest) (*v1.ListTaskExecutionsResponse, error) {
	list, total, err := s.uc.ListExecutions(ctx, req.TaskId, int(req.Page), int(req.PageSize))
	if err != nil {
		return nil, err
	}
	return &v1.ListTaskExecutionsResponse{
		Executions: toProtoExecutions(list),
		Total:      total,
		Page:       req.Page,
		PageSize:   req.PageSize,
	}, nil
}

// CancelExecution cancels a task execution.
func (s *TaskService) CancelExecution(ctx context.Context, req *v1.CancelExecutionRequest) (*v1.TaskExecution, error) {
	exec, err := s.uc.CancelExecution(ctx, req.Id)
	if err != nil {
		return nil, err
	}
	return toProtoExecution(exec), nil
}

// RetryExecution retries a task execution.
func (s *TaskService) RetryExecution(ctx context.Context, req *v1.RetryExecutionRequest) (*v1.TaskExecution, error) {
	exec, err := s.uc.RetryExecution(ctx, req.Id)
	if err != nil {
		return nil, err
	}
	return toProtoExecution(exec), nil
}

// BatchStartTasks batch starts tasks.
func (s *TaskService) BatchStartTasks(ctx context.Context, req *v1.BatchTaskOperationRequest) (*v1.BatchTaskOperationResult, error) {
	result, err := s.uc.BatchStart(ctx, req.Ids)
	if err != nil {
		return nil, err
	}
	return &v1.BatchTaskOperationResult{
		Total:   result.Total,
		Success: result.Success,
		Failed:  result.Failed,
	}, nil
}

// BatchStopTasks batch stops tasks.
func (s *TaskService) BatchStopTasks(ctx context.Context, req *v1.BatchTaskOperationRequest) (*v1.BatchTaskOperationResult, error) {
	result, err := s.uc.BatchStop(ctx, req.Ids)
	if err != nil {
		return nil, err
	}
	return &v1.BatchTaskOperationResult{
		Total:   result.Total,
		Success: result.Success,
		Failed:  result.Failed,
	}, nil
}

// BatchDeleteTasks batch deletes tasks.
func (s *TaskService) BatchDeleteTasks(ctx context.Context, req *v1.BatchTaskOperationRequest) (*v1.BatchTaskOperationResult, error) {
	result, err := s.uc.BatchDelete(ctx, req.Ids)
	if err != nil {
		return nil, err
	}
	return &v1.BatchTaskOperationResult{
		Total:   result.Total,
		Success: result.Success,
		Failed:  result.Failed,
	}, nil
}
