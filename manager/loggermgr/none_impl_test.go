package loggermgr

import (
	"context"
	"testing"
)

// TestNewLoggerManagerNoneImpl 测试创建空日志管理器
func TestNewLoggerManagerNoneImpl(t *testing.T) {
	mgr := NewLoggerManagerNoneImpl()
	if mgr == nil {
		t.Fatal("NewLoggerManagerNoneImpl() returned nil")
	}
}

// TestNoneLoggerManagerBaseInterface 测试空日志管理器基础接口
func TestNoneLoggerManagerBaseInterface(t *testing.T) {
	mgr := NewLoggerManagerNoneImpl()

	t.Run("ManagerName", func(t *testing.T) {
		name := mgr.ManagerName()
		if name != "none-logger" {
			t.Errorf("ManagerName() = %v, want 'none-logger'", name)
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

// TestNoneLoggerManagerLogger 测试获取 Logger
func TestNoneLoggerManagerLogger(t *testing.T) {
	mgr := NewLoggerManagerNoneImpl()

	logger := mgr.Logger("test")
	if logger == nil {
		t.Fatal("Logger() returned nil")
	}

	// 确保返回的是 noneLoggerImpl
	if _, ok := logger.(*noneLoggerImpl); !ok {
		t.Errorf("Logger() returned unexpected type: %T", logger)
	}
}

// TestNoneLoggerManagerSetGlobalLevel 测试设置全局日志级别
func TestNoneLoggerManagerSetGlobalLevel(t *testing.T) {
	mgr := NewLoggerManagerNoneImpl()

	// 设置不同级别不应出错
	levels := []LogLevel{DebugLevel, InfoLevel, WarnLevel, ErrorLevel, FatalLevel}
	for _, level := range levels {
		t.Run(level.String(), func(t *testing.T) {
			mgr.SetGlobalLevel(level)
			// 空实现不应有任何副作用
		})
	}
}

// TestNoneLoggerManagerShutdown 测试关闭
func TestNoneLoggerManagerShutdown(t *testing.T) {
	mgr := NewLoggerManagerNoneImpl()

	ctx := context.Background()
	err := mgr.Shutdown(ctx)
	if err != nil {
		t.Errorf("Shutdown() error = %v, want nil", err)
	}
}

// TestNoneLoggerManagerShutdownNilContext 测试 nil 上下文关闭
func TestNoneLoggerManagerShutdownNilContext(t *testing.T) {
	mgr := NewLoggerManagerNoneImpl()

	err := mgr.Shutdown(nil)
	if err != nil {
		t.Errorf("Shutdown(nil) error = %v, want nil", err)
	}
}

// TestNoneLoggerManagerMultipleLoggers 测试获取多个 Logger
func TestNoneLoggerManagerMultipleLoggers(t *testing.T) {
	mgr := NewLoggerManagerNoneImpl()

	names := []string{"test1", "test2", "test3", ""}
	for _, name := range names {
		t.Run(name, func(t *testing.T) {
			logger := mgr.Logger(name)
			if logger == nil {
				t.Errorf("Logger(%q) returned nil", name)
			}
			// 所有 Logger 应该都是 noneLoggerImpl
			if _, ok := logger.(*noneLoggerImpl); !ok {
				t.Errorf("Logger(%q) returned unexpected type: %T", name, logger)
			}
		})
	}
}

// TestNoneLoggerOutput 测试空日志输出器的各种日志方法
func TestNoneLoggerOutput(t *testing.T) {
	logger := newNoneLoggerImpl()

	tests := []struct {
		name string
		fn   func(msg string, args ...any)
	}{
		{"Debug", logger.Debug},
		{"Info", logger.Info},
		{"Warn", logger.Warn},
		{"Error", logger.Error},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 空实现不应 panic 或出错
			tt.fn("test message", "key", "value")
			tt.fn("test message") // 无参数
			tt.fn("")             // 空消息
		})
	}
}

// TestNoneLoggerFatal 测试 Fatal 方法
func TestNoneLoggerFatal(t *testing.T) {
	logger := newNoneLoggerImpl()

	// 注意：由于 Fatal 在空实现中不会退出程序，我们可以安全地调用它
	// 如果使用 os.Exit(1)，测试会被中断
	logger.Fatal("fatal message", "key", "value")
	logger.Fatal("fatal message")
	logger.Fatal("")
}

// TestNoneLoggerWith 测试 With 方法
func TestNoneLoggerWith(t *testing.T) {
	logger := newNoneLoggerImpl()

	// With 应该返回 Logger 实例（可能是自身或新实例）
	newLogger := logger.With("key1", "value1", "key2", "value2")
	if newLogger == nil {
		t.Fatal("With() returned nil")
	}

	// 验证返回的 Logger 可以正常使用
	newLogger.Debug("test")
	newLogger.Info("test")
	newLogger.Warn("test")
	newLogger.Error("test")

	// 多次 With
	logger2 := logger.With("a", "b").With("c", "d")
	if logger2 == nil {
		t.Fatal("Multiple With() calls returned nil")
	}
}

// TestNoneLoggerSetLevel 测试 SetLevel 方法
func TestNoneLoggerSetLevel(t *testing.T) {
	logger := newNoneLoggerImpl()

	levels := []LogLevel{DebugLevel, InfoLevel, WarnLevel, ErrorLevel, FatalLevel}
	for _, level := range levels {
		t.Run(level.String(), func(t *testing.T) {
			logger.SetLevel(level)
			// 空实现不应有任何副作用
		})
	}
}

// TestNoneLoggerNoPanic 测试确保空实现不会 panic
func TestNoneLoggerNoPanic(t *testing.T) {
	logger := newNoneLoggerImpl()
	mgr := NewLoggerManagerNoneImpl()

	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Unexpected panic: %v", r)
		}
	}()

	// 测试所有方法
	mgr.ManagerName()
	mgr.Health()
	mgr.OnStart()
	mgr.OnStop()
	mgr.Logger("test")
	mgr.SetGlobalLevel(InfoLevel)
	mgr.Shutdown(nil)

	logger.Debug("test")
	logger.Info("test")
	logger.Warn("test")
	logger.Error("test")
	logger.Fatal("test")
	logger.With("key", "value")
	logger.SetLevel(DebugLevel)
}

