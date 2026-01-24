package limitermgr

import (
	"fmt"

	"github.com/lite-lake/litecore-go/manager/cachemgr"
	"github.com/lite-lake/litecore-go/manager/configmgr"
	"github.com/lite-lake/litecore-go/manager/loggermgr"
	"github.com/lite-lake/litecore-go/manager/telemetrymgr"

	"github.com/lite-lake/litecore-go/common"
)

// Build 创建限流管理器实例
// driverType: 驱动类型 ("redis", "memory", "none")
// driverConfig: 驱动配置 (根据驱动类型不同而不同)
// loggerMgr: 日志管理器
// telemetryMgr: 遥测管理器
// cacheMgr: 缓存管理器（Redis 实现需要）
//
// 返回 ILimiterManager 接口实例和可能的错误
func Build(
	driverType string,
	driverConfig map[string]any,
	loggerMgr loggermgr.ILoggerManager,
	telemetryMgr telemetrymgr.ITelemetryManager,
	cacheMgr cachemgr.ICacheManager,
) (ILimiterManager, error) {
	switch driverType {
	case "redis":
		_, err := parseRedisLimiterConfig(driverConfig)
		if err != nil {
			return nil, err
		}

		mgr := NewLimiterManagerRedisImpl(loggerMgr, telemetryMgr, cacheMgr)
		return mgr, nil

	case "memory":
		_, err := parseMemoryLimiterConfig(driverConfig)
		if err != nil {
			return nil, err
		}

		mgr := NewLimiterManagerMemoryImpl(loggerMgr, telemetryMgr)
		return mgr, nil

	default:
		if driverType == "" {
			return nil, fmt.Errorf("driver type is required")
		}
		return nil, fmt.Errorf("unsupported driver type: %s (must be redis or memory)", driverType)
	}
}

// BuildWithConfigProvider 从配置提供者创建限流管理器实例
// 自动从配置提供者读取 limiter.driver 和对应驱动配置
//
// 参数：
//   - configProvider: 配置管理器
//   - loggerMgr: 日志管理器
//   - telemetryMgr: 遥测管理器
//   - cacheMgr: 缓存管理器（Redis 实现需要）
//
// 返回 ILimiterManager 接口实例和可能的错误
func BuildWithConfigProvider(
	configProvider configmgr.IConfigManager,
	loggerMgr loggermgr.ILoggerManager,
	telemetryMgr telemetrymgr.ITelemetryManager,
	cacheMgr cachemgr.ICacheManager,
) (ILimiterManager, error) {
	if configProvider == nil {
		return nil, fmt.Errorf("configProvider cannot be nil")
	}

	driverType, err := configProvider.Get("limiter.driver")
	if err != nil {
		return nil, fmt.Errorf("failed to get limiter.driver: %w", err)
	}

	driverTypeStr, err := common.GetString(driverType)
	if err != nil {
		return nil, fmt.Errorf("limiter.driver: %w", err)
	}

	var driverConfig map[string]any

	switch driverTypeStr {
	case "redis":
		redisConfig, err := configProvider.Get("limiter.redis_config")
		if err != nil {
			return nil, fmt.Errorf("failed to get limiter.redis_config: %w", err)
		}
		driverConfig, err = common.GetMap(redisConfig)
		if err != nil {
			return nil, fmt.Errorf("limiter.redis_config: %w", err)
		}

	case "memory":
		memoryConfig, err := configProvider.Get("limiter.memory_config")
		if err != nil {
			return nil, fmt.Errorf("failed to get limiter.memory_config: %w", err)
		}
		driverConfig, err = common.GetMap(memoryConfig)
		if err != nil {
			return nil, fmt.Errorf("limiter.memory_config: %w", err)
		}

	default:
		return nil, fmt.Errorf("unsupported driver type: %s (must be redis or memory)", driverTypeStr)
	}

	return Build(driverTypeStr, driverConfig, loggerMgr, telemetryMgr, cacheMgr)
}
