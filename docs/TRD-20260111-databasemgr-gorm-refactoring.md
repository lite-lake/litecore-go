# DatabaseMgr GORM 改造计划（无向后兼容版本）

## 1. 改造背景

### 1.1 当前状况分析

当前 DatabaseMgr 基于 `database/sql` 标准库实现，提供以下功能：
- ✅ 支持多种数据库驱动（MySQL、PostgreSQL、SQLite）
- ✅ 连接池管理
- ✅ 健康检查
- ✅ 事务支持
- ✅ 生命周期管理
- ❌ **缺少 ORM 能力**：需要手写 SQL，容易出错且不安全
- ❌ **缺少模型定义**：没有结构化模型和关联关系管理
- ❌ **缺少钩子机制**：没有 BeforeCreate/AfterUpdate 等生命周期钩子
- ❌ **缺少自动迁移**：需要手动管理数据库 schema
- ❌ **缺少查询构建器**：复杂查询需要拼接 SQL 字符串

### 1.2 为什么选择 GORM？

**GORM** 是 Go 语言最流行的 ORM 框架，具有以下优势：

1. **全功能 ORM**
   - 自动模型定义和关联关系（Has One、Has Many、Many To Many、Belongs To）
   - 强大的查询构建器（链式调用、条件组合、预加载）
   - 自动创建和迁移表结构

2. **开发效率**
   - 减少手写 SQL 的工作量
   - 类型安全的查询 API
   - 丰富的钩子机制（BeforeCreate、AfterCreate、BeforeUpdate、AfterUpdate 等）

3. **企业级特性**
   - 事务支持（嵌套事务、保存点）
   - 乐观锁、复合主键
   - 多数据库支持（MySQL、PostgreSQL、SQLite、SQL Server）
   - 性能优化（批量插入、预加载、懒加载）

4. **生态成熟**
   - 活跃的社区维护
   - 丰富的插件生态
   - 详尽的文档和示例

5. **行业标准**
   - Go 社区事实标准的 ORM 框架
   - 被众多大型项目采用
   - 持续更新和维护

### 1.3 改造目标

**彻底改造** DatabaseMgr，完全基于 GORM 实现：

1. ✅ **移除 database/sql 依赖**：完全使用 GORM API
2. ✅ **简化接口设计**：直接暴露 GORM 功能，无需适配层
3. ✅ **增强功能**：利用 GORM 特性，提供强大的数据库操作能力
4. ✅ **保持架构**：延续现有的 Manager 架构和生命周期管理

---

## 2. 架构设计

### 2.1 整体架构

```
┌─────────────────────────────────────────────────────┐
│           DatabaseManager (Public Interface)         │
│  - DB() *gorm.DB                                      │
│  - Model(model any) *gorm.DB                          │
│  - Table(name string) *gorm.DB                        │
│  - Transaction(fn func(*gorm.DB) error) error         │
│  - AutoMigrate(models ...any) error                   │
│  - WithContext(ctx context.Context) *gorm.DB          │
│  - Exec(sql string, values ...any) result.Result      │
│  - Raw(sql string, values ...any) *gorm.DB            │
└─────────────────────────────────────────────────────┘
                         ↓
┌─────────────────────────────────────────────────────┐
│              Drivers (Internal)                       │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐  │
│  │MySQLManager │  │PostgreSQL   │  │  SQLite     │  │
│  │             │  │Manager      │  │ Manager     │  │
│  │- gormDB     │  │- gormDB     │  │- gormDB     │  │
│  └─────────────┘  └─────────────┘  └─────────────┘  │
└─────────────────────────────────────────────────────┘
                         ↓
┌─────────────────────────────────────────────────────┐
│                 GORM Library                          │
│  - gorm.io/gorm                                       │
│  - gorm.io/driver/mysql                               │
│  - gorm.io/driver/postgres                            │
│  - gorm.io/driver/sqlite                              │
└─────────────────────────────────────────────────────┘
```

### 2.2 接口设计

#### 2.2.1 DatabaseManager 接口（GORM 原生版）

```go
package databasemgr

import (
    "context"
    "gorm.io/gorm"
)

// DatabaseManager 数据库管理器接口（完全基于 GORM）
type DatabaseManager interface {
    // ========== 生命周期方法 ==========
    // ManagerName 返回管理器名称
    ManagerName() string

    // Health 检查管理器健康状态
    Health() error

    // OnStart 在服务器启动时触发
    OnStart() error

    // OnStop 在服务器停止时触发
    OnStop() error

    // ========== GORM 核心方法 ==========
    // DB 获取 GORM 数据库实例
    DB() *gorm.DB

    // Model 指定模型进行操作
    Model(model any) *gorm.DB

    // Table 指定表名进行操作
    Table(name string) *gorm.DB

    // WithContext 设置上下文
    WithContext(ctx context.Context) *gorm.DB

    // ========== 事务管理 ==========
    // Transaction 执行事务
    Transaction(fn func(*gorm.DB) error, opts ...*interface{}) error

    // Begin 开启事务
    Begin(opts ...*interface{}) *gorm.DB

    // ========== 迁移管理 ==========
    // AutoMigrate 自动迁移
    AutoMigrate(models ...any) error

    // Migrator 获取迁移器
    Migrator() gorm.Migrator

    // ========== 连接管理 ==========
    // Driver 获取数据库驱动类型
    Driver() string

    // Ping 检查数据库连接
    Ping(ctx context.Context) error

    // Stats 获取连接池统计信息
    Stats() gorm.SQLStats

    // Close 关闭数据库连接
    Close() error

    // ========== 原生 SQL 支持 ==========
    // Exec 执行原生 SQL
    Exec(sql string, values ...any) *gorm.DB

    // Raw 执行原生查询
    Raw(sql string, values ...any) *gorm.DB
}
```

#### 2.2.2 ExtendedDatabaseManager 扩展接口（可选）

