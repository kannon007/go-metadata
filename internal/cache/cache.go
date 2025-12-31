package cache

import (
	"context"
	"time"
)

// Cache 统一缓存接口
type Cache interface {
	// Get 获取缓存值
	Get(ctx context.Context, key string) ([]byte, error)
	// Set 设置缓存值
	Set(ctx context.Context, key string, value []byte, ttl time.Duration) error
	// Delete 删除缓存
	Delete(ctx context.Context, key string) error
	// Exists 检查key是否存在
	Exists(ctx context.Context, key string) (bool, error)
	// SetNX 仅当key不存在时设置（用于分布式锁）
	SetNX(ctx context.Context, key string, value []byte, ttl time.Duration) (bool, error)
	// Expire 设置过期时间
	Expire(ctx context.Context, key string, ttl time.Duration) error
	// TTL 获取剩余过期时间
	TTL(ctx context.Context, key string) (time.Duration, error)
	// Incr 自增
	Incr(ctx context.Context, key string) (int64, error)
	// Close 关闭连接
	Close() error
}

// ErrKeyNotFound key不存在错误
type ErrKeyNotFound struct {
	Key string
}

func (e *ErrKeyNotFound) Error() string {
	return "cache: key not found: " + e.Key
}

// IsKeyNotFound 判断是否为key不存在错误
func IsKeyNotFound(err error) bool {
	_, ok := err.(*ErrKeyNotFound)
	return ok
}
