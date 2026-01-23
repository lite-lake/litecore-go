# Logger

统一的日志记录接口和默认实现，支持多级别日志、动态级别控制和上下文字段传递。

## 特性

- **统一接口** - 定义了 Debug、Info、Warn、Error、Fatal 五个标准级别
- **灵活级别控制** - 支持运行时动态调整日志级别
- **上下文字段** - 通过 With 方法添加额外字段到日志中
- **序列化支持** - LogLevel 类型支持 YAML/JSON 序列化
- **Zap 集成** - 提供与 uber-go/zap 框架的级别转换

## 快速开始

```go
package main

import "github.com/lite-lake/litecore-go/logger"

func main() {
    // 创建日志记录器
    log := logger.NewDefaultLogger("MyService")

    // 设置日志级别
    log.SetLevel(logger.DebugLevel)

    // 记录各级别日志
    log.Debug("调试信息", "request_id", "12345")
    log.Info("操作成功", "user_id", 100)
    log.Warn("查询较慢", "duration", "500ms")
    log.Error("连接失败", "host", "example.com")

    // 使用 With 添加上下文字段
    log.With("service", "payment").Info("处理订单", "order_id", "1001")
}
```

## 日志级别

日志级别从低到高依次为：Debug、Info、Warn、Error、Fatal。

设置某个级别后，只会记录该级别及以上的日志。

| 级别 | 值 | 说明 |
|------|-----|------|
| Debug | 0 | 调试信息，详细的技术细节 |
| Info | 1 | 一般信息，业务流程记录 |
| Warn | 2 | 警告信息，可能的问题 |
| Error | 3 | 错误信息，操作失败 |
| Fatal | 4 | 致命错误，程序退出 |

```go
// 只记录 Error 及以上级别的日志
logger.SetLevel(logger.ErrorLevel)
logger.Debug("这条不会被记录")  // 0 < 3，不记录
logger.Info("这条不会被记录")   // 1 < 3，不记录
logger.Warn("这条不会被记录")   // 2 < 3，不记录
logger.Error("这条会被记录")    // 3 = 3，记录
```

## 接口定义

```go
type ILogger interface {
    Debug(msg string, args ...any)
    Info(msg string, args ...any)
    Warn(msg string, args ...any)
    Error(msg string, args ...any)
    Fatal(msg string, args ...any)
    With(args ...any) ILogger
    SetLevel(level LogLevel)
}
```

## With 方法

With 方法返回一个新的 Logger 实例，带有额外的上下文字段。

```go
// 基础使用
logger.With("request_id", "12345").Info("处理请求")

// 链式调用，累积字段
logger.With("service", "payment").
       With("version", "1.0").
       Info("服务启动", "port", 8080)

// 原始 Logger 不受影响
log1 := logger.With("key1", "value1")
log2 := logger.With("key2", "value2")  // log2 不包含 key1
```

## 日志级别解析

```go
// 从字符串解析
level := logger.ParseLogLevel("debug")      // 返回 DebugLevel
level := logger.ParseLogLevel("INFO")        // 返回 InfoLevel
level := logger.ParseLogLevel("warning")     // 返回 WarnLevel
level := logger.ParseLogLevel("invalid")     // 返回默认的 InfoLevel

// 验证级别字符串是否有效
valid := logger.IsValidLogLevel("debug")   // true
valid := logger.IsValidLogLevel("invalid") // false
```

## 序列化支持

LogLevel 类型实现了 `encoding.TextMarshaler` 和 `encoding.TextUnmarshaler` 接口，可直接用于 YAML/JSON 配置。

```go
// 序列化
level := logger.InfoLevel
data, _ := level.MarshalText()  // 返回 []byte("info")

// 反序列化
var level logger.LogLevel
level.UnmarshalText([]byte("debug"))  // level = DebugLevel
```

## 与 Zap 集成

```go
import "go.uber.org/zap/zapcore"

// 将 LogLevel 转换为 zapcore.Level
zapLevel := logger.LogLevelToZap(logger.InfoLevel)
// zapLevel = zapcore.InfoLevel

// 将 zapcore.Level 转换为 LogLevel
logLevel := logger.ZapToLogLevel(zapcore.DebugLevel)
// logLevel = DebugLevel
```

## 自定义 Logger 实现

通过实现 ILogger 接口，可以创建自定义的日志记录器：

```go
type MyLogger struct {
    // 你的字段
}

func (l *MyLogger) Debug(msg string, args ...any) {
    // 自定义实现
}

func (l *MyLogger) Info(msg string, args ...any) {
    // 自定义实现
}

// ... 实现其他接口方法

var _ logger.ILogger = (*MyLogger)(nil)
```

## 使用规范

### 日志级别选择

- **Debug**: 开发调试信息，生产环境通常关闭
- **Info**: 正常业务流程，如请求开始、资源创建
- **Warn**: 降级处理、慢查询、重试等需要注意但不影响功能的情况
- **Error**: 业务错误、操作失败，需要人工关注
- **Fatal**: 致命错误，需要立即终止程序

### 字段传递规范

使用键值对方式传递字段：

```go
// 推荐：结构化字段
logger.Info("用户登录", "user_id", 123, "ip", "192.168.1.1")

// 推荐：使用 With 添加公共字段
log := logger.With("service", "payment")
log.Info("处理订单", "order_id", "1001")
log.Error("支付失败", "order_id", "1002")
```

### 敏感信息处理

日志中不应包含密码、token、密钥等敏感信息：

```go
// 不推荐
logger.Info("用户登录", "password", "secret123")

// 推荐
logger.Info("用户登录", "user_id", 123, "has_password", true)
```

## API 参考

### 接口

| 方法 | 说明 |
|------|------|
| `Debug(msg, args...)` | 记录调试级别日志 |
| `Info(msg, args...)` | 记录信息级别日志 |
| `Warn(msg, args...)` | 记录警告级别日志 |
| `Error(msg, args...)` | 记录错误级别日志 |
| `Fatal(msg, args...)` | 记录致命错误并退出 |
| `With(args...)` | 返回带额外字段的新 Logger |
| `SetLevel(level)` | 设置日志级别 |

### 函数

| 函数 | 说明 |
|------|------|
| `NewDefaultLogger(name)` | 创建默认日志记录器 |
| `ParseLogLevel(s)` | 从字符串解析日志级别 |
| `IsValidLogLevel(s)` | 检查日志级别字符串是否有效 |
| `LogLevelToZap(level)` | 转换 LogLevel 到 zapcore.Level |
| `ZapToLogLevel(level)` | 转换 zapcore.Level 到 LogLevel |

### 方法

| 方法 | 说明 |
|------|------|
| `LogLevel.String()` | 返回日志级别的字符串表示 |
| `LogLevel.Validate()` | 验证日志级别是否有效 |
| `LogLevel.Int()` | 返回日志级别的整数值 |
| `LogLevel.MarshalText()` | 序列化为文本 |
| `LogLevel.UnmarshalText(data)` | 从文本反序列化 |
