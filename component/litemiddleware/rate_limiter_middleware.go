package litemiddleware

import (
	"fmt"
	"github.com/lite-lake/litecore-go/manager/limitermgr"
	"github.com/lite-lake/litecore-go/manager/loggermgr"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/lite-lake/litecore-go/common"
)

const (
	// 速率限制响应头
	RateLimitLimitHeader     = "X-RateLimit-Limit"
	RateLimitRemainingHeader = "X-RateLimit-Remaining"
	RateLimitResetHeader     = "X-RateLimit-Reset"
)

type KeyFunc func(c *gin.Context) string
type SkipFunc func(c *gin.Context) bool

type RateLimiterConfig struct {
	Name      *string        // 中间件名称
	Order     *int           // 执行顺序
	Limit     *int           // 时间窗口内最大请求数
	Window    *time.Duration // 时间窗口大小
	KeyFunc   KeyFunc        // 自定义key生成函数（可选，默认按IP）
	SkipFunc  SkipFunc       // 跳过限流的条件（可选）
	KeyPrefix *string        // key前缀
}

type rateLimiterMiddleware struct {
	LimiterMgr limitermgr.ILimiterManager `inject:""`
	LoggerMgr  loggermgr.ILoggerManager   `inject:""`
	config     *RateLimiterConfig
}

func NewRateLimiterMiddleware(config *RateLimiterConfig) common.IBaseMiddleware {
	cfg := config
	if cfg == nil {
		cfg = &RateLimiterConfig{}
	}

	defaultCfg := DefaultRateLimiterConfig()

	if cfg.Name == nil {
		cfg.Name = defaultCfg.Name
	}
	if cfg.Order == nil {
		cfg.Order = defaultCfg.Order
	}
	if cfg.Limit == nil {
		cfg.Limit = defaultCfg.Limit
	}
	if cfg.Window == nil {
		cfg.Window = defaultCfg.Window
	}
	if cfg.KeyPrefix == nil {
		cfg.KeyPrefix = defaultCfg.KeyPrefix
	}
	if cfg.KeyFunc == nil {
		cfg.KeyFunc = func(c *gin.Context) string {
			return c.ClientIP()
		}
	}

	return &rateLimiterMiddleware{
		config: cfg,
	}
}

// DefaultRateLimiterConfig 默认限流配置
func DefaultRateLimiterConfig() *RateLimiterConfig {
	defaultOrder := OrderRateLimiter
	name := "RateLimiterMiddleware"
	limit := 100
	window := time.Minute
	keyPrefix := "rate_limit"
	return &RateLimiterConfig{
		Name:      &name,
		Order:     &defaultOrder,
		Limit:     &limit,
		Window:    &window,
		KeyPrefix: &keyPrefix,
	}
}

// NewRateLimiterMiddlewareWithDefaults 使用默认配置创建限流中间件
func NewRateLimiterMiddlewareWithDefaults() common.IBaseMiddleware {
	return NewRateLimiterMiddleware(nil)
}

func (m *rateLimiterMiddleware) MiddlewareName() string {
	if m.config.Name != nil && *m.config.Name != "" {
		return *m.config.Name
	}
	return "RateLimiterMiddleware"
}

func (m *rateLimiterMiddleware) Order() int {
	if m.config.Order != nil {
		return *m.config.Order
	}
	return OrderRateLimiter
}

func (m *rateLimiterMiddleware) Wrapper() gin.HandlerFunc {
	return func(c *gin.Context) {
		if m.config.SkipFunc != nil && m.config.SkipFunc(c) {
			c.Next()
			return
		}

		if m.LimiterMgr == nil {
			m.LoggerMgr.Ins().Warn("Rate limiter manager not initialized, skipping rate limit check")
			c.Next()
			return
		}

		key := m.config.KeyFunc(c)
		fullKey := fmt.Sprintf("%s:%s", *m.config.KeyPrefix, key)

		ctx := c.Request.Context()

		allowed, err := m.LimiterMgr.Allow(ctx, fullKey, *m.config.Limit, *m.config.Window)
		if err != nil {
			m.LoggerMgr.Ins().Error("Rate limit check failed", "error", err, "key", fullKey)
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Rate limiter service error",
				"code":  "INTERNAL_SERVER_ERROR",
			})
			c.Abort()
			return
		}

		remaining, _ := m.LimiterMgr.GetRemaining(ctx, fullKey, *m.config.Limit, *m.config.Window)

		c.Header(RateLimitLimitHeader, fmt.Sprintf("%d", *m.config.Limit))
		c.Header(RateLimitRemainingHeader, fmt.Sprintf("%d", remaining))

		if !allowed {
			m.LoggerMgr.Ins().Warn("Request rate limited", "key", fullKey, "limit", *m.config.Limit, "window", *m.config.Window)

			c.Header("Retry-After", fmt.Sprintf("%d", int(m.config.Window.Seconds())))

			c.JSON(http.StatusTooManyRequests, gin.H{
				"error": fmt.Sprintf("Too many requests, please try again after %v", *m.config.Window),
				"code":  "RATE_LIMIT_EXCEEDED",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

func (m *rateLimiterMiddleware) OnStart() error {
	return nil
}

func (m *rateLimiterMiddleware) OnStop() error {
	return nil
}
