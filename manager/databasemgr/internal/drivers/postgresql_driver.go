package drivers

import (
	"fmt"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"com.litelake.litecore/manager/databasemgr/internal/config"
)

// PostgreSQLManager PostgreSQL 数据库管理器
type PostgreSQLManager struct {
	*GormBaseManager
	config *config.DatabaseConfig
}

// NewPostgreSQLManager 创建 PostgreSQL 数据库管理器
func NewPostgreSQLManager(cfg *config.DatabaseConfig) (*PostgreSQLManager, error) {
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid database config: %w", err)
	}

	if cfg.PostgreSQLConfig == nil {
		return nil, fmt.Errorf("postgresql_config is required")
	}

	if cfg.PostgreSQLConfig.DSN == "" {
		return nil, fmt.Errorf("postgresql DSN is required")
	}

	// GORM 配置
	gormConfig := &gorm.Config{
		SkipDefaultTransaction: true,
		Logger:                 logger.Default.LogMode(logger.Silent),
	}

	// 打开数据库连接
	db, err := gorm.Open(postgres.Open(cfg.PostgreSQLConfig.DSN), gormConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to open postgresql database: %w", err)
	}

	// 配置连接池
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get sql.DB: %w", err)
	}

	if cfg.PostgreSQLConfig.PoolConfig != nil {
		sqlDB.SetMaxOpenConns(cfg.PostgreSQLConfig.PoolConfig.MaxOpenConns)
		sqlDB.SetMaxIdleConns(cfg.PostgreSQLConfig.PoolConfig.MaxIdleConns)
		sqlDB.SetConnMaxLifetime(cfg.PostgreSQLConfig.PoolConfig.ConnMaxLifetime)
		sqlDB.SetConnMaxIdleTime(cfg.PostgreSQLConfig.PoolConfig.ConnMaxIdleTime)
	}

	baseMgr := NewGormBaseManager("postgresql-database", "postgresql", db)

	return &PostgreSQLManager{
		GormBaseManager: baseMgr,
		config:          cfg,
	}, nil
}
