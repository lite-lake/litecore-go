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
	allowOrigins := []string{"https://example.com", "https://app.example.com"}
	allowMethods := []string{"GET", "POST", "PUT", "DELETE"}
	allowHeaders := []string{"Content-Type", "Authorization"}
	exposeHeaders := []string{"Content-Length"}
	allowCredentials := true
	maxAge := 8 * time.Hour
	cfg := &litemiddleware.CorsConfig{
		AllowOrigins:     &allowOrigins,
		AllowMethods:     &allowMethods,
		AllowHeaders:     &allowHeaders,
		ExposeHeaders:    &exposeHeaders,
		AllowCredentials: &allowCredentials,
		MaxAge:           &maxAge,
	}
	cors := litemiddleware.NewCorsMiddleware(cfg)

	_ = cors
}

// 示例 3: 自定义请求日志配置
func Example_customRequestLoggerConfig() {
	// 自定义请求日志配置
	enable := true
	logBody := true
	maxBodySize := 2048
	skipPaths := []string{"/health", "/metrics", "/ping"}
	logHeaders := []string{"User-Agent", "Content-Type", "X-Request-ID"}
	successLogLevel := "debug"
	cfg := &litemiddleware.RequestLoggerConfig{
		Enable:          &enable,
		LogBody:         &logBody,
		MaxBodySize:     &maxBodySize,
		SkipPaths:       &skipPaths,
		LogHeaders:      &logHeaders,
		SuccessLogLevel: &successLogLevel,
	}
	reqLogger := litemiddleware.NewRequestLoggerMiddleware(cfg)

	_ = reqLogger
}

// 示例 4: 自定义安全头配置
func Example_customSecurityHeadersConfig() {
	// 自定义安全头配置
	frameOptions := "SAMEORIGIN"
	contentTypeOptions := "nosniff"
	xssProtection := "1; mode=block"
	referrerPolicy := "strict-origin-when-cross-origin"
	contentSecurityPolicy := "default-src 'self'; script-src 'self' 'unsafe-inline'"
	strictTransportSecurity := "max-age=31536000; includeSubDomains"
	cfg := &litemiddleware.SecurityHeadersConfig{
		FrameOptions:            &frameOptions,
		ContentTypeOptions:      &contentTypeOptions,
		XSSProtection:           &xssProtection,
		ReferrerPolicy:          &referrerPolicy,
		ContentSecurityPolicy:   &contentSecurityPolicy,
		StrictTransportSecurity: &strictTransportSecurity,
	}
	security := litemiddleware.NewSecurityHeadersMiddleware(cfg)

	_ = security
}

// 示例 5: 自定义 Recovery 配置
func Example_customRecoveryConfig() {
	// 自定义 Recovery 配置
	printStack := true
	customErrorBody := true
	errorMessage := "服务器内部错误，请稍后重试"
	errorCode := "SERVER_ERROR"
	cfg := &litemiddleware.RecoveryConfig{
		PrintStack:      &printStack,
		CustomErrorBody: &customErrorBody,
		ErrorMessage:    &errorMessage,
		ErrorCode:       &errorCode,
	}
	recovery := litemiddleware.NewRecoveryMiddleware(cfg)

	_ = recovery
}

// 示例 6: 自定义限流配置
func Example_customRateLimiterConfig() {
	// 按自定义配置创建限流中间件
	limit := 100
	window := time.Minute
	keyPrefix := "api"
	cfg := &litemiddleware.RateLimiterConfig{
		Limit:     &limit,
		Window:    &window,
		KeyPrefix: &keyPrefix,
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
	limit1 := 100
	window1 := time.Minute
	keyPrefix1 := "ip"
	byIP := litemiddleware.NewRateLimiterMiddleware(&litemiddleware.RateLimiterConfig{
		Limit:     &limit1,
		Window:    &window1,
		KeyPrefix: &keyPrefix1,
	})
	// 按路径限流
	limit2 := 200
	window2 := time.Minute
	keyPrefix2 := "path"
	byPath := litemiddleware.NewRateLimiterMiddleware(&litemiddleware.RateLimiterConfig{
		Limit:     &limit2,
		Window:    &window2,
		KeyPrefix: &keyPrefix2,
		KeyFunc: func(c *gin.Context) string {
			return c.Request.URL.Path
		},
	})
	// 按请求头限流
	limit3 := 50
	window3 := time.Minute
	keyPrefix3 := "header"
	byHeader := litemiddleware.NewRateLimiterMiddleware(&litemiddleware.RateLimiterConfig{
		Limit:     &limit3,
		Window:    &window3,
		KeyPrefix: &keyPrefix3,
		KeyFunc: func(c *gin.Context) string {
			return c.GetHeader("X-User-ID")
		},
	})
	// 按用户ID限流
	limit4 := 10
	window4 := time.Minute
	keyPrefix4 := "user"
	byUserID := litemiddleware.NewRateLimiterMiddleware(&litemiddleware.RateLimiterConfig{
		Limit:     &limit4,
		Window:    &window4,
		KeyPrefix: &keyPrefix4,
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
	allowOrigins := []string{
		"https://example.com",
		"https://www.example.com",
	}
	allowMethods := []string{
		"GET",
		"POST",
		"PUT",
		"DELETE",
		"OPTIONS",
	}
	allowHeaders := []string{
		"Origin",
		"Content-Type",
		"Authorization",
		"Accept",
	}
	allowCredentials := true
	maxAge := 12 * time.Hour
	cfg := &litemiddleware.CorsConfig{
		AllowOrigins:     &allowOrigins,
		AllowMethods:     &allowMethods,
		AllowHeaders:     &allowHeaders,
		AllowCredentials: &allowCredentials,
		MaxAge:           &maxAge,
	}
	cors := litemiddleware.NewCorsMiddleware(cfg)

	_ = cors
}

// 示例 9: 关闭请求日志
func Example_disableRequestLogger() {
	// 完全禁用请求日志
	enable := false
	cfg := &litemiddleware.RequestLoggerConfig{
		Enable: &enable,
	}
	reqLogger := litemiddleware.NewRequestLoggerMiddleware(cfg)

	_ = reqLogger
}

// 示例 10: 关闭 Recovery 堆栈打印（生产环境可能不需要）
func Example_recoveryWithoutStack() {
	// 不打印堆栈信息（生产环境可能为了性能）
	printStack := false
	customErrorBody := true
	errorMessage := "系统错误"
	errorCode := "SYSTEM_ERROR"
	cfg := &litemiddleware.RecoveryConfig{
		PrintStack:      &printStack,
		CustomErrorBody: &customErrorBody,
		ErrorMessage:    &errorMessage,
		ErrorCode:       &errorCode,
	}
	recovery := litemiddleware.NewRecoveryMiddleware(cfg)

	_ = recovery
}
