package liteadmin

import "regexp"

// Admin 委托认证协议头（PRD《Admin Center 统一管理平台设计方案》§6.5、附录 A）。
// 本文件是全仓库头名的唯一权威：AC 侧、上游服务侧、框架中间件一律从此引用，
// 禁止在任何其他文件手打头名字符串（防笔误漂移）。
const (
	// HeaderAdminKey 委托密钥头：值须 ∈ 上游服务配置 admin_delegation.api_keys（多把并存支持轮换）
	HeaderAdminKey = "X-Lite-Admin-Key"
	// HeaderAdminID 管理员 ID 头：即 AC admins 表主键，格式 ^[0-9]{1,18}$
	HeaderAdminID = "X-Lite-Admin-Id"
	// HeaderAdminUsername 管理员用户名头：建议携带，仅透写入日志，不参与校验
	HeaderAdminUsername = "X-Lite-Admin-Username"
)

// Trace 头（沿用《SOP-gate-trace-id》，非本协议新增；AC 内网直连不过 gate，
// 由 AC 侧客户端负责生成/透传，保证 request_id 全链路可查）
const (
	// HeaderGateTraceID 网关链路追踪 ID 请求头
	HeaderGateTraceID = "X-Lite-Gate-Trace-Id"
	// HeaderTraceparent W3C Trace Context 请求头
	HeaderTraceparent = "traceparent"
)

// adminIDPattern 管理员 ID 格式正则：1~18 位纯数字（预留 bigint 全量程）
var adminIDPattern = regexp.MustCompile(`^[0-9]{1,18}$`)

// IsValidAdminID 校验 X-Lite-Admin-Id 的格式（仅格式校验，不验存在性——信任边界是密钥）
func IsValidAdminID(adminID string) bool {
	return adminIDPattern.MatchString(adminID)
}
