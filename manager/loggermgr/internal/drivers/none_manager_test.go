package drivers

import (
	"context"
	"testing"

	"com.litelake.litecore/manager/loggermgr/internal/loglevel"
)

func TestNewNoneLogger(t *testing.T) {
	logger := NewNoneLogger()

	if logger == nil {
		t.Fatal("NewNoneLogger() returned nil")
	}
}

func TestNoneLogger_NoPanic(t *testing.T) {
	logger := NewNoneLogger()

	// All methods should not panic
	tests := []struct {
		name string
		fn   func()
	}{
		{"Debug", func() { logger.Debug("test", "key", "value") }},
		{"Info", func() { logger.Info("test", "key", "value") }},
		{"Warn", func() { logger.Warn("test", "key", "value") }},
		{"Error", func() { logger.Error("test", "key", "value") }},
		{"Fatal", func() { logger.Fatal("test", "key", "value") }},
		{"With", func() { logger.With("key", "value") }},
		{"SetLevel", func() { logger.SetLevel(loglevel.DebugLevel) }},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("NoneLogger.%s() panicked: %v", tt.name, r)
				}
			}()
			tt.fn()
		})
	}
}

func TestNoneLogger_WithReturnsSameInstance(t *testing.T) {
	logger := NewNoneLogger()
	result := logger.With("key", "value")

	if result != logger {
		t.Error("NoneLogger.With() should return the same instance")
	}
}

func TestNoneLogger_NoOps(t *testing.T) {
	logger := NewNoneLogger()

	// Verify all methods are no-ops (they don't do anything)
	logger.Debug("debug message")
	logger.Info("info message")
	logger.Warn("warn message")
	logger.Error("error message")
	logger.Fatal("fatal message") // Should NOT exit
	logger.SetLevel(loglevel.DebugLevel)

	// If we're here, Fatal didn't exit, which is the expected behavior for NoneLogger
}

func TestNewNoneLoggerManager(t *testing.T) {
	mgr := NewNoneLoggerManager()

	if mgr == nil {
		t.Fatal("NewNoneLoggerManager() returned nil")
	}

	if mgr.BaseManager == nil {
		t.Error("NoneLoggerManager.BaseManager is nil")
	}
}

func TestNoneLoggerManager_ManagerName(t *testing.T) {
	mgr := NewNoneLoggerManager()

	expected := "none-logger"
	if got := mgr.ManagerName(); got != expected {
		t.Errorf("NoneLoggerManager.ManagerName() = %v, want %v", got, expected)
	}
}

func TestNoneLoggerManager_GetLogger(t *testing.T) {
	mgr := NewNoneLoggerManager()

	tests := []struct {
		name string
	}{
		{"test-logger"},
		{"another-logger"},
		{""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := mgr.GetLogger(tt.name)

			if logger == nil {
				t.Error("GetLogger() returned nil")
			}

			// Each call should return a new instance
			logger2 := mgr.GetLogger(tt.name)
			if logger == logger2 {
				t.Error("GetLogger() should return a new instance each time")
			}
		})
	}
}

func TestNoneLoggerManager_SetGlobalLevel(t *testing.T) {
	mgr := NewNoneLoggerManager()

	// Should not panic
	mgr.SetGlobalLevel(loglevel.DebugLevel)
	mgr.SetGlobalLevel(loglevel.FatalLevel)
}

func TestNoneLoggerManager_Shutdown(t *testing.T) {
	mgr := NewNoneLoggerManager()

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
			wantErr: false, // NoneLoggerManager doesn't validate context
		},
		{
			name:    "cancelled context",
			ctx:     func() context.Context { ctx, cancel := context.WithCancel(context.Background()); cancel(); return ctx }(),
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := mgr.Shutdown(tt.ctx)
			if (err != nil) != tt.wantErr {
				t.Errorf("NoneLoggerManager.Shutdown() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestNoneLoggerManager_CommonManager(t *testing.T) {
	mgr := NewNoneLoggerManager()

	// Test common.BaseManager interface methods
	if err := mgr.Health(); err != nil {
		t.Errorf("NoneLoggerManager.Health() error = %v, want nil", err)
	}

	if err := mgr.OnStart(); err != nil {
		t.Errorf("NoneLoggerManager.OnStart() error = %v, want nil", err)
	}

	if err := mgr.OnStop(); err != nil {
		t.Errorf("NoneLoggerManager.OnStop() error = %v, want nil", err)
	}
}

func TestNoneLogger_ConcurrentUsage(t *testing.T) {
	logger := NewNoneLogger()
	done := make(chan bool)

	// Concurrent calls to ensure thread safety
	for i := 0; i < 100; i++ {
		go func(i int) {
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("Concurrent usage panicked: %v", r)
				}
			}()

			switch i % 6 {
			case 0:
				logger.Debug("test", "i", i)
			case 1:
				logger.Info("test", "i", i)
			case 2:
				logger.Warn("test", "i", i)
			case 3:
				logger.Error("test", "i", i)
			case 4:
				logger.Fatal("test", "i", i)
			case 5:
				logger.SetLevel(loglevel.LogLevel(i % 5))
			}
			done <- true
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < 100; i++ {
		<-done
	}
}

func TestNoneLoggerManager_ConcurrentUsage(t *testing.T) {
	mgr := NewNoneLoggerManager()
	done := make(chan bool)

	// Concurrent calls to ensure thread safety
	for i := 0; i < 100; i++ {
		go func(i int) {
			_ = mgr.GetLogger("test")
			mgr.SetGlobalLevel(loglevel.LogLevel(i % 5))
			done <- true
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < 100; i++ {
		<-done
	}
}
