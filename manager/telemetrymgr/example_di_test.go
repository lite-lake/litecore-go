package telemetrymgr_test

import (
	"context"
	"fmt"

	"com.litelake.litecore/manager/telemetrymgr"
)

// Example_diPattern 演示依赖注入模式的使用方法
func Example_diPattern() {
	// 1. 创建 Manager 实例
	mgr := telemetrymgr.NewManager("default")

	// 2. 设置配置提供者（通过 Container 注入）
	// 在实际应用中，这个步骤由 Container.InjectAll() 自动完成
	// mgr.Config = configProvider

	// 3. 启动管理器（初始化驱动）
	// 在实际应用中，这个步骤由 Container 的生命周期管理自动完成
	if err := mgr.OnStart(); err != nil {
		fmt.Printf("Failed to start manager: %v\n", err)
		return
	}

	// 4. 使用观测组件
	tracer := mgr.Tracer("example")
	meter := mgr.Meter("example")
	logger := mgr.Logger("example")

	// 使用 Tracer
	ctx := context.Background()
	_, span := tracer.Start(ctx, "example-operation")
	span.End()

	// 使用 Meter
	_ = meter

	// 使用 Logger
	_ = logger

	// 5. 停止管理器（优雅关闭）
	if err := mgr.OnStop(); err != nil {
		fmt.Printf("Failed to stop manager: %v\n", err)
		return
	}

	fmt.Println("Manager started and stopped successfully")
	// Output: Manager started and stopped successfully
}

// Example_diPatternWithConfig 演示带配置的依赖注入模式
func Example_diPatternWithConfig() {
	// 创建配置提供者
	// configProvider := NewMockConfigProvider(map[string]any{
	//     "telemetry.default": map[string]any{
	//         "driver": "otel",
	//         "otel_config": map[string]any{
	//             "endpoint": "localhost:4317",
	//             "insecure": true,
	//             "traces": map[string]any{
	//                 "enabled": true,
	//             },
	//             "metrics": map[string]any{
	//                 "enabled": false,
	//             },
	//             "logs": map[string]any{
	//                 "enabled": false,
	//             },
	//         },
	//     },
	// })

	// 创建 Manager
	mgr := telemetrymgr.NewManager("default")

	// 在实际应用中，通过 Container 注入配置
	// mgr.Config = configProvider

	// 启动管理器
	if err := mgr.OnStart(); err != nil {
		fmt.Printf("Failed to start manager: %v\n", err)
		return
	}

	// 验证健康状态
	if err := mgr.Health(); err != nil {
		fmt.Printf("Manager health check failed: %v\n", err)
		return
	}

	fmt.Println("Manager is healthy")

	// 停止管理器
	if err := mgr.OnStop(); err != nil {
		fmt.Printf("Failed to stop manager: %v\n", err)
		return
	}

	// Output: Manager is healthy
}

// Example_migrationFromFactory 演示从 Factory 模式迁移到 DI 模式
func Example_migrationFromFactory() {
	// ===== 旧方式（Factory 模式）- 已废弃 =====
	// factory := telemetrymgr.NewFactory()
	// mgr := factory.Build("otel", cfg)

	// ===== 新方式（依赖注入模式）=====
	// 1. 创建 Manager
	mgr := telemetrymgr.NewManager("default")

	// 2. 通过 Container 注入配置
	// container.Register(mgr)
	// container.InjectAll()  // 自动注入 Config

	// 3. 启动 Manager
	// mgr.OnStart()  // 从 Config 加载配置并初始化驱动

	// 4. 使用 Manager
	tracer := mgr.Tracer("example")
	_ = tracer

	// 5. 停止 Manager
	// mgr.OnStop()

	fmt.Println("Migration successful")
	// Output: Migration successful
}
