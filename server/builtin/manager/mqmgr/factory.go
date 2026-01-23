package mqmgr

import (
	"fmt"

	"github.com/lite-lake/litecore-go/common"
	"github.com/lite-lake/litecore-go/server/builtin/manager/configmgr"
)

func Build(
	driverType string,
	driverConfig map[string]any,
) (IMQManager, error) {
	switch driverType {
	case "rabbitmq":
		rabbitmqConfig, err := parseRabbitMQConfig(driverConfig)
		if err != nil {
			return nil, err
		}

		mgr, err := NewMessageQueueManagerRabbitMQImpl(rabbitmqConfig)
		if err != nil {
			return nil, err
		}

		return mgr, nil

	case "memory":
		memoryConfig, err := parseMemoryConfig(driverConfig)
		if err != nil {
			return nil, err
		}

		mgr := NewMessageQueueManagerMemoryImpl(memoryConfig)
		return mgr, nil

	default:
		return nil, fmt.Errorf("unsupported driver type: %s", driverType)
	}
}

func BuildWithConfigProvider(configProvider configmgr.IConfigManager) (IMQManager, error) {
	if configProvider == nil {
		return nil, fmt.Errorf("configProvider cannot be nil")
	}

	driverType, err := configProvider.Get("mq.driver")
	if err != nil {
		return nil, fmt.Errorf("failed to get mq.driver: %w", err)
	}

	driverTypeStr, err := common.GetString(driverType)
	if err != nil {
		return nil, fmt.Errorf("mq.driver: %w", err)
	}

	var driverConfig map[string]any

	switch driverTypeStr {
	case "rabbitmq":
		rabbitmqConfig, err := configProvider.Get("mq.rabbitmq_config")
		if err != nil {
			return nil, fmt.Errorf("failed to get mq.rabbitmq_config: %w", err)
		}
		driverConfig, err = common.GetMap(rabbitmqConfig)
		if err != nil {
			return nil, fmt.Errorf("mq.rabbitmq_config: %w", err)
		}

	case "memory":
		memoryConfig, err := configProvider.Get("mq.memory_config")
		if err != nil {
			return nil, fmt.Errorf("failed to get mq.memory_config: %w", err)
		}
		driverConfig, err = common.GetMap(memoryConfig)
		if err != nil {
			return nil, fmt.Errorf("mq.memory_config: %w", err)
		}

	default:
		return nil, fmt.Errorf("unsupported driver type: %s (must be rabbitmq or memory)", driverTypeStr)
	}

	return Build(driverTypeStr, driverConfig)
}