```go
// ExtendedDatabaseManager GORM 数据库管理器扩展接口
// 提供更高级的 GORM 特定功能
type ExtendedDatabaseManager interface {
    DatabaseManager

    // ========== 高级查询 ==========
    // Preload 预加载关联关系
    Preload(query string, args ...any) *gorm.DB

    // Joins 关联查询
    Joins(query string, args ...any) *gorm.DB

    // Scopes 应用查询作用域
    Scopes(funcs ...func(*gorm.DB) *gorm.DB) *gorm.DB

    // Omit 排除字段
    Omit(fields ...string) *gorm.DB

    // Select 选择字段
    Select(query interface{}, args ...any) *gorm.DB

    // ========== 批量操作 ==========
    // CreateInBatches 批量创建
    CreateInBatches(value any, batchSize int) *gorm.DB

    // Save 保存（包含更新）
    Save(value any) *gorm.DB

    // Updates 更新
    Updates(values interface{}) *gorm.DB

    // Update 更新单个字段
    Update(column string, value interface{}) *gorm.DB

    // Delete 删除
    Delete(value any, conds ...interface{}) *gorm.DB

    // ========== 钩子管理 ==========
    // Callback 返回回调管理器
    Callback() *callback

    // AddHook 添加全局钩子
    AddHook(name string, hook interface{})

    // ========== 配置管理 ==========
    // Config 获取 GORM 配置
    Config() *gorm.Config

    // Debug 开启调试模式
    Debug() *gorm.DB

    // ============================================================
    // 说明：以上方法本质上是 *gorm.DB 的代理
    // 用户可以直接使用 dbMgr.DB() 获取 *gorm.DB 实例
    // 然后调用所有 GORM 方法
    // ============================================================
}
```

### 2.3 核心数据模型设计示例

#### 2.3.1 基础模型

```go
package models

import (
    "time"
    "gorm.io/gorm"
)

// BaseModel 基础模型（包含公共字段）
type BaseModel struct {
    ID        uint           `gorm:"primarykey" json:"id"`
    CreatedAt time.Time      `json:"created_at"`
    UpdatedAt time.Time      `json:"updated_at"`
    DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
}

// User 用户模型
type User struct {
    BaseModel
    Username string  `gorm:"type:varchar(50);uniqueIndex;not null;comment:用户名" json:"username"`
    Email    string  `gorm:"type:varchar(100);uniqueIndex;not null;comment:邮箱" json:"email"`
    Age      int     `gorm:"index;comment:年龄" json:"age"`

    // 关联关系
    Profile  *Profile  `gorm:"foreignKey:UserID;comment:用户资料" json:"profile,omitempty"`
    Orders   []Order   `gorm:"foreignKey:UserID;comment:用户订单" json:"orders,omitempty"`
    Roles    []Role    `gorm:"many2many:user_roles;comment:用户角色" json:"roles,omitempty"`
}

// Profile 用户资料模型
type Profile struct {
    BaseModel
    UserID   uint   `gorm:"uniqueIndex;not null;comment:用户ID" json:"user_id"`
    Avatar   string `gorm:"type:varchar(255);comment:头像" json:"avatar"`
    Bio      string `gorm:"type:text;comment:个人简介" json:"bio"`
    Location string `gorm:"type:varchar(100);comment:所在地" json:"location"`
}

// Order 订单模型
type Order struct {
    BaseModel
    UserID     uint    `gorm:"index;not null;comment:用户ID" json:"user_id"`
    OrderNo    string  `gorm:"type:varchar(50);uniqueIndex;not null;comment:订单号" json:"order_no"`
    TotalPrice float64 `gorm:"type:decimal(10,2);comment:订单总价" json:"total_price"`
    Status     string  `gorm:"type:varchar(20);index;comment:订单状态" json:"status"`
    Remark     string  `gorm:"type:text;comment:备注" json:"remark"`

    // 关联关系
    Items  []OrderItem `gorm:"foreignKey:OrderID;comment:订单项" json:"items,omitempty"`
}

// OrderItem 订单项模型
type OrderItem struct {
    BaseModel
    OrderID   uint    `gorm:"index;not null;comment:订单ID" json:"order_id"`
    ProductID uint    `gorm:"index;not null;comment:商品ID" json:"product_id"`
    Quantity  int     `gorm:"comment:数量" json:"quantity"`
    Price     float64 `gorm:"type:decimal(10,2);comment:单价" json:"price"`
}

// Role 角色模型
type Role struct {
    BaseModel
    Name        string `gorm:"type:varchar(50);uniqueIndex;not null;comment:角色名" json:"name"`
    Description string `gorm:"type:text;comment:角色描述" json:"description"`
}

// ========== 钩子方法 ==========

// BeforeCreate 创建前钩子
func (u *User) BeforeCreate(tx *gorm.DB) error {
    // 验证年龄
    if u.Age < 0 {
        return tx.AddError(errors.New("age cannot be negative"))
    }

    // 验证邮箱格式
    if !isValidEmail(u.Email) {
        return tx.AddError(errors.New("invalid email format"))
    }

    return nil
}

// AfterCreate 创建后钩子
func (u *User) AfterCreate(tx *gorm.DB) error {
    // 记录日志或发送欢迎邮件
    log.Printf("User %s created with ID %d", u.Username, u.ID)
    return nil
}

// BeforeUpdate 更新前钩子
func (u *User) BeforeUpdate(tx *gorm.DB) error {
    // 防止修改已删除的用户
    if u.DeletedAt.Valid {
        return tx.AddError(errors.New("cannot update deleted user"))
    }
    return nil
}

// AfterFind 查询后钩子
func (u *User) AfterFind(tx *gorm.DB) error {
    // 自动加载某些计算字段
    return nil
}
```

---

## 3. 改造方案

### 3.1 第一阶段：GORM 集成（核心改造）

#### 任务 1.1：清理旧依赖

- [ ] 移除 `database/sql` 相关依赖（可选）
  - `github.com/go-sql-driver/mysql`
  - `github.com/lib/pq`
  - `github.com/mattn/go-sqlite3`

- [ ] 添加 GORM 依赖到 `go.mod`
  ```go
  require (
      gorm.io/gorm v1.25.12
      gorm.io/driver/mysql v1.5.7
      gorm.io/driver/postgres v1.5.9
      gorm.io/driver/sqlite v1.5.7
  )
  ```

#### 任务 1.2：创建 GORM 基础管理器

- [ ] 创建 `internal/drivers/gorm_base_manager.go`
  - 实现 GORM 数据库连接创建
  - 实现 GORM 配置管理
  - 提供生命周期管理（OnStart、OnStop）
  - 实现健康检查
- [ ] 创建 `internal/drivers/gorm_base_manager_test.go`

