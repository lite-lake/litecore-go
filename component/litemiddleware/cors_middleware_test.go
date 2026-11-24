package litemiddleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestNewCorsMiddleware(t *testing.T) {
	tests := []struct {
		name string
	}{
		{
			name: "创建CORS中间件",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			middleware := NewCorsMiddleware(nil)
			assert.NotNil(t, middleware)
			assert.IsType(t, &corsMiddleware{}, middleware)
		})
	}
}

func TestCorsMiddleware_WithDefaults(t *testing.T) {
	tests := []struct {
		name string
	}{
		{
			name: "使用默认配置创建CORS中间件",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			middleware := NewCorsMiddlewareWithDefaults()
			assert.NotNil(t, middleware)
			assert.IsType(t, &corsMiddleware{}, middleware)
		})
	}
}

func TestCorsMiddleware_自定义配置(t *testing.T) {
	tests := []struct {
		name         string
		config       *CorsConfig
		expectedOrig string
	}{
		{
			name: "自定义源",
			config: func() *CorsConfig {
				allowOrigins := []string{"https://example.com"}
				return &CorsConfig{AllowOrigins: &allowOrigins}
			}(),
			expectedOrig: "https://example.com",
		},
		{
			name: "自定义多个源",
			config: func() *CorsConfig {
				allowOrigins := []string{"https://example.com", "https://test.com"}
				allowCredentials := false
				return &CorsConfig{
					AllowOrigins:     &allowOrigins,
					AllowCredentials: &allowCredentials,
				}
			}(),
			expectedOrig: "https://example.com",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			middleware := NewCorsMiddleware(tt.config).(*corsMiddleware)
			gin.SetMode(gin.TestMode)
			router := gin.New()
			router.Use(middleware.Wrapper())
			router.GET("/test", func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{"message": "ok"})
			})

			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			req.Header.Set("Origin", tt.expectedOrig)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedOrig, w.Header().Get("Access-Control-Allow-Origin"))
		})
	}
}

func TestCorsMiddleware_MiddlewareName(t *testing.T) {
	tests := []struct {
		name     string
		expected string
	}{
		{
			name:     "返回中间件名称",
			expected: "CorsMiddleware",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			middleware := NewCorsMiddleware(nil).(*corsMiddleware)
			assert.Equal(t, tt.expected, middleware.MiddlewareName())
		})
	}
}

func TestCorsMiddleware_Order(t *testing.T) {
	tests := []struct {
		name     string
		expected int
	}{
		{
			name:     "返回执行顺序",
			expected: OrderCORS,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			middleware := NewCorsMiddleware(nil).(*corsMiddleware)
			assert.Equal(t, tt.expected, middleware.Order())
		})
	}
}

