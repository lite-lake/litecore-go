# LoggerManager 依赖注入重构说明

## 概述

本次重构将 `loggermgr` 包从 **Factory 模式** 迁移到 **依赖注入（DI）模式**，以更好地适配 `container` 包的依赖注入机制。

## 重构目标

1. **移除 Factory 模式**：废弃 `factory.go` 中的 `Build()` 和 `BuildWithConfig()` 方法
2. **实现 DI 模式**：创建 `manager.go`，支持通过 `inject` 标签自动注入依赖
3. **保持向后兼容**：保留现有接口，确保旧代码可以继续使用
4. **提升可测试性**：通过依赖注入使单元测试更容易编写

## 新增文件

### 1. manager/loggermgr/manager.go
实现新的 `Manager` 结构体，支持依赖注入：

```go
type Manager struct {
    // Config 配置提供者（必须依赖）
    Config common.BaseConfigProvider `inject:""`

    // TelemetryManager 观测管理器（可选依赖）
    TelemetryManager telemetrymgr.TelemetryManager `inject:"optional"`

    // 内部状态
    name   string
    driver drivers.Driver
    level  LogLevel
    mu     sync.RWMutex
    once   sync.Once
}
```

**核心方法**：
- `NewManager(name string) *Manager` - 创建管理器实例
- `OnStart() error` - 初始化管理器（在依赖注入完成后调用）
- `OnStop() error` - 停止管理器
- `Logger(name string) Logger` - 获取 Logger 实例
- `SetGlobalLevel(level LogLevel)` - 设置全局日志级别

### 2. manager/loggermgr/internal/drivers/driver.go
定义统一的驱动接口：

```go
type Driver interface {
    Start() error
    Shutdown(ctx context.Context) error
    Health() error
    GetLogger(name string) Logger
    SetLevel(level loglevel.LogLevel)
}

type Logger interface {
    Debug(msg string, args ...any)
    Info(msg string, args ...any)
    Warn(msg string, args ...any)
    Error(msg string, args ...any)
    Fatal(msg string, args ...any)
    With(args ...any) Logger
    SetLevel(level loglevel.LogLevel)
}
```

### 3. manager/loggermgr/internal/drivers/zap_driver.go
实现 `ZapDriver`，封装 `ZapLoggerManager`：

```go
type ZapDriver struct {
    config       *config.LoggerConfig
    telemetryMgr telemetrymgr.TelemetryManager
    manager      *ZapLoggerManager
    mu           sync.RWMutex
    started      bool
}
```

### 4. manager/loggermgr/internal/drivers/none_driver.go
实现 `NoneDriver`，提供空实现：

```go
type NoneDriver struct{}

func NewNoneDriver() *NoneDriver
```

### 5. manager/loggermgr/internal/config/config.go
添加默认配置函数：

```go
func DefaultLoggerConfig() *LoggerConfig
```

返回默认配置：
- 控制台日志：启用，info 级别
- 文件日志：禁用
- 观测日志：禁用

## 配置格式

### 配置键命名规范
```
logger.{manager_name}
```

例如：
- `logger.default` - 默认日志管理器
- `logger.app` - 应用日志管理器
- `logger.audit` - 审计日志管理器

### 配置示例

```yaml
logger:
  default:
    console_enabled: true
    console_config:
      level: info
    file_enabled: true
    file_config:
      level: debug
      path: /var/log/app/app.log
      rotation:
        max_size: 100
        max_age: 30
        max_backups: 10
        compress: true
    telemetry_enabled: false
```

## 使用方式

### 旧方式（Factory 模式 - 已废弃）

```go
// Deprecated
cfg := map[string]any{
    "console_enabled": true,
    "console_config": map[string]any{
        "level": "info",
    },
}
loggerMgr := loggermgr.Build(cfg, telemetryMgr)
```

### 新方式（DI 模式 - 推荐）

```go
// 1. 创建管理器
loggerMgr := loggermgr.NewManager("default")

// 2. 注册到容器
container.Register(loggerMgr)

// 3. 执行依赖注入
container.InjectAll()

// 4. 启动管理器
if err := loggerMgr.OnStart(); err != nil {
    panic(err)
}

// 5. 使用
logger := loggerMgr.Logger("service")
logger.Info("Application started")

// 6. 关闭
loggerMgr.Shutdown(ctx)
```

### 简化方式（手动注入）

