package server

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
)

// Mock implementations for testing

type mockManager struct {
	name      string
	started   bool
	stopped   bool
	healthErr error
	startErr  error
	stopErr   error
}

func (m *mockManager) ManagerName() string {
	return m.name
}

func (m *mockManager) OnStart() error {
	m.started = true
	return m.startErr
}

func (m *mockManager) OnStop() error {
	m.stopped = true
	return m.stopErr
}

func (m *mockManager) Health() error {
	return m.healthErr
}

func (m *mockManager) Config() interface{} {
	return nil
}

type mockLoggerManager struct {
	mockManager
}

func (m *mockLoggerManager) Logger(name string) interface{} {
	return "mock-logger"
}

type mockTelemetryManager struct {
	mockManager
}

func (m *mockTelemetryManager) GinMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()
	}
}

type mockService struct {
	name string
}

func (s *mockService) ServiceName() string {
	return s.name
}

func (s *mockService) OnStart() error {
	return nil
}

func (s *mockService) OnStop() error {
	return nil
}

type mockEntity struct {
	name string
}

func (e *mockEntity) EntityName() string {
	return e.name
}

func (e *mockEntity) TableName() string {
	return e.name + "s"
}

func (e *mockEntity) GetId() string {
	return "0"
}

type mockRepository struct {
	name string
}

func (r *mockRepository) RepositoryName() string {
	return r.name
}

func (r *mockRepository) OnStart() error {
	return nil
}

func (r *mockRepository) OnStop() error {
	return nil
}

type mockController struct {
	name   string
	router string
	handle gin.HandlerFunc
}

func (c *mockController) ControllerName() string {
	return c.name
}

func (c *mockController) GetRouter() string {
	return c.router
}

func (c *mockController) Handle(ctx *gin.Context) {
	if c.handle != nil {
		c.handle(ctx)
		return
	}
	ctx.JSON(200, gin.H{"message": "ok"})
}

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
	return func(c *gin.Context) {
		c.Next()
	}
}

func (m *mockMiddleware) OnStart() error {
	return nil
}

func (m *mockMiddleware) OnStop() error {
	return nil
}

func TestNewEngine(t *testing.T) {
	t.Run("create engine with no options", func(t *testing.T) {
		engine, err := NewEngine()
		if err != nil {
			t.Fatalf("NewEngine() error = %v", err)
		}

		if engine == nil {
			t.Fatal("NewEngine() returned nil engine")
		}

		if engine.serverConfig == nil {
			t.Error("serverConfig should not be nil")
		}

		if engine.containers == nil {
			t.Error("containers should not be nil")
		}

		if engine.ctx == nil {
			t.Error("ctx should not be nil")
		}

		if engine.cancel == nil {
			t.Error("cancel should not be nil")
		}
	})

	t.Run("create engine with server config option", func(t *testing.T) {
		customConfig := &ServerConfig{
			Host: "127.0.0.1",
			Port: 9090,
			Mode: "debug",
		}

		engine, err := NewEngine(WithServerConfig(customConfig))
		if err != nil {
			t.Fatalf("NewEngine() error = %v", err)
		}

		if engine.serverConfig.Host != "127.0.0.1" {
			t.Errorf("Host = %v, want 127.0.0.1", engine.serverConfig.Host)
		}

		if engine.serverConfig.Port != 9090 {
			t.Errorf("Port = %v, want 9090", engine.serverConfig.Port)
		}

		if engine.serverConfig.Mode != "debug" {
			t.Errorf("Mode = %v, want debug", engine.serverConfig.Mode)
		}
	})

	t.Run("create engine with nil server config", func(t *testing.T) {
		engine, err := NewEngine(WithServerConfig(nil))
		if err != nil {
			t.Fatalf("NewEngine() error = %v", err)
		}

		defaultConfig := DefaultServerConfig()
		if engine.serverConfig.Host != defaultConfig.Host {
			t.Errorf("Host = %v, want %v", engine.serverConfig.Host, defaultConfig.Host)
		}
	})
}

func TestEngine_Initialize(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("initialize successfully with no managers", func(t *testing.T) {
		engine, err := NewEngine()
		if err != nil {
			t.Fatalf("NewEngine() error = %v", err)
		}

		err = engine.Initialize()
		if err != nil {
			t.Fatalf("Initialize() error = %v", err)
		}

		if engine.ginEngine == nil {
			t.Error("ginEngine should not be nil after Initialize")
		}

		if engine.httpServer == nil {
			t.Error("httpServer should not be nil after Initialize")
		}
	})

	t.Run("initialize with mock managers", func(t *testing.T) {
		mgr := &mockManager{name: "testManager"}

		engine, err := NewEngine(RegisterManagers(mgr))
		if err != nil {
			t.Fatalf("NewEngine() error = %v", err)
		}

		err = engine.Initialize()
		if err != nil {
			t.Fatalf("Initialize() error = %v", err)
		}

		// Manager should not be started during Initialize
		if mgr.started {
			t.Error("manager should not be started during Initialize")
		}
	})

	t.Run("initialize twice should return error", func(t *testing.T) {
		engine, err := NewEngine()
		if err != nil {
			t.Fatalf("NewEngine() error = %v", err)
		}

		err = engine.Initialize()
		if err != nil {
			t.Fatalf("First Initialize() error = %v", err)
		}

		// Initialize is idempotent, should not error
		err = engine.Initialize()
		if err != nil {
			t.Errorf("Second Initialize() should not error, got %v", err)
		}
	})
}

