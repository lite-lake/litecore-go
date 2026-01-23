# Container

依赖注入容器，支持分层架构的依赖管理。

## 模块职责

Container 模块提供完整的分层依赖注入解决方案，负责：

1. **分层容器管理** - 提供 Entity、Manager、Repository、Service、Controller、Middleware 六层容器
2. **依赖注入** - 通过反射和结构体标签实现自动依赖注入
3. **生命周期管理** - 管理各层组件的启动和关闭
4. **循环依赖检测** - 使用拓扑排序确保无循环依赖
5. **类型安全** - 使用泛型确保编译时类型检查

## 特性

- **分层容器** - 支持 Entity、Manager、Repository、Service、Controller、Middleware 六层容器
- **类型安全** - 使用泛型确保类型安全，编译时检查
- **自动注入** - 通过结构体标签 `inject:""` 自动注入依赖
- **循环检测** - 使用拓扑排序检测循环依赖
- **线程安全** - 所有容器操作都使用读写锁保护
- **Manager 自动初始化** - Manager 组件由 Engine 自动初始化并注入

## 快速开始

```go
// 创建容器链
entityContainer := container.NewEntityContainer()
managerContainer := container.NewManagerContainer()
repositoryContainer := container.NewRepositoryContainer(entityContainer)
serviceContainer := container.NewServiceContainer(repositoryContainer)
controllerContainer := container.NewControllerContainer(serviceContainer)
middlewareContainer := container.NewMiddlewareContainer(serviceContainer)

// 设置容器链
repositoryContainer.SetManagerContainer(managerContainer)
serviceContainer.SetManagerContainer(managerContainer)
controllerContainer.SetManagerContainer(managerContainer)
middlewareContainer.SetManagerContainer(managerContainer)

// 注册实例
container.RegisterEntity(entityContainer, &UserEntity{})
container.RegisterManager(managerContainer, &ConfigManager{})
container.RegisterRepository(repositoryContainer, &UserRepository{})
container.RegisterService(serviceContainer, &UserService{})
container.RegisterController(controllerContainer, &UserController{})
container.RegisterMiddleware(middlewareContainer, &AuthMiddleware{})

// 执行依赖注入
repositoryContainer.InjectAll()
serviceContainer.InjectAll()
controllerContainer.InjectAll()
middlewareContainer.InjectAll()
```

## 分层架构

### Entity 层

Entity 层无依赖，存储数据实体实例。

```go
type UserEntity struct {
    // 实体字段
}

func (e *UserEntity) EntityName() string {
    return "UserEntity"
}

entityContainer := container.NewEntityContainer()
container.RegisterEntity(entityContainer, &UserEntity{})
```

### Manager 层

Manager 层负责管理基础组件（如配置、日志、数据库、缓存等）。Manager 组件位于独立的 `manager` 包中，由 Engine 自动初始化并注册到容器。

**内置 Manager 组件**（按初始化顺序）：
1. ConfigManager (`manager/configmgr`) - 配置管理
2. TelemetryManager (`manager/telemetrymgr`) - 遥测监控
3. LoggerManager (`manager/loggermgr`) - 日志管理
4. DatabaseManager (`manager/databasemgr`) - 数据库管理
5. CacheManager (`manager/cachemgr`) - 缓存管理
6. LockManager (`manager/lockmgr`) - 分布式锁管理
7. LimiterManager (`manager/limitermgr`) - 限流管理
8. MQManager (`manager/mqmgr`) - 消息队列管理

**Manager 自动初始化**：

Manager 组件由 Engine 自动初始化，无需手动注册：

```go
// Engine 会自动初始化所有内置 Manager
// 1. ConfigManager
// 2. TelemetryManager
// 3. LoggerManager
// 4. DatabaseManager
// 5. CacheManager
// 6. LockManager
// 7. LimiterManager
// 8. MQManager
```

**在业务代码中使用 Manager**：

