package loggermgr

import (
	"fmt"
	"strings"

	"com.litelake.litecore/common"
)

// Build 创建日志管理器实例
// driverType: 驱动类型 ("zap", "none")
// driverConfig: 驱动配置 (根据驱动类型不同而不同)
//   - zap: 传递给 ParseLoggerConfigFromMap 的 map[string]any
//   - none: 忽略
//
// 返回 LoggerManager 接口实例和可能的错误
func Build(
	driverType string,
	driverConfig map[string]any,
) (ILoggerManager, error) {
	// 标准化驱动类型（大小写不敏感，去除空格）
	driverType = strings.ToLower(strings.TrimSpace(driverType))

	switch driverType {
	case "zap":
		// 解析 Zap 配置
		loggerConfig, err := ParseLoggerConfigFromMap(driverConfig)
		if err != nil {
			return nil, fmt.Errorf("failed to parse zap config: %w", err)
		}

		// 设置驱动类型
		loggerConfig.Driver = driverType

		// 验证配置
		if err := loggerConfig.Validate(); err != nil {
			return nil, fmt.Errorf("invalid config: %w", err)
		}

		// 创建 Zap 实现
		mgr, err := NewLoggerManagerZapImpl(loggerConfig)
		if err != nil {
			return nil, fmt.Errorf("failed to create zap manager: %w", err)
		}

		return mgr, nil

	case "none":
		mgr := NewLoggerManagerNoneImpl()
		return mgr, nil

	default:
		return nil, fmt.Errorf("unsupported driver type: %s (must be zap or none)", driverType)
	}
}

// BuildWithConfigProvider 从配置提供者创建日志管理器实例
// 自动从配置提供者读取 logger.driver 和对应驱动配置
// 配置路径：
//   - logger.driver: 驱动类型 ("zap", "none")
//   - logger.zap_config: Zap 驱动配置（当 driver=zap 时使用）
//
// 返回 LoggerManager 接口实例和可能的错误
func BuildWithConfigProvider(configProvider common.IBaseConfigProvider) (ILoggerManager, error) {
	if configProvider == nil {
		return nil, fmt.Errorf("configProvider cannot be nil")
	}

	// 1. 读取驱动类型 logger.driver
	driverType, err := configProvider.Get("logger.driver")
	if err != nil {
		return nil, fmt.Errorf("failed to get logger.driver: %w", err)
	}

	driverTypeStr, ok := driverType.(string)
	if !ok {
		return nil, fmt.Errorf("logger.driver must be a string, got %T", driverType)
	}

	// 标准化驱动类型（大小写不敏感，去除空格）
	driverTypeStr = strings.ToLower(strings.TrimSpace(driverTypeStr))

	// 2. 根据驱动类型读取对应配置
	var driverConfig map[string]any

	switch driverTypeStr {
	case "zap":
		// 读取 zap_config
		zapConfig, err := configProvider.Get("logger.zap_config")
		if err != nil {
			return nil, fmt.Errorf("failed to get logger.zap_config: %w", err)
		}
		driverConfig, ok = zapConfig.(map[string]any)
		if !ok {
			return nil, fmt.Errorf("logger.zap_config must be a map, got %T", zapConfig)
		}

	case "none":
		// none 驱动不需要配置
		driverConfig = nil

	default:
		return nil, fmt.Errorf("unsupported driver type: %s (must be zap or none)", driverTypeStr)
	}

	// 3. 调用 Build 函数创建实例
	return Build(driverTypeStr, driverConfig)
}
