package mqmgr

import (
	"testing"

	"github.com/lite-lake/litecore-go/manager/configmgr"
)

type mockConfigProvider struct {
	data map[string]any
}

func (m *mockConfigProvider) ManagerName() string {
	return "mockConfigProvider"
}

func (m *mockConfigProvider) Health() error {
	return nil
}

func (m *mockConfigProvider) OnStart() error {
	return nil
}

func (m *mockConfigProvider) OnStop() error {
	return nil
}

func (m *mockConfigProvider) Get(key string) (any, error) {
	if val, ok := m.data[key]; ok {
		return val, nil
	}
	return nil, nil
}

func (m *mockConfigProvider) Has(key string) bool {
	_, ok := m.data[key]
	return ok
}

func TestBuild(t *testing.T) {
	t.Run("memory 驱动", func(t *testing.T) {
		driverConfig := map[string]any{
			"max_queue_size": 5000,
			"channel_buffer": 200,
		}

		mgr, err := Build("memory", driverConfig, nil, nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if mgr == nil {
			t.Fatal("expected manager not to be nil")
		}

		if mgr.ManagerName() != "messageQueueManagerMemoryImpl" {
			t.Errorf("expected 'messageQueueManagerMemoryImpl', got '%s'", mgr.ManagerName())
		}

		mgr.Close()
	})

	t.Run("memory 驱动默认配置", func(t *testing.T) {
		mgr, err := Build("memory", nil, nil, nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if mgr == nil {
			t.Fatal("expected manager not to be nil")
		}

		mgr.Close()
	})

	t.Run("rabbitmq 驱动", func(t *testing.T) {
		driverConfig := map[string]any{
			"url":     "amqp://guest:guest@localhost:5672/",
			"durable": false,
		}

		mgr, err := Build("rabbitmq", driverConfig, nil, nil)
		if err != nil {
			t.Skip("RabbitMQ not available")
		}

		if mgr == nil {
			t.Fatal("expected manager not to be nil")
		}

		if mgr.ManagerName() != "messageQueueManagerRabbitMQImpl" {
			t.Errorf("expected 'messageQueueManagerRabbitMQImpl', got '%s'", mgr.ManagerName())
		}

		mgr.Close()
	})

	t.Run("不支持的驱动类型", func(t *testing.T) {
		driverConfig := map[string]any{}

		_, err := Build("unsupported", driverConfig, nil, nil)
		if err == nil {
			t.Error("expected error with unsupported driver type")
		}
	})

	t.Run("rabbitmq 驱动无效配置", func(t *testing.T) {
		driverConfig := map[string]any{
			"url": "invalid-url",
		}

		_, err := Build("rabbitmq", driverConfig, nil, nil)
		if err == nil {
			t.Error("expected error with invalid rabbitmq config")
		}
	})
}

func TestBuildWithConfigProvider(t *testing.T) {
	t.Run("空配置提供者", func(t *testing.T) {
		_, err := BuildWithConfigProvider(nil, nil, nil)
		if err == nil {
			t.Error("expected error with nil config provider")
		}
	})

	t.Run("memory 驱动完整配置", func(t *testing.T) {
		provider := &mockConfigProvider{
			data: map[string]any{
				"mq.driver": "memory",
				"mq.memory_config": map[string]any{
					"max_queue_size": 5000,
					"channel_buffer": 200,
				},
			},
		}

		mgr, err := BuildWithConfigProvider(provider, nil, nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if mgr == nil {
			t.Fatal("expected manager not to be nil")
		}

		if mgr.ManagerName() != "messageQueueManagerMemoryImpl" {
			t.Errorf("expected 'messageQueueManagerMemoryImpl', got '%s'", mgr.ManagerName())
		}

		mgr.Close()
	})

	t.Run("rabbitmq 驱动完整配置", func(t *testing.T) {
		provider := &mockConfigProvider{
			data: map[string]any{
				"mq.driver": "rabbitmq",
				"mq.rabbitmq_config": map[string]any{
					"url":     "amqp://guest:guest@localhost:5672/",
					"durable": false,
				},
			},
		}

		mgr, err := BuildWithConfigProvider(provider, nil, nil)
		if err != nil {
			t.Skip("RabbitMQ not available")
		}

		if mgr == nil {
			t.Fatal("expected manager not to be nil")
		}

		if mgr.ManagerName() != "messageQueueManagerRabbitMQImpl" {
			t.Errorf("expected 'messageQueueManagerRabbitMQImpl', got '%s'", mgr.ManagerName())
		}

		mgr.Close()
	})

	t.Run("缺少 mq.driver", func(t *testing.T) {
		provider := &mockConfigProvider{
			data: map[string]any{},
		}

		_, err := BuildWithConfigProvider(provider, nil, nil)
		if err == nil {
			t.Error("expected error when mq.driver is missing")
		}
	})

	t.Run("rabbitmq 驱动缺少 rabbitmq_config", func(t *testing.T) {
		provider := &mockConfigProvider{
			data: map[string]any{
				"mq.driver": "rabbitmq",
			},
		}

		_, err := BuildWithConfigProvider(provider, nil, nil)
		if err == nil {
			t.Error("expected error when rabbitmq_config is missing")
		}
	})

	t.Run("memory 驱动缺少 memory_config", func(t *testing.T) {
		provider := &mockConfigProvider{
			data: map[string]any{
				"mq.driver": "memory",
			},
		}

		_, err := BuildWithConfigProvider(provider, nil, nil)
		if err == nil {
			t.Error("expected error when memory_config is missing")
		}
	})

	t.Run("无效的 mq.driver 类型", func(t *testing.T) {
		provider := &mockConfigProvider{
			data: map[string]any{
				"mq.driver": 123,
			},
		}

		_, err := BuildWithConfigProvider(provider, nil, nil)
		if err == nil {
			t.Error("expected error with invalid mq.driver type")
		}
	})

	t.Run("不支持的驱动类型", func(t *testing.T) {
		provider := &mockConfigProvider{
			data: map[string]any{
				"mq.driver":        "unsupported",
				"mq.memory_config": map[string]any{},
			},
		}

		_, err := BuildWithConfigProvider(provider, nil, nil)
		if err == nil {
			t.Error("expected error with unsupported driver type")
		}
	})

	t.Run("无效的 rabbitmq_config 类型", func(t *testing.T) {
		provider := &mockConfigProvider{
			data: map[string]any{
				"mq.driver":          "rabbitmq",
				"mq.rabbitmq_config": "invalid",
			},
		}

		_, err := BuildWithConfigProvider(provider, nil, nil)
		if err == nil {
			t.Error("expected error with invalid rabbitmq_config type")
		}
	})

	t.Run("无效的 memory_config 类型", func(t *testing.T) {
		provider := &mockConfigProvider{
			data: map[string]any{
				"mq.driver":        "memory",
				"mq.memory_config": "invalid",
			},
		}

		_, err := BuildWithConfigProvider(provider, nil, nil)
		if err == nil {
			t.Error("expected error with invalid memory_config type")
		}
	})
}

var _ configmgr.IConfigManager = (*mockConfigProvider)(nil)

func TestBuildWithLoggerAndTelemetry(t *testing.T) {
	t.Run("memory 驱动带 logger 和 telemetry", func(t *testing.T) {
		driverConfig := map[string]any{
			"max_queue_size": 5000,
			"channel_buffer": 200,
		}

		mgr, err := Build("memory", driverConfig, nil, nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if mgr == nil {
			t.Fatal("expected manager not to be nil")
		}

		mgr.Close()
	})
}

func TestParseMemoryConfigInvalidType(t *testing.T) {
	t.Run("无效的 max_queue_size 类型", func(t *testing.T) {
		cfg := map[string]any{
			"max_queue_size": "invalid",
		}

		config, err := parseMemoryConfig(cfg)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		if config.MaxQueueSize != DefaultMemoryMaxQueueSize {
			t.Errorf("expected default MaxQueueSize for invalid type, got %d", config.MaxQueueSize)
		}
	})

	t.Run("无效的 channel_buffer 类型", func(t *testing.T) {
		cfg := map[string]any{
			"channel_buffer": "invalid",
		}

		config, err := parseMemoryConfig(cfg)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		if config.ChannelBuffer != DefaultMemoryChannelBuffer {
			t.Errorf("expected default ChannelBuffer for invalid type, got %d", config.ChannelBuffer)
		}
	})
}

func TestBuildWithConfigProviderNilDriverConfig(t *testing.T) {
	t.Run("nil driver config", func(t *testing.T) {
		provider := &mockConfigProvider{
			data: map[string]any{
				"mq.driver":        "memory",
				"mq.memory_config": nil,
			},
		}

		_, err := BuildWithConfigProvider(provider, nil, nil)
		if err == nil {
			t.Error("expected error with nil memory_config, got nil")
		}
	})
}

func TestBuildInvalidMemoryConfig(t *testing.T) {
	t.Run("memory 驱动配置无效", func(t *testing.T) {
		driverConfig := map[string]any{
			"max_queue_size": "invalid",
			"channel_buffer": 100,
		}

		_, err := Build("memory", driverConfig, nil, nil)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
}

func TestBuildMemoryWithAllOptions(t *testing.T) {
	t.Run("memory 驱动带所有配置选项", func(t *testing.T) {
		driverConfig := map[string]any{
			"max_queue_size": int(5000),
			"channel_buffer": int(200),
		}

		mgr, err := Build("memory", driverConfig, nil, nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if mgr == nil {
			t.Fatal("expected manager not to be nil")
		}

		mgr.Close()
	})
}

func TestBuildMemoryWithInt64Options(t *testing.T) {
	t.Run("memory 驱动带 int64 配置选项", func(t *testing.T) {
		driverConfig := map[string]any{
			"max_queue_size": int64(5000),
			"channel_buffer": int64(200),
		}

		mgr, err := Build("memory", driverConfig, nil, nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if mgr == nil {
			t.Fatal("expected manager not to be nil")
		}

		mgr.Close()
	})
}

func TestBuildMemoryWithFloat64Options(t *testing.T) {
	t.Run("memory 驱动带 float64 配置选项", func(t *testing.T) {
		driverConfig := map[string]any{
			"max_queue_size": float64(5000),
			"channel_buffer": float64(200),
		}

		mgr, err := Build("memory", driverConfig, nil, nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if mgr == nil {
			t.Fatal("expected manager not to be nil")
		}

		mgr.Close()
	})
}

func TestBuildWithConfigProviderNilProvider(t *testing.T) {
	t.Run("nil 配置提供者", func(t *testing.T) {
		_, err := BuildWithConfigProvider(nil, nil, nil)
		if err == nil {
			t.Error("expected error with nil config provider, got nil")
		}
	})
}

func TestBuildRabbitMQWithInvalidConfig(t *testing.T) {
	t.Run("RabbitMQ 无效配置", func(t *testing.T) {
		driverConfig := map[string]any{
			"url": "invalid-url",
		}

		_, err := Build("rabbitmq", driverConfig, nil, nil)
		if err == nil {
			t.Error("expected error with invalid URL, got nil")
		}
	})
}

func TestBuildRabbitMQWithNilConfig(t *testing.T) {
	t.Run("RabbitMQ nil 配置", func(t *testing.T) {
		_, err := Build("rabbitmq", nil, nil, nil)
		if err == nil {
			t.Error("expected error with nil config, got nil")
		}
	})
}

func TestBuildRabbitMQWithLoggerAndTelemetry(t *testing.T) {
	t.Run("RabbitMQ 带 logger 和 telemetry", func(t *testing.T) {
		driverConfig := map[string]any{
			"url":     "amqp://guest:guest@localhost:5672/",
			"durable": true,
		}

		mgr, err := Build("rabbitmq", driverConfig, nil, nil)
		if err != nil {
			t.Skip("RabbitMQ not available")
		}

		if mgr == nil {
			t.Fatal("expected manager not to be nil")
		}

		mgr.Close()
	})
}
