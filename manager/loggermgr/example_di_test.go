package loggermgr_test

import (
	"context"
	"testing"

	"com.litelake.litecore/manager/loggermgr"
	"com.litelake.litecore/manager/loggermgr/internal/config"
)

// Example_basicUsage 展示 LoggerManager 的基本用法（使用 DI 模式）
func Example_basicUsage() {
	// 1. 创建日志管理器
	loggerMgr := loggermgr.NewManager("default")

	// 2. 注入配置（在实际应用中由容器自动注入）
	// 这里使用 MockConfigProvider 模拟
	configProvider := &MockConfigProvider{
		configs: map[string]any{
			"logger.default": map[string]any{
				"console_enabled": true,
				"console_config": map[string]any{
					"level": "info",
				},
				"file_enabled": false,
			},
		},
	}
	loggerMgr.Config = configProvider

	// 3. 启动管理器
	if err := loggerMgr.OnStart(); err != nil {
		panic(err)
	}

	// 4. 获取 Logger 并使用
	logger := loggerMgr.Logger("example")
	logger.Info("Application started", "version", "1.0.0")

	// 5. 关闭管理器
	ctx := context.Background()
	loggerMgr.Shutdown(ctx)
}

// Example_withMultipleLoggers 展示如何使用多个命名的 Logger
func Example_withMultipleLoggers() {
	loggerMgr := loggermgr.NewManager("app")

	// 注入配置并启动
	configProvider := &MockConfigProvider{
		configs: map[string]any{
			"logger.app": map[string]any{
				"console_enabled": true,
				"console_config": map[string]any{
					"level": "debug",
				},
			},
		},
	}
	loggerMgr.Config = configProvider

	if err := loggerMgr.OnStart(); err != nil {
		panic(err)
	}
	defer loggerMgr.Shutdown(context.Background())

	// 为不同的模块创建不同的 Logger
	serviceLogger := loggerMgr.Logger("service")
	repositoryLogger := loggerMgr.Logger("repository")
	apiLogger := loggerMgr.Logger("api")

	// 各个模块使用自己的 Logger
	serviceLogger.Info("Service initialized")
	repositoryLogger.Debug("Query executed", "query", "SELECT * FROM users")
	apiLogger.Warn("Rate limit approaching", "count", 95)
}

// Example_withFields 展示如何使用 With 方法添加字段
func Example_withFields() {
	loggerMgr := loggermgr.NewManager("app")
	configProvider := &MockConfigProvider{
		configs: map[string]any{
			"logger.app": map[string]any{
				"console_enabled": true,
				"console_config":  map[string]any{"level": "info"},
			},
		},
	}
	loggerMgr.Config = configProvider

	if err := loggerMgr.OnStart(); err != nil {
		panic(err)
	}
	defer loggerMgr.Shutdown(context.Background())

	// 创建带有基础字段的 Logger
	baseLogger := loggerMgr.Logger("http")
	requestLogger := baseLogger.With("method", "GET", "endpoint", "/api/users")

	// 使用带有预设字段的 Logger
	requestLogger.Info("Processing request", "user_id", 12345)

	// 链式调用创建更多字段
	detailedLogger := requestLogger.With("duration_ms", 150)
	detailedLogger.Info("Request completed")
}

// Example_setLevel 展示如何动态设置日志级别
func Example_setLevel() {
	loggerMgr := loggermgr.NewManager("app")
	configProvider := &MockConfigProvider{
		configs: map[string]any{
			"logger.app": map[string]any{
				"console_enabled": true,
				"console_config":  map[string]any{"level": "info"},
			},
		},
	}
	loggerMgr.Config = configProvider

	if err := loggerMgr.OnStart(); err != nil {
		panic(err)
	}
	defer loggerMgr.Shutdown(context.Background())

	logger := loggerMgr.Logger("debug_test")

	// 设置为 Debug 级别
	logger.SetLevel(loggermgr.DebugLevel)
	logger.Debug("This debug message will be visible")

	// 设置为 Error 级别
	logger.SetLevel(loggermgr.ErrorLevel)
	logger.Debug("This debug message will NOT be visible")
	logger.Error("This error message will be visible")
}