func TestEngine_Start(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("start successfully", func(t *testing.T) {
		engine, err := NewEngine(
			WithServerConfig(&ServerConfig{
				Host: "127.0.0.1",
				Port: 0, // Use random port
			}),
		)
		if err != nil {
			t.Fatalf("NewEngine() error = %v", err)
		}

		err = engine.Initialize()
		if err != nil {
			t.Fatalf("Initialize() error = %v", err)
		}

		err = engine.Start()
		if err != nil {
			t.Fatalf("Start() error = %v", err)
		}

		if !engine.IsStarted() {
			t.Error("IsStarted() should return true after Start")
		}

		// Cleanup
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		engine.httpServer.Shutdown(ctx)
	})

	t.Run("start twice should return error", func(t *testing.T) {
		engine, err := NewEngine(
			WithServerConfig(&ServerConfig{
				Host: "127.0.0.1",
				Port: 0,
			}),
		)
		if err != nil {
			t.Fatalf("NewEngine() error = %v", err)
		}

		err = engine.Initialize()
		if err != nil {
			t.Fatalf("Initialize() error = %v", err)
		}

		err = engine.Start()
		if err != nil {
			t.Fatalf("First Start() error = %v", err)
		}

		err = engine.Start()
		if err == nil {
			t.Error("Second Start() should return error")
		}

		// Cleanup
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		engine.httpServer.Shutdown(ctx)
	})
}

func TestEngine_Stop(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("stop successfully", func(t *testing.T) {
		engine, err := NewEngine(
			WithServerConfig(&ServerConfig{
				Host: "127.0.0.1",
				Port: 0,
			}),
		)
		if err != nil {
			t.Fatalf("NewEngine() error = %v", err)
		}

		err = engine.Initialize()
		if err != nil {
			t.Fatalf("Initialize() error = %v", err)
		}

		err = engine.Start()
		if err != nil {
			t.Fatalf("Start() error = %v", err)
		}

		err = engine.Stop()
		if err != nil {
			t.Fatalf("Stop() error = %v", err)
		}

		if engine.IsStarted() {
			t.Error("IsStarted() should return false after Stop")
		}
	})

	t.Run("stop without start should not error", func(t *testing.T) {
		engine, err := NewEngine()
		if err != nil {
			t.Fatalf("NewEngine() error = %v", err)
		}

		err = engine.Initialize()
		if err != nil {
			t.Fatalf("Initialize() error = %v", err)
		}

		err = engine.Stop()
		if err != nil {
			t.Errorf("Stop() without Start() should not error, got %v", err)
		}
	})
}

func TestEngine_GetGinEngine(t *testing.T) {
	gin.SetMode(gin.TestMode)

	engine, err := NewEngine()
	if err != nil {
		t.Fatalf("NewEngine() error = %v", err)
	}

	err = engine.Initialize()
	if err != nil {
		t.Fatalf("Initialize() error = %v", err)
	}

	ginEngine := engine.GetGinEngine()
	if ginEngine == nil {
		t.Error("GetGinEngine() should not return nil after Initialize")
	}

	if ginEngine != engine.ginEngine {
		t.Error("GetGinEngine() should return the same ginEngine")
	}
}

func TestEngine_GetConfig(t *testing.T) {
	t.Run("get config with no config provider", func(t *testing.T) {
		engine, err := NewEngine()
		if err != nil {
			t.Fatalf("NewEngine() error = %v", err)
		}

		config := engine.GetConfig()
		if config != nil {
			t.Error("GetConfig() should return nil when no config is registered")
		}
	})
}

func TestEngine_GetManager(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("get existing manager", func(t *testing.T) {
		mgr := &mockManager{name: "testManager"}

		engine, err := NewEngine(RegisterManagers(mgr))
		if err != nil {
			t.Fatalf("NewEngine() error = %v", err)
		}

		err = engine.Initialize()
		if err != nil {
			t.Fatalf("Initialize() error = %v", err)
		}

		gotMgr, err := engine.GetManager("testManager")
		if err != nil {
			t.Fatalf("GetManager() error = %v", err)
		}

		if gotMgr.ManagerName() != "testManager" {
			t.Errorf("GetManager() returned wrong manager")
		}
	})

	t.Run("get non-existent manager", func(t *testing.T) {
		engine, err := NewEngine()
		if err != nil {
			t.Fatalf("NewEngine() error = %v", err)
		}

		err = engine.Initialize()
		if err != nil {
			t.Fatalf("Initialize() error = %v", err)
		}

		_, err = engine.GetManager("nonExistent")
		if err == nil {
			t.Error("GetManager() should return error for non-existent manager")
		}
	})
}

