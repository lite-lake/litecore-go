package infras

import (
	"com.litelake.litecore/common"
	"com.litelake.litecore/component/manager/loggermgr"
)

// LoggerManager 日志管理器接口
type LoggerManager interface {
	loggermgr.LoggerManager
}

// loggerManagerImpl 日志管理器实现
type loggerManagerImpl struct {
	loggermgr.LoggerManager
}

// NewLoggerManager 创建日志管理器
func NewLoggerManager(configProvider common.BaseConfigProvider) (LoggerManager, error) {
	mgr, err := loggermgr.BuildWithConfigProvider(configProvider)
	if err != nil {
		return nil, err
	}
	return &loggerManagerImpl{LoggerManager: mgr}, nil
}
