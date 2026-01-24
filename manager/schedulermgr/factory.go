package schedulermgr

import (
	"fmt"
	"github.com/lite-lake/litecore-go/manager/configmgr"
	"strings"

	"github.com/lite-lake/litecore-go/common"
	"gopkg.in/yaml.v3"
)

func Build(config *CronConfig) (ISchedulerManager, error) {
	if config == nil {
		config = &CronConfig{
			ValidateOnStartup: true,
		}
	}

	return NewSchedulerManagerCronImpl(config), nil
}

func BuildWithConfigProvider(configProvider configmgr.IConfigManager) (ISchedulerManager, error) {
	if configProvider == nil {
		return nil, fmt.Errorf("configProvider cannot be nil")
	}

	driverType, err := configProvider.Get("scheduler.driver")
	if err != nil {
		return nil, fmt.Errorf("failed to get scheduler.driver: %w", err)
	}

	driverTypeStr, err := common.GetString(driverType)
	if err != nil {
		return nil, fmt.Errorf("scheduler.driver: %w", err)
	}

	driverTypeStr = strings.ToLower(strings.TrimSpace(driverTypeStr))

	if driverTypeStr != "cron" {
		return nil, fmt.Errorf("unsupported scheduler driver: %s (must be cron)", driverTypeStr)
	}

	cronConfig, err := configProvider.Get("scheduler.cron_config")
	if err != nil {
		return nil, fmt.Errorf("failed to get scheduler.cron_config: %w", err)
	}

	cfg := &CronConfig{}
	if cronConfig != nil {
		data, err := yaml.Marshal(cronConfig)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal cron config: %w", err)
		}
		if err := yaml.Unmarshal(data, cfg); err != nil {
			return nil, fmt.Errorf("failed to unmarshal cron config: %w", err)
		}
	}

	return Build(cfg)
}
