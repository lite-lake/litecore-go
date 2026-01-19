package telemetrymgr

import (
	"context"
	"fmt"
	"sync"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/log"
	"go.opentelemetry.io/otel/metric"
	sdklog "go.opentelemetry.io/otel/sdk/log"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.24.0"
	"go.opentelemetry.io/otel/trace"
)

// telemetryManagerOtelImpl OpenTelemetry 观测管理器实现
type telemetryManagerOtelImpl struct {
	*telemetryManagerBaseImpl
	config         *TelemetryConfig
	tracerProvider *sdktrace.TracerProvider
	meterProvider  *sdkmetric.MeterProvider
	loggerProvider *sdklog.LoggerProvider
	resource       *resource.Resource
	mu             sync.RWMutex
	shutdownFuncs  []func(context.Context) error
	shutdownOnce   sync.Once
}

// NewTelemetryManagerOtelImpl 创建 OTEL 观测管理器实现
func NewTelemetryManagerOtelImpl(cfg *TelemetryConfig) (TelemetryManager, error) {
	if cfg.Driver != "otel" {
		return nil, fmt.Errorf("invalid driver for otel manager: %s", cfg.Driver)
	}

	ctx := context.Background()
	mgr := &telemetryManagerOtelImpl{
		telemetryManagerBaseImpl: newTelemetryManagerBaseImpl("otel-telemetry"),
		config:                   cfg,
		shutdownFuncs:            make([]func(context.Context) error, 0),
	}

	// 初始化资源
	if err := mgr.initResource(ctx); err != nil {
		return nil, fmt.Errorf("failed to initialize resource: %w", err)
	}

	// 初始化 TracerProvider
	if err := mgr.initTracerProvider(ctx); err != nil {
		return nil, fmt.Errorf("failed to initialize tracer provider: %w", err)
	}

	// 初始化 MeterProvider
	if err := mgr.initMeterProvider(ctx); err != nil {
		return nil, fmt.Errorf("failed to initialize meter provider: %w", err)
	}

	// 初始化 LoggerProvider
	if err := mgr.initLoggerProvider(ctx); err != nil {
		return nil, fmt.Errorf("failed to initialize logger provider: %w", err)
	}

	return mgr, nil
}

// initResource 初始化 OTEL 资源
func (m *telemetryManagerOtelImpl) initResource(ctx context.Context) error {
	// 构建资源属性
	attrs := []attribute.KeyValue{
		semconv.ServiceNameKey.String("litecore-app"),
	}

	// 添加配置中的资源属性
	for _, ra := range m.config.OtelConfig.ResourceAttributes {
		attrs = append(attrs, attribute.String(ra.Key, ra.Value))
	}

	// 创建资源
	res, err := resource.New(
		ctx,
		resource.WithAttributes(attrs...),
		resource.WithSchemaURL(semconv.SchemaURL),
	)
	if err != nil {
		return fmt.Errorf("failed to create resource: %w", err)
	}

	m.mu.Lock()
	m.resource = res
	m.mu.Unlock()

	return nil
}

// initTracerProvider 初始化 TracerProvider
func (m *telemetryManagerOtelImpl) initTracerProvider(ctx context.Context) error {
	if m.config.OtelConfig.Traces == nil || !m.config.OtelConfig.Traces.Enabled {
		// 如果未启用链路追踪，使用 NoOpProvider
		m.mu.Lock()
		m.tracerProvider = sdktrace.NewTracerProvider()
		m.mu.Unlock()
		return nil
	}

	// 构建 OTLP exporter 选项
	opts := []otlptracegrpc.Option{
		otlptracegrpc.WithEndpoint(m.config.OtelConfig.Endpoint),
	}

	// 配置安全连接
	if m.config.OtelConfig.Insecure {
		opts = append(opts, otlptracegrpc.WithInsecure())
	}
	// 默认使用安全连接 (TLS)

	// 配置请求头（用于认证）
	if len(m.config.OtelConfig.Headers) > 0 {
		opts = append(opts, otlptracegrpc.WithHeaders(m.config.OtelConfig.Headers))
	}

	// 创建 OTLP exporter
	exporter, err := otlptracegrpc.New(ctx, opts...)
	if err != nil {
		return fmt.Errorf("failed to create OTLP trace exporter: %w", err)
	}

	m.mu.Lock()
	// 记录 shutdown 函数
	m.shutdownFuncs = append(m.shutdownFuncs, exporter.Shutdown)

	// 创建 TracerProvider
	m.tracerProvider = sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(m.resource),
	)

	// 设置全局 TracerProvider
	otel.SetTracerProvider(m.tracerProvider)

	// 记录 shutdown 函数
	m.shutdownFuncs = append(m.shutdownFuncs, m.tracerProvider.Shutdown)
	m.mu.Unlock()

	return nil
}