```go
import (
    "github.com/lite-lake/litecore-go/common"
    "github.com/lite-lake/litecore-go/manager/configmgr"
    "github.com/lite-lake/litecore-go/manager/loggermgr"
    "github.com/lite-lake/litecore-go/manager/databasemgr"
)

type UserService struct {
    Config    configmgr.IConfigManager    `inject:""`
    LoggerMgr loggermgr.ILoggerManager   `inject:""`
    DBManager databasemgr.IDatabaseManager `inject:""`
}

func (s *UserService) ServiceName() string {
    return "UserService"
}

func (s *UserService) SomeMethod() error {
    // 使用配置管理器
    config, err := configmgr.Get[string](s.Config, "app.some.config")
    
    // 使用日志管理器（统一日志注入机制）
    s.LoggerMgr.Ins().Info("操作开始", "param", value)
    s.LoggerMgr.Ins().Error("操作失败", "error", err)
    
    // 使用数据库管理器
    var result User
    s.DBManager.DB().First(&result, id)
    
    return nil
}
```

### Repository 层

Repository 层依赖 Manager 和 Entity 层。

```go
import (
    "github.com/lite-lake/litecore-go/common"
    "github.com/lite-lake/litecore-go/manager/databasemgr"
    "github.com/lite-lake/litecore-go/manager/loggermgr"
)

type UserRepository struct {
    DBManager databasemgr.IDatabaseManager `inject:""`
    UserEntity UserEntity                   `inject:""`
    LoggerMgr loggermgr.ILoggerManager     `inject:""`
}

func (r *UserRepository) RepositoryName() string {
    return "UserRepository"
}

func (r *UserRepository) FindByID(id int64) (*User, error) {
    var user User
    result := r.DBManager.DB().First(&user, id)
    if result.Error != nil {
        r.LoggerMgr.Ins().Error("查询用户失败", "id", id, "error", result.Error)
        return nil, result.Error
    }
    return &user, nil
}

repositoryContainer := container.NewRepositoryContainer(entityContainer)
repositoryContainer.SetManagerContainer(managerContainer)
container.RegisterRepository(repositoryContainer, &UserRepository{})
```

### Service 层

Service 层依赖 Manager、Repository 和其他 Service 层。

```go
import (
    "github.com/lite-lake/litecore-go/common"
    "github.com/lite-lake/litecore-go/manager/configmgr"
    "github.com/lite-lake/litecore-go/manager/loggermgr"
)

type UserService struct {
    Config    configmgr.IConfigManager    `inject:""`
    Repo      IUserRepository             `inject:""`
    AuthService IAuthService             `inject:""` // 其他服务
    LoggerMgr loggermgr.ILoggerManager   `inject:""`
}

func (s *UserService) ServiceName() string {
    return "UserService"
}

func (s *UserService) OnStart() error {
    s.LoggerMgr.Ins().Info("UserService 启动")
    return nil
}

func (s *UserService) OnStop() error {
    s.LoggerMgr.Ins().Info("UserService 停止")
    return nil
}

func (s *UserService) CreateUser(req *CreateUserRequest) error {
    s.LoggerMgr.Ins().Info("创建用户开始", "name", req.Name)
    
    // 业务逻辑
    
    s.LoggerMgr.Ins().Info("创建用户成功", "user_id", user.ID)
    return nil
}

serviceContainer := container.NewServiceContainer(repositoryContainer)
serviceContainer.SetManagerContainer(managerContainer)
container.RegisterService(serviceContainer, &UserService{})
```

### Controller 层

Controller 层依赖 Manager 和 Service 层。

