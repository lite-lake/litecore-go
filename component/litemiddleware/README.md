# litemiddleware

内置 HTTP 中间件组件，提供开箱即用的通用中间件实现。

## 特性

- **统一接口** - 所有中间件实现 `common.IBaseMiddleware` 接口
- **灵活配置** - 配置属性使用指针类型，支持可选配置和默认值覆盖
- **依赖注入** - 通过 `inject:""` 标签自动注入 Manager 组件
- **执行顺序** - 预定义 Order 常量，支持自定义执行顺序
- **完整测试** - 所有中间件包含单元测试和示例

## 快速开始

```go
import "github.com/lite-lake/litecore-go/component/litemiddleware"

// 使用默认配置创建中间件
recovery := litemiddleware.NewRecoveryMiddlewareWithDefaults()
reqLogger := litemiddleware.NewRequestLoggerMiddlewareWithDefaults()
cors := litemiddleware.NewCorsMiddlewareWithDefaults()
security := litemiddleware.NewSecurityHeadersMiddlewareWithDefaults()

// 注册到容器
container.RegisterMiddleware(middlewareContainer, recovery)
container.RegisterMiddleware(middlewareContainer, reqLogger)
container.RegisterMiddleware(middlewareContainer, cors)
container.RegisterMiddleware(middlewareContainer, security)
```

## 可用中间件列表

| 中间件 | 功能 | Order | 依赖 |
|--------|------|-------|------|
| Recovery | panic 恢复 | 0 | LoggerManager |
| RequestLogger | 请求日志 | 50 | LoggerManager |
| CORS | 跨域处理 | 100 | 无 |
| SecurityHeaders | 安全头 | 150 | 无 |
| RateLimiter | 限流 | 200 | LimiterManager, LoggerManager |
| Telemetry | 遥测 | 250 | TelemetryManager |

## 配置说明

所有中间件配置支持以下通用字段：

| 字段 | 类型 | 说明 |
|------|------|------|
| Name | *string | 中间件名称，用于日志和标识 |
| Order | *int | 执行顺序，数值越小越先执行 |

配置属性使用指针类型（`*string`, `*int`, `*bool` 等），未配置的字段将使用默认值。这种设计允许：
- 零值区分（如 `false` 与未配置）
- 默认值覆盖
- 可选配置

## Recovery 中间件

panic 恢复中间件，捕获 panic 并记录日志，返回友好的错误响应。

### 配置选项

| 字段 | 类型 | 默认值 | 说明 |
|------|------|--------|------|
| Name | *string | "RecoveryMiddleware" | 中间件名称 |
| Order | *int | 0 | 执行顺序 |
| PrintStack | *bool | true | 是否打印堆栈信息 |
| CustomErrorBody | *bool | true | 是否使用 JSON 格式错误响应 |
| ErrorMessage | *string | "Internal server error" | 自定义错误消息 |
| ErrorCode | *string | "INTERNAL_SERVER_ERROR" | 自定义错误代码 |

### 使用示例

```go
// 使用默认配置
recovery := litemiddleware.NewRecoveryMiddlewareWithDefaults()

// 生产环境配置（不打印堆栈）
printStack := false
cfg := &litemiddleware.RecoveryConfig{
    PrintStack:      &printStack,
    CustomErrorBody: &[]bool{true}[0],
    ErrorMessage:    &[]string{"系统繁忙，请稍后重试"}[0],
    ErrorCode:       &[]string{"SYSTEM_BUSY"}[0],
}
recovery := litemiddleware.NewRecoveryMiddleware(cfg)
```

## RequestLogger 中间件

请求日志中间件，记录请求和响应信息，支持日志级别控制和路径过滤。

### 配置选项

| 字段 | 类型 | 默认值 | 说明 |
|------|------|--------|------|
| Name | *string | "RequestLoggerMiddleware" | 中间件名称 |
| Order | *int | 50 | 执行顺序 |
| Enable | *bool | true | 是否启用请求日志 |
| LogBody | *bool | true | 是否记录请求 Body |
| MaxBodySize | *int | 4096 | 最大记录 Body 大小（字节），0 表示不限制 |
| SkipPaths | *[]string | []{"/health", "/metrics"} | 跳过日志记录的路径 |
| LogHeaders | *[]string | []{"User-Agent", "Content-Type"} | 需要记录的请求头 |
| SuccessLogLevel | *string | "info" | 成功请求日志级别（debug/info） |

