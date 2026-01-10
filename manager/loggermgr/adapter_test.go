package loggermgr

import (
	"testing"

	"com.litelake.litecore/manager/loggermgr/internal/config"
	"com.litelake.litecore/manager/loggermgr/internal/drivers"
)

func TestNewLoggerAdapter(t *testing.T) {
	// Create a mock zap logger using the actual implementation
	cfg := &config.LoggerConfig{
		ConsoleEnabled: true,
		ConsoleConfig:  &config.LogLevelConfig{Level: "info"},
	}
	zapLogger, err := drivers.NewZapLogger("test", cfg, nil)
	if err != nil {
		t.Fatalf("Failed to create zap logger: %v", err)
	}

	adapter := NewLoggerAdapter(zapLogger)

	if adapter == nil {
		t.Fatal("NewLoggerAdapter() returned nil")
	}

	if adapter.driver != zapLogger {
		t.Error("NewLoggerAdapter() did not set the driver correctly")
	}
}

func TestLoggerAdapter_LoggingMethods(t *testing.T) {
	cfg := &config.LoggerConfig{
		ConsoleEnabled: true,
		ConsoleConfig:  &config.LogLevelConfig{Level: "debug"},
	}
	zapLogger, err := drivers.NewZapLogger("test", cfg, nil)
	if err != nil {
		t.Fatalf("Failed to create zap logger: %v", err)
	}

	adapter := NewLoggerAdapter(zapLogger)

	// Should not panic
	tests := []struct {
		name string
		fn   func()
	}{
		{"Debug", func() { adapter.Debug("debug message", "key", "value") }},
		{"Info", func() { adapter.Info("info message", "key", "value") }},
		{"Warn", func() { adapter.Warn("warn message", "key", "value") }},
		{"Error", func() { adapter.Error("error message", "key", "value") }},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.fn()
		})
	}
}

func TestLoggerAdapter_With(t *testing.T) {
	cfg := &config.LoggerConfig{
		ConsoleEnabled: true,
		ConsoleConfig:  &config.LogLevelConfig{Level: "debug"},
	}
	zapLogger, err := drivers.NewZapLogger("test", cfg, nil)
	if err != nil {
		t.Fatalf("Failed to create zap logger: %v", err)
	}

	adapter := NewLoggerAdapter(zapLogger)
	newAdapter := adapter.With("service", "test-service", "version", "1.0.0")

	if newAdapter == nil {
		t.Fatal("With() returned nil")
	}

	// Should return a new adapter instance
	if newAdapter == adapter {
		t.Error("With() should return a new adapter instance")
	}

	// The new adapter should be a LoggerAdapter
	if _, ok := newAdapter.(*LoggerAdapter); !ok {
		t.Error("With() should return a LoggerAdapter")
	}
}

func TestLoggerAdapter_SetLevel(t *testing.T) {
	cfg := &config.LoggerConfig{
		ConsoleEnabled: true,
		ConsoleConfig:  &config.LogLevelConfig{Level: "info"},
	}
	zapLogger, err := drivers.NewZapLogger("test", cfg, nil)
	if err != nil {
		t.Fatalf("Failed to create zap logger: %v", err)
	}

	adapter := NewLoggerAdapter(zapLogger)

	// Should not panic
	adapter.SetLevel(DebugLevel)
	adapter.SetLevel(InfoLevel)
	adapter.SetLevel(WarnLevel)
	adapter.SetLevel(ErrorLevel)
	adapter.SetLevel(FatalLevel)
}

