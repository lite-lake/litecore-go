// Package dtos 定义留言板应用的数据传输对象
package dtos

import "time"

// CreateMessageRequest 创建留言请求
type CreateMessageRequest struct {
	Nickname string `json:"nickname" binding:"required,min=2,max=20"`
	Content  string `json:"content" binding:"required,min=5,max=500"`
}

// UpdateStatusRequest 更新留言状态请求
type UpdateStatusRequest struct {
	Status string `form:"status" binding:"required,oneof=pending approved rejected"`
}

// LoginRequest 管理员登录请求
type LoginRequest struct {
	Password string `json:"password" binding:"required"`
}

// MessageResponse 留言响应
type MessageResponse struct {
	ID        uint      `json:"id"`
	Nickname  string    `json:"nickname"`
	Content   string    `json:"content"`
	Status    string    `json:"status,omitempty"` // 管理端返回，用户端可选
	CreatedAt time.Time `json:"created_at"`
}

// LoginResponse 登录响应
type LoginResponse struct {
	Token string `json:"token"`
}

// ToMessageResponse 将 Message 实体转换为 MessageResponse
func ToMessageResponse(id uint, nickname, content, status string, createdAt time.Time) MessageResponse {
	return MessageResponse{
		ID:        id,
		Nickname:  nickname,
		Content:   content,
		Status:    status,
		CreatedAt: createdAt,
	}
}
