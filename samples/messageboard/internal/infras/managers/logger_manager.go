package managers

import (
	"github.com/lite-lake/litecore-go/common"
	"github.com/lite-lake/litecore-go/component/manager/loggermgr"
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
