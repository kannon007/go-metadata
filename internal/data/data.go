package data

import (
	"go-metadata/internal/biz"
	"go-metadata/internal/conf"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/wire"
)

// ProviderSet is data providers.
var ProviderSet = wire.NewSet(
	NewData,
	NewDataSourceRepo,
	NewTaskRepo,
	NewTemplateRepo,
)

// Data is the data layer struct.
type Data struct {
	log *log.Helper
}

// NewData creates a new Data.
func NewData(c *conf.Data, logger log.Logger) (*Data, func(), error) {
	cleanup := func() {
		log.NewHelper(logger).Info("closing the data resources")
	}
	return &Data{
		log: log.NewHelper(logger),
	}, cleanup, nil
}

// dataSourceRepo implements biz.DataSourceRepo.
type dataSourceRepo struct {
	data *Data
	log  *log.Helper
}

// NewDataSourceRepo creates a new DataSourceRepo.
func NewDataSourceRepo(data *Data, logger log.Logger) biz.DataSourceRepo {
	return &dataSourceRepo{
		data: data,
		log:  log.NewHelper(logger),
	}
}

// taskRepo implements biz.TaskRepo.
type taskRepo struct {
	data *Data
	log  *log.Helper
}

// NewTaskRepo creates a new TaskRepo.
func NewTaskRepo(data *Data, logger log.Logger) biz.TaskRepo {
	return &taskRepo{
		data: data,
		log:  log.NewHelper(logger),
	}
}

// templateRepo implements biz.TemplateRepo.
type templateRepo struct {
	data *Data
	log  *log.Helper
}

// NewTemplateRepo creates a new TemplateRepo.
func NewTemplateRepo(data *Data, logger log.Logger) biz.TemplateRepo {
	return &templateRepo{
		data: data,
		log:  log.NewHelper(logger),
	}
}
