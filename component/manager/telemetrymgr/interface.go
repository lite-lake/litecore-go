package telemetrymgr

import (
	"context"

	"go.opentelemetry.io/otel/log"
	"go.opentelemetry.io/otel/metric"
	sdklog "go.opentelemetry.io/otel/sdk/log"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
)

// TelemetryManager 观测管理器接口
// 统一提供 Traces、Metrics、Logs 三大观测能力
type TelemetryManager interface {
	// ========== 生命周期管理（符合 BaseManager 接口） ==========
	// ManagerName 返回管理器名称
	ManagerName() string

	// Health 检查管理器健康状态
	Health() error

	// OnStart 在服务器启动时触发
	OnStart() error

	// OnStop 在服务器停止时触发
	OnStop() error

	// ========== Tracing ==========
	// Tracer 获取 Tracer 实例
	Tracer(name string) trace.Tracer

	// TracerProvider 获取 TracerProvider
	TracerProvider() *sdktrace.TracerProvider

	// ========== Metrics ==========
	// Meter 获取 Meter 实例
	Meter(name string) metric.Meter

	// MeterProvider 获取 MeterProvider
	MeterProvider() *sdkmetric.MeterProvider

	// ========== Logging ==========
	// Logger 获取 Logger 实例
	Logger(name string) log.Logger

	// LoggerProvider 获取 LoggerProvider
	LoggerProvider() *sdklog.LoggerProvider

	// ========== 生命周期 ==========
	// Shutdown 关闭观测管理器，刷新所有待处理的数据
	Shutdown(ctx context.Context) error
}
