# Server

提供统一的 HTTP 服务引擎，支持自动依赖注入、生命周期管理和中间件集成。

## 特性

- **容器管理** - 集成 Manager、Entity、Repository、Service、Controller、Middleware 六层容器
- **内置组件** - Manager 作为内置组件（位于 `manager/` 目录），由引擎自动初始化和注入
- **自动注入** - 按依赖顺序自动处理组件注入（Manager → Entity → Repository → Service → Controller/Middleware）
- **生命周期管理** - 统一管理各层组件的启动和停止，支持健康检查
- **中间件集成** - 自动排序并注册全局中间件到 Gin 引擎，支持通过配置自定义名称和执行顺序
- **路由管理** - 自动注册控制器路由，通过 BaseController 统一管理
- **优雅关闭** - 支持信号处理，超时控制的安全关闭机制
- **启动日志** - 支持异步启动日志，记录各阶段启动状态和耗时

## 模块结构

```
litecore-go/
├── server/                    # Server 模块
│   ├── builtin.go             # 内置组件初始化（Manager 自动注入）
│   ├── config.go              # 服务器配置
│   ├── engine.go              # 引擎核心
│   ├── lifecycle.go           # 生命周期管理
│   ├── middleware.go          # 中间件管理
│   ├── router.go              # 路由管理
│   ├── async_startup_logger.go # 异步启动日志
│   └── startup_phase.go       # 启动阶段定义
│
├── manager/                   # Manager 组件（独立模块）
│   ├── configmgr/             # 配置管理器
│   ├── databasemgr/           # 数据库管理器
│   ├── cachemgr/              # 缓存管理器
│   ├── loggermgr/             # 日志管理器
│   ├── lockmgr/               # 锁管理器
│   ├── limitermgr/            # 限流管理器
│   ├── telemetrymgr/          # 遥测管理器
│   └── mqmgr/                 # 消息队列管理器
│
└── component/
    └── litemiddleware/        # 内置中间件组件
        ├── rate_limiter_middleware.go   # 限流中间件
        ├── cors_middleware.go           # CORS 中间件
        ├── recovery_middleware.go       # 恢复中间件
        ├── request_logger_middleware.go # 请求日志中间件
        ├── security_headers_middleware.go # 安全头中间件
        └── telemetry_middleware.go      # 遥测中间件
```

## 快速开始

### 方式一：使用 CLI 生成的应用引擎（推荐）

```go
package main

import (
    "log"

    "github.com/lite-lake/litecore-go/samples/messageboard/internal/application"
)

func main() {
    // 创建应用引擎（由 CLI 工具自动生成）
    engine, err := application.NewEngine()
    if err != nil {
        log.Fatalf("Failed to create engine: %v", err)
    }

    // 一键启动
    if err := engine.Run(); err != nil {
        log.Fatalf("Engine run failed: %v", err)
    }
}
```

### 方式二：手动创建引擎

```go
package main

import (
    "github.com/lite-lake/litecore-go/server"
    "github.com/lite-lake/litecore-go/container"
)

func main() {
    // 创建容器
    entityContainer := container.NewEntityContainer()
    repositoryContainer := container.NewRepositoryContainer(entityContainer)
    serviceContainer := container.NewServiceContainer(repositoryContainer)
    controllerContainer := container.NewControllerContainer(serviceContainer)
    middlewareContainer := container.NewMiddlewareContainer(serviceContainer)

    // 注册其他组件（实体、仓储、服务、控制器、中间件）...

    // Manager 由引擎自动初始化和注入（位于 manager/ 目录）

    // 创建并启动引擎
    engine := server.NewEngine(
        &server.BuiltinConfig{
            Driver:   "yaml",
            FilePath: "config.yaml",
        },
        entityContainer,
        repositoryContainer,
        serviceContainer,
        controllerContainer,
        middlewareContainer,
    )
    if err := engine.Run(); err != nil {
        panic(err)
    }
}
```

## 核心功能

### 创建引擎

使用 CLI 生成的 `NewEngine()` 创建服务引擎（推荐）：

```go
// 由 CLI 工具自动生成
engine, err := application.NewEngine()
if err != nil {
    log.Fatalf("Failed to create engine: %v", err)
}
```

