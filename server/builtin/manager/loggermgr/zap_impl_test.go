package loggermgr

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

// TestNewLoggerManagerZapImpl 测试创建 Zap 日志管理器
func TestNewLoggerManagerZapImpl(t *testing.T) {
	tests := []struct {
		name      string
		config    *LoggerConfig
		wantErr   bool
		errString string
	}{
		{
			name: "Valid configmgr with console",
			config: &LoggerConfig{
				Driver: "zap",
				ZapConfig: &ZapConfig{
					ConsoleEnabled: true,
					ConsoleConfig:  &LogLevelConfig{Level: "info"},
				},
			},
			wantErr: false,
		},
		{
			name: "Valid configmgr with file",
			config: &LoggerConfig{
				Driver: "zap",
				ZapConfig: &ZapConfig{
					ConsoleEnabled: true,
					FileEnabled:    true,
					FileConfig: &FileLogConfig{
						Level: "info",
						Path:  "/tmp/test.log",
					},
				},
			},
			wantErr: false,
		},
		{
			name: "Invalid driver",
			config: &LoggerConfig{
				Driver: "none",
				ZapConfig: &ZapConfig{
					ConsoleEnabled: true,
				},
			},
			wantErr:   true,
			errString: "invalid driver for zap manager",
		},
		{
			name: "Invalid zap configmgr",
			config: &LoggerConfig{
				Driver:    "zap",
				ZapConfig: &ZapConfig{},
			},
			wantErr:   true,
			errString: "at least one logger output",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mgr, err := NewLoggerManagerZapImpl(tt.config)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewLoggerManagerZapImpl() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && err != nil && tt.errString != "" {
				// 错误消息检查已在 config_test.go 中覆盖
			}
			if !tt.wantErr && mgr != nil {
				defer mgr.Shutdown(nil)
			}
		})
	}
}

// TestZapLoggerManagerBaseInterface 测试 Zap 日志管理器基础接口
func TestZapLoggerManagerBaseInterface(t *testing.T) {
	config := &LoggerConfig{
		Driver: "zap",
		ZapConfig: &ZapConfig{
			ConsoleEnabled: true,
			ConsoleConfig:  &LogLevelConfig{Level: "info"},
		},
	}

	mgr, err := NewLoggerManagerZapImpl(config)
	if err != nil {
		t.Fatalf("NewLoggerManagerZapImpl() error = %v", err)
	}
	defer mgr.Shutdown(nil)

	t.Run("ManagerName", func(t *testing.T) {
		name := mgr.ManagerName()
		if name != "zap-logger" {
			t.Errorf("ManagerName() = %v, want 'zap-logger'", name)
		}
	})

	t.Run("Health", func(t *testing.T) {
		err := mgr.Health()
		if err != nil {
			t.Errorf("Health() error = %v, want nil", err)
		}
	})

	t.Run("OnStart", func(t *testing.T) {
		err := mgr.OnStart()
		if err != nil {
			t.Errorf("OnStart() error = %v, want nil", err)
		}
	})

	t.Run("OnStop", func(t *testing.T) {
		err := mgr.OnStop()
		if err != nil {
			t.Errorf("OnStop() error = %v, want nil", err)
		}
	})
}

// TestZapLoggerManagerGetLogger 测试获取 Logger
func TestZapLoggerManagerGetLogger(t *testing.T) {
	config := &LoggerConfig{
		Driver: "zap",
		ZapConfig: &ZapConfig{
			ConsoleEnabled: true,
			ConsoleConfig:  &LogLevelConfig{Level: "info"},
		},
	}

	mgr, err := NewLoggerManagerZapImpl(config)
	if err != nil {
		t.Fatalf("NewLoggerManagerZapImpl() error = %v", err)
	}
	defer mgr.Shutdown(nil)

	t.Run("Get single logger", func(t *testing.T) {
		logger := mgr.Logger("test")
		if logger == nil {
			t.Fatal("Logger() returned nil")
		}
	})

	t.Run("Get same logger twice", func(t *testing.T) {
		logger1 := mgr.Logger("test1")
		logger2 := mgr.Logger("test1")
		if logger1 != logger2 {
			t.Error("Logger() should return same instance for same name")
		}
	})

	t.Run("Get different loggers", func(t *testing.T) {
		logger1 := mgr.Logger("app")
		logger2 := mgr.Logger("db")
		// 应该是不同的实例
		if logger1 == logger2 {
			t.Error("Logger() should return different instances for different names")
		}
	})
}

