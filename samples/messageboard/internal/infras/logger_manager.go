package infras

import (
	"com.litelake.litecore/common"
	"com.litelake.litecore/manager/loggermgr"
)

// NewLoggerManager 创建日志管理器
func NewLoggerManager(configProvider common.BaseConfigProvider) (loggermgr.LoggerManager, error) {
	return loggermgr.BuildWithConfigProvider(configProvider)
}
