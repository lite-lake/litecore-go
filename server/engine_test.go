package server

import (
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/lite-lake/litecore-go/container"
	"github.com/lite-lake/litecore-go/server/builtin"
)

// TestNewEngine 测试创建引擎
func TestNewEngine(t *testing.T) {
	t.Run("创建新引擎_所有容器正确设置", func(t *testing.T) {
		entityContainer := container.NewEntityContainer()
		repositoryContainer := container.NewRepositoryContainer(entityContainer)
		serviceContainer := container.NewServiceContainer(repositoryContainer)
		controllerContainer := container.NewControllerContainer(serviceContainer)
		middlewareContainer := container.NewMiddlewareContainer(serviceContainer)

		builtinConfig := &builtin.Config{
			Driver:   "yaml",
			FilePath: "test.yaml",
		}

		engine := NewEngine(
			builtinConfig,
			entityContainer,
			repositoryContainer,
			serviceContainer,
			controllerContainer,
			middlewareContainer,
		)

		if engine == nil {
			t.Fatal("期望 Engine 不为 nil")
		}

		if engine.Entity != entityContainer {
			t.Error("Entity 容器未正确设置")
		}

		if engine.Repository != repositoryContainer {
			t.Error("Repository 容器未正确设置")
		}

		if engine.Service != serviceContainer {
			t.Error("Service 容器未正确设置")
		}

		if engine.Controller != controllerContainer {
			t.Error("Controller 容器未正确设置")
		}

		if engine.Middleware != middlewareContainer {
			t.Error("Middleware 容器未正确设置")
		}

		if engine.started {
			t.Error("新引擎应该处于未启动状态")
		}
	})
}

// TestEngineInitialize 测试引擎初始化
func TestEngineInitialize(t *testing.T) {
	t.Run("初始化_成功", func(t *testing.T) {
		gin.SetMode(gin.TestMode)

		entityContainer := container.NewEntityContainer()
		repositoryContainer := container.NewRepositoryContainer(entityContainer)
		serviceContainer := container.NewServiceContainer(repositoryContainer)
		controllerContainer := container.NewControllerContainer(serviceContainer)
		middlewareContainer := container.NewMiddlewareContainer(serviceContainer)

		builtinConfig := &builtin.Config{
			Driver:   "yaml",
			FilePath: "/tmp/test-config.yaml",
		}

		engine := NewEngine(
			builtinConfig,
			entityContainer,
			repositoryContainer,
			serviceContainer,
			controllerContainer,
			middlewareContainer,
		)

		if err := engine.Initialize(); err != nil {
			t.Fatalf("初始化失败: %v", err)
		}

		if engine.ginEngine == nil {
			t.Error("Gin 引擎未创建")
		}

		if engine.httpServer == nil {
			t.Error("HTTP 服务器未创建")
		}
	})
}

// TestEngineStart 测试引擎启动
func TestEngineStart(t *testing.T) {
	t.Run("启动_成功", func(t *testing.T) {
		gin.SetMode(gin.TestMode)

		entityContainer := container.NewEntityContainer()
		repositoryContainer := container.NewRepositoryContainer(entityContainer)
		serviceContainer := container.NewServiceContainer(repositoryContainer)
		controllerContainer := container.NewControllerContainer(serviceContainer)
		middlewareContainer := container.NewMiddlewareContainer(serviceContainer)

		builtinConfig := &builtin.Config{
			Driver:   "yaml",
			FilePath: "/tmp/test-config.yaml",
		}

		engine := NewEngine(
			builtinConfig,
			entityContainer,
			repositoryContainer,
			serviceContainer,
			controllerContainer,
			middlewareContainer,
		)

		engine.serverConfig = &serverConfig{
			Host:            "127.0.0.1",
			Port:            18080,
			Mode:            "test",
			ReadTimeout:     1 * time.Second,
			WriteTimeout:    1 * time.Second,
			IdleTimeout:     10 * time.Second,
			ShutdownTimeout: 2 * time.Second,
		}

		if err := engine.Initialize(); err != nil {
			t.Fatalf("初始化失败: %v", err)
		}

		if err := engine.Start(); err != nil {
			t.Fatalf("启动失败: %v", err)
		}

		if !engine.started {
			t.Error("引擎应该处于已启动状态")
		}

		go func() {
			time.Sleep(100 * time.Millisecond)
			_ = engine.Stop()
		}()
	})

	t.Run("重复启动_返回错误", func(t *testing.T) {
		gin.SetMode(gin.TestMode)

		entityContainer := container.NewEntityContainer()
		repositoryContainer := container.NewRepositoryContainer(entityContainer)
		serviceContainer := container.NewServiceContainer(repositoryContainer)
		controllerContainer := container.NewControllerContainer(serviceContainer)
		middlewareContainer := container.NewMiddlewareContainer(serviceContainer)

		builtinConfig := &builtin.Config{
			Driver:   "yaml",
			FilePath: "/tmp/test-config.yaml",
		}

		engine := NewEngine(
			builtinConfig,
			entityContainer,
			repositoryContainer,
			serviceContainer,
			controllerContainer,
			middlewareContainer,
		)

		engine.serverConfig = &serverConfig{
			Host:            "127.0.0.1",
			Port:            18081,
			Mode:            "test",
			ReadTimeout:     1 * time.Second,
			WriteTimeout:    1 * time.Second,
			IdleTimeout:     10 * time.Second,
			ShutdownTimeout: 2 * time.Second,
		}

		if err := engine.Initialize(); err != nil {
			t.Fatalf("初始化失败: %v", err)
		}

		if err := engine.Start(); err != nil {
			t.Fatalf("第一次启动失败: %v", err)
		}

		go func() {
			time.Sleep(100 * time.Millisecond)
			_ = engine.Stop()
		}()

		err := engine.Start()
		if err == nil {
			t.Error("期望重复启动返回错误")
		}

		if err != nil && !strings.Contains(err.Error(), "already started") {
			t.Errorf("期望错误包含 'already started', 实际: %v", err)
		}
	})
}
