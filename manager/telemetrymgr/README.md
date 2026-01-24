# telemetrymgr

可观测性管理模块，提供统一的 Traces、Metrics、Logs 三大信号管理能力。

## 特性

- **统一接口** - 提供 `ITelemetryManager` 接口，统一管理可观测性组件
- **多驱动支持** - 支持 none（空实现）和 otel（OpenTelemetry）两种驱动
- **生命周期管理** - 集成 `OnStart/OnStop/Shutdown` 生命周期钩子，支持优雅关闭
- **灵活配置** - 支持从配置提供者或直接配置创建管理器实例
- **完整可观测性** - 支持链路追踪、指标收集和结构化日志
- **OpenTelemetry 集成** - 原生支持 OpenTelemetry 协议，可连接到 OTLP 兼容的收集器

## 快速开始

### 使用 none 驱动（空实现）

```go
import (
    "context"
    "github.com/lite-lake/litecore-go/manager/telemetrymgr"
)

func main() {
    mgr, err := telemetrymgr.Build("none", nil)
    if err != nil {
        log.Fatal(err)
    }
    defer mgr.Shutdown(context.Background())

    tracer := mgr.Tracer("my-service")
    ctx, span := tracer.Start(context.Background(), "operation")
    defer span.End()
}
```

### 使用 OpenTelemetry 驱动

```go
import (
    "go.opentelemetry.io/otel/attribute"
    "github.com/lite-lake/litecore-go/manager/telemetrymgr"
)

func main() {
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

    tracer := mgr.Tracer("my-service")
    ctx, span := tracer.Start(context.Background(), "operation")
    defer span.End()
    span.SetAttributes(attribute.String("user.id", "123"))
}
```

### 从配置提供者创建

```go
mgr, err := telemetrymgr.BuildWithConfigProvider(configProvider)
if err != nil {
    log.Fatal(err)
}
defer mgr.Shutdown(context.Background())
```

配置路径：
- `telemetry.driver`: 驱动类型（"otel" 或 "none"）
- `telemetry.otel_config`: OTel 配置对象

### 配置示例

在 `config.yaml` 中配置 telemetry：

```yaml
telemetry:
  driver: "otel"
  otel_config:
    endpoint: "localhost:4317"
    insecure: false
    headers:
      Authorization: "Bearer your-token"
    resource_attributes:
      - key: "service.name"
        value: "my-service"
      - key: "service.version"
        value: "1.0.0"
    traces:
      enabled: true
    metrics:
      enabled: false
    logs:
      enabled: false
```

## 支持的遥测驱动

### none 驱动

空实现，不产生任何可观测性数据。适用于开发测试环境或不需要遥测的场景。所有方法都返回 no-op 实现，调用开销极小。

**特点：**
- 零开销
- 无网络连接
- 适用于本地开发或测试环境

### otel 驱动

完整的 OpenTelemetry 实现，支持连接到 OTLP 兼容的收集器（如 Jaeger、Tempo、Prometheus 等）。

**支持的信号：**
- **Traces** - 链路追踪（完整支持）
- **Metrics** - 指标收集（预留接口，当前使用 noop provider）
- **Logs** - 结构化日志（预留接口，当前使用 noop provider）

**特点：**
- 连接到 OTLP 收集器（gRPC 协议）
- 支持资源属性自定义
- 支持请求头认证
- 支持 TLS/非 TLS 连接
- 三大信号独立启用/禁用
- 批处理导出，减少网络开销

## OpenTelemetry 集成

### OTLP 端点配置

otel 驱动通过 OTLP（OpenTelemetry Protocol）将数据发送到收集器。默认端点为 `localhost:4317`（gRPC 协议）。

**常用收集器端点：**

| 收集器 | 端点 |
|--------|------|
| Jaeger | `localhost:4317` |
| Grafana Tempo | `localhost:4317` |
| Grafana Agent | `localhost:4317` |
| OpenTelemetry Collector | `localhost:4317` |

### 资源属性

资源属性用于标识产生遥测数据的服务。建议配置以下属性：

```go
config := map[string]any{
    "resource_attributes": []telemetrymgr.ResourceAttribute{
        {Key: "service.name", Value: "my-service"},
        {Key: "service.version", Value: "1.0.0"},
        {Key: "deployment.environment", Value: "production"},
        {Key: "host.name", Value: "server-01"},
    },
}
```

**语义约定（推荐）：**
- `service.name` - 服务名称（必填）
- `service.version` - 服务版本
- `deployment.environment` - 部署环境（development/staging/production）
- `host.name` - 主机名

