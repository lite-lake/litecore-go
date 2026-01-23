# Component 组件

Component 目录提供了开箱即用的内置组件，包括控制器、中间件和服务，基于 5 层分层架构和依赖注入机制设计。

## 目录结构

```
component/
├── litecontroller/      # 控制器组件
│   ├── health_controller.go           # 健康检查控制器
│   ├── metrics_controller.go         # 指标控制器
│   ├── pprof_controllers.go          # pprof 性能分析控制器
│   ├── resource_html_controller.go   # HTML 资源控制器
│   └── resource_static_controller.go # 静态资源控制器
│
├── litemiddleware/     # 中间件组件
│   ├── cors_middleware.go            # CORS 跨域中间件
│   ├── recovery_middleware.go        # Panic 恢复中间件
│   ├── request_logger_middleware.go  # 请求日志中间件
│   ├── security_headers_middleware.go # 安全头中间件
│   ├── rate_limiter_middleware.go    # 限流中间件
│   ├── telemetry_middleware.go       # 遥测中间件
│   └── constants.go                  # 中间件执行顺序常量
│
└── liteservice/       # 服务组件
    └── html_template_service.go      # HTML 模板渲染服务
```

## 组件说明

### litecontroller（控制器组件）

控制器负责处理 HTTP 请求并返回响应。所有控制器都实现 `common.IBaseController` 接口，支持依赖注入。

| 控制器 | 功能 | 路由 | 说明 |
|--------|------|------|------|
| HealthController | 健康检查 | `/health` | 检查所有 Manager 的健康状态 |
| MetricsController | 指标统计 | `/metrics` | 返回服务器运行指标和组件数量 |
| PprofController | 性能分析 | `/debug/pprof/*` | 提供性能分析工具（CPU/内存/协程等） |
| ResourceHTMLController | HTML 资源 | 自定义 | 渲染 HTML 模板页面 |
| ResourceStaticController | 静态资源 | 自定义 | 提供静态文件服务 |

**使用示例：**

```go
// 创建健康检查控制器
health := litecontroller.NewHealthController()

// 注册到容器
controllerContainer.RegisterController(health)
```

---

### litemiddleware（中间件组件）

中间件用于在请求处理前后执行通用逻辑。所有中间件都实现 `common.IBaseMiddleware` 接口，支持通过配置自定义 Name 和 Order，配置属性使用指针类型以支持可选配置。

| 中间件 | 功能 | 默认 Order | 说明 |
|--------|------|-----------|------|
| Recovery | Panic 恢复 | 0 | 捕获 panic 并返回错误响应 |
| RequestLogger | 请求日志 | 50 | 记录 HTTP 请求和响应 |
| CORS | 跨域处理 | 100 | 处理 CORS 预检请求和响应头 |
| SecurityHeaders | 安全头 | 150 | 添加 HTTP 安全响应头 |
| RateLimiter | 限流 | 200 | 基于滑动窗口的请求限流 |
| Telemetry | 遥测 | 250 | 集成 OpenTelemetry 追踪 |

**核心特性：**

1. **可选配置（指针类型）**：可以只配置需要修改的属性，未配置的属性自动使用默认值
2. **自定义 Name 和 Order**：支持通过配置自定义中间件名称和执行顺序
3. **依赖注入**：所有中间件支持通过 `inject:""` 标签注入 Manager
4. **灵活组合**：支持任意属性组合配置

**使用示例：**

```go
// 使用默认配置
cors := litemiddleware.NewCorsMiddleware(nil)

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

// 注册到容器
middlewareContainer.RegisterMiddleware(cors)
middlewareContainer.RegisterMiddleware(limiter)
```

详细配置说明请参考 [litemiddleware/README.md](./litemiddleware/README.md)

---

### liteservice（服务组件）

服务提供业务逻辑和基础设施功能。所有服务都实现 `common.IBaseService` 接口，支持依赖注入和生命周期管理。

| 服务 | 功能 | 说明 |
|------|------|------|
| HTMLTemplateService | HTML 模板渲染 | 加载和渲染 HTML 模板 |

**使用示例：**

```go
// 创建 HTML 模板服务
htmlService := liteservice.NewHTMLTemplateService("templates/*")

// 注册到容器
serviceContainer.RegisterService(htmlService)

// 在控制器中使用
ctx.HTML(200, "index.html", gin.H{"title": "Welcome"})
```

---

## 依赖注入

所有组件都支持通过 `inject:""` 标签进行依赖注入：

```go
type HealthController struct {
    ManagerContainer common.IBaseManager      `inject:""`
    LoggerMgr        loggermgr.ILoggerManager `inject:""`
}

type rateLimiterMiddleware struct {
    LimiterMgr limitermgr.ILimiterManager `inject:""`
    LoggerMgr  loggermgr.ILoggerManager   `inject:""`
}
```

