# Component 组件

提供开箱即用的内置组件，包括控制器、中间件和服务，基于 5 层分层架构和依赖注入机制设计。

## 特性

- **内置控制器** - 健康检查、指标统计、性能分析、静态资源等
- **内置中间件** - 跨域处理、安全头、限流、遥测、请求日志等
- **内置服务** - HTML 模板渲染
- **依赖注入** - 支持通过 `inject:""` 标签注入 Manager
- **生命周期管理** - 实现 OnStart/OnStop 钩子方法
- **配置灵活** - 中间件支持可选配置和自定义执行顺序

## 快速开始

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
recovery := litemiddleware.NewRecoveryMiddleware(nil)
cors := litemiddleware.NewCorsMiddleware(nil)
reqLogger := litemiddleware.NewRequestLoggerMiddleware(nil)
security := litemiddleware.NewSecurityHeadersMiddleware(nil)
limiter := litemiddleware.NewRateLimiterMiddleware(nil)

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

serviceContainer.RegisterService(htmlService)
```

## litecontroller（控制器组件）

| 控制器 | 路由 | 功能 |
|--------|------|------|
| HealthController | `/health` | 检查所有 Manager 的健康状态 |
| MetricsController | `/metrics` | 返回服务器运行指标和组件数量 |
| PprofIndexController | `/debug/pprof` | 性能分析首页 |
| PprofHeapController | `/debug/pprof/heap` | 堆内存分析 |
| PprofGoroutineController | `/debug/pprof/goroutine` | 协程分析 |
| PprofAllocsController | `/debug/pprof/allocs` | 内存分配分析 |
| PprofBlockController | `/debug/pprof/block` | 阻塞分析 |
| PprofMutexController | `/debug/pprof/mutex` | 锁竞争分析 |
| PprofProfileController | `/debug/pprof/profile` | CPU 采样 |
| PprofTraceController | `/debug/pprof/trace` | 协程追踪 |
| ResourceStaticController | 自定义 | 静态文件服务 |
| ResourceHTMLController | - | HTML 模板渲染 |

**依赖注入：**
```go
type HealthController struct {
    ManagerContainer common.IBaseManager      `inject:""`
    LoggerMgr        loggermgr.ILoggerManager `inject:""`
}
```

## litemiddleware（中间件组件）

| 中间件 | 默认 Order | 功能 | 依赖 |
|--------|-----------|------|------|
| Recovery | 0 | Panic 恢复 | LoggerMgr |
| RequestLogger | 50 | 请求日志 | LoggerMgr |
| CORS | 100 | 跨域处理 | - |
| SecurityHeaders | 150 | 安全头 | - |
| RateLimiter | 200 | 限流 | LimiterMgr, LoggerMgr |
| Telemetry | 250 | 遥测 | TelemetryManager |

### 可选配置

配置结构体使用指针类型，支持只配置需要修改的属性：

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
```

### 中间件执行顺序

预定义的执行顺序（按 Order 值从小到大）：
- 0: Recovery（最先执行）
- 50: RequestLogger
- 100: CORS
- 150: SecurityHeaders
- 200: RateLimiter
- 250: Telemetry

业务自定义中间件建议从 Order 350 开始。

## liteservice（服务组件）

### HTMLTemplateService

HTML 模板渲染服务，支持生命周期管理。

```go
htmlService := liteservice.NewHTMLTemplateService("templates/*")
serviceContainer.RegisterService(htmlService)

// 在控制器中使用
ctx.HTML(200, "index.html", gin.H{"title": "Welcome"})
```

**依赖注入：**
```go
type HTMLTemplateService struct {
    LoggerMgr loggermgr.ILoggerManager `inject:""`
}
```

**生命周期方法：**
- `OnStart()` - 启动时加载模板
- `OnStop()` - 停止时清理资源

## 在 messageboard 中的使用示例

messageboard 项目通过封装的方式使用组件：

```go
// 封装中间件（internal/middlewares/xxx.go）
type IRecoveryMiddleware interface {
    common.IBaseMiddleware
}

func NewRecoveryMiddleware() IRecoveryMiddleware {
    return litemiddleware.NewRecoveryMiddlewareWithDefaults()
}

// 封装控制器（internal/controllers/xxx.go）
type sysHealthControllerImpl struct {
    componentController litecontroller.IHealthController
    LoggerMgr          loggermgr.ILoggerManager `inject:""`
}

func NewSysHealthController() ISysHealthController {
    return &sysHealthControllerImpl{
        componentController: litecontroller.NewHealthController(),
    }
}
```

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
```

## 相关文档

- [AGENTS.md](../AGENTS.md) - 项目开发指南
- [manager/README.md](../manager/README.md) - Manager 组件说明
- [container/README.md](../container/README.md) - 依赖注入容器说明
- [docs/SOP-package-document.md](../docs/SOP-package-document.md) - 文档撰写规范
