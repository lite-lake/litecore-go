// Package entities 定义留言板应用的数据实体
package entities

import (
	"fmt"
	"time"

	"com.litelake.litecore/common"
)

// Message 留言实体
type Message struct {
	ID        uint      `gorm:"primarykey" json:"id"`
	Nickname  string    `gorm:"type:varchar(20);not null" json:"nickname"`
	Content   string    `gorm:"type:varchar(500);not null" json:"content"`
	Status    string    `gorm:"type:varchar(20);default:'pending'" json:"status"` // pending, approved, rejected
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// EntityName 实现 BaseEntity 接口
func (m *Message) EntityName() string {
	return "Message"
}

// TableName 指定表名
func (Message) TableName() string {
	return "messages"
}

// GetId 实现 BaseEntity 接口
func (m *Message) GetId() string {
	return fmt.Sprintf("%d", m.ID)
}

// IsApproved 检查留言是否已审核通过
func (m *Message) IsApproved() bool {
	return m.Status == "approved"
}

// IsPending 检查留言是否待审核
func (m *Message) IsPending() bool {
	return m.Status == "pending"
}

// IsRejected 检查留言是否已拒绝
func (m *Message) IsRejected() bool {
	return m.Status == "rejected"
}

var _ common.BaseEntity = (*Message)(nil)
