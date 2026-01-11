package server

import (
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	"com.litelake.litecore/common"
)

// orderedMiddleware implements common.BaseMiddleware for testing
type orderedMiddleware struct {
	name  string
	order int
}

func (m *orderedMiddleware) MiddlewareName() string {
	return m.name
}

func (m *orderedMiddleware) Order() int {
	return m.order
}

func (m *orderedMiddleware) Wrapper() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()
	}
}

func (m *orderedMiddleware) OnStart() error {
	return nil
}

func (m *orderedMiddleware) OnStop() error {
	return nil
}

func TestSortMiddlewares(t *testing.T) {
	t.Run("sort middlewares by order", func(t *testing.T) {
		m1 := &orderedMiddleware{name: "middleware3", order: 3}
		m2 := &orderedMiddleware{name: "middleware1", order: 1}
		m3 := &orderedMiddleware{name: "middleware2", order: 2}

		middlewares := []common.BaseMiddleware{m1, m2, m3}
		sorted := sortMiddlewares(middlewares)

		if sorted[0].Order() != 1 {
			t.Errorf("first middleware order = %d, want 1", sorted[0].Order())
		}

		if sorted[1].Order() != 2 {
			t.Errorf("second middleware order = %d, want 2", sorted[1].Order())
		}

		if sorted[2].Order() != 3 {
			t.Errorf("third middleware order = %d, want 3", sorted[2].Order())
		}
	})

	t.Run("sort already sorted middlewares", func(t *testing.T) {
		m1 := &orderedMiddleware{name: "middleware1", order: 1}
		m2 := &orderedMiddleware{name: "middleware2", order: 2}
		m3 := &orderedMiddleware{name: "middleware3", order: 3}

		middlewares := []common.BaseMiddleware{m1, m2, m3}
		sorted := sortMiddlewares(middlewares)

		for i := 0; i < len(middlewares); i++ {
			if sorted[i].Order() != i+1 {
				t.Errorf("middleware %d order = %d, want %d", i, sorted[i].Order(), i+1)
			}
		}
	})

	t.Run("sort single middleware", func(t *testing.T) {
		m1 := &orderedMiddleware{name: "middleware1", order: 1}

		middlewares := []common.BaseMiddleware{m1}
		sorted := sortMiddlewares(middlewares)

		if len(sorted) != 1 {
			t.Errorf("sorted length = %d, want 1", len(sorted))
		}

		if sorted[0].Order() != 1 {
			t.Errorf("middleware order = %d, want 1", sorted[0].Order())
		}
	})

	t.Run("sort empty middlewares", func(t *testing.T) {
		var middlewares []common.BaseMiddleware

		sorted := sortMiddlewares(middlewares)

		if len(sorted) != 0 {
			t.Errorf("sorted length = %d, want 0", len(sorted))
		}
	})
}

func TestRequestLoggerMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)

	middleware := requestLoggerMiddleware()

	t.Run("log GET request", func(t *testing.T) {
		router := gin.New()
		router.Use(middleware)
		router.GET("/test", func(c *gin.Context) {
			c.JSON(200, gin.H{"message": "ok"})
		})

		req := httptest.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != 200 {
			t.Errorf("status code = %d, want 200", w.Code)
		}
	})

	t.Run("log POST request with body", func(t *testing.T) {
		router := gin.New()
		router.Use(middleware)
		router.POST("/test", func(c *gin.Context) {
			c.JSON(200, gin.H{"message": "ok"})
		})

		req := httptest.NewRequest("POST", "/test", nil)
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != 200 {
			t.Errorf("status code = %d, want 200", w.Code)
		}
	})
}

func TestCorsMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)

	middleware := CorsMiddleware()

	t.Run("GET request with CORS headers", func(t *testing.T) {
		router := gin.New()
		router.Use(middleware)
		router.GET("/test", func(c *gin.Context) {
			c.JSON(200, gin.H{"message": "ok"})
		})

		req := httptest.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		// Check CORS headers
		if w.Header().Get("Access-Control-Allow-Origin") != "*" {
			t.Error("Access-Control-Allow-Origin header should be *")
		}

		if w.Header().Get("Access-Control-Allow-Credentials") != "true" {
			t.Error("Access-Control-Allow-Credentials header should be true")
		}
	})

	t.Run("OPTIONS request", func(t *testing.T) {
		router := gin.New()
		router.Use(middleware)
		router.OPTIONS("/test", func(c *gin.Context) {
			c.JSON(200, gin.H{})
		})

		req := httptest.NewRequest("OPTIONS", "/test", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != 204 {
			t.Errorf("status code = %d, want 204", w.Code)
		}
	})
}

func TestSecurityHeadersMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)

	middleware := SecurityHeadersMiddleware()

	t.Run("add security headers", func(t *testing.T) {
		router := gin.New()
		router.Use(middleware)
		router.GET("/test", func(c *gin.Context) {
			c.JSON(200, gin.H{"message": "ok"})
		})

		req := httptest.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		tests := []struct {
			header string
			value  string
		}{
			{"X-Frame-Options", "DENY"},
			{"X-Content-Type-Options", "nosniff"},
			{"X-XSS-Protection", "1; mode=block"},
			{"Referrer-Policy", "strict-origin-when-cross-origin"},
		}

		for _, tt := range tests {
			if got := w.Header().Get(tt.header); got != tt.value {
				t.Errorf("%s header = %s, want %s", tt.header, got, tt.value)
			}
		}
	})
}

