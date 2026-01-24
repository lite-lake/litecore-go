// Package repositories 定义数据访问层
package repositories

import (
	"github.com/lite-lake/litecore-go/common"
	"github.com/lite-lake/litecore-go/manager/configmgr"
	"github.com/lite-lake/litecore-go/manager/databasemgr"
	"github.com/lite-lake/litecore-go/samples/messageboard/internal/entities"
)

// IMessageRepository 留言仓储接口
type IMessageRepository interface {
	common.IBaseRepository
	Create(message *entities.Message) error            // 创建留言
	GetByID(id uint) (*entities.Message, error)        // 根据 ID 获取留言
	GetApprovedMessages() ([]*entities.Message, error) // 获取已审核通过的留言列表
	GetAllMessages() ([]*entities.Message, error)      // 获取所有留言列表
	UpdateStatus(id uint, status string) error         // 更新留言状态
	Delete(id uint) error                              // 删除留言
	CountByStatus(status string) (int64, error)        // 根据状态统计留言数量
}

type messageRepositoryImpl struct {
	Config  configmgr.IConfigManager     `inject:""` // 配置管理器
	Manager databasemgr.IDatabaseManager `inject:""` // 数据库管理器
}

// NewMessageRepository 创建留言仓储实例
func NewMessageRepository() IMessageRepository {
	return &messageRepositoryImpl{}
}

// RepositoryName 返回仓储名称
func (r *messageRepositoryImpl) RepositoryName() string {
	return "MessageRepository"
}

// OnStart 启动时自动迁移数据库表结构
func (r *messageRepositoryImpl) OnStart() error {
	return r.Manager.AutoMigrate(&entities.Message{})
}

// OnStop 停止时清理资源
func (r *messageRepositoryImpl) OnStop() error {
	return nil
}

// Create 在数据库中创建新留言
func (r *messageRepositoryImpl) Create(message *entities.Message) error {
	db := r.Manager.DB()
	return db.Create(message).Error
}

// GetByID 根据 ID 获取留言记录
func (r *messageRepositoryImpl) GetByID(id uint) (*entities.Message, error) {
	db := r.Manager.DB()
	var message entities.Message
	err := db.First(&message, id).Error
	if err != nil {
		return nil, err
	}
	return &message, nil
}

// GetApprovedMessages 获取已审核通过的留言列表，按创建时间倒序排列
func (r *messageRepositoryImpl) GetApprovedMessages() ([]*entities.Message, error) {
	db := r.Manager.DB()
	var messages []*entities.Message
	err := db.Where("status = ?", "approved").
		Order("created_at DESC").
		Find(&messages).Error
	return messages, err
}

// GetAllMessages 获取所有留言列表，按创建时间倒序排列
func (r *messageRepositoryImpl) GetAllMessages() ([]*entities.Message, error) {
	db := r.Manager.DB()
	var messages []*entities.Message
	err := db.Order("created_at DESC").Find(&messages).Error
	return messages, err
}

// UpdateStatus 更新指定留言的状态
func (r *messageRepositoryImpl) UpdateStatus(id uint, status string) error {
	db := r.Manager.DB()
	return db.Model(&entities.Message{}).
		Where("id = ?", id).
		Update("status", status).Error
}

// Delete 删除指定的留言记录
func (r *messageRepositoryImpl) Delete(id uint) error {
	db := r.Manager.DB()
	return db.Delete(&entities.Message{}, id).Error
}

// CountByStatus 统计指定状态的留言数量
func (r *messageRepositoryImpl) CountByStatus(status string) (int64, error) {
	db := r.Manager.DB()
	var count int64
	err := db.Model(&entities.Message{}).
		Where("status = ?", status).
		Count(&count).Error
	return count, err
}

var _ IMessageRepository = (*messageRepositoryImpl)(nil)
