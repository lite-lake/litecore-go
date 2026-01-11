# Manager 完全重构方案

**文档编号**: TRD-20260111-manager-refactoring
**创建日期**: 2025-01-11
**文档版本**: 1.0
**项目**: litecore-go

---

## 一、重构目标

将 Manager 从 **Factory 模式** 重构为 **依赖注入（DI）模式**，完全适配 Container 的设计机制。

### 当前架构问题
```
Factory.Build(cfg, dep1, dep2) → Manager (手动传递依赖)
```

### 重构后架构
```
Container.Register(Manager) → Container.InjectAll() → 自动注入依赖
```

---

## 二、核心设计原则

### 2.1 依赖注入原则
- 所有依赖通过 `inject` 标签声明
- Container 自动解析和注入依赖
- 支持跨层依赖和同层依赖
- 支持可选依赖 (`inject:"optional"`)

### 2.2 延迟初始化原则
- 构造函数只做最小初始化
- 配置读取和依赖初始化在 `OnStart()` 中完成
- 保证注入顺序正确后再初始化

### 2.3 配置获取原则
- 从 `BaseConfigProvider` 获取配置
- 使用结构化配置而非 `map[string]any`
- 支持配置验证和默认值

---

## 三、架构改造方案

### 3.1 Manager 层次结构

```
┌─────────────────────────────────────────┐
│     TelemetryManager (无依赖)            │ ← 基础设施层，最先初始化
├─────────────────────────────────────────┤
│     LoggerManager (依赖 Telemetry)      │ ← 依赖 Telemetry
├─────────────────────────────────────────┤
│     DatabaseManager (无依赖)             │ ← 基础设施层
├─────────────────────────────────────────┤
│     CacheManager (依赖 Logger, Telemetry)│ ← 依赖上层 Manager
└─────────────────────────────────────────┘
```

### 3.2 依赖声明规范

```go
type XxxManager struct {
    // 基础依赖（必须）
    Config common.BaseConfigProvider `inject:""`

    // Manager 层依赖（可选）
    TelemetryManager telemetrymgr.TelemetryManager `inject:"optional"`
    LoggerManager    loggermgr.LoggerManager      `inject:"optional"`

    // 内部状态（不注入）
    name     string
    driver   AnyDriver
    mu       sync.RWMutex
    once     sync.Once
}
```

---

## 四、各 Manager 改造方案

### 4.1 TelemetryManager（无依赖，最先初始化）

**文件结构**：
```
manager/telemetrymgr/
├── manager.go          # Manager 实现（新增）
├── telemetry.go        # 接口保持不变
└── internal/
    ├── config/         # 配置解析
    └── drivers/        # OTEL 驱动实现
```

