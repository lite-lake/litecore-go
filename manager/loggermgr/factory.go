package loggermgr

import (
	"fmt"

	"com.litelake.litecore/common"
	"com.litelake.litecore/manager/loggermgr/internal/config"
	"com.litelake.litecore/manager/loggermgr/internal/drivers"
	"com.litelake.litecore/manager/telemetrymgr"
)

// Build 创建日志管理器实例
// cfg: 日志配置内容（包含 telemetry_enabled、console_enabled、file_enabled 等配置项）
// telemetryMgr: 可选的观测管理器，用于发送观测日志
func Build(cfg map[string]any, telemetryMgr telemetrymgr.TelemetryManager) common.Manager {
	// 解析日志配置
	loggerConfig, err := config.ParseLoggerConfigFromMap(cfg)
	if err != nil {
		// 配置解析失败，返回 none 管理器作为降级
		return NewNoneLoggerManagerAdapter(drivers.NewNoneLoggerManager())
	}

	// 验证配置
	if err := loggerConfig.Validate(); err != nil {
		// 配置验证失败，返回 none 管理器作为降级
		return NewNoneLoggerManagerAdapter(drivers.NewNoneLoggerManager())
	}

	// 创建 zap 日志管理器
	mgr, err := drivers.NewZapLoggerManager(loggerConfig, telemetryMgr)
	if err != nil {
		// 初始化失败，降级到 none 管理器
		return NewNoneLoggerManagerAdapter(drivers.NewNoneLoggerManager())
	}

	// 返回适配后的管理器
	return NewLoggerManagerAdapter(mgr)
}

// BuildWithConfig 使用配置结构体创建日志管理器
// loggerConfig: 日志配置结构体
// telemetryMgr: 可选的观测管理器，用于发送观测日志
func BuildWithConfig(loggerConfig *config.LoggerConfig, telemetryMgr telemetrymgr.TelemetryManager) (common.Manager, error) {
	if err := loggerConfig.Validate(); err != nil {
		return nil, fmt.Errorf("invalid logger config: %w", err)
	}

	mgr, err := drivers.NewZapLoggerManager(loggerConfig, telemetryMgr)
	if err != nil {
		return nil, fmt.Errorf("failed to create zap logger manager: %w", err)
	}

	// 返回适配后的管理器
	return NewLoggerManagerAdapter(mgr), nil
}
