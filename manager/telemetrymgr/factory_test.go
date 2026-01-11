package telemetrymgr

import (
	"context"
	"testing"

	"com.litelake.litecore/common"
)

func TestNewFactory(t *testing.T) {
	f := NewFactory()

	if f == nil {
		t.Fatal("NewFactory() returned nil")
	}
}

func TestFactory_Build_NoneDriver(t *testing.T) {
	f := NewFactory()

	mgr := f.Build("none", nil)
	if mgr == nil {
		t.Fatal("Build() returned nil manager")
	}
	if mgr.ManagerName() != "none-telemetry" {
		t.Errorf("Build() manager name = %v, want 'none-telemetry'", mgr.ManagerName())
	}

	// Cleanup
	_ = mgr.OnStart()
	_ = mgr.OnStop()
}

func TestFactory_Build_NoneDriverWithConfig(t *testing.T) {
	f := NewFactory()

	// none driver should ignore config
	cfg := map[string]any{
		"some": "config",
	}
	mgr := f.Build("none", cfg)
	if mgr == nil {
		t.Fatal("Build() returned nil manager")
	}
	if mgr.ManagerName() != "none-telemetry" {
		t.Errorf("Build() manager name = %v, want 'none-telemetry'", mgr.ManagerName())
	}
}

func TestFactory_Build_OtelDriver(t *testing.T) {
	f := NewFactory()

	otelCfg := map[string]any{
		"endpoint": "http://localhost:4317",
		"traces":   map[string]any{"enabled": true},
	}
	mgr := f.Build("otel", otelCfg)
	if mgr == nil {
		t.Fatal("Build() returned nil manager")
	}
	if mgr.ManagerName() != "otel-telemetry" {
		t.Errorf("Build() manager name = %v, want 'otel-telemetry'", mgr.ManagerName())
	}

	// Cleanup
	_ = mgr.OnStart()
	_ = mgr.OnStop()
}

func TestFactory_Build_OtelDriverWithoutEndpoint(t *testing.T) {
	f := NewFactory()

	// Missing endpoint should cause fallback to none driver
	otelCfg := map[string]any{
		"traces": map[string]any{"enabled": true},
	}
	mgr := f.Build("otel", otelCfg)
	if mgr == nil {
		t.Fatal("Build() returned nil manager")
	}
	// Should fallback to none driver
	if mgr.ManagerName() != "none-telemetry" {
		t.Errorf("Build() manager name = %v, want 'none-telemetry' (fallback)", mgr.ManagerName())
	}
}

func TestFactory_Build_OtelDriverWithNilConfig(t *testing.T) {
	f := NewFactory()

	// nil config should cause fallback to none driver (no endpoint)
	mgr := f.Build("otel", nil)
	if mgr == nil {
		t.Fatal("Build() returned nil manager")
	}
	// Should fallback to none driver
	if mgr.ManagerName() != "none-telemetry" {
		t.Errorf("Build() manager name = %v, want 'none-telemetry' (fallback)", mgr.ManagerName())
	}
}

func TestFactory_Build_OtelDriverWithFullConfig(t *testing.T) {
	f := NewFactory()

	otelCfg := map[string]any{
		"endpoint": "http://localhost:4317",
		"insecure": true,
		"headers": map[string]any{
			"authorization": "Bearer token",
		},
		"resource_attributes": []any{
			map[string]any{"key": "service.name", "value": "test-service"},
			map[string]any{"key": "env", "value": "test"},
		},
		"traces":  map[string]any{"enabled": true},
		"metrics": map[string]any{"enabled": true},
		"logs":    map[string]any{"enabled": false},
	}
	mgr := f.Build("otel", otelCfg)
	if mgr == nil {
		t.Fatal("Build() returned nil manager")
	}
	if mgr.ManagerName() != "otel-telemetry" {
		t.Errorf("Build() manager name = %v, want 'otel-telemetry'", mgr.ManagerName())
	}

	// Test that it implements TelemetryManager
	if tm, ok := mgr.(TelemetryManager); ok {
		tracer := tm.Tracer("test")
		if tracer == nil {
			t.Error("TelemetryManager.Tracer() returned nil")
		}

		ctx := context.Background()
		if err := tm.Shutdown(ctx); err != nil {
			t.Errorf("TelemetryManager.Shutdown() error = %v", err)
		}
	} else {
		t.Error("Manager does not implement TelemetryManager interface")
	}

	// Cleanup
	_ = mgr.OnStart()
	_ = mgr.OnStop()
}

