package cachemgr

import (
	"context"
	"fmt"
	"sync"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"

	"com.litelake.litecore/common"
	"com.litelake.litecore/manager/cachemgr/internal/config"
	"com.litelake.litecore/manager/cachemgr/internal/drivers"
	"com.litelake.litecore/manager/loggermgr"
	"com.litelake.litecore/manager/telemetrymgr"
)

// Manager 缓存管理器
// 实现依赖注入模式，支持多种缓存驱动（Redis、Memory、None）
type Manager struct {
	// 依赖注入字段
	Config            common.BaseConfigProvider      `inject:""`
	LoggerManager     loggermgr.LoggerManager       `inject:"optional"`
	TelemetryManager  telemetrymgr.TelemetryManager `inject:"optional"`

	// 内部状态
	name              string
	driver            drivers.Driver
	logger            loggermgr.Logger
	tracer            trace.Tracer
	meter             metric.Meter
	cacheHitCounter   metric.Int64Counter
	cacheMissCounter  metric.Int64Counter
	operationDuration metric.Float64Histogram
	mu                sync.RWMutex
	once              sync.Once
}

// NewManager 创建缓存管理器
// name: 管理器名称，用于配置键前缀（如 "default" 对应配置键 "cache.default"）
func NewManager(name string) *Manager {
	return &Manager{
		name:   name,
		driver: drivers.NewNoneDriver(),
	}
}

// ManagerName 返回管理器名称
func (m *Manager) ManagerName() string {
	return m.name
}

// OnStart 初始化管理器（依赖注入完成后调用）
// 加载配置、创建驱动、初始化观测组件
func (m *Manager) OnStart() error {
	var initErr error
	m.once.Do(func() {
		// 1. 从 Config 获取配置
		cfg, err := m.loadConfig()
		if err != nil {
			initErr = fmt.Errorf("load config failed: %w", err)
			return
		}

		// 2. 创建驱动
		driver, err := m.createDriver(cfg)
		if err != nil {
			initErr = fmt.Errorf("create driver failed: %w", err)
			return
		}
		m.driver = driver

		// 3. 初始化 Logger
		if m.LoggerManager != nil {
			m.logger = m.LoggerManager.Logger("cachemgr")
		}

		// 4. 初始化 Telemetry
		if m.TelemetryManager != nil {
			m.tracer = m.TelemetryManager.Tracer("cachemgr")
			m.meter = m.TelemetryManager.Meter("cachemgr")

			// 创建指标
			m.cacheHitCounter, _ = m.meter.Int64Counter(
				"cache.hit",
				metric.WithDescription("Cache hit count"),
				metric.WithUnit("{hit}"),
			)
			m.cacheMissCounter, _ = m.meter.Int64Counter(
				"cache.miss",
				metric.WithDescription("Cache miss count"),
				metric.WithUnit("{miss}"),
			)
			m.operationDuration, _ = m.meter.Float64Histogram(
				"cache.operation.duration",
				metric.WithDescription("Cache operation duration in seconds"),
				metric.WithUnit("s"),
			)
		}

		// 5. 启动驱动
		if err := m.driver.Start(); err != nil {
			initErr = fmt.Errorf("start driver failed: %w", err)
			return
		}
	})
	return initErr
}

// loadConfig 从 ConfigProvider 加载配置
func (m *Manager) loadConfig() (*config.CacheConfig, error) {
	if m.Config == nil {
		return config.DefaultConfig(), nil
	}

	cfgKey := fmt.Sprintf("cache.%s", m.name)
	cfgData, err := m.Config.Get(cfgKey)
	if err != nil {
		// 配置不存在，返回默认配置
		return config.DefaultConfig(), nil
	}

	// 类型断言：将 any 转换为 map[string]any
	cfgMap, ok := cfgData.(map[string]any)
	if !ok {
		return config.DefaultConfig(), nil
	}

	return config.ParseCacheConfigFromMap(cfgMap)
}

// createDriver 根据配置创建缓存驱动
func (m *Manager) createDriver(cfg *config.CacheConfig) (drivers.Driver, error) {
	switch cfg.Driver {
	case "redis":
		if cfg.RedisConfig == nil {
			return nil, fmt.Errorf("redis config is required")
		}
		return drivers.NewRedisDriver(cfg.RedisConfig)
	case "memory":
		if cfg.MemoryConfig == nil {
			return nil, fmt.Errorf("memory config is required")
		}
		return drivers.NewMemoryDriver(cfg.MemoryConfig)
	case "none":
		return drivers.NewNoneDriver(), nil
	default:
		return drivers.NewNoneDriver(), nil
	}
}

// OnStop 停止管理器
func (m *Manager) OnStop() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	return m.driver.Stop()
}

// Health 健康检查
func (m *Manager) Health() error {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return m.driver.Health()
}

// ========== 缓存操作 ==========

// Get 获取缓存值
func (m *Manager) Get(ctx context.Context, key string, dest any) error {
	var hit bool
	var getErr error

	err := m.recordOperation(ctx, "get", key, func() error {
		getErr = m.driver.Get(ctx, key, dest)
		hit = (getErr == nil)
		return getErr
	})

	m.recordCacheHit(ctx, hit)
	return err
}

// Set 设置缓存值
func (m *Manager) Set(ctx context.Context, key string, value any, expiration time.Duration) error {
	return m.recordOperation(ctx, "set", key, func() error {
		return m.driver.Set(ctx, key, value, expiration)
	})
}

