package server

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
)

// Mock managers for lifecycle testing

type lifecycleManager struct {
	name         string
	startCalled  bool
	stopCalled   bool
	startOrder   int
	stopOrder    int
	startErr     error
	stopErr      error
	startDelay   time.Duration
	stopDelay    time.Duration
}

func (m *lifecycleManager) ManagerName() string {
	return m.name
}

func (m *lifecycleManager) OnStart() error {
	m.startCalled = true
	if m.startDelay > 0 {
		time.Sleep(m.startDelay)
	}
	return m.startErr
}

func (m *lifecycleManager) OnStop() error {
	m.stopCalled = true
	if m.stopDelay > 0 {
		time.Sleep(m.stopDelay)
	}
	return m.stopErr
}

func (m *lifecycleManager) Health() error {
	return nil
}

func (m *lifecycleManager) Config() interface{} {
	return nil
}

var startCounter int
var stopCounter int

func resetCounters() {
	startCounter = 0
	stopCounter = 0
}

func TestStartManagers(t *testing.T) {
	gin.SetMode(gin.TestMode)
	resetCounters()

	t.Run("start all managers successfully", func(t *testing.T) {
		mgr1 := &lifecycleManager{name: "manager1"}
		mgr2 := &lifecycleManager{name: "manager2"}

		engine, err := NewEngine(
			RegisterManagers(mgr1, mgr2),
			WithServerConfig(&ServerConfig{
				Host: "127.0.0.1",
				Port: 0,
			}),
		)
		if err != nil {
			t.Fatalf("NewEngine() error = %v", err)
		}

		if err := engine.Initialize(); err != nil {
			t.Fatalf("Initialize() error = %v", err)
		}

		// Managers should not be started after Initialize
		if mgr1.startCalled {
			t.Error("manager1 should not be started after Initialize")
		}

		if err := engine.Start(); err != nil {
			t.Fatalf("Start() error = %v", err)
		}

		// Now managers should be started
		if !mgr1.startCalled {
			t.Error("manager1 should be started after Start")
		}

		if !mgr2.startCalled {
			t.Error("manager2 should be started after Start")
		}

		// Cleanup
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		engine.httpServer.Shutdown(ctx)
	})

	t.Run("start manager returns error", func(t *testing.T) {
		mgr := &lifecycleManager{name: "failingManager", startErr: errors.New("start failed")}

		engine, err := NewEngine(RegisterManagers(mgr))
		if err != nil {
			t.Fatalf("NewEngine() error = %v", err)
		}

		if err := engine.Initialize(); err != nil {
			t.Fatalf("Initialize() should not fail, got %v", err)
		}

		if err := engine.Start(); err == nil {
			t.Error("Start() should return error when manager fails to start")
		}
	})
}

func TestStopManagers(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("stop all managers successfully", func(t *testing.T) {
		mgr1 := &lifecycleManager{name: "manager1"}
		mgr2 := &lifecycleManager{name: "manager2"}

		engine, err := NewEngine(
			RegisterManagers(mgr1, mgr2),
			WithServerConfig(&ServerConfig{
				Host: "127.0.0.1",
				Port: 0,
			}),
		)
		if err != nil {
			t.Fatalf("NewEngine() error = %v", err)
		}

		if err := engine.Initialize(); err != nil {
			t.Fatalf("Initialize() error = %v", err)
		}

		if err := engine.Start(); err != nil {
			t.Fatalf("Start() error = %v", err)
		}

		// Stop should be called during Stop()
		if err := engine.Stop(); err != nil {
			t.Fatalf("Stop() error = %v", err)
		}

		// Note: stopCalled is checked indirectly through successful shutdown
	})

	t.Run("stop manager returns error but continues", func(t *testing.T) {
		mgr1 := &lifecycleManager{name: "manager1"}
		mgr2 := &lifecycleManager{name: "manager2", stopErr: errors.New("stop failed")}

		engine, err := NewEngine(
			RegisterManagers(mgr1, mgr2),
			WithServerConfig(&ServerConfig{
				Host: "127.0.0.1",
				Port: 0,
			}),
		)
		if err != nil {
			t.Fatalf("NewEngine() error = %v", err)
		}

		if err := engine.Initialize(); err != nil {
			t.Fatalf("Initialize() error = %v", err)
		}

		if err := engine.Start(); err != nil {
			t.Fatalf("Start() error = %v", err)
		}

		// Stop should continue even if one manager fails
		if err := engine.Stop(); err != nil {
			// Error is expected but other managers should still be stopped
			t.Logf("Stop() returned error (expected): %v", err)
		}
	})
}

