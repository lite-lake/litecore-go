package drivers

import (
	"context"
	"testing"

	"com.litelake.litecore/manager/loggermgr/internal/loglevel"
)

func TestNewBaseManager(t *testing.T) {
	name := "test-manager"
	mgr := NewBaseManager(name)

	if mgr == nil {
		t.Fatal("NewBaseManager() returned nil")
	}

	if mgr.ManagerName() != name {
		t.Errorf("ManagerName() = %v, want %v", mgr.ManagerName(), name)
	}
}

func TestBaseManager_ManagerName(t *testing.T) {
	tests := []struct {
		name string
	}{
		{"test-manager-1"},
		{"test-manager-2"},
		{""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mgr := NewBaseManager(tt.name)
			if got := mgr.ManagerName(); got != tt.name {
				t.Errorf("BaseManager.ManagerName() = %v, want %v", got, tt.name)
			}
		})
	}
}

func TestBaseManager_Health(t *testing.T) {
	mgr := NewBaseManager("test-manager")

	if err := mgr.Health(); err != nil {
		t.Errorf("BaseManager.Health() error = %v, want nil", err)
	}
}

func TestBaseManager_OnStart(t *testing.T) {
	mgr := NewBaseManager("test-manager")

	if err := mgr.OnStart(); err != nil {
		t.Errorf("BaseManager.OnStart() error = %v, want nil", err)
	}
}

func TestBaseManager_OnStop(t *testing.T) {
	mgr := NewBaseManager("test-manager")

	if err := mgr.OnStop(); err != nil {
		t.Errorf("BaseManager.OnStop() error = %v, want nil", err)
	}
}

func TestBaseManager_Shutdown(t *testing.T) {
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
			wantErr: false, // BaseManager.Shutdown doesn't validate context
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mgr := NewBaseManager("test-manager")
			err := mgr.Shutdown(tt.ctx)
			if (err != nil) != tt.wantErr {
				t.Errorf("BaseManager.Shutdown() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateContext(t *testing.T) {
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
			wantErr: true,
		},
		{
			name:    "context with timeout",
			ctx:     context.TODO(),
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateContext(tt.ctx)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateContext() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestNewBaseLogger(t *testing.T) {
	name := "test-logger"
	level := loglevel.InfoLevel

	logger := NewBaseLogger(name, level)

	if logger == nil {
		t.Fatal("NewBaseLogger() returned nil")
	}

	if logger.LoggerName() != name {
		t.Errorf("LoggerName() = %v, want %v", logger.LoggerName(), name)
	}

	if logger.GetLevel() != level {
		t.Errorf("GetLevel() = %v, want %v", logger.GetLevel(), level)
	}
}

func TestBaseLogger_LoggerName(t *testing.T) {
	tests := []struct {
		name string
	}{
		{"logger-1"},
		{"logger-2"},
		{""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := NewBaseLogger(tt.name, loglevel.InfoLevel)
			if got := logger.LoggerName(); got != tt.name {
				t.Errorf("BaseLogger.LoggerName() = %v, want %v", got, tt.name)
			}
		})
	}
}

func TestBaseLogger_SetLevel(t *testing.T) {
	logger := NewBaseLogger("test", loglevel.InfoLevel)

	// Set to debug
	logger.SetLevel(loglevel.DebugLevel)
	if got := logger.GetLevel(); got != loglevel.DebugLevel {
		t.Errorf("After SetLevel(DebugLevel), GetLevel() = %v, want %v", got, loglevel.DebugLevel)
	}

	// Set to error
	logger.SetLevel(loglevel.ErrorLevel)
	if got := logger.GetLevel(); got != loglevel.ErrorLevel {
		t.Errorf("After SetLevel(ErrorLevel), GetLevel() = %v, want %v", got, loglevel.ErrorLevel)
	}
}

func TestBaseLogger_GetLevel(t *testing.T) {
	tests := []struct {
		name  string
		level loglevel.LogLevel
	}{
		{"debug level", loglevel.DebugLevel},
		{"info level", loglevel.InfoLevel},
		{"warn level", loglevel.WarnLevel},
		{"error level", loglevel.ErrorLevel},
		{"fatal level", loglevel.FatalLevel},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := NewBaseLogger("test", tt.level)
			if got := logger.GetLevel(); got != tt.level {
				t.Errorf("BaseLogger.GetLevel() = %v, want %v", got, tt.level)
			}
		})
	}
}

func TestBaseLogger_IsEnabled(t *testing.T) {
	tests := []struct {
		name          string
		loggerLevel   loglevel.LogLevel
		checkLevel    loglevel.LogLevel
		expected      bool
	}{
		{"debug logger, check debug - enabled", loglevel.DebugLevel, loglevel.DebugLevel, true},
		{"debug logger, check info - enabled", loglevel.DebugLevel, loglevel.InfoLevel, true},
		{"debug logger, check fatal - enabled", loglevel.DebugLevel, loglevel.FatalLevel, true},
		{"info logger, check debug - disabled", loglevel.InfoLevel, loglevel.DebugLevel, false},
		{"info logger, check info - enabled", loglevel.InfoLevel, loglevel.InfoLevel, true},
		{"info logger, check warn - enabled", loglevel.InfoLevel, loglevel.WarnLevel, true},
		{"warn logger, check debug - disabled", loglevel.WarnLevel, loglevel.DebugLevel, false},
		{"warn logger, check warn - enabled", loglevel.WarnLevel, loglevel.WarnLevel, true},
		{"warn logger, check error - enabled", loglevel.WarnLevel, loglevel.ErrorLevel, true},
		{"error logger, check warn - disabled", loglevel.ErrorLevel, loglevel.WarnLevel, false},
		{"error logger, check error - enabled", loglevel.ErrorLevel, loglevel.ErrorLevel, true},
		{"error logger, check fatal - enabled", loglevel.ErrorLevel, loglevel.FatalLevel, true},
		{"fatal logger, check debug - disabled", loglevel.FatalLevel, loglevel.DebugLevel, false},
		{"fatal logger, check fatal - enabled", loglevel.FatalLevel, loglevel.FatalLevel, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := NewBaseLogger("test", tt.loggerLevel)
			if got := logger.IsEnabled(tt.checkLevel); got != tt.expected {
				t.Errorf("BaseLogger.IsEnabled() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestBaseLogger_ConcurrentAccess(t *testing.T) {
	logger := NewBaseLogger("test", loglevel.InfoLevel)
	done := make(chan bool)

	// Concurrent reads
	for i := 0; i < 10; i++ {
		go func() {
			_ = logger.GetLevel()
			_ = logger.IsEnabled(loglevel.DebugLevel)
			done <- true
		}()
	}

	// Concurrent writes
	for i := 0; i < 10; i++ {
		go func(i int) {
			logger.SetLevel(loglevel.LogLevel(i % 5))
			done <- true
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < 20; i++ {
		<-done
	}

	// Logger should still be functional
	_ = logger.GetLevel()
	logger.SetLevel(loglevel.DebugLevel)
}
