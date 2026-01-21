# Database Manager - 数据库管理器

提供统一的数据库管理功能,基于 GORM 支持多种数据库驱动。

## 特性

- **多数据库支持** - 支持 MySQL、PostgreSQL、SQLite 和 None(空实现)驱动
- **连接池管理** - 统一的连接池配置和统计监控
- **可观测性集成** - 内置日志、链路追踪和指标收集
- **事务管理** - 支持事务操作和自动回滚
- **自动迁移** - 基于 GORM 的数据库 Schema 迁移能力
- **配置驱动** - 支持通过配置提供者创建实例
- **依赖注入** - 集成日志管理器和可观测性管理器

## 快速开始

```go
package main

import (
    "log"
    "github.com/lite-lake/litecore-go/manager/databasemgr"
)

func main() {
    // 方式1: 使用工厂函数直接创建
    cfg := map[string]any{
        "dsn": "root:password@tcp(localhost:3306)/mydb?charset=utf8mb4&parseTime=True&loc=Local",
        "pool_config": map[string]any{
            "max_open_conns":     100,
            "max_idle_conns":     10,
            "conn_max_lifetime":  3600,
            "conn_max_idle_time": 600,
        },
    }
    dbMgr, err := databasemgr.Build("mysql", cfg)
    if err != nil {
        log.Fatal(err)
    }
    defer dbMgr.Close()

    // 使用 GORM 进行数据库操作
    type User struct {
        ID   uint   `gorm:"primarykey"`
        Name string `gorm:"size:255"`
    }

    // 自动迁移
    if err := dbMgr.AutoMigrate(&User{}); err != nil {
        log.Fatal(err)
    }

    // 创建记录
    user := User{Name: "Alice"}
    if err := dbMgr.DB().Create(&user).Error; err != nil {
        log.Fatal(err)
    }

    // 查询记录
    var result User
    if err := dbMgr.DB().First(&result, user.ID).Error; err != nil {
        log.Fatal(err)
    }

    log.Printf("User: %+v\n", result)
}
```

## 创建数据库管理器

### 使用 Build 函数

`Build` 函数是最简单的创建方式,直接指定驱动类型和配置:

```go
// MySQL
mysqlCfg := map[string]any{
    "dsn": "root:password@tcp(localhost:3306)/mydb?charset=utf8mb4&parseTime=True&loc=Local",
}
dbMgr, err := databasemgr.Build("mysql", mysqlCfg)

// PostgreSQL
postgresqlCfg := map[string]any{
    "dsn": "host=localhost port=5432 user=postgres password=password dbname=mydb sslmode=disable",
}
dbMgr, err := databasemgr.Build("postgresql", postgresqlCfg)

// SQLite
sqliteCfg := map[string]any{
    "dsn": "file:./cache.db?cache=shared&mode=rwc",
}
dbMgr, err := databasemgr.Build("sqlite", sqliteCfg)

// None (空实现,用于测试或不需要数据库的场景)
dbMgr := databasemgr.Build("none", nil)
```

### 使用 BuildWithConfigProvider

`BuildWithConfigProvider` 从配置提供者读取配置,适合依赖注入场景:

```go
import "github.com/lite-lake/litecore-go/common"

// 创建配置提供者
provider := config.NewYamlConfigProvider("config.yaml")

// 从配置创建数据库管理器
dbMgr, err := databasemgr.BuildWithConfigProvider(provider)
if err != nil {
    log.Fatal(err)
}
defer dbMgr.Close()
```

配置文件示例(YAML):

```yaml
database:
  driver: mysql

  mysql_config:
    dsn: "root:password@tcp(localhost:3306)/mydb?charset=utf8mb4&parseTime=True&loc=Local"
    pool_config:
      max_open_conns: 100
      max_idle_conns: 10
      conn_max_lifetime: 3600  # 秒
      conn_max_idle_time: 600   # 秒

  observability_config:
    slow_query_threshold: 1s   # 慢查询阈值
    log_sql: false              # 是否记录完整 SQL
    sample_rate: 1.0            # 采样率 (0.0-1.0)
```

### 使用构造函数

对于更精细的控制,可以使用各驱动的构造函数:

```go
// MySQL
mysqlCfg := &databasemgr.MySQLConfig{
    DSN: "root:password@tcp(localhost:3306)/mydb?charset=utf8mb4&parseTime=True&loc=Local",
    PoolConfig: &databasemgr.PoolConfig{
        MaxOpenConns:    100,
        MaxIdleConns:    10,
        ConnMaxLifetime: 3600 * time.Second,
        ConnMaxIdleTime: 600 * time.Second,
    },
}
dbMgr, err := databasemgr.NewDatabaseManagerMySQLImpl(mysqlCfg)

// PostgreSQL
postgresqlCfg := &databasemgr.PostgreSQLConfig{
    DSN: "host=localhost port=5432 user=postgres password=password dbname=mydb sslmode=disable",
    PoolConfig: &databasemgr.PoolConfig{...},
}
dbMgr, err := databasemgr.NewDatabaseManagerPostgreSQLImpl(postgresqlCfg)

// SQLite
sqliteCfg := &databasemgr.SQLiteConfig{
    DSN: "file:./cache.db?cache=shared&mode=rwc",
    PoolConfig: &databasemgr.PoolConfig{
        MaxOpenConns: 1,
        MaxIdleConns: 1,
    },
}
dbMgr, err := databasemgr.NewDatabaseManagerSQLiteImpl(sqliteCfg)

// None
dbMgr := databasemgr.NewDatabaseManagerNoneImpl()
```

## GORM 核心

### 获取 DB 实例

所有 GORM 操作都通过 `DB()` 方法获取数据库实例:

```go
db := dbMgr.DB()

// 简单查询
var user User
db.First(&user, 1)

// 条件查询
db.Where("name = ?", "Alice").First(&user)

// 创建记录
db.Create(&User{Name: "Bob"})

// 更新记录
db.Model(&user).Update("Name", "Charlie")

// 删除记录
db.Delete(&user)
```

### 指定模型

```go
// 使用 Model 方法
dbMgr.Model(&User{}).Where("age > ?", 18).Find(&users)

// 使用 Table 方法(直接操作表名)
dbMgr.Table("users").Count(&count)
```

### 使用上下文

```go
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()

// 所有操作都会使用这个上下文
dbMgr.WithContext(ctx).Find(&users)
```

## 事务管理

### 自动事务

`Transaction` 方法会自动处理提交和回滚:

```go
err := dbMgr.Transaction(func(tx *gorm.DB) error {
    // 创建用户
    if err := tx.Create(&User{Name: "Alice"}).Error; err != nil {
        return err  // 返回错误会触发回滚
    }

    // 创建订单
    if err := tx.Create(&Order{UserID: 1}).Error; err != nil {
        return err  // 返回错误会触发回滚
    }

    return nil  // 返回 nil 会提交事务
})
if err != nil {
    log.Error("transaction failed", err)
}
```

### 手动事务

```go
// 开始事务
tx := dbMgr.Begin()

// 执行操作
if err := tx.Create(&User{Name: "Bob"}).Error; err != nil {
    tx.Rollback()  // 回滚
    log.Fatal(err)
}

if err := tx.Create(&Order{UserID: 2}).Error; err != nil {
    tx.Rollback()  // 回滚
    log.Fatal(err)
}

// 提交事务
tx.Commit()
```

## 迁移管理

### 自动迁移

```go
// 自动迁移表结构
err := dbMgr.AutoMigrate(&User{}, &Product{}, &Order{})
if err != nil {
    log.Fatal(err)
}
```

### 使用 Migrator

```go
migrator := dbMgr.Migrator()

// 检查表是否存在
if migrator.HasTable(&User{}) {
    log.Println("Users table exists")
}

// 创建表
migrator.CreateTable(&User{})

// 删除表(慎用)
migrator.DropTable(&User{})

// 重命名表
migrator.RenameTable(&User{}, "users_new")

// 添加列
migrator.AddColumn(&User{}, "Email")
```

## 连接管理

### 健康检查

```go
if err := dbMgr.Health(); err != nil {
    log.Error("database health check failed", err)
}
```

### Ping 检查

```go
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()

if err := dbMgr.Ping(ctx); err != nil {
    log.Error("database ping failed", err)
}
```

