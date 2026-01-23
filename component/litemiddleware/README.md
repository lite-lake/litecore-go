# 中间件配置指南

## 配置设计

所有内置中间件都支持灵活的配置方式。配置属性使用指针类型，支持可选配置，未配置的属性自动使用默认值。

## 核心特性

### 1. 可选配置（指针类型）

所有配置属性都使用指针类型，这意味着：
- **可选配置**：可以只配置需要修改的属性
- **自动默认值**：未配置的属性自动使用默认值
- **灵活组合**：支持任意属性组合

### 2. 默认值覆盖机制

每个 `NewXxxMiddleware` 函数内部都会：
1. 接收 `*XxxConfig`（可以为 nil）
2. 获取默认配置
3. 用用户配置覆盖默认配置中的对应字段

```go
func NewCorsMiddleware(config *CorsConfig) common.IBaseMiddleware {
    cfg := config
    if cfg == nil {
        cfg = &CorsConfig{}
    }
    
    defaultCfg := DefaultCorsConfig()
    
    // 只覆盖用户配置的字段
    if cfg.Name == nil {
        cfg.Name = defaultCfg.Name
    }
    if cfg.AllowOrigins == nil {
        cfg.AllowOrigins = defaultCfg.AllowOrigins
    }
    // ...
    
    return &corsMiddleware{cfg: cfg}
}
```

---

## 支持的中间件

### 1. CORS 中间件 (`cors_middleware.go`)

**配置结构：**
```go
type CorsConfig struct {
    Name             *string       // 中间件名称
    Order            *int          // 执行顺序
    AllowOrigins     *[]string     // 允许的源
    AllowMethods     *[]string     // 允许的 HTTP 方法
    AllowHeaders     *[]string     // 允许的请求头
    ExposeHeaders    *[]string     // 暴露的响应头
    AllowCredentials *bool         // 是否允许携带凭证
    MaxAge           *time.Duration // 预检请求缓存时间
}
```

**默认配置：**
- `AllowOrigins: []string{"*"}`
- `AllowMethods: []string{"GET", "POST", "PUT", "DELETE", "PATCH", "OPTIONS"}`
- `AllowHeaders: []string{"Origin", "Content-Type", "Authorization", ...}`
- `AllowCredentials: true`
- `MaxAge: 12 * time.Hour`

**使用示例：**
```go
// 使用默认配置（允许所有源）
devCors := litemiddleware.NewCorsMiddleware(nil)

// 只修改允许的源（其他使用默认值）
allowOrigins := []string{"https://example.com"}
prodCors := litemiddleware.NewCorsMiddleware(&litemiddleware.CorsConfig{
    AllowOrigins: &allowOrigins,
})

// 修改多个属性
allowOrigins := []string{"https://example.com", "https://app.example.com"}
allowMethods := []string{"GET", "POST", "PUT"}
allowCredentials := true
customCors := litemiddleware.NewCorsMiddleware(&litemiddleware.CorsConfig{
    AllowOrigins:     &allowOrigins,
    AllowMethods:     &allowMethods,
    AllowCredentials: &allowCredentials,
})
```

---

### 2. RequestLogger 中间件 (`request_logger_middleware.go`)

**配置结构：**
```go
type RequestLoggerConfig struct {
    Name            *string  // 中间件名称
    Order           *int     // 执行顺序
    Enable          *bool    // 是否启用请求日志
    LogBody         *bool    // 是否记录请求 Body
    MaxBodySize     *int     // 最大记录 Body 大小（字节）
    SkipPaths       *[]string // 跳过日志记录的路径
    LogHeaders      *[]string // 需要记录的请求头
    SuccessLogLevel *string  // 成功请求日志级别（debug/info）
}
```

**默认配置：**
- `Enable: true`
- `LogBody: true`
- `MaxBodySize: 4096`
- `SkipPaths: []string{"/health", "/metrics"}`
- `LogHeaders: []string{"User-Agent", "Content-Type"}`
- `SuccessLogLevel: "info"`

**使用示例：**
```go
// 使用默认配置
reqLogger := litemiddleware.NewRequestLoggerMiddleware(nil)

// 只禁用 Body 记录
logBody := false
prodLogger := litemiddleware.NewRequestLoggerMiddleware(&litemiddleware.RequestLoggerConfig{
    LogBody: &logBody,
})

// 完全自定义
enable := true
logBody := false
maxBodySize := 2048
skipPaths := []string{"/health", "/metrics", "/ping"}
logHeaders := []string{"User-Agent", "X-Request-ID"}
successLogLevel := "debug"
customLogger := litemiddleware.NewRequestLoggerMiddleware(&litemiddleware.RequestLoggerConfig{
    Enable:          &enable,
    LogBody:         &logBody,
    MaxBodySize:     &maxBodySize,
    SkipPaths:       &skipPaths,
    LogHeaders:      &logHeaders,
    SuccessLogLevel: &successLogLevel,
})
```

