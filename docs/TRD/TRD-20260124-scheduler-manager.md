# 定时任务管理器技术需求文档 (TRD)

| 文档编号 | TRD-20260124 |
|---------|-------------|
| 文档名称 | 定时任务管理器 (Scheduler Manager) |
| 版本     | 1.0 |
| 日期     | 2026-01-24 |
| 状态     | 草稿 |

## 1. 背景与目标

### 1.1 背景

当前 litecore-go 框架已提供完整的 5 层架构和消息队列监听功能。然而，对于周期性任务（如数据清理、统计报表、缓存预热等），缺乏统一的调度机制。

业务系统中存在大量定时执行场景，需要：
- 定义统一的定时器接口
- 自动注册和发现定时器
- 自动启动和停止定时调度
- 支持标准 Crontab 表达式
- 依赖注入支持
- 与现有架构无缝集成

### 1.2 目标

引入 **Scheduler 层**，作为与 Controller、Listener 并列的独立层，用于处理定时任务调度。实现以下目标：

1. **统一接口定义**：定义 `IBaseScheduler` 接口，规范定时器行为
2. **容器化管理**：实现 `SchedulerContainer`，统一管理所有定时器
3. **Crontab 支持**：支持标准 6 段式 Crontab 表达式（秒 分 时 日 月 周）
4. **时区支持**：每个定时器可独立配置时区，默认使用服务器本地时间
5. **依赖注入**：支持注入 Manager、Service，复用现有 DI 机制
6. **生命周期管理**：在 Engine 启动/停止时自动启动/停止所有定时器
7. **配置验证**：程序加载时检查所有定时器配置，失败直接 panic

### 1.3 非目标

- 不提供复杂的高级调度功能（如依赖关系、任务链等）
- 不改变现有的依赖注入架构
- 不提供任务执行历史记录和监控（可后续扩展）
- 不支持任务的暂停/恢复（仅支持启动/停止）

## 2. 设计方案

### 2.1 架构设计

#### 2.1.1 七层架构

引入 Scheduler 层后，架构从 6 层扩展为 7 层：

```
┌─────────────────────────────────────────────────────────┐
│                        外部触发                           │
│  (HTTP请求、MQ消息、定时触发)                             │
└─────────────────────────────────────────────────────────┘
                             │
          ┌──────────────────┼──────────────────┐
          │                  │                  │
  ┌───────▼────────┐ ┌───────▼────────┐ ┌────────▼────────┐
  │   Controller   │ │    Listener     │ │   Scheduler     │
  │                │ │                 │ │   (新增)        │
  │ HTTP 路由处理  │ │ MQ 消息消费     │ │ 定时任务执行    │
  └────────┬───────┘ └────────┬────────┘ └────────┬────────┘
           │                  │                  │
           └──────────────────┼──────────────────┘
                              │
                    ┌─────────▼────────┐
                    │     Service      │  业务逻辑
                    └─────────┬────────┘
                              │
               ┌──────────────┼──────────────┐
               │              │              │
     ┌─────────▼────────┐ ┌──▼─────────────▼────┐ ┌──▼────────┐
     │   Repository     │ │   Manager           │ │          │
     │                  │ │                     │ │          │
     │ 数据访问         │ │ 配置/缓存/日志等    │ │          │
     └────────┬─────────┘ │   + SchedulerMgr    │ │          │
              │           └─────────────────────┘ │          │
     ┌────────▼────────┘                        │          │
     │     Entity      │                        │          │
     │   数据模型       │                        │          │
     └─────────────────┘                        └──────────┘
```

#### 2.1.2 依赖关系

**允许的依赖方向**：

| 依赖层级 | 可依赖的层级 |
|---------|-------------|
| Scheduler | Manager, Service |
| Listener | Manager, Service |
| Controller | Manager, Service |
| Middleware | Manager, Service |
| Service | Manager, Repository, Entity |
| Repository | Manager, Entity |
| Entity | 无 |

**依赖规则说明**：
- **Scheduler 只能依赖 Manager 和 Service**，不能直接访问 Repository
- **Service 是唯一可以访问 Repository 的层**，封装所有数据访问逻辑
- **Controller、Listener、Scheduler 同级**，分别处理 HTTP 请求、MQ 消息和定时任务，但都不能跨层访问 Repository
- **Controller、Middleware、Listener、Scheduler 都不能注入 Repository**（统一架构规则，由依赖注入检查强制执行）
- **遵循分层架构原则**：上层通过 Service 访问数据，避免绕过业务逻辑层

**Scheduler 的依赖注入规则**：
- ✅ 可注入：所有 Manager（LoggerManager, DatabaseManager 等）
- ✅ 可注入：所有 Service
- ✅ 可注入：同层其他 Scheduler（需注意循环依赖）
- ❌ 不可注入：Repository（违反分层架构，必须通过 Service 访问数据）
- ❌ 不可注入：Controller、Middleware、Listener（避免与其他处理层耦合）

### 2.2 接口定义

#### 2.2.1 基础接口

**文件位置**：`common/base_scheduler.go`

