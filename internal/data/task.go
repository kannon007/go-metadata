package data

import (
	"context"

	"go-metadata/internal/biz"
)

func (r *taskRepo) Create(ctx context.Context, task *biz.CollectionTask) (*biz.CollectionTask, error) {
	// TODO: implement database operation
	task.ID = "task-1"
	return task, nil
}

func (r *taskRepo) Update(ctx context.Context, task *biz.CollectionTask) (*biz.CollectionTask, error) {
	// TODO: implement database operation
	return task, nil
}

func (r *taskRepo) Delete(ctx context.Context, id string) error {
	// TODO: implement database operation
	return nil
}

func (r *taskRepo) Get(ctx context.Context, id string) (*biz.CollectionTask, error) {
	// TODO: implement database operation
	return &biz.CollectionTask{ID: id, Name: "Test Task"}, nil
}

func (r *taskRepo) List(ctx context.Context, page, pageSize int) ([]*biz.CollectionTask, int64, error) {
	// TODO: implement database operation
	return []*biz.CollectionTask{}, 0, nil
}

func (r *taskRepo) UpdateStatus(ctx context.Context, id string, status biz.TaskStatus) error {
	// TODO: implement database operation
	return nil
}

func (r *taskRepo) GetExecution(ctx context.Context, id string) (*biz.TaskExecution, error) {
	// TODO: implement database operation
	return &biz.TaskExecution{ID: id}, nil
}

func (r *taskRepo) ListExecutions(ctx context.Context, taskID string, page, pageSize int) ([]*biz.TaskExecution, int64, error) {
	// TODO: implement database operation
	return []*biz.TaskExecution{}, 0, nil
}