---

### 3. SecurityHeaders 中间件 (`security_headers_middleware.go`)

**配置结构：**
```go
type SecurityHeadersConfig struct {
    Name                    *string // 中间件名称
    Order                   *int    // 执行顺序
    FrameOptions            *string // X-Frame-Options
    ContentTypeOptions      *string // X-Content-Type-Options
    XSSProtection           *string // X-XSS-Protection
    ReferrerPolicy          *string // Referrer-Policy
    ContentSecurityPolicy   *string // Content-Security-Policy
    StrictTransportSecurity *string // Strict-Transport-Security
}
```

**默认配置：**
- `FrameOptions: "DENY"`
- `ContentTypeOptions: "nosniff"`
- `XSSProtection: "1; mode=block"`
- `ReferrerPolicy: "strict-origin-when-cross-origin"`

**使用示例：**
```go
// 使用默认配置
security := litemiddleware.NewSecurityHeadersMiddleware(nil)

// 添加 CSP
csp := "default-src 'self'; script-src 'self'"
customSecurity := litemiddleware.NewSecurityHeadersMiddleware(&litemiddleware.SecurityHeadersConfig{
    ContentSecurityPolicy: &csp,
})

// 完全自定义
frameOptions := "SAMEORIGIN"
contentTypeOptions := "nosniff"
xssProtection := "1; mode=block"
referrerPolicy := "no-referrer"
csp := "default-src 'self'"
hsts := "max-age=31536000; includeSubDomains"
prodSecurity := litemiddleware.NewSecurityHeadersMiddleware(&litemiddleware.SecurityHeadersConfig{
    FrameOptions:            &frameOptions,
    ContentTypeOptions:      &contentTypeOptions,
    XSSProtection:           &xssProtection,
    ReferrerPolicy:          &referrerPolicy,
    ContentSecurityPolicy:   &csp,
    StrictTransportSecurity: &hsts,
})
```

---

### 4. Recovery 中间件 (`recovery_middleware.go`)

**配置结构：**
```go
type RecoveryConfig struct {
    Name            *string // 中间件名称
    Order           *int    // 执行顺序
    PrintStack      *bool   // 是否打印堆栈信息
    CustomErrorBody *bool   // 是否使用自定义错误响应
    ErrorMessage    *string // 自定义错误消息
    ErrorCode       *string // 自定义错误代码
}
```

**默认配置：**
- `PrintStack: true`
- `CustomErrorBody: true`
- `ErrorMessage: "内部服务器错误"`
- `ErrorCode: "INTERNAL_SERVER_ERROR"`

**使用示例：**
```go
// 使用默认配置（打印堆栈）
recovery := litemiddleware.NewRecoveryMiddleware(nil)

// 生产环境（不打印堆栈）
printStack := false
customErrorBody := true
errorMessage := "系统繁忙，请稍后重试"
errorCode := "SYSTEM_BUSY"
prodRecovery := litemiddleware.NewRecoveryMiddleware(&litemiddleware.RecoveryConfig{
    PrintStack:      &printStack,
    CustomErrorBody: &customErrorBody,
    ErrorMessage:    &errorMessage,
    ErrorCode:       &errorCode,
})
```

---

### 5. RateLimiter 中间件 (`rate_limiter_middleware.go`)

**配置结构：**
```go
type RateLimiterConfig struct {
    Name      *string       // 中间件名称
    Order     *int          // 执行顺序
    Limit     *int          // 时间窗口内最大请求数
    Window    *time.Duration // 时间窗口大小
    KeyFunc   KeyFunc       // 自定义key生成函数
    SkipFunc  SkipFunc      // 跳过限流的条件
    KeyPrefix *string       // key前缀
}
```

**默认配置：**
- `Limit: 100`
- `Window: time.Minute`
- `KeyPrefix: "rate_limit"`
- `KeyFunc: 按IP生成key`

**使用示例：**
```go
// 使用默认配置（按IP限流）
limiter := litemiddleware.NewRateLimiterMiddleware(nil)

// 自定义限流规则
limit := 200
window := time.Minute
keyPrefix := "api"
customLimiter := litemiddleware.NewRateLimiterMiddleware(&litemiddleware.RateLimiterConfig{
    Limit:     &limit,
    Window:    &window,
    KeyPrefix: &keyPrefix,
    KeyFunc: func(c *gin.Context) string {
        return c.GetHeader("X-User-ID")
    },
    SkipFunc: func(c *gin.Context) bool {
        return c.GetHeader("X-Internal") == "true"
    },
})
```

