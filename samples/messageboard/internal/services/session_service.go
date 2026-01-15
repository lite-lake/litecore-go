// Package services 定义业务逻辑层
package services

import (
	"context"
	"errors"
	"fmt"
	"time"

	"com.litelake.litecore/common"
	"com.litelake.litecore/config"
	"com.litelake.litecore/manager/cachemgr"
	"com.litelake.litecore/samples/messageboard/internal/dtos"
	"github.com/google/uuid"
)

// ISessionService 会话服务接口
type ISessionService interface {
	common.BaseService
	CreateSession() (string, error)
	ValidateSession(token string) (*dtos.AdminSession, error)
	DeleteSession(token string) error
}

type sessionService struct {
	Config   common.BaseConfigProvider `inject:""`
	CacheMgr cachemgr.CacheManager     `inject:""`
	timeout  time.Duration
}

// NewSessionService 创建会话服务
func NewSessionService() ISessionService {
	return &sessionService{}
}

func (s *sessionService) ServiceName() string {
	return "SessionService"
}

func (s *sessionService) OnStart() error {
	timeout, err := config.Get[int](s.Config, "app.admin.session_timeout")
	if err != nil {
		return fmt.Errorf("failed to get session_timeout from config: %w", err)
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
		return "", fmt.Errorf("failed to store session: %w", err)
	}

	return token, nil
}

func (s *sessionService) ValidateSession(token string) (*dtos.AdminSession, error) {
	ctx := context.Background()
	sessionKey := fmt.Sprintf("session:%s", token)

	var session dtos.AdminSession
	if err := s.CacheMgr.Get(ctx, sessionKey, &session); err != nil {
		return nil, errors.New("session not found")
	}

	if time.Now().After(session.ExpiresAt) {
		s.DeleteSession(token)
		return nil, errors.New("session expired")
	}

	return &session, nil
}

func (s *sessionService) DeleteSession(token string) error {
	ctx := context.Background()
	sessionKey := fmt.Sprintf("session:%s", token)
	return s.CacheMgr.Delete(ctx, sessionKey)
}

var _ ISessionService = (*sessionService)(nil)
