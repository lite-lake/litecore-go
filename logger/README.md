# Logger

统一的日志记录接口和默认实现，支持多级别日志、动态级别控制和上下文字段传递。

## 特性

- **统一接口** - 定义了 Debug、Info、Warn、Error、Fatal 五个标准级别
- **灵活级别控制** - 支持运行时动态调整日志级别
- **上下文字段** - 通过 With 方法添加额外字段到日志中
- **序列化支持** - LogLevel 类型支持 YAML/JSON 序列化
- **Zap 集成** - 提供与 uber-go/zap 框架的级别转换
- **多种日志格式** - 支持 Gin 风格、JSON、Default 三种日志格式
- **彩色输出** - 控制台支持彩色日志，自动检测终端支持
- **启动日志** - 记录应用启动过程和耗时统计

## 日志格式

日志管理器支持三种控制台日志格式，通过配置文件中的 `format` 字段指定：

### Gin 格式（默认）

Gin 风格的控制台输出格式，使用竖线分隔符和固定宽度对齐，适合开发调试。

**格式**：
```
{时间} | {级别} | {消息} | {字段1}={值1} {字段2}={值2} ...
```

**示例**：
```
2026-01-24 15:04:05.123 | INFO  | 开始依赖注入 | count=23
2026-01-24 15:04:05.456 | WARN  | 慢查询检测 | duration=1.2s
2026-01-24 15:04:05.789 | ERROR | 数据库连接失败 | error="connection refused"
```

**特点**：
- 时间固定宽度 23 字符：`2006-01-24 15:04:05.000`
- 级别固定宽度 5 字符，右对齐，带颜色
- 字段格式：`key=value`，字符串值用引号包裹
- 支持彩色输出（根据终端自动检测）

### JSON 格式

JSON 格式日志，适合日志分析和监控系统集成。

**示例**：
```json
{"time":"2026-01-24T15:04:05.123Z","level":"info","msg":"开始依赖注入","count":23}
```

### Default 格式

Zap 默认的 ConsoleEncoder 格式，向后兼容旧版本。

### 格式选择建议

- **开发环境**：使用 `gin` 格式，可读性高，便于调试
- **生产环境**：控制台使用 `gin` 格式，文件日志使用 JSON 格式
- **日志分析**：使用 `json` 格式，便于 ELK、Loki 等系统采集
- **向后兼容**：使用 `default` 格式恢复原有行为

## 快速开始

### 基础使用

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

### 框架集成使用

在框架中使用日志管理器：

```go
type MessageService struct {
    LoggerMgr loggermgr.ILoggerManager `inject:""`
    logger     logger.ILogger
}

func (s *MessageService) initLogger() {
    if s.LoggerMgr != nil {
        s.logger = s.LoggerMgr.Logger("MessageService")
    }
}

func (s *MessageService) CreateMessage(msg string) error {
    s.initLogger()
    
    s.logger.Info("创建消息", "content", msg)
    
    // 业务逻辑...
    
    s.logger.Debug("消息创建完成", "id", 123)
    return nil
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

## 配置说明

日志系统通过 YAML 配置文件进行配置，位于 `logger` 配置节点下。

### 完整配置示例

```yaml
logger:
  driver: "zap"              # 驱动类型：zap, default, none
  zap_config:
    # 观测日志（OpenTelemetry）
    telemetry_enabled: false # 是否启用观测日志
    telemetry_config:
      level: "info"         # 日志级别：debug, info, warn, error, fatal

    # 控制台日志
    console_enabled: true    # 是否启用控制台日志
    console_config:
      level: "info"          # 日志级别：debug, info, warn, error, fatal
      format: "gin"          # 格式：gin | json | default
      color: true            # 是否启用颜色输出
      time_format: "2006-01-24 15:04:05.000"  # 时间格式

    # 文件日志
    file_enabled: true       # 是否启用文件日志
    file_config:
      level: "debug"         # 日志级别
      path: "./logs/app.log" # 日志文件路径
      rotation:              # 日志轮转配置
        max_size: 100        # 单个日志文件最大大小（MB）
        max_age: 30          # 日志文件保留天数
        max_backups: 10      # 保留的旧日志文件最大数量
        compress: true       # 是否压缩旧日志文件