// Example_configuration 展示完整的配置示例
func Example_configuration() {
	// 完整的配置示例
	cfg := map[string]any{
		// 控制台日志配置
		"console_enabled": true,
		"console_config": map[string]any{
			"level": "info", // debug, info, warn, error, fatal
		},

		// 文件日志配置
		"file_enabled": true,
		"file_config": map[string]any{
			"level": "debug",
			"path":  "/var/log/app/app.log",
			"rotation": map[string]any{
				"max_size":    100,  // MB
				"max_age":     30,   // days
				"max_backups": 10,   // number of backups
				"compress":    true, // compress old files
			},
		},

		// 观测日志配置（集成 OTEL）
		"telemetry_enabled": false,
		"telemetry_config": map[string]any{
			"level": "info",
		},
	}

	loggerMgr := loggermgr.NewManager("app")
	configProvider := &MockConfigProvider{
		configs: map[string]any{
			"logger.app": cfg,
		},
	}
	loggerMgr.Config = configProvider

	if err := loggerMgr.OnStart(); err != nil {
		panic(err)
	}
	defer loggerMgr.Shutdown(context.Background())

	logger := loggerMgr.Logger("configured")
	logger.Info("Logger initialized with full configuration")
}

// Example_defaultConfig 展示默认配置的行为
func Example_defaultConfig() {
	// 不提供配置，使用默认配置
	loggerMgr := loggermgr.NewManager("default")

	// 使用空的配置提供者
	configProvider := &MockConfigProvider{
		configs: map[string]any{},
	}
	loggerMgr.Config = configProvider

	if err := loggerMgr.OnStart(); err != nil {
		panic(err)
	}
	defer loggerMgr.Shutdown(context.Background())

	// 默认配置：控制台输出，info 级别
	logger := loggerMgr.Logger("default")
	logger.Info("Using default configuration")
}

// Example_multipleManagers 展示如何使用多个独立的日志管理器
func Example_multipleManagers() {
	// 创建两个独立的管理器
	appLoggerMgr := loggermgr.NewManager("app")
	auditLoggerMgr := loggermgr.NewManager("audit")

	// 为应用日志配置
	appConfig := &MockConfigProvider{
		configs: map[string]any{
			"logger.app": map[string]any{
				"console_enabled": true,
				"console_config":  map[string]any{"level": "debug"},
				"file_enabled": true,
				"file_config": map[string]any{
					"level": "info",
					"path":  "/var/log/app/app.log",
				},
			},
		},
	}
	appLoggerMgr.Config = appConfig

	// 为审计日志配置（只记录到文件）
	auditConfig := &MockConfigProvider{
		configs: map[string]any{
			"logger.audit": map[string]any{
				"console_enabled": false,
				"file_enabled":    true,
				"file_config": map[string]any{
					"level": "info",
					"path":  "/var/log/audit/audit.log",
				},
			},
		},
	}
	auditLoggerMgr.Config = auditConfig

	// 启动两个管理器
	if err := appLoggerMgr.OnStart(); err != nil {
		panic(err)
	}
	if err := auditLoggerMgr.OnStart(); err != nil {
		panic(err)
	}
	defer appLoggerMgr.Shutdown(context.Background())
	defer auditLoggerMgr.Shutdown(context.Background())

	// 使用应用日志
	appLogger := appLoggerMgr.Logger("service")
	appLogger.Debug("Debug information for troubleshooting")

	// 使用审计日志
	auditLogger := auditLoggerMgr.Logger("security")
	auditLogger.Info("User action", "action", "login", "user_id", 12345)
}

// MockConfigProvider 模拟配置提供者（用于示例）
type MockConfigProvider struct {
	configs map[string]any
}

func (m *MockConfigProvider) ConfigProviderName() string {
	return "mock"
}

func (m *MockConfigProvider) Get(key string) (any, error) {
	if cfg, ok := m.configs[key]; ok {
		return cfg, nil
	}
	return nil, nil // 配置不存在返回 nil
}

func (m *MockConfigProvider) Has(key string) bool {
	_, ok := m.configs[key]
	return ok
}

// TestExampleDIUsage 测试 DI 模式的使用
func TestExampleDIUsage(t *testing.T) {
	// 创建管理器
	mgr := loggermgr.NewManager("test")

	// 注入配置
	configProvider := &MockConfigProvider{
		configs: map[string]any{
			"logger.test": map[string]any{
				"console_enabled": true,
				"console_config":  map[string]any{"level": "debug"},
			},
		},
	}
	mgr.Config = configProvider

	// 启动
	if err := mgr.OnStart(); err != nil {
		t.Fatalf("OnStart() failed: %v", err)
	}

	// 使用
	logger := mgr.Logger("test")
	logger.Info("Test message", "key", "value")

	// 关闭（在测试环境中 sync 可能失败，这是正常的）
	if err := mgr.Shutdown(context.Background()); err != nil {
		// 测试环境中 sync 到 stdout/stderr 可能失败
		// 这是预期行为，不是真正的错误
		t.Logf("Shutdown() failed (expected in test): %v", err)
	}

	// 检查健康状态
	if err := mgr.Health(); err != nil {
		t.Errorf("Health() failed: %v", err)
	}
}

