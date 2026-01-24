# Container

依赖注入容器，提供 5 层分层架构的自动依赖管理。

## 特性

- **5 层容器** - Entity、Manager、Repository、Service、Controller、Middleware 六层容器
- **类型安全** - 泛型注册与获取，编译时类型检查
- **自动注入** - 通过 `inject:""` 结构体标签自动注入依赖
- **拓扑排序** - Service 层使用 Kahn 算法检测并解决循环依赖
- **线程安全** - 所有容器操作使用读写锁保护
- **Manager 自动初始化** - Engine 自动初始化并注册内置 Manager

## 快速开始

```go
// 创建容器链
entityContainer := container.NewEntityContainer()
repositoryContainer := container.NewRepositoryContainer(entityContainer)
serviceContainer := container.NewServiceContainer(repositoryContainer)
controllerContainer := container.NewControllerContainer(serviceContainer)
middlewareContainer := container.NewMiddlewareContainer(serviceContainer)

// 注册实例
container.RegisterEntity(entityContainer, &Message{})
container.RegisterRepository[IMessageRepository](repositoryContainer, NewMessageRepository())
container.RegisterService[IMessageService](serviceContainer, NewMessageService())
container.RegisterController[IMessageController](controllerContainer, NewMessageController())
container.RegisterMiddleware[IAuthMiddleware](middlewareContainer, NewAuthMiddleware())

// 创建 Engine（自动初始化 Manager 和执行依赖注入）
engine := server.NewEngine(
    &server.BuiltinConfig{Driver: "yaml", FilePath: "configs/config.yaml"},
    entityContainer, repositoryContainer, serviceContainer,
    controllerContainer, middlewareContainer,
)
engine.Initialize()
engine.Run()
```

## 分层容器

### Entity 容器

Entity 层无依赖，仅存储数据实体。

```go
entityContainer := container.NewEntityContainer()
container.RegisterEntity(entityContainer, &Message{})
```

### Repository 容器

依赖 Manager 和 Entity 层。

```go
repositoryContainer := container.NewRepositoryContainer(entityContainer)
container.RegisterRepository[IMessageRepository](repositoryContainer, NewMessageRepository())
```

### Service 容器

依赖 Manager、Repository 和其他 Service 层，支持拓扑排序。

```go
serviceContainer := container.NewServiceContainer(repositoryContainer)
container.RegisterService[IMessageService](serviceContainer, NewMessageService())
container.RegisterService[IAuthService](serviceContainer, NewAuthService())
```

### Controller 容器

依赖 Manager 和 Service 层。

```go
controllerContainer := container.NewControllerContainer(serviceContainer)
container.RegisterController[IMessageController](controllerContainer, NewMessageController())
```

### Middleware 容器

依赖 Manager 和 Service 层。

```go
middlewareContainer := container.NewMiddlewareContainer(serviceContainer)
container.RegisterMiddleware[IAuthMiddleware](middlewareContainer, NewAuthMiddleware())
```

## 依赖注入

### 基本用法

使用 `inject:""` 标签标记需要注入的字段：

```go
type MessageService struct {
    LoggerMgr loggermgr.ILoggerManager   `inject:""`
    DBManager databasemgr.IDatabaseManager `inject:""`
    Repo      IMessageRepository           `inject:""`
}
```

### 可选依赖

使用 `inject:"optional"` 标记可选依赖，注入失败不会报错：

```go
type MessageService struct {
    CacheManager cachemgr.ICacheManager `inject:"optional"`
}
```

## 容器 API

### EntityContainer

```go
func NewEntityContainer() *EntityContainer
func RegisterEntity[T common.IBaseEntity](e *EntityContainer, impl T) error
func (e *EntityContainer) GetByName(name string) (common.IBaseEntity, error)
func (e *EntityContainer) GetByType(typ reflect.Type) ([]common.IBaseEntity, error)
func (e *EntityContainer) GetAll() []common.IBaseEntity
func (e *EntityContainer) Count() int
```

### RepositoryContainer

