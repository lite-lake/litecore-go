package databasemgr

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// sqliteImpl SQLite 数据库管理器实现
type sqliteImpl struct {
	*databaseManagerBaseImpl
}

// NewDatabaseManagerSQLiteImpl 创建 SQLite 数据库管理器
func NewDatabaseManagerSQLiteImpl(cfg *SQLiteConfig) (DatabaseManager, error) {
	if cfg == nil {
		return nil, fmt.Errorf("sqlite config is required")
	}

	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid sqlite config: %w", err)
	}

	if cfg.DSN == "" {
		return nil, fmt.Errorf("sqlite DSN is required")
	}

	// GORM 配置
	gormConfig := &gorm.Config{
		SkipDefaultTransaction: true,
		Logger:                 logger.Default.LogMode(logger.Silent),
	}

	// 打开数据库连接
	db, err := gorm.Open(sqlite.Open(cfg.DSN), gormConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to open sqlite database: %w", err)
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
	baseImpl := newDatabaseManagerBaseImpl("sqlite", "sqlite", db)

	// 创建完整配置用于初始化可观测性
	fullCfg := &DatabaseConfig{
		Driver:       "sqlite",
		SQLiteConfig: cfg,
	}

	// 初始化可观测性
	baseImpl.initObservability(fullCfg)

	// 测试连接
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := sqlDB.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping sqlite database: %w", err)
	}

	return &sqliteImpl{
		databaseManagerBaseImpl: baseImpl,
	}, nil
}

// ManagerName 返回管理器名称
func (s *sqliteImpl) ManagerName() string {
	return s.name
}

// Health 健康检查
func (s *sqliteImpl) Health() error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.sqlDB == nil {
		return fmt.Errorf("database not initialized")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return s.sqlDB.PingContext(ctx)
}

// OnStart 启动时初始化
func (s *sqliteImpl) OnStart() error {
	return s.Health()
}

// OnStop 停止时清理
func (s *sqliteImpl) OnStop() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.sqlDB == nil {
		return nil
	}

	err := s.sqlDB.Close()
	s.sqlDB = nil
	s.db = nil
	return err
}

// ========== GORM 核心 ==========

// DB 获取 GORM 数据库实例
func (s *sqliteImpl) DB() *gorm.DB {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.db
}

// Model 指定模型进行操作
func (s *sqliteImpl) Model(value any) *gorm.DB {
	return s.DB().Model(value)
}

// Table 指定表名进行操作
func (s *sqliteImpl) Table(name string) *gorm.DB {
	return s.DB().Table(name)
}

// WithContext 设置上下文
func (s *sqliteImpl) WithContext(ctx context.Context) *gorm.DB {
	return s.DB().WithContext(ctx)
}

// ========== 事务管理 ==========

// Transaction 执行事务
func (s *sqliteImpl) Transaction(fn func(*gorm.DB) error, opts ...*sql.TxOptions) error {
	return s.DB().Transaction(fn, opts...)
}

// Begin 开启事务
func (s *sqliteImpl) Begin(opts ...*sql.TxOptions) *gorm.DB {
	return s.DB().Begin(opts...)
}

// ========== 迁移管理 ==========

// AutoMigrate 自动迁移
func (s *sqliteImpl) AutoMigrate(models ...any) error {
	return s.DB().AutoMigrate(models...)
}

// Migrator 获取迁移器
func (s *sqliteImpl) Migrator() gorm.Migrator {
	return s.DB().Migrator()
}

// ========== 连接管理 ==========

// Driver 获取驱动类型
func (s *sqliteImpl) Driver() string {
	return s.driver
}

// Ping 检查数据库连接
func (s *sqliteImpl) Ping(ctx context.Context) error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.sqlDB == nil {
		return fmt.Errorf("database not initialized")
	}

	return s.sqlDB.PingContext(ctx)
}

// Stats 获取连接池统计信息
func (s *sqliteImpl) Stats() sql.DBStats {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.sqlDB == nil {
		return sql.DBStats{}
	}

	return s.sqlDB.Stats()
}

// Close 关闭数据库连接
func (s *sqliteImpl) Close() error {
	return s.OnStop()
}

// ========== 原生 SQL ==========

// Exec 执行原生 SQL
func (s *sqliteImpl) Exec(sql string, values ...any) *gorm.DB {
	return s.DB().Exec(sql, values...)
}

// Raw 执行原生查询
func (s *sqliteImpl) Raw(sql string, values ...any) *gorm.DB {
	return s.DB().Raw(sql, values...)
}
