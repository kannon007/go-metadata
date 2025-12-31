package service

import (
	"context"

	v1 "go-metadata/api/metadata/v1"
	"go-metadata/internal/biz"

	"github.com/go-kratos/kratos/v2/log"
	"google.golang.org/protobuf/types/known/emptypb"
)

// DataSourceService implements the DataSourceService gRPC/HTTP service.
type DataSourceService struct {
	v1.UnimplementedDataSourceServiceServer

	uc  *biz.DataSourceUsecase
	log *log.Helper
}

// NewDataSourceService creates a new DataSourceService.
func NewDataSourceService(uc *biz.DataSourceUsecase, logger log.Logger) *DataSourceService {
	return &DataSourceService{
		uc:  uc,
		log: log.NewHelper(logger),
	}
}

// CreateDataSource creates a new data source.
func (s *DataSourceService) CreateDataSource(ctx context.Context, req *v1.CreateDataSourceRequest) (*v1.DataSource, error) {
	ds, err := s.uc.Create(ctx, &biz.DataSource{
		Name:        req.Name,
		Type:        biz.DataSourceType(req.Type.String()),
		Description: req.Description,
		Config:      toBizConnectionConfig(req.Config),
		Tags:        req.Tags,
	})
	if err != nil {
		return nil, err
	}
	return toProtoDataSource(ds), nil
}

// UpdateDataSource updates an existing data source.
func (s *DataSourceService) UpdateDataSource(ctx context.Context, req *v1.UpdateDataSourceRequest) (*v1.DataSource, error) {
	ds, err := s.uc.Update(ctx, &biz.DataSource{
		ID:          req.Id,
		Name:        req.Name,
		Description: req.Description,
		Config:      toBizConnectionConfig(req.Config),
		Tags:        req.Tags,
	})
	if err != nil {
		return nil, err
	}
	return toProtoDataSource(ds), nil
}

// DeleteDataSource deletes a data source.
func (s *DataSourceService) DeleteDataSource(ctx context.Context, req *v1.DeleteDataSourceRequest) (*emptypb.Empty, error) {
	if err := s.uc.Delete(ctx, req.Id); err != nil {
		return nil, err
	}
	return &emptypb.Empty{}, nil
}

// GetDataSource gets a data source by ID.
func (s *DataSourceService) GetDataSource(ctx context.Context, req *v1.GetDataSourceRequest) (*v1.DataSource, error) {
	ds, err := s.uc.Get(ctx, req.Id)
	if err != nil {
		return nil, err
	}
	return toProtoDataSource(ds), nil
}

// ListDataSources lists data sources.
func (s *DataSourceService) ListDataSources(ctx context.Context, req *v1.ListDataSourcesRequest) (*v1.ListDataSourcesResponse, error) {
	list, total, err := s.uc.List(ctx, int(req.Page), int(req.PageSize))
	if err != nil {
		return nil, err
	}
	return &v1.ListDataSourcesResponse{
		DataSources: toProtoDataSources(list),
		Total:       total,
		Page:        req.Page,
		PageSize:    req.PageSize,
	}, nil
}

// TestConnection tests the connection of a data source.
func (s *DataSourceService) TestConnection(ctx context.Context, req *v1.TestConnectionRequest) (*v1.ConnectionTestResult, error) {
	result, err := s.uc.TestConnection(ctx, req.Id)
	if err != nil {
		return nil, err
	}
	return toProtoConnectionTestResult(result), nil
}

// TestConnectionWithConfig tests connection with provided config.
func (s *DataSourceService) TestConnectionWithConfig(ctx context.Context, req *v1.TestConnectionWithConfigRequest) (*v1.ConnectionTestResult, error) {
	result, err := s.uc.TestConnectionWithConfig(ctx, biz.DataSourceType(req.Type.String()), toBizConnectionConfig(req.Config))
	if err != nil {
		return nil, err
	}
	return toProtoConnectionTestResult(result), nil
}

// RefreshConnectionStatus refreshes the connection status.
func (s *DataSourceService) RefreshConnectionStatus(ctx context.Context, req *v1.RefreshConnectionStatusRequest) (*emptypb.Empty, error) {
	if err := s.uc.RefreshConnectionStatus(ctx, req.Id); err != nil {
		return nil, err
	}
	return &emptypb.Empty{}, nil
}

// UpdateDataSourceStatus updates the status of a data source.
func (s *DataSourceService) UpdateDataSourceStatus(ctx context.Context, req *v1.UpdateDataSourceStatusRequest) (*emptypb.Empty, error) {
	if err := s.uc.UpdateStatus(ctx, req.Id, biz.DataSourceStatus(req.Status.String())); err != nil {
		return nil, err
	}
	return &emptypb.Empty{}, nil
}

// GetDataSourceStats gets data source statistics.
func (s *DataSourceService) GetDataSourceStats(ctx context.Context, req *emptypb.Empty) (*v1.DataSourceStats, error) {
	stats, err := s.uc.GetStats(ctx)
	if err != nil {
		return nil, err
	}
	return &v1.DataSourceStats{
		TotalCount:    stats.TotalCount,
		ActiveCount:   stats.ActiveCount,
		InactiveCount: stats.InactiveCount,
		ErrorCount:    stats.ErrorCount,
	}, nil
}

// CanDeleteDataSource checks if a data source can be deleted.
func (s *DataSourceService) CanDeleteDataSource(ctx context.Context, req *v1.CanDeleteDataSourceRequest) (*v1.CanDeleteResult, error) {
	result, err := s.uc.CanDelete(ctx, req.Id)
	if err != nil {
		return nil, err
	}
	return &v1.CanDeleteResult{
		CanDelete:       result.CanDelete,
		Reason:          result.Reason,
		AssociatedTasks: result.AssociatedTasks,
	}, nil
}

// BatchUpdateStatus batch updates data source status.
func (s *DataSourceService) BatchUpdateStatus(ctx context.Context, req *v1.BatchUpdateStatusRequest) (*v1.BatchOperationResult, error) {
	result, err := s.uc.BatchUpdateStatus(ctx, req.Ids, biz.DataSourceStatus(req.Status.String()))
	if err != nil {
		return nil, err
	}
	return &v1.BatchOperationResult{
		Total:   result.Total,
		Success: result.Success,
		Failed:  result.Failed,
		Errors:  result.Errors,
	}, nil
}

// BatchImportDataSources batch imports data sources.
func (s *DataSourceService) BatchImportDataSources(ctx context.Context, req *v1.BatchImportRequest) (*v1.BatchImportResult, error) {
	// TODO: implement batch import
	return &v1.BatchImportResult{}, nil
}

// BatchExportDataSources batch exports data sources.
func (s *DataSourceService) BatchExportDataSources(ctx context.Context, req *v1.BatchExportRequest) (*v1.BatchExportResponse, error) {
	// TODO: implement batch export
	return &v1.BatchExportResponse{}, nil
}
