# CacheManager 可观测性改造方案

**文档编号**: TRD-20260111
**创建日期**: 2025-01-11
**负责人**: kentzhu
**状态**: 设计阶段

## 1. 背景

当前 `cachemgr` 仅提供基础的缓存操作功能，缺乏日志记录和可观测性支持。为了更好地监控缓存操作、排查问题和分析性能，需要接入 `loggermgr` 和 `telemetrymgr`。

## 2. 现状分析

### 2.1 cachemgr 当前架构

```
manager/cachemgr/
├── interface.go           # CacheManager 接口定义
├── factory.go             # 工厂方法创建管理器
├── cache_adapter.go       # 适配器，将内部驱动适配到接口
└── internal/drivers/
    ├── base_manager.go    # 基础管理器
    ├── redis_manager.go   # Redis 驱动实现
    ├── memory_manager.go  # 内存驱动实现
    └── none_manager.go    # 空驱动实现
```

**关键特点**：
- 采用适配器模式，`CacheManagerAdapter` 实现 `CacheManager` 接口
- 内部驱动 (`RedisManager`, `MemoryManager`) 实现统一的 `CacheDriver` 接口
- 所有驱动继承 `BaseManager` 实现 `common.Manager` 接口
- **当前没有任何日志和可观测性能力**

### 2.2 loggermgr 架构分析

**关键特点**：
- `LoggerManager` 接口提供 `Logger(name) Logger` 方法获取日志实例
- 工厂方法 `Build(cfg, telemetryMgr)` 接受 `TelemetryManager` 参数
- `ZapLoggerManager` 内部使用 `OTELCore` 将日志输出到 OTEL
- 支持多个日志核心：Console, File, OTEL

### 2.3 telemetrymgr 架构分析

**关键特点**：
- `TelemetryManager` 接口提供三大观测能力：
  - `Tracer(name) trace.Tracer` - 链路追踪
  - `Meter(name) metric.Meter` - 指标收集
  - `Logger(name) log.Logger` - OTEL 日志
- `OtelManager` 实现完整的 OpenTelemetry 功能
- 已被 `loggermgr` 成功集成

## 3. 设计方案

### 3.1 架构设计

采用 **分层接入** 的方式，在适配器层统一处理日志和可观测性，保持驱动层的简洁性。

```
┌─────────────────────────────────────────────────────────┐
│                 CacheManagerAdapter                      │
│  ├─ logger: Logger (来自 loggermgr)                      │
│  ├─ tracer: Tracer (来自 telemetrymgr)                   │
│  ├─ meter: Meter (来自 telemetrymgr)                     │
│  └─ driver: CacheDriver                                   │
└─────────────────────────────────────────────────────────┘
                          ↓
                          ↓ 调用驱动 + 记录观测数据
                          ↓
┌─────────────────────────────────────────────────────────┐
│              RedisManager / MemoryManager                │
│           (保持无依赖，专注缓存逻辑)                       │
└─────────────────────────────────────────────────────────┘
```

### 3.2 设计原则

1. **保持驱动层纯净**：内部驱动（RedisManager, MemoryManager）不依赖 loggermgr/telemetrymgr
2. **适配器层增强**：在 CacheManagerAdapter 层统一处理日志和可观测性
3. **可选依赖**：loggermgr 和 telemetrymgr 为可选依赖，降级时不影响核心功能
4. **最小侵入**：对现有接口的改动最小化

### 3.3 接口设计

#### 3.3.1 扩展 CacheManagerAdapter

```go
// cache_adapter.go
type CacheManagerAdapter struct {
    driver        CacheDriver
    logger        loggermgr.Logger      // 新增：日志实例
    tracer        trace.Tracer           // 新增：链路追踪
    meter         metric.Meter           // 新增：指标收集
    cacheHitCounter metric.Int64Counter // 新增：缓存命中计数器
    cacheMissCounter metric.Int64Counter // 新增：缓存未命中计数器
    operationDuration metric.Float64Histogram // 新增：操作耗时
}
```

#### 3.3.2 新增观测辅助方法

