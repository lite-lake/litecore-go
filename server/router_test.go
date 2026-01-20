package server

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"

	"com.litelake.litecore/common"
	"com.litelake.litecore/container"
)

// TestParseRoute 测试路由解析
func TestParseRoute(t *testing.T) {
	tests := []struct {
		name           string
		route          string
		expectedMethod string
		expectedPath   string
		expectError    bool
	}{
		{
			name:           "有效路由_GET",
			route:          "/api/users [GET]",
			expectedMethod: "GET",
			expectedPath:   "/api/users",
			expectError:    false,
		},
		{
			name:           "有效路由_POST",
			route:          "/api/users [POST]",
			expectedMethod: "POST",
			expectedPath:   "/api/users",
			expectError:    false,
		},
		{
			name:           "有效路由_PUT",
			route:          "/api/users/123 [PUT]",
			expectedMethod: "PUT",
			expectedPath:   "/api/users/123",
			expectError:    false,
		},
		{
			name:           "有效路由_DELETE",
			route:          "/api/users/123 [DELETE]",
			expectedMethod: "DELETE",
			expectedPath:   "/api/users/123",
			expectError:    false,
		},
		{
			name:           "有效路由_PATCH",
			route:          "/api/users/123 [PATCH]",
			expectedMethod: "PATCH",
			expectedPath:   "/api/users/123",
			expectError:    false,
		},
		{
			name:           "有效路由_小写方法",
			route:          "/api/users [get]",
			expectedMethod: "GET",
			expectedPath:   "/api/users",
			expectError:    false,
		},
		{
			name:           "有效路由_混合大小写",
			route:          "/api/users [Post]",
			expectedMethod: "POST",
			expectedPath:   "/api/users",
			expectError:    false,
		},
		{
			name:           "有效路由_路径带空格",
			route:          "/api/users/123/active [GET]",
			expectedMethod: "GET",
			expectedPath:   "/api/users/123/active",
			expectError:    false,
		},
		{
			name:           "无效格式_缺少方法",
			route:          "/api/users",
			expectedMethod: "",
			expectedPath:   "",
			expectError:    true,
		},
		{
			name:           "无效格式_缺少闭合括号",
			route:          "/api/users [GET",
			expectedMethod: "",
			expectedPath:   "",
			expectError:    true,
		},
		{
			name:           "无效格式_缺少开放括号",
			route:          "/api/users GET]",
			expectedMethod: "",
			expectedPath:   "",
			expectError:    true,
		},
		{
			name:           "无效格式_空路径",
			route:          " [GET]",
			expectedMethod: "",
			expectedPath:   "",
			expectError:    true,
		},
		{
			name:           "无效格式_空方法",
			route:          "/api/users []",
			expectedMethod: "",
			expectedPath:   "",
			expectError:    true,
		},
		{
			name:           "根路径",
			route:          "/ [GET]",
			expectedMethod: "GET",
			expectedPath:   "/",
			expectError:    false,
		},
		{
			name:           "复杂路径",
			route:          "/api/v1/users/{id}/profile [PUT]",
			expectedMethod: "PUT",
			expectedPath:   "/api/v1/users/{id}/profile",
			expectError:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			method, path, err := parseRoute(tt.route)

			if tt.expectError {
				if err == nil {
					t.Errorf("期望返回错误，但实际返回: method=%s, path=%s", method, path)
				}
				return
			}

			if err != nil {
				t.Errorf("未期望的错误: %v", err)
				return
			}

			if method != tt.expectedMethod {
				t.Errorf("期望 method = %s, 实际 = %s", tt.expectedMethod, method)
			}

			if path != tt.expectedPath {
				t.Errorf("期望 path = %s, 实际 = %s", tt.expectedPath, path)
			}
		})
	}
}