```go
package common

import "time"

// IBaseScheduler 基础定时器接口
// 所有 Scheduler 类必须继承此接口并实现相关方法
// 用于定义定时器的基础行为和契约
type IBaseScheduler interface {
    // SchedulerName 返回定时器名称
    // 格式：xxxScheduler（小驼峰）
    // 示例："cleanupScheduler"
    SchedulerName() string

    // GetRule 返回 Crontab 定时规则
    // 使用标准 6 段式格式：秒 分 时 日 月 周
    // 示例："0 */5 * * * *" 表示每 5 分钟执行一次
    //      "0 0 2 * * *" 表示每天凌晨 2 点执行
    //      "0 0 * * * 1" 表示每周一凌晨执行
    GetRule() string

    // GetTimezone 返回定时器使用的时区
    // 返回空字符串时使用服务器本地时间
    // 支持标准时区名称，如 "Asia/Shanghai", "UTC", "America/New_York"
    // 默认值：空字符串（服务器本地时间）
    GetTimezone() string

    // OnTick 定时触发时调用
    // tickID: 计划执行时间的 Unix 时间戳（秒级），可用于去重或日志追踪
    // 返回: 执行错误（返回 error 不会触发重试，仅记录日志）
    OnTick(tickID int64) error

    // OnStart 在服务器启动时触发
    // 用于初始化定时器状态、连接资源等
    OnStart() error

    // OnStop 在服务器停止时触发
    // 用于清理资源、保存状态等
    OnStop() error
}
```

#### 2.2.2 定时器示例

**文件位置**：`samples/messageboard/internal/schedulers/cleanup_scheduler.go`

```go
package schedulers

import (
    "github.com/lite-lake/litecore-go/common"
    "github.com/lite-lake/litecore-go/manager/loggermgr"
    "github.com/lite-lake/litecore-go/samples/messageboard/internal/services"
)

// ICleanupScheduler 清理定时器接口
type ICleanupScheduler interface {
    common.IBaseScheduler
}

type cleanupSchedulerImpl struct {
    MessageService services.IMessageService `inject:""`
    LoggerMgr      loggermgr.ILoggerManager `inject:""`
    logger         loggermgr.ILogger
}

// NewCleanupScheduler 创建定时器实例
func NewCleanupScheduler() ICleanupScheduler {
    return &cleanupSchedulerImpl{}
}

// SchedulerName 返回定时器名称
func (s *cleanupSchedulerImpl) SchedulerName() string {
    return "cleanupScheduler"
}

// GetRule 返回 Crontab 规则
// 每天凌晨 2 点执行
func (s *cleanupSchedulerImpl) GetRule() string {
    return "0 0 2 * * *"
}

// GetTimezone 返回时区
// 使用上海时区
func (s *cleanupSchedulerImpl) GetTimezone() string {
    return "Asia/Shanghai"
}

// OnTick 定时触发
func (s *cleanupSchedulerImpl) OnTick(tickID int64) error {
    s.initLogger()
    s.logger.Info("开始清理任务", "tick_id", tickID)

    // 业务处理逻辑...
    // 使用注入的 Service 处理业务
    count, err := s.MessageService.CleanupExpiredMessages()
    if err != nil {
        s.logger.Error("清理任务失败", "error", err)
        return err
    }

    s.logger.Info("清理任务完成", "deleted_count", count)
    return nil
}

// OnStart 启动时初始化
func (s *cleanupSchedulerImpl) OnStart() error {
    s.initLogger()
    s.logger.Info("清理定时器已启动")
    return nil
}

// OnStop 停止时清理
func (s *cleanupSchedulerImpl) OnStop() error {
    s.initLogger()
    s.logger.Info("清理定时器已停止")
    return nil
}

// initLogger 初始化日志器
func (s *cleanupSchedulerImpl) initLogger() {
    if s.logger == nil && s.LoggerMgr != nil {
        s.logger = s.LoggerMgr.Logger(s.SchedulerName())
    }
}

var _ ICleanupScheduler = (*cleanupSchedulerImpl)(nil)
var _ common.IBaseScheduler = (*cleanupSchedulerImpl)(nil)
```

### 2.3 Crontab 表达式解析

#### 2.3.1 表达式格式

使用标准 6 段式 Crontab 表达式：

```
┌─────────────── 秒 (0-59)
│ ┌───────────── 分 (0-59)
│ │ ┌─────────── 时 (0-23)
│ │ │ ┌───────── 日 (1-31)
│ │ │ │ ┌─────── 月 (1-12)
│ │ │ │ │ ┌───── 周 (0-6, 0=周日)
│ │ │ │ │ │
* * * * * *
```

**支持的特殊字符**：

| 字符 | 含义 | 示例 |
|-----|------|------|
| `*` | 任意值 | `* * * * * *` 每秒执行 |
| `,` | 多个值 | `0,30 * * * * *` 每分钟的第 0 秒和第 30 秒 |
| `-` | 范围 | `0-29 * * * * *` 每分钟的第 0 到 29 秒 |
| `/` | 步长 | `*/10 * * * * *` 每 10 秒执行一次 |
| `?` | 不指定（仅用于日和周） | `0 0 0 * * ?` 不指定星期几 |

#### 2.3.2 常用表达式示例

| 表达式 | 说明 |
|--------|------|
| `0 * * * * *` | 每分钟的第 0 秒执行 |
| `*/5 * * * * *` | 每 5 秒执行一次 |
| `0 */5 * * * *` | 每 5 分钟执行一次 |
| `0 0 * * * *` | 每小时的第 0 分第 0 秒执行 |
| `0 0 0 * * *` | 每天凌晨执行 |
| `0 0 0 * * 0` | 每周日凌晨执行 |
| `0 0 0 1 * *` | 每月 1 号凌晨执行 |
| `0 0 2 * * *` | 每天凌晨 2 点执行 |
| `0 0 12 * * 1-5` | 周一到周五中午 12 点执行 |
| `0 30 8,18 * * *` | 每天 8:30 和 18:30 执行 |

### 2.4 Manager 接口定义

**文件位置**：`manager/schedulermgr/interface.go`

