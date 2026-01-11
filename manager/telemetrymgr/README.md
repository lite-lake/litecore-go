# telemetrymgr

可观测性管理模块，提供统一的 Traces、Metrics、Logs 三大信号管理能力。

## 特性

- **统一接口** - 提供 `TelemetryManager` 接口，统一管理可观测性组件
- **多驱动支持** - 支持 none（空实现）和 otel（OpenTelemetry）两种驱动
- **生命周期管理** - 集成 `OnStart/OnStop` 生命周期钩子，支持优雅关闭
- **灵活配置** - 支持从配置提供者或直接配置创建管理器实例
- **完整可观测性** - 支持链路追踪、指标收集和结构化日志
- **线程安全** - 所有实现都是并发安全的

## 快速开始

### 使用 none 驱动（空实现）

```go
import (
    "context"
    "com.litelake.litecore/manager/telemetrymgr"
)

func main() {
    // 创建空实现管理器
    mgr, err := telemetrymgr.Build("none", nil)
    if err != nil {
        log.Fatal(err)
    }
    defer mgr.Shutdown(context.Background())

    // 使用 tracer（no-op）
    tracer := mgr.Tracer("my-service")
    ctx, span := tracer.Start(context.Background(), "operation")
    defer span.End()
}
```

### 使用 OpenTelemetry 驱动

```go
import (
    "com.litelake.litecore/manager/telemetrymgr"
)

func main() {
    // 创建 OTel 管理器
    mgr, err := telemetrymgr.Build("otel", map[string]any{
        "endpoint": "localhost:4317",
        "insecure": true,
        "headers": map[string]any{
            "authorization": "Bearer token",
        },
        "traces": map[string]any{
            "enabled": true,
        },
    })
    if err != nil {
        log.Fatal(err)
    }
    defer mgr.Shutdown(context.Background())

    // 使用 tracer
    tracer := mgr.Tracer("my-service")
    ctx, span := tracer.Start(context.Background(), "operation")
    defer span.End()
    span.SetAttributes(attribute.String("user.id", "123"))
}
```

### 从配置提供者创建

```go
// 配置提供者会读取以下配置：
// - telemetry.driver: "otel" 或 "none"
// - telemetry.otel_config: OTel 配置对象

mgr, err := telemetrymgr.BuildWithConfigProvider(configProvider)
if err != nil {
    log.Fatal(err)
}
defer mgr.Shutdown(context.Background())
```

## 配置

### OTel 驱动配置

| 字段 | 类型 | 默认值 | 说明 |
|------|------|--------|------|
| `endpoint` | string | `localhost:4317` | OTLP 端点地址 |
| `insecure` | bool | `false` | 是否使用不安全连接（非 TLS） |
| `headers` | map[string]string | `nil` | 请求头（用于认证） |
| `resource_attributes` | []ResourceAttribute | `nil` | 资源属性 |
| `traces.enabled` | bool | `false` | 是否启用链路追踪 |
| `metrics.enabled` | bool | `false` | 是否启用指标收集 |
| `logs.enabled` | bool | `false` | 是否启用结构化日志 |

### 资源属性配置

```go
resourceAttributes := []telemetrymgr.ResourceAttribute{
    {Key: "service.name", Value: "my-service"},
    {Key: "service.version", Value: "1.0.0"},
    {Key: "deployment.environment", Value: "production"},
}
```

## 核心 API

### 创建管理器

#### Build() - 从直接配置创建

```go
func Build(driverType string, driverConfig map[string]any) (TelemetryManager, error)
```

参数：
- `driverType`: 驱动类型（"otel" 或 "none"）
- `driverConfig`: 驱动配置（根据驱动类型不同而不同）

#### BuildWithConfigProvider() - 从配置提供者创建

```go
func BuildWithConfigProvider(configProvider common.BaseConfigProvider) (TelemetryManager, error)
```

配置路径：
- `telemetry.driver`: 驱动类型
- `telemetry.otel_config`: OTel 配置

### Tracer - 链路追踪

```go
// 获取 Tracer 实例
tracer := mgr.Tracer("my-service")

// 创建 span
ctx, span := tracer.Start(ctx, "operation-name")
defer span.End()

// 设置属性
span.SetAttributes(attribute.String("key", "value"))

// 添加事件
span.AddEvent("event-name", attribute.String("details", "..."))

// 设置状态
span.SetStatus(codes.Error, "operation failed")
```