或手动创建（传入配置和所有容器实例）：

```go
engine := server.NewEngine(
    &builtin.Config{
        Driver:   "yaml",
        FilePath: "config.yaml",
    },
    entityContainer,
    repositoryContainer,
    serviceContainer,
    controllerContainer,
    middlewareContainer,
)
```

### 启动服务

提供两种启动方式：

**方式一：一步启动（推荐）**

```go
// Run() = Initialize() + Start() + WaitForShutdown()
if err := engine.Run(); err != nil {
    log.Fatalf("Engine run failed: %v", err)
}
```

**启动日志：**

Engine 支持启动日志功能，记录各阶段的启动状态和耗时：

| 启动阶段 | 说明 |
|---------|------|
| 配置加载 | 加载配置文件 |
| 管理器初始化 | 初始化所有 Manager |
| 依赖注入 | 执行各层组件的依赖注入 |
| 路由注册 | 注册中间件和控制器路由 |
| 组件启动 | 启动各层组件 |
| 运行中 | HTTP 服务器运行 |
| 关闭中 | 优雅关闭各层组件 |

日志格式（Gin 风格）：
```
2026-01-24 15:04:05.123 | INFO  | 开始初始化内置组件
2026-01-24 15:04:05.456 | INFO  | 初始化完成: ConfigManager
2026-01-24 15:04:05.789 | INFO  | 管理器初始化完成 | count=8 | duration=1.2s
2026-01-24 15:04:06.123 | INFO  | 注册中间件 | middleware=RecoveryMiddleware | type=全局
2026-01-24 15:04:06.456 | INFO  | 中间件注册完成 | middleware_count=6
2026-01-24 15:04:06.789 | INFO  | HTTP 服务器启动成功 | address=0.0.0.0:8080
```

启动日志配置（通过 BuiltinConfig）：
```go
type StartupLogConfig struct {
    Enabled bool // 是否启用启动日志
    Async   bool // 是否异步输出（默认 true）
    Buffer  int  // 缓冲区大小（默认 100）
}
```

**方式二：分步启动（需要自定义初始化时）**

```go
// 1. 初始化（依赖注入、创建 Gin 引擎、注册中间件和路由）
if err := engine.Initialize(); err != nil {
    log.Fatalf("Failed to initialize engine: %v", err)
}

// 2. 启动服务（启动各层组件和 HTTP 服务器）
if err := engine.Start(); err != nil {
    log.Fatalf("Failed to start engine: %v", err)
}

// 3. 等待关闭信号
engine.WaitForShutdown()
```

### 生命周期管理

Engine 按以下顺序管理组件生命周期：

**启动顺序：**
1. Manager 层（内置组件，由 `server.Initialize()` 自动初始化）
   - ConfigManager
   - TelemetryManager
   - LoggerManager
   - DatabaseManager
   - CacheManager
   - LockManager
   - LimiterManager
   - MQManager
2. Entity 层（按注册顺序）
3. Repository 层（按注册顺序）
4. Service 层（按注册顺序）
5. Controller 和 Middleware 层（按注册顺序）
6. HTTP 服务器

**停止顺序（反转启动顺序）：**
1. HTTP 服务器（优雅关闭）
2. Controller 和 Middleware 层（反转顺序）
3. Service 层（反转顺序）
4. Repository 层（反转顺序）
5. Entity 层（反转顺序）
6. Manager 层（反转顺序，自动清理）

```go
// 手动停止
if err := engine.Stop(); err != nil {
    log.Printf("Stop error: %v", err)
}
```

### 依赖注入

Engine 在初始化时自动按以下顺序执行依赖注入：

1. **Manager 层**（由 `server.Initialize()` 自动初始化并注入）
   - ConfigManager（最先初始化，其他 Manager 依赖它）
   - TelemetryManager（依赖 ConfigManager）
   - LoggerManager（依赖 ConfigManager、TelemetryManager）
   - DatabaseManager（依赖 ConfigManager）
   - CacheManager（依赖 ConfigManager）
   - LockManager（依赖 ConfigManager）
   - LimiterManager（依赖 ConfigManager）
   - MQManager（依赖 ConfigManager）
