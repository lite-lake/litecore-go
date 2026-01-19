package middleware

import (
	"github.com/gin-gonic/gin"

	"com.litelake.litecore/common"
)

// TelemetryMiddleware 遥测中间件
type TelemetryMiddleware struct {
	order   int
	manager common.IBaseManager
}

// NewTelemetryMiddleware 创建遥测中间件
func NewTelemetryMiddleware(manager common.IBaseManager) common.IBaseMiddleware {
	return &TelemetryMiddleware{order: 50, manager: manager}
}

// MiddlewareName 返回中间件名称
func (m *TelemetryMiddleware) MiddlewareName() string {
	return "TelemetryMiddleware"
}

// Order 返回执行顺序
func (m *TelemetryMiddleware) Order() int {
	return m.order
}

// Wrapper 返回 Gin 中间件函数
func (m *TelemetryMiddleware) Wrapper() gin.HandlerFunc {
	if otelMiddleware, ok := m.manager.(interface {
		GinMiddleware() gin.HandlerFunc
	}); ok {
		return otelMiddleware.GinMiddleware()
	}
	return func(c *gin.Context) {
		c.Next()
	}
}

// OnStart 服务器启动时触发
func (m *TelemetryMiddleware) OnStart() error {
	return nil
}

// OnStop 服务器停止时触发
func (m *TelemetryMiddleware) OnStop() error {
	return nil
}
