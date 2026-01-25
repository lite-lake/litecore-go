// Package dtos 定义留言板应用的数据传输对象
package dtos

import "time"

// CreateMessageRequest 创建留言请求
type CreateMessageRequest struct {
	Nickname string `json:"nickname" binding:"required,min=2,max=20"` // 昵称，2-20 字符
	Content  string `json:"content" binding:"required,min=5,max=500"` // 留言内容，5-500 字符
}

// UpdateStatusRequest 更新留言状态请求
type UpdateStatusRequest struct {
	Status string `form:"status" binding:"required,oneof=pending approved rejected"` // 目标状态：pending、approved、rejected
}

// LoginRequest 管理员登录请求
type LoginRequest struct {
	Password string `json:"password" binding:"required"` // 管理员密码
}

// MessageResponse 留言响应
type MessageResponse struct {
	ID        string    `json:"id"`               // 留言 ID
	Nickname  string    `json:"nickname"`         // 昵称
	Content   string    `json:"content"`          // 留言内容
	Status    string    `json:"status,omitempty"` // 状态（管理端返回，用户端可选）
	CreatedAt time.Time `json:"created_at"`       // 创建时间
}

// LoginResponse 登录响应
type LoginResponse struct {
	Token string `json:"token"` // 认证令牌
}

// ToMessageResponse 将留言实体转换为响应 DTO
func ToMessageResponse(id, nickname, content, status string, createdAt time.Time) MessageResponse {
	return MessageResponse{
		ID:        id,
		Nickname:  nickname,
		Content:   content,
		Status:    status,
		CreatedAt: createdAt,
	}
}
