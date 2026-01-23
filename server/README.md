# Server

提供统一的 HTTP 服务引擎，支持自动依赖注入、生命周期管理和中间件集成。

## 特性

- **容器管理** - 集成 Entity、Repository、Service、Controller、Middleware 五层容器
- **内置组件** - Config 和 Manager 作为内置组件，由引擎自动初始化和注入
- **自动注入** - 按依赖顺序自动处理组件注入（Entity → Repository → Service → Controller/Middleware）
- **生命周期管理** - 统一管理各层组件的启动和停止，支持健康检查
- **中间件集成** - 自动排序并注册全局中间件到 Gin 引擎
- **路由管理** - 自动注册控制器路由，通过 BaseController 统一管理
- **优雅关闭** - 支持信号处理，超时控制的安全关闭机制

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
    "github.com/lite-lake/litecore-go/server/builtin"
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

    // Config 和 Manager 由引擎自动初始化和注入

    // 创建并启动引擎
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
1. Config 和 Manager（内置组件，自动初始化）
   - ConfigManager
   - DatabaseManager
   - CacheManager
   - LoggerManager
   - LockManager
   - LimiterManager
   - TelemetryManager
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
6. Manager 和 Config（内置组件，自动清理）

```go
// 手动停止
if err := engine.Stop(); err != nil {
    log.Printf("Stop error: %v", err)
}
```

### 依赖注入

Engine 在初始化时自动按以下顺序执行依赖注入：

1. **Config 和 Manager**（内置组件，自动初始化）
   - ConfigManager
   - DatabaseManager
   - CacheManager
   - LoggerManager
   - LockManager
   - LimiterManager
   - TelemetryManager
2. **Entity 层**（无依赖）
3. **Repository 层**（依赖 Config、Manager、Entity）
4. **Service 层**（依赖 Config、Manager、Repository 和同层 Service）
5. **Controller 层**（依赖 Config、Manager、Service）
6. **Middleware 层**（依赖 Config、Manager、Service）

各层组件通过 `inject:""` 标签声明依赖，Config 和 Manager 由引擎自动注入：

```go
type UserServiceImpl struct {
    // 内置组件（引擎自动注入）
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

中间件按 `Order()` 排序后自动注册到 Gin 引擎：

```go
type AuthMiddleware struct {
    Order int // 越小越先执行
}

func (m *AuthMiddleware) Order() int {
    return 10
}

func (m *AuthMiddleware) Wrapper() gin.HandlerFunc {
    return gin.HandlerFunc(func(c *gin.Context) {
        // 中间件逻辑
        c.Next()
    })
}
```

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

#### 基本使用

框架提供了限流器中间件，支持多种限流策略：

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

#### 其他限流策略

**按路径限流：**

```go
func NewRateLimiterByPathMiddleware() common.IBaseMiddleware {
    return middleware.NewRateLimiterByPath(1000, time.Minute)
}
```

**按用户 ID 限流（需在认证中间件后执行）：**

```go
func NewRateLimiterByUserMiddleware() common.IBaseMiddleware {
    return middleware.NewRateLimiterByUserID(50, time.Minute)
}
```

**按 Header 限流：**

```go
func NewRateLimiterByAPIKeyMiddleware() common.IBaseMiddleware {
    return middleware.NewRateLimiterByHeader(100, time.Minute, "X-API-Key")
}
```

**自定义限流配置：**

```go
func NewCustomRateLimiterMiddleware() common.IBaseMiddleware {
    return middleware.NewRateLimiter(&middleware.RateLimiterConfig{
        Limit:     200,
        Window:    time.Minute,
        KeyPrefix: "custom",
        KeyFunc: func(c *gin.Context) string {
            return c.GetHeader("X-Tenant-ID")
        },
        SkipFunc: func(c *gin.Context) bool {
            return c.Request.URL.Path == "/health"
        },
    })
}
```

#### 响应头说明

限流器中间件会在响应中添加以下头：

| 响应头 | 说明 |
|--------|------|
| X-RateLimit-Limit | 时间窗口内的最大请求数 |
| X-RateLimit-Remaining | 剩余可用次数 |
| X-RateLimit-Reset | 限流窗口重置时间（秒） |
| Retry-After | 被限流时建议的等待时间（秒） |

#### 中间件 Order 建议

限流器中间件建议使用 Order = 90，在认证中间件之前执行：

```go
const (
    OrderRecovery  = 10
    OrderLogger    = 20
    OrderCors      = 30
    OrderTelemetry = 40
    OrderSecurity  = 50
    OrderRateLimit = 90
    OrderAuth      = 100
)
```


## API

### Engine

服务引擎主结构，提供完整的服务管理功能。

#### 构造函数

```go
func NewEngine(
    builtinConfig *builtin.Config,
    entity *container.EntityContainer,
    repository *container.RepositoryContainer,
    service *container.ServiceContainer,
    controller *container.ControllerContainer,
    middleware *container.MiddlewareContainer,
) *Engine
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

### 2. 中间件排序

为中间件设置合理的 Order 值，控制执行顺序：

```go
const (
    OrderRecovery    = 10  // panic 恢复
    OrderLogger      = 20  // 日志记录
    OrderCors        = 30  // 跨域处理
    OrderTelemetry   = 40  // 遥测监控
    OrderSecurity    = 50  // 安全头
    OrderRateLimit   = 90  // 限流
    OrderAuth        = 100 // 认证
)
```

来自 messageboard 的实际中间件示例：

```go
// Recovery 中间件
func (m *RecoveryMiddleware) Order() int {
    return 10
}

// 请求日志中间件
func (m *RequestLoggerMiddleware) Order() int {
    return 20
}

// CORS 中间件
func (m *CorsMiddleware) Order() int {
    return 30
}

// 遥测中间件
func (m *TelemetryMiddleware) Order() int {
    return 40
}

// 安全头中间件
func (m *SecurityHeadersMiddleware) Order() int {
    return 50
}

// 限流中间件
func (m *RateLimiterMiddleware) Order() int {
    return 90
}

// 认证中间件
func (m *AuthMiddleware) Order() int {
    return 100
}
```

**限流器中间件 Order 建议值：** 90（在认证之前执行，避免无效请求消耗认证资源）

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
2. **线程安全**：Engine 使用读写锁保护内部状态，外部访问需加锁
3. **重入保护**：Start 方法已实现重入保护，重复调用返回错误
4. **路由定义**：所有路由必须通过 BaseController 的 `GetRouter()` 方法定义
5. **中间件顺序**：中间件按 Order 升序排序，越小的值越先执行

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