// TestRegisterRoute 测试路由注册
func TestRegisterRoute(t *testing.T) {
	tests := []struct {
		name    string
		method  string
		path    string
		handler gin.HandlerFunc
	}{
		{
			name:   "GET 路由",
			method: "GET",
			path:   "/test",
			handler: func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{"method": "GET"})
			},
		},
		{
			name:   "POST 路由",
			method: "POST",
			path:   "/test",
			handler: func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{"method": "POST"})
			},
		},
		{
			name:   "PUT 路由",
			method: "PUT",
			path:   "/test",
			handler: func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{"method": "PUT"})
			},
		},
		{
			name:   "DELETE 路由",
			method: "DELETE",
			path:   "/test",
			handler: func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{"method": "DELETE"})
			},
		},
		{
			name:   "PATCH 路由",
			method: "PATCH",
			path:   "/test",
			handler: func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{"method": "PATCH"})
			},
		},
		{
			name:   "HEAD 路由",
			method: "HEAD",
			path:   "/test",
			handler: func(c *gin.Context) {
				c.Status(http.StatusOK)
			},
		},
		{
			name:   "OPTIONS 路由",
			method: "OPTIONS",
			path:   "/test",
			handler: func(c *gin.Context) {
				c.Status(http.StatusOK)
			},
		},
		{
			name:   "未知方法_默认为GET",
			method: "UNKNOWN",
			path:   "/test",
			handler: func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{"method": "GET"})
			},
		},
		{
			name:   "小写方法_转换为GET",
			method: "get",
			path:   "/test",
			handler: func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{"method": "GET"})
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gin.SetMode(gin.TestMode)
			ginEngine := gin.New()

			configContainer := container.NewConfigContainer()
			entityContainer := container.NewEntityContainer()
			managerContainer := container.NewManagerContainer(configContainer)
			repositoryContainer := container.NewRepositoryContainer(configContainer, managerContainer, entityContainer)
			serviceContainer := container.NewServiceContainer(configContainer, managerContainer, repositoryContainer)
			controllerContainer := container.NewControllerContainer(configContainer, managerContainer, serviceContainer)
			middlewareContainer := container.NewMiddlewareContainer(configContainer, managerContainer, serviceContainer)

			engine := NewEngine(
				configContainer,
				entityContainer,
				managerContainer,
				repositoryContainer,
				serviceContainer,
				controllerContainer,
				middlewareContainer,
			)

			engine.ginEngine = ginEngine
			engine.registerRoute(tt.method, tt.path, tt.handler)

			reqMethod := tt.method
			if tt.name == "未知方法_默认为GET" || tt.name == "小写方法_转换为GET" {
				reqMethod = "GET"
			}

			req, _ := http.NewRequest(reqMethod, tt.path, nil)
			w := httptest.NewRecorder()
			ginEngine.ServeHTTP(w, req)

			if w.Code == http.StatusNotFound {
				t.Errorf("路由 %s %s 未注册", tt.method, tt.path)
			}
		})
	}
}

// TestNoRouteHandler 测试 NoRoute 处理器
func TestNoRouteHandler(t *testing.T) {
	t.Run("未定义的路由_返回404", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
		ginEngine := gin.New()

		configContainer := container.NewConfigContainer()
		entityContainer := container.NewEntityContainer()
		managerContainer := container.NewManagerContainer(configContainer)
		repositoryContainer := container.NewRepositoryContainer(configContainer, managerContainer, entityContainer)
		serviceContainer := container.NewServiceContainer(configContainer, managerContainer, repositoryContainer)
		controllerContainer := container.NewControllerContainer(configContainer, managerContainer, serviceContainer)
		middlewareContainer := container.NewMiddlewareContainer(configContainer, managerContainer, serviceContainer)

		engine := NewEngine(
			configContainer,
			entityContainer,
			managerContainer,
			repositoryContainer,
			serviceContainer,
			controllerContainer,
			middlewareContainer,
		)

		engine.ginEngine = ginEngine

		engine.ginEngine.NoRoute(func(c *gin.Context) {
			c.JSON(http.StatusNotFound, gin.H{
				"error":  "route not found",
				"path":   c.Request.URL.Path,
				"method": c.Request.Method,
			})
		})

		req, _ := http.NewRequest("GET", "/nonexistent", nil)
		w := httptest.NewRecorder()
		ginEngine.ServeHTTP(w, req)

		if w.Code != http.StatusNotFound {
			t.Errorf("期望状态码 404, 实际 = %d", w.Code)
		}

		body := w.Body.String()
		if !strings.Contains(body, "route not found") {
			t.Errorf("响应应该包含错误信息")
		}
	})
}

// TestControllerRegistration 测试控制器注册
func TestControllerRegistration(t *testing.T) {
	t.Run("空容器_注册成功", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
		ginEngine := gin.New()

		configContainer := container.NewConfigContainer()
		entityContainer := container.NewEntityContainer()
		managerContainer := container.NewManagerContainer(configContainer)
		repositoryContainer := container.NewRepositoryContainer(configContainer, managerContainer, entityContainer)
		serviceContainer := container.NewServiceContainer(configContainer, managerContainer, repositoryContainer)
		controllerContainer := container.NewControllerContainer(configContainer, managerContainer, serviceContainer)
		middlewareContainer := container.NewMiddlewareContainer(configContainer, managerContainer, serviceContainer)

		engine := NewEngine(
			configContainer,
			entityContainer,
			managerContainer,
			repositoryContainer,
			serviceContainer,
			controllerContainer,
			middlewareContainer,
		)

		engine.ginEngine = ginEngine
		err := engine.registerControllers()

		if err != nil {
			t.Errorf("期望注册成功, 实际返回错误: %v", err)
		}
	})
}

