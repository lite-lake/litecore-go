# Component 组件

提供开箱即用的内置组件，基于 5 层分层架构和依赖注入机制设计。

## 特性

- **开箱即用** - 内置常用控制器、中间件和服务
- **统一接口** - 遵循 IBaseController/IBaseMiddleware/IBaseService 标准
- **依赖注入** - 支持通过 `inject:""` 标签注入 Manager 和其他组件
- **灵活配置** - 中间件支持可选配置和自定义执行顺序
- **生命周期管理** - 实现 OnStart/OnStop 钩子方法
- **易于扩展** - 可通过封装模式自定义组件行为

## 组件概述

Component 组件位于 `component/` 目录下，提供三层内置组件：

```
component/
├── litecontroller/   # 控制器组件
├── litemiddleware/   # 中间件组件
└── liteservice/      # 服务组件
```

### 设计原则

1. **统一命名规范**
   - 包名统一使用 `lite` 前缀
   - 具体组件使用大驼峰命名
   - 接口使用 `I` 前缀

2. **标准接口实现**
   - 所有控制器实现 `common.IBaseController`
   - 所有中间件实现 `common.IBaseMiddleware`
   - 所有服务实现 `common.IBaseService`

3. **配置灵活**
   - 使用指针类型实现可选配置
   - 提供 `DefaultXxxConfig()` 设置默认值
   - 提供 `NewXxxWithDefaults()` 快速创建

4. **生命周期管理**
   - 实现 `OnStart()` 和 `OnStop()` 钩子
   - 支持资源初始化和清理

## 可用组件列表

### litecontroller（控制器组件）

| 组件 | 路由 | 功能 | 依赖 |
|------|------|------|------|
| HealthController | `/health [GET]` | 健康检查，检测所有 Manager 状态 | ManagerContainer, LoggerMgr |
| MetricsController | `/metrics [GET]` | 返回服务器运行指标和组件数量 | ManagerContainer, ServiceContainer, LoggerMgr |
| PprofIndexController | `/debug/pprof [GET]` | 性能分析首页 | LoggerMgr |
| PprofHeapController | `/debug/pprof/heap [GET]` | 堆内存分析 | LoggerMgr |
| PprofGoroutineController | `/debug/pprof/goroutine [GET]` | 协程分析 | LoggerMgr |
| PprofAllocsController | `/debug/pprof/allocs [GET]` | 内存分配分析 | LoggerMgr |
| PprofBlockController | `/debug/pprof/block [GET]` | 阻塞分析 | LoggerMgr |
| PprofMutexController | `/debug/pprof/mutex [GET]` | 锁竞争分析 | LoggerMgr |
| PprofProfileController | `/debug/pprof/profile [GET]` | CPU 采样 | LoggerMgr |
| PprofTraceController | `/debug/pprof/trace [GET]` | 协程追踪 | LoggerMgr |
| ResourceStaticController | 自定义 | 静态文件服务 | LoggerMgr |
| ResourceHTMLController | - | HTML 模板渲染（已废弃，使用 liteservice） | LoggerMgr |

### litemiddleware（中间件组件）

| 组件 | 默认 Order | 功能 | 依赖 |
|------|-----------|------|------|
| Recovery | 0 | Panic 恢复，打印堆栈 | LoggerMgr |
| RequestLogger | 50 | 请求日志，记录请求/响应信息 | LoggerMgr |
| CORS | 100 | 跨域处理，设置 CORS 头 | - |
| SecurityHeaders | 150 | 安全头，X-Frame-Options 等 | - |
| RateLimiter | 200 | 限流，基于时间窗口 | LimiterMgr, LoggerMgr |
| Telemetry | 250 | 遥测，OpenTelemetry 集成 | TelemetryManager |

### liteservice（服务组件）

| 组件 | 功能 | 依赖 |
|------|------|------|
| HTMLTemplateService | HTML 模板渲染服务 | LoggerMgr |

## 统一接口规范

### 命名规范

```
组件包：lite<type>（litecontroller、litemiddleware、liteservice）
具体组件：<Name><Type>（HealthController、RecoveryMiddleware、HTMLTemplateService）
接口：I<Name><Type>（IHealthController、IRecoveryMiddleware、IHTMLTemplateService）
配置结构：<Name>Config（RecoveryConfig、CorsConfig）
工厂函数：New<Name><Type>（NewHealthController、NewRecoveryMiddleware）
默认配置：Default<Name>Config（DefaultRecoveryConfig）
默认创建：New<Name><Type>WithDefaults（NewRecoveryMiddlewareWithDefaults）
```

### 控制器接口

```go
type IBaseController interface {
    ControllerName() string  // 控制器名称
    GetRouter() string      // 路由定义，如 "/health [GET]"
    Handle(ctx *gin.Context) // 请求处理
}
```

### 中间件接口

