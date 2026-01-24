# 消息队列监听器技术需求文档 (TRD)

| 文档编号 | TRD-20260124 |
|---------|-------------|
| 文档名称 | 消息队列监听器 (Message Listener) |
| 版本     | 1.1 |
| 日期     | 2026-01-24 |
| 状态     | 草稿 |

## 1. 背景与目标

### 1.1 背景

当前 litecore-go 框架已支持消息队列功能（通过 `mqmgr.IMQManager`），提供消息发布和订阅能力。然而，消费端的使用较为原始，需要手动调用 `SubscribeWithCallback` 并管理订阅生命周期。

在业务系统中，存在大量异步处理场景（如消息审核、邮件发送、数据同步等），这些场景需要：
- 定义统一的监听器接口
- 自动注册和发现监听器
- 自动启动和停止订阅
- 依赖注入支持
- 与现有架构无缝集成

### 1.2 目标

引入 **Listener 层**，作为与 Controller 并列的独立层，用于处理消息队列的消费监听。实现以下目标：

1. **统一接口定义**：定义 `IBaseListener` 接口，规范监听器行为
2. **容器化管理**：实现 `ListenerContainer`，统一管理所有监听器
3. **自动发现**：通过 CLI 工具自动扫描 `listeners/` 目录，生成注册代码
4. **依赖注入**：支持注入 Manager、Service，复用现有 DI 机制
5. **生命周期管理**：在 Engine 启动/停止时自动启动/停止所有监听器
6. **配置驱动**：通过配置文件控制队列订阅行为

### 1.3 非目标

- 不提供新的消息队列实现（继续使用 Memory/RabbitMQ）
- 不改变现有的依赖注入架构
- 不修改 MQManager 的接口和行为

## 2. 设计方案

### 2.1 架构设计

#### 2.1.1 六层架构

引入 Listener 层后，架构从 5 层扩展为 6 层：

```
┌─────────────────────────────────────────────────────────┐
│                        外部请求                           │
└─────────────────────────────────────────────────────────┘
                            │
                   ┌────────▼────────┐
                   │   Middleware    │  HTTP 请求处理
                   └────────┬────────┘
                            │
         ┌──────────────────┼──────────────────┐
         │                  │                  │
 ┌───────▼────────┐ ┌───────▼────────┐ ┌────────▼────────┐
 │   Controller   │ │    Listener     │ │                 │
 │                │ │   (新增)        │ │                 │
 │ HTTP 路由处理  │ │ MQ 消息消费     │ │                 │
 └────────┬───────┘ └────────┬────────┘ │                 │
          │                  │          │                 │
          └──────────────────┼──────────┘                 │
                             │                            │
                   ┌─────────▼────────┐                   │
                   │     Service      │  业务逻辑          │
                   └─────────┬────────┘                   │
                             │                            │
              ┌──────────────┼──────────────┐              │
              │              │              │              │
    ┌─────────▼────────┐ ┌───▼────────┐ ┌──▼──────────────▼────┐
    │   Repository     │ │   Manager  │ │                    │
    │                  │ │            │ │                    │
    │ 数据访问         │ │ 配置/缓存  │ │                    │
    └────────┬─────────┘ │ 日志等     │ │                    │
             │           └────────────┘ │                    │
             │                         │                    │
    ┌────────▼────────┐                │                    │
    │     Entity     │                │                    │
    │   数据模型      │                │                    │
    └────────────────┘                └────────────────────┘
```

#### 2.1.2 依赖关系

**允许的依赖方向**：

| 依赖层级 | 可依赖的层级 |
|---------|-------------|
| Listener | Manager, Service |
| Controller | Manager, Service |
| Middleware | Manager, Service |
| Service | Manager, Repository, Entity |
| Repository | Manager, Entity |
| Entity | 无 |

**依赖规则说明**：
- **Listener 只能依赖 Manager 和 Service**，不能直接访问 Repository
- **Service 是唯一可以访问 Repository 的层**，封装所有数据访问逻辑
- **Controller 和 Listener 同级**，分别处理 HTTP 请求和 MQ 消息，但都不能跨层访问 Repository
- **Controller、Middleware、Listener 都不能注入 Repository**（统一架构规则，由依赖注入检查强制执行）
- **遵循分层架构原则**：上层通过 Service 访问数据，避免绕过业务逻辑层

**Listener 的依赖注入规则**：
- ✅ 可注入：所有 Manager（MQManager, LoggerManager, DatabaseManager 等）
- ✅ 可注入：所有 Service
- ✅ 可注入：同层其他 Listener（需注意循环依赖）
- ❌ 不可注入：Repository（违反分层架构，必须通过 Service 访问数据）
- ❌ 不可注入：Controller、Middleware（避免与 HTTP 处理层耦合）

### 2.2 接口定义

#### 2.2.1 基础接口

**文件位置**：`common/base_listener.go`

