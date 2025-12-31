package auth

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// 认证相关错误
var (
	ErrInvalidToken     = errors.New("invalid token")
	ErrTokenExpired     = errors.New("token expired")
	ErrTokenNotFound    = errors.New("token not found")
	ErrInvalidClaims    = errors.New("invalid claims")
	ErrPermissionDenied = errors.New("permission denied")
	ErrUserNotFound     = errors.New("user not found")
	ErrInvalidPassword  = errors.New("invalid password")
)

// Role 用户角色
type Role string

const (
	RoleAdmin    Role = "admin"    // 管理员：拥有所有权限
	RoleOperator Role = "operator" // 操作员：可以管理数据源和任务
	RoleViewer   Role = "viewer"   // 查看者：只能查看数据
)

// AllRoles 返回所有角色
func AllRoles() []Role {
	return []Role{RoleAdmin, RoleOperator, RoleViewer}
}

// IsValidRole 检查角色是否有效
func IsValidRole(r Role) bool {
	for _, validRole := range AllRoles() {
		if r == validRole {
			return true
		}
	}
	return false
}

// Permission 权限类型
type Permission string

const (
	// 数据源权限
	PermissionDataSourceCreate Permission = "datasource:create"
	PermissionDataSourceRead   Permission = "datasource:read"
	PermissionDataSourceUpdate Permission = "datasource:update"
	PermissionDataSourceDelete Permission = "datasource:delete"

	// 任务权限
	PermissionTaskCreate  Permission = "task:create"
	PermissionTaskRead    Permission = "task:read"
	PermissionTaskUpdate  Permission = "task:update"
	PermissionTaskDelete  Permission = "task:delete"
	PermissionTaskExecute Permission = "task:execute"

	// 系统权限
	PermissionSystemAdmin Permission = "system:admin"
	PermissionAuditRead   Permission = "audit:read"
)

// RolePermissions 角色权限映射
var RolePermissions = map[Role][]Permission{
	RoleAdmin: {
		PermissionDataSourceCreate, PermissionDataSourceRead, PermissionDataSourceUpdate, PermissionDataSourceDelete,
		PermissionTaskCreate, PermissionTaskRead, PermissionTaskUpdate, PermissionTaskDelete, PermissionTaskExecute,
		PermissionSystemAdmin, PermissionAuditRead,
	},
	RoleOperator: {
		PermissionDataSourceCreate, PermissionDataSourceRead, PermissionDataSourceUpdate,
		PermissionTaskCreate, PermissionTaskRead, PermissionTaskUpdate, PermissionTaskExecute,
	},
	RoleViewer: {
		PermissionDataSourceRead,
		PermissionTaskRead,
	},
}

// HasPermission 检查角色是否拥有指定权限
func HasPermission(role Role, permission Permission) bool {
	permissions, ok := RolePermissions[role]
	if !ok {
		return false
	}
	for _, p := range permissions {
		if p == permission {
			return true
		}
	}
	return false
}


// User 用户信息
type User struct {
	ID       string   `json:"id"`
	Username string   `json:"username"`
	Email    string   `json:"email"`
	Role     Role     `json:"role"`
	Roles    []Role   `json:"roles,omitempty"` // 支持多角色
	Enabled  bool     `json:"enabled"`
}

// HasRole 检查用户是否拥有指定角色
func (u *User) HasRole(role Role) bool {
	if u.Role == role {
		return true
	}
	for _, r := range u.Roles {
		if r == role {
			return true
		}
	}
	return false
}

// HasPermission 检查用户是否拥有指定权限
func (u *User) HasPermission(permission Permission) bool {
	// 检查主角色
	if HasPermission(u.Role, permission) {
		return true
	}
	// 检查附加角色
	for _, role := range u.Roles {
		if HasPermission(role, permission) {
			return true
		}
	}
	return false
}

// IsAdmin 检查用户是否是管理员
func (u *User) IsAdmin() bool {
	return u.HasRole(RoleAdmin)
}

// Claims JWT声明
type Claims struct {
	jwt.RegisteredClaims
	UserID   string `json:"user_id"`
	Username string `json:"username"`
	Email    string `json:"email"`
	Role     Role   `json:"role"`
	Roles    []Role `json:"roles,omitempty"`
}

// ToUser 转换为用户信息
func (c *Claims) ToUser() *User {
	return &User{
		ID:       c.UserID,
		Username: c.Username,
		Email:    c.Email,
		Role:     c.Role,
		Roles:    c.Roles,
		Enabled:  true,
	}
}

// JWTConfig JWT配置
type JWTConfig struct {
	Secret     string        `json:"secret"`
	Expire     time.Duration `json:"expire"`
	Issuer     string        `json:"issuer"`
	RefreshExp time.Duration `json:"refresh_exp"`
}

// JWTManager JWT管理器
type JWTManager struct {
	config *JWTConfig
}

// NewJWTManager 创建JWT管理器
func NewJWTManager(config *JWTConfig) *JWTManager {
	if config.Issuer == "" {
		config.Issuer = "go-metadata"
	}
	if config.Expire == 0 {
		config.Expire = 24 * time.Hour
	}
	if config.RefreshExp == 0 {
		config.RefreshExp = 7 * 24 * time.Hour
	}
	return &JWTManager{config: config}
}

// GenerateToken 生成JWT令牌
func (m *JWTManager) GenerateToken(user *User) (string, error) {
	now := time.Now()
	claims := &Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    m.config.Issuer,
			Subject:   user.ID,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(m.config.Expire)),
			NotBefore: jwt.NewNumericDate(now),
		},
		UserID:   user.ID,
		Username: user.Username,
		Email:    user.Email,
		Role:     user.Role,
		Roles:    user.Roles,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(m.config.Secret))
}

// GenerateRefreshToken 生成刷新令牌
func (m *JWTManager) GenerateRefreshToken(user *User) (string, error) {
	now := time.Now()
	claims := &Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    m.config.Issuer,
			Subject:   user.ID,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(m.config.RefreshExp)),
			NotBefore: jwt.NewNumericDate(now),
		},
		UserID:   user.ID,
		Username: user.Username,
		Role:     user.Role,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(m.config.Secret))
}

// ParseToken 解析JWT令牌
func (m *JWTManager) ParseToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrInvalidToken
		}
		return []byte(m.config.Secret), nil
	})

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrTokenExpired
		}
		return nil, ErrInvalidToken
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, ErrInvalidClaims
	}

	return claims, nil
}

// ValidateToken 验证令牌并返回用户信息
func (m *JWTManager) ValidateToken(tokenString string) (*User, error) {
	claims, err := m.ParseToken(tokenString)
	if err != nil {
		return nil, err
	}
	return claims.ToUser(), nil
}
