// Package telemetrymgr 提供可观测性管理功能，支持分布式链路追踪、指标收集和日志关联。
//
// 核心特性：
//   - OpenTelemetry 集成：基于业界标准的 OpenTelemetry 协议，支持多种后端
//   - 驱动抽象设计：统一接口支持多种实现，当前支持 OTEL 和 None 驱动
//   - 优雅降级机制：配置解析失败或初始化错误时自动降级到空实现
//   - 生命周期管理：完整的启动、健康检查和优雅关闭支持
//   - 线程安全：所有操作均保证并发安全
//
// 基本用法：
//
//	factory := telemetrymgr.NewFactory()
//
//	// 配置 OTEL 驱动
//	config := map[string]any{
//	    "endpoint": "http://localhost:4317",
//	    "resource_attributes": []any{
//	        map[string]any{"key": "service.name", "value": "my-service"},
//	    },
//	    "traces":  map[string]any{"enabled": true},
//	    "metrics": map[string]any{"enabled": true},
//	    "logs":    map[string]any{"enabled": true},
//	}
//
//	mgr := factory.Build("otel", config)
//	mgr.OnStart()
//	defer mgr.OnStop()
//
//	// 使用 Tracer 进行链路追踪
//	if telemetryMgr, ok := mgr.(telemetrymgr.TelemetryManager); ok {
//	    tracer := telemetryMgr.Tracer("my-service")
//	    ctx, span := tracer.Start(context.Background(), "operation")
//	    defer span.End()
//	    // ... 业务逻辑
//	}
//
// 支持的驱动类型：
//
//   - "otel": OpenTelemetry 实现，需要提供 endpoint 等配置
//   - "none": 空实现，适用于不需要遥测功能的场景
//
// 配置说明：
//
// OTEL 驱动配置包含以下字段：
//   - endpoint (string, 必需): OTLP collector 端点地址
//   - insecure (bool, 可选): 是否使用不安全连接，默认 false（使用 TLS）
//   - resource_attributes ([]ResourceAttribute, 可选): 资源属性列表
//   - headers (map[string]string, 可选): 认证请求头
//   - traces (FeatureConfig, 可选): 链路追踪配置
//   - metrics (FeatureConfig, 可选): 指标配置
//   - logs (FeatureConfig, 可选): 日志配置
package telemetrymgr