**manager.go（新增）**：
```go
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
    sdklog "go.opentelemetry.io/otel/sdk/log"
    sdkmetric "go.opentelemetry.io/otel/sdk/metric"
    sdktrace "go.opentelemetry.io/otel/sdk/trace"
    "go.opentelemetry.io/otel/trace"
)

// Manager 观测管理器
type Manager struct {
    // 依赖注入字段
    Config common.BaseConfigProvider `inject:""`

    // 内部状态
    name      string
    driver    drivers.Driver
    tracer    trace.Tracer
    meter     metric.Meter
    logger    log.Logger
    mu        sync.RWMutex
    once      sync.Once
}

// NewManager 创建观测管理器
func NewManager(name string) *Manager {
    return &Manager{
        name:   name,
        driver: drivers.NewNoneDriver(),
    }
}

// ManagerName 返回管理器名称
func (m *Manager) ManagerName() string {
    return m.name
}

// OnStart 初始化管理器（依赖注入完成后调用）
func (m *Manager) OnStart() error {
    return m.once.Do(func() error {
        // 1. 从 Config 获取配置
        cfg, err := m.loadConfig()
        if err != nil {
            return fmt.Errorf("load config failed: %w", err)
        }

        // 2. 创建驱动
        driver, err := drivers.NewOTELDriver(cfg)
        if err != nil {
            return fmt.Errorf("create driver failed: %w", err)
        }
        m.driver = driver

        // 3. 初始化观测组件
        if err := m.driver.Start(); err != nil {
            return fmt.Errorf("start driver failed: %w", err)
        }

        // 4. 创建默认的 Tracer/Meter/Logger
        m.tracer = m.driver.TracerProvider().Tracer(m.name)
        m.meter = m.driver.MeterProvider().Meter(m.name)
        m.logger = m.driver.LoggerProvider().Logger(m.name)

        return nil
    })
}

// loadConfig 从 ConfigProvider 加载配置
func (m *Manager) loadConfig() (*config.TelemetryConfig, error) {
    if m.Config == nil {
        // 返回默认配置（禁用观测）
        return config.DefaultConfig(), nil
    }

    // 获取配置键
    cfgKey := fmt.Sprintf("telemetry.%s", m.name)
    cfgData, ok := m.Config.Get(cfgKey)
    if !ok {
        return config.DefaultConfig(), nil
    }

    // 解析配置
    return config.ParseFromMap(cfgData)
}

// OnStop 停止管理器
func (m *Manager) OnStop() error {
    ctx := context.Background()
    return m.driver.Shutdown(ctx)
}

// Health 健康检查
func (m *Manager) Health() error {
    return m.driver.Health()
}

// ========== Tracing ==========

// Tracer 获取 Tracer 实例
func (m *Manager) Tracer(name string) trace.Tracer {
    m.mu.RLock()
    defer m.mu.RUnlock()

    if m.driver == nil {
        return trace.NewNoOpTracerProvider().Tracer(name)
    }
    return m.driver.TracerProvider().Tracer(name)
}

// TracerProvider 获取 TracerProvider
func (m *Manager) TracerProvider() *sdktrace.TracerProvider {
    m.mu.RLock()
    defer m.mu.RUnlock()
    return m.driver.TracerProvider()
}

// ========== Metrics ==========

// Meter 获取 Meter 实例
func (m *Manager) Meter(name string) metric.Meter {
    m.mu.RLock()
    defer m.mu.RUnlock()

    if m.driver == nil {
        return metric.NewNoopMeterProvider().Meter(name)
    }
    return m.driver.MeterProvider().Meter(name)
}

// MeterProvider 获取 MeterProvider
func (m *Manager) MeterProvider() *sdkmetric.MeterProvider {
    m.mu.RLock()
    defer m.mu.RUnlock()
    return m.driver.MeterProvider()
}

// ========== Logging ==========

// Logger 获取 Logger 实例
func (m *Manager) Logger(name string) log.Logger {
    m.mu.RLock()
    defer m.mu.RUnlock()

    if m.driver == nil {
        return log.NewNoopLoggerProvider().Logger(name)
    }
    return m.driver.LoggerProvider().Logger(name)
}

// LoggerProvider 获取 LoggerProvider
func (m *Manager) LoggerProvider() *sdklog.LoggerProvider {
    m.mu.RLock()
    defer m.mu.RUnlock()
    return m.driver.LoggerProvider()
}

// Shutdown 关闭观测管理器
func (m *Manager) Shutdown(ctx context.Context) error {
    return m.OnStop()
}

// 确保 Manager 实现 TelemetryManager 和 common.BaseManager
var _ TelemetryManager = (*Manager)(nil)
var _ common.BaseManager = (*Manager)(nil)
```

**删除文件**：
- `factory.go` - 移除 Factory 模式

---

### 4.2 LoggerManager（依赖 TelemetryManager）