```go
package common

import (
    "context"
    "github.com/lite-lake/litecore-go/manager/mqmgr"
)

// IBaseListener 基础监听器接口
// 所有 Listener 类必须继承此接口并实现相关方法
// 用于定义监听器的基础行为和契约
type IBaseListener interface {
    // ListenerName 返回监听器名称
    // 格式：xxxListenerImpl（小驼峰，带 Impl 后缀）
    ListenerName() string

    // GetQueue 返回监听的队列名称
    // 返回值示例："message.created", "user.registered"
    GetQueue() string

    // GetSubscribeOptions 返回订阅选项
    // 可配置是否持久化、是否自动确认、并发消费者数量等
    GetSubscribeOptions() []mqmgr.SubscribeOption

    // Handle 处理队列消息
    // ctx: 上下文
    // msg: 消息对象，包含 ID、Body、Headers
    // 返回: 处理错误（返回 error 会触发 Nack）
    Handle(ctx context.Context, msg mqmgr.Message) error

    // OnStart 在服务器启动时触发
    OnStart() error
    // OnStop 在服务器停止时触发
    OnStop() error
}
```

#### 2.2.2 监听器示例

**文件位置**：`samples/messageboard/internal/listeners/message_created_listener.go`

```go
package listeners

import (
    "context"
    "github.com/lite-lake/litecore-go/common"
    "github.com/lite-lake/litecore-go/manager/loggermgr"
    "github.com/lite-lake/litecore-go/manager/mqmgr"
    "github.com/lite-lake/litecore-go/samples/messageboard/internal/services"
)

// IMessageCreatedListener 消息创建监听器接口
type IMessageCreatedListener interface {
    common.IBaseListener
}

type messageCreatedListenerImpl struct {
    MessageService services.IMessageService `inject:""`
    LoggerMgr      loggermgr.ILoggerManager `inject:""`
}

// NewMessageCreatedListener 创建监听器实例
func NewMessageCreatedListener() IMessageCreatedListener {
    return &messageCreatedListenerImpl{}
}

// ListenerName 返回监听器名称
func (l *messageCreatedListenerImpl) ListenerName() string {
    return "messageCreatedListenerImpl"
}

// GetQueue 返回监听的队列名称
func (l *messageCreatedListenerImpl) GetQueue() string {
    return "message.created"
}

// GetSubscribeOptions 返回订阅选项
func (l *messageCreatedListenerImpl) GetSubscribeOptions() []mqmgr.SubscribeOption {
    return []mqmgr.SubscribeOption{
        mqmgr.WithAutoAck(true),          // 自动确认：返回 nil 自动 Ack，返回 error 自动 Nack
        mqmgr.WithSubscribeDurable(true),  // 持久化队列
        mqmgr.WithConcurrency(5),          // 并发消费者数量（可选，默认1）
    }
}

// OnStart 启动时初始化
func (l *messageCreatedListenerImpl) OnStart() error {
    return nil
}

// OnStop 停止时清理
func (l *messageCreatedListenerImpl) OnStop() error {
    return nil
}

// Handle 处理队列消息
// 返回 nil: 消息处理成功，自动 Ack（从队列移除）
// 返回 error: 消息处理失败，自动 Nack（重新入队重试）
func (l *messageCreatedListenerImpl) Handle(ctx context.Context, msg mqmgr.Message) error {
    l.LoggerMgr.Ins().Info("收到消息创建事件",
        "message_id", msg.ID(),
        "body", string(msg.Body()),
        "headers", msg.Headers())

    // 业务处理逻辑...
    // 处理成功返回 nil，处理失败返回 error

    return nil
}

var _ IMessageCreatedListener = (*messageCreatedListenerImpl)(nil)
var _ common.IBaseListener = (*messageCreatedListenerImpl)(nil)
```

### 2.3 容器实现

#### 2.3.1 ListenerContainer

**文件位置**：`container/listener_container.go`

