package cache

import (
	"context"
	"sync"
	"time"
)

// item 缓存项
type item struct {
	value      []byte
	expiration int64 // 过期时间戳（纳秒），0表示永不过期
}

// isExpired 检查是否过期
func (i *item) isExpired() bool {
	if i.expiration == 0 {
		return false
	}
	return time.Now().UnixNano() > i.expiration
}

// MemoryCache 本地内存缓存实现
type MemoryCache struct {
	items           map[string]*item
	mu              sync.RWMutex
	cleanupInterval time.Duration
	stopCleanup     chan struct{}
}

// MemoryCacheOption 内存缓存配置选项
type MemoryCacheOption func(*MemoryCache)

// WithCleanupInterval 设置清理间隔
func WithCleanupInterval(d time.Duration) MemoryCacheOption {
	return func(c *MemoryCache) {
		c.cleanupInterval = d
	}
}

// NewMemoryCache 创建内存缓存
func NewMemoryCache(opts ...MemoryCacheOption) *MemoryCache {
	c := &MemoryCache{
		items:           make(map[string]*item),
		cleanupInterval: 5 * time.Minute,
		stopCleanup:     make(chan struct{}),
	}
	for _, opt := range opts {
		opt(c)
	}
	go c.cleanupLoop()
	return c
}


// cleanupLoop 定期清理过期项
func (c *MemoryCache) cleanupLoop() {
	ticker := time.NewTicker(c.cleanupInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			c.deleteExpired()
		case <-c.stopCleanup:
			return
		}
	}
}

// deleteExpired 删除过期项
func (c *MemoryCache) deleteExpired() {
	c.mu.Lock()
	defer c.mu.Unlock()
	for k, v := range c.items {
		if v.isExpired() {
			delete(c.items, k)
		}
	}
}

// Get 获取缓存值
func (c *MemoryCache) Get(ctx context.Context, key string) ([]byte, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	it, ok := c.items[key]
	if !ok || it.isExpired() {
		return nil, &ErrKeyNotFound{Key: key}
	}
	// 返回副本，避免外部修改
	result := make([]byte, len(it.value))
	copy(result, it.value)
	return result, nil
}

// Set 设置缓存值
func (c *MemoryCache) Set(ctx context.Context, key string, value []byte, ttl time.Duration) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	var exp int64
	if ttl > 0 {
		exp = time.Now().Add(ttl).UnixNano()
	}
	// 存储副本
	v := make([]byte, len(value))
	copy(v, value)
	c.items[key] = &item{value: v, expiration: exp}
	return nil
}

// Delete 删除缓存
func (c *MemoryCache) Delete(ctx context.Context, key string) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.items, key)
	return nil
}

// Exists 检查key是否存在
func (c *MemoryCache) Exists(ctx context.Context, key string) (bool, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	it, ok := c.items[key]
	if !ok || it.isExpired() {
		return false, nil
	}
	return true, nil
}

// SetNX 仅当key不存在时设置
func (c *MemoryCache) SetNX(ctx context.Context, key string, value []byte, ttl time.Duration) (bool, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	it, ok := c.items[key]
	if ok && !it.isExpired() {
		return false, nil
	}
	var exp int64
	if ttl > 0 {
		exp = time.Now().Add(ttl).UnixNano()
	}
	v := make([]byte, len(value))
	copy(v, value)
	c.items[key] = &item{value: v, expiration: exp}
	return true, nil
}

// Expire 设置过期时间
func (c *MemoryCache) Expire(ctx context.Context, key string, ttl time.Duration) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	it, ok := c.items[key]
	if !ok || it.isExpired() {
		return &ErrKeyNotFound{Key: key}
	}
	if ttl > 0 {
		it.expiration = time.Now().Add(ttl).UnixNano()
	} else {
		it.expiration = 0
	}
	return nil
}

// TTL 获取剩余过期时间
func (c *MemoryCache) TTL(ctx context.Context, key string) (time.Duration, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	it, ok := c.items[key]
	if !ok || it.isExpired() {
		return 0, &ErrKeyNotFound{Key: key}
	}
	if it.expiration == 0 {
		return -1, nil // 永不过期
	}
	return time.Duration(it.expiration - time.Now().UnixNano()), nil
}

// Incr 自增
func (c *MemoryCache) Incr(ctx context.Context, key string) (int64, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	it, ok := c.items[key]
	if !ok || it.isExpired() {
		c.items[key] = &item{value: []byte("1"), expiration: 0}
		return 1, nil
	}
	var val int64
	for _, b := range it.value {
		val = val*10 + int64(b-'0')
	}
	val++
	it.value = []byte(string(rune('0'+val%10)) + string(it.value[1:]))
	// 简化处理：直接用fmt
	it.value = []byte(formatInt64(val))
	return val, nil
}

func formatInt64(n int64) string {
	if n == 0 {
		return "0"
	}
	var buf [20]byte
	i := len(buf)
	for n > 0 {
		i--
		buf[i] = byte('0' + n%10)
		n /= 10
	}
	return string(buf[i:])
}

// Close 关闭缓存
func (c *MemoryCache) Close() error {
	close(c.stopCleanup)
	return nil
}
