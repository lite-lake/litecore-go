# 中间件配置化改造总结

## 改造内容

本次改造为所有内置中间件添加了配置支持，使其在业务系统中可以灵活自定义配置。

## 改造的中间件

### 1. CORS 中间件 (`cors_middleware.go`)

**新增配置结构：**
```go
type CorsConfig struct {
    AllowOrigins     []string      // 允许的源
    AllowMethods     []string      // 允许的 HTTP 方法
    AllowHeaders     []string      // 允许的请求头
    ExposeHeaders    []string      // 暴露的响应头
    AllowCredentials bool          // 是否允许携带凭证
    MaxAge           time.Duration // 预检请求缓存时间
}
```

**新增构造函数：**
- `NewCorsMiddleware(config *CorsConfig)` - 使用自定义配置
- `NewCorsMiddlewareWithDefaults()` - 使用默认配置

**解决的问题：**
- ✅ 修复了硬编码 `Allow-Origin: *` 的安全问题
- ✅ 支持自定义允许的源
- ✅ 支持自定义所有 CORS 相关配置

**使用示例：**
```go
// 开发环境（允许所有源）
devCors := middleware.NewCorsMiddlewareWithDefaults()

// 生产环境（仅允许特定域名）
prodCors := middleware.NewCorsMiddleware(&middleware.CorsConfig{
    AllowOrigins: []string{"https://example.com"},
    AllowCredentials: true,
    MaxAge: 12 * time.Hour,
})
```

---

### 2. RequestLogger 中间件 (`request_logger_middleware.go`)

**新增配置结构：**
```go
type RequestLoggerConfig struct {
    Enable          bool     // 是否启用请求日志
    LogBody         bool     // 是否记录请求 Body
    MaxBodySize     int      // 最大记录 Body 大小（字节），0 表示不限制
    SkipPaths       []string // 跳过日志记录的路径
    LogHeaders      []string // 需要记录的请求头
    SuccessLogLevel string   // 成功请求日志级别（debug/info）
}
```

**新增构造函数：**
- `NewRequestLoggerMiddleware(config *RequestLoggerConfig)` - 使用自定义配置
- `NewRequestLoggerMiddlewareWithDefaults()` - 使用默认配置

**解决的问题：**
- ✅ 支持禁用日志记录
- ✅ 支持控制是否记录 Body
- ✅ 支持限制 Body 大小，避免内存占用过大
- ✅ 支持跳过某些路径（如健康检查）

**使用示例：**
```go
// 默认配置（启用日志，记录 Body）
reqLogger := middleware.NewRequestLoggerMiddlewareWithDefaults()

// 生产环境（不记录 Body，跳过健康检查）
prodLogger := middleware.NewRequestLoggerMiddleware(&middleware.RequestLoggerConfig{
    Enable:      true,
    LogBody:     false,
    MaxBodySize: 2048,
    SkipPaths:   []string{"/health", "/metrics"},
})
```

---

### 3. SecurityHeaders 中间件 (`security_headers_middleware.go`)

**新增配置结构：**
```go
type SecurityHeadersConfig struct {
    FrameOptions              string // X-Frame-Options (DENY, SAMEORIGIN, ALLOW-FROM)
    ContentTypeOptions        string // X-Content-Type-Options (nosniff)
    XSSProtection             string // X-XSS-Protection (1; mode=block)
    ReferrerPolicy            string // Referrer-Policy (strict-origin-when-cross-origin, no-referrer, etc)
    ContentSecurityPolicy     string // Content-Security-Policy
    StrictTransportSecurity   string // Strict-Transport-Security (max-age=31536000; includeSubDomains)
}
```

**新增构造函数：**
- `NewSecurityHeadersMiddleware(config *SecurityHeadersConfig)` - 使用自定义配置
- `NewSecurityHeadersMiddlewareWithDefaults()` - 使用默认配置

**解决的问题：**
- ✅ 支持自定义安全头
- ✅ 支持添加 CSP 和 HSTS 头

**使用示例：**
```go
// 默认配置（基本安全头）
security := middleware.NewSecurityHeadersMiddlewareWithDefaults()

// 生产环境（包含 CSP 和 HSTS）
prodSecurity := middleware.NewSecurityHeadersMiddleware(&middleware.SecurityHeadersConfig{
    FrameOptions:              "SAMEORIGIN",
    ContentSecurityPolicy:     "default-src 'self'; script-src 'self' 'unsafe-inline'",
    StrictTransportSecurity:  "max-age=31536000; includeSubDomains",
})
```

---

### 4. Recovery 中间件 (`recovery_middleware.go`)

**新增配置结构：**
```go
type RecoveryConfig struct {
    PrintStack      bool   // 是否打印堆栈信息
    CustomErrorBody bool   // 是否使用自定义错误响应
    ErrorMessage    string // 自定义错误消息
    ErrorCode       string // 自定义错误代码
}
```

**新增构造函数：**
- `NewRecoveryMiddleware(config *RecoveryConfig)` - 使用自定义配置
- `NewRecoveryMiddlewareWithDefaults()` - 使用默认配置

**解决的问题：**
- ✅ 支持控制是否打印堆栈（生产环境可关闭）
- ✅ 支持自定义错误消息和代码

**使用示例：**
```go
// 默认配置（打印堆栈）
recovery := middleware.NewRecoveryMiddlewareWithDefaults()

// 生产环境（不打印堆栈，自定义错误）
prodRecovery := middleware.NewRecoveryMiddleware(&middleware.RecoveryConfig{
    PrintStack:      false,
    CustomErrorBody: true,
    ErrorMessage:    "系统繁忙，请稍后重试",
    ErrorCode:       "SYSTEM_BUSY",
})
```