**文件结构：**
```go
// internal/drivers/gorm_base_manager.go
package drivers

import (
    "context"
    "fmt"
    "sync"
    "time"

    "gorm.io/gorm"
    "gorm.io/gorm/logger"

    "com.litelake.litecore/common"
)

// GormBaseManager GORM 基础管理器
type GormBaseManager struct {
    name   string
    driver string
    db     *gorm.DB
    sqlDB  *sql.DB  // 用于连接池管理
    mu     sync.RWMutex
}

// NewGormBaseManager 创建 GORM 基础管理器
func NewGormBaseManager(name, driver string, db *gorm.DB) *GormBaseManager {
    sqlDB, _ := db.DB()
    return &GormBaseManager{
        name:   name,
        driver: driver,
        db:     db,
        sqlDB:  sqlDB,
    }
}

// ManagerName 返回管理器名称
func (m *GormBaseManager) ManagerName() string {
    return m.name
}

// DB 获取 GORM 数据库实例
func (m *GormBaseManager) DB() *gorm.DB {
    m.mu.RLock()
    defer m.mu.RUnlock()
    return m.db
}

// Driver 获取驱动类型
func (m *GormBaseManager) Driver() string {
    return m.driver
}

// Ping 检查数据库连接
func (m *GormBaseManager) Ping(ctx context.Context) error {
    sqlDB, err := m.db.DB()
    if err != nil {
        return err
    }
    return sqlDB.PingContext(ctx)
}

// Health 检查健康状态
func (m *GormBaseManager) Health() error {
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()
    return m.Ping(ctx)
}

// Stats 获取连接池统计信息
func (m *GormBaseManager) Stats() gorm.SQLStats {
    sqlDB, err := m.db.DB()
    if err != nil {
        return gorm.SQLStats{}
    }
    return sqlDB.Stats()
}

// Close 关闭数据库连接
func (m *GormBaseManager) Close() error {
    m.mu.Lock()
    defer m.mu.Unlock()

    sqlDB, err := m.db.DB()
    if err != nil {
        return err
    }

    if sqlDB != nil {
        return sqlDB.Close()
    }

    return nil
}

// OnStart 启动时的初始化
func (m *GormBaseManager) OnStart() error {
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()

    if err := m.Ping(ctx); err != nil {
        return fmt.Errorf("failed to ping database: %w", err)
    }

    return nil
}

// OnStop 停止时的清理
func (m *GormBaseManager) OnStop() error {
    return m.Close()
}
```

#### 任务 1.3：重写 MySQL 驱动

- [ ] 删除旧的 `internal/drivers/mysql_driver.go`
- [ ] 创建新的 `internal/drivers/mysql_driver.go`
  - 使用 `gorm.Open(mysql.Open(dsn))`
  - 实现 GORM 配置
  - 连接池配置

**新实现：**
```go
// internal/drivers/mysql_driver.go
package drivers

import (
    "fmt"

    "gorm.io/driver/mysql"
    "gorm.io/gorm"
    "gorm.io/gorm/logger"

    "com.litelake.litecore/manager/databasemgr/internal/config"
)

// MySQLManager MySQL 数据库管理器
type MySQLManager struct {
    *GormBaseManager
    config *config.DatabaseConfig
}

// NewMySQLManager 创建 MySQL 数据库管理器
func NewMySQLManager(cfg *config.DatabaseConfig) (*MySQLManager, error) {
    if err := cfg.Validate(); err != nil {
        return nil, fmt.Errorf("invalid database config: %w", err)
    }

    if cfg.MySQLConfig == nil {
        return nil, fmt.Errorf("mysql_config is required")
    }

    // GORM 配置
    gormConfig := &gorm.Config{
        SkipDefaultTransaction: true,
        Logger:                 logger.Default.LogMode(logger.Silent),
    }

    // 打开数据库连接
    db, err := gorm.Open(mysql.Open(cfg.MySQLConfig.DSN), gormConfig)
    if err != nil {
        return nil, fmt.Errorf("failed to open mysql database: %w", err)
    }

    // 配置连接池
    sqlDB, err := db.DB()
    if err != nil {
        return nil, fmt.Errorf("failed to get sql.DB: %w", err)
    }

    if cfg.MySQLConfig.PoolConfig != nil {
        sqlDB.SetMaxOpenConns(cfg.MySQLConfig.PoolConfig.MaxOpenConns)
        sqlDB.SetMaxIdleConns(cfg.MySQLConfig.PoolConfig.MaxIdleConns)
        sqlDB.SetConnMaxLifetime(cfg.MySQLConfig.PoolConfig.ConnMaxLifetime)
        sqlDB.SetConnMaxIdleTime(cfg.MySQLConfig.PoolConfig.ConnMaxIdleTime)
    }

    baseMgr := NewGormBaseManager("mysql-database", "mysql", db)

    return &MySQLManager{
        GormBaseManager: baseMgr,
        config:          cfg,
    }, nil
}
```

#### 任务 1.4：重写 PostgreSQL 驱动

- [ ] 删除旧的 `internal/drivers/postgresql_driver.go`
- [ ] 创建新的 `internal/drivers/postgresql_driver.go`
  - 使用 `gorm.Open(postgres.Open(dsn))`

#### 任务 1.5：重写 SQLite 驱动

- [ ] 删除旧的 `internal/drivers/sqlite_driver.go`
- [ ] 创建新的 `internal/drivers/sqlite_driver.go`
  - 使用 `gorm.Open(sqlite.Open(dsn))`

#### 任务 1.6：重写 None 驱动

- [ ] 删除旧的 `internal/drivers/none_manager.go`
- [ ] 创建新的 `internal/drivers/none_manager.go`
  - 使用 GORM 的Dialector 实现

### 3.2 第二阶段：接口重构（简化设计）

#### 任务 2.1：重写 DatabaseManager 接口

- [ ] 删除旧的 `interface.go`
- [ ] 创建新的 `interface.go`
  - 完全基于 GORM 设计
  - 移除所有 `database/sql` 相关方法
  - 直接暴露 GORM 功能

```go
// manager/databasemgr/interface.go
package databasemgr

import (
    "context"
    "gorm.io/gorm"
)

// DatabaseManager 数据库管理器接口
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
    Transaction(fn func(*gorm.DB) error, opts ...*interface{}) error
    Begin(opts ...*interface{}) *gorm.DB

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

#### 任务 2.2：删除适配器

- [ ] **删除** `database_adapter.go`
- [ ] 直接在驱动中实现 DatabaseManager 接口

```go
// MySQLManager 直接实现 DatabaseManager 接口
var _ DatabaseManager = (*MySQLManager)(nil)
```

#### 任务 2.3：简化工厂模式

- [ ] 重写 `factory.go`
  - 移除适配器逻辑
  - 直接返回实现了 DatabaseManager 接口的管理器

```go
// manager/databasemgr/factory.go
package databasemgr

import (
    "com.litelake.litecore/common"
    "com.litelake.litecore/manager/databasemgr/internal/config"
    "com.litelake.litecore/manager/databasemgr/internal/drivers"
)

// Factory 数据库管理器工厂
type Factory struct{}

// NewFactory 创建数据库管理器工厂
func NewFactory() *Factory {
    return &Factory{}
}

