package telemetrymgr

import (
	"fmt"

	"com.litelake.litecore/manager/telemetrymgr/internal/config"
	"com.litelake.litecore/manager/telemetrymgr/internal/drivers"
)

// Factory 观测管理器工厂
//
// Deprecated: Factory 模式已废弃，请使用依赖注入模式替代。
// 使用 NewManager() 创建 Manager 实例，并通过 Container 注入依赖。
// 迁移示例：
//
//	旧方式：
//	  factory := telemetrymgr.NewFactory()
//	  mgr := factory.Build("otel", cfg)
//
//	新方式：
//	  mgr := telemetrymgr.NewManager("default")
//	  container.Register(mgr)
//	  container.InjectAll()
//	  mgr.OnStart()
type Factory struct{}

// NewFactory 创建观测管理器工厂
//
// Deprecated: 请使用 NewManager() 替代
func NewFactory() *Factory {
	return &Factory{}
}

// Build 创建观测管理器实例
//
// Deprecated: 请使用依赖注入模式替代
// driver: 驱动类型，支持 "none", "otel"
// cfg: 驱动专属的配置数据
//   - driver="otel": cfg 是 OTEL 配置内容 (endpoint, traces, metrics 等)
//   - driver="none": cfg 可以为 nil 或空
// 返回 TelemetryManager 接口，可直接使用 Tracer/Meter/Logger 等方法
func (f *Factory) Build(driver string, cfg map[string]any) TelemetryManager {
	switch driver {
	case "otel":
		// 解析 OTEL 配置
		otelConfig, err := config.ParseOtelConfigFromMap(cfg)
		if err != nil {
			// 配置解析失败，返回 none 驱动作为降级
			return drivers.NewNoneManager()
		}

		// 创建 TelemetryConfig
		telemetryConfig := &config.TelemetryConfig{
			Driver:     driver,
			OtelConfig: otelConfig,
		}

		// 验证配置
		if err := telemetryConfig.Validate(); err != nil {
			// 配置验证失败，返回 none 驱动作为降级
			return drivers.NewNoneManager()
		}

		// 创建 OTEL 管理器
		mgr, err := drivers.NewOtelManager(telemetryConfig)
		if err != nil {
			// OTEL 初始化失败，降级到 none 驱动
			return drivers.NewNoneManager()
		}
		return mgr

	case "none":
		// none 驱动无需配置
		return drivers.NewNoneManager()

	default:
		// 未知驱动类型，返回 none 驱动作为降级
		return drivers.NewNoneManager()
	}
}

// BuildWithConfig 使用配置结构体创建观测管理器
//
// Deprecated: 请使用依赖注入模式替代
// 返回 TelemetryManager 接口，可直接使用 Tracer/Meter/Logger 等方法
func (f *Factory) BuildWithConfig(telemetryConfig *config.TelemetryConfig) (TelemetryManager, error) {
	if err := telemetryConfig.Validate(); err != nil {
		return nil, fmt.Errorf("invalid telemetry config: %w", err)
	}

	switch telemetryConfig.Driver {
	case "otel":
		mgr, err := drivers.NewOtelManager(telemetryConfig)
		if err != nil {
			return nil, fmt.Errorf("failed to create otel manager: %w", err)
		}
		return mgr, nil
	case "none":
		return drivers.NewNoneManager(), nil
	default:
		return nil, fmt.Errorf("unsupported telemetry driver: %s", telemetryConfig.Driver)
	}
}