---

### 5. RateLimiter 中间件 (`rate_limiter_middleware.go`)

**已有配置结构，新增：**
- `DefaultRateLimiterConfig()` - 返回默认配置
- `NewRateLimiterWithDefaults()` - 使用默认配置

**使用示例：**
```go
// 使用默认配置
limiter := middleware.NewRateLimiterWithDefaults()

// 按自定义配置
customLimiter := middleware.NewRateLimiter(&middleware.RateLimiterConfig{
    Limit:     200,
    Window:    time.Minute,
    KeyPrefix: "custom",
})

// 使用便捷函数
ipLimiter := middleware.NewRateLimiterByIP(100, time.Minute)
pathLimiter := middleware.NewRateLimiterByPath(1000, time.Minute)
```

---

## 依赖注入

所有中间件的依赖注入保持不变，仍然使用 `inject:""` 标签注入 Manager：

```go
type RequestLoggerMiddleware struct {
    order     int
    LoggerMgr loggermgr.ILoggerManager `inject:""`
    cfg       *RequestLoggerConfig
}

type RateLimiterMiddleware struct {
    order      int
    LimiterMgr limitermgr.ILimiterManager `inject:""`
    LoggerMgr  loggermgr.ILoggerManager   `inject:""`
    config     *RateLimiterConfig
}
```

依赖注入会在容器初始化时自动完成，不受配置影响。

---

## 业务系统使用方式

### 方式 1：使用默认配置（快速开发）

```go
// internal/middlewares/cors_middleware.go
func NewCorsMiddleware() ICorsMiddleware {
    return componentMiddleware.NewCorsMiddlewareWithDefaults()
}
```

### 方式 2：使用自定义配置（生产环境）

```go
// internal/middlewares/cors_middleware.go
import "time"

func NewProductionCorsMiddleware() ICorsMiddleware {
    cfg := &componentMiddleware.CorsConfig{
        AllowOrigins:     []string{"https://example.com"},
        AllowCredentials: true,
        MaxAge:           12 * time.Hour,
    }
    return componentMiddleware.NewCorsMiddleware(cfg)
}
```

### 方式 3：直接使用 component 层中间件

```go
import "github.com/lite-lake/litecore-go/component/litemiddleware"

// 在容器中注册
container.RegisterMiddleware(middlewareContainer, middleware.NewCorsMiddlewareWithDefaults())
```

---

## 中间件执行顺序

预定义的中间件执行顺序（按 Order 值从小到大）：

| 中间件 | Order | 说明 |
|--------|-------|------|
| Recovery | 0 | panic 恢复（最先执行） |
| RequestLogger | 50 | 请求日志 |
| CORS | 100 | 跨域处理 |
| SecurityHeaders | 150 | 安全头 |
| RateLimiter | 200 | 限流 |
| Telemetry | 250 | 遥测 |

业务自定义中间件建议从 Order 350 开始。

---

## 测试

所有中间件都有完整的单元测试，测试包括：
- 默认配置测试
- 自定义配置测试
- 功能正确性测试
- 边界条件测试

运行测试：
```bash
go test ./component/litemiddleware/... -v
```

---

## 向后兼容性

**不兼容的改动：**
- 所有 `NewXxxMiddleware()` 函数现在需要传入 `*XxxConfig` 参数
- 需要更新所有调用这些函数的代码

**兼容性处理：**
- 提供了 `NewXxxMiddlewareWithDefaults()` 函数，使用默认配置
- 示例项目的中间件包装层已更新

**迁移步骤：**
1. 将 `NewXxxMiddleware()` 调用改为 `NewXxxMiddlewareWithDefaults()`
2. 或使用 `NewXxxMiddleware(&middleware.XxxConfig{})` 传入自定义配置

---

## 文件清单

### 修改的文件
- `component/middleware/cors_middleware.go` - 添加 CORS 配置支持
- `component/middleware/recovery_middleware.go` - 添加 Recovery 配置支持
- `component/middleware/request_logger_middleware.go` - 添加 RequestLogger 配置支持
- `component/middleware/security_headers_middleware.go` - 添加 SecurityHeaders 配置支持
- `component/middleware/rate_limiter_middleware.go` - 添加 WithDefaults 函数

### 新增的文件
- `component/middleware/example_test.go` - 中间件使用示例

### 测试文件
- `component/middleware/cors_middleware_test.go` - CORS 测试（已更新）
- `component/middleware/security_headers_middleware_test.go` - SecurityHeaders 测试（已更新）

### 示例项目
- `samples/messageboard/internal/middlewares/*.go` - 更新所有包装中间件

---

## 使用建议

### 开发环境
使用默认配置，快速开发：
```go
container.RegisterMiddleware(middlewareContainer, middleware.NewCorsMiddlewareWithDefaults())
```

### 生产环境
自定义配置，增强安全性：
```go
container.RegisterMiddleware(middlewareContainer, middleware.NewCorsMiddleware(&middleware.CorsConfig{
    AllowOrigins: []string{"https://yourdomain.com"},
    AllowCredentials: true,
}))
```

### 性能优化
- 关闭 RequestLogger 的 Body 记录（大文件场景）
- 关闭 Recovery 的堆栈打印（生产环境）
- 合理设置 RateLimiter 限制

---

## 总结

本次改造实现了：
1. ✅ 所有中间件都支持自定义配置
2. ✅ 提供默认配置，简化使用
3. ✅ 保持依赖注入机制不变
4. ✅ 修复了 CORS 安全问题
5. ✅ 优化了 RequestLogger 性能
6. ✅ 所有测试通过

业务系统现在可以灵活地根据环境（开发/测试/生产）配置不同的中间件参数。
