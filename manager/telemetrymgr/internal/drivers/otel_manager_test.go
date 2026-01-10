package drivers

import (
	"context"
	"testing"

	"com.litelake.litecore/common"
	"com.litelake.litecore/manager/telemetrymgr/internal/config"
	"go.opentelemetry.io/otel/trace"
)

func TestNewOtelManager(t *testing.T) {
	tests := []struct {
		name    string
		cfg     *config.TelemetryConfig
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid otel config",
			cfg: &config.TelemetryConfig{
				Driver: "otel",
				OtelConfig: &config.OtelConfig{
					Endpoint: "http://localhost:4317",
					Traces:   &config.FeatureConfig{Enabled: false},
				},
			},
			wantErr: false,
		},
		{
			name: "valid otel config with traces enabled",
			cfg: &config.TelemetryConfig{
				Driver: "otel",
				OtelConfig: &config.OtelConfig{
					Endpoint: "localhost:4317",
					Insecure: true,
					Traces:   &config.FeatureConfig{Enabled: true},
				},
			},
			wantErr: false,
		},
		{
			name: "invalid driver",
			cfg: &config.TelemetryConfig{
				Driver: "none",
			},
			wantErr: true,
			errMsg:  "invalid driver for otel manager: none",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mgr, err := NewOtelManager(tt.cfg)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewOtelManager() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && err.Error() != tt.errMsg {
				t.Errorf("NewOtelManager() error message = %v, want %v", err.Error(), tt.errMsg)
			}
			if !tt.wantErr {
				if mgr == nil {
					t.Error("NewOtelManager() returned nil manager")
				}
				if mgr.ManagerName() != "otel-telemetry" {
					t.Errorf("NewOtelManager() name = %v, want 'otel-telemetry'", mgr.ManagerName())
				}
				// Cleanup
				ctx := context.Background()
				_ = mgr.Shutdown(ctx)
			}
		})
	}
}

func TestOtelManager_ManagerName(t *testing.T) {
	cfg := &config.TelemetryConfig{
		Driver: "otel",
		OtelConfig: &config.OtelConfig{
			Endpoint: "http://localhost:4317",
			Traces:   &config.FeatureConfig{Enabled: false},
		},
	}

	mgr, err := NewOtelManager(cfg)
	if err != nil {
		t.Fatalf("NewOtelManager() error = %v", err)
	}
	defer mgr.Shutdown(context.Background())

	want := "otel-telemetry"
	if got := mgr.ManagerName(); got != want {
		t.Errorf("OtelManager.ManagerName() = %v, want %v", got, want)
	}
}

func TestOtelManager_Tracer(t *testing.T) {
	cfg := &config.TelemetryConfig{
		Driver: "otel",
		OtelConfig: &config.OtelConfig{
			Endpoint: "http://localhost:4317",
			Traces:   &config.FeatureConfig{Enabled: false},
		},
	}

	mgr, err := NewOtelManager(cfg)
	if err != nil {
		t.Fatalf("NewOtelManager() error = %v", err)
	}
	defer mgr.Shutdown(context.Background())

	tracer := mgr.Tracer("test-service")
	if tracer == nil {
		t.Fatal("OtelManager.Tracer() returned nil")
	}

	// Test that tracer works
	ctx, span := tracer.Start(context.Background(), "test-operation")
	if ctx == nil {
		t.Error("Tracer.Start() returned nil context")
	}
	if span == nil {
		t.Error("Tracer.Start() returned nil span")
	}
	span.End()
}

func TestOtelManager_TracerMultipleCalls(t *testing.T) {
	cfg := &config.TelemetryConfig{
		Driver: "otel",
		OtelConfig: &config.OtelConfig{
			Endpoint: "http://localhost:4317",
			Traces:   &config.FeatureConfig{Enabled: false},
		},
	}

	mgr, err := NewOtelManager(cfg)
	if err != nil {
		t.Fatalf("NewOtelManager() error = %v", err)
	}
	defer mgr.Shutdown(context.Background())

	tracer1 := mgr.Tracer("service1")
	tracer2 := mgr.Tracer("service2")

	// Different service names should return different tracers
	if tracer1 == tracer2 {
		t.Error("OtelManager.Tracer() should return different tracers for different names")
	}

	// Same service name should return the same tracer
	tracer3 := mgr.Tracer("service1")
	if tracer1 != tracer3 {
		t.Error("OtelManager.Tracer() should return the same tracer for the same name")
	}
}

