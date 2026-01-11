package server

import (
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestParseRouter(t *testing.T) {
	tests := []struct {
		name          string
		router        string
		expectedRoute string
		expectedMethod string
	}{
		{
			name:          "simple route",
			router:        "/api/users",
			expectedRoute: "/api/users",
			expectedMethod: "GET",
		},
		{
			name:          "route with uppercase method in brackets",
			router:        "/api/users [POST]",
			expectedRoute: "/api/users",
			expectedMethod: "POST",
		},
		{
			name:          "route with lowercase method in brackets",
			router:        "/api/users [post]",
			expectedRoute: "/api/users",
			expectedMethod: "POST",
		},
		{
			name:          "route with space and method",
			router:        "/api/users DELETE",
			expectedRoute: "/api/users",
			expectedMethod: "DELETE",
		},
		{
			name:          "route with space and uppercase method",
			router:        "/api/users PUT",
			expectedRoute: "/api/users",
			expectedMethod: "PUT",
		},
		{
			name:          "PATCH method",
			router:        "/api/users [PATCH]",
			expectedRoute: "/api/users",
			expectedMethod: "PATCH",
		},
		{
			name:          "HEAD method",
			router:        "/api/users [HEAD]",
			expectedRoute: "/api/users",
			expectedMethod: "HEAD",
		},
		{
			name:          "OPTIONS method",
			router:        "/api/users [OPTIONS]",
			expectedRoute: "/api/users",
			expectedMethod: "OPTIONS",
		},
		{
			name:          "route with leading/trailing spaces",
			router:        "  /api/users  ",
			expectedRoute: "/api/users",
			expectedMethod: "GET",
		},
		{
			name:          "route with path params",
			router:        "/api/users/:id",
			expectedRoute: "/api/users/:id",
			expectedMethod: "GET",
		},
		{
			name:          "route with wildcard",
			router:        "/api/files/*filepath",
			expectedRoute: "/api/files/*filepath",
			expectedMethod: "GET",
		},
		{
			name:          "nested route",
			router:        "/api/v1/users/:id/posts",
			expectedRoute: "/api/v1/users/:id/posts",
			expectedMethod: "GET",
		},
		{
			name:          "invalid method defaults to GET",
			router:        "/api/users INVALID",
			expectedRoute: "/api/users",
			expectedMethod: "GET",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			route, method := parseRouter(tt.router)

			if route != tt.expectedRoute {
				t.Errorf("route = %s, want %s", route, tt.expectedRoute)
			}

			if method != tt.expectedMethod {
				t.Errorf("method = %s, want %s", method, tt.expectedMethod)
			}
		})
	}
}

