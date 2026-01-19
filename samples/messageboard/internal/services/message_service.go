// Package services 定义业务逻辑层
package services

import (
	"errors"
	"fmt"
	"time"

	"com.litelake.litecore/common"
	"com.litelake.litecore/samples/messageboard/internal/entities"
	"com.litelake.litecore/samples/messageboard/internal/repositories"
)

// IMessageService 留言服务接口
type IMessageService interface {
	common.IBaseService
	CreateMessage(nickname, content string) (*entities.Message, error)
	GetApprovedMessages() ([]*entities.Message, error)
	GetAllMessages() ([]*entities.Message, error)
	UpdateMessageStatus(id uint, status string) error
	DeleteMessage(id uint) error
	GetStatistics() (map[string]int64, error)
}

type messageService struct {
	Config     common.IBaseConfigProvider       `inject:""`
	Repository repositories.IMessageRepository `inject:""`
}

// NewMessageService 创建留言服务
func NewMessageService() IMessageService {
	return &messageService{}
}

func (s *messageService) ServiceName() string {
	return "MessageService"
}

func (s *messageService) OnStart() error {
	return nil
}

func (s *messageService) OnStop() error {
	return nil
}

func (s *messageService) CreateMessage(nickname, content string) (*entities.Message, error) {
	if len(nickname) < 2 || len(nickname) > 20 {
		return nil, errors.New("昵称长度必须在 2-20 个字符之间")
	}
	if len(content) < 5 || len(content) > 500 {
		return nil, errors.New("留言内容长度必须在 5-500 个字符之间")
	}

	message := &entities.Message{
		Nickname:  nickname,
		Content:   content,
		Status:    "pending",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := s.Repository.Create(message); err != nil {
		return nil, fmt.Errorf("failed to create message: %w", err)
	}

	return message, nil
}

func (s *messageService) GetApprovedMessages() ([]*entities.Message, error) {
	messages, err := s.Repository.GetApprovedMessages()
	if err != nil {
		return nil, fmt.Errorf("failed to get approved messages: %w", err)
	}
	return messages, nil
}

func (s *messageService) GetAllMessages() ([]*entities.Message, error) {
	messages, err := s.Repository.GetAllMessages()
	if err != nil {
		return nil, fmt.Errorf("failed to get all messages: %w", err)
	}
	return messages, nil
}

func (s *messageService) UpdateMessageStatus(id uint, status string) error {
	if status != "pending" && status != "approved" && status != "rejected" {
		return errors.New("invalid status value")
	}

	message, err := s.Repository.GetByID(id)
	if err != nil {
		return fmt.Errorf("message not found: %w", err)
	}
	if message == nil {
		return errors.New("message not found")
	}

	if err := s.Repository.UpdateStatus(id, status); err != nil {
		return fmt.Errorf("failed to update message status: %w", err)
	}

	return nil
}

func (s *messageService) DeleteMessage(id uint) error {
	message, err := s.Repository.GetByID(id)
	if err != nil {
		return fmt.Errorf("message not found: %w", err)
	}
	if message == nil {
		return errors.New("message not found")
	}

	if err := s.Repository.Delete(id); err != nil {
		return fmt.Errorf("failed to delete message: %w", err)
	}

	return nil
}

func (s *messageService) GetStatistics() (map[string]int64, error) {
	pendingCount, err := s.Repository.CountByStatus("pending")
	if err != nil {
		return nil, err
	}

	approvedCount, err := s.Repository.CountByStatus("approved")
	if err != nil {
		return nil, err
	}

	rejectedCount, err := s.Repository.CountByStatus("rejected")
	if err != nil {
		return nil, err
	}

	return map[string]int64{
		"pending":  pendingCount,
		"approved": approvedCount,
		"rejected": rejectedCount,
		"total":    pendingCount + approvedCount + rejectedCount,
	}, nil
}

var _ IMessageService = (*messageService)(nil)
