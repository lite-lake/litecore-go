package mqmgr

import (
	"testing"
	"time"
)

func TestDefaultConfig(t *testing.T) {
	t.Run("返回默认配置", func(t *testing.T) {
		config := DefaultConfig()

		if config.Driver != "memory" {
			t.Errorf("expected driver 'memory', got '%s'", config.Driver)
		}

		if config.RabbitMQConfig == nil {
			t.Error("expected RabbitMQConfig not to be nil")
		}

		if config.RabbitMQConfig.URL != DefaultRabbitMQURL {
			t.Errorf("expected URL '%s', got '%s'", DefaultRabbitMQURL, config.RabbitMQConfig.URL)
		}

		if config.RabbitMQConfig.Durable != DefaultRabbitMQDurable {
			t.Errorf("expected Durable %v, got %v", DefaultRabbitMQDurable, config.RabbitMQConfig.Durable)
		}

		if config.MemoryConfig == nil {
			t.Error("expected MemoryConfig not to be nil")
		}

		if config.MemoryConfig.MaxQueueSize != DefaultMemoryMaxQueueSize {
			t.Errorf("expected MaxQueueSize %d, got %d", DefaultMemoryMaxQueueSize, config.MemoryConfig.MaxQueueSize)
		}

		if config.MemoryConfig.ChannelBuffer != DefaultMemoryChannelBuffer {
			t.Errorf("expected ChannelBuffer %d, got %d", DefaultMemoryChannelBuffer, config.MemoryConfig.ChannelBuffer)
		}
	})
}