```go
package schedulermgr

import (
    "github.com/lite-lake/litecore-go/common"
)

// ISchedulerManager 定时任务管理器接口
type ISchedulerManager interface {
    common.IBaseManager

    // ValidateScheduler 验证定时器配置是否正确
    // 在程序加载时调用，配置错误直接 panic
    // scheduler: 待验证的定时器实例
    // 返回: 验证错误（调用方负责 panic）
    ValidateScheduler(scheduler common.IBaseScheduler) error

    // RegisterScheduler 注册定时器
    // 在 SchedulerManager.OnStart() 时由容器调用
    // scheduler: 待注册的定时器实例
    // 返回: 注册错误
    RegisterScheduler(scheduler common.IBaseScheduler) error

    // UnregisterScheduler 注销定时器
    // 在 SchedulerManager.OnStop() 时由容器调用
    // scheduler: 待注销的定时器实例
    // 返回: 注销错误
    UnregisterScheduler(scheduler common.IBaseScheduler) error
}
```

### 2.5 容器实现

#### 2.5.1 SchedulerContainer

**文件位置**：`container/scheduler_container.go`

```go
package container

import (
    "reflect"

    "github.com/lite-lake/litecore-go/common"
)

// SchedulerContainer 定时器层容器
type SchedulerContainer struct {
    *InjectableLayerContainer[common.IBaseScheduler]
    serviceContainer *ServiceContainer
}

// NewSchedulerContainer 创建新的定时器容器
func NewSchedulerContainer(service *ServiceContainer) *SchedulerContainer {
    return &SchedulerContainer{
        InjectableLayerContainer: NewInjectableLayerContainer(func(s common.IBaseScheduler) string {
            return s.SchedulerName()
        }),
        serviceContainer: service,
    }
}

// SetManagerContainer 设置管理器容器
func (c *SchedulerContainer) SetManagerContainer(container *ManagerContainer) {
    c.InjectableLayerContainer.SetManagerContainer(container)
}

// RegisterScheduler 泛型注册函数，按接口类型注册
func RegisterScheduler[T common.IBaseScheduler](c *SchedulerContainer, impl T) error {
    ifaceType := reflect.TypeOf((*T)(nil)).Elem()
    return c.RegisterByType(ifaceType, impl)
}

// GetScheduler 按接口类型获取
func GetScheduler[T common.IBaseScheduler](c *SchedulerContainer) (T, error) {
    ifaceType := reflect.TypeOf((*T)(nil)).Elem()
    impl := c.GetByType(ifaceType)
    if impl == nil {
        var zero T
        return zero, &InstanceNotFoundError{
            Name:  ifaceType.Name(),
            Layer: "Scheduler",
        }
    }
    return impl.(T), nil
}

// InjectAll 执行依赖注入
func (c *SchedulerContainer) InjectAll() error {
    c.checkManagerContainer("Scheduler")

    if c.InjectableLayerContainer.base.container.IsInjected() {
        return nil
    }

    c.InjectableLayerContainer.base.sources = c.InjectableLayerContainer.base.buildSources(c, c.managerContainer, c.serviceContainer)
    return c.InjectableLayerContainer.base.injectAll(c)
}

// GetDependency 根据类型获取依赖实例（实现ContainerSource接口）
func (c *SchedulerContainer) GetDependency(fieldType reflect.Type) (interface{}, error) {
    baseRepositoryType := reflect.TypeOf((*common.IBaseRepository)(nil)).Elem()
    if fieldType == baseRepositoryType || fieldType.Implements(baseRepositoryType) {
        return nil, &DependencyNotFoundError{
            FieldType:     fieldType,
            ContainerType: "Repository",
            Message:       "Scheduler cannot directly inject Repository, must access data through Service",
        }
    }

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

    return nil, nil
}

// ValidateAll 验证所有定时器配置
// 在程序加载时调用，配置错误直接 panic
// 注意：此方法只做基础验证，Crontab 表达式解析由 SchedulerManager 完成
func (c *SchedulerContainer) ValidateAll() {
    schedulers := c.GetAll()
    if len(schedulers) == 0 {
        return
    }

    for _, scheduler := range schedulers {
        if err := c.validateScheduler(scheduler); err != nil {
            panic(fmt.Sprintf("scheduler %s validation failed: %v", scheduler.SchedulerName(), err))
        }
    }
}

// validateScheduler 验证单个定时器（基础验证）
// Crontab 表达式的完整解析由 SchedulerManager.ValidateScheduler() 完成
func (c *SchedulerContainer) validateScheduler(scheduler common.IBaseScheduler) error {
    // 验证 Crontab 表达式非空
    rule := scheduler.GetRule()
    if rule == "" {
        return fmt.Errorf("rule cannot be empty")
    }

    // 验证时区格式（如果不为空）
    timezone := scheduler.GetTimezone()
    if timezone != "" {
        _, err := time.LoadLocation(timezone)
        if err != nil {
            return fmt.Errorf("invalid timezone: %w", err)
        }
    }

    // 注意：不在这里解析 Crontab 表达式，避免容器层依赖管理器实现
    // Crontab 表达式的解析和验证由 SchedulerManager.ValidateScheduler() 完成
    return nil
}
```

#### 2.5.2 自动生成代码

**文件位置**：`samples/messageboard/internal/application/scheduler_container.go`（CLI 生成）