2. **Entity 层**（无依赖）
3. **Repository 层**（依赖 Manager、Entity）
4. **Service 层**（依赖 Manager、Repository 和同层 Service）
5. **Controller 层**（依赖 Manager、Service）
6. **Middleware 层**（依赖 Manager、Service）

各层组件通过 `inject:""` 标签声明依赖，Manager 由引擎自动注入：

```go
type UserServiceImpl struct {
    // 内置组件（引擎自动注入，来自 manager 包）
    Config     configmgr.IConfigManager      `inject:""`
    DBManager  databasemgr.IDatabaseManager  `inject:""`
    LoggerMgr  loggermgr.ILoggerManager     `inject:""`
    LockMgr    lockmgr.ILockManager         `inject:""`
    LimiterMgr limitermgr.ILimiterManager    `inject:""`

    // 业务依赖
    UserRepo   repository.IUserRepository   `inject:""`
}
```

### 中间件管理

中间件按 `Order()` 排序后自动注册到 Gin 引擎。内置中间件支持通过配置自定义名称和执行顺序：

```go
type AuthMiddleware struct {
    Name  string
    Order int // 越小越先执行
}

func (m *AuthMiddleware) MiddlewareName() string {
    return m.Name
}

func (m *AuthMiddleware) Order() int {
    return m.Order
}

func (m *AuthMiddleware) Wrapper() gin.HandlerFunc {
    return gin.HandlerFunc(func(c *gin.Context) {
        // 中间件逻辑
        c.Next()
    })
}
```

#### 使用内置中间件

内置中间件位于 `component/litemiddleware` 包，支持灵活配置：

```go
import "github.com/lite-lake/litecore-go/component/litemiddleware"

// 使用默认配置（按 IP 限流）
rateLimiter := litemiddleware.NewRateLimiterMiddleware(nil)

// 自定义配置
limit := 200
window := time.Minute
order := 90
name := "APILimiter"
customLimiter := litemiddleware.NewRateLimiterMiddleware(&litemiddleware.RateLimiterConfig{
    Name:      &name,
    Order:     &order,
    Limit:     &limit,
    Window:    &window,
})
```

#### 预定义的中间件 Order

| 中间件 | Order | 说明 |
|--------|-------|------|
| Recovery | 0 | panic 恢复（最先执行） |
| RequestLogger | 50 | 请求日志 |
| CORS | 100 | 跨域处理 |
| SecurityHeaders | 150 | 安全头 |
| RateLimiter | 200 | 限流（认证前执行） |
| Telemetry | 250 | 遥测 |
| Auth | 300 | 认证（预留） |

业务自定义中间件建议从 Order 350 开始。

### 路由管理

**通过控制器定义路由**

所有路由都必须通过控制器（Controller）的 `GetRouter()` 方法定义，格式为：`/path [METHOD]`

```go
type UserController struct {
    Config    common.BaseConfigProvider  `inject:""`
    UserService service.IUserService `inject:""`
}

func (ctrl *UserController) GetRouter() string {
    return "/users [GET]"
}

func (ctrl *UserController) Handle(c *gin.Context) {
    users, err := ctrl.UserService.List()
    if err != nil {
        c.JSON(500, gin.H{"error": err.Error()})
        return
    }
    c.JSON(200, users)
}
```

来自 messageboard 的实际控制器示例：

```go
type MessageListController struct {
    Config  config.BaseConfigProvider `inject:""`
    MsgSvc  service.IMessageService   `inject:""`
}

func (ctrl *MessageListController) GetRouter() string {
    return "/api/messages [GET]"
}

func (ctrl *MessageListController) Handle(c *gin.Context) {
    messages, err := ctrl.MsgSvc.List()
    if err != nil {
        c.JSON(500, gin.H{"error": err.Error()})
        return
    }
    c.JSON(200, gin.H{"data": messages})
}
```

### 锁和限流管理器使用

#### LockMgr（锁管理器）

LockMgr 提供分布式锁功能，支持 Redis 和 Memory 两种实现：

**配置文件示例：**

