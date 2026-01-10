# Logger Manager

提供灵活、高性能的日志管理功能，支持多种输出方式和日志级别。

## 特性

- **多输出支持** - 支持控制台、文件和观测日志（OpenTelemetry）三种输出方式，可独立配置
- **灵活配置** - 支持通过配置文件或代码配置日志级别、输出格式和日志轮转策略
- **高性能** - 基于 zap 高性能日志库，支持结构化日志和并发写入
- **日志轮转** - 支持按大小、时间和备份数量自动轮转日志文件
- **零成本降级** - 配置失败时自动降级到空日志器，避免影响程序运行
- **线程安全** - 所有操作都是线程安全的，支持并发使用

## 快速开始

```go
package main

import (
    "context"
    "com.litelake.litecore/manager/loggermgr"
)

func main() {
    // 配置日志（启用控制台输出）
    cfg := map[string]any{
        "console_enabled": true,
        "console_config": map[string]any{
            "level": "info",
        },
    }

    // 构建日志管理器
    mgr := loggermgr.Build(cfg, nil)

    // 获取 Logger 实例
    logger := mgr.(loggermgr.LoggerManager).Logger("my-app")

    // 记录日志
    logger.Info("Application started", "version", "1.0.0")
    logger.Warn("Configuration loaded", "config_file", "app.yaml")

    // 关闭日志管理器
    _ = mgr.(loggermgr.LoggerManager).Shutdown(context.Background())
}
```

## 日志级别

包支持五个日志级别，按优先级从低到高：

- **DebugLevel** - 调试信息，用于开发阶段
- **InfoLevel** - 一般信息，记录程序运行状态
- **WarnLevel** - 警告信息，表示潜在问题
- **ErrorLevel** - 错误信息，记录程序错误
- **FatalLevel** - 致命错误，记录后退出程序

### 使用日志级别

```go
// 解析日志级别字符串
level := loggermgr.ParseLogLevel("debug")

// 使用常量
logger.SetLevel(loggermgr.InfoLevel)

// 记录不同级别的日志
logger.Debug("Debug information", "detail", "value")
logger.Info("Information message", "status", "ok")
logger.Warn("Warning message", "issue", "potential problem")
logger.Error("Error occurred", "error", err)
logger.Fatal("Fatal error", "reason", "critical failure")
```

## 控制台输出

控制台输出支持彩色显示和时间戳格式化，适合开发环境使用。

```go
cfg := map[string]any{
    "console_enabled": true,
    "console_config": map[string]any{
        "level": "debug", // 设置日志级别
    },
}

mgr := loggermgr.Build(cfg, nil)
```

### 控制台配置选项

- `console_enabled` - 是否启用控制台输出（布尔值）
- `console_config.level` - 控制台日志级别（debug/info/warn/error/fatal）

## 文件输出

文件输出支持日志轮转和压缩，适合生产环境使用。

```go
cfg := map[string]any{
    "file_enabled": true,
    "file_config": map[string]any{
        "level": "info",
        "path":  "./logs/app.log",
        "rotation": map[string]any{
            "max_size":    "100MB",  // 单个文件最大 100MB
            "max_age":     "30d",    // 保留 30 天
            "max_backups": 10,       // 最多保留 10 个备份
            "compress":    true,     // 压缩旧文件
        },
    },
}

mgr := loggermgr.Build(cfg, nil)
```

### 文件配置选项

- `file_enabled` - 是否启用文件输出（布尔值）
- `file_config.level` - 文件日志级别
- `file_config.path` - 日志文件路径
- `file_config.rotation.max_size` - 单个日志文件最大大小（MB），支持格式：100MB、1GB、100
- `file_config.rotation.max_age` - 日志文件保留天数，支持格式：30d、48h
- `file_config.rotation.max_backups` - 保留的旧日志文件最大数量（整数）
- `file_config.rotation.compress` - 是否压缩旧日志文件（布尔值）

### 日志轮转

日志轮转会在以下情况下触发：

1. 当前日志文件大小达到 `max_size` 限制
2. 日志文件保留时间超过 `max_age` 限制
3. 备份文件数量超过 `max_backups` 限制

轮转后的文件名格式：`app.log.2024-01-01.001.gz`

## 观测日志

观测日志支持将日志发送到 OpenTelemetry 后端，用于分布式追踪和监控。

```go
cfg := map[string]any{
    "telemetry_enabled": true,
    "telemetry_config": map[string]any{
        "level": "info",
    },
    "console_enabled": true,
    "console_config": map[string]any{
        "level": "debug",
    },
}

// 需要传入 TelemetryManager 实例
mgr := factory.Build(cfg, telemetryMgr)
```

### 观测配置选项

- `telemetry_enabled` - 是否启用观测日志（布尔值）
- `telemetry_config.level` - 观测日志级别

## 组合输出

可以同时启用多种输出方式，日志会同时发送到所有配置的输出。

```go
cfg := map[string]any{
    "telemetry_enabled": true,
    "telemetry_config": map[string]any{
        "level": "info",
    },
    "console_enabled": true,
    "console_config": map[string]any{
        "level": "debug",
    },
    "file_enabled": true,
    "file_config": map[string]any{
        "level": "warn",
        "path":  "./logs/app.log",
        "rotation": map[string]any{
            "max_size":    "100MB",
            "max_age":     "30d",
            "max_backups": 10,
            "compress":    true,
        },
    },
}

mgr := factory.Build(cfg, telemetryMgr)
```

