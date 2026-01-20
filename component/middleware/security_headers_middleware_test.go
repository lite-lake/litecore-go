package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestNewSecurityHeadersMiddleware(t *testing.T) {
	tests := []struct {
		name string
	}{
		{
			name: "创建安全头中间件",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			middleware := NewSecurityHeadersMiddleware()
			assert.NotNil(t, middleware)
			assert.IsType(t, &SecurityHeadersMiddleware{}, middleware)
		})
	}
}

func TestSecurityHeadersMiddleware_MiddlewareName(t *testing.T) {
	tests := []struct {
		name     string
		expected string
	}{
		{
			name:     "返回中间件名称",
			expected: "SecurityHeadersMiddleware",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			middleware := NewSecurityHeadersMiddleware().(*SecurityHeadersMiddleware)
			assert.Equal(t, tt.expected, middleware.MiddlewareName())
		})
	}
}

func TestSecurityHeadersMiddleware_Order(t *testing.T) {
	tests := []struct {
		name     string
		expected int
	}{
		{
			name:     "返回执行顺序",
			expected: 40,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			middleware := NewSecurityHeadersMiddleware().(*SecurityHeadersMiddleware)
			assert.Equal(t, tt.expected, middleware.Order())
		})
	}
}

func TestSecurityHeadersMiddleware_Wrapper(t *testing.T) {
	tests := []struct {
		name string
	}{
		{
			name: "设置所有安全头",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gin.SetMode(gin.TestMode)
			router := gin.New()

			middleware := NewSecurityHeadersMiddleware().(*SecurityHeadersMiddleware)
			router.Use(middleware.Wrapper())

			router.GET("/test", func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{"message": "ok"})
			})

			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusOK, w.Code)

			headers := w.Header()
			assert.Equal(t, "DENY", headers.Get("X-Frame-Options"))
			assert.Equal(t, "nosniff", headers.Get("X-Content-Type-Options"))
			assert.Equal(t, "1; mode=block", headers.Get("X-XSS-Protection"))
			assert.Equal(t, "strict-origin-when-cross-origin", headers.Get("Referrer-Policy"))
		})
	}
}

func TestSecurityHeadersMiddleware_Wrapper_AllMethods(t *testing.T) {
	tests := []struct {
		name   string
		method string
	}{
		{
			name:   "GET请求",
			method: http.MethodGet,
		},
		{
			name:   "POST请求",
			method: http.MethodPost,
		},
		{
			name:   "PUT请求",
			method: http.MethodPut,
		},
		{
			name:   "DELETE请求",
			method: http.MethodDelete,
		},
		{
			name:   "PATCH请求",
			method: http.MethodPatch,
		},
		{
			name:   "OPTIONS请求",
			method: http.MethodOptions,
		},
		{
			name:   "HEAD请求",
			method: http.MethodHead,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gin.SetMode(gin.TestMode)
			router := gin.New()

			middleware := NewSecurityHeadersMiddleware().(*SecurityHeadersMiddleware)
			router.Use(middleware.Wrapper())

			router.Any("/test", func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{"message": "ok"})
			})

			req := httptest.NewRequest(tt.method, "/test", nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			headers := w.Header()
			assert.Equal(t, "DENY", headers.Get("X-Frame-Options"))
			assert.Equal(t, "nosniff", headers.Get("X-Content-Type-Options"))
			assert.Equal(t, "1; mode=block", headers.Get("X-XSS-Protection"))
			assert.Equal(t, "strict-origin-when-cross-origin", headers.Get("Referrer-Policy"))
		})
	}
}

func TestSecurityHeadersMiddleware_XFrameOptions(t *testing.T) {
	tests := []struct {
		name          string
		expectedValue string
	}{
		{
			name:          "X-Frame-Options头设置为DENY",
			expectedValue: "DENY",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gin.SetMode(gin.TestMode)
			router := gin.New()

			middleware := NewSecurityHeadersMiddleware().(*SecurityHeadersMiddleware)
			router.Use(middleware.Wrapper())

			router.GET("/test", func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{"message": "ok"})
			})

			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedValue, w.Header().Get("X-Frame-Options"))
		})
	}
}

func TestSecurityHeadersMiddleware_XContentTypeOptions(t *testing.T) {
	tests := []struct {
		name          string
		expectedValue string
	}{
		{
			name:          "X-Content-Type-Options头设置为nosniff",
			expectedValue: "nosniff",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gin.SetMode(gin.TestMode)
			router := gin.New()

			middleware := NewSecurityHeadersMiddleware().(*SecurityHeadersMiddleware)
			router.Use(middleware.Wrapper())

			router.GET("/test", func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{"message": "ok"})
			})

			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedValue, w.Header().Get("X-Content-Type-Options"))
		})
	}
}

func TestSecurityHeadersMiddleware_XXSSProtection(t *testing.T) {
	tests := []struct {
		name          string
		expectedValue string
	}{
		{
			name:          "X-XSS-Protection头设置正确",
			expectedValue: "1; mode=block",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gin.SetMode(gin.TestMode)
			router := gin.New()

			middleware := NewSecurityHeadersMiddleware().(*SecurityHeadersMiddleware)
			router.Use(middleware.Wrapper())

			router.GET("/test", func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{"message": "ok"})
			})

			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedValue, w.Header().Get("X-XSS-Protection"))
		})
	}
}

