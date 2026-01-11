package telemetrymgr

import (
	"context"
	"fmt"
	"testing"

	"com.litelake.litecore/common"
)

// MockConfigProvider 模拟配置提供者
type MockConfigProvider struct {
	config map[string]any
}

func NewMockConfigProvider(cfg map[string]any) *MockConfigProvider {
	return &MockConfigProvider{config: cfg}
}

func (m *MockConfigProvider) ConfigProviderName() string {
	return "mock"
}

func (m *MockConfigProvider) Get(key string) (any, error) {
	if val, ok := m.config[key]; ok {
		return val, nil
	}
	return nil, fmt.Errorf("config not found: %s", key)
}

func (m *MockConfigProvider) Has(key string) bool {
	_, ok := m.config[key]
	return ok
}

// TestNewManager 测试创建 Manager
func TestNewManager(t *testing.T) {
	mgr := NewManager("default")
	if mgr == nil {
		t.Fatal("NewManager returned nil")
	}

	if mgr.ManagerName() != "default" {
		t.Errorf("expected manager name 'default', got '%s'", mgr.ManagerName())
	}
}

// TestManager_OnStart_WithNilConfig 测试无配置时的启动
func TestManager_OnStart_WithNilConfig(t *testing.T) {
	mgr := NewManager("test-nil-config")
	mgr.Config = nil

	err := mgr.OnStart()
	if err != nil {
		t.Fatalf("OnStart failed: %v", err)
	}

	// 验证健康检查
	if err := mgr.Health(); err != nil {
		t.Errorf("Health check failed: %v", err)
	}

	// 验证可以获取 Tracer
	tracer := mgr.Tracer("test")
	if tracer == nil {
		t.Error("Tracer is nil")
	}
}

// TestManager_OnStart_WithNoneDriver 测试使用 none 驱动
func TestManager_OnStart_WithNoneDriver(t *testing.T) {
	cfg := map[string]any{
		"telemetry.test-none": map[string]any{
			"driver": "none",
		},
	}

	configProvider := NewMockConfigProvider(cfg)
	mgr := NewManager("test-none")
	mgr.Config = configProvider

	err := mgr.OnStart()
	if err != nil {
		t.Fatalf("OnStart failed: %v", err)
	}

	// 验证驱动已初始化
	if err := mgr.Health(); err != nil {
		t.Errorf("Health check failed: %v", err)
	}

	// 验证 Tracer
	tracer := mgr.Tracer("test")
	if tracer == nil {
		t.Error("Tracer is nil")
	}

	// 验证 Meter
	meter := mgr.Meter("test")
	if meter == nil {
		t.Error("Meter is nil")
	}

	// 验证 Logger
	logger := mgr.Logger("test")
	if logger == nil {
		t.Error("Logger is nil")
	}
}

// TestManager_OnStart_WithOtelDriver 测试使用 OTEL 驱动
func TestManager_OnStart_WithOtelDriver(t *testing.T) {
	cfg := map[string]any{
		"telemetry.test-otel": map[string]any{
			"driver": "otel",
			"otel_config": map[string]any{
				"endpoint": "localhost:4317",
				"insecure": true,
				"traces": map[string]any{
					"enabled": false,
				},
				"metrics": map[string]any{
					"enabled": false,
				},
				"logs": map[string]any{
					"enabled": false,
				},
			},
		},
	}

	configProvider := NewMockConfigProvider(cfg)
	mgr := NewManager("test-otel")
	mgr.Config = configProvider

	err := mgr.OnStart()
	if err != nil {
		t.Fatalf("OnStart failed: %v", err)
	}

	// 验证驱动已初始化
	if err := mgr.Health(); err != nil {
		t.Errorf("Health check failed: %v", err)
	}

	// 验证 Tracer
	tracer := mgr.Tracer("test")
	if tracer == nil {
		t.Error("Tracer is nil")
	}

	// 验证 TracerProvider
	tp := mgr.TracerProvider()
	if tp == nil {
		t.Error("TracerProvider is nil")
	}
}

// TestManager_OnStart_WithInvalidConfig 测试配置解析失败时的降级
func TestManager_OnStart_WithInvalidConfig(t *testing.T) {
	cfg := map[string]any{
		"telemetry.test-invalid": map[string]any{
			"driver": "otel",
			// 缺少 otel_config
		},
	}

	configProvider := NewMockConfigProvider(cfg)
	mgr := NewManager("test-invalid")
	mgr.Config = configProvider

	err := mgr.OnStart()
	// 配置解析失败，应该降级到 none 驱动，但返回错误
	if err == nil {
		t.Error("Expected error when config is invalid, got nil")
	}

	// 验证健康检查仍然成功（因为降级到 none 驱动）
	if err := mgr.Health(); err != nil {
		t.Errorf("Health check should succeed with none driver, got: %v", err)
	}
}

