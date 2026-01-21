package server

import (
	"net/http"
	"net/http/httptest"
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
