# databasemgr - 数据库管理器

提供统一的数据库管理功能，基于 GORM 支持多种数据库驱动。

## 特性

- **多驱动支持** - MySQL、PostgreSQL、SQLite 和 None（空实现）
- **统一接口** - 完全基于 GORM，提供 `IDatabaseManager` 统一接口
- **连接池管理** - 统一的连接池配置和统计监控
- **可观测性集成** - 内置日志、链路追踪和指标收集（支持 OpenTelemetry）
- **事务管理** - 支持自动事务和手动事务
- **自动迁移** - 基于 GORM 的数据库 Schema 迁移能力
- **配置驱动** - 支持通过配置提供者创建实例
- **依赖注入** - 集成日志管理器和可观测性管理器

## 快速开始

```go
import (
    "github.com/lite-lake/litecore-go/manager/databasemgr"
)

// 使用工厂函数创建
cfg := map[string]any{
    "dsn": "root:password@tcp(localhost:3306)/mydb?charset=utf8mb4&parseTime=True&loc=Local",
}

dbMgr, err := databasemgr.Build("mysql", cfg)
if err != nil {
    log.Fatal(err)
}
defer dbMgr.Close()

// 定义模型
type User struct {
    ID   uint   `gorm:"primarykey"`
    Name string `gorm:"size:255"`
}

// 自动迁移
dbMgr.AutoMigrate(&User{})

// 创建记录
dbMgr.DB().Create(&User{Name: "Alice"})

// 查询记录
var user User
dbMgr.DB().First(&user, 1)
```

## 创建管理器

### 使用 Build 函数

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

// None（空实现，用于测试）
dbMgr := databasemgr.NewDatabaseManagerNoneImpl()
```

### 使用 BuildWithConfigProvider

从配置提供者读取配置，适合依赖注入场景：

```go
import "github.com/lite-lake/litecore-go/manager/configmgr"

provider, err := configmgr.Build("yaml", "config.yaml")
if err != nil {
    log.Fatal(err)
}

dbMgr, err := databasemgr.BuildWithConfigProvider(provider)
if err != nil {
    log.Fatal(err)
}
defer dbMgr.Close()
```

配置文件示例（YAML）：

```yaml
database:
  driver: mysql

  mysql_config:
    dsn: "root:password@tcp(localhost:3306)/mydb?charset=utf8mb4&parseTime=True&loc=Local"
    pool_config:
      max_open_conns: 100
      max_idle_conns: 10
      conn_max_lifetime: 3600
      conn_max_idle_time: 600

  observability_config:
    slow_query_threshold: 1s
    log_sql: false
    sample_rate: 1.0
```

### 使用构造函数

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
```

## GORM 核心

### 获取 DB 实例

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

### 指定模型和表

```go
// 使用 Model 方法
dbMgr.Model(&User{}).Where("age > ?", 18).Find(&users)

// 使用 Table 方法
dbMgr.Table("users").Count(&count)
```

### 使用上下文

```go
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()

dbMgr.WithContext(ctx).Find(&users)
```

## 事务管理

### 自动事务

```go
err := dbMgr.Transaction(func(tx *gorm.DB) error {
    if err := tx.Create(&User{Name: "Alice"}).Error; err != nil {
        return err // 返回错误会触发回滚
    }
    if err := tx.Create(&Order{UserID: 1}).Error; err != nil {
        return err // 返回错误会触发回滚
    }
    return nil // 返回 nil 会提交事务
})
```

### 手动事务

```go
tx := dbMgr.Begin()
if err := tx.Create(&User{Name: "Bob"}).Error; err != nil {
    tx.Rollback()
    log.Fatal(err)
}
tx.Commit()
```

## 迁移管理

### 自动迁移

```go
dbMgr.AutoMigrate(&User{}, &Product{}, &Order{})
```

### 使用 Migrator

```go
migrator := dbMgr.Migrator()

// 检查表是否存在
if migrator.HasTable(&User{}) {
    log.Println("Users table exists")
}

// 创建/删除/重命名表
migrator.CreateTable(&User{})
migrator.DropTable(&User{})
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
```

## 可观测性

### 指标

收集的指标包括：
- `db.query.duration` - 查询耗时直方图
- `db.query.count` - 查询计数器
- `db.query.error_count` - 错误计数器
- `db.query.slow_count` - 慢查询计数器
- `db.transaction.count` - 事务计数器
- `db.connection.pool` - 连接池状态

### 可观测性配置

```go
cfg := &databasemgr.DatabaseConfig{
    Driver: "mysql",
    MySQLConfig: &databasemgr.MySQLConfig{
        DSN: "...",
    },
    ObservabilityConfig: &databasemgr.ObservabilityConfig{
        SlowQueryThreshold: 1 * time.Second,
        LogSQL:             false,
        SampleRate:         1.0,
    },
}
```

## API

### 工厂函数

```go
func Build(driverType string, driverConfig map[string]any) (IDatabaseManager, error)
func BuildWithConfigProvider(configProvider configmgr.IConfigManager) (IDatabaseManager, error)
```

### 构造函数

```go
func NewDatabaseManagerMySQLImpl(cfg *MySQLConfig) (IDatabaseManager, error)
func NewDatabaseManagerPostgreSQLImpl(cfg *PostgreSQLConfig) (IDatabaseManager, error)
func NewDatabaseManagerSQLiteImpl(cfg *SQLiteConfig) (IDatabaseManager, error)
func NewDatabaseManagerNoneImpl() IDatabaseManager
```

### IDatabaseManager 接口

#### GORM 核心

```go
DB() *gorm.DB
Model(value any) *gorm.DB
Table(name string) *gorm.DB
WithContext(ctx context.Context) *gorm.DB
```

#### 事务管理

```go
Transaction(fn func(*gorm.DB) error, opts ...*sql.TxOptions) error
Begin(opts ...*sql.TxOptions) *gorm.DB
```

#### 迁移管理

```go
AutoMigrate(models ...any) error
Migrator() gorm.Migrator
```

#### 连接管理

```go
Driver() string
Ping(ctx context.Context) error
Stats() sql.DBStats
Close() error
```

#### 原生 SQL

```go
Exec(sql string, values ...any) *gorm.DB
Raw(sql string, values ...any) *gorm.DB
```

## 配置

### PoolConfig

```go
type PoolConfig struct {
    MaxOpenConns    int           // 最大打开连接数
    MaxIdleConns    int           // 最大空闲连接数
    ConnMaxLifetime time.Duration // 连接最大存活时间
    ConnMaxIdleTime time.Duration // 连接最大空闲时间
}
```

### ObservabilityConfig

```go
type ObservabilityConfig struct {
    SlowQueryThreshold time.Duration // 慢查询阈值
    LogSQL             bool          // 是否记录完整 SQL
    SampleRate         float64       // 采样率 (0.0-1.0)
}
```

## 最佳实践

### 连接池配置

- **MySQL/PostgreSQL**：MaxOpenConns 设置为 CPU 核心数的 2-4 倍
- **SQLite**：MaxOpenConns 通常设置为 1，避免写锁冲突
- 使用 `Stats()` 定期检查连接池状态

### 事务使用

- 优先使用 `Transaction` 方法
- 保持事务简短，避免长时间持有锁
- 在事务中避免外部调用

### DSN 格式

**MySQL**：`user:password@tcp(localhost:3306)/dbname?charset=utf8mb4&parseTime=True&loc=Local`

**PostgreSQL**：`host=localhost port=5432 user=postgres password=password dbname=dbname sslmode=disable`

**SQLite**：`file:./cache.db?cache=shared&mode=rwc`
