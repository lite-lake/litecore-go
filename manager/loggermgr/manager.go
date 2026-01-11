package loggermgr

import (
	"context"
	"fmt"
	"sync"

	"com.litelake.litecore/common"
	"com.litelake.litecore/manager/loggermgr/internal/config"
	"com.litelake.litecore/manager/loggermgr/internal/drivers"
	"com.litelake.litecore/manager/loggermgr/internal/loglevel"
	"com.litelake.litecore/manager/telemetrymgr"
)

// Manager 日志管理器
// 实现依赖注入模式，支持从容器自动注入依赖
type Manager struct {
	// Config 配置提供者（必须依赖）
	Config common.BaseConfigProvider `inject:""`

	// TelemetryManager 观测管理器（可选依赖）
	// 如果可用，将集成 OTEL 的 TracerProvider
	TelemetryManager telemetrymgr.TelemetryManager `inject:"optional"`

	// name 管理器名称
	name string

	// driver 日志驱动实例
	driver drivers.Driver

	// level 当前日志级别
	level LogLevel

	// mu 保护并发访问的互斥锁
	mu sync.RWMutex

	// once 确保 OnStart 只执行一次
	once sync.Once
}

// NewManager 创建日志管理器
// name: 管理器名称，用于配置键前缀（如 "default" 对应配置键 "logger.default"）
//
// 该构造函数只做最小初始化，实际的配置加载和驱动创建在 OnStart 中完成
// 这样确保依赖注入完成后再初始化管理器
func NewManager(name string) *Manager {
	return &Manager{
		name:   name,
		driver: drivers.NewNoneDriver(),
		level:  InfoLevel,
	}
}

// ManagerName 返回管理器名称
// 实现 common.BaseManager 接口
func (m *Manager) ManagerName() string {
	return m.name
}

// OnStart 初始化管理器
// 实现 common.BaseManager 接口
//
// 该方法在依赖注入完成后由容器调用，执行以下步骤：
// 1. 从 ConfigProvider 加载配置
// 2. 获取 TelemetryManager 的 TracerProvider（如果可用）
// 3. 创建并启动日志驱动
// 4. 设置日志级别
//
// 如果配置加载失败或驱动创建失败，将降级到 NoneDriver（不输出日志）
func (m *Manager) OnStart() (err error) {
	m.once.Do(func() {
		// 1. 从 Config 加载配置
		cfg, loadErr := m.loadConfig()
		if loadErr != nil {
			err = fmt.Errorf("load config failed: %w", loadErr)
			return
		}

		// 2. 获取 TelemetryManager 的 TracerProvider（如果可用）
		var otelTracerProvider interface{}
		if m.TelemetryManager != nil {
			otelTracerProvider = m.TelemetryManager.TracerProvider()
		}

		// 3. 创建 zap 驱动
		driver, driverErr := drivers.NewZapDriver(cfg, otelTracerProvider)
		if driverErr != nil {
			// 驱动创建失败，使用 none 驱动降级
			m.driver = drivers.NewNoneDriver()
			err = fmt.Errorf("create zap driver failed, fallback to none driver: %w", driverErr)
			return
		}
		m.driver = driver

		// 4. 启动驱动
		if startErr := m.driver.Start(); startErr != nil {
			m.driver = drivers.NewNoneDriver()
			err = fmt.Errorf("start driver failed: %w", startErr)
			return
		}

		// 5. 设置日志级别（使用控制台日志级别作为默认级别）
		if cfg.ConsoleConfig != nil && cfg.ConsoleConfig.Level != "" {
			m.level = LogLevel(loglevel.ParseLogLevel(cfg.ConsoleConfig.Level))
		}
	})

	return err
}

// loadConfig 从 ConfigProvider 加载配置
//
// 配置键格式：logger.{manager_name}
// 例如：manager 名称为 "default"，配置键为 "logger.default"
//
// 如果 ConfigProvider 为 nil 或配置不存在，返回默认配置（仅控制台输出，info 级别）
func (m *Manager) loadConfig() (*config.LoggerConfig, error) {
	if m.Config == nil {
		return config.DefaultLoggerConfig(), nil
	}

	// 构建配置键：logger.{name}
	cfgKey := fmt.Sprintf("logger.%s", m.name)
	cfgData, err := m.Config.Get(cfgKey)
	if err != nil {
		// 配置不存在，使用默认配置
		return config.DefaultLoggerConfig(), nil
	}

	// 类型断言：将 any 转换为 map[string]any
	cfgMap, ok := cfgData.(map[string]any)
	if !ok {
		// 配置格式不正确，使用默认配置
		return config.DefaultLoggerConfig(), nil
	}

	// 解析配置
	cfg, parseErr := config.ParseLoggerConfigFromMap(cfgMap)
	if parseErr != nil {
		return nil, fmt.Errorf("parse config from map failed: %w", parseErr)
	}

	return cfg, nil
}

