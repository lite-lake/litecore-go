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

// telemetryManagerNoneImpl 空观测管理器实现
// 在不需要遥测功能时使用，提供空实现以避免条件判断
type telemetryManagerNoneImpl struct {
	*telemetryManagerBaseImpl
	tracerProvider *sdktrace.TracerProvider
	meterProvider  *sdkmetric.MeterProvider
	loggerProvider *sdklog.LoggerProvider
}

// NewTelemetryManagerNoneImpl 创建空观测管理器实现
func NewTelemetryManagerNoneImpl() TelemetryManager {
	return &telemetryManagerNoneImpl{
		telemetryManagerBaseImpl: newTelemetryManagerBaseImpl("none-telemetry"),
		tracerProvider:           sdktrace.NewTracerProvider(),
		meterProvider:            sdkmetric.NewMeterProvider(),
		loggerProvider:           sdklog.NewLoggerProvider(),
	}
}

// Tracer 获取 Tracer 实例
// 返回 NoOpTracer，不会产生任何追踪数据
func (m *telemetryManagerNoneImpl) Tracer(name string) trace.Tracer {
	return m.tracerProvider.Tracer(name)
}

// TracerProvider 获取 TracerProvider
func (m *telemetryManagerNoneImpl) TracerProvider() *sdktrace.TracerProvider {
	return m.tracerProvider
}

// Meter 获取 Meter 实例
// 返回 NoOpMeter，不会产生任何指标数据
func (m *telemetryManagerNoneImpl) Meter(name string) metric.Meter {
	return m.meterProvider.Meter(name)
}

// MeterProvider 获取 MeterProvider
func (m *telemetryManagerNoneImpl) MeterProvider() *sdkmetric.MeterProvider {
	return m.meterProvider
}

// Logger 获取 Logger 实例
// 返回 NoOpLogger，不会产生任何日志数据
func (m *telemetryManagerNoneImpl) Logger(name string) log.Logger {
	return m.loggerProvider.Logger(name)
}

// LoggerProvider 获取 LoggerProvider
func (m *telemetryManagerNoneImpl) LoggerProvider() *sdklog.LoggerProvider {
	return m.loggerProvider
}

// Shutdown 关闭观测管理器
// 空实现，无需清理资源
func (m *telemetryManagerNoneImpl) Shutdown(ctx context.Context) error {
	return nil
}

// Close 关闭观测管理器（别名方法）
func (m *telemetryManagerNoneImpl) Close() error {
	return nil
}

// 确保实现 TelemetryManager 接口
var _ TelemetryManager = (*telemetryManagerNoneImpl)(nil)