```go
type IBaseMiddleware interface {
    MiddlewareName() string          // 中间件名称
    Order() int                      // 执行顺序
    Wrapper() gin.HandlerFunc        // Gin 中间件函数
    OnStart() error                  // 启动钩子
    OnStop() error                   // 停止钩子
}
```

### 服务接口

```go
type IBaseService interface {
    ServiceName() string  // 服务名称
    OnStart() error       // 启动钩子
    OnStop() error        // 停止钩子
}
```

## 配置支持（Name/Order）

### 中间件配置结构

所有中间件配置结构体均包含 `Name` 和 `Order` 字段：

```go
type <Component>Config struct {
    Name  *string  // 中间件名称
    Order *int     // 执行顺序
    // ... 其他配置字段
}
```

### 可选配置模式

使用指针类型实现可选配置：

```go
// 自定义部分配置
order := 300
name := "CustomRecovery"
limiter := litemiddleware.NewRateLimiterMiddleware(&litemiddleware.RateLimiterConfig{
    Name:  &name,
    Order: &order,
    // 其他字段使用默认值
})

// 使用默认配置
recovery := litemiddleware.NewRecoveryMiddlewareWithDefaults()
// 或
recovery := litemiddleware.NewRecoveryMiddleware(nil)
```

### 中间件执行顺序

预定义的执行顺序（按 Order 值从小到大）：

| Order | 中间件 | 说明 |
|-------|--------|------|
| 0 | Recovery | Panic 恢复（最先执行） |
| 50 | RequestLogger | 请求日志 |
| 100 | CORS | 跨域处理 |
| 150 | SecurityHeaders | 安全头 |
| 200 | RateLimiter | 限流 |
| 250 | Telemetry | 遥测 |
| 300+ | 自定义 | 业务中间件（建议从 350 开始） |

## 依赖注入方式

### 注入 Manager

组件通过 `inject:""` 标签自动注入 Manager：

```go
type HealthController struct {
    ManagerContainer common.IBaseManager      `inject:""`
    LoggerMgr        loggermgr.ILoggerManager `inject:""`
}

type RecoveryMiddleware struct {
    LoggerMgr loggermgr.ILoggerManager `inject:""`
}

type RateLimiterMiddleware struct {
    LimiterMgr limitermgr.ILimiterManager `inject:""`
    LoggerMgr  loggermgr.ILoggerManager   `inject:""`
}
```

### 注入其他组件

```go
type CustomController struct {
    LoggerMgr        loggermgr.ILoggerManager `inject:""`
    DBManager        databasemgr.IDatabaseManager `inject:""`
    CacheMgr         cachemgr.ICacheManager `inject:""`
    LimiterMgr       limitermgr.ILimiterManager `inject:""`
    TelemetryManager telemetrymgr.ITelemetryManager `inject:""`
}
```

## 快速开始

### 基本用法

```go
import (
    "github.com/lite-lake/litecore-go/component/litecontroller"
    "github.com/lite-lake/litecore-go/component/litemiddleware"
    "github.com/lite-lake/litecore-go/component/liteservice"
)

// 创建控制器
health := litecontroller.NewHealthController()
metrics := litecontroller.NewMetricsController()
static := litecontroller.NewResourceStaticController("/static", "./static")

// 创建中间件（使用默认配置）
recovery := litemiddleware.NewRecoveryMiddlewareWithDefaults()
cors := litemiddleware.NewCorsMiddlewareWithDefaults()
reqLogger := litemiddleware.NewRequestLoggerMiddlewareWithDefaults()
security := litemiddleware.NewSecurityHeadersMiddlewareWithDefaults()
limiter := litemiddleware.NewRateLimiterMiddlewareWithDefaults()
telemetry := litemiddleware.NewTelemetryMiddlewareWithDefaults()

// 创建服务
htmlService := liteservice.NewHTMLTemplateService("templates/*")

// 注册到容器
controllerContainer.RegisterController(health)
controllerContainer.RegisterController(metrics)
controllerContainer.RegisterController(static)

middlewareContainer.RegisterMiddleware(recovery)
middlewareContainer.RegisterMiddleware(cors)
middlewareContainer.RegisterMiddleware(reqLogger)
middlewareContainer.RegisterMiddleware(security)
middlewareContainer.RegisterMiddleware(limiter)
middlewareContainer.RegisterMiddleware(telemetry)

serviceContainer.RegisterService(htmlService)
```

### 自定义配置

```go
// 自定义限流规则
limit := 200
window := time.Minute
keyPrefix := "api"
limiter := litemiddleware.NewRateLimiterMiddleware(&litemiddleware.RateLimiterConfig{
    Limit:     &limit,
    Window:    &window,
    KeyPrefix: &keyPrefix,
    KeyFunc: func(c *gin.Context) string {
        return c.GetHeader("X-User-ID")
    },
})

// 自定义 CORS 配置
allowOrigins := []string{"https://example.com"}
allowCredentials := true
cors := litemiddleware.NewCorsMiddleware(&litemiddleware.CorsConfig{
    AllowOrigins:     &allowOrigins,
    AllowCredentials: &allowCredentials,
})

// 自定义安全头配置
frameOptions := "SAMEORIGIN"
security := litemiddleware.NewSecurityHeadersMiddleware(&litemiddleware.SecurityHeadersConfig{
    FrameOptions: &frameOptions,
})
```

