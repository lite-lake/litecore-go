package liteadmin

// SubjectTypeAdmin 委托会话主体类型（与 SOP-service-route-and-auth §3.2 统一 Session
// 的 subject_type=admin 一致）
const SubjectTypeAdmin = "admin"

// DelegatedIdentity 委托管理员身份（轻量结构体）
// 注意：*entities.Session 与会话 ContextKey 均为应用级定义（各应用自己的 entities/constants
// 包），本包不引用任何应用实体；接入服务在 AdminDelegationMiddleware 的 Resolve 回调里把
// 本结构映射为自己的会话类型并注入自己的 ContextKey（PRD §4.4 回调注入设计）。
type DelegatedIdentity struct {
	// SubjectType 主体类型，恒为 SubjectTypeAdmin（"admin"）
	SubjectType string
	// SubjectID 主体 ID，即请求头 X-Lite-Admin-Id 的值
	SubjectID string
	// Username 管理员用户名，即请求头 X-Lite-Admin-Username 的值（可为空，仅日志用途）
	Username string
}

// NewDelegatedSession 构造委托身份的字段助手（减少各接入服务样板代码）：
// SubjectType 固定填 "admin"，SubjectID 取自 X-Lite-Admin-Id，Username 取自 X-Lite-Admin-Username。
func NewDelegatedSession(adminID, username string) DelegatedIdentity {
	return DelegatedIdentity{
		SubjectType: SubjectTypeAdmin,
		SubjectID:   adminID,
		Username:    username,
	}
}
