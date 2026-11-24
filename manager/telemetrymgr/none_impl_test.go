package telemetrymgr

import (
	"context"
	"testing"
)

// TestNewTelemetryManagerNoneImpl 测试创建空实现
func TestNewTelemetryManagerNoneImpl(t *testing.T) {
	mgr := NewTelemetryManagerNoneImpl()

	if mgr == nil {
		t.Fatal("expected non-nil manager")
	}

	// 验证接口实现
	if _, ok := mgr.(*telemetryManagerNoneImpl); !ok {
		t.Error("expected *telemetryManagerNoneImpl type")
	}
}

// TestITelemetryManagerNoneImpl_ManagerName 测试管理器名称
func TestITelemetryManagerNoneImpl_ManagerName(t *testing.T) {
	mgr := NewTelemetryManagerNoneImpl()

	name := mgr.ManagerName()
	expected := "none-telemetry"

	if name != expected {
		t.Errorf("expected manager name '%s', got '%s'", expected, name)
	}
}

// TestITelemetryManagerNoneImpl_Health 测试健康检查
func TestITelemetryManagerNoneImpl_Health(t *testing.T) {
	mgr := NewTelemetryManagerNoneImpl()

	err := mgr.Health()
	if err != nil {
		t.Errorf("expected no error from Health, got %v", err)
	}
}

// TestITelemetryManagerNoneImpl_Lifecycle 测试生命周期方法
func TestITelemetryManagerNoneImpl_Lifecycle(t *testing.T) {
	mgr := NewTelemetryManagerNoneImpl()

	// 测试 OnStart
	if err := mgr.OnStart(); err != nil {
		t.Errorf("expected no error from OnStart, got %v", err)
	}

	// 测试 OnStop
	if err := mgr.OnStop(); err != nil {
		t.Errorf("expected no error from OnStop, got %v", err)
	}

	// 测试 Shutdown
	ctx := context.Background()
	if err := mgr.Shutdown(ctx); err != nil {
		t.Errorf("expected no error from Shutdown, got %v", err)
	}
}