func TestLoggerAdapter_ConcurrentUsage(t *testing.T) {
	cfg := &config.LoggerConfig{
		ConsoleEnabled: true,
		ConsoleConfig:  &config.LogLevelConfig{Level: "debug"},
	}
	zapLogger, err := drivers.NewZapLogger("test", cfg, nil)
	if err != nil {
		t.Fatalf("Failed to create zap logger: %v", err)
	}

	adapter := NewLoggerAdapter(zapLogger)
	done := make(chan bool)

	// Concurrent usage
	for i := 0; i < 50; i++ {
		go func(i int) {
			adapter.Info("concurrent message", "i", i)
			_ = adapter.With("key", "value")
			adapter.SetLevel(LogLevel(i % 5))
			done <- true
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < 50; i++ {
		<-done
	}
}

func TestLoggerAdapter_Interface(t *testing.T) {
	// Verify LoggerAdapter implements Logger interface
	cfg := &config.LoggerConfig{
		ConsoleEnabled: true,
		ConsoleConfig:  &config.LogLevelConfig{Level: "info"},
	}
	zapLogger, err := drivers.NewZapLogger("test", cfg, nil)
	if err != nil {
		t.Fatalf("Failed to create zap logger: %v", err)
	}

	var logger Logger = NewLoggerAdapter(zapLogger)

	// Should implement all interface methods
	logger.Debug("test")
	logger.Info("test")
	logger.Warn("test")
	logger.Error("test")
	logger.SetLevel(InfoLevel)
	_ = logger.With("key", "value")
}

func TestLoggerAdapter_Fatal(t *testing.T) {
	cfg := &config.LoggerConfig{
		ConsoleEnabled: true,
		ConsoleConfig:  &config.LogLevelConfig{Level: "debug"},
	}
	zapLogger, err := drivers.NewZapLogger("test", cfg, nil)
	if err != nil {
		t.Fatalf("Failed to create zap logger: %v", err)
	}

	_ = NewLoggerAdapter(zapLogger)

	// Note: Fatal calls os.Exit, so we cannot actually test the full behavior
	// We can only verify the method exists and can be called
	// In a real scenario, this would exit the program
	// We'll just verify it doesn't panic before the underlying driver.Fatal is called
	// Since we can't actually let it exit, we skip this test in the actual run
	t.Skip("Fatal method calls os.Exit, skipping to prevent test termination")
}

func TestLoggerAdapter_WithMultipleChains(t *testing.T) {
	cfg := &config.LoggerConfig{
		ConsoleEnabled: true,
		ConsoleConfig:  &config.LogLevelConfig{Level: "debug"},
	}
	zapLogger, err := drivers.NewZapLogger("test", cfg, nil)
	if err != nil {
		t.Fatalf("Failed to create zap logger: %v", err)
	}

	adapter := NewLoggerAdapter(zapLogger)

	// Chain multiple With calls
	adapter1 := adapter.With("service", "test-service")
	adapter2 := adapter1.With("version", "1.0.0")
	adapter3 := adapter2.With("environment", "production")

	if adapter1 == nil || adapter2 == nil || adapter3 == nil {
		t.Error("With() should not return nil")
	}

	// Verify each adapter is a LoggerAdapter
	if _, ok := adapter1.(*LoggerAdapter); !ok {
		t.Error("adapter1 should be a LoggerAdapter")
	}
	if _, ok := adapter2.(*LoggerAdapter); !ok {
		t.Error("adapter2 should be a LoggerAdapter")
	}
	if _, ok := adapter3.(*LoggerAdapter); !ok {
		t.Error("adapter3 should be a LoggerAdapter")
	}

	// Verify they are different instances
	if adapter1 == adapter2 || adapter2 == adapter3 {
		t.Error("Chained With() calls should return new instances")
	}
}

func TestLoggerAdapter_WithSingleKey(t *testing.T) {
	cfg := &config.LoggerConfig{
		ConsoleEnabled: true,
		ConsoleConfig:  &config.LogLevelConfig{Level: "debug"},
	}
	zapLogger, err := drivers.NewZapLogger("test", cfg, nil)
	if err != nil {
		t.Fatalf("Failed to create zap logger: %v", err)
	}

	adapter := NewLoggerAdapter(zapLogger)

	// With single key (should be handled gracefully)
	result := adapter.With("single")
	if result == nil {
		t.Error("With() with single key should not return nil")
	}
}

func TestLoggerAdapter_MultipleLogLevels(t *testing.T) {
	cfg := &config.LoggerConfig{
		ConsoleEnabled: true,
		ConsoleConfig:  &config.LogLevelConfig{Level: "info"},
	}
	zapLogger, err := drivers.NewZapLogger("test", cfg, nil)
	if err != nil {
		t.Fatalf("Failed to create zap logger: %v", err)
	}

	adapter := NewLoggerAdapter(zapLogger)

	// Test all log levels can be set
	levels := []LogLevel{DebugLevel, InfoLevel, WarnLevel, ErrorLevel, FatalLevel}
	for _, level := range levels {
		adapter.SetLevel(level)
	}
}
