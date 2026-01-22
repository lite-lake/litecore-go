// Package services 定义业务逻辑层
package services

import (
	"fmt"

	"github.com/lite-lake/litecore-go/common"
	"github.com/lite-lake/litecore-go/samples/messageboard/internal/dtos"
	"github.com/lite-lake/litecore-go/server/builtin/manager/configmgr"
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
	Config         configmgr.IConfigManager `inject:""`
	SessionService ISessionService          `inject:""`
	Logger         common.ILogger           `inject:""`
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
	storedPassword, err := configmgr.Get[string](s.Config, "app.admin.password")
	if err != nil {
		if s.Logger != nil {
			s.Logger.Error("获取管理员密码失败", "error", err)
		}
		return false
	}
	return hash.Hash.BcryptVerify(password, storedPassword)
}

func (s *authService) Login(password string) (string, error) {
	if !s.VerifyPassword(password) {
		if s.Logger != nil {
			s.Logger.Warn("登录失败：密码错误")
		}
		return "", fmt.Errorf("invalid password")
	}

	token, err := s.SessionService.CreateSession()
	if err != nil {
		if s.Logger != nil {
			s.Logger.Error("登录失败：创建会话失败", "error", err)
		}
		return "", fmt.Errorf("failed to create session: %w", err)
	}

	if s.Logger != nil {
		s.Logger.Info("登录成功", "token", token)
	}

	return token, nil
}

func (s *authService) Logout(token string) error {
	if s.Logger != nil {
		s.Logger.Info("退出登录", "token", token)
	}
	return s.SessionService.DeleteSession(token)
}

func (s *authService) ValidateToken(token string) (*dtos.AdminSession, error) {
	return s.SessionService.ValidateSession(token)
}

var _ IAuthService = (*authService)(nil)
