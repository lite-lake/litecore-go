package middlewares

import (
	"github.com/lite-lake/litecore-go/common"
	componentMiddleware "github.com/lite-lake/litecore-go/component/middleware"
)

type ITelemetryMiddleware interface {
	common.IBaseMiddleware
}

func NewTelemetryMiddleware() ITelemetryMiddleware {
	return componentMiddleware.NewTelemetryMiddlewareWithDefaults()
}
