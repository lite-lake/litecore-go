# Manager 依赖注入缺陷修复技术需求文档

| 文档版本 | 日期 | 作者 |
|---------|------|------|
| 1.0 | 2026-01-24 | opencode |

## 1. 问题概述

### 1.1 问题描述

`manager/` 层的所有 Manager 组件定义了 `inject:""` 标签用于依赖注入，但这些依赖**从未被注入**，导致：

1. **可观测性功能完全失效** - `Logger` 和 `telemetryMgr` 始终为 `nil`
2. **跨 Manager 依赖不工作** - 如 `limitermgr` 依赖 `cachemgr`，但 `CacheMgr` 为 `nil`
3. **代码注释与实际行为不一致** - Factory 注释写"需要通过容器注入"，但实际未实现

### 1.2 受影响范围

| Manager | 受影响功能 | 后果 |
|---------|-----------|------|
| cachemgr | 日志记录、指标、链路追踪 | 无任何可观测性输出 |
| databasemgr | 日志记录、指标、慢查询检测、链路追踪 | 无任何可观测性输出 |
| lockmgr | 日志记录、指标、链路追踪、Redis 缓存依赖 | Redis 锁功能失效 |
| limitermgr | 日志记录、指标、链路追踪、Redis 缓存依赖 | Redis 限流器报错 |
| mqmgr | 日志记录、指标、链路追踪 | 无任何可观测性输出 |
| loggermgr | 正常（构造函数显式传入） | ✅ 不受影响 |
| telemetrymgr | 正常（无依赖注入需求） | ✅ 不受影响 |
| configmgr | 正常（无依赖注入需求） | ✅ 不受影响 |

### 1.3 根本原因

**架构设计矛盾：**

1. **Manager 层理论上不支持依赖注入**（AGENTS.md 明确指出 "Manager → Config + other Managers"）
2. **代码却定义了 inject 标签**（impl_base.go 中定义）
3. **容器从未对 Manager 执行注入**（Engine.autoInject 只注入 Repository/Service/Controller/Middleware/Listener/Scheduler）
4. **Factory 方法不传入依赖**（Build/BuildWithConfigProvider 只接收配置参数）

**依赖链断裂示例（以 limitermgr 为例）：**

```
limitermgr/redis_impl.go:14
  CacheMgr cachemgr.ICacheManager `inject:""`  // ← 定义了标签

limitermgr/redis_impl.go:82
  if r.CacheMgr == nil {  // ← 永远为 nil
      return fmt.Errorf("cache manager is not initialized")
  }
```

**可观测性失效示例（以 databasemgr 为例）：**

```go
// databasemgr/impl_base.go:24-25
Logger       logger.ILogger                 `inject:""`
telemetryMgr telemetrymgr.ITelemetryManager `inject:""`

// 初始化可观测性时
func (b *databaseManagerBaseImpl) initObservability(cfg *DatabaseConfig) {
    if b.telemetryMgr == nil {  // ← 永远为 nil
        return  // 直接返回，不初始化指标
    }
    // ... 指标初始化代码永远不会执行
}
```

## 2. 当前实现分析

### 2.1 现有 inject 标签位置

| 文件 | 字段 | 类型 |
|------|------|------|
| manager/cachemgr/impl_base.go:20 | Logger | `logger.ILogger` |
| manager/cachemgr/impl_base.go:22 | telemetryMgr | `telemetrymgr.ITelemetryManager` |
| manager/databasemgr/impl_base.go:24 | Logger | `logger.ILogger` |
| manager/databasemgr/impl_base.go:25 | telemetryMgr | `telemetrymgr.ITelemetryManager` |
| manager/lockmgr/impl_base.go:22 | Logger | `logger.ILogger` |
| manager/lockmgr/impl_base.go:24 | telemetryMgr | `telemetrymgr.ITelemetryManager` |
| manager/lockmgr/impl_base.go:26 | cacheMgr | `cachemgr.ICacheManager` |
| manager/limitermgr/impl_base.go:21 | Logger | `logger.ILogger` |
| manager/limitermgr/impl_base.go:23 | telemetryMgr | `telemetrymgr.ITelemetryManager` |
| manager/limitermgr/impl_base.go:25 | cacheMgr | `cachemgr.ICacheManager` |
| manager/limitermgr/redis_impl.go:14 | CacheMgr | `cachemgr.ICacheManager` |
| manager/mqmgr/impl_base.go:18 | Logger | `logger.ILogger` |
| manager/mqmgr/impl_base.go:19 | telemetryMgr | `telemetrymgr.ITelemetryManager` |