### 使用示例

```go
// 使用默认配置
reqLogger := litemiddleware.NewRequestLoggerMiddlewareWithDefaults()

// 禁用 Body 记录，调整日志级别为 debug
logBody := false
successLogLevel := "debug"
cfg := &litemiddleware.RequestLoggerConfig{
    LogBody:         &logBody,
    SuccessLogLevel: &successLogLevel,
}
reqLogger := litemiddleware.NewRequestLoggerMiddleware(cfg)
```

## CORS 中间件

跨域资源共享中间件，支持灵活的跨域配置。

### 配置选项

| 字段 | 类型 | 默认值 | 说明 |
|------|------|--------|------|
| Name | *string | "CorsMiddleware" | 中间件名称 |
| Order | *int | 100 | 执行顺序 |
| AllowOrigins | *[]string | []{"*"} | 允许的源 |
| AllowMethods | *[]string | []{"GET", "POST", "PUT", "DELETE", "PATCH", "OPTIONS"} | 允许的 HTTP 方法 |
| AllowHeaders | *[]string | []{"Origin", "Content-Type", "Authorization", "X-Requested-With", "Accept", "Cache-Control"} | 允许的请求头 |
| ExposeHeaders | *[]string | nil | 暴露的响应头 |
| AllowCredentials | *bool | true | 是否允许携带凭证 |
| MaxAge | *time.Duration | 12h | 预检请求缓存时间 |

### 使用示例

```go
// 使用默认配置（允许所有源）
cors := litemiddleware.NewCorsMiddlewareWithDefaults()

// 生产环境配置（仅允许特定域名）
allowOrigins := []string{"https://example.com", "https://app.example.com"}
allowCredentials := true
cfg := &litemiddleware.CorsConfig{
    AllowOrigins:     &allowOrigins,
    AllowCredentials: &allowCredentials,
}
cors := litemiddleware.NewCorsMiddleware(cfg)
```

## SecurityHeaders 中间件

安全头中间件，自动添加常见的安全 HTTP 头。

### 配置选项

| 字段 | 类型 | 默认值 | 说明 |
|------|------|--------|------|
| Name | *string | "SecurityHeadersMiddleware" | 中间件名称 |
| Order | *int | 150 | 执行顺序 |
| FrameOptions | *string | "DENY" | X-Frame-Options |
| ContentTypeOptions | *string | "nosniff" | X-Content-Type-Options |
| XSSProtection | *string | "1; mode=block" | X-XSS-Protection |
| ReferrerPolicy | *string | "strict-origin-when-cross-origin" | Referrer-Policy |
| ContentSecurityPolicy | *string | nil | Content-Security-Policy |
| StrictTransportSecurity | *string | nil | Strict-Transport-Security |

### 使用示例

```go
// 使用默认配置
security := litemiddleware.NewSecurityHeadersMiddlewareWithDefaults()

// 添加 CSP 和 HSTS
csp := "default-src 'self'; script-src 'self'"
hsts := "max-age=31536000; includeSubDomains"
cfg := &litemiddleware.SecurityHeadersConfig{
    ContentSecurityPolicy:   &csp,
    StrictTransportSecurity: &hsts,
}
security := litemiddleware.NewSecurityHeadersMiddleware(cfg)
```

## RateLimiter 中间件

限流中间件，基于时间窗口的请求频率控制，支持多种限流策略。

### 配置选项

| 字段 | 类型 | 默认值 | 说明 |
|------|------|--------|------|
| Name | *string | "RateLimiterMiddleware" | 中间件名称 |
| Order | *int | 200 | 执行顺序 |
| Limit | *int | 100 | 时间窗口内最大请求数 |
| Window | *time.Duration | 1m | 时间窗口大小 |
| KeyFunc | KeyFunc | func(c) string { c.ClientIP() } | 自定义 key 生成函数 |
| SkipFunc | SkipFunc | nil | 跳过限流的条件 |
| KeyPrefix | *string | "rate_limit" | key 前缀 |

### 使用示例

