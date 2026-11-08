# Common - 公共基础接口

 定义 5 层依赖注入架构的基础接口，规范各层的行为契约和生命周期管理，并提供类型转换工具函数。交互层包含 4 种组件类型。

## 概述

common 包是框架的核心基础模块，定义了标准化的分层架构接口，所有业务组件都必须实现对应的基础接口。通过接口约束，确保架构的一致性和可维护性。

 ## 架构层次

 ```
 ┌─────────────────────────────────────────────────────────────┐
 │                    交互层 (Interaction Layer)              │
 │  Controller/Middleware/Listener/Scheduler               │
 │  (HTTP 请求、MQ 消息、定时任务的统一处理层)                 │
 ├─────────────────────────────────────────────────────────────┤
 │                       Service                              │
 │                    (业务逻辑层)                              │
 ├─────────────────────────────────────────────────────────────┤
 │                     Repository                              │
 │                    (数据访问层)                              │
 ├─────────────────────────────────────────────────────────────┤
 │                       Entity                                │
 │                    (数据实体层)                              │
 └─────────────────────────────────────────────────────────────┘
                         ↑
              ┌──────────┴──────────┐
              │  内置管理器层 (Manager Layer) │
              │ (基础设施管理器)     │
              └─────────────────────┘
      configmgr、databasemgr、loggermgr、cachemgr、
       lockmgr、limitermgr、mqmgr、telemetrymgr、schedulermgr
 ```

## 特性

 - **标准接口定义** - 定义 Entity、Manager、Repository、Service、Controller、Middleware、Listener、Scheduler 的标准接口
  - **实体基类** - 提供 3 种预定义的实体基类，支持 CUID2 ID 自动生成和时间戳自动填充
  - **生命周期管理** - 提供统一的 OnStart 和 OnStop 钩子方法
  - **命名规范** - 每层接口要求实现对应的名称方法，便于调试和日志
  - **依赖注入支持** - 为容器提供标准接口类型，支持类型安全的依赖注入
  - **类型转换工具** - 提供安全的类型转换函数，避免 panic 并支持默认值
  - **HTTP 状态码常量** - 定义完整的 HTTP 状态码常量，便于统一使用
  - **5层架构规范** - 明确各层的职责边界和依赖关系，确保架构清晰

 ## 快速开始

 ### 使用实体基类（推荐）

 框架提供了 3 种预定义的实体基类，支持 CUID2 ID 自动生成和时间戳自动填充：

 ```go
 import "github.com/lite-lake/litecore-go/common"

 // 方式 1：最常用的基类（ID + 创建时间 + 更新时间）
 type Message struct {
     common.BaseEntityWithTimestamps  // 自动生成 ID、CreatedAt、UpdatedAt
     Nickname  string `gorm:"type:varchar(20);not null"`
     Content   string `gorm:"type:varchar(500);not null"`
     Status    string `gorm:"type:varchar(20);default:'pending'"`
 }

 func (m *Message) EntityName() string {
     return "Message"
 }

 func (m *Message) TableName() string {
     return "messages"
 }

 func (m *Message) GetId() string {
     return m.ID  // ID 由基类提供，类型为 string
 }

 var _ common.IBaseEntity = (*Message)(nil)
 ```

 ```go
 // 方式 2：仅需要 ID 和创建时间（如日志实体）
 type AuditLog struct {
     common.BaseEntityWithCreatedAt  // 自动生成 ID、CreatedAt
     Action  string
     Details string
 }

 var _ common.IBaseEntity = (*AuditLog)(nil)
 ```

 ```go
 // 方式 3：只需要 ID（如配置表）
 type SystemConfig struct {
     common.BaseEntityOnlyID  // 自动生成 ID
     Key   string
     Value string
 }

 var _ common.IBaseEntity = (*SystemConfig)(nil)
 ```

 **基类特性**：
 - **CUID2 ID**：25 位字符串，时间有序、高唯一性、分布式安全
 - **数据库存储**：varchar(32)，预留更多兼容空间
 - **自动填充**：通过 GORM Hook 自动设置 ID 和时间戳
 - **类型安全**：ID 类型为 string，避免类型转换

 **注意事项**：
 - GORM 不会自动调用嵌入结构体的 Hook，必须手动调用父类 Hook
 - Repository 中查询 ID 时使用 `Where("id = ?", id)` 而不是 `First(entity, id)`

 ### 自定义实体（不使用基类）

 如果需要自定义 ID 生成逻辑或不使用时间戳，可以手动定义实体：