```go
package container

import (
    "reflect"

    "github.com/lite-lake/litecore-go/common"
)

// ListenerContainer 监听器层容器
type ListenerContainer struct {
    *InjectableLayerContainer[common.IBaseListener]
    serviceContainer *ServiceContainer
}

// NewListenerContainer 创建新的监听器容器
// 与 ControllerContainer 一致，通过 SetManagerContainer 方法设置 ManagerContainer
func NewListenerContainer(service *ServiceContainer) *ListenerContainer {
    return &ListenerContainer{
        InjectableLayerContainer: NewInjectableLayerContainer(func(l common.IBaseListener) string {
            return l.ListenerName()
        }),
        serviceContainer: service,
    }
}

// SetManagerContainer 设置管理器容器
func (l *ListenerContainer) SetManagerContainer(container *ManagerContainer) {
    l.InjectableLayerContainer.SetManagerContainer(container)
}

// RegisterListener 泛型注册函数，按接口类型注册
func RegisterListener[T common.IBaseListener](c *ListenerContainer, impl T) error {
    ifaceType := reflect.TypeOf((*T)(nil)).Elem()
    return c.RegisterByType(ifaceType, impl)
}

// GetListener 按接口类型获取
func GetListener[T common.IBaseListener](c *ListenerContainer) (T, error) {
    ifaceType := reflect.TypeOf((*T)(nil)).Elem()
    impl := c.GetByType(ifaceType)
    if impl == nil {
        var zero T
        return zero, &InstanceNotFoundError{
            Name:  ifaceType.Name(),
            Layer: "Listener",
        }
    }
    return impl.(T), nil
}

// InjectAll 执行依赖注入
func (c *ListenerContainer) InjectAll() error {
    c.checkManagerContainer("Listener")

    if c.InjectableLayerContainer.base.container.IsInjected() {
        return nil
    }

    c.InjectableLayerContainer.base.sources = c.InjectableLayerContainer.base.buildSources(c, c.managerContainer, c.serviceContainer)
    return c.InjectableLayerContainer.base.injectAll(c)
}

// GetDependency 根据类型获取依赖实例（实现ContainerSource接口）
func (c *ListenerContainer) GetDependency(fieldType reflect.Type) (interface{}, error) {
    if dep, err := resolveDependencyFromManager(fieldType, c.managerContainer); dep != nil || err != nil {
        return dep, err
    }

    baseServiceType := reflect.TypeOf((*common.IBaseService)(nil)).Elem()
    if fieldType == baseServiceType || fieldType.Implements(baseServiceType) {
        if c.serviceContainer == nil {
            return nil, &DependencyNotFoundError{
                FieldType:     fieldType,
                ContainerType: "Service",
            }
        }
        impl := c.serviceContainer.GetByType(fieldType)
        if impl == nil {
            return nil, &DependencyNotFoundError{
                FieldType:     fieldType,
                ContainerType: "Service",
            }
        }
        return impl, nil
    }

    // Listener cannot directly inject Repository
    // Must access data through Service layer, following layered architecture principles
    baseRepositoryType := reflect.TypeOf((*common.IBaseRepository)(nil)).Elem()
    if fieldType == baseRepositoryType || fieldType.Implements(baseRepositoryType) {
        return nil, &DependencyNotFoundError{
            FieldType:     fieldType,
            ContainerType: "Repository",
            Message:       "Listener cannot directly inject Repository, must access data through Service",
        }
    }

    return nil, nil
}
```

#### 2.3.2 自动生成代码

**文件位置**：`samples/messageboard/internal/application/listener_container.go`（CLI 生成）

```go
// Code generated by litecore/cli. DO NOT EDIT.
package application

import (
    "github.com/lite-lake/litecore-go/container"
    listeners "github.com/lite-lake/litecore-go/samples/messageboard/internal/listeners"
)

// InitListenerContainer 初始化监听器容器
func InitListenerContainer(serviceContainer *container.ServiceContainer) *container.ListenerContainer {
    listenerContainer := container.NewListenerContainer(serviceContainer)

    // 使用泛型注册函数注册监听器
    container.RegisterListener[listeners.IMessageCreatedListener](listenerContainer, listeners.NewMessageCreatedListener())
    container.RegisterListener[listeners.IMessageAuditListener](listenerContainer, listeners.NewMessageAuditListener())

    return listenerContainer
}
```

**泛型注册说明**：
- `RegisterListener[T]` 是泛型注册函数，按接口类型注册监听器实现
- `T` 是监听器接口类型（如 `IMessageCreatedListener`）
- 第二个参数是监听器实现实例（通过 `NewXXXListener()` 创建）
- 自动生成的代码会遍历 `listeners/` 目录，为每个监听器生成注册语句

### 2.4 Engine 扩展

#### 2.4.1 Engine 结构体修改

**文件位置**：`server/engine.go`

```go
// Engine 服务引擎
type Engine struct {
    // ... 现有字段

    // Listener 监听器容器（新增）
    Listener *container.ListenerContainer
}
```

#### 2.4.2 NewEngine 构造函数修改

```go
func NewEngine(
    builtinConfig *BuiltinConfig,
    entity *container.EntityContainer,
    repository *container.RepositoryContainer,
    service *container.ServiceContainer,
    controller *container.ControllerContainer,
    middleware *container.MiddlewareContainer,
    listener *container.ListenerContainer,  // 新增参数
) *Engine {
    // ... 现有逻辑

    return &Engine{
        // ... 现有字段
        Listener: listener,
        // ...
    }
}
```

#### 2.4.3 autoInject 方法扩展

