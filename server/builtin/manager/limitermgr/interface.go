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

	// GetRemaining 获取剩余可访问次数
	// ctx: 上下文
	// key: 限流键
	// limit: 时间窗口内的最大请求数
	// window: 时间窗口大小
	// 返回: 剩余次数
	GetRemaining(ctx context.Context, key string, limit int, window time.Duration) (int, error)
}