func TestCorsMiddleware_Wrapper(t *testing.T) {
	tests := []struct {
		name               string
		method             string
		expectedStatusCode int
		checkHeaders       bool
	}{
		{
			name:               "GET请求设置CORS头",
			method:             http.MethodGet,
			expectedStatusCode: http.StatusOK,
			checkHeaders:       true,
		},
		{
			name:               "POST请求设置CORS头",
			method:             http.MethodPost,
			expectedStatusCode: http.StatusOK,
			checkHeaders:       true,
		},
		{
			name:               "OPTIONS请求返回NoContent",
			method:             http.MethodOptions,
			expectedStatusCode: http.StatusNoContent,
			checkHeaders:       true,
		},
		{
			name:               "PUT请求设置CORS头",
			method:             http.MethodPut,
			expectedStatusCode: http.StatusOK,
			checkHeaders:       true,
		},
		{
			name:               "DELETE请求设置CORS头",
			method:             http.MethodDelete,
			expectedStatusCode: http.StatusOK,
			checkHeaders:       true,
		},
		{
			name:               "PATCH请求设置CORS头",
			method:             http.MethodPatch,
			expectedStatusCode: http.StatusOK,
			checkHeaders:       true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gin.SetMode(gin.TestMode)
			router := gin.New()
			middleware := NewCorsMiddleware(nil).(*corsMiddleware)
			router.Use(middleware.Wrapper())
			router.GET("/test", func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{"message": "ok"})
			})
			router.POST("/test", func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{"message": "ok"})
			})
			router.PUT("/test", func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{"message": "ok"})
			})
			router.DELETE("/test", func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{"message": "ok"})
			})
			router.PATCH("/test", func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{"message": "ok"})
			})
			router.OPTIONS("/test", func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{"message": "ok"})
			})

			req := httptest.NewRequest(tt.method, "/test", nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatusCode, w.Code)

			if tt.checkHeaders {
				assert.Equal(t, "*", w.Header().Get("Access-Control-Allow-Origin"))
				// 默认配置 AllowOrigins 为 ["*"]，安全修复后 AllowCredentials 自动降级为 false
				assert.Empty(t, w.Header().Get("Access-Control-Allow-Credentials"))
				assert.Contains(t, w.Header().Get("Access-Control-Allow-Headers"), "Content-Type")
				assert.Contains(t, w.Header().Get("Access-Control-Allow-Methods"), "GET")
				assert.Contains(t, w.Header().Get("Access-Control-Allow-Methods"), "POST")
				assert.Contains(t, w.Header().Get("Access-Control-Allow-Methods"), "PUT")
				assert.Contains(t, w.Header().Get("Access-Control-Allow-Methods"), "DELETE")
				assert.Contains(t, w.Header().Get("Access-Control-Allow-Methods"), "PATCH")
				assert.Contains(t, w.Header().Get("Access-Control-Allow-Methods"), "OPTIONS")
			}
		})
	}
}

func TestCorsMiddleware_OnStart(t *testing.T) {
	tests := []struct {
		name string
	}{
		{
			name: "服务器启动回调",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			middleware := NewCorsMiddleware(nil).(*corsMiddleware)
			err := middleware.OnStart()
			assert.NoError(t, err)
		})
	}
}

// TestCorsMiddleware_SecurityGuard_WildcardWithCredentials
// 安全护栏测试：AllowOrigins 含 "*" 时，显式设置 AllowCredentials: true 应被自动降级为 false
func TestCorsMiddleware_SecurityGuard_WildcardWithCredentials(t *testing.T) {
	allowOrigins := []string{"*"}
	allowCredentials := true
	config := &CorsConfig{
		AllowOrigins:     &allowOrigins,
		AllowCredentials: &allowCredentials,
	}

	middleware := NewCorsMiddleware(config).(*corsMiddleware)

	// 安全护栏应将 AllowCredentials 降级为 false
	assert.False(t, *middleware.cfg.AllowCredentials, "AllowCredentials 含 \"*\" 时应被自动降级为 false")
}

// TestCorsMiddleware_NoOriginReflection
// 安全修复验证：使用默认配置时，恶意 Origin 不应被反射到 ACAO 头
func TestCorsMiddleware_NoOriginReflection(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	middleware := NewCorsMiddlewareWithDefaults().(*corsMiddleware)
	router.Use(middleware.Wrapper())
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "ok"})
	})

	// 模拟恶意 Origin
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Origin", "https://evil-attacker.com")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// ACAO 应为字面值 "*"，而非反射恶意 Origin
	assert.Equal(t, "*", w.Header().Get("Access-Control-Allow-Origin"))
	assert.NotEqual(t, "https://evil-attacker.com", w.Header().Get("Access-Control-Allow-Origin"))
	// 默认配置不应允许凭证
	assert.Empty(t, w.Header().Get("Access-Control-Allow-Credentials"))
}

func TestCorsMiddleware_OnStop(t *testing.T) {
	tests := []struct {
		name string
	}{
		{
			name: "服务器停止回调",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			middleware := NewCorsMiddleware(nil).(*corsMiddleware)
			err := middleware.OnStop()
			assert.NoError(t, err)
		})
	}
}
