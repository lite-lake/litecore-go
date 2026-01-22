package telemetrymgr

import (
	"context"
	"sync"
	"testing"
)

// TestNewTelemetryManagerOtelImpl 测试创建 OTel 实现
func TestNewTelemetryManagerOtelImpl(t *testing.T) {
	tests := []struct {
		name    string
		config  *TelemetryConfig
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid otel configmgr",
			config: &TelemetryConfig{
				Driver: "otel",
				OtelConfig: &OtelConfig{
					Endpoint: "localhost:4317",
					Traces:   &FeatureConfig{Enabled: false},
					Metrics:  &FeatureConfig{Enabled: false},
					Logs:     &FeatureConfig{Enabled: false},
				},
			},
			wantErr: false,
		},
		{
			name: "valid otel configmgr with traces enabled",
			config: &TelemetryConfig{
				Driver: "otel",
				OtelConfig: &OtelConfig{
					Endpoint: "otel-collector:4317",
					Insecure: true,
					Traces:   &FeatureConfig{Enabled: true},
					Metrics:  &FeatureConfig{Enabled: false},
					Logs:     &FeatureConfig{Enabled: false},
				},
			},
			wantErr: false,
		},
		{
			name: "invalid driver for otel manager",
			config: &TelemetryConfig{
				Driver: "none",
				OtelConfig: &OtelConfig{
					Endpoint: "localhost:4317",
				},
			},
			wantErr: true,
			errMsg:  "invalid driver for otel manager",
		},
		{
			name: "missing endpoint uses default",
			config: &TelemetryConfig{
				Driver: "otel",
				OtelConfig: &OtelConfig{
					Endpoint: "", // 空端点会被忽略，代码使用它但不验证
					Traces:   &FeatureConfig{Enabled: false},
					Metrics:  &FeatureConfig{Enabled: false},
					Logs:     &FeatureConfig{Enabled: false},
				},
			},
			wantErr: false, // NewTelemetryManagerOtelImpl 不验证配置，假设已验证
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mgr, err := NewTelemetryManagerOtelImpl(tt.config)
			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error, got nil")
				}
				if tt.errMsg != "" && err != nil {
					if len(err.Error()) < len(tt.errMsg) || err.Error()[:len(tt.errMsg)] != tt.errMsg {
						t.Errorf("expected error containing '%s', got '%s'", tt.errMsg, err.Error())
					}
				}
				return
			}

			if err != nil {
				t.Errorf("expected no error, got %v", err)
				return
			}

			if mgr == nil {
				t.Fatal("expected non-nil manager")
			}

			// 验证类型
			if _, ok := mgr.(*telemetryManagerOtelImpl); !ok {
				t.Error("expected *telemetryManagerOtelImpl type")
			}

			// 清理资源
			mgr.Shutdown(context.Background())
		})
	}
}

// TestITelemetryManagerOtelImpl_ManagerName 测试管理器名称
func TestITelemetryManagerOtelImpl_ManagerName(t *testing.T) {
	config := &TelemetryConfig{
		Driver: "otel",
		OtelConfig: &OtelConfig{
			Endpoint: "localhost:4317",
			Traces:   &FeatureConfig{Enabled: false},
			Metrics:  &FeatureConfig{Enabled: false},
			Logs:     &FeatureConfig{Enabled: false},
		},
	}

	mgr, err := NewTelemetryManagerOtelImpl(config)
	if err != nil {
		t.Fatalf("failed to create manager: %v", err)
	}
	defer mgr.Shutdown(context.Background())

	name := mgr.ManagerName()
	expected := "otel-telemetry"

	if name != expected {
		t.Errorf("expected manager name '%s', got '%s'", expected, name)
	}
}

