package litemiddleware_test

import (
	"time"

	"github.com/gin-gonic/gin"

	"github.com/lite-lake/litecore-go/component/litemiddleware"
)

// 示例 1: 使用默认配置创建中间件
func Example_newMiddlewareWithDefaults() {
	// 使用默认配置
	cors := litemiddleware.NewCorsMiddlewareWithDefaults()
	recovery := litemiddleware.NewRecoveryMiddlewareWithDefaults()
	security := litemiddleware.NewSecurityHeadersMiddlewareWithDefaults()
	reqLogger := litemiddleware.NewRequestLoggerMiddlewareWithDefaults()

	_ = cors
	_ = recovery
	_ = security
	_ = reqLogger
}

// 示例 2: 自定义 CORS 配置
func Example_customCorsConfig() {
	// 自定义 CORS 配置
	cfg := &litemiddleware.CorsConfig{
		AllowOrigins:     []string{"https://example.com", "https://app.example.com"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE"},
		AllowHeaders:     []string{"Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           8 * time.Hour,
	}
	cors := litemiddleware.NewCorsMiddleware(cfg)

	_ = cors
}

// 示例 3: 自定义请求日志配置
func Example_customRequestLoggerConfig() {
	// 自定义请求日志配置
	cfg := &litemiddleware.RequestLoggerConfig{
		Enable:          true,
		LogBody:         true,
		MaxBodySize:     2048,
		SkipPaths:       []string{"/health", "/metrics", "/ping"},
		LogHeaders:      []string{"User-Agent", "Content-Type", "X-Request-ID"},
		SuccessLogLevel: "debug",
	}
	reqLogger := litemiddleware.NewRequestLoggerMiddleware(cfg)

	_ = reqLogger
}

// 示例 4: 自定义安全头配置
func Example_customSecurityHeadersConfig() {
	// 自定义安全头配置
	cfg := &litemiddleware.SecurityHeadersConfig{
		FrameOptions:            "SAMEORIGIN",
		ContentTypeOptions:      "nosniff",
		XSSProtection:           "1; mode=block",
		ReferrerPolicy:          "strict-origin-when-cross-origin",
		ContentSecurityPolicy:   "default-src 'self'; script-src 'self' 'unsafe-inline'",
		StrictTransportSecurity: "max-age=31536000; includeSubDomains",
	}
	security := litemiddleware.NewSecurityHeadersMiddleware(cfg)

	_ = security
}

// 示例 5: 自定义 Recovery 配置
func Example_customRecoveryConfig() {
	// 自定义 Recovery 配置
	cfg := &litemiddleware.RecoveryConfig{
		PrintStack:      true,
		CustomErrorBody: true,
		ErrorMessage:    "服务器内部错误，请稍后重试",
		ErrorCode:       "SERVER_ERROR",
	}
	recovery := litemiddleware.NewRecoveryMiddleware(cfg)

	_ = recovery
}

// 示例 6: 自定义限流配置
func Example_customRateLimiterConfig() {
	// 按自定义配置创建限流中间件
	cfg := &litemiddleware.RateLimiterConfig{
		Limit:     100,
		Window:    time.Minute,
		KeyPrefix: "api",
		KeyFunc: func(c *gin.Context) string {
			// 自定义 key 生成逻辑
			// 可以基于请求路径、用户ID、IP等生成唯一key
			return c.ClientIP()
		},
		SkipFunc: func(c *gin.Context) bool {
			// 跳过某些请求的限流检查
			// 例如：内部IP、白名单用户等
			return false
		},
	}
	limiter := litemiddleware.NewRateLimiterMiddleware(cfg)

	_ = limiter
}

// 示例 7: 使用配置创建不同类型的限流中间件
func Example_rateLimiterConfigurations() {
	// 按IP限流
	byIP := litemiddleware.NewRateLimiterMiddleware(&litemiddleware.RateLimiterConfig{
		Limit:     100,
		Window:    time.Minute,
		KeyPrefix: "ip",
	})
	// 按路径限流
	byPath := litemiddleware.NewRateLimiterMiddleware(&litemiddleware.RateLimiterConfig{
		Limit:     200,
		Window:    time.Minute,
		KeyPrefix: "path",
		KeyFunc: func(c *gin.Context) string {
			return c.Request.URL.Path
		},
	})
	// 按请求头限流
	byHeader := litemiddleware.NewRateLimiterMiddleware(&litemiddleware.RateLimiterConfig{
		Limit:     50,
		Window:    time.Minute,
		KeyPrefix: "header",
		KeyFunc: func(c *gin.Context) string {
			return c.GetHeader("X-User-ID")
		},
	})
	// 按用户ID限流
	byUserID := litemiddleware.NewRateLimiterMiddleware(&litemiddleware.RateLimiterConfig{
		Limit:     10,
		Window:    time.Minute,
		KeyPrefix: "user",
		KeyFunc: func(c *gin.Context) string {
			if userID, exists := c.Get("user_id"); exists {
				if uid, ok := userID.(string); ok {
					return uid
				}
			}
			return c.ClientIP()
		},
	})

	_ = byIP
	_ = byPath
	_ = byHeader
	_ = byUserID
}

// 示例 8: 闭包配置生产环境 CORS
func Example_productionCorsConfig() {
	// 生产环境 CORS 配置（仅允许特定域名）
	cfg := &litemiddleware.CorsConfig{
		AllowOrigins: []string{
			"https://example.com",
			"https://www.example.com",
		},
		AllowMethods: []string{
			"GET",
			"POST",
			"PUT",
			"DELETE",
			"OPTIONS",
		},
		AllowHeaders: []string{
			"Origin",
			"Content-Type",
			"Authorization",
			"Accept",
		},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}
	cors := litemiddleware.NewCorsMiddleware(cfg)

	_ = cors
}

// 示例 9: 关闭请求日志
func Example_disableRequestLogger() {
	// 完全禁用请求日志
	cfg := &litemiddleware.RequestLoggerConfig{
		Enable: false,
	}
	reqLogger := litemiddleware.NewRequestLoggerMiddleware(cfg)

	_ = reqLogger
}

// 示例 10: 关闭 Recovery 堆栈打印（生产环境可能不需要）
func Example_recoveryWithoutStack() {
	// 不打印堆栈信息（生产环境可能为了性能）
	cfg := &litemiddleware.RecoveryConfig{
		PrintStack:      false,
		CustomErrorBody: true,
		ErrorMessage:    "系统错误",
		ErrorCode:       "SYSTEM_ERROR",
	}
	recovery := litemiddleware.NewRecoveryMiddleware(cfg)

	_ = recovery
}
