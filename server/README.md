# Server

提供统一的 HTTP 服务引擎，支持自动依赖注入、生命周期管理和中间件集成。

## 特性

- **内置组件自动初始化** - 自动初始化 8 个内置 Manager（Config、Telemetry、Logger、Database、Cache、Lock、Limiter、MQ）
- **5 层依赖注入架构** - Entity → Repository → Service → Controller → Middleware，支持层级依赖和同层依赖
- **生命周期管理** - 统一管理各层组件的启动和停止，启动和停止顺序可预测
- **中间件集成** - 自动排序并注册全局中间件，支持通过配置自定义名称和执行顺序
- **路由管理** - 自动注册控制器路由，支持 OpenAPI 风格的路由定义 `/path [METHOD]`
- **优雅关闭** - 支持信号处理，超时控制的安全关闭机制
- **启动日志** - 支持异步启动日志，记录各阶段启动状态和耗时

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
// 1. 初始化（初始化内置组件、依赖注入、创建 Gin 引擎、注册中间件和路由）
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

### 启动日志

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
2026-01-24 15:04:06.123 | INFO  | 开始依赖注入
2026-01-24 15:04:06.456 | INFO  | [Repository 层] MessageRepository: 注入完成
2026-01-24 15:04:06.789 | INFO  | 依赖注入完成 | count=23
2026-01-24 15:04:07.123 | INFO  | HTTP 服务器启动成功 | address=0.0.0.0:8080
```

### 生命周期管理

Engine 按以下顺序管理组件生命周期：

**Initialize() 初始化顺序：**
1. 管理器初始化（按顺序初始化 8 个内置 Manager）
   - ConfigManager（必须最先初始化）
   - TelemetryManager（依赖 ConfigManager）
   - LoggerManager（依赖 ConfigManager、TelemetryManager）
   - DatabaseManager（依赖 ConfigManager）
   - CacheManager（依赖 ConfigManager）
   - LockManager（依赖 ConfigManager）
   - LimiterManager（依赖 ConfigManager）
   - MQManager（依赖 ConfigManager）
2. 依赖注入（按层顺序）
   - Repository 层（依赖 Manager、Entity）
   - Service 层（依赖 Manager、Repository 和同层 Service）
   - Controller 层（依赖 Manager、Service）
   - Middleware 层（依赖 Manager、Service）
3. 创建 Gin 引擎
4. 注册中间件和路由

**Start() 启动顺序：**
1. Manager 层（按注册顺序）
2. Repository 层（按注册顺序）
3. Service 层（按注册顺序）
4. Middleware 层（按注册顺序）
5. HTTP 服务器

**Stop() 停止顺序（反转启动顺序）：**
1. HTTP 服务器（优雅关闭）
2. Middleware 层（反转注册顺序）
3. Service 层（反转注册顺序）
4. Repository 层（反转注册顺序）
5. Manager 层（反转注册顺序）

```go
// 手动停止
if err := engine.Stop(); err != nil {
    log.Printf("Stop error: %v", err)
}
```

### 依赖注入

Engine 在初始化时自动按以下顺序执行依赖注入：

1. **Manager 层**（由 `server.Initialize()` 自动初始化并注入）
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
    Config    configmgr.IConfigManager  `inject:""`
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

**支持的路由格式：**

```go
// 正确格式（OpenAPI 风格）
"/users [GET]"
"/users/:id [GET]"
"/users [POST]"
"/users/:id [PUT]"
"/users/:id [DELETE]"
"/static/*filepath [GET]"

// 错误格式（会被忽略）
""                // 空字符串
"users"           // 缺少方法
"/users[GET]"     // 缺少空格
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
| `Initialize() error` | 初始化引擎（初始化内置组件、依赖注入、创建 Gin 引擎、注册中间件和路由） |
| `Start() error` | 启动引擎（启动各层组件和 HTTP 服务器） |
| `Stop() error` | 停止引擎（优雅关闭） |
| `Run() error` | 一键启动（Initialize + Start + WaitForShutdown） |
| `WaitForShutdown()` | 等待关闭信号 |

## 最佳实践

### 1. 使用 CLI 工具生成应用引擎

推荐使用 CLI 工具生成容器初始化代码，确保容器注册顺序和依赖注入正确：

```go
// 由 CLI 工具自动生成 application/engine.go
func NewEngine() (*server.Engine, error) {
    entityContainer := InitEntityContainer()
    repositoryContainer := InitRepositoryContainer(entityContainer)
    serviceContainer := InitServiceContainer(repositoryContainer)
    controllerContainer := InitControllerContainer(serviceContainer)
    middlewareContainer := InitMiddlewareContainer(serviceContainer)

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

### 3. 路由命名规范

控制器路由使用 OpenAPI 风格：`/path [METHOD]`

```go
// 正确
"/messages [GET]"
"/messages/:id [GET]"
"/messages [POST]"
"/messages/:id [PUT]"
"/messages/:id [DELETE]"
```

### 4. 错误处理

启动和停止时的错误应妥善处理：

```go
if err := engine.Run(); err != nil {
    log.Fatalf("Engine run failed: %v", err)
}
```

## 信号处理

Engine 自动处理以下信号，触发优雅关闭：

- `SIGINT`（Ctrl+C）
- `SIGTERM`
- `SIGQUIT`

关闭流程：

1. 捕获信号
2. HTTP 服务器优雅关闭（等待现有请求完成）
3. 停止 Middleware 层（反转注册顺序）
4. 停止 Service 层（反转注册顺序）
5. 停止 Repository 层（反转注册顺序）
6. 停止 Manager 层（反转注册顺序）

## 注意事项

1. **依赖注入**：确保组件使用 `inject:""` 标签声明依赖
2. **Manager 引用**：Manager 组件位于 `manager/` 目录，导入路径为 `github.com/lite-lake/litecore-go/manager/xxxmgr`
3. **线程安全**：Engine 使用读写锁保护内部状态
4. **重入保护**：Start 方法已实现重入保护，重复调用返回错误
5. **路由定义**：所有路由必须通过 BaseController 的 `GetRouter()` 方法定义
6. **中间件顺序**：中间件按 Order 升序排序，越小的值越先执行
7. **中间件配置**：内置中间件支持通过配置自定义名称和执行顺序
