package auth

import (
	"context"
)

// 上下文键类型
type contextKey string

const (
	userContextKey    contextKey = "user"
	tokenContextKey   contextKey = "token"
	requestIDKey      contextKey = "request_id"
	clientIPKey       contextKey = "client_ip"
	userAgentKey      contextKey = "user_agent"
)

// WithUser 将用户信息存入上下文
func WithUser(ctx context.Context, user *User) context.Context {
	return context.WithValue(ctx, userContextKey, user)
}

// UserFromContext 从上下文获取用户信息
func UserFromContext(ctx context.Context) (*User, bool) {
	user, ok := ctx.Value(userContextKey).(*User)
	return user, ok
}

// MustUserFromContext 从上下文获取用户信息，如果不存在则panic
func MustUserFromContext(ctx context.Context) *User {
	user, ok := UserFromContext(ctx)
	if !ok {
		panic("user not found in context")
	}
	return user
}

// WithToken 将令牌存入上下文
func WithToken(ctx context.Context, token string) context.Context {
	return context.WithValue(ctx, tokenContextKey, token)
}

// TokenFromContext 从上下文获取令牌
func TokenFromContext(ctx context.Context) (string, bool) {
	token, ok := ctx.Value(tokenContextKey).(string)
	return token, ok
}

// WithRequestID 将请求ID存入上下文
func WithRequestID(ctx context.Context, requestID string) context.Context {
	return context.WithValue(ctx, requestIDKey, requestID)
}

// RequestIDFromContext 从上下文获取请求ID
func RequestIDFromContext(ctx context.Context) string {
	requestID, ok := ctx.Value(requestIDKey).(string)
	if !ok {
		return ""
	}
	return requestID
}

// WithClientIP 将客户端IP存入上下文
func WithClientIP(ctx context.Context, ip string) context.Context {
	return context.WithValue(ctx, clientIPKey, ip)
}

// ClientIPFromContext 从上下文获取客户端IP
func ClientIPFromContext(ctx context.Context) string {
	ip, ok := ctx.Value(clientIPKey).(string)
	if !ok {
		return ""
	}
	return ip
}

// WithUserAgent 将User-Agent存入上下文
func WithUserAgent(ctx context.Context, userAgent string) context.Context {
	return context.WithValue(ctx, userAgentKey, userAgent)
}

// UserAgentFromContext 从上下文获取User-Agent
func UserAgentFromContext(ctx context.Context) string {
	userAgent, ok := ctx.Value(userAgentKey).(string)
	if !ok {
		return ""
	}
	return userAgent
}

// RequestInfo 请求信息
type RequestInfo struct {
	RequestID string
	ClientIP  string
	UserAgent string
	User      *User
}

// RequestInfoFromContext 从上下文获取请求信息
func RequestInfoFromContext(ctx context.Context) *RequestInfo {
	info := &RequestInfo{
		RequestID: RequestIDFromContext(ctx),
		ClientIP:  ClientIPFromContext(ctx),
		UserAgent: UserAgentFromContext(ctx),
	}
	if user, ok := UserFromContext(ctx); ok {
		info.User = user
	}
	return info
}
