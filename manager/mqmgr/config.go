package mqmgr

import (
	"fmt"
	"strings"
	"time"
)

const (
	// DefaultRabbitMQURL 默认 RabbitMQ 连接地址
	DefaultRabbitMQURL = "amqp://guest:guest@localhost:5672/"
	// DefaultRabbitMQDurable 默认 RabbitMQ 队列持久化配置
	DefaultRabbitMQDurable = true
	// DefaultMemoryMaxQueueSize 默认内存队列最大容量
	DefaultMemoryMaxQueueSize = 10000
	// DefaultMemoryChannelBuffer 默认内存通道缓冲区大小
	DefaultMemoryChannelBuffer = 100
)

// DefaultConfig 返回默认配置
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

// MQConfig 消息队列配置
type MQConfig struct {
	// Driver 驱动类型：rabbitmq 或 memory
	Driver string `yaml:"driver"`
	// RabbitMQConfig RabbitMQ 配置
	RabbitMQConfig *RabbitMQConfig `yaml:"rabbitmq_config"`
	// MemoryConfig 内存队列配置
	MemoryConfig *MemoryConfig `yaml:"memory_config"`
}

// RabbitMQConfig RabbitMQ 配置
type RabbitMQConfig struct {
	// URL 连接地址
	URL string `yaml:"url"`
	// Durable 是否持久化
	Durable bool `yaml:"durable"`
}

// MemoryConfig 内存队列配置
type MemoryConfig struct {
	// MaxQueueSize 最大队列大小
	MaxQueueSize int `yaml:"max_queue_size"`
	// ChannelBuffer 通道缓冲区大小
	ChannelBuffer int `yaml:"channel_buffer"`
}

// Validate 验证配置
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

// ParseMQConfigFromMap 从 map 解析消息队列配置
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

// parseRabbitMQConfig 解析 RabbitMQ 配置
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

// parseMemoryConfig 解析内存队列配置
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

// toInt 将任意值转换为整数
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

// parseDuration 将任意值转换为时长
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
