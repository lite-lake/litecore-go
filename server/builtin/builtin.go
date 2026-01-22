package builtin

import (
	"fmt"

	"github.com/lite-lake/litecore-go/server/builtin/manager/cachemgr"
	"github.com/lite-lake/litecore-go/server/builtin/manager/configmgr"
	"github.com/lite-lake/litecore-go/server/builtin/manager/databasemgr"
	"github.com/lite-lake/litecore-go/server/builtin/manager/loggermgr"
	"github.com/lite-lake/litecore-go/server/builtin/manager/telemetrymgr"

	"github.com/lite-lake/litecore-go/util/logger"
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
	LoggerRegistry *logger.LoggerRegistry // 日志注册器

	ConfigProvider configmgr.IConfigManager       // 配置管理器
	LoggerManager  loggermgr.ILoggerManager       // 日志管理器
	TelemetryMgr   telemetrymgr.ITelemetryManager // 追踪管理器
	DatabaseMgr    databasemgr.IDatabaseManager   // 数据库管理器
	CacheMgr       cachemgr.ICacheManager         // 缓存管理器
}

func Initialize(cfg *Config) (*Components, error) {
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configmgr: %w", err)
	}

	loggerRegistry := logger.NewLoggerRegistry()

	configProvider, err := configmgr.NewConfigManager(cfg.Driver, cfg.FilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to create configmgr provider: %w", err)
	}

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

func (c *Components) GetConfigProvider() configmgr.IConfigManager {
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
