package auth

import (
	"context"
	"encoding/json"
	"net/http"
	"sync"
	"time"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/middleware"
)

// RateLimitConfig 限流配置
type RateLimitConfig struct {
	Enabled       bool          `json:"enabled"`
	RequestsPerIP int           `json:"requests_per_ip"` // 每个IP在时间窗口内的最大请求数
	BurstSize     int           `json:"burst_size"`      // 突发请求数
	Window        time.Duration `json:"window"`          // 时间窗口
	CleanupPeriod time.Duration `json:"cleanup_period"`  // 清理周期
}

// DefaultRateLimitConfig 默认限流配置
func DefaultRateLimitConfig() *RateLimitConfig {
	return &RateLimitConfig{
		Enabled:       true,
		RequestsPerIP: 100,
		BurstSize:     20,
		Window:        time.Minute,
		CleanupPeriod: 5 * time.Minute,
	}
}

// clientInfo 客户端信息
type clientInfo struct {
	tokens     int       // 当前令牌数
	lastRefill time.Time // 上次补充时间
	requests   int       // 当前窗口内的请求数
	windowStart time.Time // 窗口开始时间
}

// RateLimiter 限流器
type RateLimiter struct {
	config  *RateLimitConfig
	clients map[string]*clientInfo
	mu      sync.RWMutex
	log     *log.Helper
	stopCh  chan struct{}
}

// NewRateLimiter 创建限流器
func NewRateLimiter(config *RateLimitConfig, logger log.Logger) *RateLimiter {
	if config == nil {
		config = DefaultRateLimitConfig()
	}

	rl := &RateLimiter{
		config:  config,
		clients: make(map[string]*clientInfo),
		log:     log.NewHelper(logger),
		stopCh:  make(chan struct{}),
	}

	// 启动清理协程
	if config.CleanupPeriod > 0 {
		go rl.cleanupLoop()
	}

	return rl
}

// cleanupLoop 定期清理过期的客户端记录
func (rl *RateLimiter) cleanupLoop() {
	ticker := time.NewTicker(rl.config.CleanupPeriod)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			rl.cleanup()
		case <-rl.stopCh:
			return
		}
	}
}

// cleanup 清理过期的客户端记录
func (rl *RateLimiter) cleanup() {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	expireTime := now.Add(-rl.config.Window * 2)

	for ip, info := range rl.clients {
		if info.lastRefill.Before(expireTime) {
			delete(rl.clients, ip)
		}
	}
}

// Stop 停止限流器
func (rl *RateLimiter) Stop() {
	close(rl.stopCh)
}

// Allow 检查是否允许请求
func (rl *RateLimiter) Allow(clientIP string) bool {
	if !rl.config.Enabled {
		return true
	}

	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	info, exists := rl.clients[clientIP]

	if !exists {
		// 新客户端，初始化
		rl.clients[clientIP] = &clientInfo{
			tokens:      rl.config.BurstSize - 1,
			lastRefill:  now,
			requests:    1,
			windowStart: now,
		}
		return true
	}

	// 检查是否需要重置窗口
	if now.Sub(info.windowStart) >= rl.config.Window {
		info.requests = 0
		info.windowStart = now
		info.tokens = rl.config.BurstSize
	}

	// 补充令牌（令牌桶算法）
	elapsed := now.Sub(info.lastRefill)
	tokensToAdd := int(elapsed.Seconds() * float64(rl.config.RequestsPerIP) / rl.config.Window.Seconds())
	if tokensToAdd > 0 {
		info.tokens += tokensToAdd
		if info.tokens > rl.config.BurstSize {
			info.tokens = rl.config.BurstSize
		}
		info.lastRefill = now
	}

	// 检查是否超过限制
	if info.requests >= rl.config.RequestsPerIP {
		return false
	}

	// 检查令牌
	if info.tokens <= 0 {
		return false
	}

	// 消耗令牌
	info.tokens--
	info.requests++

	return true
}

// GetClientStats 获取客户端统计信息
func (rl *RateLimiter) GetClientStats(clientIP string) (requests int, remaining int) {
	rl.mu.RLock()
	defer rl.mu.RUnlock()

	info, exists := rl.clients[clientIP]
	if !exists {
		return 0, rl.config.RequestsPerIP
	}

	return info.requests, rl.config.RequestsPerIP - info.requests
}

// Handler 返回Kratos中间件处理函数
func (rl *RateLimiter) Handler() middleware.Middleware {
	return func(handler middleware.Handler) middleware.Handler {
		return func(ctx context.Context, req interface{}) (interface{}, error) {
			clientIP := ClientIPFromContext(ctx)
			if clientIP == "" {
				clientIP = "unknown"
			}

			if !rl.Allow(clientIP) {
				rl.log.WithContext(ctx).Warnf("Rate limit exceeded for IP: %s", clientIP)
				return nil, ErrRateLimitExceeded
			}

			return handler(ctx, req)
		}
	}
}

