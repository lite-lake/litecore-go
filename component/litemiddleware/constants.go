package litemiddleware

const (
	OrderRecovery        = 0   // panic 恢复中间件（最先执行）
	OrderRequestLogger   = 50  // 请求日志中间件
	OrderCORS            = 100 // CORS 跨域中间件
	OrderSecurityHeaders = 150 // 安全头中间件
	OrderRateLimiter     = 200 // 限流中间件（认证前执行）
	OrderTelemetry       = 250 // 遥测中间件
	OrderAuth            = 300 // 认证中间件

	// 预留空间用于业务中间件：350, 400, 450...
)