func TestOtelManager_TracerProvider(t *testing.T) {
	cfg := &config.TelemetryConfig{
		Driver: "otel",
		OtelConfig: &config.OtelConfig{
			Endpoint: "http://localhost:4317",
			Traces:   &config.FeatureConfig{Enabled: false},
		},
	}

	mgr, err := NewOtelManager(cfg)
	if err != nil {
		t.Fatalf("NewOtelManager() error = %v", err)
	}
	defer mgr.Shutdown(context.Background())

	tp := mgr.TracerProvider()
	if tp == nil {
		t.Fatal("OtelManager.TracerProvider() returned nil")
	}
}

func TestOtelManager_MeterProvider(t *testing.T) {
	cfg := &config.TelemetryConfig{
		Driver: "otel",
		OtelConfig: &config.OtelConfig{
			Endpoint: "http://localhost:4317",
			Traces:   &config.FeatureConfig{Enabled: false},
		},
	}

	mgr, err := NewOtelManager(cfg)
	if err != nil {
		t.Fatalf("NewOtelManager() error = %v", err)
	}
	defer mgr.Shutdown(context.Background())

	mp := mgr.MeterProvider()
	if mp == nil {
		t.Fatal("OtelManager.MeterProvider() returned nil")
	}
}

func TestOtelManager_Health(t *testing.T) {
	cfg := &config.TelemetryConfig{
		Driver: "otel",
		OtelConfig: &config.OtelConfig{
			Endpoint: "http://localhost:4317",
			Traces:   &config.FeatureConfig{Enabled: false},
		},
	}

	mgr, err := NewOtelManager(cfg)
	if err != nil {
		t.Fatalf("NewOtelManager() error = %v", err)
	}
	defer mgr.Shutdown(context.Background())

	if err := mgr.Health(); err != nil {
		t.Errorf("OtelManager.Health() error = %v, want nil", err)
	}
}

func TestOtelManager_OnStart(t *testing.T) {
	cfg := &config.TelemetryConfig{
		Driver: "otel",
		OtelConfig: &config.OtelConfig{
			Endpoint: "http://localhost:4317",
			Traces:   &config.FeatureConfig{Enabled: false},
		},
	}

	mgr, err := NewOtelManager(cfg)
	if err != nil {
		t.Fatalf("NewOtelManager() error = %v", err)
	}
	defer mgr.Shutdown(context.Background())

	if err := mgr.OnStart(); err != nil {
		t.Errorf("OtelManager.OnStart() error = %v, want nil", err)
	}
}

func TestOtelManager_OnStop(t *testing.T) {
	cfg := &config.TelemetryConfig{
		Driver: "otel",
		OtelConfig: &config.OtelConfig{
			Endpoint: "http://localhost:4317",
			Traces:   &config.FeatureConfig{Enabled: false},
		},
	}

	mgr, err := NewOtelManager(cfg)
	if err != nil {
		t.Fatalf("NewOtelManager() error = %v", err)
	}

	if err := mgr.OnStop(); err != nil {
		t.Errorf("OtelManager.OnStop() error = %v, want nil", err)
	}
}

