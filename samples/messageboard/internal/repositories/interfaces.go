// Package repositories 定义数据访问层接口
package repositories

import (
	"com.litelake.litecore/common"
	"com.litelake.litecore/samples/messageboard/internal/entities"
)

// IMessageRepository 留言仓储接口
type IMessageRepository interface {
	common.BaseRepository
	Create(message *entities.Message) error
	GetByID(id uint) (*entities.Message, error)
	GetApprovedMessages() ([]*entities.Message, error)
	GetAllMessages() ([]*entities.Message, error)
	UpdateStatus(id uint, status string) error
	Delete(id uint) error
	CountByStatus(status string) (int64, error)
}