// HTTPHandler 返回HTTP中间件处理函数
func (rl *RateLimiter) HTTPHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		clientIP := getClientIP(r)

		if !rl.Allow(clientIP) {
			rl.log.Warnf("Rate limit exceeded for IP: %s, path: %s", clientIP, r.URL.Path)
			
			// 设置限流相关的响应头
			requests, remaining := rl.GetClientStats(clientIP)
			w.Header().Set("X-RateLimit-Limit", string(rune(rl.config.RequestsPerIP)))
			w.Header().Set("X-RateLimit-Remaining", string(rune(remaining)))
			w.Header().Set("X-RateLimit-Reset", time.Now().Add(rl.config.Window).Format(time.RFC3339))
			
			writeRateLimitResponse(w, requests, remaining, rl.config.Window)
			return
		}

		// 设置限流相关的响应头
		_, remaining := rl.GetClientStats(clientIP)
		w.Header().Set("X-RateLimit-Limit", string(rune(rl.config.RequestsPerIP)))
		w.Header().Set("X-RateLimit-Remaining", string(rune(remaining)))

		next.ServeHTTP(w, r)
	})
}

// writeRateLimitResponse 写入限流响应
func writeRateLimitResponse(w http.ResponseWriter, requests, remaining int, window time.Duration) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Retry-After", string(rune(int(window.Seconds()))))
	w.WriteHeader(http.StatusTooManyRequests)
	
	json.NewEncoder(w).Encode(map[string]interface{}{
		"code":      "RATE_LIMIT_EXCEEDED",
		"message":   "Too many requests, please try again later",
		"requests":  requests,
		"remaining": remaining,
		"retry_after": int(window.Seconds()),
	})
}

// ErrRateLimitExceeded 限流错误
var ErrRateLimitExceeded = &RateLimitError{
	Code:    "RATE_LIMIT_EXCEEDED",
	Message: "Too many requests, please try again later",
}

// RateLimitError 限流错误
type RateLimitError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func (e *RateLimitError) Error() string {
	return e.Message
}


// EndpointRateLimitConfig 端点限流配置
type EndpointRateLimitConfig struct {
	Path          string        `json:"path"`
	Method        string        `json:"method"`
	RequestsPerIP int           `json:"requests_per_ip"`
	Window        time.Duration `json:"window"`
}

// EndpointRateLimiter 端点级别限流器
type EndpointRateLimiter struct {
	defaultConfig *RateLimitConfig
	endpoints     map[string]*RateLimitConfig // path:method -> config
	limiters      map[string]*RateLimiter     // path:method:ip -> limiter
	mu            sync.RWMutex
	log           *log.Helper
	logger        log.Logger
}

// NewEndpointRateLimiter 创建端点级别限流器
func NewEndpointRateLimiter(defaultConfig *RateLimitConfig, logger log.Logger) *EndpointRateLimiter {
	if defaultConfig == nil {
		defaultConfig = DefaultRateLimitConfig()
	}

	erl := &EndpointRateLimiter{
		defaultConfig: defaultConfig,
		endpoints:     make(map[string]*RateLimitConfig),
		limiters:      make(map[string]*RateLimiter),
		log:           log.NewHelper(logger),
		logger:        logger,
	}

	// 设置默认的端点限流配置
	erl.setupDefaultEndpoints()

	return erl
}

// setupDefaultEndpoints 设置默认的端点限流配置
func (erl *EndpointRateLimiter) setupDefaultEndpoints() {
	// 登录接口限制更严格（防止暴力破解）
	erl.SetEndpointConfig("/api/v1/login", "POST", &RateLimitConfig{
		Enabled:       true,
		RequestsPerIP: 10,
		BurstSize:     5,
		Window:        time.Minute,
	})

	// 批量操作接口限制
	erl.SetEndpointConfig("/api/v1/datasources/batch", "POST", &RateLimitConfig{
		Enabled:       true,
		RequestsPerIP: 20,
		BurstSize:     5,
		Window:        time.Minute,
	})

	// 导出接口限制
	erl.SetEndpointConfig("/api/v1/datasources/export", "GET", &RateLimitConfig{
		Enabled:       true,
		RequestsPerIP: 10,
		BurstSize:     3,
		Window:        time.Minute,
	})

	// 任务执行接口限制
	erl.SetEndpointConfig("/api/v1/tasks/*/start", "POST", &RateLimitConfig{
		Enabled:       true,
		RequestsPerIP: 30,
		BurstSize:     10,
		Window:        time.Minute,
	})
}