**manager.go（新增）**：
```go
package loggermgr

import (
    "context"
    "fmt"
    "sync"

    "com.litelake.litecore/common"
    "com.litelake.litecore/manager/loggermgr/internal/config"
    "com.litelake.litecore/manager/loggermgr/internal/drivers"
    "com.litelake.litecore/manager/telemetrymgr"
    sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

// Manager 日志管理器
type Manager struct {
    // 依赖注入字段
    Config            common.BaseConfigProvider     `inject:""`
    TelemetryManager  telemetrymgr.TelemetryManager `inject:"optional"`

    // 内部状态
    name      string
    driver    drivers.Driver
    level     LogLevel
    mu        sync.RWMutex
    once      sync.Once
}

// NewManager 创建日志管理器
func NewManager(name string) *Manager {
    return &Manager{
        name:   name,
        driver: drivers.NewNoneDriver(),
        level:  InfoLevel,
    }
}

// ManagerName 返回管理器名称
func (m *Manager) ManagerName() string {
    return m.name
}

// OnStart 初始化管理器（依赖注入完成后调用）
func (m *Manager) OnStart() error {
    return m.once.Do(func() error {
        // 1. 从 Config 获取配置
        cfg, err := m.loadConfig()
        if err != nil {
            return fmt.Errorf("load config failed: %w", err)
        }

        // 2. 获取 TelemetryManager（如果可用）
        var otelTracerProvider *sdktrace.TracerProvider
        if m.TelemetryManager != nil {
            otelTracerProvider = m.TelemetryManager.TracerProvider()
        }

        // 3. 创建驱动
        driver, err := drivers.NewZapDriver(cfg, otelTracerProvider)
        if err != nil {
            return fmt.Errorf("create driver failed: %w", err)
        }
        m.driver = driver

        // 4. 启动驱动
        if err := m.driver.Start(); err != nil {
            return fmt.Errorf("start driver failed: %w", err)
        }

        // 5. 设置日志级别
        m.level = ParseLogLevel(cfg.Level)

        return nil
    })
}

// loadConfig 从 ConfigProvider 加载配置
func (m *Manager) loadConfig() (*config.LoggerConfig, error) {
    if m.Config == nil {
        return config.DefaultConfig(), nil
    }

    cfgKey := fmt.Sprintf("logger.%s", m.name)
    cfgData, ok := m.Config.Get(cfgKey)
    if !ok {
        return config.DefaultConfig(), nil
    }

    return config.ParseFromMap(cfgData)
}

// OnStop 停止管理器
func (m *Manager) OnStop() error {
    ctx := context.Background()
    return m.driver.Shutdown(ctx)
}

// Health 健康检查
func (m *Manager) Health() error {
    return m.driver.Health()
}

// Logger 获取指定名称的 Logger 实例
func (m *Manager) Logger(name string) Logger {
    m.mu.RLock()
    defer m.mu.RUnlock()

    driverLogger := m.driver.GetLogger(name)
    return &LoggerAdapter{driver: driverLogger}
}

// SetGlobalLevel 设置全局日志级别
func (m *Manager) SetGlobalLevel(level LogLevel) {
    m.mu.Lock()
    defer m.mu.Unlock()
    m.level = level
    m.driver.SetLevel(level)
}

// Shutdown 关闭日志管理器
func (m *Manager) Shutdown(ctx context.Context) error {
    return m.OnStop()
}

// LoggerAdapter Logger 适配器
type LoggerAdapter struct {
    driver drivers.Logger
}

// Debug 记录调试级别日志
func (l *LoggerAdapter) Debug(msg string, args ...any) {
    l.driver.Debug(msg, args...)
}

// Info 记录信息级别日志
func (l *LoggerAdapter) Info(msg string, args ...any) {
    l.driver.Info(msg, args...)
}

// Warn 记录警告级别日志
func (l *LoggerAdapter) Warn(msg string, args ...any) {
    l.driver.Warn(msg, args...)
}

// Error 记录错误级别日志
func (l *LoggerAdapter) Error(msg string, args ...any) {
    l.driver.Error(msg, args...)
}

// Fatal 记录致命错误级别日志
func (l *LoggerAdapter) Fatal(msg string, args ...any) {
    l.driver.Fatal(msg, args...)
}

// With 返回一个带有额外字段的新 Logger
func (l *LoggerAdapter) With(args ...any) Logger {
    return &LoggerAdapter{driver: l.driver.With(args...)}
}

// SetLevel 设置日志级别
func (l *LoggerAdapter) SetLevel(level LogLevel) {
    l.driver.SetLevel(level)
}

// 确保 Manager 实现 LoggerManager 和 common.BaseManager
var _ LoggerManager = (*Manager)(nil)
var _ common.BaseManager = (*Manager)(nil)
```

---

### 4.3 DatabaseManager（无依赖）

