# Container

 依赖注入容器，支持 5 层分层架构的自动依赖管理，交互层包含 4 种组件容器。

 ## 特性

 - **5 层容器** - 内置管理器层、Entity、Repository、Service、交互层（Controller/Middleware/Listener/Scheduler 4 种容器）
 - **类型安全** - 泛型注册与获取，编译时类型检查
 - **自动注入** - 通过 `inject:""` 结构体标签自动注入依赖
 - **拓扑排序** - Service 层使用 Kahn 算法检测并解决循环依赖
 - **线程安全** - 所有容器操作使用读写锁保护
 - **分层依赖** - 严格的依赖层级，交互层禁止直接注入 Repository
 - **Manager 自动初始化** - Engine 自动初始化并注册内置 Manager

## 快速开始

```go
package application

import (
	"github.com/lite-lake/litecore-go/container"
	"github.com/lite-lake/litecore-go/common"
	"github.com/lite-lake/litecore-go/server"
	entities "yourproject/internal/entities"
	repositories "yourproject/internal/repositories"
	services "yourproject/internal/services"
	controllers "yourproject/internal/controllers"
)

// InitEntityContainer 初始化实体容器
func InitEntityContainer() *container.EntityContainer {
	entityContainer := container.NewEntityContainer()
	container.RegisterEntity[common.IBaseEntity](entityContainer, &entities.Message{})
	return entityContainer
}

// InitRepositoryContainer 初始化仓储容器
func InitRepositoryContainer(entityContainer *container.EntityContainer) *container.RepositoryContainer {
	repositoryContainer := container.NewRepositoryContainer(entityContainer)
	container.RegisterRepository[repositories.IMessageRepository](repositoryContainer, repositories.NewMessageRepository())
	return repositoryContainer
}

// InitServiceContainer 初始化服务容器
func InitServiceContainer(repositoryContainer *container.RepositoryContainer) *container.ServiceContainer {
	serviceContainer := container.NewServiceContainer(repositoryContainer)
	container.RegisterService[services.IMessageService](serviceContainer, services.NewMessageService())
	return serviceContainer
}

// InitControllerContainer 初始化控制器容器
func InitControllerContainer(serviceContainer *container.ServiceContainer) *container.ControllerContainer {
	controllerContainer := container.NewControllerContainer(serviceContainer)
	container.RegisterController[controllers.IMessageController](controllerContainer, controllers.NewMessageController())
	return controllerContainer
}

// NewEngine 创建并启动应用引擎
func NewEngine() (*server.Engine, error) {
	entityContainer := InitEntityContainer()
	repositoryContainer := InitRepositoryContainer(entityContainer)
	serviceContainer := InitServiceContainer(repositoryContainer)
	controllerContainer := InitControllerContainer(serviceContainer)

	return server.NewEngine(
		&server.BuiltinConfig{Driver: "yaml", FilePath: "configs/config.yaml"},
		entityContainer, repositoryContainer, serviceContainer,
		controllerContainer, nil, nil, nil,
	), nil
}
```

 ## 分层架构

 容器支持 5 层分层架构，严格遵循依赖方向：

 ```
 ┌─────────────────────────────────────────────────────────────┐
 │                交互层 (Interaction Layer)                  │
 │  Controller/Middleware/Listener/Scheduler               │
 │  (HTTP 请求、MQ 消息、定时任务的统一处理)                    │
 ├─────────────────────────────────────────────────────────────┤
 │                    Service Layer                          │
 │              (业务逻辑和数据处理)                          │
 │            【支持服务间依赖 + 拓扑排序】                    │
 ├─────────────────────────────────────────────────────────────┤
 │                 Repository Layer                          │
 │              (数据访问和持久化)                            │
 ├─────────────────────────────────────────────────────────────┤
 │                    Entity Layer                            │
 │              (数据模型和领域对象)                           │
 └─────────────────────────────────────────────────────────────┘
             ↑                                              ↑
             └───────────────── Manager Layer ───────────────┘
  (configmgr、loggermgr、databasemgr、cachemgr、lockmgr、
   limitermgr、mqmgr、telemetrymgr、schedulermgr)
 ```

 ### 依赖规则

 | 层 | 可依赖的层 | 说明 |
 |---|---|---|
 | Entity | 无 | 纯数据模型，无依赖 |
 | Manager | 其他 Manager | 基础能力组件，可相互依赖 |
 | Repository | Manager + Entity | 数据访问层 |
 | Service | Manager + Repository + Service | 业务逻辑，支持服务间依赖 |
 | 交互层 (Controller) | Manager + Service | HTTP 请求处理 |
 | 交互层 (Middleware) | Manager + Service | 请求拦截器 |
 | 交互层 (Scheduler) | Manager + Service | 定时任务 |
 | 交互层 (Listener) | Manager + Service | 事件监听器 |

