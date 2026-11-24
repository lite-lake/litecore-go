package server

import (
	"crypto/tls"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

// newForwardedContext 构造带指定 Host / TLS / 代理头的测试上下文
func newForwardedContext(t *testing.T, host string, withTLS bool, headers map[string]string) *gin.Context {
	t.Helper()
	gin.SetMode(gin.TestMode)

	req := httptest.NewRequest(http.MethodGet, "/echo", nil)
	req.Host = host
	if withTLS {
		req.TLS = &tls.ConnectionState{}
	}
	for k, v := range headers {
		req.Header.Set(k, v)
	}

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	return c
}

// TestClientHost 验证 ClientHost 的取值规则：
// 优先 X-Forwarded-Host（清理空白、多值取最右合法段、非法段跳过），回退 c.Request.Host。
func TestClientHost(t *testing.T) {
	cases := []struct {
		name      string
		host      string
		forwarded string
		want      string
	}{
		{name: "无代理头_回退RequestHost", host: "api.example.net", want: "api.example.net"},
		{name: "单值合法_优先代理头", host: "app-demo-api.node1.example.net", forwarded: "api.example.net", want: "api.example.net"},
		{name: "单值清理空白", host: "mesh.internal", forwarded: "  api.example.net  ", want: "api.example.net"},
		{name: "带端口原样保留", host: "mesh.internal", forwarded: "api.example.net:8443", want: "api.example.net:8443"},
		{name: "多值取最右合法段", host: "mesh.internal", forwarded: "internal.a.example.net, api.example.net", want: "api.example.net"},
		{name: "最右段非法_向左取合法段", host: "mesh.internal", forwarded: "api.example.net, evil host", want: "api.example.net"},
		{name: "最右段非法_URI字符注入", host: "api.example.net", forwarded: "api.example.net, evil.example/a?b=1", want: "api.example.net"},
		{name: "全部段非法_回退RequestHost", host: "api.example.net", forwarded: "a b, c/d", want: "api.example.net"},
		{name: "空串头_回退RequestHost", host: "api.example.net", forwarded: "", want: "api.example.net"},
		{name: "空段跳过_取合法段", host: "mesh.internal", forwarded: "  , api.example.net", want: "api.example.net"},
		{name: "非法端口_回退RequestHost", host: "api.example.net", forwarded: "api.example.net:84a8", want: "api.example.net"},
		{name: "端口为空_回退RequestHost", host: "api.example.net", forwarded: "api.example.net:", want: "api.example.net"},
		{name: "多冒号非法_回退RequestHost", host: "api.example.net", forwarded: "host:80:90", want: "api.example.net"},
		{name: "IPv6字面量带端口合法", host: "mesh.internal", forwarded: "[2001:db8::1]:8443", want: "[2001:db8::1]:8443"},
		{name: "IPv6字面量无端口合法", host: "mesh.internal", forwarded: "[2001:db8::1]", want: "[2001:db8::1]"},
		{name: "IPv6字面量缺右括号非法", host: "api.example.net", forwarded: "[2001:db8::1", want: "api.example.net"},
		{name: "IPv6字面量非法端口", host: "api.example.net", forwarded: "[2001:db8::1]:x", want: "api.example.net"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			headers := map[string]string{}
			if tc.forwarded != "" {
				headers["X-Forwarded-Host"] = tc.forwarded
			}
			c := newForwardedContext(t, tc.host, false, headers)
			if got := ClientHost(c); got != tc.want {
				t.Errorf("期望 ClientHost %q, 实际 %q", tc.want, got)
			}
		})
	}
}

// TestClientHost_NilRequest 验证请求缺省时的 fail-safe 行为
func TestClientHost_NilRequest(t *testing.T) {
	c := &gin.Context{}
	if got := ClientHost(c); got != "" {
		t.Errorf("期望空串, 实际 %q", got)
	}
	if got := ClientHost(nil); got != "" {
		t.Errorf("期望空串, 实际 %q", got)
	}
}

// TestClientProto 验证 ClientProto 的取值规则：
// 优先 X-Forwarded-Proto（仅 http/https，多值取最右合法段），回退按 TLS 推断。
func TestClientProto(t *testing.T) {
	cases := []struct {
		name      string
		forwarded string
		withTLS   bool
		want      string
	}{
		{name: "无头无TLS_推断http", want: "http"},
		{name: "无头有TLS_推断https", withTLS: true, want: "https"},
		{name: "头https_优先", forwarded: "https", want: "https"},
		{name: "头大写HTTPS_归一化小写", forwarded: "HTTPS", want: "https"},
		{name: "头http_覆盖TLS推断", forwarded: "http", withTLS: true, want: "http"},
		{name: "非法值ftp_回退TLS推断", forwarded: "ftp", withTLS: true, want: "https"},
		{name: "非法值ftp无TLS_回退http", forwarded: "ftp", want: "http"},
		{name: "多值取最右合法段", forwarded: "https, http", want: "http"},
		{name: "最右非法_向左取合法段", forwarded: "https, ftp", want: "https"},
		{name: "空串头_回退TLS推断", forwarded: "", withTLS: true, want: "https"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			headers := map[string]string{}
			if tc.forwarded != "" {
				headers["X-Forwarded-Proto"] = tc.forwarded
			}
			c := newForwardedContext(t, "api.example.net", tc.withTLS, headers)
			if got := ClientProto(c); got != tc.want {
				t.Errorf("期望 ClientProto %q, 实际 %q", tc.want, got)
			}
		})
	}
}

// TestClientProto_NilRequest 验证请求缺省时的 fail-safe 行为
func TestClientProto_NilRequest(t *testing.T) {
	c := &gin.Context{}
	if got := ClientProto(c); got != "http" {
		t.Errorf("期望 \"http\", 实际 %q", got)
	}
	if got := ClientProto(nil); got != "http" {
		t.Errorf("期望 \"http\", 实际 %q", got)
	}
}