```go
// autoInject 自动依赖注入
func (e *Engine) autoInject() error {
    e.logPhaseStart(PhaseInjection, "开始依赖注入")

    // 1. Entity 层

    // 2. Repository 层
    // ... 现有代码

    // 3. Service 层
    // ... 现有代码

    // 4. Controller 层
    // ... 现有代码

    // 5. Listener 层（新增）
    e.Listener.SetManagerContainer(e.Manager)
    if err := e.Listener.InjectAll(); err != nil {
        return fmt.Errorf("listener inject failed: %w", err)
    }
    listeners := e.Listener.GetAll()
    for _, listener := range listeners {
        e.logStartup(PhaseInjection, fmt.Sprintf("[%s 层] %s: 注入完成", "Listener", listener.ListenerName()))
    }

    // 6. Middleware 层
    // ... 现有代码

    totalCount := len(repos) + len(svcs) + len(ctrls) + len(listeners) + len(mws)
    e.logPhaseEnd(PhaseInjection, "依赖注入完成", logger.F("count", totalCount))

    return nil
}
```

#### 2.4.4 启动监听器

**文件位置**：`server/lifecycle.go`

```go
// startListeners 启动所有监听器
func (e *Engine) startListeners() error {
    e.logPhaseStart(PhaseStartup, "Starting Listener layer")

    listeners := e.Listener.GetAll()
    if len(listeners) == 0 {
        e.getLogger().Info("No listeners registered, skipping Listener startup")
        return nil
    }

    mqManager, err := container.GetManager[mqmgr.IMQManager](e.Manager)
    if err != nil {
        return fmt.Errorf("MQManager not initialized but %d listeners exist: %w", len(listeners), err)
    }

    startedCount := 0

    for _, listener := range listeners {
        queue := listener.GetQueue()
        opts := listener.GetSubscribeOptions()

        e.getLogger().Info("Starting message listener",
            "listener", listener.ListenerName(),
            "queue", queue)

        err := mqManager.SubscribeWithCallback(
            e.ctx,
            queue,
            listener.Handle,
            opts...,
        )
        if err != nil {
            return fmt.Errorf("failed to start listener %s: %w", listener.ListenerName(), err)
        }

        e.logStartup(PhaseStartup, listener.ListenerName()+": started")
        startedCount++
    }

    e.logPhaseEnd(PhaseStartup, "Listener layer startup completed", logger.F("count", startedCount))
    return nil
}
```

#### 2.4.5 停止监听器

```go
// stopListeners 停止所有监听器
func (e *Engine) stopListeners() []error {
    listeners := e.Listener.GetAll()
    var errors []error

    for i := len(listeners) - 1; i >= 0; i-- {
        listener := listeners[i]
        if err := e.unsubscribeListener(listener); err != nil {
            errors = append(errors, fmt.Errorf("failed to stop listener %s: %w", listener.ListenerName(), err))
        }
    }

    return errors
}

// unsubscribeListener 取消监听器订阅
func (e *Engine) unsubscribeListener(listener common.IBaseListener) error {
    mqManager, err := container.GetManager[mqmgr.IMQManager](e.Manager)
    if err != nil {
        return fmt.Errorf("get MQManager failed: %w", err)
    }

    // MQManager will automatically cancel all subscriptions when Close() is called
    // Here we mainly log the event
    e.getLogger().Info("Stopping message listener", "listener", listener.ListenerName())
    return nil
}
```

#### 2.4.6 Start 方法修改

```go
func (e *Engine) Start() error {
    e.mu.Lock()
    defer e.mu.Unlock()

    if e.started {
        return fmt.Errorf("engine already started")
    }

    // 1. 启动所有 Manager

    // 2. 启动所有 Repository

    // 3. 启动所有 Service
    if err := e.startServices(); err != nil {
        return fmt.Errorf("start services failed: %w", err)
    }

    // 4. 启动所有 Middleware
    if err := e.startMiddlewares(); err != nil {
        return fmt.Errorf("start middlewares failed: %w", err)
    }

    // 停止异步日志器
    if e.asyncLogger != nil {
        e.asyncLogger.Stop()
        e.asyncLogger = nil
    }

    // 5. 启动所有 Listener（新增）
    if err := e.startListeners(); err != nil {
        return fmt.Errorf("start listeners failed: %w", err)
    }

    // 6. 启动 HTTP 服务器
    // ...
}
```

#### 2.4.7 Stop 方法修改

