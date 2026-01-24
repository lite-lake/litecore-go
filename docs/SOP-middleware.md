# 中间件开发规范 (SOP-Middleware)

## 概述

本文档规范了 litecore-go 中间件的开发流程、使用方式和最佳实践。中间件是 HTTP 请求处理链中的重要组成部分，必须遵循统一的规范以确保系统的一致性和可维护性。

### 中间件作用

中间件在 HTTP 请求到达 Controller 之前和响应返回之后执行，用于处理横切关注点，如认证、日志、限流、安全等。

### 组件位置

- **系统中间件**：`component/litemiddleware` 包，提供开箱即用的通用中间件实现
- **业务中间件**：`internal/middlewares` 包，实现业务特定的中间件逻辑

### 执行流程

```
请求 → Recovery → RequestLogger → CORS → SecurityHeaders → RateLimiter → Telemetry → Auth → [业务中间件] → Controller → 响应
```

---

## 规范

### 1. 接口定义

所有中间件必须实现 `common.IBaseMiddleware` 接口：

```go
type IBaseMiddleware interface {
    MiddlewareName() string        // 返回中间件名称
    Order() int                    // 返回执行顺序
    Wrapper() gin.HandlerFunc      // 返回 Gin 中间件函数
    OnStart() error                // 服务器启动时调用
    OnStop() error                 // 服务器停止时调用
}
```

### 2. 命名规范

- **接口名称**：`I` 前缀 + 中间件名 + `Middleware`，例如 `IAuthMiddleware`
- **结构体名称**：小写开头的 camelCase，例如 `authMiddleware`
- **工厂函数名称**：`New` + 中间件名 + `Middleware`，例如 `NewAuthMiddleware`
- **配置结构体**：`XxxConfig`，例如 `RateLimiterConfig`
- **默认配置函数**：`DefaultXxxConfig()`，例如 `DefaultRateLimiterConfig()`
- **默认实例函数**：`NewXxxMiddlewareWithDefaults()`，例如 `NewRateLimiterMiddlewareWithDefaults()`

### 3. 基本结构模板

#### 3.1 自定义业务中间件

```go
package middlewares

import (
    "github.com/gin-gonic/gin"
    "github.com/lite-lake/litecore-go/common"
)

// IMyMiddleware 自定义中间件接口
type IMyMiddleware interface {
    common.IBaseMiddleware
}

type myMiddleware struct {
    order int
}

// NewMyMiddleware 创建中间件
func NewMyMiddleware() IMyMiddleware {
    return &myMiddleware{
        order: 350,
    }
}

func (m *myMiddleware) MiddlewareName() string {
    return "MyMiddleware"
}

func (m *myMiddleware) Order() int {
    return m.order
}

func (m *myMiddleware) Wrapper() gin.HandlerFunc {
    return func(c *gin.Context) {
        c.Next()
    }
}

func (m *myMiddleware) OnStart() error {
    return nil
}

func (m *myMiddleware) OnStop() error {
    return nil
}

var _ IMyMiddleware = (*myMiddleware)(nil)
```

#### 3.2 封装系统中间件

```go
package middlewares

import (
    "github.com/lite-lake/litecore-go/common"
    litemiddleware "github.com/lite-lake/litecore-go/component/litemiddleware"
)

// ICorsMiddleware CORS 跨域中间件接口
type ICorsMiddleware interface {
    common.IBaseMiddleware
}

// NewCorsMiddleware 使用默认配置创建 CORS 中间件
func NewCorsMiddleware() ICorsMiddleware {
    return litemiddleware.NewCorsMiddlewareWithDefaults()
}
```

### 4. 依赖注入规范

中间件可以依赖以下组件：

- **Manager 组件**：通过 `inject:""` 标签自动注入
- **Service 组件**：通过 `inject:""` 标签自动注入

```go
type authMiddleware struct {
    order       int
    AuthService services.IAuthService `inject:""`
    LoggerMgr   loggermgr.ILoggerManager `inject:""`
}
```

### 5. 配置规范

系统中间件支持通过配置对象进行自定义配置。配置对象使用**指针类型字段**，支持可选配置和默认值机制。

#### 5.1 配置结构特征

```go
type MiddlewareConfig struct {
    Name  *string  // 中间件名称（可选）
    Order *int     // 执行顺序（可选）
    // ... 其他配置字段
}
```

#### 5.2 创建中间件的方式

**方式 1：使用默认配置**

