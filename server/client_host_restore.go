package server

import (
	"net"
	"strings"
	"sync"

	"github.com/gin-gonic/gin"

	"github.com/lite-lake/litecore-go/common"
)

// orderClientHostRestore client host restore 中间件执行顺序：
// 位于 system 中间件区（0-49）——recovery（0）之后、request-logger（50）之前，
// 早于 CORS / authz 与全部应用中间件，保证应用读到的 c.Request.Host 已完成还原。
const orderClientHostRestore = 10

// systemMiddleware 内置 system 中间件标记接口：
// 由 Engine 直接注册（不经应用 Middleware 容器），注册日志 type 标记为 "system"。
type systemMiddleware interface {
	isSystemMiddleware()
}

// clientHostRestoreMiddleware 客户端 Host 还原中间件（v1.2 二级链路适配，内部PRD）。
//
// v1.2 语义：外部流量经 A 网关（公网域名 host）转发到 B 网关时 Host 头被改写为
// mesh 域名（override_host），客户端真实域名经 X-Forwarded-Host 传递。二级链路下
// 应用收到的 Host 将是 mesh 域名，导致按 Host 语义工作的逻辑（如 sc-api 的
// AWS SigV4 签名包含 Host）校验失败。本中间件把 c.Request.Host 还原为
// X-Forwarded-Host 的清理值，使存量应用零改动保持正确。
//
// 信任模型（与 X-Forwarded-For / DefaultTrustedProxyCIDRs 一致，D3 信任域）：
//   - 仅当直连对端 ∈ DefaultTrustedProxyCIDRs（gate/BFF 容器、mesh 出口、本机）时
//     才采信 X-Forwarded-Host；不受信（公网直连）对端携带的该头视为伪造，直接忽略；
//   - 边缘网关对代理头删头重建，公网客户端无法借 A 网关注入自定义取值。
//
// 行为影响：单级链路（Host 本就等于公网域名 = X-Forwarded-Host）为 no-op；
// BFF / 二级链路下 Host 由内部名还原为公网域名（全仓审计确认无应用依赖内部 Host，
// 属预期统一）。
type clientHostRestoreMiddleware struct{}

// NewClientHostRestoreMiddleware 创建客户端 Host 还原中间件
// （Engine 已在 system 区内置注册，应用通常无需自行挂载；导出仅供测试与独立引擎复用）。
func NewClientHostRestoreMiddleware() common.IBaseMiddleware {
	return &clientHostRestoreMiddleware{}
}

func (m *clientHostRestoreMiddleware) isSystemMiddleware() {}

// MiddlewareName 返回中间件名称
func (m *clientHostRestoreMiddleware) MiddlewareName() string {
	return "ClientHostRestoreMiddleware"
}

// Order 返回执行顺序（system 区：recovery 之后、request-logger 之前）
func (m *clientHostRestoreMiddleware) Order() int {
	return orderClientHostRestore
}

// Wrapper 返回 Gin 中间件函数：受信直连对端携带 X-Forwarded-Host 时
// 将 c.Request.Host 还原为清理值；其余情况不做任何修改。
func (m *clientHostRestoreMiddleware) Wrapper() gin.HandlerFunc {
	return func(c *gin.Context) {
		if isTrustedPeer(c) {
			if host := cleanForwardedHost(c.Request.Header.Get("X-Forwarded-Host")); host != "" {
				c.Request.Host = host
			}
		}
		c.Next()
	}
}

// OnStart 服务器启动时触发（无外部资源，直接返回）
func (m *clientHostRestoreMiddleware) OnStart() error { return nil }

// OnStop 服务器停止时触发（无外部资源，直接返回）
func (m *clientHostRestoreMiddleware) OnStop() error { return nil }

// trustedProxyNets 预解析的受信代理网段：对齐 DefaultTrustedProxyCIDRs，
// 进程内仅解析一次（sync.Once），避免逐请求 ParseCIDR。
var (
	trustedProxyNetsOnce sync.Once
	trustedProxyNets     []*net.IPNet
)

// parseTrustedProxyNets 惰性解析并返回受信代理网段。
// DefaultTrustedProxyCIDRs 常量被改坏时（applyClientIPTrustPolicy 启动期已对同一
// 常量 fail-fast 校验，此处兜底）返回 nil，还原功能 fail-safe 关闭。
func parseTrustedProxyNets() []*net.IPNet {
	trustedProxyNetsOnce.Do(func() {
		nets := make([]*net.IPNet, 0, len(DefaultTrustedProxyCIDRs))
		for _, cidr := range DefaultTrustedProxyCIDRs {
			_, network, err := net.ParseCIDR(cidr)
			if err != nil {
				trustedProxyNets = nil
				return
			}
			nets = append(nets, network)
		}
		trustedProxyNets = nets
	})
	return trustedProxyNets
}

// directPeerIP 提取直连对端 IP（仅解析 RemoteAddr，不受任何代理头影响）
func directPeerIP(c *gin.Context) net.IP {
	if c.Request == nil {
		return nil
	}
	addr := strings.TrimSpace(c.Request.RemoteAddr)
	if addr == "" {
		return nil
	}
	// RemoteAddr 正常为 host:port；容忍无端口的裸 IP 写法
	if host, _, err := net.SplitHostPort(addr); err == nil {
		return net.ParseIP(host)
	}
	return net.ParseIP(addr)
}

// isTrustedPeer 判断直连对端是否属于受信代理网段（DefaultTrustedProxyCIDRs）
func isTrustedPeer(c *gin.Context) bool {
	ip := directPeerIP(c)
	if ip == nil {
		return false
	}
	for _, network := range parseTrustedProxyNets() {
		if network.Contains(ip) {
			return true
		}
	}
	return false
}
