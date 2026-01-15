// Package services 定义业务逻辑层
package services

import (
	"fmt"

	"com.litelake.litecore/common"
	"com.litelake.litecore/config"
	"com.litelake.litecore/samples/messageboard/internal/dtos"
)

// IAuthService 认证服务接口
type IAuthService interface {
	common.BaseService
	VerifyPassword(password string) bool
	Login(password string) (string, error)
	Logout(token string) error
	ValidateToken(token string) (*dtos.AdminSession, error)
}

type authService struct {
	Config         common.BaseConfigProvider `inject:""`
	SessionService ISessionService           `inject:""`
}

// NewAuthService 创建认证服务
func NewAuthService() IAuthService {
	return &authService{}
}

func (s *authService) ServiceName() string {
	return "AuthService"
}

func (s *authService) OnStart() error {
	return nil
}

func (s *authService) OnStop() error {
	return nil
}

func (s *authService) VerifyPassword(password string) bool {
	storedPassword, err := config.Get[string](s.Config, "app.admin.password")
	if err != nil {
		return false
	}
	return password == storedPassword
}

func (s *authService) Login(password string) (string, error) {
	if !s.VerifyPassword(password) {
		return "", fmt.Errorf("invalid password")
	}

	token, err := s.SessionService.CreateSession()
	if err != nil {
		return "", fmt.Errorf("failed to create session: %w", err)
	}

	return token, nil
}

func (s *authService) Logout(token string) error {
	return s.SessionService.DeleteSession(token)
}

func (s *authService) ValidateToken(token string) (*dtos.AdminSession, error) {
	return s.SessionService.ValidateSession(token)
}

var _ IAuthService = (*authService)(nil)
