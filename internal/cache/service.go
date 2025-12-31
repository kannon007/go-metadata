package cache

import (
	"context"
	"encoding/json"
	"time"
)

// Service 缓存服务，提供类型安全的缓存操作
type Service struct {
	cache Cache
}

// NewService 创建缓存服务
func NewService(cache Cache) *Service {
	return &Service{cache: cache}
}

// GetString 获取字符串
func (s *Service) GetString(ctx context.Context, key string) (string, error) {
	data, err := s.cache.Get(ctx, key)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// SetString 设置字符串
func (s *Service) SetString(ctx context.Context, key, value string, ttl time.Duration) error {
	return s.cache.Set(ctx, key, []byte(value), ttl)
}

// GetJSON 获取JSON对象
func (s *Service) GetJSON(ctx context.Context, key string, dest interface{}) error {
	data, err := s.cache.Get(ctx, key)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, dest)
}

// SetJSON 设置JSON对象
func (s *Service) SetJSON(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}
	return s.cache.Set(ctx, key, data, ttl)
}

// Delete 删除缓存
func (s *Service) Delete(ctx context.Context, key string) error {
	return s.cache.Delete(ctx, key)
}

// Exists 检查key是否存在
func (s *Service) Exists(ctx context.Context, key string) (bool, error) {
	return s.cache.Exists(ctx, key)
}

// GetOrSet 获取缓存，不存在则调用loader加载并缓存
func (s *Service) GetOrSet(ctx context.Context, key string, ttl time.Duration, loader func() ([]byte, error)) ([]byte, error) {
	data, err := s.cache.Get(ctx, key)
	if err == nil {
		return data, nil
	}
	if !IsKeyNotFound(err) {
		return nil, err
	}
	// 加载数据
	data, err = loader()
	if err != nil {
		return nil, err
	}
	// 缓存数据
	if err := s.cache.Set(ctx, key, data, ttl); err != nil {
		return data, nil // 缓存失败不影响返回
	}
	return data, nil
}

// GetOrSetJSON 获取JSON缓存，不存在则调用loader加载并缓存
func (s *Service) GetOrSetJSON(ctx context.Context, key string, dest interface{}, ttl time.Duration, loader func() (interface{}, error)) error {
	err := s.GetJSON(ctx, key, dest)
	if err == nil {
		return nil
	}
	if !IsKeyNotFound(err) {
		return err
	}
	// 加载数据
	value, err := loader()
	if err != nil {
		return err
	}
	// 缓存数据
	_ = s.SetJSON(ctx, key, value, ttl)
	// 将value赋值给dest
	data, _ := json.Marshal(value)
	return json.Unmarshal(data, dest)
}

// Lock 获取分布式锁
func (s *Service) Lock(ctx context.Context, key string, ttl time.Duration) (bool, error) {
	return s.cache.SetNX(ctx, key, []byte("1"), ttl)
}

// Unlock 释放分布式锁
func (s *Service) Unlock(ctx context.Context, key string) error {
	return s.cache.Delete(ctx, key)
}

// Incr 自增计数器
func (s *Service) Incr(ctx context.Context, key string) (int64, error) {
	return s.cache.Incr(ctx, key)
}

// Raw 获取底层Cache接口
func (s *Service) Raw() Cache {
	return s.cache
}
