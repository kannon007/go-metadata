package auth

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/middleware"
	"github.com/go-kratos/kratos/v2/transport"
	khttp "github.com/go-kratos/kratos/v2/transport/http"
	"github.com/google/uuid"
)

// AuthMiddleware 认证中间件配置
type AuthMiddleware struct {
	jwtManager   *JWTManager
	skipPaths    map[string]bool
	log          *log.Helper
}

// NewAuthMiddleware 创建认证中间件
func NewAuthMiddleware(jwtManager *JWTManager, logger log.Logger) *AuthMiddleware {
	return &AuthMiddleware{
		jwtManager: jwtManager,
		skipPaths: map[string]bool{
			"/health":         true,
			"/api/v1/login":   true,
			"/api/v1/refresh": true,
		},
		log: log.NewHelper(logger),
	}
}

// AddSkipPath 添加跳过认证的路径
func (m *AuthMiddleware) AddSkipPath(path string) {
	m.skipPaths[path] = true
}

// Handler 返回Kratos中间件处理函数
func (m *AuthMiddleware) Handler() middleware.Middleware {
	return func(handler middleware.Handler) middleware.Handler {
		return func(ctx context.Context, req interface{}) (interface{}, error) {
			// 生成请求ID
			requestID := uuid.New().String()
			ctx = WithRequestID(ctx, requestID)

			// 获取传输信息
			if tr, ok := transport.FromServerContext(ctx); ok {
				// 获取客户端IP和User-Agent
				if ht, ok := tr.(*khttp.Transport); ok {
					ctx = WithClientIP(ctx, getClientIP(ht.Request()))
					ctx = WithUserAgent(ctx, ht.Request().UserAgent())
				}

				// 检查是否跳过认证
				path := tr.Operation()
				if m.skipPaths[path] {
					return handler(ctx, req)
				}

				// 获取Authorization头
				header := tr.RequestHeader()
				authHeader := header.Get("Authorization")
				if authHeader == "" {
					return nil, ErrTokenNotFound
				}

				// 解析Bearer令牌
				token := strings.TrimPrefix(authHeader, "Bearer ")
				if token == authHeader {
					return nil, ErrInvalidToken
				}

				// 验证令牌
				user, err := m.jwtManager.ValidateToken(token)
				if err != nil {
					m.log.WithContext(ctx).Warnf("Token validation failed: %v", err)
					return nil, err
				}

				// 检查用户是否启用
				if !user.Enabled {
					return nil, ErrPermissionDenied
				}

				// 将用户信息和令牌存入上下文
				ctx = WithUser(ctx, user)
				ctx = WithToken(ctx, token)
			}

			return handler(ctx, req)
		}
	}
}

// HTTPHandler 返回HTTP中间件处理函数
func (m *AuthMiddleware) HTTPHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 生成请求ID
		requestID := uuid.New().String()
		ctx := WithRequestID(r.Context(), requestID)
		ctx = WithClientIP(ctx, getClientIP(r))
		ctx = WithUserAgent(ctx, r.UserAgent())

		// 设置响应头
		w.Header().Set("X-Request-ID", requestID)

		// 检查是否跳过认证
		if m.skipPaths[r.URL.Path] {
			next.ServeHTTP(w, r.WithContext(ctx))
			return
		}

		// 获取Authorization头
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			writeErrorResponse(w, http.StatusUnauthorized, "TOKEN_NOT_FOUND", "Authorization header is required")
			return
		}

		// 解析Bearer令牌
		token := strings.TrimPrefix(authHeader, "Bearer ")
		if token == authHeader {
			writeErrorResponse(w, http.StatusUnauthorized, "INVALID_TOKEN", "Invalid authorization header format")
			return
		}

		// 验证令牌
		user, err := m.jwtManager.ValidateToken(token)
		if err != nil {
			m.log.WithContext(ctx).Warnf("Token validation failed: %v", err)
			switch err {
			case ErrTokenExpired:
				writeErrorResponse(w, http.StatusUnauthorized, "TOKEN_EXPIRED", "Token has expired")
			case ErrInvalidToken, ErrInvalidClaims:
				writeErrorResponse(w, http.StatusUnauthorized, "INVALID_TOKEN", "Invalid token")
			default:
				writeErrorResponse(w, http.StatusUnauthorized, "AUTH_ERROR", err.Error())
			}
			return
		}

		// 检查用户是否启用
		if !user.Enabled {
			writeErrorResponse(w, http.StatusForbidden, "USER_DISABLED", "User account is disabled")
			return
		}

		// 将用户信息和令牌存入上下文
		ctx = WithUser(ctx, user)
		ctx = WithToken(ctx, token)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// getClientIP 获取客户端IP
func getClientIP(r *http.Request) string {
	// 优先从X-Forwarded-For获取
	xff := r.Header.Get("X-Forwarded-For")
	if xff != "" {
		ips := strings.Split(xff, ",")
		if len(ips) > 0 {
			return strings.TrimSpace(ips[0])
		}
	}

	// 其次从X-Real-IP获取
	xri := r.Header.Get("X-Real-IP")
	if xri != "" {
		return xri
	}

	// 最后从RemoteAddr获取
	ip := r.RemoteAddr
	if idx := strings.LastIndex(ip, ":"); idx != -1 {
		ip = ip[:idx]
	}
	return ip
}

// writeErrorResponse 写入错误响应
func writeErrorResponse(w http.ResponseWriter, statusCode int, code, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"code":    code,
		"message": message,
	})
}