```go
// Code generated by litecore/cli. DO NOT EDIT.
package application

import (
    "github.com/lite-lake/litecore-go/container"
    schedulers "github.com/lite-lake/litecore-go/samples/messageboard/internal/schedulers"
)

// InitSchedulerContainer 初始化定时器容器
func InitSchedulerContainer(serviceContainer *container.ServiceContainer) *container.SchedulerContainer {
    schedulerContainer := container.NewSchedulerContainer(serviceContainer)

    // 使用泛型注册函数注册定时器
    container.RegisterScheduler[schedulers.ICleanupScheduler](schedulerContainer, schedulers.NewCleanupScheduler())
    container.RegisterScheduler[schedulers.IStatisticsScheduler](schedulerContainer, schedulers.NewStatisticsScheduler())

    return schedulerContainer
}
```

### 2.6 Engine 扩展

#### 2.6.1 Engine 结构体修改

**文件位置**：`server/engine.go`

```go
// Engine 服务引擎
type Engine struct {
    // ... 现有字段

    // Scheduler 定时器容器（新增）
    Scheduler *container.SchedulerContainer
}
```

#### 2.6.2 NewEngine 构造函数修改

```go
func NewEngine(
    builtinConfig *BuiltinConfig,
    entity *container.EntityContainer,
    repository *container.RepositoryContainer,
    service *container.ServiceContainer,
    controller *container.ControllerContainer,
    middleware *container.MiddlewareContainer,
    listener *container.ListenerContainer,
    scheduler *container.SchedulerContainer,  // 新增参数
) *Engine {
    // ... 现有逻辑

    return &Engine{
        // ... 现有字段
        Scheduler: scheduler,
        // ...
    }
}
```

#### 2.6.3 Initialize 方法修改

```go
import (
    "fmt"
    "time"

    "github.com/lite-lake/litecore-go/container"
    "github.com/lite-lake/litecore-go/logger"
    "github.com/lite-lake/litecore-go/manager/schedulermgr"
)

func (e *Engine) Initialize() error {
    e.mu.Lock()
    defer e.mu.Unlock()

    // 初始化启动时间统计
    e.startupStartTime = time.Now()

    // 初始化前使用默认日志器
    e.setLogger(logger.NewDefaultLogger("Engine"))
    e.isStartup = true

    // 1. 初始化内置组件
    builtInManagerContainer, err := Initialize(e.builtinConfig)
    if err != nil {
        return fmt.Errorf("failed to initialize builtin components: %w", err)
    }
    e.Manager = builtInManagerContainer

    // 切换到结构化日志
    if loggerMgr, err := container.GetManager[loggermgr.ILoggerManager](e.Manager); err == nil {
        e.setLogger(loggerMgr.Ins())
        e.isStartup = false
        e.getLogger().Info("切换到结构化日志系统")

        // 初始化异步日志器
        if e.startupLogConfig.Async {
            e.asyncLogger = NewAsyncStartupLogger(e.getLogger(), e.startupLogConfig.Buffer)
        }
    }

    // 2. 验证 Scheduler 配置（在依赖注入之前）
    if e.Scheduler != nil {
        e.logPhaseStart(PhaseValidation, "开始验证 Scheduler 配置")

        // 2.1 Container 层基础验证（规则非空、时区格式）
        e.Scheduler.ValidateAll()

        // 2.2 Manager 层完整验证（Crontab 表达式解析）
        schedulerMgr, err := container.GetManager[schedulermgr.ISchedulerManager](e.Manager)
        if err == nil {
            schedulers := e.Scheduler.GetAll()
            for _, scheduler := range schedulers {
                if err := schedulerMgr.ValidateScheduler(scheduler); err != nil {
                    panic(fmt.Sprintf("scheduler %s crontab validation failed: %v", scheduler.SchedulerName(), err))
                }
            }
        }

        e.logPhaseEnd(PhaseValidation, "Scheduler 配置验证完成", logger.F("count", e.Scheduler.Count()))
    }

    // 3. 自动依赖注入
    if err := e.autoInject(); err != nil {
        return fmt.Errorf("auto inject failed: %w", err)
    }

    // ... 后续逻辑
}
```

