package limitermgr

import (
	"context"
	"time"

	"github.com/lite-lake/litecore-go/common"
)

// ILimiterManager 限流管理器接口
type ILimiterManager interface {
	common.IBaseManager

	// Allow 检查是否允许通过限流
	// ctx: 上下文
	// key: 限流键（如用户ID、IP等）
	// limit: 时间窗口内的最大请求数
	// window: 时间窗口大小
	// 返回: 允许返回 true，否则返回 false
	Allow(ctx context.Context, key string, limit int, window time.Duration) (bool, error)

	// AllowN 带成本检查是否允许通过限流（内部PRD：一次调用消耗 cost 个
	// 配额，网关 ingest 批量上报按记录计数——cost=批内条数）。语义（双实现一致）：
	// 窗口内已用量 + cost ≤ limit → 允许并消耗 cost；否则整批拒绝且不消耗；
	// cost=0（空批）恒允许且不消耗。
	// ctx: 上下文
	// key: 限流键（如用户ID、IP等）
	// limit: 时间窗口内的最大配额数
	// window: 时间窗口大小
	// cost: 本次调用消耗的配额数（0=空批不消耗）
	// 返回: 允许返回 true，否则返回 false
	AllowN(ctx context.Context, key string, limit int, window time.Duration, cost int) (bool, error)

	// GetRemaining 获取剩余可访问次数
	// ctx: 上下文
	// key: 限流键
	// limit: 时间窗口内的最大请求数
	// window: 时间窗口大小
	// 返回: 剩余次数
	GetRemaining(ctx context.Context, key string, limit int, window time.Duration) (int, error)
}
