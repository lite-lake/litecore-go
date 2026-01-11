package drivers

import (
	"fmt"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"com.litelake.litecore/manager/databasemgr/internal/config"
)

// SQLiteManager SQLite 数据库管理器
type SQLiteManager struct {
	*GormBaseManager
	config *config.DatabaseConfig
}

// NewSQLiteManager 创建 SQLite 数据库管理器
func NewSQLiteManager(cfg *config.DatabaseConfig) (*SQLiteManager, error) {
	if cfg == nil {
		return nil, fmt.Errorf("database config is required")
	}
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid database config: %w", err)
	}

	if cfg.SQLiteConfig == nil {
		return nil, fmt.Errorf("sqlite_config is required")
	}

	if cfg.SQLiteConfig.DSN == "" {
		return nil, fmt.Errorf("sqlite DSN is required")
	}

	// GORM 配置
	gormConfig := &gorm.Config{
		SkipDefaultTransaction: true,
		Logger:                 logger.Default.LogMode(logger.Silent),
	}

	// 打开数据库连接
	db, err := gorm.Open(sqlite.Open(cfg.SQLiteConfig.DSN), gormConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to open sqlite database: %w", err)
	}

	// 配置连接池
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get sql.DB: %w", err)
	}

	if cfg.SQLiteConfig.PoolConfig != nil {
		sqlDB.SetMaxOpenConns(cfg.SQLiteConfig.PoolConfig.MaxOpenConns)
		sqlDB.SetMaxIdleConns(cfg.SQLiteConfig.PoolConfig.MaxIdleConns)
		sqlDB.SetConnMaxLifetime(cfg.SQLiteConfig.PoolConfig.ConnMaxLifetime)
		sqlDB.SetConnMaxIdleTime(cfg.SQLiteConfig.PoolConfig.ConnMaxIdleTime)
	}

	baseMgr := NewGormBaseManager("sqlite-database", "sqlite", db)

	return &SQLiteManager{
		GormBaseManager: baseMgr,
		config:          cfg,
	}, nil
}
