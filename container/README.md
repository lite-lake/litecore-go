# Container - 分层依赖注入容器

提供分层依赖注入容器，严格约束架构分层并管理组件生命周期。

## 特性

- **分层架构** - 定义 Entity/Repository/Service/Controller/Middleware 五层容器，严格单向依赖
- **依赖注入** - 通过 inject 标签自动注入依赖，支持接口类型匹配和可选依赖
- **同层依赖** - Service 层支持同层依赖，自动拓扑排序确定注入顺序
- **类型安全** - 使用泛型 API 注册，编译时类型检查，接口实现校验
- **错误检测** - 自动检测循环依赖、依赖缺失、接口未实现等错误
- **并发安全** - 使用 RWMutex 保护容器内部状态，支持多线程并发读取

## 快速开始

```go
package main

import (
    "log"
    "reflect"

    "github.com/lite-lake/litecore-go/common"
    "github.com/lite-lake/litecore-go/container"
)

func main() {
    // 1. 创建容器（按依赖顺序）
    entityContainer := container.NewEntityContainer()
    repositoryContainer := container.NewRepositoryContainer(entityContainer)
    serviceContainer := container.NewServiceContainer(repositoryContainer)
    controllerContainer := container.NewControllerContainer(serviceContainer)
    middlewareContainer := container.NewMiddlewareContainer(serviceContainer)

    // 2. 注册实例（使用泛型 API）
    userEntity := &User{}
    container.RegisterEntity[User](entityContainer, userEntity)

    userRepo := &UserRepositoryImpl{}
    container.RegisterRepository[UserRepository](repositoryContainer, userRepo)

    userService := &UserServiceImpl{}
    container.RegisterService[UserService](serviceContainer, userService)

    userController := &UserControllerImpl{}
    container.RegisterController[UserController](controllerContainer, userController)

    // 3. 执行依赖注入（按层次从下到上）
    if err := repositoryContainer.InjectAll(); err != nil {
        log.Fatalf("Repository injection failed: %v", err)
    }
    if err := serviceContainer.InjectAll(); err != nil {
        log.Fatalf("Service injection failed: %v", err)
    }
    if err := controllerContainer.InjectAll(); err != nil {
        log.Fatalf("Controller injection failed: %v", err)
    }
    if err := middlewareContainer.InjectAll(); err != nil {
        log.Fatalf("Middleware injection failed: %v", err)
    }

    // 4. 获取实例使用
    userService, err := serviceContainer.GetByType(reflect.TypeOf((*UserService)(nil)).Elem())
    if err != nil {
        log.Fatalf("Get service failed: %v", err)
    }
    userService.Handle()
}
```

## 容器层次

系统定义以下五层容器，遵循单向依赖原则：

```
┌─────────────────────────────────────────────────────────────┐
│  Controller Layer   (BaseController)                        │
│  Middleware Layer   (BaseMiddleware)                        │
├─────────────────────────────────────────────────────────────┤
│  Service Layer      (BaseService)                           │
├─────────────────────────────────────────────────────────────┤
│  Repository Layer   (BaseRepository)                        │
├─────────────────────────────────────────────────────────────┤
│  Entity Layer       (BaseEntity)                            │
└─────────────────────────────────────────────────────────────┘
                        ↑ 内置组件
┌─────────────────────────────────────────────────────────────┐
│  Config & Manager  (内置组件，由引擎自动初始化)              │
└─────────────────────────────────────────────────────────────┘
```

| 层级       | 可依赖的层级                              |
| ---------- | ----------------------------------------- |
| Entity     | 无依赖                                    |
| Repository | Entity, Config, Manager (内置)           |
| Service    | Repository, Config, Manager (内置), 其他 Service |
| Controller | Service, Config, Manager (内置)         |
| Middleware | Service, Config, Manager (内置)         |

## 注册实例

### 泛型 API 注册（推荐）

