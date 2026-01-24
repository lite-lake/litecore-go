// Package litemiddleware 提供 HTTP 中间件组件，实现开箱即用的通用中间件功能。
//
// 核心特性：
//   - 统一接口：所有中间件实现 common.IBaseMiddleware 接口，保持一致的使用方式
//   - 灵活配置：配置属性使用指针类型，支持可选配置和默认值覆盖
//   - 依赖注入：通过 inject:"" 标签自动注入 Manager 组件（LoggerManager、LimiterManager、TelemetryManager）
//   - 执行顺序：预定义 Order 常量，支持自定义执行顺序，确保中间件按预期顺序执行
//   - 完整测试：所有中间件包含单元测试和示例代码，便于理解和使用
//
// 基本用法：
//
//	// 使用默认配置创建中间件
//	recovery := litemiddleware.NewRecoveryMiddlewareWithDefaults()
//	reqLogger := litemiddleware.NewRequestLoggerMiddlewareWithDefaults()
//	cors := litemiddleware.NewCorsMiddlewareWithDefaults()
//	security := litemiddleware.NewSecurityHeadersMiddlewareWithDefaults()
//	limiter := litemiddleware.NewRateLimiterMiddlewareWithDefaults()
//	telemetry := litemiddleware.NewTelemetryMiddlewareWithDefaults()
//
//	// 注册到容器
//	container.RegisterMiddleware(middlewareContainer, recovery)
//	container.RegisterMiddleware(middlewareContainer, reqLogger)
//	container.RegisterMiddleware(middlewareContainer, cors)
//	container.RegisterMiddleware(middlewareContainer, security)
//	container.RegisterMiddleware(middlewareContainer, limiter)
//	container.RegisterMiddleware(middlewareContainer, telemetry)
//
// 自定义配置：
//
//	// 自定义 CORS 配置
//	allowOrigins := []string{"https://example.com"}
//	allowCredentials := true
//	cors := litemiddleware.NewCorsMiddleware(&litemiddleware.CorsConfig{
//	    AllowOrigins:     &allowOrigins,
//	    AllowCredentials: &allowCredentials,
//	})
//
//	// 自定义限流配置
//	limit := 100
//	window := time.Minute
//	keyPrefix := "api"
//	limiter := litemiddleware.NewRateLimiterMiddleware(&litemiddleware.RateLimiterConfig{
//	    Limit:     &limit,
//	    Window:    &window,
//	    KeyPrefix: &keyPrefix,
//	})
//
// 执行顺序：
//
//	// 自定义中间件名称和执行顺序
//	name := "CustomMiddleware"
//	order := 350
//	customMiddleware := litemiddleware.NewCorsMiddleware(&litemiddleware.CorsConfig{
//	    Name:  &name,
//	    Order: &order,
//	})
//
// 配置属性说明：
//
// 所有中间件配置支持以下通用字段：
//   - Name (*string): 中间件名称，用于日志和标识
//   - Order (*int): 执行顺序，数值越小越先执行
//
// 配置属性使用指针类型（*string, *int, *bool 等），未配置的字段将使用默认值。
// 这种设计允许零值区分（如 false 与未配置）、默认值覆盖和可选配置。
//
// 可用中间件列表：
//
// | 中间件 | 功能 | 默认 Order | 依赖 |
// |--------|------|-----------|------|
// | Recovery | panic 恢复 | 0 | LoggerManager |
// | RequestLogger | 请求日志 | 50 | LoggerManager |
// | CORS | 跨域处理 | 100 | 无 |
// | SecurityHeaders | 安全头 | 150 | 无 |
// | RateLimiter | 限流 | 200 | LimiterManager, LoggerManager |
// | Telemetry | 遥测 | 250 | TelemetryManager |
//
// 预定义 Order 常量：
//   - OrderRecovery = 0: panic 恢复中间件（最先执行）
//   - OrderRequestLogger = 50: 请求日志中间件
//   - OrderCORS = 100: CORS 跨域中间件
//   - OrderSecurityHeaders = 150: 安全头中间件
//   - OrderRateLimiter = 200: 限流中间件
//   - OrderTelemetry = 250: 遥测中间件
//   - OrderAuth = 300: 认证中间件（预留）
//
// 业务自定义中间件建议从 Order 350 开始。
package litemiddleware
