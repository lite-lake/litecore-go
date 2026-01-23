# 日志格式美化改造技术需求文档

| 文档版本 | 日期 | 作者 |
|---------|------|------|
| 1.0 | 2026-01-24 | opencode |

## 1. 背景

### 1.1 现状问题

当前日志格式存在以下问题：

1. **格式不统一**
   - Builtin 初始化阶段使用标准库日志：`2026/01/24 02:17:42 [Builtin] INFO: 开始初始化内置组件 []`
   - Zap 阶段使用 ConsoleEncoder：`2026-01-24 02:17:42.750 | INFO | 切换到结构化日志系统 | {"logger": "zap"}`

2. **JSON 字段可读性差**
   - 字段堆砌在末尾：`{"logger": "zap", "[count 8]": ["duration","48.083µs"]}`
   - 键值对格式不规范：`[middleware RecoveryMiddleware]: [type 全局]`

3. **颜色利用率低**
   - 代码已实现颜色检测和编码器，但视觉效果不明显
   - 缺少图标、状态码颜色等视觉元素

4. **缺少视觉层次**
   - 没有分组、缩进等结构化展示
   - 启动过程缺少进度感

### 1.2 改造目标

- 统一日志格式，参考 Gin 框架的竖线分隔符风格
- 提升控制台可读性，支持彩色输出
- 保持文件日志的结构化（JSON 格式）
- 向后兼容，不破坏现有接口

## 2. 技术方案

### 2.1 整体架构

```
┌─────────────────────────────────────────────────────────┐
│                    Logger Manager                        │
├─────────────────────────────────────────────────────────┤
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐    │
│  │  Console    │  │    File     │  │  Telemetry  │    │
│  │  Encoder    │  │  Encoder    │  │   Encoder   │    │
│  └─────────────┘  └─────────────┘  └─────────────┘    │
│         │               │               │              │
│         ▼               ▼               ▼              │
│  ┌─────────────────────────────────────────────┐       │
│  │         zapcore.NewTee(cores...)           │       │
│  └─────────────────────────────────────────────┘       │
└─────────────────────────────────────────────────────────┘
```

### 2.2 日志格式定义

#### 2.2.1 控制台日志（Gin 风格）

**通用格式**：
```
{时间} | {级别} | {消息} | {字段1}={值1} {字段2}={值2} ...
```

**示例**：
```
2026-01-24 02:17:42.750 | INFO | 开始依赖注入 | count=23
2026-01-24 02:17:42.750 | INFO | 注册路由 | method=GET path=/api/messages
2026-01-24 02:17:42.750 | ERROR | HTTP server error | error="listen tcp 0.0.0.0:8080: bind: address already in use"
```

**字段格式**：
- 时间：固定宽度 `2006-01-24 15:04:05.000`（23字符）
- 级别：固定宽度 5 字符，右对齐，带颜色
- 消息：主描述信息
- 字段：`key=value` 格式，字符串值用引号包裹

#### 2.2.2 请求日志（兼容 Gin 格式）

**格式**：
```
{时间} | {状态码} | {耗时} | {客户端IP} | {方法} | {路径}
```

**示例**：
```
2026-01-24 02:17:43.123 | 200 | 1.234ms | 127.0.0.1 | GET | /api/messages
2026-01-24 02:17:43.456 | 404 | 0.456ms | 192.168.1.1 | POST | /api/unknown
```

**状态码颜色**：
- 2xx：绿色
- 3xx：黄色
- 4xx：橙色
- 5xx：红色

#### 2.2.3 文件日志（JSON 格式）

保持现有 JSON 格式，用于日志分析和监控：

```json
{
  "time": "2026-01-24T02:17:42.750Z",
  "level": "info",
  "msg": "开始依赖注入",
  "count": 23,
  "duration": "48.083µs"
}
```

### 2.3 颜色方案

#### 2.3.1 日志级别颜色