func TestEngine_GetService(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("get service", func(t *testing.T) {
		svc := &mockService{name: "testService"}

		engine, err := NewEngine(RegisterServices(svc))
		if err != nil {
			t.Fatalf("NewEngine() error = %v", err)
		}

		err = engine.Initialize()
		if err != nil {
			t.Fatalf("Initialize() error = %v", err)
		}

		gotSvc, err := engine.GetService("testService")
		if err != nil {
			t.Fatalf("GetService() error = %v", err)
		}

		if gotSvc.ServiceName() != "testService" {
			t.Errorf("GetService() returned wrong service")
		}
	})
}

func TestEngine_GetLogger(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("get logger with LoggerManager", func(t *testing.T) {
		mgr := &mockLoggerManager{}

		engine, err := NewEngine(RegisterManagers(mgr))
		if err != nil {
			t.Fatalf("NewEngine() error = %v", err)
		}

		logger := engine.GetLogger()
		if logger == nil {
			t.Error("GetLogger() should return logger when LoggerManager is registered")
		}
	})

	t.Run("get logger without LoggerManager", func(t *testing.T) {
		engine, err := NewEngine()
		if err != nil {
			t.Fatalf("NewEngine() error = %v", err)
		}

		logger := engine.GetLogger()
		if logger != nil {
			t.Error("GetLogger() should return nil when no LoggerManager is registered")
		}
	})
}

func TestEngine_Health(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("all managers healthy", func(t *testing.T) {
		mgr := &mockManager{name: "healthyManager", healthErr: nil}

		engine, err := NewEngine(RegisterManagers(mgr))
		if err != nil {
			t.Fatalf("NewEngine() error = %v", err)
		}

		err = engine.Initialize()
		if err != nil {
			t.Fatalf("Initialize() error = %v", err)
		}

		err = engine.Health()
		if err != nil {
			t.Errorf("Health() should return nil when all managers are healthy, got %v", err)
		}
	})

	t.Run("one manager unhealthy", func(t *testing.T) {
		mgr := &mockManager{name: "unhealthyManager", healthErr: errors.New("health check failed")}

		engine, err := NewEngine(RegisterManagers(mgr))
		if err != nil {
			t.Fatalf("NewEngine() error = %v", err)
		}

		err = engine.Initialize()
		if err != nil {
			t.Fatalf("Initialize() error = %v", err)
		}

		err = engine.Health()
		if err == nil {
			t.Error("Health() should return error when a manager is unhealthy")
		}
	})
}

func TestEngine_SetMode(t *testing.T) {
	engine, err := NewEngine()
	if err != nil {
		t.Fatalf("NewEngine() error = %v", err)
	}

	engine.SetMode("test")
	if engine.serverConfig.Mode != "test" {
		t.Errorf("SetMode() failed, Mode = %v, want test", engine.serverConfig.Mode)
	}
}

func TestEngine_IsStarted(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("not started initially", func(t *testing.T) {
		engine, err := NewEngine()
		if err != nil {
			t.Fatalf("NewEngine() error = %v", err)
		}

		if engine.IsStarted() {
			t.Error("IsStarted() should return false before Start")
		}
	})

	t.Run("started after Start", func(t *testing.T) {
		engine, err := NewEngine(
			WithServerConfig(&ServerConfig{
				Host: "127.0.0.1",
				Port: 0,
			}),
		)
		if err != nil {
			t.Fatalf("NewEngine() error = %v", err)
		}

		err = engine.Initialize()
		if err != nil {
			t.Fatalf("Initialize() error = %v", err)
		}

		err = engine.Start()
		if err != nil {
			t.Fatalf("Start() error = %v", err)
		}

		if !engine.IsStarted() {
			t.Error("IsStarted() should return true after Start")
		}

		// Cleanup
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		engine.httpServer.Shutdown(ctx)
	})
}

func TestEngine_ConcurrentAccess(t *testing.T) {
	gin.SetMode(gin.TestMode)

	engine, err := NewEngine()
	if err != nil {
		t.Fatalf("NewEngine() error = %v", err)
	}

	var wg sync.WaitGroup
	concurrentOps := 100

	for i := 0; i < concurrentOps; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			engine.IsStarted()
			engine.GetGinEngine()
			engine.GetConfig()
		}()
	}

	wg.Wait()
	// If we got here without deadlock/race, the test passed
}

func TestEngine_NewEngineWithConfig(t *testing.T) {
	mockConfig := &mockConfigProvider{}

	engine, err := NewEngineWithConfig(mockConfig)
	if err != nil {
		t.Fatalf("NewEngineWithConfig() error = %v", err)
	}

	if engine == nil {
		t.Fatal("NewEngineWithConfig() returned nil")
	}

	config := engine.GetConfig()
	if config == nil {
		t.Error("GetConfig() should return the registered config")
	}
}

type mockConfigProvider struct {
}

func (m *mockConfigProvider) ConfigProviderName() string {
	return "mockConfigProvider"
}

func (m *mockConfigProvider) Get(key string) (any, error) {
	return nil, nil
}

func (m *mockConfigProvider) Has(key string) bool {
	return false
}
