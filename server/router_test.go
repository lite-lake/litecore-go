package server

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
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
			name:           "有效路由_逗号分隔多方法",
			route:          "/oauth2/consent [GET,POST]",
			expectedMethod: "GET|POST",
			expectedPath:   "/oauth2/consent",
			expectError:    false,
		},
		{
			name:           "有效路由_逗号分隔多方法带空格",
			route:          "/oauth2/consent [GET, POST, PUT]",
			expectedMethod: "GET| POST| PUT",
			expectedPath:   "/oauth2/consent",
			expectError:    false,
		},
		{
			name:           "有效路由_ANY方法",
			route:          "/api/*any [ANY]",
			expectedMethod: "ANY",
			expectedPath:   "/api/*any",
			expectError:    false,
		},
		{
			name:           "有效路由_竖线分隔多方法（兼容原有写法）",
			route:          "/api/users [GET|POST|DELETE]",
			expectedMethod: "GET|POST|DELETE",
			expectedPath:   "/api/users",
			expectError:    false,
		},
		{
			name:           "有效路由_混合分隔符",
			route:          "/api/users [GET,POST|PUT]",
			expectedMethod: "GET|POST|PUT",
			expectedPath:   "/api/users",
			expectError:    false,
		},
		{
			name:           "无效格式_缺少方法",
			route:          "/api/users",
			expectedMethod: "",
			expectedPath:   "",
			expectError:    true,
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

// TestNoRouteHandler 测试 NoRoute 处理器
func TestNoRouteHandler(t *testing.T) {
	t.Run("未定义的路由_返回404", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
		ginEngine := gin.New()

		engine := &Engine{}
		engine.ginEngine = ginEngine
		engine.serverConfig = defaultServerConfig()

		req, _ := http.NewRequest("GET", "/not-found", nil)
		w := httptest.NewRecorder()
		ginEngine.ServeHTTP(w, req)

		if w.Code != http.StatusNotFound {
			t.Errorf("期望 404, 实际: %d", w.Code)
		}
	})
}

// TestMultiMethodRouteRegistration 测试多方法路由注册
func TestMultiMethodRouteRegistration(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine := &Engine{
		ginEngine: gin.New(),
	}

	// 测试逗号分隔的多方法注册
	testPath := "/test/multi"
	testHandler := func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	}

	// 注册GET,POST方法
	methods := "GET,POST"
	path := testPath
	for _, method := range strings.Split(methods, ",") {
		method = strings.TrimSpace(method)
		engine.registerRoute(method, path, testHandler)
	}

	// 测试GET请求
	reqGet, _ := http.NewRequest("GET", testPath, nil)
	wGet := httptest.NewRecorder()
	engine.ginEngine.ServeHTTP(wGet, reqGet)
	if wGet.Code != http.StatusOK {
		t.Errorf("GET请求期望200, 实际: %d", wGet.Code)
	}

	// 测试POST请求
	reqPost, _ := http.NewRequest("POST", testPath, nil)
	wPost := httptest.NewRecorder()
	engine.ginEngine.ServeHTTP(wPost, reqPost)
	if wPost.Code != http.StatusOK {
		t.Errorf("POST请求期望200, 实际: %d", wPost.Code)
	}

	// 测试PUT请求（未注册）
	reqPut, _ := http.NewRequest("PUT", testPath, nil)
	wPut := httptest.NewRecorder()
	engine.ginEngine.ServeHTTP(wPut, reqPut)
	if wPut.Code != http.StatusNotFound {
		t.Errorf("PUT请求期望404, 实际: %d", wPut.Code)
	}
}
