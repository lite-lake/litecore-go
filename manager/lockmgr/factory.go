package lockmgr

import (
	"fmt"
	"github.com/lite-lake/litecore-go/manager/cachemgr"
	"github.com/lite-lake/litecore-go/manager/configmgr"
	"github.com/lite-lake/litecore-go/manager/loggermgr"
	"github.com/lite-lake/litecore-go/manager/telemetrymgr"

	"github.com/lite-lake/litecore-go/common"
)

// Build 创建锁管理器实例
// 参数：
//   - driverType: 驱动类型 ("redis", "memory")
//   - driverConfig: 驱动配置 (根据驱动类型不同而不同)
//   - loggerMgr: 日志管理器
//   - telemetryMgr: 遥测管理器
//   - cacheMgr: 缓存管理器（Redis 模式必需，Memory 模式可以为 nil）
//
// 返回 ILockManager 接口实例和可能的错误
func Build(
	driverType string,
	driverConfig map[string]any,
	loggerMgr loggermgr.ILoggerManager,
	telemetryMgr telemetrymgr.ITelemetryManager,
	cacheMgr cachemgr.ICacheManager,
) (ILockManager, error) {
	if driverType == "" {
		driverType = "memory"
	}

	switch driverType {
	case "redis":
		// 契约自检：liteauth 硬性契约要求 lock=redis 时底座 cache 必须同为 redis 实现
		// （锁实际复用全局 cacheMgr 连接；cache=none/memory 会导致锁等待超时或静默退化为进程内锁）
		if kind := cachemgr.DriverKind(cacheMgr); kind != "redis" {
			if kind == "" {
				kind = "nil"
			}
			return nil, fmt.Errorf(
				"lockmgr: lock.driver=redis requires cache.driver=redis（liteauth 硬性契约：锁实际复用全局 cacheMgr，"+
					"见 liteauth-go README「接入硬性契约」；违反后果：cache=none → SetNX 恒失败 → 锁等待超时，"+
					"cache=memory → 锁退化进程内锁；当前 cache 底座=%s）", kind)
		}

		redisConfig, err := parseRedisLockConfig(driverConfig)
		if err != nil {
			return nil, err
		}

		mgr := NewLockManagerRedisImpl(loggerMgr, telemetryMgr, cacheMgr, redisConfig)
		return mgr, nil

	case "memory":
		memoryConfig, err := parseMemoryLockConfig(driverConfig)
		if err != nil {
			return nil, err
		}

		mgr := NewLockManagerMemoryImpl(loggerMgr, telemetryMgr, memoryConfig)
		return mgr, nil

	default:
		return nil, fmt.Errorf("unsupported driver type: %s", driverType)
	}
}

// BuildWithConfigProvider 从配置提供者创建锁管理器实例
// 自动从配置提供者读取 lock.driver 和对应驱动配置
// 参数：
//   - configProvider: 配置提供者
//   - loggerMgr: 日志管理器
//   - telemetryMgr: 遥测管理器
//   - cacheMgr: 缓存管理器（Redis 模式必需，Memory 模式可以为 nil）
//
// 返回 ILockManager 接口实例和可能的错误
func BuildWithConfigProvider(
	configProvider configmgr.IConfigManager,
	loggerMgr loggermgr.ILoggerManager,
	telemetryMgr telemetrymgr.ITelemetryManager,
	cacheMgr cachemgr.ICacheManager,
) (ILockManager, error) {
	if configProvider == nil {
		return nil, fmt.Errorf("configProvider cannot be nil")
	}

	driverType, err := configProvider.Get("lock.driver")
	if err != nil {
		return nil, fmt.Errorf("failed to get lock.driver: %w", err)
	}

	driverTypeStr, err := common.GetString(driverType)
	if err != nil {
		return nil, fmt.Errorf("lock.driver: %w", err)
	}

	var driverConfig map[string]any

	switch driverTypeStr {
	case "redis":
		redisConfig, err := configProvider.Get("lock.redis_config")
		if err != nil {
			return nil, fmt.Errorf("failed to get lock.redis_config: %w", err)
		}
		driverConfig, err = common.GetMap(redisConfig)
		if err != nil {
			return nil, fmt.Errorf("lock.redis_config: %w", err)
		}

	case "memory":
		memoryConfig, err := configProvider.Get("lock.memory_config")
		if err != nil {
			return nil, fmt.Errorf("failed to get lock.memory_config: %w", err)
		}
		driverConfig, err = common.GetMap(memoryConfig)
		if err != nil {
			return nil, fmt.Errorf("lock.memory_config: %w", err)
		}

	default:
		return nil, fmt.Errorf("unsupported driver type: %s (must be redis or memory)", driverTypeStr)
	}

	return Build(driverTypeStr, driverConfig, loggerMgr, telemetryMgr, cacheMgr)
}
