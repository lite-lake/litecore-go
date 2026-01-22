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
	router string
	Logger ILogger `inject:""`
}

func (m *mockController) ControllerName() string {
	return m.name
}

func (m *mockController) GetRouter() string {
	return m.router
}

func (m *mockController) Handle(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, gin.H{
		"message": "success",
	})
}

type testController struct {
	Logger ILogger `inject:""`
}

func (t *testController) ControllerName() string {
	return "TestController"
}

func (t *testController) GetRouter() string {
	return "/test [GET]"
}

func (t *testController) Handle(ctx *gin.Context) {
	ctx.JSON(http.StatusCreated, gin.H{
		"controller": t.ControllerName(),
	})
}

func TestIBaseController_基础接口实现(t *testing.T) {
	controller := &mockController{
		name:   "TestController",
		router: "/test [GET]",
	}

	assert.Equal(t, "TestController", controller.ControllerName())
	assert.Equal(t, "/test [GET]", controller.GetRouter())
}

func TestIBaseController_Handle方法(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		controller IBaseController
		wantCode   int
	}{
		{
			name:       "返回成功状态",
			controller: &mockController{name: "Success", router: "/success [GET]"},
			wantCode:   http.StatusOK,
		},
		{
			name:       "返回创建状态",
			controller: &testController{},
			wantCode:   http.StatusCreated,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			tt.controller.Handle(c)

			assert.Equal(t, tt.wantCode, w.Code)
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
			controller: &mockController{name: "", router: "/test [GET]"},
		},
		{
			name:       "带有空路由的控制器",
			controller: &mockController{name: "Test", router: ""},
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
		name:   "CombinedController",
		router: "/combined [POST]",
	}

	var iface IBaseController = controller
	assert.Equal(t, "CombinedController", iface.ControllerName())
	assert.Equal(t, "/combined [POST]", iface.GetRouter())
}

func TestIBaseController_路由格式(t *testing.T) {
	tests := []struct {
		name       string
		controller IBaseController
		wantRouter string
	}{
		{
			name:       "GET请求路由",
			controller: &mockController{name: "GetTest", router: "/api/users [GET]"},
			wantRouter: "/api/users [GET]",
		},
		{
			name:       "POST请求路由",
			controller: &mockController{name: "PostTest", router: "/api/users [POST]"},
			wantRouter: "/api/users [POST]",
		},
		{
			name:       "PUT请求路由",
			controller: &mockController{name: "PutTest", router: "/api/users/:id [PUT]"},
			wantRouter: "/api/users/:id [PUT]",
		},
		{
			name:       "DELETE请求路由",
			controller: &mockController{name: "DeleteTest", router: "/api/users/:id [DELETE]"},
			wantRouter: "/api/users/:id [DELETE]",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.wantRouter, tt.controller.GetRouter())
		})
	}
}