### 请求头认证

如果 OTLP 端点需要认证，可以通过 `headers` 配置：

```go
config := map[string]any{
    "headers": map[string]any{
        "Authorization": "Bearer my-api-key",
        "X-Custom-Header": "custom-value",
    },
}
```

### TLS/非 TLS 连接

默认使用 TLS 安全连接。对于开发环境或内网部署，可以禁用 TLS：

```go
config := map[string]any{
    "endpoint": "localhost:4317",
    "insecure": true,  // 禁用 TLS，使用非安全连接
}
```

**注意：** 生产环境建议使用 TLS 加密连接。

### 独立启用/禁用信号

三大信号可以独立启用或禁用：

```go
config := map[string]any{
    "traces": map[string]any{
        "enabled": true,   // 启用链路追踪
    },
    "metrics": map[string]any{
        "enabled": false,  // 禁用指标收集
    },
    "logs": map[string]any{
        "enabled": false,  // 禁用日志观测
    },
}
```

## 配置说明

### OTel 驱动配置

| 字段 | 类型 | 默认值 | 说明 |
|------|------|--------|------|
| `endpoint` | string | `localhost:4317` | OTLP 端点地址（gRPC 协议） |
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

### 配置验证

`Build` 和 `BuildWithConfigProvider` 会自动验证配置。如果配置无效，会返回错误。

**验证规则：**
- `driver` 必须是 "otel" 或 "none"
- 当 `driver` 为 "otel" 时，必须提供 `otel_config`
- `otel_config.endpoint` 不能为空

## API 说明

### 创建管理器

#### Build() - 从直接配置创建

```go
func Build(driverType string, driverConfig map[string]any) (ITelemetryManager, error)
```

**参数：**
- `driverType`: 驱动类型（"otel" 或 "none"）
- `driverConfig`: 驱动配置（根据驱动类型不同而不同）

**返回：**
- `ITelemetryManager`: 遥测管理器接口实例
- `error`: 创建失败的错误

#### BuildWithConfigProvider() - 从配置提供者创建

```go
func BuildWithConfigProvider(configProvider configmgr.IConfigManager) (ITelemetryManager, error)
```

**参数：**
- `configProvider`: 配置管理器接口实例

**配置路径：**
- `telemetry.driver`: 驱动类型
- `telemetry.otel_config`: OTel 配置（当 driver=otel 时）

### Tracer - 链路追踪

```go
tracer := mgr.Tracer("my-service")
ctx, span := tracer.Start(ctx, "operation-name")
defer span.End()

// 设置属性
span.SetAttributes(
    attribute.String("user.id", "123"),
    attribute.Int("retry.count", 3),
)

// 添加事件
span.AddEvent("cache.miss", attribute.String("key", "user:123"))

// 设置状态
span.SetStatus(codes.Error, "operation failed")
```

### Meter - 指标收集

```go
meter := mgr.Meter("my-service")

// Counter 计数器
counter, _ := meter.Float64Counter("requests_total")
counter.Add(ctx, 1, attribute.String("path", "/api/users"))

// Histogram 直方图
histogram, _ := meter.Float64Histogram("request_duration")
histogram.Record(ctx, 123.45, attribute.String("status", "success"))

// UpUpDownCounter 可上下计数
upDownCounter, _ := meter.Int64UpDownCounter("active_connections")
upDownCounter.Add(ctx, 1, attribute.String("type", "websocket"))
```

**注意：** 当前 Metrics 使用 noop provider，不会发送数据到 OTLP 端点。此功能为预留接口。

### Logger - 结构化日志

```go
logger := mgr.Logger("my-service")
logger.Emit(ctx, log.Record{
    Timestamp:        time.Now(),
    Severity:         log.SeverityInfo,
    Body:             attribute.String("message", "operation completed"),
    Attributes:       []attribute.KeyValue{
        attribute.String("user.id", "123"),
        attribute.String("action", "login"),
    },
})
```

**注意：** 当前 Logs 使用 noop provider，不会发送数据到 OTLP 端点。此功能为预留接口。

### 生命周期管理

```go
// OnStart - 服务器启动时调用
if err := mgr.OnStart(); err != nil {
    log.Fatal(err)
}

// Health - 检查管理器健康状态
if err := mgr.Health(); err != nil {
    log.Printf("manager unhealthy: %v", err)
}

// OnStop - 服务器停止时调用（会自动触发 Shutdown）
if err := mgr.OnStop(); err != nil {
    log.Printf("manager stop failed: %v", err)
}

// Shutdown - 优雅关闭，刷新所有待处理的数据
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()
if err := mgr.Shutdown(ctx); err != nil {
    log.Printf("manager shutdown failed: %v", err)
}
```

