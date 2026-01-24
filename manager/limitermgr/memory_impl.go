package limitermgr

import (
	"context"
	"sync"
	"time"

	"github.com/lite-lake/litecore-go/manager/loggermgr"
	"github.com/lite-lake/litecore-go/manager/telemetrymgr"
)

// limiterEntry 限流条目，存储单个限流键的状态
type limiterEntry struct {
	mu        sync.RWMutex  // 读写锁，保证并发安全
	window    []time.Time   // 时间窗口内所有请求的时间戳
	limit     int           // 时间窗口内的最大请求数
	windowDur time.Duration // 时间窗口大小
}

// limiterManagerMemoryImpl 内存限流管理器实现
// 使用本地内存存储限流状态，适用于单实例场景
type limiterManagerMemoryImpl struct {
	*limiterManagerBaseImpl          // 基类，提供可观测性功能
	limiters                sync.Map // 存储所有限流条目，key为限流键，value为*limiterEntry
	name                    string   // 管理器名称
}

// NewLimiterManagerMemoryImpl 创建内存限流管理器实例
// 参数：
//   - loggerMgr: 日志管理器
//   - telemetryMgr: 遥测管理器
//
// 返回 ILimiterManager 接口实例
func NewLimiterManagerMemoryImpl(
	loggerMgr loggermgr.ILoggerManager,
	telemetryMgr telemetrymgr.ITelemetryManager,
) ILimiterManager {
	impl := &limiterManagerMemoryImpl{
		limiterManagerBaseImpl: newILimiterManagerBaseImpl(loggerMgr, telemetryMgr, nil),
		limiters:               sync.Map{},
		name:                   "limiterManagerMemoryImpl",
	}
	impl.initObservability()
	return impl
}

// ManagerName 返回管理器名称
func (m *limiterManagerMemoryImpl) ManagerName() string {
	return m.name
}

// Health 检查管理器健康状态
// 内存限流管理器始终返回 nil，因为它无外部依赖
func (m *limiterManagerMemoryImpl) Health() error {
	return nil
}

// OnStart 启动管理器时的回调
// 内存限流管理器无需额外初始化，返回 nil
func (m *limiterManagerMemoryImpl) OnStart() error {
	return nil
}

// OnStop 停止管理器时的回调
// 内存限流管理器无需清理资源，返回 nil
func (m *limiterManagerMemoryImpl) OnStop() error {
	return nil
}

// Allow 检查是否允许通过限流
// 使用滑动窗口算法，统计时间窗口内的请求数量
// 参数：
//   - ctx: 上下文
//   - key: 限流键，标识限流对象
//   - limit: 时间窗口内的最大请求数
//   - window: 时间窗口大小
//
// 返回: 允许返回 true，否则返回 false
func (m *limiterManagerMemoryImpl) Allow(ctx context.Context, key string, limit int, window time.Duration) (bool, error) {
	var result bool

	err := m.recordOperation(ctx, "memory", "allow", key, func() error {
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

		value, _ := m.limiters.LoadOrStore(key, &limiterEntry{
			limit:     limit,
			windowDur: window,
		})
		entry := value.(*limiterEntry)

		entry.mu.Lock()
		now := time.Now()
		entry.limit = limit
		entry.windowDur = window
		cutoff := now.Add(-window)

		validWindow := make([]time.Time, 0, len(entry.window))
		for _, t := range entry.window {
			if t.After(cutoff) {
				validWindow = append(validWindow, t)
			}
		}
		entry.window = validWindow

		if len(entry.window) < limit {
			entry.window = append(entry.window, now)
			result = true
			m.recordAllowance(ctx, "memory", true)
		} else {
			result = false
			m.recordAllowance(ctx, "memory", false)
		}
		entry.mu.Unlock()

		return nil
	})

	return result, err
}

// GetRemaining 获取剩余可访问次数
// 计算时间窗口内还可通过的请求数量
// 参数：
//   - ctx: 上下文
//   - key: 限流键
//   - limit: 时间窗口内的最大请求数
//   - window: 时间窗口大小
//
// 返回: 剩余次数
func (m *limiterManagerMemoryImpl) GetRemaining(ctx context.Context, key string, limit int, window time.Duration) (int, error) {
	var result int

	err := m.recordOperation(ctx, "memory", "get_remaining", key, func() error {
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

		value, ok := m.limiters.Load(key)
		if !ok {
			result = limit
			return nil
		}

		entry := value.(*limiterEntry)

		entry.mu.RLock()
		now := time.Now()
		entry.limit = limit
		entry.windowDur = window
		cutoff := now.Add(-window)

		count := 0
		for _, t := range entry.window {
			if t.After(cutoff) {
				count++
			}
		}
		result = limit - count
		entry.mu.RUnlock()

		return nil
	})

	return result, err
}

var _ ILimiterManager = (*limiterManagerMemoryImpl)(nil)
