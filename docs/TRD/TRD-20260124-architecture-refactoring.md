# 架构重构技术需求文档

| 文档版本 | 日期 | 作者 |
|---------|------|------|
| 1.0 | 2026-01-24 | opencode |

## 1. 背景

### 1.1 重构目标

为了提升框架的可维护性和可扩展性，对 litecore-go 进行了以下架构重构：

1. **Manager 组件独立化** - 将 Manager 组件从 `server/builtin/manager` 迁移至 `manager` 目录，作为独立的基础能力层
2. **组件分层重组** - 将组件按类型重组为 `component/litecontroller`, `component/litemiddleware`, `component/liteservice`，提高代码组织性
3. **新增 Manager** - 新增 mqmgr、limitermgr、lockmgr 管理器，丰富基础能力

### 1.2 重构收益

- **清晰的模块边界** - Manager 组件与 Server 核心解耦，便于独立开发和测试
- **统一的包结构** - 组件按类型分层，降低开发者的理解成本
- **更好的可扩展性** - 新增 Manager 和组件更加便捷
- **增强的功能** - 新增限流、锁、消息队列等基础能力

## 2. Manager 组件重构

### 2.1 迁移前后对比

**迁移前（旧结构）**：
```
server/
  ├── builtin/
  │   ├── manager/
  │   │   ├── configmgr/
  │   │   ├── cachemgr/
  │   │   ├── databasemgr/
  │   │   ├── loggermgr/
  │   │   └── telemetrymgr/
  │   └── builtin.go
  ├── engine.go
  └── ...
```

**迁移后（新结构）**：
```
manager/
  ├── README.md
  ├── configmgr/
  ├── cachemgr/
  ├── databasemgr/
  ├── loggermgr/
  ├── telemetrymgr/
  ├── limitermgr/         # 新增
  ├── lockmgr/            # 新增
  └── mqmgr/              # 新增

server/
  ├── builtin.go
  ├── engine.go
  └── ...
```

### 2.2 Manager 列表

| Manager | 说明 | 接口 |
|---------|------|------|
| configmgr | 配置管理器 | `IConfigManager` |
| telemetrymgr | 可观测性管理器 | `ITelemetryManager` |
| loggermgr | 日志管理器 | `ILoggerManager` |
| databasemgr | 数据库管理器 | `IDatabaseManager` |
| cachemgr | 缓存管理器 | `ICacheManager` |
| lockmgr | 锁管理器 | `ILockManager` |
| limitermgr | 限流管理器 | `ILimiterManager` |
| mqmgr | 消息队列管理器 | `IMQManager` |

### 2.3 初始化顺序

Manager 初始化顺序（在 `server/builtin.go` 的 `Initialize` 函数中）：

1. ConfigManager（必须最先初始化，其他管理器依赖它）
2. TelemetryManager（依赖配置管理器）
3. LoggerManager（依赖配置管理器和遥测管理器）
4. DatabaseManager（依赖配置管理器）
5. CacheManager（依赖配置管理器）
6. LockManager（依赖配置管理器）
7. LimiterManager（依赖配置管理器）
8. MQManager（依赖配置管理器）

### 2.4 依赖关系图

```
configmgr
    ↓
    ├── loggermgr
    ├── cachemgr
    │       └── configmgr
    ├── databasemgr
    │       └── configmgr
    ├── limitermgr
    │       ├── configmgr
    │       └── cachemgr (可选)
    ├── lockmgr
    │       ├── configmgr
    │       └── cachemgr (可选)
    ├── mqmgr
    │       └── configmgr
    ├── telemetrymgr
    │       └── configmgr
    └── loggermgr
            └── configmgr
```

## 3. 组件分层重组

### 3.1 迁移前后对比

**迁移前（旧结构）**：
```
component/
  ├── controller/         # 控制器
  ├── middleware/        # 中间件
  ├── service/           # 服务
  └── ...
```

**迁移后（新结构）**：
```
component/
  ├── litecontroller/    # 控制器（HTTP 处理器）
  ├── litemiddleware/    # 中间件（请求拦截器）
  ├── liteservice/       # 服务（业务逻辑层）
  └── README.md
```

### 3.2 litecontroller（控制器层）

**职责**：处理 HTTP 请求，返回响应

**内置控制器**：
- `HealthController` - 健康检查
- `MetricsController` - Prometheus 指标
- `PProfController` - 性能分析
- `ResourceHTMLController` - HTML 资源
- `ResourceStaticController` - 静态资源

**接口**：`common.IBaseController`

```go
type IBaseController interface {
    ControllerName() string
    GetRouter() string
    Handle(c *gin.Context)
    OnStart() error
    OnStop() error
}
```

### 3.3 litemiddleware（中间件层）

**职责**：拦截请求，执行前置/后置逻辑

**内置中间件**：
- `RecoveryMiddleware` - 恢复（错误处理）
- `CORSMiddleware` - 跨域资源共享
- `RateLimiterMiddleware` - 限流
- `AuthMiddleware` - 认证（示例）
- `RequestLoggerMiddleware` - 请求日志
- `SecurityHeadersMiddleware` - 安全头
- `TelemetryMiddleware` - 可观测性

**接口**：`common.IBaseMiddleware`

```go
type IBaseMiddleware interface {
    MiddlewareName() string
    Order() int
    Wrapper() gin.HandlerFunc
    OnStart() error
    OnStop() error
}
```

**新增特性**：
- 支持通过配置自定义 `Name` 和 `Order`
- 示例配置：

```go
config := &RateLimiterConfig{
    Name:  pointer.String("CustomRateLimiter"),
    Order: pointer.Int(100),
    Limit: pointer.Int(200),
    Window: pointer.Duration(time.Minute),
}
middleware := NewRateLimiterMiddleware(config)
```