## 与中间件集成

### 使用 TelemetryMiddleware

`litemiddleware` 提供了 `TelemetryMiddleware`，可以自动为 HTTP 请求创建链路追踪 span。

```go
import "github.com/lite-lake/litecore-go/component/litemiddleware"

// 创建中间件
telemetryMiddleware := litemiddleware.NewTelemetryMiddleware(nil)

// 注册到容器
middlewareContainer.RegisterMiddleware(telemetryMiddleware)
```

**TelemetryMiddleware 会：**
- 自动为每个 HTTP 请求创建 span
- 记录请求方法、路径、状态码等属性
- 将 span 上下文传播到请求上下文

### 在 Service 层使用

```go
type MessageService struct {
    TelemetryMgr telemetrymgr.ITelemetryManager `inject:""`
    tracer       trace.Tracer
}

func (s *MessageService) initTracer() {
    if s.TelemetryMgr != nil {
        s.tracer = s.TelemetryMgr.Tracer("message-service")
    }
}

func (s *MessageService) CreateMessage(ctx context.Context, msg *Message) error {
    s.initTracer()
    
    ctx, span := s.tracer.Start(ctx, "CreateMessage")
    defer span.End()
    
    // 业务逻辑
    span.SetAttributes(attribute.String("message.id", msg.ID))
    
    return nil
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

### 6. 传播上下文

在分布式系统中，确保 span 上下文在服务间传播：

```go
// HTTP 请求携带 trace context
req, _ := http.NewRequest("GET", "http://api.example.com/data", nil)
span := trace.SpanFromContext(ctx)
trace.ContextWithSpan(req.Context(), span)
```

### 7. 合理设置采样率

在高流量场景下，可以通过采样率减少数据量：

```go
// TODO: 当前版本未暴露采样率配置
// 未来版本将支持采样率配置
```

## 性能考虑

- **none 驱动**：开销极小，几乎可以忽略
- **otel 驱动未启用特性时**：使用 no-op provider，开销很小
- **otel 驱动启用特性后**：
  - 建立网络连接到 OTLP 端点
  - 使用批处理导出，减少网络开销
  - Traces 使用批量导出，默认批处理大小和超时可优化
- **并发安全**：所有方法都可以安全地并发调用

## 线程安全

`ITelemetryManager` 接口的所有实现都是并发安全的，可以在多个 goroutine 中同时使用。

```go
for i := 0; i < 10; i++ {
    go func() {
        tracer := mgr.Tracer("worker")
        ctx, span := tracer.Start(context.Background(), "task")
        defer span.End()
    }()
}
```

## 常见问题

### Q: 为什么 Metrics 和 Logs 不发送数据？

A: 当前版本的 Metrics 和 Logs 使用 noop provider，不会将数据发送到 OTLP 端点。此功能为预留接口，未来版本将支持完整的 Metrics 和 Logs 导出。

### Q: 如何验证 OTel 集成是否正常工作？

A: 可以通过以下方式验证：
1. 检查 `mgr.Health()` 是否返回错误
2. 在 Jaeger 或其他收集器中查看是否有 span 数据
3. 查看管理器日志，确认 exporter 初始化成功

### Q: OTel 端点连接失败会怎样？

A: 如果 OTel 端点连接失败，`NewTelemetryManagerOtelImpl` 会返回错误。建议在创建管理器时检查错误。

### Q: 如何在生产环境使用 TLS 连接？

A: 在配置中不设置 `insecure` 或设置为 `false`：

```go
config := map[string]any{
    "endpoint": "otel-collector.example.com:4317",
    "insecure": false,  // 使用 TLS
}
```

### Q: 是否支持多个 OTel 端点？

A: 当前版本不支持多个端点。如果需要发送到多个端点，需要使用 OpenTelemetry Collector 作为中间聚合层。

## 相关文档

- [OpenTelemetry 官方文档](https://opentelemetry.io/docs/)
- [OTLP 协议规范](https://opentelemetry.io/docs/reference/specification/protocol/otlp/)
- [Jaeger 文档](https://www.jaegertracing.io/docs/)
- [Grafana Tempo 文档](https://grafana.com/docs/tempo/latest/)
- [litemiddleware README](../../component/litemiddleware/README.md)