func TestFactory_Build_UnknownDriver(t *testing.T) {
	f := NewFactory()

	mgr := f.Build("unknown", nil)
	if mgr == nil {
		t.Fatal("Build() returned nil manager")
	}
	// Should fallback to none driver
	if mgr.ManagerName() != "none-telemetry" {
		t.Errorf("Build() manager name = %v, want 'none-telemetry' (fallback)", mgr.ManagerName())
	}
}

func TestFactory_Build_EmptyDriver(t *testing.T) {
	f := NewFactory()

	mgr := f.Build("", nil)
	if mgr == nil {
		t.Fatal("Build() returned nil manager")
	}
	// Should fallback to none driver
	if mgr.ManagerName() != "none-telemetry" {
		t.Errorf("Build() manager name = %v, want 'none-telemetry' (fallback)", mgr.ManagerName())
	}
}

func TestFactory_Build_OtelDriverWithInvalidConfigType(t *testing.T) {
	f := NewFactory()

	otelCfg := map[string]any{
		"endpoint": 123, // Invalid type
	}
	mgr := f.Build("otel", otelCfg)
	if mgr == nil {
		t.Fatal("Build() returned nil manager")
	}
	// In non-strict mode, invalid types are ignored, so endpoint will be empty
	// and it should fallback to none driver
	if mgr.ManagerName() != "none-telemetry" {
		t.Errorf("Build() manager name = %v, want 'none-telemetry' (fallback)", mgr.ManagerName())
	}
}

func TestFactory_Build_Lifecycle(t *testing.T) {
	f := NewFactory()

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
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mgr := f.Build(tt.driver, tt.cfg)
			if mgr == nil {
				t.Fatal("Build() returned nil manager")
			}

			// Test lifecycle
			if err := mgr.OnStart(); err != nil {
				t.Errorf("OnStart() error = %v", err)
			}

			if err := mgr.Health(); err != nil {
				t.Errorf("Health() error = %v", err)
			}

			if err := mgr.OnStop(); err != nil {
				t.Errorf("OnStop() error = %v", err)
			}
		})
	}
}

func TestFactory_Build_ManagerImplementsCommonManager(t *testing.T) {
	f := NewFactory()

	// Test that all drivers return managers implementing common.BaseManager
	drivers := []struct {
		name   string
		driver string
		cfg    map[string]any
	}{
		{"none", "none", nil},
		{"otel", "otel", map[string]any{"endpoint": "http://localhost:4317"}},
	}

	for _, d := range drivers {
		t.Run(d.name, func(t *testing.T) {
			mgr := f.Build(d.driver, d.cfg)
			if mgr == nil {
				t.Fatal("Build() returned nil manager")
			}

			// Compile-time check that mgr implements common.BaseManager
			var _ common.BaseManager = mgr

			// Runtime check
			if _, ok := mgr.(common.BaseManager); !ok {
				t.Error("Manager does not implement common.BaseManager interface")
			}
		})
	}
}

func TestFactory_Build_ManagerImplementsTelemetryManager(t *testing.T) {
	f := NewFactory()

	// Test that all drivers return managers implementing TelemetryManager
	drivers := []struct {
		name   string
		driver string
		cfg    map[string]any
	}{
		{"none", "none", nil},
		{"otel", "otel", map[string]any{"endpoint": "http://localhost:4317"}},
	}

	for _, d := range drivers {
		t.Run(d.name, func(t *testing.T) {
			mgr := f.Build(d.driver, d.cfg)
			if mgr == nil {
				t.Fatal("Build() returned nil manager")
			}

			// Runtime check
			if _, ok := mgr.(TelemetryManager); !ok {
				t.Error("Manager does not implement TelemetryManager interface")
			}
		})
	}
}

func TestFactory_Build_TelemetryManagerMethods(t *testing.T) {
	f := NewFactory()

	otelCfg := map[string]any{
		"endpoint": "http://localhost:4317",
		"traces":   map[string]any{"enabled": true},
	}
	mgr := f.Build("otel", otelCfg)
	if mgr == nil {
		t.Fatal("Build() returned nil manager")
	}
	defer mgr.OnStop()

	tm, ok := mgr.(TelemetryManager)
	if !ok {
		t.Fatal("Manager does not implement TelemetryManager interface")
	}

	// Test Tracer method
	tracer := tm.Tracer("test-service")
	if tracer == nil {
		t.Error("Tracer() returned nil")
	}

	// Test Tracer with span
	ctx, span := tracer.Start(context.Background(), "test-operation")
	if ctx == nil {
		t.Error("Tracer.Start() returned nil context")
	}
	if span == nil {
		t.Error("Tracer.Start() returned nil span")
	}
	span.End()

	// Test Shutdown method
	ctx2 := context.Background()
	if err := tm.Shutdown(ctx2); err != nil {
		t.Errorf("Shutdown() error = %v", err)
	}
}
