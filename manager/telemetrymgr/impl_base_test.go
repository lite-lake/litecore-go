package telemetrymgr

import (
	"context"
	"testing"
)

// TestNewTelemetryManagerBaseImpl 测试创建基础实现
func TestNewTelemetryManagerBaseImpl(t *testing.T) {
	base := newTelemetryManagerBaseImpl("test-manager")

	if base == nil {
		t.Fatal("expected non-nil base implementation")
	}

	if base.name != "test-manager" {
		t.Errorf("expected name 'test-manager', got '%s'", base.name)
	}
}

// TestTelemetryManagerBaseImpl_ManagerName 测试 ManagerName 方法
func TestTelemetryManagerBaseImpl_ManagerName(t *testing.T) {
	tests := []struct {
		name     string
		manager  string
		expected string
	}{
		{
			name:     "simple name",
			manager:  "test",
			expected: "test",
		},
		{
			name:     "name with spaces",
			manager:  "test manager",
			expected: "test manager",
		},
		{
			name:     "name with special chars",
			manager:  "test-manager-123",
			expected: "test-manager-123",
		},
		{
			name:     "empty name",
			manager:  "",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			base := newTelemetryManagerBaseImpl(tt.manager)
			if base.ManagerName() != tt.expected {
				t.Errorf("expected name '%s', got '%s'", tt.expected, base.ManagerName())
			}
		})
	}
}

// TestTelemetryManagerBaseImpl_Health 测试 Health 方法
func TestTelemetryManagerBaseImpl_Health(t *testing.T) {
	base := newTelemetryManagerBaseImpl("test-manager")

	err := base.Health()
	if err != nil {
		t.Errorf("expected no error from Health, got %v", err)
	}
}

// TestTelemetryManagerBaseImpl_OnStart 测试 OnStart 方法
func TestTelemetryManagerBaseImpl_OnStart(t *testing.T) {
	base := newTelemetryManagerBaseImpl("test-manager")

	err := base.OnStart()
	if err != nil {
		t.Errorf("expected no error from OnStart, got %v", err)
	}
}

// TestTelemetryManagerBaseImpl_OnStop 测试 OnStop 方法
func TestTelemetryManagerBaseImpl_OnStop(t *testing.T) {
	base := newTelemetryManagerBaseImpl("test-manager")

	err := base.OnStop()
	if err != nil {
		t.Errorf("expected no error from OnStop, got %v", err)
	}
}

// TestTelemetryManagerBaseImpl_Lifecycle 测试完整的生命周期
func TestTelemetryManagerBaseImpl_Lifecycle(t *testing.T) {
	base := newTelemetryManagerBaseImpl("test-manager")

	// 按顺序调用生命周期方法
	if err := base.OnStart(); err != nil {
		t.Errorf("OnStart failed: %v", err)
	}

	if err := base.Health(); err != nil {
		t.Errorf("Health failed: %v", err)
	}

	if err := base.OnStop(); err != nil {
		t.Errorf("OnStop failed: %v", err)
	}
}

// TestValidateContext 测试 ValidateContext 函数
func TestValidateContext(t *testing.T) {
	tests := []struct {
		name    string
		ctx     context.Context
		wantErr bool
		errMsg  string
	}{
		{
			name:    "valid context",
			ctx:     context.Background(),
			wantErr: false,
		},
		{
			name:    "valid context with values",
			ctx:     context.WithValue(context.Background(), "key", "value"),
			wantErr: false,
		},
		{
			name:    "valid context with timeout",
			ctx:     context.Background(),
			wantErr: false,
		},
		{
			name:    "nil context",
			ctx:     nil,
			wantErr: true,
			errMsg:  "context cannot be nil",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateContext(tt.ctx)
			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error, got nil")
				}
				if tt.errMsg != "" && err != nil {
					if len(err.Error()) < len(tt.errMsg) || err.Error()[:len(tt.errMsg)] != tt.errMsg {
						t.Errorf("expected error containing '%s', got '%s'", tt.errMsg, err.Error())
					}
				}
			} else {
				if err != nil {
					t.Errorf("expected no error, got %v", err)
				}
			}
		})
	}
}

