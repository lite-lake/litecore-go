package server

import (
	"testing"
	"time"

	"github.com/lite-lake/litecore-go/container"
)

// TestStartManagers 测试管理器启动
func TestStartManagers(t *testing.T) {
	t.Run("空容器_启动成功", func(t *testing.T) {
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

		err := engine.startManagers()
		if err != nil {
			t.Fatalf("启动管理器失败: %v", err)
		}
	})
}

// TestStartRepositories 测试仓储启动
func TestStartRepositories(t *testing.T) {
	t.Run("空容器_启动成功", func(t *testing.T) {
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

		err := engine.startRepositories()
		if err != nil {
			t.Fatalf("启动仓储失败: %v", err)
		}
	})
}

// TestStartServices 测试服务启动
func TestStartServices(t *testing.T) {
	t.Run("空容器_启动成功", func(t *testing.T) {
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

		err := engine.startServices()
		if err != nil {
			t.Fatalf("启动服务失败: %v", err)
		}
	})
}

// TestStopManagers 测试管理器停止
func TestStopManagers(t *testing.T) {
	t.Run("空容器_停止成功", func(t *testing.T) {
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

		errors := engine.stopManagers()
		if len(errors) > 0 {
			t.Errorf("停止管理器时发生错误: %v", errors)
		}
	})
}

// TestStopServices 测试服务停止
func TestStopServices(t *testing.T) {
	t.Run("空容器_停止成功", func(t *testing.T) {
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

		errors := engine.stopServices()
		if len(errors) > 0 {
			t.Errorf("停止服务时发生错误: %v", errors)
		}
	})
}

// TestStopRepositories 测试仓储停止
func TestStopRepositories(t *testing.T) {
	t.Run("空容器_停止成功", func(t *testing.T) {
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

		errors := engine.stopRepositories()
		if len(errors) > 0 {
			t.Errorf("停止仓储时发生错误: %v", errors)
		}
	})
}

// TestStop 测试引擎停止
func TestStop(t *testing.T) {
	t.Run("停止未启动的引擎_无错误", func(t *testing.T) {
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

		err := engine.Stop()
		if err != nil {
			t.Errorf("停止未启动的引擎不应返回错误: %v", err)
		}
	})
}

// TestLifecycleTimeout 测试生命周期超时
func TestLifecycleTimeout(t *testing.T) {
	t.Run("关闭超时_在限制时间内完成", func(t *testing.T) {
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

		engine.started = true
		engine.shutdownTimeout = 1 * time.Second

		start := time.Now()
		err := engine.Stop()
		duration := time.Since(start)

		if err != nil {
			t.Errorf("期望停止成功, 实际错误: %v", err)
		}

		if duration > 2*time.Second {
			t.Errorf("关闭时间过长: %v", duration)
		}
	})
}

// BenchmarkStartManagers 基准测试管理器启动性能
func BenchmarkStartManagers(b *testing.B) {
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

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = engine.startManagers()
	}
}