```go
import "github.com/lite-lake/litecore-go/common"

// 定义实体，实现 IBaseEntity 接口
type User struct {
    ID   string `gorm:"primaryKey"`
    Name string
}

func (u *User) EntityName() string {
    return "User"
}

func (u *User) TableName() string {
    return "users"
}

func (u *User) GetId() string {
    return u.ID
}

// 定义服务，实现 IBaseService 接口
type UserService struct {
    Config    configmgr.IConfigManager    `inject:""`
    LoggerMgr loggermgr.ILoggerManager   `inject:""`
}

func (s *UserService) ServiceName() string {
    return "UserService"
}

func (s *UserService) OnStart() error {
    return nil
}

func (s *UserService) OnStop() error {
    return nil
}

// 定义控制器，实现 IBaseController 接口
type UserController struct {
    Service   UserService                `inject:""`
    LoggerMgr loggermgr.ILoggerManager   `inject:""`
}

func (c *UserController) ControllerName() string {
    return "UserController"
}

func (c *UserController) GetRouter() string {
    return "/api/users [GET]"
}

func (c *UserController) Handle(ctx *gin.Context) {
    ctx.JSON(common.HTTPStatusOK, gin.H{"message": "success"})
}

// 定义中间件，实现 IBaseMiddleware 接口
type AuthMiddleware struct {
    Service   UserService                `inject:""`
}

func (m *AuthMiddleware) MiddlewareName() string {
    return "AuthMiddleware"
}

func (m *AuthMiddleware) Order() int {
    return 100
}

func (m *AuthMiddleware) Wrapper() gin.HandlerFunc {
    return func(ctx *gin.Context) {
        // 中间件逻辑
        ctx.Next()
    }
}

func (m *AuthMiddleware) OnStart() error {
    return nil
}

func (m *AuthMiddleware) OnStop() error {
    return nil
}

// 定义监听器，实现 IBaseListener 接口
type UserCreatedListener struct {
    LoggerMgr loggermgr.ILoggerManager   `inject:""`
}

func (l *UserCreatedListener) ListenerName() string {
    return "UserCreatedListener"
}

func (l *UserCreatedListener) GetQueue() string {
    return "user.created"
}

func (l *UserCreatedListener) GetSubscribeOptions() []common.ISubscribeOption {
    return nil
}

func (l *UserCreatedListener) Handle(ctx context.Context, msg common.IMessageListener) error {
    return nil
}

func (l *UserCreatedListener) OnStart() error {
    return nil
}

func (l *UserCreatedListener) OnStop() error {
    return nil
}

// 定义定时器，实现 IBaseScheduler 接口
type CleanupScheduler struct {
    Service   UserService                `inject:""`
}

func (s *CleanupScheduler) SchedulerName() string {
    return "CleanupScheduler"
}

func (s *CleanupScheduler) GetRule() string {
    return "0 0 2 * * *"
}

func (s *CleanupScheduler) GetTimezone() string {
    return "Asia/Shanghai"
}

func (s *CleanupScheduler) OnTick(tickID int64) error {
    return nil
}

func (s *CleanupScheduler) OnStart() error {
    return nil
}

func (s *CleanupScheduler) OnStop() error {
    return nil
}
```

## 核心接口

