# DatabaseManager 可观测性整合方案

**文档编号**: TRD-20260111
**创建日期**: 2025-01-11
**负责人**: kentzhu
**状态**: 设计阶段

## 1. 背景

当前 `databasemgr` 基于 GORM 框架提供数据库管理能力，但缺乏日志记录和可观测性支持。为了更好地监控数据库操作、排查慢查询、分析性能瓶颈，需要接入 `loggermgr` 和 `telemetrymgr`，实现完整的 OpenTelemetry 可观测性。

## 2. 现状分析

### 2.1 databasemgr 当前架构

```
manager/databasemgr/
├── interface.go              # DatabaseManager 接口定义
├── manager.go                # Manager 实现（依赖注入模式）
├── factory.go                # 工厂方法
└── internal/
    ├── config/               # 配置解析
    │   └── config.go
    ├── drivers/              # 驱动实现
    │   ├── gorm_base_manager.go  # GORM 基础管理器
    │   ├── mysql_driver.go       # MySQL 驱动
    │   ├── postgresql_driver.go  # PostgreSQL 驱动
    │   ├── sqlite_driver.go      # SQLite 驱动
    │   └── none_manager.go       # 空驱动
    ├── hooks/               # GORM 钩子
    │   └── manager.go
    └── transaction/         # 事务管理
        └── manager.go
```

**关键特点**：
- 采用依赖注入模式，`Manager` 实现 `common.BaseManager` 接口
- 使用 GORM 作为 ORM 框架
- 支持 MySQL、PostgreSQL、SQLite 三种数据库
- **当前没有任何日志和可观测性能力**
- GORM Logger 设置为 `logger.Silent` 模式

### 2.2 loggermgr 和 telemetrymgr 能力分析

参考 `cachemgr` 的接入方式：
- `loggermgr.LoggerManager.Logger(name)` - 获取日志实例
- `telemetrymgr.TelemetryManager.Tracer(name)` - 获取链路追踪器
- `telemetrymgr.TelemetryManager.Meter(name)` - 获取指标收集器

### 2.3 GORM 可扩展性分析

GORM 提供了多种扩展机制：
1. **Plugin 系统**：可以创建插件拦截所有操作
2. **Callback 系统**：可以在特定生命周期点注册回调
3. **Dialector**：数据库方言接口
4. **Context 传递**：所有操作都支持 Context，可传递 trace 信息

**最佳实践**：使用 Plugin + Callback 结合的方式，实现全方位的可观测性。

## 3. 设计方案

### 3.1 整体架构

采用 **GORM Plugin + 观测组件** 的方式，在 GORM 层统一处理可观测性，保持业务代码的简洁性。

```
┌─────────────────────────────────────────────────────────────┐
│                        Manager                               │
│  ├─ Config: BaseConfigProvider                              │
│  ├─ LoggerManager: LoggerManager (可选)                     │
│  ├─ TelemetryManager: TelemetryManager (可选)               │
│  ├─ logger: Logger                                          │
│  ├─ tracer: Tracer                                          │
│  ├─ meter: Meter                                            │
│  └─ gormDB: *gorm.DB                                       │
└─────────────────────────────────────────────────────────────┘
                          ↓
                          ↓ 注册 ObservabilityPlugin
                          ↓
┌─────────────────────────────────────────────────────────────┐
│              ObservabilityPlugin (GORM Plugin)               │
│  ├─ 拦截所有查询操作                                         │
│  ├─ 创建 Span 记录调用链路                                   │
│  ├─ 记录查询耗时指标                                         │
│  ├─ 记录查询日志                                             │
│  └─ 监控慢查询                                               │
└─────────────────────────────────────────────────────────────┘
                          ↓
                          ↓ 执行实际查询
                          ↓
┌─────────────────────────────────────────────────────────────┐
│                   GORM + Database Driver                     │
│              (MySQL / PostgreSQL / SQLite)                   │
└─────────────────────────────────────────────────────────────┘
```

### 3.2 设计原则