## 依赖注入

### 基本用法

使用 `inject:""` 标签标记需要注入的字段：

```go
import (
	"github.com/lite-lake/litecore-go/manager/configmgr"
	"github.com/lite-lake/litecore-go/manager/databasemgr"
	"github.com/lite-lake/litecore-go/manager/loggermgr"
)

type MessageServiceImpl struct {
	ConfigMgr   configmgr.IConfigManager    `inject:""`
	LoggerMgr   loggermgr.ILoggerManager   `inject:""`
	DBManager   databasemgr.IDatabaseManager `inject:""`
	MessageRepo IMessageRepository          `inject:""`
	AuthService IAuthService               `inject:""`
}

func (s *MessageServiceImpl) SomeMethod() error {
	s.logger = s.LoggerMgr.Ins()
	s.logger.Info("处理消息")
	// ...
}
```

### 注入顺序

依赖注入按以下优先级解析：

1. **本层容器** - 优先从当前容器查找（Service → Service）
2. **下层容器** - 从下层容器查找（Service → Repository）
3. **Manager 容器** - 从 Manager 容器查找（所有层 → Manager）

### Service 层拓扑排序

Service 层支持服务间依赖，并使用拓扑排序确保注入顺序：

```go
// ServiceA 依赖 ServiceB
type ServiceA struct {
	ServiceB IServiceB `inject:""`
}

// ServiceB 依赖 ServiceC
type ServiceB struct {
	ServiceC IServiceC `inject:""`
}

// ServiceC 无依赖
type ServiceC struct{}

// 注入顺序：ServiceC → ServiceB → ServiceA
serviceContainer.InjectAll()
```

如果存在循环依赖，系统会抛出 `CircularDependencyError`。

## 分层容器

### Entity 容器

Entity 层无依赖，仅存储数据实体。

```go
entityContainer := container.NewEntityContainer()
container.RegisterEntity[common.IBaseEntity](entityContainer, &Message{})

// 按名称获取
entity, err := entityContainer.GetByName("Message")

// 按类型获取（可返回多个）
entities, err := entityContainer.GetByType(reflect.TypeOf(&Message{}))

// 获取所有
all := entityContainer.GetAll()
```

### Manager 容器

Manager 容器存储管理器实例，由 Engine 自动初始化。

```go
managerContainer := container.NewManagerContainer()
container.RegisterManager[configmgr.IConfigManager](managerContainer, configMgr)

// 按类型获取
cfg, err := container.GetManager[configmgr.IConfigManager](managerContainer)

// 获取所有（按名称排序）
all := managerContainer.GetAllSorted()
```

### Repository 容器

Repository 层依赖 Manager 和 Entity 层。

```go
repositoryContainer := container.NewRepositoryContainer(entityContainer)
repositoryContainer.SetManagerContainer(managerContainer)

container.RegisterRepository[IMessageRepository](repositoryContainer, repo)

// 执行依赖注入
repositoryContainer.InjectAll()

// 按类型获取
repo, err := container.GetRepository[IMessageRepository](repositoryContainer)
```

### Service 容器

Service 层依赖 Manager、Repository 和其他 Service 层，支持拓扑排序。

