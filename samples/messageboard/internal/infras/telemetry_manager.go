package infras

import (
	"com.litelake.litecore/common"
	"com.litelake.litecore/manager/telemetrymgr"
)

// NewTelemetryManager 创建遥测管理器
func NewTelemetryManager(configProvider common.BaseConfigProvider) (telemetrymgr.TelemetryManager, error) {
	return telemetrymgr.BuildWithConfigProvider(configProvider)
}
