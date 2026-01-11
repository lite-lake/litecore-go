package drivers

import (
	"testing"
)

func TestNewBaseConfigProvider(t *testing.T) {
	data := map[string]any{
		"key1": "value1",
		"key2": 123,
	}

	provider := NewBaseConfigProvider(data)

	if provider == nil {
		t.Fatal("NewBaseConfigProvider returned nil")
	}

	if provider.configData == nil {
		t.Error("configData is nil")
	}

	if provider.configData["key1"] != "value1" {
		t.Errorf("expected key1 to be 'value1', got %v", provider.configData["key1"])
	}
}

func TestBaseConfigProvider_Get_EmptyKey(t *testing.T) {
	data := map[string]any{
		"key1": "value1",
	}
	provider := NewBaseConfigProvider(data)

	result, err := provider.Get("")

	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	if result == nil {
		t.Error("expected result to be configData, got nil")
	}

	resultMap, ok := result.(map[string]any)
	if !ok {
		t.Error("expected result to be map[string]any")
	}

	if resultMap["key1"] != "value1" {
		t.Errorf("expected key1 to be 'value1', got %v", resultMap["key1"])
	}
}

func TestBaseConfigProvider_Get_SimpleKey(t *testing.T) {
	data := map[string]any{
		"name":     "test",
		"port":     8080,
		"enabled":  true,
		"count":    42,
		"floatNum": 3.14,
	}
	provider := NewBaseConfigProvider(data)

	tests := []struct {
		name string
		key  string
		want any
	}{
		{"string key", "name", "test"},
		{"int key", "port", 8080},
		{"bool key", "enabled", true},
		{"float key", "floatNum", 3.14},
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

func TestBaseConfigProvider_Get_NestedKey(t *testing.T) {
	data := map[string]any{
		"database": map[string]any{
			"host": "localhost",
			"port": 3306,
			"credentials": map[string]any{
				"username": "admin",
				"password": "secret",
			},
		},
		"server": map[string]any{
			"host": "0.0.0.0",
			"port": 8080,
		},
	}
	provider := NewBaseConfigProvider(data)

	tests := []struct {
		name string
		key  string
		want any
	}{
		{"first level nested", "database.host", "localhost"},
		{"first level nested int", "database.port", 3306},
		{"second level nested", "database.credentials.username", "admin"},
		{"second level nested password", "database.credentials.password", "secret"},
		{"different first level", "server.port", 8080},
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

func TestBaseConfigProvider_Get_ArrayIndex(t *testing.T) {
	data := map[string]any{
		"servers": []any{
			map[string]any{"host": "server1", "port": 8001},
			map[string]any{"host": "server2", "port": 8002},
			map[string]any{"host": "server3", "port": 8003},
		},
		"items":   []any{"item1", "item2", "item3"},
		"numbers": []any{1, 2, 3, 4, 5},
	}
	provider := NewBaseConfigProvider(data)

	tests := []struct {
		name string
		key  string
		want any
	}{
		{"array index 0", "servers[0].host", "server1"},
		{"array index 1", "servers[1].port", 8002},
		{"array index 2", "servers[2].host", "server3"},
		{"simple array", "items[0]", "item1"},
		{"simple array middle", "items[1]", "item2"},
		{"number array", "numbers[3]", 4},
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

func TestBaseConfigProvider_Get_EntireArrayElement(t *testing.T) {
	data := map[string]any{
		"servers": []any{
			map[string]any{"host": "server1", "port": 8001},
			map[string]any{"host": "server2", "port": 8002},
		},
	}
	provider := NewBaseConfigProvider(data)

	// Get entire array element without accessing nested properties
	got, err := provider.Get("servers[0]")
	if err != nil {
		t.Fatalf("Get('servers[0]') error = %v", err)
	}

	// Type assert and verify contents
	element, ok := got.(map[string]any)
	if !ok {
		t.Fatalf("got is not map[string]any, got %T", got)
	}

	if element["host"] != "server1" {
		t.Errorf("element['host'] = %v, want 'server1'", element["host"])
	}
	if element["port"] != 8001 {
		t.Errorf("element['port'] = %v, want 8001", element["port"])
	}
}

func TestBaseConfigProvider_Get_KeyNotFound(t *testing.T) {
	data := map[string]any{
		"existing": "value",
		"nested": map[string]any{
			"key": "value",
		},
	}
	provider := NewBaseConfigProvider(data)

	tests := []struct {
		name        string
		key         string
		errorSubstr string
	}{
		{"non-existent key", "nonexistent", "not found"},
		{"non-existent nested", "nested.nonexistent", "not found"},
		{"non-existent deep nested", "nested.key.deeper", "expected object"},
		{"invalid path", "invalid..path", "not found"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := provider.Get(tt.key)
			if err == nil {
				t.Errorf("Get(%q) expected error, got nil", tt.key)
				return
			}
			if !contains(err.Error(), tt.errorSubstr) {
				t.Errorf("Get(%q) error = %v, expected to contain %v", tt.key, err, tt.errorSubstr)
			}
		})
	}
}

func TestBaseConfigProvider_Get_InvalidArrayIndex(t *testing.T) {
	data := map[string]any{
		"servers": []any{
			map[string]any{"host": "server1"},
			map[string]any{"host": "server2"},
		},
		"notArray": "string",
	}
	provider := NewBaseConfigProvider(data)

	tests := []struct {
		name        string
		key         string
		errorSubstr string
	}{
		{"index out of bounds", "servers[10].host", "out of bounds"},
		{"negative index", "servers[-1].host", "expected object"},
		{"index on non-array", "notArray[0]", "not an array"},
		// Note: "servers[abc]" won't match array index regex (requires digits),
		// so it's treated as "servers" key and returns the array
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := provider.Get(tt.key)
			if err == nil {
				t.Errorf("Get(%q) expected error, got nil", tt.key)
				return
			}
			if !contains(err.Error(), tt.errorSubstr) {
				t.Errorf("Get(%q) error = %v, expected to contain %v", tt.key, err, tt.errorSubstr)
			}
		})
	}
}

func TestBaseConfigProvider_Get_NestedArrayAccess(t *testing.T) {
	data := map[string]any{
		"environments": map[string]any{
			"production": []any{
				map[string]any{"host": "prod1", "port": 8080},
				map[string]any{"host": "prod2", "port": 8081},
			},
			"staging": []any{
				map[string]any{"host": "stage1", "port": 9080},
			},
		},
	}
	provider := NewBaseConfigProvider(data)

	tests := []struct {
		name string
		key  string
		want any
	}{
		{"production first server host", "environments.production[0].host", "prod1"},
		{"production second server port", "environments.production[1].port", 8081},
		{"staging only server", "environments.staging[0].host", "stage1"},
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

func TestBaseConfigProvider_Get_NonMapInPath(t *testing.T) {
	data := map[string]any{
		"string": "not a map",
		"number": 123,
	}
	provider := NewBaseConfigProvider(data)

	tests := []struct {
		name        string
		key         string
		errorSubstr string
	}{
		{"string as map", "string.key", "expected object"},
		{"number as map", "number.key", "expected object"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := provider.Get(tt.key)
			if err == nil {
				t.Errorf("Get(%q) expected error, got nil", tt.key)
				return
			}
			if !contains(err.Error(), tt.errorSubstr) {
				t.Errorf("Get(%q) error = %v, expected to contain %v", tt.key, err, tt.errorSubstr)
			}
		})
	}
}

func TestBaseConfigProvider_Get_ComplexPath(t *testing.T) {
	data := map[string]any{
		"app": map[string]any{
			"servers": []any{
				map[string]any{
					"name":  "web",
					"ports": []int{80, 443},
				},
				map[string]any{
					"name":  "api",
					"ports": []int{8080, 8443},
				},
			},
		},
	}
	provider := NewBaseConfigProvider(data)

	// Get entire array
	arr, err := provider.Get("app.servers")
	if err != nil {
		t.Errorf("Get('app.servers') error = %v", err)
	}

	servers, ok := arr.([]any)
	if !ok {
		t.Error("expected []any type")
	}
	if len(servers) != 2 {
		t.Errorf("expected 2 servers, got %d", len(servers))
	}
}

func TestBaseConfigProvider_Has(t *testing.T) {
	data := map[string]any{
		"existing": "value",
		"nested": map[string]any{
			"key": "value",
		},
		"servers": []any{
			map[string]any{"host": "server1"},
		},
	}
	provider := NewBaseConfigProvider(data)

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
		{"non-existing deep nested", "nested.key.deeper", false},
		{"empty key", "", true},
		{"invalid array index", "servers[10]", false},
		{"invalid path", "invalid..path", false},
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

func TestParsePath(t *testing.T) {
	provider := NewBaseConfigProvider(map[string]any{})

	tests := []struct {
		name       string
		path       string
		wantLen    int
		wantErr    bool
		firstKey   string
		firstIndex int
		hasIndex   bool
	}{
		{"simple key", "key", 1, false, "key", -1, false},
		{"nested key", "a.b.c", 3, false, "a", -1, false},
		{"array index", "servers[0]", 1, false, "servers", 0, true},
		{"nested with array", "a.b[0].c", 3, false, "a", -1, false},
		{"array index in middle", "servers[0].port", 2, false, "servers", 0, true},
		{"complex path", "app.servers[0].ports[1]", 3, false, "app", -1, false},
		{"empty string", "", 0, true, "", -1, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parts, err := provider.parsePath(tt.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("parsePath() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && len(parts) != tt.wantLen {
				t.Errorf("parsePath() returned %d parts, want %d", len(parts), tt.wantLen)
			}
			if !tt.wantErr && len(parts) > 0 {
				if parts[0].key != tt.firstKey {
					t.Errorf("parsePath()[0].key = %v, want %v", parts[0].key, tt.firstKey)
				}
				if parts[0].hasIndex != tt.hasIndex {
					t.Errorf("parsePath()[0].hasIndex = %v, want %v", parts[0].hasIndex, tt.hasIndex)
				}
			}
		})
	}
}

// Helper function
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(s) > 0 && len(substr) > 0 && containsHelper(s, substr)))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
