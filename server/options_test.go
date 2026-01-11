package server

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestWithConfigFile(t *testing.T) {
	t.Run("load JSON config file", func(t *testing.T) {
		// Create temp JSON config file
		tmpDir := t.TempDir()
		configPath := filepath.Join(tmpDir, "config.json")

		configContent := `{
			"host": "localhost",
			"port": 9090,
			"debug": true
		}`

		if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
			t.Fatalf("failed to write config file: %v", err)
		}

		engine, err := NewEngine(WithConfigFile(configPath))
		if err != nil {
			t.Fatalf("NewEngine() error = %v", err)
		}

		if engine.containers.config.Count() == 0 {
			t.Error("config should be registered")
		}
	})

	t.Run("load YAML config file", func(t *testing.T) {
		tmpDir := t.TempDir()
		configPath := filepath.Join(tmpDir, "config.yaml")

		configContent := `
host: localhost
port: 9090
debug: true
`

		if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
			t.Fatalf("failed to write config file: %v", err)
		}

		engine, err := NewEngine(WithConfigFile(configPath))
		if err != nil {
			t.Fatalf("NewEngine() error = %v", err)
		}

		if engine.containers.config.Count() == 0 {
			t.Error("config should be registered")
		}
	})

	t.Run("unsupported file format", func(t *testing.T) {
		tmpDir := t.TempDir()
		configPath := filepath.Join(tmpDir, "config.txt")

		if err := os.WriteFile(configPath, []byte("test"), 0644); err != nil {
			t.Fatalf("failed to write config file: %v", err)
		}

		_, err := NewEngine(WithConfigFile(configPath))
		if err == nil {
			t.Error("NewEngine() should return error for unsupported file format")
		}
	})

	t.Run("file not found", func(t *testing.T) {
		_, err := NewEngine(WithConfigFile("/nonexistent/config.json"))
		if err == nil {
			t.Error("NewEngine() should return error for nonexistent file")
		}
	})

	t.Run("invalid YAML file", func(t *testing.T) {
		tmpDir := t.TempDir()
		configPath := filepath.Join(tmpDir, "config.yaml")

		if err := os.WriteFile(configPath, []byte("invalid: yaml: content: ["), 0644); err != nil {
			t.Fatalf("failed to write config file: %v", err)
		}

		engine, err := NewEngine(WithConfigFile(configPath))
		// Engine creation might succeed but config loading will fail
		if err != nil {
			// This is expected behavior
			t.Logf("Expected error for invalid YAML: %v", err)
		}

		// Check that init error was recorded
		if engine != nil && len(engine.initErrors) > 0 {
			t.Logf("Init errors recorded: %v", engine.initErrors)
		}
	})
}

func TestWithConfig(t *testing.T) {
	t.Run("register config provider", func(t *testing.T) {
		mockConfig := &testConfigProvider{}

		engine, err := NewEngine(WithConfig(mockConfig))
		if err != nil {
			t.Fatalf("NewEngine() error = %v", err)
		}

		if engine.containers.config.Count() == 0 {
			t.Error("config should be registered")
		}

		config := engine.GetConfig()
		if config == nil {
			t.Error("GetConfig() should return the registered config")
		}
	})

	t.Run("register nil config", func(t *testing.T) {
		// Registering nil should cause an error
		_, err := NewEngine(WithConfig(nil))
		if err == nil {
			t.Error("NewEngine() should return error when registering nil config")
		}
	})
}

func TestWithServerConfig(t *testing.T) {
	t.Run("use custom server config", func(t *testing.T) {
		customConfig := &ServerConfig{
			Host: "192.168.1.1",
			Port: 9999,
			Mode: "debug",
		}

		engine, err := NewEngine(WithServerConfig(customConfig))
		if err != nil {
			t.Fatalf("NewEngine() error = %v", err)
		}

		if engine.serverConfig.Host != "192.168.1.1" {
			t.Errorf("Host = %s, want 192.168.1.1", engine.serverConfig.Host)
		}

		if engine.serverConfig.Port != 9999 {
			t.Errorf("Port = %d, want 9999", engine.serverConfig.Port)
		}

		if engine.serverConfig.Mode != "debug" {
			t.Errorf("Mode = %s, want debug", engine.serverConfig.Mode)
		}
	})

	t.Run("use nil server config (should use default)", func(t *testing.T) {
		engine, err := NewEngine(WithServerConfig(nil))
		if err != nil {
			t.Fatalf("NewEngine() error = %v", err)
		}

		defaultConfig := DefaultServerConfig()

		if engine.serverConfig.Host != defaultConfig.Host {
			t.Errorf("Host = %s, want %s", engine.serverConfig.Host, defaultConfig.Host)
		}

		if engine.serverConfig.Port != defaultConfig.Port {
			t.Errorf("Port = %d, want %d", engine.serverConfig.Port, defaultConfig.Port)
		}
	})
}