### 连接池统计

```go
stats := dbMgr.Stats()
log.Printf("OpenConnections: %d", stats.OpenConnections)
log.Printf("InUse: %d", stats.InUse)
log.Printf("Idle: %d", stats.Idle)
log.Printf("WaitCount: %d", stats.WaitCount)
log.Printf("WaitDuration: %v", stats.WaitDuration)
```

### 关闭连接

```go
defer dbMgr.Close()
```

## 原生 SQL

### 执行查询

```go
type Result struct {
    ID   int
    Name string
}

var results []Result
dbMgr.Raw("SELECT id, name FROM users WHERE age > ?", 18).Scan(&results)
```

### 执行命令

```go
dbMgr.Exec("UPDATE users SET status = ? WHERE id = ?", "active", 1)
```

## 可观测性

### 日志

Database Manager 会自动记录:
- 所有数据库操作的耗时
- 慢查询(Warn 级别)
- 错误(Error 级别)

### 链路追踪

所有数据库操作都会自动创建 span,包含:
- 操作类型(query/create/update/delete)
- 表名
- 执行状态
- 错误信息(如果有)

### 指标

收集的指标包括:
- `db.query.duration` - 查询耗时直方图
- `db.query.count` - 查询计数器
- `db.query.error_count` - 错误计数器
- `db.query.slow_count` - 慢查询计数器
- `db.transaction.count` - 事务计数器
- `db.connection.pool` - 连接池状态

### 配置可观测性

```go
cfg := &databasemgr.DatabaseConfig{
    Driver: "mysql",
    MySQLConfig: &databasemgr.MySQLConfig{
        DSN: "...",
    },
    ObservabilityConfig: &databasemgr.ObservabilityConfig{
        SlowQueryThreshold: 1 * time.Second,  // 慢查询阈值
        LogSQL:             false,             // 是否记录完整 SQL(生产环境建议关闭)
        SampleRate:         1.0,               // 采样率(0.0-1.0)
    },
}
```

## API

### 工厂函数

```go
// Build 创建数据库管理器实例
func Build(driverType string, driverConfig map[string]any) (DatabaseManager, error)

// BuildWithConfigProvider 从配置提供者创建数据库管理器实例
func BuildWithConfigProvider(configProvider common.BaseConfigProvider) (DatabaseManager, error)
```

### 构造函数

```go
// NewDatabaseManagerMySQLImpl 创建 MySQL 数据库管理器
func NewDatabaseManagerMySQLImpl(cfg *MySQLConfig) (DatabaseManager, error)

// NewDatabaseManagerPostgreSQLImpl 创建 PostgreSQL 数据库管理器
func NewDatabaseManagerPostgreSQLImpl(cfg *PostgreSQLConfig) (DatabaseManager, error)

// NewDatabaseManagerSQLiteImpl 创建 SQLite 数据库管理器
func NewDatabaseManagerSQLiteImpl(cfg *SQLiteConfig) (DatabaseManager, error)

// NewDatabaseManagerNoneImpl 创建空数据库管理器
func NewDatabaseManagerNoneImpl() DatabaseManager
```

### DatabaseManager 接口

#### 生命周期管理

```go
// ManagerName 返回管理器名称
ManagerName() string

// Health 检查管理器健康状态
Health() error

// OnStart 在服务器启动时触发
OnStart() error

// OnStop 在服务器停止时触发
OnStop() error
```

#### GORM 核心

```go
// DB 获取 GORM 数据库实例
DB() *gorm.DB

// Model 指定模型进行操作
Model(value any) *gorm.DB

// Table 指定表名进行操作
Table(name string) *gorm.DB

// WithContext 设置上下文
WithContext(ctx context.Context) *gorm.DB
```

#### 事务管理

```go
// Transaction 执行事务
Transaction(fn func(*gorm.DB) error, opts ...*sql.TxOptions) error

// Begin 开启事务
Begin(opts ...*sql.TxOptions) *gorm.DB
```

#### 迁移管理

```go
// AutoMigrate 自动迁移
AutoMigrate(models ...any) error

// Migrator 获取迁移器
Migrator() gorm.Migrator
```

#### 连接管理

