package databasemgr

import (
	"fmt"

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
func (f *Factory) Build(driver string, cfg map[string]any) common.BaseManager {
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

// BuildWithConfig 使用配置结构体创建数据库管理器
func (f *Factory) BuildWithConfig(databaseConfig *config.DatabaseConfig) (DatabaseManager, error) {
	if databaseConfig == nil {
		return nil, fmt.Errorf("database config is required")
	}
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
		return drivers.NewNoneDatabaseManager(), nil
	default:
		return drivers.NewNoneDatabaseManager(), nil
	}
}