## Logger 实例

### 获取 Logger

```go
// 获取指定名称的 Logger 实例
logger := mgr.(loggermgr.LoggerManager).Logger("my-app")

// Logger 实例会被缓存，重复获取相同名称的 Logger 会返回同一实例
logger2 := mgr.(loggermgr.LoggerManager).Logger("my-app")
// logger == logger2
```

### 使用结构化日志

```go
// 使用键值对记录结构化日志
logger.Info("User logged in",
    "user_id", 12345,
    "username", "john.doe",
    "ip_address", "192.168.1.1")

logger.Error("Failed to process request",
    "request_id", "req-123",
    "error", err,
    "retry_count", 3)
```

### 添加固定字段

使用 `With` 方法创建带有固定字段的新 Logger：

```go
// 创建带有固定字段的 Logger
serviceLogger := logger.With(
    "service", "api-server",
    "version", "1.0.0",
)

// 所有日志都会包含固定字段
serviceLogger.Info("Request received", "path", "/api/users")
// 输出包含：service=api-server, version=1.0.0, path=/api/users
```

### 动态设置日志级别

```go
// 运行时动态调整日志级别
logger.SetLevel(loggermgr.DebugLevel)

// 设置全局日志级别
mgr.(loggermgr.LoggerManager).SetGlobalLevel(loggermgr.WarnLevel)
```

## API

### 构建函数

使用包级函数创建日志管理器实例。

```go
// 从配置 map 构建日志管理器
func Build(cfg map[string]any, telemetryMgr telemetrymgr.TelemetryManager) common.Manager

// 从配置结构体构建日志管理器
func BuildWithConfig(loggerConfig *config.LoggerConfig, telemetryMgr telemetrymgr.TelemetryManager) (common.Manager, error)
```

### LoggerManager

日志管理器接口，用于管理 Logger 实例。

```go
// 获取指定名称的 Logger 实例
func Logger(name string) Logger

// 设置全局日志级别
func SetGlobalLevel(level LogLevel)

// 关闭日志管理器
func Shutdown(ctx context.Context) error
```

### Logger

日志接口，提供日志记录功能。

```go
// 记录调试级别日志
func Debug(msg string, args ...any)

// 记录信息级别日志
func Info(msg string, args ...any)

// 记录警告级别日志
func Warn(msg string, args ...any)

// 记录错误级别日志
func Error(msg string, args ...any)

// 记录致命错误级别日志，然后退出程序
func Fatal(msg string, args ...any)

// 返回一个带有额外字段的新 Logger
func With(args ...any) Logger

// 设置日志级别
func SetLevel(level LogLevel)
```

## 错误处理

日志管理器采用零成本降级策略，确保配置失败不会影响程序运行：

```go
// 配置解析失败时，返回空日志管理器
cfg := map[string]any{
    "console_enabled": false,
    "file_enabled":    false,
}
mgr := loggermgr.Build(cfg, nil)
// mgr 是 NoneLoggerManager，所有日志操作都是空操作

// 日志初始化失败时，自动降级到控制台输出
cfg := map[string]any{
    "file_enabled": true,
    "file_config": map[string]any{
        "path": "/invalid/path/app.log", // 无效路径
    },
}
mgr := loggermgr.Build(cfg, nil)
// 自动降级到控制台输出，避免程序崩溃
```

## 性能考虑

- **Logger 缓存** - Logger 实例会被缓存，重复获取相同名称的 Logger 会返回同一实例
- **异步写入** - 日志写入采用异步方式，避免阻塞主程序
- **级别检查** - 日志级别检查在调用前完成，避免不必要的字符串格式化
- **线程安全** - 所有操作都是线程安全的，支持并发写入

```go
// 并发使用示例
for i := 0; i < 100; i++ {
    go func(i int) {
        logger.Info("Concurrent message", "index", i)
    }(i)
}
```

## 最佳实践

1. **使用有意义的 Logger 名称**

```go
// 好的做法
logger := mgr.Logger("api-server")
logger := mgr.Logger("database.connection")

// 不好的做法
logger := mgr.Logger("logger1")
logger := mgr.Logger("abc")
```

2. **使用结构化日志**

```go
// 好的做法
logger.Info("User logged in", "user_id", 12345, "ip", "192.168.1.1")

// 不好的做法
logger.Info(fmt.Sprintf("User %d logged in from %s", 12345, "192.168.1.1"))
```

3. **合理设置日志级别**

```go
// 开发环境
logger.SetLevel(loggermgr.DebugLevel)

// 生产环境
logger.SetLevel(loggermgr.InfoLevel)
```

4. **使用 With 添加固定字段**

```go
// 为服务添加固定字段
serviceLogger := logger.With("service", "api-server", "version", "1.0.0")

// 为请求添加固定字段
requestLogger := serviceLogger.With("request_id", reqID)
```

5. **正确关闭日志管理器**

```go
defer func() {
    if err := mgr.Shutdown(context.Background()); err != nil {
        log.Printf("Failed to shutdown logger: %v", err)
    }
}()
```