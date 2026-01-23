package litemiddleware

import (
	"github.com/gin-gonic/gin"

	"github.com/lite-lake/litecore-go/common"
)

// SecurityHeadersConfig 安全头配置
type SecurityHeadersConfig struct {
	Name                    *string // 中间件名称
	Order                   *int    // 执行顺序
	FrameOptions            *string // X-Frame-Options (DENY, SAMEORIGIN, ALLOW-FROM)
	ContentTypeOptions      *string // X-Content-Type-Options (nosniff)
	XSSProtection           *string // X-XSS-Protection (1; mode=block)
	ReferrerPolicy          *string // Referrer-Policy (strict-origin-when-cross-origin, no-referrer, etc)
	ContentSecurityPolicy   *string // Content-Security-Policy
	StrictTransportSecurity *string // Strict-Transport-Security (max-age=31536000; includeSubDomains)
}

// DefaultSecurityHeadersConfig 默认安全头配置
func DefaultSecurityHeadersConfig() *SecurityHeadersConfig {
	defaultOrder := OrderSecurityHeaders
	name := "SecurityHeadersMiddleware"
	frameOptions := "DENY"
	contentTypeOptions := "nosniff"
	xssProtection := "1; mode=block"
	referrerPolicy := "strict-origin-when-cross-origin"
	return &SecurityHeadersConfig{
		Name:               &name,
		Order:              &defaultOrder,
		FrameOptions:       &frameOptions,
		ContentTypeOptions: &contentTypeOptions,
		XSSProtection:      &xssProtection,
		ReferrerPolicy:     &referrerPolicy,
	}
}

// securityHeadersMiddleware 安全头中间件
type securityHeadersMiddleware struct {
	cfg *SecurityHeadersConfig
}

// NewSecurityHeadersMiddleware 创建安全头中间件
func NewSecurityHeadersMiddleware(config *SecurityHeadersConfig) common.IBaseMiddleware {
	cfg := config
	if cfg == nil {
		cfg = &SecurityHeadersConfig{}
	}

	defaultCfg := DefaultSecurityHeadersConfig()

	if cfg.Name == nil {
		cfg.Name = defaultCfg.Name
	}
	if cfg.Order == nil {
		cfg.Order = defaultCfg.Order
	}
	if cfg.FrameOptions == nil {
		cfg.FrameOptions = defaultCfg.FrameOptions
	}
	if cfg.ContentTypeOptions == nil {
		cfg.ContentTypeOptions = defaultCfg.ContentTypeOptions
	}
	if cfg.XSSProtection == nil {
		cfg.XSSProtection = defaultCfg.XSSProtection
	}
	if cfg.ReferrerPolicy == nil {
		cfg.ReferrerPolicy = defaultCfg.ReferrerPolicy
	}
	if cfg.ContentSecurityPolicy == nil {
		cfg.ContentSecurityPolicy = defaultCfg.ContentSecurityPolicy
	}
	if cfg.StrictTransportSecurity == nil {
		cfg.StrictTransportSecurity = defaultCfg.StrictTransportSecurity
	}

	return &securityHeadersMiddleware{cfg: cfg}
}

// NewSecurityHeadersMiddlewareWithDefaults 使用默认配置创建安全头中间件
func NewSecurityHeadersMiddlewareWithDefaults() common.IBaseMiddleware {
	return NewSecurityHeadersMiddleware(nil)
}

// MiddlewareName 返回中间件名称
func (m *securityHeadersMiddleware) MiddlewareName() string {
	if m.cfg.Name != nil && *m.cfg.Name != "" {
		return *m.cfg.Name
	}
	return "SecurityHeadersMiddleware"
}

// Order 返回执行顺序
func (m *securityHeadersMiddleware) Order() int {
	if m.cfg.Order != nil {
		return *m.cfg.Order
	}
	return OrderSecurityHeaders
}

// Wrapper 返回 Gin 中间件函数
func (m *securityHeadersMiddleware) Wrapper() gin.HandlerFunc {
	return func(c *gin.Context) {
		if m.cfg.FrameOptions != nil && *m.cfg.FrameOptions != "" {
			c.Writer.Header().Set("X-Frame-Options", *m.cfg.FrameOptions)
		}

		if m.cfg.ContentTypeOptions != nil && *m.cfg.ContentTypeOptions != "" {
			c.Writer.Header().Set("X-Content-Type-Options", *m.cfg.ContentTypeOptions)
		}

		if m.cfg.XSSProtection != nil && *m.cfg.XSSProtection != "" {
			c.Writer.Header().Set("X-XSS-Protection", *m.cfg.XSSProtection)
		}

		if m.cfg.ReferrerPolicy != nil && *m.cfg.ReferrerPolicy != "" {
			c.Writer.Header().Set("Referrer-Policy", *m.cfg.ReferrerPolicy)
		}

		if m.cfg.ContentSecurityPolicy != nil && *m.cfg.ContentSecurityPolicy != "" {
			c.Writer.Header().Set("Content-Security-Policy", *m.cfg.ContentSecurityPolicy)
		}

		if m.cfg.StrictTransportSecurity != nil && *m.cfg.StrictTransportSecurity != "" {
			c.Writer.Header().Set("Strict-Transport-Security", *m.cfg.StrictTransportSecurity)
		}

		c.Next()
	}
}

// OnStart 服务器启动时触发
func (m *securityHeadersMiddleware) OnStart() error {
	return nil
}

// OnStop 服务器停止时触发
func (m *securityHeadersMiddleware) OnStop() error {
	return nil
}