#### 2.6.4 autoInject 方法扩展

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

    // 5. Scheduler 层（新增）
    e.Scheduler.SetManagerContainer(e.Manager)
    if err := e.Scheduler.InjectAll(); err != nil {
        return fmt.Errorf("scheduler inject failed: %w", err)
    }
    schedulers := e.Scheduler.GetAll()
    for _, scheduler := range schedulers {
        e.logStartup(PhaseInjection, fmt.Sprintf("[%s 层] %s: 注入完成", "Scheduler", scheduler.SchedulerName()))
    }

    // 6. Middleware 层
    // ... 现有代码

    // 7. Listener 层
    // ... 现有代码

    totalCount := len(repos) + len(svcs) + len(ctrls) + len(listeners) + len(mws) + len(schedulers)
    e.logPhaseEnd(PhaseInjection, "依赖注入完成", logger.F("count", totalCount))

    return nil
}
```

#### 2.6.5 启动定时器

**文件位置**：`server/lifecycle.go`

```go
// startSchedulers 启动所有定时器
func (e *Engine) startSchedulers() error {
    e.logPhaseStart(PhaseStartup, "开始启动 Scheduler 层")

    if e.Scheduler == nil {
        e.getLogger().Info("未配置 Scheduler 层，跳过启动")
        return nil
    }

    schedulers := e.Scheduler.GetAll()
    if len(schedulers) == 0 {
        e.getLogger().Info("没有注册的 Scheduler，跳过启动")
        return nil
    }

    schedulerMgr, err := container.GetManager[schedulermgr.ISchedulerManager](e.Manager)
    if err != nil {
        return fmt.Errorf("SchedulerManager 未初始化，但存在 %d 个 Scheduler: %w", len(schedulers), err)
    }

    startedCount := 0

    for _, scheduler := range schedulers {
        e.getLogger().Info("注册定时器",
            logger.F("scheduler", scheduler.SchedulerName()),
            logger.F("rule", scheduler.GetRule()),
            logger.F("timezone", scheduler.GetTimezone()))

        // 注册定时器
        if err := schedulerMgr.RegisterScheduler(scheduler); err != nil {
            return fmt.Errorf("注册定时器 %s 失败: %w", scheduler.SchedulerName(), err)
        }

        // 调用定时器的 OnStart
        if err := scheduler.OnStart(); err != nil {
            return fmt.Errorf("启动定时器 %s 失败: %w", scheduler.SchedulerName(), err)
        }

        e.logStartup(PhaseStartup, scheduler.SchedulerName()+": 启动完成")
        startedCount++
    }

    e.logPhaseEnd(PhaseStartup, "Scheduler 层启动完成", logger.F("count", startedCount))
    return nil
}
```

#### 2.6.6 停止定时器

```go
// stopSchedulers 停止所有定时器
func (e *Engine) stopSchedulers() []error {
    if e.Scheduler == nil {
        return nil
    }

    schedulers := e.Scheduler.GetAll()
    var errors []error

    // 按相反顺序停止
    for i := len(schedulers) - 1; i >= 0; i-- {
        scheduler := schedulers[i]

        // 调用定时器的 OnStop
        if err := scheduler.OnStop(); err != nil {
            errors = append(errors, fmt.Errorf("停止定时器 %s 失败: %w", scheduler.SchedulerName(), err))
        }

        // 注销定时器
        schedulerMgr, err := container.GetManager[schedulermgr.ISchedulerManager](e.Manager)
        if err != nil {
            errors = append(errors, fmt.Errorf("获取 SchedulerManager 失败: %w", err))
            continue
        }

        if err := schedulerMgr.UnregisterScheduler(scheduler); err != nil {
            errors = append(errors, fmt.Errorf("注销定时器 %s 失败: %w", scheduler.SchedulerName(), err))
        }
    }

    return errors
}
```

#### 2.6.7 Start 方法修改

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

    // 4. 启动所有 Middleware

    // 5. 启动所有 Scheduler（新增）
    if err := e.startSchedulers(); err != nil {
        return fmt.Errorf("start schedulers failed: %w", err)
    }

    // 6. 启动所有 Listener

    // 停止异步日志器
    if e.asyncLogger != nil {
        e.asyncLogger.Stop()
        e.asyncLogger = nil
    }

    // 7. 启动 HTTP 服务器
}
```

#### 2.6.8 Stop 方法修改

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

    schedulerErrors := e.stopSchedulers()  // 新增
    e.logStartup(PhaseShutdown, "Scheduler 层停止完成")

    serviceErrors := e.stopServices()
    e.logStartup(PhaseShutdown, "Service 层停止完成")

    repositoryErrors := e.stopRepositories()
    e.logStartup(PhaseShutdown, "Repository 层停止完成")

    managerErrors := e.stopManagers()
    e.logStartup(PhaseShutdown, "Manager 层停止完成")

    allErrors := make([]error, 0)
    allErrors = append(allErrors, middlewareErrors...)
    allErrors = append(allErrors, listenerErrors...)
    allErrors = append(allErrors, schedulerErrors...)
    allErrors = append(allErrors, serviceErrors...)
    allErrors = append(allErrors, repositoryErrors...)
    allErrors = append(allErrors, managerErrors...)

    // ... 错误处理逻辑
}
```

### 2.7 CLI 工具扩展

#### 2.7.1 analyzer 扩展

**文件位置**：`cli/analyzer/analyzer.go`

```go
const (
    LayerEntity     Layer = "entity"
    LayerRepository Layer = "repository"
    LayerService    Layer = "service"
    LayerController Layer = "controller"
    LayerMiddleware Layer = "middleware"
    LayerListener   Layer = "listener"
    LayerScheduler  Layer = "scheduler"  // 新增
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
        if strings.Contains(part, "listeners") {
            return LayerListener
        }
        if strings.Contains(part, "schedulers") {  // 新增
            return LayerScheduler
        }
    }

    return ""
}

// IsLitecoreLayer 判断是否为 Litecore 标准层
func IsLitecoreLayer(layer Layer) bool {
    switch layer {
    case LayerEntity, LayerRepository, LayerService,
        LayerController, LayerMiddleware, LayerListener, LayerScheduler:  // 新增
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
    case LayerListener:
        return "IBaseListener"
    case LayerScheduler:  // 新增
        return "IBaseScheduler"
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
    case LayerListener:
        return "ListenerContainer"
    case LayerScheduler:  // 新增
        return "SchedulerContainer"
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
    case LayerListener:
        return "RegisterListener"
    case LayerScheduler:  // 新增
        return "RegisterScheduler"
    default:
        return "Register"
    }
}
```

## 3. 实现细节

### 3.1 定时任务执行流程

```
┌────────────────────────────────────────────────────────────┐
│  1. 应用启动                                                │
│  └─> Engine.Start()                                        │
│      └─> startSchedulers()                                  │
│          └─> foreach scheduler:                            │
│              SchedulerManager.RegisterScheduler(scheduler)  │
│              scheduler.OnStart()                            │
│              └─> SchedulerManager 启动定时器                │
└────────────────────────────────────────────────────────────┘
                            │
                            ▼
┌────────────────────────────────────────────────────────────┐
│  2. 定时触发                                                │
│  └─> SchedulerManager 监控时间                              │
│      └─> 到达触发时间                                      │
│          └─> 启动独立协程                                  │
│              └─> scheduler.OnTick(tickID)                 │
│                  tickID = 计划执行时间的 Unix 时间戳        │
└────────────────────────────────────────────────────────────┘
                            │
                            ▼