func TestMQConfig_Validate(t *testing.T) {
	t.Run("空驱动类型", func(t *testing.T) {
		config := &MQConfig{
			Driver: "",
		}

		err := config.Validate()
		if err == nil {
			t.Error("expected error with empty driver")
		}
	})

	t.Run("不支持的驱动类型", func(t *testing.T) {
		config := &MQConfig{
			Driver: "unsupported",
		}

		err := config.Validate()
		if err == nil {
			t.Error("expected error with unsupported driver")
		}
	})

	t.Run("驱动类型大小写和空格处理", func(t *testing.T) {
		config := &MQConfig{
			Driver: "  RabbitMQ  ",
			RabbitMQConfig: &RabbitMQConfig{
				URL:     DefaultRabbitMQURL,
				Durable: true,
			},
		}

		err := config.Validate()
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		if config.Driver != "rabbitmq" {
			t.Errorf("expected driver 'rabbitmq', got '%s'", config.Driver)
		}
	})

	t.Run("rabbitmq 驱动缺少配置", func(t *testing.T) {
		config := &MQConfig{
			Driver:         "rabbitmq",
			RabbitMQConfig: nil,
		}

		err := config.Validate()
		if err == nil {
			t.Error("expected error when RabbitMQConfig is nil")
		}
	})

	t.Run("rabbitmq 驱动完整配置", func(t *testing.T) {
		config := &MQConfig{
			Driver: "rabbitmq",
			RabbitMQConfig: &RabbitMQConfig{
				URL:     "amqp://localhost:5672/",
				Durable: true,
			},
		}

		err := config.Validate()
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})

	t.Run("memory 驱动缺少配置", func(t *testing.T) {
		config := &MQConfig{
			Driver:       "memory",
			MemoryConfig: nil,
		}

		err := config.Validate()
		if err == nil {
			t.Error("expected error when MemoryConfig is nil")
		}
	})

	t.Run("memory 驱动完整配置", func(t *testing.T) {
		config := &MQConfig{
			Driver: "memory",
			MemoryConfig: &MemoryConfig{
				MaxQueueSize:  1000,
				ChannelBuffer: 100,
			},
		}

		err := config.Validate()
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
}

func TestParseMQConfigFromMap(t *testing.T) {
	t.Run("空 map", func(t *testing.T) {
		config, err := ParseMQConfigFromMap(nil)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		if config.Driver != "memory" {
			t.Errorf("expected default driver 'memory', got '%s'", config.Driver)
		}
	})

	t.Run("解析 driver", func(t *testing.T) {
		cfg := map[string]any{
			"driver": "rabbitmq",
		}

		config, err := ParseMQConfigFromMap(cfg)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		if config.Driver != "rabbitmq" {
			t.Errorf("expected driver 'rabbitmq', got '%s'", config.Driver)
		}
	})

	t.Run("解析 rabbitmq_config", func(t *testing.T) {
		cfg := map[string]any{
			"driver": "rabbitmq",
			"rabbitmq_config": map[string]any{
				"url":     "amqp://custom:5672/",
				"durable": false,
			},
		}

		config, err := ParseMQConfigFromMap(cfg)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		if config.RabbitMQConfig.URL != "amqp://custom:5672/" {
			t.Errorf("expected URL 'amqp://custom:5672/', got '%s'", config.RabbitMQConfig.URL)
		}

		if config.RabbitMQConfig.Durable != false {
			t.Errorf("expected Durable false, got %v", config.RabbitMQConfig.Durable)
		}
	})

	t.Run("解析 memory_config", func(t *testing.T) {
		cfg := map[string]any{
			"driver": "memory",
			"memory_config": map[string]any{
				"max_queue_size": 5000,
				"channel_buffer": 200,
			},
		}

		config, err := ParseMQConfigFromMap(cfg)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		if config.MemoryConfig.MaxQueueSize != 5000 {
			t.Errorf("expected MaxQueueSize 5000, got %d", config.MemoryConfig.MaxQueueSize)
		}

		if config.MemoryConfig.ChannelBuffer != 200 {
			t.Errorf("expected ChannelBuffer 200, got %d", config.MemoryConfig.ChannelBuffer)
		}
	})
}

func TestParseRabbitMQConfig(t *testing.T) {
	t.Run("空配置使用默认值", func(t *testing.T) {
		config, err := parseRabbitMQConfig(nil)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		if config.URL != DefaultRabbitMQURL {
			t.Errorf("expected default URL, got '%s'", config.URL)
		}

		if config.Durable != DefaultRabbitMQDurable {
			t.Errorf("expected default Durable, got %v", config.Durable)
		}
	})

	t.Run("自定义 URL", func(t *testing.T) {
		cfg := map[string]any{
			"url": "amqp://custom:5672/",
		}

		config, err := parseRabbitMQConfig(cfg)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		if config.URL != "amqp://custom:5672/" {
			t.Errorf("expected URL 'amqp://custom:5672/', got '%s'", config.URL)
		}
	})

	t.Run("URL 去除空格", func(t *testing.T) {
		cfg := map[string]any{
			"url": "  amqp://custom:5672/  ",
		}

		config, err := parseRabbitMQConfig(cfg)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		if config.URL != "amqp://custom:5672/" {
			t.Errorf("expected trimmed URL, got '%s'", config.URL)
		}
	})

	t.Run("自定义 Durable", func(t *testing.T) {
		cfg := map[string]any{
			"durable": false,
		}

		config, err := parseRabbitMQConfig(cfg)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		if config.Durable != false {
			t.Errorf("expected Durable false, got %v", config.Durable)
		}
	})
}

func TestParseMemoryConfig(t *testing.T) {
	t.Run("空配置使用默认值", func(t *testing.T) {
		config, err := parseMemoryConfig(nil)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		if config.MaxQueueSize != DefaultMemoryMaxQueueSize {
			t.Errorf("expected default MaxQueueSize, got %d", config.MaxQueueSize)
		}

		if config.ChannelBuffer != DefaultMemoryChannelBuffer {
			t.Errorf("expected default ChannelBuffer, got %d", config.ChannelBuffer)
		}
	})

	t.Run("自定义 MaxQueueSize", func(t *testing.T) {
		cfg := map[string]any{
			"max_queue_size": 5000,
		}

		config, err := parseMemoryConfig(cfg)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		if config.MaxQueueSize != 5000 {
			t.Errorf("expected MaxQueueSize 5000, got %d", config.MaxQueueSize)
		}
	})

	t.Run("自定义 ChannelBuffer", func(t *testing.T) {
		cfg := map[string]any{
			"channel_buffer": 200,
		}

		config, err := parseMemoryConfig(cfg)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		if config.ChannelBuffer != 200 {
			t.Errorf("expected ChannelBuffer 200, got %d", config.ChannelBuffer)
		}
	})

	t.Run("负值忽略", func(t *testing.T) {
		cfg := map[string]any{
			"max_queue_size": -1,
		}

		config, err := parseMemoryConfig(cfg)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		if config.MaxQueueSize != DefaultMemoryMaxQueueSize {
			t.Errorf("expected default MaxQueueSize for negative value, got %d", config.MaxQueueSize)
		}
	})

	t.Run("零值忽略", func(t *testing.T) {
		cfg := map[string]any{
			"max_queue_size": 0,
		}

		config, err := parseMemoryConfig(cfg)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		if config.MaxQueueSize != DefaultMemoryMaxQueueSize {
			t.Errorf("expected default MaxQueueSize for zero value, got %d", config.MaxQueueSize)
		}
	})
}

func TestToInt(t *testing.T) {
	t.Run("int 类型", func(t *testing.T) {
		result, ok := toInt(42)
		if !ok || result != 42 {
			t.Errorf("expected (42, true), got (%d, %v)", result, ok)
		}
	})

	t.Run("int64 类型", func(t *testing.T) {
		result, ok := toInt(int64(42))
		if !ok || result != 42 {
			t.Errorf("expected (42, true), got (%d, %v)", result, ok)
		}
	})

	t.Run("float64 整数类型", func(t *testing.T) {
		result, ok := toInt(42.0)
		if !ok || result != 42 {
			t.Errorf("expected (42, true), got (%d, %v)", result, ok)
		}
	})

	t.Run("float64 小数类型", func(t *testing.T) {
		result, ok := toInt(42.5)
		if ok || result != 0 {
			t.Errorf("expected (0, false), got (%d, %v)", result, ok)
		}
	})

	t.Run("不支持的类型", func(t *testing.T) {
		result, ok := toInt("42")
		if ok || result != 0 {
			t.Errorf("expected (0, false), got (%d, %v)", result, ok)
		}
	})

	t.Run("nil 类型", func(t *testing.T) {
		result, ok := toInt(nil)
		if ok || result != 0 {
			t.Errorf("expected (0, false), got (%d, %v)", result, ok)
		}
	})
}

func TestParseDuration(t *testing.T) {
	t.Run("int 类型", func(t *testing.T) {
		duration, err := parseDuration(10)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		if duration != 10*time.Second {
			t.Errorf("expected 10s, got %v", duration)
		}
	})

	t.Run("int64 类型", func(t *testing.T) {
		duration, err := parseDuration(int64(10))
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		if duration != 10*time.Second {
			t.Errorf("expected 10s, got %v", duration)
		}
	})

	t.Run("float64 整数类型", func(t *testing.T) {
		duration, err := parseDuration(10.0)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		if duration != 10*time.Second {
			t.Errorf("expected 10s, got %v", duration)
		}
	})

	t.Run("float64 小数类型（会被截断）", func(t *testing.T) {
		duration, err := parseDuration(2.7)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		if duration != 2*time.Second {
			t.Errorf("expected 2s (truncated), got %v", duration)
		}
	})

	t.Run("字符串类型 - 秒", func(t *testing.T) {
		duration, err := parseDuration("10s")
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		if duration != 10*time.Second {
			t.Errorf("expected 10s, got %v", duration)
		}
	})

	t.Run("字符串类型 - 分钟", func(t *testing.T) {
		duration, err := parseDuration("5m")
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		if duration != 5*time.Minute {
			t.Errorf("expected 5m, got %v", duration)
		}
	})

	t.Run("字符串类型 - 小时", func(t *testing.T) {
		duration, err := parseDuration("2h")
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		if duration != 2*time.Hour {
			t.Errorf("expected 2h, got %v", duration)
		}
	})

	t.Run("字符串类型 - 毫秒", func(t *testing.T) {
		duration, err := parseDuration("500ms")
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		if duration != 500*time.Millisecond {
			t.Errorf("expected 500ms, got %v", duration)
		}
	})

	t.Run("字符串类型 - 无效格式", func(t *testing.T) {
		_, err := parseDuration("invalid")
		if err == nil {
			t.Error("expected error with invalid duration format")
		}
	})

	t.Run("不支持的类型", func(t *testing.T) {
		_, err := parseDuration(true)
		if err == nil {
			t.Error("expected error with unsupported type")
		}
	})

	t.Run("nil 类型", func(t *testing.T) {
		_, err := parseDuration(nil)
		if err == nil {
			t.Error("expected error with nil type")
		}
	})
}

func TestParseMQConfigFromMapInvalidDriver(t *testing.T) {
	t.Run("无效的 driver 类型", func(t *testing.T) {
		cfg := map[string]any{
			"driver": 123,
		}

		config, err := ParseMQConfigFromMap(cfg)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		if config.Driver != "memory" {
			t.Errorf("expected default driver 'memory', got '%s'", config.Driver)
		}
	})
}

func TestParseRabbitMQConfigAllFields(t *testing.T) {
	t.Run("所有字段都设置", func(t *testing.T) {
		cfg := map[string]any{
			"url":     "amqp://custom:5672/",
			"durable": false,
		}

		config, err := parseRabbitMQConfig(cfg)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		if config.URL != "amqp://custom:5672/" {
			t.Errorf("expected URL 'amqp://custom:5672/', got '%s'", config.URL)
		}

		if config.Durable != false {
			t.Errorf("expected Durable false, got %v", config.Durable)
		}
	})
}

func TestParseMemoryConfigAllFields(t *testing.T) {
	t.Run("所有字段都设置", func(t *testing.T) {
		cfg := map[string]any{
			"max_queue_size": int(5000),
			"channel_buffer": int(200),
		}

		config, err := parseMemoryConfig(cfg)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		if config.MaxQueueSize != 5000 {
			t.Errorf("expected MaxQueueSize 5000, got %d", config.MaxQueueSize)
		}

		if config.ChannelBuffer != 200 {
			t.Errorf("expected ChannelBuffer 200, got %d", config.ChannelBuffer)
		}
	})
}

func TestParseMemoryConfigWithInt64(t *testing.T) {
	t.Run("使用 int64 类型", func(t *testing.T) {
		cfg := map[string]any{
			"max_queue_size": int64(5000),
			"channel_buffer": int64(200),
		}

		config, err := parseMemoryConfig(cfg)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		if config.MaxQueueSize != 5000 {
			t.Errorf("expected MaxQueueSize 5000, got %d", config.MaxQueueSize)
		}

		if config.ChannelBuffer != 200 {
			t.Errorf("expected ChannelBuffer 200, got %d", config.ChannelBuffer)
		}
	})
}

func TestParseMemoryConfigWithFloat64(t *testing.T) {
	t.Run("使用 float64 类型", func(t *testing.T) {
		cfg := map[string]any{
			"max_queue_size": float64(5000),
			"channel_buffer": float64(200),
		}

		config, err := parseMemoryConfig(cfg)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		if config.MaxQueueSize != 5000 {
			t.Errorf("expected MaxQueueSize 5000, got %d", config.MaxQueueSize)
		}

		if config.ChannelBuffer != 200 {
			t.Errorf("expected ChannelBuffer 200, got %d", config.ChannelBuffer)
		}
	})
}

func TestParseMemoryConfigWithZeroMaxQueueSize(t *testing.T) {
	t.Run("零值 max_queue_size 忽略", func(t *testing.T) {
		cfg := map[string]any{
			"max_queue_size": 0,
			"channel_buffer": 100,
		}

		config, err := parseMemoryConfig(cfg)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		if config.MaxQueueSize != DefaultMemoryMaxQueueSize {
			t.Errorf("expected default MaxQueueSize for zero value, got %d", config.MaxQueueSize)
		}

		if config.ChannelBuffer != 100 {
			t.Errorf("expected ChannelBuffer 100, got %d", config.ChannelBuffer)
		}
	})
}

func TestParseMemoryConfigWithZeroChannelBuffer(t *testing.T) {
	t.Run("零值 channel_buffer 忽略", func(t *testing.T) {
		cfg := map[string]any{
			"max_queue_size": 1000,
			"channel_buffer": 0,
		}

		config, err := parseMemoryConfig(cfg)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		if config.MaxQueueSize != 1000 {
			t.Errorf("expected MaxQueueSize 1000, got %d", config.MaxQueueSize)
		}

		if config.ChannelBuffer != DefaultMemoryChannelBuffer {
			t.Errorf("expected default ChannelBuffer for zero value, got %d", config.ChannelBuffer)
		}
	})
}

func TestParseRabbitMQConfigWithNil(t *testing.T) {
	t.Run("nil 配置使用默认值", func(t *testing.T) {
		config, err := parseRabbitMQConfig(nil)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		if config.URL != DefaultRabbitMQURL {
			t.Errorf("expected default URL, got '%s'", config.URL)
		}

		if config.Durable != DefaultRabbitMQDurable {
			t.Errorf("expected default Durable, got %v", config.Durable)
		}
	})
}

func TestParseMemoryConfigWithNil(t *testing.T) {
	t.Run("nil 配置使用默认值", func(t *testing.T) {
		config, err := parseMemoryConfig(nil)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		if config.MaxQueueSize != DefaultMemoryMaxQueueSize {
			t.Errorf("expected default MaxQueueSize, got %d", config.MaxQueueSize)
		}

		if config.ChannelBuffer != DefaultMemoryChannelBuffer {
			t.Errorf("expected default ChannelBuffer, got %d", config.ChannelBuffer)
		}
	})
}

func TestToIntFloatNonInteger(t *testing.T) {
	t.Run("float64 非整数", func(t *testing.T) {
		result, ok := toInt(1.5)
		if ok {
			t.Errorf("expected (0, false), got (%d, %v)", result, ok)
		}
	})
}

func TestToIntBool(t *testing.T) {
	t.Run("bool 类型", func(t *testing.T) {
		result, ok := toInt(true)
		if ok {
			t.Errorf("expected (0, false), got (%d, %v)", result, ok)
		}
	})
}
