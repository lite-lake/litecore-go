package litemiddleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/lite-lake/litecore-go/common"
)

// ZoneGuardConfig 分区守卫配置
// 守卫一个「页面 + API」分区（如 admin 区：/admin 页面 + /api/admin 接口），
// 页面未登录 302 到登录页，API 未登录返回 401（见《SOP-service-route-and-auth.md》§3.1）
type ZoneGuardConfig struct {
	// Name 中间件名称（可选，默认按 API 前缀生成）
	Name string
	// Order 执行顺序（认证授权档位建议 50-99）
	Order int
	// APIPrefix API 分区前缀（如 "/api/admin"），精确路径与前缀路径均守卫
	APIPrefix string
	// PagePrefix 页面分区前缀（如 "/admin"），精确路径与前缀路径均守卫
	PagePrefix string
	// LoginPage 页面登录页路径（豁免，如 "/admin/login"）
	LoginPage string
	// LoginAPI 登录接口路径（豁免，如 "/api/admin/auth/login"）
	LoginAPI string
	// CookieName 会话 Cookie 名
	CookieName string
	// CookieSecure 清除会话 Cookie 时是否带 Secure 属性（prod 环境应为 true）
	CookieSecure bool
	// LoginRedirect 页面未登录时的 302 目标（如 "/admin/login"）
	LoginRedirect string
	// Resolve 会话校验回调（三态）：
	//   (true, nil)            → 已认证，回调应自行向 gin.Context 注入会话信息供控制器使用
	//   (false, nil)           → 未认证（无效/过期/主体类型不符）→ 401/302 并清除会话 Cookie
	//   (false, 非 nil error)  → 认证子系统瞬时故障（如 DB 抖动）→ 503 且保留 Cookie，
	//                            不清会话、不引导登录——严格区分「服务暂不可用」与「未登录」，
	//                            防止把在线用户误踢下线（AC 2026-10 事故：
	//                            SQLite busy 被当未登录清 Cookie）
	Resolve func(sessionID string, ctx *gin.Context) (bool, error)
}

// ZoneGuardMiddleware 分区守卫中间件
// 路径自门禁：命中分区前缀的请求必须携带有效会话，登录入口本身豁免
type ZoneGuardMiddleware struct {
	config ZoneGuardConfig
}

// NewZoneGuardMiddleware 创建分区守卫中间件
func NewZoneGuardMiddleware(config *ZoneGuardConfig) *ZoneGuardMiddleware {
	return &ZoneGuardMiddleware{config: *config}
}

// MiddlewareName 返回中间件名称
func (m *ZoneGuardMiddleware) MiddlewareName() string {
	if m.config.Name != "" {
		return m.config.Name
	}
	return "ZoneGuardMiddleware[" + m.config.APIPrefix + "]"
}

// Order 返回中间件顺序
func (m *ZoneGuardMiddleware) Order() int {
	return m.config.Order
}

// Wrapper 返回 Gin 中间件处理函数
// 页面未登录 302 到登录页，API 未登录返回 401，无效会话 Cookie 即时清除
func (m *ZoneGuardMiddleware) Wrapper() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		// 已通过 Admin 委托认证的请求视为已认证（AdminDelegationMiddleware 在更早档位完成
		// 密钥校验并置委托标记，双模式过渡期两者并存，见 SOP-service-route-and-auth §3.8）：
		// Cookie 守卫逻辑不适用于无 Cookie 的委托流量，直接放行
		if _, delegated := ctx.Get(ContextKeyAdminDelegated); delegated {
			ctx.Next()
			return
		}

		path := ctx.Request.URL.Path

		isAPI := path == m.config.APIPrefix || strings.HasPrefix(path, m.config.APIPrefix+"/")
		isPage := path == m.config.PagePrefix || strings.HasPrefix(path, m.config.PagePrefix+"/")
		if !isAPI && !isPage {
			ctx.Next()
			return
		}

		// 登录页与登录接口本身豁免
		if path == m.config.LoginPage || path == m.config.LoginAPI {
			ctx.Next()
			return
		}

		sessionID, err := ctx.Cookie(m.config.CookieName)
		if err == nil && sessionID != "" && m.config.Resolve != nil {
			authed, rerr := m.config.Resolve(sessionID, ctx)
			if rerr != nil {
				// 认证子系统瞬时故障：503（前端按服务不可用降级重试；
				// 保留 Cookie，会话是否有效留待恢复后判定）
				ctx.AbortWithStatusJSON(http.StatusServiceUnavailable,
					common.ErrorWithCode(http.StatusServiceUnavailable, "认证服务暂时不可用，请稍后重试"))
				return
			}
			if authed {
				ctx.Next()
				return
			}
		}

		m.clearSessionCookie(ctx)

		if isAPI {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, common.UnauthorizedError("未登录或会话已过期"))
			return
		}

		ctx.Redirect(http.StatusFound, m.config.LoginRedirect)
		ctx.Abort()
	}
}

// clearSessionCookie 清除无效/过期的会话 Cookie
func (m *ZoneGuardMiddleware) clearSessionCookie(ctx *gin.Context) {
	ctx.SetSameSite(http.SameSiteLaxMode)
	ctx.SetCookie(m.config.CookieName, "", -1, "/", "", m.config.CookieSecure, true)
}

// OnStart 中间件启动时调用
func (m *ZoneGuardMiddleware) OnStart() error {
	return nil
}

// OnStop 中间件停止时调用
func (m *ZoneGuardMiddleware) OnStop() error {
	return nil
}

var _ common.IBaseMiddleware = (*ZoneGuardMiddleware)(nil)