┌────────────────────────────────────────────────────────────┐
│  3. 业务处理                                                │
│  └─> scheduler.OnTick() 方法执行                            │
│      └─> 使用注入的 Service 处理业务逻辑                    │
│          └─> Service 内部通过 Repository 访问数据        │
└────────────────────────────────────────────────────────────┘
                            │
                            ▼
┌────────────────────────────────────────────────────────────┐
│  4. 返回结果                                                │
│  └─> return nil   ──> 任务执行成功（记录 info 日志）       │
│  └─> return error ──> 任务执行失败（记录 error 日志）     │
│      └─> 不自动重试，仅记录日志                            │
└────────────────────────────────────────────────────────────┘
```

### 3.2 并发控制

**完全并发执行**：
- 每次触发定时任务时，启动独立的 goroutine 执行 `OnTick()`
- 不等待前一次执行完成，允许并发执行
- 业务系统自行处理并发安全问题

**示例场景**：
```go
// 定时器配置：每 5 秒执行一次
GetRule() -> "*/5 * * * * *"

// 执行时间线：
t=00s  -> OnTick(1706091600) 启动 goroutine A
t=05s  -> OnTick(1706091605) 启动 goroutine B（即使 A 还在执行）
t=10s  -> OnTick(1706091610) 启动 goroutine C（即使 A、B 还在执行）
...
```

**并发安全建议**：
- 使用互斥锁保护共享状态
- 使用分布式锁处理跨实例并发
- 任务幂等设计（使用 tickID 去重）

### 3.3 配置管理

#### 3.3.1 全局配置

```yaml
# configs/config.yaml
scheduler:
  driver: "cron"  # 目前只有 cron 实现
  cron_config:
    # 启动时是否检查所有 Scheduler 配置（默认 true）
    validate_on_startup: true
```

#### 3.3.2 Scheduler 配置

Scheduler 的配置通过接口方法返回，不需要额外配置文件：

```go
type cleanupSchedulerImpl struct {
    // ...
}

func (s *cleanupSchedulerImpl) GetRule() string {
    return "0 0 2 * * *"  // Crontab 规则
}

func (s *cleanupSchedulerImpl) GetTimezone() string {
    return "Asia/Shanghai"  // 时区
}
```

### 3.4 时区处理

#### 3.4.1 时区默认值

- `GetTimezone()` 返回空字符串 → 使用服务器本地时间
- `GetTimezone()` 返回具体时区 → 使用该时区

#### 3.4.2 时区示例

```go
// 1. 使用服务器本地时间
func (s *schedulerImpl) GetTimezone() string {
    return ""  // 或 "Local"
}

// 2. 使用 UTC 时间
func (s *schedulerImpl) GetTimezone() string {
    return "UTC"
}

// 3. 使用上海时区
func (s *schedulerImpl) GetTimezone() string {
    return "Asia/Shanghai"
}
```

#### 3.4.3 时区转换示例

```go
// 定时器规则：每天凌晨 2 点执行
rule := "0 0 2 * * *"

// 上海时区：Asia/Shanghai
// UTC 时区：02:00 UTC = 10:00 Shanghai（相差 8 小时）
// 本地时间：根据服务器设置
```

### 3.5 日志记录

Scheduler 应使用注入的 `LoggerManager` 记录日志，遵循框架的日志使用规范：

```go
type mySchedulerImpl struct {
    LoggerMgr loggermgr.ILoggerManager `inject:""`
    logger     loggermgr.ILogger
}

func (s *mySchedulerImpl) initLogger() {
    if s.logger == nil && s.LoggerMgr != nil {
        s.logger = s.LoggerMgr.Logger(s.SchedulerName())
    }
}

