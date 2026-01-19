package middleware

import (
	"github.com/gin-gonic/gin"

	"com.litelake.litecore/common"
)

// RecoveryMiddleware panic 恢复中间件
type RecoveryMiddleware struct {
	order int
}

// NewRecoveryMiddleware 创建 panic 恢复中间件
func NewRecoveryMiddleware() common.BaseMiddleware {
	return &RecoveryMiddleware{order: 10}
}

// MiddlewareName 返回中间件名称
func (m *RecoveryMiddleware) MiddlewareName() string {
	return "RecoveryMiddleware"
}

// Order 返回执行顺序
func (m *RecoveryMiddleware) Order() int {
	return m.order
}

// Wrapper 返回 Gin 中间件函数
func (m *RecoveryMiddleware) Wrapper() gin.HandlerFunc {
	return gin.Recovery()
}

// OnStart 服务器启动时触发
func (m *RecoveryMiddleware) OnStart() error {
	return nil
}

// OnStop 服务器停止时触发
func (m *RecoveryMiddleware) OnStop() error {
	return nil
}
