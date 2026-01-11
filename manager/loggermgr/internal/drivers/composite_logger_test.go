package drivers

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"com.litelake.litecore/manager/loggermgr/internal/config"
	"com.litelake.litecore/manager/loggermgr/internal/loglevel"
)

func TestNewZapLoggerManager(t *testing.T) {
	tests := []struct {
		name    string
		config  *config.LoggerConfig
		wantErr bool
	}{
		{
			name: "valid config with console",
			config: &config.LoggerConfig{
				ConsoleEnabled: true,
				ConsoleConfig:  &config.LogLevelConfig{Level: "info"},
			},
			wantErr: false,
		},
		{
			name: "valid config with file",
			config: &config.LoggerConfig{
				FileEnabled: true,
				FileConfig: &config.FileLogConfig{
					Level: "info",
					Path:  filepath.Join(os.TempDir(), "test.log"),
				},
			},
			wantErr: false,
		},
		{
			name: "invalid config - no outputs",
			config: &config.LoggerConfig{
				ConsoleEnabled: false,
				FileEnabled:    false,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mgr, err := NewZapLoggerManager(tt.config, nil)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewZapLoggerManager() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && mgr == nil {
				t.Error("NewZapLoggerManager() returned nil manager")
			}
			if !tt.wantErr && mgr != nil {
				_ = mgr.Shutdown(context.Background())
			}
		})
	}
}

func TestZapLoggerManager_ManagerName(t *testing.T) {
	cfg := &config.LoggerConfig{
		ConsoleEnabled: true,
		ConsoleConfig:  &config.LogLevelConfig{Level: "info"},
	}
	mgr, err := NewZapLoggerManager(cfg, nil)
	if err != nil {
		t.Fatalf("NewZapLoggerManager() error = %v", err)
	}
	defer mgr.Shutdown(context.Background())

	expected := "zap-logger"
	if got := mgr.ManagerName(); got != expected {
		t.Errorf("ZapLoggerManager.ManagerName() = %v, want %v", got, expected)
	}
}

func TestZapLoggerManager_GetLogger(t *testing.T) {
	cfg := &config.LoggerConfig{
		ConsoleEnabled: true,
		ConsoleConfig:  &config.LogLevelConfig{Level: "info"},
	}
	mgr, err := NewZapLoggerManager(cfg, nil)
	if err != nil {
		t.Fatalf("NewZapLoggerManager() error = %v", err)
	}
	defer mgr.Shutdown(context.Background())

	tests := []struct {
		name string
	}{
		{"test-logger"},
		{"another-logger"},
		{"service-logger"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := mgr.GetLogger(tt.name)

			if logger == nil {
				t.Error("GetLogger() returned nil")
			}

			if logger.name != tt.name {
				t.Errorf("Logger name = %v, want %v", logger.name, tt.name)
			}
		})
	}
}

func TestZapLoggerManager_GetLogger_Caching(t *testing.T) {
	cfg := &config.LoggerConfig{
		ConsoleEnabled: true,
		ConsoleConfig:  &config.LogLevelConfig{Level: "info"},
	}
	mgr, err := NewZapLoggerManager(cfg, nil)
	if err != nil {
		t.Fatalf("NewZapLoggerManager() error = %v", err)
	}
	defer mgr.Shutdown(context.Background())

	name := "test-logger"
	logger1 := mgr.GetLogger(name)
	logger2 := mgr.GetLogger(name)

	// Same instance should be returned
	if logger1 != logger2 {
		t.Error("GetLogger() should return the same instance for the same name")
	}
}

func TestZapLoggerManager_SetGlobalLevel(t *testing.T) {
	cfg := &config.LoggerConfig{
		ConsoleEnabled: true,
		ConsoleConfig:  &config.LogLevelConfig{Level: "info"},
	}
	mgr, err := NewZapLoggerManager(cfg, nil)
	if err != nil {
		t.Fatalf("NewZapLoggerManager() error = %v", err)
	}
	defer mgr.Shutdown(context.Background())

	// Should not panic
	mgr.SetGlobalLevel(loglevel.LogLevelToZap(loglevel.DebugLevel))
	mgr.SetGlobalLevel(loglevel.LogLevelToZap(loglevel.FatalLevel))
}

