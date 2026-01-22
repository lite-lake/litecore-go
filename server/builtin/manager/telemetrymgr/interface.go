package telemetrymgr

import (
	"context"

	"github.com/lite-lake/litecore-go/common"
	"go.opentelemetry.io/otel/log"
	"go.opentelemetry.io/otel/metric"
	sdklog "go.opentelemetry.io/otel/sdk/log"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
)

// ITelemetryManager 观测管理器接口
// 统一提供 Traces、Metrics、Logs 三大观测能力
type ITelemetryManager interface {
	common.IBaseManager

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
