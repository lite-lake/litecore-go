package litemiddleware

const (
	OrderRecovery        = 0   // panic 恢复中间件（最先执行）
	OrderRequestLogger   = 50  // 请求日志中间件
	OrderCORS            = 100 // CORS 跨域中间件
	OrderSecurityHeaders = 150 // 安全头中间件
	OrderRateLimiter     = 200 // 限流中间件（认证前执行）
	OrderTelemetry       = 250 // 遥测中间件
	OrderAuth            = 300 // 认证中间件
	OrderCSRF            = 350 // CSRF防护中间件

	// OrderAdminDelegation Admin 委托认证中间件（Admin Center 委托协议，PRD §6.5）
	// 与 OrderAuth 同值同档（「认证槽位」）：无头化服务中它独占该分区认证（取代 ZoneGuard 的
	// API 守卫角色），二者分区互斥、不在同一分区并存注册；双模式过渡期由接入服务以更早的
	// 贴靠档位（如 54，先于本地会话守卫 55）注册本组件
	OrderAdminDelegation = 300

	// 预留空间用于业务中间件：400, 450...
)