使用泛型辅助函数注册，类型安全且代码简洁：

```go
// Entity 层
userEntity := &User{}
container.RegisterEntity[User](entityContainer, userEntity)

// Repository 层
userRepo := &UserRepositoryImpl{}
container.RegisterRepository[UserRepository](repositoryContainer, userRepo)

// Service 层
userService := &UserServiceImpl{}
container.RegisterService[UserService](serviceContainer, userService)

// Controller 层
userController := &UserControllerImpl{}
container.RegisterController[UserController](controllerContainer, userController)

// Middleware 层
authMiddleware := &AuthMiddlewareImpl{}
container.RegisterMiddleware[AuthMiddleware](middlewareContainer, authMiddleware)
```

### 注册规则

1. **接口唯一性**：每个接口类型只能注册一个实现
2. **实现校验**：注册时会检查实现是否真正实现了接口
3. **并发安全**：注册方法使用写锁，获取方法使用读锁

```go
// 错误：重复注册相同接口
err := container.RegisterService[UserService](serviceContainer, &UserServiceImpl{})
// 第二次注册会返回 InterfaceAlreadyRegisteredError

// 错误：实现未实现接口
err := container.RegisterService[UserService](serviceContainer, &InvalidServiceImpl{})
// 会返回 ImplementationDoesNotImplementInterfaceError
```

## 依赖注入

### 声明依赖

在结构体字段上使用 `inject` 标签声明需要注入的依赖，Config 和 Manager 由引擎自动注入：

```go
type UserServiceImpl struct {
    // 内置组件（由引擎自动注入）
    Config    common.BaseConfigProvider `inject:""`
    DBManager DatabaseManager           `inject:""`
    CacheMgr  CacheManager              `inject:""`

    // 业务依赖
    UserRepo  UserRepository            `inject:""`
    OrderSvc  OrderService              `inject:""`  // 同层依赖

    // 可选依赖
    OptionalService IOtherService       `inject:"optional"`
}

func (s *UserServiceImpl) ServiceName() string {
    return "user"
}
```

### 注入规则

1. **接口匹配**：字段类型是接口，注册实例实现了该接口
2. **精确匹配**：字段类型与注册实例类型完全一致
3. **唯一性要求**：匹配结果必须唯一，否则报错
4. **可选依赖**：使用 `inject:"optional"` 标记可选依赖，找不到时不报错

```go
type UserServiceImpl struct {
    Config   common.BaseConfigProvider `inject:""`
    CacheMgr CacheManager              `inject:"optional"` // 可选依赖
}
```

### 两阶段注入

容器采用 **注册-注入分离** 的两阶段模式：

**阶段 1：注册阶段 (RegisterByType)**

- 仅将实例加入容器的 items map（以接口类型为键）
- 不执行任何依赖注入操作
- 可按任意顺序注册
- 注册时检查接口唯一性和实现校验

**阶段 2：注入阶段 (InjectAll)**

- 遍历容器内所有已注册实例
- 反射解析实例字段，执行依赖注入
- 对于同层依赖，按拓扑顺序注入
- 检测循环依赖和缺失依赖，失败时报错

## 同层依赖

Service 层支持同层依赖，容器会自动构建依赖图并进行拓扑排序。

### 示例场景

```go
// UserService 依赖 OrderService，OrderService 依赖 PaymentService
type UserServiceImpl struct {
    OrderSvc OrderService `inject:""`
}

type OrderServiceImpl struct {
    PaymentSvc PaymentService `inject:""`
}

type PaymentServiceImpl struct {
    // 无同层依赖
}
```

### 注入顺序