// OnStop 停止管理器
// 实现 common.BaseManager 接口
//
// 刷新所有待处理的日志并关闭日志驱动
func (m *Manager) OnStop() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	ctx := context.Background()
	if err := m.driver.Shutdown(ctx); err != nil {
		return fmt.Errorf("shutdown driver failed: %w", err)
	}

	return nil
}

// Health 检查管理器健康状态
// 实现 common.BaseManager 接口
func (m *Manager) Health() error {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return m.driver.Health()
}

// Logger 获取指定名称的 Logger 实例
// 实现 LoggerManager 接口
//
// name: Logger 名称，用于标识日志来源（如 "service", "repository"）
// 返回的 Logger 实例是线程安全的
func (m *Manager) Logger(name string) Logger {
	m.mu.RLock()
	defer m.mu.RUnlock()

	driverLogger := m.driver.GetLogger(name)
	return newDriverLoggerAdapter(driverLogger)
}

// newDriverLoggerAdapter 创建驱动 Logger 的适配器
// 将内部 drivers.Logger 适配为对外暴露的 Logger 接口
func newDriverLoggerAdapter(driverLogger drivers.Logger) Logger {
	// 如果是 ZapLogger，使用现有的 LoggerAdapter
	if zapLogger, ok := driverLogger.(*drivers.ZapLogger); ok {
		return NewLoggerAdapter(zapLogger)
	}
	// 否则使用通用的适配器
	return &genericLoggerAdapter{driver: driverLogger}
}

// genericLoggerAdapter 通用日志适配器
// 适配任何实现了 drivers.Logger 接口的日志记录器
type genericLoggerAdapter struct {
	driver drivers.Logger
}

// Debug 记录调试级别日志
func (a *genericLoggerAdapter) Debug(msg string, args ...any) {
	a.driver.Debug(msg, args...)
}

// Info 记录信息级别日志
func (a *genericLoggerAdapter) Info(msg string, args ...any) {
	a.driver.Info(msg, args...)
}

// Warn 记录警告级别日志
func (a *genericLoggerAdapter) Warn(msg string, args ...any) {
	a.driver.Warn(msg, args...)
}

// Error 记录错误级别日志
func (a *genericLoggerAdapter) Error(msg string, args ...any) {
	a.driver.Error(msg, args...)
}

// Fatal 记录致命错误级别日志
func (a *genericLoggerAdapter) Fatal(msg string, args ...any) {
	a.driver.Fatal(msg, args...)
}

// With 返回一个带有额外字段的新 Logger
func (a *genericLoggerAdapter) With(args ...any) Logger {
	return &genericLoggerAdapter{driver: a.driver.With(args...)}
}

// SetLevel 设置日志级别
func (a *genericLoggerAdapter) SetLevel(level LogLevel) {
	internalLevel := loglevel.LogLevel(level)
	a.driver.SetLevel(internalLevel)
}

// SetGlobalLevel 设置全局日志级别
// 实现 LoggerManager 接口
//
// level: 日志级别（DebugLevel, InfoLevel, WarnLevel, ErrorLevel, FatalLevel）
// 所有从该管理器获取的 Logger 实例都会受到影响
func (m *Manager) SetGlobalLevel(level LogLevel) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.level = level
	// 将 LogLevel 转换为内部 loglevel.LogLevel
	internalLevel := loglevel.LogLevel(level)
	m.driver.SetLevel(internalLevel)
}

// Shutdown 关闭日志管理器
// 实现 LoggerManager 接口
//
// 刷新所有待处理的日志并关闭日志驱动
// 该方法是 OnStop 的别名，提供更直观的关闭接口
func (m *Manager) Shutdown(ctx context.Context) error {
	return m.OnStop()
}

// 确保 Manager 实现 LoggerManager 和 common.BaseManager 接口
var _ LoggerManager = (*Manager)(nil)
var _ common.BaseManager = (*Manager)(nil)

