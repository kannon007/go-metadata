package biz

import (
	"context"

	"github.com/go-kratos/kratos/v2/log"
)

// TemplateRepo is a Template repo interface.
type TemplateRepo interface {
	Create(ctx context.Context, tpl *DataSourceTemplate) (*DataSourceTemplate, error)
	Update(ctx context.Context, tpl *DataSourceTemplate) (*DataSourceTemplate, error)
	Delete(ctx context.Context, id string) error
	Get(ctx context.Context, id string) (*DataSourceTemplate, error)
	List(ctx context.Context, page, pageSize int) ([]*DataSourceTemplate, int64, error)
	GetSystemTemplates(ctx context.Context, dsType string) ([]*DataSourceTemplate, error)
}

// TemplateUsecase is a Template usecase.
type TemplateUsecase struct {
	repo   TemplateRepo
	dsRepo DataSourceRepo
	log    *log.Helper
}

// NewTemplateUsecase creates a new TemplateUsecase.
func NewTemplateUsecase(repo TemplateRepo, dsRepo DataSourceRepo, logger log.Logger) *TemplateUsecase {
	return &TemplateUsecase{repo: repo, dsRepo: dsRepo, log: log.NewHelper(logger)}
}

// Create creates a Template.
func (uc *TemplateUsecase) Create(ctx context.Context, tpl *DataSourceTemplate) (*DataSourceTemplate, error) {
	return uc.repo.Create(ctx, tpl)
}

// Update updates a Template.
func (uc *TemplateUsecase) Update(ctx context.Context, tpl *DataSourceTemplate) (*DataSourceTemplate, error) {
	return uc.repo.Update(ctx, tpl)
}

// Delete deletes a Template.
func (uc *TemplateUsecase) Delete(ctx context.Context, id string) error {
	return uc.repo.Delete(ctx, id)
}

// Get gets a Template by ID.
func (uc *TemplateUsecase) Get(ctx context.Context, id string) (*DataSourceTemplate, error) {
	return uc.repo.Get(ctx, id)
}

// List lists Templates.
func (uc *TemplateUsecase) List(ctx context.Context, page, pageSize int) ([]*DataSourceTemplate, int64, error) {
	return uc.repo.List(ctx, page, pageSize)
}

// GetSystemTemplates gets system templates.
func (uc *TemplateUsecase) GetSystemTemplates(ctx context.Context, dsType string) ([]*DataSourceTemplate, error) {
	return uc.repo.GetSystemTemplates(ctx, dsType)
}

// Apply applies a template to create a data source.
func (uc *TemplateUsecase) Apply(ctx context.Context, templateID string, ds *DataSource) (*DataSource, error) {
	tpl, err := uc.repo.Get(ctx, templateID)
	if err != nil {
		return nil, err
	}
	ds.Type = tpl.Type
	if ds.Config == nil {
		ds.Config = tpl.ConfigTemplate
	}
	return uc.dsRepo.Create(ctx, ds)
}

// SaveAsTemplate saves a data source as a template.
func (uc *TemplateUsecase) SaveAsTemplate(ctx context.Context, dsID, name, description string) (*DataSourceTemplate, error) {
	ds, err := uc.dsRepo.Get(ctx, dsID)
	if err != nil {
		return nil, err
	}
	return uc.repo.Create(ctx, &DataSourceTemplate{
		Name:           name,
		Type:           ds.Type,
		Description:    description,
		ConfigTemplate: ds.Config,
	})
}