```go
// 注册顺序不限
container.RegisterService[UserService](serviceContainer, &UserServiceImpl{})
container.RegisterService[OrderService](serviceContainer, &OrderServiceImpl{})
container.RegisterService[PaymentService](serviceContainer, &PaymentServiceImpl{})

// InjectAll 自动处理依赖顺序
// 内部执行流程：
// 1. 构建依赖图：UserService → OrderService → PaymentService
// 2. 拓扑排序：[PaymentService, OrderService, UserService]
// 3. 按顺序注入：先 PaymentService，再 OrderService，最后 UserService
err := serviceContainer.InjectAll()
```

## 获取实例

### 按接口类型获取

```go
// 使用 GetByType 获取实例
userService, err := serviceContainer.GetByType(reflect.TypeOf((*UserService)(nil)).Elem())
if err != nil {
    log.Fatal(err)
}

// 使用实例
userService.Handle()
```

### 获取所有实例

```go
// 获取所有 Service
allServices := serviceContainer.GetAll()
for _, svc := range allServices {
    log.Printf("Service: %s", svc.ServiceName())
}
```

## 错误处理

### 错误类型

```go
// 依赖缺失错误
type DependencyNotFoundError struct {
    InstanceName  string       // 当前实例名称
    FieldName     string       // 缺失依赖的字段名
    FieldType     reflect.Type // 期望的依赖类型
    ContainerType string       // 应该从哪个容器查找
}

// 循环依赖错误
type CircularDependencyError struct {
    Cycle []string // 循环依赖链
}

// 接口已被注册错误
type InterfaceAlreadyRegisteredError struct {
    InterfaceType reflect.Type // 接口类型
    ExistingImpl  interface{}  // 已存在的实现
    NewImpl       interface{}  // 新的实现
}

// 实现未实现接口错误
type ImplementationDoesNotImplementInterfaceError struct {
    InterfaceType  reflect.Type // 接口类型
    Implementation interface{}  // 实现
}

// 实例未找到错误
type InstanceNotFoundError struct {
    Name  string // 实例名称
    Layer string // 层级名称
}
```

### 错误处理示例

```go
if err := serviceContainer.InjectAll(); err != nil {
    switch e := err.(type) {
    case *container.DependencyNotFoundError:
        log.Fatalf("Missing dependency: %v.%s needs %v from %s",
            e.InstanceName, e.FieldName, e.FieldType, e.ContainerType)
    case *container.CircularDependencyError:
        log.Fatalf("Circular dependency detected: %s", strings.Join(e.Cycle, " → "))
    case *container.InterfaceAlreadyRegisteredError:
        log.Fatalf("Interface %v already registered", e.InterfaceType)
    default:
        log.Fatal(err)
    }
}
```

## API

### 泛型辅助函数（推荐）

```go
// Entity 层
func RegisterEntity[T common.BaseEntity](e *EntityContainer, impl T) error

// Repository 层
func RegisterRepository[T common.BaseRepository](r *RepositoryContainer, impl T) error

// Service 层
func RegisterService[T common.BaseService](s *ServiceContainer, impl T) error

// Controller 层
func RegisterController[T common.BaseController](c *ControllerContainer, impl T) error

// Middleware 层
func RegisterMiddleware[T common.BaseMiddleware](m *MiddlewareContainer, impl T) error
```

### EntityContainer

```go
func NewEntityContainer() *EntityContainer
func (e *EntityContainer) RegisterByType(ifaceType reflect.Type, impl common.BaseEntity) error
func (e *EntityContainer) InjectAll() error
func (e *EntityContainer) GetAll() []common.BaseEntity
func (e *EntityContainer) GetByType(typ reflect.Type) (common.BaseEntity, error)
func (e *EntityContainer) Count() int
```

### RepositoryContainer

```go
func NewRepositoryContainer(entity *EntityContainer) *RepositoryContainer
func (r *RepositoryContainer) RegisterByType(ifaceType reflect.Type, impl common.BaseRepository) error
func (r *RepositoryContainer) InjectAll() error
func (r *RepositoryContainer) GetAll() []common.BaseRepository
func (r *RepositoryContainer) GetByType(typ reflect.Type) (common.BaseRepository, error)
func (r *RepositoryContainer) Count() int
```