### 2.2 当前 Factory 方法签名

```go
// manager/cachemgr/factory.go
func Build(driverType string, driverConfig map[string]any) (ICacheManager, error)
func BuildWithConfigProvider(configProvider configmgr.IConfigManager) (ICacheManager, error)

// manager/databasemgr/factory.go
func Build(driverType string, driverConfig map[string]any) (IDatabaseManager, error)
func BuildWithConfigProvider(configProvider configmgr.IConfigManager) (IDatabaseManager, error)

// manager/lockmgr/factory.go
func Build(driverType string, driverConfig map[string]any) (ILockManager, error)
func BuildWithConfigProvider(configProvider configmgr.IConfigManager) (ILockManager, error)

// manager/limitermgr/factory.go
func Build(driverType string, driverConfig map[string]any) (ILimiterManager, error)
func BuildWithConfigProvider(configProvider configmgr.IConfigManager) (ILimiterManager, error)

// manager/mqmgr/factory.go
func Build(driverType string, driverConfig map[string]any) (IMQManager, error)
func BuildWithConfigProvider(configProvider configmgr.IConfigManager) (IMQManager, error)
```

### 2.3 当前 builtin.go 初始化逻辑

```go
// server/builtin.go:Initialize
func Initialize(cfg *BuiltinConfig) (*container.ManagerContainer, error) {
    cntr := container.NewManagerContainer()

    // 1. 初始化配置管理器
    configManager, err := configmgr.Build(cfg.Driver, cfg.FilePath)
    container.RegisterManager[configmgr.IConfigManager](cntr, configManager)

    // 2. 初始化遥测管理器
    telemetryMgr, err := telemetrymgr.BuildWithConfigProvider(configManager)
    container.RegisterManager[telemetrymgr.ITelemetryManager](cntr, telemetryMgr)

    // 3. 初始化日志管理器
    loggerManager, err := loggermgr.BuildWithConfigProvider(configManager, telemetryMgr)
    container.RegisterManager[loggermgr.ILoggerManager](cntr, loggerManager)

    // 4-9. 初始化其他管理器（都不传入依赖）
    databaseMgr, err := databasemgr.BuildWithConfigProvider(configManager)
    cacheMgr, err := cachemgr.BuildWithConfigProvider(configManager)
    lockMgr, err := lockmgr.BuildWithConfigProvider(configManager)
    limiterMgr, err := limitermgr.BuildWithConfigProvider(configManager)
    mqMgr, err := mqmgr.BuildWithConfigProvider(configManager)
    schedulerMgr, err := schedulermgr.BuildWithConfigProvider(configManager)
    // ...
}
```

**问题：** Factory 方法不接受 Logger、TelemetryMgr 等依赖参数，impl_base 的字段始终为 nil。

### 2.4 ManagerContainer 不支持注入

```go
// container/manager_container.go
type ManagerContainer struct {
    container *TypedContainer[common.IBaseManager]
}

// ❌ 没有 InjectAll 方法
// ✅ 只有 RegisterByType、GetByType 等注册/获取方法
```

```go
// server/engine.go:autoInject
func (e *Engine) autoInject() error {
    // ❌ 没有 e.Manager.InjectAll()
    // ✅ 只对以下层执行注入
    e.Repository.InjectAll()
    e.Service.InjectAll()
    e.Controller.InjectAll()
    e.Middleware.InjectAll()
    e.Listener.InjectAll()
    e.Scheduler.InjectAll()
}
```

## 3. 修复方案

### 3.1 设计原则

**按架构设计，Manager 层不支持依赖注入，依赖应该在 Factory 构造时显式传入。**

### 3.2 修复步骤

#### 步骤 1：修改 impl_base.go，移除 inject 标签

将字段从：
```go
Logger       logger.ILogger                 `inject:""`
telemetryMgr telemetrymgr.ITelemetryManager `inject:""`
cacheMgr     cachemgr.ICacheManager         `inject:""`
```

改为：
```go
loggerMgr    loggermgr.ILoggerManager
telemetryMgr telemetrymgr.ITelemetryManager
cacheMgr     cachemgr.ICacheManager
```

#### 步骤 2：修改 Factory 方法签名，接收依赖参数