// TestTelemetryManagerBaseImpl_DefaultBehavior 测试默认行为
func TestTelemetryManagerBaseImpl_DefaultBehavior(t *testing.T) {
	base := newTelemetryManagerBaseImpl("test-manager")

	t.Run("Health always returns nil", func(t *testing.T) {
		for i := 0; i < 10; i++ {
			if err := base.Health(); err != nil {
				t.Errorf("iteration %d: expected no error, got %v", i, err)
			}
		}
	})

	t.Run("OnStart always returns nil", func(t *testing.T) {
		for i := 0; i < 10; i++ {
			if err := base.OnStart(); err != nil {
				t.Errorf("iteration %d: expected no error, got %v", i, err)
			}
		}
	})

	t.Run("OnStop always returns nil", func(t *testing.T) {
		for i := 0; i < 10; i++ {
			if err := base.OnStop(); err != nil {
				t.Errorf("iteration %d: expected no error, got %v", i, err)
			}
		}
	})
}

// TestTelemetryManagerBaseImpl_EmbeddingInImplementations 测试在具体实现中的嵌入
func TestTelemetryManagerBaseImpl_EmbeddingInImplementations(t *testing.T) {
	t.Run("none implementation", func(t *testing.T) {
		mgr := NewTelemetryManagerNoneImpl()

		// 通过嵌入的基础实现调用方法
		name := mgr.ManagerName()
		if name != "none-telemetry" {
			t.Errorf("expected manager name 'none-telemetry', got '%s'", name)
		}

		// 基础方法应该正常工作
		if err := mgr.Health(); err != nil {
			t.Errorf("Health check failed: %v", err)
		}

		if err := mgr.OnStart(); err != nil {
			t.Errorf("OnStart failed: %v", err)
		}

		if err := mgr.OnStop(); err != nil {
			t.Errorf("OnStop failed: %v", err)
		}
	})

	t.Run("otel implementation", func(t *testing.T) {
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

		// 通过嵌入的基础实现调用方法
		name := mgr.ManagerName()
		if name != "otel-telemetry" {
			t.Errorf("expected manager name 'otel-telemetry', got '%s'", name)
		}

		// 基础方法应该正常工作
		if err := mgr.Health(); err != nil {
			t.Errorf("Health check failed: %v", err)
		}

		if err := mgr.OnStart(); err != nil {
			t.Errorf("OnStart failed: %v", err)
		}

		if err := mgr.OnStop(); err != nil {
			t.Errorf("OnStop failed: %v", err)
		}
	})
}

// TestTelemetryManagerBaseImpl_ConcurrentAccess 测试并发访问
func TestTelemetryManagerBaseImpl_ConcurrentAccess(t *testing.T) {
	base := newTelemetryManagerBaseImpl("test-manager")

	done := make(chan bool)

	// 并发调用所有方法
	for i := 0; i < 10; i++ {
		go func() {
			for j := 0; j < 100; j++ {
				base.ManagerName()
			}
			done <- true
		}()

		go func() {
			for j := 0; j < 100; j++ {
				base.Health()
			}
			done <- true
		}()

		go func() {
			for j := 0; j < 100; j++ {
				base.OnStart()
			}
			done <- true
		}()

		go func() {
			for j := 0; j < 100; j++ {
				base.OnStop()
			}
			done <- true
		}()
	}

	// 等待所有 goroutine 完成
	for i := 0; i < 40; i++ {
		<-done
	}

	// 验证管理器仍然正常工作
	if err := base.Health(); err != nil {
		t.Errorf("health check failed after concurrent access: %v", err)
	}
}

// TestTelemetryManagerBaseImpl_NameConsistency 测试名称一致性
func TestTelemetryManagerBaseImpl_NameConsistency(t *testing.T) {
	base := newTelemetryManagerBaseImpl("consistent-name")

	// 多次调用应该返回相同的名称
	for i := 0; i < 100; i++ {
		if name := base.ManagerName(); name != "consistent-name" {
			t.Errorf("iteration %d: expected name 'consistent-name', got '%s'", i, name)
		}
	}
}

// BenchmarkTelemetryManagerBaseImpl_ManagerName 基准测试 ManagerName
func BenchmarkTelemetryManagerBaseImpl_ManagerName(b *testing.B) {
	base := newTelemetryManagerBaseImpl("benchmark-manager")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		base.ManagerName()
	}
}

// BenchmarkValidateContext 基准测试 ValidateContext
func BenchmarkValidateContext(b *testing.B) {
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ValidateContext(ctx)
	}
}
