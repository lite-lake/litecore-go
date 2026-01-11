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
	"github.com/google/uuid"
)

// AdminSession 管理员会话信息
// 存储在缓存中，不持久化到数据库
type AdminSession struct {
	Token     string    `json:"token"`
	ExpiresAt time.Time `json:"expires_at"`
}

// SessionService 会话管理服务
type SessionService struct {
	Config     common.BaseConfigProvider `inject:""`
	CacheMgr   cachemgr.CacheManager     `inject:""`
	timeout    time.Duration
}

// NewSessionService 创建会话服务实例
func NewSessionService() *SessionService {
	return &SessionService{}
}

// ServiceName 实现 BaseService 接口
func (s *SessionService) ServiceName() string {
	return "SessionService"
}

// OnStart 实现 BaseService 接口
func (s *SessionService) OnStart() error {
	// 从配置读取会话超时时间
	timeout, err := config.Get[int](s.Config, "app.admin.session_timeout")
	if err != nil {
		return fmt.Errorf("failed to get session_timeout from config: %w", err)
	}
	s.timeout = time.Duration(timeout) * time.Second
	return nil
}

// OnStop 实现 BaseService 接口
func (s *SessionService) OnStop() error {
	return nil
}

// CreateSession 创建会话并返回令牌
func (s *SessionService) CreateSession() (string, error) {
	token := uuid.New().String()
	session := &AdminSession{
		Token:     token,
		ExpiresAt: time.Now().Add(s.timeout),
	}

	// 使用缓存管理器存储会话
	ctx := context.Background()
	sessionKey := fmt.Sprintf("session:%s", token)
	if err := s.CacheMgr.Set(ctx, sessionKey, session, s.timeout); err != nil {
		return "", fmt.Errorf("failed to store session: %w", err)
	}

	return token, nil
}

// ValidateSession 验证会话
func (s *SessionService) ValidateSession(token string) (*AdminSession, error) {
	ctx := context.Background()
	sessionKey := fmt.Sprintf("session:%s", token)

	var session AdminSession
	if err := s.CacheMgr.Get(ctx, sessionKey, &session); err != nil {
		return nil, errors.New("session not found")
	}

	// 检查是否过期
	if time.Now().After(session.ExpiresAt) {
		s.DeleteSession(token)
		return nil, errors.New("session expired")
	}

	return &session, nil
}

// DeleteSession 删除会话
func (s *SessionService) DeleteSession(token string) error {
	ctx := context.Background()
	sessionKey := fmt.Sprintf("session:%s", token)
	return s.CacheMgr.Delete(ctx, sessionKey)
}

var _ common.BaseService = (*SessionService)(nil)
