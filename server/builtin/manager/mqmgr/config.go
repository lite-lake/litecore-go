package mqmgr

import (
	"fmt"
	"strings"
	"time"
)

const (
	DefaultRabbitMQURL         = "amqp://guest:guest@localhost:5672/"
	DefaultRabbitMQDurable     = true
	DefaultMemoryMaxQueueSize  = 10000
	DefaultMemoryChannelBuffer = 100
)

func DefaultConfig() *MQConfig {
	return &MQConfig{
		Driver: "memory",
		RabbitMQConfig: &RabbitMQConfig{
			URL:     DefaultRabbitMQURL,
			Durable: DefaultRabbitMQDurable,
		},
		MemoryConfig: &MemoryConfig{
			MaxQueueSize:  DefaultMemoryMaxQueueSize,
			ChannelBuffer: DefaultMemoryChannelBuffer,
		},
	}
}

type MQConfig struct {
	Driver         string          `yaml:"driver"`
	RabbitMQConfig *RabbitMQConfig `yaml:"rabbitmq_config"`
	MemoryConfig   *MemoryConfig   `yaml:"memory_config"`
}

type RabbitMQConfig struct {
	URL     string `yaml:"url"`
	Durable bool   `yaml:"durable"`
}

type MemoryConfig struct {
	MaxQueueSize  int `yaml:"max_queue_size"`
	ChannelBuffer int `yaml:"channel_buffer"`
}

func (c *MQConfig) Validate() error {
	if c.Driver == "" {
		return fmt.Errorf("driver is required")
	}

	c.Driver = strings.ToLower(strings.TrimSpace(c.Driver))

	switch c.Driver {
	case "rabbitmq", "memory":
	default:
		return fmt.Errorf("unsupported driver: %s (must be rabbitmq or memory)", c.Driver)
	}

	if c.Driver == "rabbitmq" && c.RabbitMQConfig == nil {
		return fmt.Errorf("rabbitmq_config is required when driver is rabbitmq")
	}

	if c.Driver == "memory" && c.MemoryConfig == nil {
		return fmt.Errorf("memory_config is required when driver is memory")
	}

	return nil
}

func ParseMQConfigFromMap(cfg map[string]any) (*MQConfig, error) {
	config := &MQConfig{
		Driver: "memory",
		RabbitMQConfig: &RabbitMQConfig{
			URL:     DefaultRabbitMQURL,
			Durable: DefaultRabbitMQDurable,
		},
		MemoryConfig: &MemoryConfig{
			MaxQueueSize:  DefaultMemoryMaxQueueSize,
			ChannelBuffer: DefaultMemoryChannelBuffer,
		},
	}

	if cfg == nil {
		return config, nil
	}

	if driver, ok := cfg["driver"].(string); ok {
		config.Driver = strings.ToLower(strings.TrimSpace(driver))
	}

	if rabbitmqConfigMap, ok := cfg["rabbitmq_config"].(map[string]any); ok {
		rabbitmqConfig, err := parseRabbitMQConfig(rabbitmqConfigMap)
		if err != nil {
			return nil, fmt.Errorf("failed to parse rabbitmq_config: %w", err)
		}
		config.RabbitMQConfig = rabbitmqConfig
	}

	if memoryConfigMap, ok := cfg["memory_config"].(map[string]any); ok {
		memoryConfig, err := parseMemoryConfig(memoryConfigMap)
		if err != nil {
			return nil, fmt.Errorf("failed to parse memory_config: %w", err)
		}
		config.MemoryConfig = memoryConfig
	}

	return config, nil
}

func parseRabbitMQConfig(cfg map[string]any) (*RabbitMQConfig, error) {
	config := &RabbitMQConfig{
		URL:     DefaultRabbitMQURL,
		Durable: DefaultRabbitMQDurable,
	}

	if url, ok := cfg["url"].(string); ok {
		config.URL = strings.TrimSpace(url)
	}

	if durable, ok := cfg["durable"].(bool); ok {
		config.Durable = durable
	}

	return config, nil
}

func parseMemoryConfig(cfg map[string]any) (*MemoryConfig, error) {
	config := &MemoryConfig{
		MaxQueueSize:  DefaultMemoryMaxQueueSize,
		ChannelBuffer: DefaultMemoryChannelBuffer,
	}

	if maxSize, ok := cfg["max_queue_size"]; ok {
		if v, ok := toInt(maxSize); ok && v > 0 {
			config.MaxQueueSize = v
		}
	}

	if buffer, ok := cfg["channel_buffer"]; ok {
		if v, ok := toInt(buffer); ok && v > 0 {
			config.ChannelBuffer = v
		}
	}

	return config, nil
}

func toInt(v any) (int, bool) {
	switch val := v.(type) {
	case int:
		return val, true
	case int64:
		return int(val), true
	case float64:
		if val == float64(int64(val)) {
			return int(val), true
		}
		return 0, false
	default:
		return 0, false
	}
}

func parseDuration(v any) (time.Duration, error) {
	switch val := v.(type) {
	case int:
		return time.Duration(val) * time.Second, nil
	case int64:
		return time.Duration(val) * time.Second, nil
	case float64:
		return time.Duration(val) * time.Second, nil
	case string:
		if duration, err := time.ParseDuration(val); err == nil {
			return duration, nil
		}
		return 0, fmt.Errorf("invalid duration format: %s", val)
	default:
		return 0, fmt.Errorf("unsupported duration type: %T", v)
	}
}
