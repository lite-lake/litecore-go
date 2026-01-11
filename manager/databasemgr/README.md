# Database Manager

数据库管理器，支持 MySQL、PostgreSQL 和 SQLite，基于 GORM 实现。

## 特性

- **多数据库支持** - 支持 MySQL、PostgreSQL、SQLite，以及无数据库模式
- **工厂模式创建** - 通过 Factory 统一创建不同驱动的管理器
- **连接池管理** - 支持连接池配置和状态监控
- **事务支持** - 完整的事务管理和自动迁移功能
- **生命周期管理** - 集成服务启停接口，支持健康检查
- **可观测性** - 集成 OpenTelemetry，支持链路追踪、指标和日志

## 快速开始

```go
package main

import (
    "log"

    "com.litelake.litecore/manager/databasemgr"
    "com.litelake.litecore/manager/databasemgr/internal/config"
)

func main() {
    // 创建工厂
    factory := databasemgr.NewFactory()

    // 配置 SQLite 内存数据库
    cfg := &config.DatabaseConfig{
        Driver: "sqlite",
        SQLiteConfig: &config.SQLiteConfig{
            DSN: ":memory:",
        },
    }

    // 创建管理器
    mgr, err := factory.BuildWithConfig(cfg)
    if err != nil {
        log.Fatal(err)
    }
    defer mgr.Close()

    // 使用 GORM 操作数据库
    db := mgr.DB()
    db.AutoMigrate(&User{})

    db.Create(&User{Name: "Alice", Age: 30})
    db.Create(&User{Name: "Bob", Age: 25})

    var users []User
    db.Find(&users)
}

type User struct {
    ID   uint
    Name string
    Age  int
}
```

## 创建管理器

### 使用配置结构体（推荐）

```go
cfg := &config.DatabaseConfig{
    Driver: "mysql",
    MySQLConfig: &config.MySQLConfig{
        DSN: "root:password@tcp(localhost:3306)/mydb?charset=utf8mb4&parseTime=True&loc=Local",
        PoolConfig: &config.PoolConfig{
            MaxOpenConns:    20,
            MaxIdleConns:    10,
            ConnMaxLifetime: 30 * time.Second,
            ConnMaxIdleTime: 5 * time.Minute,
        },
    },
}

mgr, err := factory.BuildWithConfig(cfg)
if err != nil {
    log.Fatal(err)
}
```

### 使用 map 配置

```go
cfg := map[string]any{
    "driver": "postgresql",
    "postgresql_config": map[string]any{
        "dsn": "host=localhost port=5432 user=postgres password=password dbname=mydb sslmode=disable",
        "pool_config": map[string]any{
            "max_open_conns": 20,
            "max_idle_conns": 10,
        },
    },
}

mgr := factory.Build("", cfg)
// 无效配置会返回 NoneDatabaseManager
if mgr.ManagerName() == "none-database" {
    log.Fatal("Failed to create database manager")
}
```

## GORM 操作

DatabaseManager 封装了完整的 GORM 功能，支持所有 GORM 操作：

```go
db := mgr.DB()

// 创建
db.Create(&User{Name: "Alice"})

// 查询
var user User
db.First(&user, 1)

var users []User
db.Where("age > ?", 18).Find(&users)

// 更新
db.Model(&user).Update("age", 31)

// 删除
db.Delete(&user)
```

也可以使用便捷方法：

```go
// 指定模型
mgr.Model(&User{}).Where("age > ?", 18).Find(&users)

// 指定表名
mgr.Table("users").Count(&count)

// 带上下文
ctx := context.Background()
mgr.WithContext(ctx).First(&user, 1)
```

## 事务管理

### 自动事务

```go
err := mgr.Transaction(func(tx *gorm.DB) error {
    if err := tx.Create(&User{Name: "Alice"}).Error; err != nil {
        return err // 返回错误会回滚
    }
    if err := tx.Create(&User{Name: "Bob"}).Error; err != nil {
        return err
    }
    return nil // 返回 nil 会提交
})
if err != nil {
    log.Fatal(err)
}
```

### 手动事务

```go
tx := mgr.Begin()
defer func() {
    if r := recover(); r != nil {
        tx.Rollback()
        panic(r)
    }
}()

if err := tx.Create(&User{Name: "Alice"}).Error; err != nil {
    tx.Rollback()
    log.Fatal(err)
}

if err := tx.Create(&User{Name: "Bob"}).Error; err != nil {
    tx.Rollback()
    log.Fatal(err)
}

tx.Commit()
```

