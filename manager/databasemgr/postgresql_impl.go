package databasemgr

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// postgresqlImpl PostgreSQL 数据库管理器实现
type postgresqlImpl struct {
	*databaseManagerBaseImpl
}

// NewDatabaseManagerPostgreSQLImpl 创建 PostgreSQL 数据库管理器
func NewDatabaseManagerPostgreSQLImpl(cfg *PostgreSQLConfig) (DatabaseManager, error) {
	if cfg == nil {
		return nil, fmt.Errorf("postgresql config is required")
	}

	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid postgresql config: %w", err)
	}

	if cfg.DSN == "" {
		return nil, fmt.Errorf("postgresql DSN is required")
	}

	// GORM 配置
	gormConfig := &gorm.Config{
		SkipDefaultTransaction: true,
		Logger:                 logger.Default.LogMode(logger.Silent),
	}

	// 打开数据库连接
	db, err := gorm.Open(postgres.Open(cfg.DSN), gormConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to open postgresql database: %w", err)
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
	baseImpl := newDatabaseManagerBaseImpl("postgresql", "postgresql", db)

	// 创建完整配置用于初始化可观测性
	fullCfg := &DatabaseConfig{
		Driver:           "postgresql",
		PostgreSQLConfig: cfg,
	}

	// 初始化可观测性
	baseImpl.initObservability(fullCfg)

	// 测试连接
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := sqlDB.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping postgresql database: %w", err)
	}

	return &postgresqlImpl{
		databaseManagerBaseImpl: baseImpl,
	}, nil
}

// ManagerName 返回管理器名称
func (p *postgresqlImpl) ManagerName() string {
	return p.name
}

// Health 健康检查
func (p *postgresqlImpl) Health() error {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if p.sqlDB == nil {
		return fmt.Errorf("database not initialized")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return p.sqlDB.PingContext(ctx)
}

// OnStart 启动时初始化
func (p *postgresqlImpl) OnStart() error {
	return p.Health()
}

// OnStop 停止时清理
func (p *postgresqlImpl) OnStop() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.sqlDB == nil {
		return nil
	}

	err := p.sqlDB.Close()
	p.sqlDB = nil
	p.db = nil
	return err
}

// ========== GORM 核心 ==========

// DB 获取 GORM 数据库实例
func (p *postgresqlImpl) DB() *gorm.DB {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.db
}

// Model 指定模型进行操作
func (p *postgresqlImpl) Model(value any) *gorm.DB {
	return p.DB().Model(value)
}

// Table 指定表名进行操作
func (p *postgresqlImpl) Table(name string) *gorm.DB {
	return p.DB().Table(name)
}

// WithContext 设置上下文
func (p *postgresqlImpl) WithContext(ctx context.Context) *gorm.DB {
	return p.DB().WithContext(ctx)
}

// ========== 事务管理 ==========

// Transaction 执行事务
func (p *postgresqlImpl) Transaction(fn func(*gorm.DB) error, opts ...*sql.TxOptions) error {
	return p.DB().Transaction(fn, opts...)
}

// Begin 开启事务
func (p *postgresqlImpl) Begin(opts ...*sql.TxOptions) *gorm.DB {
	return p.DB().Begin(opts...)
}

// ========== 迁移管理 ==========

// AutoMigrate 自动迁移
func (p *postgresqlImpl) AutoMigrate(models ...any) error {
	return p.DB().AutoMigrate(models...)
}

// Migrator 获取迁移器
func (p *postgresqlImpl) Migrator() gorm.Migrator {
	return p.DB().Migrator()
}

// ========== 连接管理 ==========

// Driver 获取驱动类型
func (p *postgresqlImpl) Driver() string {
	return p.driver
}

// Ping 检查数据库连接
func (p *postgresqlImpl) Ping(ctx context.Context) error {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if p.sqlDB == nil {
		return fmt.Errorf("database not initialized")
	}

	return p.sqlDB.PingContext(ctx)
}

// Stats 获取连接池统计信息
func (p *postgresqlImpl) Stats() sql.DBStats {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if p.sqlDB == nil {
		return sql.DBStats{}
	}

	return p.sqlDB.Stats()
}

// Close 关闭数据库连接
func (p *postgresqlImpl) Close() error {
	return p.OnStop()
}

// ========== 原生 SQL ==========

// Exec 执行原生 SQL
func (p *postgresqlImpl) Exec(sql string, values ...any) *gorm.DB {
	return p.DB().Exec(sql, values...)
}

// Raw 执行原生查询
func (p *postgresqlImpl) Raw(sql string, values ...any) *gorm.DB {
	return p.DB().Raw(sql, values...)
}
