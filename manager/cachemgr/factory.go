package cachemgr

import (
	"fmt"
	"time"

	"com.litelake.litecore/common"
	"com.litelake.litecore/manager/cachemgr/internal/config"
	"com.litelake.litecore/manager/cachemgr/internal/drivers"
)

// Build 创建缓存管理器实例
// cfg: 缓存配置内容（包含 driver、redis_config、memory_config 等配置项）
// 返回实现 common.Manager 接口的缓存管理器
func Build(cfg map[string]any) common.Manager {
	// 解析缓存配置
	cacheConfig, err := config.ParseCacheConfigFromMap(cfg)
	if err != nil {
		// 配置解析失败，返回 none 管理器作为降级
		return NewCacheManagerAdapter(drivers.NewNoneManager())
	}

	// 验证配置
	if err := cacheConfig.Validate(); err != nil {
		// 配置验证失败，返回 none 管理器作为降级
		return NewCacheManagerAdapter(drivers.NewNoneManager())
	}

	// 根据驱动类型创建相应的管理器
	var mgr common.Manager
	switch cacheConfig.Driver {
	case "redis":
		redisMgr, err := drivers.NewRedisManager(cacheConfig.RedisConfig)
		if err != nil {
			// Redis 初始化失败，降级到 none 管理器
			return NewCacheManagerAdapter(drivers.NewNoneManager())
		}
		mgr = NewCacheManagerAdapter(redisMgr)
	case "memory":
		memoryMgr := drivers.NewMemoryManager(
			cacheConfig.MemoryConfig.MaxAge,
			cacheConfig.MemoryConfig.MaxAge,
		)
		mgr = NewCacheManagerAdapter(memoryMgr)
	case "none":
		mgr = NewCacheManagerAdapter(drivers.NewNoneManager())
	default:
		// 未知驱动，返回 none 管理器
		mgr = NewCacheManagerAdapter(drivers.NewNoneManager())
	}

	return mgr
}

// BuildWithConfig 使用配置结构体创建缓存管理器
// cacheConfig: 缓存配置结构体
// 返回实现 common.Manager 接口的缓存管理器和可能的错误
func BuildWithConfig(cacheConfig *config.CacheConfig) (common.Manager, error) {
	if err := cacheConfig.Validate(); err != nil {
		return nil, fmt.Errorf("invalid cache config: %w", err)
	}

	var mgr common.Manager
	switch cacheConfig.Driver {
	case "redis":
		redisMgr, err := drivers.NewRedisManager(cacheConfig.RedisConfig)
		if err != nil {
			return nil, fmt.Errorf("failed to create redis manager: %w", err)
		}
		mgr = NewCacheManagerAdapter(redisMgr)
	case "memory":
		memoryMgr := drivers.NewMemoryManager(
			cacheConfig.MemoryConfig.MaxAge,
			cacheConfig.MemoryConfig.MaxAge,
		)
		mgr = NewCacheManagerAdapter(memoryMgr)
	case "none":
		mgr = NewCacheManagerAdapter(drivers.NewNoneManager())
	default:
		return nil, fmt.Errorf("unsupported driver: %s", cacheConfig.Driver)
	}

	return mgr, nil
}

// BuildRedis 创建 Redis 缓存管理器（便捷方法）
func BuildRedis(host string, port int, password string, db int) common.Manager {
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
		return NewCacheManagerAdapter(drivers.NewNoneManager())
	}

	return NewCacheManagerAdapter(redisMgr)
}

// BuildMemory 创建内存缓存管理器（便捷方法）
func BuildMemory(defaultExpiration, cleanupInterval time.Duration) common.Manager {
	memoryMgr := drivers.NewMemoryManager(defaultExpiration, cleanupInterval)
	return NewCacheManagerAdapter(memoryMgr)
}

// BuildNone 创建空缓存管理器（便捷方法）
func BuildNone() common.Manager {
	return NewCacheManagerAdapter(drivers.NewNoneManager())
}
