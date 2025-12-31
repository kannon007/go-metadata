package server

import (
	v1 "go-metadata/api/metadata/v1"
	"go-metadata/internal/conf"
	"go-metadata/internal/service"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/middleware/logging"
	"github.com/go-kratos/kratos/v2/middleware/recovery"
	"github.com/go-kratos/kratos/v2/middleware/validate"
	"github.com/go-kratos/kratos/v2/transport/http"
)

// NewHTTPServer new an HTTP server.
func NewHTTPServer(
	c *conf.Server,
	logger log.Logger,
	datasource *service.DataSourceService,
	task *service.TaskService,
	template *service.TemplateService,
	user *service.UserService,
) *http.Server {
	var opts = []http.ServerOption{
		http.Middleware(
			recovery.Recovery(),
			logging.Server(logger),
			validate.Validator(),
		),
	}
	if c.Http.Network != "" {
		opts = append(opts, http.Network(c.Http.Network))
	}
	if c.Http.Addr != "" {
		opts = append(opts, http.Address(c.Http.Addr))
	}
	if c.Http.Timeout != nil {
		opts = append(opts, http.Timeout(c.Http.Timeout.AsDuration()))
	}
	srv := http.NewServer(opts...)

	// 注册生成的 HTTP 服务
	v1.RegisterDataSourceServiceHTTPServer(srv, datasource)
	v1.RegisterTaskServiceHTTPServer(srv, task)
	v1.RegisterTemplateServiceHTTPServer(srv, template)
	v1.RegisterUserServiceHTTPServer(srv, user)

	return srv
}