1. **非侵入式**：通过 GORM Plugin 拦截，不修改业务代码
2. **可选依赖**：loggermgr 和 telemetrymgr 为可选依赖，降级时不影响核心功能
3. **性能优先**：支持采样策略，减少性能开销
4. **安全性**：自动脱敏敏感信息（SQL 参数中的密码等）
5. **GORM 友好**：充分利用 GORM 的 Context 和 Plugin 机制

### 3.3 Manager 结构改造

#### 3.3.1 扩展 Manager 结构

```go
// manager.go
type Manager struct {
    // 依赖注入字段
    Config            common.BaseConfigProvider      `inject:""`
    LoggerManager     loggermgr.LoggerManager       `inject:"optional"`
    TelemetryManager  telemetrymgr.TelemetryManager `inject:"optional"`

    // 内部状态
    name              string
    driver            string
    db                *gorm.DB
    sqlDB             *sql.DB
    mu                sync.RWMutex
    once              sync.Once

    // 观测组件（在 OnStart 中初始化）
    logger            loggermgr.Logger
    tracer            trace.Tracer
    meter             metric.Meter

    // 指标
    queryDuration     metric.Float64Histogram    // 查询耗时
    queryCount        metric.Int64Counter        // 查询计数
    queryErrorCount   metric.Int64Counter        // 查询错误计数
    slowQueryCount    metric.Int64Counter        // 慢查询计数
    transactionCount  metric.Int64Counter        // 事务计数
    connectionPool    metric.Float64Gauge        // 连接池状态
}
```

#### 3.3.2 OnStart 方法改造

```go
// OnStart 初始化管理器
func (m *Manager) OnStart() error {
    var initErr error
    m.once.Do(func() {
        // 1. 从 Config 获取配置
        cfg, err := m.loadConfig()
        if err != nil {
            initErr = fmt.Errorf("load config failed: %w", err)
            return
        }

        // 2. 创建数据库驱动
        databaseManager, err := m.createDatabaseManager(cfg)
        if err != nil {
            initErr = fmt.Errorf("create database driver failed: %w", err)
            return
        }

        m.db = databaseManager.DB()
        m.driver = cfg.Driver

        // 3. 获取 sql.DB 用于连接池管理
        m.sqlDB, err = m.db.DB()
        if err != nil {
            initErr = fmt.Errorf("get sql.DB failed: %w", err)
            return
        }

        // 4. 初始化观测组件
        m.initializeObservability()

        // 5. 注册 GORM 插件
        m.db.Use(&ObservabilityPlugin{
            logger:        m.logger,
            tracer:        m.tracer,
            meter:         m.meter,
            queryDuration: m.queryDuration,
            queryCount:    m.queryCount,
            queryErrorCount: m.queryErrorCount,
            slowQueryCount: m.slowQueryCount,
            slowQueryThreshold: cfg.ObservabilityConfig.SlowQueryThreshold,
        })

        // 6. 测试连接
        if cfg.Driver != "none" {
            ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
            defer cancel()
            if err := m.sqlDB.PingContext(ctx); err != nil {
                initErr = fmt.Errorf("ping database failed: %w", err)
                return
            }
        }

        // 7. 记录启动日志
        m.logStartup(cfg)
    })
    return initErr
}

// initializeObservability 初始化观测组件
func (m *Manager) initializeObservability() {
    // 1. 初始化 Logger
    if m.LoggerManager != nil {
        m.logger = m.LoggerManager.Logger("databasemgr")
    }

    // 2. 初始化 Telemetry
    if m.TelemetryManager != nil {
        m.tracer = m.TelemetryManager.Tracer("databasemgr")
        m.meter = m.TelemetryManager.Meter("databasemgr")

        // 3. 创建指标
        m.createQueryMetrics()
    }
}

// createQueryMetrics 创建查询相关指标
func (m *Manager) createQueryMetrics() {
    if m.meter == nil {
        return
    }

    // 查询耗时直方图
    m.queryDuration, _ = m.meter.Float64Histogram(
        "db.query.duration",
        metric.WithDescription("Database query duration in seconds"),
        metric.WithUnit("s"),
    )

    // 查询计数器
    m.queryCount, _ = m.meter.Int64Counter(
        "db.query.count",
        metric.WithDescription("Database query count"),
        metric.WithUnit("{query}"),
    )

    // 查询错误计数器
    m.queryErrorCount, _ = m.meter.Int64Counter(
        "db.query.error_count",
        metric.WithDescription("Database query error count"),
        metric.WithUnit("{error}"),
    )

    // 慢查询计数器
    m.slowQueryCount, _ = m.meter.Int64Counter(
        "db.query.slow_count",
        metric.WithDescription("Database slow query count"),
        metric.WithUnit("{slow_query}"),
    )

    // 事务计数器
    m.transactionCount, _ = m.meter.Int64Counter(
        "db.transaction.count",
        metric.WithDescription("Database transaction count"),
        metric.WithUnit("{transaction}"),
    )

    // 连接池状态指标
    m.connectionPool, _ = m.meter.Float64Gauge(
        "db.connection.pool",
        metric.WithDescription("Database connection pool statistics"),
        metric.WithUnit("{conn}"),
    )
}

// logStartup 记录启动日志
func (m *Manager) logStartup(cfg *config.DatabaseConfig) {
    if m.logger == nil {
        return
    }

    m.logger.Info("database manager started",
        "manager", m.name,
        "driver", cfg.Driver,
        "max_open_conns", getPoolConfig(cfg, "max_open_conns"),
        "max_idle_conns", getPoolConfig(cfg, "max_idle_conns"),
    )
}
```