```go
import (
    "github.com/lite-lake/litecore-go/common"
    "github.com/lite-lake/litecore-go/manager/loggermgr"
)

type UserController struct {
    UserService IUserService            `inject:""`
    LoggerMgr loggermgr.ILoggerManager   `inject:""`
}

func (c *UserController) ControllerName() string {
    return "UserController"
}

func (c *UserController) GetRouter() string {
    return "/api/users/:id [GET]"
}

func (c *UserController) Handle(ctx *gin.Context) {
    id := ctx.Param("id")
    c.LoggerMgr.Ins().Info("获取用户", "id", id)
    
    user, err := c.UserService.GetUser(id)
    if err != nil {
        c.LoggerMgr.Ins().Error("获取用户失败", "id", id, "error", err)
        ctx.JSON(common.HTTPStatusInternalServerError, gin.H{"error": err.Error()})
        return
    }
    
    ctx.JSON(common.HTTPStatusOK, user)
}

controllerContainer := container.NewControllerContainer(serviceContainer)
controllerContainer.SetManagerContainer(managerContainer)
container.RegisterController(controllerContainer, &UserController{})
```

### Middleware 层

Middleware 层依赖 Manager 和 Service 层。

```go
import (
    "github.com/lite-lake/litecore-go/common"
    "github.com/lite-lake/litecore-go/manager/loggermgr"
)

type AuthMiddleware struct {
    UserService IUserService            `inject:""`
    LoggerMgr loggermgr.ILoggerManager   `inject:""`
}

func (m *AuthMiddleware) MiddlewareName() string {
    return "AuthMiddleware"
}

func (m *AuthMiddleware) Handle(ctx *gin.Context) {
    token := ctx.GetHeader("Authorization")
    
    user, err := m.UserService.ValidateToken(token)
    if err != nil {
        m.LoggerMgr.Ins().Warn("认证失败", "token", token, "error", err)
        ctx.JSON(common.HTTPStatusUnauthorized, gin.H{"error": "unauthorized"})
        ctx.Abort()
        return
    }
    
    ctx.Set("user", user)
    ctx.Next()
}

middlewareContainer := container.NewMiddlewareContainer(serviceContainer)
middlewareContainer.SetManagerContainer(managerContainer)
container.RegisterMiddleware(middlewareContainer, &AuthMiddleware{})
```

## 依赖注入

### 基本用法

使用 `inject:""` 标签标记需要注入的字段。

```go
type UserService struct {
    ConfigManager `inject:""`
    UserRepository `inject:""`
}
```

### 可选依赖

使用 `inject:"optional"` 标记可选依赖。

```go
type UserService struct {
    ConfigManager `inject:""`
    CacheService `inject:"optional"` // 可选依赖
}
```

## API

### EntityContainer

```go
func NewEntityContainer() *EntityContainer
func RegisterEntity[T common.IBaseEntity](e *EntityContainer, impl T) error
func (e *EntityContainer) GetByName(name string) (common.IBaseEntity, error)
func (e *EntityContainer) GetAll() []common.IBaseEntity
func (e *EntityContainer) GetByType(typ reflect.Type) ([]common.IBaseEntity, error)
func (e *EntityContainer) Count() int
```

### ManagerContainer

```go
func NewManagerContainer() *ManagerContainer
func RegisterManager[T common.IBaseManager](m *ManagerContainer, impl T) error
func GetManager[T common.IBaseManager](m *ManagerContainer) (T, error)
func (m *ManagerContainer) RegisterByType(ifaceType reflect.Type, impl common.IBaseManager) error
func (m *ManagerContainer) GetByType(ifaceType reflect.Type) common.IBaseManager
func (m *ManagerContainer) GetAll() []common.IBaseManager
func (m *ManagerContainer) GetAllSorted() []common.IBaseManager
func (m *ManagerContainer) GetNames() []string
func (m *ManagerContainer) Count() int
```

### RepositoryContainer

```go
func NewRepositoryContainer(entity *EntityContainer) *RepositoryContainer
func RegisterRepository[T common.IBaseRepository](r *RepositoryContainer, impl T) error
func GetRepository[T common.IBaseRepository](r *RepositoryContainer) (T, error)
func (r *RepositoryContainer) SetManagerContainer(container *ManagerContainer)
func (r *RepositoryContainer) InjectAll() error
func (r *RepositoryContainer) RegisterByType(ifaceType reflect.Type, impl common.IBaseRepository) error
func (r *RepositoryContainer) GetByType(ifaceType reflect.Type) common.IBaseRepository
func (r *RepositoryContainer) GetAll() []common.IBaseRepository
func (r *RepositoryContainer) GetAllSorted() []common.IBaseRepository
func (r *RepositoryContainer) Count() int
```