### IBaseEntity - 实体层接口

 定义数据实体的标准接口，所有实体必须实现：

 ```go
 type IBaseEntity interface {
     EntityName() string  // 返回实体名称，用于标识和调试
     TableName() string   // 返回数据库表名
     GetId() string       // 返回实体唯一标识
 }
 ```

 **命名规范**：
 - 实体结构体使用 PascalCase（如 `Message`、`User`）
 - 方法返回实体名称（如 `"Message"`）

 ### 实体基类（BaseEntity）

 框架提供了 3 种预定义的实体基类，支持 CUID2 ID 自动生成和时间戳自动填充：

 #### 1. BaseEntityOnlyID - 仅 ID

 **适用场景**：无需时间戳的实体（如配置表、字典表）

 ```go
 type BaseEntityOnlyID struct {
     ID string `gorm:"type:varchar(32);primarykey" json:"id"`
 }

 func (b *BaseEntityOnlyID) BeforeCreate(tx *gorm.DB) error {
     if b.ID == "" {
         newID, err := id.NewCUID2()
         if err != nil {
             return err
         }
         b.ID = newID
     }
     return nil
 }
 ```

 #### 2. BaseEntityWithCreatedAt - ID + 创建时间

 **适用场景**：只需要记录创建时间的实体（如日志、审计记录）

 ```go
 type BaseEntityWithCreatedAt struct {
     BaseEntityOnlyID
     CreatedAt time.Time `gorm:"type:timestamp;not null" json:"created_at"`
 }

 func (b *BaseEntityWithCreatedAt) BeforeCreate(tx *gorm.DB) error {
     if err := b.BaseEntityOnlyID.BeforeCreate(tx); err != nil {
         return err
     }
     if b.CreatedAt.IsZero() {
         b.CreatedAt = time.Now()
     }
     return nil
 }
 ```

 #### 3. BaseEntityWithTimestamps - ID + 创建时间 + 更新时间（最常用）

 **适用场景**：需要追踪创建和修改时间的实体

 ```go
 type BaseEntityWithTimestamps struct {
     BaseEntityWithCreatedAt
     UpdatedAt time.Time `gorm:"type:timestamp;not null" json:"updated_at"`
 }

 func (b *BaseEntityWithTimestamps) BeforeCreate(tx *gorm.DB) error {
     if err := b.BaseEntityWithCreatedAt.BeforeCreate(tx); err != nil {
         return err
     }
     if b.UpdatedAt.IsZero() {
         b.UpdatedAt = time.Now()
     }
     return nil
 }

 func (b *BaseEntityWithTimestamps) BeforeUpdate(tx *gorm.DB) error {
     b.UpdatedAt = time.Now()
     return nil
 }
 ```

 **使用示例**：

 ```go
 // 使用 BaseEntityWithTimestamps（最常用）
 type Message struct {
     common.BaseEntityWithTimestamps
     Nickname string `gorm:"type:varchar(20);not null" json:"nickname"`
     Content  string `gorm:"type:varchar(500);not null" json:"content"`
     Status   string `gorm:"type:varchar(20);default:'pending'" json:"status"`
 }

 func (m *Message) EntityName() string {
     return "Message"
 }

 func (m *Message) TableName() string {
     return "messages"
 }

 func (m *Message) GetId() string {
     return m.ID
 }

 var _ common.IBaseEntity = (*Message)(nil)
 ```

 **Repository 层无需修改**：

 ```go
 func (r *messageRepositoryImpl) Create(message *entities.Message) error {
     db := r.Manager.DB()
     return db.Create(message).Error  // Hook 自动填充 ID、CreatedAt、UpdatedAt
 }
 ```

 **Service 层代码简化**：

 ```go
 func (s *messageServiceImpl) CreateMessage(nickname, content string) (*entities.Message, error) {
     // 验证逻辑...

     message := &entities.Message{
         Nickname: nickname,
         Content:  content,
         Status:   "pending",
         // 不再需要手动设置 CreatedAt 和 UpdatedAt
     }

     if err := s.Repository.Create(message); err != nil {
         return nil, fmt.Errorf("failed to create message: %w", err)
     }

     return message, nil
 }
 ```

 **重要注意事项**：

 1. **GORM Hook 继承**：GORM 不会自动调用嵌入结构体的 Hook，必须手动调用父类的 Hook 方法
 2. **ID 类型**：使用 string 类型，CUID2 生成 25 位字符串，数据库存储 varchar(32)
 3. **Repository 查询**：使用 `Where("id = ?", id)` 而不是 `First(entity, id)`
 4. **并发安全**：CUID2 生成器可以在 goroutine 中并发使用

 **性能考虑**：
 - CUID2 生成比自增 ID 慢（约 10μs vs < 1μs）
 - 批量插入 1000 条记录可能需要 10ms 的额外开销
 - 如果性能是关键因素，可以考虑：
   - 使用 goroutine 并发生成 ID
   - 缓存一批预生成的 ID
   - 对关键表保留自增 ID（不使用基类）

### IBaseManager - 管理器层接口

定义资源管理器的标准接口，提供健康检查和生命周期管理：

```go
type IBaseManager interface {
    ManagerName() string  // 返回管理器名称
    Health() error        // 检查健康状态
    OnStart() error       // 启动时触发
    OnStop() error        // 停止时触发
}
```

**注意**：Manager 组件位于 `manager/` 目录下，由引擎自动初始化和注入。常见管理器包括：
- `configmgr` - 配置管理
- `databasemgr` - 数据库管理
- `loggermgr` - 日志管理
- `cachemgr` - 缓存管理
- `lockmgr` - 锁管理
- `limitermgr` - 限流管理
- `mqmgr` - 消息队列管理
- `telemetrymgr` - 可观测性管理

### IBaseRepository - 存储库层接口

定义数据访问层的标准接口：

```go
type IBaseRepository interface {
    RepositoryName() string  // 返回存储库名称
    OnStart() error          // 启动时触发
    OnStop() error           // 停止时触发
}
```

**命名规范**：
- 接口使用 `I` + 功能名 + `Repository`（如 `IMessageRepository`）
- 实现使用小驼峰 + `Impl` 后缀（如 `messageRepositoryImpl`）
- RepositoryName 返回实现类的名称（如 `"MessageRepository"`）

### IBaseService - 服务层接口

定义业务逻辑层的标准接口：

```go
type IBaseService interface {
    ServiceName() string  // 返回服务名称
    OnStart() error      // 启动时触发
    OnStop() error       // 停止时触发
}
```

**命名规范**：
- 接口使用 `I` + 功能名 + `Service`（如 `IMessageService`）
- 实现使用小驼峰 + `Impl` 后缀（如 `messageServiceImpl`）
- ServiceName 返回实现类的名称（如 `"MessageService"`）

### IBaseController - 控制器层接口

