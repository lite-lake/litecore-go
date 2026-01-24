// Package services 定义业务逻辑层
package services

import (
	"fmt"

	"github.com/lite-lake/litecore-go/common"
	"github.com/lite-lake/litecore-go/manager/configmgr"
	"github.com/lite-lake/litecore-go/manager/loggermgr"
	"github.com/lite-lake/litecore-go/samples/messageboard/internal/dtos"
	"github.com/lite-lake/litecore-go/util/hash"
)

// IAuthService 认证服务接口
type IAuthService interface {
	common.IBaseService
	VerifyPassword(password string) bool                    // 验证管理员密码
	Login(password string) (string, error)                  // 管理员登录，返回 token
	Logout(token string) error                              // 管理员退出登录
	ValidateToken(token string) (*dtos.AdminSession, error) // 验证 token 有效性
}

type authServiceImpl struct {
	Config         configmgr.IConfigManager `inject:""` // 配置管理器
	LoggerMgr      loggermgr.ILoggerManager `inject:""` // 日志管理器
	SessionService ISessionService          `inject:""` // 会话服务
}

// NewAuthService 创建认证服务实例
func NewAuthService() IAuthService {
	return &authServiceImpl{}
}

// ServiceName 返回服务名称
func (s *authServiceImpl) ServiceName() string {
	return "AuthService"
}

// OnStart 启动时初始化
func (s *authServiceImpl) OnStart() error {
	return nil
}

// OnStop 停止时清理
func (s *authServiceImpl) OnStop() error {
	return nil
}

// VerifyPassword 验证管理员密码是否正确
func (s *authServiceImpl) VerifyPassword(password string) bool {
	storedPassword, err := configmgr.Get[string](s.Config, "app.admin.password")
	if err != nil {
		s.LoggerMgr.Ins().Error("Failed to get admin password", "error", err)
		return false
	}
	return hash.Hash.BcryptVerify(password, storedPassword)
}

// Login 管理员登录，验证密码后创建会话
func (s *authServiceImpl) Login(password string) (string, error) {
	if !s.VerifyPassword(password) {
		s.LoggerMgr.Ins().Warn("Login failed: invalid password")
		return "", fmt.Errorf("invalid password")
	}

	token, err := s.SessionService.CreateSession()
	if err != nil {
		s.LoggerMgr.Ins().Error("Login failed: failed to create session", "error", err)
		return "", fmt.Errorf("failed to create session: %w", err)
	}

	s.LoggerMgr.Ins().Info("Login successful", "token", token)

	return token, nil
}

// Logout 管理员退出登录，删除会话
func (s *authServiceImpl) Logout(token string) error {
	s.LoggerMgr.Ins().Info("Logout", "token", token)
	return s.SessionService.DeleteSession(token)
}

// ValidateToken 验证 token 有效性
func (s *authServiceImpl) ValidateToken(token string) (*dtos.AdminSession, error) {
	return s.SessionService.ValidateSession(token)
}

var _ IAuthService = (*authServiceImpl)(nil)
