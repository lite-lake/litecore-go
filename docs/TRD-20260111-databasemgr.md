# DatabaseManager 开发计划

## 1. 概述

DatabaseManager 是 LiteCore 框架的数据库管理组件，负责数据库连接的创建、管理和生命周期控制。支持 MySQL、PostgreSQL 和 SQLite 三种数据库驱动，提供统一的接口和配置方式。

## 2. 架构设计

### 2.1 整体结构

```
manager/databasemgr/
├── doc.go                    # 包文档
├── interface.go              # 数据库管理器接口定义
├── factory.go                # 工厂函数，用于创建管理器实例
├── database_adapter.go       # 数据库适配器
├── README.md                 # 使用文档
└── internal/
    ├── config/               # 配置解析和验证
    │   ├── config.go         # 配置结构体和解析逻辑
    │   └── config_test.go    # 配置测试
    └── drivers/              # 驱动实现
        ├── base_manager.go   # 基础管理器（实现 common.Manager）
        ├── base_manager_test.go
        ├── none_manager.go   # 空实现（降级方案）
        ├── none_manager_test.go
        ├── mysql_driver.go   # MySQL 驱动实现
        ├── mysql_driver_test.go
        ├── postgresql_driver.go  # PostgreSQL 驱动实现
        ├── postgresql_driver_test.go
        └── sqlite_driver.go   # SQLite 驱动实现
        └── sqlite_driver_test.go
```

### 2.2 接口设计

#### DatabaseManager 接口

```go
type DatabaseManager interface {
    // 继承 common.Manager 接口
    ManagerName() string
    Health() error
    OnStart() error
    OnStop() error

    // 获取数据库连接
    DB() *sql.DB

    // 获取数据库驱动类型
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

### 2.3 配置设计

#### DatabaseConfig 结构

```go
type DatabaseConfig struct {
    Driver          string            `yaml:"driver"`           // 驱动类型: mysql, postgresql, sqlite
    SQLiteConfig    *SQLiteConfig     `yaml:"sqlite_config"`    // SQLite 配置
    PostgreSQLConfig *PostgreSQLConfig `yaml:"postgresql_config"` // PostgreSQL 配置
    MySQLConfig     *MySQLConfig      `yaml:"mysql_config"`     // MySQL 配置
}

// PoolConfig 数据库连接池配置（所有驱动通用）
type PoolConfig struct {
    MaxOpenConns    int           `yaml:"max_open_conns"`     // 最大打开连接数，0 表示无限制
    MaxIdleConns    int           `yaml:"max_idle_conns"`     // 最大空闲连接数
    ConnMaxLifetime time.Duration `yaml:"conn_max_lifetime"`  // 连接最大存活时间
    ConnMaxIdleTime time.Duration `yaml:"conn_max_idle_time"` // 连接最大空闲时间
}

type SQLiteConfig struct {
    DSN        string      `yaml:"dsn"`         // SQLite DSN，如: file:./data.db?cache=shared&mode=rwc
    PoolConfig *PoolConfig `yaml:"pool_config"` // 连接池配置（可选）
}

type PostgreSQLConfig struct {
    DSN        string      `yaml:"dsn"`         // PostgreSQL DSN，如: host=localhost port=5432 user=postgres password=password dbname=lite_demo sslmode=disable
    PoolConfig *PoolConfig `yaml:"pool_config"` // 连接池配置（可选）
}