### 3.4 liteservice（服务层）

**职责**：实现业务逻辑，封装数据访问

**内置服务**：
- `HTMLTemplateService` - HTML 模板服务

**接口**：`common.IBaseService`

```go
type IBaseService interface {
    ServiceName() string
    OnStart() error
    OnStop() error
}
```

## 4. 新增 Manager 组件

### 4.1 limitermgr（限流管理器）

**功能**：提供基于时间窗口的请求频率控制

**支持的驱动**：
- `memory` - 内存限流
- `redis` - Redis 分布式限流

**核心功能**：
- `Allow(ctx, key, limit, window)` - 限流检查
- `GetRemaining(ctx, key, limit, window)` - 获取剩余配额

**使用场景**：
- API 限流
- 防止暴力破解
- 资源访问控制

**接口**：`ILimiterManager`

### 4.2 lockmgr（锁管理器）

**功能**：提供分布式锁功能

**支持的驱动**：
- `memory` - 内存锁
- `redis` - Redis 分布式锁

**核心功能**：
- `Lock(ctx, key, ttl)` - 阻塞锁获取
- `TryLock(ctx, key, ttl)` - 非阻塞锁尝试
- `Unlock(ctx, key)` - 锁释放

**使用场景**：
- 分布式任务调度
- 资源互斥访问
- 幂等性控制

**接口**：`ILockManager`

### 4.3 mqmgr（消息队列管理器）

**功能**：提供消息队列功能，支持异步消息处理

**支持的驱动**：
- `rabbitmq` - RabbitMQ
- `memory` - 内存队列（用于开发和测试）

**核心功能**：
- `Publish(ctx, exchange, routingKey, message)` - 消息发布
- `Subscribe(ctx, queue, callback)` - 消息订阅
- `Ack(ctx, deliveryTag)` - 消息确认
- `Nack(ctx, deliveryTag, requeue)` - 消息拒绝

**接口**：`IMQManager`

## 5. 技术细节

### 5.1 包路径变化

| 类型 | 迁移前 | 迁移后 |
|------|--------|--------|
| ConfigManager | `server/builtin/manager/configmgr` | `manager/configmgr` |
| LoggerManager | `server/builtin/manager/loggermgr` | `manager/loggermgr` |
| Controller | `component/controller` | `component/litecontroller` |
| Middleware | `component/middleware` | `component/litemiddleware` |
| Service | `component/service` | `component/liteservice` |

### 5.2 依赖注入模式

**Manager 注入**：

```go
type MyService struct {
    LoggerMgr  loggermgr.ILoggerManager  `inject:""`
    CacheMgr   cachemgr.ICacheManager   `inject:""`
    DBMgr      databasemgr.IDatabaseManager `inject:""`
}
```

**组件注入**：

```go
type MyController struct {
    Service    MyService `inject:""`
    LoggerMgr  loggermgr.ILoggerManager `inject:""`
}
```

### 5.3 初始化流程

```go
// 1. 创建容器
entityContainer := container.NewEntityContainer()
repositoryContainer := container.NewRepositoryContainer(entityContainer)
serviceContainer := container.NewServiceContainer(repositoryContainer)
controllerContainer := container.NewControllerContainer(serviceContainer)
middlewareContainer := container.NewMiddlewareContainer(serviceContainer)

// 2. 创建引擎（自动初始化 Manager 组件）
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

// 3. 运行引擎
engine.Run()
```

## 6. 向后兼容性

### 6.1 代码兼容性

- **Manager 接口保持不变** - 所有 Manager 接口定义保持一致
- **组件接口保持不变** - Controller、Middleware、Service 接口定义保持一致
- **依赖注入方式不变** - 继续使用 `inject:""` 标签

### 6.2 配置兼容性

- **配置路径不变** - 配置文件中的路径保持一致
- **配置格式不变** - YAML 配置格式保持一致

### 6.3 迁移指南

**代码迁移**：

```go
// 旧代码
import "github.com/lite-lake/litecore-go/server/builtin/manager/loggermgr"

// 新代码
import "github.com/lite-lake/litecore-go/manager/loggermgr"
```

**组件迁移**：

```go
// 旧代码
import "github.com/lite-lake/litecore-go/component/controller"

// 新代码
import "github.com/lite-lake/litecore-go/component/litecontroller"
```

## 7. 验收标准

### 7.1 功能验收

- [ ] 所有 Manager 组件正常初始化
- [ ] 所有内置组件正常注入
- [ ] 新增 Manager（limitermgr、lockmgr、mqmgr）功能正常
- [ ] 中间件支持自定义 Name 和 Order

### 7.2 架构验收

- [ ] Manager 组件与 Server 核心解耦
- [ ] 组件按类型分层（litecontroller、litemiddleware、liteservice）
- [ ] 包结构清晰，符合约定

### 7.3 兼容性验收

- [ ] 现有应用无需修改即可运行
- [ ] 配置文件格式兼容
- [ ] 依赖注入方式兼容

## 8. 实施计划

| 阶段 | 任务 | 产出 |
|------|------|------|
| 1 | Manager 组件迁移 | manager/ 目录 |
| 2 | 组件分层重组 | component/litecontroller/litemiddleware/liteservice |
| 3 | 新增 Manager | limitermgr、lockmgr、mqmgr |
| 4 | 更新依赖 | go.mod 更新 |
| 5 | 文档更新 | AGENTS.md、README 更新 |
| 6 | 测试验证 | 单元测试、集成测试 |

## 9. 附录

### 9.1 相关文档

- AGENTS.md - 开发指南
- manager/README.md - Manager 组件文档
- component/README.md - 组件文档

### 9.2 示例项目

- samples/messageboard - 消息板示例