**manager.go（新增）**：
```go
package databasemgr

import (
    "context"
    "database/sql"
    "fmt"
    "sync"
    "time"

    "gorm.io/gorm"

    "com.litelake.litecore/common"
    "com.litelake.litecore/manager/databasemgr/internal/config"
    "com.litelake.litecore/manager/databasemgr/internal/drivers"
)

// Manager 数据库管理器
type Manager struct {
    // 依赖注入字段
    Config common.BaseConfigProvider `inject:""`

    // 内部状态
    name      string
    driver    string
    db        *gorm.DB
    sqlDB     *sql.DB
    mu        sync.RWMutex
    once      sync.Once
}

// NewManager 创建数据库管理器
func NewManager(name string) *Manager {
    return &Manager{
        name:   name,
        driver: "none",
    }
}

// ManagerName 返回管理器名称
func (m *Manager) ManagerName() string {
    return m.name
}

// OnStart 初始化管理器
func (m *Manager) OnStart() error {
    return m.once.Do(func() error {
        // 1. 从 Config 获取配置
        cfg, err := m.loadConfig()
        if err != nil {
            return fmt.Errorf("load config failed: %w", err)
        }

        // 2. 创建驱动
        driver, err := drivers.NewGormDriver(cfg)
        if err != nil {
            return fmt.Errorf("create driver failed: %w", err)
        }

        // 3. 获取 GORM 实例
        m.db = driver.DB()
        m.driver = cfg.Driver

        // 4. 获取 sql.DB 用于连接池管理
        m.sqlDB, err = m.db.DB()
        if err != nil {
            return fmt.Errorf("get sql.DB failed: %w", err)
        }

        // 5. 测试连接
        ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
        defer cancel()
        if err := m.sqlDB.PingContext(ctx); err != nil {
            return fmt.Errorf("ping database failed: %w", err)
        }

        return nil
    })
}

// loadConfig 从 ConfigProvider 加载配置
func (m *Manager) loadConfig() (*config.DatabaseConfig, error) {
    if m.Config == nil {
        return nil, fmt.Errorf("config provider is required")
    }

    cfgKey := fmt.Sprintf("database.%s", m.name)
    cfgData, ok := m.Config.Get(cfgKey)
    if !ok {
        return nil, fmt.Errorf("config not found: %s", cfgKey)
    }

    return config.ParseFromMap(cfgData)
}

// OnStop 停止管理器
func (m *Manager) OnStop() error {
    m.mu.Lock()
    defer m.mu.Unlock()

    if m.sqlDB == nil {
        return nil
    }

    err := m.sqlDB.Close()
    m.sqlDB = nil
    m.db = nil
    return err
}

// Health 健康检查
func (m *Manager) Health() error {
    m.mu.RLock()
    defer m.mu.RUnlock()

    if m.sqlDB == nil {
        return fmt.Errorf("database not initialized")
    }

    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()
    return m.sqlDB.PingContext(ctx)
}

// ========== GORM 核心 ==========

// DB 获取 GORM 数据库实例
func (m *Manager) DB() *gorm.DB {
    m.mu.RLock()
    defer m.mu.RUnlock()
    return m.db
}

// Model 指定模型进行操作
func (m *Manager) Model(value any) *gorm.DB {
    return m.DB().Model(value)
}

// Table 指定表名进行操作
func (m *Manager) Table(name string) *gorm.DB {
    return m.DB().Table(name)
}

// WithContext 设置上下文
func (m *Manager) WithContext(ctx context.Context) *gorm.DB {
    return m.DB().WithContext(ctx)
}

// ========== 事务管理 ==========

// Transaction 执行事务
func (m *Manager) Transaction(fn func(*gorm.DB) error, opts ...*sql.TxOptions) error {
    return m.DB().Transaction(fn, opts...)
}

// Begin 开启事务
func (m *Manager) Begin(opts ...*sql.TxOptions) *gorm.DB {
    return m.DB().Begin(opts...)
}

// ========== 迁移管理 ==========

// AutoMigrate 自动迁移
func (m *Manager) AutoMigrate(models ...any) error {
    return m.DB().AutoMigrate(models...)
}

// Migrator 获取迁移器
func (m *Manager) Migrator() gorm.Migrator {
    return m.DB().Migrator()
}

// ========== 连接管理 ==========

// Driver 获取数据库驱动类型
func (m *Manager) Driver() string {
    return m.driver
}

// Ping 检查数据库连接
func (m *Manager) Ping(ctx context.Context) error {
    m.mu.RLock()
    defer m.mu.RUnlock()

    if m.sqlDB == nil {
        return fmt.Errorf("database not initialized")
    }

    return m.sqlDB.PingContext(ctx)
}

// Stats 获取连接池统计信息
func (m *Manager) Stats() sql.DBStats {
    m.mu.RLock()
    defer m.mu.RUnlock()

    if m.sqlDB == nil {
        return sql.DBStats{}
    }

    return m.sqlDB.Stats()
}

// Close 关闭数据库连接
func (m *Manager) Close() error {
    return m.OnStop()
}

// ========== 原生 SQL ==========

// Exec 执行原生 SQL
func (m *Manager) Exec(sql string, values ...any) *gorm.DB {
    return m.DB().Exec(sql, values...)
}

// Raw 执行原生查询
func (m *Manager) Raw(sql string, values ...any) *gorm.DB {
    return m.DB().Raw(sql, values...)
}

// 确保 Manager 实现 DatabaseManager 和 common.BaseManager
var _ DatabaseManager = (*Manager)(nil)
var _ common.BaseManager = (*Manager)(nil)
```

