// Package telemetrymgr 提供统一的可观测性管理功能，支持 Traces、Metrics、Logs 三大信号。
//
// 核心特性：
//   - 统一接口 - 提供 TelemetryManager 接口，统一管理可观测性组件
//   - 多驱动支持 - 支持 none（空实现）和 otel（OpenTelemetry）两种驱动
//   - 生命周期管理 - 集成 OnStart/OnStop 生命周期钩子，支持优雅关闭
//   - 灵活配置 - 支持从配置提供者或直接配置创建管理器实例
//   - 可观测性集成 - 完整支持链路追踪、指标收集和结构化日志
//
// 基本用法：
//
//	// 使用默认的 none 驱动创建管理器
//	mgr, err := telemetrymgr.Build("none", nil)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	defer mgr.Shutdown(context.Background())
//
//	// 使用 OpenTelemetry 驱动
//	mgr, err = telemetrymgr.Build("otel", map[string]any{
//	    "endpoint": "localhost:4317",
//	    "insecure": true,
//	})
//	if err != nil {
//	    log.Fatal(err)
//	}
//	defer mgr.Shutdown(context.Background())
//
// 驱动类型：
//
// 支持 "none" 和 "otel" 两种驱动：
//   - none: 空实现，不产生任何可观测性数据，适用于不需要遥测的场景
//   - otel: OpenTelemetry 实现，连接到 OTLP 收集器（如 Jaeger、Tempo 等）
//
// 从配置提供者创建：
//
//	// 从配置提供者读取配置
//	mgr, err := telemetrymgr.BuildWithConfigProvider(configProvider)
//	if err != nil {
//	    log.Fatal(err)
//	}
//
// 配置路径：
//   - telemetry.driver: 驱动类型 ("otel", "none")
//   - telemetry.otel_config: OTel 驱动配置（当 driver=otel 时使用）
//
// 使用 Tracer：
//
//	tracer := mgr.Tracer("my-service")
//	ctx, span := tracer.Start(ctx, "operation-name")
//	defer span.End()
//	span.SetAttributes(attribute.String("key", "value"))
//
// 使用 Meter：
//
//	meter := mgr.Meter("my-service")
//	counter, _ := meter.Float64Counter("requests_total")
//	counter.Add(ctx, 1, attribute.String("path", "/api/users"))
//
// 使用 Logger：
//
//	logger := mgr.Logger("my-service")
//	logger.Emit(context.Background(), log.Record{...})
//
// 优雅关闭：
//
//	// OnStop 会自动调用 Shutdown，也可以手动调用
//	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
//	defer cancel()
//	mgr.Shutdown(ctx)
package telemetrymgr
