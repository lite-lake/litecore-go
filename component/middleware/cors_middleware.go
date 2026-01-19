package middleware

import (
	"github.com/gin-gonic/gin"

	"com.litelake.litecore/common"
)

// CorsMiddleware CORS 跨域中间件
type CorsMiddleware struct {
	order int
}

// NewCorsMiddleware 创建 CORS 中间件
func NewCorsMiddleware() common.BaseMiddleware {
	return &CorsMiddleware{order: 30}
}

// MiddlewareName 返回中间件名称
func (m *CorsMiddleware) MiddlewareName() string {
	return "CorsMiddleware"
}

// Order 返回执行顺序
func (m *CorsMiddleware) Order() int {
	return m.order
}

// Wrapper 返回 Gin 中间件函数
func (m *CorsMiddleware) Wrapper() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE, PATCH")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}

// OnStart 服务器启动时触发
func (m *CorsMiddleware) OnStart() error {
	return nil
}

// OnStop 服务器停止时触发
func (m *CorsMiddleware) OnStop() error {
	return nil
}
