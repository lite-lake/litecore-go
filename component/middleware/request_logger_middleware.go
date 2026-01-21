package middleware

import (
	"bytes"
	"fmt"
	"io"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/lite-lake/litecore-go/common"
	"github.com/lite-lake/litecore-go/component/manager/loggermgr"
)

// RequestLoggerMiddleware 请求日志中间件
type RequestLoggerMiddleware struct {
	order     int
	loggerMgr loggermgr.ILoggerManager `inject:""`
	logger    loggermgr.ILogger
}

// NewRequestLoggerMiddleware 创建请求日志中间件
func NewRequestLoggerMiddleware() common.IBaseMiddleware {
	return &RequestLoggerMiddleware{order: 20}
}

// initLogger 初始化日志器（依赖注入后调用）
func (m *RequestLoggerMiddleware) initLogger() {
	if m.loggerMgr != nil {
		m.logger = m.loggerMgr.Logger("request_logger_middleware")
	}
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
		m.initLogger()
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
		requestID := m.getRequestID(c)

		if len(c.Errors) > 0 {
			if m.logger != nil {
				for _, e := range c.Errors {
					m.logger.Error("请求处理错误",
						"request_id", requestID,
						"method", method,
						"path", path,
						"client_ip", clientIP,
						"status", status,
						"latency", latency,
						"error", e.Error(),
						"stack", e.Type,
					)
				}
			} else {
				for _, e := range c.Errors {
					println("[ERROR]", e.Error())
				}
			}
		} else {
			if m.logger != nil {
				m.logger.Info("请求处理完成",
					"request_id", requestID,
					"method", method,
					"path", path,
					"client_ip", clientIP,
					"status", status,
					"latency", latency,
				)
			} else {
				println("[INFO]", clientIP, method, path, status, latency)
			}
		}
	}
}

// getRequestID 获取请求ID
func (m *RequestLoggerMiddleware) getRequestID(c *gin.Context) string {
	requestID := c.GetHeader("X-Request-Id")
	if requestID == "" {
		requestID = c.GetHeader("X-Request-ID")
	}
	if requestID == "" {
		requestID = fmt.Sprintf("%d", time.Now().UnixNano())
	}
	c.Set("request_id", requestID)
	return requestID
}

// OnStart 服务器启动时触发
func (m *RequestLoggerMiddleware) OnStart() error {
	return nil
}

// OnStop 服务器停止时触发
func (m *RequestLoggerMiddleware) OnStop() error {
	return nil
}
