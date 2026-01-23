package litemiddleware

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
	Name      string        // 中间件名称
	Order     *int          // 执行顺序（指针类型用于判断是否设置）
	Limit     int           // 时间窗口内最大请求数
	Window    time.Duration // 时间窗口大小
	KeyFunc   KeyFunc       // 自定义key生成函数（可选，默认按IP）
	SkipFunc  SkipFunc      // 跳过限流的条件（可选）
	KeyPrefix string        // key前缀，默认 "rate_limit"
}

type rateLimiterMiddleware struct {
	LimiterMgr limitermgr.ILimiterManager `inject:""`
	LoggerMgr  loggermgr.ILoggerManager   `inject:""`
	config     *RateLimiterConfig
}

func NewRateLimiterMiddleware(config *RateLimiterConfig) common.IBaseMiddleware {
	if config == nil {
		config = DefaultRateLimiterConfig()
	}
	if config.KeyFunc == nil {
		config.KeyFunc = func(c *gin.Context) string {
			return c.ClientIP()
		}
	}
	if config.KeyPrefix == "" {
		config.KeyPrefix = "rate_limit"
	}

	return &rateLimiterMiddleware{
		config: config,
	}
}

// DefaultRateLimiterConfig 默认限流配置
func DefaultRateLimiterConfig() *RateLimiterConfig {
	defaultOrder := OrderRateLimiter
	return &RateLimiterConfig{
		Name:      "RateLimiterMiddleware",
		Order:     &defaultOrder,
		Limit:     100,
		Window:    time.Minute,
		KeyPrefix: "rate_limit",
	}
}

// NewRateLimiterMiddlewareWithDefaults 使用默认配置创建限流中间件
func NewRateLimiterMiddlewareWithDefaults() common.IBaseMiddleware {
	return NewRateLimiterMiddleware(nil)
}

func (m *rateLimiterMiddleware) MiddlewareName() string {
	if m.config.Name != "" {
		return m.config.Name
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

func (m *rateLimiterMiddleware) OnStart() error {
	return nil
}

func (m *rateLimiterMiddleware) OnStop() error {
	return nil
}
