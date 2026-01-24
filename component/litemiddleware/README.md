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

## 中间件列表

### Recovery 中间件

Panic 恢复中间件，捕获 panic 并记录日志，返回友好的错误响应。

```go
// 使用默认配置
recovery := litemiddleware.NewRecoveryMiddleware(nil)

// 自定义配置（生产环境不打印堆栈）
printStack := false
cfg := &litemiddleware.RecoveryConfig{
    PrintStack:      &printStack,
    CustomErrorBody: &[]bool{true}[0],
    ErrorMessage:    &[]string{"系统繁忙，请稍后重试"}[0],
    ErrorCode:       &[]string{"SYSTEM_BUSY"}[0],
}
recovery := litemiddleware.NewRecoveryMiddleware(cfg)
```

### RequestLogger 中间件

请求日志中间件，记录请求和响应信息，支持日志级别控制和路径过滤。

```go
// 使用默认配置（记录 Body，跳过 /health 和 /metrics）
reqLogger := litemiddleware.NewRequestLoggerMiddleware(nil)

// 禁用 Body 记录，调整日志级别为 debug
logBody := false
successLogLevel := "debug"
cfg := &litemiddleware.RequestLoggerConfig{
    LogBody:         &logBody,
    SuccessLogLevel: &successLogLevel,
}
reqLogger := litemiddleware.NewRequestLoggerMiddleware(cfg)
```

### CORS 中间件

跨域资源共享中间件，支持灵活的跨域配置。

```go
// 使用默认配置（允许所有源）
cors := litemiddleware.NewCorsMiddleware(nil)

// 生产环境配置（仅允许特定域名）
allowOrigins := []string{"https://example.com", "https://app.example.com"}
allowCredentials := true
cfg := &litemiddleware.CorsConfig{
    AllowOrigins:     &allowOrigins,
    AllowCredentials: &allowCredentials,
}
cors := litemiddleware.NewCorsMiddleware(cfg)
```

### SecurityHeaders 中间件

安全头中间件，自动添加常见的安全 HTTP 头。

```go
// 使用默认配置（X-Frame-Options: DENY, X-Content-Type-Options: nosniff 等）
security := litemiddleware.NewSecurityHeadersMiddleware(nil)

// 添加 CSP 和 HSTS
csp := "default-src 'self'; script-src 'self'"
hsts := "max-age=31536000; includeSubDomains"
cfg := &litemiddleware.SecurityHeadersConfig{
    ContentSecurityPolicy:   &csp,
    StrictTransportSecurity: &hsts,
}
security := litemiddleware.NewSecurityHeadersMiddleware(cfg)
```

### RateLimiter 中间件

限流中间件，基于时间窗口的请求频率控制，支持多种限流策略。

```go
// 使用默认配置（按 IP 限流，每分钟 100 次请求）
limiter := litemiddleware.NewRateLimiterMiddleware(nil)

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

### Telemetry 中间件

遥测中间件，集成 OpenTelemetry 进行链路追踪和指标采集。

```go
// 使用默认配置（依赖注入 TelemetryManager）
telemetry := litemiddleware.NewTelemetryMiddleware(nil)
```

## API

### 配置结构

| 配置结构 | 说明 |
|----------|------|
| `CorsConfig` | CORS 跨域配置 |
| `RecoveryConfig` | Panic 恢复配置 |
| `RequestLoggerConfig` | 请求日志配置 |
| `SecurityHeadersConfig` | 安全头配置 |
| `RateLimiterConfig` | 限流配置 |
| `TelemetryConfig` | 遥测配置 |

### 构造函数

| 函数 | 说明 |
|------|------|
| `NewCorsMiddleware(*CorsConfig)` | 创建 CORS 中间件 |
| `NewRecoveryMiddleware(*RecoveryConfig)` | 创建 Recovery 中间件 |
| `NewRequestLoggerMiddleware(*RequestLoggerConfig)` | 创建 RequestLogger 中间件 |
| `NewSecurityHeadersMiddleware(*SecurityHeadersConfig)` | 创建 SecurityHeaders 中间件 |
| `NewRateLimiterMiddleware(*RateLimiterConfig)` | 创建 RateLimiter 中间件 |
| `NewTelemetryMiddleware(*TelemetryConfig)` | 创建 Telemetry 中间件 |

### 便捷函数

| 函数 | 说明 |
|------|------|
| `NewCorsMiddlewareWithDefaults()` | 使用默认配置创建 CORS 中间件 |
| `NewRecoveryMiddlewareWithDefaults()` | 使用默认配置创建 Recovery 中间件 |
| `NewRequestLoggerMiddlewareWithDefaults()` | 使用默认配置创建 RequestLogger 中间件 |
| `NewSecurityHeadersMiddlewareWithDefaults()` | 使用默认配置创建 SecurityHeaders 中间件 |
| `NewRateLimiterMiddlewareWithDefaults()` | 使用默认配置创建 RateLimiter 中间件 |
| `NewTelemetryMiddlewareWithDefaults()` | 使用默认配置创建 Telemetry 中间件 |

### 接口方法

所有中间件实现以下方法：

```go
type IBaseMiddleware interface {
    MiddlewareName() string  // 返回中间件名称
    Order() int              // 返回执行顺序
    Wrapper() gin.HandlerFunc // 返回 Gin 中间件函数
    OnStart() error           // 服务器启动时触发
    OnStop() error            // 服务器停止时触发
}
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