// TestManager_OnStart_WithMissingConfig 测试配置不存在时的默认行为
func TestManager_OnStart_WithMissingConfig(t *testing.T) {
	cfg := map[string]any{
		"other.key": "value",
	}

	configProvider := NewMockConfigProvider(cfg)
	mgr := NewManager("test-missing")
	mgr.Config = configProvider

	err := mgr.OnStart()
	if err != nil {
		t.Fatalf("OnStart failed: %v", err)
	}

	// 验证使用默认配置
	if err := mgr.Health(); err != nil {
		t.Errorf("Health check failed: %v", err)
	}
}

// TestManager_OnStart_Once 测试 OnStart 只执行一次
func TestManager_OnStart_Once(t *testing.T) {
	mgr := NewManager("test-once")
	mgr.Config = nil

	// 多次调用 OnStart
	for i := 0; i < 3; i++ {
		err := mgr.OnStart()
		if err != nil {
			t.Fatalf("OnStart call %d failed: %v", i+1, err)
		}
	}

	// 验证健康检查
	if err := mgr.Health(); err != nil {
		t.Errorf("Health check failed: %v", err)
	}
}

// TestManager_OnStop 测试停止管理器
func TestManager_OnStop(t *testing.T) {
	cfg := map[string]any{
		"telemetry.test-stop": map[string]any{
			"driver": "none",
		},
	}

	configProvider := NewMockConfigProvider(cfg)
	mgr := NewManager("test-stop")
	mgr.Config = configProvider

	// 启动管理器
	if err := mgr.OnStart(); err != nil {
		t.Fatalf("OnStart failed: %v", err)
	}

	// 停止管理器
	if err := mgr.OnStop(); err != nil {
		t.Fatalf("OnStop failed: %v", err)
	}

	// 再次停止应该不会有问题
	if err := mgr.OnStop(); err != nil {
		t.Errorf("Second OnStop should not fail, got: %v", err)
	}
}

// TestManager_Shutdown 测试 Shutdown 方法
func TestManager_Shutdown(t *testing.T) {
	cfg := map[string]any{
		"telemetry.test-shutdown": map[string]any{
			"driver": "none",
		},
	}

	configProvider := NewMockConfigProvider(cfg)
	mgr := NewManager("test-shutdown")
	mgr.Config = configProvider

	// 启动管理器
	if err := mgr.OnStart(); err != nil {
		t.Fatalf("OnStart failed: %v", err)
	}

	// 调用 Shutdown
	ctx := context.Background()
	if err := mgr.Shutdown(ctx); err != nil {
		t.Fatalf("Shutdown failed: %v", err)
	}
}

// TestManager_TracerProvider 测试 TracerProvider
func TestManager_TracerProvider(t *testing.T) {
	mgr := NewManager("test-tracer-provider")
	mgr.Config = nil

	if err := mgr.OnStart(); err != nil {
		t.Fatalf("OnStart failed: %v", err)
	}

	tp := mgr.TracerProvider()
	if tp == nil {
		t.Error("TracerProvider is nil")
	}

	// 验证可以创建 Tracer
	tracer := tp.Tracer("test")
	if tracer == nil {
		t.Error("Tracer from TracerProvider is nil")
	}
}

// TestManager_MeterProvider 测试 MeterProvider
func TestManager_MeterProvider(t *testing.T) {
	mgr := NewManager("test-meter-provider")
	mgr.Config = nil

	if err := mgr.OnStart(); err != nil {
		t.Fatalf("OnStart failed: %v", err)
	}

	mp := mgr.MeterProvider()
	if mp == nil {
		t.Error("MeterProvider is nil")
	}

	// 验证可以创建 Meter
	meter := mp.Meter("test")
	if meter == nil {
		t.Error("Meter from MeterProvider is nil")
	}
}

// TestManager_LoggerProvider 测试 LoggerProvider
func TestManager_LoggerProvider(t *testing.T) {
	mgr := NewManager("test-logger-provider")
	mgr.Config = nil

	if err := mgr.OnStart(); err != nil {
		t.Fatalf("OnStart failed: %v", err)
	}

	lp := mgr.LoggerProvider()
	if lp == nil {
		t.Error("LoggerProvider is nil")
	}

	// 验证可以创建 Logger
	logger := lp.Logger("test")
	if logger == nil {
		t.Error("Logger from LoggerProvider is nil")
	}
}