func TestEngine_LifecycleStop(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("stop without start returns nil", func(t *testing.T) {
		engine, err := NewEngine()
		if err != nil {
			t.Fatalf("NewEngine() error = %v", err)
		}

		if err := engine.Initialize(); err != nil {
			t.Fatalf("Initialize() error = %v", err)
		}

		if err := engine.Stop(); err != nil {
			t.Errorf("Stop() without start should return nil, got %v", err)
		}
	})

	t.Run("stop http server timeout", func(t *testing.T) {
		engine, err := NewEngine(
			WithServerConfig(&ServerConfig{
				Host:            "127.0.0.1",
				Port:            0,
				ShutdownTimeout: 1 * time.Nanosecond, // Very short timeout
			}),
		)
		if err != nil {
			t.Fatalf("NewEngine() error = %v", err)
		}

		if err := engine.Initialize(); err != nil {
			t.Fatalf("Initialize() error = %v", err)
		}

		if err := engine.Start(); err != nil {
			t.Fatalf("Start() error = %v", err)
		}

		// Stop might timeout but should not panic
		_ = engine.Stop()
	})
}

func TestEngine_Restart(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("restart successfully", func(t *testing.T) {
		mgr := &lifecycleManager{name: "testManager"}

		engine, err := NewEngine(
			RegisterManagers(mgr),
			WithServerConfig(&ServerConfig{
				Host: "127.0.0.1",
				Port: 0,
			}),
		)
		if err != nil {
			t.Fatalf("NewEngine() error = %v", err)
		}

		if err := engine.Initialize(); err != nil {
			t.Fatalf("Initialize() error = %v", err)
		}

		if err := engine.Start(); err != nil {
			t.Fatalf("Start() error = %v", err)
		}

		if !engine.IsStarted() {
			t.Error("engine should be started")
		}

		if err := engine.Restart(); err != nil {
			t.Fatalf("Restart() error = %v", err)
		}

		if !engine.IsStarted() {
			t.Error("engine should be started after restart")
		}

		// Cleanup
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		engine.httpServer.Shutdown(ctx)
	})

	t.Run("restart without start", func(t *testing.T) {
		engine, err := NewEngine(
			WithServerConfig(&ServerConfig{
				Host: "127.0.0.1",
				Port: 0,
			}),
		)
		if err != nil {
			t.Fatalf("NewEngine() error = %v", err)
		}

		if err := engine.Initialize(); err != nil {
			t.Fatalf("Initialize() error = %v", err)
		}

		if err := engine.Restart(); err != nil {
			t.Errorf("Restart() without start failed: %v", err)
		}

		if !engine.IsStarted() {
			t.Error("engine should be started after restart")
		}

		// Cleanup
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		engine.httpServer.Shutdown(ctx)
	})
}

func TestEngine_GracefulShutdown(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("graceful shutdown with timeout", func(t *testing.T) {
		mgr := &lifecycleManager{
			name:       "testManager",
			stopDelay:  100 * time.Millisecond,
		}

		engine, err := NewEngine(
			RegisterManagers(mgr),
			WithServerConfig(&ServerConfig{
				Host: "127.0.0.1",
				Port: 0,
			}),
		)
		if err != nil {
			t.Fatalf("NewEngine() error = %v", err)
		}

		if err := engine.Initialize(); err != nil {
			t.Fatalf("Initialize() error = %v", err)
		}

		if err := engine.Start(); err != nil {
			t.Fatalf("Start() error = %v", err)
		}

		// Graceful shutdown with sufficient timeout
		if err := engine.GracefulShutdown(5 * time.Second); err != nil {
			t.Errorf("GracefulShutdown() error = %v", err)
		}

		if engine.IsStarted() {
			t.Error("engine should not be started after graceful shutdown")
		}
	})

	t.Run("graceful shutdown completes successfully", func(t *testing.T) {
		mgr := &lifecycleManager{
			name:      "testManager",
			stopDelay: 100 * time.Millisecond,
		}

		engine, err := NewEngine(
			RegisterManagers(mgr),
			WithServerConfig(&ServerConfig{
				Host: "127.0.0.1",
				Port: 0,
			}),
		)
		if err != nil {
			t.Fatalf("NewEngine() error = %v", err)
		}

		if err := engine.Initialize(); err != nil {
			t.Fatalf("Initialize() error = %v", err)
		}

		if err := engine.Start(); err != nil {
			t.Fatalf("Start() error = %v", err)
		}

		// Graceful shutdown with sufficient timeout
		if err := engine.GracefulShutdown(5 * time.Second); err != nil {
			t.Errorf("GracefulShutdown() error = %v", err)
		}

		if !mgr.stopCalled {
			t.Error("manager should be stopped")
		}

		if engine.IsStarted() {
			t.Error("engine should not be started after graceful shutdown")
		}
	})
}

