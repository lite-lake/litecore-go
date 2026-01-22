package loggermgr

import (
	"testing"

	"github.com/lite-lake/litecore-go/logger"
	"github.com/lite-lake/litecore-go/server/builtin/manager/telemetrymgr"
	"github.com/stretchr/testify/assert"
)

func TestNewDriverZapLoggerManager(t *testing.T) {
	t.Run("valid_console_config", func(t *testing.T) {
		cfg := &DriverZapConfig{
			ConsoleEnabled: true,
			ConsoleConfig:  &LogLevelConfig{Level: "info"},
		}

		mgr, err := NewDriverZapLoggerManager(cfg, nil)
		assert.NoError(t, err)
		assert.NotNil(t, mgr)
		assert.Equal(t, "LoggerZapManager", mgr.ManagerName())

		log := mgr.Ins()
		assert.NotNil(t, log)
		log.Info("test message", "key", "value")
	})

	t.Run("valid_file_config", func(t *testing.T) {
		cfg := &DriverZapConfig{
			FileEnabled: true,
			FileConfig: &FileLogConfig{
				Level: "debug",
				Path:  "/tmp/test_app.log",
				Rotation: &RotationConfig{
					MaxSize:    100,
					MaxAge:     30,
					MaxBackups: 10,
					Compress:   true,
				},
			},
		}

		mgr, err := NewDriverZapLoggerManager(cfg, nil)
		assert.NoError(t, err)
		assert.NotNil(t, mgr)

		log := mgr.Ins()
		assert.NotNil(t, log)
		log.Debug("debug message", "key", "value")
	})

	t.Run("both_outputs", func(t *testing.T) {
		cfg := &DriverZapConfig{
			ConsoleEnabled: true,
			ConsoleConfig:  &LogLevelConfig{Level: "info"},
			FileEnabled:    true,
			FileConfig: &FileLogConfig{
				Level: "warn",
				Path:  "/tmp/test_app.log",
				Rotation: &RotationConfig{
					MaxSize:    100,
					MaxAge:     30,
					MaxBackups: 10,
					Compress:   true,
				},
			},
		}

		mgr, err := NewDriverZapLoggerManager(cfg, nil)
		assert.NoError(t, err)
		assert.NotNil(t, mgr)

		log := mgr.Ins()
		assert.NotNil(t, log)
		log.Info("info message", "key", "value")
	})

	t.Run("nil_config", func(t *testing.T) {
		mgr, err := NewDriverZapLoggerManager(nil, nil)
		assert.Error(t, err)
		assert.Nil(t, mgr)
	})

	t.Run("no_output_enabled", func(t *testing.T) {
		cfg := &DriverZapConfig{
			ConsoleEnabled: false,
			FileEnabled:    false,
		}

		mgr, err := NewDriverZapLoggerManager(cfg, nil)
		assert.Error(t, err)
		assert.Nil(t, mgr)
	})

	t.Run("file_enabled_but_no_config", func(t *testing.T) {
		cfg := &DriverZapConfig{
			ConsoleEnabled: false,
			FileEnabled:    true,
			FileConfig:     nil,
		}

		mgr, err := NewDriverZapLoggerManager(cfg, nil)
		assert.Error(t, err)
		assert.Nil(t, mgr)
	})

	t.Run("telemetry_enabled_but_no_mgr", func(t *testing.T) {
		cfg := &DriverZapConfig{
			TelemetryEnabled: true,
			TelemetryConfig:  &LogLevelConfig{Level: "info"},
			ConsoleEnabled:   true,
			ConsoleConfig:    &LogLevelConfig{Level: "info"},
		}

		mgr, err := NewDriverZapLoggerManager(cfg, nil)
		assert.Error(t, err)
		assert.Nil(t, mgr)
		assert.Contains(t, err.Error(), "telemetry_manager is required")
	})

	t.Run("telemetry_with_mgr", func(t *testing.T) {
		cfg := &DriverZapConfig{
			TelemetryEnabled: true,
			TelemetryConfig:  &LogLevelConfig{Level: "info"},
			ConsoleEnabled:   true,
			ConsoleConfig:    &LogLevelConfig{Level: "info"},
		}

		telemetryMgr := telemetrymgr.NewTelemetryManagerNoneImpl()
		mgr, err := NewDriverZapLoggerManager(cfg, telemetryMgr)
		assert.NoError(t, err)
		assert.NotNil(t, mgr)

		log := mgr.Ins()
		log.Info("test message", "key", "value")
	})

	t.Run("telemetry_only", func(t *testing.T) {
		cfg := &DriverZapConfig{
			TelemetryEnabled: true,
			TelemetryConfig:  &LogLevelConfig{Level: "info"},
		}

		telemetryMgr := telemetrymgr.NewTelemetryManagerNoneImpl()
		mgr, err := NewDriverZapLoggerManager(cfg, telemetryMgr)
		assert.NoError(t, err)
		assert.NotNil(t, mgr)
	})

	t.Run("log_levels", func(t *testing.T) {
		cfg := &DriverZapConfig{
			ConsoleEnabled: true,
			ConsoleConfig:  &LogLevelConfig{Level: "debug"},
		}

		mgr, err := NewDriverZapLoggerManager(cfg, nil)
		assert.NoError(t, err)

		log := mgr.Ins()
		log.Debug("debug message")
		log.Info("info message")
		log.Warn("warn message")
		log.Error("error message")
	})

	t.Run("with_context", func(t *testing.T) {
		cfg := &DriverZapConfig{
			ConsoleEnabled: true,
			ConsoleConfig:  &LogLevelConfig{Level: "info"},
		}

		mgr, err := NewDriverZapLoggerManager(cfg, nil)
		assert.NoError(t, err)

		log := mgr.Ins()
		logWithCtx := log.With("service", "test-service", "version", "1.0.0")
		logWithCtx.Info("message with context")
	})

	t.Run("set_level", func(t *testing.T) {
		cfg := &DriverZapConfig{
			ConsoleEnabled: true,
			ConsoleConfig:  &LogLevelConfig{Level: "info"},
		}

		mgr, err := NewDriverZapLoggerManager(cfg, nil)
		assert.NoError(t, err)

		log := mgr.Ins()
		log.SetLevel(logger.DebugLevel)
		log.Debug("debug message after set level")
		log.SetLevel(logger.WarnLevel)
	})
}