```go
// Driver 获取驱动类型
Driver() string

// Ping 检查数据库连接
Ping(ctx context.Context) error

// Stats 获取连接池统计信息
Stats() sql.DBStats

// Close 关闭数据库连接
Close() error
```

#### 原生 SQL

```go
// Exec 执行原生 SQL
Exec(sql string, values ...any) *gorm.DB

// Raw 执行原生查询
Raw(sql string, values ...any) *gorm.DB
```

## 配置

### DatabaseConfig

```go
type DatabaseConfig struct {
    Driver              string                 // 驱动类型
    SQLiteConfig        *SQLiteConfig          // SQLite 配置
    PostgreSQLConfig    *PostgreSQLConfig      // PostgreSQL 配置
    MySQLConfig         *MySQLConfig           // MySQL 配置
    ObservabilityConfig *ObservabilityConfig   // 可观测性配置
}
```

### PoolConfig

```go
type PoolConfig struct {
    MaxOpenConns    int           // 最大打开连接数
    MaxIdleConns    int           // 最大空闲连接数
    ConnMaxLifetime time.Duration // 连接最大存活时间
    ConnMaxIdleTime time.Duration // 连接最大空闲时间
}
```

### MySQLConfig

```go
type MySQLConfig struct {
    DSN        string      // MySQL DSN
    PoolConfig *PoolConfig // 连接池配置
}
```

### PostgreSQLConfig

```go
type PostgreSQLConfig struct {
    DSN        string      // PostgreSQL DSN
    PoolConfig *PoolConfig // 连接池配置
}
```

### SQLiteConfig

```go
type SQLiteConfig struct {
    DSN        string      // SQLite DSN
    PoolConfig *PoolConfig // 连接池配置
}
```

### ObservabilityConfig

```go
type ObservabilityConfig struct {
    SlowQueryThreshold time.Duration // 慢查询阈值
    LogSQL             bool          // 是否记录完整 SQL
    SampleRate         float64       // 采样率(0.0-1.0)
}
```

## 最佳实践

### 连接池配置

- **MySQL/PostgreSQL**: 根据应用并发量调整,通常 MaxOpenConns 设置为 CPU 核心数的 2-4 倍
- **SQLite**: MaxOpenConns 通常设置为 1,避免写锁冲突
- **监控**: 使用 `Stats()` 定期检查连接池状态,避免连接泄漏

### 事务使用

- 优先使用 `Transaction` 方法,自动处理回滚
- 保持事务简短,避免长时间持有锁
- 在事务中避免外部调用(如 HTTP 请求)

### 错误处理

- 始终检查数据库操作的错误
- 使用 `Health()` 方法实现健康检查端点
- 对于连接错误,考虑实现重试机制

### 性能优化

- 使用索引优化查询性能
- 批量操作使用 `CreateInBatches` 或 `Clauses`
- 合理使用 `Preload` 避免 N+1 查询
- 对于只读查询,使用 `Read().Unscoped()` 提高性能

### 安全建议

- 使用参数化查询,避免 SQL 注入
- 敏感信息(如密码)不要记录在日志中
- 生产环境关闭 `LogSQL`,避免敏感数据泄漏
- 使用环境变量管理数据库凭证

### DSN 格式

**MySQL**:
```
user:password@tcp(localhost:3306)/dbname?charset=utf8mb4&parseTime=True&loc=Local
```

**PostgreSQL**:
```
host=localhost port=5432 user=postgres password=password dbname=dbname sslmode=disable
```

**SQLite**:
```
file:./cache.db?cache=shared&mode=rwc
```

## 错误处理

所有错误都遵循 Go 的错误处理惯例:

```go
if err := dbMgr.AutoMigrate(&User{}); err != nil {
    log.Printf("AutoMigrate failed: %v", err)
    return err
}
```

常见错误:
- 配置错误: 检查 DSN 和连接池配置
- 连接错误: 检查数据库服务是否运行,网络是否可达
- 迁移错误: 检查表结构定义是否正确
- 查询错误: 检查 SQL 语法和数据类型匹配

## 线程安全

DatabaseManager 的所有方法都是线程安全的,可以并发使用。但对于需要多次操作的场景,建议使用事务确保一致性。
