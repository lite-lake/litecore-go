package databasemgr

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// databaseManagerMysqlImpl MySQL 数据库管理器实现
type databaseManagerMysqlImpl struct {
	*databaseManagerBaseImpl
}

// NewDatabaseManagerMySQLImpl 创建 MySQL 数据库管理器
func NewDatabaseManagerMySQLImpl(cfg *MySQLConfig) (IDatabaseManager, error) {
	if cfg == nil {
		return nil, fmt.Errorf("mysql configmgr is required")
	}

	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid mysql configmgr: %w", err)
	}

	if cfg.DSN == "" {
		return nil, fmt.Errorf("mysql DSN is required")
	}

	// GORM 配置
	gormConfig := &gorm.Config{
		SkipDefaultTransaction: true,
		Logger:                 logger.Default.LogMode(logger.Silent),
	}

	// 打开数据库连接
	db, err := gorm.Open(mysql.Open(cfg.DSN), gormConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to open mysql database: %w", err)
	}

	// 配置连接池
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get sql.DB: %w", err)
	}

	if cfg.PoolConfig != nil {
		sqlDB.SetMaxOpenConns(cfg.PoolConfig.MaxOpenConns)
		sqlDB.SetMaxIdleConns(cfg.PoolConfig.MaxIdleConns)
		sqlDB.SetConnMaxLifetime(cfg.PoolConfig.ConnMaxLifetime)
		sqlDB.SetConnMaxIdleTime(cfg.PoolConfig.ConnMaxIdleTime)
	}

	// 创建基础实现
	baseImpl := newIDatabaseManagerBaseImpl("databaseManagerMysqlImpl", "mysql", db)

	// 创建完整配置用于初始化可观测性
	fullCfg := &DatabaseConfig{
		Driver:      "mysql",
		MySQLConfig: cfg,
	}

	// 初始化可观测性
	baseImpl.initObservability(fullCfg)

	// 测试连接
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := sqlDB.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping mysql database: %w", err)
	}

	return &databaseManagerMysqlImpl{
		databaseManagerBaseImpl: baseImpl,
	}, nil
}

// ManagerName 返回管理器名称
func (m *databaseManagerMysqlImpl) ManagerName() string {
	return m.name
}

// Health 健康检查
func (m *databaseManagerMysqlImpl) Health() error {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.sqlDB == nil {
		return fmt.Errorf("database not initialized")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return m.sqlDB.PingContext(ctx)
}

// OnStart 启动时初始化
func (m *databaseManagerMysqlImpl) OnStart() error {
	return m.Health()
}

// OnStop 停止时清理
func (m *databaseManagerMysqlImpl) OnStop() error {
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

// ========== GORM 核心 ==========

// DB 获取 GORM 数据库实例
func (m *databaseManagerMysqlImpl) DB() *gorm.DB {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.db
}

// Model 指定模型进行操作
func (m *databaseManagerMysqlImpl) Model(value any) *gorm.DB {
	return m.DB().Model(value)
}

// Table 指定表名进行操作
func (m *databaseManagerMysqlImpl) Table(name string) *gorm.DB {
	return m.DB().Table(name)
}

// WithContext 设置上下文
func (m *databaseManagerMysqlImpl) WithContext(ctx context.Context) *gorm.DB {
	return m.DB().WithContext(ctx)
}

// ========== 事务管理 ==========

// Transaction 执行事务
func (m *databaseManagerMysqlImpl) Transaction(fn func(*gorm.DB) error, opts ...*sql.TxOptions) error {
	return m.DB().Transaction(fn, opts...)
}

// Begin 开启事务
func (m *databaseManagerMysqlImpl) Begin(opts ...*sql.TxOptions) *gorm.DB {
	return m.DB().Begin(opts...)
}

// ========== 迁移管理 ==========

// AutoMigrate 自动迁移
func (m *databaseManagerMysqlImpl) AutoMigrate(models ...any) error {
	return m.DB().AutoMigrate(models...)
}

// Migrator 获取迁移器
func (m *databaseManagerMysqlImpl) Migrator() gorm.Migrator {
	return m.DB().Migrator()
}

// ========== 连接管理 ==========

// Driver 获取驱动类型
func (m *databaseManagerMysqlImpl) Driver() string {
	return m.driver
}

// Ping 检查数据库连接
func (m *databaseManagerMysqlImpl) Ping(ctx context.Context) error {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.sqlDB == nil {
		return fmt.Errorf("database not initialized")
	}

	return m.sqlDB.PingContext(ctx)
}

// Stats 获取连接池统计信息
func (m *databaseManagerMysqlImpl) Stats() sql.DBStats {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.sqlDB == nil {
		return sql.DBStats{}
	}

	return m.sqlDB.Stats()
}

// Close 关闭数据库连接
func (m *databaseManagerMysqlImpl) Close() error {
	return m.OnStop()
}

// ========== 原生 SQL ==========

// Exec 执行原生 SQL
func (m *databaseManagerMysqlImpl) Exec(sql string, values ...any) *gorm.DB {
	return m.DB().Exec(sql, values...)
}

// Raw 执行原生查询
func (m *databaseManagerMysqlImpl) Raw(sql string, values ...any) *gorm.DB {
	return m.DB().Raw(sql, values...)
}
