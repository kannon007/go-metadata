package auth

import (
	"context"
	"net/http"
	"strings"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/middleware"
)

// RBACMiddleware RBAC权限中间件
type RBACMiddleware struct {
	pathPermissions map[string]map[string]Permission // path -> method -> permission
	log             *log.Helper
}

// NewRBACMiddleware 创建RBAC中间件
func NewRBACMiddleware(logger log.Logger) *RBACMiddleware {
	m := &RBACMiddleware{
		pathPermissions: make(map[string]map[string]Permission),
		log:             log.NewHelper(logger),
	}
	m.setupDefaultPermissions()
	return m
}

// setupDefaultPermissions 设置默认权限映射
func (m *RBACMiddleware) setupDefaultPermissions() {
	// 数据源API权限
	m.AddPermission("/api/v1/datasources", "GET", PermissionDataSourceRead)
	m.AddPermission("/api/v1/datasources", "POST", PermissionDataSourceCreate)
	m.AddPermission("/api/v1/datasources/*", "GET", PermissionDataSourceRead)
	m.AddPermission("/api/v1/datasources/*", "PUT", PermissionDataSourceUpdate)
	m.AddPermission("/api/v1/datasources/*", "DELETE", PermissionDataSourceDelete)
	m.AddPermission("/api/v1/datasources/*/test", "POST", PermissionDataSourceRead)
	m.AddPermission("/api/v1/datasources/batch", "POST", PermissionDataSourceUpdate)
	m.AddPermission("/api/v1/datasources/export", "GET", PermissionDataSourceRead)
	m.AddPermission("/api/v1/datasources/import", "POST", PermissionDataSourceCreate)

	// 任务API权限
	m.AddPermission("/api/v1/tasks", "GET", PermissionTaskRead)
	m.AddPermission("/api/v1/tasks", "POST", PermissionTaskCreate)
	m.AddPermission("/api/v1/tasks/*", "GET", PermissionTaskRead)
	m.AddPermission("/api/v1/tasks/*", "PUT", PermissionTaskUpdate)
	m.AddPermission("/api/v1/tasks/*", "DELETE", PermissionTaskDelete)
	m.AddPermission("/api/v1/tasks/*/start", "POST", PermissionTaskExecute)
	m.AddPermission("/api/v1/tasks/*/stop", "POST", PermissionTaskExecute)
	m.AddPermission("/api/v1/tasks/*/pause", "POST", PermissionTaskExecute)
	m.AddPermission("/api/v1/tasks/*/resume", "POST", PermissionTaskExecute)
	m.AddPermission("/api/v1/tasks/*/retry", "POST", PermissionTaskExecute)

	// 审计日志API权限
	m.AddPermission("/api/v1/audit", "GET", PermissionAuditRead)
	m.AddPermission("/api/v1/audit/*", "GET", PermissionAuditRead)

	// 系统管理API权限
	m.AddPermission("/api/v1/system/*", "GET", PermissionSystemAdmin)
	m.AddPermission("/api/v1/system/*", "POST", PermissionSystemAdmin)
	m.AddPermission("/api/v1/system/*", "PUT", PermissionSystemAdmin)
	m.AddPermission("/api/v1/system/*", "DELETE", PermissionSystemAdmin)
}

// AddPermission 添加路径权限映射
func (m *RBACMiddleware) AddPermission(path, method string, permission Permission) {
	if m.pathPermissions[path] == nil {
		m.pathPermissions[path] = make(map[string]Permission)
	}
	m.pathPermissions[path][method] = permission
}

// GetRequiredPermission 获取路径所需的权限
func (m *RBACMiddleware) GetRequiredPermission(path, method string) (Permission, bool) {
	// 精确匹配
	if methods, ok := m.pathPermissions[path]; ok {
		if perm, ok := methods[method]; ok {
			return perm, true
		}
	}

	// 通配符匹配
	for pattern, methods := range m.pathPermissions {
		if matchPath(pattern, path) {
			if perm, ok := methods[method]; ok {
				return perm, true
			}
		}
	}

	return "", false
}

