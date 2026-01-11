# Database Manager - 数据库管理器

提供基于 GORM 的灵活、高性能数据库管理功能，支持 MySQL、PostgreSQL 和 SQLite 三种数据库驱动。

## 特性

- **完全基于 GORM** - 提供 GORM 的全部功能，包括链式查询、自动迁移、钩子等
- **多驱动支持** - 支持 MySQL、PostgreSQL 和 SQLite 三种数据库驱动
- **连接池管理** - 自动管理数据库连接池，支持连接池参数配置
- **自动迁移** - 支持数据库表结构的自动创建和更新
- **事务管理** - 支持事务操作，包括嵌套事务和保存点
- **健康检查** - 提供数据库连接健康检查功能
- **零成本降级** - 配置失败时自动降级到空数据库管理器，避免影响程序运行
- **线程安全** - 所有操作都是线程安全的，支持并发访问

## 快速开始

```go
package main

import (
    "fmt"

    "com.litelake.litecore/manager/databasemgr"
    "gorm.io/gorm"
)

// User 用户模型
type User struct {
    ID   uint   `gorm:"primaryKey"`
    Name string `gorm:"size:255"`
    Age  int    `gorm:"index"`
}

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

    // 自动迁移表结构
    err := dbMgr.AutoMigrate(&User{})
    if err != nil {
        panic(err)
    }

    // 使用 GORM 创建记录
    err = dbMgr.DB().Create(&User{Name: "John", Age: 30}).Error
    if err != nil {
        panic(err)
    }

    // 查询记录
    var users []User
    err = dbMgr.DB().Where("age > ?", 18).Find(&users).Error
    if err != nil {
        panic(err)
    }

    fmt.Printf("找到 %d 个用户\n", len(users))

    // 使用事务
    err = dbMgr.Transaction(func(tx *gorm.DB) error {
        // 创建用户
        if err := tx.Create(&User{Name: "Alice", Age: 25}).Error; err != nil {
            return err
        }

        // 更新用户
        if err := tx.Model(&User{}).Where("name = ?", "John").Update("age", 31).Error; err != nil {
            return err
        }

        return nil
    })
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
| **mysql** | MySQL | 生产环境、高并发场景 | gorm.io/driver/mysql |
| **postgresql** | PostgreSQL | 企业级应用、高级功能需求 | gorm.io/driver/postgres |
| **sqlite** | SQLite | 嵌入式应用、测试环境 | gorm.io/driver/sqlite |
| **none** | - | 降级场景、配置失败时 | - |

## GORM 集成

DatabaseManager 完全基于 GORM 构建，提供 GORM 的所有功能。

### 核心方法

```go
// DB 获取 GORM 数据库实例
db := dbMgr.DB()

// Model 指定模型进行操作
dbMgr.Model(&User{}).Where("age > ?", 18).Find(&users)

// Table 指定表名进行操作
dbMgr.Table("users").Where("age > ?", 18).Find(&results)

// WithContext 设置上下文
ctx := context.Background()
dbMgr.WithContext(ctx).Find(&users)

// Transaction 执行事务
err := dbMgr.Transaction(func(tx *gorm.DB) error {
    return tx.Create(&User{Name: "John"}).Error
})

// Begin 开启事务
tx := dbMgr.Begin()
tx.Create(&User{Name: "Alice"})
tx.Commit()

// AutoMigrate 自动迁移
err := dbMgr.AutoMigrate(&User{}, &Product{})

// Migrator 获取迁移器
migrator := dbMgr.Migrator()
migrator.CreateTable(&User{})

// Exec 执行原生 SQL
dbMgr.Exec("DELETE FROM users WHERE age < ?", 18)

// Raw 执行原生查询
dbMgr.Raw("SELECT * FROM users WHERE age > ?", 18).Scan(&results)
```

### 链式查询

```go
// 使用 GORM 的链式查询
var users []User
dbMgr.DB().
    Select("name, age").
    Where("age > ?", 18).
    Order("age DESC").
    Limit(10).
    Find(&users)
```

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
fmt.Printf("Open Connections: %d\n", stats.MaxOpenConnections)
fmt.Printf("In Use: %d\n", stats.InUse)
fmt.Printf("Idle: %d\n", stats.Idle)
fmt.Printf("Wait Count: %d\n", stats.WaitCount)
fmt.Printf("Wait Duration: %v\n", stats.WaitDuration)
```

## 事务管理

### 基本事务

