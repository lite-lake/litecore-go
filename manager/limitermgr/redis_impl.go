package limitermgr

import (
	"context"
	"fmt"
	"time"

	"github.com/lite-lake/litecore-go/manager/cachemgr"
	"github.com/lite-lake/litecore-go/manager/loggermgr"
	"github.com/lite-lake/litecore-go/manager/telemetrymgr"
)

// limiterManagerRedisImpl Redis 限流管理器实现
// 使用 Redis 存储限流状态，支持分布式限流
type limiterManagerRedisImpl struct {
	*limiterManagerBaseImpl        // 基类，提供可观测性功能
	name                    string // 管理器名称
}

const limiterKeyPrefix = "limiter:" // Redis 限流键前缀

// NewLimiterManagerRedisImpl 创建 Redis 限流管理器实例
// 参数：
//   - loggerMgr: 日志管理器
//   - telemetryMgr: 遥测管理器
//   - cacheMgr: 缓存管理器
//
// 返回 ILimiterManager 接口实例
func NewLimiterManagerRedisImpl(
	loggerMgr loggermgr.ILoggerManager,
	telemetryMgr telemetrymgr.ITelemetryManager,
	cacheMgr cachemgr.ICacheManager,
) ILimiterManager {
	impl := &limiterManagerRedisImpl{
		limiterManagerBaseImpl: newILimiterManagerBaseImpl(loggerMgr, telemetryMgr, cacheMgr),
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
	if r.cacheMgr == nil {
		return fmt.Errorf("cache manager is not initialized")
	}
	return r.cacheMgr.Health()
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
	return r.AllowN(ctx, key, limit, window, 1)
}

// AllowN 带成本检查是否允许通过限流（内部PRD）
// 固定窗口计数器，消耗语义与 memory 实现一致：已用量 + cost ≤ limit →
// 允许并消耗 cost；否则整批拒绝且不消耗；cost=0（空批）恒允许且不触碰 Redis。
// 实现：cachemgr 门面无原子 check-and-consume 口，先 GetInt64 预判（拒绝路径
// 零消耗），INCRBY 后以累加值复核——并发下预判通过但累加超限者仍拒绝
// （该极端并发下已消耗的成本随窗口过期自愈）。
// 参数：
//   - ctx: 上下文
//   - key: 限流键，标识限流对象
//   - limit: 时间窗口内的最大配额数
//   - window: 时间窗口大小
//   - cost: 本次调用消耗的配额数
//
// 返回: 允许返回 true，否则返回 false
func (r *limiterManagerRedisImpl) AllowN(ctx context.Context, key string, limit int, window time.Duration, cost int) (bool, error) {
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
		if err := ValidateCost(cost); err != nil {
			return err
		}

		if r.cacheMgr == nil {
			return fmt.Errorf("cache manager is not initialized")
		}

		cacheKey := limiterKeyPrefix + key

		if cost == 0 {
			result = true
			r.recordAllowance(ctx, "redis", true)
			return nil
		}

		// 预判：拒绝路径不消耗（键不存在/读取故障按 0 继续，INCRBY 后复核兜底）
		if count, err := r.cacheMgr.GetInt64(ctx, cacheKey); err == nil && count+int64(cost) > int64(limit) {
			result = false
			r.recordAllowance(ctx, "redis", false)
			return nil
		}

		count, err := r.cacheMgr.Increment(ctx, cacheKey, int64(cost))
		if err != nil {
			return fmt.Errorf("failed to increment counter: %w", err)
		}

		// 累加值恰为本次成本 = 本次 INCRBY 新建键，此时才设置窗口过期
		if count == int64(cost) {
			if err := r.cacheMgr.Expire(ctx, cacheKey, window); err != nil {
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

		if r.cacheMgr == nil {
			return fmt.Errorf("cache manager is not initialized")
		}

		cacheKey := limiterKeyPrefix + key

		// 计数器由 Allow(Allow 内 INCRBY) 写入，是纯整数而非 gob；用 GetInt64 读取
		// 与写入格式对齐（用 Get(gob) 会对纯整数反序列化报 unexpected EOF，见 rate-limiter 日志）。
		// 键不存在或 Redis 故障：按剩余=limit fail-open（与既有语义一致）
		count, err := r.cacheMgr.GetInt64(ctx, cacheKey)
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