```

### 配置项说明

#### LogLevelConfig

| 字段 | 类型 | 默认值 | 说明 |
|------|------|--------|------|
| `level` | string | "info" | 日志级别：debug, info, warn, error, fatal |
| `format` | string | "gin" | 日志格式：gin（默认）\| json \| default |
| `color` | bool | true | 是否启用颜色输出（根据终端自动检测） |
| `time_format` | string | "2006-01-24 15:04:05.000" | 时间格式 |

#### FileLogConfig

| 字段 | 类型 | 默认值 | 说明 |
|------|------|--------|------|
| `level` | string | "info" | 日志级别 |
| `path` | string | "./logs/app.log" | 日志文件路径 |
| `rotation.max_size` | int | 100 | 单个日志文件最大大小（MB） |
| `rotation.max_age` | int | 30 | 日志文件保留天数 |
| `rotation.max_backups` | int | 10 | 保留的旧日志文件最大数量 |
| `rotation.compress` | bool | true | 是否压缩旧日志文件 |

### 颜色方案

Gin 格式支持彩色输出，根据终端能力自动启用：

| 日志级别 | ANSI 颜色 | 说明 |
|---------|----------|------|
| DEBUG | 灰色 | 开发调试信息 |
| INFO | 绿色 | 正常业务流程 |
| WARN | 黄色 | 降级处理、慢查询 |
| ERROR | 红色 | 业务错误、操作失败 |
| FATAL | 红色+粗体 | 致命错误 |

### 使用不同格式

```yaml
# 开发环境：Gin 格式 + 彩色输出
console_config:
  level: "debug"
  format: "gin"
  color: true

# 生产环境控制台：Gin 格式 + 仅 Info 及以上
console_config:
  level: "info"
  format: "gin"
  color: false

# 生产环境文件：JSON 格式
file_config:
  level: "debug"
  path: "./logs/app.log"
```

## 启动日志

框架提供启动日志功能，记录应用启动各阶段的详细信息和耗时统计。

### 启动日志配置

```yaml
server:
  startup_log:
    enabled: true     # 是否启用启动日志
    async: true       # 是否异步日志
    buffer: 100       # 日志缓冲区大小
```

### 启动日志输出

启动日志会记录以下信息：

- **配置加载阶段**：配置文件读取和解析
- **管理器初始化阶段**：各 Manager 组件初始化
- **依赖注入阶段**：Entity、Repository、Service、Controller、Middleware 各层注入
- **路由注册阶段**：HTTP 路由注册
- **组件启动阶段**：各组件 OnStart 方法调用
- **启动完成汇总**：总耗时和各阶段耗时统计

**示例输出**：
```
2026-01-24 15:04:05.123 | INFO  | 开始依赖注入
2026-01-24 15:04:05.124 | INFO  | [Repository 层] MessageRepository: 注入完成
2026-01-24 15:04:05.125 | INFO  | [Service 层] MessageService: 注入完成
2026-01-24 15:04:05.126 | INFO  | [Controller 层] MessageController: 注入完成
2026-01-24 15:04:05.127 | INFO  | [Middleware 层] RecoveryMiddleware: 注入完成
2026-01-24 15:04:05.128 | INFO  | 依赖注入完成 | count=4 | duration=5ms
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

### 格式选择建议

根据不同环境和使用场景选择合适的日志格式：

**开发环境**：
```yaml
console_config:
  level: "debug"      # 记录详细调试信息
  format: "gin"       # 使用 Gin 格式，可读性高
  color: true         # 启用颜色，便于区分
```

**生产环境**：
```yaml
console_config:
  level: "info"       # 仅记录 Info 及以上
  format: "gin"       # 控制台使用 Gin 格式
  color: false        # 关闭颜色，减少干扰

file_config:
  level: "debug"      # 文件记录 Debug 级别
  path: "./logs/app.log"
```

**日志分析环境**：
```yaml
console_config:
  level: "info"
  format: "json"      # 使用 JSON 格式，便于采集分析
  color: false

file_config:
  level: "debug"
  path: "./logs/app.log"
```

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

## Gin 格式优势

Gin 格式是默认的控制台日志格式，具有以下优势：

### 1. 可读性强

- **统一格式**：所有日志采用统一的竖线分隔符格式
- **固定宽度**：时间、级别等字段使用固定宽度，对齐整齐
- **简洁清晰**：字段采用 `key=value` 格式，一目了然