// TestITelemetryManagerOtelImpl_Health 测试健康检查
func TestITelemetryManagerOtelImpl_Health(t *testing.T) {
	config := &TelemetryConfig{
		Driver: "otel",
		OtelConfig: &OtelConfig{
			Endpoint: "localhost:4317",
			Traces:   &FeatureConfig{Enabled: false},
			Metrics:  &FeatureConfig{Enabled: false},
			Logs:     &FeatureConfig{Enabled: false},
		},
	}

	mgr, err := NewTelemetryManagerOtelImpl(config)
	if err != nil {
		t.Fatalf("failed to create manager: %v", err)
	}
	defer mgr.Shutdown(context.Background())

	err = mgr.Health()
	if err != nil {
		t.Errorf("expected no error from Health, got %v", err)
	}
}

// TestITelemetryManagerOtelImpl_Lifecycle 测试生命周期方法
func TestITelemetryManagerOtelImpl_Lifecycle(t *testing.T) {
	config := &TelemetryConfig{
		Driver: "otel",
		OtelConfig: &OtelConfig{
			Endpoint: "localhost:4317",
			Traces:   &FeatureConfig{Enabled: false},
			Metrics:  &FeatureConfig{Enabled: false},
			Logs:     &FeatureConfig{Enabled: false},
		},
	}

	mgr, err := NewTelemetryManagerOtelImpl(config)
	if err != nil {
		t.Fatalf("failed to create manager: %v", err)
	}

	// 测试 OnStart
	if err := mgr.OnStart(); err != nil {
		t.Errorf("expected no error from OnStart, got %v", err)
	}

	// 测试 Health
	if err := mgr.Health(); err != nil {
		t.Errorf("expected no error from Health, got %v", err)
	}

	// 测试 OnStop
	if err := mgr.OnStop(); err != nil {
		t.Errorf("expected no error from OnStop, got %v", err)
	}
}

// TestITelemetryManagerOtelImpl_Providers 测试 Provider 方法
func TestITelemetryManagerOtelImpl_Providers(t *testing.T) {
	config := &TelemetryConfig{
		Driver: "otel",
		OtelConfig: &OtelConfig{
			Endpoint: "localhost:4317",
			Traces:   &FeatureConfig{Enabled: false},
			Metrics:  &FeatureConfig{Enabled: false},
			Logs:     &FeatureConfig{Enabled: false},
		},
	}

	mgr, err := NewTelemetryManagerOtelImpl(config)
	if err != nil {
		t.Fatalf("failed to create manager: %v", err)
	}
	defer mgr.Shutdown(context.Background())

	t.Run("TracerProvider", func(t *testing.T) {
		tp := mgr.TracerProvider()
		if tp == nil {
			t.Error("expected non-nil TracerProvider")
		}
	})

	t.Run("MeterProvider", func(t *testing.T) {
		mp := mgr.MeterProvider()
		if mp == nil {
			t.Error("expected non-nil MeterProvider")
		}
	})

	t.Run("LoggerProvider", func(t *testing.T) {
		lp := mgr.LoggerProvider()
		if lp == nil {
			t.Error("expected non-nil LoggerProvider")
		}
	})
}

// TestITelemetryManagerOtelImpl_Tracer 测试 Tracer 方法
func TestITelemetryManagerOtelImpl_Tracer(t *testing.T) {
	config := &TelemetryConfig{
		Driver: "otel",
		OtelConfig: &OtelConfig{
			Endpoint: "localhost:4317",
			Traces:   &FeatureConfig{Enabled: false},
			Metrics:  &FeatureConfig{Enabled: false},
			Logs:     &FeatureConfig{Enabled: false},
		},
	}

	mgr, err := NewTelemetryManagerOtelImpl(config)
	if err != nil {
		t.Fatalf("failed to create manager: %v", err)
	}
	defer mgr.Shutdown(context.Background())

	tracer := mgr.Tracer("test-tracer")
	if tracer == nil {
		t.Error("expected non-nil tracer")
	}

	// 测试不同名称的 tracer
	names := []string{"", "test", "tracer-with-dash", "tracer_with_underscore"}
	for _, name := range names {
		tracer := mgr.Tracer(name)
		if tracer == nil {
			t.Errorf("expected non-nil tracer for name '%s'", name)
		}
	}
}

