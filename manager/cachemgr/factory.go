package cachemgr

import (
	"fmt"
)

// Build 创建缓存管理器实例
// driverType: 驱动类型 ("redis", "memory", "none")
// driverConfig: 驱动配置 (根据驱动类型不同而不同)
//   - redis: 传递给 parseRedisConfig 的 map[string]any
//   - memory: 传递给 parseMemoryConfig 的 map[string]any
//   - none: 忽略
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