### Meter - 指标收集

```go
// 获取 Meter 实例
meter := mgr.Meter("my-service")

// 创建计数器
counter, _ := meter.Float64Counter("requests_total")
counter.Add(ctx, 1, attribute.String("path", "/api/users"))

// 创建直方图
histogram, _ := meter.Float64Histogram("request_duration")
histogram.Record(ctx, 123.45, attribute.String("status", "success"))
```

### Logger - 结构化日志

```go
// 获取 Logger 实例
logger := mgr.Logger("my-service")

// 发送日志记录
logger.Emit(ctx, log.Record{
    Timestamp:        time.Now(),
    Severity:         log.SeverityInfo,
    Body:             attribute.String("message", "operation completed"),
    Attributes:       []attribute.KeyValue{attribute.String("user.id", "123")},
})
```

## 驱动类型

### none 驱动

空实现，不产生任何可观测性数据。适用于：
- 开发测试环境
- 不需要遥测的场景
- 性能敏感的应用

所有方法都返回 no-op 实现，调用开销极小。

### otel 驱动

完整的 OpenTelemetry 实现，支持：
- 连接到 OTLP 收集器（Jaeger、Tempo、Prometheus 等）
- 发送链路追踪数据
- 发送指标数据
- 发送日志数据

支持的功能：
- 资源属性自定义
- 请求头认证
- TLS/非 TLS 连接
- 三大信号独立启用/禁用

## 生命周期管理

```go
// 启动时
if err := mgr.OnStart(); err != nil {
    log.Fatal(err)
}

// 健康检查
if err := mgr.Health(); err != nil {
    log.Printf("manager unhealthy: %v", err)
}

// 停止时（会自动调用 Shutdown）
if err := mgr.OnStop(); err != nil {
    log.Printf("manager stop failed: %v", err)
}

// 或手动关闭
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()
if err := mgr.Shutdown(ctx); err != nil {
    log.Printf("manager shutdown failed: %v", err)
}
```

## 错误处理

### 创建失败

```go
mgr, err := telemetrymgr.Build("otel", map[string]any{
    "endpoint": "", // 空端点会导致验证失败
})
if err != nil {
    // 错误会包含详细的失败原因
    log.Printf("failed to create manager: %v", err)
    return
}
```

### 关闭失败

```go
// 关闭失败不会影响程序运行，但可能导致数据丢失
if err := mgr.Shutdown(ctx); err != nil {
    log.Printf("warning: manager shutdown incomplete: %v", err)
}
```

## 最佳实践

### 1. 使用 defer 确保资源释放

```go
mgr, err := telemetrymgr.Build("otel", config)
if err != nil {
    return err
}
defer mgr.Shutdown(context.Background())
```

### 2. 设置合理的关闭超时

```go
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()
mgr.Shutdown(ctx)
```

### 3. 为不同服务使用不同的名称

```go
userServiceTracer := mgr.Tracer("user-service")
orderServiceTracer := mgr.Tracer("order-service")
```

### 4. 在开发环境使用 none 驱动

```go
driver := "none"
if os.Getenv("ENV") == "production" {
    driver = "otel"
}

mgr, err := telemetrymgr.Build(driver, config)
```

### 5. 资源属性包含服务信息

```go
config := map[string]any{
    "resource_attributes": []telemetrymgr.ResourceAttribute{
        {Key: "service.name", Value: "my-service"},
        {Key: "service.version", Value: version},
        {Key: "deployment.environment", Value: environment},
    },
}
```

## 性能考虑

- **none 驱动**：开销极小，几乎可以忽略
- **otel 驱动未启用特性时**：使用 no-op provider，开销很小
- **otel 驱动启用特性后**：会建立网络连接，使用批处理减少开销
- **并发安全**：所有方法都可以安全地并发调用

## 线程安全

`TelemetryManager` 接口的所有实现都是并发安全的，可以在多个 goroutine 中同时使用。

```go
// 多个 goroutine 可以安全地使用同一个管理器
for i := 0; i < 10; i++ {
    go func() {
        tracer := mgr.Tracer("worker")
        ctx, span := tracer.Start(context.Background(), "task")
        defer span.End()
        // ... 执行任务
    }()
}
```