## 迁移管理

```go
// 自动迁移
err := mgr.AutoMigrate(&User{}, &Product{}, &Order{})
if err != nil {
    log.Fatal(err)
}

// 使用迁移器进行高级操作
migrator := mgr.Migrator()

// 创建表
migrator.CreateTable(&User{})

// 添加列
migrator.AddColumn(&User{}, "Email", "string")

// 创建索引
migrator.CreateIndex(&User{}, "Name")
```

## 连接池管理

### 配置连接池

```go
cfg := &config.DatabaseConfig{
    Driver: "mysql",
    MySQLConfig: &config.MySQLConfig{
        DSN: "user:password@tcp(localhost:3306)/dbname",
        PoolConfig: &config.PoolConfig{
            MaxOpenConns:    20,             // 最大打开连接数
            MaxIdleConns:    10,             // 最大空闲连接数
            ConnMaxLifetime: 30 * time.Second, // 连接最大存活时间
            ConnMaxIdleTime: 5 * time.Minute,  // 连接最大空闲时间
        },
    },
}
```

### 监控连接池

```go
stats := mgr.Stats()
fmt.Printf("Open Connections: %d\n", stats.OpenConnections)
fmt.Printf("In Use: %d\n", stats.InUse)
fmt.Printf("Idle: %d\n", stats.Idle)
```

## 可观测性

Manager 支持完整的 OpenTelemetry 可观测性，包括链路追踪、指标和日志。

### 启用可观测性

通过依赖注入容器接入 loggermgr 和 telemetrymgr：

```go
import (
    "com.litelake.litecore/manager/databasemgr"
    "com.litelake.litecore/manager/loggermgr"
    "com.litelake.litecore/manager/telemetrymgr"
)

// 注册到容器
container.Register("config", configProvider)
container.Register("logger.default", loggermgr.NewManager("default"))
container.Register("telemetry.default", telemetrymgr.NewManager("default"))
container.Register("database.default", databasemgr.NewManager("default"))

// 启动容器
container.Start()
defer container.Stop()
```

### 可观测性配置

在数据库配置中添加可观测性选项：

```go
cfg := &config.DatabaseConfig{
    Driver: "mysql",
    MySQLConfig: &config.MySQLConfig{
        DSN: "user:password@tcp(localhost:3306)/dbname",
    },
    ObservabilityConfig: &config.ObservabilityConfig{
        SlowQueryThreshold: 1 * time.Second,  // 慢查询阈值
        LogSQL:             false,              // 是否记录完整 SQL
        SampleRate:         0.1,                // 10% 采样率
    },
}
```

### 配置选项

| 参数 | 类型 | 默认值 | 说明 |
|------|------|--------|------|
| `SlowQueryThreshold` | Duration | 1s | 慢查询阈值，超过此时长的查询会被标记 |
| `LogSQL` | bool | false | 是否在日志中记录完整 SQL（生产环境建议关闭） |
| `SampleRate` | float64 | 1.0 | 采样率（0.0-1.0），1.0 表示全采样 |

### 自动采集的指标

| 指标名称 | 类型 | 描述 | 属性 |
|---------|------|------|------|
| `db.query.duration` | Histogram | 查询耗时（秒） | operation, table, status |
| `db.query.count` | Counter | 查询总数 | operation, table, status |
| `db.query.error_count` | Counter | 查询错误数 | operation, table, status |
| `db.query.slow_count` | Counter | 慢查询数 | operation, table |
| `db.connection.pool` | Gauge | 连接池状态 | state (open/in_use/idle) |

### 日志级别

- **Debug** - 正常数据库操作（仅包含操作类型、表名、耗时）
- **Warn** - 慢查询告警
- **Error** - 数据库操作失败

### SQL 脱敏

当 `LogSQL=true` 时，插件会自动脱敏 SQL 语句中的敏感信息：
- 密码字段（password, pwd, token, secret, api_key）
- 限制 SQL 语句长度（最大 500 字符）

### 示例输出

```
[DEBUG] database operation success operation=query table=users duration=0.002
[WARN] slow database query detected operation=query table=orders duration=1.234 threshold=1.000
[ERROR] database operation failed operation=update table=users error="duplicate key" duration=0.005
```