```go
// 记录缓存操作（带链路追踪和指标）
func (a *CacheManagerAdapter) recordOperation(
    ctx context.Context,
    operation string,
    key string,
    fn func() error,
) error {
    // 1. 创建 span
    ctx, span := a.tracer.Start(ctx, "cache."+operation,
        trace.WithAttributes(
            attribute.String("cache.key", key),
            attribute.String("cache.driver", a.driver.ManagerName()),
        ),
    )
    defer span.End()

    // 2. 执行操作
    start := time.Now()
    err := fn()
    duration := time.Since(start).Seconds()

    // 3. 记录指标
    a.operationDuration.Record(ctx, duration,
        metric.WithAttributes(
            attribute.String("operation", operation),
            attribute.String("status", getStatus(err)),
        ),
    )

    // 4. 记录日志
    if err != nil {
        a.logger.Error("cache operation failed",
            "operation", operation,
            "key", key,
            "error", err.Error(),
            "duration", duration,
        )
        span.RecordError(err)
        span.SetStatus(codes.Error, err.Error())
    } else {
        a.logger.Debug("cache operation success",
            "operation", operation,
            "key", key,
            "duration", duration,
        )
    }

    return err
}

// 记录缓存命中/未命中
func (a *CacheManagerAdapter) recordCacheHit(ctx context.Context, hit bool) {
    if hit {
        a.cacheHitCounter.Add(ctx, 1)
    } else {
        a.cacheMissCounter.Add(ctx, 1)
    }
}
```

### 3.4 工厂方法改造

#### 3.4.1 扩展 Build 函数签名

```go
// factory.go

// Build 创建缓存管理器实例
// cfg: 缓存配置内容
// loggerMgr: 可选的日志管理器
// telemetryMgr: 可选的观测管理器
func Build(
    cfg map[string]any,
    loggerMgr loggermgr.LoggerManager,
    telemetryMgr telemetrymgr.TelemetryManager,
) common.Manager {
    // 1. 解析配置创建驱动
    cacheConfig, err := config.ParseCacheConfigFromMap(cfg)
    if err != nil {
        return NewCacheManagerAdapter(drivers.NewNoneManager(), nil, nil)
    }

    var driver CacheDriver
    switch cacheConfig.Driver {
    case "redis":
        driver, err = drivers.NewRedisManager(cacheConfig.RedisConfig)
    case "memory":
        driver = drivers.NewMemoryManager(...)
    case "none":
        driver = drivers.NewNoneManager()
    }

    if err != nil || driver == nil {
        return NewCacheManagerAdapter(drivers.NewNoneManager(), nil, nil)
    }

    // 2. 获取 logger 和 tracer（可选）
    var logger loggermgr.Logger
    if loggerMgr != nil {
        logger = loggerMgr.Logger("cachemgr")
    }

    var tracer trace.Tracer
    var meter metric.Meter
    if telemetryMgr != nil {
        tracer = telemetryMgr.Tracer("cachemgr")
        meter = telemetryMgr.Meter("cachemgr")
    }

    // 3. 创建增强的适配器
    return NewCacheManagerAdapter(driver, logger, tracer, meter)
}
```

### 3.5 使用示例

```go
// 创建管理器
loggerMgr := loggermgr.Build(loggerCfg, telemetryMgr)
cacheMgr := cachemgr.Build(cacheCfg, loggerMgr, telemetryMgr)

// 使用缓存（自动记录日志和指标）
ctx := context.Background()
var result User
err := cacheMgr.Get(ctx, "user:123", &result)
// 内部会自动：
// - 创建 span 记录调用链路
// - 记录操作耗时指标
// - 记录缓存命中/未命中
// - 输出日志（成功 Debug 级别，失败 Error 级别）
```

## 4. 可观测性指标设计

### 4.1 Metrics 指标

| 指标名称 | 类型 | 描述 | 属性 |
|---------|------|------|------|
| `cache.hit` | Counter | 缓存命中次数 | `driver` |
| `cache.miss` | Counter | 缓存未命中次数 | `driver` |
| `cache.operation.duration` | Histogram | 操作耗时（秒） | `operation`, `status` |
| `cache.operation.size` | Histogram | 批量操作大小 | `operation` |

### 4.2 Traces Span

每个缓存操作创建一个 span，包含以下属性：

- `cache.operation`: 操作类型 (get/set/delete/...)
- `cache.key`: 缓存键（脱敏处理）
- `cache.driver`: 驱动类型 (redis/memory)
- `cache.hit`: 是否命中

