// Package services 定义业务逻辑层接口
package services

import (
	"com.litelake.litecore/common"
	"com.litelake.litecore/samples/messageboard/internal/entities"
)

// IMessageService 留言服务接口
type IMessageService interface {
	common.BaseService
	CreateMessage(nickname, content string) (*entities.Message, error)
	GetApprovedMessages() ([]*entities.Message, error)
	GetAllMessages() ([]*entities.Message, error)
	UpdateMessageStatus(id uint, status string) error
	DeleteMessage(id uint) error
	GetStatistics() (map[string]int64, error)
}

// IAuthService 认证服务接口
type IAuthService interface {
	common.BaseService
	VerifyPassword(password string) bool
	Login(password string) (string, error)
	Logout(token string) error
	ValidateToken(token string) (*AdminSession, error)
}

// ISessionService 会话服务接口
type ISessionService interface {
	common.BaseService
	CreateSession() (string, error)
	ValidateSession(token string) (*AdminSession, error)
	DeleteSession(token string) error
}