```go
import litemiddleware "github.com/lite-lake/litecore-go/component/litemiddleware"

cors := litemiddleware.NewCorsMiddlewareWithDefaults()
rateLimiter := litemiddleware.NewRateLimiterMiddlewareWithDefaults()
```

**方式 2：自定义配置（部分字段）**

```go
limit := 100
window := time.Minute
cfg := &litemiddleware.RateLimiterConfig{
    Limit:  &limit,
    Window: &window,
}
rateLimiter := litemiddleware.NewRateLimiterMiddleware(cfg)
```

**方式 3：自定义配置（包含 Name 和 Order）**

```go
name := "MyRateLimiter"
order := 250
limit := 100
window := time.Minute
cfg := &litemiddleware.RateLimiterConfig{
    Name:   &name,
    Order:  &order,
    Limit:  &limit,
    Window: &window,
}
rateLimiter := litemiddleware.NewRateLimiterMiddleware(cfg)
```

### 6. Order 分配规范

#### 6.1 预定义 Order 范围

所有预定义的 Order 常量位于 `component/litemiddleware/constants.go`：

```go
// 系统中间件（0-300）
OrderRecovery        = 0   // panic 恢复（最先执行）
OrderRequestLogger   = 50  // 请求日志
OrderCORS            = 100 // CORS 跨域
OrderSecurityHeaders = 150 // 安全头
OrderRateLimiter     = 200 // 限流（认证前执行）
OrderTelemetry       = 250 // 遥测
OrderAuth            = 300 // 认证

// 预留空间用于业务中间件：350, 400, 450...
```

#### 6.2 Order 选择原则

1. **基础中间件**（0-300）：系统必备，不应修改
2. **业务中间件**（350+）：自定义中间件，从 350 开始
3. **认证相关**：通常在 300-400 范围
4. **日志相关**：根据执行时间选择，越早越好
5. **性能敏感**：尽量靠前，尽早拒绝无效请求
6. **限流中间件**：通常在 200（认证前），避免无认证请求耗尽配额

### 7. 错误处理规范

```go
func (m *myMiddleware) Wrapper() gin.HandlerFunc {
    return func(c *gin.Context) {
        if err := m.checkSomething(); err != nil {
            m.LoggerMgr.Ins().Warn("检查失败", "error", err)
            c.JSON(http.StatusBadRequest, gin.H{
                "error": "检查失败",
                "code":  "VALIDATION_ERROR",
            })
            c.Abort()
            return
        }
        c.Next()
    }
}
```

---

## 注意点

### 1. 中间件依赖

#### 限流中间件需要 LimiterManager

限流中间件需要依赖 `LimiterManager`，确保在配置中启用限流管理器：

```yaml
limiter:
  driver: "memory"  # 或 "redis"
  memory_config:
    max_backups: 1000
```

#### 遥测中间件需要 TelemetryManager

遥测中间件需要依赖 `TelemetryManager`，需要在配置中启用：

```yaml
telemetry:
  enabled: true
  service_name: "my-service"
```

### 2. 日志使用规范

中间件中必须使用依赖注入的 `LoggerManager`，禁止使用标准库 log 或 fmt.Printf：

```go
type myMiddleware struct {
    LoggerMgr loggermgr.ILoggerManager `inject:""`
}

func (m *myMiddleware) Wrapper() gin.HandlerFunc {
    return func(c *gin.Context) {
        if m.LoggerMgr != nil {
            m.LoggerMgr.Ins().Info("处理请求")
        }
        c.Next()
    }
}
```

### 3. 上下文数据传递

使用 `c.Set()` 设置，`c.Get()` 获取：

```go
// 在认证中间件中设置
c.Set("user_id", user.ID)

// 在后续中间件或控制器中获取
if userID, exists := c.Get("user_id"); exists {
    uid := userID.(string)
}
```

### 4. 跳过某些路由

在 Wrapper 中检查路由：

```go
func (m *myMiddleware) Wrapper() gin.HandlerFunc {
    return func(c *gin.Context) {
        if c.Request.URL.Path == "/health" {
            c.Next()
            return
        }
        // 正常处理
    }
}
```

### 5. 恢复 Panic

Recovery 中间件（Order=0）会自动捕获所有 panic，但业务中间件不应依赖这个机制：

