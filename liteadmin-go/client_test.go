package liteadmin

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const testTraceID = "4bf92f3577b34da6a3ce929d0e0e4736"

func TestUpstreamClient_NewRequest_头注入(t *testing.T) {
	t.Run("完整身份与 W3C trace 同时携带 traceparent", func(t *testing.T) {
		client := NewUpstreamClient("https://app-demo.node2.v-prod.mesh.example.net/", "test-key")
		req, err := client.NewRequest(context.Background(), http.MethodGet, "/api/admin/messages", NewDelegatedSession("42", "kent"), testTraceID, nil)
		require.NoError(t, err)

		assert.Equal(t, "https://app-demo.node2.v-prod.mesh.example.net/api/admin/messages", req.URL.String())
		assert.Equal(t, "test-key", req.Header.Get(HeaderAdminKey))
		assert.Equal(t, "42", req.Header.Get(HeaderAdminID))
		assert.Equal(t, "kent", req.Header.Get(HeaderAdminUsername))
		assert.Equal(t, testTraceID, req.Header.Get(HeaderGateTraceID))

		traceparent := req.Header.Get(HeaderTraceparent)
		assert.Regexp(t, regexp.MustCompile(`^00-`+testTraceID+`-[0-9a-f]{16}-01$`), traceparent)
	})

	t.Run("用户名为空时不携带该头", func(t *testing.T) {
		client := NewUpstreamClient("https://upstream.example.com", "k")
		req, err := client.NewRequest(context.Background(), http.MethodGet, "/api/admin/overview", NewDelegatedSession("1", ""), testTraceID, nil)
		require.NoError(t, err)
		assert.Empty(t, req.Header.Get(HeaderAdminUsername))
	})

	t.Run("非 W3C 格式 trace 只传播 GateTraceID 不造 traceparent", func(t *testing.T) {
		client := NewUpstreamClient("https://upstream.example.com", "k")
		req, err := client.NewRequest(context.Background(), http.MethodGet, "/api/admin/overview", NewDelegatedSession("1", ""), "legacy-trace-id", nil)
		require.NoError(t, err)
		assert.Equal(t, "legacy-trace-id", req.Header.Get(HeaderGateTraceID))
		assert.Empty(t, req.Header.Get(HeaderTraceparent))
	})

	t.Run("trace 为空时自动生成 32 位小写 hex", func(t *testing.T) {
		client := NewUpstreamClient("https://upstream.example.com", "k")
		req, err := client.NewRequest(context.Background(), http.MethodGet, "/api/admin/overview", NewDelegatedSession("1", ""), "", nil)
		require.NoError(t, err)
		got := req.Header.Get(HeaderGateTraceID)
		assert.Regexp(t, regexp.MustCompile(`^[0-9a-f]{32}$`), got)
		assert.NotEqual(t, strings.Repeat("0", 32), got)
		assert.NotEmpty(t, req.Header.Get(HeaderTraceparent))
	})

	t.Run("路径非 admin 分区拒绝", func(t *testing.T) {
		client := NewUpstreamClient("https://upstream.example.com", "k")
		_, err := client.NewRequest(context.Background(), http.MethodGet, "/api/user/orders", NewDelegatedSession("1", ""), testTraceID, nil)
		assert.ErrorContains(t, err, "/api/admin")
	})

	t.Run("未配置密钥 fail-closed", func(t *testing.T) {
		client := NewUpstreamClient("https://upstream.example.com", "")
		_, err := client.NewRequest(context.Background(), http.MethodGet, "/api/admin/overview", NewDelegatedSession("1", ""), testTraceID, nil)
		assert.ErrorContains(t, err, "委托密钥")
	})
}

func TestUpstreamClient_Do(t *testing.T) {
	var got *http.Request
	var gotBody string
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got = r
		b, _ := io.ReadAll(r.Body)
		gotBody = string(b)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"code":200}`))
	}))
	defer upstream.Close()

	client := NewUpstreamClient(upstream.URL, "roundtrip-key")
	resp, err := client.Do(context.Background(), http.MethodPost, "/api/admin/messages", NewDelegatedSession("7", "kent"), testTraceID, strings.NewReader(`{"title":"hi"}`))
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, http.MethodPost, got.Method)
	assert.Equal(t, `{"title":"hi"}`, gotBody)
	assert.Equal(t, "roundtrip-key", got.Header.Get(HeaderAdminKey))
	assert.Equal(t, "7", got.Header.Get(HeaderAdminID))
}
