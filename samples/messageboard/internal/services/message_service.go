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

// MessageService 留言业务服务
type MessageService struct {
	Config     common.BaseConfigProvider  `inject:""`
	Repository *repositories.MessageRepository `inject:""`
}

// NewMessageService 创建留言服务实例
func NewMessageService() *MessageService {
	return &MessageService{}
}

// ServiceName 实现 BaseService 接口
func (s *MessageService) ServiceName() string {
	return "MessageService"
}

// OnStart 实现 BaseService 接口
func (s *MessageService) OnStart() error {
	return nil
}

// OnStop 实现 BaseService 接口
func (s *MessageService) OnStop() error {
	return nil
}

// CreateMessage 创建留言
func (s *MessageService) CreateMessage(nickname, content string) (*entities.Message, error) {
	// 验证输入
	if len(nickname) < 2 || len(nickname) > 20 {
		return nil, errors.New("昵称长度必须在 2-20 个字符之间")
	}
	if len(content) < 5 || len(content) > 500 {
		return nil, errors.New("留言内容长度必须在 5-500 个字符之间")
	}

	// 创建留言实体
	message := &entities.Message{
		Nickname:  nickname,
		Content:   content,
		Status:    "pending", // 默认待审核
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// 保存到数据库
	if err := s.Repository.Create(message); err != nil {
		return nil, fmt.Errorf("failed to create message: %w", err)
	}

	return message, nil
}

// GetApprovedMessages 获取已审核通过的留言列表
func (s *MessageService) GetApprovedMessages() ([]*entities.Message, error) {
	messages, err := s.Repository.GetApprovedMessages()
	if err != nil {
		return nil, fmt.Errorf("failed to get approved messages: %w", err)
	}
	return messages, nil
}

// GetAllMessages 获取所有留言（管理端）
func (s *MessageService) GetAllMessages() ([]*entities.Message, error) {
	messages, err := s.Repository.GetAllMessages()
	if err != nil {
		return nil, fmt.Errorf("failed to get all messages: %w", err)
	}
	return messages, nil
}

// UpdateMessageStatus 更新留言状态
func (s *MessageService) UpdateMessageStatus(id uint, status string) error {
	// 验证状态值
	if status != "pending" && status != "approved" && status != "rejected" {
		return errors.New("invalid status value")
	}

	// 检查留言是否存在
	message, err := s.Repository.GetByID(id)
	if err != nil {
		return fmt.Errorf("message not found: %w", err)
	}
	if message == nil {
		return errors.New("message not found")
	}

	// 更新状态
	if err := s.Repository.UpdateStatus(id, status); err != nil {
		return fmt.Errorf("failed to update message status: %w", err)
	}

	return nil
}

// DeleteMessage 删除留言
func (s *MessageService) DeleteMessage(id uint) error {
	// 检查留言是否存在
	message, err := s.Repository.GetByID(id)
	if err != nil {
		return fmt.Errorf("message not found: %w", err)
	}
	if message == nil {
		return errors.New("message not found")
	}

	// 删除留言
	if err := s.Repository.Delete(id); err != nil {
		return fmt.Errorf("failed to delete message: %w", err)
	}

	return nil
}

// GetStatistics 获取留言统计信息
func (s *MessageService) GetStatistics() (map[string]int64, error) {
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

var _ common.BaseService = (*MessageService)(nil)