### 2. 视觉层次

- **彩色区分**：不同级别的日志使用不同颜色，快速识别
- **状态码颜色**：HTTP 状态码根据范围使用不同颜色（2xx 绿色，4xx 橙色，5xx 红色）
- **信息层次**：时间、级别、消息、字段层次分明

### 3. 开发友好

- **快速扫描**：竖线分隔符和固定宽度便于快速扫描日志
- **字段提取**：`key=value` 格式便于提取关键信息
- **调试体验**：彩色输出和清晰格式提升调试体验

### 4. 标准兼容

- **Gin 风格**：兼容 Gin 框架的日志格式
- **请求日志**：支持 HTTP 请求日志的标准化格式
- **生态兼容**：与现有日志工具和监控系统集成

### 5. 性能优化

- **高效编码**：使用 `zapcore.Encoder` 接口，性能优秀
- **缓冲复用**：使用 `buffer.Pool` 复用缓冲区，减少 GC 压力
- **异步日志**：支持异步日志写入，降低阻塞

## 向后兼容性

日志系统保持良好的向后兼容性：

- **默认格式**：默认使用 `gin` 格式，无需修改现有代码
- **格式切换**：提供 `default` 选项恢复原有格式
- **接口不变**：`ILogger` 接口保持不变，无需修改业务代码
- **配置兼容**：未指定 `format` 字段时，自动使用 `gin` 格式

如需恢复原有格式，在配置文件中设置：

```yaml
console_config:
  format: "default"  # 使用原有格式
```

## 技术实现

### GinConsoleEncoder

`GinConsoleEncoder` 是 Gin 格式日志的核心编码器，位于 `manager/loggermgr/encoder_gin.go`。

**核心特性**：

- 实现 `zapcore.Encoder` 接口
- 支持自定义时间格式
- 支持彩色输出（基于 ANSI 颜色代码）
- 支持多种字段类型（字符串、数字、布尔、时间等）
- 使用固定宽度和竖线分隔符

**颜色支持**：

自动检测终端是否支持颜色输出，检测逻辑包括：

- 检查 `NO_COLOR` 环境变量
- 检查 `TERM` 环境变量
- 检查 `CI` 环境变量
- 检查标准输出是否为终端设备

**字段格式化**：

| 类型 | 格式 | 示例 |
|------|------|------|
| 字符串 | 带引号 | `"value"` |
| 数字 | 十进制 | `123` |
| 布尔 | 小写 | `true` |
| 时间 | ISO 格式 | `2006-01-02T15:04:05Z` |
| 错误 | 带引号 | `"error message"` |

### 异步启动日志

`AsyncStartupLogger` 是启动日志的核心实现，位于 `server/async_startup_logger.go`。

**核心特性**：

- 使用 channel 缓冲日志事件
- 后台 goroutine 异步刷新日志
- 支持优雅关闭，确保日志不丢失
- 线程安全，支持并发写入

**使用场景**：

- 记录应用启动各阶段的详细日志
- 统计各阶段耗时
- 提供启动过程可视化
- 支持关闭过程日志追踪

**配置项**：

| 字段 | 类型 | 默认值 | 说明 |
|------|------|--------|------|
| `enabled` | bool | true | 是否启用启动日志 |
| `async` | bool | true | 是否异步日志 |
| `buffer` | int | 100 | 日志缓冲区大小 |

### 命名日志器

为每个服务或模块创建独立的日志器，便于日志追踪和过滤：

```go
type MessageService struct {
    logger logger.ILogger
}

func (s *MessageService) initLogger(loggerMgr loggermgr.ILoggerManager) {
    s.logger = loggerMgr.Logger("MessageService")
}
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

### 日志管理器接口

框架通过 `ILoggerManager` 接口提供日志管理功能：

| 方法 | 说明 |
|------|------|
| `ManagerName()` | 返回管理器名称 |
| `Health()` | 健康检查 |
| `OnStart()` | 启动回调 |
| `OnStop()` | 停止回调 |
| `Ins()` | 获取全局日志实例 |

### 获取命名日志器

```go
// 通过管理器获取命名日志器
logger := loggerMgr.Logger("UserService")

// 使用命名日志器
logger.Info("用户登录", "user_id", 123)
```