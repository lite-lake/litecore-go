package liteadmin

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
	"time"
)

// w3cTraceIDPattern W3C trace-id 格式：32 位小写 hex（见《SOP-gate-trace-id》§3）
var w3cTraceIDPattern = regexp.MustCompile(`^[0-9a-f]{32}$`)

// DefaultTimeout 默认上游调用超时（PRD §6.3：默认 15s）
const DefaultTimeout = 15 * time.Second

// UpstreamClient AC 侧上游服务客户端：构造并发起对上游 /api/admin/** 的委托调用。
//
// 职责边界（PRD §4.1 信任链第 3 段）：注入委托头与 trace 头；入站的
// X-Lite-Admin-* / Authorization / Cookie 剥离不属于本包（透传引擎在改写请求时处理）。
type UpstreamClient struct {
	// BaseURL 上游服务内网基址（如 https://app-demo.node2.v-prod.mesh.example.net，末尾不带 /）
	BaseURL string
	// AdminKey 该上游服务的委托密钥（X-Lite-Admin-Key 的值，与上游 admin_delegation.api_keys 配对）
	AdminKey string
	// HTTPClient 发起调用的 HTTP 客户端；为 nil 时使用带 DefaultTimeout 超时的默认客户端
	HTTPClient *http.Client
}

// NewUpstreamClient 创建上游委托调用客户端
func NewUpstreamClient(baseURL, adminKey string) *UpstreamClient {
	return &UpstreamClient{
		BaseURL:    strings.TrimRight(baseURL, "/"),
		AdminKey:   adminKey,
		HTTPClient: &http.Client{Timeout: DefaultTimeout},
	}
}

// NewRequest 构造发往上游 /api/admin/** 的委托调用请求（不发送，便于调用方定制）。
//
// 参数：
//   - method / adminPath：HTTP 方法与上游管理路径（adminPath 以 /api/admin 开头）；
//   - identity：委托管理员身份（X-Lite-Admin-Id / X-Lite-Admin-Username 的来源）；
//   - traceID：链路追踪 ID（X-Lite-Gate-Trace-Id）；为空则自动生成 32 位小写 hex；
//   - body：请求体（可为 nil）。
//
// trace 语义依《SOP-gate-trace-id》：ID 为 W3C 32 位小写 hex 时同时携带 traceparent
// （同一 ID 空间，parent-id 每次新生成）；其余格式只传播 X-Lite-Gate-Trace-Id，
// 宁缺勿假，不造合法外形的假 W3C 头。
func (c *UpstreamClient) NewRequest(ctx context.Context, method, adminPath string, identity DelegatedIdentity, traceID string, body io.Reader) (*http.Request, error) {
	if !strings.HasPrefix(adminPath, "/api/admin") {
		return nil, fmt.Errorf("上游路径必须以 /api/admin 开头，收到 %q", adminPath)
	}
	if c.AdminKey == "" {
		return nil, fmt.Errorf("未配置委托密钥（AdminKey），拒绝构造委托请求")
	}

	req, err := http.NewRequestWithContext(ctx, method, c.BaseURL+adminPath, body)
	if err != nil {
		return nil, fmt.Errorf("构造上游请求失败: %w", err)
	}

	// 委托头（身份断言只能出自 AC，X-Lite-Admin-Username 仅透写日志，不参与上游校验）
	req.Header.Set(HeaderAdminKey, c.AdminKey)
	req.Header.Set(HeaderAdminID, identity.SubjectID)
	if identity.Username != "" {
		req.Header.Set(HeaderAdminUsername, identity.Username)
	}

	// trace 头：无则生成，W3C 格式同时携带 traceparent
	traceID = resolveTraceID(traceID)
	req.Header.Set(HeaderGateTraceID, traceID)
	if tp := buildTraceparent(traceID); tp != "" {
		req.Header.Set(HeaderTraceparent, tp)
	}

	return req, nil
}

// Do 构造并发起委托调用（NewRequest + HTTPClient.Do 的便捷组合）
func (c *UpstreamClient) Do(ctx context.Context, method, adminPath string, identity DelegatedIdentity, traceID string, body io.Reader) (*http.Response, error) {
	req, err := c.NewRequest(ctx, method, adminPath, identity, traceID, body)
	if err != nil {
		return nil, err
	}
	client := c.HTTPClient
	if client == nil {
		client = &http.Client{Timeout: DefaultTimeout}
	}
	return client.Do(req)
}

// resolveTraceID 规范化追踪 ID：为空时生成 32 位小写 hex（crypto/rand，全零重试）；
// 非空时原样保留（旧格式 `[A-Za-z0-9]{1,64}` 依《SOP-gate-trace-id》兼容传播，
// 仅影响是否携带 traceparent，不改写调用方的链路 ID）
func resolveTraceID(traceID string) string {
	traceID = strings.TrimSpace(traceID)
	if traceID != "" {
		return traceID
	}
	for {
		buf := make([]byte, 16)
		if _, err := rand.Read(buf); err != nil {
			// crypto/rand 失败极罕见（系统熵源故障）；不 panic，退化为全零让下游按非法处理
			return strings.Repeat("0", 32)
		}
		id := hex.EncodeToString(buf)
		if id != strings.Repeat("0", 32) {
			return id
		}
	}
}

// buildTraceparent 由 32 位小写 hex 的 trace-id 组装 W3C traceparent
// （version=00，parent-id 每跳新生成 16 hex，flags=01）；非 W3C 格式返回空串
func buildTraceparent(traceID string) string {
	if !w3cTraceIDPattern.MatchString(traceID) || traceID == strings.Repeat("0", 32) {
		return ""
	}
	buf := make([]byte, 8)
	if _, err := rand.Read(buf); err != nil {
		return ""
	}
	return "00-" + traceID + "-" + hex.EncodeToString(buf) + "-01"
}