```go
func (e *Engine) Stop() error {
    e.mu.Lock()
    defer e.mu.Unlock()

    if !e.started {
        return nil
    }

    e.logStartup(PhaseShutdown, "HTTP 服务器关闭...")

    ctx, cancel := context.WithTimeout(context.Background(), e.shutdownTimeout)
    defer cancel()

    if e.httpServer != nil {
        if err := e.httpServer.Shutdown(ctx); err != nil {
            return fmt.Errorf("HTTP server shutdown error: %w", err)
        }
    }

    e.logPhaseStart(PhaseShutdown, "开始停止各层组件")

    // 按相反顺序停止各层
    middlewareErrors := e.stopMiddlewares()
    e.logStartup(PhaseShutdown, "Middleware 层停止完成")

    listenerErrors := e.stopListeners()
    e.logStartup(PhaseShutdown, "Listener 层停止完成")

    serviceErrors := e.stopServices()
    e.logStartup(PhaseShutdown, "Service 层停止完成")

    repositoryErrors := e.stopRepositories()
    e.logStartup(PhaseShutdown, "Repository 层停止完成")

    managerErrors := e.stopManagers()
    e.logStartup(PhaseShutdown, "Manager 层停止完成")

    allErrors := make([]error, 0)
    allErrors = append(allErrors, middlewareErrors...)
    allErrors = append(allErrors, listenerErrors...)
    allErrors = append(allErrors, serviceErrors...)
    allErrors = append(allErrors, repositoryErrors...)
    allErrors = append(allErrors, managerErrors...)

    // ... 错误处理逻辑
}
```

### 2.5 CLI 工具扩展

#### 2.5.1 analyzer 扩展

**文件位置**：`cli/analyzer/analyzer.go`

```go
const (
    LayerEntity     Layer = "entity"
    LayerRepository Layer = "repository"
    LayerService    Layer = "service"
    LayerController Layer = "controller"
    LayerMiddleware Layer = "middleware"
    LayerListener   Layer = "listener"  // 新增
)

// detectLayer 检测代码层
func (a *Analyzer) detectLayer(filename, packageName string) Layer {
    parts := strings.FieldsFunc(filename, func(r rune) bool {
        return r == '/' || r == '\\'
    })

    for _, part := range parts {
        if strings.Contains(part, "entities") {
            return LayerEntity
        }
        if strings.Contains(part, "repositories") {
            return LayerRepository
        }
        if strings.Contains(part, "services") {
            return LayerService
        }
        if strings.Contains(part, "controllers") {
            return LayerController
        }
        if strings.Contains(part, "middlewares") {
            return LayerMiddleware
        }
        if strings.Contains(part, "listeners") {  // 新增
            return LayerListener
        }
    }

    return ""
}

// IsLitecoreLayer 判断是否为 Litecore 标准层
func IsLitecoreLayer(layer Layer) bool {
    switch layer {
    case LayerEntity, LayerRepository, LayerService,
        LayerController, LayerMiddleware, LayerListener:  // 新增
        return true
    default:
        return false
    }
}

// GetBaseInterface 获取层对应的基础接口
func GetBaseInterface(layer Layer) string {
    switch layer {
    case LayerEntity:
        return "BaseEntity"
    case LayerRepository:
        return "BaseRepository"
    case LayerService:
        return "BaseService"
    case LayerController:
        return "BaseController"
    case LayerMiddleware:
        return "BaseMiddleware"
    case LayerListener:  // 新增
        return "IBaseListener"
    default:
        return ""
    }
}

// GetContainerName 获取容器名称
func GetContainerName(layer Layer) string {
    switch layer {
    case LayerEntity:
        return "EntityContainer"
    case LayerRepository:
        return "RepositoryContainer"
    case LayerService:
        return "ServiceContainer"
    case LayerController:
        return "ControllerContainer"
    case LayerMiddleware:
        return "MiddlewareContainer"
    case LayerListener:  // 新增
        return "ListenerContainer"
    default:
        return ""
    }
}

// GetRegisterFunction 获取注册函数名
func GetRegisterFunction(layer Layer) string {
    switch layer {
    case LayerEntity:
        return "RegisterEntity"
    case LayerRepository:
        return "RegisterRepository"
    case LayerService:
        return "RegisterService"
    case LayerController:
        return "RegisterController"
    case LayerMiddleware:
        return "RegisterMiddleware"
    case LayerListener:  // 新增
        return "RegisterListener"
    default:
        return "Register"
    }
}
```

#### 2.5.2 template 扩展

在 CLI 模板中添加 Listener 容器初始化代码生成。

## 3. 实现细节

### 3.1 消息处理流程

```
┌────────────────────────────────────────────────────────────┐
│  1. 应用启动                                                │
│  └─> Engine.Start()                                        │
│      └─> startListeners()                                  │
│          └─> foreach listener:                             │
│              MQManager.SubscribeWithCallback(queue, handle)  │
└────────────────────────────────────────────────────────────┘
                           │
                           ▼
┌────────────────────────────────────────────────────────────┐
│  2. 消息到达                                                │
│  └─> MQManager 接收消息                                    │
│      └─> 调用 listener.Handle(ctx, msg)                    │
└────────────────────────────────────────────────────────────┘
                           │
                           ▼
┌────────────────────────────────────────────────────────────┐
│  3. 业务处理                                                │
│  └─> listener.Handle() 方法执行                            │
│      └─> 使用注入的 Service 处理业务逻辑                    │
│          └─> Service 内部通过 Repository 访问数据        │
└────────────────────────────────────────────────────────────┘
                           │
                           ▼
┌────────────────────────────────────────────────────────────┐
│  4. 返回结果                                                │
│  └─> return nil   ──> 自动 Ack（AutoAck=true）            │
│  └─> return error ──> 自动 Nack 并重新入队                │
└────────────────────────────────────────────────────────────┘
```