// Build 创建数据库管理器实例
func (f *Factory) Build(driver string, cfg map[string]any) common.Manager {
    databaseConfig, err := config.ParseDatabaseConfigFromMap(cfg)
    if err != nil {
        return drivers.NewNoneDatabaseManager()
    }

    if driver != "" {
        databaseConfig.Driver = driver
    }

    if err := databaseConfig.Validate(); err != nil {
        return drivers.NewNoneDatabaseManager()
    }

    switch databaseConfig.Driver {
    case "mysql":
        mgr, err := drivers.NewMySQLManager(databaseConfig)
        if err != nil {
            return drivers.NewNoneDatabaseManager()
        }
        return mgr

    case "postgresql":
        mgr, err := drivers.NewPostgreSQLManager(databaseConfig)
        if err != nil {
            return drivers.NewNoneDatabaseManager()
        }
        return mgr

    case "sqlite":
        mgr, err := drivers.NewSQLiteManager(databaseConfig)
        if err != nil {
            return drivers.NewNoneDatabaseManager()
        }
        return mgr

    case "none":
        return drivers.NewNoneDatabaseManager()

    default:
        return drivers.NewNoneDatabaseManager()
    }
}

// BuildWithConfig 使用配置结构体创建
func (f *Factory) BuildWithConfig(databaseConfig *config.DatabaseConfig) (DatabaseManager, error) {
    if err := databaseConfig.Validate(); err != nil {
        return nil, err
    }

    switch databaseConfig.Driver {
    case "mysql":
        return drivers.NewMySQLManager(databaseConfig)
    case "postgresql":
        return drivers.NewPostgreSQLManager(databaseConfig)
    case "sqlite":
        return drivers.NewSQLiteManager(databaseConfig)
    case "none":
        return drivers.NewNoneDatabaseManager()
    default:
        return drivers.NewNoneDatabaseManager(), nil
    }
}
```

### 3.3 第三阶段：高级特性（增强功能）

#### 任务 3.1：实现钩子管理器

- [ ] 创建 `internal/hooks/manager.go`
  - 提供全局钩子注册
  - 支持模型级别钩子
  - 钩子执行顺序控制

```go
// internal/hooks/manager.go
package hooks

import (
    "gorm.io/gorm"
)

// Manager 钩子管理器
type Manager struct {
    callbacks map[string][]interface{}
}

// NewManager 创建钩子管理器
func NewManager() *Manager {
    return &Manager{
        callbacks: make(map[string][]interface{}),
    }
}

// Register 注册钩子
func (m *Manager) Register(name string, hook interface{}) {
    if m.callbacks[name] == nil {
        m.callbacks[name] = []interface{}{}
    }
    m.callbacks[name] = append(m.callbacks[name], hook)
}

// ApplyTo 应用钩子到 GORM
func (m *Manager) ApplyTo(db *gorm.DB) {
    // 应用钩子逻辑
}
```

#### 任务 3.2：实现自动迁移

- [ ] 创建 `internal/migration/migrator.go`
  - 实现 `AutoMigrate()` 功能
  - 支持版本化迁移（可选）
  - 迁移回滚支持（可选）

```go
// internal/migration/migrator.go
package migration

import (
    "fmt"

    "gorm.io/gorm"
)

// Migrator 迁移管理器
type Migrator struct {
    db *gorm.DB
}

// NewMigrator 创建迁移管理器
func NewMigrator(db *gorm.DB) *Migrator {
    return &Migrator{db: db}
}

// AutoMigrate 自动迁移
func (m *Migrator) AutoMigrate(models ...interface{}) error {
    return m.db.AutoMigrate(models...)
}

// CreateTables 创建表
func (m *Migrator) CreateTables(models ...interface{}) error {
    return m.db.Migrator().CreateTable(models...)
}

// DropTables 删除表
func (m *Migrator) DropTables(models ...interface{}) error {
    return m.db.Migrator().DropTable(models...)
}