### ServiceContainer

```go
func NewServiceContainer(repository *RepositoryContainer) *ServiceContainer
func RegisterService[T common.IBaseService](s *ServiceContainer, impl T) error
func GetService[T common.IBaseService](s *ServiceContainer) (T, error)
func (s *ServiceContainer) SetManagerContainer(container *ManagerContainer)
func (s *ServiceContainer) InjectAll() error
func (s *ServiceContainer) RegisterByType(ifaceType reflect.Type, impl common.IBaseService) error
func (s *ServiceContainer) GetByType(ifaceType reflect.Type) common.IBaseService
func (s *ServiceContainer) GetAll() []common.IBaseService
func (s *ServiceContainer) GetAllSorted() []common.IBaseService
func (s *ServiceContainer) Count() int
```

### ControllerContainer

```go
func NewControllerContainer(service *ServiceContainer) *ControllerContainer
func RegisterController[T common.IBaseController](c *ControllerContainer, impl T) error
func GetController[T common.IBaseController](c *ControllerContainer) (T, error)
func (c *ControllerContainer) SetManagerContainer(container *ManagerContainer)
func (c *ControllerContainer) InjectAll() error
func (c *ControllerContainer) RegisterByType(ifaceType reflect.Type, impl common.IBaseController) error
func (c *ControllerContainer) GetByType(ifaceType reflect.Type) common.IBaseController
func (c *ControllerContainer) GetAll() []common.IBaseController
func (c *ControllerContainer) GetAllSorted() []common.IBaseController
func (c *ControllerContainer) Count() int
```

### MiddlewareContainer

```go
func NewMiddlewareContainer(service *ServiceContainer) *MiddlewareContainer
func RegisterMiddleware[T common.IBaseMiddleware](m *MiddlewareContainer, impl T) error
func GetMiddleware[T common.IBaseMiddleware](m *MiddlewareContainer) (T, error)
func (m *MiddlewareContainer) SetManagerContainer(container *ManagerContainer)
func (m *MiddlewareContainer) InjectAll() error
func (m *MiddlewareContainer) RegisterByType(ifaceType reflect.Type, impl common.IBaseMiddleware) error
func (m *MiddlewareContainer) GetByType(ifaceType reflect.Type) common.IBaseMiddleware
func (m *MiddlewareContainer) GetAll() []common.IBaseMiddleware
func (m *MiddlewareContainer) GetAllSorted() []common.IBaseMiddleware
func (m *MiddlewareContainer) Count() int
```

## Manager 初始化机制

### 自动初始化

Manager 组件由 Engine 自动初始化，开发者无需手动创建和注册 Manager。初始化顺序如下：

1. **ConfigManager** - 必须最先初始化，其他 Manager 依赖它
2. **TelemetryManager** - 依赖 ConfigManager
3. **LoggerManager** - 依赖 ConfigManager 和 TelemetryManager
4. **DatabaseManager** - 依赖 ConfigManager
5. **CacheManager** - 依赖 ConfigManager
6. **LockManager** - 依赖 ConfigManager
7. **LimiterManager** - 依赖 ConfigManager
8. **MQManager** - 依赖 ConfigManager

```go
// Engine 在 Initialize() 时自动初始化所有 Manager
engine := server.NewEngine(
    &server.BuiltinConfig{
        Driver:   "yaml",
        FilePath: "configs/config.yaml",
    },
    entityContainer,
    repositoryContainer,
    serviceContainer,
    controllerContainer,
    middlewareContainer,
)
```

### 手动扩展 Manager

如需自定义 Manager，可以在 Engine 初始化后手动注册：

