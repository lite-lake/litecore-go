package cachemgr

import (
	"fmt"

	"com.litelake.litecore/common"
)

// Build 创建缓存管理器实例
// driverType: 驱动类型 ("redis", "memory", "none")
// driverConfig: 驱动配置 (根据驱动类型不同而不同)
//   - redis: 传递给 parseRedisConfig 的 map[string]any
//   - memory: 传递给 parseMemoryConfig 的 map[string]any
//   - none: 忽略
//
// 返回 CacheManager 接口实例和可能的错误
// 注意：loggerMgr 和 telemetryMgr 需要通过容器注入
func Build(
	driverType string,
	driverConfig map[string]any,
) (CacheManager, error) {
	switch driverType {
	case "redis":
		redisConfig, err := parseRedisConfig(driverConfig)
		if err != nil {
			return nil, err
		}

		mgr, err := NewCacheManagerRedisImpl(redisConfig)
		if err != nil {
			return nil, err
		}

		return mgr, nil

	case "memory":
		memoryConfig, err := parseMemoryConfig(driverConfig)
		if err != nil {
			return nil, err
		}

		// 使用 MemoryConfig 中的 MaxAge 作为缓存过期时间
		mgr := NewCacheManagerMemoryImpl(
			memoryConfig.MaxAge,
			memoryConfig.MaxAge/2, // 清理间隔设为过期时间的一半
		)

		return mgr, nil

	case "none":
		mgr := NewCacheManagerNoneImpl()
		return mgr, nil

	default:
		return nil, fmt.Errorf("unsupported driver type: %s", driverType)
	}
}

// BuildWithConfigProvider 从配置提供者创建缓存管理器实例
// 自动从配置提供者读取 cache.driver 和对应驱动配置
// 配置路径：
//   - cache.driver: 驱动类型 ("redis", "memory", "none")
//   - cache.redis_config: Redis 驱动配置（当 driver=redis 时使用）
//   - cache.memory_config: Memory 驱动配置（当 driver=memory 时使用）
//
// 返回 CacheManager 接口实例和可能的错误
// 注意：loggerMgr 和 telemetryMgr 需要通过容器注入
func BuildWithConfigProvider(configProvider common.BaseConfigProvider) (CacheManager, error) {
	if configProvider == nil {
		return nil, fmt.Errorf("configProvider cannot be nil")
	}

	// 1. 读取驱动类型 cache.driver
	driverType, err := configProvider.Get("cache.driver")
	if err != nil {
		return nil, fmt.Errorf("failed to get cache.driver: %w", err)
	}

	driverTypeStr, ok := driverType.(string)
	if !ok {
		return nil, fmt.Errorf("cache.driver must be a string, got %T", driverType)
	}

	// 2. 根据驱动类型读取对应配置
	var driverConfig map[string]any

	switch driverTypeStr {
	case "redis":
		redisConfig, err := configProvider.Get("cache.redis_config")
		if err != nil {
			return nil, fmt.Errorf("failed to get cache.redis_config: %w", err)
		}
		driverConfig, ok = redisConfig.(map[string]any)
		if !ok {
			return nil, fmt.Errorf("cache.redis_config must be a map, got %T", redisConfig)
		}

	case "memory":
		memoryConfig, err := configProvider.Get("cache.memory_config")
		if err != nil {
			return nil, fmt.Errorf("failed to get cache.memory_config: %w", err)
		}
		driverConfig, ok = memoryConfig.(map[string]any)
		if !ok {
			return nil, fmt.Errorf("cache.memory_config must be a map, got %T", memoryConfig)
		}

	case "none":
		// none 驱动不需要配置
		driverConfig = nil

	default:
		return nil, fmt.Errorf("unsupported driver type: %s (must be redis, memory, or none)", driverTypeStr)
	}

	// 3. 调用 Build 函数创建实例
	return Build(driverTypeStr, driverConfig)
}
