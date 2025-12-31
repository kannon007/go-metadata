package biz

import (
	"context"

	"github.com/go-kratos/kratos/v2/log"
)

// DataSourceRepo is a DataSource repo interface.
type DataSourceRepo interface {
	Create(ctx context.Context, ds *DataSource) (*DataSource, error)
	Update(ctx context.Context, ds *DataSource) (*DataSource, error)
	Delete(ctx context.Context, id string) error
	Get(ctx context.Context, id string) (*DataSource, error)
	List(ctx context.Context, page, pageSize int) ([]*DataSource, int64, error)
	UpdateStatus(ctx context.Context, id string, status DataSourceStatus) error
	GetStats(ctx context.Context) (*DataSourceStats, error)
	CanDelete(ctx context.Context, id string) (*CanDeleteResult, error)
	BatchUpdateStatus(ctx context.Context, ids []string, status DataSourceStatus) (*BatchOperationResult, error)
}

// DataSourceUsecase is a DataSource usecase.
type DataSourceUsecase struct {
	repo DataSourceRepo
	log  *log.Helper
}

// NewDataSourceUsecase creates a new DataSourceUsecase.
func NewDataSourceUsecase(repo DataSourceRepo, logger log.Logger) *DataSourceUsecase {
	return &DataSourceUsecase{repo: repo, log: log.NewHelper(logger)}
}

// Create creates a DataSource.
func (uc *DataSourceUsecase) Create(ctx context.Context, ds *DataSource) (*DataSource, error) {
	return uc.repo.Create(ctx, ds)
}

// Update updates a DataSource.
func (uc *DataSourceUsecase) Update(ctx context.Context, ds *DataSource) (*DataSource, error) {
	return uc.repo.Update(ctx, ds)
}

// Delete deletes a DataSource.
func (uc *DataSourceUsecase) Delete(ctx context.Context, id string) error {
	return uc.repo.Delete(ctx, id)
}

// Get gets a DataSource by ID.
func (uc *DataSourceUsecase) Get(ctx context.Context, id string) (*DataSource, error) {
	return uc.repo.Get(ctx, id)
}

// List lists DataSources.
func (uc *DataSourceUsecase) List(ctx context.Context, page, pageSize int) ([]*DataSource, int64, error) {
	return uc.repo.List(ctx, page, pageSize)
}

// TestConnection tests the connection of a DataSource.
func (uc *DataSourceUsecase) TestConnection(ctx context.Context, id string) (*ConnectionTestResult, error) {
	// TODO: implement connection test
	return &ConnectionTestResult{Success: true, Message: "Connection successful"}, nil
}

// TestConnectionWithConfig tests connection with provided config.
func (uc *DataSourceUsecase) TestConnectionWithConfig(ctx context.Context, dsType DataSourceType, config *ConnectionConfig) (*ConnectionTestResult, error) {
	// TODO: implement connection test with config
	return &ConnectionTestResult{Success: true, Message: "Connection successful"}, nil
}

// RefreshConnectionStatus refreshes the connection status.
func (uc *DataSourceUsecase) RefreshConnectionStatus(ctx context.Context, id string) error {
	// TODO: implement refresh connection status
	return nil
}

// UpdateStatus updates the status of a DataSource.
func (uc *DataSourceUsecase) UpdateStatus(ctx context.Context, id string, status DataSourceStatus) error {
	return uc.repo.UpdateStatus(ctx, id, status)
}

// GetStats gets DataSource statistics.
func (uc *DataSourceUsecase) GetStats(ctx context.Context) (*DataSourceStats, error) {
	return uc.repo.GetStats(ctx)
}

// CanDelete checks if a DataSource can be deleted.
func (uc *DataSourceUsecase) CanDelete(ctx context.Context, id string) (*CanDeleteResult, error) {
	return uc.repo.CanDelete(ctx, id)
}

// BatchUpdateStatus batch updates DataSource status.
func (uc *DataSourceUsecase) BatchUpdateStatus(ctx context.Context, ids []string, status DataSourceStatus) (*BatchOperationResult, error) {
	return uc.repo.BatchUpdateStatus(ctx, ids, status)
}