```go
// 使用默认配置（按 IP 限流，每分钟 100 次请求）
limiter := litemiddleware.NewRateLimiterMiddlewareWithDefaults()

// 按用户 ID 限流，自定义跳过条件
limit := 10
window := time.Minute
keyPrefix := "user"
cfg := &litemiddleware.RateLimiterConfig{
    Limit:     &limit,
    Window:    &window,
    KeyPrefix: &keyPrefix,
    KeyFunc: func(c *gin.Context) string {
        if userID, exists := c.Get("user_id"); exists {
            return userID.(string)
        }
        return c.ClientIP()
    },
    SkipFunc: func(c *gin.Context) bool {
        return c.GetHeader("X-Internal") == "true"
    },
}
limiter := litemiddleware.NewRateLimiterMiddleware(cfg)
```

### 响应头

| 响应头 | 说明 |
|--------|------|
| X-RateLimit-Limit | 时间窗口内最大请求数 |
| X-RateLimit-Remaining | 剩余可用请求数 |
| X-RateLimit-Reset | 窗口重置时间 |
| Retry-After | 限流时重试等待时间 |

## Telemetry 中间件

遥测中间件，集成 OpenTelemetry 进行链路追踪和指标采集。

### 配置选项

| 字段 | 类型 | 默认值 | 说明 |
|------|------|--------|------|
| Name | string | "TelemetryMiddleware" | 中间件名称 |
| Order | *int | 250 | 执行顺序 |

### 使用示例

```go
// 使用默认配置（依赖注入 TelemetryManager）
telemetry := litemiddleware.NewTelemetryMiddlewareWithDefaults()
```

## 执行顺序

预定义的中间件执行顺序（按 Order 值从小到大）：

| Order | 中间件 | 说明 |
|-------|--------|------|
| 0 | Recovery | panic 恢复（最先执行） |
| 50 | RequestLogger | 请求日志 |
| 100 | CORS | 跨域处理 |
| 150 | SecurityHeaders | 安全头 |
| 200 | RateLimiter | 限流 |
| 250 | Telemetry | 遥测 |
| 300 | Auth | 认证（预留） |

业务自定义中间件建议从 Order 350 开始。

## 依赖注入

中间件通过依赖注入获取 Manager 组件：

```go
type requestLoggerMiddleware struct {
    LoggerMgr loggermgr.ILoggerManager `inject:""`
    cfg       *RequestLoggerConfig
}

type rateLimiterMiddleware struct {
    LimiterMgr limitermgr.ILimiterManager `inject:""`
    LoggerMgr  loggermgr.ILoggerManager   `inject:""`
    config     *RateLimiterConfig
}

type telemetryMiddleware struct {
    TelemetryManager telemetrymgr.ITelemetryManager `inject:""`
    cfg              *TelemetryConfig
}
```

## 业务层封装

在业务项目中，可以定义自己的中间件接口并封装 litemiddleware 的实现：

```go
// internal/middlewares/cors_middleware.go
package middlewares

import (
    "github.com/lite-lake/litecore-go/common"
    "github.com/lite-lake/litecore-go/component/litemiddleware"
)

type ICorsMiddleware interface {
    common.IBaseMiddleware
}

func NewCorsMiddleware() ICorsMiddleware {
    return litemiddleware.NewCorsMiddlewareWithDefaults()
}

func NewProductionCorsMiddleware() ICorsMiddleware {
    allowOrigins := []string{"https://example.com"}
    allowCredentials := true
    return litemiddleware.NewCorsMiddleware(&litemiddleware.CorsConfig{
        AllowOrigins:     &allowOrigins,
        AllowCredentials: &allowCredentials,
    })
}
```

## 性能优化建议

- **RequestLogger**：生产环境关闭 Body 记录（`LogBody: false`）
- **Recovery**：生产环境关闭堆栈打印（`PrintStack: false`）
- **RateLimiter**：合理设置 Limit 和 Window，避免过度限制
- **SkipFunc**：使用 SkipFunc 跳过内部请求或健康检查的限流

## 版本历史

### v2.0.0 (2026-01-24)

- **目录重构**：从 `component/middleware` 迁移至 `component/litemiddleware`
- **包名变更**：`middleware` → `litemiddleware`
- **配置增强**：所有中间件支持通过配置自定义 Name 和 Order
- **配置重构**：配置属性改为指针类型，支持可选配置
- **新增功能**：RateLimiter 限流中间件
- **示例完善**：添加了完整的使用示例和测试用例
