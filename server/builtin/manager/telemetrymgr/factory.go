package telemetrymgr

import (
	"fmt"
	"strings"

	"github.com/lite-lake/litecore-go/common"
	"github.com/lite-lake/litecore-go/server/builtin/manager/configmgr"
)

// Build 创建观测管理器实例
// driverType: 驱动类型 ("otel", "none")
// driverConfig: 驱动配置 (根据驱动类型不同而不同)
//   - otel: 传递给 parseOtelConfig 的 map[string]any
//   - none: 忽略
//
// 返回 ITelemetryManager 接口实例和可能的错误
func Build(
	driverType string,
	driverConfig map[string]any,
) (ITelemetryManager, error) {
	// 标准化驱动类型（大小写不敏感，去除空格）
	driverType = strings.ToLower(strings.TrimSpace(driverType))

	switch driverType {
	case "otel":
		// 解析 OTEL 配置
		otelConfig, err := parseOtelConfig(driverConfig)
		if err != nil {
			return nil, fmt.Errorf("failed to parse otel configmgr: %w", err)
		}

		// 创建完整的 TelemetryConfig
		config := &TelemetryConfig{
			Driver:     driverType,
			OtelConfig: otelConfig,
		}

		// 验证配置
		if err := config.Validate(); err != nil {
			return nil, fmt.Errorf("invalid configmgr: %w", err)
		}

		// 创建 OTEL 实现
		mgr, err := NewTelemetryManagerOtelImpl(config)
		if err != nil {
			return nil, fmt.Errorf("failed to create otel manager: %w", err)
		}

		return mgr, nil

	case "none":
		mgr := NewTelemetryManagerNoneImpl()
		return mgr, nil

	default:
		return nil, fmt.Errorf("unsupported driver type: %s (must be otel or none)", driverType)
	}
}

// BuildWithConfigProvider 从配置提供者创建观测管理器实例
// 自动从配置提供者读取 telemetry.driver 和对应驱动配置
// 配置路径：
//   - telemetry.driver: 驱动类型 ("otel", "none")
//   - telemetry.otel_config: OTEL 驱动配置（当 driver=otel 时使用）
//
// 返回 ITelemetryManager 接口实例和可能的错误
func BuildWithConfigProvider(configProvider configmgr.IConfigManager) (ITelemetryManager, error) {
	if configProvider == nil {
		return nil, fmt.Errorf("configProvider cannot be nil")
	}

	// 1. 读取驱动类型 telemetry.driver
	driverType, err := configProvider.Get("telemetry.driver")
	if err != nil {
		return nil, fmt.Errorf("failed to get telemetry.driver: %w", err)
	}

	driverTypeStr, err := common.GetString(driverType)
	if err != nil {
		return nil, fmt.Errorf("telemetry.driver: %w", err)
	}

	// 标准化驱动类型（大小写不敏感，去除空格）
	driverTypeStr = strings.ToLower(strings.TrimSpace(driverTypeStr))

	// 2. 根据驱动类型读取对应配置
	var driverConfig map[string]any

	switch driverTypeStr {
	case "otel":
		// 读取 otel_config
		otelConfig, err := configProvider.Get("telemetry.otel_config")
		if err != nil {
			return nil, fmt.Errorf("failed to get telemetry.otel_config: %w", err)
		}
		driverConfig, err = common.GetMap(otelConfig)
		if err != nil {
			return nil, fmt.Errorf("telemetry.otel_config: %w", err)
		}

	case "none":
		// none 驱动不需要配置
		driverConfig = nil

	default:
		return nil, fmt.Errorf("unsupported driver type: %s (must be otel or none)", driverTypeStr)
	}

	// 3. 调用 Build 函数创建实例
	return Build(driverTypeStr, driverConfig)
}
