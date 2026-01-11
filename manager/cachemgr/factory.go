package cachemgr

import (
	"fmt"

	"com.litelake.litecore/manager/loggermgr"
	"com.litelake.litecore/manager/telemetrymgr"
)

// Build 创建缓存管理器实例
// driverType: 驱动类型 ("redis", "memory", "none")
// driverConfig: 驱动配置 (根据驱动类型不同而不同)
//   - redis: 传递给 parseRedisConfig 的 map[string]any
//   - memory: 传递给 parseMemoryConfig 的 map[string]any
//   - none: 忽略
// loggerMgr: 可选的日志管理器
// telemetryMgr: 可选的观测管理器
// 返回 CacheManager 接口实例和可能的错误
func Build(
	driverType string,
	driverConfig map[string]any,
	loggerMgr loggermgr.LoggerManager,
	telemetryMgr telemetrymgr.TelemetryManager,
) (CacheManager, error) {
	switch driverType {
	case "redis":
		redisConfig, err := parseRedisConfig(driverConfig)
		if err != nil {
			return nil, err
		}

		mgr, err := NewCacheManagerRedisImpl(redisConfig, loggerMgr, telemetryMgr)
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
			loggerMgr,
			telemetryMgr,
		)

		return mgr, nil

	case "none":
		mgr := NewCacheManagerNoneImpl(loggerMgr, telemetryMgr)
		return mgr, nil

	default:
		return nil, fmt.Errorf("unsupported driver type: %s", driverType)
	}
}