// TestExampleMultipleLoggers 测试多个 Logger
func TestExampleMultipleLoggers(t *testing.T) {
	mgr := loggermgr.NewManager("multi")

	configProvider := &MockConfigProvider{
		configs: map[string]any{
			"logger.multi": map[string]any{
				"console_enabled": true,
				"console_config":  map[string]any{"level": "info"},
			},
		},
	}
	mgr.Config = configProvider

	if err := mgr.OnStart(); err != nil {
		t.Fatalf("OnStart() failed: %v", err)
	}
	defer mgr.Shutdown(context.Background())

	// 创建多个 Logger
	logger1 := mgr.Logger("module1")
	logger2 := mgr.Logger("module2")

	// 使用
	logger1.Info("Module1 message")
	logger2.Info("Module2 message")
}

// TestExampleSetLevel 测试动态设置级别
func TestExampleSetLevel(t *testing.T) {
	mgr := loggermgr.NewManager("level")

	configProvider := &MockConfigProvider{
		configs: map[string]any{
			"logger.level": map[string]any{
				"console_enabled": true,
				"console_config":  map[string]any{"level": "info"},
			},
		},
	}
	mgr.Config = configProvider

	if err := mgr.OnStart(); err != nil {
		t.Fatalf("OnStart() failed: %v", err)
	}
	defer mgr.Shutdown(context.Background())

	logger := mgr.Logger("test")

	// 设置全局级别
	mgr.SetGlobalLevel(loggermgr.ErrorLevel)

	// Debug 和 Info 消息不会输出
	logger.Debug("Should not appear")
	logger.Info("Should not appear")

	// Error 消息会输出
	logger.Error("Should appear")
}

// TestExampleWithFields 测试 With 方法
func TestExampleWithFields(t *testing.T) {
	mgr := loggermgr.NewManager("with")

	configProvider := &MockConfigProvider{
		configs: map[string]any{
			"logger.with": map[string]any{
				"console_enabled": true,
				"console_config":  map[string]any{"level": "info"},
			},
		},
	}
	mgr.Config = configProvider

	if err := mgr.OnStart(); err != nil {
		t.Fatalf("OnStart() failed: %v", err)
	}
	defer mgr.Shutdown(context.Background())

	baseLogger := mgr.Logger("http")
	requestLogger := baseLogger.With("method", "GET", "path", "/api/users")

	// 使用带预设字段的 Logger
	requestLogger.Info("Processing request", "user_id", 123)

	// 链式调用
	detailedLogger := requestLogger.With("status", 200)
	detailedLogger.Info("Request completed")
}

// TestExampleDefaultConfig 测试默认配置
func TestExampleDefaultConfig(t *testing.T) {
	mgr := loggermgr.NewManager("default")

	// 空配置提供者
	configProvider := &MockConfigProvider{
		configs: map[string]any{},
	}
	mgr.Config = configProvider

	if err := mgr.OnStart(); err != nil {
		t.Fatalf("OnStart() failed: %v", err)
	}
	defer mgr.Shutdown(context.Background())

	logger := mgr.Logger("test")
	logger.Info("Using default config")
}

// TestExampleFileConfig 测试文件配置
func TestExampleFileConfig(t *testing.T) {
	// 这个测试验证配置解析功能
	cfg := &config.LoggerConfig{
		ConsoleEnabled: true,
		ConsoleConfig:  &config.LogLevelConfig{Level: "info"},
		FileEnabled:    false,
	}

	if err := cfg.Validate(); err != nil {
		t.Fatalf("Validate() failed: %v", err)
	}

	// 测试默认配置
	defaultCfg := config.DefaultLoggerConfig()
	if err := defaultCfg.Validate(); err != nil {
		t.Fatalf("DefaultConfig Validate() failed: %v", err)
	}

	// 验证默认值
	if !defaultCfg.ConsoleEnabled {
		t.Error("DefaultConfig should have console enabled")
	}
	if defaultCfg.ConsoleConfig.Level != "info" {
		t.Errorf("DefaultConfig console level should be 'info', got '%s'", defaultCfg.ConsoleConfig.Level)
	}
}
