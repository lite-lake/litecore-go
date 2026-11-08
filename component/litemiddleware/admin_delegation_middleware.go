package litemiddleware

import (
	"crypto/subtle"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/lite-lake/liteadmin-go"

	"github.com/lite-lake/litecore-go/common"
)

// ContextKeyAdminDelegated 委托认证标记的 gin.Context 键
// AdminDelegationMiddleware 校验通过后置入（值为 true），供两类下游组件消费：
//   - CSRFMiddleware：检测到标记即跳过双提交校验（委托流量无 Cookie，精确豁免，SOP §3.5/§3.8）；
//   - ZoneGuardMiddleware：双模式过渡期检测到标记直接放行（委托认证在更早档位已完成校验）。
//
// 标记只在 key 校验通过后才置入，进程内可信。
const ContextKeyAdminDelegated = "lite_admin_delegated"

// 委托认证失败原因（传给失败钩子与告警钩子，机器可读）
const (
	// DelegationFailureMissingKey 未携带 X-Lite-Admin-Key（严格模式下出现即异常调用）
	DelegationFailureMissingKey = "missing_key"
	// DelegationFailureInvalidKey 密钥不匹配（疑似伪造或轮换不同步）
	DelegationFailureInvalidKey = "invalid_key"
	// DelegationFailureInvalidID X-Lite-Admin-Id 格式非法
	DelegationFailureInvalidID = "invalid_admin_id"
	// DelegationFailureRejected 应用层 Resolve 回调拒绝
	DelegationFailureRejected = "resolve_rejected"
)

// AdminDelegationResolve 应用层会话注入回调：key 校验通过后由接入服务把委托身份
// 映射为自己的会话实体（如 *entities.Session{SubjectType: "admin"}）并注入自己的
// ContextKeySession，返回 false 视为拒绝（401）。
//
// 设计先例：ZoneGuardConfig.Resolve（回调注入）——*entities.Session 与会话 ContextKey
// 均为应用级定义，框架层不引用应用实体（PRD《Admin Center 统一管理平台设计方案》§4.4）。
type AdminDelegationResolve func(identity liteadmin.DelegatedIdentity, ctx *gin.Context) bool

// AdminDelegationFailureHook 失败处理钩子（限流计数/告警均由接入服务自行实现，框架只暴露接口）
type AdminDelegationFailureHook func(ctx *gin.Context, reason string)

// AdminDelegationConfig Admin 委托认证配置（协议详见 SOP-service-route-and-auth §3.8）
type AdminDelegationConfig struct {
	// Name 中间件名称（可选，默认 AdminDelegationMiddleware）
	Name string
	// Order 执行顺序（默认 OrderAdminDelegation = 300 认证槽位；双模式过渡期建议贴靠 54，
	// 先于本地会话守卫（55）执行）
	Order int
	// APIPrefix 守卫的 API 分区前缀（默认 "/api/admin"，精确路径与前缀路径均守卫）
	APIPrefix string
	// APIKeys 委托密钥数组（多把并存支持轮换，任一命中即通过）；严格模式下未配置属部署错误
	APIKeys []string
	// FallbackToSession 双模式开关（过渡期服务保留本地登录时置 true）：
	//   true  —— 请求未携带委托头时放行给后续中间件（应用自己的会话守卫继续处理）；
	//   false —— 无头化全量要求模式：未携带委托头一律 401（该分区无登录豁免路径）
	FallbackToSession bool
	// Resolve 应用层会话注入回调（见 AdminDelegationResolve）；nil 时仅置委托标记不注入会话
	Resolve AdminDelegationResolve
	// OnAuthFailure 认证失败钩子：每次失败必触发（接入服务在钩子内做限流计数，框架不做具体限流）
	OnAuthFailure AdminDelegationFailureHook
	// OnAlert 告警钩子：每次失败触发（告警通道与去频策略由接入服务自行实现）
	OnAlert AdminDelegationFailureHook
}

// AdminDelegationMiddleware Admin 委托认证中间件（Admin Center 委托协议，上游服务侧守卫）
//
// 守卫 /api/admin/**：校验 X-Lite-Admin-Key（crypto/subtle 常数时间比较，多把并存）、
// X-Lite-Admin-Id 格式（^[0-9]{1,18}$）；通过后置委托标记（ContextKeyAdminDelegated）
// 并经应用层回调（Resolve）注入会话 → 控制器零改动；失败 → 401 + APIResponse{code:401}。
type AdminDelegationMiddleware struct {
	config AdminDelegationConfig
}

