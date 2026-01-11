package config

import (
	"testing"
)

func TestTelemetryConfigValidate(t *testing.T) {
	tests := []struct {
		name    string
		cfg     *TelemetryConfig
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid otel config",
			cfg: &TelemetryConfig{
				Driver: "otel",
				OtelConfig: &OtelConfig{
					Endpoint: "http://localhost:4317",
				},
			},
			wantErr: false,
		},
		{
			name: "valid none config",
			cfg: &TelemetryConfig{
				Driver:     "none",
				OtelConfig: nil,
			},
			wantErr: false,
		},
		{
			name: "empty driver",
			cfg: &TelemetryConfig{
				Driver: "",
			},
			wantErr: true,
			errMsg:  "telemetry driver is required",
		},
		{
			name: "otel driver without config",
			cfg: &TelemetryConfig{
				Driver:     "otel",
				OtelConfig: nil,
			},
			wantErr: true,
			errMsg:  "otel_config is required when driver is 'otel'",
		},
		{
			name: "otel driver without endpoint",
			cfg: &TelemetryConfig{
				Driver: "otel",
				OtelConfig: &OtelConfig{
					Endpoint: "",
				},
			},
			wantErr: true,
			errMsg:  "otel endpoint is required",
		},
		{
			name: "unsupported driver",
			cfg: &TelemetryConfig{
				Driver: "unsupported",
			},
			wantErr: true,
			errMsg:  "unsupported telemetry driver: unsupported",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.cfg.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && err.Error() != tt.errMsg {
				t.Errorf("Validate() error message = %v, want %v", err.Error(), tt.errMsg)
			}
		})
	}
}