func TestBuild(t *testing.T) {
	t.Run("zap_driver", func(t *testing.T) {
		cfg := &Config{
			Driver: "zap",
			ZapConfig: &DriverZapConfig{
				ConsoleEnabled: true,
				ConsoleConfig:  &LogLevelConfig{Level: "info"},
			},
		}

		mgr, err := Build(cfg, nil)
		assert.NoError(t, err)
		assert.NotNil(t, mgr)
		assert.Equal(t, "LoggerZapManager", mgr.ManagerName())
	})

	t.Run("default_driver", func(t *testing.T) {
		cfg := &Config{
			Driver: "default",
		}

		mgr, err := Build(cfg, nil)
		assert.NoError(t, err)
		assert.NotNil(t, mgr)
		assert.Equal(t, "LoggerDefaultManager", mgr.ManagerName())
	})

	t.Run("none_driver", func(t *testing.T) {
		cfg := &Config{
			Driver: "none",
		}

		mgr, err := Build(cfg, nil)
		assert.NoError(t, err)
		assert.NotNil(t, mgr)
		assert.Equal(t, "LoggerNoneManager", mgr.ManagerName())
	})

	t.Run("unknown_driver", func(t *testing.T) {
		cfg := &Config{
			Driver: "unknown",
		}

		mgr, err := Build(cfg, nil)
		assert.Error(t, err)
		assert.Nil(t, mgr)
	})

	t.Run("nil_config", func(t *testing.T) {
		mgr, err := Build(nil, nil)
		assert.Error(t, err)
		assert.Nil(t, mgr)
	})
}

