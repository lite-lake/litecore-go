package managers

import (
	"github.com/lite-lake/litecore-go/common"
	"github.com/lite-lake/litecore-go/component/manager/telemetrymgr"
)

type ITelemetryManager interface {
	telemetrymgr.ITelemetryManager
}

type telemetryManagerImpl struct {
	telemetrymgr.ITelemetryManager
}

func NewTelemetryManager(configProvider common.IBaseConfigProvider) (ITelemetryManager, error) {
	mgr, err := telemetrymgr.BuildWithConfigProvider(configProvider)
	if err != nil {
		return nil, err
	}
	return &telemetryManagerImpl{mgr}, nil
}
