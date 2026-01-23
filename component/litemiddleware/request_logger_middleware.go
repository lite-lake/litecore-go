package litemiddleware

import (
	"bytes"
	"fmt"
	"io"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/lite-lake/litecore-go/common"
	"github.com/lite-lake/litecore-go/server/builtin/manager/loggermgr"
)

// RequestLoggerConfig 请求日志配置
type RequestLoggerConfig struct {
	Name            string   // 中间件名称
	Order           *int     // 执行顺序（指针类型用于判断是否设置）
	Enable          bool     // 是否启用请求日志
	LogBody         bool     // 是否记录请求 Body
	MaxBodySize     int      // 最大记录 Body 大小（字节），0 表示不限制
	SkipPaths       []string // 跳过日志记录的路径
	LogHeaders      []string // 需要记录的请求头
	SuccessLogLevel string   // 成功请求日志级别（debug/info）
}

// DefaultRequestLoggerConfig 默认请求日志配置
func DefaultRequestLoggerConfig() *RequestLoggerConfig {
	defaultOrder := OrderRequestLogger
	return &RequestLoggerConfig{
		Name:            "RequestLoggerMiddleware",
		Order:           &defaultOrder,
		Enable:          true,
		LogBody:         true,
		MaxBodySize:     4096,
		SkipPaths:       []string{"/health", "/metrics"},
		LogHeaders:      []string{"User-Agent", "Content-Type"},
		SuccessLogLevel: "info",
	}
}

// requestLoggerMiddleware 请求日志中间件
type requestLoggerMiddleware struct {
	LoggerMgr loggermgr.ILoggerManager `inject:""`
	cfg       *RequestLoggerConfig
}

// NewRequestLoggerMiddleware 创建请求日志中间件
func NewRequestLoggerMiddleware(config *RequestLoggerConfig) common.IBaseMiddleware {
	if config == nil {
		config = DefaultRequestLoggerConfig()
	}
	return &requestLoggerMiddleware{cfg: config}
}

// NewRequestLoggerMiddlewareWithDefaults 使用默认配置创建请求日志中间件
func NewRequestLoggerMiddlewareWithDefaults() common.IBaseMiddleware {
	return NewRequestLoggerMiddleware(nil)
}

// MiddlewareName 返回中间件名称
func (m *requestLoggerMiddleware) MiddlewareName() string {
	if m.cfg.Name != "" {
		return m.cfg.Name
	}
	return "RequestLoggerMiddleware"
}

// Order 返回执行顺序
func (m *requestLoggerMiddleware) Order() int {
	if m.cfg.Order != nil {
		return *m.cfg.Order
	}
	return OrderRequestLogger
}

// Wrapper 返回 Gin 中间件函数
func (m *requestLoggerMiddleware) Wrapper() gin.HandlerFunc {
	return func(c *gin.Context) {
		if !m.cfg.Enable {
			c.Next()
			return
		}

		for _, skipPath := range m.cfg.SkipPaths {
			if c.Request.URL.Path == skipPath {
				c.Next()
				return
			}
		}

		start := time.Now()

		var bodyBytes []byte
		if m.cfg.LogBody && c.Request.Body != nil && c.Request.Method != "GET" {
			var err error
			bodyBytes, err = io.ReadAll(c.Request.Body)
			if err == nil {
				c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
			}
		}

		c.Next()

		latency := time.Since(start)
		status := c.Writer.Status()
		clientIP := c.ClientIP()
		method := c.Request.Method
		path := c.Request.URL.Path
		requestID := m.getRequestID(c)

		logFunc := m.LoggerMgr.Ins().Info
		if m.cfg.SuccessLogLevel == "debug" {
			logFunc = m.LoggerMgr.Ins().Debug
		}

		if len(c.Errors) > 0 {
			for _, e := range c.Errors {
				fields := []interface{}{
					"request_id", requestID,
					"method", method,
					"path", path,
					"client_ip", clientIP,
					"status", status,
					"latency", latency,
					"error", e.Error(),
					"stack", e.Type,
				}

				if m.cfg.LogBody && len(bodyBytes) > 0 {
					bodyStr := string(bodyBytes)
					if m.cfg.MaxBodySize > 0 && len(bodyStr) > m.cfg.MaxBodySize {
						bodyStr = bodyStr[:m.cfg.MaxBodySize] + "...(truncated)"
					}
					fields = append(fields, "body", bodyStr)
				}

				if len(m.cfg.LogHeaders) > 0 {
					headers := make(map[string]string)
					for _, key := range m.cfg.LogHeaders {
						headers[key] = c.GetHeader(key)
					}
					fields = append(fields, "headers", headers)
				}

				m.LoggerMgr.Ins().Error("请求处理错误", fields...)
			}
		} else {
			fields := []interface{}{
				"request_id", requestID,
				"method", method,
				"path", path,
				"client_ip", clientIP,
				"status", status,
				"latency", latency,
			}

			if m.cfg.LogBody && len(bodyBytes) > 0 {
				bodyStr := string(bodyBytes)
				if m.cfg.MaxBodySize > 0 && len(bodyStr) > m.cfg.MaxBodySize {
					bodyStr = bodyStr[:m.cfg.MaxBodySize] + "...(truncated)"
				}
				fields = append(fields, "body", bodyStr)
			}

			if len(m.cfg.LogHeaders) > 0 {
				headers := make(map[string]string)
				for _, key := range m.cfg.LogHeaders {
					headers[key] = c.GetHeader(key)
				}
				fields = append(fields, "headers", headers)
			}

			logFunc("请求处理完成", fields...)
		}
	}
}

// getRequestID 获取请求ID
func (m *requestLoggerMiddleware) getRequestID(c *gin.Context) string {
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
func (m *requestLoggerMiddleware) OnStart() error {
	return nil
}

// OnStop 服务器停止时触发
func (m *requestLoggerMiddleware) OnStop() error {
	return nil
}