```yaml
lock:
  driver: "memory"              # redis, memory
  redis_config:
    host: "localhost"
    port: 6379
    password: ""
    db: 0
    max_idle_conns: 10
    max_open_conns: 100
    conn_max_lifetime: 30s
  memory_config:
    max_backups: 1000
```

**Service 层使用示例：**

```go
type OrderService struct {
    LockMgr    lockmgr.ILockManager `inject:""`
    OrderRepo  repository.IOrderRepository `inject:""`
}

func (s *OrderService) CreateOrder(order *entities.Order) error {
    ctx := context.Background()
    lockKey := fmt.Sprintf("order:user:%d", order.UserID)

    if err := s.LockMgr.Lock(ctx, lockKey, 10*time.Second); err != nil {
        return fmt.Errorf("获取锁失败: %w", err)
    }
    defer s.LockMgr.Unlock(ctx, lockKey)

    return s.OrderRepo.Create(order)
}

func (s *OrderService) TryProcessOrder(orderID uint) (bool, error) {
    ctx := context.Background()
    lockKey := fmt.Sprintf("order:process:%d", orderID)

    locked, err := s.LockMgr.TryLock(ctx, lockKey, 5*time.Second)
    if err != nil {
        return false, err
    }
    if !locked {
        return false, nil
    }
    defer s.LockMgr.Unlock(ctx, lockKey)

    return true, s.OrderRepo.Process(orderID)
}
```

#### LimiterMgr（限流管理器）

LimiterMgr 提供限流功能，支持 Redis 和 Memory 两种实现：

**配置文件示例：**

```yaml
limiter:
  driver: "memory"              # redis, memory
  redis_config:
    host: "localhost"
    port: 6379
    password: ""
    db: 0
    max_idle_conns: 10
    max_open_conns: 100
    conn_max_lifetime: 30s
  memory_config:
    max_backups: 1000
```

**Service 层使用示例：**

```go
type APIService struct {
    LimiterMgr limitermgr.ILimiterManager `inject:""`
}

func (s *APIService) CallAPI(apiKey string) error {
    ctx := context.Background()
    key := fmt.Sprintf("api:%s", apiKey)

    allowed, err := s.LimiterMgr.Allow(ctx, key, 100, time.Minute)
    if err != nil {
        return fmt.Errorf("限流检查失败: %w", err)
    }
    if !allowed {
        remaining, _ := s.LimiterMgr.GetRemaining(ctx, key, 100, time.Minute)
        return fmt.Errorf("请求过于频繁，剩余次数: %d", remaining)
    }

    return s.doAPICall(apiKey)
}

func (s *APIService) GetUserQuota(userID string) (int, error) {
    ctx := context.Background()
    key := fmt.Sprintf("user:%s:quota", userID)

    remaining, err := s.LimiterMgr.GetRemaining(ctx, key, 1000, time.Hour)
    if err != nil {
        return 0, err
    }
    return remaining, nil
}
```

### 限流器中间件集成

框架提供了 `rate_limiter_middleware` 中间件（位于 `component/litemiddleware` 包），支持灵活的限流配置。

#### 基本使用

```go
import "github.com/lite-lake/litecore-go/component/litemiddleware"

// 使用默认配置（按 IP 限流，100 次/分钟）
rateLimiter := litemiddleware.NewRateLimiterMiddleware(nil)

// 自定义配置
limit := 200
window := time.Minute
order := 90
name := "APILimiter"
keyPrefix := "api"
customLimiter := litemiddleware.NewRateLimiterMiddleware(&litemiddleware.RateLimiterConfig{
    Name:      &name,
    Order:     &order,
    Limit:     &limit,
    Window:    &window,
    KeyPrefix: &keyPrefix,
    KeyFunc: func(c *gin.Context) string {
        // 按 Header 中的 X-User-ID 限流
        return c.GetHeader("X-User-ID")
    },
    SkipFunc: func(c *gin.Context) bool {
        // 跳过健康检查
        return c.Request.URL.Path == "/health"
    },
})
```

**配置文件示例：**

