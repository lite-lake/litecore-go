package telemetrymgr

import (
	"testing"
)

// TestDefaultConfig 测试默认配置
func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig()

	if config.Driver != "none" {
		t.Errorf("expected driver 'none', got '%s'", config.Driver)
	}

	if config.OtelConfig == nil {
		t.Fatal("otel config should not be nil")
	}

	if config.OtelConfig.Endpoint != DefaultOtelEndpoint {
		t.Errorf("expected endpoint '%s', got '%s'", DefaultOtelEndpoint, config.OtelConfig.Endpoint)
	}

	if config.OtelConfig.Insecure != DefaultOtelInsecure {
		t.Errorf("expected insecure %v, got %v", DefaultOtelInsecure, config.OtelConfig.Insecure)
	}
}

// TestTelemetryConfigValidate 测试配置验证
func TestTelemetryConfigValidate(t *testing.T) {
	tests := []struct {
		name    string
		config  *TelemetryConfig
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid none driver",
			config: &TelemetryConfig{
				Driver: "none",
			},
			wantErr: false,
		},
		{
			name: "valid otel driver",
			config: &TelemetryConfig{
				Driver: "otel",
				OtelConfig: &OtelConfig{
					Endpoint: "localhost:4317",
					Traces:   &FeatureConfig{Enabled: false},
					Metrics:  &FeatureConfig{Enabled: false},
					Logs:     &FeatureConfig{Enabled: false},
				},
			},
			wantErr: false,
		},
		{
			name: "empty driver",
			config: &TelemetryConfig{
				Driver: "",
			},
			wantErr: true,
			errMsg:  "telemetry driver is required",
		},
		{
			name: "otel driver without otel config",
			config: &TelemetryConfig{
				Driver:     "otel",
				OtelConfig: nil,
			},
			wantErr: true,
			errMsg:  "otel_config is required when driver is 'otel'",
		},
		{
			name: "otel driver without endpoint",
			config: &TelemetryConfig{
				Driver: "otel",
				OtelConfig: &OtelConfig{
					Endpoint: "",
					Traces:   &FeatureConfig{Enabled: false},
					Metrics:  &FeatureConfig{Enabled: false},
					Logs:     &FeatureConfig{Enabled: false},
				},
			},
			wantErr: true,
			errMsg:  "otel endpoint is required",
		},
		{
			name: "unsupported driver",
			config: &TelemetryConfig{
				Driver: "invalid",
			},
			wantErr: true,
			errMsg:  "unsupported telemetry driver",
		},
		{
			name: "driver name normalized - uppercase OTEL",
			config: &TelemetryConfig{
				Driver: "OTEL",
				OtelConfig: &OtelConfig{
					Endpoint: "localhost:4317",
					Traces:   &FeatureConfig{Enabled: false},
					Metrics:  &FeatureConfig{Enabled: false},
					Logs:     &FeatureConfig{Enabled: false},
				},
			},
			wantErr: false,
		},
		{
			name: "driver name normalized - with spaces",
			config: &TelemetryConfig{
				Driver: "  otel  ",
				OtelConfig: &OtelConfig{
					Endpoint: "localhost:4317",
					Traces:   &FeatureConfig{Enabled: false},
					Metrics:  &FeatureConfig{Enabled: false},
					Logs:     &FeatureConfig{Enabled: false},
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error containing '%s', got nil", tt.errMsg)
				}
				if tt.errMsg != "" && err != nil {
					// 检查错误消息是否包含预期内容
					if len(err.Error()) < len(tt.errMsg) || err.Error()[:len(tt.errMsg)] != tt.errMsg {
						t.Errorf("expected error containing '%s', got '%s'", tt.errMsg, err.Error())
					}
				}
			} else {
				if err != nil {
					t.Errorf("expected no error, got %v", err)
				}
			}
		})
	}
}

