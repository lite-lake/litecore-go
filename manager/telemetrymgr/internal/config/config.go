package config

import (
	"fmt"
)

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
		return fmt.Errorf("unsupported telemetry driver: %s", c.Driver)
	}

	return nil
}

// DefaultConfig 返回默认配置（禁用观测）
func DefaultConfig() *TelemetryConfig {
	return &TelemetryConfig{
		Driver: "none",
		OtelConfig: &OtelConfig{
			Insecure: false,
			Headers:  make(map[string]string),
			Traces:   &FeatureConfig{Enabled: false},
			Metrics:  &FeatureConfig{Enabled: false},
			Logs:     &FeatureConfig{Enabled: false},
		},
	}
}

// ParseOtelConfigFromMap 从 ConfigMap 解析 OTEL 配置
// cfg 是 OTEL 驱动的配置内容（不包含 driver 字段）
// strict: 是否严格模式，严格模式下类型断言失败会返回错误
func ParseOtelConfigFromMap(cfg map[string]any, strict ...bool) (*OtelConfig, error) {
	isStrict := len(strict) > 0 && strict[0]
	otelCfg := &OtelConfig{
		Insecure: false, // 默认使用安全连接（TLS）
		Headers:  make(map[string]string),
		Traces:   &FeatureConfig{Enabled: false},
		Metrics:  &FeatureConfig{Enabled: false},
		Logs:     &FeatureConfig{Enabled: false},
	}

	if cfg == nil {
		return otelCfg, nil
	}

	// 已知的字段名集合
	knownFields := map[string]bool{
		"endpoint":            true,
		"insecure":            true,
		"resource_attributes": true,
		"headers":             true,
		"traces":              true,
		"metrics":             true,
		"logs":                true,
	}

	// 在严格模式下检查未知字段
	if isStrict {
		for key := range cfg {
			if !knownFields[key] {
				return nil, fmt.Errorf("unknown config field: %s", key)
			}
		}
	}

	// 解析 endpoint
	if endpoint, ok := cfg["endpoint"].(string); ok {
		otelCfg.Endpoint = endpoint
	} else if isStrict && cfg["endpoint"] != nil {
		return nil, fmt.Errorf("invalid type for 'endpoint': expected string, got %T", cfg["endpoint"])
	}

	// 解析 insecure
	if insecure, ok := cfg["insecure"].(bool); ok {
		otelCfg.Insecure = insecure
	} else if isStrict && cfg["insecure"] != nil {
		return nil, fmt.Errorf("invalid type for 'insecure': expected bool, got %T", cfg["insecure"])
	}

	// 解析 headers
	if headersMap, ok := cfg["headers"].(map[string]any); ok {
		for k, v := range headersMap {
			if strVal, ok := v.(string); ok {
				otelCfg.Headers[k] = strVal
			} else if isStrict {
				return nil, fmt.Errorf("invalid type for header '%s': expected string, got %T", k, v)
			}
		}
	} else if isStrict && cfg["headers"] != nil {
		return nil, fmt.Errorf("invalid type for 'headers': expected map[string]any, got %T", cfg["headers"])
	}

	// 解析 resource_attributes
	if attrsList, ok := cfg["resource_attributes"].([]any); ok {
		for i, attr := range attrsList {
			if attrMap, ok := attr.(map[string]any); ok {
				ra := ResourceAttribute{}
				if key, ok := attrMap["key"].(string); ok {
					ra.Key = key
				} else if isStrict && attrMap["key"] != nil {
					return nil, fmt.Errorf("invalid type for resource_attributes[%d].key: expected string, got %T", i, attrMap["key"])
				}
				if value, ok := attrMap["value"].(string); ok {
					ra.Value = value
				} else if isStrict && attrMap["value"] != nil {
					return nil, fmt.Errorf("invalid type for resource_attributes[%d].value: expected string, got %T", i, attrMap["value"])
				}
				otelCfg.ResourceAttributes = append(otelCfg.ResourceAttributes, ra)
			} else if isStrict {
				return nil, fmt.Errorf("invalid type for resource_attributes[%d]: expected map[string]any, got %T", i, attr)
			}
		}
	} else if isStrict && cfg["resource_attributes"] != nil {
		return nil, fmt.Errorf("invalid type for 'resource_attributes': expected []any, got %T", cfg["resource_attributes"])
	}

	// 解析 traces
	if tracesMap, ok := cfg["traces"].(map[string]any); ok {
		if enabled, ok := tracesMap["enabled"].(bool); ok {
			otelCfg.Traces.Enabled = enabled
		} else if isStrict && tracesMap["enabled"] != nil {
			return nil, fmt.Errorf("invalid type for traces.enabled: expected bool, got %T", tracesMap["enabled"])
		}
	} else if isStrict && cfg["traces"] != nil {
		return nil, fmt.Errorf("invalid type for 'traces': expected map[string]any, got %T", cfg["traces"])
	}

	// 解析 metrics
	if metricsMap, ok := cfg["metrics"].(map[string]any); ok {
		if enabled, ok := metricsMap["enabled"].(bool); ok {
			otelCfg.Metrics.Enabled = enabled
		} else if isStrict && metricsMap["enabled"] != nil {
			return nil, fmt.Errorf("invalid type for metrics.enabled: expected bool, got %T", metricsMap["enabled"])
		}
	} else if isStrict && cfg["metrics"] != nil {
		return nil, fmt.Errorf("invalid type for 'metrics': expected map[string]any, got %T", cfg["metrics"])
	}

	// 解析 logs
	if logsMap, ok := cfg["logs"].(map[string]any); ok {
		if enabled, ok := logsMap["enabled"].(bool); ok {
			otelCfg.Logs.Enabled = enabled
		} else if isStrict && logsMap["enabled"] != nil {
			return nil, fmt.Errorf("invalid type for logs.enabled: expected bool, got %T", logsMap["enabled"])
		}
	} else if isStrict && cfg["logs"] != nil {
		return nil, fmt.Errorf("invalid type for 'logs': expected map[string]any, got %T", cfg["logs"])
	}

	return otelCfg, nil
}
