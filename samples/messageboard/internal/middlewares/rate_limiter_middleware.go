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
	return litemiddleware.NewRateLimiterMiddleware(&litemiddleware.RateLimiterConfig{
		Limit:     100,
		Window:    time.Minute,
		KeyPrefix: "ip",
	})
}
