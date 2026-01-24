# Logger Manager

日志管理器，提供统一的日志管理接口，支持多种日志驱动。

## 特性

- **多驱动支持** - 支持 zap（高性能）、default（简单）、none（空实现）三种日志驱动
- **多输出支持** - 同时支持控制台、文件和观测日志三种输出方式
- **多格式支持** - 控制台输出支持 Gin 风格、JSON、默认三种格式
- **彩色输出** - 自动检测终端支持，可配置彩色分级显示
- **日志轮转** - 支持按大小、时间轮转日志文件，并可压缩旧日志
- **OpenTelemetry 集成** - 支持将日志输出到可观测平台
- **依赖注入** - 支持通过 DI 注入到各层组件
- **线程安全** - 支持并发日志写入

## 快速开始

### 通过配置文件创建

```yaml
logger:
  driver: zap
  zap_config:
    console_enabled: true
    console_config:
      level: info
      format: gin
      color: true
    file_enabled: false
    file_config:
      level: debug
      path: ./logs/app.log
```

### 依赖注入使用

```go
package services

import (
    "github.com/lite-lake/litecore-go/manager/loggermgr"
)

type MyService struct {
    LoggerMgr loggermgr.ILoggerManager `inject:""`
    logger    loggermgr.ILogger
}

func (s *MyService) OnStart() error {
    s.logger = s.LoggerMgr.Ins()
    return nil
}

func (s *MyService) DoWork() {
    s.logger.Info("开始处理", "task_id", 123)
    s.logger.Debug("处理详情", "items", 10)
    s.logger.Error("处理失败", "error", err)
}
```

### 直接创建

```go
import (
    "github.com/lite-lake/litecore-go/manager/loggermgr"
)

cfg := &loggermgr.Config{
    Driver: "zap",
    ZapConfig: &loggermgr.DriverZapConfig{
        ConsoleEnabled: true,
        ConsoleConfig:  &loggermgr.LogLevelConfig{Level: "info"},
    },
}

mgr, err := loggermgr.Build(cfg, nil)
if err != nil {
    panic(err)
}

log := mgr.Ins()
log.Info("应用启动", "port", 8080)
```

## 日志驱动

### Zap 驱动

高性能日志驱动，支持多输出和轮转，适用于生产环境。

```go
cfg := &loggermgr.DriverZapConfig{
    ConsoleEnabled: true,
    ConsoleConfig:  &loggermgr.LogLevelConfig{Level: "info"},
    FileEnabled: true,
    FileConfig: &loggermgr.FileLogConfig{
        Level: "debug",
        Path:  "./logs/app.log",
        Rotation: &loggermgr.RotationConfig{
            MaxSize:    100,
            MaxAge:     30,
            MaxBackups: 10,
            Compress:   true,
        },
    },
}

mgr, err := loggermgr.NewDriverZapLoggerManager(cfg, telemetryMgr)
```

### Default 驱动

简单日志驱动，输出到标准输出，适用于开发环境。

```go
mgr := loggermgr.NewDriverDefaultLoggerManager()
log := mgr.Ins()
log.Info("简单日志输出")
```

### None 驱动

空日志驱动，不输出任何日志，适用于测试环境。

```go
mgr := loggermgr.NewDriverNoneLoggerManager()
log := mgr.Ins()
log.Info("这条日志不会被输出")
```

## 日志格式

### Gin 格式（推荐）

竖线分隔符，适合控制台输出，支持彩色显示。

```
2026-01-24 15:04:05.123 | INFO  | 开始依赖注入 | count=23
2026-01-24 15:04:05.456 | WARN  | 慢查询检测 | duration=1.2s
2026-01-24 15:04:05.789 | ERROR | 数据库连接失败 | error="connection refused"
```

**颜色配置**：
- `color: true`：根据终端自动检测（默认）
- `color: false`：关闭彩色输出

**日志级别颜色**：
- DEBUG：灰色
- INFO：绿色
- WARN：黄色
- ERROR：红色
- FATAL：红色+粗体

### JSON 格式

适合日志分析和监控系统。

```json
{"time":"2026-01-24T15:04:05.123Z","level":"INFO","msg":"开始依赖注入","count":23}
```

### Default 格式

默认 ConsoleEncoder 格式。

```json
{"level":"info","time":"2026-01-24T15:04:05.123Z","msg":"开始依赖注入","count":23}
```

## 日志级别

支持五个日志级别：

| 级别 | 说明 | 使用场景 |
|------|------|----------|
| Debug | 调试信息 | 开发调试 |
| Info | 正常业务流程 | 请求开始/完成、资源创建 |
| Warn | 降级处理、慢查询、重试 | 需要关注但不影响运行 |
| Error | 业务错误、操作失败 | 需人工关注 |
| Fatal | 致命错误 | 需立即终止程序 |

## API

### 接口

#### ILoggerManager

```go
type ILoggerManager interface {
    common.IBaseManager
    Ins() logger.ILogger
}
```

### 工厂函数

#### Build

根据配置创建日志管理器。

```go
func Build(config *Config, telemetryMgr telemetrymgr.ITelemetryManager) (ILoggerManager, error)
```

#### BuildWithConfigProvider

通过配置提供者创建日志管理器（引擎自动调用）。

