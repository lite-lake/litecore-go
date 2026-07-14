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
	// 如果已有有效的CSRF Cookie，复用令牌，不覆盖Cookie
	if existingToken, err := ctx.Cookie(csrfCookieName); err == nil && existingToken != "" {
		ctx.Set(csrfContextKey, existingToken)
		return
	}

	// 没有Cookie时生成新的随机令牌
	token := rand.Rand.RandomString(csrfTokenLength)
	ctx.Set(csrfContextKey, token)

	// 设置Cookie
	http.SetCookie(ctx.Writer, &http.Cookie{
		Name:     csrfCookieName,
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
	token := rand.Rand.RandomString(csrfTokenLength)
	ctx.Set(csrfContextKey, token)

	http.SetCookie(ctx.Writer, &http.Cookie{
		Name:     csrfCookieName,
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
	cookieToken, err := ctx.Cookie(csrfCookieName)
	if err != nil {
		return false
	}

	// 比较令牌是否相等
	return subtle.ConstantTimeCompare([]byte(token), []byte(cookieToken)) == 1
}

var _ common.IBaseMiddleware = (*CSRFMiddleware)(nil)