### 3.2 错误处理

**Handle 方法返回值的处理**：

| 场景                | 处理方式                    | 说明                          |
|---------------------|----------------------------|-------------------------------|
| Handle返回nil       | AutoAck=true时自动确认      | 消息成功处理，从队列移除      |
| Handle返回error     | Nack并立即重新入队           | 失败后立即重试，失败N次后抛弃  |
| Handle发生panic     | 捕获panic，记录error日志     | 防止监听器崩溃，继续处理后续  |
| 连续失败N次         | 发送到死信队列或直接抛弃     | 避免无限重试导致队列堵塞      |

**错误处理策略**：

1. **立即重试机制**
   - Handle返回error时，MQManager自动调用`Nack(ctx, msg, true)`将消息重新入队
   - 消息立即回到队列头部，等待下次消费

2. **失败次数限制**
   - 消息头部维护重试次数计数（x-retry-count）
   - 每次失败后计数+1，达到阈值（默认3次）后不再重试
   - 失败N次后：发送到死信队列（如果有）或直接抛弃

3. **Panic捕获**
   - MQManager在调用Handle方法时使用recover()捕获panic
   - Panic被捕获后记录error级别日志，记录panic信息和堆栈
   - Panic等同于返回error，触发重新入队机制

**重试次数示例**：
```go
// 在消息头中维护重试次数
retryCount := 0
if count, ok := msg.Headers()["x-retry-count"].(int); ok {
    retryCount = count
}

if retryCount >= maxRetryCount {
    // 超过最大重试次数，发送到死信队列或抛弃
    return fmt.Errorf("max retry count (%d) exceeded", maxRetryCount)
}

// 重试前增加计数
msg.Headers()["x-retry-count"] = retryCount + 1
```

### 3.3 配置管理

Listener 的订阅行为通过 `GetSubscribeOptions()` 方法配置，不需要额外的配置文件。

**订阅选项说明**：

| 选项                | 类型   | 默认值 | 说明                          |
|---------------------|-------|--------|-------------------------------|
| WithAutoAck         | bool  | true   | 是否自动确认消息（推荐 true）  |
| WithSubscribeDurable| bool  | false  | 队列是否持久化                 |
| WithConcurrency     | int   | 1      | 并发消费者数量（工作池大小）   |

但需要确保 MQManager 已正确配置：

```yaml
# configs/config.yaml
mq:
  driver: "rabbitmq"  # 或 "memory"
  rabbitmq_config:
    url: "amqp://guest:guest@localhost:5672/"
    durable: true
  memory_config:
    max_queue_size: 10000
    channel_buffer: 100
```

**并发消费者说明**：

- 使用 `WithConcurrency(n)` 配置单个队列的并发消费者数量
- MQManager会创建一个大小为n的工作池（worker pool）
- 每个worker独立调用 `listener.Handle(ctx, msg)` 处理消息
- 并发消费可以提高消息处理吞吐量，但需要注意：
  - 消息处理顺序不保证（并发时）
  - 需确保 `Handle()` 方法是线程安全的
  - 避免过大的并发数导致资源耗尽

### 3.4 日志记录

Listener 应使用注入的 `LoggerManager` 记录日志，遵循框架的日志使用规范：

```go
type myListenerImpl struct {
    LoggerMgr loggermgr.ILoggerManager `inject:""`
}

func (l *myListenerImpl) Handle(ctx context.Context, msg mqmgr.Message) error {
    l.LoggerMgr.Ins().Info("开始处理消息",
        "queue", l.GetQueue(),
        "msg_id", msg.ID())

    // 业务逻辑...

    l.LoggerMgr.Ins().Info("消息处理完成", "msg_id", msg.ID())
    return nil
}
```

**日志级别使用建议**：
- **Debug**：详细的消息内容、中间处理步骤
- **Info**：消息接收、处理完成、业务关键节点
- **Warn**：重试、降级处理、格式异常
- **Error**：处理失败、依赖服务异常
- **Fatal**：严重错误导致监听器无法继续运行

### 3.5 测试策略

#### 3.5.1 单元测试

```go
func TestMessageCreatedListener_Handle(t *testing.T) {
    // 创建 Mock 依赖
    mockService := &mockMessageService{}
    mockLogger := &mockLoggerManager{}

    // 创建监听器
    listener := &messageCreatedListenerImpl{
        MessageService: mockService,
        LoggerMgr:      mockLogger,
    }

    // 创建测试消息
    msg := &mockMessage{
        id:   "msg-123",
        body: []byte(`{"id": "123", "content": "test"}`),
    }

    // 调用 Handle
    err := listener.Handle(context.Background(), msg)

    // 断言
    assert.NoError(t, err)
    mockService.AssertCalled(t, "ProcessMessage", msg)
}
```