// TestNoneLoggerManagerLifecycle 测试完整的生命周期
func TestNoneLoggerManagerLifecycle(t *testing.T) {
	mgr := NewLoggerManagerNoneImpl()

	// 启动
	err := mgr.OnStart()
	if err != nil {
		t.Errorf("OnStart() error = %v", err)
	}

	// 获取 logger 并使用
	logger := mgr.Logger("app")
	logger.Info("Application started")
	logger.Debug("Debug info")
	logger.Warn("Warning message")
	logger.Error("Error message")

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
	if err != nil {
		t.Errorf("Shutdown() error = %v", err)
	}
}

// TestNoneLoggerNilSafe 测试 nil 安全性
func TestNoneLoggerNilSafe(t *testing.T) {
	logger := newNoneLoggerImpl()

	// 空参数测试
	logger.Debug("")
	logger.Info("")
	logger.Warn("")
	logger.Error("")
	logger.Fatal("")

	// nil 参数测试（变参）
	logger.Debug("test", nil)
	logger.Info("test", nil)

	// 奇数个参数
	logger.With("key1")                   // 只有一个参数
	logger.With("key1", "value1", "key2") // 三个参数
}

// TestNoneLoggerWithEmptyArgs 测试 With 方法空参数
func TestNoneLoggerWithEmptyArgs(t *testing.T) {
	logger := newNoneLoggerImpl()

	// 无参数
	newLogger := logger.With()
	if newLogger == nil {
		t.Error("With() with no args returned nil")
	}

	// 空字符串参数
	newLogger = logger.With("", "")
	if newLogger == nil {
		t.Error("With() with empty string args returned nil")
	}
}

// BenchmarkNoneLoggerOutput 性能测试
func BenchmarkNoneLoggerOutput(b *testing.B) {
	logger := newNoneLoggerImpl()

	b.Run("Debug", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			logger.Debug("benchmark message", "iter", i)
		}
	})

	b.Run("Info", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			logger.Info("benchmark message", "iter", i)
		}
	})

	b.Run("Warn", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			logger.Warn("benchmark message", "iter", i)
		}
	})

	b.Run("Error", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			logger.Error("benchmark message", "iter", i)
		}
	})

	b.Run("With", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			logger.With("key", "value")
		}
	})
}

// BenchmarkNoneLoggerManagerGetLogger 性能测试 - 获取 Logger
func BenchmarkNoneLoggerManagerGetLogger(b *testing.B) {
	mgr := NewLoggerManagerNoneImpl()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = mgr.Logger("test")
	}
}