定义 HTTP 处理层的标准接口：

```go
type IBaseController interface {
    ControllerName() string              // 返回控制器名称
    GetRouter() string                   // 返回路由定义（OpenAPI @Router 规范）
    Handle(ctx *gin.Context)             // 处理请求
}
```

**命名规范**：
- 接口使用 `I` + 功能名 + `Controller`（如 `IMsgCreateController`）
- 实现使用小驼峰 + `Impl` 后缀（如 `msgCreateControllerImpl`）
- ControllerName 返回实现类的名称（如 `"msgCreateControllerImpl"`）
- GetRouter 返回路由格式：`/path [METHOD]`（如 `/api/messages [POST]`）

### IBaseMiddleware - 中间件层接口

定义中间件的标准接口：

```go
type IBaseMiddleware interface {
    MiddlewareName() string        // 返回中间件名称
    Order() int                   // 返回执行顺序，数值越小越先执行
    Wrapper() gin.HandlerFunc     // 返回 Gin 中间件函数
    OnStart() error               // 启动时触发
    OnStop() error                // 停止时触发
}
```

**命名规范**：
- 接口使用 `I` + 功能名 + `Middleware`（如 `IAuthMiddleware`）
- 实现使用小驼峰 + `Impl` 后缀（如 `authMiddlewareImpl`）
- MiddlewareName 返回中间件名称（如 `"AuthMiddleware"`）
- Order 返回执行顺序，数值越小越先执行（如 100）

### IBaseListener - 监听器层接口

定义消息监听器的标准接口，用于处理消息队列事件：

```go
type IBaseListener interface {
    ListenerName() string                    // 返回监听器名称
    GetQueue() string                        // 返回监听的队列名称
    GetSubscribeOptions() []ISubscribeOption // 返回订阅选项
    Handle(ctx context.Context, msg IMessageListener) error  // 处理消息
    OnStart() error                          // 启动时触发
    OnStop() error                           // 停止时触发
}
```

**IMessageListener 接口**：
```go
type IMessageListener interface {
    ID() string              // 获取消息 ID
    Body() []byte           // 获取消息体
    Headers() map[string]any // 获取消息头
}
```

**命名规范**：
- 接口使用 `I` + 功能名 + `Listener`（如 `IMessageCreatedListener`）
- 实现使用小驼峰 + `Impl` 后缀（如 `messageCreatedListenerImpl`）
- ListenerName 返回监听器名称（如 `"MessageCreatedListener"`）
- GetQueue 返回队列名称（如 `"message.created"`）
- Handle 方法处理消息，返回 error 会触发 Nack

### IBaseScheduler - 定时器层接口

定义定时任务的标准接口，用于处理周期性任务：

```go
type IBaseScheduler interface {
    SchedulerName() string  // 返回定时器名称
    GetRule() string       // 返回 Crontab 定时规则（6 段式）
    GetTimezone() string   // 返回时区（空字符串使用服务器本地时间）
    OnTick(tickID int64) error  // 定时触发时调用
    OnStart() error       // 启动时触发
    OnStop() error        // 停止时触发
}
```

**命名规范**：
- 接口使用 `I` + 功能名 + `Scheduler`（如 `ICleanupScheduler`）
- 实现使用小驼峰 + `Impl` 后缀（如 `cleanupSchedulerImpl`）
- SchedulerName 返回定时器名称（如 `"cleanupScheduler"`）
- GetRule 返回 Crontab 规则（如 `"0 0 2 * * *"` 表示每天凌晨 2 点）
- GetTimezone 返回时区（如 `"Asia/Shanghai"`、`"UTC"`）
- OnTick 方法接收 tickID（Unix 时间戳秒级）

 ## 依赖规则

 各层之间有明确的依赖关系：

 - **Entity 层**：无依赖
 - **Repository 层**：可依赖 Entity、Manager
 - **Service 层**：可依赖 Repository、Entity、Manager、其他 Service
 - **交互层 (Controller/Middleware/Listener/Scheduler)**：可依赖 Service、Manager

 **原则**：上层可以依赖下层，下层不能依赖上层。交互层统一处理 HTTP 请求、MQ 消息和定时任务，必须通过 Service 层访问数据。

## 依赖注入

所有基础接口都支持依赖注入，使用 `inject:""` 标签：

```go
type UserServiceImpl struct {
    // 内置管理器（由引擎自动注入）
    Config     configmgr.IConfigManager    `inject:""`
    LoggerMgr  loggermgr.ILoggerManager   `inject:""`
    DBManager  databasemgr.IDatabaseManager `inject:""`

    // 业务依赖
    UserRepo   IUserRepository            `inject:""`
    CacheMgr   cachemgr.ICacheManager     `inject:""`
}
```

## 生命周期管理

实现 IBaseManager、IBaseRepository、IBaseService、IBaseMiddleware 接口的组件，会在以下时机调用生命周期方法：

