package drivers

import (
	"os"
	"path/filepath"
	"testing"
)

// Helper function to create a temporary JSON file
func createTempJSONFile(t *testing.T, content string) string {
	t.Helper()
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "config.json")

	err := os.WriteFile(filePath, []byte(content), 0644)
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}

	return filePath
}

func TestNewJsonConfigProvider_Success(t *testing.T) {
	jsonContent := `{
		"name": "test-app",
		"port": 8080,
		"enabled": true,
		"database": {
			"host": "localhost",
			"port": 3306,
			"credentials": {
				"username": "admin",
				"password": "secret"
			}
		},
		"servers": [
			{"host": "server1", "port": 8001},
			{"host": "server2", "port": 8002}
		]
	}`

	filePath := createTempJSONFile(t, jsonContent)

	provider, err := NewJsonConfigProvider(filePath)
	if err != nil {
		t.Fatalf("NewJsonConfigProvider() error = %v", err)
	}

	if provider == nil {
		t.Fatal("NewJsonConfigProvider() returned nil")
	}

	// Verify provider is JsonConfigProvider type
	jsonProvider, ok := provider.(*JsonConfigProvider)
	if !ok {
		t.Fatal("provider is not *JsonConfigProvider")
	}

	if jsonProvider.base == nil {
		t.Error("provider.base is nil")
	}
}

func TestNewJsonConfigProvider_FileNotFound(t *testing.T) {
	_, err := NewJsonConfigProvider("/nonexistent/path/config.json")

	if err == nil {
		t.Error("expected error for non-existent file, got nil")
	}

	if !contains(err.Error(), "failed to read json file") {
		t.Errorf("expected file read error, got: %v", err)
	}
}

func TestNewJsonConfigProvider_InvalidJSON(t *testing.T) {
	invalidJSON := `{
		"name": "test",
		"port": 8080,
	}` // Missing closing brace

	filePath := createTempJSONFile(t, invalidJSON)

	_, err := NewJsonConfigProvider(filePath)

	if err == nil {
		t.Error("expected error for invalid JSON, got nil")
	}

	if !contains(err.Error(), "failed to parse json") {
		t.Errorf("expected parse error, got: %v", err)
	}
}

func TestNewJsonConfigProvider_EmptyFile(t *testing.T) {
	filePath := createTempJSONFile(t, "")

	_, err := NewJsonConfigProvider(filePath)

	if err == nil {
		t.Error("expected error for empty file, got nil")
	}
}

func TestNewJsonConfigProvider_NonObjectJSON(t *testing.T) {
	nonObjectJSON := `["array", "values"]`

	filePath := createTempJSONFile(t, nonObjectJSON)

	_, err := NewJsonConfigProvider(filePath)

	if err == nil {
		t.Error("expected error for non-object JSON, got nil")
	}
}

