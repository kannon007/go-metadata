package biz

import (
	"context"

	"github.com/go-kratos/kratos/v2/log"
)

// TaskRepo is a Task repo interface.
type TaskRepo interface {
	Create(ctx context.Context, task *CollectionTask) (*CollectionTask, error)
	Update(ctx context.Context, task *CollectionTask) (*CollectionTask, error)
	Delete(ctx context.Context, id string) error
	Get(ctx context.Context, id string) (*CollectionTask, error)
	List(ctx context.Context, page, pageSize int) ([]*CollectionTask, int64, error)
	UpdateStatus(ctx context.Context, id string, status TaskStatus) error
	GetExecution(ctx context.Context, id string) (*TaskExecution, error)
	ListExecutions(ctx context.Context, taskID string, page, pageSize int) ([]*TaskExecution, int64, error)
}

// TaskUsecase is a Task usecase.
type TaskUsecase struct {
	repo TaskRepo
	log  *log.Helper
}

// NewTaskUsecase creates a new TaskUsecase.
func NewTaskUsecase(repo TaskRepo, logger log.Logger) *TaskUsecase {
	return &TaskUsecase{repo: repo, log: log.NewHelper(logger)}
}

// Create creates a Task.
func (uc *TaskUsecase) Create(ctx context.Context, task *CollectionTask) (*CollectionTask, error) {
	return uc.repo.Create(ctx, task)
}

// Update updates a Task.
func (uc *TaskUsecase) Update(ctx context.Context, task *CollectionTask) (*CollectionTask, error) {
	return uc.repo.Update(ctx, task)
}

// Delete deletes a Task.
func (uc *TaskUsecase) Delete(ctx context.Context, id string) error {
	return uc.repo.Delete(ctx, id)
}

// Get gets a Task by ID.
func (uc *TaskUsecase) Get(ctx context.Context, id string) (*CollectionTask, error) {
	return uc.repo.Get(ctx, id)
}

// List lists Tasks.
func (uc *TaskUsecase) List(ctx context.Context, page, pageSize int) ([]*CollectionTask, int64, error) {
	return uc.repo.List(ctx, page, pageSize)
}

// Start starts a Task.
func (uc *TaskUsecase) Start(ctx context.Context, id string) error {
	return uc.repo.UpdateStatus(ctx, id, "TASK_STATUS_ACTIVE")
}

// Stop stops a Task.
func (uc *TaskUsecase) Stop(ctx context.Context, id string) error {
	return uc.repo.UpdateStatus(ctx, id, "TASK_STATUS_INACTIVE")
}

// Pause pauses a Task.
func (uc *TaskUsecase) Pause(ctx context.Context, id string) error {
	return uc.repo.UpdateStatus(ctx, id, "TASK_STATUS_PAUSED")
}

// Resume resumes a Task.
func (uc *TaskUsecase) Resume(ctx context.Context, id string) error {
	return uc.repo.UpdateStatus(ctx, id, "TASK_STATUS_ACTIVE")
}

// ExecuteNow executes a Task immediately.
func (uc *TaskUsecase) ExecuteNow(ctx context.Context, id string) (*TaskExecution, error) {
	// TODO: implement execute now
	return &TaskExecution{ID: "exec-1", TaskID: id, Status: "EXECUTION_STATUS_RUNNING"}, nil
}

// GetStatus gets the status of a Task.
func (uc *TaskUsecase) GetStatus(ctx context.Context, id string) (*TaskStatusInfo, error) {
	task, err := uc.repo.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	return &TaskStatusInfo{
		TaskID:         task.ID,
		Status:         task.Status,
		AllowedActions: []string{"start", "stop", "pause"},
	}, nil
}

// GetExecution gets a TaskExecution by ID.
func (uc *TaskUsecase) GetExecution(ctx context.Context, id string) (*TaskExecution, error) {
	return uc.repo.GetExecution(ctx, id)
}

// ListExecutions lists TaskExecutions.
func (uc *TaskUsecase) ListExecutions(ctx context.Context, taskID string, page, pageSize int) ([]*TaskExecution, int64, error) {
	return uc.repo.ListExecutions(ctx, taskID, page, pageSize)
}

// CancelExecution cancels a TaskExecution.
func (uc *TaskUsecase) CancelExecution(ctx context.Context, id string) (*TaskExecution, error) {
	// TODO: implement cancel execution
	return &TaskExecution{ID: id, Status: "EXECUTION_STATUS_CANCELLED"}, nil
}

// RetryExecution retries a TaskExecution.
func (uc *TaskUsecase) RetryExecution(ctx context.Context, id string) (*TaskExecution, error) {
	// TODO: implement retry execution
	return &TaskExecution{ID: id, Status: "EXECUTION_STATUS_PENDING"}, nil
}

// BatchStart batch starts Tasks.
func (uc *TaskUsecase) BatchStart(ctx context.Context, ids []string) (*BatchOperationResult, error) {
	// TODO: implement batch start
	return &BatchOperationResult{Total: int32(len(ids)), Success: int32(len(ids))}, nil
}

// BatchStop batch stops Tasks.
func (uc *TaskUsecase) BatchStop(ctx context.Context, ids []string) (*BatchOperationResult, error) {
	// TODO: implement batch stop
	return &BatchOperationResult{Total: int32(len(ids)), Success: int32(len(ids))}, nil
}

// BatchDelete batch deletes Tasks.
func (uc *TaskUsecase) BatchDelete(ctx context.Context, ids []string) (*BatchOperationResult, error) {
	// TODO: implement batch delete
	return &BatchOperationResult{Total: int32(len(ids)), Success: int32(len(ids))}, nil
}
