package cache

import (
	"time"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/wire"
	"github.com/redis/go-redis/v9"

	"go-metadata/internal/conf"
)

// ProviderSet is cache providers.
var ProviderSet = wire.NewSet(NewCache, NewService)

// NewCache 根据配置创建缓存实例
func NewCache(c *conf.Data, logger log.Logger, rdb *redis.Client) Cache {
	helper := log.NewHelper(logger)

	cacheType := "auto"
	prefix := "metadata"
	cleanupInterval := 5 * time.Minute

	if c.Cache != nil {
		if c.Cache.Type != "" {
			cacheType = c.Cache.Type
		}
		if c.Cache.Prefix != "" {
			prefix = c.Cache.Prefix
		}
		if c.Cache.CleanupInterval != nil {
			cleanupInterval = c.Cache.CleanupInterval.AsDuration()
		}
	}

	// 根据配置类型选择缓存实现
	switch cacheType {
	case "redis":
		if rdb == nil {
			helper.Warn("redis client is nil, fallback to memory cache")
			return NewMemoryCache(WithCleanupInterval(cleanupInterval))
		}
		helper.Info("using redis cache")
		return NewRedisCache(rdb, WithPrefix(prefix))
	case "memory":
		helper.Info("using memory cache")
		return NewMemoryCache(WithCleanupInterval(cleanupInterval))
	default: // auto
		if rdb != nil && c.Redis != nil {
			helper.Info("using redis cache (auto)")
			return NewRedisCache(rdb, WithPrefix(prefix))
		}
		helper.Info("using memory cache (auto)")
		return NewMemoryCache(WithCleanupInterval(cleanupInterval))
	}
}
