# Database Manager - 数据库管理器

提供灵活、高性能的数据库管理功能，支持 MySQL、PostgreSQL 和 SQLite 三种数据库驱动。

## 特性

- **多驱动支持** - 支持 MySQL、PostgreSQL 和 SQLite 三种数据库驱动
- **连接池管理** - 自动管理数据库连接池，支持连接池参数配置
- **健康检查** - 提供数据库连接健康检查功能
- **事务支持** - 支持事务操作，保证数据一致性
- **零成本降级** - 配置失败时自动降级到空数据库管理器，避免影响程序运行
- **线程安全** - 所有操作都是线程安全的，支持并发访问

## 快速开始

```go
package main

import (
    "context"
    "fmt"

    "com.litelake.litecore/manager/databasemgr"
)

func main() {
    // 创建工厂实例
    factory := databasemgr.NewFactory()

    // 配置数据库（使用 SQLite）
    cfg := map[string]any{
        "driver": "sqlite",
        "sqlite_config": map[string]any{
            "dsn": "file:./data.db?cache=shared&mode=rwc",
        },
    }

    // 构建数据库管理器
    mgr := factory.Build("sqlite", cfg)
    dbMgr := mgr.(databasemgr.DatabaseManager)

    // 检查驱动类型
    if dbMgr.Driver() == "none" {
        fmt.Println("数据库初始化失败，使用降级模式")
        return
    }

    // 执行查询
    ctx := context.Background()
    rows, err := dbMgr.DB().QueryContext(ctx, "SELECT * FROM users")
    if err != nil {
        panic(err)
    }
    defer rows.Close()

    // 使用事务
    tx, err := dbMgr.BeginTx(ctx, nil)
    if err != nil {
        panic(err)
    }
    defer tx.Rollback()

    _, err = tx.Exec("INSERT INTO users (name) VALUES (?)", "John")
    if err != nil {
        panic(err)
    }

    err = tx.Commit()
    if err != nil {
        panic(err)
    }

    // 关闭连接
    _ = dbMgr.Close()
}
```

## 数据库驱动

包支持四种数据库驱动，每种驱动适用于不同的场景：

| 驱动 | 数据库 | 适合场景 | 底层驱动 |
|------|--------|----------|----------|
| **mysql** | MySQL | 生产环境、高并发场景 | go-sql-driver/mysql |
| **postgresql** | PostgreSQL | 企业级应用、高级功能需求 | lib/pq |
| **sqlite** | SQLite | 嵌入式应用、测试环境 | mattn/go-sqlite3 |
| **none** | - | 降级场景、配置失败时 | - |

## 工厂模式

### Build 方法

使用 map 配置构建数据库管理器，配置失败时自动降级为 none 驱动：

```go
factory := databasemgr.NewFactory()

// 使用 map 配置
cfg := map[string]any{
    "driver": "sqlite",
    "sqlite_config": map[string]any{
        "dsn": "file:./data.db?cache=shared&mode=rwc",
        "pool_config": map[string]any{
            "max_open_conns": 1,
        },
    },
}

mgr := factory.Build("sqlite", cfg)
dbMgr := mgr.(databasemgr.DatabaseManager)
```

### BuildWithConfig 方法

使用结构体配置构建数据库管理器，配置失败时返回错误：

```go
factory := databasemgr.NewFactory()

// 使用结构体配置
config := &config.DatabaseConfig{
    Driver: "sqlite",
    SQLiteConfig: &config.SQLiteConfig{
        DSN: "file:./data.db?cache=shared&mode=rwc",
        PoolConfig: &config.PoolConfig{
            MaxOpenConns: 1,
            MaxIdleConns: 1,
        },
    },
}

mgr, err := factory.BuildWithConfig(config)
if err != nil {
    panic(err)
}
dbMgr := mgr.(databasemgr.DatabaseManager)
```

## 配置说明

### MySQL 配置

MySQL 适合生产环境和高并发场景。

```go
cfg := map[string]any{
    "driver": "mysql",
    "mysql_config": map[string]any{
        "dsn": "root:password@tcp(localhost:3306)/lite_demo?charset=utf8mb4&parseTime=True&loc=Local",
        "pool_config": map[string]any{
            "max_open_conns":    10,
            "max_idle_conns":    5,
            "conn_max_lifetime": "30s",
            "conn_max_idle_time": "5m",
        },
    },
}
```

