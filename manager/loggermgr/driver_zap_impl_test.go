package loggermgr

import (
	"github.com/lite-lake/litecore-go/manager/telemetrymgr"
	"testing"

	"github.com/lite-lake/litecore-go/logger"
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

func TestLoggerFormats(t *testing.T) {
	t.Run("Gin格式输出", func(t *testing.T) {
		cfg := &DriverZapConfig{
			ConsoleEnabled: true,
			ConsoleConfig: &LogLevelConfig{
				Level:      "debug",
				Format:     "gin",
				Color:      true,
				TimeFormat: "2006-01-02 15:04:05.000",
			},
		}

		mgr, err := NewDriverZapLoggerManager(cfg, nil)
		assert.NoError(t, err)
		assert.NotNil(t, mgr)

		log := mgr.Ins()
		assert.NotNil(t, log)

		assert.NotPanics(t, func() {
			log.Debug("debug message", "user_id", 1001, "action", "login")
			log.Info("info message", "request_id", "abc123", "path", "/api/users")
			log.Warn("warn message", "retry_count", 3, "max_retries", 5)
			log.Error("error message", "error_code", 500, "error_msg", "internal error")
		})

		logWithCtx := log.With("service", "user-service", "version", "1.0.0")
		assert.NotPanics(t, func() {
			logWithCtx.Info("message with context", "status", "running")
		})
	})

	t.Run("Json格式输出", func(t *testing.T) {
		cfg := &DriverZapConfig{
			ConsoleEnabled: true,
			ConsoleConfig: &LogLevelConfig{
				Level:      "info",
				Format:     "json",
				Color:      false,
				TimeFormat: "2006-01-02T15:04:05.000Z07:00",
			},
		}

		mgr, err := NewDriverZapLoggerManager(cfg, nil)
		assert.NoError(t, err)
		assert.NotNil(t, mgr)

		log := mgr.Ins()
		assert.NotNil(t, log)

		assert.NotPanics(t, func() {
			log.Info("info message", "key1", "value1", "key2", 123)
			log.Warn("warn message", "warning", "something")
			log.Error("error message", "err", "error details")
		})
	})

	t.Run("Default格式输出", func(t *testing.T) {
		cfg := &DriverZapConfig{
			ConsoleEnabled: true,
			ConsoleConfig: &LogLevelConfig{
				Level:      "info",
				Format:     "default",
				Color:      false,
				TimeFormat: "2006-01-02 15:04:05",
			},
		}

		mgr, err := NewDriverZapLoggerManager(cfg, nil)
		assert.NoError(t, err)
		assert.NotNil(t, mgr)

		log := mgr.Ins()
		assert.NotNil(t, log)

		assert.NotPanics(t, func() {
			log.Debug("debug message")
			log.Info("info message", "data", "test")
			log.Warn("warn message")
			log.Error("error message", "error", "test error")
		})
	})

	t.Run("颜色控制-开启颜色", func(t *testing.T) {
		cfg := &DriverZapConfig{
			ConsoleEnabled: true,
			ConsoleConfig: &LogLevelConfig{
				Level:  "info",
				Format: "gin",
				Color:  true,
			},
		}

		mgr, err := NewDriverZapLoggerManager(cfg, nil)
		assert.NoError(t, err)
		assert.NotNil(t, mgr)

		log := mgr.Ins()
		assert.NotPanics(t, func() {
			log.Info("info message", "key", "value")
			log.Warn("warn message", "key", "value")
			log.Error("error message", "key", "value")
		})
	})

	t.Run("颜色控制-关闭颜色", func(t *testing.T) {
		cfg := &DriverZapConfig{
			ConsoleEnabled: true,
			ConsoleConfig: &LogLevelConfig{
				Level:  "info",
				Format: "gin",
				Color:  false,
			},
		}

		mgr, err := NewDriverZapLoggerManager(cfg, nil)
		assert.NoError(t, err)
		assert.NotNil(t, mgr)

		log := mgr.Ins()
		assert.NotPanics(t, func() {
			log.Info("info message", "key", "value")
			log.Warn("warn message", "key", "value")
			log.Error("error message", "key", "value")
		})
	})

	t.Run("自定义时间格式", func(t *testing.T) {
		timeFormats := []string{
			"2006-01-02 15:04:05",
			"2006/01/02 15:04:05",
			"15:04:05.000",
			"2006-01-02T15:04:05Z07:00",
		}

		for _, tf := range timeFormats {
			t.Run(tf, func(t *testing.T) {
				cfg := &DriverZapConfig{
					ConsoleEnabled: true,
					ConsoleConfig: &LogLevelConfig{
						Level:      "info",
						Format:     "gin",
						Color:      false,
						TimeFormat: tf,
					},
				}

				mgr, err := NewDriverZapLoggerManager(cfg, nil)
				assert.NoError(t, err)
				assert.NotNil(t, mgr)

				log := mgr.Ins()
				assert.NotPanics(t, func() {
					log.Info("test time format", "format", tf)
				})
			})
		}
	})

	t.Run("文件日志与控制台日志并存", func(t *testing.T) {
		tempDir := "/tmp/test_logger_integration"
		logPath := tempDir + "/app.log"

		cfg := &DriverZapConfig{
			ConsoleEnabled: true,
			ConsoleConfig: &LogLevelConfig{
				Level:  "debug",
				Format: "gin",
				Color:  false,
			},
			FileEnabled: true,
			FileConfig: &FileLogConfig{
				Level: "debug",
				Path:  logPath,
				Rotation: &RotationConfig{
					MaxSize:    1,
					MaxAge:     1,
					MaxBackups: 1,
					Compress:   false,
				},
			},
		}

		mgr, err := NewDriverZapLoggerManager(cfg, nil)
		assert.NoError(t, err)
		assert.NotNil(t, mgr)

		log := mgr.Ins()
		assert.NotNil(t, log)

		assert.NotPanics(t, func() {
			log.Debug("debug message", "output", "both")
			log.Info("info message", "output", "both")
			log.Warn("warn message", "output", "both")
			log.Error("error message", "output", "both")
		})

		assert.NoError(t, mgr.OnStop())
	})

	t.Run("文件日志与控制台日志并存-不同级别", func(t *testing.T) {
		tempDir := "/tmp/test_logger_integration_levels"
		logPath := tempDir + "/app.log"

		cfg := &DriverZapConfig{
			ConsoleEnabled: true,
			ConsoleConfig: &LogLevelConfig{
				Level:  "info",
				Format: "gin",
				Color:  false,
			},
			FileEnabled: true,
			FileConfig: &FileLogConfig{
				Level: "warn",
				Path:  logPath,
				Rotation: &RotationConfig{
					MaxSize:    1,
					MaxAge:     1,
					MaxBackups: 1,
					Compress:   false,
				},
			},
		}

		mgr, err := NewDriverZapLoggerManager(cfg, nil)
		assert.NoError(t, err)
		assert.NotNil(t, mgr)

		log := mgr.Ins()
		assert.NotNil(t, log)

		assert.NotPanics(t, func() {
			log.Debug("debug message - file only")
			log.Info("info message - console only")
			log.Warn("warn message - both outputs")
			log.Error("error message - both outputs")
		})

		assert.NoError(t, mgr.OnStop())
	})

	t.Run("多格式组合测试", func(t *testing.T) {
		formats := []struct {
			format string
			name   string
		}{
			{"gin", "Gin格式"},
			{"json", "Json格式"},
			{"default", "Default格式"},
		}

		for _, f := range formats {
			t.Run(f.name, func(t *testing.T) {
				cfg := &DriverZapConfig{
					ConsoleEnabled: true,
					ConsoleConfig: &LogLevelConfig{
						Level:  "info",
						Format: f.format,
						Color:  false,
					},
				}

				mgr, err := NewDriverZapLoggerManager(cfg, nil)
				assert.NoError(t, err)
				assert.NotNil(t, mgr)

				log := mgr.Ins()
				assert.NotPanics(t, func() {
					log.Debug("debug")
					log.Info("info", "format", f.format)
					log.Warn("warn", "format", f.format)
					log.Error("error", "format", f.format)
				})

				logWithCtx := log.With("test_field", "test_value")
				assert.NotPanics(t, func() {
					logWithCtx.Info("message with context", "format", f.format)
				})
			})
		}
	})

	t.Run("Gin格式-多种字段类型", func(t *testing.T) {
		cfg := &DriverZapConfig{
			ConsoleEnabled: true,
			ConsoleConfig: &LogLevelConfig{
				Level:      "debug",
				Format:     "gin",
				Color:      false,
				TimeFormat: "2006-01-02 15:04:05.000",
			},
		}

		mgr, err := NewDriverZapLoggerManager(cfg, nil)
		assert.NoError(t, err)
		assert.NotNil(t, mgr)

		log := mgr.Ins()

		assert.NotPanics(t, func() {
			log.Info("string fields", "name", "Zhang San", "email", "test@example.com")
			log.Info("number fields", "age", 25, "score", 98.5, "count", 1000)
			log.Info("boolean fields", "enabled", true, "active", false)
			log.Info("mixed fields", "id", 123, "name", "Li Si", "active", true, "balance", 999.99)
		})
	})
}
