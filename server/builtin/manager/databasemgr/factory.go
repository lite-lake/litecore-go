package databasemgr

import (
	"fmt"

	"github.com/lite-lake/litecore-go/common"
	"github.com/lite-lake/litecore-go/server/builtin/manager/configmgr"
)

// Build 创建数据库管理器实例
// driverType: 驱动类型 ("mysql", "postgresql", "sqlite", "none")
// driverConfig: 驱动配置 (根据驱动类型不同而不同)
//   - mysql: 传递给 parseMySQLConfig 的 map[string]any
//   - postgresql: 传递给 parsePostgreSQLConfig 的 map[string]any
//   - sqlite: 传递给 parseSQLiteConfig 的 map[string]any
//   - none: 忽略
//
// 返回 IDatabaseManager 接口实例和可能的错误
// 注意：loggerMgr 和 telemetryMgr 需要通过容器注入
func Build(
	driverType string,
	driverConfig map[string]any,
) (IDatabaseManager, error) {
	switch driverType {
	case "mysql":
		mysqlConfig, err := parseMySQLConfig(driverConfig)
		if err != nil {
			return nil, err
		}

		mgr, err := NewDatabaseManagerMySQLImpl(mysqlConfig)
		if err != nil {
			return nil, err
		}

		return mgr, nil

	case "postgresql":
		postgresqlConfig, err := parsePostgreSQLConfig(driverConfig)
		if err != nil {
			return nil, err
		}

		mgr, err := NewDatabaseManagerPostgreSQLImpl(postgresqlConfig)
		if err != nil {
			return nil, err
		}

		return mgr, nil

	case "sqlite":
		sqliteConfig, err := parseSQLiteConfig(driverConfig)
		if err != nil {
			return nil, err
		}

		mgr, err := NewDatabaseManagerSQLiteImpl(sqliteConfig)
		if err != nil {
			return nil, err
		}

		return mgr, nil

	case "none":
		mgr := NewDatabaseManagerNoneImpl()
		return mgr, nil

	default:
		return nil, fmt.Errorf("unsupported driver type: %s", driverType)
	}
}

// BuildWithConfigProvider 从配置提供者创建数据库管理器实例
// 自动从配置提供者读取 database.driver 和对应驱动配置
// 配置路径：
//   - database.driver: 驱动类型 ("mysql", "postgresql", "sqlite", "none")
//   - database.mysql_config: MySQL 驱动配置（当 driver=mysql 时使用）
//   - database.postgresql_config: PostgreSQL 驱动配置（当 driver=postgresql 时使用）
//   - database.sqlite_config: SQLite 驱动配置（当 driver=sqlite 时使用）
//
// 返回 IDatabaseManager 接口实例和可能的错误
// 注意：loggerMgr 和 telemetryMgr 需要通过容器注入
func BuildWithConfigProvider(configProvider configmgr.IConfigManager) (IDatabaseManager, error) {
	if configProvider == nil {
		return nil, fmt.Errorf("configProvider cannot be nil")
	}

	// 1. 读取驱动类型 database.driver
	driverType, err := configProvider.Get("database.driver")
	if err != nil {
		return nil, fmt.Errorf("failed to get database.driver: %w", err)
	}

	driverTypeStr, err := common.GetString(driverType)
	if err != nil {
		return nil, fmt.Errorf("database.driver: %w", err)
	}

	// 2. 根据驱动类型读取对应配置
	var driverConfig map[string]any

	switch driverTypeStr {
	case "mysql":
		mysqlConfig, err := configProvider.Get("database.mysql_config")
		if err != nil {
			return nil, fmt.Errorf("failed to get database.mysql_config: %w", err)
		}
		driverConfig, err = common.GetMap(mysqlConfig)
		if err != nil {
			return nil, fmt.Errorf("database.mysql_config: %w", err)
		}

	case "postgresql":
		postgresqlConfig, err := configProvider.Get("database.postgresql_config")
		if err != nil {
			return nil, fmt.Errorf("failed to get database.postgresql_config: %w", err)
		}
		driverConfig, err = common.GetMap(postgresqlConfig)
		if err != nil {
			return nil, fmt.Errorf("database.postgresql_config: %w", err)
		}

	case "sqlite":
		sqliteConfig, err := configProvider.Get("database.sqlite_config")
		if err != nil {
			return nil, fmt.Errorf("failed to get database.sqlite_config: %w", err)
		}
		driverConfig, err = common.GetMap(sqliteConfig)
		if err != nil {
			return nil, fmt.Errorf("database.sqlite_config: %w", err)
		}

	case "none":
		// none 驱动不需要配置
		driverConfig = nil

	default:
		return nil, fmt.Errorf("unsupported driver type: %s (must be mysql, postgresql, sqlite, or none)", driverTypeStr)
	}

	// 3. 调用 Build 函数创建实例
	return Build(driverTypeStr, driverConfig)
}
