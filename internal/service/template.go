package service

import (
	"context"

	v1 "go-metadata/api/metadata/v1"
	"go-metadata/internal/biz"

	"github.com/go-kratos/kratos/v2/log"
	"google.golang.org/protobuf/types/known/emptypb"
)

// TemplateService implements the TemplateService gRPC/HTTP service.
type TemplateService struct {
	v1.UnimplementedTemplateServiceServer

	uc  *biz.TemplateUsecase
	log *log.Helper
}

// NewTemplateService creates a new TemplateService.
func NewTemplateService(uc *biz.TemplateUsecase, logger log.Logger) *TemplateService {
	return &TemplateService{
		uc:  uc,
		log: log.NewHelper(logger),
	}
}

// CreateTemplate creates a new template.
func (s *TemplateService) CreateTemplate(ctx context.Context, req *v1.CreateTemplateRequest) (*v1.DataSourceTemplate, error) {
	tpl, err := s.uc.Create(ctx, &biz.DataSourceTemplate{
		Name:           req.Name,
		Type:           req.Type.String(),
		Description:    req.Description,
		ConfigTemplate: toBizConnectionConfig(req.ConfigTemplate),
	})
	if err != nil {
		return nil, err
	}
	return toProtoTemplate(tpl), nil
}

// UpdateTemplate updates an existing template.
func (s *TemplateService) UpdateTemplate(ctx context.Context, req *v1.UpdateTemplateRequest) (*v1.DataSourceTemplate, error) {
	tpl, err := s.uc.Update(ctx, &biz.DataSourceTemplate{
		ID:             req.Id,
		Name:           req.Name,
		Description:    req.Description,
		ConfigTemplate: toBizConnectionConfig(req.ConfigTemplate),
	})
	if err != nil {
		return nil, err
	}
	return toProtoTemplate(tpl), nil
}

// DeleteTemplate deletes a template.
func (s *TemplateService) DeleteTemplate(ctx context.Context, req *v1.DeleteTemplateRequest) (*emptypb.Empty, error) {
	if err := s.uc.Delete(ctx, req.Id); err != nil {
		return nil, err
	}
	return &emptypb.Empty{}, nil
}

// GetTemplate gets a template by ID.
func (s *TemplateService) GetTemplate(ctx context.Context, req *v1.GetTemplateRequest) (*v1.DataSourceTemplate, error) {
	tpl, err := s.uc.Get(ctx, req.Id)
	if err != nil {
		return nil, err
	}
	return toProtoTemplate(tpl), nil
}

// ListTemplates lists templates.
func (s *TemplateService) ListTemplates(ctx context.Context, req *v1.ListTemplatesRequest) (*v1.ListTemplatesResponse, error) {
	list, total, err := s.uc.List(ctx, int(req.Page), int(req.PageSize))
	if err != nil {
		return nil, err
	}
	return &v1.ListTemplatesResponse{
		Templates: toProtoTemplates(list),
		Total:     total,
		Page:      req.Page,
		PageSize:  req.PageSize,
	}, nil
}

// GetSystemTemplates gets system templates.
func (s *TemplateService) GetSystemTemplates(ctx context.Context, req *v1.GetSystemTemplatesRequest) (*v1.GetSystemTemplatesResponse, error) {
	list, err := s.uc.GetSystemTemplates(ctx, req.Type.String())
	if err != nil {
		return nil, err
	}
	return &v1.GetSystemTemplatesResponse{
		Templates: toProtoTemplates(list),
	}, nil
}

// ApplyTemplate applies a template to create a data source.
func (s *TemplateService) ApplyTemplate(ctx context.Context, req *v1.ApplyTemplateRequest) (*v1.DataSource, error) {
	ds, err := s.uc.Apply(ctx, req.TemplateId, &biz.DataSource{
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

// ExportTemplates exports templates.
func (s *TemplateService) ExportTemplates(ctx context.Context, req *v1.ExportTemplatesRequest) (*v1.ExportTemplatesResponse, error) {
	// TODO: implement export
	return &v1.ExportTemplatesResponse{}, nil
}

// ImportTemplates imports templates.
func (s *TemplateService) ImportTemplates(ctx context.Context, req *v1.ImportTemplatesRequest) (*v1.ImportTemplatesResult, error) {
	// TODO: implement import
	return &v1.ImportTemplatesResult{}, nil
}

// SaveAsTemplate saves a data source as a template.
func (s *TemplateService) SaveAsTemplate(ctx context.Context, req *v1.SaveAsTemplateRequest) (*v1.DataSourceTemplate, error) {
	tpl, err := s.uc.SaveAsTemplate(ctx, req.DatasourceId, req.Name, req.Description)
	if err != nil {
		return nil, err
	}
	return toProtoTemplate(tpl), nil
}
