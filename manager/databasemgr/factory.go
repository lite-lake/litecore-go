package databasemgr

import (
	"context"
	"database/sql"
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
// driver: 数据库驱动类型（mysql, postgresql, sqlite, none）
// cfg: 数据库配置内容
func (f *Factory) Build(driver string, cfg map[string]any) common.Manager {
	// 解析配置
	databaseConfig, err := config.ParseDatabaseConfigFromMap(cfg)
	if err != nil {
		// 配置解析失败，返回 none 管理器作为降级
		return NewDatabaseManagerAdapter(drivers.NewNoneDatabaseManager())
	}

	// 设置驱动类型
	if driver != "" {
		databaseConfig.Driver = driver
	}

	// 验证配置
	if err := databaseConfig.Validate(); err != nil {
		// 配置验证失败，返回 none 管理器作为降级
		return NewDatabaseManagerAdapter(drivers.NewNoneDatabaseManager())
	}

	// 根据驱动类型创建管理器
	var mgr interface {
		common.Manager
		DB() *sql.DB
		Driver() string
		Ping(ctx context.Context) error
		BeginTx(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error)
		Stats() sql.DBStats
		Close() error
	}

	switch databaseConfig.Driver {
	case "sqlite":
		sqliteMgr, err := drivers.NewSQLiteManager(databaseConfig)
		if err != nil {
			// 初始化失败，降级到 none 管理器
			return NewDatabaseManagerAdapter(drivers.NewNoneDatabaseManager())
		}
		mgr = sqliteMgr

	case "mysql":
		mysqlMgr, err := drivers.NewMySQLManager(databaseConfig)
		if err != nil {
			// 初始化失败，降级到 none 管理器
			return NewDatabaseManagerAdapter(drivers.NewNoneDatabaseManager())
		}
		mgr = mysqlMgr

	case "postgresql":
		postgresqlMgr, err := drivers.NewPostgreSQLManager(databaseConfig)
		if err != nil {
			// 初始化失败，降级到 none 管理器
			return NewDatabaseManagerAdapter(drivers.NewNoneDatabaseManager())
		}
		mgr = postgresqlMgr

	case "none":
		mgr = drivers.NewNoneDatabaseManager()

	default:
		// 未知驱动，返回 none 管理器
		return NewDatabaseManagerAdapter(drivers.NewNoneDatabaseManager())
	}

	// 返回适配后的管理器
	return NewDatabaseManagerAdapter(mgr)
}

// BuildWithConfig 使用配置结构体创建数据库管理器
// databaseConfig: 数据库配置结构体
func (f *Factory) BuildWithConfig(databaseConfig *config.DatabaseConfig) (common.Manager, error) {
	if err := databaseConfig.Validate(); err != nil {
		return nil, fmt.Errorf("invalid database config: %w", err)
	}

	var mgr interface {
		common.Manager
		DB() *sql.DB
		Driver() string
		Ping(ctx context.Context) error
		BeginTx(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error)
		Stats() sql.DBStats
		Close() error
	}

	switch databaseConfig.Driver {
	case "sqlite":
		sqliteMgr, err := drivers.NewSQLiteManager(databaseConfig)
		if err != nil {
			return nil, fmt.Errorf("failed to create sqlite manager: %w", err)
		}
		mgr = sqliteMgr

	case "mysql":
		mysqlMgr, err := drivers.NewMySQLManager(databaseConfig)
		if err != nil {
			return nil, fmt.Errorf("failed to create mysql manager: %w", err)
		}
		mgr = mysqlMgr

	case "postgresql":
		postgresqlMgr, err := drivers.NewPostgreSQLManager(databaseConfig)
		if err != nil {
			return nil, fmt.Errorf("failed to create postgresql manager: %w", err)
		}
		mgr = postgresqlMgr

	case "none":
		mgr = drivers.NewNoneDatabaseManager()

	default:
		return nil, fmt.Errorf("unsupported driver: %s", databaseConfig.Driver)
	}

	// 返回适配后的管理器
	return NewDatabaseManagerAdapter(mgr), nil
}