#### 3.5.2 集成测试

使用真实的 MQManager（Memory 实现）进行测试：

```go
func TestListenerIntegration(t *testing.T) {
    // 创建 MQManager
    mqConfig := &mqmgr.MemoryConfig{
        MaxQueueSize:  100,
        ChannelBuffer: 10,
    }
    mqMgr := mqmgr.NewMessageQueueManagerMemoryImpl(mqConfig)
    defer mqMgr.Close()

    // 创建监听器容器
    listenerContainer := container.NewListenerContainer(nil)
    container.RegisterListener[IMessageCreatedListener](
        listenerContainer,
        NewMessageCreatedListener(),
    )

    // 启动监听
    listeners := listenerContainer.GetAll()
    for _, listener := range listeners {
        opts := listener.GetSubscribeOptions()
        err := mqMgr.SubscribeWithCallback(
            context.Background(),
            listener.GetQueue(),
            listener.Handle,
            opts...,
        )
        assert.NoError(t, err)
    }

    // 发布消息
    err := mqMgr.Publish(context.Background(), "message.created", []byte("test"))
    assert.NoError(t, err)

    // 等待消息被消费
    time.Sleep(100 * time.Millisecond)
}
```

## 4. 目录结构

```
litecore-go/
├── common/
│   └── base_listener.go          # 新增：IBaseListener 接口定义
├── container/
│   ├── base_container.go         # 修改：添加 Listener 相关类型
│   └── listener_container.go     # 新增：ListenerContainer 实现
├── server/
│   ├── engine.go                  # 修改：添加 Listener 字段和方法
│   └── lifecycle.go               # 修改：添加 startListeners/stopListeners
├── cli/
│   └── analyzer/
│       └── analyzer.go            # 修改：添加 Listener 层识别
└── samples/messageboard/
    └── internal/
        ├── listeners/             # 新增：监听器目录
        │   ├── message_created_listener.go
        │   └── message_audit_listener.go
        └── application/
            └── listener_container.go  # CLI 自动生成
```

## 5. 使用示例

### 5.1 创建监听器

```bash
# 1. 创建 listeners 目录
mkdir -p internal/listeners

# 2. 创建监听器文件
cat > internal/listeners/message_created_listener.go << 'EOF'
package listeners

import (
    "context"
    "github.com/lite-lake/litecore-go/common"
    "github.com/lite-lake/litecore-go/manager/loggermgr"
    "github.com/lite-lake/litecore-go/manager/mqmgr"
)

// IMessageCreatedListener 消息创建监听器
type IMessageCreatedListener interface {
    common.IBaseListener
}

type messageCreatedListenerImpl struct {
    LoggerMgr loggermgr.ILoggerManager `inject:""`
}

func NewMessageCreatedListener() IMessageCreatedListener {
    return &messageCreatedListenerImpl{}
}

func (l *messageCreatedListenerImpl) ListenerName() string {
    return "messageCreatedListenerImpl"
}

func (l *messageCreatedListenerImpl) GetQueue() string {
    return "message.created"
}

func (l *messageCreatedListenerImpl) GetSubscribeOptions() []mqmgr.SubscribeOption {
    return []mqmgr.SubscribeOption{
        mqmgr.WithAutoAck(true),  // 自动确认：返回 nil 自动 Ack，返回 error 自动 Nack
    }
}

func (l *messageCreatedListenerImpl) Handle(ctx context.Context, msg mqmgr.Message) error {
    l.LoggerMgr.Ins().Info("处理消息",
        "id", msg.ID(),
        "body", string(msg.Body()))
    return nil
}

var _ IMessageCreatedListener = (*messageCreatedListenerImpl)(nil)
var _ common.IBaseListener = (*messageCreatedListenerImpl)(nil)
EOF
```

### 5.2 生成容器代码

```bash
# 运行 CLI 工具
go run ./cli/main.go -project . -output internal/application

# 生成结果：
# internal/application/listener_container.go
```

### 5.3 更新 engine.go

```go
package main

import (
    "github.com/lite-lake/litecore-go/server"
    "github.com/lite-lake/litecore-go/samples/messageboard/internal/application"
)

func main() {
    // 初始化各层容器
    entityContainer := application.InitEntityContainer()
    repositoryContainer := application.InitRepositoryContainer(entityContainer)
    serviceContainer := application.InitServiceContainer(repositoryContainer)
    controllerContainer := application.InitControllerContainer(serviceContainer)
    listenerContainer := application.InitListenerContainer(serviceContainer)
    middlewareContainer := application.InitMiddlewareContainer(serviceContainer)

    // 创建引擎（新增 listenerContainer 参数）
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
        listenerContainer,  // 新增
    )

    engine.Run()
}
```