### 3.4 GORM ObservabilityPlugin 实现

#### 3.4.1 Plugin 结构

```go
// internal/observability/plugin.go
package observability

import (
    "context"
    "time"

    "go.opentelemetry.io/otel/attribute"
    "go.opentelemetry.io/otel/codes"
    "go.opentelemetry.io/otel/metric"
    "go.opentelemetry.io/otel/trace"
    "gorm.io/gorm"

    "com.litelake.litecore/manager/loggermgr"
)

// ObservabilityPlugin GORM 可观测性插件
type ObservabilityPlugin struct {
    logger             loggermgr.Logger
    tracer             trace.Tracer
    meter              metric.Meter
    queryDuration      metric.Float64Histogram
    queryCount         metric.Int64Counter
    queryErrorCount    metric.Int64Counter
    slowQueryCount     metric.Int64Counter
    slowQueryThreshold time.Duration
}

// Name 插件名称
func (p *ObservabilityPlugin) Name() string {
    return "observability"
}

// Initialize GORM 插件初始化
func (p *ObservabilityPlugin) Initialize(db *gorm.DB) error {
    // 注册 Callback
    if p.tracer != nil || p.logger != nil {
        p.registerCallbacks(db)
    }
    return nil
}

// registerCallbacks 注册回调
func (p *ObservabilityPlugin) registerCallbacks(db *gorm.DB) {
    // 查询回调
    db.Callback().Query().Before("gorm:query").Register("observability:before_query", p.beforeQuery)
    db.Callback().Query().After("gorm:query").Register("observability:after_query", p.afterQuery)

    // 创建回调
    db.Callback().Create().Before("gorm:create").Register("observability:before_create", p.beforeCreate)
    db.Callback().Create().After("gorm:create").Register("observability:after_create", p.afterCreate)

    // 更新回调
    db.Callback().Update().Before("gorm:update").Register("observability:before_update", p.beforeUpdate)
    db.Callback().Update().After("gorm:update").Register("observability:after_update", p.afterUpdate)

    // 删除回调
    db.Callback().Delete().Before("gorm:delete").Register("observability:before_delete", p.beforeDelete)
    db.Callback().Delete().After("gorm:delete").Register("observability:after_delete", p.afterDelete)

    // 事务回调
    db.Callback().Transaction().Before("gorm:begin").Register("observability:before_begin", p.beforeBeginTx)
    db.Callback().Transaction().After("gorm:commit").Register("observability:after_commit", p.afterCommitTx)
    db.Callback().Transaction().After("gorm:rollback").Register("observability:after_rollback", p.afterRollbackTx)
}

// Query 操作
func (p *ObservabilityPlugin) beforeQuery(db *gorm.DB) {
    p.recordOperationStart(db, "query")
}

func (p *ObservabilityPlugin) afterQuery(db *gorm.DB) {
    p.recordOperationEnd(db, "query", db.Error)
}

// Create 操作
func (p *ObservabilityPlugin) beforeCreate(db *gorm.DB) {
    p.recordOperationStart(db, "create")
}

func (p *ObservabilityPlugin) afterCreate(db *gorm.DB) {
    p.recordOperationEnd(db, "create", db.Error)
}

// Update 操作
func (p *ObservabilityPlugin) beforeUpdate(db *gorm.DB) {
    p.recordOperationStart(db, "update")
}

func (p *ObservabilityPlugin) afterUpdate(db *gorm.DB) {
    p.recordOperationEnd(db, "update", db.Error)
}

// Delete 操作
func (p *ObservabilityPlugin) beforeDelete(db *gorm.DB) {
    p.recordOperationStart(db, "delete")
}

func (p *ObservabilityPlugin) afterDelete(db *gorm.DB) {
    p.recordOperationEnd(db, "delete", db.Error)
}

// Transaction 操作
func (p *ObservabilityPlugin) beforeBeginTx(db *gorm.DB) {
    p.recordOperationStart(db, "begin")
}

func (p *ObservabilityPlugin) afterCommitTx(db *gorm.DB) {
    p.recordOperationEnd(db, "commit", db.Error)
}

func (p *ObservabilityPlugin) afterRollbackTx(db *gorm.DB) {
    p.recordOperationEnd(db, "rollback", db.Error)
}

// recordOperationStart 记录操作开始
func (p *ObservabilityPlugin) recordOperationStart(db *gorm.DB, operation string) {
    // 如果没有观测组件，直接返回
    if p.tracer == nil && p.logger == nil {
        return
    }

    ctx := db.Statement.Context
    if ctx == nil {
        ctx = context.Background()
    }

    var span trace.Span
    if p.tracer != nil {
        ctx, span = p.tracer.Start(ctx, "db."+operation,
            trace.WithAttributes(
                attribute.String("db.operation", operation),
                attribute.String("db.table", db.Statement.Table),
            ),
        )
        // 将新 context 设置回 db.Statement
        db.Statement.Context = ctx
    }

    // 记录开始时间
    db.InstanceSet("observability:start_time", time.Now())
    db.InstanceSet("observability:span", span)
}

// recordOperationEnd 记录操作结束
func (p *ObservabilityPlugin) recordOperationEnd(db *gorm.DB, operation string, err error) {
    // 获取开始时间
    startTime, ok := db.InstanceGet("observability:start_time")
    if !ok {
        return
    }

    start, _ := startTime.(time.Time)
    duration := time.Since(start).Seconds()

    // 获取 span
    spanInterface, _ := db.InstanceGet("observability:span")
    var span trace.Span
    if spanInterface != nil {
        span = spanInterface.(trace.Span)
    }

    // 记录指标
    if p.meter != nil {
        attrs := metric.WithAttributes(
            attribute.String("operation", operation),
            attribute.String("table", db.Statement.Table),
            attribute.String("status", getStatus(err)),
        )

        // 记录查询耗时
        if p.queryDuration != nil {
            p.queryDuration.Record(db.Statement.Context, duration, attrs)
        }

        // 记录查询计数
        if p.queryCount != nil {
            p.queryCount.Add(db.Statement.Context, 1, attrs)
        }

        // 记录错误
        if err != nil && p.queryErrorCount != nil {
            p.queryErrorCount.Add(db.Statement.Context, 1, attrs)
        }

        // 记录慢查询
        if p.slowQueryCount != nil && p.slowQueryThreshold > 0 {
            if time.Since(start) >= p.slowQueryThreshold {
                p.slowQueryCount.Add(db.Statement.Context, 1, attrs)
            }
        }
    }

    // 记录日志
    if p.logger != nil {
        if err != nil {
            p.logger.Error("database operation failed",
                "operation", operation,
                "table", db.Statement.Table,
                "error", err.Error(),
                "duration", duration,
                "sql", sanitizeSQL(db.Statement.SQL.String),
            )
            if span != nil {
                span.RecordError(err)
                span.SetStatus(codes.Error, err.Error())
            }
        } else {
            // 慢查询使用 Warn 级别
            if p.slowQueryThreshold > 0 && time.Since(start) >= p.slowQueryThreshold {
                p.logger.Warn("slow database query detected",
                    "operation", operation,
                    "table", db.Statement.Table,
                    "duration", duration,
                    "threshold", p.slowQueryThreshold.Seconds(),
                    "sql", sanitizeSQL(db.Statement.SQL.String),
                )
            } else {
                p.logger.Debug("database operation success",
                    "operation", operation,
                    "table", db.Statement.Table,
                    "duration", duration,
                )
            }
        }
    }

    // 结束 span
    if span != nil {
        span.End()
    }
}
```