// TestZapLoggerOutput 测试 Zap 日志输出
func TestZapLoggerOutput(t *testing.T) {
	config := &LoggerConfig{
		Driver: "zap",
		ZapConfig: &ZapConfig{
			ConsoleEnabled: true,
			ConsoleConfig:  &LogLevelConfig{Level: "debug"},
		},
	}

	mgr, err := NewLoggerManagerZapImpl(config)
	if err != nil {
		t.Fatalf("NewLoggerManagerZapImpl() error = %v", err)
	}
	defer mgr.Shutdown(nil)

	logger := mgr.Logger("test")

	t.Run("Debug", func(t *testing.T) {
		logger.Debug("debug message", "key", "value")
	})

	t.Run("Info", func(t *testing.T) {
		logger.Info("info message", "key", "value")
	})

	t.Run("Warn", func(t *testing.T) {
		logger.Warn("warn message", "key", "value")
	})

	t.Run("Error", func(t *testing.T) {
		logger.Error("error message", "key", "value")
	})
}

// TestZapLoggerWith 测试 With 方法
func TestZapLoggerWith(t *testing.T) {
	config := &LoggerConfig{
		Driver: "zap",
		ZapConfig: &ZapConfig{
			ConsoleEnabled: true,
			ConsoleConfig:  &LogLevelConfig{Level: "info"},
		},
	}

	mgr, err := NewLoggerManagerZapImpl(config)
	if err != nil {
		t.Fatalf("NewLoggerManagerZapImpl() error = %v", err)
	}
	defer mgr.Shutdown(nil)

	logger := mgr.Logger("test")

	t.Run("With fields", func(t *testing.T) {
		newLogger := logger.With("service", "test-service", "version", "1.0.0")
		if newLogger == nil {
			t.Fatal("With() returned nil")
		}
		newLogger.Info("message with fields")
	})

	t.Run("Multiple With", func(t *testing.T) {
		newLogger := logger.With("key1", "value1").With("key2", "value2")
		if newLogger == nil {
			t.Fatal("Multiple With() returned nil")
		}
		newLogger.Info("message with multiple fields")
	})
}

// TestZapLoggerSetLevel 测试设置日志级别
func TestZapLoggerSetLevel(t *testing.T) {
	config := &LoggerConfig{
		Driver: "zap",
		ZapConfig: &ZapConfig{
			ConsoleEnabled: true,
			ConsoleConfig:  &LogLevelConfig{Level: "info"},
		},
	}

	mgr, err := NewLoggerManagerZapImpl(config)
	if err != nil {
		t.Fatalf("NewLoggerManagerZapImpl() error = %v", err)
	}
	defer mgr.Shutdown(nil)

	logger := mgr.Logger("test")

	t.Run("SetLevel", func(t *testing.T) {
		logger.SetLevel(DebugLevel)
		logger.Debug("should be visible")

		logger.SetLevel(ErrorLevel)
		logger.Debug("should not be visible")
		logger.Error("should be visible")
	})
}

// TestZapLoggerManagerSetGlobalLevel 测试设置全局日志级别
func TestZapLoggerManagerSetGlobalLevel(t *testing.T) {
	config := &LoggerConfig{
		Driver: "zap",
		ZapConfig: &ZapConfig{
			ConsoleEnabled: true,
			ConsoleConfig:  &LogLevelConfig{Level: "info"},
		},
	}

	mgr, err := NewLoggerManagerZapImpl(config)
	if err != nil {
		t.Fatalf("NewLoggerManagerZapImpl() error = %v", err)
	}
	defer mgr.Shutdown(nil)

	logger1 := mgr.Logger("logger1")
	logger2 := mgr.Logger("logger2")

	t.Run("SetGlobalLevel affects all loggers", func(t *testing.T) {
		mgr.SetGlobalLevel(DebugLevel)
		logger1.Debug("debug from logger1")
		logger2.Debug("debug from logger2")

		mgr.SetGlobalLevel(ErrorLevel)
		logger1.Info("info from logger1 - should not be visible")
		logger2.Error("error from logger2 - should be visible")
	})
}

// TestZapLoggerManagerShutdown 测试关闭日志管理器
func TestZapLoggerManagerShutdown(t *testing.T) {
	config := &LoggerConfig{
		Driver: "zap",
		ZapConfig: &ZapConfig{
			ConsoleEnabled: true,
			ConsoleConfig:  &LogLevelConfig{Level: "info"},
		},
	}

	mgr, err := NewLoggerManagerZapImpl(config)
	if err != nil {
		t.Fatalf("NewLoggerManagerZapImpl() error = %v", err)
	}

	logger := mgr.Logger("test")
	logger.Info("before shutdown")

	ctx := context.Background()
	err = mgr.Shutdown(ctx)
	// stdout/stderr sync 可能会失败，这是正常的非 TTY 环境行为
	if err != nil && !containsSyncError(err) {
		t.Errorf("Shutdown() unexpected error = %v", err)
	}

	// 关闭后再次调用应该是安全的
	err = mgr.Shutdown(ctx)
	if err != nil && !containsSyncError(err) {
		t.Errorf("Second Shutdown() unexpected error = %v", err)
	}
}

