// Package services 定义业务逻辑层
package services

import (
	"errors"
	"fmt"

	"github.com/lite-lake/litecore-go/common"
	"github.com/lite-lake/litecore-go/manager/configmgr"
	"github.com/lite-lake/litecore-go/manager/loggermgr"
	"github.com/lite-lake/litecore-go/samples/messageboard/internal/entities"
	"github.com/lite-lake/litecore-go/samples/messageboard/internal/repositories"
)

// IMessageService 留言服务接口
type IMessageService interface {
	common.IBaseService
	CreateMessage(nickname, content string) (*entities.Message, error) // 创建留言
	GetApprovedMessages() ([]*entities.Message, error)                 // 获取已审核留言列表
	GetAllMessages() ([]*entities.Message, error)                      // 获取所有留言列表
	UpdateMessageStatus(id string, status string) error                // 更新留言状态
	DeleteMessage(id string) error                                     // 删除留言
	GetStatistics() (map[string]int64, error)                          // 获取留言统计信息
}

type messageServiceImpl struct {
	Config     configmgr.IConfigManager        `inject:""` // 配置管理器
	Repository repositories.IMessageRepository `inject:""` // 留言仓储
	LoggerMgr  loggermgr.ILoggerManager        `inject:""` // 日志管理器
}

// NewMessageService 创建留言服务实例
func NewMessageService() IMessageService {
	return &messageServiceImpl{}
}

// ServiceName 返回服务名称
func (s *messageServiceImpl) ServiceName() string {
	return "MessageService"
}

// OnStart 启动时初始化
func (s *messageServiceImpl) OnStart() error {
	return nil
}

// OnStop 停止时清理
func (s *messageServiceImpl) OnStop() error {
	return nil
}

// CreateMessage 创建新留言
// 验证昵称和内容长度，初始状态为 pending
func (s *messageServiceImpl) CreateMessage(nickname, content string) (*entities.Message, error) {
	if len(nickname) < 2 || len(nickname) > 20 {
		s.LoggerMgr.Ins().Warn("Failed to create message: invalid nickname length", "nickname_length", len(nickname))
		return nil, errors.New("昵称长度必须在 2-20 个字符之间")
	}
	if len(content) < 5 || len(content) > 500 {
		s.LoggerMgr.Ins().Warn("Failed to create message: invalid content length", "content_length", len(content))

		return nil, errors.New("留言内容长度必须在 5-500 个字符之间")
	}

	message := &entities.Message{
		Nickname: nickname,
		Content:  content,
		Status:   "pending",
	}

	if err := s.Repository.Create(message); err != nil {
		s.LoggerMgr.Ins().Error("Failed to create message", "nickname", nickname, "error", err)

		return nil, fmt.Errorf("failed to create message: %w", err)
	}

	s.LoggerMgr.Ins().Info("Message created successfully", "id", message.ID, "nickname", message.Nickname, "status", message.Status)

	return message, nil
}

// GetApprovedMessages 获取已审核通过的留言列表
func (s *messageServiceImpl) GetApprovedMessages() ([]*entities.Message, error) {

	s.LoggerMgr.Ins().Debug("Getting approved messages")

	messages, err := s.Repository.GetApprovedMessages()
	if err != nil {
		s.LoggerMgr.Ins().Error("Failed to get approved messages", "error", err)
		return nil, fmt.Errorf("failed to get approved messages: %w", err)
	}

	s.LoggerMgr.Ins().Debug("Approved messages retrieved successfully", "count", len(messages))

	return messages, nil
}

// GetAllMessages 获取所有留言列表（管理员专用）
func (s *messageServiceImpl) GetAllMessages() ([]*entities.Message, error) {
	s.LoggerMgr.Ins().Debug("Getting all messages")

	messages, err := s.Repository.GetAllMessages()
	if err != nil {
		s.LoggerMgr.Ins().Error("Failed to get all messages", "error", err)
		return nil, fmt.Errorf("failed to get all messages: %w", err)
	}

	s.LoggerMgr.Ins().Debug("All messages retrieved successfully", "count", len(messages))

	return messages, nil
}

// UpdateMessageStatus 更新留言状态（管理员专用）
// 状态值必须是 pending、approved 或 rejected
func (s *messageServiceImpl) UpdateMessageStatus(id string, status string) error {
	if status != "pending" && status != "approved" && status != "rejected" {
		s.LoggerMgr.Ins().Warn("Failed to update message status: invalid status value", "id", id, "status", status)
		return errors.New("invalid status value")
	}

	s.LoggerMgr.Ins().Debug("Preparing to update message status", "id", id, "status", status)

	message, err := s.Repository.GetByID(id)
	if err != nil {
		s.LoggerMgr.Ins().Error("Failed to update message status: message not found", "id", id, "error", err)
		return fmt.Errorf("message not found: %w", err)
	}
	if message == nil {
		s.LoggerMgr.Ins().Warn("Failed to update message status: message not found", "id", id)
		return errors.New("message not found")
	}

	if err := s.Repository.UpdateStatus(id, status); err != nil {
		s.LoggerMgr.Ins().Error("Failed to update message status", "id", id, "status", status, "error", err)
		return fmt.Errorf("failed to update message status: %w", err)
	}

	s.LoggerMgr.Ins().Info("Message status updated successfully", "id", id, "old_status", message.Status, "new_status", status)

	return nil
}

// DeleteMessage 删除留言（管理员专用）
func (s *messageServiceImpl) DeleteMessage(id string) error {
	s.LoggerMgr.Ins().Debug("Preparing to delete message", "id", id)

	message, err := s.Repository.GetByID(id)
	if err != nil {
		s.LoggerMgr.Ins().Error("Failed to delete message: message not found", "id", id, "error", err)
		return fmt.Errorf("message not found: %w", err)
	}
	if message == nil {
		s.LoggerMgr.Ins().Warn("Failed to delete message: message not found", "id", id)
		return errors.New("message not found")
	}

	if err := s.Repository.Delete(id); err != nil {
		s.LoggerMgr.Ins().Error("Failed to delete message", "id", id, "nickname", message.Nickname, "error", err)
		return fmt.Errorf("failed to delete message: %w", err)
	}

	s.LoggerMgr.Ins().Info("Message deleted successfully", "id", id, "nickname", message.Nickname, "status", message.Status)

	return nil
}

// GetStatistics 获取留言统计信息
// 返回各状态留言数量及总数
func (s *messageServiceImpl) GetStatistics() (map[string]int64, error) {
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

var _ IMessageService = (*messageServiceImpl)(nil)
