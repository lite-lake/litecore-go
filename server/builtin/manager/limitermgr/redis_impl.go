package limitermgr

import (
	"context"
	"fmt"
	"time"

	"github.com/lite-lake/litecore-go/server/builtin/manager/cachemgr"
)

// limiterManagerRedisImpl Redis 限流管理器实现
// 使用 Redis 存储限流状态，支持分布式限流
type limiterManagerRedisImpl struct {
	*limiterManagerBaseImpl                        // 基类，提供可观测性功能
	CacheMgr                cachemgr.ICacheManager `inject:""` // 缓存管理器，用于 Redis 操作
	name                    string                 // 管理器名称
}

const limiterKeyPrefix = "limiter:" // Redis 限流键前缀

// NewLimiterManagerRedisImpl 创建 Redis 限流管理器实例
func NewLimiterManagerRedisImpl() ILimiterManager {
	impl := &limiterManagerRedisImpl{
		limiterManagerBaseImpl: newILimiterManagerBaseImpl(),
		name:                   "limiterManagerRedisImpl",
	}
	impl.initObservability()
	return impl
}

// ManagerName 返回管理器名称
func (r *limiterManagerRedisImpl) ManagerName() string {
	return r.name
}

// Health 检查管理器健康状态
// 检查缓存管理器是否已初始化及连接是否正常
func (r *limiterManagerRedisImpl) Health() error {
	if r.CacheMgr == nil {
		return fmt.Errorf("cache manager is not initialized")
	}
	return r.CacheMgr.Health()
}

// OnStart 启动管理器时的回调
// Redis 限流管理器无需额外初始化，返回 nil
func (r *limiterManagerRedisImpl) OnStart() error {
	return nil
}

// OnStop 停止管理器时的回调
// Redis 限流管理器无需清理资源，返回 nil
func (r *limiterManagerRedisImpl) OnStop() error {
	return nil
}

// Allow 检查是否允许通过限流
// 使用 Redis 计数器实现固定窗口限流
// 参数：
//   - ctx: 上下文
//   - key: 限流键，标识限流对象
//   - limit: 时间窗口内的最大请求数
//   - window: 时间窗口大小
//
// 返回: 允许返回 true，否则返回 false
func (r *limiterManagerRedisImpl) Allow(ctx context.Context, key string, limit int, window time.Duration) (bool, error) {
	var result bool

	err := r.recordOperation(ctx, "redis", "allow", key, func() error {
		if err := ValidateContext(ctx); err != nil {
			return err
		}
		if err := ValidateKey(key); err != nil {
			return err
		}
		if err := ValidateLimit(limit); err != nil {
			return err
		}
		if err := ValidateWindow(window); err != nil {
			return err
		}

		if r.CacheMgr == nil {
			return fmt.Errorf("cache manager is not initialized")
		}

		cacheKey := limiterKeyPrefix + key

		count, err := r.CacheMgr.Increment(ctx, cacheKey, 1)
		if err != nil {
			return fmt.Errorf("failed to increment counter: %w", err)
		}

		if count == 1 {
			if err := r.CacheMgr.Expire(ctx, cacheKey, window); err != nil {
				return fmt.Errorf("failed to set expiration: %w", err)
			}
		}

		allowed := int(count) <= limit
		result = allowed
		r.recordAllowance(ctx, "redis", allowed)

		return nil
	})

	return result, err
}

// GetRemaining 获取剩余可访问次数
// 从 Redis 中查询当前计数值，计算剩余次数
// 参数：
//   - ctx: 上下文
//   - key: 限流键
//   - limit: 时间窗口内的最大请求数
//   - window: 时间窗口大小
//
// 返回: 剩余次数
func (r *limiterManagerRedisImpl) GetRemaining(ctx context.Context, key string, limit int, window time.Duration) (int, error) {
	var result int

	err := r.recordOperation(ctx, "redis", "get_remaining", key, func() error {
		if err := ValidateContext(ctx); err != nil {
			return err
		}
		if err := ValidateKey(key); err != nil {
			return err
		}
		if err := ValidateLimit(limit); err != nil {
			return err
		}
		if err := ValidateWindow(window); err != nil {
			return err
		}

		if r.CacheMgr == nil {
			return fmt.Errorf("cache manager is not initialized")
		}

		cacheKey := limiterKeyPrefix + key

		var count int64
		err := r.CacheMgr.Get(ctx, cacheKey, &count)
		if err != nil {
			result = limit
			return nil
		}

		remaining := limit - int(count)
		if remaining < 0 {
			remaining = 0
		}
		result = remaining

		return nil
	})

	return result, err
}

var _ ILimiterManager = (*limiterManagerRedisImpl)(nil)