### 4.3 Logs 日志

| 级别 | 触发条件 | 内容 |
|------|---------|------|
| Debug | 操作成功 | 操作类型、键、耗时 |
| Info | 重要事件 (启动/关闭) | 管理器名称、配置信息 |
| Warn | 降级/重试 | 降级原因、重试次数 |
| Error | 操作失败 | 错误信息、堆栈 |

## 5. 实施计划

### 5.1 第一阶段：接口和结构改造

- [ ] 扩展 `CacheManagerAdapter` 结构，添加 logger/tracer/meter 字段
- [ ] 修改 `NewCacheManagerAdapter` 函数签名
- [ ] 修改 `Build` 函数签名（保持向后兼容）
- [ ] 添加观测辅助方法

### 5.2 第二阶段：核心方法改造

- [ ] 改造 `Get` 方法：添加 span、指标、日志
- [ ] 改造 `Set` 方法：添加 span、指标、日志
- [ ] 改造 `Delete` 方法：添加 span、指标、日志
- [ ] 改造批量操作方法：添加 span、指标、日志

### 5.3 第三阶段：测试和文档

- [ ] 单元测试：验证日志和指标记录
- [ ] 集成测试：验证与 loggermgr/telemetrymgr 的集成
- [ ] 性能测试：确保性能影响可控（<5%）
- [ ] 更新文档和使用示例

## 6. 向后兼容性

为了保持向后兼容，提供两种 Build 方法：

```go
// 新方法：完整功能
func Build(
    cfg map[string]any,
    loggerMgr loggermgr.LoggerManager,
    telemetryMgr telemetrymgr.TelemetryManager,
) common.Manager

// 兼容方法：不传入日志和观测（降级模式）
func BuildSimple(cfg map[string]any) common.Manager {
    return Build(cfg, nil, nil)
}
```

## 7. 依赖调整

### 7.1 不需要调整 loggermgr 和 telemetrymgr

当前两个管理器的接口设计已经满足需求：
- `loggermgr.LoggerManager.Logger(name)` 获取日志实例
- `telemetrymgr.TelemetryManager.Tracer(name)` 获取 Tracer
- `telemetrymgr.TelemetryManager.Meter(name)` 获取 Meter

### 7.2 新增依赖

cachemgr 需要新增以下导入：

```go
import (
    "com.litelake.litecore/manager/loggermgr"
    "com.litelake.litecore/manager/telemetrymgr"

    "go.opentelemetry.io/otel/trace"
    "go.opentelemetry.io/otel/metric"
    "go.opentelemetry.io/otel/attribute"
)
```

## 8. 风险和注意事项

### 8.1 性能风险

- **风险**：每次操作都创建 span 可能影响性能
- **缓解**：
  - 使用采样策略（如 10% 采样）
  - 对高并发操作（如 Get）采用轻量级记录
  - 提供 `DisableObservation()` 选项

### 8.2 敏感信息泄露

- **风险**：日志中记录键名可能泄露敏感信息
- **缓解**：
  - 实现键名脱敏函数
  - 提供 `SanitizeKey(key string)` 方法
  - 示例：`"user:123456"` → `"user:***"`

### 8.3 循环依赖

- **风险**：loggermgr 依赖 telemetrymgr，如果 cachemgr 也依赖两者可能导致循环依赖
- **现状**：无循环依赖风险，依赖链为单向：
  - cachemgr → loggermgr
  - cachemgr → telemetrymgr
  - loggermgr → telemetrymgr

## 9. 后续优化

### 9.1 短期优化

- [ ] 添加缓存预热日志
- [ ] 添加连接池状态监控
- [ ] 添加慢查询日志（超过阈值）

### 9.2 长期优化

- [ ] 实现缓存穿透防护（布隆过滤器）并记录相关指标
- [ ] 实现缓存雪崩预警
- [ ] 添加缓存命中率趋势分析

## 10. 参考资料

- [OpenTelemetry Specification](https://opentelemetry.io/docs/reference/specification/)
- [OpenTelemetry Go SDK](https://github.com/open-telemetry/opentelemetry-go)
- [项目 loggermgr 文档](../manager/loggermgr/README.md)
- [项目 telemetrymgr 文档](../manager/telemetrymgr/README.md)