func TestSecurityHeadersMiddleware_ReferrerPolicy(t *testing.T) {
	tests := []struct {
		name          string
		expectedValue string
	}{
		{
			name:          "Referrer-Policy头设置正确",
			expectedValue: "strict-origin-when-cross-origin",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gin.SetMode(gin.TestMode)
			router := gin.New()

			middleware := NewSecurityHeadersMiddleware().(*SecurityHeadersMiddleware)
			router.Use(middleware.Wrapper())

			router.GET("/test", func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{"message": "ok"})
			})

			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedValue, w.Header().Get("Referrer-Policy"))
		})
	}
}

func TestSecurityHeadersMiddleware_OnStart(t *testing.T) {
	tests := []struct {
		name string
	}{
		{
			name: "服务器启动回调",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			middleware := NewSecurityHeadersMiddleware().(*SecurityHeadersMiddleware)
			err := middleware.OnStart()
			assert.NoError(t, err)
		})
	}
}

func TestSecurityHeadersMiddleware_OnStop(t *testing.T) {
	tests := []struct {
		name string
	}{
		{
			name: "服务器停止回调",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			middleware := NewSecurityHeadersMiddleware().(*SecurityHeadersMiddleware)
			err := middleware.OnStop()
			assert.NoError(t, err)
		})
	}
}

func TestSecurityHeadersMiddleware_ChainedMiddlewares(t *testing.T) {
	tests := []struct {
		name string
	}{
		{
			name: "链式中间件组合",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gin.SetMode(gin.TestMode)
			router := gin.New()

			securityMiddleware := NewSecurityHeadersMiddleware().(*SecurityHeadersMiddleware)
			corsMiddleware := NewCorsMiddleware().(*CorsMiddleware)

			router.Use(securityMiddleware.Wrapper())
			router.Use(corsMiddleware.Wrapper())

			router.GET("/test", func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{"message": "ok"})
			})

			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			headers := w.Header()

			assert.Equal(t, http.StatusOK, w.Code)

			assert.Equal(t, "DENY", headers.Get("X-Frame-Options"))
			assert.Equal(t, "nosniff", headers.Get("X-Content-Type-Options"))
			assert.Equal(t, "1; mode=block", headers.Get("X-XSS-Protection"))
			assert.Equal(t, "strict-origin-when-cross-origin", headers.Get("Referrer-Policy"))
			assert.Equal(t, "*", headers.Get("Access-Control-Allow-Origin"))
			assert.Equal(t, "true", headers.Get("Access-Control-Allow-Credentials"))
		})
	}
}

func TestSecurityHeadersMiddleware_NotOverridingExistingHeaders(t *testing.T) {
	tests := []struct {
		name string
	}{
		{
			name: "不覆盖已存在的头",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gin.SetMode(gin.TestMode)
			router := gin.New()

			router.Use(func(c *gin.Context) {
				c.Writer.Header().Set("X-Frame-Options", "SAMEORIGIN")
				c.Next()
			})

			middleware := NewSecurityHeadersMiddleware().(*SecurityHeadersMiddleware)
			router.Use(middleware.Wrapper())

			router.GET("/test", func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{"message": "ok"})
			})

			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, "DENY", w.Header().Get("X-Frame-Options"))
		})
	}
}

func TestSecurityHeadersMiddleware_ErrorResponse(t *testing.T) {
	tests := []struct {
		name string
	}{
		{
			name: "错误响应也设置安全头",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gin.SetMode(gin.TestMode)
			router := gin.New()

			middleware := NewSecurityHeadersMiddleware().(*SecurityHeadersMiddleware)
			router.Use(middleware.Wrapper())

			router.GET("/test", func(c *gin.Context) {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "server error"})
			})

			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusInternalServerError, w.Code)

			headers := w.Header()
			assert.Equal(t, "DENY", headers.Get("X-Frame-Options"))
			assert.Equal(t, "nosniff", headers.Get("X-Content-Type-Options"))
			assert.Equal(t, "1; mode=block", headers.Get("X-XSS-Protection"))
			assert.Equal(t, "strict-origin-when-cross-origin", headers.Get("Referrer-Policy"))
		})
	}
}

func TestSecurityHeadersMiddleware_EmptyPath(t *testing.T) {
	tests := []struct {
		name string
	}{
		{
			name: "空路径请求",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gin.SetMode(gin.TestMode)
			router := gin.New()

			middleware := NewSecurityHeadersMiddleware().(*SecurityHeadersMiddleware)
			router.Use(middleware.Wrapper())

			router.GET("/", func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{"message": "ok"})
			})

			req := httptest.NewRequest(http.MethodGet, "/", nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			headers := w.Header()
			assert.Equal(t, "DENY", headers.Get("X-Frame-Options"))
			assert.Equal(t, "nosniff", headers.Get("X-Content-Type-Options"))
		})
	}
}

func TestSecurityHeadersMiddleware_QueryParameters(t *testing.T) {
	tests := []struct {
		name string
	}{
		{
			name: "带查询参数的请求",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gin.SetMode(gin.TestMode)
			router := gin.New()

			middleware := NewSecurityHeadersMiddleware().(*SecurityHeadersMiddleware)
			router.Use(middleware.Wrapper())

			router.GET("/test", func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{"message": "ok"})
			})

			req := httptest.NewRequest(http.MethodGet, "/test?param1=value1&param2=value2", nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			headers := w.Header()
			assert.Equal(t, "DENY", headers.Get("X-Frame-Options"))
			assert.Equal(t, "nosniff", headers.Get("X-Content-Type-Options"))
		})
	}
}