### ServiceContainer

```go
func NewServiceContainer(repository *RepositoryContainer) *ServiceContainer
func (s *ServiceContainer) RegisterByType(ifaceType reflect.Type, impl common.BaseService) error
func (s *ServiceContainer) InjectAll() error
func (s *ServiceContainer) GetAll() []common.BaseService
func (s *ServiceContainer) GetByType(typ reflect.Type) (common.BaseService, error)
func (s *ServiceContainer) Count() int
```

### ControllerContainer

```go
func NewControllerContainer(service *ServiceContainer) *ControllerContainer
func (c *ControllerContainer) RegisterByType(ifaceType reflect.Type, impl common.BaseController) error
func (c *ControllerContainer) InjectAll() error
func (c *ControllerContainer) GetAll() []common.BaseController
func (c *ControllerContainer) GetByType(typ reflect.Type) (common.BaseController, error)
func (c *ControllerContainer) Count() int
```

### MiddlewareContainer

```go
func NewMiddlewareContainer(service *ServiceContainer) *MiddlewareContainer
func (m *MiddlewareContainer) RegisterByType(ifaceType reflect.Type, impl common.BaseMiddleware) error
func (m *MiddlewareContainer) InjectAll() error
func (m *MiddlewareContainer) GetAll() []common.BaseMiddleware
func (m *MiddlewareContainer) GetByType(typ reflect.Type) (common.BaseMiddleware, error)
func (m *MiddlewareContainer) Count() int
```

## 最佳实践

### 1. 依赖声明为接口类型

```go
// 推荐：声明为接口类型
type UserServiceImpl struct {
    OrderSvc OrderService   `inject:""`
    UserRepo UserRepository `inject:""`
}

// 避免：声明为具体实现类型
type UserServiceImpl struct {
    OrderSvc *OrderServiceImpl `inject:""`
}
```

### 2. 避免循环依赖

```go
// 推荐：无环的依赖关系
UserService → OrderService → PaymentService

// 避免：循环依赖
UserService → OrderService → UserService (循环!)
```

### 3. 避免过度依赖

```go
// 推荐：依赖聚焦
type UserServiceImpl struct {
    DBManager  DatabaseManager      `inject:""`  // 内置组件
    UserRepo   UserRepository       `inject:""`
}

// 避免：依赖过多
type UserServiceImpl struct {
    Config     common.BaseConfigProvider `inject:""`  // 内置组件
    DBManager  DatabaseManager           `inject:""`  // 内置组件
    CacheMgr   CacheManager              `inject:""`  // 内置组件
    UserRepo   UserRepository            `inject:""`
    OrderRepo  OrderRepository           `inject:""`
    OrderSvc   OrderService              `inject:""`
    PaymentSvc PaymentService            `inject:""`
    // ... 更多依赖
}
```

### 4. 使用可选依赖

对于非必须的依赖，使用 `inject:"optional"` 标记：

```go
type UserServiceImpl struct {
    Config   common.BaseConfigProvider `inject:""`
    CacheMgr CacheManager              `inject:"optional"` // 可选依赖
}
```

## 并发安全

容器在访问时使用 `sync.RWMutex` 保护：

- **写入阶段**：应用启动时单线程顺序注册，无并发写入
- **读取阶段**：服务运行期间多线程并发读取

RegisterByType 使用写锁（Lock），GetByType/GetAll 使用读锁（RLock）。

## 性能考虑

1. **反射开销**：InjectAll 使用反射解析字段，仅在启动时执行一次；注册时使用反射校验接口实现
2. **拓扑排序**：复杂度 O(V + E)，V 为实例数量，E 为依赖边数
3. **并发读取**：注入完成后使用 RWMutex 保护，读取性能高
4. **类型索引**：使用 map[reflect.Type] 作为索引，查找复杂度 O(1)
