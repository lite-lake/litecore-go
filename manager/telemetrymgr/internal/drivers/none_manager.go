package drivers

import (
	"context"

	"com.litelake.litecore/common"
	"go.opentelemetry.io/otel/log"
	"go.opentelemetry.io/otel/metric"
	sdklog "go.opentelemetry.io/otel/sdk/log"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
)

// NoneManager 空观测管理器
// 在不需要遥测功能时使用，提供空实现以避免条件判断
type NoneManager struct {
	*BaseManager
	tracerProvider *sdktrace.TracerProvider
	meterProvider  *sdkmetric.MeterProvider
	loggerProvider *sdklog.LoggerProvider
}

// NewNoneManager 创建空观测管理器
func NewNoneManager() *NoneManager {
	return &NoneManager{
		BaseManager:    NewBaseManager("none-telemetry"),
		tracerProvider: sdktrace.NewTracerProvider(),
		meterProvider:  sdkmetric.NewMeterProvider(),
		loggerProvider: sdklog.NewLoggerProvider(),
	}
}

// Tracer 获取 Tracer 实例
// 返回 NoOpTracer，不会产生任何追踪数据
func (m *NoneManager) Tracer(name string) trace.Tracer {
	return m.tracerProvider.Tracer(name)
}

// TracerProvider 获取 TracerProvider
func (m *NoneManager) TracerProvider() *sdktrace.TracerProvider {
	return m.tracerProvider
}

// Meter 获取 Meter 实例
// 返回 NoOpMeter，不会产生任何指标数据
func (m *NoneManager) Meter(name string) metric.Meter {
	return m.meterProvider.Meter(name)
}

// MeterProvider 获取 MeterProvider
func (m *NoneManager) MeterProvider() *sdkmetric.MeterProvider {
	return m.meterProvider
}

// Logger 获取 Logger 实例
// 返回 NoOpLogger，不会产生任何日志数据
func (m *NoneManager) Logger(name string) log.Logger {
	return m.loggerProvider.Logger(name)
}

// LoggerProvider 获取 LoggerProvider
func (m *NoneManager) LoggerProvider() *sdklog.LoggerProvider {
	return m.loggerProvider
}

// Shutdown 关闭观测管理器
// 空实现，无需清理资源
func (m *NoneManager) Shutdown(ctx context.Context) error {
	return nil
}

// ensure NoneManager implements common.Manager interface
var _ common.Manager = (*NoneManager)(nil)

// ensure NoneManager implements telemetrymgr.TelemetryManager interface
var _ interface {
	Tracer(name string) trace.Tracer
	TracerProvider() *sdktrace.TracerProvider
	Meter(name string) metric.Meter
	MeterProvider() *sdkmetric.MeterProvider
	Logger(name string) log.Logger
	LoggerProvider() *sdklog.LoggerProvider
	Shutdown(ctx context.Context) error
} = (*NoneManager)(nil)
