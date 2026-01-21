package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/lite-lake/litecore-go/common"
)

// Helper function to create a temporary JSON file
func createTempFile(t *testing.T, filename string, content string) string {
	t.Helper()
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, filename)

	err := os.WriteFile(filePath, []byte(content), 0644)
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}

	return filePath
}

func TestNewConfigProvider_JsonDriver(t *testing.T) {
	jsonContent := `{
		"name": "test",
		"port": 8080
	}`

	filePath := createTempFile(t, "config.json", jsonContent)

	provider, err := NewConfigProvider("json", filePath)
	if err != nil {
		t.Fatalf("NewConfigProvider() error = %v", err)
	}

	if provider == nil {
		t.Fatal("NewConfigProvider() returned nil")
	}

	// Verify it's working
	name, err := provider.Get("name")
	if err != nil {
		t.Errorf("provider.Get('name') error = %v", err)
	}
	if name != "test" {
		t.Errorf("provider.Get('name') = %v, want 'test'", name)
	}
}

func TestNewConfigProvider_YamlDriver(t *testing.T) {
	yamlContent := `
name: test
port: 8080
`

	filePath := createTempFile(t, "config.yaml", yamlContent)

	provider, err := NewConfigProvider("yaml", filePath)
	if err != nil {
		t.Fatalf("NewConfigProvider() error = %v", err)
	}

	if provider == nil {
		t.Fatal("NewConfigProvider() returned nil")
	}

	// Verify it's working
	name, err := provider.Get("name")
	if err != nil {
		t.Errorf("provider.Get('name') error = %v", err)
	}
	if name != "test" {
		t.Errorf("provider.Get('name') = %v, want 'test'", name)
	}
}

func TestNewConfigProvider_UnsupportedDriver(t *testing.T) {
	filePath := createTempFile(t, "config.txt", "some content")

	_, err := NewConfigProvider("xml", filePath)
	if err == nil {
		t.Error("expected error for unsupported driver, got nil")
	}

	if err.Error() != "unsupported driver" {
		t.Errorf("expected 'unsupported driver' error, got: %v", err)
	}
}

func TestNewConfigProvider_JsonDriver_FileNotFound(t *testing.T) {
	_, err := NewConfigProvider("json", "/nonexistent/path/config.json")
	if err == nil {
		t.Error("expected error for non-existent file, got nil")
	}
}

func TestNewConfigProvider_YamlDriver_FileNotFound(t *testing.T) {
	_, err := NewConfigProvider("yaml", "/nonexistent/path/config.yaml")
	if err == nil {
		t.Error("expected error for non-existent file, got nil")
	}
}

func TestNewConfigProvider_JsonDriver_InvalidContent(t *testing.T) {
	invalidJSON := `{
		"name": "test",
		invalid
	}`

	filePath := createTempFile(t, "config.json", invalidJSON)

	_, err := NewConfigProvider("json", filePath)
	if err == nil {
		t.Error("expected error for invalid JSON, got nil")
	}
}

func TestNewConfigProvider_YamlDriver_InvalidContent(t *testing.T) {
	invalidYAML := `
name: test
port: 8080
- invalid array item
`

	filePath := createTempFile(t, "config.yaml", invalidYAML)

	_, err := NewConfigProvider("yaml", filePath)
	if err == nil {
		t.Error("expected error for invalid YAML, got nil")
	}
}

func TestNewConfigProvider_DriverCaseSensitivity(t *testing.T) {
	jsonContent := `{"name": "test"}`
	yamlContent := `name: test`

	jsonPath := createTempFile(t, "config.json", jsonContent)
	yamlPath := createTempFile(t, "config.yaml", yamlContent)

	tests := []struct {
		name        string
		driver      string
		filePath    string
		expectError bool
	}{
		{"lowercase json", "json", jsonPath, false},
		{"lowercase yaml", "yaml", yamlPath, false},
		{"uppercase JSON", "JSON", jsonPath, true},
		{"uppercase YAML", "YAML", yamlPath, true},
		{"mixed case Json", "Json", jsonPath, true},
		{"mixed case Yaml", "Yaml", yamlPath, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewConfigProvider(tt.driver, tt.filePath)
			if (err != nil) != tt.expectError {
				t.Errorf("NewConfigProvider() error = %v, expectError %v", err, tt.expectError)
			}
		})
	}
}