```yaml
app:
  name: "myapp"
  version: "1.0.0"

server:
  host: "0.0.0.0"
  port: 8080
  mode: "debug"

database:
  driver: "sqlite"
  sqlite_config:
    dsn: "./data/myapp.db"

cache:
  driver: "memory"

logger:
  driver: "zap"
  zap_config:
    console_enabled: true
    console_config:
      level: "info"

lock:
  driver: "memory"
  memory_config:
    max_backups: 1000

limiter:
  driver: "memory"
  memory_config:
    max_backups: 1000
```

#### 中间件实现

**按 IP 限流：**

```go
package middlewares

import (
    "time"

    "github.com/lite-lake/litecore-go/common"
    "github.com/lite-lake/litecore-go/component/middleware"
)

type IRateLimiterMiddleware interface {
    common.IBaseMiddleware
}

type rateLimiterMiddleware struct {
    inner common.IBaseMiddleware
}

func NewRateLimiterMiddleware() IRateLimiterMiddleware {
    return &rateLimiterMiddleware{
        inner: middleware.NewRateLimiterByIP(100, time.Minute),
    }
}

func (m *rateLimiterMiddleware) MiddlewareName() string { return "RateLimiterMiddleware" }
func (m *rateLimiterMiddleware) Order() int              { return 90 }
func (m *rateLimiterMiddleware) Wrapper() gin.HandlerFunc {
    return m.inner.Wrapper()
}
func (m *rateLimiterMiddleware) OnStart() error { return nil }
func (m *rateLimiterMiddleware) OnStop() error  { return nil }

var _ IRateLimiterMiddleware = (*rateLimiterMiddleware)(nil)
```

#### 配置参数

| 参数 | 类型 | 说明 | 默认值 |
|------|------|------|--------|
| Name | *string | 中间件名称 | "RateLimiterMiddleware" |
| Order | *int | 执行顺序 | 200 |
| Limit | *int | 时间窗口内最大请求数 | 100 |
| Window | *time.Duration | 时间窗口大小 | time.Minute |
| KeyFunc | KeyFunc | 自定义 key 生成函数 | 按 IP 生成 |
| SkipFunc | SkipFunc | 跳过限流的条件 | 无 |
| KeyPrefix | *string | key 前缀 | "rate_limit" |

#### 响应头说明

限流器中间件会在响应中添加以下头：

| 响应头 | 说明 |
|--------|------|
| X-RateLimit-Limit | 时间窗口内的最大请求数 |
| X-RateLimit-Remaining | 剩余可用次数 |
| X-RateLimit-Reset | 限流窗口重置时间（秒） |
| Retry-After | 被限流时建议的等待时间（秒） |

#### 中间件 Order 建议

限流器中间件建议使用 Order = 200，在认证中间件之前执行：

```go
const (
    OrderRecovery        = 0   // panic 恢复
    OrderRequestLogger   = 50  // 请求日志
    OrderCORS            = 100 // CORS
    OrderSecurityHeaders = 150 // 安全头
    OrderRateLimiter     = 200 // 限流
    OrderTelemetry       = 250 // 遥测
    OrderAuth            = 300 // 认证
)
```


## API

### Engine

服务引擎主结构，提供完整的服务管理功能。

#### 构造函数

```go
func NewEngine(
    builtinConfig *BuiltinConfig,
    entity *container.EntityContainer,
    repository *container.RepositoryContainer,
    service *container.ServiceContainer,
    controller *container.ControllerContainer,
    middleware *container.MiddlewareContainer,
) *Engine
```

**BuiltinConfig 配置：**

```go
type BuiltinConfig struct {
    Driver   string // 配置驱动类型（如：yaml、json 等）
    FilePath string // 配置文件路径
}
```

#### 生命周期方法

| 方法 | 说明 |
|------|------|
| `Initialize() error` | 初始化引擎（依赖注入、创建 Gin 引擎、注册中间件和路由） |
| `Start() error` | 启动引擎（启动各层组件和 HTTP 服务器） |
| `Stop() error` | 停止引擎（优雅关闭） |
| `Run() error` | 一键启动（Initialize + Start + WaitForShutdown） |
| `WaitForShutdown()` | 等待关闭信号 |

## 最佳实践

### 1. 容器注册顺序

