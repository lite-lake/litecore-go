package infras

import (
	"com.litelake.litecore/common"
	"com.litelake.litecore/component/manager/loggermgr"
)

// ILoggerManager 日志管理器接口
type ILoggerManager interface {
	loggermgr.ILoggerManager
}

// loggerManagerImpl 日志管理器实现
type loggerManagerImpl struct {
	loggermgr.ILoggerManager
}

// NewLoggerManager 创建日志管理器
func NewLoggerManager(configProvider common.IBaseConfigProvider) (ILoggerManager, error) {
	mgr, err := loggermgr.BuildWithConfigProvider(configProvider)
	if err != nil {
		return nil, err
	}
	return &loggerManagerImpl{mgr}, nil
}
