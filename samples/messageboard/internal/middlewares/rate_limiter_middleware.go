package middlewares

import (
	"time"

	"github.com/lite-lake/litecore-go/common"
	"github.com/lite-lake/litecore-go/component/middleware"
)

type IRateLimiterMiddleware interface {
	common.IBaseMiddleware
}

func NewRateLimiterMiddleware() IRateLimiterMiddleware {
	return middleware.NewRateLimiterMiddleware(&middleware.RateLimiterConfig{
		Limit:     100,
		Window:    time.Minute,
		KeyPrefix: "ip",
	})
}
