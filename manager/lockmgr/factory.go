package lockmgr

import (
	"fmt"
	"github.com/lite-lake/litecore-go/manager/configmgr"

	"github.com/lite-lake/litecore-go/common"
)

// Build 创建锁管理器实例
// driverType: 驱动类型 ("redis", "memory")
// driverConfig: 驱动配置 (根据驱动类型不同而不同)
//
// 返回 ILockManager 接口实例和可能的错误
// 注意：loggerMgr、telemetryMgr 和 cacheMgr 需要通过容器注入
func Build(
	driverType string,
	driverConfig map[string]any,
) (ILockManager, error) {
	if driverType == "" {
		driverType = "memory"
	}

	switch driverType {
	case "redis":
		redisConfig, err := parseRedisLockConfig(driverConfig)
		if err != nil {
			return nil, err
		}

		mgr := NewLockManagerRedisImpl(redisConfig)
		return mgr, nil

	case "memory":
		memoryConfig, err := parseMemoryLockConfig(driverConfig)
		if err != nil {
			return nil, err
		}

		mgr := NewLockManagerMemoryImpl(memoryConfig)
		return mgr, nil

	default:
		return nil, fmt.Errorf("unsupported driver type: %s", driverType)
	}
}

// BuildWithConfigProvider 从配置提供者创建锁管理器实例
// 自动从配置提供者读取 lock.driver 和对应驱动配置
//
// 返回 ILockManager 接口实例和可能的错误
// 注意：loggerMgr、telemetryMgr 和 cacheMgr 需要通过容器注入
func BuildWithConfigProvider(configProvider configmgr.IConfigManager) (ILockManager, error) {
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

	return Build(driverTypeStr, driverConfig)
}