---

### 4.4 CacheManager（依赖 LoggerManager, TelemetryManager）

**manager.go（新增）**：
```go
package cachemgr

import (
    "context"
    "fmt"
    "sync"
    "time"

    "go.opentelemetry.io/otel/attribute"
    "go.opentelemetry.io/otel/codes"
    "go.opentelemetry.io/otel/metric"
    "go.opentelemetry.io/otel/trace"

    "com.litelake.litecore/common"
    "com.litelake.litecore/manager/cachemgr/internal/config"
    "com.litelake.litecore/manager/cachemgr/internal/drivers"
    "com.litelake.litecore/manager/loggermgr"
    "com.litelake.litecore/manager/telemetrymgr"
)

// Manager 缓存管理器
type Manager struct {
    // 依赖注入字段
    Config            common.BaseConfigProvider      `inject:""`
    LoggerManager     loggermgr.LoggerManager       `inject:"optional"`
    TelemetryManager  telemetrymgr.TelemetryManager `inject:"optional"`

    // 内部状态
    name              string
    driver            drivers.Driver
    logger            loggermgr.Logger
    tracer            trace.Tracer
    meter             metric.Meter
    cacheHitCounter   metric.Int64Counter
    cacheMissCounter  metric.Int64Counter
    operationDuration metric.Float64Histogram
    mu                sync.RWMutex
    once              sync.Once
}

// NewManager 创建缓存管理器
func NewManager(name string) *Manager {
    return &Manager{
        name:   name,
        driver: drivers.NewNoneDriver(),
    }
}

// ManagerName 返回管理器名称
func (m *Manager) ManagerName() string {
    return m.name
}

// OnStart 初始化管理器
func (m *Manager) OnStart() error {
    return m.once.Do(func() error {
        // 1. 从 Config 获取配置
        cfg, err := m.loadConfig()
        if err != nil {
            return fmt.Errorf("load config failed: %w", err)
        }

        // 2. 创建驱动
        driver, err := drivers.NewDriver(cfg)
        if err != nil {
            return fmt.Errorf("create driver failed: %w", err)
        }
        m.driver = driver

        // 3. 初始化 Logger
        if m.LoggerManager != nil {
            m.logger = m.LoggerManager.Logger("cachemgr")
        }

        // 4. 初始化 Telemetry
        if m.TelemetryManager != nil {
            m.tracer = m.TelemetryManager.Tracer("cachemgr")
            m.meter = m.TelemetryManager.Meter("cachemgr")

            // 创建指标
            m.cacheHitCounter, _ = m.meter.Int64Counter(
                "cache.hit",
                metric.WithDescription("Cache hit count"),
                metric.WithUnit("{hit}"),
            )
            m.cacheMissCounter, _ = m.meter.Int64Counter(
                "cache.miss",
                metric.WithDescription("Cache miss count"),
                metric.WithUnit("{miss}"),
            )
            m.operationDuration, _ = m.meter.Float64Histogram(
                "cache.operation.duration",
                metric.WithDescription("Cache operation duration in seconds"),
                metric.WithUnit("s"),
            )
        }

        // 5. 启动驱动
        if err := m.driver.Start(); err != nil {
            return fmt.Errorf("start driver failed: %w", err)
        }

        return nil
    })
}

// loadConfig 从 ConfigProvider 加载配置
func (m *Manager) loadConfig() (*config.CacheConfig, error) {
    if m.Config == nil {
        return config.DefaultConfig(), nil
    }

    cfgKey := fmt.Sprintf("cache.%s", m.name)
    cfgData, ok := m.Config.Get(cfgKey)
    if !ok {
        return config.DefaultConfig(), nil
    }

    return config.ParseFromMap(cfgData)
}

// OnStop 停止管理器
func (m *Manager) OnStop() error {
    return m.driver.Stop()
}

// Health 健康检查
func (m *Manager) Health() error {
    return m.driver.Health()
}

// ========== 缓存操作 ==========

// Get 获取缓存值
func (m *Manager) Get(ctx context.Context, key string, dest any) error {
    var hit bool
    var getErr error

    err := m.recordOperation(ctx, "get", key, func() error {
        getErr = m.driver.Get(ctx, key, dest)
        hit = (getErr == nil)
        return getErr
    })

    m.recordCacheHit(ctx, hit)
    return err
}

// Set 设置缓存值
func (m *Manager) Set(ctx context.Context, key string, value any, expiration time.Duration) error {
    return m.recordOperation(ctx, "set", key, func() error {
        return m.driver.Set(ctx, key, value, expiration)
    })
}

// SetNX 仅当键不存在时才设置值
func (m *Manager) SetNX(ctx context.Context, key string, value any, expiration time.Duration) (bool, error) {
    var result bool
    var resultErr error

    err := m.recordOperation(ctx, "setnx", key, func() error {
        result, resultErr = m.driver.SetNX(ctx, key, value, expiration)
        return resultErr
    })

    return result, err
}

// Delete 删除缓存值
func (m *Manager) Delete(ctx context.Context, key string) error {
    return m.recordOperation(ctx, "delete", key, func() error {
        return m.driver.Delete(ctx, key)
    })
}

// Exists 检查键是否存在
func (m *Manager) Exists(ctx context.Context, key string) (bool, error) {
    var result bool
    var resultErr error

    err := m.recordOperation(ctx, "exists", key, func() error {
        result, resultErr = m.driver.Exists(ctx, key)
        return resultErr
    })

    return result, err
}

// Expire 设置过期时间
func (m *Manager) Expire(ctx context.Context, key string, expiration time.Duration) error {
    return m.recordOperation(ctx, "expire", key, func() error {
        return m.driver.Expire(ctx, key, expiration)
    })
}

// TTL 获取剩余过期时间
func (m *Manager) TTL(ctx context.Context, key string) (time.Duration, error) {
    var result time.Duration
    var resultErr error

    err := m.recordOperation(ctx, "ttl", key, func() error {
        result, resultErr = m.driver.TTL(ctx, key)
        return resultErr
    })

    return result, err
}

// Clear 清空所有缓存
func (m *Manager) Clear(ctx context.Context) error {
    return m.recordOperation(ctx, "clear", "", func() error {
        return m.driver.Clear(ctx)
    })
}

// GetMultiple 批量获取
func (m *Manager) GetMultiple(ctx context.Context, keys []string) (map[string]any, error) {
    var result map[string]any
    var resultErr error

    key := "batch"
    if len(keys) > 0 {
        key = keys[0]
    }

    err := m.recordOperation(ctx, "getmultiple", key, func() error {
        result, resultErr = m.driver.GetMultiple(ctx, keys)
        return resultErr
    })

    return result, err
}

// SetMultiple 批量设置
func (m *Manager) SetMultiple(ctx context.Context, items map[string]any, expiration time.Duration) error {
    key := "batch"
    for k := range items {
        key = k
        break
    }

    return m.recordOperation(ctx, "setmultiple", key, func() error {
        return m.driver.SetMultiple(ctx, items, expiration)
    })
}

// DeleteMultiple 批量删除
func (m *Manager) DeleteMultiple(ctx context.Context, keys []string) error {
    key := "batch"
    if len(keys) > 0 {
        key = keys[0]
    }

    return m.recordOperation(ctx, "deletemultiple", key, func() error {
        return m.driver.DeleteMultiple(ctx, keys)
    })
}

// Increment 自增
func (m *Manager) Increment(ctx context.Context, key string, value int64) (int64, error) {
    var result int64
    var resultErr error

    err := m.recordOperation(ctx, "increment", key, func() error {
        result, resultErr = m.driver.Increment(ctx, key, value)
        return resultErr
    })

    return result, err
}

// Decrement 自减
func (m *Manager) Decrement(ctx context.Context, key string, value int64) (int64, error) {
    var result int64
    var resultErr error

    err := m.recordOperation(ctx, "decrement", key, func() error {
        result, resultErr = m.driver.Decrement(ctx, key, value)
        return resultErr
    })

    return result, err
}

// Close 关闭缓存连接
func (m *Manager) Close() error {
    return m.OnStop()
}

// ========== 观测辅助方法 ==========

// recordOperation 记录缓存操作（带链路追踪和指标）
func (m *Manager) recordOperation(
    ctx context.Context,
    operation string,
    key string,
    fn func() error,
) error {
    // 如果没有可观测性配置，直接执行操作
    if m.tracer == nil && m.logger == nil && m.operationDuration == nil {
        return fn()
    }

    var span trace.Span
    if m.tracer != nil {
        ctx, span = m.tracer.Start(ctx, "cache."+operation,
            trace.WithAttributes(
                attribute.String("cache.key", sanitizeKey(key)),
                attribute.String("cache.driver", m.driver.Name()),
            ),
        )
        defer span.End()
    }

    start := time.Now()
    err := fn()
    duration := time.Since(start).Seconds()

    if m.operationDuration != nil {
        m.operationDuration.Record(ctx, duration,
            metric.WithAttributes(
                attribute.String("operation", operation),
                attribute.String("status", getStatus(err)),
            ),
        )
    }

    if m.logger != nil {
        if err != nil {
            m.logger.Error("cache operation failed",
                "operation", operation,
                "key", sanitizeKey(key),
                "error", err.Error(),
                "duration", duration,
            )
            if span != nil {
                span.RecordError(err)
                span.SetStatus(codes.Error, err.Error())
            }
        } else {
            m.logger.Debug("cache operation success",
                "operation", operation,
                "key", sanitizeKey(key),
                "duration", duration,
            )
        }
    }

    return err
}

// recordCacheHit 记录缓存命中/未命中
func (m *Manager) recordCacheHit(ctx context.Context, hit bool) {
    if m.meter == nil {
        return
    }

    attrs := metric.WithAttributes(
        attribute.String("cache.driver", m.driver.Name()),
    )

    if hit {
        if m.cacheHitCounter != nil {
            m.cacheHitCounter.Add(ctx, 1, attrs)
        }
    } else {
        if m.cacheMissCounter != nil {
            m.cacheMissCounter.Add(ctx, 1, attrs)
        }
    }
}

// sanitizeKey 对缓存键进行脱敏处理
func sanitizeKey(key string) string {
    if len(key) <= 10 {
        return key
    }
    return key[:5] + "***"
}

// getStatus 根据错误返回状态字符串
func getStatus(err error) string {
    if err != nil {
        return "error"
    }
    return "success"
}

// 确保 Manager 实现 CacheManager 和 common.BaseManager
var _ CacheManager = (*Manager)(nil)
var _ common.BaseManager = (*Manager)(nil)
```