func TestZapLoggerImpl_With(t *testing.T) {
	cfg := &DriverZapConfig{
		ConsoleEnabled: true,
		ConsoleConfig:  &LogLevelConfig{Level: "info"},
	}

	mgr, err := NewDriverZapLoggerManager(cfg, nil)
	assert.NoError(t, err)

	log := mgr.Ins()
	logWithCtx := log.With("service", "test-service", "version", "1.0.0")

	logWithCtx.Info("message with context", "key", "value")

	logWithCtx2 := logWithCtx.With("extra", "field")
	logWithCtx2.Info("message with extra context")
}

func TestManagerLifecycle(t *testing.T) {
	t.Run("zap_manager_lifecycle", func(t *testing.T) {
		cfg := &DriverZapConfig{
			ConsoleEnabled: true,
			ConsoleConfig:  &LogLevelConfig{Level: "info"},
		}

		mgr, err := NewDriverZapLoggerManager(cfg, nil)
		assert.NoError(t, err)

		assert.NoError(t, mgr.Health())
		assert.NoError(t, mgr.OnStart())
		assert.NoError(t, mgr.OnStop())
	})
}

func TestZapLoggerImpl_LevelFiltering(t *testing.T) {
	t.Run("级别过滤", func(t *testing.T) {
		cfg := &DriverZapConfig{
			ConsoleEnabled: true,
			ConsoleConfig:  &LogLevelConfig{Level: "warn"},
		}

		mgr, err := NewDriverZapLoggerManager(cfg, nil)
		assert.NoError(t, err)

		log := mgr.Ins()

		assert.NotPanics(t, func() {
			log.Debug("debug message")
			log.Info("info message")
			log.Warn("warn message")
			log.Error("error message")
		})
	})
}

func TestZapLoggerImpl_ConcurrentLogging(t *testing.T) {
	t.Run("并发日志", func(t *testing.T) {
		cfg := &DriverZapConfig{
			ConsoleEnabled: true,
			ConsoleConfig:  &LogLevelConfig{Level: "debug"},
		}

		mgr, err := NewDriverZapLoggerManager(cfg, nil)
		assert.NoError(t, err)

		log := mgr.Ins()

		done := make(chan bool)
		for i := 0; i < 100; i++ {
			go func(id int) {
				log.Info("concurrent message", "id", id)
				done <- true
			}(i)
		}

		for i := 0; i < 100; i++ {
			<-done
		}
	})
}

func TestZapLoggerImpl_WithChained(t *testing.T) {
	t.Run("链式 With 调用", func(t *testing.T) {
		cfg := &DriverZapConfig{
			ConsoleEnabled: true,
			ConsoleConfig:  &LogLevelConfig{Level: "info"},
		}

		mgr, err := NewDriverZapLoggerManager(cfg, nil)
		assert.NoError(t, err)

		log := mgr.Ins()

		log1 := log.With("service", "test")
		log2 := log1.With("version", "1.0")
		log3 := log2.With("env", "dev")

		assert.NotPanics(t, func() {
			log3.Info("message with chained context")
		})
	})
}

func TestBuildFileCore(t *testing.T) {
	t.Run("创建文件核心", func(t *testing.T) {
		cfg := &FileLogConfig{
			Level: "info",
			Path:  "/tmp/test_logger.log",
			Rotation: &RotationConfig{
				MaxSize:    10,
				MaxAge:     7,
				MaxBackups: 3,
				Compress:   false,
			},
		}

		core, err := buildFileCore(cfg)
		assert.NoError(t, err)
		assert.NotNil(t, core)
	})

	t.Run("空配置", func(t *testing.T) {
		core, err := buildFileCore(nil)
		assert.Error(t, err)
		assert.Nil(t, core)
	})

	t.Run("默认路径", func(t *testing.T) {
		cfg := &FileLogConfig{
			Level: "info",
		}

		core, err := buildFileCore(cfg)
		assert.NoError(t, err)
		assert.NotNil(t, core)
	})
}