```go
// 创建管理器
loggerMgr := loggermgr.NewManager("default")

// 手动注入配置
configProvider := &MyConfigProvider{}
loggerMgr.Config = configProvider

// 启动
if err := loggerMgr.OnStart(); err != nil {
    panic(err)
}

// 使用
logger := loggerMgr.Logger("service")
logger.Info("Service started")
```

## 迁移指南

### 步骤 1：替换创建方式

**之前**：
```go
loggerMgr := loggermgr.Build(cfg, telemetryMgr)
```

**之后**：
```go
loggerMgr := loggermgr.NewManager("default")
// 配置通过 ConfigProvider 注入
```

### 步骤 2：配置迁移

**之前**（直接传递配置）：
```go
cfg := map[string]any{
    "console_enabled": true,
    "console_config": map[string]any{"level": "info"},
}
loggerMgr := loggermgr.Build(cfg, nil)
```

**之后**（通过 ConfigProvider）：
```go
// 在配置文件中定义
config := &AppConfig{
    Loggers: map[string]any{
        "logger.default": map[string]any{
            "console_enabled": true,
            "console_config": map[string]any{"level": "info"},
        },
    },
}

// 注入到管理器
loggerMgr := loggermgr.NewManager("default")
container.Register(config)
container.Register(loggerMgr)
container.InjectAll()
```

### 步骤 3：依赖关系

LoggerManager 依赖 TelemetryManager（可选）：

```go
// TelemetryManager 需要先注册和初始化
telemetryMgr := telemetrymgr.NewManager("default")
container.Register(telemetryMgr)

// LoggerManager 会自动注入 TelemetryManager
loggerMgr := loggermgr.NewManager("default")
container.Register(loggerMgr)

// 执行依赖注入
container.InjectAll()
```

## 依赖关系图

```
ConfigProvider (配置层)
    ↓
    └─→ TelemetryManager (无依赖)
            ↓
            └─→ LoggerManager (依赖 TelemetryManager)
```

## 设计原则

### 1. 依赖注入原则
- 所有依赖通过 `inject` 标签声明
- Container 自动解析和注入依赖
- 支持可选依赖 (`inject:"optional"`)

### 2. 延迟初始化原则
- 构造函数只做最小初始化
- 配置读取和依赖初始化在 `OnStart()` 中完成
- 保证注入顺序正确后再初始化

### 3. 配置获取原则
- 从 `BaseConfigProvider` 获取配置
- 使用结构化配置而非 `map[string]any`
- 支持配置验证和默认值

### 4. 降级策略
- 配置加载失败：使用默认配置
- 驱动创建失败：降级到 NoneDriver
- TelemetryManager 不可用：不集成 OTEL，但不影响日志功能

## 接口兼容性

所有公开接口保持不变：

- `LoggerManager` 接口
- `Logger` 接口
- `LogLevel` 类型
- `ParseLogLevel()` 函数

现有的适配器继续可用：

- `LoggerAdapter` - 适配 `ZapLogger`
- `LoggerManagerAdapter` - 适配 `ZapLoggerManager`

## 测试

运行测试：
```bash
go test ./manager/loggermgr/... -v
```

运行示例测试：
```bash
go test ./manager/loggermgr/example_di_test.go -v
```

## 性能考虑

1. **线程安全**：所有公共方法都是线程安全的
2. **延迟初始化**：Logger 实例按需创建
3. **缓存机制**：ZapLoggerManager 缓存 Logger 实例
4. **零拷贝**：使用接口和适配器，避免不必要的类型转换

## 常见问题

### Q: 如何在没有配置的情况下使用？
A: 不提供配置时，会使用默认配置（控制台输出，info 级别）

### Q: 如何禁用日志？
A: 设置 `console_enabled: false` 和 `file_enabled: false`，会使用 NoneDriver

### Q: 如何集成 OTEL？
A: 注入 TelemetryManager，LoggerManager 会自动集成

### Q: 如何动态修改日志级别？
A: 使用 `SetGlobalLevel()` 或 `Logger.SetLevel()`

### Q: 是否支持多个独立的日志管理器？
A: 是的，可以创建多个 Manager 实例，每个使用不同的名称和配置

## 版本兼容性

- **v1.x** - Factory 模式（已废弃，但仍可用）
- **v2.x** - DI 模式（推荐）
- **v3.0** - 将完全移除 Factory 模式

## 参考资料

- [Manager 重构方案](../../docs/TRD-20260111-manager-refactoring.md)
- [Container 文档](../../container/README.md)
- [TelemetryManager 文档](../telemetrymgr/README.md)
