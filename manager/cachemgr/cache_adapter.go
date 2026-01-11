package cachemgr

import (
	"context"
	"time"

	"com.litelake.litecore/manager/cachemgr/internal/drivers"
	"com.litelake.litecore/manager/loggermgr"
	"com.litelake.litecore/manager/telemetrymgr"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"
)

// CacheManagerAdapter 缓存管理器适配器
// 将内部驱动适配到 CacheManager 接口
type CacheManagerAdapter struct {
	driver            CacheDriver
	logger            loggermgr.Logger        // 日志实例
	tracer            trace.Tracer            // 链路追踪
	meter             metric.Meter            // 指标收集
	cacheHitCounter   metric.Int64Counter     // 缓存命中计数器
	cacheMissCounter  metric.Int64Counter     // 缓存未命中计数器
	operationDuration metric.Float64Histogram // 操作耗时
}

// CacheDriver 缓存驱动接口
// 内部驱动需要实现此接口
type CacheDriver interface {
	// Manager 接口方法
	ManagerName() string
	Health() error
	OnStart() error
	OnStop() error
	Close() error

	// 缓存操作方法
	Get(ctx context.Context, key string, dest any) error
	Set(ctx context.Context, key string, value any, expiration time.Duration) error
	SetNX(ctx context.Context, key string, value any, expiration time.Duration) (bool, error)
	Delete(ctx context.Context, key string) error
	Exists(ctx context.Context, key string) (bool, error)
	Expire(ctx context.Context, key string, expiration time.Duration) error
	TTL(ctx context.Context, key string) (time.Duration, error)
	Clear(ctx context.Context) error
	GetMultiple(ctx context.Context, keys []string) (map[string]any, error)
	SetMultiple(ctx context.Context, items map[string]any, expiration time.Duration) error
	DeleteMultiple(ctx context.Context, keys []string) error
	Increment(ctx context.Context, key string, value int64) (int64, error)
	Decrement(ctx context.Context, key string, value int64) (int64, error)
}

// NewCacheManagerAdapter 创建缓存管理器适配器
// loggerMgr: 可选的日志管理器
// telemetryMgr: 可选的观测管理器
func NewCacheManagerAdapter(
	driver CacheDriver,
	loggerMgr loggermgr.LoggerManager,
	telemetryMgr telemetrymgr.TelemetryManager,
) *CacheManagerAdapter {
	adapter := &CacheManagerAdapter{
		driver: driver,
	}

	// 初始化日志
	if loggerMgr != nil {
		adapter.logger = loggerMgr.Logger("cachemgr")
	}

	// 初始化链路追踪和指标
	if telemetryMgr != nil {
		adapter.tracer = telemetryMgr.Tracer("cachemgr")
		adapter.meter = telemetryMgr.Meter("cachemgr")

		// 创建计数器
		cacheHitCounter, err := adapter.meter.Int64Counter(
			"cache.hit",
			metric.WithDescription("Cache hit count"),
			metric.WithUnit("{hit}"),
		)
		if err == nil {
			adapter.cacheHitCounter = cacheHitCounter
		}

		cacheMissCounter, err := adapter.meter.Int64Counter(
			"cache.miss",
			metric.WithDescription("Cache miss count"),
			metric.WithUnit("{miss}"),
		)
		if err == nil {
			adapter.cacheMissCounter = cacheMissCounter
		}

		operationDuration, err := adapter.meter.Float64Histogram(
			"cache.operation.duration",
			metric.WithDescription("Cache operation duration in seconds"),
			metric.WithUnit("s"),
		)
		if err == nil {
			adapter.operationDuration = operationDuration
		}
	}

	return adapter
}

// ManagerName 返回管理器名称
func (a *CacheManagerAdapter) ManagerName() string {
	return a.driver.ManagerName()
}

// Health 检查管理器健康状态
func (a *CacheManagerAdapter) Health() error {
	return a.driver.Health()
}

// OnStart 在服务器启动时触发
func (a *CacheManagerAdapter) OnStart() error {
	return a.driver.OnStart()
}

// OnStop 在服务器停止时触发
func (a *CacheManagerAdapter) OnStop() error {
	return a.driver.OnStop()
}