func TestOtelManager_Shutdown(t *testing.T) {
	tests := []struct {
		name    string
		cfg     *config.TelemetryConfig
		ctx     context.Context
		wantErr bool
	}{
		{
			name: "valid context - traces disabled",
			cfg: &config.TelemetryConfig{
				Driver: "otel",
				OtelConfig: &config.OtelConfig{
					Endpoint: "http://localhost:4317",
					Traces:   &config.FeatureConfig{Enabled: false},
				},
			},
			ctx:     context.Background(),
			wantErr: false,
		},
		{
			name: "valid context - traces enabled",
			cfg: &config.TelemetryConfig{
				Driver: "otel",
				OtelConfig: &config.OtelConfig{
					Endpoint: "localhost:4317",
					Insecure: true,
					Traces:   &config.FeatureConfig{Enabled: true},
				},
			},
			ctx:     context.Background(),
			wantErr: false,
		},
		{
			name: "nil context",
			cfg: &config.TelemetryConfig{
				Driver: "otel",
				OtelConfig: &config.OtelConfig{
					Endpoint: "http://localhost:4317",
					Traces:   &config.FeatureConfig{Enabled: false},
				},
			},
			ctx:     nil,
			wantErr: false,
		},
		{
			name: "cancelled context",
			cfg: &config.TelemetryConfig{
				Driver: "otel",
				OtelConfig: &config.OtelConfig{
					Endpoint: "http://localhost:4317",
					Traces:   &config.FeatureConfig{Enabled: false},
				},
			},
			ctx:     func() context.Context { ctx, cancel := context.WithCancel(context.Background()); cancel(); return ctx }(),
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mgr, err := NewOtelManager(tt.cfg)
			if err != nil {
				t.Fatalf("NewOtelManager() error = %v", err)
			}

			if err := mgr.Shutdown(tt.ctx); (err != nil) != tt.wantErr {
				t.Errorf("OtelManager.Shutdown() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestOtelManager_ShutdownMultipleTimes(t *testing.T) {
	cfg := &config.TelemetryConfig{
		Driver: "otel",
		OtelConfig: &config.OtelConfig{
			Endpoint: "http://localhost:4317",
			Traces:   &config.FeatureConfig{Enabled: false},
		},
	}

	mgr, err := NewOtelManager(cfg)
	if err != nil {
		t.Fatalf("NewOtelManager() error = %v", err)
	}

	ctx := context.Background()

	// Shutdown should be idempotent (sync.Once)
	for i := 0; i < 5; i++ {
		if err := mgr.Shutdown(ctx); err != nil {
			t.Errorf("OtelManager.Shutdown() iteration %d error = %v", i, err)
		}
	}
}

func TestOtelManager_Lifecycle(t *testing.T) {
	cfg := &config.TelemetryConfig{
		Driver: "otel",
		OtelConfig: &config.OtelConfig{
			Endpoint: "http://localhost:4317",
			Traces:   &config.FeatureConfig{Enabled: false},
		},
	}

	mgr, err := NewOtelManager(cfg)
	if err != nil {
		t.Fatalf("NewOtelManager() error = %v", err)
	}
	ctx := context.Background()

	// Test complete lifecycle
	if err := mgr.OnStart(); err != nil {
		t.Fatalf("OtelManager.OnStart() error = %v", err)
	}

	if err := mgr.Health(); err != nil {
		t.Fatalf("OtelManager.Health() error = %v", err)
	}

	if err := mgr.OnStop(); err != nil {
		t.Fatalf("OtelManager.OnStop() error = %v", err)
	}

	if err := mgr.Shutdown(ctx); err != nil {
		t.Fatalf("OtelManager.Shutdown() error = %v", err)
	}
}

func TestOtelManager_WithResourceAttributes(t *testing.T) {
	cfg := &config.TelemetryConfig{
		Driver: "otel",
		OtelConfig: &config.OtelConfig{
			Endpoint: "http://localhost:4317",
			ResourceAttributes: []config.ResourceAttribute{
				{Key: "service.name", Value: "test-service"},
				{Key: "environment", Value: "test"},
			},
			Traces: &config.FeatureConfig{Enabled: false},
		},
	}

	mgr, err := NewOtelManager(cfg)
	if err != nil {
		t.Fatalf("NewOtelManager() error = %v", err)
	}
	defer mgr.Shutdown(context.Background())

	if err := mgr.Health(); err != nil {
		t.Errorf("OtelManager.Health() error = %v", err)
	}
}

func TestOtelManager_WithHeaders(t *testing.T) {
	cfg := &config.TelemetryConfig{
		Driver: "otel",
		OtelConfig: &config.OtelConfig{
			Endpoint: "localhost:4317",
			Insecure: true,
			Headers: map[string]string{
				"authorization": "Bearer test-token",
			},
			Traces: &config.FeatureConfig{Enabled: true},
		},
	}

	mgr, err := NewOtelManager(cfg)
	if err != nil {
		t.Fatalf("NewOtelManager() error = %v", err)
	}
	defer mgr.Shutdown(context.Background())

	if err := mgr.Health(); err != nil {
		t.Errorf("OtelManager.Health() error = %v", err)
	}
}

func TestOtelManagerImplementsManagerInterface(t *testing.T) {
	// Compile-time check that OtelManager implements common.Manager
	var _ common.Manager = (*OtelManager)(nil)
}

func TestOtelManagerImplementsTelemetryManagerInterface(t *testing.T) {
	// Compile-time check that OtelManager implements TelemetryManager
	type telemtryManagerInterface interface {
		Tracer(name string) trace.Tracer
		Shutdown(ctx context.Context) error
	}
	var _ telemtryManagerInterface = (*OtelManager)(nil)
}
