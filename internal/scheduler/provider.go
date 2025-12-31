package scheduler

import (
	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/wire"

	"go-metadata/internal/conf"
	"go-metadata/internal/scheduler/dolphin"
	"go-metadata/internal/scheduler/types"
)

// ProviderSet is scheduler providers.
var ProviderSet = wire.NewSet(
	NewSchedulerManager,
	NewBuiltinScheduler,
	NewDolphinSchedulerAdapterFromConfig,
	ProvideSchedulerAdapter,
)

// NewDolphinSchedulerAdapterFromConfig 从配置创建DolphinScheduler适配器
func NewDolphinSchedulerAdapterFromConfig(c *conf.Scheduler, logger log.Logger) *dolphin.DolphinSchedulerAdapter {
	if c == nil || c.Type != "dolphinscheduler" {
		return nil
	}

	config := &dolphin.Config{
		Endpoint:    c.Endpoint,
		Token:       c.Properties["token"],
		ProjectName: c.Properties["project_name"],
	}

	if config.ProjectName == "" {
		config.ProjectName = "metadata-collection"
	}

	return dolphin.NewDolphinSchedulerAdapter(config, logger)
}

// ProvideSchedulerAdapter 提供调度器适配器并注册到管理器
func ProvideSchedulerAdapter(
	manager *SchedulerManager,
	builtin *BuiltinScheduler,
	dolphinAdapter *dolphin.DolphinSchedulerAdapter,
) *SchedulerManager {
	// 注册内置调度器
	manager.RegisterAdapter(types.SchedulerAdapter(builtin))

	// 注册DolphinScheduler适配器（如果配置了）
	if dolphinAdapter != nil {
		manager.RegisterAdapter(types.SchedulerAdapter(dolphinAdapter))
	}

	return manager
}