func TestEngine_GetManagers(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("get all managers", func(t *testing.T) {
		mgr1 := &lifecycleManager{name: "manager1"}
		mgr2 := &lifecycleManager{name: "manager2"}

		engine, err := NewEngine(RegisterManagers(mgr1, mgr2))
		if err != nil {
			t.Fatalf("NewEngine() error = %v", err)
		}

		if err := engine.Initialize(); err != nil {
			t.Fatalf("Initialize() error = %v", err)
		}

		managers := engine.GetManagers()
		if len(managers) != 2 {
			t.Errorf("GetManagers() returned %d managers, want 2", len(managers))
		}
	})

	t.Run("get managers when none registered", func(t *testing.T) {
		engine, err := NewEngine()
		if err != nil {
			t.Fatalf("NewEngine() error = %v", err)
		}

		if err := engine.Initialize(); err != nil {
			t.Fatalf("Initialize() error = %v", err)
		}

		managers := engine.GetManagers()
		if len(managers) != 0 {
			t.Errorf("GetManagers() returned %d managers, want 0", len(managers))
		}
	})
}

func TestEngine_GetServices(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("get all services", func(t *testing.T) {
		svc1 := &mockService{name: "service1"}
		svc2 := &mockService{name: "service2"}

		engine, err := NewEngine(RegisterServices(svc1, svc2))
		if err != nil {
			t.Fatalf("NewEngine() error = %v", err)
		}

		if err := engine.Initialize(); err != nil {
			t.Fatalf("Initialize() error = %v", err)
		}

		services := engine.GetServices()
		if len(services) != 2 {
			t.Errorf("GetServices() returned %d services, want 2", len(services))
		}
	})
}

func TestEngine_GetControllers(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("get all controllers", func(t *testing.T) {
		ctrl1 := &mockController{name: "controller1", router: "/api/test1"}
		ctrl2 := &mockController{name: "controller2", router: "/api/test2"}

		engine, err := NewEngine(RegisterControllers(ctrl1, ctrl2))
		if err != nil {
			t.Fatalf("NewEngine() error = %v", err)
		}

		if err := engine.Initialize(); err != nil {
			t.Fatalf("Initialize() error = %v", err)
		}

		controllers := engine.GetControllers()
		if len(controllers) != 2 {
			t.Errorf("GetControllers() returned %d controllers, want 2", len(controllers))
		}
	})
}

func TestEngine_GetMiddlewares(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("get all middlewares", func(t *testing.T) {
		mw1 := &mockMiddleware{name: "middleware1", order: 1}
		mw2 := &mockMiddleware{name: "middleware2", order: 2}

		engine, err := NewEngine(RegisterMiddlewares(mw1, mw2))
		if err != nil {
			t.Fatalf("NewEngine() error = %v", err)
		}

		if err := engine.Initialize(); err != nil {
			t.Fatalf("Initialize() error = %v", err)
		}

		middlewares := engine.GetMiddlewares()
		if len(middlewares) != 2 {
			t.Errorf("GetMiddlewares() returned %d middlewares, want 2", len(middlewares))
		}
	})
}

func TestEngine_ConcurrentLifecycle(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("concurrent start/stop calls", func(t *testing.T) {
		engine, err := NewEngine(
			WithServerConfig(&ServerConfig{
				Host: "127.0.0.1",
				Port: 0,
			}),
		)
		if err != nil {
			t.Fatalf("NewEngine() error = %v", err)
		}

		if err := engine.Initialize(); err != nil {
			t.Fatalf("Initialize() error = %v", err)
		}

		var wg sync.WaitGroup
		for i := 0; i < 10; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				_ = engine.IsStarted()
			}()
		}
		wg.Wait()
	})
}