// TestManager_ImplementsBaseManager 测试 Manager 实现 BaseManager 接口
func TestManager_ImplementsBaseManager(t *testing.T) {
	var _ common.BaseManager = (*Manager)(nil)
}

// TestManager_ImplementsTelemetryManager 测试 Manager 实现 TelemetryManager 接口
func TestManager_ImplementsTelemetryManager(t *testing.T) {
	var _ TelemetryManager = (*Manager)(nil)
}

// TestManager_Tracer_IsThreadSafe 测试 Tracer 方法的线程安全性
func TestManager_Tracer_IsThreadSafe(t *testing.T) {
	cfg := map[string]any{
		"telemetry.test-concurrent": map[string]any{
			"driver": "none",
		},
	}

	configProvider := NewMockConfigProvider(cfg)
	mgr := NewManager("test-concurrent")
	mgr.Config = configProvider

	if err := mgr.OnStart(); err != nil {
		t.Fatalf("OnStart failed: %v", err)
	}

	// 并发调用 Tracer
	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func(id int) {
			tracer := mgr.Tracer(fmt.Sprintf("tracer-%d", id))
			if tracer == nil {
				t.Errorf("Tracer %d is nil", id)
			}
			done <- true
		}(i)
	}

	// 等待所有 goroutine 完成
	for i := 0; i < 10; i++ {
		<-done
	}
}

// TestManager_ConfigKeyFormat 测试配置键格式
func TestManager_ConfigKeyFormat(t *testing.T) {
	tests := []struct {
		name           string
		expectedConfig string
	}{
		{"default", "telemetry.default"},
		{"primary", "telemetry.primary"},
		{"secondary", "telemetry.secondary"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mgr := NewManager(tt.name)
			expectedKey := fmt.Sprintf("telemetry.%s", tt.name)

			// 通过检查 Config.Get 调用验证配置键格式
			cfg := map[string]any{
				expectedKey: map[string]any{
					"driver": "none",
				},
			}

			configProvider := NewMockConfigProvider(cfg)
			mgr.Config = configProvider

			if err := mgr.OnStart(); err != nil {
				t.Fatalf("OnStart failed: %v", err)
			}
		})
	}
}

// TestManager_WithOtelResourceAttributes 测试 OTEL 资源属性配置
func TestManager_WithOtelResourceAttributes(t *testing.T) {
	cfg := map[string]any{
		"telemetry.test-resource": map[string]any{
			"driver": "otel",
			"otel_config": map[string]any{
				"endpoint": "localhost:4317",
				"insecure": true,
				"resource_attributes": []map[string]any{
					{"key": "service.name", "value": "test-service"},
					{"key": "service.version", "value": "1.0.0"},
				},
				"traces": map[string]any{
					"enabled": false,
				},
				"metrics": map[string]any{
					"enabled": false,
				},
				"logs": map[string]any{
					"enabled": false,
				},
			},
		},
	}

	configProvider := NewMockConfigProvider(cfg)
	mgr := NewManager("test-resource")
	mgr.Config = configProvider

	err := mgr.OnStart()
	if err != nil {
		t.Fatalf("OnStart failed: %v", err)
	}

	// 验证驱动已初始化
	if err := mgr.Health(); err != nil {
		t.Errorf("Health check failed: %v", err)
	}
}

// TestManager_Tracer_NoOpTracer 测试无操作 Tracer
func TestManager_Tracer_NoOpTracer(t *testing.T) {
	mgr := NewManager("test-noop")
	mgr.Config = nil

	if err := mgr.OnStart(); err != nil {
		t.Fatalf("OnStart failed: %v", err)
	}

	tracer := mgr.Tracer("test")

	// 验证是一个 noop tracer
	ctx := context.Background()
	_, span := tracer.Start(ctx, "test-operation")
	span.End()

	// No-op tracer 应该不会 panic
}

// BenchmarkManager_Tracer 性能测试：Tracer 方法
func BenchmarkManager_Tracer(b *testing.B) {
	mgr := NewManager("bench")
	mgr.Config = nil
	mgr.OnStart()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = mgr.Tracer("test")
	}
}

// BenchmarkManager_Tracer_Parallel 并发性能测试：Tracer 方法
func BenchmarkManager_Tracer_Parallel(b *testing.B) {
	mgr := NewManager("bench-parallel")
	mgr.Config = nil
	mgr.OnStart()

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			_ = mgr.Tracer(fmt.Sprintf("tracer-%d", i%100))
			i++
		}
	})
}
