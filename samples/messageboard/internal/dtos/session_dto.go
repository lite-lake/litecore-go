// Package dtos 定义数据传输对象
package dtos

import "time"

// AdminSession 管理员会话信息
// 存储在缓存中，不持久化到数据库
type AdminSession struct {
	Token     string    `json:"token"`
	ExpiresAt time.Time `json:"expires_at"`
}
