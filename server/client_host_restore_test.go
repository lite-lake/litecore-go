package server

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

// newHostRestoreEngine 构造已挂载 client host restore 中间件的引擎，
// 并注册回显 c.Request.Host 的路由。返回引擎与用于捕获结果的指针。
func newHostRestoreEngine(t *testing.T) (*gin.Engine, *string) {
	t.Helper()
	gin.SetMode(gin.TestMode)
	engine := gin.New()
	engine.Use(NewClientHostRestoreMiddleware().Wrapper())
	got := new(string)
	engine.GET("/echo/host", func(c *gin.Context) {
		*got = c.Request.Host
		c.Status(http.StatusOK)
	})
	return engine, got
}

// serveHostRestoreRequest 按指定直连对端 / Host / X-Forwarded-Host 发起请求
func serveHostRestoreRequest(t *testing.T, engine *gin.Engine, remoteAddr, host, forwardedHost string) {
	t.Helper()
	req := httptest.NewRequest(http.MethodGet, "/echo/host", nil)
	req.RemoteAddr = remoteAddr
	req.Host = host
	if forwardedHost != "" {
		req.Header.Set("X-Forwarded-Host", forwardedHost)
	}
	w := httptest.NewRecorder()
	engine.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("期望 200, 实际 %d", w.Code)
	}
}

// TestClientHostRestoreMiddleware_Metadata 验证中间件元数据：
// 名称正确且 Order 位于 system 区（0-49），并实现 systemMiddleware 标记。
func TestClientHostRestoreMiddleware_Metadata(t *testing.T) {
	mw := NewClientHostRestoreMiddleware()
	if mw.MiddlewareName() != "ClientHostRestoreMiddleware" {
		t.Errorf("期望名称 ClientHostRestoreMiddleware, 实际 %q", mw.MiddlewareName())
	}
	if order := mw.Order(); order < 0 || order >= 50 {
		t.Errorf("期望 Order 位于 system 区 [0,49], 实际 %d", order)
	}
	if _, ok := mw.(systemMiddleware); !ok {
		t.Error("期望实现 systemMiddleware 标记接口")
	}
}

// TestClientHostRestoreMiddleware_TrustedPeerRestore 验证受信对端的 Host 还原行为：
// 直连对端 ∈ DefaultTrustedProxyCIDRs 且请求携带 X-Forwarded-Host 时，
// c.Request.Host 被还原为清理值（v1.2 二级链路：mesh 域名 → 客户端公网域名）。
func TestClientHostRestoreMiddleware_TrustedPeerRestore(t *testing.T) {
	cases := []struct {
		name          string
		remoteAddr    string
		forwardedHost string
		want          string
	}{
		{name: "mesh网段对端还原", remoteAddr: "10.10.0.5:443", forwardedHost: "api.example.net", want: "api.example.net"},
		{name: "docker网段对端还原", remoteAddr: "172.17.0.5:8", forwardedHost: "api.example.net", want: "api.example.net"},
		{name: "办公网段对端还原", remoteAddr: "192.168.1.9:4000", forwardedHost: "api.example.net", want: "api.example.net"},
		{name: "IPv6回环对端还原", remoteAddr: "[::1]:5555", forwardedHost: "api.example.net", want: "api.example.net"},
		{name: "IPv6 ULA对端还原", remoteAddr: "[fd00::5]:5555", forwardedHost: "api.example.net", want: "api.example.net"},
		{name: "多值取最右合法段并清理空白", remoteAddr: "10.10.0.5:443", forwardedHost: " 10.0.0.1:1 , api.example.net ", want: "api.example.net"},
		{name: "带端口原样还原", remoteAddr: "10.10.0.5:443", forwardedHost: "api.example.net:8443", want: "api.example.net:8443"},
		{name: "单级链路语义_Host等于代理头_no-op", remoteAddr: "10.10.0.5:443", forwardedHost: "api.example.net", want: "api.example.net"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			engine, got := newHostRestoreEngine(t)
			serveHostRestoreRequest(t, engine, tc.remoteAddr, "app-demo-api.node1.example.net", tc.forwardedHost)
			if *got != tc.want {
				t.Errorf("期望还原后 Host %q, 实际 %q", tc.want, *got)
			}
		})
	}
}

// TestClientHostRestoreMiddleware_UntrustedPeerIgnore 验证不受信对端的伪造防护：
// 公网直连对端携带 X-Forwarded-Host 不被采信，c.Request.Host 保持不变
// （与 X-Forwarded-For 同信任模型，防公网伪造）。
func TestClientHostRestoreMiddleware_UntrustedPeerIgnore(t *testing.T) {
	cases := []struct {
		name       string
		remoteAddr string
	}{
		{name: "公网直连对端", remoteAddr: "8.8.8.8:999"},
		{name: "公网IPv6直连对端", remoteAddr: "[2001:db8::9]:999"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			engine, got := newHostRestoreEngine(t)
			serveHostRestoreRequest(t, engine, tc.remoteAddr, "app-demo-api.node1.example.net", "evil.example.com")
			if *got != "app-demo-api.node1.example.net" {
				t.Errorf("不受信对端的 X-Forwarded-Host 应被忽略, 期望 Host 保持 %q, 实际 %q",
					"app-demo-api.node1.example.net", *got)
			}
		})
	}
}

// TestClientHostRestoreMiddleware_NoHeaderNoOp 验证无 X-Forwarded-Host / 头值非法时 no-op：
// 受信对端但请求不带有效代理头时，c.Request.Host 原样保留。
func TestClientHostRestoreMiddleware_NoHeaderNoOp(t *testing.T) {
	cases := []struct {
		name          string
		forwardedHost string
	}{
		{name: "无X-Forwarded-Host头", forwardedHost: ""},
		{name: "头值全空白", forwardedHost: "   "},
		{name: "头值非法host段", forwardedHost: "bad host/evil"},
		{name: "头值空串", forwardedHost: "\t"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			engine, got := newHostRestoreEngine(t)
			serveHostRestoreRequest(t, engine, "10.10.0.5:443", "app-demo-api.node1.example.net", tc.forwardedHost)
			if *got != "app-demo-api.node1.example.net" {
				t.Errorf("无效代理头不应改动 Host, 期望 %q, 实际 %q",
					"app-demo-api.node1.example.net", *got)
			}
		})
	}
}