---

## 五、配置方案改造

### 5.1 配置键命名规范

```
{manager_type}.{manager_name}
例如：
- telemetry.default
- logger.default
- database.primary
- cache.default
```

### 5.2 配置结构规范化

```go
// internal/config/config.go
package config

import (
    "fmt"
)

// TelemetryConfig 观测配置
type TelemetryConfig struct {
    Enabled     bool    `mapstructure:"enabled"`
    ServiceName string  `mapstructure:"service_name"`
    Endpoint    string  `mapstructure:"endpoint"`
    SampleRate  float64 `mapstructure:"sample_rate"`
}

// DefaultConfig 返回默认配置
func DefaultTelemetryConfig() *TelemetryConfig {
    return &TelemetryConfig{
        Enabled:     false,
        ServiceName: "litecore",
        Endpoint:    "",
        SampleRate:  1.0,
    }
}

// ParseFromMap 从 map 解析配置
func ParseTelemetryFromMap(data any) (*TelemetryConfig, error) {
    cfg := DefaultTelemetryConfig()

    if m, ok := data.(map[string]any); ok {
        if v, ok := m["enabled"].(bool); ok {
            cfg.Enabled = v
        }
        if v, ok := m["service_name"].(string); ok {
            cfg.ServiceName = v
        }
        if v, ok := m["endpoint"].(string); ok {
            cfg.Endpoint = v
        }
        if v, ok := m["sample_rate"].(float64); ok {
            cfg.SampleRate = v
        }
    }

    return cfg, nil
}
```