## 健康检查

```go
// 简单检查
if err := mgr.Health(); err != nil {
    log.Printf("Database unhealthy: %v", err)
}

// 带超时的检查
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()
if err := mgr.Ping(ctx); err != nil {
    log.Printf("Ping failed: %v", err)
}
```

## 生命周期管理

DatabaseManager 实现了 `common.Manager` 接口，可以集成到服务生命周期中：

```go
// 启动时调用
if err := mgr.OnStart(); err != nil {
    log.Fatal(err)
}

// 停止时调用
if err := mgr.OnStop(); err != nil {
    log.Fatal(err)
}
```

## API

### 接口

| 接口 | 说明 |
|------|------|
| `DatabaseManager` | 数据库管理器核心接口 |

### 工厂方法

| 方法 | 说明 |
|------|------|
| `NewFactory()` | 创建工厂实例 |
| `Build(driver, cfg)` | 从 map 配置创建管理器（错误时返回 NoneDatabaseManager） |
| `BuildWithConfig(cfg)` | 从配置结构体创建管理器（返回详细错误） |

### 核心方法

| 分类 | 方法 | 说明 |
|------|------|------|
| 生命周期 | `ManagerName()` | 返回管理器名称 |
|  | `Health()` | 检查健康状态 |
|  | `OnStart()` | 启动初始化 |
|  | `OnStop()` | 停止清理 |
| GORM | `DB()` | 获取 GORM 实例 |
|  | `Model(value)` | 指定模型 |
|  | `Table(name)` | 指定表名 |
|  | `WithContext(ctx)` | 设置上下文 |
| 事务 | `Transaction(fn, opts)` | 执行事务 |
|  | `Begin(opts)` | 开启事务 |
| 迁移 | `AutoMigrate(models)` | 自动迁移 |
|  | `Migrator()` | 获取迁移器 |
| 连接 | `Driver()` | 获取驱动类型 |
|  | `Ping(ctx)` | 检查连接 |
|  | `Stats()` | 连接池状态 |
|  | `Close()` | 关闭连接 |
| SQL | `Exec(sql, values)` | 执行原生 SQL |
|  | `Raw(sql, values)` | 执行原生查询 |

## 配置

### 驱动类型

| 驱动 | 说明 |
|------|------|
| `mysql` | MySQL 数据库 |
| `postgresql` | PostgreSQL 数据库 |
| `sqlite` | SQLite 数据库 |
| `none` | 无数据库（空管理器） |

### DSN 格式

**SQLite**
```
file:./data.db?cache=shared&mode=rwc
```

**MySQL**
```
user:password@tcp(localhost:3306)/dbname?charset=utf8mb4&parseTime=True&loc=Local
```

**PostgreSQL**
```
host=localhost port=5432 user=postgres password=password dbname=dbname sslmode=disable
```

### 连接池配置

| 参数 | 类型 | 默认值 | 说明 |
|------|------|--------|------|
| `MaxOpenConns` | int | 10 | 最大打开连接数，0 表示无限制 |
| `MaxIdleConns` | int | 5 | 最大空闲连接数 |
| `ConnMaxLifetime` | Duration | 30s | 连接最大存活时间 |
| `ConnMaxIdleTime` | Duration | 5m | 连接最大空闲时间 |

## 错误处理

`BuildWithConfig` 方法会验证配置并返回详细的错误信息：

```go
mgr, err := factory.BuildWithConfig(cfg)
if err != nil {
    // 处理配置错误
    if strings.Contains(err.Error(), "driver is required") {
        log.Fatal("Driver must be specified")
    }
    if strings.Contains(err.Error(), "DSN is required") {
        log.Fatal("Database DSN is required")
    }
    log.Fatal(err)
}
```

`Build` 方法在配置错误时不会返回错误，而是返回 `NoneDatabaseManager`：

```go
mgr := factory.Build("", cfg)
if mgr.ManagerName() == "none-database" {
    log.Fatal("Failed to create database manager")
}
```

## 线程安全

DatabaseManager 的所有方法都是线程安全的，可以在多个 goroutine 中并发使用。

```go
// 并发安全
var wg sync.WaitGroup
for i := 0; i < 10; i++ {
    wg.Add(1)
    go func() {
        defer wg.Done()
        var users []User
        mgr.DB().Find(&users)
    }()
}
wg.Wait()
```
