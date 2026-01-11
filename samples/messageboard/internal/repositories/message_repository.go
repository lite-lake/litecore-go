// Package repositories 定义数据访问层
package repositories

import (
	"com.litelake.litecore/common"
	"com.litelake.litecore/manager/databasemgr"
	"com.litelake.litecore/samples/messageboard/internal/entities"
)

// MessageRepository 留言仓储
type MessageRepository struct {
	Config  common.BaseConfigProvider   `inject:""`
	Manager databasemgr.DatabaseManager `inject:""`
}

// NewMessageRepository 创建留言仓储实例
func NewMessageRepository() *MessageRepository {
	return &MessageRepository{}
}

// RepositoryName 实现 BaseRepository 接口
func (r *MessageRepository) RepositoryName() string {
	return "MessageRepository"
}

// OnStart 实现 BaseRepository 接口
func (r *MessageRepository) OnStart() error {
	// 自动迁移数据库表
	return r.Manager.AutoMigrate(&entities.Message{})
}

// OnStop 实现 BaseRepository 接口
func (r *MessageRepository) OnStop() error {
	return nil
}

// Create 创建留言
func (r *MessageRepository) Create(message *entities.Message) error {
	db := r.Manager.DB()
	return db.Create(message).Error
}

// GetByID 根据 ID 获取留言
func (r *MessageRepository) GetByID(id uint) (*entities.Message, error) {
	db := r.Manager.DB()
	var message entities.Message
	err := db.First(&message, id).Error
	if err != nil {
		return nil, err
	}
	return &message, nil
}

// GetApprovedMessages 获取所有已审核通过的留言
func (r *MessageRepository) GetApprovedMessages() ([]*entities.Message, error) {
	db := r.Manager.DB()
	var messages []*entities.Message
	err := db.Where("status = ?", "approved").
		Order("created_at DESC").
		Find(&messages).Error
	return messages, err
}

// GetAllMessages 获取所有留言（包括未审核）
func (r *MessageRepository) GetAllMessages() ([]*entities.Message, error) {
	db := r.Manager.DB()
	var messages []*entities.Message
	err := db.Order("created_at DESC").Find(&messages).Error
	return messages, err
}

// UpdateStatus 更新留言状态
func (r *MessageRepository) UpdateStatus(id uint, status string) error {
	db := r.Manager.DB()
	return db.Model(&entities.Message{}).
		Where("id = ?", id).
		Update("status", status).Error
}

// Delete 删除留言
func (r *MessageRepository) Delete(id uint) error {
	db := r.Manager.DB()
	return db.Delete(&entities.Message{}, id).Error
}

// CountByStatus 根据状态统计留言数量
func (r *MessageRepository) CountByStatus(status string) (int64, error) {
	db := r.Manager.DB()
	var count int64
	err := db.Model(&entities.Message{}).
		Where("status = ?", status).
		Count(&count).Error
	return count, err
}

var _ common.BaseRepository = (*MessageRepository)(nil)
