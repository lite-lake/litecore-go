package litemiddleware

import (
	"github.com/lite-lake/litecore-go/manager/loggermgr"
	"runtime/debug"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/lite-lake/litecore-go/common"
)

// RecoveryConfig panic 恢复配置
type RecoveryConfig struct {
	Name            *string // 中间件名称
	Order           *int    // 执行顺序
	PrintStack      *bool   // 是否打印堆栈信息
	CustomErrorBody *bool   // 是否使用自定义错误响应
	ErrorMessage    *string // 自定义错误消息
	ErrorCode       *string // 自定义错误代码
}

// DefaultRecoveryConfig 默认 panic 恢复配置
func DefaultRecoveryConfig() *RecoveryConfig {
	defaultOrder := OrderRecovery
	name := "RecoveryMiddleware"
	printStack := true
	customErrorBody := true
	errorMessage := "内部服务器错误"
	errorCode := "INTERNAL_SERVER_ERROR"
	return &RecoveryConfig{
		Name:            &name,
		Order:           &defaultOrder,
		PrintStack:      &printStack,
		CustomErrorBody: &customErrorBody,
		ErrorMessage:    &errorMessage,
		ErrorCode:       &errorCode,
	}
}

// recoveryMiddleware panic 恢复中间件
type recoveryMiddleware struct {
	LoggerMgr loggermgr.ILoggerManager `inject:""`
	cfg       *RecoveryConfig
}

// NewRecoveryMiddleware 创建 panic 恢复中间件
func NewRecoveryMiddleware(config *RecoveryConfig) common.IBaseMiddleware {
	cfg := config
	if cfg == nil {
		cfg = &RecoveryConfig{}
	}

	defaultCfg := DefaultRecoveryConfig()

	if cfg.Name == nil {
		cfg.Name = defaultCfg.Name
	}
	if cfg.Order == nil {
		cfg.Order = defaultCfg.Order
	}
	if cfg.PrintStack == nil {
		cfg.PrintStack = defaultCfg.PrintStack
	}
	if cfg.CustomErrorBody == nil {
		cfg.CustomErrorBody = defaultCfg.CustomErrorBody
	}
	if cfg.ErrorMessage == nil {
		cfg.ErrorMessage = defaultCfg.ErrorMessage
	}
	if cfg.ErrorCode == nil {
		cfg.ErrorCode = defaultCfg.ErrorCode
	}

	return &recoveryMiddleware{cfg: cfg}
}

// NewRecoveryMiddlewareWithDefaults 使用默认配置创建 panic 恢复中间件
func NewRecoveryMiddlewareWithDefaults() common.IBaseMiddleware {
	return NewRecoveryMiddleware(nil)
}

// MiddlewareName 返回中间件名称
func (m *recoveryMiddleware) MiddlewareName() string {
	if m.cfg.Name != nil && *m.cfg.Name != "" {
		return *m.cfg.Name
	}
	return "RecoveryMiddleware"
}

// Order 返回执行顺序
func (m *recoveryMiddleware) Order() int {
	if m.cfg.Order != nil {
		return *m.cfg.Order
	}
	return OrderRecovery
}

// Wrapper 返回 Gin 中间件函数
func (m *recoveryMiddleware) Wrapper() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				stack := debug.Stack()

				requestID := c.GetHeader("X-Request-ID")
				if requestID == "" {
					requestID = c.GetString("request_id")
				}

				clientIP := c.ClientIP()
				method := c.Request.Method
				path := c.Request.URL.Path
				userAgent := c.Request.UserAgent()
				query := c.Request.URL.RawQuery

				fields := []interface{}{
					"panic", err,
					"method", method,
					"path", path,
					"query", query,
					"ip", clientIP,
					"userAgent", userAgent,
					"requestID", requestID,
					"timestamp", time.Now().Format(time.RFC3339Nano),
				}

				if m.cfg.PrintStack != nil && *m.cfg.PrintStack {
					fields = append(fields, "stack", string(stack))
				}

				m.LoggerMgr.Ins().Error("PANIC recovered", fields...)

				if m.cfg.CustomErrorBody != nil && *m.cfg.CustomErrorBody {
					c.JSON(common.HTTPStatusInternalServerError, gin.H{
						"error": *m.cfg.ErrorMessage,
						"code":  *m.cfg.ErrorCode,
					})
				} else {
					c.String(common.HTTPStatusInternalServerError, *m.cfg.ErrorMessage)
				}
				c.Abort()
			}
		}()
		c.Next()
	}
}

// OnStart 服务器启动时触发
func (m *recoveryMiddleware) OnStart() error {
	return nil
}

// OnStop 服务器停止时触发
func (m *recoveryMiddleware) OnStop() error {
	return nil
}