```go
// 使用 Transaction 方法自动管理事务
err := dbMgr.Transaction(func(tx *gorm.DB) error {
    // 创建用户
    if err := tx.Create(&User{Name: "John"}).Error; err != nil {
        return err // 返回错误会自动回滚
    }

    // 更新用户
    if err := tx.Model(&User{}).Where("name = ?", "Alice").Update("age", 25).Error; err != nil {
        return err
    }

    return nil // 返回 nil 会自动提交
})
```

### 手动事务

```go
// 手动开启事务
tx := dbMgr.Begin()

// 执行操作
tx.Create(&User{Name: "John"})

// 提交或回滚
if err != nil {
    tx.Rollback()
} else {
    tx.Commit()
}
```

### 嵌套事务

```go
// GORM 支持嵌套事务（通过保存点实现）
err := dbMgr.Transaction(func(tx *gorm.DB) error {
    // 外层事务
    if err := tx.Create(&User{Name: "John"}).Error; err != nil {
        return err
    }

    // 内层事务（保存点）
    return tx.Transaction(func(tx2 *gorm.DB) error {
        return tx2.Create(&User{Name: "Alice"}).Error
    })
})
```

## 自动迁移

### 基本用法

```go
// 自动迁移模型
err := dbMgr.AutoMigrate(&User{}, &Product{}, &Order{})
if err != nil {
    panic(err)
}
```

### 模型定义

```go
type User struct {
    ID        uint      `gorm:"primaryKey"`
    Name      string    `gorm:"size:255;not null"`
    Email     string    `gorm:"size:255;uniqueIndex"`
    Age       int       `gorm:"index"`
    Active    bool      `gorm:"default:true"`
    CreatedAt time.Time
    UpdatedAt time.Time
}
```

### 迁移器 API

```go
// 获取迁移器
migrator := dbMgr.Migrator()

// 创建表
migrator.CreateTable(&User{})

// 删除表
migrator.DropTable(&User{})

// 重命名表
migrator.RenameTable("users", "user_accounts")

// 添加列
migrator.AddColumn(&User{}, "Avatar")

// 删除列
migrator.DropColumn(&User{}, "Avatar")

// 创建索引
migrator.CreateIndex(&User{}, "IdxEmail")

// 删除索引
migrator.DropIndex(&User{}, "IdxEmail")

// 检查列是否存在
hasColumn := migrator.HasColumn(&User{}, "Email")

// 检查索引是否存在
hasIndex := migrator.HasIndex(&User{}, "IdxEmail")
```

## API 文档

### DatabaseManager 接口

```go
type DatabaseManager interface {
    // ========== 生命周期管理 ==========
    ManagerName() string
    Health() error
    OnStart() error
    OnStop() error

    // ========== GORM 核心 ==========
    DB() *gorm.DB
    Model(value any) *gorm.DB
    Table(name string) *gorm.DB
    WithContext(ctx context.Context) *gorm.DB

    // ========== 事务管理 ==========
    Transaction(fn func(*gorm.DB) error, opts ...*sql.TxOptions) error
    Begin(opts ...*sql.TxOptions) *gorm.DB

    // ========== 迁移管理 ==========
    AutoMigrate(models ...any) error
    Migrator() gorm.Migrator

    // ========== 连接管理 ==========
    Driver() string
    Ping(ctx context.Context) error
    Stats() gorm.SQLStats
    Close() error

    // ========== 原生 SQL ==========
    Exec(sql string, values ...any) *gorm.DB
    Raw(sql string, values ...any) *gorm.DB
}
```

### 核心方法

#### DB()

获取 GORM 数据库实例，用于执行所有 GORM 操作。

```go
db := dbMgr.DB()
var users []User
db.Where("age > ?", 18).Find(&users)
```

#### Model()

指定模型进行操作。

```go
dbMgr.Model(&User{}).Where("age > ?", 18).Find(&users)
dbMgr.Model(&User{}).Update("active", true)
```

#### Table()

指定表名进行操作。

```go
dbMgr.Table("user_profiles").Where("user_id = ?", userID).First(&profile)
```

#### WithContext()

设置上下文，用于超时控制和追踪。

```go
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()

dbMgr.WithContext(ctx).Find(&users)
```

#### Transaction()

执行事务，自动处理提交和回滚。

```go
err := dbMgr.Transaction(func(tx *gorm.DB) error {
    if err := tx.Create(&User{Name: "John"}).Error; err != nil {
        return err // 自动回滚
    }
    return nil // 自动提交
})
```