// TestParseOtelConfig 测试 OTel 配置解析
func TestParseOtelConfig(t *testing.T) {
	tests := []struct {
		name    string
		input   map[string]any
		wantErr bool
		verify  func(*testing.T, *OtelConfig)
	}{
		{
			name:    "nil config returns defaults",
			input:   nil,
			wantErr: false,
			verify: func(t *testing.T, cfg *OtelConfig) {
				if cfg.Endpoint != DefaultOtelEndpoint {
					t.Errorf("expected default endpoint '%s', got '%s'", DefaultOtelEndpoint, cfg.Endpoint)
				}
				if cfg.Insecure != DefaultOtelInsecure {
					t.Errorf("expected default insecure %v, got %v", DefaultOtelInsecure, cfg.Insecure)
				}
			},
		},
		{
			name: "valid full config",
			input: map[string]any{
				"endpoint": "otel-collector:4317",
				"insecure": true,
				"headers": map[string]any{
					"authorization": "Bearer token123",
				},
				"resource_attributes": []any{
					map[string]any{
						"key":   "service.name",
						"value": "my-service",
					},
					map[string]any{
						"key":   "service.version",
						"value": "1.0.0",
					},
				},
				"traces": map[string]any{
					"enabled": true,
				},
				"metrics": map[string]any{
					"enabled": true,
				},
				"logs": map[string]any{
					"enabled": true,
				},
			},
			wantErr: false,
			verify: func(t *testing.T, cfg *OtelConfig) {
				if cfg.Endpoint != "otel-collector:4317" {
					t.Errorf("expected endpoint 'otel-collector:4317', got '%s'", cfg.Endpoint)
				}
				if !cfg.Insecure {
					t.Errorf("expected insecure true, got false")
				}
				if cfg.Headers["authorization"] != "Bearer token123" {
					t.Errorf("expected authorization header, got %v", cfg.Headers)
				}
				if len(cfg.ResourceAttributes) != 2 {
					t.Errorf("expected 2 resource attributes, got %d", len(cfg.ResourceAttributes))
				}
				if !cfg.Traces.Enabled {
					t.Errorf("expected traces enabled, got false")
				}
				if !cfg.Metrics.Enabled {
					t.Errorf("expected metrics enabled, got false")
				}
				if !cfg.Logs.Enabled {
					t.Errorf("expected logs enabled, got false")
				}
			},
		},
		{
			name: "endpoint with spaces is trimmed",
			input: map[string]any{
				"endpoint": "  otel-collector:4317  ",
			},
			wantErr: false,
			verify: func(t *testing.T, cfg *OtelConfig) {
				if cfg.Endpoint != "otel-collector:4317" {
					t.Errorf("expected trimmed endpoint, got '%s'", cfg.Endpoint)
				}
			},
		},
		{
			name: "resource attributes without key are ignored",
			input: map[string]any{
				"resource_attributes": []any{
					map[string]any{
						"value": "only-value",
					},
					map[string]any{
						"key":   "valid-key",
						"value": "valid-value",
					},
				},
			},
			wantErr: false,
			verify: func(t *testing.T, cfg *OtelConfig) {
				if len(cfg.ResourceAttributes) != 1 {
					t.Errorf("expected 1 resource attribute (invalid one ignored), got %d", len(cfg.ResourceAttributes))
				}
				if cfg.ResourceAttributes[0].Key != "valid-key" {
					t.Errorf("expected key 'valid-key', got '%s'", cfg.ResourceAttributes[0].Key)
				}
			},
		},
		{
			name: "headers with non-string values are ignored",
			input: map[string]any{
				"headers": map[string]any{
					"string-key":  "string-value",
					"number-key":  123,
					"bool-key":    true,
					"object-key":  map[string]any{},
					"slice-key":   []any{},
					"null-key":    nil,
					"valid-value": "value",
				},
			},
			wantErr: false,
			verify: func(t *testing.T, cfg *OtelConfig) {
				// 只有 string-key 和 valid-value 应该被保留
				if len(cfg.Headers) != 2 {
					t.Errorf("expected 2 headers, got %d: %v", len(cfg.Headers), cfg.Headers)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg, err := parseOtelConfig(tt.input)
			if tt.wantErr && err == nil {
				t.Errorf("expected error, got nil")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("expected no error, got %v", err)
			}
			if err == nil && tt.verify != nil {
				tt.verify(t, cfg)
			}
		})
	}
}

// TestParseTelemetryConfigFromMap 测试从 ConfigMap 解析配置
func TestParseTelemetryConfigFromMap(t *testing.T) {
	tests := []struct {
		name    string
		input   map[string]any
		wantErr bool
		verify  func(*testing.T, *TelemetryConfig)
	}{
		{
			name:    "nil config returns defaults",
			input:   nil,
			wantErr: false,
			verify: func(t *testing.T, cfg *TelemetryConfig) {
				if cfg.Driver != "none" {
					t.Errorf("expected default driver 'none', got '%s'", cfg.Driver)
				}
			},
		},
		{
			name: "empty map returns defaults",
			input: map[string]any{},
			wantErr: false,
			verify: func(t *testing.T, cfg *TelemetryConfig) {
				if cfg.Driver != "none" {
					t.Errorf("expected default driver 'none', got '%s'", cfg.Driver)
				}
			},
		},
		{
			name: "none driver config",
			input: map[string]any{
				"driver": "none",
			},
			wantErr: false,
			verify: func(t *testing.T, cfg *TelemetryConfig) {
				if cfg.Driver != "none" {
					t.Errorf("expected driver 'none', got '%s'", cfg.Driver)
				}
			},
		},
		{
			name: "otel driver config",
			input: map[string]any{
				"driver": "otel",
				"otel_config": map[string]any{
					"endpoint": "otel:4317",
					"insecure": true,
				},
			},
			wantErr: false,
			verify: func(t *testing.T, cfg *TelemetryConfig) {
				if cfg.Driver != "otel" {
					t.Errorf("expected driver 'otel', got '%s'", cfg.Driver)
				}
				if cfg.OtelConfig.Endpoint != "otel:4317" {
					t.Errorf("expected endpoint 'otel:4317', got '%s'", cfg.OtelConfig.Endpoint)
				}
				if !cfg.OtelConfig.Insecure {
					t.Errorf("expected insecure true, got false")
				}
			},
		},
		{
			name: "driver name case insensitive",
			input: map[string]any{
				"driver": "OTEL",
				"otel_config": map[string]any{
					"endpoint": "otel:4317",
				},
			},
			wantErr: false,
			verify: func(t *testing.T, cfg *TelemetryConfig) {
				if cfg.Driver != "otel" {
					t.Errorf("expected normalized driver 'otel', got '%s'", cfg.Driver)
				}
			},
		},
		{
			name: "otel driver ignores invalid otel_config type and uses default",
			input: map[string]any{
				"driver":      "otel",
				"otel_config": "invalid", // 不是 map 类型，会被忽略
			},
			wantErr: false,
			verify: func(t *testing.T, cfg *TelemetryConfig) {
				if cfg.Driver != "otel" {
					t.Errorf("expected driver 'otel', got '%s'", cfg.Driver)
				}
				// 应该使用默认的 OtelConfig
				if cfg.OtelConfig.Endpoint != DefaultOtelEndpoint {
					t.Errorf("expected default endpoint '%s', got '%s'", DefaultOtelEndpoint, cfg.OtelConfig.Endpoint)
				}
			},
		},
		{
			name: "driver with spaces is trimmed",
			input: map[string]any{
				"driver": "  none  ",
			},
			wantErr: false,
			verify: func(t *testing.T, cfg *TelemetryConfig) {
				if cfg.Driver != "none" {
					t.Errorf("expected trimmed driver 'none', got '%s'", cfg.Driver)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg, err := ParseTelemetryConfigFromMap(tt.input)
			if tt.wantErr && err == nil {
				t.Errorf("expected error, got nil")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("expected no error, got %v", err)
			}
			if err == nil && tt.verify != nil {
				tt.verify(t, cfg)
			}
		})
	}
}

// TestOtelConfigDefaults 测试 OtelConfig 默认值
func TestOtelConfigDefaults(t *testing.T) {
	config := DefaultConfig()
	otelConfig := config.OtelConfig

	if otelConfig.Traces == nil || otelConfig.Traces.Enabled {
		t.Error("expected traces disabled by default")
	}

	if otelConfig.Metrics == nil || otelConfig.Metrics.Enabled {
		t.Error("expected metrics disabled by default")
	}

	if otelConfig.Logs == nil || otelConfig.Logs.Enabled {
		t.Error("expected logs disabled by default")
	}

	if otelConfig.Headers == nil {
		t.Error("expected headers map to be initialized")
	}

	if len(otelConfig.ResourceAttributes) != 0 {
		t.Error("expected no resource attributes by default")
	}
}