```go
serviceContainer := container.NewServiceContainer(repositoryContainer)
serviceContainer.SetManagerContainer(managerContainer)

container.RegisterService[IMessageService](serviceContainer, messageService)
container.RegisterService[IAuthService](serviceContainer, authService)

// 执行依赖注入（自动拓扑排序）
serviceContainer.InjectAll()

// 按类型获取
svc, err := container.GetService[IMessageService](serviceContainer)
```

### Controller 容器

Controller 层依赖 Manager 和 Service 层。

**注意**：Controller 禁止直接注入 Repository，必须通过 Service 访问数据。

```go
controllerContainer := container.NewControllerContainer(serviceContainer)
container.RegisterController[IMessageController](controllerContainer, messageController)

// 执行依赖注入
controllerContainer.InjectAll()

// 按类型获取
ctrl, err := container.GetController[IMessageController](controllerContainer)
```

### Middleware 容器

Middleware 层依赖 Manager 和 Service 层。

**注意**：Middleware 禁止直接注入 Repository，必须通过 Service 访问数据。

```go
middlewareContainer := container.NewMiddlewareContainer(serviceContainer)
container.RegisterMiddleware[IAuthMiddleware](middlewareContainer, authMiddleware)

// 执行依赖注入
middlewareContainer.InjectAll()

// 按类型获取
mw, err := container.GetMiddleware[IAuthMiddleware](middlewareContainer)
```

### Scheduler 容器

Scheduler 层依赖 Manager 和 Service 层，用于定时任务。

**注意**：Scheduler 禁止直接注入 Repository，必须通过 Service 访问数据。

```go
schedulerContainer := container.NewSchedulerContainer(serviceContainer)
container.RegisterScheduler[ICleanupScheduler](schedulerContainer, cleanupScheduler)

// 执行依赖注入
schedulerContainer.InjectAll()

// 按类型获取
scheduler, err := container.GetScheduler[ICleanupScheduler](schedulerContainer)
```

### Listener 容器

Listener 层依赖 Manager 和 Service 层，用于事件监听。

**注意**：Listener 禁止直接注入 Repository，必须通过 Service 访问数据。

```go
listenerContainer := container.NewListenerContainer(serviceContainer)
container.RegisterListener[IMessageListener](listenerContainer, messageListener)

// 执行依赖注入
listenerContainer.InjectAll()

// 按类型获取
listener, err := container.GetListener[IMessageListener](listenerContainer)
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
func (e *EntityContainer) GetDependency(fieldType reflect.Type) (interface{}, error)
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
func (m *ManagerContainer) GetDependency(fieldType reflect.Type) (interface{}, error)
```

### RepositoryContainer

```go
func NewRepositoryContainer(entity *EntityContainer) *RepositoryContainer
func RegisterRepository[T common.IBaseRepository](r *RepositoryContainer, impl T) error
func GetRepository[T common.IBaseRepository](r *RepositoryContainer) (T, error)
func (r *RepositoryContainer) RegisterByType(ifaceType reflect.Type, impl common.IBaseRepository) error
func (r *RepositoryContainer) InjectAll() error
func (r *RepositoryContainer) GetByType(ifaceType reflect.Type) common.IBaseRepository
func (r *RepositoryContainer) GetAll() []common.IBaseRepository
func (r *RepositoryContainer) GetAllSorted() []common.IBaseRepository
func (r *RepositoryContainer) Count() int
func (r *RepositoryContainer) SetManagerContainer(container *ManagerContainer)
func (r *RepositoryContainer) GetDependency(fieldType reflect.Type) (interface{}, error)
```

### ServiceContainer