---

## 六、使用示例

### 6.1 基础使用

```go
package main

import (
    "com.litelake.litecore/container"
    "com.litelake.litecore/manager/cachemgr"
    "com.litelake.litecore/manager/databasemgr"
    "com.litelake.litecore/manager/loggermgr"
    "com.litelake.litecore/manager/telemetrymgr"
)

func main() {
    // 1. 创建容器（按依赖顺序）
    configContainer := container.NewConfigContainer()
    managerContainer := container.NewManagerContainer(configContainer)

    // 2. 加载配置
    cfg := &AppConfig{
        // ... 配置加载逻辑
    }
    configContainer.Register(cfg)

    // 3. 创建 Manager 实例（按任意顺序）
    telemetryMgr := telemetrymgr.NewManager("default")
    loggerMgr := loggermgr.NewManager("default")
    databaseMgr := databasemgr.NewManager("primary")
    cacheMgr := cachemgr.NewManager("default")

    // 4. 注册到容器（按任意顺序）
    managerContainer.Register(telemetryMgr)
    managerContainer.Register(loggerMgr)
    managerContainer.Register(databaseMgr)
    managerContainer.Register(cacheMgr)

    // 5. 执行依赖注入（自动拓扑排序）
    if err := managerContainer.InjectAll(); err != nil {
        panic(err)
    }

    // 6. 启动所有 Manager（按注入顺序自动启动）
    managers := managerContainer.GetAll()
    for _, mgr := range managers {
        if err := mgr.OnStart(); err != nil {
            panic(err)
        }
    }

    // 7. 使用 Manager
    db := databaseMgr.DB()
    // ...

    // 8. 优雅关闭
    for _, mgr := range managers {
        if err := mgr.OnStop(); err != nil {
            // 记录错误
        }
    }
}
```

