package telemetrymgr

import (
	"context"
	"fmt"
	"sync"

	"com.litelake.litecore/common"
	"com.litelake.litecore/manager/telemetrymgr/internal/config"
	"com.litelake.litecore/manager/telemetrymgr/internal/drivers"
	"go.opentelemetry.io/otel/log"
	"go.opentelemetry.io/otel/metric"
	noopmetric "go.opentelemetry.io/otel/metric/noop"
	sdklog "go.opentelemetry.io/otel/sdk/log"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
)

// Manager 观测管理器
// 实现依赖注入模式，适配 Container 的 DI 机制
type Manager struct {
	// Config 依赖注入的配置提供者
	// 注入标签为空字符串，表示必须注入
	Config common.BaseConfigProvider `inject:""`

	// 内部状态
	name   string
	driver TelemetryDriver
	tracer trace.Tracer
	meter  metric.Meter
	logger log.Logger
	mu     sync.RWMutex
	once   sync.Once
}

// NewManager 创建观测管理器实例
// name: 管理器名称，用于配置键前缀（如 "default" → "telemetry.default"）
//
// 构造函数只做最小初始化，实际配置加载和驱动初始化在 OnStart 中完成
// 这符合依赖注入的延迟初始化原则
func NewManager(name string) *Manager {
	return &Manager{
		name:   name,
		driver: drivers.NewNoneManager(),
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
// 使用 sync.Once 确保只执行一次，即使多次调用也只会初始化一次
// 初始化流程：
// 1. 从 Config 加载配置（如果 Config 为 nil，使用默认配置）
// 2. 根据配置创建驱动（优先 OTEL，失败时降级到 None）
// 3. 初始化默认的 Tracer/Meter/Logger
func (m *Manager) OnStart() error {
	var initErr error

	m.once.Do(func() {
		// 1. 加载配置
		cfg, err := m.loadConfig()
		if err != nil {
			initErr = fmt.Errorf("load config failed: %w", err)
			return
		}

		// 2. 创建驱动（优先 OTEL）
		driver, err := m.createDriver(cfg)
		if err != nil {
			// 驱动创建失败，使用 none 驱动降级
			m.driver = drivers.NewNoneManager()
			initErr = fmt.Errorf("create driver failed, using none driver: %w", err)
			return
		}

		m.mu.Lock()
		m.driver = driver
		m.mu.Unlock()

		// 3. 初始化默认的观测组件
		m.initializeComponents()
	})

	return initErr
}

// loadConfig 从 ConfigProvider 加载配置
// 支持 nil Config，此时返回默认配置（禁用观测）
func (m *Manager) loadConfig() (*config.TelemetryConfig, error) {
	if m.Config == nil {
		return config.DefaultConfig(), nil
	}

	// 配置键格式：telemetry.{manager_name}
	// 例如：telemetry.default
	cfgKey := fmt.Sprintf("telemetry.%s", m.name)

	// 检查配置是否存在
	if !m.Config.Has(cfgKey) {
		return config.DefaultConfig(), nil
	}

	// 获取配置数据
	cfgData, err := m.Config.Get(cfgKey)
	if err != nil {
		return nil, fmt.Errorf("get config failed: %w", err)
	}

	// 将配置数据转换为 map[string]any
	cfgMap, ok := cfgData.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("invalid config type for %s: expected map[string]any, got %T", cfgKey, cfgData)
	}

	// 解析配置
	return m.parseConfigFromMap(cfgMap)
}

// parseConfigFromMap 从 map 解析配置
func (m *Manager) parseConfigFromMap(cfgData map[string]any) (*config.TelemetryConfig, error) {
	// 获取驱动类型
	driverType, ok := cfgData["driver"].(string)
	if !ok {
		return nil, fmt.Errorf("missing or invalid 'driver' field in config")
	}

	telemetryConfig := &config.TelemetryConfig{
		Driver: driverType,
	}

	// 如果是 OTEL 驱动，解析 OTEL 配置
	if driverType == "otel" {
		otelConfigData, hasOtelConfig := cfgData["otel_config"]
		if !hasOtelConfig {
			return nil, fmt.Errorf("missing 'otel_config' field when driver is 'otel'")
		}

		otelCfgMap, ok := otelConfigData.(map[string]any)
		if !ok {
			return nil, fmt.Errorf("invalid 'otel_config' type: expected map[string]any, got %T", otelConfigData)
		}

		otelConfig, err := config.ParseOtelConfigFromMap(otelCfgMap)
		if err != nil {
			return nil, fmt.Errorf("parse otel config failed: %w", err)
		}

		telemetryConfig.OtelConfig = otelConfig
	}

	// 验证配置
	if err := telemetryConfig.Validate(); err != nil {
		return nil, fmt.Errorf("validate config failed: %w", err)
	}

	return telemetryConfig, nil
}

// createDriver 根据配置创建驱动
// 优先使用 OTEL 驱动，失败时降级到 None 驱动
func (m *Manager) createDriver(cfg *config.TelemetryConfig) (TelemetryDriver, error) {
	switch cfg.Driver {
	case "otel":
		// 创建 OTEL 驱动
		driver, err := drivers.NewOtelManager(cfg)
		if err != nil {
			// OTEL 初始化失败，返回 none 驱动
			return nil, fmt.Errorf("create otel driver failed: %w", err)
		}
		return driver, nil

	case "none":
		// none 驱动
		return drivers.NewNoneManager(), nil

	default:
		return nil, fmt.Errorf("unsupported driver type: %s", cfg.Driver)
	}
}

// initializeComponents 初始化默认的观测组件
// 必须在持有锁的情况下调用
func (m *Manager) initializeComponents() {
	// 创建默认的 Tracer/Meter/Logger
	m.tracer = m.driver.Tracer(m.name)
	m.meter = m.driver.Meter(m.name)
	m.logger = m.driver.Logger(m.name)
}

// OnStop 停止管理器
// 实现 common.BaseManager 接口
func (m *Manager) OnStop() error {
	m.mu.RLock()
	defer m.mu.RUnlock()

	ctx := context.Background()
	return m.driver.Shutdown(ctx)
}

// Health 检查管理器健康状态
// 实现 common.BaseManager 接口
func (m *Manager) Health() error {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.driver == nil {
		return fmt.Errorf("driver is not initialized")
	}

	return m.driver.Health()
}

// ========== Tracing ==========

// Tracer 获取 Tracer 实例
// 实现 TelemetryManager 接口
// 线程安全：使用 RWMutex 保护
func (m *Manager) Tracer(name string) trace.Tracer {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.driver == nil {
		return trace.NewNoopTracerProvider().Tracer(name)
	}

	return m.driver.Tracer(name)
}

// TracerProvider 获取 TracerProvider
// 实现 TelemetryManager 接口
// 线程安全：使用 RWMutex 保护
func (m *Manager) TracerProvider() *sdktrace.TracerProvider {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.driver == nil {
		return sdktrace.NewTracerProvider()
	}

	return m.driver.TracerProvider()
}

// ========== Metrics ==========

// Meter 获取 Meter 实例
// 实现 TelemetryManager 接口
// 线程安全：使用 RWMutex 保护
func (m *Manager) Meter(name string) metric.Meter {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.driver == nil {
		return noopmetric.NewMeterProvider().Meter(name)
	}

	return m.driver.Meter(name)
}

// MeterProvider 获取 MeterProvider
// 实现 TelemetryManager 接口
// 线程安全：使用 RWMutex 保护
func (m *Manager) MeterProvider() *sdkmetric.MeterProvider {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.driver == nil {
		return sdkmetric.NewMeterProvider()
	}

	return m.driver.MeterProvider()
}

// ========== Logging ==========

// Logger 获取 Logger 实例
// 实现 TelemetryManager 接口
// 线程安全：使用 RWMutex 保护
func (m *Manager) Logger(name string) log.Logger {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.driver == nil {
		return sdklog.NewLoggerProvider().Logger(name)
	}

	return m.driver.Logger(name)
}

// LoggerProvider 获取 LoggerProvider
// 实现 TelemetryManager 接口
// 线程安全：使用 RWMutex 保护
func (m *Manager) LoggerProvider() *sdklog.LoggerProvider {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.driver == nil {
		return sdklog.NewLoggerProvider()
	}

	return m.driver.LoggerProvider()
}

// ========== Lifecycle ==========

// Shutdown 关闭观测管理器
// 实现 TelemetryManager 接口
// 刷新所有待处理的数据并关闭连接
func (m *Manager) Shutdown(ctx context.Context) error {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.driver == nil {
		return nil
	}

	return m.driver.Shutdown(ctx)
}

// ========== Internal Driver Interface ==========

// TelemetryDriver 驱动接口
// 封装底层驱动实现，使 Manager 不直接依赖具体的驱动类型
type TelemetryDriver interface {
	// Tracer 获取 Tracer 实例
	Tracer(name string) trace.Tracer

	// TracerProvider 获取 TracerProvider
	TracerProvider() *sdktrace.TracerProvider

	// Meter 获取 Meter 实例
	Meter(name string) metric.Meter

	// MeterProvider 获取 MeterProvider
	MeterProvider() *sdkmetric.MeterProvider

	// Logger 获取 Logger 实例
	Logger(name string) log.Logger

	// LoggerProvider 获取 LoggerProvider
	LoggerProvider() *sdklog.LoggerProvider

	// Shutdown 关闭驱动
	Shutdown(ctx context.Context) error

	// Health 健康检查
	Health() error
}

// ========== Compile-time Interface Validation ==========

// 确保 Manager 实现 TelemetryManager 接口
var _ TelemetryManager = (*Manager)(nil)

// 确保 Manager 实现 common.BaseManager 接口
var _ common.BaseManager = (*Manager)(nil)