**MySQL DSN 格式**：
```
username:password@protocol(address)/dbname?param=value
```

**常用 DSN 参数**：
- `charset` - 字符集，推荐 `utf8mb4`
- `parseTime` - 是否解析时间，推荐 `True`
- `loc` - 时区，推荐 `Local`
- `timeout` - 连接超时时间
- `readTimeout` - 读取超时时间
- `writeTimeout` - 写入超时时间

### PostgreSQL 配置

PostgreSQL 适合需要高级功能的企业级应用。

```go
cfg := map[string]any{
    "driver": "postgresql",
    "postgresql_config": map[string]any{
        "dsn": "host=localhost port=5432 user=postgres password=password dbname=lite_demo sslmode=disable",
        "pool_config": map[string]any{
            "max_open_conns":    10,
            "max_idle_conns":    5,
            "conn_max_lifetime": "30s",
            "conn_max_idle_time": "5m",
        },
    },
}
```

**PostgreSQL DSN 格式**：
```
host=port user=password dbname=sslmode=connect_timeout=
```

**常用 DSN 参数**：
- `host` - 数据库主机地址
- `port` - 数据库端口，默认 `5432`
- `user` - 数据库用户名
- `password` - 数据库密码
- `dbname` - 数据库名称
- `sslmode` - SSL 模式（`disable`、`require`、`verify-ca`、`verify-full`）
- `connect_timeout` - 连接超时时间

### SQLite 配置

SQLite 适合嵌入式应用和测试环境。

```go
cfg := map[string]any{
    "driver": "sqlite",
    "sqlite_config": map[string]any{
        "dsn": "file:./data.db?cache=shared&mode=rwc",
        "pool_config": map[string]any{
            "max_open_conns":    1,  // SQLite 建议设置为 1
            "max_idle_conns":    1,
            "conn_max_lifetime": "30s",
            "conn_max_idle_time": "5m",
        },
    },
}
```

**SQLite DSN 格式**：
```
file:path?param=value
```

**常用 DSN 参数**：
- `cache=shared` - 启用共享缓存
- `mode=rwc` - 读写创建模式（read-write-create）
- `_sync=NORMAL` - 同步模式（`FULL`、`NORMAL`、`OFF`）

## 连接池配置

### 默认配置

| 参数 | 默认值 | 说明 |
|------|--------|------|
| `max_open_conns` | 10 | 最大打开连接数，0 表示无限制 |
| `max_idle_conns` | 5 | 最大空闲连接数 |
| `conn_max_lifetime` | 30s | 连接最大存活时间 |
| `conn_max_idle_time` | 5m | 连接最大空闲时间 |

### 最佳实践

- **SQLite**：设置 `max_open_conns` 为 1，因为 SQLite 不支持并发写入
- **MySQL/PostgreSQL**：根据应用负载调整连接池大小，建议初始值为 10-50
- **监控**：定期检查连接池统计信息，调整参数

```go
// 获取连接池统计信息
stats := dbMgr.Stats()
fmt.Printf("Open Connections: %d\n", stats.OpenConnections)
fmt.Printf("In Use: %d\n", stats.InUse)
fmt.Printf("Idle: %d\n", stats.Idle)
fmt.Printf("Wait Count: %d\n", stats.WaitCount)
fmt.Printf("Wait Duration: %v\n", stats.WaitDuration)
```

## API 文档

### DatabaseManager 接口

```go
type DatabaseManager interface {
    // ManagerName 返回管理器名称
    ManagerName() string

    // Health 检查管理器健康状态
    Health() error

    // OnStart 在服务器启动时触发
    OnStart() error

    // OnStop 在服务器停止时触发
    OnStop() error

    // DB 获取数据库连接
    DB() *sql.DB

    // Driver 获取数据库驱动类型
    Driver() string

    // Ping 检查数据库连接
    Ping(ctx context.Context) error

    // BeginTx 开始事务
    BeginTx(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error)

    // Stats 获取数据库连接池统计信息
    Stats() sql.DBStats

    // Close 关闭数据库连接
    Close() error
}
```

### 核心方法

#### DB()

获取底层的 `*sql.DB` 实例，用于执行原生 SQL 操作。

```go
db := dbMgr.DB()
rows, err := db.QueryContext(ctx, "SELECT * FROM users WHERE id = ?", userID)
```

#### Ping()

检查数据库连接是否正常。

```go
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()

err := dbMgr.Ping(ctx)
if err != nil {
    // 数据库连接失败
}
```

