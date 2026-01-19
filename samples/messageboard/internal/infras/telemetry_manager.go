package infras

import (
	"com.litelake.litecore/common"
	"com.litelake.litecore/component/manager/telemetrymgr"
)

// TelemetryManager 遥测管理器接口
type TelemetryManager interface {
	telemetrymgr.TelemetryManager
}

// telemetryManagerImpl 遥测管理器实现
type telemetryManagerImpl struct {
	telemetrymgr.TelemetryManager
}

// NewTelemetryManager 创建遥测管理器
func NewTelemetryManager(configProvider common.BaseConfigProvider) (TelemetryManager, error) {
	mgr, err := telemetrymgr.BuildWithConfigProvider(configProvider)
	if err != nil {
		return nil, err
	}
	return &telemetryManagerImpl{TelemetryManager: mgr}, nil
}
