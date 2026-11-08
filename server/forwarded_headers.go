package server

import (
	"strings"

	"github.com/gin-gonic/gin"
)

// ClientHost 返回客户端语义的 Host（v1.2 二级链路适配，内部PRD）。
//
// v1.2 语义：Host 头只是网关路由键——外部流量经 A 网关（公网域名）转发到 B 网关时
// Host 会被改写为 mesh 域名（override_host），客户端真实访问的域名经
// X-Forwarded-Host 传递（A 边缘写入原值、B append 继承，链路上均为单值）。
// 取值规则：
//   - 优先 X-Forwarded-Host：多值时取最右合法段（最靠近应用的代理写入的值；
//     正常链路均为单值，多值处理仅作防御），单段清理空白并做合法性校验；
//   - 无该头或所有段均非法时回退 c.Request.Host。
//
// 信任前提：边缘网关对代理头删头重建，公网客户端无法伪造；私网直连对端
// （gate/BFF 容器、mesh 出口、本机）属 D3 信任域，与 X-Forwarded-For /
// DefaultTrustedProxyCIDRs 同一信任模型。需要"仅受信对端的头生效"的强保证时，
// 请依赖内置 client host restore 中间件（直接还原 c.Request.Host，应用零改动）。
func ClientHost(c *gin.Context) string {
	if c == nil || c.Request == nil {
		return ""
	}
	if host := cleanForwardedHost(c.Request.Header.Get("X-Forwarded-Host")); host != "" {
		return host
	}
	return c.Request.Host
}

// ClientProto 返回客户端语义的协议（"http" / "https"，v1.2 二级链路适配）。
//
// 优先 X-Forwarded-Proto：多值时取最右合法段，仅接受 http/https（大小写不敏感，
// 返回统一小写），其余取值视为非法直接跳过；无有效头时按 c.Request.TLS 推断
// （TLS 非 nil 即 https，否则 http）。信任前提同 ClientHost。
func ClientProto(c *gin.Context) string {
	if c != nil && c.Request != nil {
		for _, segment := range backwardSegments(c.Request.Header.Get("X-Forwarded-Proto")) {
			switch proto := strings.ToLower(segment); proto {
			case "http", "https":
				return proto
			}
		}
		if c.Request.TLS != nil {
			return "https"
		}
	}
	return "http"
}

// cleanForwardedHost 清理 X-Forwarded-Host 取值：按逗号切分后自右向左取第一个
// 合法段；全部非法时返回空串（由调用方回退 c.Request.Host）。
func cleanForwardedHost(value string) string {
	for _, host := range backwardSegments(value) {
		if validHostSegment(host) {
			return host
		}
	}
	return ""
}

// backwardSegments 将代理头取值按逗号切分、清理空白后自右向左返回
// （多值场景下最右段由最靠近应用的代理追加，最具代表性；正常链路均为单值）。
func backwardSegments(value string) []string {
	if value == "" {
		return nil
	}
	parts := strings.Split(value, ",")
	segments := make([]string, 0, len(parts))
	for i := len(parts) - 1; i >= 0; i-- {
		segments = append(segments, strings.TrimSpace(parts[i]))
	}
	return segments
}

// validHostSegment 校验代理头中单段 host 的合法性（host[:port] 或 [IPv6]:port 形态）：
// 非空、不含空白与 URI 语义字符（防注入），端口存在时必须为纯数字。
func validHostSegment(host string) bool {
	if host == "" || strings.ContainsAny(host, " \t\r\n\"/?#@\\") {
		return false
	}
	// IPv6 字面量：[addr] 或 [addr]:port
	if strings.HasPrefix(host, "[") {
		end := strings.IndexByte(host, ']')
		if end < 2 { // 缺 ']' 或地址为空
			return false
		}
		rest := host[end+1:]
		return rest == "" || (strings.HasPrefix(rest, ":") && isNumericPort(rest[1:]))
	}
	// reg-name / IPv4：至多一个 ':'，且端口部分为纯数字
	switch strings.Count(host, ":") {
	case 0:
		return true
	case 1:
		i := strings.IndexByte(host, ':')
		return isNumericPort(host[i+1:])
	default:
		return false
	}
}

// isNumericPort 校验端口段为纯数字（非空且全为 0-9，端口范围不在此强校验）
func isNumericPort(s string) bool {
	if s == "" {
		return false
	}
	for i := 0; i < len(s); i++ {
		if s[i] < '0' || s[i] > '9' {
			return false
		}
	}
	return true
}
