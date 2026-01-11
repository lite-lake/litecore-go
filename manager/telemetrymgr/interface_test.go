package telemetrymgr

import (
	"context"
	"testing"

	"go.opentelemetry.io/otel/log"
	"go.opentelemetry.io/otel/metric"
	sdklog "go.opentelemetry.io/otel/sdk/log"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
)

// MockTelemetryManager 是一个实现 TelemetryManager 接口的模拟对象
type MockTelemetryManager struct {
	tracerProvider *sdktrace.TracerProvider
	meterProvider  *sdkmetric.MeterProvider
	loggerProvider *sdklog.LoggerProvider
}

func NewMockTelemetryManager() *MockTelemetryManager {
	return &MockTelemetryManager{
		tracerProvider: sdktrace.NewTracerProvider(),
		meterProvider:  sdkmetric.NewMeterProvider(),
		loggerProvider: sdklog.NewLoggerProvider(),
	}
}

func (m *MockTelemetryManager) Tracer(name string) trace.Tracer {
	return m.tracerProvider.Tracer(name)
}

func (m *MockTelemetryManager) TracerProvider() *sdktrace.TracerProvider {
	return m.tracerProvider
}

func (m *MockTelemetryManager) Meter(name string) metric.Meter {
	return m.meterProvider.Meter(name)
}

func (m *MockTelemetryManager) MeterProvider() *sdkmetric.MeterProvider {
	return m.meterProvider
}

func (m *MockTelemetryManager) Logger(name string) log.Logger {
	return m.loggerProvider.Logger(name)
}

func (m *MockTelemetryManager) LoggerProvider() *sdklog.LoggerProvider {
	return m.loggerProvider
}

func (m *MockTelemetryManager) Shutdown(ctx context.Context) error {
	return nil
}

func TestTelemetryManager_Tracer(t *testing.T) {
	mock := NewMockTelemetryManager()

	tracer := mock.Tracer("test-service")
	if tracer == nil {
		t.Fatal("Tracer() returned nil")
	}

	// 测试 tracer 可以正常使用
	ctx, span := tracer.Start(context.Background(), "test-operation")
	if ctx == nil {
		t.Error("Tracer.Start() returned nil context")
	}
	if span == nil {
		t.Error("Tracer.Start() returned nil span")
	}
	span.End()
}

func TestTelemetryManager_TracerProvider(t *testing.T) {
	mock := NewMockTelemetryManager()

	tp := mock.TracerProvider()
	if tp == nil {
		t.Fatal("TracerProvider() returned nil")
	}
}

func TestTelemetryManager_Meter(t *testing.T) {
	mock := NewMockTelemetryManager()

	meter := mock.Meter("test-service")
	if meter == nil {
		t.Fatal("Meter() returned nil")
	}
}

func TestTelemetryManager_MeterProvider(t *testing.T) {
	mock := NewMockTelemetryManager()

	mp := mock.MeterProvider()
	if mp == nil {
		t.Fatal("MeterProvider() returned nil")
	}
}

func TestTelemetryManager_Logger(t *testing.T) {
	mock := NewMockTelemetryManager()

	logger := mock.Logger("test-service")
	if logger == nil {
		t.Fatal("Logger() returned nil")
	}
}

func TestTelemetryManager_LoggerProvider(t *testing.T) {
	mock := NewMockTelemetryManager()

	lp := mock.LoggerProvider()
	if lp == nil {
		t.Fatal("LoggerProvider() returned nil")
	}
}

func TestTelemetryManager_Shutdown(t *testing.T) {
	mock := NewMockTelemetryManager()

	ctx := context.Background()
	if err := mock.Shutdown(ctx); err != nil {
		t.Errorf("Shutdown() error = %v, want nil", err)
	}
}

func TestTelemetryManager_InterfaceCompliance(t *testing.T) {
	// 测试 MockTelemetryManager 实现了 TelemetryManager 接口
	var _ TelemetryManager = (*MockTelemetryManager)(nil)

	mock := NewMockTelemetryManager()

	// 测试所有方法
	_ = mock.Tracer("test")
	_ = mock.TracerProvider()
	_ = mock.Meter("test")
	_ = mock.MeterProvider()
	_ = mock.Logger("test")
	_ = mock.LoggerProvider()
	_ = mock.Shutdown(context.Background())
}

func TestTelemetryManager_RealImplementations(t *testing.T) {
	// 测试实际的管理器实现了 TelemetryManager 接口
	factory := NewFactory()

	tests := []struct {
		name   string
		driver string
		cfg    map[string]any
	}{
		{
			name:   "none driver",
			driver: "none",
			cfg:    nil,
		},
		{
			name:   "otel driver",
			driver: "otel",
			cfg: map[string]any{
				"endpoint": "http://localhost:4317",
				"traces":   map[string]any{"enabled": false},
				"metrics":  map[string]any{"enabled": false},
				"logs":     map[string]any{"enabled": false},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mgr := factory.Build(tt.driver, tt.cfg)
			if mgr == nil {
				t.Fatal("Build() returned nil manager")
			}
			defer mgr.OnStop()

			tm, ok := mgr.(TelemetryManager)
			if !ok {
				t.Fatal("Manager does not implement TelemetryManager interface")
			}

			// 测试 Tracer 相关方法
			tracer := tm.Tracer("test-service")
			if tracer == nil {
				t.Error("TelemetryManager.Tracer() returned nil")
			}
			if tp := tm.TracerProvider(); tp == nil {
				t.Error("TelemetryManager.TracerProvider() returned nil")
			}

			// 测试 Meter 相关方法
			meter := tm.Meter("test-service")
			if meter == nil {
				t.Error("TelemetryManager.Meter() returned nil")
			}
			if mp := tm.MeterProvider(); mp == nil {
				t.Error("TelemetryManager.MeterProvider() returned nil")
			}

			// 测试 Logger 相关方法
			logger := tm.Logger("test-service")
			if logger == nil {
				t.Error("TelemetryManager.Logger() returned nil")
			}
			if lp := tm.LoggerProvider(); lp == nil {
				t.Error("TelemetryManager.LoggerProvider() returned nil")
			}

			// 测试使用 Tracer
			ctx, span := tracer.Start(context.Background(), "test-operation")
			if ctx == nil {
				t.Error("Tracer.Start() returned nil context")
			}
			if span == nil {
				t.Error("Tracer.Start() returned nil span")
			}
			span.End()

			// 测试 Shutdown 方法
			ctx2 := context.Background()
			if err := tm.Shutdown(ctx2); err != nil {
				t.Errorf("TelemetryManager.Shutdown() error = %v", err)
			}
		})
	}
}
