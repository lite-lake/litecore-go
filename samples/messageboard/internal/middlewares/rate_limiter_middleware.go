package middlewares

import (
	"time"

	"github.com/lite-lake/litecore-go/common"
	"github.com/lite-lake/litecore-go/component/litemiddleware"
)

// IRateLimiterMiddleware 限流中间件接口
type IRateLimiterMiddleware interface {
	common.IBaseMiddleware
}

// NewRateLimiterMiddleware 创建限流中间件
// 配置：每 IP 每分钟最多 100 次请求
func NewRateLimiterMiddleware() IRateLimiterMiddleware {
	limit := 100
	window := time.Minute
	keyPrefix := "ip"
	return litemiddleware.NewRateLimiterMiddleware(&litemiddleware.RateLimiterConfig{
		Limit:     &limit,
		Window:    &window,
		KeyPrefix: &keyPrefix,
	})
}
