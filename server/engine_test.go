package server

import (
	"os"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/lite-lake/litecore-go/container"
)

// TestNewEngine 测试创建引擎
func TestNewEngine(t *testing.T) {
	t.Run("创建新引擎_所有容器正确设置", func(t *testing.T) {
		entityContainer := container.NewEntityContainer()
		repositoryContainer := container.NewRepositoryContainer(entityContainer)
		serviceContainer := container.NewServiceContainer(repositoryContainer)
		controllerContainer := container.NewControllerContainer(serviceContainer)
		middlewareContainer := container.NewMiddlewareContainer(serviceContainer)

		builtinConfig := &BuiltinConfig{
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
			nil,
			nil,
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

		configFile := `server:
  port: 8080
telemetry:
  driver: none
logger:
  driver: none
database:
  driver: none
cache:
  driver: none
lock:
  driver: memory
  memory_config:
    ttl: 60
limiter:
  driver: memory
  memory_config:
    max_requests: 1000
    window: 60
mq:
  driver: memory
  memory_config:
    max_queue_size: 10000
    channel_buffer: 100
scheduler:
  driver: cron
  cron_config:
    validate_on_startup: true
`
		configPath := t.TempDir() + "/test-config.yaml"
		if err := os.WriteFile(configPath, []byte(configFile), 0644); err != nil {
			t.Fatalf("创建配置文件失败: %v", err)
		}

		builtinConfig := &BuiltinConfig{
			Driver:   "yaml",
			FilePath: configPath,
		}

		engine := NewEngine(
			builtinConfig,
			entityContainer,
			repositoryContainer,
			serviceContainer,
			controllerContainer,
			middlewareContainer,
			nil,
			nil,
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

		_ = engine.Stop()
	})

	t.Run("重复启动_返回错误", func(t *testing.T) {
		gin.SetMode(gin.TestMode)

		entityContainer := container.NewEntityContainer()
		repositoryContainer := container.NewRepositoryContainer(entityContainer)
		serviceContainer := container.NewServiceContainer(repositoryContainer)
		controllerContainer := container.NewControllerContainer(serviceContainer)
		middlewareContainer := container.NewMiddlewareContainer(serviceContainer)

		configFile := `server:
  port: 8080
telemetry:
  driver: none
logger:
  driver: none
database:
  driver: none
cache:
  driver: none
lock:
  driver: memory
  memory_config:
    ttl: 60
limiter:
  driver: memory
  memory_config:
    max_requests: 1000
    window: 60
mq:
  driver: memory
  memory_config:
    max_queue_size: 10000
    channel_buffer: 100
scheduler:
  driver: cron
  cron_config:
    validate_on_startup: true
`
		configPath := t.TempDir() + "/test-config.yaml"
		if err := os.WriteFile(configPath, []byte(configFile), 0644); err != nil {
			t.Fatalf("创建配置文件失败: %v", err)
		}

		builtinConfig := &BuiltinConfig{
			Driver:   "yaml",
			FilePath: configPath,
		}

		engine := NewEngine(
			builtinConfig,
			entityContainer,
			repositoryContainer,
			serviceContainer,
			controllerContainer,
			middlewareContainer,
			nil,
			nil,
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

		configFile := `server:
  port: 8080
telemetry:
  driver: none
logger:
  driver: none
database:
  driver: none
cache:
  driver: none
lock:
  driver: memory
  memory_config:
    ttl: 60
limiter:
  driver: memory
  memory_config:
    max_requests: 1000
    window: 60
mq:
  driver: memory
  memory_config:
    max_queue_size: 10000
    channel_buffer: 100
scheduler:
  driver: cron
  cron_config:
    validate_on_startup: true
`
		configPath := t.TempDir() + "/test-config.yaml"
		if err := os.WriteFile(configPath, []byte(configFile), 0644); err != nil {
			t.Fatalf("创建配置文件失败: %v", err)
		}

		builtinConfig := &BuiltinConfig{
			Driver:   "yaml",
			FilePath: configPath,
		}

		engine := NewEngine(
			builtinConfig,
			entityContainer,
			repositoryContainer,
			serviceContainer,
			controllerContainer,
			middlewareContainer,
			nil,
			nil,
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

		_ = engine.Stop()
	})

	t.Run("重复启动_返回错误", func(t *testing.T) {
		gin.SetMode(gin.TestMode)

		entityContainer := container.NewEntityContainer()
		repositoryContainer := container.NewRepositoryContainer(entityContainer)
		serviceContainer := container.NewServiceContainer(repositoryContainer)
		controllerContainer := container.NewControllerContainer(serviceContainer)
		middlewareContainer := container.NewMiddlewareContainer(serviceContainer)

		configFile := `server:
  port: 8080
telemetry:
  driver: none
logger:
  driver: none
database:
  driver: none
cache:
  driver: none
lock:
  driver: memory
  memory_config:
    ttl: 60
limiter:
  driver: memory
  memory_config:
    max_requests: 1000
    window: 60
mq:
  driver: memory
  memory_config:
    max_queue_size: 10000
    channel_buffer: 100
scheduler:
  driver: cron
  cron_config:
    validate_on_startup: true
`
		configPath := t.TempDir() + "/test-config.yaml"
		if err := os.WriteFile(configPath, []byte(configFile), 0644); err != nil {
			t.Fatalf("创建配置文件失败: %v", err)
		}

		builtinConfig := &BuiltinConfig{
			Driver:   "yaml",
			FilePath: configPath,
		}

		engine := NewEngine(
			builtinConfig,
			entityContainer,
			repositoryContainer,
			serviceContainer,
			controllerContainer,
			middlewareContainer,
			nil,
			nil,
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

		err := engine.Start()
		if err == nil {
			t.Error("期望重复启动返回错误")
		}

		if err != nil && !strings.Contains(err.Error(), "already started") {
			t.Errorf("期望错误包含 'already started', 实际: %v", err)
		}

		_ = engine.Stop()
	})
}

// TestEngineLoadServerConfig 测试从配置文件加载 server 配置
func TestEngineLoadServerConfig(t *testing.T) {
	t.Run("配置文件覆盖默认值", func(t *testing.T) {
		gin.SetMode(gin.TestMode)

		entityContainer := container.NewEntityContainer()
		repositoryContainer := container.NewRepositoryContainer(entityContainer)
		serviceContainer := container.NewServiceContainer(repositoryContainer)
		controllerContainer := container.NewControllerContainer(serviceContainer)
		middlewareContainer := container.NewMiddlewareContainer(serviceContainer)

		configFile := `server:
  host: "127.0.0.1"
  port: 9999
  mode: "test"
  read_timeout: "5s"
  write_timeout: "8s"
  idle_timeout: "30s"
  shutdown_timeout: "10s"
  startup_log:
    enabled: false
    async: false
    buffer: 50
telemetry:
  driver: none
logger:
  driver: none
database:
  driver: none
cache:
  driver: none
lock:
  driver: memory
  memory_config:
    ttl: 60
limiter:
  driver: memory
  memory_config:
    max_requests: 1000
    window: 60
mq:
  driver: memory
  memory_config:
    max_queue_size: 10000
    channel_buffer: 100
scheduler:
  driver: cron
  cron_config:
    validate_on_startup: true
`
		configPath := t.TempDir() + "/test-server-config.yaml"
		if err := os.WriteFile(configPath, []byte(configFile), 0644); err != nil {
			t.Fatalf("创建配置文件失败: %v", err)
		}

		builtinConfig := &BuiltinConfig{
			Driver:   "yaml",
			FilePath: configPath,
		}

		engine := NewEngine(
			builtinConfig,
			entityContainer,
			repositoryContainer,
			serviceContainer,
			controllerContainer,
			middlewareContainer,
			nil,
			nil,
		)

		if err := engine.Initialize(); err != nil {
			t.Fatalf("初始化失败: %v", err)
		}

		if engine.serverConfig.Host != "127.0.0.1" {
			t.Errorf("期望 host 为 '127.0.0.1', 实际: %s", engine.serverConfig.Host)
		}

		if engine.serverConfig.Port != 9999 {
			t.Errorf("期望 port 为 9999, 实际: %d", engine.serverConfig.Port)
		}

		if engine.serverConfig.Mode != "test" {
			t.Errorf("期望 mode 为 'test', 实际: %s", engine.serverConfig.Mode)
		}

		if engine.serverConfig.ReadTimeout != 5*time.Second {
			t.Errorf("期望 read_timeout 为 5s, 实际: %v", engine.serverConfig.ReadTimeout)
		}

		if engine.serverConfig.WriteTimeout != 8*time.Second {
			t.Errorf("期望 write_timeout 为 8s, 实际: %v", engine.serverConfig.WriteTimeout)
		}

		if engine.serverConfig.IdleTimeout != 30*time.Second {
			t.Errorf("期望 idle_timeout 为 30s, 实际: %v", engine.serverConfig.IdleTimeout)
		}

		if engine.serverConfig.ShutdownTimeout != 10*time.Second {
			t.Errorf("期望 shutdown_timeout 为 10s, 实际: %v", engine.serverConfig.ShutdownTimeout)
		}

		if engine.serverConfig.StartupLog == nil {
			t.Fatal("期望 startup_log 配置被加载")
		}

		if engine.serverConfig.StartupLog.Enabled != false {
			t.Errorf("期望 startup_log.enabled 为 false, 实际: %v", engine.serverConfig.StartupLog.Enabled)
		}

		if engine.serverConfig.StartupLog.Async != false {
			t.Errorf("期望 startup_log.async 为 false, 实际: %v", engine.serverConfig.StartupLog.Async)
		}

		if engine.serverConfig.StartupLog.Buffer != 50 {
			t.Errorf("期望 startup_log.buffer 为 50, 实际: %d", engine.serverConfig.StartupLog.Buffer)
		}

		_ = engine.Stop()
	})
}
