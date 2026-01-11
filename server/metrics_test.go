package server

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func setupMetricsEngine(enablePprof bool) (*Engine, *gin.Engine, error) {
	gin.SetMode(gin.TestMode)

	engine, err := NewEngine(
		WithServerConfig(&ServerConfig{
			EnableMetrics: true,
			EnablePprof:   enablePprof,
		}),
	)
	if err != nil {
		return nil, nil, err
	}

	if err := engine.Initialize(); err != nil {
		return nil, nil, err
	}

	return engine, engine.ginEngine, nil
}

func TestMetricsHandler(t *testing.T) {
	_, ginEngine, err := setupMetricsEngine(false)
	if err != nil {
		t.Fatalf("setupMetricsEngine() error = %v", err)
	}

	req := httptest.NewRequest("GET", "/metrics", nil)
	w := httptest.NewRecorder()
	ginEngine.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status code = %d, want %d", w.Code, http.StatusOK)
	}

	var response map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if response["server"] != "litecore-go" {
		t.Errorf("server = %v, want litecore-go", response["server"])
	}

	if response["status"] != "running" {
		t.Errorf("status = %v, want running", response["status"])
	}

	if response["version"] != "1.0.0" {
		t.Errorf("version = %v, want 1.0.0", response["version"])
	}
}

func TestRegisterMetricsRoute(t *testing.T) {
	engine, err := NewEngine(
		WithServerConfig(&ServerConfig{
			EnableMetrics: true,
		}),
	)
	if err != nil {
		t.Fatalf("NewEngine() error = %v", err)
	}

	if err := engine.Initialize(); err != nil {
		t.Fatalf("Initialize() error = %v", err)
	}

	// Check that /metrics route is registered
	routes := engine.GetRouteInfo()
	metricsFound := false
	for _, route := range routes {
		if route.Path == "/metrics" && route.Method == "GET" {
			metricsFound = true
			break
		}
	}

	if !metricsFound {
		t.Error("/metrics route should be registered when EnableMetrics is true")
	}
}

func TestRegisterPprofRoutes(t *testing.T) {
	t.Run("ppprof routes registered when enabled", func(t *testing.T) {
		engine, err := NewEngine(
			WithServerConfig(&ServerConfig{
				EnablePprof: true,
			}),
		)
		if err != nil {
			t.Fatalf("NewEngine() error = %v", err)
		}

		if err := engine.Initialize(); err != nil {
			t.Fatalf("Initialize() error = %v", err)
		}

		routes := engine.GetRouteInfo()
		pprofRoutes := []string{
			"/debug/pprof/",
			"/debug/pprof/cmdline",
			"/debug/pprof/profile",
			"/debug/pprof/symbol",
			"/debug/pprof/trace",
			"/debug/pprof/allocs",
			"/debug/pprof/block",
			"/debug/pprof/goroutine",
			"/debug/pprof/heap",
			"/debug/pprof/mutex",
			"/debug/pprof/threadcreate",
		}

		for _, expectedRoute := range pprofRoutes {
			found := false
			for _, route := range routes {
				if route.Path == expectedRoute && route.Method == "GET" {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("pprof route %s should be registered", expectedRoute)
			}
		}
	})

	t.Run("pprof routes not registered when disabled", func(t *testing.T) {
		engine, err := NewEngine(
			WithServerConfig(&ServerConfig{
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
		for _, route := range routes {
			if len(route.Path) > 10 && route.Path[:10] == "/debug/pprof" {
				t.Errorf("pprof route %s should not be registered when EnablePprof is false", route.Path)
			}
		}
	})
}

func TestPprofHandlers(t *testing.T) {
	_, ginEngine, err := setupMetricsEngine(true)
	if err != nil {
		t.Fatalf("setupMetricsEngine() error = %v", err)
	}

	t.Run("pprof index handler", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/debug/pprof/", nil)
		w := httptest.NewRecorder()
		ginEngine.ServeHTTP(w, req)

		// pprof index returns HTML, so we just check it's accessible
		if w.Code != http.StatusOK {
			t.Errorf("status code = %d, want %d", w.Code, http.StatusOK)
		}
	})

	t.Run("pprof heap handler", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/debug/pprof/heap", nil)
		w := httptest.NewRecorder()
		ginEngine.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("status code = %d, want %d", w.Code, http.StatusOK)
		}
	})

	t.Run("pprof goroutine handler", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/debug/pprof/goroutine", nil)
		w := httptest.NewRecorder()
		ginEngine.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("status code = %d, want %d", w.Code, http.StatusOK)
		}
	})

	t.Run("pprof cmdline handler", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/debug/pprof/cmdline", nil)
		w := httptest.NewRecorder()
		ginEngine.ServeHTTP(w, req)

		// cmdline handler should return 200
		if w.Code != http.StatusOK {
			t.Errorf("status code = %d, want %d", w.Code, http.StatusOK)
		}
	})
}