### 5.4 启动应用

```bash
# 编译并运行
go run cmd/server/main.go

# 输出示例：
# 2026-01-24 10:00:00.000 | INFO  | [Listener layer] messageCreatedListenerImpl: injection completed
# 2026-01-24 10:00:00.001 | INFO  | Listener layer startup completed | count=1
# 2026-01-24 10:00:00.002 | INFO  | HTTP server listening | addr=:8080
```

### 5.5 测试监听

```bash
# 发送测试消息
curl -X POST http://localhost:8080/api/messages \
  -H "Content-Type: application/json" \
  -d '{"nickname":"test","content":"hello world"}'

# 观察日志输出（假设 Service 中发布了 MQ 消息）
# 2026-01-24 10:00:05.000 | INFO  | Processing message | id=xxx | body={"id":"123",...}
```

## 6. 影响范围与兼容性

### 6.1 影响范围

**新增文件**：
- `common/base_listener.go`
- `container/listener_container.go`

**修改文件**：
- `container/base_container.go`（添加 Listener 相关类型别名）
- `server/engine.go`（添加 Listener 字段和构造参数）
- `server/lifecycle.go`（添加 startListeners/stopListeners）
- `cli/analyzer/analyzer.go`（添加 Listener 层识别）
- `cli/generator/template.go`（添加 Listener 容器代码生成模板）

**无影响文件**：
- 所有 Manager 实现
- 所有现有 Controller、Service、Repository
- MQManager 接口和实现

### 6.2 兼容性

**向后兼容**：
- 现有项目无需修改即可继续使用
- Listener 层为可选组件，不使用时传入 nil 即可

**API 变更**：
- `server.NewEngine()` 函数签名新增 `listener *container.ListenerContainer` 参数
  ```go
  // 旧版
  func NewEngine(
      builtinConfig *BuiltinConfig,
      entity *container.EntityContainer,
      repository *container.RepositoryContainer,
      service *container.ServiceContainer,
      controller *container.ControllerContainer,
      middleware *container.MiddlewareContainer,
  ) *Engine

  // 新版（兼容，listener 可传 nil）
  func NewEngine(
      builtinConfig *BuiltinConfig,
      entity *container.EntityContainer,
      repository *container.RepositoryContainer,
      service *container.ServiceContainer,
      controller *container.ControllerContainer,
      middleware *container.MiddlewareContainer,
      listener *container.ListenerContainer,  // 新增
  ) *Engine
  ```

**迁移指南**：
1. 旧代码无需修改，listener 参数传 nil 即可
2. 如需使用 Listener，创建 `internal/listeners/` 目录并实现监听器
3. 运行 CLI 工具生成 `listener_container.go`
4. 在 `cmd/server/main.go` 中传入 `listenerContainer`

## 7. 后续优化方向

### 7.1 已实现功能

1. ✅ **重试策略**：支持配置消息重试次数，失败N次后抛弃
2. ✅ **死信队列**：处理多次重试失败的消息（预留接口）
3. ✅ **并发控制**：支持配置单个队列的并发消费者数量
4. ✅ **Panic捕获**：自动捕获并记录panic，防止监听器崩溃

### 7.2 短期优化

1. **重试延迟**：支持配置重试之间的延迟时间（避免立即重试导致雪崩）
2. **监控指标**：添加 Listener 级别的指标（处理速率、错误率、队列积压等）
3. **配置化重试次数**：通过配置文件设置全局或单个队列的最大重试次数
4. **死信队列实现**：完整的死信队列机制（转发到指定队列或持久化到数据库）

### 7.3 中期优化

1. **限流保护**：防止 Listener 处理速度过慢导致队列积压
2. **优先级队列**：支持高优先级消息优先处理
3. **批量处理**：支持批量消费消息
4. **热重载**：运行时动态添加/删除监听器

### 7.4 长期优化

1. **事件溯源**：支持消息持久化和回放
2. **分布式事务**：支持 Saga 模式的长事务
3. **多租户**：支持租户级别的队列隔离
4. **可视化监控**：提供 Listener 运行状态的 Dashboard

## 8. 附录

### 8.1 术语表

| 术语 | 说明 |
|------|------|
| Listener | 监听器，负责消费 MQ 消息的组件 |
| Queue | 队列，消息队列的存储单元 |
| AutoAck | 自动确认，消息处理后自动确认 |
| Durable | 持久化，消息/队列在重启后保留 |
| Nack | 否定确认，拒绝消息并可选择重新入队 |

### 8.2 参考资料

- [MQManager 文档](../manager/mqmgr/README.md)
- [容器架构文档](../container/README.md)
- [依赖注入设计](AGENTS.md)
- [CLI 工具文档](../cli/README.md)

---

**文档结束**