---

### 6. Telemetry 中间件 (`telemetry_middleware.go`)

**配置结构：**
```go
type TelemetryConfig struct {
    Name  *string // 中间件名称
    Order *int    // 执行顺序
}
```

**默认配置：**
- 无需额外配置，通过 DI 注入 TelemetryManager

**使用示例：**
```go
// 使用默认配置
telemetry := litemiddleware.NewTelemetryMiddleware(nil)
```

---

## 依赖注入

所有中间件的依赖注入使用 `inject:""` 标签注入 Manager：

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
```

依赖注入会在容器初始化时自动完成。

---

## 业务系统使用方式

### 方式 1：使用默认配置（快速开发）

```go
// internal/middlewares/cors_middleware.go
func NewCorsMiddleware() ICorsMiddleware {
    return litemiddleware.NewCorsMiddleware(nil)
}
```

### 方式 2：使用自定义配置（生产环境）

```go
// internal/middlewares/cors_middleware.go
func NewProductionCorsMiddleware() ICorsMiddleware {
    allowOrigins := []string{"https://example.com", "https://www.example.com"}
    allowCredentials := true
    maxAge := 12 * time.Hour
    
    return litemiddleware.NewCorsMiddleware(&litemiddleware.CorsConfig{
        AllowOrigins:     &allowOrigins,
        AllowCredentials: &allowCredentials,
        MaxAge:           &maxAge,
    })
}
```

### 方式 3：配置驱动（从配置文件读取）

```go
// internal/middlewares/cors_middleware.go
func NewConfigurableCorsMiddleware(cfg config.CorsConfig) ICorsMiddleware {
    var allowOrigins []string
    if cfg.AllowOrigins != nil {
        allowOrigins = cfg.AllowOrigins
    }
    
    allowCredentials := true
    if cfg.AllowCredentials != nil {
        allowCredentials = *cfg.AllowCredentials
    }
    
    maxAge := 12 * time.Hour
    if cfg.MaxAge != nil {
        maxAge = *cfg.MaxAge
    }
    
    return litemiddleware.NewCorsMiddleware(&litemiddleware.CorsConfig{
        AllowOrigins:     &allowOrigins,
        AllowCredentials: &allowCredentials,
        MaxAge:           &maxAge,
    })
}
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

## 文件清单

### 中间件实现文件
- `cors_middleware.go` - CORS 跨域中间件
- `recovery_middleware.go` - Panic 恢复中间件
- `request_logger_middleware.go` - 请求日志中间件
- `security_headers_middleware.go` - 安全头中间件
- `rate_limiter_middleware.go` - 限流中间件
- `telemetry_middleware.go` - 遥测中间件

### 测试文件
- `cors_middleware_test.go` - CORS 测试
- `rate_limiter_middleware_test.go` - RateLimiter 测试
- `security_headers_middleware_test.go` - SecurityHeaders 测试
- `telemetry_middleware_test.go` - Telemetry 测试
- `example_test.go` - 使用示例

### 示例项目
- `samples/messageboard/internal/middlewares/` - 示例项目中间件封装

---

## 使用建议

### 开发环境
使用默认配置，快速开发：
```go
container.RegisterMiddleware(middlewareContainer, litemiddleware.NewCorsMiddleware(nil))
```

### 生产环境
自定义配置，增强安全性：
```go
allowOrigins := []string{"https://yourdomain.com"}
allowCredentials := true
container.RegisterMiddleware(middlewareContainer, litemiddleware.NewCorsMiddleware(&litemiddleware.CorsConfig{
    AllowOrigins:     &allowOrigins,
    AllowCredentials: &allowCredentials,
}))
```

### 配置文件驱动
建议将中间件配置纳入配置文件管理，支持不同环境使用不同配置：
```yaml
cors:
  allow_origins:
    - https://example.com
    - https://app.example.com
  allow_credentials: true
  max_age: 12h

request_logger:
  enable: true
  log_body: false
  max_body_size: 2048
  skip_paths:
    - /health
    - /metrics
```

### 性能优化
- 关闭 RequestLogger 的 Body 记录（大文件场景）
- 关闭 Recovery 的堆栈打印（生产环境）
- 合理设置 RateLimiter 限制

---

## 总结

中间件配置设计特性：
1. ✅ 所有配置属性都使用指针类型（可选配置）
2. ✅ 未配置的属性自动使用默认值
3. ✅ 支持依赖注入机制
4. ✅ CORS 支持灵活的跨域配置
5. ✅ RequestLogger 支持性能优化配置
6. ✅ 所有中间件都有完整的单元测试

业务系统可以灵活地根据环境（开发/测试/生产）配置不同的中间件参数，支持任意属性组合。