type MySQLConfig struct {
    DSN        string      `yaml:"dsn"`         // MySQL DSN，如: root:password@tcp(localhost:3306)/lite_demo?charset=utf8mb4&parseTime=True&loc=Local
    PoolConfig *PoolConfig `yaml:"pool_config"` // 连接池配置（可选）
}
```

## 3. 开发任务

### 3.1 第一阶段：基础框架

#### 任务 1.1：创建目录结构
- [ ] 创建 `manager/databasemgr/` 目录
- [ ] 创建 `internal/config/` 目录
- [ ] 创建 `internal/drivers/` 目录

#### 任务 1.2：实现接口定义
- [ ] 创建 `interface.go`，定义 DatabaseManager 接口
- [ ] 创建 `doc.go`，编写包文档

#### 任务 1.3：实现配置解析
- [ ] 创建 `internal/config/config.go`
  - 定义 DatabaseConfig 结构体
  - 定义 SQLiteConfig、PostgreSQLConfig、MySQLConfig 结构体
  - 实现 ParseDatabaseConfigFromMap() 函数
  - 实现 Validate() 方法
- [ ] 创建 `internal/config/config_test.go`
  - 测试配置解析
  - 测试配置验证
  - 测试默认值

#### 任务 1.4：实现基础管理器
- [ ] 创建 `internal/drivers/base_manager.go`
  - 实现 common.Manager 接口
  - 提供公共方法
- [ ] 创建 `internal/drivers/base_manager_test.go`

### 3.2 第二阶段：驱动实现

#### 任务 2.1：实现 None 驱动
- [ ] 创建 `internal/drivers/none_manager.go`
  - 空实现，用于降级场景
  - 所有操作返回错误或空值
- [ ] 创建 `internal/drivers/none_manager_test.go`

#### 任务 2.2：实现 SQLite 驱动
- [ ] 创建 `internal/drivers/sqlite_driver.go`
  - 实现数据库连接创建
  - 实现连接池配置
  - 实现 health check
- [ ] 创建 `internal/drivers/sqlite_driver_test.go`
  - 测试连接创建
  - 测试查询操作
  - 测试事务操作

#### 任务 2.3：实现 PostgreSQL 驱动
- [ ] 创建 `internal/drivers/postgresql_driver.go`
  - 实现数据库连接创建
  - 实现连接池配置
  - 实现 health check
- [ ] 创建 `internal/drivers/postgresql_driver_test.go`
  - 测试连接创建
  - 测试查询操作
  - 测试事务操作

#### 任务 2.4：实现 MySQL 驱动
- [ ] 创建 `internal/drivers/mysql_driver.go`
  - 实现数据库连接创建
  - 实现连接池配置
  - 实现 health check
- [ ] 创建 `internal/drivers/mysql_driver_test.go`
  - 测试连接创建
  - 测试查询操作
  - 测试事务操作

### 3.3 第三阶段：工厂和适配器

#### 任务 3.1：实现工厂函数
- [ ] 创建 `factory.go`
  - 实现 Build() 函数
  - 实现 BuildWithConfig() 函数
  - 实现降级逻辑（失败时返回 none 驱动）

#### 任务 3.2：实现适配器
- [ ] 创建 `database_adapter.go`
  - 将内部驱动适配到 DatabaseManager 接口
  - 实现 common.Manager 接口方法
  - 实现 DatabaseManager 特有方法

### 3.4 第四阶段：文档和测试

#### 任务 4.1：编写使用文档
- [ ] 创建 `README.md`
  - 快速开始
  - 配置说明
  - 使用示例
  - 最佳实践

#### 任务 4.2：集成测试
- [ ] 创建 `integration_test.go`
  - 测试完整的初始化流程
  - 测试配置加载
  - 测试多驱动切换
  - 测试降级场景

#### 任务 4.3：性能测试（可选）
- [ ] 创建 `benchmark_test.go`
  - 测试连接池性能
  - 测试查询性能
  - 测试并发性能

## 4. 技术细节

### 4.1 依赖库

```go
import (
    "database/sql"
    "context"
    "time"

    // MySQL 驱动
    _ "github.com/go-sql-driver/mysql"

    // PostgreSQL 驱动
    _ "github.com/lib/pq"

    // SQLite 驱动
    _ "github.com/mattn/go-sqlite3"

    "com.litelake.litecore/common"
)
```

### 4.2 数据源字符串（DSN）格式

#### MySQL DSN
```
username:password@protocol(address)/dbname?param=value
```
示例：`root:password@tcp(localhost:3306)/lite_demo?charset=utf8mb4&parseTime=True&loc=Local`

#### PostgreSQL DSN
```
host=port user=password dbname=sslmode=connect_timeout=
```
示例：`host=localhost port=5432 user=postgres password=password dbname=lite_demo sslmode=disable`

#### SQLite DSN
```
file:path?param=value
```
示例：`file:./data.db?cache=shared&mode=rwc`

### 4.3 连接池配置

默认值：
- MaxOpenConns: 10
- MaxIdleConns: 5
- ConnMaxLifetime: 30s
- ConnMaxIdleTime: 5m

### 4.4 错误处理策略

1. **配置解析失败**：返回 none 驱动，记录错误日志
2. **连接创建失败**：返回 none 驱动，记录错误日志
3. **查询执行失败**：返回原始错误，由调用方处理
4. **Health check 失败**：返回错误，但不影响管理器运行

### 4.5 降级策略

当以下情况发生时，自动降级到 none 驱动：
- 配置解析失败
- 驱动初始化失败
- 连接创建失败

none 驱动行为：
- 所有 DB 操作返回 "database not available" 错误
- Health() 返回错误
- OnStart() 和 OnStop() 无操作

## 5. 测试策略

### 5.1 单元测试

每个驱动都需要独立的单元测试：
- 配置解析测试
- 连接创建测试
- 基本操作测试
- 错误处理测试

### 5.2 集成测试

- 测试完整初始化流程
- 测试多驱动场景
- 测试降级机制
- 测试生命周期管理

### 5.3 测试数据库

使用内存数据库进行测试：
- SQLite: `file::memory:?cache=shared`
- PostgreSQL: 使用 testcontainers（可选）
- MySQL: 使用 testcontainers（可选）

## 6. 使用示例

### 6.1 配置示例

#### MySQL 配置示例

```yaml
database:
  driver: mysql
  mysql_config:
    dsn: "root:password@tcp(localhost:3306)/lite_demo?charset=utf8mb4&parseTime=True&loc=Local&timeout=10s&readTimeout=30s"
    pool_config:
      max_open_conns: 10
      max_idle_conns: 5
      conn_max_lifetime: 30s
      conn_max_idle_time: 5m
