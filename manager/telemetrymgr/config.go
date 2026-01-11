package telemetrymgr

import (
	"fmt"
	"strings"
)

const (
	// 默认值
	DefaultOtelEndpoint = "localhost:4317"
	DefaultOtelInsecure = false
)

// DefaultConfig 返回默认配置（禁用观测）
func DefaultConfig() *TelemetryConfig {
	return &TelemetryConfig{
		Driver: "none",
		OtelConfig: &OtelConfig{
			Endpoint: DefaultOtelEndpoint,
			Insecure: DefaultOtelInsecure,
			Headers:  make(map[string]string),
			Traces:   &FeatureConfig{Enabled: false},
			Metrics:  &FeatureConfig{Enabled: false},
			Logs:     &FeatureConfig{Enabled: false},
		},
	}
}

// TelemetryConfig 观测管理配置
type TelemetryConfig struct {
	Driver     string      `yaml:"driver"`      // 驱动类型: none, otel
	OtelConfig *OtelConfig `yaml:"otel_config"` // OTEL 驱动配置
}

// OtelConfig OpenTelemetry 配置
type OtelConfig struct {
	Endpoint           string              `yaml:"endpoint"`            // OTLP 端点，如 http://localhost:4317
	Insecure           bool                `yaml:"insecure"`            // 是否使用不安全连接（默认false，使用TLS）
	ResourceAttributes []ResourceAttribute `yaml:"resource_attributes"` // 资源属性
	Headers            map[string]string   `yaml:"headers"`             // 请求头（用于认证）
	Traces             *FeatureConfig      `yaml:"traces"`              // 链路追踪配置
	Metrics            *FeatureConfig      `yaml:"metrics"`             // 指标配置
	Logs               *FeatureConfig      `yaml:"logs"`                // 日志配置
}

// ResourceAttribute 资源属性
type ResourceAttribute struct {
	Key   string `yaml:"key"`
	Value string `yaml:"value"`
}

// FeatureConfig 功能配置
type FeatureConfig struct {
	Enabled bool `yaml:"enabled"`
}

// Validate 验证配置
func (c *TelemetryConfig) Validate() error {
	if c.Driver == "" {
		return fmt.Errorf("telemetry driver is required")
	}

	// 标准化驱动名称
	c.Driver = strings.ToLower(strings.TrimSpace(c.Driver))

	switch c.Driver {
	case "otel":
		if c.OtelConfig == nil {
			return fmt.Errorf("otel_config is required when driver is 'otel'")
		}
		if c.OtelConfig.Endpoint == "" {
			return fmt.Errorf("otel endpoint is required")
		}
	case "none":
		// none 驱动无需额外配置
	default:
		return fmt.Errorf("unsupported telemetry driver: %s (must be otel or none)", c.Driver)
	}

	return nil
}

// ParseTelemetryConfigFromMap 从 ConfigMap 解析观测配置
func ParseTelemetryConfigFromMap(cfg map[string]any) (*TelemetryConfig, error) {
	config := &TelemetryConfig{
		Driver: "none", // 默认使用 none 驱动
		OtelConfig: &OtelConfig{
			Endpoint: DefaultOtelEndpoint,
			Insecure: DefaultOtelInsecure,
			Headers:  make(map[string]string),
			Traces:   &FeatureConfig{Enabled: false},
			Metrics:  &FeatureConfig{Enabled: false},
			Logs:     &FeatureConfig{Enabled: false},
		},
	}

	if cfg == nil {
		return config, nil
	}

	// 解析 driver
	if driver, ok := cfg["driver"].(string); ok {
		config.Driver = strings.ToLower(strings.TrimSpace(driver))
	}

	// 解析 otel_config
	if otelConfigMap, ok := cfg["otel_config"].(map[string]any); ok {
		otelConfig, err := parseOtelConfig(otelConfigMap)
		if err != nil {
			return nil, fmt.Errorf("failed to parse otel_config: %w", err)
		}
		config.OtelConfig = otelConfig
	}

	return config, nil
}

// parseOtelConfig 解析 OTEL 配置
func parseOtelConfig(cfg map[string]any) (*OtelConfig, error) {
	otelConfig := &OtelConfig{
		Endpoint: DefaultOtelEndpoint,
		Insecure: DefaultOtelInsecure,
		Headers:  make(map[string]string),
		Traces:   &FeatureConfig{Enabled: false},
		Metrics:  &FeatureConfig{Enabled: false},
		Logs:     &FeatureConfig{Enabled: false},
	}

	if cfg == nil {
		return otelConfig, nil
	}

	// 解析 endpoint
	if endpoint, ok := cfg["endpoint"].(string); ok {
		otelConfig.Endpoint = strings.TrimSpace(endpoint)
	}

	// 解析 insecure
	if insecure, ok := cfg["insecure"].(bool); ok {
		otelConfig.Insecure = insecure
	}

	// 解析 headers
	if headersMap, ok := cfg["headers"].(map[string]any); ok {
		for k, v := range headersMap {
			if strVal, ok := v.(string); ok {
				otelConfig.Headers[k] = strVal
			}
		}
	}

	// 解析 resource_attributes
	if attrsList, ok := cfg["resource_attributes"].([]any); ok {
		for _, attr := range attrsList {
			if attrMap, ok := attr.(map[string]any); ok {
				ra := ResourceAttribute{}
				if key, ok := attrMap["key"].(string); ok {
					ra.Key = key
				}
				if value, ok := attrMap["value"].(string); ok {
					ra.Value = value
				}
				if ra.Key != "" {
					otelConfig.ResourceAttributes = append(otelConfig.ResourceAttributes, ra)
				}
			}
		}
	}

	// 解析 traces
	if tracesMap, ok := cfg["traces"].(map[string]any); ok {
		if enabled, ok := tracesMap["enabled"].(bool); ok {
			if otelConfig.Traces == nil {
				otelConfig.Traces = &FeatureConfig{}
			}
			otelConfig.Traces.Enabled = enabled
		}
	}

	// 解析 metrics
	if metricsMap, ok := cfg["metrics"].(map[string]any); ok {
		if enabled, ok := metricsMap["enabled"].(bool); ok {
			if otelConfig.Metrics == nil {
				otelConfig.Metrics = &FeatureConfig{}
			}
			otelConfig.Metrics.Enabled = enabled
		}
	}

	// 解析 logs
	if logsMap, ok := cfg["logs"].(map[string]any); ok {
		if enabled, ok := logsMap["enabled"].(bool); ok {
			if otelConfig.Logs == nil {
				otelConfig.Logs = &FeatureConfig{}
			}
			otelConfig.Logs.Enabled = enabled
		}
	}

	return otelConfig, nil
}