// NewAdminDelegationMiddleware 创建 Admin 委托认证中间件
func NewAdminDelegationMiddleware(config *AdminDelegationConfig) *AdminDelegationMiddleware {
	cfg := AdminDelegationConfig{}
	if config != nil {
		cfg = *config
	}
	if cfg.APIPrefix == "" {
		cfg.APIPrefix = "/api/admin"
	}
	return &AdminDelegationMiddleware{config: cfg}
}

// MiddlewareName 返回中间件名称
func (m *AdminDelegationMiddleware) MiddlewareName() string {
	if m.config.Name != "" {
		return m.config.Name
	}
	return "AdminDelegationMiddleware"
}

// Order 返回中间件顺序（默认认证槽位 OrderAdminDelegation）
func (m *AdminDelegationMiddleware) Order() int {
	if m.config.Order != 0 {
		return m.config.Order
	}
	return OrderAdminDelegation
}

// Wrapper 返回 Gin 中间件处理函数
func (m *AdminDelegationMiddleware) Wrapper() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		path := ctx.Request.URL.Path
		// 仅守卫目标分区（精确路径与前缀路径），其余请求直接放行
		if path != m.config.APIPrefix && !strings.HasPrefix(path, m.config.APIPrefix+"/") {
			ctx.Next()
			return
		}

		// 未配置委托密钥：严格模式属部署错误（fail-closed 拒绝，参照 /api/open API Key 守卫惯例）；
		// 双模式视为未启用委托，放行给后续会话守卫
		if len(m.config.APIKeys) == 0 {
			if m.config.FallbackToSession {
				ctx.Next()
				return
			}
			ctx.AbortWithStatusJSON(http.StatusInternalServerError, common.InternalError("管理员委托密钥未配置"))
			return
		}

		key := ctx.GetHeader(liteadmin.HeaderAdminKey)

		// 未携带委托头：双模式回落给后续会话守卫；严格模式全量要求委托头
		if key == "" {
			if m.config.FallbackToSession {
				ctx.Next()
				return
			}
			m.reject(ctx, DelegationFailureMissingKey, "缺少管理员委托凭据")
			return
		}

		// 委托密钥校验：常数时间比较，任一配置命中即通过（多把并存支持轮换）
		if !m.matchKey(key) {
			m.reject(ctx, DelegationFailureInvalidKey, "无效的管理员委托凭据")
			return
		}

		// 管理员 ID 格式校验（仅格式，不验存在性——信任边界是密钥）
		adminID := ctx.GetHeader(liteadmin.HeaderAdminID)
		if !liteadmin.IsValidAdminID(adminID) {
			m.reject(ctx, DelegationFailureInvalidID, "无效的管理员委托身份")
			return
		}

		// 应用层回调：由接入服务构造自己的会话实体并注入自己的 ContextKey
		identity := liteadmin.NewDelegatedSession(adminID, ctx.GetHeader(liteadmin.HeaderAdminUsername))
		if m.config.Resolve != nil && !m.config.Resolve(identity, ctx) {
			m.reject(ctx, DelegationFailureRejected, "未登录或会话已过期")
			return
		}

		// 置委托标记：CSRF 中间件据此精确豁免双提交校验；双模式下的 ZoneGuard 据此放行
		ctx.Set(ContextKeyAdminDelegated, true)
		ctx.Next()
	}
}

// matchKey 常数时间逐一比较配置密钥（遍历全部配置，避免短路造成的时序差异）
func (m *AdminDelegationMiddleware) matchKey(provided string) bool {
	matched := false
	for _, candidate := range m.config.APIKeys {
		if subtle.ConstantTimeCompare([]byte(provided), []byte(candidate)) == 1 {
			matched = true
		}
	}
	return matched
}

// reject 认证失败统一出口：触发钩子（限流计数/告警由接入服务实现）并返回 401
func (m *AdminDelegationMiddleware) reject(ctx *gin.Context, reason, message string) {
	if m.config.OnAuthFailure != nil {
		m.config.OnAuthFailure(ctx, reason)
	}
	if m.config.OnAlert != nil {
		m.config.OnAlert(ctx, reason)
	}
	ctx.AbortWithStatusJSON(http.StatusUnauthorized, common.UnauthorizedError(message))
}

// OnStart 中间件启动时调用
func (m *AdminDelegationMiddleware) OnStart() error {
	return nil
}

// OnStop 中间件停止时调用
func (m *AdminDelegationMiddleware) OnStop() error {
	return nil
}

var _ common.IBaseMiddleware = (*AdminDelegationMiddleware)(nil)