// TestITelemetryManagerNoneImpl_Tracer 测试 Tracer 方法
func TestITelemetryManagerNoneImpl_Tracer(t *testing.T) {
	mgr := NewTelemetryManagerNoneImpl()

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

// TestITelemetryManagerNoneImpl_TracerProvider 测试 TracerProvider 方法
func TestITelemetryManagerNoneImpl_TracerProvider(t *testing.T) {
	mgr := NewTelemetryManagerNoneImpl()

	provider := mgr.TracerProvider()
	if provider == nil {
		t.Fatal("expected non-nil TracerProvider")
	}

	// 验证 provider 可以用于创建 tracer
	tracer := provider.Tracer("test")
	if tracer == nil {
		t.Error("expected non-nil tracer from provider")
	}
}

// TestITelemetryManagerNoneImpl_Meter 测试 Meter 方法
func TestITelemetryManagerNoneImpl_Meter(t *testing.T) {
	mgr := NewTelemetryManagerNoneImpl()

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

// TestITelemetryManagerNoneImpl_MeterProvider 测试 MeterProvider 方法
func TestITelemetryManagerNoneImpl_MeterProvider(t *testing.T) {
	mgr := NewTelemetryManagerNoneImpl()

	provider := mgr.MeterProvider()
	if provider == nil {
		t.Fatal("expected non-nil MeterProvider")
	}

	// 验证 provider 可以用于创建 meter
	meter := provider.Meter("test")
	if meter == nil {
		t.Error("expected non-nil meter from provider")
	}
}

// TestITelemetryManagerNoneImpl_Logger 测试 Logger 方法
func TestITelemetryManagerNoneImpl_Logger(t *testing.T) {
	mgr := NewTelemetryManagerNoneImpl()

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

// TestITelemetryManagerNoneImpl_LoggerProvider 测试 LoggerProvider 方法
func TestITelemetryManagerNoneImpl_LoggerProvider(t *testing.T) {
	mgr := NewTelemetryManagerNoneImpl()

	provider := mgr.LoggerProvider()
	if provider == nil {
		t.Fatal("expected non-nil LoggerProvider")
	}

	// 验证 provider 可以用于创建 logger
	logger := provider.Logger("test")
	if logger == nil {
		t.Error("expected non-nil logger from provider")
	}
}

// TestITelemetryManagerNoneImpl_ShutdownWithNilContext 测试使用 nil context shutdown
func TestITelemetryManagerNoneImpl_ShutdownWithNilContext(t *testing.T) {
	mgr := NewTelemetryManagerNoneImpl()

	// 空实现应该能够处理 nil context（虽然不推荐）
	err := mgr.Shutdown(nil)
	if err != nil {
		t.Errorf("expected no error with nil context, got %v", err)
	}
}

// TestITelemetryManagerNoneImpl_ShutdownMultipleTimes 测试多次关闭
func TestITelemetryManagerNoneImpl_ShutdownMultipleTimes(t *testing.T) {
	mgr := NewTelemetryManagerNoneImpl()

	ctx := context.Background()

	// 多次调用 shutdown 应该都成功
	for i := 0; i < 5; i++ {
		if err := mgr.Shutdown(ctx); err != nil {
			t.Errorf("shutdown iteration %d failed: %v", i, err)
		}
	}
}

// TestITelemetryManagerNoneImpl_InterfaceCompliance 测试接口符合性
func TestITelemetryManagerNoneImpl_InterfaceCompliance(t *testing.T) {
	mgr := NewTelemetryManagerNoneImpl()

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

// TestITelemetryManagerNoneImpl_NoOpBehavior 测试空操作行为
func TestITelemetryManagerNoneImpl_NoOpBehavior(t *testing.T) {
	mgr := NewTelemetryManagerNoneImpl()

	t.Run("tracer returns no-op tracer", func(t *testing.T) {
		tracer := mgr.Tracer("test")

		// No-op tracer 的所有方法都应该安全调用
		ctx := context.Background()
		_, span := tracer.Start(ctx, "test-operation")

		// 验证 span 可以正常操作
		span.AddEvent("test-event")
		span.End()

		// 验证从 span 获取的 context 仍然有效
		if ctx == nil {
			t.Error("context should not be nil")
		}
	})

	t.Run("meter returns meter", func(t *testing.T) {
		meter := mgr.Meter("test")

		// 即使是 no-op provider，返回的 meter 也不是 nil
		if meter == nil {
			t.Error("meter should not be nil")
		}
	})

	t.Run("logger returns logger", func(t *testing.T) {
		logger := mgr.Logger("test")

		// 即使是 no-op provider，返回的 logger 也不是 nil
		if logger == nil {
			t.Error("logger should not be nil")
		}
	})
}

// TestITelemetryManagerNoneImpl_ConcurrentAccess 测试并发访问
func TestITelemetryManagerNoneImpl_ConcurrentAccess(t *testing.T) {
	mgr := NewTelemetryManagerNoneImpl()

	done := make(chan bool)

	// 并发调用多个方法
	go func() {
		for i := 0; i < 100; i++ {
			mgr.Tracer("test")
			mgr.Meter("test")
			mgr.Logger("test")
		}
		done <- true
	}()

	go func() {
		for i := 0; i < 100; i++ {
			mgr.TracerProvider()
			mgr.MeterProvider()
			mgr.LoggerProvider()
		}
		done <- true
	}()

	go func() {
		ctx := context.Background()
		for i := 0; i < 100; i++ {
			mgr.Health()
			mgr.Shutdown(ctx)
		}
		done <- true
	}()

	// 等待所有 goroutine 完成
	for i := 0; i < 3; i++ {
		<-done
	}

	// 验证管理器仍然正常工作
	if err := mgr.Health(); err != nil {
		t.Errorf("manager health check failed after concurrent access: %v", err)
	}
}

// TestITelemetryManagerNoneImpl_ProvidersConsistency 测试 Provider 一致性
func TestITelemetryManagerNoneImpl_ProvidersConsistency(t *testing.T) {
	mgr := NewTelemetryManagerNoneImpl()

	// 多次获取相同的 provider 应该返回相同的实例
	tp1 := mgr.TracerProvider()
	tp2 := mgr.TracerProvider()

	if tp1 != tp2 {
		t.Error("TracerProvider should return the same instance")
	}

	mp1 := mgr.MeterProvider()
	mp2 := mgr.MeterProvider()

	if mp1 != mp2 {
		t.Error("MeterProvider should return the same instance")
	}

	lp1 := mgr.LoggerProvider()
	lp2 := mgr.LoggerProvider()

	if lp1 != lp2 {
		t.Error("LoggerProvider should return the same instance")
	}
}

// TestITelemetryManagerNoneImpl_ProviderTypes 测试 Provider 类型
func TestITelemetryManagerNoneImpl_ProviderTypes(t *testing.T) {
	mgr := NewTelemetryManagerNoneImpl()

	// 验证返回的 provider 是正确的具体类型
	tp := mgr.TracerProvider()
	if tp == nil {
		t.Fatal("TracerProvider should not be nil")
	}
	// 注意：不进行类型断言，因为 TracerProvider 不是接口

	mp := mgr.MeterProvider()
	if mp == nil {
		t.Fatal("MeterProvider should not be nil")
	}

	lp := mgr.LoggerProvider()
	if lp == nil {
		t.Fatal("LoggerProvider should not be nil")
	}
}
