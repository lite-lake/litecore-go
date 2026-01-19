package infras

import (
	"com.litelake.litecore/common"
	"com.litelake.litecore/component/manager/telemetrymgr"
)

// ITelemetryManager 遥测管理器接口
type ITelemetryManager interface {
	telemetrymgr.ITelemetryManager
}

// telemetryManagerImpl 遥测管理器实现
type telemetryManagerImpl struct {
	telemetrymgr.ITelemetryManager
}

// NewTelemetryManager 创建遥测管理器
func NewTelemetryManager(configProvider common.IBaseConfigProvider) (ITelemetryManager, error) {
	mgr, err := telemetrymgr.BuildWithConfigProvider(configProvider)
	if err != nil {
		return nil, err
	}
	return &telemetryManagerImpl{mgr}, nil
}