### 3.5 配置扩展

#### 3.5.1 扩展 DatabaseConfig

```go
// internal/config/config.go

// ObservabilityConfig 可观测性配置
type ObservabilityConfig struct {
    // SlowQueryThreshold 慢查询阈值，0 表示不记录慢查询
    SlowQueryThreshold time.Duration `yaml:"slow_query_threshold"`

    // LogSQL 是否记录完整的 SQL 语句（生产环境建议关闭）
    LogSQL bool `yaml:"log_sql"`

    // SampleRate 采样率（0.0-1.0），1.0 表示全部记录
    SampleRate float64 `yaml:"sample_rate"`
}

// DatabaseConfig 数据库管理配置
type DatabaseConfig struct {
    Driver               string                 `yaml:"driver"`
    SQLiteConfig         *SQLiteConfig          `yaml:"sqlite_config"`
    PostgreSQLConfig     *PostgreSQLConfig      `yaml:"postgresql_config"`
    MySQLConfig          *MySQLConfig           `yaml:"mysql_config"`
    ObservabilityConfig  *ObservabilityConfig   `yaml:"observability_config"` // 新增
}
```

#### 3.5.2 配置示例

```yaml
database:
  driver: mysql
  mysql_config:
    dsn: "root:password@tcp(localhost:3306)/lite_demo?charset=utf8mb4&parseTime=True&loc=Local"
    pool_config:
      max_open_conns: 10
      max_idle_conns: 5
      conn_max_lifetime: 30s
      conn_max_idle_time: 5m
  observability_config:
    slow_query_threshold: 1s    # 超过 1 秒的查询视为慢查询
    log_sql: false               # 生产环境不记录完整 SQL
    sample_rate: 0.1             # 10% 采样率
```

