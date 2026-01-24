# databasemgr - 数据库管理器

提供统一的数据库管理功能，基于 GORM 支持 MySQL、PostgreSQL 和 SQLite。

## 特性

- **多数据库支持** - MySQL、PostgreSQL、SQLite 和 None（空实现，用于测试）
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

dbMgr, err := databasemgr.Build("mysql", cfg, nil, nil)
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
dbMgr, err := databasemgr.Build("mysql", mysqlCfg, loggerMgr, telemetryMgr)

// PostgreSQL
postgresqlCfg := map[string]any{
    "dsn": "host=localhost port=5432 user=postgres password=password dbname=mydb sslmode=disable",
}
dbMgr, err := databasemgr.Build("postgresql", postgresqlCfg, loggerMgr, telemetryMgr)

// SQLite
sqliteCfg := map[string]any{
    "dsn": "file:./cache.db?cache=shared&mode=rwc",
}
dbMgr, err := databasemgr.Build("sqlite", sqliteCfg, loggerMgr, telemetryMgr)

// None（空实现，用于测试）
dbMgr := databasemgr.NewDatabaseManagerNoneImpl(loggerMgr, telemetryMgr)
```

### 使用 BuildWithConfigProvider

从配置提供者读取配置，适合依赖注入场景：

```go
import "github.com/lite-lake/litecore-go/manager/configmgr"

provider, err := configmgr.Build("yaml", "config.yaml")
if err != nil {
    log.Fatal(err)
}

dbMgr, err := databasemgr.BuildWithConfigProvider(provider, loggerMgr, telemetryMgr)
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
dbMgr, err := databasemgr.NewDatabaseManagerMySQLImpl(mysqlCfg, loggerMgr, telemetryMgr)
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

databasemgr 内置了完整的可观测性功能，包括慢查询日志、SQL 日志、指标收集和链路追踪。

### 慢查询日志

当查询耗时超过 `slow_query_threshold` 时，自动记录慢查询日志：

```go
// 配置慢查询阈值为 1 秒
observability_config:
  slow_query_threshold: "1s"
```

慢查询日志示例：
```
2026-01-24 15:04:05.456 | WARN  | 慢查询检测 | operation=query table=users duration=1.2s threshold=1s
```

### SQL 日志

启用 `log_sql` 可以记录完整的 SQL 语句（生产环境建议关闭）：

```go
observability_config:
  log_sql: true
```

SQL 日志会自动进行脱敏处理，隐藏密码、token 等敏感信息：
```go
// 原始 SQL: SELECT * FROM users WHERE password = 'secret123'
// 脱敏后:   SELECT * FROM users WHERE password = '***'
```

支持的脱敏字段：
- password、pwd
- token
- secret
- api_key

### 指标收集

自动收集以下指标：

| 指标名称 | 类型 | 说明 |
|---------|------|------|
| `db.query.duration` | Histogram | 查询耗时（秒） |
| `db.query.count` | Counter | 查询计数 |
| `db.query.error_count` | Counter | 错误计数 |
| `db.query.slow_count` | Counter | 慢查询计数 |
| `db.transaction.count` | Counter | 事务计数 |
| `db.connection.pool` | Gauge | 连接池状态 |

指标包含以下属性：
- `operation` - 操作类型（query、create、update、delete）
- `table` - 表名
- `status` - 状态（success、error）

### 链路追踪

自动为所有数据库操作创建链路追踪 Span，支持 OpenTelemetry：
```go
Span 名称: db.{operation}
属性:
  - db.operation: 操作类型
  - db.table: 表名
```

### 采样率

通过 `sample_rate` 控制可观测性数据采集频率，减少性能开销：

```go
observability_config:
  sample_rate: 0.1  # 仅采集 10% 的数据
```

### 日志级别

不同级别的数据库操作使用不同的日志级别：

| 级别 | 场景 | 示例 |
|------|------|------|
| Debug | 正常操作成功 | `database operation success` |
| Warn | 慢查询 | `slow database query detected` |
| Error | 操作失败 | `database operation failed` |

### 可观测性配置示例

```yaml
database:
  driver: mysql
  mysql_config:
    dsn: "root:password@tcp(localhost:3306)/mydb"
  observability_config:
    slow_query_threshold: "1s"  # 超过 1 秒的查询记录为慢查询
    log_sql: false               # 生产环境建议关闭
    sample_rate: 1.0             # 采样率 100%
```

### 可观测性配置结构

```go
type ObservabilityConfig struct {
    SlowQueryThreshold time.Duration // 慢查询阈值，0 表示不记录慢查询
    LogSQL             bool          // 是否记录完整 SQL（生产环境建议关闭）
    SampleRate         float64       // 采样率 (0.0-1.0)
}
```

## 支持的数据库

### MySQL

```yaml
database:
  driver: mysql
  mysql_config:
    dsn: "user:password@tcp(localhost:3306)/dbname?charset=utf8mb4&parseTime=True&loc=Local"
    pool_config:
      max_open_conns: 100
      max_idle_conns: 10
      conn_max_lifetime: "30s"
      conn_max_idle_time: "5m"
  observability_config:
    slow_query_threshold: "1s"
    log_sql: false
    sample_rate: 1.0
```

