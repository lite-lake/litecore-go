package drivers

import (
	"context"
	"testing"

	"com.litelake.litecore/common"
	"go.opentelemetry.io/otel/trace"
)

func TestNewNoneManager(t *testing.T) {
	nm := NewNoneManager()

	if nm == nil {
		t.Fatal("NewNoneManager() returned nil")
	}
	if nm.BaseManager == nil {
		t.Error("NewNoneManager() BaseManager is nil")
	}
	if nm.name != "none-telemetry" {
		t.Errorf("NewNoneManager() name = %v, want 'none-telemetry'", nm.name)
	}
}

func TestNoneManager_ManagerName(t *testing.T) {
	nm := NewNoneManager()

	want := "none-telemetry"
	if got := nm.ManagerName(); got != want {
		t.Errorf("NoneManager.ManagerName() = %v, want %v", got, want)
	}
}

func TestNoneManager_Tracer(t *testing.T) {
	nm := NewNoneManager()

	tracer := nm.Tracer("test-service")
	if tracer == nil {
		t.Fatal("NoneManager.Tracer() returned nil")
	}

	// Test that tracer returns a noop tracer (it should not panic)
	ctx, span := tracer.Start(context.Background(), "test-operation")
	if ctx == nil {
		t.Error("Tracer.Start() returned nil context")
	}
	if span == nil {
		t.Error("Tracer.Start() returned nil span")
	}
	span.End()
}

func TestNoneManager_TracerMultipleCalls(t *testing.T) {
	nm := NewNoneManager()

	tracer1 := nm.Tracer("service1")
	tracer2 := nm.Tracer("service2")

	// Different service names should return different tracers
	if tracer1 == tracer2 {
		t.Error("NoneManager.Tracer() should return different tracers for different names")
	}

	// Same service name should return the same tracer
	tracer3 := nm.Tracer("service1")
	if tracer1 != tracer3 {
		t.Error("NoneManager.Tracer() should return the same tracer for the same name")
	}
}

func TestNoneManager_Shutdown(t *testing.T) {
	tests := []struct {
		name    string
		ctx     context.Context
		wantErr bool
	}{
		{
			name:    "valid context",
			ctx:     context.Background(),
			wantErr: false,
		},
		{
			name:    "nil context",
			ctx:     nil,
			wantErr: false,
		},
		{
			name:    "cancelled context",
			ctx:     func() context.Context { ctx, cancel := context.WithCancel(context.Background()); cancel(); return ctx }(),
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			nm := NewNoneManager()
			if err := nm.Shutdown(tt.ctx); (err != nil) != tt.wantErr {
				t.Errorf("NoneManager.Shutdown() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestNoneManager_ShutdownMultipleTimes(t *testing.T) {
	nm := NewNoneManager()
	ctx := context.Background()

	// Shutdown should be idempotent
	for i := 0; i < 5; i++ {
		if err := nm.Shutdown(ctx); err != nil {
			t.Errorf("NoneManager.Shutdown() iteration %d error = %v", i, err)
		}
	}
}

func TestNoneManager_Health(t *testing.T) {
	nm := NewNoneManager()

	if err := nm.Health(); err != nil {
		t.Errorf("NoneManager.Health() error = %v, want nil", err)
	}
}

func TestNoneManager_OnStart(t *testing.T) {
	nm := NewNoneManager()

	if err := nm.OnStart(); err != nil {
		t.Errorf("NoneManager.OnStart() error = %v, want nil", err)
	}
}

func TestNoneManager_OnStop(t *testing.T) {
	nm := NewNoneManager()

	if err := nm.OnStop(); err != nil {
		t.Errorf("NoneManager.OnStop() error = %v, want nil", err)
	}
}

func TestNoneManager_Lifecycle(t *testing.T) {
	nm := NewNoneManager()
	ctx := context.Background()

	// Test complete lifecycle
	if err := nm.OnStart(); err != nil {
		t.Fatalf("NoneManager.OnStart() error = %v", err)
	}

	if err := nm.Health(); err != nil {
		t.Fatalf("NoneManager.Health() error = %v", err)
	}

	if err := nm.OnStop(); err != nil {
		t.Fatalf("NoneManager.OnStop() error = %v", err)
	}

	if err := nm.Shutdown(ctx); err != nil {
		t.Fatalf("NoneManager.Shutdown() error = %v", err)
	}
}

func TestNoneManagerImplementsManagerInterface(t *testing.T) {
	// Compile-time check that NoneManager implements common.Manager
	var _ common.Manager = (*NoneManager)(nil)
}

func TestNoneManagerImplementsTelemetryManagerInterface(t *testing.T) {
	// Compile-time check that NoneManager implements TelemetryManager
	type telemtryManagerInterface interface {
		Tracer(name string) trace.Tracer
		Shutdown(ctx context.Context) error
	}
	var _ telemtryManagerInterface = (*NoneManager)(nil)
}
