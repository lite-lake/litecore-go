# Server

提供统一的 HTTP 服务引擎，支持自动依赖注入、生命周期管理和中间件集成。

## 特性

- **容器管理** - 集成 Config、Entity、Manager、Repository、Service、Controller、Middleware 七层容器
- **自动注入** - 按依赖顺序自动处理组件注入（Entity → Manager → Repository → Service → Controller/Middleware）
- **生命周期管理** - 统一管理各层组件的启动和停止，支持健康检查
- **中间件集成** - 自动排序并注册全局中间件到 Gin 引擎
- **路由管理** - 自动注册控制器路由，支持自定义路由扩展
- **优雅关闭** - 支持信号处理，超时控制的安全关闭机制

## 快速开始

```go
package main

import (
    "com.litelake.litecore/common"
    "com.litelake.litecore/config"
    "com.litelake.litecore/container"
    "com.litelake.litecore/databasemgr"
    "com.litelake.litecore/server"
)

func main() {
    // 创建容器
    configContainer := container.NewConfigContainer()
    entityContainer := container.NewEntityContainer()
    managerContainer := container.NewManagerContainer(configContainer)
    repositoryContainer := container.NewRepositoryContainer(configContainer, managerContainer, entityContainer)
    serviceContainer := container.NewServiceContainer(configContainer, managerContainer, repositoryContainer)
    controllerContainer := container.NewControllerContainer(configContainer, managerContainer, serviceContainer)
    middlewareContainer := container.NewMiddlewareContainer(configContainer, managerContainer, serviceContainer)

    // 注册配置
    configProvider, _ := config.NewConfigProvider("yaml", "config.yaml")
    container.RegisterConfig[common.BaseConfigProvider](configContainer, configProvider)

    // 注册管理器
    dbMgr := databasemgr.NewDatabaseManager()
    container.RegisterManager[databasemgr.DatabaseManager](managerContainer, dbMgr)

    // 注册其他组件（实体、仓储、服务、控制器、中间件）...

    // 创建并启动引擎
    engine := server.NewEngine(
        configContainer,
        entityContainer,
        managerContainer,
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

使用 `NewEngine` 创建服务引擎，需要传入所有容器实例：

```go
engine := server.NewEngine(
    configContainer,
    entityContainer,
    managerContainer,
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
    panic(err)
}
```

**方式二：分步启动**

```go
// 1. 初始化（依赖注入、创建 Gin 引擎、注册中间件和路由）
if err := engine.Initialize(); err != nil {
    panic(err)
}

// 2. 启动服务（启动各层组件和 HTTP 服务器）
if err := engine.Start(); err != nil {
    panic(err)
}

// 3. 等待关闭信号
engine.WaitForShutdown()
```

### 生命周期管理

Engine 按以下顺序管理组件生命周期：

**启动顺序：**
1. Manager 层（按注册顺序）
2. Repository 层（按注册顺序）
3. Service 层（按注册顺序）
4. HTTP 服务器

**停止顺序（反转启动顺序）：**
1. HTTP 服务器（优雅关闭）
2. Service 层（反转顺序）
3. Repository 层（反转顺序）
4. Manager 层（反转顺序）

```go
// 手动停止
if err := engine.Stop(); err != nil {
    log.Printf("Stop error: %v", err)
}

// 重启
if err := engine.Restart(); err != nil {
    log.Printf("Restart error: %v", err)
}

