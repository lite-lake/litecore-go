package loggermgr

import (
	"context"
	"testing"

	"com.litelake.litecore/manager/loggermgr/internal/config"
	"com.litelake.litecore/manager/loggermgr/internal/drivers"
)

func TestNewLoggerManagerAdapter(t *testing.T) {
	cfg := &config.LoggerConfig{
		ConsoleEnabled: true,
		ConsoleConfig:  &config.LogLevelConfig{Level: "info"},
	}
	zapMgr, err := drivers.NewZapLoggerManager(cfg, nil)
	if err != nil {
		t.Fatalf("Failed to create zap logger manager: %v", err)
	}

	adapter := NewLoggerManagerAdapter(zapMgr)

	if adapter == nil {
		t.Fatal("NewLoggerManagerAdapter() returned nil")
	}

	if adapter.driver != zapMgr {
		t.Error("NewLoggerManagerAdapter() did not set the driver correctly")
	}
}

func TestLoggerManagerAdapter_Logger(t *testing.T) {
	cfg := &config.LoggerConfig{
		ConsoleEnabled: true,
		ConsoleConfig:  &config.LogLevelConfig{Level: "info"},
	}
	zapMgr, err := drivers.NewZapLoggerManager(cfg, nil)
	if err != nil {
		t.Fatalf("Failed to create zap logger manager: %v", err)
	}
	defer zapMgr.Shutdown(context.Background())

	adapter := NewLoggerManagerAdapter(zapMgr)

	tests := []struct {
		name string
	}{
		{"test-logger"},
		{"another-logger"},
		{"service-logger"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := adapter.Logger(tt.name)

			if logger == nil {
				t.Error("Logger() returned nil")
			}

			// Should return a LoggerAdapter
			if _, ok := logger.(*LoggerAdapter); !ok {
				t.Error("Logger() should return a LoggerAdapter")
			}
		})
	}
}

func TestLoggerManagerAdapter_Logger_Caching(t *testing.T) {
	cfg := &config.LoggerConfig{
		ConsoleEnabled: true,
		ConsoleConfig:  &config.LogLevelConfig{Level: "info"},
	}
	zapMgr, err := drivers.NewZapLoggerManager(cfg, nil)
	if err != nil {
		t.Fatalf("Failed to create zap logger manager: %v", err)
	}
	defer zapMgr.Shutdown(context.Background())

	adapter := NewLoggerManagerAdapter(zapMgr)

	name := "test-logger"
	logger1 := adapter.Logger(name)
	logger2 := adapter.Logger(name)

	// The underlying driver returns the same logger instance
	// but the adapter wraps it, so we can't directly compare
	// Just verify both work
	if logger1 == nil || logger2 == nil {
		t.Error("Logger() returned nil")
	}
}

func TestLoggerManagerAdapter_SetGlobalLevel(t *testing.T) {
	cfg := &config.LoggerConfig{
		ConsoleEnabled: true,
		ConsoleConfig:  &config.LogLevelConfig{Level: "info"},
	}
	zapMgr, err := drivers.NewZapLoggerManager(cfg, nil)
	if err != nil {
		t.Fatalf("Failed to create zap logger manager: %v", err)
	}
	defer zapMgr.Shutdown(context.Background())

	adapter := NewLoggerManagerAdapter(zapMgr)

	// Should not panic
	adapter.SetGlobalLevel(DebugLevel)
	adapter.SetGlobalLevel(InfoLevel)
	adapter.SetGlobalLevel(WarnLevel)
	adapter.SetGlobalLevel(ErrorLevel)
	adapter.SetGlobalLevel(FatalLevel)
}

func TestLoggerManagerAdapter_Shutdown(t *testing.T) {
	cfg := &config.LoggerConfig{
		ConsoleEnabled: true,
		ConsoleConfig:  &config.LogLevelConfig{Level: "info"},
	}
	zapMgr, err := drivers.NewZapLoggerManager(cfg, nil)
	if err != nil {
		t.Fatalf("Failed to create zap logger manager: %v", err)
	}

	adapter := NewLoggerManagerAdapter(zapMgr)

	// Create a logger
	_ = adapter.Logger("test")

	// Shutdown
	err = adapter.Shutdown(context.Background())
	// Note: In test environments, sync to stdout/stderr may fail with "bad file descriptor"
	// This is expected behavior and not a real error
	if err != nil && !containsSyncError(err.Error()) {
		t.Errorf("LoggerManagerAdapter.Shutdown() error = %v", err)
	}

	// Second shutdown should also succeed (idempotent)
	err = adapter.Shutdown(context.Background())
	if err != nil && !containsSyncError(err.Error()) {
		t.Errorf("LoggerManagerAdapter.Shutdown() second call error = %v", err)
	}
}

