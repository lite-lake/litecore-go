# Telemetry Manager

提供可观测性管理功能，支持分布式链路追踪、指标收集和日志关联。

## 特性

- **OpenTelemetry 集成** - 基于业界标准的 OpenTelemetry 协议，支持多种后端
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
    // 创建工厂
    factory := telemetrymgr.NewFactory()

    // 配置 OTEL 驱动
    config := map[string]any{
        "endpoint": "http://localhost:4317",
        "resource_attributes": []any{
            map[string]any{"key": "service.name", "value": "my-service"},
            map[string]any{"key": "environment", "value": "production"},
        },
        "traces":  map[string]any{"enabled": true},
        "metrics": map[string]any{"enabled": true},
        "logs":    map[string]any{"enabled": true},
    }

    // 创建并启动管理器
    mgr := factory.Build("otel", config)
    if err := mgr.OnStart(); err != nil {
        panic(err)
    }
    defer mgr.OnStop()

    fmt.Println("Telemetry manager started:", mgr.ManagerName())

    // 使用 Tracer 进行链路追踪
    if telemetryMgr, ok := mgr.(telemetrymgr.TelemetryManager); ok {
        tracer := telemetryMgr.Tracer("my-service")
        ctx, span := tracer.Start(context.Background(), "operation")
        defer span.End()

        // 在这里执行业务逻辑
        _ = ctx
    }
}
```

## 创建管理器

### 使用 OTEL 驱动

```go
factory := telemetrymgr.NewFactory()

config := map[string]any{
    "endpoint": "http://localhost:4317",
    "insecure": false, // 默认使用 TLS
    "traces":   map[string]any{"enabled": true},
}

mgr := factory.Build("otel", config)
mgr.OnStart()
defer mgr.OnStop()
```

### 使用 None 驱动

```go
factory := telemetrymgr.NewFactory()

// none 驱动不需要配置
mgr := factory.Build("none", nil)

mgr.OnStart()
defer mgr.OnStop()
```

### 使用配置结构体

```go
factory := telemetrymgr.NewFactory()

telemetryConfig := &config.TelemetryConfig{
    Driver: "otel",
    OtelConfig: &config.OtelConfig{
        Endpoint: "http://localhost:4317",
        Traces:   &config.FeatureConfig{Enabled: true},
        Metrics:  &config.FeatureConfig{Enabled: true},
    },
}

mgr, err := factory.BuildWithConfig(telemetryConfig)
if err != nil {
    panic(err)
}
```

## 使用 Tracer

TelemetryManager 接口提供了获取 Tracer 实例的方法：

```go
if telemetryMgr, ok := mgr.(telemetrymgr.TelemetryManager); ok {
    tracer := telemetryMgr.Tracer("my-service")

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
}
```

## 获取 Provider

某些场景可能需要直接访问 TracerProvider：

```go
// 获取 TracerProvider
tp := telemetrymgr.TracerProvider(mgr)
if tp != nil {
    // 直接使用 TracerProvider
    _ = tp
}
```

## 优雅关闭

```go
if telemetryMgr, ok := mgr.(telemetrymgr.TelemetryManager); ok {
    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()

    if err := telemetryMgr.Shutdown(ctx); err != nil {
        log.Printf("Failed to shutdown telemetry manager: %v", err)
    }
}
```

## API

### 接口

#### TelemetryManager

```go
type TelemetryManager interface {
    // Tracer 获取 Tracer 实例
    Tracer(name string) trace.Tracer

    // Shutdown 关闭观测管理器，刷新所有待处理的数据
    Shutdown(ctx context.Context) error
}
```

#### TracerProviderGetter

```go
type TracerProviderGetter interface {
    TracerProvider() *sdktrace.TracerProvider
}
```

### 工厂方法

#### NewFactory

```go
func NewFactory() *Factory
```

创建观测管理器工厂。

#### Build

```go
func (f *Factory) Build(driver string, cfg map[string]any) common.Manager
```

创建观测管理器实例。

- `driver`: 驱动类型，支持 "none", "otel"
- `cfg`: 驱动专属的配置数据

#### BuildWithConfig

```go
func (f *Factory) BuildWithConfig(telemetryConfig *config.TelemetryConfig) (common.Manager, error)
```

使用配置结构体创建观测管理器。

### 辅助函数

#### TracerProvider

```go
func TracerProvider(mgr any) *sdktrace.TracerProvider
```

获取 TracerProvider，用于需要直接使用 TracerProvider 的场景。

## 配置说明

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

`Build` 方法采用优雅降级策略：

1. 配置解析失败 → 降级到 None 驱动
2. 配置验证失败 → 降级到 None 驱动
3. OTEL 初始化失败 → 降级到 None 驱动
4. 未知驱动类型 → 降级到 None 驱动

`BuildWithConfig` 方法在错误时返回 error，需要调用方处理：

```go
mgr, err := factory.BuildWithConfig(telemetryConfig)
if err != nil {
    return fmt.Errorf("failed to create telemetry manager: %w", err)
}
```

## 线程安全

所有管理器实现均保证并发安全：
- OtelManager 使用 `sync.RWMutex` 保护内部状态
- NoneManager 是无状态的，天然并发安全
- Shutdown 使用 `sync.Once` 确保只执行一次

## 健康检查

Manager 接口提供 `Health()` 方法用于健康检查：

```go
if err := mgr.Health(); err != nil {
    log.Printf("Telemetry manager unhealthy: %v", err)
}
```
