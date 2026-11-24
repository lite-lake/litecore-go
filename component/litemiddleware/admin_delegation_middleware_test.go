package litemiddleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/lite-lake/liteadmin-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	testKeyPrimary = "primary-key-aaaaaaaaaaaaaaaaaaaa"
	testKeyRotated = "rotated-key-bbbbbbbbbbbbbbbbbbbb"
)

// newTestDelegation 构造被测委托守卫（默认严格全量模式 + 双把密钥并存）
func newTestDelegation(mutate func(cfg *AdminDelegationConfig)) *AdminDelegationMiddleware {
	cfg := &AdminDelegationConfig{
		Name:      "TestDelegation",
		APIPrefix: "/api/admin",
		APIKeys:   []string{testKeyPrimary, testKeyRotated},
	}
	if mutate != nil {
		mutate(cfg)
	}
	return NewAdminDelegationMiddleware(cfg)
}

// performDelegation 发起请求并捕获放行结果与后续处理器捕获的 gin.Context
func performDelegation(t *testing.T, mw *AdminDelegationMiddleware, method, path string, headers map[string]string) (*httptest.ResponseRecorder, bool, *gin.Context) {
	t.Helper()
	gin.SetMode(gin.TestMode)

	called := false
	var captured *gin.Context
	router := gin.New()
	router.Use(mw.Wrapper())
	router.Any("/*any", func(ctx *gin.Context) {
		called = true
		captured = ctx
		ctx.Status(http.StatusOK)
	})

	req := httptest.NewRequest(method, path, nil)
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	return w, called, captured
}

// delegationHeaders 组装委托头
func delegationHeaders(key, adminID, username string) map[string]string {
	h := map[string]string{liteadmin.HeaderAdminKey: key}
	if adminID != "" {
		h[liteadmin.HeaderAdminID] = adminID
	}
	if username != "" {
		h[liteadmin.HeaderAdminUsername] = username
	}
	return h
}

func TestAdminDelegationMiddleware_key校验矩阵(t *testing.T) {
	tests := []struct {
		name     string
		mode     string // strict / fallback
		headers  map[string]string
		wantCode int
		wantNext bool
		wantMark bool // 后续处理器中是否可见委托标记
	}{
		{name: "正确的主密钥放行", mode: "strict", headers: delegationHeaders(testKeyPrimary, "42", "kent"), wantCode: http.StatusOK, wantNext: true, wantMark: true},
		{name: "轮换中的新密钥放行", mode: "strict", headers: delegationHeaders(testKeyRotated, "42", ""), wantCode: http.StatusOK, wantNext: true, wantMark: true},
		{name: "错误密钥 401", mode: "strict", headers: delegationHeaders("wrong-key", "42", ""), wantCode: http.StatusUnauthorized},
		{name: "密钥头缺失 401（严格模式）", mode: "strict", headers: map[string]string{liteadmin.HeaderAdminID: "42"}, wantCode: http.StatusUnauthorized},
		{name: "Id 缺失 401", mode: "strict", headers: delegationHeaders(testKeyPrimary, "", "kent"), wantCode: http.StatusUnauthorized},
		{name: "Id 含字母 401", mode: "strict", headers: delegationHeaders(testKeyPrimary, "42a", ""), wantCode: http.StatusUnauthorized},
		{name: "Id 超 18 位 401", mode: "strict", headers: delegationHeaders(testKeyPrimary, "1234567890123456789", ""), wantCode: http.StatusUnauthorized},
		{name: "非守卫分区放行且无标记", mode: "strict", headers: map[string]string{liteadmin.HeaderAdminKey: "anything"}, wantCode: http.StatusOK, wantNext: true},
		{name: "分区精确路径受守卫", mode: "strict", headers: map[string]string{}, wantCode: http.StatusUnauthorized},
		{name: "双模式无头回落放行且无标记", mode: "fallback", headers: map[string]string{}, wantCode: http.StatusOK, wantNext: true},
		{name: "双模式有效委托放行并置标记", mode: "fallback", headers: delegationHeaders(testKeyPrimary, "42", ""), wantCode: http.StatusOK, wantNext: true, wantMark: true},
		{name: "双模式带了错密钥仍 401（不静默回落）", mode: "fallback", headers: delegationHeaders("wrong-key", "42", ""), wantCode: http.StatusUnauthorized},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mw := newTestDelegation(func(cfg *AdminDelegationConfig) {
				cfg.FallbackToSession = tt.mode == "fallback"
			})
			path := "/api/admin/messages"
			if tt.name == "分区精确路径受守卫" {
				path = "/api/admin"
			} else if tt.name == "非守卫分区放行且无标记" {
				path = "/api/user/messages"
			}

			w, called, captured := performDelegation(t, mw, http.MethodGet, path, tt.headers)

			assert.Equal(t, tt.wantCode, w.Code)
			assert.Equal(t, tt.wantNext, called)
			wantMark := false
			if tt.wantMark {
				wantMark = true
			}
			if captured != nil {
				_, marked := captured.Get(ContextKeyAdminDelegated)
				assert.Equal(t, wantMark, marked, "委托标记应仅在密钥校验通过后置入")
			}
			if !tt.wantNext {
				// 失败响应应为 APIResponse{code:401}
				assert.Contains(t, w.Body.String(), `"code":401`)
			}
		})
	}
}

