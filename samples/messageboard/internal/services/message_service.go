// Package services 定义业务逻辑层
package services

import (
	"errors"
	"fmt"
	"github.com/lite-lake/litecore-go/server/builtin/manager/configmgr"
	"time"

	"github.com/lite-lake/litecore-go/common"
	"github.com/lite-lake/litecore-go/samples/messageboard/internal/entities"
	"github.com/lite-lake/litecore-go/samples/messageboard/internal/repositories"
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
	Config     configmgr.IConfigManager        `inject:""`
	Repository repositories.IMessageRepository `inject:""`
	Logger     common.ILogger                  `inject:""`
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
		if s.Logger != nil {
			s.Logger.Warn("创建留言失败：昵称长度不符合要求", "nickname_length", len(nickname))
		}
		return nil, errors.New("昵称长度必须在 2-20 个字符之间")
	}
	if len(content) < 5 || len(content) > 500 {
		if s.Logger != nil {
			s.Logger.Warn("创建留言失败：内容长度不符合要求", "content_length", len(content))
		}
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
		if s.Logger != nil {
			s.Logger.Error("创建留言失败", "nickname", nickname, "error", err)
		}
		return nil, fmt.Errorf("failed to create message: %w", err)
	}

	if s.Logger != nil {
		s.Logger.Info("创建留言成功", "id", message.ID, "nickname", message.Nickname, "status", message.Status)
	}

	return message, nil
}

func (s *messageService) GetApprovedMessages() ([]*entities.Message, error) {
	if s.Logger != nil {
		s.Logger.Debug("获取已审核留言列表")
	}

	messages, err := s.Repository.GetApprovedMessages()
	if err != nil {
		if s.Logger != nil {
			s.Logger.Error("获取已审核留言失败", "error", err)
		}
		return nil, fmt.Errorf("failed to get approved messages: %w", err)
	}

	if s.Logger != nil {
		s.Logger.Debug("获取已审核留言成功", "count", len(messages))
	}

	return messages, nil
}

func (s *messageService) GetAllMessages() ([]*entities.Message, error) {
	if s.Logger != nil {
		s.Logger.Debug("获取所有留言列表")
	}

	messages, err := s.Repository.GetAllMessages()
	if err != nil {
		if s.Logger != nil {
			s.Logger.Error("获取所有留言失败", "error", err)
		}
		return nil, fmt.Errorf("failed to get all messages: %w", err)
	}

	if s.Logger != nil {
		s.Logger.Debug("获取所有留言成功", "count", len(messages))
	}

	return messages, nil
}

func (s *messageService) UpdateMessageStatus(id uint, status string) error {
	if status != "pending" && status != "approved" && status != "rejected" {
		if s.Logger != nil {
			s.Logger.Warn("更新留言状态失败：无效的状态值", "id", id, "status", status)
		}
		return errors.New("invalid status value")
	}

	if s.Logger != nil {
		s.Logger.Debug("准备更新留言状态", "id", id, "status", status)
	}

	message, err := s.Repository.GetByID(id)
	if err != nil {
		if s.Logger != nil {
			s.Logger.Error("更新留言状态失败：留言不存在", "id", id, "error", err)
		}
		return fmt.Errorf("message not found: %w", err)
	}
	if message == nil {
		if s.Logger != nil {
			s.Logger.Warn("更新留言状态失败：留言不存在", "id", id)
		}
		return errors.New("message not found")
	}

	if err := s.Repository.UpdateStatus(id, status); err != nil {
		if s.Logger != nil {
			s.Logger.Error("更新留言状态失败", "id", id, "status", status, "error", err)
		}
		return fmt.Errorf("failed to update message status: %w", err)
	}

	if s.Logger != nil {
		s.Logger.Info("更新留言状态成功", "id", id, "old_status", message.Status, "new_status", status)
	}

	return nil
}

func (s *messageService) DeleteMessage(id uint) error {
	if s.Logger != nil {
		s.Logger.Debug("准备删除留言", "id", id)
	}

	message, err := s.Repository.GetByID(id)
	if err != nil {
		if s.Logger != nil {
			s.Logger.Error("删除留言失败：留言不存在", "id", id, "error", err)
		}
		return fmt.Errorf("message not found: %w", err)
	}
	if message == nil {
		if s.Logger != nil {
			s.Logger.Warn("删除留言失败：留言不存在", "id", id)
		}
		return errors.New("message not found")
	}

	if err := s.Repository.Delete(id); err != nil {
		if s.Logger != nil {
			s.Logger.Error("删除留言失败", "id", id, "nickname", message.Nickname, "error", err)
		}
		return fmt.Errorf("failed to delete message: %w", err)
	}

	if s.Logger != nil {
		s.Logger.Info("删除留言成功", "id", id, "nickname", message.Nickname, "status", message.Status)
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