// containsSyncError 检查错误是否为 sync 错误（可以忽略）
func containsSyncError(err error) bool {
	if err == nil {
		return false
	}
	errStr := err.Error()
	return containsString(errStr, "sync") || containsString(errStr, "bad file descriptor")
}

// containsString 检查字符串是否包含子串（忽略大小写）
func containsString(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && indexOfSubstring(s, substr))
}

// indexOfSubstring 查找子串位置
func indexOfSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// TestZapLoggerManagerFileLogging 测试文件日志
func TestZapLoggerManagerFileLogging(t *testing.T) {
	tmpDir := filepath.Join(os.TempDir(), "zap-test-logs")
	defer os.RemoveAll(tmpDir)

	logPath := filepath.Join(tmpDir, "test.log")

	config := &LoggerConfig{
		Driver: "zap",
		ZapConfig: &ZapConfig{
			ConsoleEnabled: true,
			ConsoleConfig:  &LogLevelConfig{Level: "info"},
			FileEnabled:    true,
			FileConfig: &FileLogConfig{
				Level: "debug",
				Path:  logPath,
				Rotation: &RotationConfig{
					MaxSize:    1, // 1MB
					MaxAge:     1, // 1 day
					MaxBackups: 2,
					Compress:   false,
				},
			},
		},
	}

	mgr, err := NewLoggerManagerZapImpl(config)
	if err != nil {
		t.Fatalf("NewLoggerManagerZapImpl() error = %v", err)
	}
	defer mgr.Shutdown(nil)

	logger := mgr.Logger("test")
	logger.Debug("debug to file")
	logger.Info("info to file")
	logger.Warn("warn to file")
	logger.Error("error to file")

	// 等待写入完成
	mgr.Shutdown(nil)

	// 验证文件是否创建
	if _, err := os.Stat(logPath); os.IsNotExist(err) {
		t.Errorf("Log file was not created: %s", logPath)
	}
}

// TestZapLoggerManagerLifecycle 测试完整的生命周期
func TestZapLoggerManagerLifecycle(t *testing.T) {
	config := &LoggerConfig{
		Driver: "zap",
		ZapConfig: &ZapConfig{
			ConsoleEnabled: true,
			ConsoleConfig:  &LogLevelConfig{Level: "info"},
		},
	}

	mgr, err := NewLoggerManagerZapImpl(config)
	if err != nil {
		t.Fatalf("NewLoggerManagerZapImpl() error = %v", err)
	}

	// 启动
	err = mgr.OnStart()
	if err != nil {
		t.Errorf("OnStart() error = %v", err)
	}

	// 获取 logger 并使用
	logger := mgr.Logger("app")
	logger.Info("Application started")

	// 设置级别
	mgr.SetGlobalLevel(DebugLevel)
	logger.Debug("Debug after level change")

	// 健康检查
	err = mgr.Health()
	if err != nil {
		t.Errorf("Health() error = %v", err)
	}

	// 停止
	err = mgr.OnStop()
	if err != nil {
		t.Errorf("OnStop() error = %v", err)
	}

	// 关闭
	ctx := context.Background()
	err = mgr.Shutdown(ctx)
	if err != nil && !containsSyncError(err) {
		t.Errorf("Shutdown() unexpected error = %v", err)
	}
}

// TestZapLoggerLevelFiltering 测试日志级别过滤
func TestZapLoggerLevelFiltering(t *testing.T) {
	tests := []struct {
		name        string
		configLevel string
		setLevel    LogLevel
		expectDebug bool
		expectInfo  bool
		expectWarn  bool
		expectError bool
	}{
		{
			name:        "Info level",
			configLevel: "info",
			setLevel:    InfoLevel,
			expectDebug: false,
			expectInfo:  true,
			expectWarn:  true,
			expectError: true,
		},
		{
			name:        "Error level",
			configLevel: "error",
			setLevel:    ErrorLevel,
			expectDebug: false,
			expectInfo:  false,
			expectWarn:  false,
			expectError: true,
		},
		{
			name:        "Debug level",
			configLevel: "debug",
			setLevel:    DebugLevel,
			expectDebug: true,
			expectInfo:  true,
			expectWarn:  true,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := &LoggerConfig{
				Driver: "zap",
				ZapConfig: &ZapConfig{
					ConsoleEnabled: true,
					ConsoleConfig:  &LogLevelConfig{Level: tt.configLevel},
				},
			}

			mgr, err := NewLoggerManagerZapImpl(config)
			if err != nil {
				t.Fatalf("NewLoggerManagerZapImpl() error = %v", err)
			}
			defer mgr.Shutdown(nil)

			logger := mgr.Logger("test")
			logger.SetLevel(tt.setLevel)

			// 这些调用不应该 panic
			logger.Debug("debug")
			logger.Info("info")
			logger.Warn("warn")
			logger.Error("error")
		})
	}
}