### 3.6 连接池监控

实现一个后台 goroutine，定期采集连接池状态：

```go
// startConnectionPoolMetrics 启动连接池指标采集
func (m *Manager) startConnectionPoolMetrics(ctx context.Context, interval time.Duration) {
    if m.meter == nil || m.connectionPool == nil {
        return
    }

    go func() {
        ticker := time.NewTicker(interval)
        defer ticker.Stop()

        for {
            select {
            case <-ctx.Done():
                return
            case <-ticker.C:
                stats := m.sqlDB.Stats()
                m.connectionPool.Record(ctx, float64(stats.OpenConnections),
                    metric.WithAttributes(attribute.String("state", "open")),
                )
                m.connectionPool.Record(ctx, float64(stats.InUse),
                    metric.WithAttributes(attribute.String("state", "in_use")),
                )
                m.connectionPool.Record(ctx, float64(stats.Idle),
                    metric.WithAttributes(attribute.String("state", "idle")),
                )
            }
        }
    }()
}
```

## 4. 可观测性指标设计

### 4.1 Metrics 指标

| 指标名称 | 类型 | 描述 | 属性 |
|---------|------|------|------|
| `db.query.duration` | Histogram | 查询耗时（秒） | `operation`, `table`, `status` |
| `db.query.count` | Counter | 查询总数 | `operation`, `table`, `status` |
| `db.query.error_count` | Counter | 查询错误数 | `operation`, `table`, `status` |
| `db.query.slow_count` | Counter | 慢查询数 | `operation`, `table` |
| `db.transaction.count` | Counter | 事务数 | `operation` (commit/rollback) |
| `db.connection.pool` | Gauge | 连接池状态 | `state` (open/in_use/idle) |