func TestRegisterControllers(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("register GET controller", func(t *testing.T) {
		ctrl := &mockController{
			name:   "testController",
			router: "/api/test",
			handle: func(c *gin.Context) {
				c.JSON(200, gin.H{"method": "GET"})
			},
		}

		engine, err := NewEngine(RegisterControllers(ctrl))
		if err != nil {
			t.Fatalf("NewEngine() error = %v", err)
		}

		if err := engine.Initialize(); err != nil {
			t.Fatalf("Initialize() error = %v", err)
		}

		req := httptest.NewRequest("GET", "/api/test", nil)
		w := httptest.NewRecorder()
		engine.ginEngine.ServeHTTP(w, req)

		if w.Code != 200 {
			t.Errorf("status code = %d, want 200", w.Code)
		}
	})

	t.Run("register POST controller", func(t *testing.T) {
		ctrl := &mockController{
			name:   "testController",
			router: "/api/test [POST]",
			handle: func(c *gin.Context) {
				c.JSON(201, gin.H{"method": "POST"})
			},
		}

		engine, err := NewEngine(RegisterControllers(ctrl))
		if err != nil {
			t.Fatalf("NewEngine() error = %v", err)
		}

		if err := engine.Initialize(); err != nil {
			t.Fatalf("Initialize() error = %v", err)
		}

		req := httptest.NewRequest("POST", "/api/test", nil)
		w := httptest.NewRecorder()
		engine.ginEngine.ServeHTTP(w, req)

		if w.Code != 201 {
			t.Errorf("status code = %d, want 201", w.Code)
		}
	})

	t.Run("register PUT controller", func(t *testing.T) {
		ctrl := &mockController{
			name:   "testController",
			router: "/api/test [PUT]",
			handle: func(c *gin.Context) {
				c.JSON(200, gin.H{"method": "PUT"})
			},
		}

		engine, err := NewEngine(RegisterControllers(ctrl))
		if err != nil {
			t.Fatalf("NewEngine() error = %v", err)
		}

		if err := engine.Initialize(); err != nil {
			t.Fatalf("Initialize() error = %v", err)
		}

		req := httptest.NewRequest("PUT", "/api/test", nil)
		w := httptest.NewRecorder()
		engine.ginEngine.ServeHTTP(w, req)

		if w.Code != 200 {
			t.Errorf("status code = %d, want 200", w.Code)
		}
	})

	t.Run("register DELETE controller", func(t *testing.T) {
		ctrl := &mockController{
			name:   "testController",
			router: "/api/test [DELETE]",
			handle: func(c *gin.Context) {
				c.JSON(204, nil)
			},
		}

		engine, err := NewEngine(RegisterControllers(ctrl))
		if err != nil {
			t.Fatalf("NewEngine() error = %v", err)
		}

		if err := engine.Initialize(); err != nil {
			t.Fatalf("Initialize() error = %v", err)
		}

		req := httptest.NewRequest("DELETE", "/api/test", nil)
		w := httptest.NewRecorder()
		engine.ginEngine.ServeHTTP(w, req)

		if w.Code != 204 {
			t.Errorf("status code = %d, want 204", w.Code)
		}
	})

	t.Run("register PATCH controller", func(t *testing.T) {
		ctrl := &mockController{
			name:   "testController",
			router: "/api/test [PATCH]",
			handle: func(c *gin.Context) {
				c.JSON(200, gin.H{"method": "PATCH"})
			},
		}

		engine, err := NewEngine(RegisterControllers(ctrl))
		if err != nil {
			t.Fatalf("NewEngine() error = %v", err)
		}

		if err := engine.Initialize(); err != nil {
			t.Fatalf("Initialize() error = %v", err)
		}

		req := httptest.NewRequest("PATCH", "/api/test", nil)
		w := httptest.NewRecorder()
		engine.ginEngine.ServeHTTP(w, req)

		if w.Code != 200 {
			t.Errorf("status code = %d, want 200", w.Code)
		}
	})
}

func TestEngine_RegisterRoute(t *testing.T) {
	gin.SetMode(gin.TestMode)

	engine, err := NewEngine()
	if err != nil {
		t.Fatalf("NewEngine() error = %v", err)
	}

	if err := engine.Initialize(); err != nil {
		t.Fatalf("Initialize() error = %v", err)
	}

	t.Run("register GET route", func(t *testing.T) {
		handler := func(c *gin.Context) {
			c.JSON(200, gin.H{"message": "GET"})
		}

		engine.RegisterRoute("GET", "/custom/get", handler)

		req := httptest.NewRequest("GET", "/custom/get", nil)
		w := httptest.NewRecorder()
		engine.ginEngine.ServeHTTP(w, req)

		if w.Code != 200 {
			t.Errorf("status code = %d, want 200", w.Code)
		}
	})

	t.Run("register POST route", func(t *testing.T) {
		handler := func(c *gin.Context) {
			c.JSON(201, gin.H{"message": "POST"})
		}

		engine.RegisterRoute("POST", "/custom/post", handler)

		req := httptest.NewRequest("POST", "/custom/post", nil)
		w := httptest.NewRecorder()
		engine.ginEngine.ServeHTTP(w, req)

		if w.Code != 201 {
			t.Errorf("status code = %d, want 201", w.Code)
		}
	})

	t.Run("register PUT route", func(t *testing.T) {
		handler := func(c *gin.Context) {
			c.JSON(200, gin.H{"message": "PUT"})
		}

		engine.RegisterRoute("PUT", "/custom/put", handler)

		req := httptest.NewRequest("PUT", "/custom/put", nil)
		w := httptest.NewRecorder()
		engine.ginEngine.ServeHTTP(w, req)

		if w.Code != 200 {
			t.Errorf("status code = %d, want 200", w.Code)
		}
	})

	t.Run("register DELETE route", func(t *testing.T) {
		handler := func(c *gin.Context) {
			c.Status(204)
		}

		engine.RegisterRoute("DELETE", "/custom/delete", handler)

		req := httptest.NewRequest("DELETE", "/custom/delete", nil)
		w := httptest.NewRecorder()
		engine.ginEngine.ServeHTTP(w, req)

		if w.Code != 204 {
			t.Errorf("status code = %d, want 204", w.Code)
		}
	})

	t.Run("register invalid method defaults to GET", func(t *testing.T) {
		handler := func(c *gin.Context) {
			c.JSON(200, gin.H{"message": "default"})
		}

		engine.RegisterRoute("INVALID", "/custom/default", handler)

		req := httptest.NewRequest("GET", "/custom/default", nil)
		w := httptest.NewRecorder()
		engine.ginEngine.ServeHTTP(w, req)

		if w.Code != 200 {
			t.Errorf("status code = %d, want 200", w.Code)
		}
	})
}

