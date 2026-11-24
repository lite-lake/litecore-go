package litemiddleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

// newTestZoneGuard 构造被测分区守卫中间件（admin 形态）
func newTestZoneGuard(resolve func(sessionID string, ctx *gin.Context) (bool, error)) *ZoneGuardMiddleware {
	return NewZoneGuardMiddleware(&ZoneGuardConfig{
		Name:          "TestAdminGuard",
		Order:         55,
		APIPrefix:     "/api/admin",
		PagePrefix:    "/admin",
		LoginPage:     "/admin/login",
		LoginAPI:      "/api/admin/auth/login",
		CookieName:    "test_admin_session",
		LoginRedirect: "/admin/login",
		Resolve:       resolve,
	})
}

// performZoneRequest 发起请求并捕获放行结果
func performZoneRequest(t *testing.T, guard *ZoneGuardMiddleware, method, path, cookieValue string) (*httptest.ResponseRecorder, bool) {
	t.Helper()
	gin.SetMode(gin.TestMode)

	called := false
	router := gin.New()
	router.Use(guard.Wrapper())
	router.NoRoute(func(ctx *gin.Context) {
		called = true
		ctx.Status(http.StatusOK)
	})

	req := httptest.NewRequest(method, path, nil)
	if cookieValue != "" {
		req.AddCookie(&http.Cookie{Name: "test_admin_session", Value: cookieValue})
	}
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	return w, called
}

func TestZoneGuardMiddleware_未认证请求(t *testing.T) {
	tests := []struct {
		name         string
		method       string
		path         string
		wantStatus   int
		wantRedirect string
		wantCalled   bool
	}{
		{name: "页面精确路径未登录跳转登录页", method: http.MethodGet, path: "/admin", wantStatus: http.StatusFound, wantRedirect: "/admin/login"},
		{name: "页面子路径未登录跳转登录页", method: http.MethodGet, path: "/admin/settings", wantStatus: http.StatusFound, wantRedirect: "/admin/login"},
		{name: "API 精确路径未登录返回 401", method: http.MethodGet, path: "/api/admin", wantStatus: http.StatusUnauthorized},
		{name: "API 子路径未登录返回 401", method: http.MethodGet, path: "/api/admin/messages", wantStatus: http.StatusUnauthorized},
		{name: "登录页豁免", method: http.MethodGet, path: "/admin/login", wantStatus: http.StatusOK, wantCalled: true},
		{name: "登录接口豁免", method: http.MethodPost, path: "/api/admin/auth/login", wantStatus: http.StatusOK, wantCalled: true},
		{name: "非本分区路径放行", method: http.MethodGet, path: "/api/user/messages", wantStatus: http.StatusOK, wantCalled: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			guard := newTestZoneGuard(func(sessionID string, ctx *gin.Context) (bool, error) { return false, nil })

			w, called := performZoneRequest(t, guard, tt.method, tt.path, "")

			assert.Equal(t, tt.wantStatus, w.Code)
			assert.Equal(t, tt.wantCalled, called)
			if tt.wantRedirect != "" {
				assert.Equal(t, tt.wantRedirect, w.Header().Get("Location"))
			}
		})
	}
}

func TestZoneGuardMiddleware_会话校验(t *testing.T) {
	t.Run("校验通过放行并注入上下文", func(t *testing.T) {
		guard := newTestZoneGuard(func(sessionID string, ctx *gin.Context) (bool, error) {
			ctx.Set("subject", "admin-1")
			return true, nil
		})

		gin.SetMode(gin.TestMode)
		var subject any
		router := gin.New()
		router.Use(guard.Wrapper())
		router.NoRoute(func(ctx *gin.Context) {
			subject, _ = ctx.Get("subject")
			ctx.Status(http.StatusOK)
		})
		req := httptest.NewRequest(http.MethodGet, "/api/admin/messages", nil)
		req.AddCookie(&http.Cookie{Name: "test_admin_session", Value: "sid-ok"})
		router.ServeHTTP(httptest.NewRecorder(), req)

		assert.Equal(t, "admin-1", subject)
	})

	t.Run("校验失败等同未登录", func(t *testing.T) {
		guard := newTestZoneGuard(func(sessionID string, ctx *gin.Context) (bool, error) { return false, nil })

		w, called := performZoneRequest(t, guard, http.MethodGet, "/api/admin/messages", "sid-bad")

		assert.Equal(t, http.StatusUnauthorized, w.Code)
		assert.False(t, called)
	})

	t.Run("校验回调瞬时故障 → 503 且保留会话 Cookie（不清会话不引导登录）", func(t *testing.T) {
		guard := newTestZoneGuard(func(sessionID string, ctx *gin.Context) (bool, error) {
			return false, assert.AnError
		})

		w, called := performZoneRequest(t, guard, http.MethodGet, "/api/admin/messages", "sid-ok")

		assert.Equal(t, http.StatusServiceUnavailable, w.Code)
		assert.False(t, called)
		assert.NotContains(t, w.Header().Get("Set-Cookie"), "Max-Age=0")
	})

	t.Run("未配置 Resolve 回调时一律未认证（fail-closed）", func(t *testing.T) {
		guard := NewZoneGuardMiddleware(&ZoneGuardConfig{
			APIPrefix:     "/api/admin",
			PagePrefix:    "/admin",
			CookieName:    "test_admin_session",
			LoginRedirect: "/admin/login",
		})

		w, _ := performZoneRequest(t, guard, http.MethodGet, "/api/admin/messages", "sid-ok")

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})
}

func TestZoneGuardMiddleware_清除无效会话Cookie(t *testing.T) {
	guard := newTestZoneGuard(func(sessionID string, ctx *gin.Context) (bool, error) { return false, nil })

	// prod 配置：清除 Cookie 应带 Secure 属性
	prodGuard := newTestZoneGuard(func(sessionID string, ctx *gin.Context) (bool, error) { return false, nil })
	prodGuard.config.CookieSecure = true

	tests := []struct {
		name       string
		guard      *ZoneGuardMiddleware
		path       string
		wantSecure bool
	}{
		{name: "API 401 响应清除 Cookie", guard: guard, path: "/api/admin/messages", wantSecure: false},
		{name: "prod 下清除 Cookie 带 Secure", guard: prodGuard, path: "/api/admin/messages", wantSecure: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w, _ := performZoneRequest(t, tt.guard, http.MethodGet, tt.path, "sid-bad")

			cookies := w.Result().Cookies()
			found := false
			for _, c := range cookies {
				if c.Name == "test_admin_session" {
					found = true
					assert.Equal(t, -1, c.MaxAge, "应立即过期")
					assert.Equal(t, tt.wantSecure, c.Secure)
				}
			}
			assert.True(t, found, "应下发清除 Cookie")
		})
	}
}

func TestZoneGuardMiddleware_名称与顺序(t *testing.T) {
	t.Run("显式名称优先", func(t *testing.T) {
		guard := newTestZoneGuard(nil)
		assert.Equal(t, "TestAdminGuard", guard.MiddlewareName())
		assert.Equal(t, 55, guard.Order())
	})
	t.Run("默认名称按 API 前缀生成", func(t *testing.T) {
		guard := NewZoneGuardMiddleware(&ZoneGuardConfig{APIPrefix: "/api/admin"})
		assert.Equal(t, "ZoneGuardMiddleware[/api/admin]", guard.MiddlewareName())
	})
}