// Get 获取缓存值
func (a *CacheManagerAdapter) Get(ctx context.Context, key string, dest any) error {
	var hit bool
	var getErr error

	err := a.recordOperation(ctx, "get", key, func() error {
		getErr = a.driver.Get(ctx, key, dest)
		// 判断是否命中缓存（没有错误表示命中）
		hit = (getErr == nil)
		return getErr
	})

	// 记录缓存命中/未命中
	a.recordCacheHit(ctx, hit)

	return err
}

// Set 设置缓存值
func (a *CacheManagerAdapter) Set(ctx context.Context, key string, value any, expiration time.Duration) error {
	return a.recordOperation(ctx, "set", key, func() error {
		return a.driver.Set(ctx, key, value, expiration)
	})
}

// SetNX 仅当键不存在时才设置值
func (a *CacheManagerAdapter) SetNX(ctx context.Context, key string, value any, expiration time.Duration) (bool, error) {
	var result bool
	var resultErr error

	err := a.recordOperation(ctx, "setnx", key, func() error {
		result, resultErr = a.driver.SetNX(ctx, key, value, expiration)
		return resultErr
	})

	return result, err
}

// Delete 删除缓存值
func (a *CacheManagerAdapter) Delete(ctx context.Context, key string) error {
	return a.recordOperation(ctx, "delete", key, func() error {
		return a.driver.Delete(ctx, key)
	})
}

// Exists 检查键是否存在
func (a *CacheManagerAdapter) Exists(ctx context.Context, key string) (bool, error) {
	var result bool
	var resultErr error

	err := a.recordOperation(ctx, "exists", key, func() error {
		result, resultErr = a.driver.Exists(ctx, key)
		return resultErr
	})

	return result, err
}

// Expire 设置过期时间
func (a *CacheManagerAdapter) Expire(ctx context.Context, key string, expiration time.Duration) error {
	return a.recordOperation(ctx, "expire", key, func() error {
		return a.driver.Expire(ctx, key, expiration)
	})
}

// TTL 获取剩余过期时间
func (a *CacheManagerAdapter) TTL(ctx context.Context, key string) (time.Duration, error) {
	var result time.Duration
	var resultErr error

	err := a.recordOperation(ctx, "ttl", key, func() error {
		result, resultErr = a.driver.TTL(ctx, key)
		return resultErr
	})

	return result, err
}

// Clear 清空所有缓存
func (a *CacheManagerAdapter) Clear(ctx context.Context) error {
	// Clear 操作不需要 key，使用特殊处理
	if a.tracer == nil && a.logger == nil && a.operationDuration == nil {
		return a.driver.Clear(ctx)
	}

	var span trace.Span
	if a.tracer != nil {
		ctx, span = a.tracer.Start(ctx, "cache.clear",
			trace.WithAttributes(
				attribute.String("cache.driver", a.driver.ManagerName()),
			),
		)
		defer span.End()
	}

	start := time.Now()
	err := a.driver.Clear(ctx)
	duration := time.Since(start).Seconds()

	if a.operationDuration != nil {
		a.operationDuration.Record(ctx, duration,
			metric.WithAttributes(
				attribute.String("operation", "clear"),
				attribute.String("status", getStatus(err)),
			),
		)
	}

	if a.logger != nil {
		if err != nil {
			a.logger.Error("cache clear failed",
				"error", err.Error(),
				"duration", duration,
			)
			if span != nil {
				span.RecordError(err)
				span.SetStatus(codes.Error, err.Error())
			}
		} else {
			a.logger.Info("cache clear success",
				"duration", duration,
			)
		}
	}

	return err
}

// GetMultiple 批量获取
func (a *CacheManagerAdapter) GetMultiple(ctx context.Context, keys []string) (map[string]any, error) {
	var result map[string]any
	var resultErr error

	// 批量操作使用第一个键作为 key（如果没有则为 empty）
	key := "batch"
	if len(keys) > 0 {
		key = keys[0]
	}

	err := a.recordOperation(ctx, "getmultiple", key, func() error {
		result, resultErr = a.driver.GetMultiple(ctx, keys)
		return resultErr
	})

	return result, err
}

// SetMultiple 批量设置
func (a *CacheManagerAdapter) SetMultiple(ctx context.Context, items map[string]any, expiration time.Duration) error {
	// 获取第一个键作为代表
	key := "batch"
	for k := range items {
		key = k
		break
	}

	return a.recordOperation(ctx, "setmultiple", key, func() error {
		return a.driver.SetMultiple(ctx, items, expiration)
	})
}