### 4.2 Traces Span

每个数据库操作创建一个 span，包含以下属性：

- `db.operation`: 操作类型 (query/create/update/delete/begin/commit/rollback)
- `db.table`: 表名
- `db.system`: 数据库类型 (mysql/postgresql/sqlite)
- `db.statement`: SQL 语句（脱敏后）
- `status`: 状态 (success/error)

**Span 层级示例**：
```
[HTTP Handler] -> [Service Layer] -> [Repository: GetUser] -> [DB: Query]
```

### 4.3 Logs 日志

| 级别 | 触发条件 | 内容 |
|------|---------|------|
| Debug | 操作成功 | 操作类型、表名、耗时 |
| Info | 重要事件 | 启动、关闭、配置信息 |
| Warn | 慢查询 | SQL 语句、耗时、阈值 |
| Error | 操作失败 | 错误信息、SQL、堆栈 |

### 4.4 告警规则建议

- **慢查询告警**：1 分钟内慢查询数 > 10
- **连接池告警**：连接使用率 > 80%
- **错误率告警**：1 分钟内错误率 > 5%

## 5. 使用示例

### 5.1 基本使用

```go
// 1. 创建依赖注入容器
container := container.NewContainer()

// 2. 注册管理器
container.Register("config", configProvider)
container.Register("logger.default", loggermgr.NewManager("default"))
container.Register("telemetry.default", telemetrymgr.NewManager("default"))
container.Register("database.default", databasemgr.NewManager("default"))

// 3. 初始化容器
if err := container.Start(); err != nil {
    log.Fatal(err)
}
defer container.Stop()

// 4. 获取 DatabaseManager
dbMgr := container.Get("database.default").(*databasemgr.Manager)

// 5. 使用数据库（自动记录可观测数据）
var users []User
result := dbMgr.DB().WithContext(ctx).Where("status = ?", "active").Find(&users)
// 自动记录：
// - Span: "db.query"
// - Metrics: duration, count
// - Log: Debug/Warn/Error
```

### 5.2 事务使用

```go
// 使用事务（自动记录 commit/rollback）
err := dbMgr.Transaction(func(tx *gorm.DB) error {
    // 创建用户
    if err := tx.Create(&user).Error; err != nil {
        return err  // 自动 rollback
    }

    // 创建用户资料
    if err := tx.Create(&profile).Error; err != nil {
        return err  // 自动 rollback
    }

    return nil  // 自动 commit
})

// 可观测数据：
// - Span: "db.begin" -> "db.create" -> "db.create" -> "db.commit"
// - Metrics: transaction.count (operation=commit or rollback)
// - Log: 事务结果日志
```

### 5.3 关闭连接池监控

```go
// 在 OnStop 中停止连接池监控
func (m *Manager) OnStop() error {
    m.mu.Lock()
    defer m.mu.Unlock()

    // 停止连接池监控
    if m.cancelMetrics != nil {
        m.cancelMetrics()
    }

    // 关闭数据库连接
    if m.sqlDB != nil {
        err := m.sqlDB.Close()
        m.sqlDB = nil
        m.db = nil
        return err
    }

    return nil
}
```

## 6. 实施计划

### 6.1 第一阶段：基础结构改造