依赖注入会在容器初始化时自动完成。

---

## 中间件执行顺序

预定义的中间件执行顺序（按 Order 值从小到大）：

| Order | 中间件 | 说明 |
|-------|--------|------|
| 0 | Recovery | panic 恢复（最先执行） |
| 50 | RequestLogger | 请求日志 |
| 100 | CORS | 跨域处理 |
| 150 | SecurityHeaders | 安全头 |
| 200 | RateLimiter | 限流 |
| 250 | Telemetry | 遥测 |

业务自定义中间件建议从 Order 350 开始。

---

## 架构设计

### 5 层分层架构

```
┌─────────────────────────────────────────────────────────┐
│                   Controller Layer                       │
│              (HTTP 请求处理和响应)                        │
├─────────────────────────────────────────────────────────┤
│                  Middleware Layer                        │
│              (请求预处理和后处理)                         │
├─────────────────────────────────────────────────────────┤
│                   Service Layer                          │
│              (业务逻辑和数据处理)                         │
├─────────────────────────────────────────────────────────┤
│                 Repository Layer                         │
│              (数据访问和持久化)                           │
├─────────────────────────────────────────────────────────┤
│                   Entity Layer                           │
│              (数据模型和领域对象)                         │
└─────────────────────────────────────────────────────────┘
```

### 依赖规则

- Entity（无依赖）
- Manager → Config + 其他 Managers
- Repository → Config + Manager + Entity
- Service → Config + Manager + Repository + 其他 Services
- Controller → Config + Manager + Service
- Middleware → Config + Manager + Service

---

## 快速开始

### 1. 使用默认组件

```go
// 创建控制器
health := litecontroller.NewHealthController()
metrics := litecontroller.NewMetricsController()

// 创建中间件
cors := litemiddleware.NewCorsMiddleware(nil)
recovery := litemiddleware.NewRecoveryMiddleware(nil)
reqLogger := litemiddleware.NewRequestLoggerMiddleware(nil)
security := litemiddleware.NewSecurityHeadersMiddleware(nil)
limiter := litemiddleware.NewRateLimiterMiddleware(nil)

// 创建服务
htmlService := liteservice.NewHTMLTemplateService("templates/*")

// 注册到容器
controllerContainer.RegisterController(health)
controllerContainer.RegisterController(metrics)

middlewareContainer.RegisterMiddleware(cors)
middlewareContainer.RegisterMiddleware(recovery)
middlewareContainer.RegisterMiddleware(reqLogger)
middlewareContainer.RegisterMiddleware(security)
middlewareContainer.RegisterMiddleware(limiter)

serviceContainer.RegisterService(htmlService)
```

### 2. 自定义配置

```go
// 自定义 CORS 配置
allowOrigins := []string{"https://example.com"}
allowCredentials := true
customCors := litemiddleware.NewCorsMiddleware(&litemiddleware.CorsConfig{
    AllowOrigins:     &allowOrigins,
    AllowCredentials: &allowCredentials,
})

// 自定义限流配置
limit := 500
window := time.Minute
keyPrefix := "api_prod"
customLimiter := litemiddleware.NewRateLimiterMiddleware(&litemiddleware.RateLimiterConfig{
    Limit:     &limit,
    Window:    &window,
    KeyPrefix: &keyPrefix,
})
```

### 3. 配置文件驱动

建议将组件配置纳入配置文件管理，支持不同环境使用不同配置：

```yaml
# config.yaml
cors:
  allow_origins:
    - https://example.com
  allow_credentials: true
  max_age: 12h

rate_limiter:
  limit: 500
  window: 1m
  key_prefix: "api_prod"

request_logger:
  enable: true
  log_body: false
  max_body_size: 2048
```

---

## 测试

所有组件都有完整的单元测试，运行测试：

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

---

## 版本历史

### v1.0.0 (2026-01-24)

- **目录重构**：统一包名规范
  - `controller` → `litecontroller`
  - `middleware` → `litemiddleware`
  - `service` → `liteservice`

- **中间件增强**：
  - 所有中间件支持通过配置自定义 Name 和 Order
  - 配置重构为指针类型，支持可选配置
  - 新增 RateLimiter 中间件（限流功能）

- **依赖注入优化**：
  - 统一日志注入机制
  - 支持通过 DI 注入 Manager

---

## 相关文档

- [litemiddleware/README.md](./litemiddleware/README.md) - 中间件详细配置指南
- [AGENTS.md](../AGENTS.md) - 项目开发指南
- [manager/README.md](../manager/README.md) - Manager 组件说明