// matchPath 路径匹配（支持*通配符）
func matchPath(pattern, path string) bool {
	patternParts := strings.Split(pattern, "/")
	pathParts := strings.Split(path, "/")

	if len(patternParts) != len(pathParts) {
		return false
	}

	for i, part := range patternParts {
		if part == "*" {
			continue
		}
		if part != pathParts[i] {
			return false
		}
	}

	return true
}

// Handler 返回Kratos中间件处理函数
func (m *RBACMiddleware) Handler() middleware.Middleware {
	return func(handler middleware.Handler) middleware.Handler {
		return func(ctx context.Context, req interface{}) (interface{}, error) {
			// 获取用户信息
			user, ok := UserFromContext(ctx)
			if !ok {
				// 如果没有用户信息，说明是跳过认证的路径，直接放行
				return handler(ctx, req)
			}

			// 管理员直接放行
			if user.IsAdmin() {
				return handler(ctx, req)
			}

			// 这里需要从transport获取路径和方法
			// 由于Kratos中间件的限制，具体的权限检查在HTTP层处理
			return handler(ctx, req)
		}
	}
}

// HTTPHandler 返回HTTP中间件处理函数
func (m *RBACMiddleware) HTTPHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		// 获取用户信息
		user, ok := UserFromContext(ctx)
		if !ok {
			// 如果没有用户信息，说明是跳过认证的路径，直接放行
			next.ServeHTTP(w, r)
			return
		}

		// 管理员直接放行
		if user.IsAdmin() {
			next.ServeHTTP(w, r)
			return
		}

		// 获取所需权限
		requiredPerm, found := m.GetRequiredPermission(r.URL.Path, r.Method)
		if !found {
			// 如果没有配置权限要求，默认放行
			next.ServeHTTP(w, r)
			return
		}

		// 检查用户是否拥有所需权限
		if !user.HasPermission(requiredPerm) {
			m.log.WithContext(ctx).Warnf("Permission denied: user=%s, path=%s, method=%s, required=%s",
				user.Username, r.URL.Path, r.Method, requiredPerm)
			writeErrorResponse(w, http.StatusForbidden, "PERMISSION_DENIED",
				"You don't have permission to access this resource")
			return
		}

		next.ServeHTTP(w, r)
	})
}

// RequirePermission 创建需要特定权限的中间件
func RequirePermission(permission Permission, logger log.Logger) func(http.Handler) http.Handler {
	log := log.NewHelper(logger)
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()

			// 获取用户信息
			user, ok := UserFromContext(ctx)
			if !ok {
				writeErrorResponse(w, http.StatusUnauthorized, "UNAUTHORIZED", "Authentication required")
				return
			}

			// 管理员直接放行
			if user.IsAdmin() {
				next.ServeHTTP(w, r)
				return
			}

			// 检查权限
			if !user.HasPermission(permission) {
				log.WithContext(ctx).Warnf("Permission denied: user=%s, required=%s", user.Username, permission)
				writeErrorResponse(w, http.StatusForbidden, "PERMISSION_DENIED",
					"You don't have permission to access this resource")
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// RequireRole 创建需要特定角色的中间件
func RequireRole(role Role, logger log.Logger) func(http.Handler) http.Handler {
	log := log.NewHelper(logger)
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()

			// 获取用户信息
			user, ok := UserFromContext(ctx)
			if !ok {
				writeErrorResponse(w, http.StatusUnauthorized, "UNAUTHORIZED", "Authentication required")
				return
			}

			// 检查角色
			if !user.HasRole(role) {
				log.WithContext(ctx).Warnf("Role denied: user=%s, required=%s", user.Username, role)
				writeErrorResponse(w, http.StatusForbidden, "ROLE_DENIED",
					"You don't have the required role to access this resource")
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// RequireAdmin 创建需要管理员角色的中间件
func RequireAdmin(logger log.Logger) func(http.Handler) http.Handler {
	return RequireRole(RoleAdmin, logger)
}