### PostgreSQL

```yaml
database:
  driver: postgresql
  postgresql_config:
    dsn: "host=localhost port=5432 user=postgres password=password dbname=dbname sslmode=disable"
    pool_config:
      max_open_conns: 50
      max_idle_conns: 10
      conn_max_lifetime: "30s"
      conn_max_idle_time: "5m"
  observability_config:
    slow_query_threshold: "1s"
    log_sql: false
    sample_rate: 1.0
```

### SQLite

```yaml
database:
  driver: sqlite
  sqlite_config:
    dsn: "file:./data.db?cache=shared&mode=rwc"
    pool_config:
      max_open_conns: 1  # SQLite 通常设置为 1
      max_idle_conns: 1
      conn_max_lifetime: "30s"
      conn_max_idle_time: "5m"
  observability_config:
    slow_query_threshold: "1s"
    log_sql: false
    sample_rate: 1.0
```

### None（空实现）

```yaml
database:
  driver: none
```

用于测试场景，所有操作都是空操作。

## API

### 工厂函数

```go
func Build(driverType string, driverConfig map[string]any, loggerMgr loggermgr.ILoggerManager, telemetryMgr telemetrymgr.ITelemetryManager) (IDatabaseManager, error)
func BuildWithConfigProvider(configProvider configmgr.IConfigManager, loggerMgr loggermgr.ILoggerManager, telemetryMgr telemetrymgr.ITelemetryManager) (IDatabaseManager, error)
```

### 构造函数

```go
func NewDatabaseManagerMySQLImpl(cfg *MySQLConfig, loggerMgr loggermgr.ILoggerManager, telemetryMgr telemetrymgr.ITelemetryManager) (IDatabaseManager, error)
func NewDatabaseManagerPostgreSQLImpl(cfg *PostgreSQLConfig, loggerMgr loggermgr.ILoggerManager, telemetryMgr telemetrymgr.ITelemetryManager) (IDatabaseManager, error)
func NewDatabaseManagerSQLiteImpl(cfg *SQLiteConfig, loggerMgr loggermgr.ILoggerManager, telemetryMgr telemetrymgr.ITelemetryManager) (IDatabaseManager, error)
func NewDatabaseManagerNoneImpl(loggerMgr loggermgr.ILoggerManager, telemetryMgr telemetrymgr.ITelemetryManager) IDatabaseManager
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

### DatabaseConfig

```go
type DatabaseConfig struct {
    Driver              string               // 驱动类型: mysql, postgresql, sqlite, none
    SQLiteConfig        *SQLiteConfig        // SQLite 配置
    PostgreSQLConfig    *PostgreSQLConfig    // PostgreSQL 配置
    MySQLConfig         *MySQLConfig         // MySQL 配置
    ObservabilityConfig *ObservabilityConfig // 可观测性配置
    AutoMigrate         bool                 // 是否自动迁移数据库表结构
}
```

### PoolConfig

```go
type PoolConfig struct {
    MaxOpenConns    int           // 最大打开连接数，0 表示无限制
    MaxIdleConns    int           // 最大空闲连接数
    ConnMaxLifetime time.Duration // 连接最大存活时间
    ConnMaxIdleTime time.Duration // 连接最大空闲时间
}
```

### ObservabilityConfig

```go
type ObservabilityConfig struct {
    SlowQueryThreshold time.Duration // 慢查询阈值
    LogSQL             bool          // 是否记录完整 SQL（生产环境建议关闭）
    SampleRate         float64       // 采样率 (0.0-1.0)
}
```

### MySQLConfig

```go
type MySQLConfig struct {
    DSN        string      // MySQL DSN
    PoolConfig *PoolConfig // 连接池配置（可选）
}
```

### PostgreSQLConfig

```go
type PostgreSQLConfig struct {
    DSN        string      // PostgreSQL DSN
    PoolConfig *PoolConfig // 连接池配置（可选）
}
```

### SQLiteConfig

```go
type SQLiteConfig struct {
    DSN        string      // SQLite DSN
    PoolConfig *PoolConfig // 连接池配置（可选）
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

### 可观测性配置

- **生产环境**：关闭 `log_sql`，设置适当的 `slow_query_threshold`（如 1s）
- **开发环境**：开启 `log_sql`，设置较短的慢查询阈值（如 100ms）
- **高并发场景**：适当降低 `sample_rate`（如 0.1），减少性能开销

## 在 Repository 中使用

```go
type messageRepositoryImpl struct {
    Config  configmgr.IConfigManager     `inject:""`
    Manager databasemgr.IDatabaseManager `inject:""`
}

func (r *messageRepositoryImpl) GetByID(id uint) (*Message, error) {
    db := r.Manager.DB()
    var message Message
    err := db.First(&message, id).Error
    if err != nil {
        return nil, err
    }
    return &message, nil
}

func (r *messageRepositoryImpl) Create(message *Message) error {
    db := r.Manager.DB()
    return db.Create(message).Error
}

func (r *messageRepositoryImpl) GetApprovedMessages() ([]*Message, error) {
    db := r.Manager.DB()
    var messages []*Message
    err := db.Where("status = ?", "approved").
        Order("created_at DESC").
        Find(&messages).Error
    return messages, err
}
```