### 6.2 在 Repository/Service 中使用

```go
// UserRepository 用户仓储
type UserRepository struct {
    Config    common.BaseConfigProvider   `inject:""`
    DBManager databasemgr.DatabaseManager `inject:""`
}

// 使用
func (r *UserRepository) FindByID(id int64) (*User, error) {
    var user User
    err := r.DBManager.DB().First(&user, id).Error
    return &user, err
}
```

---

## 七、改造步骤

### Phase 1：基础设施改造（第1周）
1. 移除所有 Factory 文件
2. 在各 Manager 包中创建 `manager.go`
3. 实现依赖注入结构

### Phase 2：配置系统改造（第2周）
1. 改造 config 包，支持从 `BaseConfigProvider` 获取配置
2. 实现配置验证和默认值
3. 更新配置文档

### Phase 3：依赖注入集成（第3周）
1. TelemetryManager 改造（无依赖）
2. LoggerManager 改造（依赖 Telemetry）
3. DatabaseManager 改造（无依赖）
4. CacheManager 改造（依赖 Logger, Telemetry）

### Phase 4：测试和文档（第4周）
1. 更新单元测试
2. 更新集成测试
3. 更新使用文档和示例

---

## 八、优势总结

### 8.1 架构优势
- ✅ 完全适配 Container 的 DI 机制
- ✅ 自动依赖解析，无需手动管理依赖顺序
- ✅ 支持同层依赖和跨层依赖
- ✅ 支持可选依赖，灵活性高

### 8.2 代码质量
- ✅ 移除 Factory 样板代码
- ✅ 结构更清晰，职责更明确
- ✅ 更容易测试（Mock 依赖更容易）

### 8.3 可维护性
- ✅ 依赖关系一目了然（通过 inject 标签）
- ✅ 新增 Manager 只需添加 inject 标签
- ✅ 符合 SOLID 原则

---

## 九、依赖关系图

```
ConfigProvider (配置层)
    ↓
    ├─→ TelemetryManager (无依赖)
    │       ↓
    │       └─→ LoggerManager (依赖 TelemetryManager)
    │               ↓
    │               └─→ CacheManager (依赖 LoggerManager, TelemetryManager)
    │
    └─→ DatabaseManager (无依赖)
```

---

## 十、注意事项

### 10.1 线程安全
- Manager 实例在 `InjectAll()` 后会并发访问
- 使用 `sync.RWMutex` 保护内部状态
- 使用 `sync.Once` 确保只初始化一次

### 10.2 错误处理
- `OnStart()` 返回错误会导致启动失败
- 缺少必须依赖会返回 `DependencyNotFoundError`
- 循环依赖会返回 `CircularDependencyError`

### 10.3 配置验证
- 在 `loadConfig()` 中验证配置
- 提供合理的默认值
- 配置缺失时返回明确错误

---

**文档结束**
