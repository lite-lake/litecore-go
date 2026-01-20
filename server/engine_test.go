package server

import (
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/gin-gonic/gin"

	"com.litelake.litecore/common"
	"com.litelake.litecore/container"
)

// TestNewEngine 测试创建引擎
func TestNewEngine(t *testing.T) {
	t.Run("创建新引擎_所有容器正确设置", func(t *testing.T) {
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

		if engine == nil {
			t.Fatal("期望 Engine 不为 nil")
		}

		if engine.Config != configContainer {
			t.Error("Config 容器未正确设置")
		}

		if engine.Entity != entityContainer {
			t.Error("Entity 容器未正确设置")
		}

		if engine.Manager != managerContainer {
			t.Error("Manager 容器未正确设置")
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

		if !strings.Contains(err.Error(), "already started") {
			t.Errorf("期望错误包含 'already started', 实际: %v", err)
		}
	})
}

// TestEngineRun 测试完整运行流程
func TestEngineRun(t *testing.T) {
	t.Run("完整流程_初始化启动", func(t *testing.T) {
		gin.SetMode(gin.TestMode)

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

		engine.serverConfig = &serverConfig{
			Host:            "127.0.0.1",
			Port:            18083,
			Mode:            "test",
			ReadTimeout:     1 * time.Second,
			WriteTimeout:    1 * time.Second,
			IdleTimeout:     10 * time.Second,
			ShutdownTimeout: 2 * time.Second,
		}

		err := engine.Initialize()
		if err != nil {
			t.Errorf("期望 Initialize 成功, 实际错误: %v", err)
		}

		err = engine.Start()
		if err != nil {
			t.Errorf("期望 Start 成功, 实际错误: %v", err)
		}

		go func() {
			time.Sleep(100 * time.Millisecond)
			_ = engine.Stop()
		}()
	})
}

// TestConcurrentOperations 测试并发操作
func TestConcurrentOperations(t *testing.T) {
	t.Run("并发初始化_无竞争", func(t *testing.T) {
		gin.SetMode(gin.TestMode)

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

		engine.serverConfig = &serverConfig{
			Host:            "127.0.0.1",
			Port:            18084,
			Mode:            "test",
			ReadTimeout:     1 * time.Second,
			WriteTimeout:    1 * time.Second,
			IdleTimeout:     10 * time.Second,
			ShutdownTimeout: 2 * time.Second,
		}

		var wg sync.WaitGroup
		errChan := make(chan error, 10)

		for i := 0; i < 10; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				if err := engine.Initialize(); err != nil {
					errChan <- err
				}
			}()
		}

		wg.Wait()
		close(errChan)

		for err := range errChan {
			if err != nil {
				t.Errorf("并发初始化错误: %v", err)
			}
		}
	})
}

// TestInitializeGinEngineServices 测试 Gin 引擎服务初始化
func TestInitializeGinEngineServices(t *testing.T) {
	t.Run("空容器_初始化成功", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
		ginEngine := gin.New()

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

		engine.ginEngine = ginEngine
		engine.initializeGinEngineServices()
	})
}

// TestHTTPTimedout 测试 HTTP 超时配置
func TestHTTPTimedout(t *testing.T) {
	t.Run("HTTP超时_配置正确应用", func(t *testing.T) {
		gin.SetMode(gin.TestMode)

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

		customConfig := &serverConfig{
			Host:            "127.0.0.1",
			Port:            18085,
			Mode:            "test",
			ReadTimeout:     2 * time.Second,
			WriteTimeout:    3 * time.Second,
			IdleTimeout:     5 * time.Second,
			ShutdownTimeout: 10 * time.Second,
		}

		engine.serverConfig = customConfig
		engine.shutdownTimeout = customConfig.ShutdownTimeout

		if err := engine.Initialize(); err != nil {
			t.Fatalf("初始化失败: %v", err)
		}

		if engine.httpServer.ReadTimeout != customConfig.ReadTimeout {
			t.Errorf("期望 ReadTimeout = %v, 实际 = %v", customConfig.ReadTimeout, engine.httpServer.ReadTimeout)
		}

		if engine.httpServer.WriteTimeout != customConfig.WriteTimeout {
			t.Errorf("期望 WriteTimeout = %v, 实际 = %v", customConfig.WriteTimeout, engine.httpServer.WriteTimeout)
		}

		if engine.httpServer.IdleTimeout != customConfig.IdleTimeout {
			t.Errorf("期望 IdleTimeout = %v, 实际 = %v", customConfig.IdleTimeout, engine.httpServer.IdleTimeout)
		}
	})
}

// BenchmarkEngineCreation 基准测试引擎创建性能
func BenchmarkEngineCreation(b *testing.B) {
	configContainer := container.NewConfigContainer()
	entityContainer := container.NewEntityContainer()
	managerContainer := container.NewManagerContainer(configContainer)
	repositoryContainer := container.NewRepositoryContainer(configContainer, managerContainer, entityContainer)
	serviceContainer := container.NewServiceContainer(configContainer, managerContainer, repositoryContainer)
	controllerContainer := container.NewControllerContainer(configContainer, managerContainer, serviceContainer)
	middlewareContainer := container.NewMiddlewareContainer(configContainer, managerContainer, serviceContainer)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = NewEngine(
			configContainer,
			entityContainer,
			managerContainer,
			repositoryContainer,
			serviceContainer,
			controllerContainer,
			middlewareContainer,
		)
	}
}

// BenchmarkEngineInitialization 基准测试引擎初始化性能
func BenchmarkEngineInitialization(b *testing.B) {
	gin.SetMode(gin.TestMode)

	configContainer := container.NewConfigContainer()
	entityContainer := container.NewEntityContainer()
	managerContainer := container.NewManagerContainer(configContainer)
	repositoryContainer := container.NewRepositoryContainer(configContainer, managerContainer, entityContainer)
	serviceContainer := container.NewServiceContainer(configContainer, managerContainer, repositoryContainer)
	controllerContainer := container.NewControllerContainer(configContainer, managerContainer, serviceContainer)
	middlewareContainer := container.NewMiddlewareContainer(configContainer, managerContainer, serviceContainer)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		engine := NewEngine(
			configContainer,
			entityContainer,
			managerContainer,
			repositoryContainer,
			serviceContainer,
			controllerContainer,
			middlewareContainer,
		)
		_ = engine.Initialize()
	}
}

// testGinService 测试用的 Gin 服务
type testGinService struct {
	name      string
	ginEngine *gin.Engine
}

func (s *testGinService) ServiceName() string {
	return s.name
}

func (s *testGinService) OnStart() error {
	return nil
}

func (s *testGinService) OnStop() error {
	return nil
}

func (s *testGinService) SetGinEngine(eng *gin.Engine) {
	s.ginEngine = eng
}

var _ common.IBaseService = (*testGinService)(nil)
