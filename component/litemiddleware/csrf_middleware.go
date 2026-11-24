package litemiddleware

import (
	"crypto/subtle"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/lite-lake/litecore-go/common"
	"github.com/lite-lake/litecore-go/util/rand"
)

const (
	csrfTokenLength   = 32
	csrfCookieName    = "XSRF-TOKEN"
	csrfHeaderName    = "X-XSRF-TOKEN"
	csrfFormFieldName = "csrf_token"
	csrfContextKey    = "csrf_token"
)

// CSRFMiddleware CSRF防护中间件
type CSRFMiddleware struct {
	// Cookie名称（默认 XSRF-TOKEN，可按需覆盖以分离不同域的 token）
	CookieName string
	// 排除的路径前缀
	ExcludePaths []string
	// Cookie域名
	CookieDomain string
	// Cookie路径
	CookiePath string
	// Cookie是否仅HTTPS
	CookieSecure bool
	// Cookie是否HttpOnly
	CookieHttpOnly bool
	// Cookie有效期
	CookieMaxAge time.Duration
	// 是否使用SameSite
	SameSite http.SameSite
}

// NewCSRFMiddleware 创建CSRF中间件
func NewCSRFMiddleware() *CSRFMiddleware {
	return &CSRFMiddleware{
		CookieName:     csrfCookieName,
		ExcludePaths:   []string{"/api/oauth2/callback", "/.well-known/", "/api/oauth2/token", "/api/oauth2/revoke"},
		CookiePath:     "/",
		CookieSecure:   true,
		CookieHttpOnly: false,
		CookieMaxAge:   24 * time.Hour,
		SameSite:       http.SameSiteStrictMode,
	}
}

// MiddlewareName 中间件名称
func (m *CSRFMiddleware) MiddlewareName() string {
	return "CSRFMiddleware"
}

// Order 中间件顺序，在认证中间件之后
func (m *CSRFMiddleware) Order() int {
	return OrderCSRF
}

// Wrapper 中间件包装函数
func (m *CSRFMiddleware) Wrapper() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		// 跳过排除的路径
		for _, path := range m.ExcludePaths {
			if strings.HasPrefix(ctx.Request.URL.Path, path) {
				ctx.Next()
				return
			}
		}

		// Admin 委托流量精确豁免（SOP-service-route-and-auth §3.5/§3.8）：
		// AdminDelegationMiddleware 校验通过后置委托标记，此类请求凭密钥认证、无 Cookie
		// 会话上下文，不存在 CSRF 劫持面——跳过双提交校验，也不下发/轮换 XSRF-TOKEN Cookie；
		// 无标记的 Cookie 会话流量行为完全不变（双模式并存服务安全）
		if _, delegated := ctx.Get(ContextKeyAdminDelegated); delegated {
			ctx.Next()
			return
		}

		// GET、HEAD、OPTIONS、TRACE方法不需要校验
		if ctx.Request.Method == http.MethodGet ||
			ctx.Request.Method == http.MethodHead ||
			ctx.Request.Method == http.MethodOptions ||
			ctx.Request.Method == http.MethodTrace {
			// 确保CSRF令牌存在（复用已有Cookie或生成新令牌）
			m.ensureCSRFCookie(ctx)
			ctx.Next()
			return
		}

		// 校验CSRF令牌
		if !m.validateCSRFToken(ctx) {
			ctx.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"code":    http.StatusForbidden,
				"message": "CSRF token验证失败",
			})
			return
		}

		// 校验通过后轮换CSRF令牌
		m.rotateCSRFCookie(ctx)
		ctx.Next()
	}
}

// IssueCSRFCookie 确保响应携带 CSRF 令牌 Cookie（已持有效令牌则复用不覆盖，否则生成
// 新令牌下发）。供令牌豁免路径（如登录成功响应）主动引导双提交：Cookie 形态与
// ensureCSRFCookie/rotateCSRFCookie 产出完全一致（非 HttpOnly，SameSite/Path/Secure
// 随组件配置）。须在响应体写出前调用。
func (m *CSRFMiddleware) IssueCSRFCookie(ctx *gin.Context) {
	m.ensureCSRFCookie(ctx)
}

// OnStart 启动时初始化
func (m *CSRFMiddleware) OnStart() error {
	return nil
}

// OnStop 停止时清理
func (m *CSRFMiddleware) OnStop() error {
	return nil
}

// ensureCSRFCookie 确保CSRF令牌Cookie存在，已有则复用避免覆盖
func (m *CSRFMiddleware) ensureCSRFCookie(ctx *gin.Context) {
	cookieName := m.getCookieName()
	// 如果已有有效的CSRF Cookie，复用令牌，不覆盖Cookie
	if existingToken, err := ctx.Cookie(cookieName); err == nil && existingToken != "" {
		ctx.Set(csrfContextKey, existingToken)
		return
	}

	// 没有Cookie时生成新的随机令牌
	token := rand.Rand.RandomString(csrfTokenLength)
	ctx.Set(csrfContextKey, token)

	// 设置Cookie
	http.SetCookie(ctx.Writer, &http.Cookie{
		Name:     cookieName,
		Value:    token,
		Domain:   m.CookieDomain,
		Path:     m.CookiePath,
		MaxAge:   int(m.CookieMaxAge.Seconds()),
		Secure:   m.CookieSecure,
		HttpOnly: m.CookieHttpOnly,
		SameSite: m.SameSite,
	})
}

// rotateCSRFCookie 轮换CSRF令牌Cookie（始终生成新令牌）
func (m *CSRFMiddleware) rotateCSRFCookie(ctx *gin.Context) {
	cookieName := m.getCookieName()
	token := rand.Rand.RandomString(csrfTokenLength)
	ctx.Set(csrfContextKey, token)

	http.SetCookie(ctx.Writer, &http.Cookie{
		Name:     cookieName,
		Value:    token,
		Domain:   m.CookieDomain,
		Path:     m.CookiePath,
		MaxAge:   int(m.CookieMaxAge.Seconds()),
		Secure:   m.CookieSecure,
		HttpOnly: m.CookieHttpOnly,
		SameSite: m.SameSite,
	})
}

// validateCSRFToken 校验CSRF令牌
func (m *CSRFMiddleware) validateCSRFToken(ctx *gin.Context) bool {
	// 从Header获取令牌
	token := ctx.GetHeader(csrfHeaderName)

	// 如果Header没有，从表单获取
	if token == "" {
		token = ctx.PostForm(csrfFormFieldName)
	}

	// 如果都没有，返回失败
	if token == "" {
		return false
	}

	// 从Cookie获取令牌
	cookieToken, err := ctx.Cookie(m.getCookieName())
	if err != nil {
		return false
	}

	// 比较令牌是否相等
	return subtle.ConstantTimeCompare([]byte(token), []byte(cookieToken)) == 1
}

// getCookieName 返回 Cookie 名称（兼容未设置时使用默认值）
func (m *CSRFMiddleware) getCookieName() string {
	if m.CookieName != "" {
		return m.CookieName
	}
	return csrfCookieName
}

var _ common.IBaseMiddleware = (*CSRFMiddleware)(nil)