使用 CLI 工具生成时，容器注册顺序已自动处理。手动创建时，请按照依赖顺序创建和注册容器：

```go
// 1. 创建容器（按依赖顺序）
entityContainer := container.NewEntityContainer()
repositoryContainer := container.NewRepositoryContainer(entityContainer)
serviceContainer := container.NewServiceContainer(repositoryContainer)
controllerContainer := container.NewControllerContainer(serviceContainer)
middlewareContainer := container.NewMiddlewareContainer(serviceContainer)

// 2. 注册组件（按层级顺序）
// Entity 层
// Repository 层
// Service 层
// Controller 层
// Middleware 层

// Config 和 Manager 由引擎自动初始化
```

推荐使用 CLI 工具生成容器初始化代码：

```go
// 由 CLI 工具自动生成 application/engine.go
func NewEngine() (*server.Engine, error) {
    entityContainer := InitEntityContainer()
    repositoryContainer := InitRepositoryContainer(entityContainer)
    serviceContainer := InitServiceContainer(repositoryContainer)
    controllerContainer := InitControllerContainer(serviceContainer)
    middlewareContainer := InitMiddlewareContainer(serviceContainer)

    // 自动注入依赖
    if err := repositoryContainer.InjectAll(); err != nil {
        return nil, err
    }
    if err := serviceContainer.InjectAll(); err != nil {
        return nil, err
    }
    if err := controllerContainer.InjectAll(); err != nil {
        return nil, err
    }
    if err := middlewareContainer.InjectAll(); err != nil {
        return nil, err
    }

    return server.NewEngine(
        &server.BuiltinConfig{
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

### 2. 中间件排序

为中间件设置合理的 Order 值，控制执行顺序：

```go
const (
    OrderRecovery        = 0   // panic 恢复
    OrderRequestLogger   = 50  // 日志记录
    OrderCORS            = 100 // 跨域处理
    OrderSecurityHeaders = 150 // 安全头
    OrderRateLimiter     = 200 // 限流
    OrderTelemetry       = 250 // 遥测监控
    OrderAuth            = 300 // 认证
)
```

来自 messageboard 的实际中间件示例：

```go
// Recovery 中间件
func (m *RecoveryMiddleware) Order() int {
    return litemiddleware.OrderRecovery
}

// 请求日志中间件
func (m *RequestLoggerMiddleware) Order() int {
    return litemiddleware.OrderRequestLogger
}

// CORS 中间件
func (m *CorsMiddleware) Order() int {
    return litemiddleware.OrderCORS
}

// 遥测中间件
func (m *TelemetryMiddleware) Order() int {
    return litemiddleware.OrderTelemetry
}

// 安全头中间件
func (m *SecurityHeadersMiddleware) Order() int {
    return litemiddleware.OrderSecurityHeaders
}

// 限流中间件
func (m *RateLimiterMiddleware) Order() int {
    return litemiddleware.OrderRateLimiter
}

// 认证中间件
func (m *AuthMiddleware) Order() int {
    return litemiddleware.OrderAuth
}
```

**限流器中间件 Order 建议值：** 200（在认证之前执行，避免无效请求消耗认证资源）

#### 通过配置自定义 Order

所有内置中间件都支持通过配置自定义 Order：

```go
// 修改限流中间件的执行顺序为 90
order := 90
rateLimiter := litemiddleware.NewRateLimiterMiddleware(&litemiddleware.RateLimiterConfig{
    Order: &order,
})
```

### 3. 路由命名规范

控制器路由使用 OpenAPI 风格：`/path [METHOD]`

```go
// 正确
"/messages [GET]"
"/messages/:id [GET]"
"/messages [POST]"
"/messages/:id [PUT]"
"/messages/:id [DELETE]"

// 错误（会被忽略）
""                // 空字符串
"messages"        // 缺少方法
"/messages[GET]"  // 缺少空格
```

来自 messageboard 的实际控制器路由示例：

```go
// 页面控制器
func (ctrl *PageHomeController) GetRouter() string {
    return "/ [GET]"
}

func (ctrl *PageAdminController) GetRouter() string {
    return "/admin [GET]"
}

