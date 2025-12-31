package cache

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

// RedisCache Redis缓存实现
type RedisCache struct {
	client *redis.Client
	prefix string
}

// RedisCacheOption Redis缓存配置选项
type RedisCacheOption func(*RedisCache)

// WithPrefix 设置key前缀
func WithPrefix(prefix string) RedisCacheOption {
	return func(c *RedisCache) {
		c.prefix = prefix
	}
}

// NewRedisCache 创建Redis缓存
func NewRedisCache(client *redis.Client, opts ...RedisCacheOption) *RedisCache {
	c := &RedisCache{
		client: client,
		prefix: "",
	}
	for _, opt := range opts {
		opt(c)
	}
	return c
}

// key 生成带前缀的key
func (c *RedisCache) key(k string) string {
	if c.prefix == "" {
		return k
	}
	return c.prefix + ":" + k
}

// Get 获取缓存值
func (c *RedisCache) Get(ctx context.Context, key string) ([]byte, error) {
	val, err := c.client.Get(ctx, c.key(key)).Bytes()
	if err == redis.Nil {
		return nil, &ErrKeyNotFound{Key: key}
	}
	return val, err
}

// Set 设置缓存值
func (c *RedisCache) Set(ctx context.Context, key string, value []byte, ttl time.Duration) error {
	return c.client.Set(ctx, c.key(key), value, ttl).Err()
}

// Delete 删除缓存
func (c *RedisCache) Delete(ctx context.Context, key string) error {
	return c.client.Del(ctx, c.key(key)).Err()
}

// Exists 检查key是否存在
func (c *RedisCache) Exists(ctx context.Context, key string) (bool, error) {
	n, err := c.client.Exists(ctx, c.key(key)).Result()
	return n > 0, err
}

// SetNX 仅当key不存在时设置
func (c *RedisCache) SetNX(ctx context.Context, key string, value []byte, ttl time.Duration) (bool, error) {
	return c.client.SetNX(ctx, c.key(key), value, ttl).Result()
}

// Expire 设置过期时间
func (c *RedisCache) Expire(ctx context.Context, key string, ttl time.Duration) error {
	return c.client.Expire(ctx, c.key(key), ttl).Err()
}

// TTL 获取剩余过期时间
func (c *RedisCache) TTL(ctx context.Context, key string) (time.Duration, error) {
	ttl, err := c.client.TTL(ctx, c.key(key)).Result()
	if err != nil {
		return 0, err
	}
	if ttl == -2 {
		return 0, &ErrKeyNotFound{Key: key}
	}
	return ttl, nil
}

// Incr 自增
func (c *RedisCache) Incr(ctx context.Context, key string) (int64, error) {
	return c.client.Incr(ctx, c.key(key)).Result()
}

// Close 关闭连接（由外部管理redis client生命周期）
func (c *RedisCache) Close() error {
	return nil
}