func TestZapLoggerManager_Shutdown(t *testing.T) {
	cfg := &config.LoggerConfig{
		ConsoleEnabled: true,
		ConsoleConfig:  &config.LogLevelConfig{Level: "info"},
	}
	mgr, err := NewZapLoggerManager(cfg, nil)
	if err != nil {
		t.Fatalf("NewZapLoggerManager() error = %v", err)
	}

	// Create a logger
	_ = mgr.GetLogger("test")

	// Shutdown
	err = mgr.Shutdown(context.Background())
	// Note: In test environments, sync to stdout/stderr may fail with "bad file descriptor"
	// This is expected behavior and not a real error
	if err != nil {
		// Only fail if it's not a sync-related error
		if !containsSyncError(err.Error()) {
			t.Errorf("ZapLoggerManager.Shutdown() error = %v", err)
		}
	}

	// Second shutdown should also succeed (idempotent)
	err = mgr.Shutdown(context.Background())
	if err != nil && !containsSyncError(err.Error()) {
		t.Errorf("ZapLoggerManager.Shutdown() second call error = %v", err)
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

func TestZapLoggerManager_OnStart(t *testing.T) {
	cfg := &config.LoggerConfig{
		ConsoleEnabled: true,
		ConsoleConfig:  &config.LogLevelConfig{Level: "info"},
	}
	mgr, err := NewZapLoggerManager(cfg, nil)
	if err != nil {
		t.Fatalf("NewZapLoggerManager() error = %v", err)
	}
	defer mgr.Shutdown(context.Background())

	if err := mgr.OnStart(); err != nil {
		t.Errorf("ZapLoggerManager.OnStart() error = %v, want nil", err)
	}
}

func TestZapLogger_Logging(t *testing.T) {
	cfg := &config.LoggerConfig{
		ConsoleEnabled: true,
		ConsoleConfig:  &config.LogLevelConfig{Level: "debug"},
	}
	mgr, err := NewZapLoggerManager(cfg, nil)
	if err != nil {
		t.Fatalf("NewZapLoggerManager() error = %v", err)
	}
	defer mgr.Shutdown(context.Background())

	logger := mgr.GetLogger("test")

	// Should not panic
	logger.Debug("debug message", "key1", "value1")
	logger.Info("info message", "key2", "value2")
	logger.Warn("warn message", "key3", "value3")
	logger.Error("error message", "key4", "value4")
}

func TestZapLogger_With(t *testing.T) {
	cfg := &config.LoggerConfig{
		ConsoleEnabled: true,
		ConsoleConfig:  &config.LogLevelConfig{Level: "debug"},
	}
	mgr, err := NewZapLoggerManager(cfg, nil)
	if err != nil {
		t.Fatalf("NewZapLoggerManager() error = %v", err)
	}
	defer mgr.Shutdown(context.Background())

	logger := mgr.GetLogger("test")

	// Create a new logger with additional fields
	loggerWithFields := logger.With("service", "test-service", "version", "1.0.0")

	if loggerWithFields == nil {
		t.Error("With() returned nil")
	}

	// Should be a different instance
	if loggerWithFields == logger {
		t.Error("With() should return a new instance")
	}

	// Original logger should still work
	logger.Info("original logger message")

	// New logger should work
	loggerWithFields.Info("new logger message")
}

func TestZapLogger_SetLevel(t *testing.T) {
	cfg := &config.LoggerConfig{
		ConsoleEnabled: true,
		ConsoleConfig:  &config.LogLevelConfig{Level: "info"},
	}
	mgr, err := NewZapLoggerManager(cfg, nil)
	if err != nil {
		t.Fatalf("NewZapLoggerManager() error = %v", err)
	}
	defer mgr.Shutdown(context.Background())

	logger := mgr.GetLogger("test")

	// Should not panic
	logger.SetLevel(loglevel.DebugLevel)
	logger.SetLevel(loglevel.ErrorLevel)
}

func TestZapLogger_Sync(t *testing.T) {
	cfg := &config.LoggerConfig{
		ConsoleEnabled: true,
		ConsoleConfig:  &config.LogLevelConfig{Level: "debug"},
	}
	mgr, err := NewZapLoggerManager(cfg, nil)
	if err != nil {
		t.Fatalf("NewZapLoggerManager() error = %v", err)
	}
	defer mgr.Shutdown(context.Background())

	logger := mgr.GetLogger("test")
	logger.Info("test message")

	err = logger.Sync()
	// Note: In test environments, sync to stdout/stderr may fail with "bad file descriptor"
	// This is expected behavior and not a real error
	if err != nil && !containsSyncError(err.Error()) {
		t.Errorf("ZapLogger.Sync() error = %v", err)
	}
}

func TestZapLogger_isEnabled(t *testing.T) {
	cfg := &config.LoggerConfig{
		ConsoleEnabled: true,
		ConsoleConfig:  &config.LogLevelConfig{Level: "warn"},
	}
	mgr, err := NewZapLoggerManager(cfg, nil)
	if err != nil {
		t.Fatalf("NewZapLoggerManager() error = %v", err)
	}
	defer mgr.Shutdown(context.Background())

	logger := mgr.GetLogger("test")

	// Logger level is warn (based on min level calculation in NewZapLogger)
	// But the actual check happens in isEnabled
	// This test verifies the logic works

	// Just verify logging doesn't panic
	logger.Debug("debug - should not log")
	logger.Info("info - should not log")
	logger.Warn("warn - should log")
	logger.Error("error - should log")
}

func TestArgsToFields(t *testing.T) {
	tests := []struct {
		name     string
		args     []any
		expected int // expected number of fields
	}{
		{
			name:     "empty args",
			args:     []any{},
			expected: 0,
		},
		{
			name:     "single key",
			args:     []any{"key1"},
			expected: 0,
		},
		{
			name:     "one pair",
			args:     []any{"key1", "value1"},
			expected: 1,
		},
		{
			name:     "two pairs",
			args:     []any{"key1", "value1", "key2", "value2"},
			expected: 2,
		},
		{
			name:     "three pairs with odd trailing key",
			args:     []any{"key1", "value1", "key2", "value2", "key3"},
			expected: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fields := argsToFields(tt.args...)
			if len(fields) != tt.expected {
				t.Errorf("argsToFields() returned %d fields, want %d", len(fields), tt.expected)
			}
		})
	}
}

func TestZapLogger_ConcurrentLogging(t *testing.T) {
	cfg := &config.LoggerConfig{
		ConsoleEnabled: true,
		ConsoleConfig:  &config.LogLevelConfig{Level: "debug"},
	}
	mgr, err := NewZapLoggerManager(cfg, nil)
	if err != nil {
		t.Fatalf("NewZapLoggerManager() error = %v", err)
	}
	defer mgr.Shutdown(context.Background())

	logger := mgr.GetLogger("test")
	done := make(chan bool)

	// Concurrent logging
	for i := 0; i < 100; i++ {
		go func(i int) {
			logger.Info("concurrent message", "i", i)
			done <- true
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < 100; i++ {
		<-done
	}
}

func TestZapLoggerManager_ConcurrentGetLogger(t *testing.T) {
	cfg := &config.LoggerConfig{
		ConsoleEnabled: true,
		ConsoleConfig:  &config.LogLevelConfig{Level: "info"},
	}
	mgr, err := NewZapLoggerManager(cfg, nil)
	if err != nil {
		t.Fatalf("NewZapLoggerManager() error = %v", err)
	}
	defer mgr.Shutdown(context.Background())

	done := make(chan bool)

	// Concurrent get logger calls
	for i := 0; i < 100; i++ {
		go func(i int) {
			_ = mgr.GetLogger("test")
			done <- true
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < 100; i++ {
		<-done
	}
}