// SetNX 仅当键不存在时才设置值
// 返回值表示是否设置成功：true 表示设置成功，false 表示键已存在
func (m *Manager) SetNX(ctx context.Context, key string, value any, expiration time.Duration) (bool, error) {
	var result bool
	var resultErr error

	err := m.recordOperation(ctx, "setnx", key, func() error {
		result, resultErr = m.driver.SetNX(ctx, key, value, expiration)
		return resultErr
	})

	return result, err
}

// Delete 删除缓存值
func (m *Manager) Delete(ctx context.Context, key string) error {
	return m.recordOperation(ctx, "delete", key, func() error {
		return m.driver.Delete(ctx, key)
	})
}

// Exists 检查键是否存在
func (m *Manager) Exists(ctx context.Context, key string) (bool, error) {
	var result bool
	var resultErr error

	err := m.recordOperation(ctx, "exists", key, func() error {
		result, resultErr = m.driver.Exists(ctx, key)
		return resultErr
	})

	return result, err
}

// Expire 设置过期时间
func (m *Manager) Expire(ctx context.Context, key string, expiration time.Duration) error {
	return m.recordOperation(ctx, "expire", key, func() error {
		return m.driver.Expire(ctx, key, expiration)
	})
}

// TTL 获取剩余过期时间
func (m *Manager) TTL(ctx context.Context, key string) (time.Duration, error) {
	var result time.Duration
	var resultErr error

	err := m.recordOperation(ctx, "ttl", key, func() error {
		result, resultErr = m.driver.TTL(ctx, key)
		return resultErr
	})

	return result, err
}

// Clear 清空所有缓存（慎用）
func (m *Manager) Clear(ctx context.Context) error {
	return m.recordOperation(ctx, "clear", "", func() error {
		return m.driver.Clear(ctx)
	})
}

// GetMultiple 批量获取
func (m *Manager) GetMultiple(ctx context.Context, keys []string) (map[string]any, error) {
	var result map[string]any
	var resultErr error

	key := "batch"
	if len(keys) > 0 {
		key = keys[0]
	}

	err := m.recordOperation(ctx, "getmultiple", key, func() error {
		result, resultErr = m.driver.GetMultiple(ctx, keys)
		return resultErr
	})

	return result, err
}

// SetMultiple 批量设置
func (m *Manager) SetMultiple(ctx context.Context, items map[string]any, expiration time.Duration) error {
	key := "batch"
	for k := range items {
		key = k
		break
	}

	return m.recordOperation(ctx, "setmultiple", key, func() error {
		return m.driver.SetMultiple(ctx, items, expiration)
	})
}

// DeleteMultiple 批量删除
func (m *Manager) DeleteMultiple(ctx context.Context, keys []string) error {
	key := "batch"
	if len(keys) > 0 {
		key = keys[0]
	}

	return m.recordOperation(ctx, "deletemultiple", key, func() error {
		return m.driver.DeleteMultiple(ctx, keys)
	})
}

// Increment 自增
func (m *Manager) Increment(ctx context.Context, key string, value int64) (int64, error) {
	var result int64
	var resultErr error

	err := m.recordOperation(ctx, "increment", key, func() error {
		result, resultErr = m.driver.Increment(ctx, key, value)
		return resultErr
	})

	return result, err
}

// Decrement 自减
func (m *Manager) Decrement(ctx context.Context, key string, value int64) (int64, error) {
	var result int64
	var resultErr error

	err := m.recordOperation(ctx, "decrement", key, func() error {
		result, resultErr = m.driver.Decrement(ctx, key, value)
		return resultErr
	})

	return result, err
}

// Close 关闭缓存连接
func (m *Manager) Close() error {
	return m.OnStop()
}

// ========== 观测辅助方法 ==========

// recordOperation 记录缓存操作（带链路追踪和指标）
func (m *Manager) recordOperation(
	ctx context.Context,
	operation string,
	key string,
	fn func() error,
) error {
	// 如果没有可观测性配置，直接执行操作
	if m.tracer == nil && m.logger == nil && m.operationDuration == nil {
		return fn()
	}

	var span trace.Span
	if m.tracer != nil {
		ctx, span = m.tracer.Start(ctx, "cache."+operation,
			trace.WithAttributes(
				attribute.String("cache.key", sanitizeKey(key)),
				attribute.String("cache.driver", m.driver.Name()),
			),
		)
		defer span.End()
	}

	start := time.Now()
	err := fn()
	duration := time.Since(start).Seconds()

	if m.operationDuration != nil {
		m.operationDuration.Record(ctx, duration,
			metric.WithAttributes(
				attribute.String("operation", operation),
				attribute.String("status", getStatus(err)),
			),
		)
	}

	if m.logger != nil {
		if err != nil {
			m.logger.Error("cache operation failed",
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
			m.logger.Debug("cache operation success",
				"operation", operation,
				"key", sanitizeKey(key),
				"duration", duration,
			)
		}
	}

	return err
}

// recordCacheHit 记录缓存命中/未命中
func (m *Manager) recordCacheHit(ctx context.Context, hit bool) {
	if m.meter == nil {
		return
	}

	attrs := metric.WithAttributes(
		attribute.String("cache.driver", m.driver.Name()),
	)

	if hit {
		if m.cacheHitCounter != nil {
			m.cacheHitCounter.Add(ctx, 1, attrs)
		}
	} else {
		if m.cacheMissCounter != nil {
			m.cacheMissCounter.Add(ctx, 1, attrs)
		}
	}
}

// sanitizeKey 和 getStatus 函数已在 cache_adapter.go 中定义
// 这里直接使用它们，避免重复声明

// 确保 Manager 实现 CacheManager 和 common.BaseManager
var _ CacheManager = (*Manager)(nil)
var _ common.BaseManager = (*Manager)(nil)