```go
func (m *myMiddleware) Wrapper() gin.HandlerFunc {
    return func(c *gin.Context) {
        defer func() {
            if err := recover(); err != nil {
                m.LoggerMgr.Ins().Error("中间件 panic", "error", err)
                c.JSON(500, gin.H{"error": "内部错误"})
                c.Abort()
            }
        }()
        // 正常处理
    }
}
```

### 6. 响应头处理

响应头必须在调用 `c.Next()` 之前设置：

```go
func (m *myMiddleware) Wrapper() gin.HandlerFunc {
    return func(c *gin.Context) {
        c.Header("X-Custom-Header", "value")
        c.Next()
        // 后置处理
    }
}
```

### 7. 请求 Body 读取

读取请求 Body 后需要重新设置：

```go
var bodyBytes []byte
if c.Request.Body != nil {
    bodyBytes, _ = io.ReadAll(c.Request.Body)
    c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
}
```

---

## 完整示例

### 示例 1：认证中间件（业务层）

```go
package middlewares

import (
    "strings"

    "github.com/gin-gonic/gin"
    "github.com/lite-lake/litecore-go/common"
    "github.com/lite-lake/litecore-go/samples/messageboard/internal/services"
)

// IAuthMiddleware 认证中间件接口
type IAuthMiddleware interface {
    common.IBaseMiddleware
}

type authMiddleware struct {
    AuthService services.IAuthService `inject:""`
}

// NewAuthMiddleware 创建认证中间件实例
func NewAuthMiddleware() IAuthMiddleware {
    return &authMiddleware{}
}

// MiddlewareName 返回中间件名称
func (m *authMiddleware) MiddlewareName() string {
    return "AuthMiddleware"
}

// Order 返回中间件执行顺序
func (m *authMiddleware) Order() int {
    return 300
}

// Wrapper 返回中间件处理函数
func (m *authMiddleware) Wrapper() gin.HandlerFunc {
    return func(c *gin.Context) {
        if !strings.HasPrefix(c.Request.URL.Path, "/api/admin") {
            c.Next()
            return
        }

        if c.Request.URL.Path == "/api/admin/login" {
            c.Next()
            return
        }

        authHeader := c.GetHeader("Authorization")
        if authHeader == "" {
            c.JSON(common.HTTPStatusUnauthorized, gin.H{
                "code":    common.HTTPStatusUnauthorized,
                "message": "未提供认证令牌",
            })
            c.Abort()
            return
        }

        parts := strings.SplitN(authHeader, " ", 2)
        if len(parts) != 2 || parts[0] != "Bearer" {
            c.JSON(common.HTTPStatusUnauthorized, gin.H{
                "code":    common.HTTPStatusUnauthorized,
                "message": "认证令牌格式错误",
            })
            c.Abort()
            return
        }

        token := parts[1]
        session, err := m.AuthService.ValidateToken(token)
        if err != nil {
            c.JSON(common.HTTPStatusUnauthorized, gin.H{
                "code":    common.HTTPStatusUnauthorized,
                "message": "认证令牌无效或已过期",
            })
            c.Abort()
            return
        }

        c.Set("admin_session", session)
        c.Next()
    }
}

func (m *authMiddleware) OnStart() error {
    return nil
}

func (m *authMiddleware) OnStop() error {
    return nil
}

var _ IAuthMiddleware = (*authMiddleware)(nil)
```

### 示例 2：限流中间件（封装系统中间件）

```go
package middlewares

import (
    "time"

    "github.com/lite-lake/litecore-go/common"
    litemiddleware "github.com/lite-lake/litecore-go/component/litemiddleware"
)

type IRateLimiterMiddleware interface {
    common.IBaseMiddleware
}

// NewRateLimiterMiddleware 创建限流中间件
// 配置：每 IP 每分钟最多 100 次请求
func NewRateLimiterMiddleware() IRateLimiterMiddleware {
    limit := 100
    window := time.Minute
    keyPrefix := "ip"
    return litemiddleware.NewRateLimiterMiddleware(&litemiddleware.RateLimiterConfig{
        Limit:     &limit,
        Window:    &window,
        KeyPrefix: &keyPrefix,
    })
}
```

### 示例 3：自定义限流策略

