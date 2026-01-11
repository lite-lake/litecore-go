package databasemgr

import (
	"context"
	"database/sql"
	"fmt"
	"sync"
	"time"

	"gorm.io/gorm"

	"com.litelake.litecore/common"
	"com.litelake.litecore/manager/databasemgr/internal/config"
	"com.litelake.litecore/manager/databasemgr/internal/drivers"
)

// Manager 数据库管理器（依赖注入模式）
type Manager struct {
	// 依赖注入字段
	Config common.BaseConfigProvider `inject:""`

	// 内部状态
	name   string
	driver string
	db     *gorm.DB
	sqlDB  *sql.DB
	mu     sync.RWMutex
	once   sync.Once
}

// NewManager 创建数据库管理器
func NewManager(name string) *Manager {
	return &Manager{
		name:   name,
		driver: "none",
	}
}

// ManagerName 返回管理器名称
func (m *Manager) ManagerName() string {
	return m.name
}

// OnStart 初始化管理器（依赖注入完成后调用）
func (m *Manager) OnStart() error {
	var initErr error
	m.once.Do(func() {
		// 1. 从 Config 获取配置
		cfg, err := m.loadConfig()
		if err != nil {
			initErr = fmt.Errorf("load config failed: %w", err)
			return
		}

		// 2. 根据驱动类型创建相应的数据库管理器
		var databaseManager DatabaseManager
		switch cfg.Driver {
		case "mysql":
			databaseManager, err = drivers.NewMySQLManager(cfg)
		case "postgresql":
			databaseManager, err = drivers.NewPostgreSQLManager(cfg)
		case "sqlite":
			databaseManager, err = drivers.NewSQLiteManager(cfg)
		case "none":
			databaseManager = drivers.NewNoneDatabaseManager()
		default:
			initErr = fmt.Errorf("unsupported database driver: %s", cfg.Driver)
			return
		}

		if err != nil {
			initErr = fmt.Errorf("create database driver failed: %w", err)
			return
		}

		// 3. 获取 GORM 实例
		m.db = databaseManager.DB()
		m.driver = cfg.Driver

		// 4. 获取 sql.DB 用于连接池管理
		m.sqlDB, err = m.db.DB()
		if err != nil {
			initErr = fmt.Errorf("get sql.DB failed: %w", err)
			return
		}

		// 5. 测试连接（none 驱动除外）
		if cfg.Driver != "none" {
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()
			if err := m.sqlDB.PingContext(ctx); err != nil {
				initErr = fmt.Errorf("ping database failed: %w", err)
				return
			}
		}
	})
	return initErr
}

// loadConfig 从 ConfigProvider 加载配置
func (m *Manager) loadConfig() (*config.DatabaseConfig, error) {
	if m.Config == nil {
		return nil, fmt.Errorf("config provider is required")
	}

	// 获取配置键：database.{manager_name}
	cfgKey := fmt.Sprintf("database.%s", m.name)
	cfgData, err := m.Config.Get(cfgKey)
	if err != nil {
		return nil, fmt.Errorf("get config failed: %s: %w", cfgKey, err)
	}

	// 将配置数据转换为 map
	cfgMap, ok := cfgData.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("invalid config type for %s: expected map[string]any, got %T", cfgKey, cfgData)
	}

	return config.ParseFromMap(cfgMap)
}

// OnStop 停止管理器
func (m *Manager) OnStop() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.sqlDB == nil {
		return nil
	}

	err := m.sqlDB.Close()
	m.sqlDB = nil
	m.db = nil
	return err
}

// Health 健康检查
func (m *Manager) Health() error {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.sqlDB == nil {
		return fmt.Errorf("database not initialized")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return m.sqlDB.PingContext(ctx)
}

// ========== GORM 核心 ==========

// DB 获取 GORM 数据库实例
func (m *Manager) DB() *gorm.DB {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.db
}

// Model 指定模型进行操作
func (m *Manager) Model(value any) *gorm.DB {
	return m.DB().Model(value)
}

// Table 指定表名进行操作
func (m *Manager) Table(name string) *gorm.DB {
	return m.DB().Table(name)
}

// WithContext 设置上下文
func (m *Manager) WithContext(ctx context.Context) *gorm.DB {
	return m.DB().WithContext(ctx)
}

// ========== 事务管理 ==========

// Transaction 执行事务
func (m *Manager) Transaction(fn func(*gorm.DB) error, opts ...*sql.TxOptions) error {
	return m.DB().Transaction(fn, opts...)
}

// Begin 开启事务
func (m *Manager) Begin(opts ...*sql.TxOptions) *gorm.DB {
	return m.DB().Begin(opts...)
}

// ========== 迁移管理 ==========

// AutoMigrate 自动迁移
func (m *Manager) AutoMigrate(models ...any) error {
	return m.DB().AutoMigrate(models...)
}

// Migrator 获取迁移器
func (m *Manager) Migrator() gorm.Migrator {
	return m.DB().Migrator()
}

// ========== 连接管理 ==========

// Driver 获取数据库驱动类型
func (m *Manager) Driver() string {
	return m.driver
}

// Ping 检查数据库连接
func (m *Manager) Ping(ctx context.Context) error {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.sqlDB == nil {
		return fmt.Errorf("database not initialized")
	}

	return m.sqlDB.PingContext(ctx)
}

// Stats 获取连接池统计信息
func (m *Manager) Stats() sql.DBStats {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.sqlDB == nil {
		return sql.DBStats{}
	}

	return m.sqlDB.Stats()
}

// Close 关闭数据库连接
func (m *Manager) Close() error {
	return m.OnStop()
}

// ========== 原生 SQL ==========

// Exec 执行原生 SQL
func (m *Manager) Exec(sql string, values ...any) *gorm.DB {
	return m.DB().Exec(sql, values...)
}

// Raw 执行原生查询
func (m *Manager) Raw(sql string, values ...any) *gorm.DB {
	return m.DB().Raw(sql, values...)
}

// 确保 Manager 实现 DatabaseManager 和 common.BaseManager
var _ DatabaseManager = (*Manager)(nil)
var _ common.BaseManager = (*Manager)(nil)
