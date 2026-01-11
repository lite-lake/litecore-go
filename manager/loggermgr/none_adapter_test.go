package loggermgr

import (
	"context"
	"testing"

	"com.litelake.litecore/manager/loggermgr/internal/drivers"
)

func TestNewNoneLoggerManagerAdapter(t *testing.T) {
	noneMgr := drivers.NewNoneLoggerManager()
	adapter := NewNoneLoggerManagerAdapter(noneMgr)

	if adapter == nil {
		t.Fatal("NewNoneLoggerManagerAdapter() returned nil")
	}

	if adapter.driver != noneMgr {
		t.Error("NewNoneLoggerManagerAdapter() did not set the driver correctly")
	}
}

func TestNoneLoggerManagerAdapter_Logger(t *testing.T) {
	noneMgr := drivers.NewNoneLoggerManager()
	adapter := NewNoneLoggerManagerAdapter(noneMgr)

	tests := []struct {
		name string
	}{
		{"test-logger"},
		{"another-logger"},
		{""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := adapter.Logger(tt.name)

			if logger == nil {
				t.Error("Logger() returned nil")
			}

			// Should return a NoneLoggerAdapter
			if _, ok := logger.(*NoneLoggerAdapter); !ok {
				t.Error("Logger() should return a NoneLoggerAdapter")
			}
		})
	}
}

func TestNoneLoggerManagerAdapter_SetGlobalLevel(t *testing.T) {
	noneMgr := drivers.NewNoneLoggerManager()
	adapter := NewNoneLoggerManagerAdapter(noneMgr)

	// Should not panic
	adapter.SetGlobalLevel(DebugLevel)
	adapter.SetGlobalLevel(FatalLevel)
}

func TestNoneLoggerManagerAdapter_Shutdown(t *testing.T) {
	noneMgr := drivers.NewNoneLoggerManager()
	adapter := NewNoneLoggerManagerAdapter(noneMgr)

	tests := []struct {
		name    string
		ctx     context.Context
		wantErr bool
	}{
		{
			name:    "normal context",
			ctx:     context.Background(),
			wantErr: false,
		},
		{
			name:    "nil context",
			ctx:     nil,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := adapter.Shutdown(tt.ctx)
			if (err != nil) != tt.wantErr {
				t.Errorf("NoneLoggerManagerAdapter.Shutdown() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestNoneLoggerManagerAdapter_ManagerName(t *testing.T) {
	noneMgr := drivers.NewNoneLoggerManager()
	adapter := NewNoneLoggerManagerAdapter(noneMgr)

	expected := "none-logger"
	if got := adapter.ManagerName(); got != expected {
		t.Errorf("NoneLoggerManagerAdapter.ManagerName() = %v, want %v", got, expected)
	}
}

func TestNoneLoggerManagerAdapter_CommonManager(t *testing.T) {
	noneMgr := drivers.NewNoneLoggerManager()
	adapter := NewNoneLoggerManagerAdapter(noneMgr)

	// Test common.BaseManager interface methods
	if err := adapter.Health(); err != nil {
		t.Errorf("NoneLoggerManagerAdapter.Health() error = %v, want nil", err)
	}

	if err := adapter.OnStart(); err != nil {
		t.Errorf("NoneLoggerManagerAdapter.OnStart() error = %v, want nil", err)
	}

	if err := adapter.OnStop(); err != nil {
		t.Errorf("NoneLoggerManagerAdapter.OnStop() error = %v, want nil", err)
	}
}

func TestNoneLoggerAdapter_NoPanic(t *testing.T) {
	noneLogger := drivers.NewNoneLogger()
	adapter := &NoneLoggerAdapter{driver: noneLogger}

	// All methods should not panic
	tests := []struct {
		name string
		fn   func()
	}{
		{"Debug", func() { adapter.Debug("test", "key", "value") }},
		{"Info", func() { adapter.Info("test", "key", "value") }},
		{"Warn", func() { adapter.Warn("test", "key", "value") }},
		{"Error", func() { adapter.Error("test", "key", "value") }},
		{"Fatal", func() { adapter.Fatal("test", "key", "value") }},
		{"With", func() { _ = adapter.With("key", "value") }},
		{"SetLevel", func() { adapter.SetLevel(DebugLevel) }},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("NoneLoggerAdapter.%s() panicked: %v", tt.name, r)
				}
			}()
			tt.fn()
		})
	}
}

func TestNoneLoggerAdapter_WithReturnsSameInstance(t *testing.T) {
	noneLogger := drivers.NewNoneLogger()
	adapter := &NoneLoggerAdapter{driver: noneLogger}

	result := adapter.With("key", "value")

	if result != adapter {
		t.Error("NoneLoggerAdapter.With() should return the same instance")
	}
}

func TestNoneLoggerAdapter_NoOps(t *testing.T) {
	noneLogger := drivers.NewNoneLogger()
	adapter := &NoneLoggerAdapter{driver: noneLogger}

	// Verify all methods are no-ops
	adapter.Debug("debug message")
	adapter.Info("info message")
	adapter.Warn("warn message")
	adapter.Error("error message")
	adapter.Fatal("fatal message") // Should NOT exit
	adapter.SetLevel(DebugLevel)

	// If we're here, Fatal didn't exit, which is the expected behavior
}

func TestNoneLoggerAdapter_ConcurrentUsage(t *testing.T) {
	noneLogger := drivers.NewNoneLogger()
	adapter := &NoneLoggerAdapter{driver: noneLogger}
	done := make(chan bool)

	// Concurrent calls to ensure thread safety
	for i := 0; i < 100; i++ {
		go func(i int) {
			adapter.Info("test", "i", i)
			_ = adapter.With("key", "value")
			adapter.SetLevel(LogLevel(i % 5))
			done <- true
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < 100; i++ {
		<-done
	}
}

func TestNoneLoggerAdapter_Interface(t *testing.T) {
	noneLogger := drivers.NewNoneLogger()
	var logger Logger = &NoneLoggerAdapter{driver: noneLogger}

	// Should implement all interface methods without panicking
	logger.Debug("test")
	logger.Info("test")
	logger.Warn("test")
	logger.Error("test")
	logger.Fatal("test") // Should NOT exit
	logger.SetLevel(InfoLevel)
	_ = logger.With("key", "value")
}

func TestNoneLoggerManagerAdapter_Interface(t *testing.T) {
	noneMgr := drivers.NewNoneLoggerManager()
	adapter := NewNoneLoggerManagerAdapter(noneMgr)

	// Test LoggerManager interface
	logger := adapter.Logger("test")
	if logger == nil {
		t.Error("Logger() returned nil")
	}

	adapter.SetGlobalLevel(InfoLevel)

	ctx := context.Background()
	if err := adapter.Shutdown(ctx); err != nil {
		t.Errorf("Shutdown() error = %v", err)
	}
}