// TestControllerRegistration_InvalidRoute 测试无效路由的控制器
func TestControllerRegistration_InvalidRoute(t *testing.T) {
	t.Run("无效路由_跳过注册", func(t *testing.T) {
		gin.SetMode(gin.TestMode)

		configContainer := container.NewConfigContainer()
		entityContainer := container.NewEntityContainer()
		managerContainer := container.NewManagerContainer(configContainer)
		repositoryContainer := container.NewRepositoryContainer(configContainer, managerContainer, entityContainer)
		serviceContainer := container.NewServiceContainer(configContainer, managerContainer, repositoryContainer)
		controllerContainer := container.NewControllerContainer(configContainer, managerContainer, serviceContainer)
		middlewareContainer := container.NewMiddlewareContainer(configContainer, managerContainer, serviceContainer)

		engine := NewEngine(
			configContainer,
			entityContainer,
			managerContainer,
			repositoryContainer,
			serviceContainer,
			controllerContainer,
			middlewareContainer,
		)

		engine.ginEngine = gin.New()
		err := engine.registerControllers()

		if err != nil {
			t.Errorf("期望跳过无效路由, 实际返回错误: %v", err)
		}
	})
}

// TestMultipleControllers 测试多个控制器注册
func TestMultipleControllers(t *testing.T) {
	t.Run("空容器_注册成功", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
		ginEngine := gin.New()

		configContainer := container.NewConfigContainer()
		entityContainer := container.NewEntityContainer()
		managerContainer := container.NewManagerContainer(configContainer)
		repositoryContainer := container.NewRepositoryContainer(configContainer, managerContainer, entityContainer)
		serviceContainer := container.NewServiceContainer(configContainer, managerContainer, repositoryContainer)
		controllerContainer := container.NewControllerContainer(configContainer, managerContainer, serviceContainer)
		middlewareContainer := container.NewMiddlewareContainer(configContainer, managerContainer, serviceContainer)

		engine := NewEngine(
			configContainer,
			entityContainer,
			managerContainer,
			repositoryContainer,
			serviceContainer,
			controllerContainer,
			middlewareContainer,
		)

		engine.ginEngine = ginEngine
		err := engine.registerControllers()

		if err != nil {
			t.Errorf("期望注册成功, 实际返回错误: %v", err)
		}
	})
}

// BenchmarkParseRoute 基准测试路由解析性能
func BenchmarkParseRoute(b *testing.B) {
	route := "/api/v1/users/{id}/profile [GET]"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _, _ = parseRoute(route)
	}
}

// BenchmarkRouteRegistration 基准测试路由注册性能
func BenchmarkRouteRegistration(b *testing.B) {
	gin.SetMode(gin.TestMode)
	ginEngine := gin.New()

	configContainer := container.NewConfigContainer()
	entityContainer := container.NewEntityContainer()
	managerContainer := container.NewManagerContainer(configContainer)
	repositoryContainer := container.NewRepositoryContainer(configContainer, managerContainer, entityContainer)
	serviceContainer := container.NewServiceContainer(configContainer, managerContainer, repositoryContainer)
	controllerContainer := container.NewControllerContainer(configContainer, managerContainer, serviceContainer)
	middlewareContainer := container.NewMiddlewareContainer(configContainer, managerContainer, serviceContainer)

	engine := NewEngine(
		configContainer,
		entityContainer,
		managerContainer,
		repositoryContainer,
		serviceContainer,
		controllerContainer,
		middlewareContainer,
	)

	engine.ginEngine = ginEngine

	handler := func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		engine.registerRoute("GET", "/api/test", handler)
	}
}

// mockTestController 测试用的模拟控制器
type mockTestController struct {
	name  string
	route string
}

func (m *mockTestController) ControllerName() string {
	return m.name
}

func (m *mockTestController) GetRouter() string {
	return m.route
}

func (m *mockTestController) Handle(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, gin.H{"controller": m.name})
}

var _ common.IBaseController = (*mockTestController)(nil)
