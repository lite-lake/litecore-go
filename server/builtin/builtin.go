package builtin

import (
	"fmt"
	"github.com/lite-lake/litecore-go/container"
	"github.com/lite-lake/litecore-go/server/builtin/manager/cachemgr"
	"github.com/lite-lake/litecore-go/server/builtin/manager/configmgr"
	"github.com/lite-lake/litecore-go/server/builtin/manager/databasemgr"
	"github.com/lite-lake/litecore-go/server/builtin/manager/loggermgr"
	"github.com/lite-lake/litecore-go/server/builtin/manager/telemetrymgr"
)

type Config struct {
	Driver   string
	FilePath string
}

func (c *Config) Validate() error {
	if c.Driver == "" {
		return fmt.Errorf("configmgr driver cannot be empty")
	}
	if c.FilePath == "" {
		return fmt.Errorf("configmgr file path cannot be empty")
	}
	return nil
}

// Components 内置组件
type Components struct {
	ConfigMgr    configmgr.IConfigManager       // 配置管理器
	LoggerMgr    loggermgr.ILoggerManager       // 日志管理器
	TelemetryMgr telemetrymgr.ITelemetryManager // 追踪管理器
	DatabaseMgr  databasemgr.IDatabaseManager   // 数据库管理器
	CacheMgr     cachemgr.ICacheManager         // 缓存管理器
}

func Initialize(cfg *Config) (*container.ManagerContainer, error) {

	cntr := container.NewManagerContainer()

	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configmgr: %w", err)
	}

	configManager, err := configmgr.NewConfigManager(cfg.Driver, cfg.FilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to create config manager: %w", err)
	}
	if err := container.RegisterManager[configmgr.IConfigManager](cntr, configManager); err != nil {
		return nil, fmt.Errorf("failed to register config manager: %w", err)
	}

	telemetryMgr, err := telemetrymgr.BuildWithConfigProvider(configManager)
	if err != nil {
		return nil, fmt.Errorf("failed to create telemetry manager: %w", err)
	}
	if err := container.RegisterManager[telemetrymgr.ITelemetryManager](cntr, telemetryMgr); err != nil {
		return nil, fmt.Errorf("failed to register telemetry manager: %w", err)
	}

	loggerManager, err := loggermgr.BuildWithConfigProvider(configManager, telemetryMgr)
	if err != nil {
		return nil, fmt.Errorf("failed to create logger manager: %w", err)
	}
	if err := container.RegisterManager[loggermgr.ILoggerManager](cntr, loggerManager); err != nil {
		return nil, fmt.Errorf("failed to register logger manager: %w", err)
	}

	databaseMgr, err := databasemgr.BuildWithConfigProvider(configManager)
	if err != nil {
		return nil, fmt.Errorf("failed to create database manager: %w", err)
	}
	if err := container.RegisterManager[databasemgr.IDatabaseManager](cntr, databaseMgr); err != nil {
		return nil, fmt.Errorf("failed to register database manager: %w", err)
	}

	cacheMgr, err := cachemgr.BuildWithConfigProvider(configManager)
	if err != nil {
		return nil, fmt.Errorf("failed to create cache manager: %w", err)
	}
	if err := container.RegisterManager[cachemgr.ICacheManager](cntr, cacheMgr); err != nil {
		return nil, fmt.Errorf("failed to register cache manager: %w", err)
	}

	return cntr, nil
}

func (c *Components) GetConfigProvider() configmgr.IConfigManager {
	return c.ConfigMgr
}

func (c *Components) GetManagers() []interface{} {
	return []interface{}{
		c.LoggerMgr,
		c.TelemetryMgr,
		c.DatabaseMgr,
		c.CacheMgr,
	}
}
