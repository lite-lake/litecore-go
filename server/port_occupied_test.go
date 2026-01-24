package server

import (
	"fmt"
	"net"
	"os"
	"testing"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/lite-lake/litecore-go/container"
)

// TestEnginePortOccupied 测试端口占用时的错误处理
func TestEnginePortOccupied(t *testing.T) {
	port := 18083

	listener, err := net.Listen("tcp", fmt.Sprintf("127.0.0.1:%d", port))
	if err != nil {
		t.Fatalf("无法监听端口 %d: %v", port, err)
	}
	defer listener.Close()

	configFile := `telemetry:
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
	configPath := t.TempDir() + "/test-port-occupied-config.yaml"
	if err := os.WriteFile(configPath, []byte(configFile), 0644); err != nil {
		t.Fatalf("创建配置文件失败: %v", err)
	}

	builtinConfig := &BuiltinConfig{
		Driver:   "yaml",
		FilePath: configPath,
	}

	entityContainer := container.NewEntityContainer()
	repositoryContainer := container.NewRepositoryContainer(entityContainer)
	serviceContainer := container.NewServiceContainer(repositoryContainer)
	controllerContainer := container.NewControllerContainer(serviceContainer)
	middlewareContainer := container.NewMiddlewareContainer(serviceContainer)

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
		Port:            port,
		Mode:            "test",
		ReadTimeout:     1 * time.Second,
		WriteTimeout:    1 * time.Second,
		IdleTimeout:     10 * time.Second,
		ShutdownTimeout: 2 * time.Second,
	}

	if err := engine.Initialize(); err != nil {
		t.Fatalf("初始化失败: %v", err)
	}

	startTime := time.Now()
	err = engine.Start()
	duration := time.Since(startTime)

	if err == nil {
		t.Fatal("启动应该失败但成功了")
	}

	if duration > 200*time.Millisecond {
		t.Fatalf("启动耗时 %v，超过 200ms 预期", duration)
	}

	if duration < 50*time.Millisecond {
		t.Logf("警告: 启动耗时 %v，小于 50ms 预期（但可以接受）", duration)
	}

	t.Logf("启动在 %v 内失败", duration)

	if !containsSubstring(err.Error(), "address already in use") {
		t.Errorf("错误信息不正确: %v", err)
	}

	t.Logf("错误信息正确: %v", err)
}

func TestEnginePortOccupiedIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("跳过集成测试")
	}

	gin.SetMode(gin.TestMode)

	port := 18084

	listener, err := net.Listen("tcp", fmt.Sprintf("127.0.0.1:%d", port))
	if err != nil {
		t.Fatalf("无法监听端口 %d: %v", port, err)
	}
	defer listener.Close()

	configFile := fmt.Sprintf(`telemetry:
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
`)
	configPath := t.TempDir() + "/test-port-occupied-integration-config.yaml"
	if err := os.WriteFile(configPath, []byte(configFile), 0644); err != nil {
		t.Fatalf("创建配置文件失败: %v", err)
	}

	builtinConfig := &BuiltinConfig{
		Driver:   "yaml",
		FilePath: configPath,
	}

	entityContainer := container.NewEntityContainer()
	repositoryContainer := container.NewRepositoryContainer(entityContainer)
	serviceContainer := container.NewServiceContainer(repositoryContainer)
	controllerContainer := container.NewControllerContainer(serviceContainer)
	middlewareContainer := container.NewMiddlewareContainer(serviceContainer)

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
		Port:            port,
		Mode:            "test",
		ReadTimeout:     1 * time.Second,
		WriteTimeout:    1 * time.Second,
		IdleTimeout:     10 * time.Second,
		ShutdownTimeout: 2 * time.Second,
	}

	if err := engine.Initialize(); err != nil {
		t.Fatalf("初始化失败: %v", err)
	}

	startErr := engine.Start()
	if startErr == nil {
		t.Fatal("启动应该失败但成功了")
	}

	t.Logf("✅ 端口占用错误检测成功: %v", startErr)
}

func TestEngineNormalStartup(t *testing.T) {
	gin.SetMode(gin.TestMode)

	port := 18085

	configFile := `telemetry:
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
	configPath := t.TempDir() + "/test-normal-startup-config.yaml"
	if err := os.WriteFile(configPath, []byte(configFile), 0644); err != nil {
		t.Fatalf("创建配置文件失败: %v", err)
	}

	builtinConfig := &BuiltinConfig{
		Driver:   "yaml",
		FilePath: configPath,
	}

	entityContainer := container.NewEntityContainer()
	repositoryContainer := container.NewRepositoryContainer(entityContainer)
	serviceContainer := container.NewServiceContainer(repositoryContainer)
	controllerContainer := container.NewControllerContainer(serviceContainer)
	middlewareContainer := container.NewMiddlewareContainer(serviceContainer)

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
		Port:            port,
		Mode:            "test",
		ReadTimeout:     1 * time.Second,
		WriteTimeout:    1 * time.Second,
		IdleTimeout:     10 * time.Second,
		ShutdownTimeout: 2 * time.Second,
	}

	if err := engine.Initialize(); err != nil {
		t.Fatalf("初始化失败: %v", err)
	}

	startTime := time.Now()
	if err := engine.Start(); err != nil {
		t.Fatalf("启动失败: %v", err)
	}
	duration := time.Since(startTime)

	if duration > 200*time.Millisecond {
		t.Logf("警告: 正常启动耗时 %v，超过 200ms 预期", duration)
	}

	t.Logf("✅ 正常启动成功，耗时: %v", duration)

	_ = engine.Stop()
}

func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
