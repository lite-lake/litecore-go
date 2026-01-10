package telemetrymgr_test

import (
	"context"
	"fmt"

	"com.litelake.litecore/manager/telemetrymgr"
)

// Example_basicUsage 展示基本用法
func Example_basicUsage() {
	factory := telemetrymgr.NewFactory()

	// 创建 OTEL 驱动
	// 配置直接是 OTEL 的配置内容，不需要包含 driver 字段
	otelCfg := map[string]any{
		"endpoint": "http://localhost:4317",
		"resource_attributes": []any{
			map[string]any{"key": "service.name", "value": "my-service"},
			map[string]any{"key": "environment", "value": "production"},
		},
		"traces":  map[string]any{"enabled": true},
		"metrics": map[string]any{"enabled": true},
		"logs":    map[string]any{"enabled": true},
	}

	mgr := factory.Build("otel", otelCfg)
	mgr.OnStart()
	defer mgr.OnStop()

	fmt.Println("Telemetry manager started:", mgr.ManagerName())
	// Output: Telemetry manager started: otel-telemetry
}

// Example_noneDriver 展示 none 驱动用法
func Example_noneDriver() {
	factory := telemetrymgr.NewFactory()

	// none 驱动不需要配置，可以传 nil
	mgr := factory.Build("none", nil)

	mgr.OnStart()
	defer mgr.OnStop()

	fmt.Println("Telemetry manager started:", mgr.ManagerName())
	// Output: Telemetry manager started: none-telemetry
}

// Example_withTracer 展示如何使用 Tracer
func Example_withTracer() {
	factory := telemetrymgr.NewFactory()

	// 创建 OTEL 驱动
	otelCfg := map[string]any{
		"endpoint": "http://localhost:4317",
		"traces":   map[string]any{"enabled": true},
	}

	mgr := factory.Build("otel", otelCfg)
	mgr.OnStart()
	defer mgr.OnStop()

	// 使用 Tracer
	if telemetryMgr, ok := mgr.(telemetrymgr.TelemetryManager); ok {
		tracer := telemetryMgr.Tracer("my-service")
		ctx, span := tracer.Start(context.Background(), "example-operation")
		defer span.End()

		// 在这里执行业务逻辑
		_ = ctx
	}

	fmt.Println("Tracer created successfully")
}

// Example_withEmptyConfig 展示 OTEL 驱动使用空配置（会使用默认值）
func Example_withEmptyConfig() {
	factory := telemetrymgr.NewFactory()

	// OTEL 驱动使用空配置
	// 注意：这会导致验证失败，因为 endpoint 是必需的
	// 所以会降级到 none 驱动
	mgr := factory.Build("otel", nil)

	fmt.Println("Manager:", mgr.ManagerName())
	// Output: Manager: none-telemetry
}