func TestParseOtelConfigFromMap(t *testing.T) {
	tests := []struct {
		name    string
		cfg     map[string]any
		strict  bool
		want    *OtelConfig
		wantErr bool
		errMsg  string
	}{
		{
			name: "nil config - non-strict",
			cfg:  nil,
			want: &OtelConfig{
				Insecure: false,
				Headers:  map[string]string{},
				Traces:   &FeatureConfig{Enabled: false},
				Metrics:  &FeatureConfig{Enabled: false},
				Logs:     &FeatureConfig{Enabled: false},
			},
			wantErr: false,
		},
		{
			name:   "nil config - strict",
			cfg:    nil,
			strict: true,
			want: &OtelConfig{
				Insecure: false,
				Headers:  map[string]string{},
				Traces:   &FeatureConfig{Enabled: false},
				Metrics:  &FeatureConfig{Enabled: false},
				Logs:     &FeatureConfig{Enabled: false},
			},
			wantErr: false,
		},
		{
			name: "full valid config",
			cfg: map[string]any{
				"endpoint": "http://localhost:4317",
				"insecure": true,
				"headers": map[string]any{
					"authorization": "Bearer token",
				},
				"resource_attributes": []any{
					map[string]any{"key": "service.name", "value": "my-service"},
					map[string]any{"key": "env", "value": "production"},
				},
				"traces":  map[string]any{"enabled": true},
				"metrics": map[string]any{"enabled": true},
				"logs":    map[string]any{"enabled": false},
			},
			want: &OtelConfig{
				Endpoint: "http://localhost:4317",
				Insecure: true,
				Headers:  map[string]string{"authorization": "Bearer token"},
				ResourceAttributes: []ResourceAttribute{
					{Key: "service.name", Value: "my-service"},
					{Key: "env", Value: "production"},
				},
				Traces:  &FeatureConfig{Enabled: true},
				Metrics: &FeatureConfig{Enabled: true},
				Logs:    &FeatureConfig{Enabled: false},
			},
			wantErr: false,
		},
		{
			name: "minimal config",
			cfg: map[string]any{
				"endpoint": "http://localhost:4317",
			},
			want: &OtelConfig{
				Endpoint: "http://localhost:4317",
				Insecure: false,
				Headers:  map[string]string{},
				Traces:   &FeatureConfig{Enabled: false},
				Metrics:  &FeatureConfig{Enabled: false},
				Logs:     &FeatureConfig{Enabled: false},
			},
			wantErr: false,
		},
		{
			name: "invalid endpoint type - strict",
			cfg: map[string]any{
				"endpoint": 123,
			},
			strict:  true,
			wantErr: true,
			errMsg:  "invalid type for 'endpoint': expected string, got int",
		},
		{
			name: "invalid insecure type - strict",
			cfg: map[string]any{
				"insecure": "true",
			},
			strict:  true,
			wantErr: true,
			errMsg:  "invalid type for 'insecure': expected bool, got string",
		},
		{
			name: "invalid headers type - strict",
			cfg: map[string]any{
				"headers": "invalid",
			},
			strict:  true,
			wantErr: true,
			errMsg:  "invalid type for 'headers': expected map[string]any, got string",
		},
		{
			name: "invalid header value type - strict",
			cfg: map[string]any{
				"headers": map[string]any{
					"key": 123,
				},
			},
			strict:  true,
			wantErr: true,
			errMsg:  "invalid type for header 'key': expected string, got int",
		},
		{
			name: "invalid resource_attributes type - strict",
			cfg: map[string]any{
				"resource_attributes": "invalid",
			},
			strict:  true,
			wantErr: true,
			errMsg:  "invalid type for 'resource_attributes': expected []any, got string",
		},
		{
			name: "invalid resource attribute item type - strict",
			cfg: map[string]any{
				"resource_attributes": []any{"invalid"},
			},
			strict:  true,
			wantErr: true,
			errMsg:  "invalid type for resource_attributes[0]: expected map[string]any, got string",
		},
		{
			name: "invalid resource attribute key type - strict",
			cfg: map[string]any{
				"resource_attributes": []any{
					map[string]any{"key": 123, "value": "val"},
				},
			},
			strict:  true,
			wantErr: true,
			errMsg:  "invalid type for resource_attributes[0].key: expected string, got int",
		},
		{
			name: "invalid resource attribute value type - strict",
			cfg: map[string]any{
				"resource_attributes": []any{
					map[string]any{"key": "key", "value": 123},
				},
			},
			strict:  true,
			wantErr: true,
			errMsg:  "invalid type for resource_attributes[0].value: expected string, got int",
		},
		{
			name: "invalid traces type - strict",
			cfg: map[string]any{
				"traces": "invalid",
			},
			strict:  true,
			wantErr: true,
			errMsg:  "invalid type for 'traces': expected map[string]any, got string",
		},
		{
			name: "invalid traces.enabled type - strict",
			cfg: map[string]any{
				"traces": map[string]any{"enabled": "true"},
			},
			strict:  true,
			wantErr: true,
			errMsg:  "invalid type for traces.enabled: expected bool, got string",
		},
		{
			name: "invalid metrics type - strict",
			cfg: map[string]any{
				"metrics": "invalid",
			},
			strict:  true,
			wantErr: true,
			errMsg:  "invalid type for 'metrics': expected map[string]any, got string",
		},
		{
			name: "invalid logs type - strict",
			cfg: map[string]any{
				"logs": "invalid",
			},
			strict:  true,
			wantErr: true,
			errMsg:  "invalid type for 'logs': expected map[string]any, got string",
		},
		{
			name: "unknown field - strict",
			cfg: map[string]any{
				"unknown_field": "value",
			},
			strict:  true,
			wantErr: true,
			errMsg:  "unknown config field: unknown_field",
		},
		{
			name: "unknown field - non-strict",
			cfg: map[string]any{
				"unknown_field": "value",
				"endpoint":      "http://localhost:4317",
			},
			want: &OtelConfig{
				Endpoint: "http://localhost:4317",
				Insecure: false,
				Headers:  map[string]string{},
				Traces:   &FeatureConfig{Enabled: false},
				Metrics:  &FeatureConfig{Enabled: false},
				Logs:     &FeatureConfig{Enabled: false},
			},
			wantErr: false,
		},
		{
			name: "invalid endpoint type - non-strict",
			cfg: map[string]any{
				"endpoint": 123,
			},
			want: &OtelConfig{
				Insecure: false,
				Headers:  map[string]string{},
				Traces:   &FeatureConfig{Enabled: false},
				Metrics:  &FeatureConfig{Enabled: false},
				Logs:     &FeatureConfig{Enabled: false},
			},
			wantErr: false,
		},
		{
			name: "empty resource_attributes list",
			cfg: map[string]any{
				"resource_attributes": []any{},
			},
			want: &OtelConfig{
				Insecure:           false,
				Headers:            map[string]string{},
				ResourceAttributes: []ResourceAttribute{},
				Traces:             &FeatureConfig{Enabled: false},
				Metrics:            &FeatureConfig{Enabled: false},
				Logs:               &FeatureConfig{Enabled: false},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got *OtelConfig
			var err error

			if tt.strict {
				got, err = ParseOtelConfigFromMap(tt.cfg, true)
			} else {
				got, err = ParseOtelConfigFromMap(tt.cfg)
			}

			if (err != nil) != tt.wantErr {
				t.Errorf("ParseOtelConfigFromMap() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && err.Error() != tt.errMsg {
				t.Errorf("ParseOtelConfigFromMap() error message = %v, want %v", err.Error(), tt.errMsg)
			}
			if !tt.wantErr && !compareOtelConfig(got, tt.want) {
				t.Errorf("ParseOtelConfigFromMap() got = %+v, want %+v", got, tt.want)
			}
		})
	}
}

// Helper function to compare OtelConfig
func compareOtelConfig(a, b *OtelConfig) bool {
	if a.Endpoint != b.Endpoint {
		return false
	}
	if a.Insecure != b.Insecure {
		return false
	}
	if len(a.Headers) != len(b.Headers) {
		return false
	}
	for k, v := range a.Headers {
		if b.Headers[k] != v {
			return false
		}
	}
	if len(a.ResourceAttributes) != len(b.ResourceAttributes) {
		return false
	}
	for i, ra := range a.ResourceAttributes {
		if ra.Key != b.ResourceAttributes[i].Key || ra.Value != b.ResourceAttributes[i].Value {
			return false
		}
	}
	if a.Traces.Enabled != b.Traces.Enabled {
		return false
	}
	if a.Metrics.Enabled != b.Metrics.Enabled {
		return false
	}
	if a.Logs.Enabled != b.Logs.Enabled {
		return false
	}
	return true
}
