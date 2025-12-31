package data

import (
	"context"

	"go-metadata/internal/biz"
)

func (r *dataSourceRepo) Create(ctx context.Context, ds *biz.DataSource) (*biz.DataSource, error) {
	// TODO: implement database operation
	ds.ID = "ds-1"
	return ds, nil
}

func (r *dataSourceRepo) Update(ctx context.Context, ds *biz.DataSource) (*biz.DataSource, error) {
	// TODO: implement database operation
	return ds, nil
}

func (r *dataSourceRepo) Delete(ctx context.Context, id string) error {
	// TODO: implement database operation
	return nil
}

func (r *dataSourceRepo) Get(ctx context.Context, id string) (*biz.DataSource, error) {
	// TODO: implement database operation
	return &biz.DataSource{ID: id, Name: "Test DataSource"}, nil
}

func (r *dataSourceRepo) List(ctx context.Context, page, pageSize int) ([]*biz.DataSource, int64, error) {
	// TODO: implement database operation
	return []*biz.DataSource{}, 0, nil
}

func (r *dataSourceRepo) UpdateStatus(ctx context.Context, id string, status biz.DataSourceStatus) error {
	// TODO: implement database operation
	return nil
}

func (r *dataSourceRepo) GetStats(ctx context.Context) (*biz.DataSourceStats, error) {
	// TODO: implement database operation
	return &biz.DataSourceStats{}, nil
}

func (r *dataSourceRepo) CanDelete(ctx context.Context, id string) (*biz.CanDeleteResult, error) {
	// TODO: implement database operation
	return &biz.CanDeleteResult{CanDelete: true}, nil
}

func (r *dataSourceRepo) BatchUpdateStatus(ctx context.Context, ids []string, status biz.DataSourceStatus) (*biz.BatchOperationResult, error) {
	// TODO: implement database operation
	return &biz.BatchOperationResult{Total: int32(len(ids)), Success: int32(len(ids))}, nil
}