| 级别 | ANSI 颜色 | 代码 | 预览 |
|------|-----------|------|------|
| DEBUG | 灰色 | `\033[90m` | DEBUG |
| INFO | 绿色 | `\033[32m` | INFO |
| WARN | 黄色 | `\033[33m` | WARN |
| ERROR | 红色 | `\033[31m` | ERROR |
| FATAL | 红色+粗体 | `\033[1;31m` | FATAL |

#### 2.3.2 HTTP 状态码颜色

| 状态码范围 | 颜色 | 示例 |
|-----------|------|------|
| 2xx | 绿色 | 200 201 |
| 3xx | 黄色 | 301 302 |
| 4xx | 橙色 | 400 404 |
| 5xx | 红色 | 500 502 |

#### 2.3.3 其他元素

- 竖线分隔符：灰色 `\033[90m|\033[0m`
- 消息：无颜色
- 字段值（字符串）：白色

### 2.4 配置扩展

#### 2.4.1 ConsoleConfig 新增字段

```go
type LogLevelConfig struct {
    Level      string `yaml:"level"`       // 日志级别: debug, info, warn, error, fatal
    Format     string `yaml:"format"`      // 格式: gin | json | default
    Color      bool   `yaml:"color"`       // 是否启用颜色
    TimeFormat string `yaml:"time_format"` // 时间格式（默认：2006-01-24 15:04:05.000）
}
```

#### 2.4.2 配置示例

```yaml
logger:
  driver: "zap"
  zap_config:
    console_enabled: true
    console_config:
      level: "info"
      format: "gin"        # gin | json | default
      color: true          # 是否启用颜色
      time_format: "2006-01-24 15:04:05.000"
    file_enabled: true
    file_config:
      level: "info"
      path: "./logs/app.log"
```

## 3. 实现细节

### 3.1 新增 GinConsoleEncoder

**文件位置**：`manager/loggermgr/encoder_gin.go`

**核心逻辑**：
1. 实现 `zapcore.Encoder` 接口
2. 解析 Entry（时间、级别、消息、调用信息）
3. 格式化 Fields 为 `key=value`
4. 添加颜色 ANSI 代码
5. 使用固定宽度对齐

**伪代码**：
```go
func (e *ginConsoleEncoder) Clone() zapcore.Encoder {
    return &ginConsoleEncoder{
        EncoderConfig: e.EncoderConfig,
        format: e.format,
        color: e.color,
    }
}

func (e *ginConsoleEncoder) EncodeEntry(entry zapcore.Entry, fields []zapcore.Field) (*buffer.Buffer, error) {
    buf := buffer.Get()

    // 1. 时间（固定 23 字符）
    buf.WriteString(entry.Time.Format(e.TimeFormat))

    // 2. 分隔符
    buf.WriteString(" | ")

    // 3. 级别（固定 5 字符，颜色）
    if e.color {
        buf.WriteString(e.levelColor(entry.Level))
    }
    buf.WriteString(fmt.Sprintf("%-5s", levelString(entry.Level)))
    if e.color {
        buf.WriteString(colorReset)
    }

    // 4. 消息
    buf.WriteString(" | ")
    buf.WriteString(entry.Message)

    // 5. 字段（key=value 格式）
    for _, field := range fields {
        buf.WriteString(fmt.Sprintf(" %s=%s", field.Key, formatValue(field)))
    }

    buf.WriteString("\n")
    return buf, nil
}
```

### 3.2 修改 buildConsoleCore

**文件位置**：`manager/loggermgr/driver_zap_impl.go`