```go
func BuildWithConfigProvider(configProvider configmgr.IConfigManager, telemetryMgr telemetrymgr.ITelemetryManager) (ILoggerManager, error)
```

#### NewDriverZapLoggerManager

创建 Zap 驱动日志管理器。

```go
func NewDriverZapLoggerManager(cfg *DriverZapConfig, telemetryMgr telemetrymgr.ITelemetryManager) (ILoggerManager, error)
```

#### NewDriverDefaultLoggerManager

创建默认驱动日志管理器。

```go
func NewDriverDefaultLoggerManager() ILoggerManager
```

#### NewDriverNoneLoggerManager

创建空驱动日志管理器。

```go
func NewDriverNoneLoggerManager() ILoggerManager
```

### 配置类型

#### Config

```go
type Config struct {
    Driver    string           `yaml:"driver"`     // 驱动类型: zap, default, none
    ZapConfig *DriverZapConfig `yaml:"zap_config"` // Zap 驱动配置
}
```

#### DriverZapConfig

```go
type DriverZapConfig struct {
    TelemetryEnabled bool            `yaml:"telemetry_enabled"` // 是否启用观测日志
    TelemetryConfig  *LogLevelConfig `yaml:"telemetry_config"`  // 观测日志配置
    ConsoleEnabled   bool            `yaml:"console_enabled"`   // 是否启用控制台日志
    ConsoleConfig    *LogLevelConfig `yaml:"console_config"`    // 控制台日志配置
    FileEnabled      bool            `yaml:"file_enabled"`      // 是否启用文件日志
    FileConfig       *FileLogConfig  `yaml:"file_config"`       // 文件日志配置
}
```

#### LogLevelConfig

```go
type LogLevelConfig struct {
    Level      string `yaml:"level"`       // 日志级别: debug, info, warn, error, fatal
    Format     string `yaml:"format"`      // 日志格式: gin | json | default
    Color      bool   `yaml:"color"`       // 是否启用颜色输出（默认根据终端自动检测）
    TimeFormat string `yaml:"time_format"` // 时间格式（默认：2006-01-02 15:04:05.000）
}
```

#### FileLogConfig

```go
type FileLogConfig struct {
    Level    string          `yaml:"level"`    // 日志级别
    Path     string          `yaml:"path"`     // 日志文件路径
    Rotation *RotationConfig `yaml:"rotation"` // 日志轮转配置
}
```

#### RotationConfig

```go
type RotationConfig struct {
    MaxSize    int  `yaml:"max_size"`    // 单个日志文件最大大小（MB）
    MaxAge     int  `yaml:"max_age"`     // 日志文件保留天数
    MaxBackups int  `yaml:"max_backups"` // 保留的旧日志文件最大数量
    Compress   bool `yaml:"compress"`    // 是否压缩旧日志文件
}
```

## 高级用法

### 带上下文的日志

```go
log := mgr.Ins()

// 创建带上下文的 Logger
contextLogger := log.With("service", "user-service", "version", "1.0.0")
contextLogger.Info("处理用户请求", "user_id", 123)

// 链式调用
log.With("module", "auth").With("action", "login").Info("用户登录")
```

### 动态调整日志级别

```go
log := mgr.Ins()

// 设置为 Debug 级别
log.SetLevel(logger.DebugLevel)
log.Debug("调试信息")

// 设置为 Error 级别
log.SetLevel(logger.ErrorLevel)
log.Debug("这条不会被输出")
log.Error("错误信息")
```

### OpenTelemetry 集成

```go
cfg := &loggermgr.Config{
    Driver: "zap",
    ZapConfig: &loggermgr.DriverZapConfig{
        TelemetryEnabled: true,
        TelemetryConfig:  &loggermgr.LogLevelConfig{Level: "info"},
        ConsoleEnabled:   true,
        ConsoleConfig:    &loggermgr.LogLevelConfig{Level: "info"},
    },
}

mgr, err := loggermgr.Build(cfg, telemetryMgr)
```

## 敏感信息处理

日志输出时注意保护敏感信息：

```go
// 不推荐：直接输出密码
log.Error("登录失败", "password", userPassword)

// 推荐：脱敏处理
log.Error("登录失败", "user", username, "error", err)
```

## 注意事项

1. **Fatal 级别** - 调用 `Fatal` 方法后程序会退出，仅在严重错误时使用
2. **并发安全** - 所有日志操作都是线程安全的，可以在多个 goroutine 中并发使用
3. **文件权限** - 确保日志文件目录有写入权限，目录不存在时会自动创建
4. **性能考虑** - 高并发场景下建议使用 zap 驱动以获得最佳性能
5. **观测集成** - 启用观测输出需要正确配置 TelemetryManager

## 最佳实践

1. **合理选择驱动** - 开发环境使用 default，生产环境使用 zap
2. **分级输出** - 控制台输出 info 及以上级别，文件输出 debug 及以上级别
3. **使用上下文** - 通过 `With` 方法添加服务、版本等元信息
4. **结构化日志** - 使用 key-value 格式而非拼接字符串
5. **避免过载** - 生产环境避免使用 debug 级别
6. **依赖注入** - 通过 DI 注入 ILoggerManager，在 OnStart 中获取 logger 实例