#### BeginTx()

开始一个新事务。

```go
tx, err := dbMgr.BeginTx(ctx, &sql.TxOptions{
    Isolation: sql.LevelSerializable,
    ReadOnly:  false,
})
if err != nil {
    return err
}
defer tx.Rollback()

// 执行事务操作
_, err = tx.Exec("UPDATE accounts SET balance = balance - ? WHERE id = ?", amount, fromID)
if err != nil {
    return err
}

// 提交事务
err = tx.Commit()
```

#### Stats()

获取连接池统计信息，用于监控和调优。

```go
stats := dbMgr.Stats()
fmt.Printf("Open Connections: %d\n", stats.OpenConnections)
fmt.Printf("In Use: %d\n", stats.InUse)
fmt.Printf("Idle: %d\n", stats.Idle)
```

#### Driver()

获取当前使用的数据库驱动类型，可用于检查降级状态。

```go
if dbMgr.Driver() == "none" {
    return errors.New("database not available")
}
```

#### Close()

关闭数据库连接并释放资源。

```go
err := dbMgr.Close()
if err != nil {
    log.Printf("failed to close database: %v", err)
}
```

## 错误处理

### 零成本降级策略

数据库管理器采用零成本降级策略，确保配置失败时程序仍可运行：

| 场景 | 行为 |
|------|------|
| 配置解析失败 | 返回 none 驱动的管理器 |
| 连接创建失败 | 返回 none 驱动的管理器 |
| 驱动类型未知 | 返回 none 驱动的管理器 |
| 查询执行失败 | 返回原始错误，由调用方处理 |
| Health check 失败 | 返回错误，但不影响管理器运行 |

### 错误处理示例

```go
factory := databasemgr.NewFactory()
mgr := factory.Build("sqlite", cfg)
dbMgr := mgr.(databasemgr.DatabaseManager)

// 检查是否降级
if dbMgr.Driver() == "none" {
    // 数据库不可用，使用降级逻辑
    return errors.New("database not available")
}

// 执行查询并处理错误
rows, err := dbMgr.DB().QueryContext(ctx, "SELECT * FROM users")
if err != nil {
    return fmt.Errorf("failed to query users: %w", err)
}
defer rows.Close()
```

## 性能考虑

- **连接池配置**：合理配置连接池大小，避免连接泄漏和资源浪费
- **使用连接池**：复用数据库连接，减少连接创建开销
- **使用 context**：使用 context 控制查询超时，避免长时间阻塞
- **避免长事务**：及时提交或回滚事务，减少锁竞争
- **使用预编译语句**：使用预编译语句提高查询性能

### 性能优化示例

```go
// 使用 context 控制超时
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()

rows, err := db.QueryContext(ctx, "SELECT * FROM users")

// 使用预编译语句
stmt, err := db.Prepare("SELECT * FROM users WHERE id = ?")
if err != nil {
    return err
}
defer stmt.Close()

rows, err = stmt.Query(userID)
```

## 安全性

- **敏感信息保护**：不要在日志中打印密码，使用环境变量管理敏感信息
- **SSL/TLS 支持**：PostgreSQL 和 MySQL 支持 SSL/TLS 加密连接
- **参数化查询**：使用参数化查询避免 SQL 注入

### 安全配置示例

```go
// 使用环境变量配置
dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?sslmode=require",
    os.Getenv("DB_USER"),
    os.Getenv("DB_PASSWORD"),
    os.Getenv("DB_HOST"),
    3306,
    os.Getenv("DB_NAME"),
)

cfg := map[string]any{
    "driver": "mysql",
    "mysql_config": map[string]any{
        "dsn": dsn,
    },
}
```

## 最佳实践

1. **使用连接池**：不要每次查询都创建新连接
2. **及时释放资源**：使用 defer 关闭 rows 和 stmt
3. **使用事务**：保证数据一致性
4. **处理错误**：检查所有可能的错误
5. **监控连接池**：定期检查连接池统计信息
6. **使用 context**：控制查询超时
7. **使用预编译语句**：提高查询性能
8. **避免长事务**：及时提交或回滚事务

## 测试

运行测试：

```bash
# 运行所有测试
go test ./manager/databasemgr/...

# 运行测试并显示覆盖率
go test -cover ./manager/databasemgr/...

# 运行特定驱动的测试
go test ./manager/databasemgr/internal/drivers/... -run TestSQLite
```