func (s *mySchedulerImpl) OnTick(tickID int64) error {
    s.initLogger()

    s.logger.Info("开始执行定时任务",
        "scheduler", s.SchedulerName(),
        "tick_id", tickID)

    // 业务逻辑...

    s.logger.Info("定时任务执行完成", "tick_id", tickID)
    return nil
}
```

**日志级别使用建议**：
- **Debug**：详细的任务处理步骤
- **Info**：任务启动、完成、关键节点
- **Warn**：重试、降级处理、配置异常
- **Error**：任务执行失败、依赖服务异常
- **Fatal**：严重错误导致定时器无法继续运行

### 3.6 错误处理

#### 3.6.1 OnTick 返回值处理

| 场景                | 处理方式                    | 说明                          |
|---------------------|----------------------------|-------------------------------|
| OnTick返回nil       | 记录 info 日志             | 任务执行成功                  |
| OnTick返回error     | 记录 error 日志             | 任务执行失败，不自动重试      |
| OnTick发生panic     | 捕获 panic，记录 error 日志 | 防止定时器崩溃，继续调度      |

#### 3.6.2 配置验证失败

- 验证失败 → 直接 panic
- 程序终止，启动失败
- 需修复配置后重新启动

```go
// 在 Engine.Initialize() 中
if e.Scheduler != nil {
    e.Scheduler.ValidateAll()  // 失败直接 panic
}
```

### 3.7 测试策略

#### 3.7.1 单元测试

```go
func TestCleanupScheduler_OnTick(t *testing.T) {
    // 创建 Mock 依赖
    mockService := &mockMessageService{}
    mockLogger := &mockLoggerManager{}

    // 创建定时器
    scheduler := &cleanupSchedulerImpl{
        MessageService: mockService,
        LoggerMgr:      mockLogger,
    }

    // 调用 OnTick
    err := scheduler.OnTick(1706091600)

    // 断言
    assert.NoError(t, err)
    mockService.AssertCalled(t, "CleanupExpiredMessages")
}
```

#### 3.7.2 集成测试

```go
func TestSchedulerIntegration(t *testing.T) {
    // 创建 SchedulerManager
    schedulerMgr := schedulermgr.NewSchedulerManagerCronImpl(&schedulermgr.CronConfig{
        ValidateOnStartup: true,
    })

    // 创建定时器容器
    schedulerContainer := container.NewSchedulerContainer(nil)
    container.RegisterScheduler[ICleanupScheduler](
        schedulerContainer,
        NewCleanupScheduler(),
    )

    // 验证配置
    schedulerContainer.ValidateAll()

    // 注册定时器
    schedulers := schedulerContainer.GetAll()
    for _, scheduler := range schedulers {
        err := schedulerMgr.RegisterScheduler(scheduler)
        assert.NoError(t, err)
        err = scheduler.OnStart()
        assert.NoError(t, err)
    }

    // 等待定时触发（使用快速规则进行测试）
    // ...

    // 清理
    for _, scheduler := range schedulers {
        scheduler.OnStop()
        schedulerMgr.UnregisterScheduler(scheduler)
    }
}
```

## 4. 目录结构

```
litecore-go/
├── common/
│   └── base_scheduler.go          # 新增：IBaseScheduler 接口定义
├── container/
│   ├── base_container.go          # 修改：添加 Scheduler 相关类型
│   └── scheduler_container.go     # 新增：SchedulerContainer 实现
├── manager/
│   └── schedulermgr/
│       ├── doc.go                 # 包文档
│       ├── interface.go           # ISchedulerManager 接口
│       ├── impl_base.go           # 基础实现（日志、遥测）
│       ├── cron_impl.go           # Crontab 定时器实现
│       ├── crontab_parser.go      # Crontab 表达式解析器
│       ├── config.go              # 配置定义
│       └── factory.go             # 工厂方法
├── server/
│   ├── engine.go                  # 修改：添加 Scheduler 字段和方法
│   ├── lifecycle.go               # 修改：添加 startSchedulers/stopSchedulers
│   └── builtin.go                 # 修改：初始化 SchedulerManager
├── cli/
│   └── analyzer/
│       └── analyzer.go            # 修改：添加 Scheduler 层识别
└── samples/messageboard/
    └── internal/
        ├── schedulers/            # 新增：定时器目录
        │   ├── cleanup_scheduler.go
        │   └── statistics_scheduler.go
        └── application/
            └── scheduler_container.go  # CLI 自动生成
```

## 5. 使用示例

### 5.1 创建定时器

```bash
# 1. 创建 schedulers 目录
mkdir -p internal/schedulers

# 2. 创建定时器文件
cat > internal/schedulers/cleanup_scheduler.go << 'EOF'
package schedulers

import (
    "github.com/lite-lake/litecore-go/common"
    "github.com/lite-lake/litecore-go/manager/loggermgr"
    "github.com/lite-lake/litecore-go/samples/messageboard/internal/services"
)

// ICleanupScheduler 清理定时器
type ICleanupScheduler interface {
    common.IBaseScheduler
}

type cleanupSchedulerImpl struct {
    MessageService services.IMessageService `inject:""`
    LoggerMgr      loggermgr.ILoggerManager `inject:""`
    logger         loggermgr.ILogger
}

func NewCleanupScheduler() ICleanupScheduler {
    return &cleanupSchedulerImpl{}
}

func (s *cleanupSchedulerImpl) SchedulerName() string {
    return "cleanupScheduler"
}

func (s *cleanupSchedulerImpl) GetRule() string {
    return "0 0 2 * * *"  // 每天凌晨 2 点
}

func (s *cleanupSchedulerImpl) GetTimezone() string {
    return "Asia/Shanghai"
}

func (s *cleanupSchedulerImpl) OnTick(tickID int64) error {
    s.initLogger()
    s.logger.Info("开始清理任务", "tick_id", tickID)

    count, err := s.MessageService.CleanupExpiredMessages()
    if err != nil {
        s.logger.Error("清理任务失败", "error", err)
        return err
    }

    s.logger.Info("清理任务完成", "deleted_count", count)
    return nil
}

func (s *cleanupSchedulerImpl) OnStart() error {
    s.initLogger()
    s.logger.Info("清理定时器已启动")
    return nil
}

func (s *cleanupSchedulerImpl) OnStop() error {
    s.initLogger()
    s.logger.Info("清理定时器已停止")
    return nil
}

func (s *cleanupSchedulerImpl) initLogger() {
    if s.logger == nil && s.LoggerMgr != nil {
        s.logger = s.LoggerMgr.Logger(s.SchedulerName())
    }
}

var _ ICleanupScheduler = (*cleanupSchedulerImpl)(nil)
var _ common.IBaseScheduler = (*cleanupSchedulerImpl)(nil)
EOF
```

### 5.2 生成容器代码

```bash
# 运行 CLI 工具
go run ./cli/main.go -project . -output internal/application

# 生成结果：
# internal/application/scheduler_container.go
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
    middlewareContainer := application.InitMiddlewareContainer(serviceContainer)
    listenerContainer := application.InitListenerContainer(serviceContainer)
    schedulerContainer := application.InitSchedulerContainer(serviceContainer)  // 新增

    // 创建引擎（新增 schedulerContainer 参数）
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
        listenerContainer,
        schedulerContainer,  // 新增
    )

    engine.Run()
}
```

### 5.4 配置文件

```yaml
# configs/config.yaml
server:
  mode: "debug"
  address: ":8080"