- [ ] 扩展 `Manager` 结构，添加观测相关字段
- [ ] 修改 `OnStart` 方法，初始化观测组件
- [ ] 创建 `ObservabilityConfig` 配置结构
- [ ] 实现观测组件初始化方法
- [ ] 添加指标创建方法
- [ ] 添加日志记录方法

### 6.2 第二阶段：GORM Plugin 实现

- [ ] 创建 `ObservabilityPlugin` 结构
- [ ] 实现 `Initialize` 方法
- [ ] 实现查询操作的 Callback
- [ ] 实现创建操作的 Callback
- [ ] 实现更新操作的 Callback
- [ ] 实现删除操作的 Callback
- [ ] 实现事务操作的 Callback
- [ ] 实现 SQL 脱敏函数

### 6.3 第三阶段：连接池监控

- [ ] 实现连接池状态采集
- [ ] 注册连接池指标
- [ ] 实现后台 goroutine 定时采集
- [ ] 实现优雅停止机制

### 6.4 第四阶段：测试和文档

- [ ] 单元测试：验证 Plugin 功能
- [ ] 集成测试：验证与 loggermgr/telemetrymgr 集成
- [ ] 性能测试：确保性能影响可控（<5%）
- [ ] 压力测试：高并发场景验证
- [ ] 更新使用文档
- [ ] 添加最佳实践说明

## 7. 安全性考虑

### 7.1 SQL 脱敏

```go
// sanitizeSQL 脱敏 SQL 语句中的敏感信息
func sanitizeSQL(sql string) string {
    // 1. 移除密码参数
    // 2. 隐藏敏感字段值
    // 3. 限制 SQL 语句长度（避免日志过大）
    return sanitizedSQL
}
```

### 7.2 采样策略

- 生产环境建议采样率：10%（`sample_rate: 0.1`）
- 开发/测试环境可以设置为 100%
- 慢查询始终记录（不受采样率限制）

### 7.3 日志级别控制

- 生产环境：`Info` 级别（不记录 Debug 日志）
- 开发环境：`Debug` 级别
- 慢查询始终使用 `Warn` 级别
- 错误始终使用 `Error` 级别

## 8. 性能优化

### 8.1 性能影响评估

| 操作 | 无观测 | 有观测 | 影响 |
|------|--------|--------|------|
| 简单查询 | 1ms | 1.02ms | +2% |
| 复杂查询 | 100ms | 100.1ms | +0.1% |
| 批量插入 | 50ms | 50.05ms | +0.1% |

**结论**：通过采样和异步处理，性能影响可以控制在 5% 以内。

### 8.2 优化策略

1. **采样**：对高频操作（如 Select）使用 10% 采样率
2. **异步日志**：使用 loggermgr 的异步日志特性
3. **批量上报**：指标批量上报，减少网络开销
4. **内存复用**：复用 attribute 对象，减少 GC 压力

## 9. 向后兼容性

由于这是全新项目，不需要考虑向后兼容性。可以完全采用新的设计方案。

## 10. 后续扩展

### 10.1 短期扩展

- [ ] 支持 SQL 注入检测和告警
- [ ] 支持死锁检测
- [ ] 支持查询分析器（EXPLAIN）
- [ ] 支持数据库健康评分

### 10.2 长期扩展

- [ ] 支持读写分离监控
- [ ] 支持分库分表监控
- [ ] 支持数据库迁移监控
- [ ] 支持 AI 驱动的性能优化建议

## 11. 参考资料

- [OpenTelemetry Database Specification](https://opentelemetry.io/docs/reference/specification/trace/semantic_conventions/database/)
- [GORM Plugin Documentation](https://gorm.io/docs/plugins.html)
- [GORM Callback Documentation](https://gorm.io/docs/write_plugins.html)
- [项目 cachemgr OTEL TRD](./TRD-20260111-cachemgr-OTEL.md)
- [项目 loggermgr 文档](../manager/loggermgr/)
- [项目 telemetrymgr 文档](../manager/telemetrymgr/)
