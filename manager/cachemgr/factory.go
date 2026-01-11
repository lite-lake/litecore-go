package cachemgr

import (
	"fmt"
	"time"

	"com.litelake.litecore/common"
	"com.litelake.litecore/manager/cachemgr/internal/config"
	"com.litelake.litecore/manager/cachemgr/internal/drivers"
	"com.litelake.litecore/manager/loggermgr"
	"com.litelake.litecore/manager/telemetrymgr"
)

// Deprecated: Factory 模式已废弃，请使用依赖注入模式
// 使用 container.ManagerContainer 和 cachemgr.NewManager() 代替
// 例如：
//
//	container := container.NewManagerContainer(configContainer)
//	mgr := cachemgr.NewManager("default")
//	container.Register(mgr)
//	container.InjectAll()
//	mgr.OnStart()
//
// 本文件将在未来版本中移除

// Build 创建缓存管理器实例
// cfg: 缓存配置内容（包含 driver、redis_config、memory_config 等配置项）
// loggerMgr: 可选的日志管理器
// telemetryMgr: 可选的观测管理器
// 返回实现 common.BaseManager 接口的缓存管理器
func Build(
	cfg map[string]any,
	loggerMgr loggermgr.LoggerManager,
	telemetryMgr telemetrymgr.TelemetryManager,
) common.BaseManager {
	// 解析缓存配置
	cacheConfig, err := config.ParseCacheConfigFromMap(cfg)
	if err != nil {
		// 配置解析失败，返回 none 管理器作为降级
		return NewCacheManagerAdapter(drivers.NewNoneManager(), loggerMgr, telemetryMgr)
	}

	// 验证配置
	if err := cacheConfig.Validate(); err != nil {
		// 配置验证失败，返回 none 管理器作为降级
		return NewCacheManagerAdapter(drivers.NewNoneManager(), loggerMgr, telemetryMgr)
	}

	// 根据驱动类型创建相应的管理器
	var mgr common.BaseManager
	switch cacheConfig.Driver {
	case "redis":
		redisMgr, err := drivers.NewRedisManager(cacheConfig.RedisConfig)
		if err != nil {
			// Redis 初始化失败，降级到 none 管理器
			return NewCacheManagerAdapter(drivers.NewNoneManager(), loggerMgr, telemetryMgr)
		}
		mgr = NewCacheManagerAdapter(redisMgr, loggerMgr, telemetryMgr)
	case "memory":
		memoryMgr := drivers.NewMemoryManager(
			cacheConfig.MemoryConfig.MaxAge,
			cacheConfig.MemoryConfig.MaxAge,
		)
		mgr = NewCacheManagerAdapter(memoryMgr, loggerMgr, telemetryMgr)
	case "none":
		mgr = NewCacheManagerAdapter(drivers.NewNoneManager(), loggerMgr, telemetryMgr)
	default:
		// 未知驱动，返回 none 管理器
		mgr = NewCacheManagerAdapter(drivers.NewNoneManager(), loggerMgr, telemetryMgr)
	}

	return mgr
}

// BuildWithConfig 使用配置结构体创建缓存管理器
// cacheConfig: 缓存配置结构体
// loggerMgr: 可选的日志管理器
// telemetryMgr: 可选的观测管理器
// 返回实现 common.BaseManager 接口的缓存管理器和可能的错误
func BuildWithConfig(
	cacheConfig *config.CacheConfig,
	loggerMgr loggermgr.LoggerManager,
	telemetryMgr telemetrymgr.TelemetryManager,
) (common.BaseManager, error) {
	if err := cacheConfig.Validate(); err != nil {
		return nil, fmt.Errorf("invalid cache config: %w", err)
	}

	var mgr common.BaseManager
	switch cacheConfig.Driver {
	case "redis":
		redisMgr, err := drivers.NewRedisManager(cacheConfig.RedisConfig)
		if err != nil {
			return nil, fmt.Errorf("failed to create redis manager: %w", err)
		}
		mgr = NewCacheManagerAdapter(redisMgr, loggerMgr, telemetryMgr)
	case "memory":
		memoryMgr := drivers.NewMemoryManager(
			cacheConfig.MemoryConfig.MaxAge,
			cacheConfig.MemoryConfig.MaxAge,
		)
		mgr = NewCacheManagerAdapter(memoryMgr, loggerMgr, telemetryMgr)
	case "none":
		mgr = NewCacheManagerAdapter(drivers.NewNoneManager(), loggerMgr, telemetryMgr)
	default:
		return nil, fmt.Errorf("unsupported driver: %s", cacheConfig.Driver)
	}

	return mgr, nil
}

// BuildRedis 创建 Redis 缓存管理器（便捷方法）
// loggerMgr: 可选的日志管理器
// telemetryMgr: 可选的观测管理器
func BuildRedis(
	host string, port int, password string, db int,
	loggerMgr loggermgr.LoggerManager,
	telemetryMgr telemetrymgr.TelemetryManager,
) common.BaseManager {
	cfg := &config.RedisConfig{
		Host:            host,
		Port:            port,
		Password:        password,
		DB:              db,
		MaxIdleConns:    config.DefaultRedisMaxIdleConns,
		MaxOpenConns:    config.DefaultRedisMaxOpenConns,
		ConnMaxLifetime: config.DefaultRedisConnMaxLifetime,
	}

	redisMgr, err := drivers.NewRedisManager(cfg)
	if err != nil {
		return NewCacheManagerAdapter(drivers.NewNoneManager(), loggerMgr, telemetryMgr)
	}

	return NewCacheManagerAdapter(redisMgr, loggerMgr, telemetryMgr)
}

// BuildMemory 创建内存缓存管理器（便捷方法）
// loggerMgr: 可选的日志管理器
// telemetryMgr: 可选的观测管理器
func BuildMemory(
	defaultExpiration, cleanupInterval time.Duration,
	loggerMgr loggermgr.LoggerManager,
	telemetryMgr telemetrymgr.TelemetryManager,
) common.BaseManager {
	memoryMgr := drivers.NewMemoryManager(defaultExpiration, cleanupInterval)
	return NewCacheManagerAdapter(memoryMgr, loggerMgr, telemetryMgr)
}

// BuildNone 创建空缓存管理器（便捷方法）
// loggerMgr: 可选的日志管理器
// telemetryMgr: 可选的观测管理器
func BuildNone(
	loggerMgr loggermgr.LoggerManager,
	telemetryMgr telemetrymgr.TelemetryManager,
) common.BaseManager {
	return NewCacheManagerAdapter(drivers.NewNoneManager(), loggerMgr, telemetryMgr)
}