```go
// Engine 初始化后
err := engine.Initialize()
if err != nil {
    panic(err)
}

// 手动注册自定义 Manager
customManager := NewCustomManager()
if err := container.RegisterManager(engine.Manager, customManager); err != nil {
    panic(err)
}
```

## 日志注入机制

### 统一日志注入

所有层组件都可以通过依赖注入 `loggermgr.ILoggerManager` 来使用日志。框架提供统一的日志接口，支持多种日志驱动。

### 在各层中使用日志

```go
import (
    "github.com/lite-lake/litecore-go/manager/loggermgr"
)

// Repository 层
type UserRepository struct {
    LoggerMgr loggermgr.ILoggerManager `inject:""`
}

func (r *UserRepository) FindByID(id int64) (*User, error) {
    r.LoggerMgr.Ins().Debug("查询用户", "id", id)
    // ...
}

// Service 层
type UserService struct {
    LoggerMgr loggermgr.ILoggerManager `inject:""`
}

func (s *UserService) CreateUser(req *CreateUserRequest) error {
    s.LoggerMgr.Ins().Info("创建用户开始", "name", req.Name)
    // ...
    s.LoggerMgr.Ins().Error("创建用户失败", "error", err)
    // ...
}

// Controller 层
type UserController struct {
    LoggerMgr loggermgr.ILoggerManager `inject:""`
}

func (c *UserController) Handle(ctx *gin.Context) {
    c.LoggerMgr.Ins().Info("处理请求", "path", ctx.Request.URL.Path)
    // ...
}

// Middleware 层
type AuthMiddleware struct {
    LoggerMgr loggermgr.ILoggerManager `inject:""`
}

func (m *AuthMiddleware) Handle(ctx *gin.Context) {
    token := ctx.GetHeader("Authorization")
    m.LoggerMgr.Ins().Warn("认证失败", "token", token)
    // ...
}
```

### 日志级别

- **Debug** - 开发调试信息
- **Info** - 正常业务流程
- **Warn** - 降级处理、重试
- **Error** - 业务错误、操作失败
- **Fatal** - 致命错误，需要立即终止

### 结构化日志

使用结构化日志，推荐格式：

```go
// 推荐：使用键值对
s.LoggerMgr.Ins().Info("用户登录", 
    "user_id", user.ID, 
    "ip", clientIP,
    "user_agent", userAgent)

// 推荐：使用 logger.F() 函数
s.LoggerMgr.Ins().Error("操作失败", 
    logger.F("error", err),
    logger.F("user_id", user.ID),
    logger.F("operation", "create"))
```

## 错误处理

### DependencyNotFoundError

依赖未找到错误。

```go
type DependencyNotFoundError struct {
    InstanceName  string       // 当前实例名称
    FieldName     string       // 缺失依赖的字段名
    FieldType     reflect.Type // 期望的依赖类型
    ContainerType string       // 应该从哪个容器查找
}
```

### CircularDependencyError

循环依赖错误。

```go
type CircularDependencyError struct {
    Cycle []string // 循环依赖链
}
```

### AmbiguousMatchError

多重匹配错误。

```go
type AmbiguousMatchError struct {
    InstanceName string
    FieldName    string
    FieldType    reflect.Type
    Candidates   []string // 匹配的候选实例名称
}
```

### 其他错误类型

- `DuplicateRegistrationError` - 重复注册错误
- `InstanceNotFoundError` - 实例未找到错误
- `InterfaceAlreadyRegisteredError` - 接口已被注册错误
- `ImplementationDoesNotImplementInterfaceError` - 实现未实现接口错误
- `InterfaceNotRegisteredError` - 接口未注册错误
- `ManagerContainerNotSetError` - ManagerContainer 未设置错误
- `UninjectedFieldError` - 未注入字段错误

## 最佳实践

### 依赖注入顺序

确保按以下顺序设置容器链和执行注入：

1. 创建所有容器
2. 设置 ManagerContainer 到所有需要它的容器
3. 注册所有实例
4. 按从底到顶的顺序执行注入：
    - `repositoryContainer.InjectAll()`
    - `serviceContainer.InjectAll()`
    - `controllerContainer.InjectAll()`
    - `middlewareContainer.InjectAll()`