以 `cachemgr` 为例：

```go
// manager/cachemgr/factory.go

// Build 创建缓存管理器实例（新增依赖参数）
func Build(
    driverType string,
    driverConfig map[string]any,
    loggerMgr loggermgr.ILoggerManager,
    telemetryMgr telemetrymgr.ITelemetryManager,
) (ICacheManager, error)

// BuildWithConfigProvider 从配置提供者创建缓存管理器实例（新增依赖参数）
func BuildWithConfigProvider(
    configProvider configmgr.IConfigManager,
    loggerMgr loggermgr.ILoggerManager,
    telemetryMgr telemetrymgr.ITelemetryManager,
) (ICacheManager, error)
```

#### 步骤 3：修改 Factory 内部调用，传入依赖

```go
// manager/cachemgr/factory.go

func Build(
    driverType string,
    driverConfig map[string]any,
    loggerMgr loggermgr.ILoggerManager,
    telemetryMgr telemetrymgr.ITelemetryManager,
) (ICacheManager, error) {
    switch driverType {
    case "redis":
        redisConfig, err := parseRedisConfig(driverConfig)
        if err != nil {
            return nil, err
        }

        // ← 传入依赖
        mgr, err := NewCacheManagerRedisImpl(redisConfig, loggerMgr, telemetryMgr)
        if err != nil {
            return nil, err
        }
        return mgr, nil

    case "memory":
        // Memory 不需要日志和遥测，可以传入 nil
        memoryConfig, err := parseMemoryConfig(driverConfig)
        if err != nil {
            return nil, err
        }
        mgr := NewCacheManagerMemoryImpl(
            memoryConfig.MaxAge,
            memoryConfig.MaxAge/2,
        )
        // ← MemoryImpl 也需要支持设置依赖（可选）
        return mgr, nil

    case "none":
        mgr := NewCacheManagerNoneImpl()
        return mgr, nil

    default:
        return nil, fmt.Errorf("unsupported driver type: %s", driverType)
    }
}

func BuildWithConfigProvider(
    configProvider configmgr.IConfigManager,
    loggerMgr loggermgr.ILoggerManager,
    telemetryMgr telemetrymgr.ITelemetryManager,
) (ICacheManager, error) {
    // ... 解析配置 ...

    // ← 调用 Build 时传入依赖
    return Build(driverTypeStr, driverConfig, loggerMgr, telemetryMgr)
}
```

#### 步骤 4：修改 Manager 实现构造函数

```go
// manager/cachemgr/redis_impl.go

// NewCacheManagerRedisImpl 创建 Redis 实现（新增依赖参数）
func NewCacheManagerRedisImpl(
    cfg *RedisConfig,
    loggerMgr loggermgr.ILoggerManager,
    telemetryMgr telemetrymgr.ITelemetryManager,
) (ICacheManager, error) {
    // ... 创建 Redis 客户端 ...

    impl := &cacheManagerRedisImpl{
        cacheManagerBaseImpl: newICacheManagerBaseImpl(loggerMgr, telemetryMgr), // ← 传入依赖
        client:               client,
        name:                 "cacheManagerRedisImpl",
    }
    impl.initObservability() // ← 现在可以正常初始化
    return impl, nil
}

// manager/cachemgr/impl_base.go

// newICacheManagerBaseImpl 创建基类（接收依赖参数）
func newICacheManagerBaseImpl(
    loggerMgr loggermgr.ILoggerManager,
    telemetryMgr telemetrymgr.ITelemetryManager,
) *cacheManagerBaseImpl {
    return &cacheManagerBaseImpl{
        loggerMgr:     loggerMgr,
        telemetryMgr:  telemetryMgr,
    }
}
```

#### 步骤 5：修改 builtin.go，按依赖顺序传入依赖

