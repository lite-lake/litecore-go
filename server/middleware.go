package server

import (
	"bytes"
	"io"
	"time"

	"github.com/gin-gonic/gin"

	"com.litelake.litecore/common"
)

// registerMiddlewares 注册中间件
func (e *Engine) registerMiddlewares() error {
	// 1. 注册 panic 恢复中间件
	if e.serverConfig.EnableRecovery {
		e.ginEngine.Use(gin.Recovery())
	}

	// 2. 注册请求日志中间件
	e.ginEngine.Use(requestLoggerMiddleware())

	// 3. 注册可观测性中间件（如果有 TelemetryManager）
	if telemetry := e.getTelemetryManager(); telemetry != nil {
		if otelMiddleware, ok := telemetry.(interface {
			GinMiddleware() gin.HandlerFunc
		}); ok {
			e.ginEngine.Use(otelMiddleware.GinMiddleware())
		}
	}

	// 4. 注册业务中间件（全局）
	middlewares := e.middleware.GetAll()

	// 按顺序排序
	sortedMiddlewares := sortMiddlewares(middlewares)

	for _, mw := range sortedMiddlewares {
		// 注册全局中间件
		e.ginEngine.Use(mw.Wrapper())
	}

	return nil
}

// requestLoggerMiddleware 请求日志中间件
func requestLoggerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 记录开始时间
		start := time.Now()

		// 读取请求体（如果需要）
		var bodyBytes []byte
		if c.Request.Body != nil && c.Request.Method != "GET" {
			bodyBytes, _ = io.ReadAll(c.Request.Body)
			c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
		}

		// 处理请求
		c.Next()

		// 记录日志
		latency := time.Since(start)
		status := c.Writer.Status()
		clientIP := c.ClientIP()
		method := c.Request.Method
		path := c.Request.URL.Path

		// 这里可以使用 LoggerManager 记录日志
		// 暂时使用简单的格式化输出
		if len(c.Errors) > 0 {
			// 有错误
			for _, e := range c.Errors {
				println("[ERROR]", e.Error())
			}
		} else {
			// 正常请求
			println("[INFO]", clientIP, method, path, status, latency)
		}
	}
}

// sortMiddlewares 按 Order 排序中间件
func sortMiddlewares(middlewares []common.BaseMiddleware) []common.BaseMiddleware {
	// 简单的冒泡排序（因为中间件数量通常不多）
	sorted := make([]common.BaseMiddleware, len(middlewares))
	copy(sorted, middlewares)

	n := len(sorted)
	for i := 0; i < n-1; i++ {
		for j := 0; j < n-i-1; j++ {
			if sorted[j].Order() > sorted[j+1].Order() {
				sorted[j], sorted[j+1] = sorted[j+1], sorted[j]
			}
		}
	}

	return sorted
}

// getTelemetryManager 获取遥测管理器
func (e *Engine) getTelemetryManager() common.BaseManager {
	managers := e.manager.GetAll()
	for _, mgr := range managers {
		if mgr.ManagerName() == "TelemetryManager" {
			return mgr
		}
	}
	return nil
}

// Use 注册全局中间件
func (e *Engine) Use(middleware ...gin.HandlerFunc) {
	e.ginEngine.Use(middleware...)
}

// RegisterMiddleware 注册自定义中间件
func (e *Engine) RegisterMiddleware(middleware gin.HandlerFunc) {
	e.ginEngine.Use(middleware)
}

// CorsMiddleware CORS 跨域中间件
func CorsMiddleware() gin.HandlerFunc {
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

// SecurityHeadersMiddleware 安全头中间件
func SecurityHeadersMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("X-Frame-Options", "DENY")
		c.Writer.Header().Set("X-Content-Type-Options", "nosniff")
		c.Writer.Header().Set("X-XSS-Protection", "1; mode=block")
		c.Writer.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
		c.Next()
	}
}
