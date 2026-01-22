// Package services 定义业务逻辑层
package services

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	
	"github.com/lite-lake/litecore-go/common"
	"github.com/lite-lake/litecore-go/samples/messageboard/internal/dtos"
	"github.com/lite-lake/litecore-go/server/builtin/manager/cachemgr"
	"github.com/lite-lake/litecore-go/server/builtin/manager/configmgr"
	"github.com/lite-lake/litecore-go/server/builtin/manager/loggermgr"
)

// ISessionService 会话服务接口
type ISessionService interface {
	common.IBaseService
	CreateSession() (string, error)
	ValidateSession(token string) (*dtos.AdminSession, error)
	DeleteSession(token string) error
}

type sessionService struct {
	Config    configmgr.IConfigManager `inject:""`
	CacheMgr  cachemgr.ICacheManager   `inject:""`
	LoggerMgr loggermgr.ILoggerManager `inject:""`
	timeout   time.Duration
}

// NewSessionService 创建会话服务
func NewSessionService() ISessionService {
	return &sessionService{}
}

func (s *sessionService) ServiceName() string {
	return "SessionService"
}

func (s *sessionService) OnStart() error {
	timeout, err := configmgr.Get[int](s.Config, "app.admin.session_timeout")
	if err != nil {
		return fmt.Errorf("failed to get session_timeout from configmgr: %w", err)
	}
	s.timeout = time.Duration(timeout) * time.Second
	return nil
}

func (s *sessionService) OnStop() error {
	return nil
}

func (s *sessionService) CreateSession() (string, error) {
	token := uuid.New().String()
	session := &dtos.AdminSession{
		Token:     token,
		ExpiresAt: time.Now().Add(s.timeout),
	}

	ctx := context.Background()
	sessionKey := fmt.Sprintf("session:%s", token)
	if err := s.CacheMgr.Set(ctx, sessionKey, session, s.timeout); err != nil {
		s.LoggerMgr.Ins().Error("创建会话失败", "token", token, "error", err)
		return "", fmt.Errorf("failed to store session: %w", err)
	}
	s.LoggerMgr.Ins().Info("创建会话成功", "token", token, "expires_at", session.ExpiresAt)

	return token, nil
}

func (s *sessionService) ValidateSession(token string) (*dtos.AdminSession, error) {
	ctx := context.Background()
	sessionKey := fmt.Sprintf("session:%s", token)

	var session dtos.AdminSession
	if err := s.CacheMgr.Get(ctx, sessionKey, &session); err != nil {
		s.LoggerMgr.Ins().Warn("验证会话失败：会话不存在", "token", token)
		return nil, errors.New("session not found")
	}

	if time.Now().After(session.ExpiresAt) {
		s.LoggerMgr.Ins().Warn("验证会话失败：会话已过期", "token", token)
		s.DeleteSession(token)
		return nil, errors.New("session expired")
	}

	s.LoggerMgr.Ins().Debug("验证会话成功", "token", token)

	return &session, nil
}

func (s *sessionService) DeleteSession(token string) error {
	s.LoggerMgr.Ins().Info("删除会话", "token", token)

	ctx := context.Background()
	sessionKey := fmt.Sprintf("session:%s", token)
	return s.CacheMgr.Delete(ctx, sessionKey)
}

var _ ISessionService = (*sessionService)(nil)