```go
func NewServiceContainer(repository *RepositoryContainer) *ServiceContainer
func RegisterService[T common.IBaseService](s *ServiceContainer, impl T) error
func GetService[T common.IBaseService](s *ServiceContainer) (T, error)
func (s *ServiceContainer) RegisterByType(ifaceType reflect.Type, impl common.IBaseService) error
func (s *ServiceContainer) InjectAll() error
func (s *ServiceContainer) GetByType(ifaceType reflect.Type) common.IBaseService
func (s *ServiceContainer) GetAll() []common.IBaseService
func (s *ServiceContainer) GetAllSorted() []common.IBaseService
func (s *ServiceContainer) Count() int
func (s *ServiceContainer) SetManagerContainer(container *ManagerContainer)
func (s *ServiceContainer) GetDependency(fieldType reflect.Type) (interface{}, error)
```

### ControllerContainer

```go
func NewControllerContainer(service *ServiceContainer) *ControllerContainer
func RegisterController[T common.IBaseController](c *ControllerContainer, impl T) error
func GetController[T common.IBaseController](c *ControllerContainer) (T, error)
func (c *ControllerContainer) InjectAll() error
func (c *ControllerContainer) RegisterByType(ifaceType reflect.Type, impl common.IBaseController) error
func (c *ControllerContainer) GetByType(ifaceType reflect.Type) common.IBaseController
func (c *ControllerContainer) GetAll() []common.IBaseController
func (c *ControllerContainer) GetAllSorted() []common.IBaseController
func (c *ControllerContainer) Count() int
func (c *ControllerContainer) SetManagerContainer(container *ManagerContainer)
func (c *ControllerContainer) GetDependency(fieldType reflect.Type) (interface{}, error)
```

### MiddlewareContainer

```go
func NewMiddlewareContainer(service *ServiceContainer) *MiddlewareContainer
func RegisterMiddleware[T common.IBaseMiddleware](m *MiddlewareContainer, impl T) error
func GetMiddleware[T common.IBaseMiddleware](m *MiddlewareContainer) (T, error)
func (m *MiddlewareContainer) InjectAll() error
func (m *MiddlewareContainer) RegisterByType(ifaceType reflect.Type, impl common.IBaseMiddleware) error
func (m *MiddlewareContainer) GetByType(ifaceType reflect.Type) common.IBaseMiddleware
func (m *MiddlewareContainer) GetAll() []common.IBaseMiddleware
func (m *MiddlewareContainer) GetAllSorted() []common.IBaseMiddleware
func (m *MiddlewareContainer) Count() int
func (m *MiddlewareContainer) SetManagerContainer(container *ManagerContainer)
func (m *MiddlewareContainer) GetDependency(fieldType reflect.Type) (interface{}, error)
```

### SchedulerContainer

```go
func NewSchedulerContainer(service *ServiceContainer) *SchedulerContainer
func RegisterScheduler[T common.IBaseScheduler](c *SchedulerContainer, impl T) error
func GetScheduler[T common.IBaseScheduler](c *SchedulerContainer) (T, error)
func (c *SchedulerContainer) InjectAll() error
func (c *SchedulerContainer) RegisterByType(ifaceType reflect.Type, impl common.IBaseScheduler) error
func (c *SchedulerContainer) GetByType(ifaceType reflect.Type) common.IBaseScheduler
func (c *SchedulerContainer) GetAll() []common.IBaseScheduler
func (c *SchedulerContainer) GetAllSorted() []common.IBaseScheduler
func (c *SchedulerContainer) Count() int
func (c *SchedulerContainer) SetManagerContainer(container *ManagerContainer)
func (c *SchedulerContainer) GetDependency(fieldType reflect.Type) (interface{}, error)
```

### ListenerContainer

```go
func NewListenerContainer(service *ServiceContainer) *ListenerContainer
func RegisterListener[T common.IBaseListener](l *ListenerContainer, impl T) error
func GetListener[T common.IBaseListener](l *ListenerContainer) (T, error)
func (l *ListenerContainer) InjectAll() error
func (l *ListenerContainer) RegisterByType(ifaceType reflect.Type, impl common.IBaseListener) error
func (l *ListenerContainer) GetByType(ifaceType reflect.Type) common.IBaseListener
func (l *ListenerContainer) GetAll() []common.IBaseListener
func (l *ListenerContainer) GetAllSorted() []common.IBaseListener
func (l *ListenerContainer) Count() int
func (l *ListenerContainer) SetManagerContainer(container *ManagerContainer)
func (l *ListenerContainer) GetDependency(fieldType reflect.Type) (interface{}, error)
```

