package notificationmgr

import (
	"fmt"
	"time"

	"github.com/lite-lake/litecore-go/common"
	"github.com/lite-lake/litecore-go/manager/configmgr"
	"github.com/lite-lake/litecore-go/manager/loggermgr"
)

// BuildWithConfigProvider 从配置提供者创建通知管理器实例
// 读取配置节：
//   - notification.enabled: 是否启用 (bool)
//   - notification.url: Webhook URL (string)
//   - notification.timeout: HTTP 超时 (string, 如 "5s")
//   - app.name: 应用名称 (string)
//   - app.env: 应用环境 (string, 如 "local", "dev", "prod")
//
// 部署环境信息（部署环境、部署网区、部署主机、部署服务）从 DEPLOY_ 环境变量自动读取，无需配置。
func BuildWithConfigProvider(
	configProvider configmgr.IConfigManager,
	loggerMgr loggermgr.ILoggerManager,
) (INotificationManager, error) {
	if configProvider == nil {
		return nil, fmt.Errorf("configProvider cannot be nil")
	}

	// 读取通知配置
	enabled := getBoolConfig(configProvider, "notification.enabled", false)

	var url string
	var timeout time.Duration

	if enabled {
		url = getStringConfig(configProvider, "notification.url", "")
		if url == "" {
			return nil, fmt.Errorf("notification.url is required when notification is enabled")
		}

		timeoutStr := getStringConfig(configProvider, "notification.timeout", "5s")
		parsedTimeout, err := time.ParseDuration(timeoutStr)
		if err != nil {
			return nil, fmt.Errorf("invalid notification.timeout: %w", err)
		}
		timeout = parsedTimeout
	} else {
		timeout = 5 * time.Second
	}

	// 读取应用元信息
	appName := getStringConfig(configProvider, "app.name", "unknown")
	appEnv := getStringConfig(configProvider, "app.env", "unknown")

	cfg := &notificationConfig{
		Enabled: enabled,
		URL:     url,
		Timeout: timeout,
	}

	return newNotificationManager(cfg, loggerMgr, appName, appEnv), nil
}

// getStringConfig 从配置中获取字符串值
func getStringConfig(configProvider configmgr.IConfigManager, key string, defaultValue string) string {
	val, err := configProvider.Get(key)
	if err != nil {
		return defaultValue
	}
	str, err := common.GetString(val)
	if err != nil {
		return defaultValue
	}
	return str
}

// getBoolConfig 从配置中获取布尔值
func getBoolConfig(configProvider configmgr.IConfigManager, key string, defaultValue bool) bool {
	val, err := configProvider.Get(key)
	if err != nil {
		return defaultValue
	}
	boolVal, ok := val.(bool)
	if !ok {
		return defaultValue
	}
	return boolVal
}
