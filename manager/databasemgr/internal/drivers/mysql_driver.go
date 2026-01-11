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
	if cfg == nil {
		return nil, fmt.Errorf("database config is required")
	}
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid database config: %w", err)
	}

	if cfg.MySQLConfig == nil {
		return nil, fmt.Errorf("mysql_config is required")
	}

	if cfg.MySQLConfig.DSN == "" {
		return nil, fmt.Errorf("mysql DSN is required")
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