func TestJsonConfigProvider_Get_SimpleValues(t *testing.T) {
	jsonContent := `{
		"string": "value",
		"number": 42,
		"float": 3.14,
		"bool": true,
		"null": null
	}`

	filePath := createTempJSONFile(t, jsonContent)

	provider, err := NewJsonConfigProvider(filePath)
	if err != nil {
		t.Fatalf("NewJsonConfigProvider() error = %v", err)
	}

	tests := []struct {
		name string
		key  string
		want any
	}{
		{"string value", "string", "value"},
		{"number value", "number", 42.0},
		{"float value", "float", 3.14},
		{"bool value", "bool", true},
		{"null value", "null", nil},
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

func TestJsonConfigProvider_Get_NestedValues(t *testing.T) {
	jsonContent := `{
		"database": {
			"host": "localhost",
			"port": 3306,
			"credentials": {
				"username": "admin",
				"password": "secret"
			}
		}
	}`

	filePath := createTempJSONFile(t, jsonContent)

	provider, err := NewJsonConfigProvider(filePath)
	if err != nil {
		t.Fatalf("NewJsonConfigProvider() error = %v", err)
	}

	tests := []struct {
		name string
		key  string
		want any
	}{
		{"nested host", "database.host", "localhost"},
		{"nested port", "database.port", 3306.0},
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

func TestJsonConfigProvider_Get_ArrayValues(t *testing.T) {
	jsonContent := `{
		"servers": [
			{"name": "server1", "port": 8001},
			{"name": "server2", "port": 8002},
			{"name": "server3", "port": 8003}
		],
		"tags": ["web", "api", "microservice"],
		"numbers": [1, 2, 3, 4, 5]
	}`

	filePath := createTempJSONFile(t, jsonContent)

	provider, err := NewJsonConfigProvider(filePath)
	if err != nil {
		t.Fatalf("NewJsonConfigProvider() error = %v", err)
	}

	tests := []struct {
		name string
		key  string
		want any
	}{
		{"array first element", "servers[0].name", "server1"},
		{"array second element", "servers[1].port", 8002.0},
		{"array last element", "servers[2].name", "server3"},
		{"simple array", "tags[0]", "web"},
		{"simple array middle", "tags[1]", "api"},
		{"numbers array", "numbers[3]", 4.0},
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

func TestJsonConfigProvider_Get_EmptyKey(t *testing.T) {
	jsonContent := `{
		"name": "test",
		"port": 8080
	}`

	filePath := createTempJSONFile(t, jsonContent)

	provider, err := NewJsonConfigProvider(filePath)
	if err != nil {
		t.Fatalf("NewJsonConfigProvider() error = %v", err)
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

func TestJsonConfigProvider_Get_NotFound(t *testing.T) {
	jsonContent := `{
		"existing": "value"
	}`

	filePath := createTempJSONFile(t, jsonContent)

	provider, err := NewJsonConfigProvider(filePath)
	if err != nil {
		t.Fatalf("NewJsonConfigProvider() error = %v", err)
	}

	_, err = provider.Get("nonexistent")
	if err == nil {
		t.Error("Get('nonexistent') expected error, got nil")
	}
}

func TestJsonConfigProvider_Has(t *testing.T) {
	jsonContent := `{
		"existing": "value",
		"nested": {
			"key": "value"
		},
		"servers": [
			{"host": "server1"}
		]
	}`

	filePath := createTempJSONFile(t, jsonContent)

	provider, err := NewJsonConfigProvider(filePath)
	if err != nil {
		t.Fatalf("NewJsonConfigProvider() error = %v", err)
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

func TestJsonConfigProvider_ComplexStructure(t *testing.T) {
	jsonContent := `{
		"app": {
			"name": "myapp",
			"version": "1.0.0",
			"environments": {
				"production": [
					{
						"host": "prod1.example.com",
						"port": 443,
						"ssl": true
					},
					{
						"host": "prod2.example.com",
						"port": 443,
						"ssl": true
					}
				],
				"development": [
					{
						"host": "localhost",
						"port": 8080,
						"ssl": false
					}
				]
			}
		}
	}`

	filePath := createTempJSONFile(t, jsonContent)

	provider, err := NewJsonConfigProvider(filePath)
	if err != nil {
		t.Fatalf("NewJsonConfigProvider() error = %v", err)
	}

	tests := []struct {
		name string
		key  string
		want any
	}{
		{"app name", "app.name", "myapp"},
		{"app version", "app.version", "1.0.0"},
		{"prod first host", "app.environments.production[0].host", "prod1.example.com"},
		{"prod second port", "app.environments.production[1].port", 443.0},
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

func TestJsonConfigProvider_WhitespaceHandling(t *testing.T) {
	jsonContent := `{
		"key1": "value1",
		"key2":     "value2"    ,
		"nested": {
			"key3":    "value3"
		}
	}`

	filePath := createTempJSONFile(t, jsonContent)

	provider, err := NewJsonConfigProvider(filePath)
	if err != nil {
		t.Fatalf("NewJsonConfigProvider() error = %v", err)
	}

	tests := []struct {
		name string
		key  string
		want any
	}{
		{"key1", "key1", "value1"},
		{"key2", "key2", "value2"},
		{"nested key3", "nested.key3", "value3"},
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

func TestJsonConfigProvider_SpecialCharacters(t *testing.T) {
	jsonContent := `{
		"unicode": "Hello 世界",
		"escaped": "Line1\nLine2\tTabbed",
		"quote": "He said \"hello\"",
		"backslash": "path\\to\\file"
	}`

	filePath := createTempJSONFile(t, jsonContent)

	provider, err := NewJsonConfigProvider(filePath)
	if err != nil {
		t.Fatalf("NewJsonConfigProvider() error = %v", err)
	}

	tests := []struct {
		name string
		key  string
		want any
	}{
		{"unicode", "unicode", "Hello 世界"},
		{"escaped", "escaped", "Line1\nLine2\tTabbed"},
		{"quote", "quote", `He said "hello"`},
		{"backslash", "backslash", `path\to\file`},
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