// 健康检查
if err := engine.Health(); err != nil {
    log.Printf("Health check failed: %v", err)
}
```

### 依赖注入

Engine 在初始化时自动按以下顺序执行依赖注入：

1. **Entity 层**（无依赖）
2. **Manager 层**（依赖 Config 和同层 Manager）
3. **Repository 层**（依赖 Config、Manager、Entity）
4. **Service 层**（依赖 Config、Manager、Repository 和同层 Service）
5. **Controller 层**（依赖 Config、Manager、Service）
6. **Middleware 层**（依赖 Config、Manager、Service）

各层组件通过 `inject:""` 标签声明依赖：

```go
type UserServiceImpl struct {
    Config    common.BaseConfigProvider  `inject:""`
    DBManager databasemgr.DatabaseManager `inject:""`
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

**自动注册控制器路由**

控制器通过 `GetRouter()` 方法定义路由，格式为：`/path [METHOD]`

```go
func (ctrl *UserController) GetRouter() string {
    return "/users [GET]"
}
```

**自定义路由扩展**

初始化后获取 Gin 引擎进行扩展：

```go
if err := engine.Initialize(); err != nil {
    panic(err)
}

r := engine.GetGinEngine()
r.GET("/health", func(c *gin.Context) {
    c.JSON(200, gin.H{"status": "ok"})
})

// 注册路由组
api := r.Group("/api/v1")
{
    api.GET("/users", listUsers)
    api.POST("/users", createUser)
}

if err := engine.Run(); err != nil {
    panic(err)
}
```

**获取路由信息**

```go
routes := engine.GetRouteInfo()
for _, route := range routes {
    fmt.Printf("%-6s %s\n", route.Method, route.Path)
}
```

## API

### Engine

服务引擎主结构，提供完整的服务管理功能。

#### 构造函数

```go
func NewEngine(
    config *container.ConfigContainer,
    entity *container.EntityContainer,
    manager *container.ManagerContainer,
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
| `Restart() error` | 重启引擎 |
| `Run() error` | 一键启动（Initialize + Start + WaitForShutdown） |
| `WaitForShutdown()` | 等待关闭信号 |

#### 配置方法

| 方法 | 说明 |
|------|------|
| `SetMode(mode string)` | 设置 Gin 运行模式（debug/release/test） |
| `IsStarted() bool` | 检查引擎是否已启动 |

#### 获取组件

| 方法 | 返回值 | 说明 |
|------|--------|------|
| `GetGinEngine()` | `*gin.Engine` | 获取 Gin 引擎用于自定义扩展 |
| `GetConfig()` | `common.BaseConfigProvider` | 获取配置提供者 |
| `GetLogger()` | `interface{}` | 获取日志记录器（如果有 LoggerManager） |
| `GetManagers()` | `[]common.BaseManager` | 获取所有管理器 |
| `GetServices()` | `[]common.BaseService` | 获取所有服务 |
| `GetControllers()` | `[]common.BaseController` | 获取所有控制器 |
| `GetMiddlewares()` | `[]common.BaseMiddleware` | 获取所有中间件 |

#### 路由方法

| 方法 | 说明 |
|------|------|
| `RegisterRoute(method, path string, handler gin.HandlerFunc)` | 注册自定义路由 |
| `RegisterGroup(groupPath string, handlers ...gin.HandlerFunc)` | 注册路由组 |
| `GetRouteInfo()` | 获取所有已注册的路由信息 |

#### 健康检查

| 方法 | 说明 |
|------|------|
| `Health() error` | 检查所有 Manager 的健康状态 |

### ServerConfig

服务器配置结构。

```go
type ServerConfig struct {
    Host            string        // 监听地址，默认 0.0.0.0
    Port            int           // 监听端口，默认 8080
    Mode            string        // 运行模式：debug/release/test，默认 release
    ReadTimeout     time.Duration // 读取超时，默认 10s
    WriteTimeout    time.Duration // 写入超时，默认 10s
    IdleTimeout     time.Duration // 空闲超时，默认 60s
    EnableRecovery  bool          // 是否启用 panic 恢复，默认 true
    ShutdownTimeout time.Duration // 关闭超时，默认 30s
}

func DefaultServerConfig() *ServerConfig
func (c *ServerConfig) Address() string
```

## 最佳实践

### 1. 容器注册顺序

按照依赖顺序创建和注册容器：

```go
// 1. 创建容器（按依赖顺序）
configContainer := container.NewConfigContainer()
entityContainer := container.NewEntityContainer()
managerContainer := container.NewManagerContainer(configContainer)
repositoryContainer := container.NewRepositoryContainer(configContainer, managerContainer, entityContainer)
serviceContainer := container.NewServiceContainer(configContainer, managerContainer, repositoryContainer)
controllerContainer := container.NewControllerContainer(configContainer, managerContainer, serviceContainer)
middlewareContainer := container.NewMiddlewareContainer(configContainer, managerContainer, serviceContainer)

// 2. 注册组件（按层级顺序）
// Config 层
// Manager 层
// Entity 层
// Repository 层
// Service 层
// Controller 层
// Middleware 层
```

### 2. 中间件排序

为中间件设置合理的 Order 值，控制执行顺序：

```go
const (
    OrderRecovery    = 10  // panic 恢复
    OrderLogger      = 20  // 日志记录
    OrderCors        = 30  // 跨域处理
    OrderAuth        = 100 // 认证
    OrderRateLimit   = 90  // 限流
)
```

### 3. 路由命名规范

控制器路由使用 OpenAPI 风格：`/path [METHOD]`

```go
// 正确
"/users [GET]"
"/users/:id [GET]"
"/users [POST]"
"/users/:id [PUT]"
"/users/:id [DELETE]"

// 错误（会被忽略）
""                // 空字符串
"users"           // 缺少方法
"/users[GET]"      // 缺少空格
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
2. HTTP 服务器优雅关闭（等待现有请求完成，超时时间由 ShutdownTimeout 配置）
3. 停止 Service 层（反转注册顺序）
4. 停止 Repository 层（反转注册顺序）
5. 停止 Manager 层（反转注册顺序）

## 注意事项

1. **依赖注入**：确保组件使用 `inject:""` 标签声明依赖
2. **线程安全**：Engine 使用读写锁保护内部状态，外部访问需加锁
3. **重入保护**：Start 方法已实现重入保护，重复调用返回错误
4. **自定义路由**：必须在 Initialize 后、Run 前调用 GetGinEngine 进行扩展
5. **中间件顺序**：中间件按 Order 升序排序，越小的值越先执行
