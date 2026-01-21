// Package repositories 定义数据访问层
package repositories

import (
	"github.com/lite-lake/litecore-go/common"
	"github.com/lite-lake/litecore-go/samples/messageboard/internal/entities"
	"github.com/lite-lake/litecore-go/samples/messageboard/internal/infras/managers"
)

// IMessageRepository 留言仓储接口
type IMessageRepository interface {
	common.IBaseRepository
	Create(message *entities.Message) error
	GetByID(id uint) (*entities.Message, error)
	GetApprovedMessages() ([]*entities.Message, error)
	GetAllMessages() ([]*entities.Message, error)
	UpdateStatus(id uint, status string) error
	Delete(id uint) error
	CountByStatus(status string) (int64, error)
}

type messageRepository struct {
	Config  common.IBaseConfigProvider `inject:""`
	Manager managers.IDatabaseManager  `inject:""`
}

// NewMessageRepository 创建留言仓储
func NewMessageRepository() IMessageRepository {
	return &messageRepository{}
}

func (r *messageRepository) RepositoryName() string {
	return "MessageRepository"
}

func (r *messageRepository) OnStart() error {
	return r.Manager.AutoMigrate(&entities.Message{})
}

func (r *messageRepository) OnStop() error {
	return nil
}

func (r *messageRepository) Create(message *entities.Message) error {
	db := r.Manager.DB()
	return db.Create(message).Error
}

func (r *messageRepository) GetByID(id uint) (*entities.Message, error) {
	db := r.Manager.DB()
	var message entities.Message
	err := db.First(&message, id).Error
	if err != nil {
		return nil, err
	}
	return &message, nil
}

func (r *messageRepository) GetApprovedMessages() ([]*entities.Message, error) {
	db := r.Manager.DB()
	var messages []*entities.Message
	err := db.Where("status = ?", "approved").
		Order("created_at DESC").
		Find(&messages).Error
	return messages, err
}

func (r *messageRepository) GetAllMessages() ([]*entities.Message, error) {
	db := r.Manager.DB()
	var messages []*entities.Message
	err := db.Order("created_at DESC").Find(&messages).Error
	return messages, err
}

func (r *messageRepository) UpdateStatus(id uint, status string) error {
	db := r.Manager.DB()
	return db.Model(&entities.Message{}).
		Where("id = ?", id).
		Update("status", status).Error
}

func (r *messageRepository) Delete(id uint) error {
	db := r.Manager.DB()
	return db.Delete(&entities.Message{}, id).Error
}

func (r *messageRepository) CountByStatus(status string) (int64, error) {
	db := r.Manager.DB()
	var count int64
	err := db.Model(&entities.Message{}).
		Where("status = ?", status).
		Count(&count).Error
	return count, err
}

var _ IMessageRepository = (*messageRepository)(nil)