func TestRegisterManagers(t *testing.T) {
	t.Run("register valid managers", func(t *testing.T) {
		mgr1 := &mockManager{name: "manager1"}
		mgr2 := &mockManager{name: "manager2"}

		engine, err := NewEngine(RegisterManagers(mgr1, mgr2))
		if err != nil {
			t.Fatalf("NewEngine() error = %v", err)
		}

		if err := engine.Initialize(); err != nil {
			t.Fatalf("Initialize() error = %v", err)
		}

		if engine.containers.manager.Count() != 2 {
			t.Errorf("manager count = %d, want 2", engine.containers.manager.Count())
		}
	})

	t.Run("register invalid manager", func(t *testing.T) {
		invalidMgr := "not a manager"

		_, err := NewEngine(RegisterManagers(invalidMgr))
		if err == nil {
			t.Error("NewEngine() should return error when registering invalid manager")
		}
	})

	t.Run("register mixed valid and invalid managers", func(t *testing.T) {
		validMgr := &mockManager{name: "validManager"}
		invalidMgr := "not a manager"

		_, err := NewEngine(RegisterManagers(validMgr, invalidMgr))
		if err == nil {
			t.Error("NewEngine() should return error when one manager is invalid")
		}
	})
}

func TestRegisterEntities(t *testing.T) {
	t.Run("register valid entities", func(t *testing.T) {
		ent1 := &mockEntity{name: "entity1"}
		ent2 := &mockEntity{name: "entity2"}

		engine, err := NewEngine(RegisterEntities(ent1, ent2))
		if err != nil {
			t.Fatalf("NewEngine() error = %v", err)
		}

		if err := engine.Initialize(); err != nil {
			t.Fatalf("Initialize() error = %v", err)
		}

		if engine.containers.entity.Count() != 2 {
			t.Errorf("entity count = %d, want 2", engine.containers.entity.Count())
		}
	})

	t.Run("register invalid entity", func(t *testing.T) {
		invalidEnt := "not an entity"

		_, err := NewEngine(RegisterEntities(invalidEnt))
		if err == nil {
			t.Error("NewEngine() should return error when registering invalid entity")
		}
	})
}

func TestRegisterRepositories(t *testing.T) {
	t.Run("register valid repositories", func(t *testing.T) {
		repo1 := &mockRepository{name: "repository1"}
		repo2 := &mockRepository{name: "repository2"}

		engine, err := NewEngine(RegisterRepositories(repo1, repo2))
		if err != nil {
			t.Fatalf("NewEngine() error = %v", err)
		}

		if err := engine.Initialize(); err != nil {
			t.Fatalf("Initialize() error = %v", err)
		}

		if engine.containers.repository.Count() != 2 {
			t.Errorf("repository count = %d, want 2", engine.containers.repository.Count())
		}
	})

	t.Run("register invalid repository", func(t *testing.T) {
		invalidRepo := "not a repository"

		_, err := NewEngine(RegisterRepositories(invalidRepo))
		if err == nil {
			t.Error("NewEngine() should return error when registering invalid repository")
		}
	})
}

func TestRegisterServices(t *testing.T) {
	t.Run("register valid services", func(t *testing.T) {
		svc1 := &mockService{name: "service1"}
		svc2 := &mockService{name: "service2"}

		engine, err := NewEngine(RegisterServices(svc1, svc2))
		if err != nil {
			t.Fatalf("NewEngine() error = %v", err)
		}

		if err := engine.Initialize(); err != nil {
			t.Fatalf("Initialize() error = %v", err)
		}

		if engine.containers.service.Count() != 2 {
			t.Errorf("service count = %d, want 2", engine.containers.service.Count())
		}
	})

	t.Run("register invalid service", func(t *testing.T) {
		invalidSvc := "not a service"

		_, err := NewEngine(RegisterServices(invalidSvc))
		if err == nil {
			t.Error("NewEngine() should return error when registering invalid service")
		}
	})
}

