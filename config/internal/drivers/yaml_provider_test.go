package drivers

import (
	"os"
	"path/filepath"
	"testing"
)

// Helper function to create a temporary YAML file
func createTempYAMLFile(t *testing.T, content string) string {
	t.Helper()
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "config.yaml")

	err := os.WriteFile(filePath, []byte(content), 0644)
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}

	return filePath
}

func TestNewYamlConfigProvider_Success(t *testing.T) {
	yamlContent := `
name: test-app
port: 8080
enabled: true
database:
  host: localhost
  port: 3306
  credentials:
    username: admin
    password: secret
servers:
  - host: server1
    port: 8001
  - host: server2
    port: 8002
`

	filePath := createTempYAMLFile(t, yamlContent)

	provider, err := NewYamlConfigProvider(filePath)
	if err != nil {
		t.Fatalf("NewYamlConfigProvider() error = %v", err)
	}

	if provider == nil {
		t.Fatal("NewYamlConfigProvider() returned nil")
	}

	// Verify provider is YamlConfigProvider type
	yamlProvider, ok := provider.(*YamlConfigProvider)
	if !ok {
		t.Fatal("provider is not *YamlConfigProvider")
	}

	if yamlProvider.base == nil {
		t.Error("provider.base is nil")
	}
}

func TestNewYamlConfigProvider_FileNotFound(t *testing.T) {
	_, err := NewYamlConfigProvider("/nonexistent/path/config.yaml")

	if err == nil {
		t.Error("expected error for non-existent file, got nil")
	}

	if !contains(err.Error(), "failed to read yaml file") {
		t.Errorf("expected file read error, got: %v", err)
	}
}

func TestNewYamlConfigProvider_InvalidYAML(t *testing.T) {
	invalidYAML := `
name: test
port: 8080
- invalid array item in object
`

	filePath := createTempYAMLFile(t, invalidYAML)

	_, err := NewYamlConfigProvider(filePath)

	if err == nil {
		t.Error("expected error for invalid YAML, got nil")
	}

	if !contains(err.Error(), "failed to parse yaml") {
		t.Errorf("expected parse error, got: %v", err)
	}
}

func TestNewYamlConfigProvider_EmptyFile(t *testing.T) {
	filePath := createTempYAMLFile(t, "")

	provider, err := NewYamlConfigProvider(filePath)
	if err != nil {
		t.Fatalf("expected empty file to be valid, got error: %v", err)
	}

	if provider == nil {
		t.Fatal("provider should not be nil for empty file")
	}

	// Empty file creates empty config, so empty key should return empty map
	result, err := provider.Get("")
	if err != nil {
		t.Errorf("Get('') on empty config error = %v", err)
	}
	if result == nil {
		t.Error("Get('') should return empty map, not nil")
	}
}

func TestYamlConfigProvider_Get_SimpleValues(t *testing.T) {
	yamlContent := `
string: value
number: 42
float: 3.14
bool: true
negative: -10
`

	filePath := createTempYAMLFile(t, yamlContent)

	provider, err := NewYamlConfigProvider(filePath)
	if err != nil {
		t.Fatalf("NewYamlConfigProvider() error = %v", err)
	}

	tests := []struct {
		name  string
		key   string
		want  any
	}{
		{"string value", "string", "value"},
		{"number value", "number", 42},
		{"float value", "float", 3.14},
		{"bool value", "bool", true},
		{"negative number", "negative", -10},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := provider.Get(tt.key)
			if err != nil {
				t.Errorf("Get(%q) error = %v", tt.key, err)
				return
			}
			if got != tt.want {
				t.Errorf("Get(%q) = %v, want %v", tt.key, got, tt.want)
			}
		})
	}
}

func TestYamlConfigProvider_Get_NestedValues(t *testing.T) {
	yamlContent := `
database:
  host: localhost
  port: 3306
  credentials:
    username: admin
    password: secret
`

	filePath := createTempYAMLFile(t, yamlContent)

	provider, err := NewYamlConfigProvider(filePath)
	if err != nil {
		t.Fatalf("NewYamlConfigProvider() error = %v", err)
	}

	tests := []struct {
		name  string
		key   string
		want  any
	}{
		{"nested host", "database.host", "localhost"},
		{"nested port", "database.port", 3306},
		{"deep nested username", "database.credentials.username", "admin"},
		{"deep nested password", "database.credentials.password", "secret"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := provider.Get(tt.key)
			if err != nil {
				t.Errorf("Get(%q) error = %v", tt.key, err)
				return
			}
			if got != tt.want {
				t.Errorf("Get(%q) = %v, want %v", tt.key, got, tt.want)
			}
		})
	}
}