// initMeterProvider 初始化 MeterProvider
func (m *telemetryManagerOtelImpl) initMeterProvider(ctx context.Context) error {
	// 确保 Metrics 字段已初始化
	if m.config.OtelConfig.Metrics == nil {
		m.config.OtelConfig.Metrics = &FeatureConfig{Enabled: false}
	}

	if !m.config.OtelConfig.Metrics.Enabled {
		// 如果未启用指标，使用 noop provider
		m.mu.Lock()
		m.meterProvider = sdkmetric.NewMeterProvider()
		m.mu.Unlock()
		return nil
	}

	// TODO: 实现 OTLP metrics exporter
	// 当前先使用 noop provider
	m.mu.Lock()
	m.meterProvider = sdkmetric.NewMeterProvider()
	m.mu.Unlock()

	return nil
}

// initLoggerProvider 初始化 LoggerProvider
func (m *telemetryManagerOtelImpl) initLoggerProvider(ctx context.Context) error {
	// 确保 Logs 字段已初始化
	if m.config.OtelConfig.Logs == nil {
		m.config.OtelConfig.Logs = &FeatureConfig{Enabled: false}
	}

	if !m.config.OtelConfig.Logs.Enabled {
		// 如果未启用日志，使用 noop provider
		m.mu.Lock()
		m.loggerProvider = sdklog.NewLoggerProvider()
		m.mu.Unlock()
		return nil
	}

	// TODO: 实现 OTLP logs exporter
	// 当前先使用 noop provider
	m.mu.Lock()
	m.loggerProvider = sdklog.NewLoggerProvider()
	m.mu.Unlock()

	return nil
}

// Tracer 获取 Tracer 实例
func (m *telemetryManagerOtelImpl) Tracer(name string) trace.Tracer {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if m.tracerProvider == nil {
		return trace.NewNoopTracerProvider().Tracer(name)
	}
	return m.tracerProvider.Tracer(name)
}

// TracerProvider 获取 TracerProvider
func (m *telemetryManagerOtelImpl) TracerProvider() *sdktrace.TracerProvider {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.tracerProvider
}

// Meter 获取 Meter 实例
func (m *telemetryManagerOtelImpl) Meter(name string) metric.Meter {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if m.meterProvider == nil {
		return sdkmetric.NewMeterProvider().Meter(name)
	}
	return m.meterProvider.Meter(name)
}

// MeterProvider 获取 MeterProvider
func (m *telemetryManagerOtelImpl) MeterProvider() *sdkmetric.MeterProvider {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.meterProvider
}

// Logger 获取 Logger 实例
func (m *telemetryManagerOtelImpl) Logger(name string) log.Logger {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if m.loggerProvider == nil {
		return sdklog.NewLoggerProvider().Logger(name)
	}
	return m.loggerProvider.Logger(name)
}

// LoggerProvider 获取 LoggerProvider
func (m *telemetryManagerOtelImpl) LoggerProvider() *sdklog.LoggerProvider {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.loggerProvider
}

// Health 检查管理器健康状态
func (m *telemetryManagerOtelImpl) Health() error {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// 检查资源是否正确初始化
	if m.resource == nil {
		return fmt.Errorf("resource is not initialized")
	}

	// 检查 TracerProvider 是否正常工作
	if m.tracerProvider == nil {
		return fmt.Errorf("tracer provider is not initialized")
	}

	// TODO: 可以添加 exporter 连接状态检查
	// 例如: 发送一个测试 span 来验证连接

	return nil
}

// OnStart 在服务器启动时触发
func (m *telemetryManagerOtelImpl) OnStart() error {
	// OTEL 管理器在创建时就已经初始化完成
	// 这里可以添加启动时的额外逻辑
	return nil
}

// OnStop 在服务器停止时触发
// 使用默认的超时时间
func (m *telemetryManagerOtelImpl) OnStop() error {
	// 触发优雅关闭
	ctx, cancel := context.WithTimeout(context.Background(), 30)
	defer cancel()
	return m.Shutdown(ctx)
}

// Shutdown 关闭观测管理器
// 使用 sync.Once 确保只关闭一次
func (m *telemetryManagerOtelImpl) Shutdown(ctx context.Context) error {
	var shutdownErr error

	m.shutdownOnce.Do(func() {
		m.mu.Lock()
		defer m.mu.Unlock()

		// 按相反顺序执行 shutdown 函数（类似 defer）
		for i := len(m.shutdownFuncs) - 1; i >= 0; i-- {
			if err := m.shutdownFuncs[i](ctx); err != nil {
				shutdownErr = fmt.Errorf("shutdown error: %w", err)
			}
		}
	})

	return shutdownErr
}

// Close 关闭观测管理器（别名方法）
func (m *telemetryManagerOtelImpl) Close() error {
	ctx, cancel := context.WithTimeout(context.Background(), 30)
	defer cancel()
	return m.Shutdown(ctx)
}

// 确保实现 TelemetryManager 接口
var _ TelemetryManager = (*telemetryManagerOtelImpl)(nil)