// API 控制器
func (ctrl *MessageListController) GetRouter() string {
    return "/api/messages [GET]"
}

func (ctrl *MessageCreateController) GetRouter() string {
    return "/api/messages [POST]"
}

func (ctrl *MessageDeleteController) GetRouter() string {
    return "/api/messages/:id [DELETE]"
}

// 静态资源控制器
func (ctrl *ResStaticController) GetRouter() string {
    return "/static/*filepath [GET]"
}

// 系统健康检查控制器
func (ctrl *SysHealthController) GetRouter() string {
    return "/health [GET]"
}

func (ctrl *SysMetricsController) GetRouter() string {
    return "/metrics [GET]"
}
```

### 4. 错误处理

启动和停止时的错误应妥善处理：

```go
if err := engine.Run(); err != nil {
    log.Fatalf("Engine run failed: %v", err)
}
```

来自 messageboard 的实际错误处理示例：

```go
func main() {
    engine, err := application.NewEngine()
    if err != nil {
        log.Fatalf("Failed to create engine: %v", err)
    }

    if err := engine.Initialize(); err != nil {
        log.Fatalf("Failed to initialize engine: %v", err)
    }

    if err := engine.Start(); err != nil {
        log.Fatalf("Failed to start engine: %v", err)
    }

    engine.WaitForShutdown()
}
```

## 信号处理

Engine 自动处理以下信号，触发优雅关闭：

- `SIGINT`（Ctrl+C）
- `SIGTERM`
- `SIGQUIT`

关闭流程：

1. 捕获信号
2. HTTP 服务器优雅关闭（等待现有请求完成，超时时间由 ShutdownTimeout 配置）
3. 停止 Service 层（反转注册顺序）
4. 停止 Repository 层（反转注册顺序）
5. 停止 Manager 层（反转注册顺序）

## 注意事项

1. **依赖注入**：确保组件使用 `inject:""` 标签声明依赖
2. **Manager 引用**：Manager 组件位于 `manager/` 目录，导入路径为 `github.com/lite-lake/litecore-go/manager/xxxmgr`
3. **线程安全**：Engine 使用读写锁保护内部状态，外部访问需加锁
4. **重入保护**：Start 方法已实现重入保护，重复调用返回错误
5. **路由定义**：所有路由必须通过 BaseController 的 `GetRouter()` 方法定义
6. **中间件顺序**：中间件按 Order 升序排序，越小的值越先执行
7. **中间件配置**：内置中间件支持通过配置自定义名称和执行顺序

来自 messageboard 的实际注意事项示例：

**依赖注入声明**

```go
type MessageServiceImpl struct {
    Config   configmgr.IConfigManager   `inject:""`
    DBMgr    databasemgr.IDatabaseManager `inject:""`
    LoggerMgr loggermgr.ILoggerManager   `inject:""`
    Messages repository.IMessageRepository `inject:""`
}

type AuthServiceImpl struct {
    Config  configmgr.IConfigManager `inject:""`
    DBMgr   databasemgr.IDatabaseManager `inject:""`
    LoggerMgr loggermgr.ILoggerManager   `inject:""`
    Sessions repository.ISessionRepository `inject:""`
}
```

**路由定义规范**

所有路由都应通过控制器定义：

```go
type HealthCheckController struct {
    Config  configmgr.IConfigManager `inject:""`
}

func (ctrl *HealthCheckController) GetRouter() string {
    return "/health [GET]"
}

func (ctrl *HealthCheckController) Handle(c *gin.Context) {
    c.JSON(200, gin.H{"status": "ok"})
}
```

**内置中间件使用**

内置中间件位于 `component/litemiddleware` 包：

```go
import "github.com/lite-lake/litecore-go/component/litemiddleware"

// 使用默认配置
rateLimiter := litemiddleware.NewRateLimiterMiddleware(nil)

// 自定义配置
order := 90
name := "APILimiter"
limit := 200
window := time.Minute
customLimiter := litemiddleware.NewRateLimiterMiddleware(&litemiddleware.RateLimiterConfig{
    Name:   &name,
    Order:  &order,
    Limit:  &limit,
    Window: &window,
})
```