func TestEngine_RegisterGroup(t *testing.T) {
	gin.SetMode(gin.TestMode)

	engine, err := NewEngine()
	if err != nil {
		t.Fatalf("NewEngine() error = %v", err)
	}

	if err := engine.Initialize(); err != nil {
		t.Fatalf("Initialize() error = %v", err)
	}

	t.Run("register route group", func(t *testing.T) {
		group := engine.RegisterGroup("/api/v1")

		group.GET("/users", func(c *gin.Context) {
			c.JSON(200, gin.H{"endpoint": "users"})
		})

		group.POST("/users", func(c *gin.Context) {
			c.JSON(201, gin.H{"endpoint": "create user"})
		})

		// Test GET endpoint
		req := httptest.NewRequest("GET", "/api/v1/users", nil)
		w := httptest.NewRecorder()
		engine.ginEngine.ServeHTTP(w, req)

		if w.Code != 200 {
			t.Errorf("GET status code = %d, want 200", w.Code)
		}

		// Test POST endpoint
		req = httptest.NewRequest("POST", "/api/v1/users", nil)
		w = httptest.NewRecorder()
		engine.ginEngine.ServeHTTP(w, req)

		if w.Code != 201 {
			t.Errorf("POST status code = %d, want 201", w.Code)
		}
	})

	t.Run("register group with middleware", func(t *testing.T) {
		middleware := func(c *gin.Context) {
			c.Header("X-Group-Middleware", "true")
			c.Next()
		}

		group := engine.RegisterGroup("/api/v2", middleware)

		group.GET("/test", func(c *gin.Context) {
			c.JSON(200, gin.H{"endpoint": "test"})
		})

		req := httptest.NewRequest("GET", "/api/v2/test", nil)
		w := httptest.NewRecorder()
		engine.ginEngine.ServeHTTP(w, req)

		if w.Code != 200 {
			t.Errorf("status code = %d, want 200", w.Code)
		}

		if w.Header().Get("X-Group-Middleware") != "true" {
			t.Error("X-Group-Middleware header should be set")
		}
	})
}