```go
// server/builtin.go:Initialize

func Initialize(cfg *BuiltinConfig) (*container.ManagerContainer, error) {
    cntr := container.NewManagerContainer()

    // 1. 初始化配置管理器
    configManager, err := configmgr.Build(cfg.Driver, cfg.FilePath)
    container.RegisterManager[configmgr.IConfigManager](cntr, configManager)

    // 2. 初始化遥测管理器
    telemetryMgr, err := telemetrymgr.BuildWithConfigProvider(configManager)
    container.RegisterManager[telemetrymgr.ITelemetryManager](cntr, telemetryMgr)

    // 3. 初始化日志管理器
    loggerManager, err := loggermgr.BuildWithConfigProvider(configManager, telemetryMgr)
    container.RegisterManager[loggermgr.ILoggerManager](cntr, loggerManager)

    // 4. 初始化缓存管理器（传入 loggerManager 和 telemetry）
    cacheMgr, err := cachemgr.BuildWithConfigProvider(configManager, loggerManager, telemetryMgr)
    container.RegisterManager[cachemgr.ICacheManager](cntr, cacheMgr)

    // 5. 初始化数据库管理器（传入 loggerManager 和 telemetry）
    databaseMgr, err := databasemgr.BuildWithConfigProvider(configManager, loggerManager, telemetryMgr)
    container.RegisterManager[databasemgr.IDatabaseManager](cntr, databaseMgr)

    // 6. 初始化锁管理器（传入 loggerManager、telemetry 和 cacheMgr）
    lockMgr, err := lockmgr.BuildWithConfigProvider(configManager, loggerManager, telemetryMgr, cacheMgr)
    container.RegisterManager[lockmgr.ILockManager](cntr, lockMgr)

    // 7. 初始化限流管理器（传入 loggerManager、telemetry 和 cacheMgr）
    limiterMgr, err := limitermgr.BuildWithConfigProvider(configManager, loggerManager, telemetryMgr, cacheMgr)
    container.RegisterManager[limitermgr.ILimiterManager](cntr, limiterMgr)

    // 8. 初始化消息队列管理器（传入 loggerManager 和 telemetry）
    mqMgr, err := mqmgr.BuildWithConfigProvider(configManager, loggerManager, telemetryMgr)
    container.RegisterManager[mqmgr.IMQManager](cntr, mqMgr)

    // 9. 初始化定时任务管理器（传入 loggerManager）
    schedulerMgr, err := schedulermgr.BuildWithConfigProvider(configManager, loggerManager)
    container.RegisterManager[schedulermgr.ISchedulerManager](cntr, schedulerMgr)

    return cntr, nil
}
```

### 3.3 依赖顺序

```
ConfigManager (无依赖)
    ↓
TelemetryManager (依赖 ConfigManager)
    ↓
LoggerManager (依赖 ConfigManager, TelemetryManager)
    ↓
DatabaseManager (依赖 LoggerManager, TelemetryManager)
CacheManager (依赖 LoggerManager, TelemetryManager)
    ↓
LockManager (依赖 LoggerManager, TelemetryManager, CacheManager)
LimiterManager (依赖 LoggerManager, TelemetryManager, CacheManager)
MQManager (依赖 LoggerManager, TelemetryManager)
SchedulerManager (依赖 LoggerManager)
```

## 4. 需要修改的文件清单

### 4.1 Manager 层文件

| 文件 | 修改内容 |
|------|---------|
| manager/cachemgr/impl_base.go | 1. 移除字段 `inject:""` 标签，字段类型改为 `loggerMgr loggermgr.ILoggerManager`<br>2. 修改 `newICacheManagerBaseImpl` 接收 loggerMgr 和 telemetryMgr 参数 |
| manager/cachemgr/redis_impl.go | 修改 `NewCacheManagerRedisImpl` 接收 loggerMgr 和 telemetryMgr 参数 |
| manager/cachemgr/memory_impl.go | 可选：支持传入依赖（若 Memory 也需要日志） |
| manager/cachemgr/none_impl.go | 无需修改 |
| manager/cachemgr/factory.go | 1. `Build` 函数新增 loggerMgr 和 telemetryMgr 参数<br>2. `BuildWithConfigProvider` 函数新增 loggerMgr 和 telemetryMgr 参数<br>3. 修改内部调用，传入依赖<br>4. 更新注释（移除"需要通过容器注入"） |

| 文件 | 修改内容 |
|------|---------|
| manager/databasemgr/impl_base.go | 1. 移除字段 `inject:""` 标签，字段类型改为 `loggerMgr loggermgr.ILoggerManager`<br>2. 修改 `newIDatabaseManagerBaseImpl` 接收 loggerMgr 和 telemetryMgr 参数 |
| manager/databasemgr/mysql_impl.go | 修改 `NewDatabaseManagerMySQLImpl` 接收 loggerMgr 和 telemetryMgr 参数 |
| manager/databasemgr/postgresql_impl.go | 修改 `NewDatabaseManagerPostgreSQLImpl` 接收 loggerMgr 和 telemetryMgr 参数 |
| manager/databasemgr/sqlite_impl.go | 修改 `NewDatabaseManagerSQLiteImpl` 接收 loggerMgr 和 telemetryMgr 参数 |
| manager/databasemgr/none_impl.go | 无需修改 |
| manager/databasemgr/factory.go | 1. `Build` 函数新增 loggerMgr 和 telemetryMgr 参数<br>2. `BuildWithConfigProvider` 函数新增 loggerMgr 和 telemetryMgr 参数<br>3. 修改内部调用，传入依赖<br>4. 更新注释 |

