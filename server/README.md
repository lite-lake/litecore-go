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
    "github.com/lite-lake/litecore-go/common"
    "github.com/lite-lake/litecore-go/config"
    "github.com/lite-lake/litecore-go/container"
    "github.com/lite-lake/litecore-go/databasemgr"
    "github.com/lite-lake/litecore-go/server"
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

或手动创建（传入所有容器实例）：

```go
engine := server.NewEngine(
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
2. **Entity 层**（无依赖）
3. **Repository 层**（依赖 Config、Manager、Entity）
4. **Service 层**（依赖 Config、Manager、Repository 和同层 Service）
5. **Controller 层**（依赖 Config、Manager、Service）
6. **Middleware 层**（依赖 Config、Manager、Service）

各层组件通过 `inject:""` 标签声明依赖，Config 和 Manager 由引擎自动注入：

```go
type UserServiceImpl struct {
    // 内置组件（引擎自动注入）
    Config    common.BaseConfigProvider  `inject:""`
    DBManager databasemgr.DatabaseManager `inject:""`

    // 业务依赖
    UserRepo  repository.IUserRepository `inject:""`
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



## API

### Engine

服务引擎主结构，提供完整的服务管理功能。

#### 构造函数

```go
func NewEngine(
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

    return server.NewEngine(entityContainer, repositoryContainer, serviceContainer, controllerContainer, middlewareContainer), nil
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

// 认证中间件
func (m *AuthMiddleware) Order() int {
    return 100
}
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
2. **线程安全**：Engine 使用读写锁保护内部状态，外部访问需加锁
3. **重入保护**：Start 方法已实现重入保护，重复调用返回错误
4. **路由定义**：所有路由必须通过 BaseController 的 `GetRouter()` 方法定义
5. **中间件顺序**：中间件按 Order 升序排序，越小的值越先执行

来自 messageboard 的实际注意事项示例：

**依赖注入声明**

```go
type MessageServiceImpl struct {
    Config   config.BaseConfigProvider   `inject:""`
    DBMgr    databasemgr.DatabaseManager `inject:""`
    Messages repository.IMessageRepository `inject:""`
}

type AuthServiceImpl struct {
    Config  config.BaseConfigProvider `inject:""`
    DBMgr   databasemgr.DatabaseManager `inject:""`
    Sessions repository.ISessionRepository `inject:""`
}
```

**路由定义规范**

所有路由都应通过控制器定义：

```go
type HealthCheckController struct {
    Config  config.BaseConfigProvider `inject:""`
}

func (ctrl *HealthCheckController) GetRouter() string {
    return "/health [GET]"
}

func (ctrl *HealthCheckController) Handle(c *gin.Context) {
    c.JSON(200, gin.H{"status": "ok"})
}
```