func TestEngine_GetRouteInfo(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("get route info with no routes", func(t *testing.T) {
		engine, err := NewEngine()
		if err != nil {
			t.Fatalf("NewEngine() error = %v", err)
		}

		if err := engine.Initialize(); err != nil {
			t.Fatalf("Initialize() error = %v", err)
		}

		// Disable default routes
		engine.serverConfig.EnableHealth = false
		engine.serverConfig.EnableMetrics = false
		engine.serverConfig.EnablePprof = false

		routes := engine.GetRouteInfo()

		// Should have some system routes even without custom routes
		if routes == nil {
			t.Error("GetRouteInfo() should not return nil")
		}
	})

	t.Run("get route info with controllers", func(t *testing.T) {
		ctrl := &mockController{
			name:   "testController",
			router: "/api/test",
		}

		engine, err := NewEngine(
			RegisterControllers(ctrl),
			WithServerConfig(&ServerConfig{
				EnableHealth: false,
				EnableMetrics: false,
				EnablePprof: false,
			}),
		)
		if err != nil {
			t.Fatalf("NewEngine() error = %v", err)
		}

		if err := engine.Initialize(); err != nil {
			t.Fatalf("Initialize() error = %v", err)
		}

		routes := engine.GetRouteInfo()

		found := false
		for _, route := range routes {
			if route.Path == "/api/test" && route.Method == "GET" {
				found = true
				break
			}
		}

		if !found {
			t.Error("GetRouteInfo() should include /api/test route")
		}
	})

	t.Run("get route info structure", func(t *testing.T) {
		engine, err := NewEngine()
		if err != nil {
			t.Fatalf("NewEngine() error = %v", err)
		}

		if err := engine.Initialize(); err != nil {
			t.Fatalf("Initialize() error = %v", err)
		}

		routes := engine.GetRouteInfo()

		for _, route := range routes {
			if route.Method == "" {
				t.Error("RouteInfo.Method should not be empty")
			}

			if route.Path == "" {
				t.Error("RouteInfo.Path should not be empty")
			}

			if route.Handler == "" {
				t.Error("RouteInfo.Handler should not be empty")
			}
		}
	})
}

func TestEngine_RegisterControllerHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("register single controller handler", func(t *testing.T) {
		engine, err := NewEngine()
		if err != nil {
			t.Fatalf("NewEngine() error = %v", err)
		}

		if err := engine.Initialize(); err != nil {
			t.Fatalf("Initialize() error = %v", err)
		}

		ctrl := &mockController{
			name:   "testController",
			router: "/custom/controller [POST]",
			handle: func(c *gin.Context) {
				c.JSON(201, gin.H{"registered": "manually"})
			},
		}

		engine.RegisterControllerHandler(ctrl)

		req := httptest.NewRequest("POST", "/custom/controller", nil)
		w := httptest.NewRecorder()
		engine.ginEngine.ServeHTTP(w, req)

		if w.Code != 201 {
			t.Errorf("status code = %d, want 201", w.Code)
		}
	})
}

func TestRouteInfo_Structure(t *testing.T) {
	info := RouteInfo{
		Method:  "GET",
		Path:    "/api/test",
		Handler: "github.com/example/package.(*Controller).Handle-fm",
	}

	if info.Method != "GET" {
		t.Errorf("Method = %s, want GET", info.Method)
	}

	if info.Path != "/api/test" {
		t.Errorf("Path = %s, want /api/test", info.Path)
	}

	if info.Handler == "" {
		t.Error("Handler should not be empty")
	}
}

func TestControllerRegistrationWithDifferentMethods(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		method        string
		router        string
		expectedStatus int
	}{
		{"GET", "/test/get", 200},
		{"POST", "/test/post", 201},
		{"PUT", "/test/put", 200},
		{"DELETE", "/test/delete", 204},
		{"PATCH", "/test/patch", 200},
		{"HEAD", "/test/head", 200},
		{"OPTIONS", "/test/options", 200},
	}

	for _, tt := range tests {
		t.Run(tt.method, func(t *testing.T) {
			lowerMethod := toLower(tt.method)
			ctrl := &mockController{
				name:   tt.method + "Controller",
				router: "/test/" + lowerMethod + " [" + tt.method + "]",
				handle: func(c *gin.Context) {
					c.Status(tt.expectedStatus)
				},
			}

			engine, err := NewEngine(RegisterControllers(ctrl))
			if err != nil {
				t.Fatalf("NewEngine() error = %v", err)
			}

			if err := engine.Initialize(); err != nil {
				t.Fatalf("Initialize() error = %v", err)
			}

			req := httptest.NewRequest(tt.method, "/test/"+lowerMethod, nil)
			w := httptest.NewRecorder()
			engine.ginEngine.ServeHTTP(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("status code = %d, want %d", w.Code, tt.expectedStatus)
			}
		})
	}
}

func toLower(s string) string {
	// Simple lowercase conversion for testing
	switch s {
	case "GET":
		return "get"
	case "POST":
		return "post"
	case "PUT":
		return "put"
	case "DELETE":
		return "delete"
	case "PATCH":
		return "patch"
	case "HEAD":
		return "head"
	case "OPTIONS":
		return "options"
	default:
		return s
	}
}