**变更**：
```go
func buildConsoleCore(cfg *LogLevelConfig) (zapcore.Core, error) {
    level := parseLogLevel(cfg.Level)

    var encoder zapcore.Encoder

    switch cfg.Format {
    case "gin":
        encoder = NewGinConsoleEncoder(zapcore.EncoderConfig{
            TimeKey:          "time",
            LevelKey:         "level",
            NameKey:          "logger",
            MessageKey:       "msg",
            LineEnding:       zapcore.DefaultLineEnding,
            EncodeTime:       customTimeEncoder,
            EncodeLevel:      zapcore.CapitalLevelEncoder,
        }, cfg.Color)
    case "json":
        encoder = zapcore.NewJSONEncoder(zapcore.EncoderConfig{...})
    default:
        encoder = zapcore.NewConsoleEncoder(zapcore.EncoderConfig{...})
    }

    stdoutCore := zapcore.NewCore(encoder, zapcore.AddSync(os.Stdout), level)
    stderrCore := zapcore.NewCore(encoder, zapcore.AddSync(os.Stderr), zapcore.ErrorLevel)

    return zapcore.NewTee(stdoutCore, stderrCore), nil
}
```

### 3.3 颜色检测优化

复用现有 `detectColorSupport()` 逻辑，添加配置覆盖：

```go
func detectColorSupport(configColor bool) bool {
    if !configColor {
        return false
    }
    // 现有检测逻辑...
}
```

### 3.4 默认值处理

- `Format` 默认值：`"gin"`
- `Color` 默认值：`true`（由 `detectColorSupport()` 动态决定）
- `TimeFormat` 默认值：`"2006-01-24 15:04:05.000"`

## 4. 兼容性处理

### 4.1 向后兼容

- 默认格式改为 `gin`，提供 `"default"` 选项恢复旧格式
- 配置文件中未指定 `format` 字段时，默认使用 `gin`
- 现有日志输出代码无需修改

### 4.2 测试策略

1. **单元测试**
   - 测试各日志级别的输出格式
   - 测试字段格式化（字符串、数字、布尔）
   - 测试颜色编码

2. **集成测试**
   - 测试控制台输出与文件输出的一致性
   - 测试多核心（Console + File + Telemetry）

3. **视觉测试**
   - 手动验证颜色显示
   - 验证对齐效果

## 5. 验收标准

### 5.1 功能验收

- [ ] 控制台日志符合 Gin 风格格式
- [ ] 日志级别显示正确颜色
- [ ] HTTP 状态码显示对应颜色
- [ ] 字段格式为 `key=value`
- [ ] 时间固定 23 字符宽度
- [ ] 文件日志保持 JSON 格式

### 5.2 配置验收

- [ ] 支持配置文件切换格式（gin/json/default）
- [ ] 支持配置颜色开关
- [ ] 支持自定义时间格式

### 5.3 兼容性验收

- [ ] 现有日志调用代码无需修改
- [ ] 不破坏 Telemetry 输出
- [ ] 不影响文件日志轮转

## 6. 实施计划

| 阶段 | 任务 | 产出 |
|------|------|------|
| 1 | 新增 `encoder_gin.go` | GinConsoleEncoder 实现 |
| 2 | 修改 `driver_zap_impl.go` | 集成新编码器 |
| 3 | 扩展 `config.go` | 新增配置字段 |
| 4 | 编写单元测试 | 测试覆盖率 > 80% |
| 5 | 更新配置文件 | messageboard/config.yaml 示例 |
| 6 | 集成测试 | 验证整体输出效果 |
| 7 | 文档更新 | AGENTS.md 更新日志使用规范 |

## 7. 附录

### 7.1 参考实现

- Gin 框架日志格式：https://github.com/gin-gonic/gin/blob/master/logger.go
- Uber Zap 编码器：https://github.com/uber-go/zap/blob/master/zapcore/encoder.go

### 7.2 ANSI 颜色代码速查

```
Reset:   \033[0m
Red:     \033[31m
Green:   \033[32m
Yellow:  \033[33m
Blue:    \033[34m
Magenta: \033[35m
Cyan:    \033[36m
White:   \033[37m
Gray:    \033[90m
Bold:    \033[1m
```