// TestITelemetryManagerOtelImpl_Meter 测试 Meter 方法
func TestITelemetryManagerOtelImpl_Meter(t *testing.T) {
	config := &TelemetryConfig{
		Driver: "otel",
		OtelConfig: &OtelConfig{
			Endpoint: "localhost:4317",
			Traces:   &FeatureConfig{Enabled: false},
			Metrics:  &FeatureConfig{Enabled: false},
			Logs:     &FeatureConfig{Enabled: false},
		},
	}

	mgr, err := NewTelemetryManagerOtelImpl(config)
	if err != nil {
		t.Fatalf("failed to create manager: %v", err)
	}
	defer mgr.Shutdown(context.Background())

	meter := mgr.Meter("test-meter")
	if meter == nil {
		t.Error("expected non-nil meter")
	}

	// 测试不同名称的 meter
	names := []string{"", "test", "meter-with-dash", "meter_with_underscore"}
	for _, name := range names {
		meter := mgr.Meter(name)
		if meter == nil {
			t.Errorf("expected non-nil meter for name '%s'", name)
		}
	}
}

// TestITelemetryManagerOtelImpl_Logger 测试 Logger 方法
func TestITelemetryManagerOtelImpl_Logger(t *testing.T) {
	config := &TelemetryConfig{
		Driver: "otel",
		OtelConfig: &OtelConfig{
			Endpoint: "localhost:4317",
			Traces:   &FeatureConfig{Enabled: false},
			Metrics:  &FeatureConfig{Enabled: false},
			Logs:     &FeatureConfig{Enabled: false},
		},
	}

	mgr, err := NewTelemetryManagerOtelImpl(config)
	if err != nil {
		t.Fatalf("failed to create manager: %v", err)
	}
	defer mgr.Shutdown(context.Background())

	logger := mgr.Logger("test-logger")
	if logger == nil {
		t.Error("expected non-nil logger")
	}

	// 测试不同名称的 logger
	names := []string{"", "test", "logger-with-dash", "logger_with_underscore"}
	for _, name := range names {
		logger := mgr.Logger(name)
		if logger == nil {
			t.Errorf("expected non-nil logger for name '%s'", name)
		}
	}
}

// TestITelemetryManagerOtelImpl_Shutdown 测试 Shutdown 方法
func TestITelemetryManagerOtelImpl_Shutdown(t *testing.T) {
	config := &TelemetryConfig{
		Driver: "otel",
		OtelConfig: &OtelConfig{
			Endpoint: "localhost:4317",
			Traces:   &FeatureConfig{Enabled: false},
			Metrics:  &FeatureConfig{Enabled: false},
			Logs:     &FeatureConfig{Enabled: false},
		},
	}

	mgr, err := NewTelemetryManagerOtelImpl(config)
	if err != nil {
		t.Fatalf("failed to create manager: %v", err)
	}

	ctx := context.Background()

	// 第一次 shutdown
	if err := mgr.Shutdown(ctx); err != nil {
		t.Errorf("first shutdown failed: %v", err)
	}

	// 第二次 shutdown 应该也是安全的（使用 sync.Once）
	if err := mgr.Shutdown(ctx); err != nil {
		t.Errorf("second shutdown failed: %v", err)
	}
}

