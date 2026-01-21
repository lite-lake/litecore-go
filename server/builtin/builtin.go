package builtin

import (
	"fmt"

	"github.com/lite-lake/litecore-go/common"
	"github.com/lite-lake/litecore-go/component/manager/cachemgr"
	"github.com/lite-lake/litecore-go/component/manager/databasemgr"
	"github.com/lite-lake/litecore-go/component/manager/loggermgr"
	"github.com/lite-lake/litecore-go/component/manager/telemetrymgr"
	"github.com/lite-lake/litecore-go/config"
	"github.com/lite-lake/litecore-go/util/logger"
)

type Config struct {
	Driver   string
	FilePath string
}

func (c *Config) Validate() error {
	if c.Driver == "" {
		return fmt.Errorf("config driver cannot be empty")
	}
	if c.FilePath == "" {
		return fmt.Errorf("config file path cannot be empty")
	}
	return nil
}

type Components struct {
	ConfigProvider common.IBaseConfigProvider
	LoggerManager  loggermgr.ILoggerManager
	LoggerRegistry *logger.LoggerRegistry
	TelemetryMgr   telemetrymgr.ITelemetryManager
	DatabaseMgr    databasemgr.IDatabaseManager
	CacheMgr       cachemgr.ICacheManager
}

func Initialize(cfg *Config) (*Components, error) {
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	configProvider, err := config.NewConfigProvider(cfg.Driver, cfg.FilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to create config provider: %w", err)
	}

	loggerRegistry := logger.NewLoggerRegistry()

	loggerManager, err := loggermgr.BuildWithConfigProvider(configProvider)
	if err != nil {
		return nil, fmt.Errorf("failed to create logger manager: %w", err)
	}

	loggerRegistry.SetLoggerManager(loggerManager)

	telemetryMgr, err := telemetrymgr.BuildWithConfigProvider(configProvider)
	if err != nil {
		return nil, fmt.Errorf("failed to create telemetry manager: %w", err)
	}

	databaseMgr, err := databasemgr.BuildWithConfigProvider(configProvider)
	if err != nil {
		return nil, fmt.Errorf("failed to create database manager: %w", err)
	}

	cacheMgr, err := cachemgr.BuildWithConfigProvider(configProvider)
	if err != nil {
		return nil, fmt.Errorf("failed to create cache manager: %w", err)
	}

	return &Components{
		ConfigProvider: configProvider,
		LoggerManager:  loggerManager,
		LoggerRegistry: loggerRegistry,
		TelemetryMgr:   telemetryMgr,
		DatabaseMgr:    databaseMgr,
		CacheMgr:       cacheMgr,
	}, nil
}

func (c *Components) GetConfigProvider() common.IBaseConfigProvider {
	return c.ConfigProvider
}

func (c *Components) GetManagers() []interface{} {
	return []interface{}{
		c.LoggerManager,
		c.TelemetryMgr,
		c.DatabaseMgr,
		c.CacheMgr,
	}
}