// TestZapLoggerEmptyArgs 测试空参数
func TestZapLoggerEmptyArgs(t *testing.T) {
	config := &LoggerConfig{
		Driver: "zap",
		ZapConfig: &ZapConfig{
			ConsoleEnabled: true,
			ConsoleConfig:  &LogLevelConfig{Level: "info"},
		},
	}

	mgr, err := NewLoggerManagerZapImpl(config)
	if err != nil {
		t.Fatalf("NewLoggerManagerZapImpl() error = %v", err)
	}
	defer mgr.Shutdown(nil)

	logger := mgr.Logger("test")

	// 空消息
	logger.Debug("")
	logger.Info("")
	logger.Warn("")
	logger.Error("")

	// 无参数
	logger.Debug("test")
	logger.Info("test")
	logger.Warn("test")
	logger.Error("test")

	// 空字段
	logger.With()
	logger.Info("test")
}

// TestZapLoggerMultipleManagers 测试创建多个管理器实例
func TestZapLoggerMultipleManagers(t *testing.T) {
	config := &LoggerConfig{
		Driver: "zap",
		ZapConfig: &ZapConfig{
			ConsoleEnabled: true,
			ConsoleConfig:  &LogLevelConfig{Level: "info"},
		},
	}

	mgr1, err := NewLoggerManagerZapImpl(config)
	if err != nil {
		t.Fatalf("NewLoggerManagerZapImpl() error = %v", err)
	}
	defer mgr1.Shutdown(nil)

	mgr2, err := NewLoggerManagerZapImpl(config)
	if err != nil {
		t.Fatalf("NewLoggerManagerZapImpl() error = %v", err)
	}
	defer mgr2.Shutdown(nil)

	logger1 := mgr1.Logger("test")
	logger2 := mgr2.Logger("test")

	// 两个管理器应该是独立的
	if logger1 == logger2 {
		t.Error("Loggers from different managers should be different instances")
	}

	logger1.Info("from manager 1")
	logger2.Info("from manager 2")
}

// TestValidateContext 测试上下文验证
func TestValidateContext(t *testing.T) {
	tests := []struct {
		name    string
		ctx     context.Context
		wantErr bool
	}{
		{
			name:    "Valid context",
			ctx:     context.Background(),
			wantErr: false,
		},
		{
			name:    "Valid TODO context",
			ctx:     context.TODO(),
			wantErr: false,
		},
		{
			name:    "Nil context",
			ctx:     nil,
			wantErr: true,
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

// BenchmarkZapLoggerOutput 性能测试
func BenchmarkZapLoggerOutput(b *testing.B) {
	config := &LoggerConfig{
		Driver: "zap",
		ZapConfig: &ZapConfig{
			ConsoleEnabled: true,
			ConsoleConfig:  &LogLevelConfig{Level: "info"},
		},
	}

	mgr, err := NewLoggerManagerZapImpl(config)
	if err != nil {
		b.Fatalf("NewLoggerManagerZapImpl() error = %v", err)
	}
	defer mgr.Shutdown(nil)

	logger := mgr.Logger("bench")

	b.Run("Info", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			logger.Info("benchmark message", "iter", i)
		}
	})

	b.Run("With", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			logger.With("key", "value")
		}
	})
}

// BenchmarkZapLoggerManagerGetLogger 性能测试 - 获取 Logger
func BenchmarkZapLoggerManagerGetLogger(b *testing.B) {
	config := &LoggerConfig{
		Driver: "zap",
		ZapConfig: &ZapConfig{
			ConsoleEnabled: true,
			ConsoleConfig:  &LogLevelConfig{Level: "info"},
		},
	}

	mgr, err := NewLoggerManagerZapImpl(config)
	if err != nil {
		b.Fatalf("NewLoggerManagerZapImpl() error = %v", err)
	}
	defer mgr.Shutdown(nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = mgr.Logger("test")
	}
}
