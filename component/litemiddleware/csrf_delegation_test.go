package litemiddleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// newCSRFTestRouter 构造 CSRF 测试引擎：markDelegated=true 时模拟 AdminDelegationMiddleware
// 已在更早档位校验通过并置委托标记
func newCSRFTestRouter(csrf *CSRFMiddleware, markDelegated bool, handler gin.HandlerFunc) *gin.Engine {
	router := gin.New()
	router.Use(func(ctx *gin.Context) {
		if markDelegated {
			ctx.Set(ContextKeyAdminDelegated, true)
		}
		ctx.Next()
	})
	router.Use(csrf.Wrapper())
	router.Any("/*any", handler)
	return router
}

// performCSRFRaw 发起请求
func performCSRFRaw(router *gin.Engine, method, path string, headers map[string]string, cookies []*http.Cookie) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, path, nil)
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	for _, c := range cookies {
		req.AddCookie(c)
	}
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	return w
}

// csrfTokenFromCookies 从响应 Cookie 中取 XSRF-TOKEN 令牌
func csrfTokenFromCookies(t *testing.T, resp *http.Response) (token string, cookie *http.Cookie) {
	t.Helper()
	for _, c := range resp.Cookies() {
		if c.Name == csrfCookieName && c.Value != "" {
			return c.Value, c
		}
	}
	t.Fatal("未下发 XSRF-TOKEN Cookie")
	return "", nil
}

func TestCSRFMiddleware_委托标记精确豁免(t *testing.T) {
	csrf := NewCSRFMiddleware()

	t.Run("带委托标记的写请求跳过双提交校验且不下发 Cookie", func(t *testing.T) {
		called := false
		router := newCSRFTestRouter(csrf, true, func(ctx *gin.Context) {
			called = true
			ctx.Status(http.StatusOK)
		})

		w := performCSRFRaw(router, http.MethodPost, "/api/admin/messages", nil, nil)

		assert.Equal(t, http.StatusOK, w.Code, "委托写请求不应被 CSRF 拦截")
		assert.True(t, called)
		assert.Empty(t, w.Result().Cookies(), "委托流量无 Cookie 上下文，不应下发/轮换 XSRF-TOKEN")
	})

	t.Run("无标记的写请求照常 403", func(t *testing.T) {
		called := false
		router := newCSRFTestRouter(csrf, false, func(ctx *gin.Context) {
			called = true
			ctx.Status(http.StatusOK)
		})

		w := performCSRFRaw(router, http.MethodPost, "/api/admin/messages", nil, nil)

		assert.Equal(t, http.StatusForbidden, w.Code, "Cookie 会话流量行为完全不变")
		assert.False(t, called)
		assert.Contains(t, w.Body.String(), "CSRF")
	})

	t.Run("无标记但双提交齐全的写请求照常通过", func(t *testing.T) {
		var passed bool
		router := newCSRFTestRouter(csrf, false, func(ctx *gin.Context) {
			passed = true
			ctx.Status(http.StatusOK)
		})

		// 先 GET 领取令牌（Cookie 会话流量的既有行为）
		wGet := performCSRFRaw(router, http.MethodGet, "/api/admin/messages", nil, nil)
		token, cookie := csrfTokenFromCookies(t, wGet.Result())
		require.NotEmpty(t, token)

		// 再按双提交约定回传（XSRF-TOKEN Cookie + X-XSRF-TOKEN 头）
		w := performCSRFRaw(router, http.MethodPost, "/api/admin/messages", map[string]string{csrfHeaderName: token}, []*http.Cookie{cookie})

		assert.Equal(t, http.StatusOK, w.Code, "双提交齐全的 Cookie 会话流量应照常通过")
		assert.True(t, passed)
	})
}