#### Begin()

手动开启事务，返回事务对象。

```go
tx := dbMgr.Begin()
tx.Create(&User{Name: "John"})
tx.Commit() // 或 tx.Rollback()
```

#### AutoMigrate()

自动迁移模型到数据库。

```go
err := dbMgr.AutoMigrate(&User{}, &Product{})
```

#### Migrator()

获取 GORM 迁移器，用于高级迁移操作。

```go
migrator := dbMgr.Migrator()
migrator.CreateTable(&User{})
migrator.AddColumn(&User{}, "Avatar")
```

#### Driver()

获取数据库驱动类型。

```go
driver := dbMgr.Driver() // "mysql", "postgresql", "sqlite", "none"
```

#### Ping()

检查数据库连接是否正常。

```go
ctx := context.Background()
err := dbMgr.Ping(ctx)
if err != nil {
    // 数据库连接失败
}
```

#### Stats()

获取连接池统计信息。

```go
stats := dbMgr.Stats()
fmt.Printf("MaxOpenConnections: %d\n", stats.MaxOpenConnections)
fmt.Printf("OpenConnections: %d\n", stats.OpenConnections)
fmt.Printf("InUse: %d\n", stats.InUse)
```

#### Close()

关闭数据库连接。

```go
err := dbMgr.Close()
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
var users []User
if err := dbMgr.DB().Find(&users).Error; err != nil {
    return fmt.Errorf("failed to query users: %w", err)
}

// 使用事务并处理错误
err := dbMgr.Transaction(func(tx *gorm.DB) error {
    if err := tx.Create(&User{Name: "John"}).Error; err != nil {
        return err
    }
    return nil
})
if err != nil {
    return fmt.Errorf("transaction failed: %w", err)
}
```

## 性能考虑

- **连接池配置**：合理配置连接池大小，避免连接泄漏和资源浪费
- **使用索引**：为常用查询字段创建索引
- **批量操作**：使用 CreateInBatches 进行批量插入
- **预加载**：使用 Preload 避免N+1查询问题
- **选择字段**：使用 Select 只查询需要的字段
- **使用 context**：使用 context 控制查询超时

### 性能优化示例

```go
// 批量插入
users := []User{{Name: "John"}, {Name: "Alice"}}
dbMgr.DB().CreateInBatches(users, 100)

// 预加载关联数据
type User struct {
    ID      uint
    Name    string
    Orders  []Order
}
dbMgr.DB().Preload("Orders").Find(&users)

// 只查询需要的字段
dbMgr.DB().Select("id", "name").Find(&users)

// 使用 context 控制超时
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()
dbMgr.WithContext(ctx).Find(&users)
```

## 安全性

- **敏感信息保护**：不要在日志中打印密码，使用环境变量管理敏感信息
- **SSL/TLS 支持**：PostgreSQL 和 MySQL 支持 SSL/TLS 加密连接
- **参数化查询**：GORM 自动使用参数化查询，避免 SQL 注入
- **输入验证**：在应用层验证用户输入

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
2. **使用事务**：保证数据一致性
3. **处理错误**：检查所有可能的错误
4. **监控连接池**：定期检查连接池统计信息
5. **使用 context**：控制查询超时
6. **使用自动迁移**：保持数据库结构与模型同步
7. **使用索引**：为常用查询字段创建索引
8. **避免长事务**：及时提交或回滚事务
9. **使用预加载**：避免 N+1 查询问题
10. **参数化查询**：GORM 自动使用参数化查询，避免 SQL 注入

## 测试

运行测试：

```bash
# 运行所有测试
go test ./manager/databasemgr/...

# 运行测试并显示覆盖率
go test -cover ./manager/databasemgr/...

# 运行特定模块的测试
go test ./manager/databasemgr/internal/config/...
go test ./manager/databasemgr/internal/hooks/...
go test ./manager/databasemgr/internal/migration/...
go test ./manager/databasemgr/internal/transaction/...

# 运行特定驱动的测试
go test ./manager/databasemgr/internal/drivers/... -run TestSQLite
```

## 架构说明

DatabaseManager 由以下核心组件构成：

- **config**: 配置解析和验证
- **drivers**: 数据库驱动实现（MySQL、PostgreSQL、SQLite、None）
- **hooks**: GORM 钩子管理
- **migration**: 数据库迁移管理
- **transaction**: 事务管理

所有组件都经过完整的单元测试覆盖，确保功能正确性和稳定性。
