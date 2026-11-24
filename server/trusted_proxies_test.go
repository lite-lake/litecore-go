package server

import (
	"net"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

// newTrustPolicyEngine 构造已应用客户端 IP 信任策略的引擎，并注册回显 ClientIP 的路由。
// 返回引擎与用于捕获解析结果的指针。
func newTrustPolicyEngine(t *testing.T) (*gin.Engine, *string) {
	t.Helper()
	gin.SetMode(gin.TestMode)
	engine := gin.New()
	if err := applyClientIPTrustPolicy(engine); err != nil {
		t.Fatalf("应用客户端 IP 信任策略失败: %v", err)
	}
	got := new(string)
	engine.GET("/echo/ip", func(c *gin.Context) {
		*got = c.ClientIP()
		c.Status(http.StatusOK)
	})
	return engine, got
}

// TestApplyClientIPTrustPolicy_Config 验证策略应用后引擎的信任配置符合规范
func TestApplyClientIPTrustPolicy_Config(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine := gin.New()
	if err := applyClientIPTrustPolicy(engine); err != nil {
		t.Fatalf("应用客户端 IP 信任策略失败: %v", err)
	}

	if len(engine.RemoteIPHeaders) != 1 || engine.RemoteIPHeaders[0] != "X-Forwarded-For" {
		t.Errorf("期望仅保留 X-Forwarded-For 通道, 实际 %v", engine.RemoteIPHeaders)
	}

	// 内置网段均应可解析（SetTrustedProxies 内部已校验，此处兜底断言防常量被改坏）
	for _, cidr := range DefaultTrustedProxyCIDRs {
		if _, _, err := net.ParseCIDR(cidr); err != nil {
			t.Errorf("内置可信网段 %q 非法: %v", cidr, err)
		}
	}
}

// TestApplyClientIPTrustPolicy_ClientIP 验证框架信任策略下 ClientIP 的解析行为：
// 私网直连解析 XFF、公网直连回退 RemoteAddr、X-Real-IP 不再被采信（内部PRD §7.1）。
func TestApplyClientIPTrustPolicy_ClientIP(t *testing.T) {
	cases := []struct {
		name       string
		remoteAddr string
		headers    map[string]string
		want       string
	}{
		{
			name:       "私网直连_解析XFF单值",
			remoteAddr: "10.10.0.5:1234",
			headers:    map[string]string{"X-Forwarded-For": "1.2.3.4"},
			want:       "1.2.3.4",
		},
		{
			name:       "私网直连_多级XFF右至左跳过受信私网",
			remoteAddr: "10.10.0.5:1234",
			headers:    map[string]string{"X-Forwarded-For": "1.2.3.4, 10.0.0.1"},
			want:       "1.2.3.4",
		},
		{
			name:       "回环直连_本地调试解析XFF",
			remoteAddr: "127.0.0.1:5555",
			headers:    map[string]string{"X-Forwarded-For": "203.0.113.7"},
			want:       "203.0.113.7",
		},
		{
			name:       "公网直连_伪造XFF被忽略回退直连对端",
			remoteAddr: "8.8.8.8:999",
			headers:    map[string]string{"X-Forwarded-For": "1.2.3.4"},
			want:       "8.8.8.8",
		},
		{
			name:       "X-Real-IP不再被采信_返回直连IP",
			remoteAddr: "10.10.0.5:1234",
			headers:    map[string]string{"X-Real-IP": "5.6.7.8"},
			want:       "10.10.0.5",
		},
		{
			name:       "无私网直连且无XFF_返回直连IP",
			remoteAddr: "192.168.1.9:4000",
			headers:    nil,
			want:       "192.168.1.9",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			engine, got := newTrustPolicyEngine(t)

			req := httptest.NewRequest(http.MethodGet, "/echo/ip", nil)
			req.RemoteAddr = tc.remoteAddr
			for k, v := range tc.headers {
				req.Header.Set(k, v)
			}

			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			if *got != tc.want {
				t.Errorf("期望 ClientIP %q, 实际 %q", tc.want, *got)
			}
		})
	}
}
