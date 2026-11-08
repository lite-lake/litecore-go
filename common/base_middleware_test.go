package common

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

type mockMiddleware struct {
	name  string
	order int
}

func (m *mockMiddleware) MiddlewareName() string {
	return m.name
}

func (m *mockMiddleware) Order() int {
	return m.order
}

func (m *mockMiddleware) Wrapper() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		ctx.Next()
	}
}

func (m *mockMiddleware) OnStart() error {
	return nil
}

func (m *mockMiddleware) OnStop() error {
	return nil
}

type authMiddleware struct{}

func (a *authMiddleware) MiddlewareName() string {
	return "AuthMiddleware"
}

func (a *authMiddleware) Order() int {
	return 1
}

func (a *authMiddleware) Wrapper() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		ctx.Set("authenticated", true)
		ctx.Next()
	}
}

func (a *authMiddleware) OnStart() error {
	return nil
}

func (a *authMiddleware) OnStop() error {
	return nil
}

type failingMiddleware struct{}

func (f *failingMiddleware) MiddlewareName() string {
	return "FailingMiddleware"
}

func (f *failingMiddleware) Order() int {
	return 0
}

func (f *failingMiddleware) Wrapper() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		ctx.AbortWithStatus(http.StatusInternalServerError)
	}
}

func (f *failingMiddleware) OnStart() error {
	return errors.New("中间件启动失败")
}

func (f *failingMiddleware) OnStop() error {
	return errors.New("中间件停止失败")
}

func TestIBaseMiddleware_基础接口实现(t *testing.T) {
	middleware := &mockMiddleware{
		name:  "TestMiddleware",
		order: 10,
	}

	assert.Equal(t, "TestMiddleware", middleware.MiddlewareName())
	assert.Equal(t, 10, middleware.Order())
	assert.NotNil(t, middleware.Wrapper())
	assert.NoError(t, middleware.OnStart())
	assert.NoError(t, middleware.OnStop())
}

func TestIBaseMiddleware_Wrapper方法(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		middleware IBaseMiddleware
		wantCode   int
		wantAuth   bool
	}{
		{
			name:       "正常执行中间件",
			middleware: &mockMiddleware{name: "Normal", order: 1},
			wantCode:   http.StatusOK,
			wantAuth:   false,
		},
		{
			name:       "认证中间件",
			middleware: &authMiddleware{},
			wantCode:   http.StatusOK,
			wantAuth:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			tt.middleware.Wrapper()(c)
			c.JSON(http.StatusOK, gin.H{"test": "data"})

			assert.Equal(t, tt.wantCode, w.Code)
			if tt.wantAuth {
				auth, exists := c.Get("authenticated")
				assert.True(t, exists)
				assert.True(t, auth.(bool))
			}
		})
	}
}

func TestIBaseMiddleware_生命周期方法(t *testing.T) {
	tests := []struct {
		name       string
		middleware IBaseMiddleware
		wantErr    bool
	}{
		{
			name:       "正常启动和停止",
			middleware: &mockMiddleware{name: "LifecycleTest", order: 5},
			wantErr:    false,
		},
		{
			name:       "启动失败的中间件",
			middleware: &failingMiddleware{},
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.middleware.OnStart()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			err = tt.middleware.OnStop()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestIBaseMiddleware_空实现(t *testing.T) {
	tests := []struct {
		name       string
		middleware IBaseMiddleware
	}{
		{
			name:       "空中间件实例",
			middleware: &mockMiddleware{},
		},
		{
			name:       "带有空名称的中间件",
			middleware: &mockMiddleware{name: "", order: 10},
		},
		{
			name:       "零顺序的中间件",
			middleware: &mockMiddleware{name: "ZeroOrder", order: 0},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.NotNil(t, tt.middleware.MiddlewareName())
			assert.NotNil(t, tt.middleware.Wrapper())
		})
	}
}

func TestIBaseMiddleware_接口组合(t *testing.T) {
	middleware := &mockMiddleware{
		name:  "CombinedMiddleware",
		order: 100,
	}

	var iface IBaseMiddleware = middleware
	assert.Equal(t, "CombinedMiddleware", iface.MiddlewareName())
	assert.Equal(t, 100, iface.Order())
	assert.NotNil(t, iface.Wrapper())
}

func TestIBaseMiddleware_执行顺序(t *testing.T) {
	tests := []struct {
		name       string
		middleware IBaseMiddleware
		wantOrder  int
	}{
		{
			name:       "高优先级中间件",
			middleware: &mockMiddleware{name: "High", order: 1},
			wantOrder:  1,
		},
		{
			name:       "中等优先级中间件",
			middleware: &mockMiddleware{name: "Medium", order: 50},
			wantOrder:  50,
		},
		{
			name:       "低优先级中间件",
			middleware: &mockMiddleware{name: "Low", order: 100},
			wantOrder:  100,
		},
		{
			name:       "负数优先级中间件",
			middleware: &mockMiddleware{name: "Negative", order: -1},
			wantOrder:  -1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.wantOrder, tt.middleware.Order())
		})
	}
}