## 错误处理

### 错误类型

| 错误类型 | 说明 |
|---------|------|
| `DependencyNotFoundError` | 依赖未找到 |
| `CircularDependencyError` | Service 层循环依赖 |
| `AmbiguousMatchError` | 多重匹配（Entity 层同类型多个实例） |
| `DuplicateRegistrationError` | 重复注册 |
| `InstanceNotFoundError` | 实例未找到 |
| `InterfaceAlreadyRegisteredError` | 接口已注册 |
| `ImplementationDoesNotImplementInterfaceError` | 实现未实现接口 |
| `InterfaceNotRegisteredError` | 接口未注册 |
| `ManagerContainerNotSetError` | ManagerContainer 未设置 |
| `UninjectedFieldError` | 标记 `inject:""` 的字段注入后仍为 nil |

### 错误处理示例

```go
svc, err := container.GetService[IMessageService](serviceContainer)
if err != nil {
	var notFound *container.InstanceNotFoundError
	if errors.As(err, &notFound) {
		log.Fatal("服务未注册:", notFound.Name)
	}
	log.Fatal("获取服务失败:", err)
}
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
	CacheMgr  cachemgr.ICacheManager     `inject:""`
}
```

## 最佳实践

### 1. 使用泛型函数注册和获取

泛型函数比按类型注册更安全，避免运行时错误：

```go
// 推荐
container.RegisterService[IMessageService](serviceContainer, svc)
svc, err := container.GetService[IMessageService](serviceContainer)

// 不推荐
serviceContainer.RegisterByType(reflect.TypeOf((*IMessageService)(nil)).Elem(), svc)
```

### 2. 按依赖顺序初始化

严格遵循分层架构顺序：

```go
entityContainer := container.NewEntityContainer()
repositoryContainer := container.NewRepositoryContainer(entityContainer)
serviceContainer := container.NewServiceContainer(repositoryContainer)
controllerContainer := container.NewControllerContainer(serviceContainer)
```

### 3. Manager 不手动注册

Manager 由 Engine 自动初始化和注册，业务代码只需声明依赖：

```go
type MyService struct {
	ConfigMgr configmgr.IConfigManager `inject:""`
}
```

### 4. 避免循环依赖

Service 层的循环依赖会被拓扑排序检测到：

```go
// 错误：循环依赖
type ServiceA struct { ServiceB IServiceB `inject:""` }
type ServiceB struct { ServiceA IServiceA `inject:""` } // ❌

// 正确：通过接口解耦
type ServiceA struct { IDataService IDataService `inject:""` }
type ServiceB struct { IDataService IDataService `inject:""` } // ✅
```

### 5. 遵循分层依赖规则

- Controller/Middleware/Scheduler/Listener 禁止直接注入 Repository
- 必须通过 Service 访问数据

```go
// 错误
type MyController struct {
	Repo IMessageRepository `inject:""` // ❌
}

// 正确
type MyController struct {
	Service IMessageService `inject:""` // ✅
}
```

### 6. 统一使用注入的日志

避免使用标准库 `log.Fatal`，统一使用注入的 `ILoggerManager`：

```go
type MyService struct {
	LoggerMgr loggermgr.ILoggerManager `inject:""`
	logger    loggermgr.ILogger
}

func (s *MyService) SomeMethod() error {
	s.logger = s.LoggerMgr.Ins()
	s.logger.Info("操作开始")
	// ...
}
```

### 7. 使用 CLI 工具生成容器代码

使用 `litecore` CLI 工具自动生成容器初始化代码：

```bash
litecore generate container
```

生成的代码位于 `internal/application/` 目录，自动处理所有容器的初始化和注册。
