package common

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

type mockController struct {
	name   string
	route  string
	handle func(ctx *gin.Context)
}

func (m *mockController) ControllerName() string {
	return m.name
}

func (m *mockController) GetRouter() string {
	return m.route
}

func (m *mockController) Handle(ctx *gin.Context) {
	if m.handle != nil {
		m.handle(ctx)
	} else {
		ctx.JSON(http.StatusOK, gin.H{"message": "ok"})
	}
}

type userController struct{}

func (u *userController) ControllerName() string {
	return "UserController"
}

func (u *userController) GetRouter() string {
	return "/api/users [GET]"
}

func (u *userController) Handle(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, gin.H{"users": []string{"user1", "user2"}})
}

type errorController struct{}

func (e *errorController) ControllerName() string {
	return "ErrorController"
}

func (e *errorController) GetRouter() string {
	return "/api/error [GET]"
}

func (e *errorController) Handle(ctx *gin.Context) {
	ctx.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
}

func TestIBaseController_基础接口实现(t *testing.T) {
	controller := &mockController{
		name:  "TestController",
		route: "/test [GET]",
	}

	assert.Equal(t, "TestController", controller.ControllerName())
	assert.Equal(t, "/test [GET]", controller.GetRouter())
	assert.NotNil(t, controller.Handle)
}

func TestIBaseController_Handle方法(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		controller IBaseController
		wantCode   int
		wantBody   string
	}{
		{
			name:       "正常响应控制器",
			controller: &mockController{name: "Normal", route: "/normal [GET]"},
			wantCode:   http.StatusOK,
			wantBody:   `{"message":"ok"}`,
		},
		{
			name:       "用户控制器",
			controller: &userController{},
			wantCode:   http.StatusOK,
			wantBody:   `{"users":["user1","user2"]}`,
		},
		{
			name:       "错误控制器",
			controller: &errorController{},
			wantCode:   http.StatusInternalServerError,
			wantBody:   `{"error":"internal error"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			tt.controller.Handle(c)

			assert.Equal(t, tt.wantCode, w.Code)
			assert.JSONEq(t, tt.wantBody, w.Body.String())
		})
	}
}

func TestIBaseController_空实现(t *testing.T) {
	tests := []struct {
		name       string
		controller IBaseController
	}{
		{
			name:       "空控制器实例",
			controller: &mockController{},
		},
		{
			name:       "带有空名称的控制器",
			controller: &mockController{name: "", route: "/test [GET]"},
		},
		{
			name:       "带有空路由的控制器",
			controller: &mockController{name: "Test", route: ""},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.NotNil(t, tt.controller.ControllerName())
			assert.NotNil(t, tt.controller.GetRouter())
		})
	}
}

func TestIBaseController_接口组合(t *testing.T) {
	controller := &mockController{
		name:  "CombinedController",
		route: "/combined [POST]",
	}

	var iface IBaseController = controller
	assert.Equal(t, "CombinedController", iface.ControllerName())
	assert.Equal(t, "/combined [POST]", iface.GetRouter())
}

func TestIBaseController_路由格式(t *testing.T) {
	tests := []struct {
		name       string
		controller IBaseController
		expected   string
	}{
		{
			name:       "GET路由",
			controller: &mockController{name: "GetTest", route: "/api/test [GET]"},
			expected:   "/api/test [GET]",
		},
		{
			name:       "POST路由",
			controller: &mockController{name: "PostTest", route: "/api/test [POST]"},
			expected:   "/api/test [POST]",
		},
		{
			name:       "PUT路由",
			controller: &mockController{name: "PutTest", route: "/api/test [PUT]"},
			expected:   "/api/test [PUT]",
		},
		{
			name:       "DELETE路由",
			controller: &mockController{name: "DeleteTest", route: "/api/test [DELETE]"},
			expected:   "/api/test [DELETE]",
		},
		{
			name:       "复杂路由",
			controller: &mockController{name: "ComplexTest", route: "/api/users/:id/profile [GET]"},
			expected:   "/api/users/:id/profile [GET]",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.controller.GetRouter())
		})
	}
}

func TestIBaseController_上下文处理(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("从上下文中获取参数", func(t *testing.T) {
		called := false
		controller := &mockController{
			name:  "ContextTest",
			route: "/test [GET]",
			handle: func(ctx *gin.Context) {
				param := ctx.Query("param")
				assert.Equal(t, "value", param)
				called = true
			},
		}

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("GET", "/test?param=value", nil)

		controller.Handle(c)
		assert.True(t, called)
	})

	t.Run("设置响应头", func(t *testing.T) {
		controller := &mockController{
			name:  "HeaderTest",
			route: "/test [GET]",
			handle: func(ctx *gin.Context) {
				ctx.Header("X-Custom-Header", "custom-value")
				ctx.JSON(http.StatusOK, gin.H{})
			},
		}

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		controller.Handle(c)
		assert.Equal(t, "custom-value", w.Header().Get("X-Custom-Header"))
	})
}
