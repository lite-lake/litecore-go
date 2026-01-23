# 中间件开发规范 (SOP-Middleware)

## 概述

本文档规范了 litecore-go 中间件的开发流程、使用方式和最佳实践。中间件是 HTTP 请求处理链中的重要组成部分，必须遵循统一的规范以确保系统的一致性和可维护性。

---

## 一、中间件开发规范

### 1.1 接口定义

所有中间件必须实现 `common.IBaseMiddleware` 接口：

```go
type IBaseMiddleware interface {
    MiddlewareName() string  // 返回中间件名称
    Order() int              // 返回执行顺序
    Wrapper() gin.HandlerFunc // 返回 Gin 中间件函数
    OnStart() error          // 服务器启动时调用
    OnStop() error           // 服务器停止时调用
}
```

### 1.2 命名规范

- **接口名称**：`I` 前缀 + 中间件名 + `Middleware`，例如 `IAuthMiddleware`
- **结构体名称**：小写开头的 camelCase，例如 `authMiddleware`
- **工厂函数名称**：`New` + 中间件名 + `Middleware`，例如 `NewAuthMiddleware`

### 1.3 基本结构模板

```go
package middlewares

import (
    "github.com/gin-gonic/gin"
    "github.com/lite-lake/litecore-go/common"
    middlewarepkg "github.com/lite-lake/litecore-go/component/litemiddleware"
)

// IMyMiddleware 自定义中间件接口
type IMyMiddleware interface {
    common.IBaseMiddleware
}

type myMiddleware struct {
    order int
    // 依赖项
    LoggerMgr loggermgr.ILoggerManager `inject:""`
    Service    services.IMyService     `inject:""`
}

// NewMyMiddleware 创建中间件
func NewMyMiddleware() IMyMiddleware {
    return &myMiddleware{
        order: 350, // 自定义中间件建议从 350 开始
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
        // 前置处理
        c.Next()
        // 后置处理
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

### 1.4 依赖注入规范

中间件可以依赖以下组件：

- **Manager 组件**：通过 `inject:""` 标签自动注入
- **Service 组件**：通过 `inject:""` 标签自动注入
- **其他中间件**：不推荐互相依赖，使用 Order 控制执行顺序

**示例：**

```go
type authMiddleware struct {
    order       int
    AuthService services.IAuthService `inject:""`
    LoggerMgr   loggermgr.ILoggerManager `inject:""`
}
```

### 1.5 错误处理规范

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

### 1.6 封装 component 中间件

如果只需要简单地使用 component 中的中间件，可以采用简洁封装方式：

```go
package middlewares