func TestNewConfigProvider_EmptyDriver(t *testing.T) {
	filePath := createTempFile(t, "config.json", `{"name": "test"}`)

	_, err := NewConfigProvider("", filePath)
	if err == nil {
		t.Error("expected error for empty driver, got nil")
	}

	if err.Error() != "unsupported driver" {
		t.Errorf("expected 'unsupported driver' error, got: %v", err)
	}
}

func TestNewConfigProvider_VerifyInterface(t *testing.T) {
	jsonContent := `{"name": "test", "port": 8080}`
	yamlContent := `name: test
port: 8080`

	jsonPath := createTempFile(t, "config.json", jsonContent)
	yamlPath := createTempFile(t, "config.yaml", yamlContent)

	tests := []struct {
		name   string
		driver string
		path   string
	}{
		{"json provider", "json", jsonPath},
		{"yaml provider", "yaml", yamlPath},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider, err := NewConfigProvider(tt.driver, tt.path)
			if err != nil {
				t.Fatalf("NewConfigProvider() error = %v", err)
			}

			// Verify it implements BaseConfigProvider interface
			var _ common.IBaseConfigProvider = provider

			// Test Get method
			name, err := provider.Get("name")
			if err != nil {
				t.Errorf("provider.Get('name') error = %v", err)
			}
			if name != "test" {
				t.Errorf("provider.Get('name') = %v, want 'test'", name)
			}

			// Test Has method
			if !provider.Has("name") {
				t.Error("provider.Has('name') returned false, expected true")
			}
			if provider.Has("nonexistent") {
				t.Error("provider.Has('nonexistent') returned true, expected false")
			}
		})
	}
}

func TestNewConfigProvider_ComplexConfiguration(t *testing.T) {
	jsonContent := `{
		"app": {
			"name": "myapp",
			"version": "1.0.0",
			"servers": [
				{"host": "localhost", "port": 8080},
				{"host": "0.0.0.0", "port": 8081}
			]
		}
	}`

	filePath := createTempFile(t, "config.json", jsonContent)

	provider, err := NewConfigProvider("json", filePath)
	if err != nil {
		t.Fatalf("NewConfigProvider() error = %v", err)
	}

	// Test nested access
	appName, err := provider.Get("app.name")
	if err != nil {
		t.Errorf("provider.Get('app.name') error = %v", err)
	}
	if appName != "myapp" {
		t.Errorf("provider.Get('app.name') = %v, want 'myapp'", appName)
	}

	// Test array access
	host, err := provider.Get("app.servers[0].host")
	if err != nil {
		t.Errorf("provider.Get('app.servers[0].host') error = %v", err)
	}
	if host != "localhost" {
		t.Errorf("provider.Get('app.servers[0].host') = %v, want 'localhost'", host)
	}
}

func TestNewConfigProvider_MultipleInstances(t *testing.T) {
	jsonContent1 := `{"name": "app1", "port": 8080}`
	jsonContent2 := `{"name": "app2", "port": 8081}`

	filePath1 := createTempFile(t, "config1.json", jsonContent1)
	filePath2 := createTempFile(t, "config2.json", jsonContent2)

	provider1, err := NewConfigProvider("json", filePath1)
	if err != nil {
		t.Fatalf("NewConfigProvider() error = %v", err)
	}

	provider2, err := NewConfigProvider("json", filePath2)
	if err != nil {
		t.Fatalf("NewConfigProvider() error = %v", err)
	}

	// Verify providers are independent
	name1, _ := provider1.Get("name")
	name2, _ := provider2.Get("name")

	if name1 != "app1" {
		t.Errorf("provider1.Get('name') = %v, want 'app1'", name1)
	}
	if name2 != "app2" {
		t.Errorf("provider2.Get('name') = %v, want 'app2'", name2)
	}
}
