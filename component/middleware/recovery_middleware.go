package middleware

import (
	"github.com/lite-lake/litecore-go/server/builtin/manager/loggermgr"
	"runtime/debug"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/lite-lake/litecore-go/common"
)

// RecoveryMiddleware panic 恢复中间件
type RecoveryMiddleware struct {
	order     int
	LoggerMgr loggermgr.ILoggerManager `inject:""`
}

// NewRecoveryMiddleware 创建 panic 恢复中间件
func NewRecoveryMiddleware() common.IBaseMiddleware {
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
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				stack := debug.Stack()

				requestID := c.GetHeader("X-Request-ID")
				if requestID == "" {
					requestID = c.GetString("request_id")
				}

				clientIP := c.ClientIP()
				method := c.Request.Method
				path := c.Request.URL.Path
				userAgent := c.Request.UserAgent()
				query := c.Request.URL.RawQuery

				m.LoggerMgr.Ins().Error(
					"PANIC recovered",
					"panic", err,
					"method", method,
					"path", path,
					"query", query,
					"ip", clientIP,
					"userAgent", userAgent,
					"requestID", requestID,
					"timestamp", time.Now().Format(time.RFC3339Nano),
					"stack", string(stack),
				)

				c.JSON(common.HTTPStatusInternalServerError, gin.H{
					"error": "内部服务器错误",
					"code":  "INTERNAL_SERVER_ERROR",
				})
				c.Abort()
			}
		}()
		c.Next()
	}
}

// OnStart 服务器启动时触发
func (m *RecoveryMiddleware) OnStart() error {
	return nil
}

// OnStop 服务器停止时触发
func (m *RecoveryMiddleware) OnStop() error {
	return nil
}
