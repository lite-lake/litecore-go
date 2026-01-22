# Logger Manager

日志管理器模块，提供统一的日志管理接口，支持多种日志驱动。

## 特性

- **多驱动支持** - 支持 zap（高性能）、default（简单）、none（空实现）三种日志驱动
- **OpenTelemetry 集成** - 支持将日志输出到观测平台
- **多输出支持** - 同时支持控制台、文件和观测日志输出
- **日志轮转** - 支持按大小、时间轮转日志文件，并可压缩旧日志
- **级别过滤** - 支持动态调整日志级别
- **线程安全** - 支持并发日志写入
- **彩色输出** - 控制台输出支持彩色分级显示

## 快速开始

```go
import (
    "github.com/lite-lake/litecore-go/server/builtin/manager/loggermgr"
    "github.com/lite-lake/litecore-go/server/builtin/manager/telemetrymgr"
)

// 创建 Zap 日志管理器
cfg := &loggermgr.Config{
    Driver: "zap",
    ZapConfig: &loggermgr.DriverZapConfig{
        ConsoleEnabled: true,
        ConsoleConfig: &loggermgr.LogLevelConfig{Level: "info"},
    },
}

mgr, err := loggermgr.Build(cfg, telemetryMgr)
if err != nil {
    panic(err)
}

// 获取日志实例并使用
log := mgr.Ins()
log.Info("应用启动", "port", 8080)
log.Warn("警告信息", "reason", "连接超时")
log.Error("错误信息", "error", err)
```

## 配置

### 驱动类型

支持三种驱动类型：

| 驱动 | 说明 | 适用场景 |
|------|------|----------|
| zap | 高性能日志驱动，支持多输出和轮转 | 生产环境 |
| default | 简单日志驱动，输出到标准输出 | 开发环境 |
| none | 空日志驱动，不输出任何日志 | 测试环境 |

### Zap 驱动配置

```yaml
driver: zap
zap_config:
  # 控制台输出配置
  console_enabled: true
  console_config:
    level: info  # debug, info, warn, error, fatal

  # 文件输出配置
  file_enabled: true
  file_config:
    level: debug
    path: ./logs/app.log
    rotation:
      max_size: 100    # 单文件最大 MB
      max_age: 30      # 保留天数
      max_backups: 10  # 保留文件数
      compress: true   # 压缩旧文件

  # 观测输出配置
  telemetry_enabled: true
  telemetry_config:
    level: info
```

## API

### 接口

#### ILoggerManager

日志管理器接口，继承自 `common.IBaseManager`。

```go
type ILoggerManager interface {
    common.IBaseManager
    Ins() logger.ILogger
}
```

### 主要函数

#### Build

根据配置创建日志管理器。

```go
func Build(config *Config, telemetryMgr telemetrymgr.ITelemetryManager) (ILoggerManager, error)
```

#### BuildWithConfigProvider

通过配置提供者创建日志管理器。

```go
func BuildWithConfigProvider(configProvider IConfigProvider, telemetryMgr telemetrymgr.ITelemetryManager) (ILoggerManager, error)
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

日志管理器配置。

```go
type Config struct {
    Driver    string           // 驱动类型: zap, default, none
    ZapConfig *DriverZapConfig // Zap 驱动配置
}
```

#### DriverZapConfig

Zap 驱动配置。

```go
type DriverZapConfig struct {
    TelemetryEnabled bool            // 是否启用观测日志
    TelemetryConfig  *LogLevelConfig // 观测日志配置
    ConsoleEnabled   bool            // 是否启用控制台日志
    ConsoleConfig    *LogLevelConfig // 控制台日志配置
    FileEnabled      bool            // 是否启用文件日志
    FileConfig       *FileLogConfig  // 文件日志配置
}
```

#### LogLevelConfig

日志级别配置。

```go
type LogLevelConfig struct {
    Level string // 日志级别: debug, info, warn, error, fatal
}
```

#### FileLogConfig

文件日志配置。

```go
type FileLogConfig struct {
    Level    string          // 日志级别
    Path     string          // 日志文件路径
    Rotation *RotationConfig // 日志轮转配置
}
```

#### RotationConfig

日志轮转配置。

```go
type RotationConfig struct {
    MaxSize    int  // 单个日志文件最大大小（MB）
    MaxAge     int  // 日志文件保留天数
    MaxBackups int  // 保留的旧日志文件最大数量
    Compress   bool // 是否压缩旧日志文件
}
```

## 使用示例

### 多输出配置

```go
cfg := &loggermgr.Config{
    Driver: "zap",
    ZapConfig: &loggermgr.DriverZapConfig{
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
    },
}
```

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

## 注意事项

1. **Fatal 级别** - 调用 `Fatal` 方法后程序会退出，仅在严重错误时使用
2. **并发安全** - 所有日志操作都是线程安全的，可以在多个 goroutine 中并发使用
3. **文件路径** - 确保日志文件目录有写入权限，目录不存在时会自动创建
4. **性能考虑** - 高并发场景下建议使用 zap 驱动以获得最佳性能
5. **观测集成** - 启用观测输出需要正确配置 TelemetryManager

## 最佳实践

1. **合理选择驱动** - 开发环境使用 default，生产环境使用 zap
2. **分级输出** - 控制台输出 info 及以上级别，文件输出 debug 及以上级别
3. **使用上下文** - 通过 `With` 方法添加服务、版本等元信息
4. **避免过载** - 生产环境避免使用 debug 级别
5. **结构化日志** - 使用 key-value 格式而非拼接字符串