1. **OnStart** - 服务器启动时调用，用于初始化资源（如连接数据库、加载缓存等）
2. **OnStop** - 服务器停止时调用，用于清理资源（如关闭连接、刷新缓存等）

生命周期方法返回 error，如果初始化失败会阻止服务器启动。

## 类型转换工具

提供安全的类型转换函数，用于从 `any` 类型中获取特定类型的值：

```go
// GetString 从 any 类型中安全获取字符串值
func GetString(value any) (string, error)

// GetStringOrDefault 从 any 类型中安全获取字符串值，失败时返回默认值
func GetStringOrDefault(value any, defaultValue string) string

// GetMap 从 any 类型中安全获取 map[string]any 值
func GetMap(value any) (map[string]any, error)

// GetMapOrDefault 从 any 类型中安全获取 map[string]any 值，失败时返回默认值
func GetMapOrDefault(value any, defaultValue map[string]any) map[string]any
```

使用示例：

```go
// 从配置中获取字符串值
name, err := common.GetString(config["name"])
if err != nil {
    log.Error("无效的名称配置")
}

// 带默认值的字符串获取
timeout := common.GetStringOrDefault(config["timeout"], "30s")

// 从配置中获取 map 值
settings, err := common.GetMap(config["settings"])
if err != nil {
    log.Error("无效的设置配置")
}
```

## HTTP 状态码常量

定义完整的 HTTP 状态码常量，便于统一使用：

```go
const (
    HTTPStatusContinue                    = 100
    HTTPStatusOK                          = 200
    HTTPStatusCreated                     = 201
    HTTPStatusNoContent                   = 204
    HTTPStatusMovedPermanently            = 301
    HTTPStatusBadRequest                  = 400
    HTTPStatusUnauthorized                = 401
    HTTPStatusForbidden                   = 403
    HTTPStatusNotFound                    = 404
    HTTPStatusInternalServerError         = 500
    HTTPStatusServiceUnavailable          = 503
    // ... 更多状态码
)
```

使用示例：

```go
ctx.JSON(common.HTTPStatusOK, gin.H{"message": "success"})
ctx.JSON(common.HTTPStatusNotFound, gin.H{"error": "not found"})
ctx.JSON(common.HTTPStatusInternalServerError, gin.H{"error": "internal error"})
```

 ## 5 层架构统一接口规范

 ### 接口命名规律

 | 层级 | 接口命名 | 实现命名 | 示例 |
 |------|---------|---------|------|
 | Entity | 不需要单独接口 | PascalCase | `Message` |
 | Repository | `I` + 功能名 + `Repository` | 小驼峰 + `Impl` | `IMessageRepository` / `messageRepositoryImpl` |
 | Service | `I` + 功能名 + `Service` | 小驼峰 + `Impl` | `IMessageService` / `messageServiceImpl` |
 | Controller | `I` + 功能名 + `Controller` | 小驼峰 + `Impl` | `IMsgCreateController` / `msgCreateControllerImpl` |
 | Middleware | `I` + 功能名 + `Middleware` | 小驼峰 + `Impl` | `IAuthMiddleware` / `authMiddlewareImpl` |
 | Listener | `I` + 功能名 + `Listener` | 小驼峰 + `Impl` | `IMessageCreatedListener` / `messageCreatedListenerImpl` |
 | Scheduler | `I` + 功能名 + `Scheduler` | 小驼峰 + `Impl` | `ICleanupScheduler` / `cleanupSchedulerImpl` |

 **说明**：交互层包含 4 种组件类型（Controller/Middleware/Listener/Scheduler），它们位于同一架构层级，职责是处理不同类型的外部交互。

### 接口方法统一规范

| 接口类型 | 名称方法 | 生命周期方法 | 特殊方法 |
|---------|---------|-------------|---------|
| IBaseEntity | `EntityName()` | - | `TableName()`, `GetId()` |
| IBaseManager | `ManagerName()` | `OnStart()`, `OnStop()` | `Health()` |
| IBaseRepository | `RepositoryName()` | `OnStart()`, `OnStop()` | - |
| IBaseService | `ServiceName()` | `OnStart()`, `OnStop()` | - |
| IBaseController | `ControllerName()` | - | `GetRouter()`, `Handle()` |
| IBaseMiddleware | `MiddlewareName()` | `OnStart()`, `OnStop()` | `Order()`, `Wrapper()` |
| IBaseListener | `ListenerName()` | `OnStart()`, `OnStop()` | `GetQueue()`, `GetSubscribeOptions()`, `Handle()` |
| IBaseScheduler | `SchedulerName()` | `OnStart()`, `OnStop()` | `GetRule()`, `GetTimezone()`, `OnTick()` |

### 依赖注入规范

所有组件统一使用 `inject:""` 标签进行依赖注入：

