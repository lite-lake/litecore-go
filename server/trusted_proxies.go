package server

import (
	"fmt"

	"github.com/gin-gonic/gin"
)

// DefaultTrustedProxyCIDRs 内置可信代理网段：覆盖 docker 网络、EasyTier mesh 与本机回环。
// 信任模型（设计决策 D3，应用零配置）：外部流量只可能经 infra-gate 进入，
// 所有合法直连对端（gate/BFF 容器、mesh 出口、本机）必属私网；公网直连不受信，
// ClientIP() fail-safe 回退为直连对端真实 IP（配置失误时即攻击者自身 IP，无法借伪造头隐匿）。
var DefaultTrustedProxyCIDRs = []string{
	"127.0.0.0/8",    // 本机回环（本地调试 curl 直连）
	"::1/128",        // IPv6 回环
	"10.0.0.0/8",     // 私网 mesh 与内网容器网络
	"172.16.0.0/12",  // docker 默认地址池
	"192.168.0.0/16", // docker 自定义网络 / 办公网
	"fc00::/7",       // IPv6 ULA
}

// applyClientIPTrustPolicy 将框架内置的客户端真实 IP 信任策略应用到 gin 引擎
// （设计决策 D3 / §7.1）：
//  1. 仅信任 DefaultTrustedProxyCIDRs 内的直连对端：受信时按右→左解析 X-Forwarded-For
//     单值，不受信时回退直连对端 IP（与 infra-gate netutil.ResolveClientIP 语义对齐）；
//  2. 显式剔除 gin 默认的 X-Real-IP 通道：X-Real-IP 全链路禁用（设计记录 §4.1）。
func applyClientIPTrustPolicy(engine *gin.Engine) error {
	if err := engine.SetTrustedProxies(DefaultTrustedProxyCIDRs); err != nil {
		return fmt.Errorf("设置可信代理网段失败: %w", err)
	}
	engine.RemoteIPHeaders = []string{"X-Forwarded-For"}
	return nil
}
