package middleware

import (
	"bytes"
	"io"
	"time"

	"github.com/gin-gonic/gin"

	"com.litelake.litecore/common"
)

// RequestLoggerMiddleware 请求日志中间件
type RequestLoggerMiddleware struct {
	order int
}

// NewRequestLoggerMiddleware 创建请求日志中间件
func NewRequestLoggerMiddleware() common.BaseMiddleware {
	return &RequestLoggerMiddleware{order: 20}
}

// MiddlewareName 返回中间件名称
func (m *RequestLoggerMiddleware) MiddlewareName() string {
	return "RequestLoggerMiddleware"
}

// Order 返回执行顺序
func (m *RequestLoggerMiddleware) Order() int {
	return m.order
}

// Wrapper 返回 Gin 中间件函数
func (m *RequestLoggerMiddleware) Wrapper() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		var bodyBytes []byte
		if c.Request.Body != nil && c.Request.Method != "GET" {
			bodyBytes, _ = io.ReadAll(c.Request.Body)
			c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
		}

		c.Next()

		latency := time.Since(start)
		status := c.Writer.Status()
		clientIP := c.ClientIP()
		method := c.Request.Method
		path := c.Request.URL.Path

		if len(c.Errors) > 0 {
			for _, e := range c.Errors {
				println("[ERROR]", e.Error())
			}
		} else {
			println("[INFO]", clientIP, method, path, status, latency)
		}
	}
}

// OnStart 服务器启动时触发
func (m *RequestLoggerMiddleware) OnStart() error {
	return nil
}

// OnStop 服务器停止时触发
func (m *RequestLoggerMiddleware) OnStop() error {
	return nil
}
