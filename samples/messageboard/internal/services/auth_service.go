// Package services 定义业务逻辑层
package services

import (
	"fmt"

	"github.com/lite-lake/litecore-go/common"
	"github.com/lite-lake/litecore-go/component/manager/loggermgr"
	"github.com/lite-lake/litecore-go/config"
	"github.com/lite-lake/litecore-go/samples/messageboard/internal/dtos"
	"github.com/lite-lake/litecore-go/util/hash"
)

// IAuthService 认证服务接口
type IAuthService interface {
	common.IBaseService
	VerifyPassword(password string) bool
	Login(password string) (string, error)
	Logout(token string) error
	ValidateToken(token string) (*dtos.AdminSession, error)
}

type authService struct {
	Config         common.IBaseConfigProvider `inject:""`
	SessionService ISessionService            `inject:""`
	LoggerMgr      loggermgr.ILoggerManager   `inject:""`
	logger         loggermgr.ILogger
}

// NewAuthService 创建认证服务
func NewAuthService() IAuthService {
	return &authService{}
}

func (s *authService) ServiceName() string {
	return "AuthService"
}

func (s *authService) OnStart() error {
	s.initLogger()
	return nil
}

func (s *authService) OnStop() error {
	return nil
}

func (s *authService) Logger() loggermgr.ILogger {
	return s.logger
}

func (s *authService) SetLoggerManager(mgr loggermgr.ILoggerManager) {
	s.LoggerMgr = mgr
	s.initLogger()
}

func (s *authService) initLogger() {
	if s.LoggerMgr != nil {
		s.logger = s.LoggerMgr.Logger("AuthService")
	}
}

func (s *authService) VerifyPassword(password string) bool {
	storedPassword, err := config.Get[string](s.Config, "app.admin.password")
	if err != nil {
		if s.logger != nil {
			s.logger.Error("获取管理员密码失败", "error", err)
		}
		return false
	}
	return hash.Hash.BcryptVerify(password, storedPassword)
}

func (s *authService) Login(password string) (string, error) {
	if !s.VerifyPassword(password) {
		if s.logger != nil {
			s.logger.Warn("登录失败：密码错误")
		}
		return "", fmt.Errorf("invalid password")
	}

	token, err := s.SessionService.CreateSession()
	if err != nil {
		if s.logger != nil {
			s.logger.Error("登录失败：创建会话失败", "error", err)
		}
		return "", fmt.Errorf("failed to create session: %w", err)
	}

	if s.logger != nil {
		s.logger.Info("登录成功", "token", token)
	}

	return token, nil
}

func (s *authService) Logout(token string) error {
	if s.logger != nil {
		s.logger.Info("退出登录", "token", token)
	}
	return s.SessionService.DeleteSession(token)
}

func (s *authService) ValidateToken(token string) (*dtos.AdminSession, error) {
	return s.SessionService.ValidateSession(token)
}

var _ IAuthService = (*authService)(nil)