```go
type messageServiceImpl struct {
    // 内置管理器（由引擎自动注入）
    Config    configmgr.IConfigManager    `inject:""`
    LoggerMgr loggermgr.ILoggerManager   `inject:""`
    DBManager databasemgr.IDatabaseManager `inject:""`

    // 业务依赖（手动注入到容器）
    Repository IMessageRepository `inject:""`
    CacheMgr  cachemgr.ICacheManager `inject:""`
}
```

### 接口编译时检查

所有实现都应在文件末尾添加编译时接口检查：

```go
var _ IMessageService = (*messageServiceImpl)(nil)
var _ common.IBaseService = (*messageServiceImpl)(nil)
```

 ### 组件注册规范

 1. **Entity**：在 `entity_container.go` 中注册
 2. **Repository**：在 `repository_container.go` 中注册
 3. **Service**：在 `service_container.go` 中注册
 4. **交互层组件**：
    - **Controller**：在 `controller_container.go` 中注册
    - **Middleware**：在 `middleware_container.go` 中注册
    - **Listener**：在 `listener_container.go` 中注册
    - **Scheduler**：在 `scheduler_container.go` 中注册

 ### 层级职责边界

 | 层级 | 职责 | 允许依赖 |
 |------|------|---------|
 | Entity | 数据模型定义 | 无 |
 | Repository | 数据访问、持久化 | Entity、Manager |
 | Service | 业务逻辑、编排 | Repository、Entity、Manager、其他 Service |
 | 交互层 (Controller) | HTTP 请求处理、响应 | Service、Manager |
 | 交互层 (Middleware) | 请求预处理、后处理 | Service、Manager |
 | 交互层 (Listener) | 消息队列事件处理 | Service、Manager |
 | 交互层 (Scheduler) | 定时任务执行 | Service、Manager |

### 生命周期方法规范

**OnStart** 方法用于：
- 初始化资源（连接数据库、连接缓存等）
- 预热缓存
- 注册定时任务
- 启动消息监听

**OnStop** 方法用于：
- 关闭数据库连接
- 刷新缓存数据
- 停止定时任务
- 取消消息订阅
- 释放其他资源

生命周期方法返回 `error`，如果初始化失败会阻止服务器启动。

### 快速开发模板

```go
// 接口定义
type IXxxService interface {
    common.IBaseService
    // 业务方法
    DoSomething() error
}

// 实现结构体
type xxxServiceImpl struct {
    Config    configmgr.IConfigManager    `inject:""`
    LoggerMgr loggermgr.ILoggerManager   `inject:""`
}

// 构造函数
func NewXxxService() IXxxService {
    return &xxxServiceImpl{}
}

// 实现 IBaseService
func (s *xxxServiceImpl) ServiceName() string { return "XxxService" }
func (s *xxxServiceImpl) OnStart() error { return nil }
func (s *xxxServiceImpl) OnStop() error { return nil }

// 实现业务方法
func (s *xxxServiceImpl) DoSomething() error {
    return nil
}

// 编译时接口检查
var _ IXxxService = (*xxxServiceImpl)(nil)
var _ common.IBaseService = (*xxxServiceImpl)(nil)
```

## 最佳实践

1. **接口实现** - 确保所有组件实现对应的基础接口（以 `I` 开头）
2. **命名规范** - 使用结构体类型名作为名称方法返回值
3. **生命周期** - 在 OnStart 中初始化资源，在 OnStop 中清理资源
4. **依赖关系** - 严格遵循分层架构的依赖规则
5. **错误处理** - 生命周期方法中的错误应该被正确处理和传播
6. **类型转换** - 使用 common 包提供的类型转换工具函数，避免直接类型断言导致的 panic
7. **HTTP 状态码** - 使用 common 包定义的 HTTP 状态码常量，保持代码一致性

## 实际应用示例

参考 `samples/messageboard` 目录下的完整示例：