// DeleteMultiple 批量删除
func (a *CacheManagerAdapter) DeleteMultiple(ctx context.Context, keys []string) error {
	key := "batch"
	if len(keys) > 0 {
		key = keys[0]
	}

	return a.recordOperation(ctx, "deletemultiple", key, func() error {
		return a.driver.DeleteMultiple(ctx, keys)
	})
}

// Increment 自增
func (a *CacheManagerAdapter) Increment(ctx context.Context, key string, value int64) (int64, error) {
	var result int64
	var resultErr error

	err := a.recordOperation(ctx, "increment", key, func() error {
		result, resultErr = a.driver.Increment(ctx, key, value)
		return resultErr
	})

	return result, err
}

// Decrement 自减
func (a *CacheManagerAdapter) Decrement(ctx context.Context, key string, value int64) (int64, error) {
	var result int64
	var resultErr error

	err := a.recordOperation(ctx, "decrement", key, func() error {
		result, resultErr = a.driver.Decrement(ctx, key, value)
		return resultErr
	})

	return result, err
}

// Close 关闭缓存连接
func (a *CacheManagerAdapter) Close() error {
	return a.driver.Close()
}

// 观测辅助方法

// recordOperation 记录缓存操作（带链路追踪和指标）
func (a *CacheManagerAdapter) recordOperation(
	ctx context.Context,
	operation string,
	key string,
	fn func() error,
) error {
	// 如果没有可观测性配置，直接执行操作
	if a.tracer == nil && a.logger == nil && a.operationDuration == nil {
		return fn()
	}

	var span trace.Span
	// 创建 span
	if a.tracer != nil {
		ctx, span = a.tracer.Start(ctx, "cache."+operation,
			trace.WithAttributes(
				attribute.String("cache.key", sanitizeKey(key)),
				attribute.String("cache.driver", a.driver.ManagerName()),
			),
		)
		defer span.End()
	}

	// 执行操作
	start := time.Now()
	err := fn()
	duration := time.Since(start).Seconds()

	// 记录指标
	if a.operationDuration != nil {
		a.operationDuration.Record(ctx, duration,
			metric.WithAttributes(
				attribute.String("operation", operation),
				attribute.String("status", getStatus(err)),
			),
		)
	}

	// 记录日志
	if a.logger != nil {
		if err != nil {
			a.logger.Error("cache operation failed",
				"operation", operation,
				"key", sanitizeKey(key),
				"error", err.Error(),
				"duration", duration,
			)
			if span != nil {
				span.RecordError(err)
				span.SetStatus(codes.Error, err.Error())
			}
		} else {
			a.logger.Debug("cache operation success",
				"operation", operation,
				"key", sanitizeKey(key),
				"duration", duration,
			)
		}
	}

	return err
}

// recordCacheHit 记录缓存命中/未命中
func (a *CacheManagerAdapter) recordCacheHit(ctx context.Context, hit bool) {
	if a.meter == nil {
		return
	}

	attrs := metric.WithAttributes(
		attribute.String("cache.driver", a.driver.ManagerName()),
	)

	if hit {
		if a.cacheHitCounter != nil {
			a.cacheHitCounter.Add(ctx, 1, attrs)
		}
	} else {
		if a.cacheMissCounter != nil {
			a.cacheMissCounter.Add(ctx, 1, attrs)
		}
	}
}

// sanitizeKey 对缓存键进行脱敏处理，避免敏感信息泄露
func sanitizeKey(key string) string {
	if len(key) <= 10 {
		return key
	}
	// 保留前5个字符，其余用***代替
	return key[:5] + "***"
}

// getStatus 根据错误返回状态字符串
func getStatus(err error) string {
	if err != nil {
		return "error"
	}
	return "success"
}

// Ensure CacheManagerAdapter implements CacheManager interface
var _ CacheManager = (*CacheManagerAdapter)(nil)

// 确保 RedisManager 实现 CacheDriver 接口
var _ CacheDriver = (*drivers.RedisManager)(nil)

// 确保 MemoryManager 实现 CacheDriver 接口
var _ CacheDriver = (*drivers.MemoryManager)(nil)

// 确保 NoneManager 实现 CacheDriver 接口
var _ CacheDriver = (*drivers.NoneManager)(nil)