// TestITelemetryManagerOtelImpl_ResourceAttributes 测试资源属性
func TestITelemetryManagerOtelImpl_ResourceAttributes(t *testing.T) {
	tests := []struct {
		name   string
		config *TelemetryConfig
		verify func(*testing.T, ITelemetryManager)
	}{
		{
			name: "with resource attributes",
			config: &TelemetryConfig{
				Driver: "otel",
				OtelConfig: &OtelConfig{
					Endpoint: "localhost:4317",
					ResourceAttributes: []ResourceAttribute{
						{Key: "service.name", Value: "test-service"},
						{Key: "service.version", Value: "1.0.0"},
						{Key: "deployment.environment", Value: "production"},
					},
					Traces:  &FeatureConfig{Enabled: false},
					Metrics: &FeatureConfig{Enabled: false},
					Logs:    &FeatureConfig{Enabled: false},
				},
			},
			verify: func(t *testing.T, mgr ITelemetryManager) {
				// 验证管理器正常创建并工作
				if err := mgr.Health(); err != nil {
					t.Errorf("health check failed: %v", err)
				}
			},
		},
		{
			name: "without resource attributes",
			config: &TelemetryConfig{
				Driver: "otel",
				OtelConfig: &OtelConfig{
					Endpoint: "localhost:4317",
					Traces:   &FeatureConfig{Enabled: false},
					Metrics:  &FeatureConfig{Enabled: false},
					Logs:     &FeatureConfig{Enabled: false},
				},
			},
			verify: func(t *testing.T, mgr ITelemetryManager) {
				// 验证管理器正常创建并工作
				if err := mgr.Health(); err != nil {
					t.Errorf("health check failed: %v", err)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mgr, err := NewTelemetryManagerOtelImpl(tt.config)
			if err != nil {
				t.Fatalf("failed to create manager: %v", err)
			}
			defer mgr.Shutdown(context.Background())

			if tt.verify != nil {
				tt.verify(t, mgr)
			}
		})
	}
}

// TestITelemetryManagerOtelImpl_Features 测试特性启用/禁用
func TestITelemetryManagerOtelImpl_Features(t *testing.T) {
	tests := []struct {
		name   string
		config *TelemetryConfig
		verify func(*testing.T, ITelemetryManager)
	}{
		{
			name: "traces enabled",
			config: &TelemetryConfig{
				Driver: "otel",
				OtelConfig: &OtelConfig{
					Endpoint: "localhost:4317",
					Traces:   &FeatureConfig{Enabled: true},
					Metrics:  &FeatureConfig{Enabled: false},
					Logs:     &FeatureConfig{Enabled: false},
				},
			},
			verify: func(t *testing.T, mgr ITelemetryManager) {
				// 验证 tracer 可以正常工作
				tracer := mgr.Tracer("test")
				if tracer == nil {
					t.Error("expected non-nil tracer")
				}
			},
		},
		{
			name: "all features enabled",
			config: &TelemetryConfig{
				Driver: "otel",
				OtelConfig: &OtelConfig{
					Endpoint: "localhost:4317",
					Traces:   &FeatureConfig{Enabled: true},
					Metrics:  &FeatureConfig{Enabled: true},
					Logs:     &FeatureConfig{Enabled: true},
				},
			},
			verify: func(t *testing.T, mgr ITelemetryManager) {
				// 验证所有 provider 都可用
				if mgr.TracerProvider() == nil {
					t.Error("expected non-nil TracerProvider")
				}
				if mgr.MeterProvider() == nil {
					t.Error("expected non-nil MeterProvider")
				}
				if mgr.LoggerProvider() == nil {
					t.Error("expected non-nil LoggerProvider")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mgr, err := NewTelemetryManagerOtelImpl(tt.config)
			if err != nil {
				t.Fatalf("failed to create manager: %v", err)
			}
			defer mgr.Shutdown(context.Background())

			if tt.verify != nil {
				tt.verify(t, mgr)
			}
		})
	}
}

// TestITelemetryManagerOtelImpl_NilFeatureConfig 测试 nil 特性配置
func TestITelemetryManagerOtelImpl_NilFeatureConfig(t *testing.T) {
	config := &TelemetryConfig{
		Driver: "otel",
		OtelConfig: &OtelConfig{
			Endpoint: "localhost:4317",
			// Metrics 和 Logs 为 nil，应该被初始化为默认值
			Traces: &FeatureConfig{Enabled: false},
		},
	}

	mgr, err := NewTelemetryManagerOtelImpl(config)
	if err != nil {
		t.Fatalf("failed to create manager: %v", err)
	}
	defer mgr.Shutdown(context.Background())

	// 验证管理器正常工作
	if err := mgr.Health(); err != nil {
		t.Errorf("health check failed: %v", err)
	}
}

// TestITelemetryManagerOtelImpl_ConcurrentAccess 测试并发访问
func TestITelemetryManagerOtelImpl_ConcurrentAccess(t *testing.T) {
	config := &TelemetryConfig{
		Driver: "otel",
		OtelConfig: &OtelConfig{
			Endpoint: "localhost:4317",
			Traces:   &FeatureConfig{Enabled: false},
			Metrics:  &FeatureConfig{Enabled: false},
			Logs:     &FeatureConfig{Enabled: false},
		},
	}

	mgr, err := NewTelemetryManagerOtelImpl(config)
	if err != nil {
		t.Fatalf("failed to create manager: %v", err)
	}
	defer mgr.Shutdown(context.Background())

	done := make(chan bool)
	var wg sync.WaitGroup

	// 并发调用多个方法
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 100; j++ {
				mgr.Tracer("test")
				mgr.Meter("test")
				mgr.Logger("test")
			}
			done <- true
		}()
	}

	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 100; j++ {
				mgr.TracerProvider()
				mgr.MeterProvider()
				mgr.LoggerProvider()
			}
			done <- true
		}()
	}

	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			ctx := context.Background()
			for j := 0; j < 50; j++ {
				mgr.Health()
				mgr.Shutdown(ctx)
			}
			done <- true
		}()
	}

	// 等待所有 goroutine 完成
	go func() {
		wg.Wait()
		close(done)
	}()

	for range done {
		// drain channel
	}

	// 验证管理器仍然正常工作
	if err := mgr.Health(); err != nil {
		t.Errorf("manager health check failed after concurrent access: %v", err)
	}
}