// RenameTable 重命名表
func (m *Migrator) RenameTable(oldName, newName string) error {
    return m.db.Migrator().RenameTable(oldName, newName)

// AddColumn 添加列
func (m *Migrator) AddColumn(model interface{}, field string) error {
    return m.db.Migrator().AddColumn(model, field)
}

// DropColumn 删除列
func (m *Migrator) DropColumn(model interface{}, field string) error {
    return m.db.Migrator().DropColumn(model, field)
}

// AlterColumn 修改列
func (m *Migrator) AlterColumn(model interface{}, field string) error {
    return m.db.Migrator().AlterColumn(model, field)
}
```

#### 任务 3.3：实现事务管理器

- [ ] 创建 `internal/transaction/manager.go`
  - 实现嵌套事务
  - 实现保存点事务
  - 事务传播行为

```go
// internal/transaction/manager.go
package transaction

import (
    "gorm.io/gorm"
)

// Manager 事务管理器
type Manager struct {
    db *gorm.DB
}

// NewManager 创建事务管理器
func NewManager(db *gorm.DB) *Manager {
    return &Manager{db: db}
}

// Transaction 执行事务
func (m *Manager) Transaction(fn func(*gorm.DB) error) error {
    return m.db.Transaction(fn)
}

// Begin 开启事务
func (m *Manager) Begin() *gorm.DB {
    return m.db.Begin()
}

// BeginTx 开启事务（带选项）
func (m *Manager) BeginTx(opts ...*interface{}) *gorm.DB {
    return m.db.Begin(opts...)
}

// NestedTransaction 嵌套事务
func (m *Manager) NestedTransaction(fn func(*gorm.DB) error) error {
    return m.db.Transaction(func(tx *gorm.DB) error {
        // 外层事务
        return tx.Transaction(func(tx2 *gorm.DB) error {
            // 内层事务（保存点）
            return fn(tx2)
        })
    })
}
```

#### 任务 3.4：实现查询构建器（可选）

- [ ] 创建 `internal/builder/query_builder.go`
  - 提供链式查询构建器
  - 支持复杂条件组合
  - 类型安全的查询构建

```go
// internal/builder/query_builder.go
package builder

import (
    "gorm.io/gorm"
)

// Builder 查询构建器
type Builder struct {
    db *gorm.DB
}

// NewBuilder 创建查询构建器
func NewBuilder(db *gorm.DB) *Builder {
    return &Builder{db: db}
}

// Where 条件
func (b *Builder) Where(query interface{}, args ...interface{}) *Builder {
    b.db = b.db.Where(query, args...)
    return b
}

// Or 条件
func (b *Builder) Or(query interface{}, args ...interface{}) *Builder {
    b.db = b.db.Or(query, args...)
    return b
}

// Order 排序
func (b *Builder) Order(value interface{}) *Builder {
    b.db = b.db.Order(value)
    return b
}

// Limit 限制数量
func (b *Builder) Limit(limit int) *Builder {
    b.db = b.db.Limit(limit)
    return b
}

// Offset 偏移量
func (b *Builder) Offset(offset int) *Builder {
    b.db = b.db.Offset(offset)
    return b
}

// Find 查询
func (b *Builder) Find(dest interface{}) *gorm.DB {
    return b.db.Find(dest)
}

// First 查询第一条
func (b *Builder) First(dest interface{}) *gorm.DB {
    return b.db.First(dest)
}

// Build 构建 GORM DB
func (b *Builder) Build() *gorm.DB {
    return b.db
}
```

### 3.4 第四阶段：文档和测试（完善生态）

#### 任务 4.1：更新文档

- [ ] 更新 `docs/TRD-20260111-databasemgr.md`
  - 更新架构设计
  - 移除 database/sql 相关内容
  - 添加 GORM 特性说明

- [ ] 重写 `README.md`
  - 完全基于 GORM 的使用示例
  - 更新快速开始指南
  - 添加模型定义示例
  - 添加查询示例
  - 添加事务示例
  - 添加关联关系示例

#### 任务 4.2：单元测试

- [ ] 更新 `internal/drivers/*_test.go`
  - 测试 GORM 连接创建
  - 测试 GORM 查询操作
  - 测试 GORM 事务操作
  - 测试钩子机制

- [ ] 创建 `manager/databasemgr/gorm_test.go`
  - 测试 DatabaseManager 接口
  - 测试工厂方法
  - 测试降级逻辑

#### 任务 4.3：集成测试

- [ ] 创建 `integration_test.go`
  - 测试完整的 CRUD 操作
  - 测试模型关联
  - 测试复杂查询
  - 测试事务操作
  - 测试自动迁移

#### 任务 4.4：使用示例

- [ ] 创建 `examples/` 目录
  - `basic_model.go` - 基础模型示例
  - `crud_operations.go` - CRUD 操作示例
  - `associations.go` - 关联关系示例
  - `transactions.go` - 事务示例
  - `migrations.go` - 迁移示例
  - `hooks.go` - 钩子示例
  - `complex_queries.go` - 复杂查询示例

#### 任务 4.5：性能测试

- [ ] 创建 `benchmark_test.go`
  - 测试查询性能
  - 测试批量操作性能
  - 测试预加载性能
  - 测试事务性能

---

## 4. 技术细节

### 4.1 依赖管理

```go
// go.mod
module com.litelake.litecore

go 1.25.0

require (
    gorm.io/gorm v1.25.12
    gorm.io/driver/mysql v1.5.7
    gorm.io/driver/postgres v1.5.9
    gorm.io/driver/sqlite v1.5.7

    // 其他依赖...
)

// 可以移除的依赖
// github.com/go-sql-driver/mysql
// github.com/lib/pq
// github.com/mattn/go-sqlite3
```

### 4.2 GORM 配置

#### 4.2.1 默认配置

```go
// internal/config/gorm_config.go
package config

import (
    "gorm.io/gorm"
    "gorm.io/gorm/logger"
)

// DefaultGormConfig 默认 GORM 配置
var DefaultGormConfig = &gorm.Config{
    // 跳过默认事务（提高性能）
    SkipDefaultTransaction: true,

    // 禁用外键约束（在迁移时）
    DisableForeignKeyConstraintWhenMigrating: true,

    // 命名策略
    NamingStrategy: schema.NamingStrategy{
        SingularTable: true,                  // 使用单数表名
        NoLowerCase:   false,                 // 转换为小写
    },

    // Logger（集成 telemetry）
    Logger: logger.Default.LogMode(logger.Silent),
}
```

#### 4.2.2 开发环境配置

```go
// 开发环境配置
DevelopmentGormConfig = &gorm.Config{
    SkipDefaultTransaction: false,
    Logger:                 logger.Default.LogMode(logger.Info),
}
```

#### 4.2.3 生产环境配置

```go
// 生产环境配置
ProductionGormConfig = &gorm.Config{
    SkipDefaultTransaction: true,
    Logger:                 logger.Default.LogMode(logger.Error),
}
```

### 4.3 连接池配置

GORM 使用 `database/sql` 的连接池，因此现有的 `PoolConfig` 继续有效：

```go
// 配置连接池
sqlDB, err := db.DB()
if err != nil {
    return err
}

// 应用连接池配置
sqlDB.SetMaxOpenConns(poolConfig.MaxOpenConns)
sqlDB.SetMaxIdleConns(poolConfig.MaxIdleConns)
sqlDB.SetConnMaxLifetime(poolConfig.ConnMaxLifetime)
sqlDB.SetConnMaxIdleTime(poolConfig.ConnMaxIdleTime)
```

### 4.4 错误处理

GORM 的错误处理方式：

```go
// GORM 错误处理
result := db.Where("age > ?", 18).Find(&users)
if result.Error != nil {
    return result.Error
}

// 检查记录是否存在
errors.Is(result.Error, gorm.ErrRecordNotFound)

// 检查影响行数
if result.RowsAffected == 0 {
    return errors.New("no records affected")
}
```

### 4.5 事务处理

#### 4.5.1 自动事务

```go
// 使用 Transaction 方法（推荐）
err := dbMgr.Transaction(func(tx *gorm.DB) error {
    // 执行操作
    if err := tx.Create(&user).Error; err != nil {
        return err // 自动回滚
    }

    if err := tx.Create(&profile).Error; err != nil {
        return err // 自动回滚
    }

    return nil // 自动提交
})
```

#### 4.5.2 手动事务

```go
// 手动控制事务
tx := dbMgr.Begin()
defer func() {
    if r := recover(); r != nil {
        tx.Rollback()
        panic(r)
    }
}()

// 执行操作
if err := tx.Create(&user).Error; err != nil {
    tx.Rollback()
    return err
}

// 提交
if err := tx.Commit().Error; err != nil {
    return err
}
```

#### 4.5.3 嵌套事务

```go
// 嵌套事务（通过保存点实现）
dbMgr.Transaction(func(tx *gorm.DB) error {
    // 外层事务

    tx.Transaction(func(tx2 *gorm.DB) error {
        // 内层事务（保存点）
        return tx2.Create(&order).Error
    })

    return nil
})
```

---

## 5. 使用示例

### 5.1 快速开始

```go
package main

import (
    "fmt"

    "com.litelake.litecore/manager/databasemgr"
    "com.litelake.litecore/models"
)

func main() {
    // 创建工厂实例
    factory := databasemgr.NewFactory()

    // 配置数据库（使用 MySQL）
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

    // 构建数据库管理器
    mgr := factory.Build("mysql", cfg)
    dbMgr := mgr.(databasemgr.DatabaseManager)

    // 自动迁移（创建表）
    err := dbMgr.AutoMigrate(
        &models.User{},
        &models.Profile{},
        &models.Order{},
        &models.OrderItem{},
        &models.Role{},
    )
    if err != nil {
        panic(err)
    }

    // 创建记录
    user := models.User{
        Username: "johndoe",
        Email:    "john@example.com",
        Age:      30,
    }
    result := dbMgr.DB().Create(&user)
    if result.Error != nil {
        panic(result.Error)
    }

    // 查询记录
    var foundUser models.User
    result = dbMgr.DB().First(&foundUser, user.ID)
    if result.Error != nil {
        panic(result.Error)
    }

    fmt.Printf("Found user: %+v\n", foundUser)
}
```

### 5.2 CRUD 操作

#### 5.2.1 Create（创建）

```go
// 创建单条记录
user := models.User{
    Username: "alice",
    Email:    "alice@example.com",
    Age:      25,
}
result := dbMgr.DB().Create(&user)
if result.Error != nil {
    panic(result.Error)
}

// 批量创建
users := []models.User{
    {Username: "bob", Email: "bob@example.com", Age: 30},
    {Username: "charlie", Email: "charlie@example.com", Age: 35},
}
result = dbMgr.DB().Create(&users)

// 使用 CreateInBatches（大批量）
result = dbMgr.DB().CreateInBatches(users, 100)
```

#### 5.2.2 Read（查询）

```go
// 查询单条记录（根据主键）
var user models.User
result := dbMgr.DB().First(&user, 1)

// 查询单条记录（根据条件）
result = dbMgr.DB().Where("username = ?", "alice").First(&user)

// 查询多条记录
var users []models.User
result = dbMgr.DB().Where("age > ?", 18).Find(&users)

// 使用 IN 查询
result = dbMgr.DB().Where("id IN ?", []int{1, 2, 3}).Find(&users)

// 使用 OR 条件
result = dbMgr.DB().Where("username = ?", "alice").Or("username = ?", "bob").Find(&users)

// 链式调用
result = dbMgr.DB().
    Where("age > ?", 18).
    Where("age < ?", 60).
    Order("age DESC").
    Limit(10).
    Find(&users)
```

#### 5.2.3 Update（更新）

```go
// 更新单个字段
result := dbMgr.DB().Model(&user).Update("age", 31)

// 更新多个字段
result = dbMgr.DB().Model(&user).Updates(models.User{
    Username: "alice_updated",
    Age:      32,
})

// 使用 Map 更新
result = dbMgr.DB().Model(&user).Updates(map[string]any{
    "username": "alice_updated",
    "age":      32,
})

// 使用 SQL 表达式
result = dbMgr.DB().Model(&user).Update("age", gorm.Expr("age + ?", 1))

// 批量更新
result = dbMgr.DB().Model(&models.User{}).Where("age > ?", 18).Update("status", "active")
```

#### 5.2.4 Delete（删除）

```go
// 删除单条记录（物理删除）
result := dbMgr.DB().Delete(&user)

// 删除根据主键
result = dbMgr.DB().Delete(&models.User{}, 1)

// 批量删除
result = dbMgr.DB().Where("age < ?", 18).Delete(&models.User{})

// 软删除（如果模型有 DeletedAt 字段）
result = dbMgr.DB().Delete(&user) // 不会真正删除，只是设置 deleted_at

// 查询包括软删除的记录
result = dbMgr.DB().Unscoped().Find(&users)

// 真正删除软删除的记录
result = dbMgr.DB().Unscoped().Delete(&user)
```

### 5.3 关联关系

#### 5.3.1 Has One（一个用户有一个资料）

```go
// 创建关联
profile := models.Profile{
    UserID:   user.ID,
    Avatar:   "avatar.jpg",
    Bio:      "Hello, world!",
    Location: "Beijing",
}
result := dbMgr.DB().Create(&profile)

// 查询关联（使用 Preload）
var userWithProfile models.User
result = dbMgr.DB().
    Preload("Profile").
    First(&userWithProfile, user.ID)

// 使用 Joins（更高效）
result = dbMgr.DB().
    Joins("Profile").
    First(&userWithProfile, user.ID)
```

#### 5.3.2 Has Many（一个用户有多个订单）

```go
// 创建关联
order := models.Order{
    UserID:     user.ID,
    OrderNo:    "ORD001",
    TotalPrice: 99.99,
    Status:     "pending",
}
result := dbMgr.DB().Create(&order)

// 查询关联
var userWithOrders models.User
result = dbMgr.DB().
    Preload("Orders").
    First(&userWithOrders, user.ID)

// 预加载并排序
result = dbMgr.DB().
    Preload("Orders", func(db *gorm.DB) *gorm.DB {
        return db.Order("created_at DESC")
    }).
    First(&userWithOrders, user.ID)
```

#### 5.3.3 Many To Many（用户和角色）

```go
// 创建关联
role := models.Role{
    Name:        "admin",
    Description: "Administrator",
}
result := dbMgr.DB().Create(&role)

// 关联用户和角色
result = dbMgr.DB().Model(&user).Association("Roles").Append(&role)

// 查询关联
var userWithRoles models.User
result = dbMgr.DB().
    Preload("Roles").
    First(&userWithRoles, user.ID)

// 替换关联
result = dbMgr.DB().Model(&user).Association("Roles").Replace(
    &models.Role{Name: "user"},
)

// 删除关联
result = dbMgr.DB().Model(&user).Association("Roles").Delete(&role)

// 清空关联
result = dbMgr.DB().Model(&user).Association("Roles").Clear()
```

### 5.4 事务操作

#### 5.4.1 简单事务

```go
// 使用 Transaction 方法
err := dbMgr.Transaction(func(tx *gorm.DB) error {
    // 创建用户
    if err := tx.Create(&user).Error; err != nil {
        return err
    }

    // 创建用户资料
    profile := models.Profile{
        UserID: user.ID,
        Bio:    "New user",
    }
    if err := tx.Create(&profile).Error; err != nil {
        return err
    }

    return nil
})
```

#### 5.4.2 嵌套事务

```go
// 外层事务
err := dbMgr.Transaction(func(tx *gorm.DB) error {
    // 创建用户
    if err := tx.Create(&user).Error; err != nil {
        return err
    }

    // 内层事务（保存点）
    tx.Transaction(func(tx2 *gorm.DB) error {
        // 创建订单
        order := models.Order{
            UserID:     user.ID,
            OrderNo:    "ORD001",
            TotalPrice: 99.99,
        }
        return tx2.Create(&order).Error
    })

    return nil
})
```

### 5.5 复杂查询

#### 5.5.1 链式查询

```go
query := dbMgr.DB().Model(&models.User{})

// 动态添加条件
if username != "" {
    query = query.Where("username LIKE ?", "%"+username+"%")
}

if minAge > 0 {
    query = query.Where("age >= ?", minAge)
}

if maxAge > 0 {
    query = query.Where("age <= ?", maxAge)
}

// 排序和分页
query = query.
    Order("created_at DESC").
    Limit(20).
    Offset(0)

// 执行查询
var users []models.User
result := query.Find(&users)
```

#### 5.5.2 使用 Scopes

```go
// 定义作用域
func AgeGreaterThan(minAge int) func(*gorm.DB) *gorm.DB {
    return func(db *gorm.DB) *gorm.DB {
        return db.Where("age > ?", minAge)
    }
}

func AgeLessThan(maxAge int) func(*gorm.DB) *gorm.DB {
    return func(db *gorm.DB) *gorm.DB {
        return db.Where("age < ?", maxAge)
    }
}

func ActiveUsers() func(*gorm.DB) *gorm.DB {
    return func(db *gorm.DB) *gorm.DB {
        return db.Where("status = ?", "active")
    }
}

// 使用作用域
var users []models.User
result := dbMgr.DB().
    Scopes(
        AgeGreaterThan(18),
        AgeLessThan(60),
        ActiveUsers(),
    ).
    Find(&users)
```

### 5.6 原生 SQL 查询

```go
// 使用 Raw 查询
var users []models.User
result := dbMgr.DB().
    Raw("SELECT * FROM users WHERE age > ?", 18).
    Scan(&users)

// 使用 Exec 执行（无返回结果）
result := dbMgr.DB().
    Exec("UPDATE users SET status = ? WHERE age < ?", "inactive", 18)

// 使用命名参数
result = dbMgr.DB().
    Raw(
        "SELECT * FROM users WHERE username = @username OR email = @email",
        sql.Named("username", "alice"),
        sql.Named("email", "alice@example.com"),
    ).
    Scan(&users)
```

### 5.7 聚合查询

```go
// Count
var count int64
result := dbMgr.DB().Model(&models.User{}).Where("age > ?", 18).Count(&count)

// Sum
var totalAge int
result = dbMgr.DB().Model(&models.User{}).Select("SUM(age)").Scan(&totalAge)

// Avg
var avgAge float64
result = dbMgr.DB().Model(&models.User{}).Select("AVG(age)").Scan(&avgAge)

// Max/Min
var maxAge int
result = dbMgr.DB().Model(&models.User{}).Select("MAX(age)").Scan(&maxAge)

// Group By
var results []struct {
    Age   int
    Count int64
}
result = dbMgr.DB().Model(&models.User{}).
    Select("age, COUNT(*) as count").
    Group("age").
    Having("count > ?", 10).
    Scan(&results)
```

---

## 6. 测试策略

### 6.1 单元测试

```go
// internal/drivers/mysql_driver_test.go
package drivers_test

import (
    "testing"

    "com.litelake.litecore/manager/databasemgr/internal/config"
    "com.litelake.litecore/manager/databasemgr/internal/drivers"
)

func TestNewMySQLManager(t *testing.T) {
    cfg := &config.DatabaseConfig{
        Driver: "mysql",
        MySQLConfig: &config.MySQLConfig{
            DSN: "root:password@tcp(localhost:3306)/test?charset=utf8mb4&parseTime=True",
        },
    }

    mgr, err := drivers.NewMySQLManager(cfg)
    if err != nil {
        t.Fatalf("failed to create MySQL manager: %v", err)
    }

    if mgr.Driver() != "mysql" {
        t.Errorf("expected driver 'mysql', got '%s'", mgr.Driver())
    }

    // 测试连接
    if err := mgr.Health(); err != nil {
        t.Errorf("health check failed: %v", err)
    }
}
```

### 6.2 集成测试

```go
// integration_test.go
package databasemgr_test

import (
    "testing"

    "com.litelake.litecore/manager/databasemgr"
    "com.litelake.litecore/models"
)

func TestUserCRUD(t *testing.T) {
    factory := databasemgr.NewFactory()

    cfg := map[string]any{
        "driver": "sqlite",
        "sqlite_config": map[string]any{
            "dsn": "file::memory:?cache=shared",
        },
    }

    mgr := factory.Build("sqlite", cfg)
    dbMgr := mgr.(databasemgr.DatabaseManager)

    // 自动迁移
    err := dbMgr.AutoMigrate(&models.User{})
    if err != nil {
        t.Fatalf("failed to migrate: %v", err)
    }

    // 创建
    user := models.User{
        Username: "test",
        Email:    "test@example.com",
        Age:      25,
    }
    result := dbMgr.DB().Create(&user)
    if result.Error != nil {
        t.Fatalf("failed to create user: %v", result.Error)
    }

    // 查询
    var foundUser models.User
    result = dbMgr.DB().First(&foundUser, user.ID)
    if result.Error != nil {
        t.Fatalf("failed to find user: %v", result.Error)
    }

    if foundUser.Username != "test" {
        t.Errorf("expected username 'test', got '%s'", foundUser.Username)
    }

    // 更新
    result = dbMgr.DB().Model(&foundUser).Update("age", 26)
    if result.Error != nil {
        t.Fatalf("failed to update user: %v", result.Error)
    }

    // 删除
    result = dbMgr.DB().Delete(&foundUser)
    if result.Error != nil {
        t.Fatalf("failed to delete user: %v", result.Error)
    }
}
```

### 6.3 性能测试

```go
// benchmark_test.go
package databasemgr_test

import (
    "testing"

    "com.litelake.litecore/manager/databasemgr"
    "com.litelake.litecore/models"
)

func BenchmarkCreate(b *testing.B) {
    factory := databasemgr.NewFactory()
    cfg := map[string]any{
        "driver": "sqlite",
        "sqlite_config": map[string]any{
            "dsn": "file::memory:?cache=shared",
        },
    }

    mgr := factory.Build("sqlite", cfg)
    dbMgr := mgr.(databasemgr.DatabaseManager)

    dbMgr.AutoMigrate(&models.User{})

    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        user := models.User{
            Username: "user",
            Email:    "user@example.com",
            Age:      25,
        }
        dbMgr.DB().Create(&user)
    }
}

func BenchmarkQuery(b *testing.B) {
    factory := databasemgr.NewFactory()
    cfg := map[string]any{
        "driver": "sqlite",
        "sqlite_config": map[string]any{
            "dsn": "file::memory:?cache=shared",
        },
    }

    mgr := factory.Build("sqlite", cfg)
    dbMgr := mgr.(databasemgr.DatabaseManager)

    dbMgr.AutoMigrate(&models.User{})

    // 准备数据
    for i := 0; i < 1000; i++ {
        user := models.User{
            Username: "user",
            Email:    "user@example.com",
            Age:      25,
        }
        dbMgr.DB().Create(&user)
    }

    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        var users []models.User
        dbMgr.DB().Where("age > ?", 18).Find(&users)
    }
}
```

---

## 7. 开发里程碑

| 阶段 | 任务 | 预计时间 | 依赖 |
|------|------|----------|------|
| **第一阶段** | GORM 集成（核心改造） | 2-3 天 | - |
| 1.1 | 清理旧依赖、添加 GORM 依赖 | 0.5 天 | - |
| 1.2 | 创建 GORM 基础管理器 | 0.5 天 | 1.1 |
| 1.3 | 重写 MySQL 驱动 | 0.5 天 | 1.2 |
| 1.4 | 重写 PostgreSQL 驱动 | 0.5 天 | 1.2 |
| 1.5 | 重写 SQLite 驱动 | 0.5 天 | 1.2 |
| 1.6 | 重写 None 驱动 | 0.5 天 | 1.2 |
| **第二阶段** | 接口重构（简化设计） | 1-2 天 | 第一阶段 |
| 2.1 | 重写 DatabaseManager 接口 | 0.5 天 | 第一阶段 |
| 2.2 | 删除适配器 | 0.5 天 | 2.1 |
| 2.3 | 简化工厂模式 | 0.5 天 | 2.1 |
| **第三阶段** | 高级特性（增强功能） | 2-3 天 | 第二阶段 |
| 3.1 | 实现钩子管理器 | 0.5 天 | 第二阶段 |
| 3.2 | 实现自动迁移 | 0.5 天 | 第二阶段 |
| 3.3 | 实现事务管理器 | 0.5 天 | 第二阶段 |
| 3.4 | 实现查询构建器（可选） | 1 天 | 第二阶段 |
| **第四阶段** | 文档和测试（完善生态） | 2-3 天 | 第三阶段 |
| 4.1 | 更新文档 | 1 天 | 第三阶段 |
| 4.2 | 单元测试 | 0.5 天 | 第三阶段 |
| 4.3 | 集成测试 | 0.5 天 | 第三阶段 |
| 4.4 | 使用示例 | 0.5 天 | 第三阶段 |
| 4.5 | 性能测试 | 0.5 天 | 第三阶段 |

**总计**：7-11 天

---

## 8. 风险评估

### 8.1 潜在风险

| 风险 | 影响 | 概率 | 缓解措施 |
|------|------|------|----------|
| 破坏性变更导致现有代码无法运行 | 高 | 高 | ✅ **已接受**：明确表示不需要向后兼容 |
| 学习曲线 | 中 | 中 | 1. 提供详细文档和示例<br>2. 团队培训 |
| GORM 版本升级 | 低 | 低 | 1. 锁定 GORM 版本<br>2. 定期评估升级 |
| 依赖库问题 | 中 | 低 | 1. GORM 是成熟项目<br>2. 活跃的社区 |

### 8.2 优势

1. **代码更简洁**
   - 移除适配器层
   - 直接暴露 GORM API
   - 减少代码量 30%+

2. **性能更好**
   - 无适配器开销
   - 直接使用 GORM 优化
   - 减少内存分配

3. **维护更简单**
   - 代码结构更清晰
   - 依赖关系更简单
   - 测试更容易

---

## 9. 总结

### 9.1 改造收益

1. **开发效率提升 70%+**
   - 完全使用 GORM API
   - 减少 SQL 编写工作量
   - 类型安全的查询

2. **代码质量提升**
   - 统一的模型定义
   - 自动迁移和 schema 管理
   - 更好的可维护性

3. **功能增强**
   - 强大的关联关系管理
   - 丰富的钩子机制
   - 支持复杂查询场景

4. **架构优化**
   - 移除适配器层
   - 直接暴露 GORM
   - 代码更简洁

### 9.2 改造成本

1. **开发成本**
   - 开发时间：7-11 天
   - 测试时间：2-3 天
   - 文档更新：1 天

2. **学习成本**
   - GORM API 学习（1-2 天）
   - 最佳实践学习（1 天）
   - 团队培训（1 天）

3. **迁移成本**
   - **不需要考虑向后兼容**
   - 直接重写相关代码
   - 清理旧代码

### 9.3 建议

1. **完全重写**
   - 不保留旧代码
   - 直接使用 GORM API
   - 清理所有 `database/sql` 相关代码

2. **充分测试**
   - 单元测试覆盖
   - 集成测试验证
   - 性能对比测试

3. **团队协作**
   - 代码审查
   - 知识分享
   - 最佳实践文档

4. **监控和反馈**
   - 性能监控
   - 错误追踪
   - 用户反馈收集

---

## 10. 附录

### 10.1 参考资料

- [GORM 官方文档](https://gorm.io/docs/)
- [GORM 中文文档](https://gorm.zh-cn.io/docs/)
- [GORM 最佳实践](https://gorm.io/docs/conventions.html)
- [GORM 性能优化](https://gorm.io/docs/performance.html)

### 10.2 常见问题

#### Q1: 为什么不需要向后兼容？

A: 因为这是一个**彻底的改造**，保留向后兼容会增加复杂度，且不利于发挥 GORM 的全部能力。

#### Q2: GORM 性能如何？

A: GORM 的性能在大多数场景下与原生 SQL 相当，通过合理配置可以进一步优化：
- 禁用默认事务
- 使用批量操作
- 使用预加载避免 N+1 查询

#### Q3: 如何在 GORM 中使用原生 SQL？

A: GORM 提供了多种原生 SQL 支持：
- `Raw()` - 执行原生查询并扫描到模型
- `Exec()` - 执行原生 SQL（无返回值）

#### Q4: 如何处理 GORM 的错误？

A:
- `result.Error` - 操作错误
- `result.RowsAffected` - 影响行数
- `errors.Is(result.Error, gorm.ErrRecordNotFound)` - 判断记录不存在

### 10.3 术语表

| 术语 | 说明 |
|------|------|
| ORM | Object-Relational Mapping，对象关系映射 |
| CRUD | Create, Read, Update, Delete |
| Preload | 预加载，提前加载关联数据 |
| Hook | 钩子，在特定事件前后执行的回调函数 |
| Migration | 数据库迁移，自动创建或更新表结构 |
| Soft Delete | 软删除，不真正删除记录，只标记为已删除 |
| Transaction | 事务，一组原子性操作的集合 |
| Nested Transaction | 嵌套事务，事务中的事务（通过保存点实现） |

---

**文档版本**：v2.0（无向后兼容版）
**创建日期**：2025-01-11
**作者**：Claude Code
**审核状态**：待审核