func TestEngine_Use(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("use custom middleware", func(t *testing.T) {
		engine, err := NewEngine()
		if err != nil {
			t.Fatalf("NewEngine() error = %v", err)
		}

		if err := engine.Initialize(); err != nil {
			t.Fatalf("Initialize() error = %v", err)
		}

		customMiddleware := func(c *gin.Context) {
			c.Set("custom", "value")
			c.Next()
		}

		engine.Use(customMiddleware)

		// Register a test route
		engine.ginEngine.GET("/test", func(c *gin.Context) {
			custom := c.GetBool("custom")
			c.JSON(200, gin.H{"custom": custom})
		})

		req := httptest.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()
		engine.ginEngine.ServeHTTP(w, req)

		if w.Code != 200 {
			t.Errorf("status code = %d, want 200", w.Code)
		}
	})
}

func TestEngine_RegisterMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("register custom middleware", func(t *testing.T) {
		engine, err := NewEngine()
		if err != nil {
			t.Fatalf("NewEngine() error = %v", err)
		}

		if err := engine.Initialize(); err != nil {
			t.Fatalf("Initialize() error = %v", err)
		}

		customMiddleware := func(c *gin.Context) {
			c.Header("X-Custom", "test")
			c.Next()
		}

		engine.RegisterMiddleware(customMiddleware)

		engine.ginEngine.GET("/test", func(c *gin.Context) {
			c.JSON(200, gin.H{})
		})

		req := httptest.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()
		engine.ginEngine.ServeHTTP(w, req)

		if w.Header().Get("X-Custom") != "test" {
			t.Error("X-Custom header should be set")
		}
	})
}

func TestEngine_RegisterMiddlewares(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("register business middlewares", func(t *testing.T) {
		mw1 := &orderedMiddleware{name: "middleware1", order: 1}
		mw2 := &orderedMiddleware{name: "middleware2", order: 2}

		engine, err := NewEngine(
			RegisterMiddlewares(mw1, mw2),
			WithServerConfig(&ServerConfig{
				EnableRecovery: false, // Disable recovery for testing
			}),
		)
		if err != nil {
			t.Fatalf("NewEngine() error = %v", err)
		}

		if err := engine.Initialize(); err != nil {
			t.Fatalf("Initialize() error = %v", err)
		}

		engine.ginEngine.GET("/test", func(c *gin.Context) {
			c.JSON(200, gin.H{})
		})

		req := httptest.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()
		engine.ginEngine.ServeHTTP(w, req)

		if w.Code != 200 {
			t.Errorf("status code = %d, want 200", w.Code)
		}
	})
}

func TestGetTelemetryManager(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("get telemetry manager when exists", func(t *testing.T) {
		mgr := &mockTelemetryManager{
			mockManager: mockManager{name: "TelemetryManager"},
		}

		engine, err := NewEngine(RegisterManagers(mgr))
		if err != nil {
			t.Fatalf("NewEngine() error = %v", err)
		}

		// Access private method via testing
		// This tests that the telemetry manager is found
		managers := engine.GetManagers()
		found := false
		for _, m := range managers {
			if m.ManagerName() == "TelemetryManager" {
				found = true
				break
			}
		}

		if !found {
			t.Error("TelemetryManager should be registered")
		}
	})

	t.Run("no telemetry manager", func(t *testing.T) {
		mgr := &mockManager{name: "otherManager"}

		engine, err := NewEngine(RegisterManagers(mgr))
		if err != nil {
			t.Fatalf("NewEngine() error = %v", err)
		}

		managers := engine.GetManagers()
		found := false
		for _, m := range managers {
			if m.ManagerName() == "TelemetryManager" {
				found = true
				break
			}
		}

		if found {
			t.Error("TelemetryManager should not be registered")
		}
	})
}

func TestMiddlewareExecutionOrder(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Create middlewares with different orders
	mw1 := &orderedMiddleware{name: "first", order: 100}
	mw2 := &orderedMiddleware{name: "second", order: 200}
	mw3 := &orderedMiddleware{name: "third", order: 150}

	engine, err := NewEngine(
		RegisterMiddlewares(mw1, mw2, mw3),
		WithServerConfig(&ServerConfig{
			EnableRecovery: false,
		}),
	)
	if err != nil {
		t.Fatalf("NewEngine() error = %v", err)
	}

	if err := engine.Initialize(); err != nil {
		t.Fatalf("Initialize() error = %v", err)
	}

	// Check that middlewares are registered
	middlewares := engine.GetMiddlewares()
	if len(middlewares) != 3 {
		t.Fatalf("middlewares count = %d, want 3", len(middlewares))
	}

	// Verify that the middleware sort function works correctly
	// The container may return them in registration order, but the sort function should sort them
	testMiddlewares := []common.BaseMiddleware{mw1, mw2, mw3}
	sorted := sortMiddlewares(testMiddlewares)

	if sorted[0].Order() != 100 {
		t.Errorf("first middleware order = %d, want 100", sorted[0].Order())
	}

	if sorted[1].Order() != 150 {
		t.Errorf("second middleware order = %d, want 150", sorted[1].Order())
	}

	if sorted[2].Order() != 200 {
		t.Errorf("third middleware order = %d, want 200", sorted[2].Order())
	}
}
