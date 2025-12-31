package data

import (
	"context"

	"go-metadata/internal/biz"
)

func (r *templateRepo) Create(ctx context.Context, tpl *biz.DataSourceTemplate) (*biz.DataSourceTemplate, error) {
	// TODO: implement database operation
	tpl.ID = "tpl-1"
	return tpl, nil
}

func (r *templateRepo) Update(ctx context.Context, tpl *biz.DataSourceTemplate) (*biz.DataSourceTemplate, error) {
	// TODO: implement database operation
	return tpl, nil
}

func (r *templateRepo) Delete(ctx context.Context, id string) error {
	// TODO: implement database operation
	return nil
}

func (r *templateRepo) Get(ctx context.Context, id string) (*biz.DataSourceTemplate, error) {
	// TODO: implement database operation
	return &biz.DataSourceTemplate{ID: id, Name: "Test Template"}, nil
}

func (r *templateRepo) List(ctx context.Context, page, pageSize int) ([]*biz.DataSourceTemplate, int64, error) {
	// TODO: implement database operation
	return []*biz.DataSourceTemplate{}, 0, nil
}

func (r *templateRepo) GetSystemTemplates(ctx context.Context, dsType string) ([]*biz.DataSourceTemplate, error) {
	// TODO: implement database operation
	return []*biz.DataSourceTemplate{}, nil
}