```go
package middlewares

import (
    "time"

    "github.com/lite-lake/litecore-go/common"
    litemiddleware "github.com/lite-lake/litecore-go/component/litemiddleware"
)

// NewLoginRateLimiter 登录接口限流（按 IP）
func NewLoginRateLimiter() common.IBaseMiddleware {
    name := "LoginRateLimiter"
    order := 200
    limit := 5
    window := time.Minute
    keyPrefix := "login"
    return litemiddleware.NewRateLimiterMiddleware(&litemiddleware.RateLimiterConfig{
        Name:      &name,
        Order:     &order,
        Limit:     &limit,
        Window:    &window,
        KeyPrefix: &keyPrefix,
        KeyFunc: func(c *gin.Context) string {
            return c.ClientIP()
        },
        SkipFunc: func(c *gin.Context) bool {
            return c.Request.URL.Path != "/api/login"
        },
    })
}

// NewUserRateLimiter API 通用限流（按用户）
func NewUserRateLimiter() common.IBaseMiddleware {
    name := "UserRateLimiter"
    order := 200
    limit := 100
    window := time.Minute
    keyPrefix := "user"
    return litemiddleware.NewRateLimiterMiddleware(&litemiddleware.RateLimiterConfig{
        Name:      &name,
        Order:     &order,
        Limit:     &limit,
        Window:    &window,
        KeyPrefix: &keyPrefix,
        KeyFunc: func(c *gin.Context) string {
            if userID, exists := c.Get("user_id"); exists {
                if uid, ok := userID.(string); ok {
                    return uid
                }
            }
            return c.ClientIP()
        },
    })
}
```

### 示例 4：完整请求日志中间件配置

```go
package middlewares

import (
    "github.com/lite-lake/litecore-go/common"
    litemiddleware "github.com/lite-lake/litecore-go/component/litemiddleware"
)

// IRequestLoggerMiddleware 请求日志中间件接口
type IRequestLoggerMiddleware interface {
    common.IBaseMiddleware
}

// NewRequestLoggerMiddleware 创建请求日志中间件
func NewRequestLoggerMiddleware() IRequestLoggerMiddleware {
    enable := true
    logBody := true
    maxBodySize := 4096
    skipPaths := []string{"/health", "/metrics"}
    logHeaders := []string{"User-Agent", "Content-Type"}
    successLogLevel := "info"
    cfg := &litemiddleware.RequestLoggerConfig{
        Enable:          &enable,
        LogBody:         &logBody,
        MaxBodySize:     &maxBodySize,
        SkipPaths:       &skipPaths,
        LogHeaders:      &logHeaders,
        SuccessLogLevel: &successLogLevel,
    }
    return litemiddleware.NewRequestLoggerMiddleware(cfg)
}
```

### 示例 5：中间件容器注册

中间件容器由代码生成器自动生成，位于 `internal/application/middleware_container.go`：

```go
// Code generated by litecore/cli. DO NOT EDIT.
package application

import (
    "github.com/lite-lake/litecore-go/container"
    middlewares "github.com/lite-lake/litecore-go/samples/messageboard/internal/middlewares"
)

// InitMiddlewareContainer 初始化中间件容器
func InitMiddlewareContainer(serviceContainer *container.ServiceContainer) *container.MiddlewareContainer {
    middlewareContainer := container.NewMiddlewareContainer(serviceContainer)
    container.RegisterMiddleware[middlewares.IAuthMiddleware](middlewareContainer, middlewares.NewAuthMiddleware())
    container.RegisterMiddleware[middlewares.ICorsMiddleware](middlewareContainer, middlewares.NewCorsMiddleware())
    container.RegisterMiddleware[middlewares.IRateLimiterMiddleware](middlewareContainer, middlewares.NewRateLimiterMiddleware())
    container.RegisterMiddleware[middlewares.IRecoveryMiddleware](middlewareContainer, middlewares.NewRecoveryMiddleware())
    container.RegisterMiddleware[middlewares.IRequestLoggerMiddleware](middlewareContainer, middlewares.NewRequestLoggerMiddleware())
    container.RegisterMiddleware[middlewares.ISecurityHeadersMiddleware](middlewareContainer, middlewares.NewSecurityHeadersMiddleware())
    container.RegisterMiddleware[middlewares.ITelemetryMiddleware](middlewareContainer, middlewares.NewTelemetryMiddleware())
    return middlewareContainer
}
```

---

## 参考资源

- [AGENTS.md](./AGENTS.md) - 整体开发规范
- [SOP-build-business-application.md](./SOP-build-business-application.md) - 业务应用构建指南
- [component/litemiddleware](../component/litemiddleware) - 系统中间件实现
- [common/base_middleware.go](../common/base_middleware.go) - 中间件接口定义
