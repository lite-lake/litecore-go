package databasemgr

import (
	"fmt"
	"time"

	"github.com/lite-lake/litecore-go/common"
	"github.com/lite-lake/litecore-go/manager/configmgr"
	"github.com/lite-lake/litecore-go/manager/loggermgr"
	"github.com/lite-lake/litecore-go/manager/telemetrymgr"
)

// migrateGateConfigurer 迁移门控配置注入接口（由 factory 调用，不对外暴露）
type migrateGateConfigurer interface {
	setMigrateGateConfig(forceMigrate bool, migrateLockTimeout time.Duration)
}

// applyMigrateGateConfig 将数据库配置中的迁移门控选项注入管理器
func applyMigrateGateConfig(mgr IDatabaseManager, cfg *DatabaseConfig) {
	if gate, ok := mgr.(migrateGateConfigurer); ok {
		gate.setMigrateGateConfig(cfg.ForceMigrate, cfg.MigrateLockTimeout)
	}
}

// parseMigrateGateConfig 从配置提供者解析迁移门控配置（force_migrate / migrate_lock_timeout）
// 使用单个 key 读取，兼容扁平 mock 与支持点路径的 configmgr 实现
func parseMigrateGateConfig(provider configmgr.IConfigManager) (*DatabaseConfig, error) {
	cfg := DefaultConfig()

	if fm, err := provider.Get("database.force_migrate"); err == nil {
		if b, ok := fm.(bool); ok {
			cfg.ForceMigrate = b
		}
	}

	if lt, err := provider.Get("database.migrate_lock_timeout"); err == nil {
		switch v := lt.(type) {
		case string:
			duration, err := time.ParseDuration(v)
			if err != nil {
				return nil, fmt.Errorf("invalid migrate_lock_timeout format: %s", v)
			}
			cfg.MigrateLockTimeout = duration
		case int:
			cfg.MigrateLockTimeout = time.Duration(v) * time.Second
		case float64:
			cfg.MigrateLockTimeout = time.Duration(v) * time.Second
		}
	}

	return cfg, nil
}

// Build 创建数据库管理器实例
// 参数：
//   - driverType: 驱动类型 ("mysql", "postgresql", "sqlite", "none")
//   - driverConfig: 驱动配置 (根据驱动类型不同而不同)
//   - mysql: 传递给 parseMySQLConfig 的 map[string]any
//   - postgresql: 传递给 parsePostgreSQLConfig 的 map[string]any
//   - sqlite: 传递给 parseSQLiteConfig 的 map[string]any
//   - none: 忽略
//   - loggerMgr: 日志管理器
//   - telemetryMgr: 遥测管理器
//
// 返回 IDatabaseManager 接口实例和可能的错误
func Build(
	driverType string,
	driverConfig map[string]any,
	loggerMgr loggermgr.ILoggerManager,
	telemetryMgr telemetrymgr.ITelemetryManager,
) (IDatabaseManager, error) {
	switch driverType {
	case "mysql":
		mysqlConfig, err := parseMySQLConfig(driverConfig)
		if err != nil {
			return nil, err
		}

		mgr, err := NewDatabaseManagerMySQLImpl(mysqlConfig, loggerMgr, telemetryMgr)
		if err != nil {
			return nil, err
		}

		return mgr, nil

	case "postgresql":
		postgresqlConfig, err := parsePostgreSQLConfig(driverConfig)
		if err != nil {
			return nil, err
		}

		mgr, err := NewDatabaseManagerPostgreSQLImpl(postgresqlConfig, loggerMgr, telemetryMgr)
		if err != nil {
			return nil, err
		}

		return mgr, nil

	case "sqlite":
		sqliteConfig, err := parseSQLiteConfig(driverConfig)
		if err != nil {
			return nil, err
		}

		mgr, err := NewDatabaseManagerSQLiteImpl(sqliteConfig, loggerMgr, telemetryMgr)
		if err != nil {
			return nil, err
		}

		return mgr, nil

	case "none":
		mgr := NewDatabaseManagerNoneImpl(loggerMgr, telemetryMgr)
		return mgr, nil

	default:
		return nil, fmt.Errorf("unsupported driver type: %s", driverType)
	}
}

// BuildWithConfigProvider 从配置提供者创建数据库管理器实例
// 自动从配置提供者读取 database.driver 和对应驱动配置
// 参数：
//   - configProvider: 配置提供者
//   - loggerMgr: 日志管理器
//   - telemetryMgr: 遥测管理器
//
// 配置路径：
//   - database.driver: 驱动类型 ("mysql", "postgresql", "sqlite", "none")
//   - database.mysql_config: MySQL 驱动配置（当 driver=mysql 时使用）
//   - database.postgresql_config: PostgreSQL 驱动配置（当 driver=postgresql 时使用）
//   - database.sqlite_config: SQLite 驱动配置（当 driver=sqlite 时使用）
//
// 返回 IDatabaseManager 接口实例和可能的错误
func BuildWithConfigProvider(
	configProvider configmgr.IConfigManager,
	loggerMgr loggermgr.ILoggerManager,
	telemetryMgr telemetrymgr.ITelemetryManager,
) (IDatabaseManager, error) {
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

	// 解析迁移门控配置（force_migrate / migrate_lock_timeout）
	gateCfg, err := parseMigrateGateConfig(configProvider)
	if err != nil {
		return nil, err
	}

	// 3. 调用 Build 函数创建实例
	mgr, err := Build(driverTypeStr, driverConfig, loggerMgr, telemetryMgr)
	if err != nil {
		return nil, err
	}

	// 注入迁移门控配置
	applyMigrateGateConfig(mgr, gateCfg)

	return mgr, nil
}
