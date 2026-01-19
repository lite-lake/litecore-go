package middleware

import (
	"github.com/gin-gonic/gin"

	"com.litelake.litecore/common"
)

// SecurityHeadersMiddleware 安全头中间件
type SecurityHeadersMiddleware struct {
	order int
}

// NewSecurityHeadersMiddleware 创建安全头中间件
func NewSecurityHeadersMiddleware() common.IBaseMiddleware {
	return &SecurityHeadersMiddleware{order: 40}
}

// MiddlewareName 返回中间件名称
func (m *SecurityHeadersMiddleware) MiddlewareName() string {
	return "SecurityHeadersMiddleware"
}

// Order 返回执行顺序
func (m *SecurityHeadersMiddleware) Order() int {
	return m.order
}

// Wrapper 返回 Gin 中间件函数
func (m *SecurityHeadersMiddleware) Wrapper() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("X-Frame-Options", "DENY")
		c.Writer.Header().Set("X-Content-Type-Options", "nosniff")
		c.Writer.Header().Set("X-XSS-Protection", "1; mode=block")
		c.Writer.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
		c.Next()
	}
}

// OnStart 服务器启动时触发
func (m *SecurityHeadersMiddleware) OnStart() error {
	return nil
}

// OnStop 服务器停止时触发
func (m *SecurityHeadersMiddleware) OnStop() error {
	return nil
}