| 文件 | 修改内容 |
|------|---------|
| manager/lockmgr/impl_base.go | 1. 移除字段 `inject:""` 标签，字段类型改为 `loggerMgr loggermgr.ILoggerManager`<br>2. 修改 `newILockManagerBaseImpl` 接收 loggerMgr、telemetryMgr 和 cacheMgr 参数 |
| manager/lockmgr/redis_impl.go | 修改 `NewLockManagerRedisImpl` 接收 loggerMgr、telemetryMgr 和 cacheMgr 参数 |
| manager/lockmgr/memory_impl.go | 修改 `NewLockManagerMemoryImpl` 接收 loggerMgr 和 telemetryMgr 参数 |
| manager/lockmgr/factory.go | 1. `Build` 函数新增 loggerMgr、telemetryMgr 和 cacheMgr 参数<br>2. `BuildWithConfigProvider` 函数新增 loggerMgr、telemetryMgr 和 cacheMgr 参数<br>3. 修改内部调用，传入依赖<br>4. 更新注释 |

| 文件 | 修改内容 |
|------|---------|
| manager/limitermgr/impl_base.go | 1. 移除字段 `inject:""` 标签，字段类型改为 `loggerMgr loggermgr.ILoggerManager`<br>2. 修改 `newILimiterManagerBaseImpl` 接收 loggerMgr、telemetryMgr 和 cacheMgr 参数 |
| manager/limitermgr/redis_impl.go | 1. 修改 `NewLimiterManagerRedisImpl` 接收 loggerMgr、telemetryMgr 和 cacheMgr 参数<br>2. 移除字段 `CacheMgr` 及其 inject 标签（已通过构造函数传入） |
| manager/limitermgr/memory_impl.go | 修改 `NewLimiterManagerMemoryImpl` 接收 loggerMgr 和 telemetryMgr 参数 |
| manager/limitermgr/factory.go | 1. `Build` 函数新增 loggerMgr、telemetryMgr 和 cacheMgr 参数<br>2. `BuildWithConfigProvider` 函数新增 loggerMgr、telemetryMgr 和 cacheMgr 参数<br>3. 修改内部调用，传入依赖<br>4. 更新注释 |

| 文件 | 修改内容 |
|------|---------|
| manager/mqmgr/impl_base.go | 1. 移除字段 `inject:""` 标签，字段类型改为 `loggerMgr loggermgr.ILoggerManager`<br>2. 修改 `newIMQManagerBaseImpl` 接收 loggerMgr 和 telemetryMgr 参数 |
| manager/mqmgr/rabbitmq_impl.go | 修改 `NewMQManagerRabbitMQImpl` 接收 loggerMgr 和 telemetryMgr 参数 |
| manager/mqmgr/memory_impl.go | 修改 `NewMQManagerMemoryImpl` 接收 loggerMgr 和 telemetryMgr 参数 |
| manager/mqmgr/factory.go | 1. `Build` 函数新增 loggerMgr 和 telemetryMgr 参数<br>2. `BuildWithConfigProvider` 函数新增 loggerMgr 和 telemetryMgr 参数<br>3. 修改内部调用，传入依赖<br>4. 更新注释 |

| 文件 | 修改内容 |
|------|---------|
| manager/schedulermgr/cron_impl.go | 修改 `NewSchedulerManagerCronImpl` 接收 loggerMgr 参数 |
| manager/schedulermgr/factory.go | 修改 `Build` 函数接收 loggerMgr 参数 |

### 4.2 Server 层文件

| 文件 | 修改内容 |
|------|---------|
| server/builtin.go | 1. 初始化 loggerMgr 和 telemetryMgr<br>2. 创建 databaseMgr、cacheMgr 时传入 loggerManager 和 telemetryMgr<br>3. 创建 lockMgr、limiterMgr 时传入 loggerManager、telemetryMgr 和 cacheMgr<br>4. 创建 mqMgr 时传入 loggerManager 和 telemetryMgr<br>5. 创建 schedulerMgr 时传入 loggerManager |