## 在 messageboard 中的使用示例

messageboard 项目通过封装的方式使用组件，实现更好的隔离和扩展性。

### 封装中间件

```go
// internal/middlewares/recovery_middleware.go
package middlewares

import (
    "github.com/lite-lake/litecore-go/common"
    "github.com/lite-lake/litecore-go/component/litemiddleware"
)

type IRecoveryMiddleware interface {
    common.IBaseMiddleware
}

func NewRecoveryMiddleware() IRecoveryMiddleware {
    return litemiddleware.NewRecoveryMiddlewareWithDefaults()
}
```

### 封装控制器

```go
// internal/controllers/sys_health_controller.go
package controllers

import (
    "github.com/lite-lake/litecore-go/component/litecontroller"
    "github.com/lite-lake/litecore-go/manager/loggermgr"
)

type ISysHealthController interface {
    common.IBaseController
}

type sysHealthControllerImpl struct {
    componentController litecontroller.IHealthController
    LoggerMgr          loggermgr.ILoggerManager `inject:""`
}

func NewSysHealthController() ISysHealthController {
    return &sysHealthControllerImpl{
        componentController: litecontroller.NewHealthController(),
    }
}

// 实现接口方法，委托给组件控制器
func (c *sysHealthControllerImpl) ControllerName() string {
    return "SysHealthController"
}

func (c *sysHealthControllerImpl) GetRouter() string {
    return "/health [GET]"
}

func (c *sysHealthControllerImpl) Handle(ctx *gin.Context) {
    c.componentController.Handle(ctx)
}
```

### 注册到容器

```go
// internal/application/middleware_container.go
middlewareContainer := container.NewMiddlewareContainer(serviceContainer)
container.RegisterMiddleware[middlewares.IRecoveryMiddleware](
    middlewareContainer,
    middlewares.NewRecoveryMiddleware(),
)

// internal/application/controller_container.go
controllerContainer := container.NewControllerContainer(serviceContainer)
container.RegisterController[controllers.ISysHealthController](
    controllerContainer,
    controllers.NewSysHealthController(),
)
```

## 各层组件特点

### litecontroller

- **职责**：HTTP 请求处理和响应
- **特点**：
  - 实现 `Handle(ctx *gin.Context)` 方法处理请求
  - 通过 `GetRouter()` 定义路由
  - 可注入 ManagerContainer 访问所有 Manager
  - 可注入 ServiceContainer 访问所有 Service
  - 支持静态文件服务和健康检查等常用功能

### litemiddleware

- **职责**：请求预处理和后处理
- **特点**：
  - 实现 `Wrapper()` 返回 `gin.HandlerFunc`
  - 通过 `Order()` 控制执行顺序
  - 支持可选配置（指针类型）
  - 实现生命周期钩子（OnStart/OnStop）
  - 统一的默认配置机制

### liteservice

- **职责**：业务逻辑和数据处理支持
- **特点**：
  - 实现生命周期钩子管理资源
  - 在 `OnStart()` 中初始化资源
  - 在 `OnStop()` 中清理资源
  - 可被 Controller 和 Middleware 注入使用

## 最佳实践

1. **使用默认配置**
   - 大多数场景下使用 `NewXxxWithDefaults()` 即可
   - 仅在需要自定义时才传入配置

2. **遵循命名规范**
   - 自定义组件遵循相同的命名规范
   - 接口使用 `I` 前缀
   - 配置使用 `<Component>Config` 格式

3. **正确设置 Order**
   - 业务中间件从 350 开始
   - 不要使用 0-250 之间的预定义 Order
   - 保持中间件执行的逻辑顺序

4. **封装组件**
   - 在项目中封装组件以隔离依赖
   - 定义本地接口方便扩展
   - 提供统一的创建函数

5. **依赖注入**
   - 使用 `inject:""` 标签声明依赖
   - 依赖由 Engine 自动注入
   - 避免在组件内部创建 Manager 实例

## 测试

```bash
# 测试所有组件
go test ./component/... -v

# 测试控制器
go test ./component/litecontroller/... -v

# 测试中间件
go test ./component/litemiddleware/... -v

# 测试服务
go test ./component/liteservice/... -v

# 测试覆盖率
go test -cover ./component/...
```

## 相关文档

- [AGENTS.md](../AGENTS.md) - 项目开发指南
- [docs/SOP-package-document.md](../docs/SOP-package-document.md) - 文档撰写规范
- [manager/README.md](../manager/README.md) - Manager 组件说明
- [container/README.md](../container/README.md) - 依赖注入容器说明
- [common/README.md](../common/README.md) - 公共接口说明