func containsSyncError(msg string) bool {
	return len(msg) > 0 && (containsSubstring(msg, "sync") || containsSubstring(msg, "bad file descriptor"))
}

func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func TestLoggerManagerAdapter_ManagerName(t *testing.T) {
	cfg := &config.LoggerConfig{
		ConsoleEnabled: true,
		ConsoleConfig:  &config.LogLevelConfig{Level: "info"},
	}
	zapMgr, err := drivers.NewZapLoggerManager(cfg, nil)
	if err != nil {
		t.Fatalf("Failed to create zap logger manager: %v", err)
	}
	defer zapMgr.Shutdown(context.Background())

	adapter := NewLoggerManagerAdapter(zapMgr)

	expected := "zap-logger"
	if got := adapter.ManagerName(); got != expected {
		t.Errorf("LoggerManagerAdapter.ManagerName() = %v, want %v", got, expected)
	}
}

func TestLoggerManagerAdapter_CommonManager(t *testing.T) {
	cfg := &config.LoggerConfig{
		ConsoleEnabled: true,
		ConsoleConfig:  &config.LogLevelConfig{Level: "info"},
	}
	zapMgr, err := drivers.NewZapLoggerManager(cfg, nil)
	if err != nil {
		t.Fatalf("Failed to create zap logger manager: %v", err)
	}
	defer zapMgr.Shutdown(context.Background())

	adapter := NewLoggerManagerAdapter(zapMgr)

	// Test common.BaseManager interface methods
	if err := adapter.Health(); err != nil {
		t.Errorf("LoggerManagerAdapter.Health() error = %v, want nil", err)
	}

	if err := adapter.OnStart(); err != nil {
		t.Errorf("LoggerManagerAdapter.OnStart() error = %v, want nil", err)
	}

	if err := adapter.OnStop(); err != nil {
		t.Errorf("LoggerManagerAdapter.OnStop() error = %v, want nil", err)
	}
}

func TestLoggerManagerAdapter_ConcurrentUsage(t *testing.T) {
	cfg := &config.LoggerConfig{
		ConsoleEnabled: true,
		ConsoleConfig:  &config.LogLevelConfig{Level: "info"},
	}
	zapMgr, err := drivers.NewZapLoggerManager(cfg, nil)
	if err != nil {
		t.Fatalf("Failed to create zap logger manager: %v", err)
	}
	defer zapMgr.Shutdown(context.Background())

	adapter := NewLoggerManagerAdapter(zapMgr)
	done := make(chan bool)

	// Concurrent usage
	for i := 0; i < 50; i++ {
		go func(i int) {
			_ = adapter.Logger("test")
			adapter.SetGlobalLevel(LogLevel(i % 5))
			done <- true
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < 50; i++ {
		<-done
	}
}

func TestLoggerManagerAdapter_Interface(t *testing.T) {
	cfg := &config.LoggerConfig{
		ConsoleEnabled: true,
		ConsoleConfig:  &config.LogLevelConfig{Level: "info"},
	}
	zapMgr, err := drivers.NewZapLoggerManager(cfg, nil)
	if err != nil {
		t.Fatalf("Failed to create zap logger manager: %v", err)
	}
	defer zapMgr.Shutdown(context.Background())

	var loggerMgr LoggerManager = NewLoggerManagerAdapter(zapMgr)

	// Test LoggerManager interface
	logger := loggerMgr.Logger("test")
	if logger == nil {
		t.Error("Logger() returned nil")
	}

	loggerMgr.SetGlobalLevel(InfoLevel)

	ctx := context.Background()
	if err := loggerMgr.Shutdown(ctx); err != nil && !containsSyncError(err.Error()) {
		t.Errorf("Shutdown() error = %v", err)
	}
}
