// Package services 定义业务逻辑层
package services

import (
	"fmt"

	"com.litelake.litecore/common"
	"com.litelake.litecore/config"
)

// AuthService 认证服务
type AuthService struct {
	Config         common.BaseConfigProvider `inject:""`
	SessionService ISessionService           `inject:""`
}

// NewAuthService 创建认证服务实例
func NewAuthService() *AuthService {
	return &AuthService{}
}

// ServiceName 实现 BaseService 接口
func (s *AuthService) ServiceName() string {
	return "AuthService"
}

// OnStart 实现 BaseService 接口
func (s *AuthService) OnStart() error {
	return nil
}

// OnStop 实现 BaseService 接口
func (s *AuthService) OnStop() error {
	return nil
}

// VerifyPassword 验证管理员密码
func (s *AuthService) VerifyPassword(password string) bool {
	// 从配置读取管理员密码
	storedPassword, err := config.Get[string](s.Config, "app.admin.password")
	if err != nil {
		return false
	}
	return password == storedPassword
}

// Login 管理员登录
func (s *AuthService) Login(password string) (string, error) {
	if !s.VerifyPassword(password) {
		return "", fmt.Errorf("invalid password")
	}

	// 创建会话
	token, err := s.SessionService.CreateSession()
	if err != nil {
		return "", fmt.Errorf("failed to create session: %w", err)
	}

	return token, nil
}

// Logout 管理员登出
func (s *AuthService) Logout(token string) error {
	return s.SessionService.DeleteSession(token)
}

// ValidateToken 验证令牌
func (s *AuthService) ValidateToken(token string) (*AdminSession, error) {
	return s.SessionService.ValidateSession(token)
}

var _ IAuthService = (*AuthService)(nil)
