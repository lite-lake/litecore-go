package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"

	"com.litelake.litecore/common"
)

type mockTelemetryManager struct{}

func (m *mockTelemetryManager) ManagerName() string {
	return "mockTelemetryManager"
}

func (m *mockTelemetryManager) Health() error {
	return nil
}

func (m *mockTelemetryManager) OnStart() error {
	return nil
}

func (m *mockTelemetryManager) OnStop() error {
	return nil
}

type mockOtelManager struct {
	*mockTelemetryManager
}

func (m *mockOtelManager) GinMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set("otel_traced", true)
		c.Next()
	}
}

func TestNewTelemetryMiddleware(t *testing.T) {
	tests := []struct {
		name    string
		manager common.IBaseManager
	}{
		{
			name:    "创建遥测中间件",
			manager: &mockTelemetryManager{},
		},
		{
			name:    "创建遥测中间件无管理器",
			manager: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			middleware := NewTelemetryMiddleware(tt.manager)
			assert.NotNil(t, middleware)
			assert.IsType(t, &TelemetryMiddleware{}, middleware)
		})
	}
}

func TestTelemetryMiddleware_MiddlewareName(t *testing.T) {
	tests := []struct {
		name     string
		expected string
	}{
		{
			name:     "返回中间件名称",
			expected: "TelemetryMiddleware",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			middleware := NewTelemetryMiddleware(&mockTelemetryManager{}).(*TelemetryMiddleware)
			assert.Equal(t, tt.expected, middleware.MiddlewareName())
		})
	}
}

func TestTelemetryMiddleware_Order(t *testing.T) {
	tests := []struct {
		name     string
		expected int
	}{
		{
			name:     "返回执行顺序",
			expected: 50,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			middleware := NewTelemetryMiddleware(&mockTelemetryManager{}).(*TelemetryMiddleware)
			assert.Equal(t, tt.expected, middleware.Order())
		})
	}
}

func TestTelemetryMiddleware_Wrapper(t *testing.T) {
	tests := []struct {
		name              string
		manager           common.IBaseManager
		hasOtelMiddleware bool
	}{
		{
			name:              "使用OTel中间件",
			manager:           &mockOtelManager{},
			hasOtelMiddleware: true,
		},
		{
			name:              "不使用OTel中间件",
			manager:           &mockTelemetryManager{},
			hasOtelMiddleware: false,
		},
		{
			name:              "无管理器",
			manager:           nil,
			hasOtelMiddleware: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gin.SetMode(gin.TestMode)
			router := gin.New()

			middleware := NewTelemetryMiddleware(tt.manager).(*TelemetryMiddleware)
			router.Use(middleware.Wrapper())

			router.GET("/test", func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{"message": "ok"})
			})

			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusOK, w.Code)

			if tt.hasOtelMiddleware {
				assert.Contains(t, w.Body.String(), "message")
			}
		})
	}
}

func TestTelemetryMiddleware_Wrapper_Context(t *testing.T) {
	tests := []struct {
		name              string
		manager           common.IBaseManager
		hasOtelMiddleware bool
	}{
		{
			name:              "OTel中间件设置上下文",
			manager:           &mockOtelManager{},
			hasOtelMiddleware: true,
		},
		{
			name:              "无OTel中间件上下文",
			manager:           &mockTelemetryManager{},
			hasOtelMiddleware: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gin.SetMode(gin.TestMode)
			router := gin.New()

			middleware := NewTelemetryMiddleware(tt.manager).(*TelemetryMiddleware)
			router.Use(middleware.Wrapper())

			var traced bool
			router.GET("/test", func(c *gin.Context) {
				traced = c.GetBool("otel_traced")
				c.JSON(http.StatusOK, gin.H{"traced": traced})
			})

			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusOK, w.Code)

			if tt.hasOtelMiddleware {
				assert.True(t, traced)
			} else {
				assert.False(t, traced)
			}
		})
	}
}

func TestTelemetryMiddleware_Wrapper_Chain(t *testing.T) {
	tests := []struct {
		name              string
		manager           common.IBaseManager
		hasOtelMiddleware bool
	}{
		{
			name:              "中间件链式调用",
			manager:           &mockOtelManager{},
			hasOtelMiddleware: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gin.SetMode(gin.TestMode)
			router := gin.New()

			middleware := NewTelemetryMiddleware(tt.manager).(*TelemetryMiddleware)

			var executionOrder []string
			router.Use(func(c *gin.Context) {
				executionOrder = append(executionOrder, "before")
				c.Next()
				executionOrder = append(executionOrder, "after")
			})
			router.Use(middleware.Wrapper())
			router.Use(func(c *gin.Context) {
				executionOrder = append(executionOrder, "wrapper")
				c.Next()
			})

			router.GET("/test", func(c *gin.Context) {
				executionOrder = append(executionOrder, "handler")
				c.JSON(http.StatusOK, gin.H{"message": "ok"})
			})

			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusOK, w.Code)
			assert.Len(t, executionOrder, 4)
			assert.Equal(t, "before", executionOrder[0])
			assert.Equal(t, "wrapper", executionOrder[1])
			assert.Equal(t, "handler", executionOrder[2])
			assert.Equal(t, "after", executionOrder[3])
		})
	}
}

func TestTelemetryMiddleware_OnStart(t *testing.T) {
	tests := []struct {
		name string
	}{
		{
			name: "服务器启动回调",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			middleware := NewTelemetryMiddleware(&mockTelemetryManager{}).(*TelemetryMiddleware)
			err := middleware.OnStart()
			assert.NoError(t, err)
		})
	}
}

func TestTelemetryMiddleware_OnStop(t *testing.T) {
	tests := []struct {
		name string
	}{
		{
			name: "服务器停止回调",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			middleware := NewTelemetryMiddleware(&mockTelemetryManager{}).(*TelemetryMiddleware)
			err := middleware.OnStop()
			assert.NoError(t, err)
		})
	}
}

func TestTelemetryMiddleware_NilManager(t *testing.T) {
	tests := []struct {
		name string
	}{
		{
			name: "空管理器处理",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gin.SetMode(gin.TestMode)
			router := gin.New()

			middleware := NewTelemetryMiddleware(nil).(*TelemetryMiddleware)
			router.Use(middleware.Wrapper())

			router.GET("/test", func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{"message": "ok"})
			})

			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusOK, w.Code)
			assert.Contains(t, w.Body.String(), "message")
		})
	}
}

type contextMockManager struct {
	*mockTelemetryManager
}

func (m *contextMockManager) GinMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		req := c.Request
		ctx := context.WithValue(req.Context(), "telemetry", "active")
		c.Request = req.WithContext(ctx)
		c.Next()
	}
}

func TestTelemetryMiddleware_ContextPropagation(t *testing.T) {
	tests := []struct {
		name            string
		manager         common.IBaseManager
		hasContextValue bool
	}{
		{
			name:            "传播上下文值",
			manager:         &contextMockManager{},
			hasContextValue: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gin.SetMode(gin.TestMode)
			router := gin.New()

			middleware := NewTelemetryMiddleware(tt.manager).(*TelemetryMiddleware)
			router.Use(middleware.Wrapper())

			var contextValue string
			router.GET("/test", func(c *gin.Context) {
				if val := c.Request.Context().Value("telemetry"); val != nil {
					contextValue = val.(string)
				}
				c.JSON(http.StatusOK, gin.H{"value": contextValue})
			})

			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusOK, w.Code)

			if tt.hasContextValue {
				assert.Equal(t, "active", contextValue)
			}
		})
	}
}
