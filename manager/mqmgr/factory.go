package mqmgr

import (
	"fmt"

	"github.com/lite-lake/litecore-go/common"
	"github.com/lite-lake/litecore-go/manager/configmgr"
	"github.com/lite-lake/litecore-go/manager/loggermgr"
	"github.com/lite-lake/litecore-go/manager/telemetrymgr"
)

// Build 根据驱动类型构建消息队列管理器
// 参数：
//   - driverType: 驱动类型 ("rabbitmq", "memory")
//   - driverConfig: 驱动配置 (根据驱动类型不同而不同)
//   - rabbitmq: 传递给 parseRabbitMQConfig 的 map[string]any
//   - memory: 传递给 parseMemoryConfig 的 map[string]any
//   - loggerMgr: 日志管理器
//   - telemetryMgr: 遥测管理器
//
// 返回 IMQManager 接口实例和可能的错误
func Build(
	driverType string,
	driverConfig map[string]any,
	loggerMgr loggermgr.ILoggerManager,
	telemetryMgr telemetrymgr.ITelemetryManager,
) (IMQManager, error) {
	switch driverType {
	case "rabbitmq":
		rabbitmqConfig, err := parseRabbitMQConfig(driverConfig)
		if err != nil {
			return nil, err
		}

		mgr, err := NewMessageQueueManagerRabbitMQImpl(rabbitmqConfig, loggerMgr, telemetryMgr)
		if err != nil {
			return nil, err
		}

		return mgr, nil

	case "memory":
		memoryConfig, err := parseMemoryConfig(driverConfig)
		if err != nil {
			return nil, err
		}

		mgr := NewMessageQueueManagerMemoryImpl(memoryConfig, loggerMgr, telemetryMgr)
		return mgr, nil

	default:
		return nil, fmt.Errorf("unsupported driver type: %s", driverType)
	}
}

// BuildWithConfigProvider 根据配置提供者构建消息队列管理器
// 自动从配置提供者读取 mq.driver 和对应驱动配置
// 参数：
//   - configProvider: 配置提供者
//   - loggerMgr: 日志管理器
//   - telemetryMgr: 遥测管理器
//
// 配置路径：
//   - mq.driver: 驱动类型 ("rabbitmq", "memory")
//   - mq.rabbitmq_config: RabbitMQ 驱动配置（当 driver=rabbitmq 时使用）
//   - mq.memory_config: Memory 驱动配置（当 driver=memory 时使用）
//
// 返回 IMQManager 接口实例和可能的错误
func BuildWithConfigProvider(
	configProvider configmgr.IConfigManager,
	loggerMgr loggermgr.ILoggerManager,
	telemetryMgr telemetrymgr.ITelemetryManager,
) (IMQManager, error) {
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

	return Build(driverTypeStr, driverConfig, loggerMgr, telemetryMgr)
}