func TestYamlConfigProvider_Get_ArrayValues(t *testing.T) {
	yamlContent := `
servers:
  - name: server1
    port: 8001
  - name: server2
    port: 8002
  - name: server3
    port: 8003
tags:
  - web
  - api
  - microservice
numbers:
  - 1
  - 2
  - 3
  - 4
  - 5
`

	filePath := createTempYAMLFile(t, yamlContent)

	provider, err := NewYamlConfigProvider(filePath)
	if err != nil {
		t.Fatalf("NewYamlConfigProvider() error = %v", err)
	}

	tests := []struct {
		name  string
		key   string
		want  any
	}{
		{"array first element", "servers[0].name", "server1"},
		{"array second element", "servers[1].port", 8002},
		{"array last element", "servers[2].name", "server3"},
		{"simple array", "tags[0]", "web"},
		{"simple array middle", "tags[1]", "api"},
		{"numbers array", "numbers[3]", 4},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := provider.Get(tt.key)
			if err != nil {
				t.Errorf("Get(%q) error = %v", tt.key, err)
				return
			}
			if got != tt.want {
				t.Errorf("Get(%q) = %v, want %v", tt.key, got, tt.want)
			}
		})
	}
}

func TestYamlConfigProvider_Get_EmptyKey(t *testing.T) {
	yamlContent := `
name: test
port: 8080
`

	filePath := createTempYAMLFile(t, yamlContent)

	provider, err := NewYamlConfigProvider(filePath)
	if err != nil {
		t.Fatalf("NewYamlConfigProvider() error = %v", err)
	}

	result, err := provider.Get("")
	if err != nil {
		t.Errorf("Get('') error = %v", err)
	}

	if result == nil {
		t.Error("Get('') returned nil")
	}

	resultMap, ok := result.(map[string]any)
	if !ok {
		t.Error("Get('') did not return map[string]any")
	}

	if resultMap["name"] != "test" {
		t.Errorf("Get('')['name'] = %v, want 'test'", resultMap["name"])
	}
}

func TestYamlConfigProvider_Get_NotFound(t *testing.T) {
	yamlContent := `
existing: value
`

	filePath := createTempYAMLFile(t, yamlContent)

	provider, err := NewYamlConfigProvider(filePath)
	if err != nil {
		t.Fatalf("NewYamlConfigProvider() error = %v", err)
	}

	_, err = provider.Get("nonexistent")
	if err == nil {
		t.Error("Get('nonexistent') expected error, got nil")
	}
}

func TestYamlConfigProvider_Has(t *testing.T) {
	yamlContent := `
existing: value
nested:
  key: value
servers:
  - host: server1
`

	filePath := createTempYAMLFile(t, yamlContent)

	provider, err := NewYamlConfigProvider(filePath)
	if err != nil {
		t.Fatalf("NewYamlConfigProvider() error = %v", err)
	}

	tests := []struct {
		name     string
		key      string
		expected bool
	}{
		{"existing key", "existing", true},
		{"existing nested key", "nested.key", true},
		{"existing array element", "servers[0].host", true},
		{"non-existing key", "nonexistent", false},
		{"non-existing nested", "nested.nonexistent", false},
		{"empty key", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := provider.Has(tt.key)
			if got != tt.expected {
				t.Errorf("Has(%q) = %v, want %v", tt.key, got, tt.expected)
			}
		})
	}
}

func TestYamlConfigProvider_ComplexStructure(t *testing.T) {
	yamlContent := `
app:
  name: myapp
  version: 1.0.0
  environments:
    production:
      - host: prod1.example.com
        port: 443
        ssl: true
      - host: prod2.example.com
        port: 443
        ssl: true
    development:
      - host: localhost
        port: 8080
        ssl: false
`

	filePath := createTempYAMLFile(t, yamlContent)

	provider, err := NewYamlConfigProvider(filePath)
	if err != nil {
		t.Fatalf("NewYamlConfigProvider() error = %v", err)
	}

	tests := []struct {
		name  string
		key   string
		want  any
	}{
		{"app name", "app.name", "myapp"},
		{"app version", "app.version", "1.0.0"},
		{"prod first host", "app.environments.production[0].host", "prod1.example.com"},
		{"prod second port", "app.environments.production[1].port", 443},
		{"dev ssl", "app.environments.development[0].ssl", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := provider.Get(tt.key)
			if err != nil {
				t.Errorf("Get(%q) error = %v", tt.key, err)
				return
			}
			if got != tt.want {
				t.Errorf("Get(%q) = %v, want %v", tt.key, got, tt.want)
			}
		})
	}
}

