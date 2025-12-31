package service

import (
	"context"
	"time"

	v1 "go-metadata/api/metadata/v1"

	"github.com/go-kratos/kratos/v2/log"
)

// UserService implements v1.UserServiceServer
type UserService struct {
	v1.UnimplementedUserServiceServer
	log *log.Helper
}

// NewUserService creates a new UserService
func NewUserService(logger log.Logger) *UserService {
	return &UserService{
		log: log.NewHelper(logger),
	}
}

// Login handles user login
func (s *UserService) Login(ctx context.Context, req *v1.LoginRequest) (*v1.LoginResponse, error) {
	s.log.WithContext(ctx).Infof("User login attempt: %s", req.Username)

	// Simple authentication - in production, use proper auth
	if req.Username == "admin" && req.Password == "ant.design" {
		return &v1.LoginResponse{
			Status:           "ok",
			Type:             req.Type,
			CurrentAuthority: "admin",
			Token:            "admin-token-" + time.Now().Format("20060102150405"),
		}, nil
	}

	if req.Username == "user" && req.Password == "ant.design" {
		return &v1.LoginResponse{
			Status:           "ok",
			Type:             req.Type,
			CurrentAuthority: "user",
			Token:            "user-token-" + time.Now().Format("20060102150405"),
		}, nil
	}

	return &v1.LoginResponse{
		Status:           "error",
		Type:             req.Type,
		CurrentAuthority: "guest",
	}, nil
}

// Logout handles user logout
func (s *UserService) Logout(ctx context.Context, req *v1.LogoutRequest) (*v1.LogoutResponse, error) {
	s.log.WithContext(ctx).Info("User logout")
	return &v1.LogoutResponse{Success: true}, nil
}

// GetCurrentUser returns current user info
func (s *UserService) GetCurrentUser(ctx context.Context, req *v1.GetCurrentUserRequest) (*v1.GetCurrentUserResponse, error) {
	return &v1.GetCurrentUserResponse{
		Data: &v1.CurrentUser{
			Name:    "Admin",
			Avatar:  "https://gw.alipayobjects.com/zos/antfincdn/XAosXuNZyF/BiazfanxmamNRoxxVxka.png",
			Userid:  "00000001",
			Email:   "admin@example.com",
			Title:   "系统管理员",
			Group:   "元数据管理平台",
			Access:  "admin",
			Address: "北京市",
			Phone:   "1380000000",
			Tags: []*v1.UserTag{
				{Key: "1", Label: "管理员"},
			},
			Country: "China",
		},
	}, nil
}

// GetNotices returns user notices
func (s *UserService) GetNotices(ctx context.Context, req *v1.GetNoticesRequest) (*v1.GetNoticesResponse, error) {
	return &v1.GetNoticesResponse{
		Data: []*v1.Notice{
			{
				Id:       "1",
				Title:    "系统初始化完成",
				Type:     "notification",
				Read:     false,
				Datetime: time.Now().Format("2006-01-02 15:04:05"),
			},
		},
	}, nil
}