func TestAdminDelegationMiddleware_会话注入回调(t *testing.T) {
	t.Run("回调被调用并收到委托身份", func(t *testing.T) {
		var gotIdentity liteadmin.DelegatedIdentity
		resolveCalled := false
		mw := newTestDelegation(func(cfg *AdminDelegationConfig) {
			cfg.Resolve = func(identity liteadmin.DelegatedIdentity, ctx *gin.Context) bool {
				resolveCalled = true
				gotIdentity = identity
				ctx.Set("app_session", identity.SubjectID) // 模拟应用注入自己的 ContextKeySession
				return true
			}
		})

		w, called, _ := performDelegation(t, mw, http.MethodGet, "/api/admin/messages", delegationHeaders(testKeyPrimary, "10086", "kent"))

		assert.Equal(t, http.StatusOK, w.Code)
		assert.True(t, called)
		assert.True(t, resolveCalled, "Session 注入回调应被调用")
		assert.Equal(t, "admin", gotIdentity.SubjectType)
		assert.Equal(t, "10086", gotIdentity.SubjectID)
		assert.Equal(t, "kent", gotIdentity.Username)
	})

	t.Run("回调拒绝视为未认证", func(t *testing.T) {
		mw := newTestDelegation(func(cfg *AdminDelegationConfig) {
			cfg.Resolve = func(identity liteadmin.DelegatedIdentity, ctx *gin.Context) bool { return false }
		})
		w, called, _ := performDelegation(t, mw, http.MethodGet, "/api/admin/messages", delegationHeaders(testKeyPrimary, "42", ""))
		assert.Equal(t, http.StatusUnauthorized, w.Code)
		assert.False(t, called)
	})

	t.Run("未配置回调时仅置标记放行", func(t *testing.T) {
		mw := newTestDelegation(nil)
		w, called, captured := performDelegation(t, mw, http.MethodGet, "/api/admin/messages", delegationHeaders(testKeyPrimary, "42", ""))
		assert.Equal(t, http.StatusOK, w.Code)
		assert.True(t, called)
		require.NotNil(t, captured)
		_, marked := captured.Get(ContextKeyAdminDelegated)
		assert.True(t, marked)
	})
}

func TestAdminDelegationMiddleware_失败钩子(t *testing.T) {
	var failureReasons, alertReasons []string
	mw := newTestDelegation(func(cfg *AdminDelegationConfig) {
		cfg.OnAuthFailure = func(ctx *gin.Context, reason string) { failureReasons = append(failureReasons, reason) }
		cfg.OnAlert = func(ctx *gin.Context, reason string) { alertReasons = append(alertReasons, reason) }
	})

	tests := []struct {
		name       string
		headers    map[string]string
		wantReason string
	}{
		{name: "缺头触发 missing_key", headers: map[string]string{}, wantReason: DelegationFailureMissingKey},
		{name: "错 key 触发 invalid_key", headers: delegationHeaders("bad", "42", ""), wantReason: DelegationFailureInvalidKey},
		{name: "Id 格式错触发 invalid_admin_id", headers: delegationHeaders(testKeyPrimary, "x", ""), wantReason: DelegationFailureInvalidID},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			failureReasons, alertReasons = nil, nil
			w, _, _ := performDelegation(t, mw, http.MethodGet, "/api/admin/messages", tt.headers)
			assert.Equal(t, http.StatusUnauthorized, w.Code)
			assert.Equal(t, []string{tt.wantReason}, failureReasons, "失败钩子应收到失败原因")
			assert.Equal(t, []string{tt.wantReason}, alertReasons, "告警钩子应收到失败原因")
		})
	}
}

