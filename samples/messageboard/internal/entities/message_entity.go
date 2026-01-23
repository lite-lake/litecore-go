// Package entities 定义留言板应用的数据实体
package entities

import (
	"fmt"
	"time"

	"github.com/lite-lake/litecore-go/common"
)

// Message 留言实体
type Message struct {
	ID        uint      `gorm:"primarykey" json:"id"`                             // 留言 ID
	Nickname  string    `gorm:"type:varchar(20);not null" json:"nickname"`        // 昵称
	Content   string    `gorm:"type:varchar(500);not null" json:"content"`        // 留言内容
	Status    string    `gorm:"type:varchar(20);default:'pending'" json:"status"` // 状态：pending 待审核，approved 已通过，rejected 已拒绝
	CreatedAt time.Time `json:"created_at"`                                       // 创建时间
	UpdatedAt time.Time `json:"updated_at"`                                       // 更新时间
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

var _ common.IBaseEntity = (*Message)(nil)
