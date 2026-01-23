# Container

依赖注入容器，支持分层架构的依赖管理。

## 特性

- **分层容器** - 支持 Entity、Manager、Repository、Service、Controller、Middleware 六层容器
- **类型安全** - 使用泛型确保类型安全，编译时检查
- **自动注入** - 通过结构体标签 `inject:""` 自动注入依赖
- **循环检测** - 使用拓扑排序检测循环依赖
- **线程安全** - 所有容器操作都使用读写锁保护

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

// 执行依赖注入
serviceContainer.InjectAll()
controllerContainer.InjectAll()
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

Manager 层负责管理基础组件（如配置、日志、数据库等）。

```go
type ConfigManager struct {
    // 管理器字段
}

func (m *ConfigManager) ManagerName() string {
    return "ConfigManager"
}

managerContainer := container.NewManagerContainer()
container.RegisterManager(managerContainer, &ConfigManager{})
```

### Repository 层

Repository 层依赖 Manager 和 Entity 层。

```go
type UserRepository struct {
    DBManager `inject:""`
    UserEntity `inject:""`
}

func (r *UserRepository) RepositoryName() string {
    return "UserRepository"
}

repositoryContainer := container.NewRepositoryContainer(entityContainer)
repositoryContainer.SetManagerContainer(managerContainer)
container.RegisterRepository(repositoryContainer, &UserRepository{})
```

### Service 层

Service 层依赖 Manager、Repository 和其他 Service 层。

```go
type UserService struct {
    ConfigManager `inject:""`
    UserRepository `inject:""`
    AuthService `inject:""` // 其他服务
}

func (s *UserService) ServiceName() string {
    return "UserService"
}

serviceContainer := container.NewServiceContainer(repositoryContainer)
serviceContainer.SetManagerContainer(managerContainer)
container.RegisterService(serviceContainer, &UserService{})
```

### Controller 层

Controller 层依赖 Manager 和 Service 层。

```go
type UserController struct {
    ConfigManager `inject:""`
    UserService `inject:""`
}

func (c *UserController) ControllerName() string {
    return "UserController"
}

controllerContainer := container.NewControllerContainer(serviceContainer)
controllerContainer.SetManagerContainer(managerContainer)
container.RegisterController(controllerContainer, &UserController{})
```

### Middleware 层

Middleware 层依赖 Manager 和 Service 层。

```go
type AuthMiddleware struct {
    ConfigManager `inject:""`
    UserService `inject:""`
}

func (m *AuthMiddleware) MiddlewareName() string {
    return "AuthMiddleware"
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

### 避免循环依赖

设计依赖关系时应避免循环依赖。如果存在循环依赖，注入时会返回 `CircularDependencyError`。

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
