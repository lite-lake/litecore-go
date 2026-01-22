package loggermgr

import (
	"fmt"
	"strings"

	"github.com/lite-lake/litecore-go/server/builtin/manager/telemetrymgr"
	"gopkg.in/yaml.v3"
)

type IConfigProvider interface {
	Get(key string) (any, error)
	Has(key string) bool
}

// NewLoggerManager 创建日志管理器
func NewLoggerManager(config *Config, telemetryMgr telemetrymgr.ITelemetryManager) (ILoggerManager, error) {
	if config == nil {
		return nil, fmt.Errorf("config cannot be nil")
	}
	switch config.Driver {
	case "zap":
		return NewDriverZapLoggerManager(config.ZapConfig, telemetryMgr)
	case "default":
		return NewDriverDefaultLoggerManager(), nil
	case "none":
		return NewDriverNoneLoggerManager(), nil
	default:
		return nil, fmt.Errorf("unknown logger driver: %s", config.Driver)
	}

}

// BuildWithConfigProvider 通过配置提供者构建日志管理器
func BuildWithConfigProvider(configProvider IConfigProvider, telemetryMgr telemetrymgr.ITelemetryManager) (ILoggerManager, error) {
	if configProvider == nil {
		return nil, fmt.Errorf("configProvider cannot be nil")
	}

	driverType, err := configProvider.Get("logger.driver")
	if err != nil {
		return nil, fmt.Errorf("failed to get logger.driver: %w", err)
	}

	driverTypeStr, ok := driverType.(string)
	if !ok {
		return nil, fmt.Errorf("logger.driver must be a string, got %T", driverType)
	}

	driverTypeStr = strings.ToLower(strings.TrimSpace(driverTypeStr))

	var cfg *Config

	switch driverTypeStr {
	case "zap":
		zapConfig, err := configProvider.Get("logger.zap_config")
		if err != nil {
			return nil, fmt.Errorf("failed to get logger.zap_config: %w", err)
		}

		zapCfg := &DriverZapConfig{}
		if zapConfig != nil {
			data, err := yaml.Marshal(zapConfig)
			if err != nil {
				return nil, fmt.Errorf("failed to marshal zap config: %w", err)
			}
			if err := yaml.Unmarshal(data, zapCfg); err != nil {
				return nil, fmt.Errorf("failed to unmarshal zap config: %w", err)
			}
		}

		cfg = &Config{
			Driver:    "zap",
			ZapConfig: zapCfg,
		}

	case "default":
		cfg = &Config{
			Driver: "default",
		}

	case "none":
		cfg = &Config{
			Driver: "none",
		}

	default:
		return nil, fmt.Errorf("unsupported driver type: %s (must be zap, default or none)", driverTypeStr)
	}

	return NewLoggerManager(cfg, telemetryMgr)
}