```

#### PostgreSQL 配置示例

```yaml
database:
  driver: postgresql
  postgresql_config:
    dsn: "host=localhost port=5432 user=postgres password=password dbname=lite_demo sslmode=disable connect_timeout=10"
    pool_config:
      max_open_conns: 10
      max_idle_conns: 5
      conn_max_lifetime: 30s
      conn_max_idle_time: 5m
```

#### SQLite 配置示例

```yaml
database:
  driver: sqlite
  sqlite_config:
    dsn: "file:./data.db?cache=shared&mode=rwc"
    pool_config:
      max_open_conns: 1  # SQLite 通常设置为 1
```

### 6.2 代码示例

```go
// 从配置创建数据库管理器
cfg := loadConfig() // 从 YAML 加载配置
dbMgr := databasemgr.Build(cfg["database"].(map[string]any))

// 获取数据库连接
db := dbMgr.DB()

// 执行查询
rows, err := db.QueryContext(ctx, "SELECT * FROM users WHERE id = ?", userID)

// 使用事务
tx, err := db.BeginTx(ctx, nil)
if err != nil {
    return err
}
defer tx.Rollback()

_, err = tx.Exec("INSERT INTO users (name) VALUES (?)", "John")
if err != nil {
    return err
}

err = tx.Commit()
if err != nil {
    return err
}
```

## 7. 开发里程碑

| 阶段 | 任务 | 预计时间 |
|------|------|----------|
| 第一阶段 | 基础框架 | 1-2 天 |
| 第二阶段 | 驱动实现 | 3-4 天 |
| 第三阶段 | 工厂和适配器 | 1-2 天 |
| 第四阶段 | 文档和测试 | 2-3 天 |

**总计**：7-11 天

## 8. 注意事项

1. **安全性**：
   - 不要在日志中打印密码
   - 使用环境变量或配置管理工具管理敏感信息
   - 支持 SSL/TLS 连接

2. **性能**：
   - 合理配置连接池大小
   - 使用连接池复用连接
   - 避免长事务

3. **兼容性**：
   - 支持多种数据库版本
   - 处理不同数据库的方言差异
   - 提供统一的错误处理

4. **可观测性**：
   - 记录连接池状态
   - 记录慢查询
   - 集成 telemetryMgr 进行追踪

## 9. 后续扩展

- [ ] 支持读写分离
- [ ] 支持分库分表
- [ ] 支持 ORM 集成（GORM、XORM）
- [ ] 支持数据库迁移工具
- [ ] 支持连接池监控指标
- [ ] 支持查询构建器