### 4.3 文档文件

| 文件 | 修改内容 |
|------|---------|
| AGENTS.md | 确认架构文档正确性（无需修改，已说明 Manager 不支持依赖注入） |
| manager/README.md | 更新 Manager 依赖注入说明，明确通过 Factory 传入而非容器注入 |
| 各 manager/README.md | 移除 "通过 `inject:"` 标签注入" 的错误说明 |

## 5. 验证标准

### 5.1 功能验证

修复完成后，验证以下功能正常工作：

1. **数据库 Manager**
   - [ ] 慢查询检测触发 WARN 日志
   - [ ] 指标 `db.query.duration` 正常记录
   - [ ] 指标 `db.query.slow_count` 正常记录
   - [ ] 链路追踪 span 正常创建

2. **缓存 Manager**
   - [ ] 操作日志正常输出
   - [ ] 指标 `cache.hit` 和 `cache.miss` 正常记录
   - [ ] 指标 `cache.operation.duration` 正常记录

3. **锁 Manager**
   - [ ] Redis 锁正常工作
   - [ ] 操作日志正常输出
   - [ ] 指标正常记录

4. **限流 Manager**
   - [ ] Redis 限流器正常工作（不再报 "cache manager is not initialized"）
   - [ ] 操作日志正常输出
   - [ ] 指标正常记录

5. **消息队列 Manager**
   - [ ] 操作日志正常输出
   - [ ] 指标正常记录

### 5.2 代码质量验证

1. **无 inject 标签残留**
   ```bash
   grep -r 'inject:""' manager/
   # 应该没有任何结果
   ```

2. **构建成功**
   ```bash
   go build -o litecore ./...
   ```

3. **测试通过**
   ```bash
   go test ./...
   ```

4. **格式检查**
   ```bash
   go fmt ./...
   ```

5. **静态检查**
   ```bash
   go vet ./...
   ```

## 6. 风险评估

| 风险 | 影响 | 缓解措施 |
|------|------|---------|
| Factory 方法签名变更，可能影响外部调用者 | 高 | 确认无外部调用者，或提供兼容层 |
| 依赖顺序变更，可能影响启动流程 | 中 | 严格按照依赖顺序初始化 |
| 某些 Manager（如 none）不需要依赖 | 低 | 接收参数后可以忽略（允许传 nil） |
| 可观测性大量日志输出 | 低 | 通过配置控制日志级别 |

## 7. 预期收益

修复完成后，预期收益：

1. **可观测性功能完全恢复** - 日志、指标、链路追踪正常工作
2. **跨 Manager 依赖正常工作** - Redis 限流器、Redis 锁等功能正常
3. **代码一致性提升** - 移除无效的 inject 标签，减少困惑
4. **架构清晰度提升** - 依赖关系通过构造函数显式声明，一目了然

## 8. 日志使用方式

### 8.1 Manager 层日志使用

Manager 层注入 `ILoggerManager` 后，通过 `LoggerMgr.Ins()` 获取 Logger 实例：

```go
type cacheManagerBaseImpl struct {
    loggerMgr    loggermgr.ILoggerManager
    telemetryMgr telemetrymgr.ITelemetryManager
}

func (c *cacheManagerRedisImpl) Set(ctx context.Context, key string, value any, ttl time.Duration) error {
    c.loggerMgr.Ins().Debug("开始设置缓存", "key", key, "ttl", ttl)
    // ...
    c.loggerMgr.Ins().Info("缓存设置成功", "key", key)
    return nil
}
```

### 8.2 与其他层保持一致

Service、Controller、Middleware、Listener、Scheduler 等层均使用 `LoggerMgr.Ins()` 方式：

```go
// Service 层
type MessageService struct {
    LoggerMgr loggermgr.ILoggerManager `inject:""`
}

func (s *MessageService) CreateMessage(...) {
    s.LoggerMgr.Ins().Info("Message created", "id", message.ID)
}
```

## 9. 参考资料

- AGENTS.md - 项目架构和编码规范
- container/manager_container.go - Manager 容器实现
- server/builtin.go - Manager 初始化逻辑
- server/engine.go:autoInject - 依赖注入逻辑（不包含 Manager）