```go
// entities/message_entity.go（使用基类）
type Message struct {
    common.BaseEntityWithTimestamps
    Nickname  string `gorm:"type:varchar(20);not null"`
    Content   string `gorm:"type:varchar(500);not null"`
    Status    string `gorm:"type:varchar(20);default:'pending'"`
}

func (m *Message) EntityName() string { return "Message" }
func (m *Message) TableName() string { return "messages" }
func (m *Message) GetId() string { return m.ID }

var _ common.IBaseEntity = (*Message)(nil)

// repositories/message_repository.go
type IMessageRepository interface {
    common.IBaseRepository
    Create(message *entities.Message) error
    GetByID(id string) (*entities.Message, error)  // ID 类型改为 string
}

type messageRepositoryImpl struct {
    Config  configmgr.IConfigManager    `inject:""`
    Manager databasemgr.IDatabaseManager `inject:""`
}

func NewMessageRepository() IMessageRepository {
    return &messageRepositoryImpl{}
}

func (r *messageRepositoryImpl) RepositoryName() string { return "MessageRepository" }
func (r *messageRepositoryImpl) OnStart() error { return nil }
func (r *messageRepositoryImpl) OnStop() error { return nil }

func (r *messageRepositoryImpl) Create(message *entities.Message) error {
    db := r.Manager.DB()
    return db.Create(message).Error
}

func (r *messageRepositoryImpl) GetByID(id string) (*entities.Message, error) {
    db := r.Manager.DB()
    var message entities.Message
    err := db.Where("id = ?", id).First(&message).Error  // 使用 Where 查询
    if err != nil {
        return nil, err
    }
    return &message, nil
}

var _ IMessageRepository = (*messageRepositoryImpl)(nil)

// services/message_service.go
type IMessageService interface {
    common.IBaseService
    CreateMessage(nickname, content string) (*entities.Message, error)  // ID 类型改为 string
    GetApprovedMessages() ([]*entities.Message, error)
    UpdateMessageStatus(id string, status string) error  // ID 类型改为 string
    DeleteMessage(id string) error  // ID 类型改为 string
}

type messageServiceImpl struct {
    Config     configmgr.IConfigManager     `inject:""`
    Repository IMessageRepository           `inject:""`
    LoggerMgr  loggermgr.ILoggerManager    `inject:""`
}

func NewMessageService() IMessageService {
    return &messageServiceImpl{}
}

func (s *messageServiceImpl) ServiceName() string { return "MessageService" }
func (s *messageServiceImpl) OnStart() error { return nil }
func (s *messageServiceImpl) OnStop() error { return nil }

func (s *messageServiceImpl) CreateMessage(nickname, content string) (*entities.Message, error) {
    // 验证逻辑...

    message := &entities.Message{
        Nickname: nickname,
        Content:  content,
        Status:   "pending",
        // 不再需要手动设置 CreatedAt 和 UpdatedAt，由 Hook 自动填充
    }

    if err := s.Repository.Create(message); err != nil {
        s.LoggerMgr.Ins().Error("Failed to create message", "nickname", nickname, "error", err)
        return nil, fmt.Errorf("failed to create message: %w", err)
    }

    s.LoggerMgr.Ins().Info("Message created successfully", "id", message.ID, "nickname", message.Nickname)
    return message, nil
}

func (s *messageServiceImpl) UpdateMessageStatus(id string, status string) error {
    // ID 类型为 string，直接使用
    message, err := s.Repository.GetByID(id)
    if err != nil {
        return fmt.Errorf("message not found: %w", err)
    }

    if err := s.Repository.UpdateStatus(id, status); err != nil {
        return fmt.Errorf("failed to update message status: %w", err)
    }

    s.LoggerMgr.Ins().Info("Message status updated successfully", "id", id, "status", status)
    return nil
}

func (s *messageServiceImpl) DeleteMessage(id string) error {
    // ID 类型为 string，直接使用
    message, err := s.Repository.GetByID(id)
    if err != nil {
        return fmt.Errorf("message not found: %w", err)
    }

    if err := s.Repository.Delete(id); err != nil {
        return fmt.Errorf("failed to delete message: %w", err)
    }

    s.LoggerMgr.Ins().Info("Message deleted successfully", "id", id)
    return nil
}

var _ IMessageService = (*messageServiceImpl)(nil)

// controllers/msg_create_controller.go
type IMsgCreateController interface {
    common.IBaseController
}

type msgCreateControllerImpl struct {
    MessageService IMessageService         `inject:""`
    LoggerMgr      loggermgr.ILoggerManager `inject:""`
}

func NewMsgCreateController() IMsgCreateController {
    return &msgCreateControllerImpl{}
}

func (c *msgCreateControllerImpl) ControllerName() string { return "msgCreateControllerImpl" }
func (c *msgCreateControllerImpl) GetRouter() string { return "/api/messages [POST]" }

func (c *msgCreateControllerImpl) Handle(ctx *gin.Context) {
    var req dtos.CreateMessageRequest
    if err := ctx.ShouldBindJSON(&req); err != nil {
        c.LoggerMgr.Ins().Error("Failed to create message: parameter binding error", "error", err)
        ctx.JSON(common.HTTPStatusBadRequest, dtos.ErrorResponse(common.HTTPStatusBadRequest, err.Error()))
        return
    }

    message, err := c.MessageService.CreateMessage(req.Nickname, req.Content)
    if err != nil {
        c.LoggerMgr.Ins().Error("Failed to create message", "nickname", req.Nickname, "error", err)
        ctx.JSON(common.HTTPStatusBadRequest, dtos.ErrorResponse(common.HTTPStatusBadRequest, err.Error()))
        return
    }

    c.LoggerMgr.Ins().Info("Message created successfully", "id", message.ID)
    ctx.JSON(common.HTTPStatusOK, dtos.SuccessResponse("留言提交成功，等待审核", gin.H{
        "id": message.ID,
    }))
}

var _ IMsgCreateController = (*msgCreateControllerImpl)(nil)

// controllers/msg_delete_controller.go
type IMsgDeleteController interface {
    common.IBaseController
}

type msgDeleteControllerImpl struct {
    MessageService IMessageService         `inject:""`
    LoggerMgr      loggermgr.ILoggerManager `inject:""`
}

func NewMsgDeleteController() IMsgDeleteController {
    return &msgDeleteControllerImpl{}
}

func (c *msgDeleteControllerImpl) ControllerName() string { return "msgDeleteControllerImpl" }
func (c *msgDeleteControllerImpl) GetRouter() string { return "/api/admin/messages/:id/delete [POST]" }

func (c *msgDeleteControllerImpl) Handle(ctx *gin.Context) {
    id := ctx.Param("id")  // ID 类型为 string，直接使用，无需解析

    if err := c.MessageService.DeleteMessage(id); err != nil {
        c.LoggerMgr.Ins().Error("Failed to delete message", "id", id, "error", err)
        ctx.JSON(common.HTTPStatusBadRequest, dtos.ErrorResponse(common.HTTPStatusBadRequest, err.Error()))
        return
    }

    c.LoggerMgr.Ins().Info("Message deleted successfully", "id", id)
    ctx.JSON(common.HTTPStatusOK, dtos.SuccessWithMessage("删除成功"))
}

var _ IMsgDeleteController = (*msgDeleteControllerImpl)(nil)

// middlewares/auth_middleware.go
type IAuthMiddleware interface {
    common.IBaseMiddleware
}

type authMiddlewareImpl struct {
    Service   IMessageService `inject:""`
}

func NewAuthMiddleware() IAuthMiddleware {
    return &authMiddlewareImpl{}
}

func (m *authMiddlewareImpl) MiddlewareName() string { return "AuthMiddleware" }
func (m *authMiddlewareImpl) Order() int { return 100 }
func (m *authMiddlewareImpl) Wrapper() gin.HandlerFunc {
    return func(ctx *gin.Context) {
        // 中间件逻辑
        ctx.Next()
    }
}
func (m *authMiddlewareImpl) OnStart() error { return nil }
func (m *authMiddlewareImpl) OnStop() error { return nil }

var _ IAuthMiddleware = (*authMiddlewareImpl)(nil)

// listeners/message_created_listener.go
type IMessageCreatedListener interface {
    common.IBaseListener
}

type messageCreatedListenerImpl struct {
    LoggerMgr loggermgr.ILoggerManager `inject:""`
}

func NewMessageCreatedListener() IMessageCreatedListener {
    return &messageCreatedListenerImpl{}
}

func (l *messageCreatedListenerImpl) ListenerName() string { return "MessageCreatedListener" }
func (l *messageCreatedListenerImpl) GetQueue() string { return "message.created" }
func (l *messageCreatedListenerImpl) GetSubscribeOptions() []common.ISubscribeOption { return nil }
func (l *messageCreatedListenerImpl) Handle(ctx context.Context, msg common.IMessageListener) error { return nil }
func (l *messageCreatedListenerImpl) OnStart() error { return nil }
func (l *messageCreatedListenerImpl) OnStop() error { return nil }

var _ IMessageCreatedListener = (*messageCreatedListenerImpl)(nil)

// schedulers/cleanup_scheduler.go
type ICleanupScheduler interface {
    common.IBaseScheduler
}

type cleanupSchedulerImpl struct {
    MessageService IMessageService `inject:""`
}

func NewCleanupScheduler() ICleanupScheduler {
    return &cleanupSchedulerImpl{}
}

func (s *cleanupSchedulerImpl) SchedulerName() string { return "cleanupScheduler" }
func (s *cleanupSchedulerImpl) GetRule() string { return "0 0 2 * * *" }
func (s *cleanupSchedulerImpl) GetTimezone() string { return "Asia/Shanghai" }
func (s *cleanupSchedulerImpl) OnTick(tickID int64) error { return nil }
func (s *cleanupSchedulerImpl) OnStart() error { return nil }
func (s *cleanupSchedulerImpl) OnStop() error { return nil }

var _ ICleanupScheduler = (*cleanupSchedulerImpl)(nil)
```

## 与其他包的关系

- **manager/** - Manager 组件实现 IBaseManager 接口，作为基础设施层提供各种能力
- **container/** - 依赖注入容器使用 common 包定义的接口类型进行类型安全的依赖注入
- **component/** - 业务组件实现 IBaseEntity、IBaseRepository、IBaseService、IBaseController、IBaseMiddleware 接口
- **util/** - 工具函数包提供特定功能的工具函数，common 包提供通用的类型转换函数
- **server/** - 服务器引擎负责管理所有组件的生命周期（OnStart/OnStop），并按规则调度 Listener 和 Scheduler
