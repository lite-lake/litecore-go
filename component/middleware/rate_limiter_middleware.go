package middleware

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/lite-lake/litecore-go/common"
	"github.com/lite-lake/litecore-go/server/builtin/manager/limitermgr"
	"github.com/lite-lake/litecore-go/server/builtin/manager/loggermgr"
)

const (
	// 响应头
	RateLimitLimitHeader     = "X-RateLimit-Limit"
	RateLimitRemainingHeader = "X-RateLimit-Remaining"
	RateLimitResetHeader     = "X-RateLimit-Reset"
)

type KeyFunc func(c *gin.Context) string
type SkipFunc func(c *gin.Context) bool

type RateLimiterConfig struct {
	Limit     int           // 时间窗口内最大请求数
	Window    time.Duration // 时间窗口大小
	KeyFunc   KeyFunc       // 自定义key生成函数（可选，默认按IP）
	SkipFunc  SkipFunc      // 跳过限流的条件（可选）
	KeyPrefix string        // key前缀，默认 "rate_limit"
}

type RateLimiterMiddleware struct {
	order      int
	LimiterMgr limitermgr.ILimiterManager `inject:""`
	LoggerMgr  loggermgr.ILoggerManager   `inject:""`
	config     *RateLimiterConfig
}

func NewRateLimiter(config *RateLimiterConfig) common.IBaseMiddleware {
	if config == nil {
		config = &RateLimiterConfig{
			Limit:     100,
			Window:    time.Minute,
			KeyPrefix: "rate_limit",
		}
	}
	if config.KeyFunc == nil {
		config.KeyFunc = func(c *gin.Context) string {
			return c.ClientIP()
		}
	}
	if config.KeyPrefix == "" {
		config.KeyPrefix = "rate_limit"
	}

	return &RateLimiterMiddleware{
		order:  40,
		config: config,
	}
}

func (m *RateLimiterMiddleware) MiddlewareName() string {
	return "RateLimiterMiddleware"
}

func (m *RateLimiterMiddleware) Order() int {
	return m.order
}

func (m *RateLimiterMiddleware) Wrapper() gin.HandlerFunc {
	return func(c *gin.Context) {
		if m.config.SkipFunc != nil && m.config.SkipFunc(c) {
			c.Next()
			return
		}

		if m.LimiterMgr == nil {
			if m.LoggerMgr != nil {
				m.LoggerMgr.Ins().Warn("限流管理器未初始化，跳过限流检查")
			}
			c.Next()
			return
		}

		key := m.config.KeyFunc(c)
		fullKey := fmt.Sprintf("%s:%s", m.config.KeyPrefix, key)

		ctx := c.Request.Context()

		allowed, err := m.LimiterMgr.Allow(ctx, fullKey, m.config.Limit, m.config.Window)
		if err != nil {
			if m.LoggerMgr != nil {
				m.LoggerMgr.Ins().Error("限流检查失败", "error", err, "key", fullKey)
			}
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "限流服务异常",
				"code":  "INTERNAL_SERVER_ERROR",
			})
			c.Abort()
			return
		}

		remaining, _ := m.LimiterMgr.GetRemaining(ctx, fullKey, m.config.Limit, m.config.Window)

		c.Header(RateLimitLimitHeader, fmt.Sprintf("%d", m.config.Limit))
		c.Header(RateLimitRemainingHeader, fmt.Sprintf("%d", remaining))

		if !allowed {
			if m.LoggerMgr != nil {
				m.LoggerMgr.Ins().Warn("请求被限流", "key", fullKey, "limit", m.config.Limit, "window", m.config.Window)
			}

			c.Header("Retry-After", fmt.Sprintf("%d", int(m.config.Window.Seconds())))

			c.JSON(http.StatusTooManyRequests, gin.H{
				"error": fmt.Sprintf("请求过于频繁，请 %v 后再试", m.config.Window),
				"code":  "RATE_LIMIT_EXCEEDED",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

func (m *RateLimiterMiddleware) OnStart() error {
	return nil
}

func (m *RateLimiterMiddleware) OnStop() error {
	return nil
}

func NewRateLimiterByIP(limit int, window time.Duration) common.IBaseMiddleware {
	return NewRateLimiter(&RateLimiterConfig{
		Limit:     limit,
		Window:    window,
		KeyPrefix: "ip",
	})
}

func NewRateLimiterByPath(limit int, window time.Duration) common.IBaseMiddleware {
	return NewRateLimiter(&RateLimiterConfig{
		Limit:     limit,
		Window:    window,
		KeyPrefix: "path",
		KeyFunc: func(c *gin.Context) string {
			return c.Request.URL.Path
		},
	})
}

func NewRateLimiterByHeader(limit int, window time.Duration, headerKey string) common.IBaseMiddleware {
	return NewRateLimiter(&RateLimiterConfig{
		Limit:     limit,
		Window:    window,
		KeyPrefix: "header",
		KeyFunc: func(c *gin.Context) string {
			return c.GetHeader(headerKey)
		},
	})
}

func NewRateLimiterByUserID(limit int, window time.Duration) common.IBaseMiddleware {
	return NewRateLimiter(&RateLimiterConfig{
		Limit:     limit,
		Window:    window,
		KeyPrefix: "user",
		KeyFunc: func(c *gin.Context) string {
			if userID, exists := c.Get("user_id"); exists {
				if uid, ok := userID.(string); ok {
					return uid
				}
			}
			return c.ClientIP()
		},
	})
}
