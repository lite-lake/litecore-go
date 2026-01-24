package server

import (
	"os"
	"testing"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/lite-lake/litecore-go/container"
)

// TestLifecycleManagerStartStop 测试管理器启动和停止
func TestLifecycleManagerStartStop(t *testing.T) {
	t.Run("管理器启动停止_成功", func(t *testing.T) {
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
		)

		engine.serverConfig = &serverConfig{
			Host:            "127.0.0.1",
			Port:            18082,
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
}