scheduler:
  driver: "cron"
  cron_config:
    validate_on_startup: true
```

### 5.5 启动应用

```bash
# 编译并运行
go run cmd/server/main.go

# 输出示例：
# 2026-01-24 10:00:00.000 | INFO  | 开始验证 Scheduler 配置
# 2026-01-24 10:00:00.001 | INFO  | Scheduler 配置验证完成 | count=1
# 2026-01-24 10:00:00.002 | INFO  | [Scheduler 层] cleanupScheduler: 注入完成
# 2026-01-24 10:00:00.003 | INFO  | 开始启动 Scheduler 层
# 2026-01-24 10:00:00.004 | INFO  | 注册定时器 | scheduler=cleanupScheduler | rule="0 0 2 * * *" | timezone=Asia/Shanghai
# 2026-01-24 10:00:00.005 | INFO  | cleanupScheduler: 启动完成
# 2026-01-24 10:00:00.006 | INFO  | Scheduler 层启动完成 | count=1
# 2026-01-24 10:00:00.007 | INFO  | HTTP server listening | addr=:8080
```

### 5.6 观察定时任务

```bash
# 到达凌晨 2 点后，观察日志：
# 2026-01-25 02:00:00.000 | INFO  | 开始清理任务 | tick_id=1737744000
# 2026-01-25 02:00:00.500 | INFO  | 清理任务完成 | deleted_count=123
```

## 6. 影响范围与兼容性

### 6.1 影响范围

**新增文件**：
- `common/base_scheduler.go`
- `container/scheduler_container.go`
- `manager/schedulermgr/`（完整目录）

**修改文件**：
- `container/base_container.go`（添加 Scheduler 相关类型）
- `server/engine.go`（添加 Scheduler 字段和构造参数）
- `server/lifecycle.go`（添加 startSchedulers/stopSchedulers）
- `server/builtin.go`（初始化 SchedulerManager）
- `cli/analyzer/analyzer.go`（添加 Scheduler 层识别）

**无影响文件**：
- 所有现有 Manager 实现
- 所有现有 Controller、Service、Repository
- Listener 和 Middleware 层

### 6.2 兼容性

**向后兼容**：
- 现有项目无需修改即可继续使用
- Scheduler 层为可选组件，不使用时传入 nil 即可

**API 变更**：
- `server.NewEngine()` 函数签名新增 `scheduler *container.SchedulerContainer` 参数
  ```go
  // 旧版
  func NewEngine(
      builtinConfig *BuiltinConfig,
      entity *container.EntityContainer,
      repository *container.RepositoryContainer,
      service *container.ServiceContainer,
      controller *container.ControllerContainer,
      middleware *container.MiddlewareContainer,
      listener *container.ListenerContainer,
  ) *Engine

  // 新版（兼容，scheduler 可传 nil）
  func NewEngine(
      builtinConfig *BuiltinConfig,
      entity *container.EntityContainer,
      repository *container.RepositoryContainer,
      service *container.ServiceContainer,
      controller *container.ControllerContainer,
      middleware *container.MiddlewareContainer,
      listener *container.ListenerContainer,
      scheduler *container.SchedulerContainer,  // 新增
  ) *Engine
  ```

**迁移指南**：
1. 旧代码无需修改，scheduler 参数传 nil 即可
2. 如需使用 Scheduler，创建 `internal/schedulers/` 目录并实现定时器
3. 运行 CLI 工具生成 `scheduler_container.go`
4. 在 `cmd/server/main.go` 中传入 `schedulerContainer`

## 7. 后续优化方向

### 7.1 已实现功能

1. ✅ **标准 Crontab 表达式**：支持 6 段式 Crontab 格式
2. ✅ **时区支持**：每个定时器可独立配置时区
3. ✅ **完全并发**：每次触发启动独立协程
4. ✅ **配置验证**：程序加载时检查配置，失败直接 panic
5. ✅ **依赖注入**：支持注入 Manager、Service

### 7.2 短期优化

1. **执行历史记录**：记录定时任务的执行历史（成功/失败/耗时）
2. **手动触发**：支持通过 API 手动触发某个定时器执行
3. **超时控制**：配置任务最大执行时长，超时自动取消
4. **健康检查**：提供定时器运行状态的健康检查接口

### 7.3 中期优化

1. **任务重试**：支持配置任务失败后的重试策略
2. **任务链**：支持任务之间的依赖关系（A 完成后触发 B）
3. **动态注册**：运行时动态添加/删除定时器
4. **分布式锁**：多实例部署时使用分布式锁避免重复执行

### 7.4 长期优化

1. **可视化监控**：提供定时任务运行状态的 Dashboard
2. **动态配置**：支持运行时修改定时规则，无需重启
3. **任务编排**：支持复杂的任务编排和依赖管理
4. **事件溯源**：支持任务执行历史持久化和回放

## 8. 附录

### 8.1 术语表

| 术语 | 说明 |
|------|------|
| Scheduler | 定时器，负责定时执行任务的组件 |
| Crontab | 定时表达式，用于定义任务的执行规则 |
| Tick | 定时触发事件 |
| Tick ID | 触发事件的唯一标识（计划执行时间的 Unix 时间戳） |
| Timezone | 时区，用于确定定时器的时间基准 |

### 8.2 参考资料

- [Crontab 表达式说明](https://en.wikipedia.org/wiki/Cron)
- [Crontab 表达式生成器](https://crontab.guru/)
- [Manager 设计文档](../manager/README.md)
- [容器架构文档](../container/README.md)
- [依赖注入设计](../AGENTS.md)
- [CLI 工具文档](../cli/README.md)

---

**文档结束**
