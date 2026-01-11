# Telemetry Manager

提供可观测性管理功能，支持分布式链路追踪、指标收集和日志关联。

## 特性

- **OpenTelemetry 集成** - 基于业界标准的 OpenTelemetry 协议，支持多种后端
- **依赖注入支持** - 通过 Container 自动注入配置和依赖
- **驱动抽象设计** - 统一接口支持多种实现，当前支持 OTEL 和 None 驱动
- **优雅降级机制** - 配置解析失败或初始化错误时自动降级到空实现
- **生命周期管理** - 完整的启动、健康检查和优雅关闭支持
- **线程安全** - 所有操作均保证并发安全

## 快速开始

```go
package main

import (
    "context"
    "fmt"

    "com.litelake.litecore/manager/telemetrymgr"
)

func main() {
    // 创建观测管理器
    mgr := telemetrymgr.NewManager("default")

    // 通过依赖注入容器初始化（推荐）
    // container.Register("config", configProvider)
    // container.Register("telemetry.default", mgr)
    // container.InjectAll()

    // 启动管理器（会从配置中读取配置）
    if err := mgr.OnStart(); err != nil {
        panic(err)
    }
    defer mgr.OnStop()

    fmt.Println("Telemetry manager started:", mgr.ManagerName())

    // 使用 Tracer 进行链路追踪
    tracer := mgr.Tracer("my-service")
    ctx, span := tracer.Start(context.Background(), "operation")
    defer span.End()

    // 在这里执行业务逻辑
    _ = ctx
}
```

## 创建管理器

### 使用依赖注入（推荐）

```go
// 创建管理器
mgr := telemetrymgr.NewManager("default")

// 注册到容器
container.Register("config", configProvider)
container.Register("telemetry.default", mgr)

// 注入依赖并启动
container.InjectAll()
mgr.OnStart()
defer mgr.OnStop()
```

### 配置 OTEL 驱动

在配置文件中添加观测配置：

```yaml
telemetry.default:
  driver: otel
  otel_config:
    endpoint: http://localhost:4317
    insecure: false  # 默认使用 TLS
    resource_attributes:
      - key: service.name
        value: my-service
      - key: environment
        value: production
    traces:
      enabled: true
    metrics:
      enabled: true
    logs:
      enabled: true
```

### 使用 None 驱动

```go
// 创建管理器
mgr := telemetrymgr.NewManager("default")

// 不提供配置或配置 driver 为 none
// 使用默认配置（禁用观测）
mgr.OnStart()
defer mgr.OnStop()
```

## 使用 Tracer

Manager 实现了 TelemetryManager 接口，提供了获取 Tracer 实例的方法：

```go
tracer := mgr.Tracer("my-service")

// 创建 Span
ctx, span := tracer.Start(context.Background(), "operation-name")
defer span.End()

// 添加属性
span.SetAttributes(
    attribute.String("user.id", "123"),
    attribute.String("action", "login"),
)

// 添加事件
span.AddEvent("user authenticated")

// 记录错误
if err != nil {
    span.RecordError(err)
    span.SetStatus(codes.Error, "operation failed")
}
```

## 获取 Provider

某些场景可能需要直接访问 TracerProvider：

```go
// 获取 TracerProvider
tp := mgr.TracerProvider()
if tp != nil {
    // 直接使用 TracerProvider
    _ = tp
}
```

## 优雅关闭

```go
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()

if err := mgr.Shutdown(ctx); err != nil {
    log.Printf("Failed to shutdown telemetry manager: %v", err)
}
```

## API

### 接口

#### TelemetryManager

```go
type TelemetryManager interface {
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

    // Shutdown 关闭观测管理器，刷新所有待处理的数据
    Shutdown(ctx context.Context) error
}
```

### 构造函数

#### NewManager

```go
func NewManager(name string) *Manager
```

创建观测管理器实例。

- `name`: 管理器名称，用于配置键前缀（如 "default" → "telemetry.default"）

## 配置说明

### 依赖注入配置

管理器通过依赖注入自动从配置中读取配置，配置键格式为 `telemetry.{manager_name}`：

```yaml
telemetry.default:
  driver: otel
  otel_config:
    endpoint: http://localhost:4317
    insecure: false
    resource_attributes:
      - key: service.name
        value: my-service
    traces:
      enabled: true
    metrics:
      enabled: true
    logs:
      enabled: true
```

### OTEL 配置结构

| 字段 | 类型 | 必需 | 说明 |
|------|------|------|------|
| endpoint | string | 是 | OTLP collector 端点地址 |
| insecure | bool | 否 | 是否使用不安全连接，默认 false（使用 TLS） |
| resource_attributes | []ResourceAttribute | 否 | 资源属性列表 |
| headers | map[string]string | 否 | 认证请求头 |
| traces | FeatureConfig | 否 | 链路追踪配置 |
| metrics | FeatureConfig | 否 | 指标配置 |
| logs | FeatureConfig | 否 | 日志配置 |

### ResourceAttribute 结构

```go
type ResourceAttribute struct {
    Key   string
    Value string
}
```

### FeatureConfig 结构

```go
type FeatureConfig struct {
    Enabled bool
}
```

## 错误处理

管理器采用优雅降级策略：

1. 配置解析失败 → 降级到 None 驱动
2. 配置验证失败 → 降级到 None 驱动
3. OTEL 初始化失败 → 降级到 None 驱动
4. 未知驱动类型 → 降级到 None 驱动

```go
// OnStart 会返回错误，但不会终止程序
if err := mgr.OnStart(); err != nil {
    log.Printf("Telemetry manager initialization failed, using none driver: %v", err)
}
```

## 线程安全

所有管理器实现均保证并发安全：
- Manager 使用 `sync.RWMutex` 保护内部状态
- Shutdown 使用 `sync.Once` 确保只执行一次

## 健康检查

Manager 接口提供 `Health()` 方法用于健康检查：

```go
if err := mgr.Health(); err != nil {
    log.Printf("Telemetry manager unhealthy: %v", err)
}
```