### Manager 使用规范

1. **不手动初始化** - Manager 由 Engine 自动初始化
2. **按需注入** - 只在需要的组件中注入 Manager
3. **统一日志** - 通过 LoggerMgr 使用日志，避免直接使用 `log.Fatal/Printf` 等标准库日志
4. **配置获取** - 使用 ConfigManager 获取配置，而不是硬编码

```go
// 推荐：通过 Manager 使用功能
type UserService struct {
    Config    configmgr.IConfigManager    `inject:""`
    LoggerMgr loggermgr.ILoggerManager   `inject:""`
    DBManager databasemgr.IDatabaseManager `inject:""`
}

// 不推荐：硬编码或直接创建 Manager
type UserService struct {
    config *some.Config  // ❌ 不推荐
    logger log.Logger    // ❌ 不推荐
}
```

### 避免循环依赖

设计依赖关系时应避免循环依赖。如果存在循环依赖，注入时会返回 `CircularDependencyError`。

依赖关系原则：
- Entity → 无依赖
- Manager → ConfigManager（其他 Manager 可以依赖其他 Manager）
- Repository → Manager + Entity
- Service → Manager + Repository + Service
- Controller → Manager + Service
- Middleware → Manager + Service

### 使用泛型函数

使用泛型注册和获取函数可以提高代码类型安全：

```go
// 推荐
container.RegisterService(serviceContainer, &UserService{})
userService, err := container.GetService[*UserService](serviceContainer)

// 不推荐
serviceContainer.RegisterByType(serviceType, &UserService{})
userService := serviceContainer.GetByType(serviceType)
```

### 日志使用规范

**禁止使用**：
- ❌ 标准库 `log.Fatal/Print/Printf/Println`
- ❌ `fmt.Printf/fmt.Println`（仅限开发调试）
- ❌ `println/print`

**推荐使用**：
- ✅ 依赖注入 `ILoggerManager`
- ✅ 使用结构化日志：`logger.Info("msg", "key", value)`
- ✅ 使用 `With` 添加上下文：`logger.With("user_id", id).Info("...")`

### 完整示例

```go
package main

import (
    "github.com/lite-lake/litecore-go/common"
    "github.com/lite-lake/litecore-go/container"
    "github.com/lite-lake/litecore-go/manager/loggermgr"
    "github.com/lite-lake/litecore-go/manager/databasemgr"
    "github.com/lite-lake/litecore-go/server"
)

type UserService struct {
    LoggerMgr loggermgr.ILoggerManager   `inject:""`
    DBManager databasemgr.IDatabaseManager `inject:""`
}

func (s *UserService) ServiceName() string {
    return "UserService"
}

func (s *UserService) OnStart() error {
    s.LoggerMgr.Ins().Info("UserService 启动")
    return nil
}

func (s *UserService) OnStop() error {
    return nil
}

func main() {
    // 创建容器
    entityContainer := container.NewEntityContainer()
    repositoryContainer := container.NewRepositoryContainer(entityContainer)
    serviceContainer := container.NewServiceContainer(repositoryContainer)
    controllerContainer := container.NewControllerContainer(serviceContainer)
    middlewareContainer := container.NewMiddlewareContainer(serviceContainer)

    // 注册 Service
    container.RegisterService(serviceContainer, &UserService{})

    // 创建 Engine（自动初始化 Manager）
    engine := server.NewEngine(
        &server.BuiltinConfig{
            Driver:   "yaml",
            FilePath: "configs/config.yaml",
        },
        entityContainer,
        repositoryContainer,
        serviceContainer,
        controllerContainer,
        middlewareContainer,
    )

    // 初始化和启动
    if err := engine.Initialize(); err != nil {
        panic(err)
    }

    if err := engine.Start(); err != nil {
        panic(err)
    }

    engine.WaitForShutdown()
}
```