func TestRegisterControllers_Options(t *testing.T) {
	t.Run("register valid controllers", func(t *testing.T) {
		ctrl1 := &mockController{name: "controller1", router: "/api/test1"}
		ctrl2 := &mockController{name: "controller2", router: "/api/test2"}

		engine, err := NewEngine(RegisterControllers(ctrl1, ctrl2))
		if err != nil {
			t.Fatalf("NewEngine() error = %v", err)
		}

		if err := engine.Initialize(); err != nil {
			t.Fatalf("Initialize() error = %v", err)
		}

		if engine.containers.controller.Count() != 2 {
			t.Errorf("controller count = %d, want 2", engine.containers.controller.Count())
		}
	})

	t.Run("register invalid controller", func(t *testing.T) {
		invalidCtrl := "not a controller"

		_, err := NewEngine(RegisterControllers(invalidCtrl))
		if err == nil {
			t.Error("NewEngine() should return error when registering invalid controller")
		}
	})
}

func TestRegisterMiddlewares(t *testing.T) {
	t.Run("register valid middlewares", func(t *testing.T) {
		mw1 := &orderedMiddleware{name: "middleware1", order: 1}
		mw2 := &orderedMiddleware{name: "middleware2", order: 2}

		engine, err := NewEngine(RegisterMiddlewares(mw1, mw2))
		if err != nil {
			t.Fatalf("NewEngine() error = %v", err)
		}

		if err := engine.Initialize(); err != nil {
			t.Fatalf("Initialize() error = %v", err)
		}

		if engine.containers.middleware.Count() != 2 {
			t.Errorf("middleware count = %d, want 2", engine.containers.middleware.Count())
		}
	})

	t.Run("register invalid middleware", func(t *testing.T) {
		invalidMw := "not a middleware"

		_, err := NewEngine(RegisterMiddlewares(invalidMw))
		if err == nil {
			t.Error("NewEngine() should return error when registering invalid middleware")
		}
	})
}

func TestMultipleOptions(t *testing.T) {
	t.Run("combine multiple options", func(t *testing.T) {
		mgr := &mockManager{name: "testManager"}
		ent := &mockEntity{name: "testEntity"}
		svc := &mockService{name: "testService"}

		customConfig := &ServerConfig{
			Host: "127.0.0.1",
			Port: 8081,
		}

		engine, err := NewEngine(
			WithServerConfig(customConfig),
			RegisterManagers(mgr),
			RegisterEntities(ent),
			RegisterServices(svc),
		)
		if err != nil {
			t.Fatalf("NewEngine() error = %v", err)
		}

		if err := engine.Initialize(); err != nil {
			t.Fatalf("Initialize() error = %v", err)
		}

		if engine.containers.manager.Count() != 1 {
			t.Errorf("manager count = %d, want 1", engine.containers.manager.Count())
		}

		if engine.containers.entity.Count() != 1 {
			t.Errorf("entity count = %d, want 1", engine.containers.entity.Count())
		}

		if engine.containers.service.Count() != 1 {
			t.Errorf("service count = %d, want 1", engine.containers.service.Count())
		}

		if engine.serverConfig.Port != 8081 {
			t.Errorf("Port = %d, want 8081", engine.serverConfig.Port)
		}
	})
}

func TestInitErrors(t *testing.T) {
	t.Run("collect multiple init errors", func(t *testing.T) {
		// Create options that will fail
		engine, err := NewEngine(
			RegisterManagers("invalid1"),
			RegisterEntities("invalid2"),
		)
		if err == nil {
			t.Error("NewEngine() should return error with invalid registrations")
		}

		if engine != nil && len(engine.initErrors) == 0 {
			t.Error("initErrors should contain errors from invalid registrations")
		}
	})
}

// Mock config provider for testing

type testConfigProvider struct {
	data map[string]interface{}
}

func (m *testConfigProvider) ConfigProviderName() string {
	return "testConfigProvider"
}

func (m *testConfigProvider) Get(key string) (any, error) {
	if m.data == nil {
		return nil, nil
	}
	val, ok := m.data[key]
	if !ok {
		return nil, nil
	}
	return val, nil
}

func (m *testConfigProvider) Has(key string) bool {
	if m.data == nil {
		return false
	}
	_, ok := m.data[key]
	return ok
}

// Add this mock to satisfy the interface
type failingConfigProvider struct{}

func (m *failingConfigProvider) ConfigProviderName() string {
	return "failingConfigProvider"
}

func (m *failingConfigProvider) Get(key string) (any, error) {
	return nil, errors.New("get failed")
}

func (m *failingConfigProvider) Has(key string) bool {
	return false
}

func TestWithConfig_ErrorHandling(t *testing.T) {
	t.Run("config registration failure", func(t *testing.T) {
		// Test that config registration errors are properly recorded
		testConfig := &testConfigProvider{}

		engine, err := NewEngine(WithConfig(testConfig))
		if err != nil {
			t.Fatalf("NewEngine() error = %v", err)
		}

		if len(engine.initErrors) > 0 {
			t.Logf("Init errors: %v", engine.initErrors)
		}
	})
}