import (
    "github.com/lite-lake/litecore-go/common"
    "github.com/lite-lake/litecore-go/component/litemiddleware"
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

如果需要自定义配置：

```go
package middlewares

import (
    "time"

    "github.com/lite-lake/litecore-go/common"
    "github.com/lite-lake/litecore-go/component/litemiddleware"
)

type IRateLimiterMiddleware interface {
    common.IBaseMiddleware
}

func NewRateLimiterMiddleware() IRateLimiterMiddleware {
    return litemiddleware.NewRateLimiterMiddleware(&litemiddleware.RateLimiterConfig{
        Limit:     100,
        Window:    time.Minute,
        KeyPrefix: "ip",
    })
}
```

---

## 二、Order 分配指南

### 2.1 预定义 Order 范围

```go
// 系统中间件（0-300）
OrderRecovery         = 0     // panic 恢复（最先执行）
OrderRequestLogger    = 50    // 请求日志
OrderCORS             = 100   // CORS 跨域
OrderSecurityHeaders  = 150   // 安全头
OrderRateLimiter      = 200   // 限流（认证前）
OrderTelemetry        = 250   // 遥测
OrderAuth             = 300   // 认证

// 业务中间件（350-1000，建议按 50 递增）
OrderContext          = 350   // 上下文设置
OrderBusinessLogic    = 400   // 业务逻辑校验
OrderPermission       = 450   // 权限检查
OrderAuditLog         = 500   // 审计日志
```

### 2.2 Order 选择原则

1. **基础中间件**（0-300）：系统必备，不应修改
2. **业务中间件**（350+）：自定义中间件，从 350 开始
3. **认证相关**：通常在 300-400 范围
4. **日志相关**：根据执行时间选择，越早越好
5. **性能敏感**：尽量靠前，尽早拒绝无效请求

### 2.3 执行顺序图

```
请求进入
  ↓
Recovery (0) ──────────── 捕获 panic
  ↓
RequestLogger (50) ────── 记录请求开始
  ↓
CORS (100) ────────────── 处理跨域
  ↓
SecurityHeaders (150) ── 添加安全头
  ↓
RateLimiter (200) ─────── 限流检查
  ↓
Telemetry (250) ───────── 遥测追踪
  ↓
Auth (300) ────────────── 认证鉴权
  ↓
[业务中间件 350+] ──────── 自定义逻辑
  ↓
[Controller]
  ↓
[业务中间件后置处理]
  ↓
RequestLogger (50) ────── 记录请求结束
  ↓
响应返回
```

### 2.4 常见场景 Order 分配

#### 场景 1：需要在认证前执行
```go
// 用户登录接口限流（不依赖认证）
Order = 250  // 在 Auth (300) 之前
```

#### 场景 2：需要在认证后执行
```go
// 权限检查（依赖用户信息）
Order = 350  // 在 Auth (300) 之后
```

#### 场景 3：需要访问数据库
```go
// 黑名单检查
Order = 400  // 较后执行，确保数据库已启动
```

---

## 三、依赖注入示例

### 3.1 注入 Manager

```go
type myMiddleware struct {
    LoggerMgr   loggermgr.ILoggerManager   `inject:""`
    ConfigMgr   configmgr.IConfigManager   `inject:""`
    LimiterMgr  limitermgr.ILimiterManager  `inject:""`
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

### 3.2 注入 Service

```go
type authMiddleware struct {
    order       int
    AuthService services.IAuthService `inject:""`
}

func (m *authMiddleware) Wrapper() gin.HandlerFunc {
    return func(c *gin.Context) {
        token := c.GetHeader("Authorization")
        user, err := m.AuthService.ValidateToken(token)
        if err != nil {
            c.JSON(401, gin.H{"error": "unauthorized"})
            c.Abort()
            return
        }
        c.Set("user", user)
        c.Next()
    }
}
```

### 3.3 注入多个依赖

```go
type auditMiddleware struct {
    order         int
    LoggerMgr     loggermgr.ILoggerManager `inject:""`
    AuditService  services.IAuditService   `inject:""`
    ConfigMgr     configmgr.IConfigManager `inject:""`
}

func (m *auditMiddleware) OnStart() error {
    // 在启动时初始化配置
    cfg, _ := m.ConfigMgr.Get("audit")
    m.auditConfig = cfg
    return nil
}
```

### 3.4 完整示例：自定义限流中间件

```go
package middlewares

import (
    "fmt"
    "net/http"
    "time"

    "github.com/gin-gonic/gin"
    "github.com/lite-lake/litecore-go/common"
    "github.com/lite-lake/litecore-go/server/builtin/manager/loggermgr"
)

type ICustomRateLimiterMiddleware interface {
    common.IBaseMiddleware
}

type customRateLimiterMiddleware struct {
    order      int
    LoggerMgr  loggermgr.ILoggerManager `inject:""`
    limit      int
    window     time.Duration
}

func NewCustomRateLimiterMiddleware(limit int, window time.Duration) ICustomRateLimiterMiddleware {
    return &customRateLimiterMiddleware{
        order:  350,  // 自定义中间件从 350 开始
        limit:  limit,
        window: window,
    }
}

func (m *customRateLimiterMiddleware) MiddlewareName() string {
    return "CustomRateLimiterMiddleware"
}

func (m *customRateLimiterMiddleware) Order() int {
    return m.order
}

func (m *customRateLimiterMiddleware) Wrapper() gin.HandlerFunc {
    return func(c *gin.Context) {
        // 简化示例：实际应使用限流器 Manager
        key := c.ClientIP()
        if m.shouldRateLimit(key) {
            m.LoggerMgr.Ins().Warn("请求被限流", "key", key)
            c.JSON(http.StatusTooManyRequests, gin.H{
                "error": "请求过于频繁",
                "code":  "RATE_LIMIT_EXCEEDED",
            })
            c.Abort()
            return
        }
        c.Next()
    }
}

func (m *customRateLimiterMiddleware) shouldRateLimit(key string) bool {
    return false
}

func (m *customRateLimiterMiddleware) OnStart() error {
    m.LoggerMgr.Ins().Info("限流中间件启动", "limit", m.limit, "window", m.window)
    return nil
}

func (m *customRateLimiterMiddleware) OnStop() error {
    m.LoggerMgr.Ins().Info("限流中间件停止")
    return nil
}

var _ ICustomRateLimiterMiddleware = (*customRateLimiterMiddleware)(nil)
```

---

## 四、注册中间件

### 4.1 中间件容器（自动生成）

中间件容器由代码生成器自动生成，位于 `internal/application/middleware_container.go`，无需手动编辑：

```go
// Code generated by litecore/cli. DO NOT EDIT.
package application

import (
    "github.com/lite-lake/litecore-go/container"
    middlewares "github.com/lite-lake/litecore-go/samples/myapp/internal/middlewares"
)

// InitMiddlewareContainer 初始化中间件容器
func InitMiddlewareContainer(serviceContainer *container.ServiceContainer) *container.MiddlewareContainer {
    middlewareContainer := container.NewMiddlewareContainer(serviceContainer)
    container.RegisterMiddleware[middlewares.IAuthMiddleware](middlewareContainer, middlewares.NewAuthMiddleware())
    container.RegisterMiddleware[middlewares.ICorsMiddleware](middlewareContainer, middlewares.NewCorsMiddleware())
    container.RegisterMiddleware[middlewares.IRateLimiterMiddleware](middlewareContainer, middlewares.NewRateLimiterMiddleware())
    return middlewareContainer
}
```

### 4.2 创建引擎（自动生成）

引擎创建函数由代码生成器自动生成，位于 `internal/application/engine.go`：

```go
// Code generated by litecore/cli. DO NOT EDIT.
package application

import (
    "github.com/lite-lake/litecore-go/server"
    "github.com/lite-lake/litecore-go/server/builtin"
)

// NewEngine 创建应用引擎
func NewEngine() (*server.Engine, error) {
    entityContainer := InitEntityContainer()
    repositoryContainer := InitRepositoryContainer(entityContainer)
    serviceContainer := InitServiceContainer(repositoryContainer)
    controllerContainer := InitControllerContainer(serviceContainer)
    middlewareContainer := InitMiddlewareContainer(serviceContainer)

    return server.NewEngine(
        &builtin.Config{
            Driver:   "yaml",
            FilePath: "configs/config.yaml",
        },
        entityContainer,
        repositoryContainer,
        serviceContainer,
        controllerContainer,
        middlewareContainer,
    ), nil
}
```

### 4.3 重新生成中间件容器

新增或修改中间件后，运行代码生成器重新生成容器代码：

```bash
go run ./cmd/generate
```

---

## 五、最佳实践

### 5.1 性能优化

- **尽早拒绝**：在中间件链前端进行简单检查（如格式验证）
- **避免阻塞**：使用异步操作处理耗时任务
- **合理缓存**：缓存重复计算的结果

### 5.2 错误处理

- **统一响应格式**：所有错误返回相同的 JSON 格式
- **记录日志**：所有异常情况都应记录日志
- **优雅降级**：中间件失败时应降级处理，不阻断请求

### 5.3 可观测性

- **请求 ID**：传递请求 ID 以便追踪
- **性能指标**：记录中间件的执行时间
- **健康检查**：提供健康检查接口

### 5.4 安全性

- **敏感信息脱敏**：不在日志中记录密码、token 等
- **最小权限原则**：中间件只获取必要的数据
- **防重放攻击**：关键操作添加防重放机制

---

## 六、常见问题

### Q1：中间件如何访问上下文数据？

使用 `c.Set()` 设置，`c.Get()` 获取：

```go
// 在认证中间件中设置
c.Set("user_id", user.ID)

// 在后续中间件或控制器中获取
if userID, exists := c.Get("user_id"); exists {
    uid := userID.(string)
}
```

### Q2：如何跳过某些路由？

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

### Q3：如何确保中间件顺序？

使用预定义的 Order 常量：

```go
import "github.com/lite-lake/litecore-go/component/litemiddleware"

type myMiddleware struct {
    order int
}

func NewMyMiddleware() IMyMiddleware {
    return &myMiddleware{
        order: litemiddleware.OrderContext,  // 使用预定义常量
    }
}
```

### Q4：中间件如何处理 panic？

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

---

## 七、参考资源

- [AGENTS.md](./AGENTS.md) - 整体开发规范
- [SOP-build-business-application.md](./SOP-build-business-application.md) - 业务应用构建指南
- [component/middleware](../component/litemiddleware) - 系统中间件实现