// SetEndpointConfig 设置端点限流配置
func (erl *EndpointRateLimiter) SetEndpointConfig(path, method string, config *RateLimitConfig) {
	key := path + ":" + method
	erl.mu.Lock()
	defer erl.mu.Unlock()
	erl.endpoints[key] = config
}

// getConfigForEndpoint 获取端点的限流配置
func (erl *EndpointRateLimiter) getConfigForEndpoint(path, method string) *RateLimitConfig {
	erl.mu.RLock()
	defer erl.mu.RUnlock()

	// 精确匹配
	key := path + ":" + method
	if config, ok := erl.endpoints[key]; ok {
		return config
	}

	// 通配符匹配
	for pattern, config := range erl.endpoints {
		patternPath := pattern[:len(pattern)-len(method)-1]
		patternMethod := pattern[len(pattern)-len(method):]
		if patternMethod == method && matchPath(patternPath, path) {
			return config
		}
	}

	return erl.defaultConfig
}

// Allow 检查是否允许请求
func (erl *EndpointRateLimiter) Allow(clientIP, path, method string) bool {
	config := erl.getConfigForEndpoint(path, method)
	if config == nil || !config.Enabled {
		return true
	}

	// 获取或创建限流器
	key := path + ":" + method + ":" + clientIP
	erl.mu.Lock()
	limiter, exists := erl.limiters[key]
	if !exists {
		limiter = NewRateLimiter(config, erl.logger)
		erl.limiters[key] = limiter
	}
	erl.mu.Unlock()

	return limiter.Allow(clientIP)
}

// HTTPHandler 返回HTTP中间件处理函数
func (erl *EndpointRateLimiter) HTTPHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		clientIP := getClientIP(r)

		if !erl.Allow(clientIP, r.URL.Path, r.Method) {
			erl.log.Warnf("Rate limit exceeded for IP: %s, path: %s, method: %s",
				clientIP, r.URL.Path, r.Method)

			config := erl.getConfigForEndpoint(r.URL.Path, r.Method)
			writeRateLimitResponse(w, config.RequestsPerIP, 0, config.Window)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// Cleanup 清理过期的限流器
func (erl *EndpointRateLimiter) Cleanup() {
	erl.mu.Lock()
	defer erl.mu.Unlock()

	// 清理所有限流器
	for key, limiter := range erl.limiters {
		limiter.Stop()
		delete(erl.limiters, key)
	}
}

// SlidingWindowRateLimiter 滑动窗口限流器
type SlidingWindowRateLimiter struct {
	config    *RateLimitConfig
	windows   map[string][]time.Time // clientIP -> request timestamps
	mu        sync.RWMutex
	log       *log.Helper
}

// NewSlidingWindowRateLimiter 创建滑动窗口限流器
func NewSlidingWindowRateLimiter(config *RateLimitConfig, logger log.Logger) *SlidingWindowRateLimiter {
	if config == nil {
		config = DefaultRateLimitConfig()
	}

	return &SlidingWindowRateLimiter{
		config:  config,
		windows: make(map[string][]time.Time),
		log:     log.NewHelper(logger),
	}
}

// Allow 检查是否允许请求（滑动窗口算法）
func (rl *SlidingWindowRateLimiter) Allow(clientIP string) bool {
	if !rl.config.Enabled {
		return true
	}

	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	windowStart := now.Add(-rl.config.Window)

	// 获取或初始化窗口
	timestamps, exists := rl.windows[clientIP]
	if !exists {
		timestamps = make([]time.Time, 0)
	}

	// 移除过期的时间戳
	validTimestamps := make([]time.Time, 0)
	for _, ts := range timestamps {
		if ts.After(windowStart) {
			validTimestamps = append(validTimestamps, ts)
		}
	}

	// 检查是否超过限制
	if len(validTimestamps) >= rl.config.RequestsPerIP {
		rl.windows[clientIP] = validTimestamps
		return false
	}

	// 添加当前请求时间戳
	validTimestamps = append(validTimestamps, now)
	rl.windows[clientIP] = validTimestamps

	return true
}

// HTTPHandler 返回HTTP中间件处理函数
func (rl *SlidingWindowRateLimiter) HTTPHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		clientIP := getClientIP(r)

		if !rl.Allow(clientIP) {
			rl.log.Warnf("Rate limit exceeded for IP: %s, path: %s", clientIP, r.URL.Path)
			writeRateLimitResponse(w, rl.config.RequestsPerIP, 0, rl.config.Window)
			return
		}

		next.ServeHTTP(w, r)
	})
}