```go
func NewRepositoryContainer(entity *EntityContainer) *RepositoryContainer
func RegisterRepository[T common.IBaseRepository](r *RepositoryContainer, impl T) error
func GetRepository[T common.IBaseRepository](r *RepositoryContainer) (T, error)
func (r *RepositoryContainer) SetManagerContainer(container *ManagerContainer)
func (r *RepositoryContainer) InjectAll() error
func (r *RepositoryContainer) GetByType(ifaceType reflect.Type) common.IBaseRepository
func (r *RepositoryContainer) GetAll() []common.IBaseRepository
func (r *RepositoryContainer) Count() int
```

### ServiceContainer

```go
func NewServiceContainer(repository *RepositoryContainer) *ServiceContainer
func RegisterService[T common.IBaseService](s *ServiceContainer, impl T) error
func GetService[T common.IBaseService](s *ServiceContainer) (T, error)
func (s *ServiceContainer) SetManagerContainer(container *ManagerContainer)
func (s *ServiceContainer) InjectAll() error
func (s *ServiceContainer) GetByType(ifaceType reflect.Type) common.IBaseService
func (s *ServiceContainer) GetAll() []common.IBaseService
func (s *ServiceContainer) Count() int
```

### ControllerContainer

```go
func NewControllerContainer(service *ServiceContainer) *ControllerContainer
func RegisterController[T common.IBaseController](c *ControllerContainer, impl T) error
func GetController[T common.IBaseController](c *ControllerContainer) (T, error)
func (c *ControllerContainer) SetManagerContainer(container *ManagerContainer)
func (c *ControllerContainer) InjectAll() error
func (c *ControllerContainer) GetByType(ifaceType reflect.Type) common.IBaseController
func (c *ControllerContainer) GetAll() []common.IBaseController
func (c *ControllerContainer) Count() int
```

### MiddlewareContainer

```go
func NewMiddlewareContainer(service *ServiceContainer) *MiddlewareContainer
func RegisterMiddleware[T common.IBaseMiddleware](m *MiddlewareContainer, impl T) error
func GetMiddleware[T common.IBaseMiddleware](m *MiddlewareContainer) (T, error)
func (m *MiddlewareContainer) SetManagerContainer(container *ManagerContainer)
func (m *MiddlewareContainer) InjectAll() error
func (m *MiddlewareContainer) GetByType(ifaceType reflect.Type) common.IBaseMiddleware
func (m *MiddlewareContainer) GetAll() []common.IBaseMiddleware
func (m *MiddlewareContainer) Count() int
```

## Manager 自动初始化

Engine 会按以下顺序自动初始化内置 Manager：

1. ConfigManager (`manager/configmgr`)
2. TelemetryManager (`manager/telemetrymgr`)
3. LoggerManager (`manager/loggermgr`)
4. DatabaseManager (`manager/databasemgr`)
5. CacheManager (`manager/cachemgr`)
6. LockManager (`manager/lockmgr`)
7. LimiterManager (`manager/limitermgr`)
8. MQManager (`manager/mqmgr`)

业务代码通过 `inject:""` 标签自动注入 Manager：

```go
type MessageService struct {
    Config    configmgr.IConfigManager    `inject:""`
    LoggerMgr loggermgr.ILoggerManager   `inject:""`
    DBManager databasemgr.IDatabaseManager `inject:""`
}
```

## 错误处理

### 错误类型

- `DependencyNotFoundError` - 依赖未找到
- `CircularDependencyError` - 循环依赖
- `AmbiguousMatchError` - 多重匹配
- `DuplicateRegistrationError` - 重复注册
- `InstanceNotFoundError` - 实例未找到
- `UninjectedFieldError` - 标记 `inject:""` 的字段注入后仍为 nil
- `InterfaceAlreadyRegisteredError` - 接口已注册
- `ImplementationDoesNotImplementInterfaceError` - 实现未实现接口
- `ManagerContainerNotSetError` - ManagerContainer 未设置

## 最佳实践

1. **使用泛型函数注册和获取** - `RegisterService[T]` / `GetService[T]` 比按类型注册更安全
2. **按依赖顺序初始化** - Entity → Repository → Service → Controller/Middleware
3. **Manager 不手动注册** - 由 Engine 自动初始化和注册
4. **避免循环依赖** - Service 之间的循环依赖会被检测到
5. **日志使用统一接口** - 通过 `ILoggerManager` 使用日志，避免使用 `log.Fatal` 等标准库函数
