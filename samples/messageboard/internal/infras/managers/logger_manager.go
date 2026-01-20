package managers

import (
	"com.litelake.litecore/common"
	"com.litelake.litecore/component/manager/loggermgr"
)

type ILoggerManager interface {
	loggermgr.ILoggerManager
}

type loggerManagerImpl struct {
	loggermgr.ILoggerManager
}

func NewLoggerManager(configProvider common.IBaseConfigProvider) (ILoggerManager, error) {
	mgr, err := loggermgr.BuildWithConfigProvider(configProvider)
	if err != nil {
		return nil, err
	}
	return &loggerManagerImpl{mgr}, nil
}