func TestEngine_GetMetrics(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("get metrics with no components", func(t *testing.T) {
		engine, err := NewEngine()
		if err != nil {
			t.Fatalf("NewEngine() error = %v", err)
		}

		if err := engine.Initialize(); err != nil {
			t.Fatalf("Initialize() error = %v", err)
		}

		metrics := engine.GetMetrics()
		if metrics == nil {
			t.Fatal("GetMetrics() should not return nil")
		}

		// Check default counts
		if metrics["managers"] != 0 {
			t.Errorf("managers = %v, want 0", metrics["managers"])
		}

		if metrics["entities"] != 0 {
			t.Errorf("entities = %v, want 0", metrics["entities"])
		}
	})

	t.Run("get metrics with components", func(t *testing.T) {
		mgr := &mockManager{name: "testManager"}
		ent := &mockEntity{name: "testEntity"}
		repo := &mockRepository{name: "testRepository"}
		svc := &mockService{name: "testService"}
		ctrl := &mockController{name: "testController", router: "/api/test"}
		mw := &mockMiddleware{name: "testMiddleware", order: 1}

		engine, err := NewEngine(
			RegisterManagers(mgr),
			RegisterEntities(ent),
			RegisterRepositories(repo),
			RegisterServices(svc),
			RegisterControllers(ctrl),
			RegisterMiddlewares(mw),
		)
		if err != nil {
			t.Fatalf("NewEngine() error = %v", err)
		}

		if err := engine.Initialize(); err != nil {
			t.Fatalf("Initialize() error = %v", err)
		}

		metrics := engine.GetMetrics()
		if metrics["managers"] != 1 {
			t.Errorf("managers = %v, want 1", metrics["managers"])
		}

		if metrics["entities"] != 1 {
			t.Errorf("entities = %v, want 1", metrics["entities"])
		}

		if metrics["repositories"] != 1 {
			t.Errorf("repositories = %v, want 1", metrics["repositories"])
		}

		if metrics["services"] != 1 {
			t.Errorf("services = %v, want 1", metrics["services"])
		}

		if metrics["controllers"] != 1 {
			t.Errorf("controllers = %v, want 1", metrics["controllers"])
		}

		if metrics["middlewares"] != 1 {
			t.Errorf("middlewares = %v, want 1", metrics["middlewares"])
		}
	})

	t.Run("get metrics with multiple components", func(t *testing.T) {
		mgr1 := &mockManager{name: "manager1"}
		mgr2 := &mockManager{name: "manager2"}
		ent1 := &mockEntity{name: "entity1"}
		ent2 := &mockEntity{name: "entity2"}

		engine, err := NewEngine(
			RegisterManagers(mgr1, mgr2),
			RegisterEntities(ent1, ent2),
		)
		if err != nil {
			t.Fatalf("NewEngine() error = %v", err)
		}

		if err := engine.Initialize(); err != nil {
			t.Fatalf("Initialize() error = %v", err)
		}

		metrics := engine.GetMetrics()
		if metrics["managers"] != 2 {
			t.Errorf("managers = %v, want 2", metrics["managers"])
		}

		if metrics["entities"] != 2 {
			t.Errorf("entities = %v, want 2", metrics["entities"])
		}
	})
}

func TestMetricsResponse_Structure(t *testing.T) {
	// MetricsResponse is a placeholder struct
	// This test ensures it compiles and can be used
	var response MetricsResponse
	_ = response
}

func TestEngine_DisabledMetrics(t *testing.T) {
	gin.SetMode(gin.TestMode)

	engine, err := NewEngine(
		WithServerConfig(&ServerConfig{
			EnableMetrics: false,
		}),
	)
	if err != nil {
		t.Fatalf("NewEngine() error = %v", err)
	}

	if err := engine.Initialize(); err != nil {
		t.Fatalf("Initialize() error = %v", err)
	}

	// Check that /metrics route is not registered
	routes := engine.GetRouteInfo()
	for _, route := range routes {
		if route.Path == "/metrics" {
			t.Error("/metrics route should not be registered when EnableMetrics is false")
		}
	}

	// But GetMetrics() should still work
	metrics := engine.GetMetrics()
	if metrics == nil {
		t.Error("GetMetrics() should still return a map even when metrics are disabled")
	}
}

func TestEngine_MetricsWithManagers(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Test with a manager that implements TelemetryManager interface
	telemetryMgr := &mockTelemetryManager{}

	engine, err := NewEngine(
		RegisterManagers(telemetryMgr),
		WithServerConfig(&ServerConfig{
			EnableMetrics: true,
		}),
	)
	if err != nil {
		t.Fatalf("NewEngine() error = %v", err)
	}

	if err := engine.Initialize(); err != nil {
		t.Fatalf("Initialize() error = %v", err)
	}

	// Metrics endpoint should be accessible
	req := httptest.NewRequest("GET", "/metrics", nil)
	w := httptest.NewRecorder()
	engine.ginEngine.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status code = %d, want %d", w.Code, http.StatusOK)
	}

	metrics := engine.GetMetrics()
	if metrics["managers"] != 1 {
		t.Errorf("managers = %v, want 1", metrics["managers"])
	}
}