func TestYamlConfigProvider_MultiLineStrings(t *testing.T) {
	yamlContent := `
single_line: simple value
literal_block: |
  This is a literal block
  preserving newlines
  and indentation
folded_block: >
  This is a folded block
  newlines are converted
  to spaces
`

	filePath := createTempYAMLFile(t, yamlContent)

	provider, err := NewYamlConfigProvider(filePath)
	if err != nil {
		t.Fatalf("NewYamlConfigProvider() error = %v", err)
	}

	tests := []struct {
		name string
		key  string
	}{
		{"single line", "single_line"},
		{"literal block", "literal_block"},
		{"folded block", "folded_block"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := provider.Get(tt.key)
			if err != nil {
				t.Errorf("Get(%q) error = %v", tt.key, err)
				return
			}
			if got == nil {
				t.Errorf("Get(%q) returned nil", tt.key)
			}
		})
	}
}

func TestYamlConfigProvider_NullValues(t *testing.T) {
	yamlContent := `
explicit_null: null
empty_string: ""
unset: ~
`

	filePath := createTempYAMLFile(t, yamlContent)

	provider, err := NewYamlConfigProvider(filePath)
	if err != nil {
		t.Fatalf("NewYamlConfigProvider() error = %v", err)
	}

	// Test that null value can be retrieved
	got, err := provider.Get("explicit_null")
	if err != nil {
		t.Errorf("Get('explicit_null') error = %v", err)
	}
	if got != nil {
		t.Errorf("Get('explicit_null') = %v, want nil", got)
	}

	// Test empty string is different from null
	got, err = provider.Get("empty_string")
	if err != nil {
		t.Errorf("Get('empty_string') error = %v", err)
	}
	if got != "" {
		t.Errorf("Get('empty_string') = %v, want ''", got)
	}
}

func TestYamlConfigProvider_AnchorAndAlias(t *testing.T) {
	yamlContent := `
defaults: &defaults
  timeout: 30
  retries: 3

production:
  <<: *defaults
  host: prod.example.com

development:
  <<: *defaults
  host: localhost
  timeout: 60
`

	filePath := createTempYAMLFile(t, yamlContent)

	provider, err := NewYamlConfigProvider(filePath)
	if err != nil {
		t.Fatalf("NewYamlConfigProvider() error = %v", err)
	}

	// Test that production inherited timeout
	got, err := provider.Get("production.timeout")
	if err != nil {
		t.Errorf("Get('production.timeout') error = %v", err)
	}
	if got != 30 {
		t.Errorf("Get('production.timeout') = %v, want 30", got)
	}

	// Test that development overrode timeout
	got, err = provider.Get("development.timeout")
	if err != nil {
		t.Errorf("Get('development.timeout') error = %v", err)
	}
	if got != 60 {
		t.Errorf("Get('development.timeout') = %v, want 60", got)
	}
}

func TestYamlConfigProvider_SpecialCharacters(t *testing.T) {
	yamlContent := `
unicode: Hello 世界
emoji: "Hello 👋"
special_chars: "test@#$%"
quotes: 'He said "hello"'
`

	filePath := createTempYAMLFile(t, yamlContent)

	provider, err := NewYamlConfigProvider(filePath)
	if err != nil {
		t.Fatalf("NewYamlConfigProvider() error = %v", err)
	}

	tests := []struct {
		name string
		key  string
		want any
	}{
		{"unicode", "unicode", "Hello 世界"},
		{"emoji", "emoji", "Hello 👋"},
		{"special chars", "special_chars", "test@#$%"},
		{"quotes", "quotes", `He said "hello"`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := provider.Get(tt.key)
			if err != nil {
				t.Errorf("Get(%q) error = %v", tt.key, err)
				return
			}
			if got != tt.want {
				t.Errorf("Get(%q) = %v, want %v", tt.key, got, tt.want)
			}
		})
	}
}

func TestYamlConfigProvider_YAML1_2Types(t *testing.T) {
	yamlContent := `
binary: !!binary SGVsbG8gV29ybGQ=
timestamp: "2024-01-15T10:30:00Z"
infinity: .inf
negative_infinity: -.inf
not_a_number: .nan
`

	filePath := createTempYAMLFile(t, yamlContent)

	provider, err := NewYamlConfigProvider(filePath)
	if err != nil {
		t.Fatalf("NewYamlConfigProvider() error = %v", err)
	}

	// Just verify these keys can be retrieved (type handling depends on yaml.v3)
	keys := []string{"binary", "timestamp", "infinity", "negative_infinity", "not_a_number"}
	for _, key := range keys {
		t.Run(key, func(t *testing.T) {
			got, err := provider.Get(key)
			if err != nil {
				t.Errorf("Get(%q) error = %v", key, err)
			}
			if got == nil {
				t.Errorf("Get(%q) returned nil", key)
			}
		})
	}
}