func TestAdminDelegationMiddleware_严格模式未配置密钥failClosed(t *testing.T) {
	// 未配置密钥且严格模式：部署错误，fail-closed 拒绝（500）
	mw := newTestDelegation(func(cfg *AdminDelegationConfig) { cfg.APIKeys = nil })
	w, called, _ := performDelegation(t, mw, http.MethodGet, "/api/admin/messages", delegationHeaders(testKeyPrimary, "42", ""))
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.False(t, called)

	// 双模式未配置密钥：视为未启用委托，回落给会话守卫
	fallback := newTestDelegation(func(cfg *AdminDelegationConfig) {
		cfg.APIKeys = nil
		cfg.FallbackToSession = true
	})
	w, called, captured := performDelegation(t, fallback, http.MethodGet, "/api/admin/messages", delegationHeaders(testKeyPrimary, "42", ""))
	assert.Equal(t, http.StatusOK, w.Code)
	assert.True(t, called)
	if captured != nil {
		_, marked := captured.Get(ContextKeyAdminDelegated)
		assert.False(t, marked, "未启用委托时不得置标记")
	}
}

func TestAdminDelegationMiddleware_双模式与会话守卫串联(t *testing.T) {
	// 双模式过渡期的真实拓扑：委托守卫（贴靠 54）先于本地会话守卫（ZoneGuard 55）执行。
	// 期望：带有效委托头的请求置标记直通 ZoneGuard；无委托头的请求回落给 ZoneGuard 的 Cookie 守卫。
	t.Run("委托请求直通会话守卫", func(t *testing.T) {
		delegation := newTestDelegation(func(cfg *AdminDelegationConfig) {
			cfg.FallbackToSession = true
			cfg.Resolve = func(identity liteadmin.DelegatedIdentity, ctx *gin.Context) bool {
				ctx.Set("app_session", identity.SubjectID)
				return true
			}
		})
		zoneGuard := NewZoneGuardMiddleware(&ZoneGuardConfig{
			APIPrefix: "/api/admin", PagePrefix: "/admin", CookieName: "t_admin_session", LoginRedirect: "/admin/login",
			Resolve: func(sessionID string, ctx *gin.Context) (bool, error) { return false, nil }, // 本地会话一律无效
		})

		gin.SetMode(gin.TestMode)
		var session any
		called := false
		router := gin.New()
		router.Use(delegation.Wrapper())
		router.Use(zoneGuard.Wrapper())
		router.GET("/api/admin/messages", func(ctx *gin.Context) {
			called = true
			session, _ = ctx.Get("app_session")
			ctx.Status(http.StatusOK)
		})

		req := httptest.NewRequest(http.MethodGet, "/api/admin/messages", nil)
		req.Header.Set(liteadmin.HeaderAdminKey, testKeyPrimary)
		req.Header.Set(liteadmin.HeaderAdminID, "42")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code, "有效委托头应直通本地会话守卫（无 Cookie 也不 401）")
		assert.True(t, called)
		assert.Equal(t, "42", session, "应用回调注入的会话应可被控制器取用")
	})

	t.Run("无委托头回落给会话守卫照常拦截", func(t *testing.T) {
		delegation := newTestDelegation(func(cfg *AdminDelegationConfig) { cfg.FallbackToSession = true })
		zoneGuard := NewZoneGuardMiddleware(&ZoneGuardConfig{
			APIPrefix: "/api/admin", PagePrefix: "/admin", CookieName: "t_admin_session", LoginRedirect: "/admin/login",
			Resolve: func(sessionID string, ctx *gin.Context) (bool, error) { return false, nil },
		})

		gin.SetMode(gin.TestMode)
		called := false
		router := gin.New()
		router.Use(delegation.Wrapper())
		router.Use(zoneGuard.Wrapper())
		router.GET("/api/admin/messages", func(ctx *gin.Context) { called = true; ctx.Status(http.StatusOK) })

		req := httptest.NewRequest(http.MethodGet, "/api/admin/messages", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code, "无委托头且无本地会话应被会话守卫拦截")
		assert.False(t, called)
	})
}

func TestAdminDelegationMiddleware_名称与顺序(t *testing.T) {
	t.Run("默认名称与认证槽位 Order", func(t *testing.T) {
		mw := NewAdminDelegationMiddleware(&AdminDelegationConfig{APIKeys: []string{"k"}})
		assert.Equal(t, "AdminDelegationMiddleware", mw.MiddlewareName())
		assert.Equal(t, OrderAdminDelegation, mw.Order())
		assert.Equal(t, 300, mw.Order(), "OrderAdminDelegation 应与 OrderAuth 同值同档")
	})
	t.Run("显式配置优先（双模式贴靠 54）", func(t *testing.T) {
		mw := newTestDelegation(func(cfg *AdminDelegationConfig) { cfg.Order = 54 })
		assert.Equal(t, 54, mw.Order())
	})
	t.Run("默认守卫前缀为 /api/admin", func(t *testing.T) {
		mw := NewAdminDelegationMiddleware(nil)
		w, called, _ := performDelegation(t, mw, http.MethodGet, "/api/admin/messages", map[string]string{})
		assert.Equal(t, http.StatusInternalServerError, w.Code, "无密钥配置时严格 fail-closed")
		assert.False(t, called)
	})
}