// TestITelemetryManagerOtelImpl_InterfaceCompliance 测试接口符合性
func TestITelemetryManagerOtelImpl_InterfaceCompliance(t *testing.T) {
	config := &TelemetryConfig{
		Driver: "otel",
		OtelConfig: &OtelConfig{
			Endpoint: "localhost:4317",
			Traces:   &FeatureConfig{Enabled: false},
			Metrics:  &FeatureConfig{Enabled: false},
			Logs:     &FeatureConfig{Enabled: false},
		},
	}

	mgr, err := NewTelemetryManagerOtelImpl(config)
	if err != nil {
		t.Fatalf("failed to create manager: %v", err)
	}
	defer mgr.Shutdown(context.Background())

	// 验证 ITelemetryManager 接口的所有方法都能正常调用
	var _ ITelemetryManager = mgr

	// 调用所有接口方法确保没有 panic
	_ = mgr.ManagerName()
	_ = mgr.Health()
	_ = mgr.OnStart()
	_ = mgr.OnStop()

	_ = mgr.Tracer("test")
	_ = mgr.TracerProvider()

	_ = mgr.Meter("test")
	_ = mgr.MeterProvider()

	_ = mgr.Logger("test")
	_ = mgr.LoggerProvider()

	ctx := context.Background()
	_ = mgr.Shutdown(ctx)
}

// TestITelemetryManagerOtelImpl_TracerOperation 测试 Tracer 操作
func TestITelemetryManagerOtelImpl_TracerOperation(t *testing.T) {
	config := &TelemetryConfig{
		Driver: "otel",
		OtelConfig: &OtelConfig{
			Endpoint: "localhost:4317",
			Traces:   &FeatureConfig{Enabled: false},
			Metrics:  &FeatureConfig{Enabled: false},
			Logs:     &FeatureConfig{Enabled: false},
		},
	}

	mgr, err := NewTelemetryManagerOtelImpl(config)
	if err != nil {
		t.Fatalf("failed to create manager: %v", err)
	}
	defer mgr.Shutdown(context.Background())

	tracer := mgr.Tracer("test-tracer")
	ctx := context.Background()

	// 创建并操作 span
	ctx, span := tracer.Start(ctx, "test-operation")
	if span == nil {
		t.Error("expected non-nil span")
	}
	span.End()

	// 验证 context 仍然有效
	if ctx == nil {
		t.Error("context should not be nil")
	}
}
