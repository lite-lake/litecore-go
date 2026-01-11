package config

import (
	"errors"
	"testing"

	"com.litelake.litecore/common"
	"com.litelake.litecore/config/internal/drivers"
)

// Mock provider for testing
type mockProvider struct {
	data map[string]any
}

func newMockProvider(data map[string]any) common.BaseConfigProvider {
	return &mockProvider{data: data}
}

func (m *mockProvider) Get(key string) (any, error) {
	if key == "" {
		return m.data, nil
	}
	val, ok := m.data[key]
	if !ok {
		return nil, errors.New("config key 'test.key' not found")
	}
	return val, nil
}

func (m *mockProvider) Has(key string) bool {
	_, ok := m.data[key]
	return ok
}

func TestIsConfigKeyNotFound(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{"nil error", nil, false},
		{"ErrKeyNotFound", ErrKeyNotFound, true},
		{"wrapped ErrKeyNotFound", errors.New("wrapped: config key not found"), true},
		{"other error", errors.New("some other error"), false},
		{"not found in message", errors.New("key 'test' not found"), true},
		{"random error", errors.New("random error"), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsConfigKeyNotFound(tt.err)
			if got != tt.expected {
				t.Errorf("IsConfigKeyNotFound() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestGet_Int(t *testing.T) {
	provider := newMockProvider(map[string]any{
		"int_val":    42,
		"float_val":  42.0,
		"string_int": "123",
		"string":     "not a number",
	})

	tests := []struct {
		name        string
		key         string
		expected    int
		expectError bool
	}{
		{"existing int", "int_val", 42, false},
		{"float to int", "float_val", 42, false},
		{"string to int", "string_int", 123, false},
		{"non-existent", "nonexistent", 0, true},
		{"invalid string", "string", 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Get[int](provider, tt.key)
			if (err != nil) != tt.expectError {
				t.Errorf("Get[int]() error = %v, expectError %v", err, tt.expectError)
				return
			}
			if !tt.expectError && got != tt.expected {
				t.Errorf("Get[int]() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestGet_Int64(t *testing.T) {
	provider := newMockProvider(map[string]any{
		"int_val":    int64(42),
		"float_val":  42.0,
		"string_int": "123",
	})

	tests := []struct {
		name        string
		key         string
		expected    int64
		expectError bool
	}{
		{"existing int64", "int_val", 42, false},
		{"float to int64", "float_val", 42, false},
		{"string to int64", "string_int", 123, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Get[int64](provider, tt.key)
			if (err != nil) != tt.expectError {
				t.Errorf("Get[int64]() error = %v, expectError %v", err, tt.expectError)
				return
			}
			if !tt.expectError && got != tt.expected {
				t.Errorf("Get[int64]() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestGet_Int32(t *testing.T) {
	provider := newMockProvider(map[string]any{
		"int_val":   int32(42),
		"float_val": 42.0,
	})

	tests := []struct {
		name        string
		key         string
		expected    int32
		expectError bool
	}{
		{"existing int32", "int_val", 42, false},
		{"float to int32", "float_val", 42, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Get[int32](provider, tt.key)
			if (err != nil) != tt.expectError {
				t.Errorf("Get[int32]() error = %v, expectError %v", err, tt.expectError)
				return
			}
			if !tt.expectError && got != tt.expected {
				t.Errorf("Get[int32]() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestGet_Float64(t *testing.T) {
	provider := newMockProvider(map[string]any{
		"float_val":    3.14,
		"int_val":      42,
		"string_float": "3.14",
		"string_int":   "42",
		"invalid":      "not a number",
	})

	tests := []struct {
		name        string
		key         string
		expected    float64
		expectError bool
	}{
		{"existing float", "float_val", 3.14, false},
		{"int to float", "int_val", 42.0, false},
		{"string to float", "string_float", 3.14, false},
		{"string int to float", "string_int", 42.0, false},
		{"invalid string", "invalid", 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Get[float64](provider, tt.key)
			if (err != nil) != tt.expectError {
				t.Errorf("Get[float64]() error = %v, expectError %v", err, tt.expectError)
				return
			}
			if !tt.expectError && got != tt.expected {
				t.Errorf("Get[float64]() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestGet_String(t *testing.T) {
	provider := newMockProvider(map[string]any{
		"string_val": "hello",
		"int_val":    42,
		"bool_val":   true,
		"float_val":  3.14,
	})

	tests := []struct {
		name        string
		key         string
		expected    string
		expectError bool
	}{
		{"existing string", "string_val", "hello", false},
		{"int to string", "int_val", "42", false},
		{"bool to string", "bool_val", "true", false},
		{"float to string", "float_val", "3.14", false},
		{"non-existent", "nonexistent", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Get[string](provider, tt.key)
			if (err != nil) != tt.expectError {
				t.Errorf("Get[string]() error = %v, expectError %v", err, tt.expectError)
				return
			}
			if !tt.expectError && got != tt.expected {
				t.Errorf("Get[string]() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestGet_Bool(t *testing.T) {
	provider := newMockProvider(map[string]any{
		"bool_val":       true,
		"string_true":    "true",
		"string_false":   "false",
		"string_yes":     "yes",
		"string_no":      "no",
		"string_one":     "1",
		"string_zero":    "0",
		"string_invalid": "not a bool",
	})

	tests := []struct {
		name        string
		key         string
		expected    bool
		expectError bool
	}{
		{"existing bool", "bool_val", true, false},
		{"string true", "string_true", true, false},
		{"string false", "string_false", false, false},
		{"string yes", "string_yes", true, true}, // lancet doesn't support "yes"/"no"
		{"string no", "string_no", false, true},  // lancet doesn't support "yes"/"no"
		{"string one", "string_one", true, false},
		{"string zero", "string_zero", false, false},
		{"invalid string", "string_invalid", false, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Get[bool](provider, tt.key)
			if (err != nil) != tt.expectError {
				t.Errorf("Get[bool]() error = %v, expectError %v", err, tt.expectError)
				return
			}
			if !tt.expectError && got != tt.expected {
				t.Errorf("Get[bool]() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestGet_TypeMismatch(t *testing.T) {
	provider := newMockProvider(map[string]any{
		"string_val": "hello",
	})

	_, err := Get[int](provider, "string_val")
	if err == nil {
		t.Error("expected type mismatch error, got nil")
		return
	}

	if !errors.Is(err, ErrTypeMismatch) {
		t.Errorf("expected ErrTypeMismatch, got: %v", err)
	}
}

func TestGet_UnsupportedType(t *testing.T) {
	provider := newMockProvider(map[string]any{
		"val": 42,
	})

	// Try to get an unsupported type (like map)
	type unsupportedType map[string]int
	_, err := Get[unsupportedType](provider, "val")
	if err == nil {
		t.Error("expected error for unsupported type, got nil")
	}
}

func TestGet_KeyNotFound(t *testing.T) {
	provider := newMockProvider(map[string]any{})

	_, err := Get[int](provider, "nonexistent")
	if err == nil {
		t.Error("expected key not found error, got nil")
		return
	}

	if !IsConfigKeyNotFound(err) {
		t.Errorf("expected key not found error, got: %v", err)
	}
}

func TestGetWithDefault_Int(t *testing.T) {
	provider := newMockProvider(map[string]any{
		"existing": 42,
	})

	tests := []struct {
		name         string
		key          string
		defaultValue int
		expected     int
	}{
		{"existing key", "existing", 99, 42},
		{"non-existent key", "nonexistent", 99, 99},
		{"zero default", "nonexistent", 0, 0},
		{"negative default", "nonexistent", -1, -1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetWithDefault(provider, tt.key, tt.defaultValue)
			if got != tt.expected {
				t.Errorf("GetWithDefault() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestGetWithDefault_String(t *testing.T) {
	provider := newMockProvider(map[string]any{
		"existing": "hello",
	})

	tests := []struct {
		name         string
		key          string
		defaultValue string
		expected     string
	}{
		{"existing key", "existing", "default", "hello"},
		{"non-existent key", "nonexistent", "default", "default"},
		{"empty default", "nonexistent", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetWithDefault(provider, tt.key, tt.defaultValue)
			if got != tt.expected {
				t.Errorf("GetWithDefault() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestGetWithDefault_Bool(t *testing.T) {
	provider := newMockProvider(map[string]any{
		"existing": true,
	})

	tests := []struct {
		name         string
		key          string
		defaultValue bool
		expected     bool
	}{
		{"existing key", "existing", false, true},
		{"non-existent key", "nonexistent", false, false},
		{"non-existent true", "nonexistent", true, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetWithDefault(provider, tt.key, tt.defaultValue)
			if got != tt.expected {
				t.Errorf("GetWithDefault() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestGetWithDefault_TypeMismatch(t *testing.T) {
	provider := newMockProvider(map[string]any{
		"string_val": "hello",
	})

	// When type mismatches, should return default value
	got := GetWithDefault[int](provider, "string_val", 99)
	if got != 99 {
		t.Errorf("GetWithDefault() with type mismatch = %v, want 99", got)
	}
}

func TestGetWithDefault_Pointer(t *testing.T) {
	provider := newMockProvider(map[string]any{
		"existing": 42,
	})

	// Test with pointer types
	defaultPtr := new(int)
	*defaultPtr = 99

	got := GetWithDefault(provider, "nonexistent", defaultPtr)
	if got != defaultPtr {
		t.Errorf("GetWithDefault() with pointer = %v, want %v", got, defaultPtr)
	}

	if *got != 99 {
		t.Errorf("GetWithDefault() pointer value = %v, want 99", *got)
	}
}

// Test with BaseConfigProvider integration
func TestGet_WithBaseProvider(t *testing.T) {
	data := map[string]any{
		"name":    "test",
		"port":    8080,
		"enabled": true,
		"rate":    3.14,
		"nested": map[string]any{
			"value": 42,
		},
	}
	provider := drivers.NewBaseConfigProvider(data)

	tests := []struct {
		name        string
		key         string
		expected    any
		expectError bool
	}{
		{"string value", "name", "test", false},
		{"int value", "port", 8080, false},
		{"bool value", "enabled", true, false},
		{"float value", "rate", 3.14, false},
		{"nested int", "nested.value", 42, false},
		{"not found", "missing", 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			switch tt.expected.(type) {
			case string:
				got, err := Get[string](provider, tt.key)
				if (err != nil) != tt.expectError {
					t.Errorf("Get[string]() error = %v, expectError %v", err, tt.expectError)
				} else if !tt.expectError && got != tt.expected {
					t.Errorf("Get[string]() = %v, want %v", got, tt.expected)
				}
			case int:
				got, err := Get[int](provider, tt.key)
				if (err != nil) != tt.expectError {
					t.Errorf("Get[int]() error = %v, expectError %v", err, tt.expectError)
				} else if !tt.expectError && got != tt.expected {
					t.Errorf("Get[int]() = %v, want %v", got, tt.expected)
				}
			case bool:
				got, err := Get[bool](provider, tt.key)
				if (err != nil) != tt.expectError {
					t.Errorf("Get[bool]() error = %v, expectError %v", err, tt.expectError)
				} else if !tt.expectError && got != tt.expected {
					t.Errorf("Get[bool]() = %v, want %v", got, tt.expected)
				}
			case float64:
				got, err := Get[float64](provider, tt.key)
				if (err != nil) != tt.expectError {
					t.Errorf("Get[float64]() error = %v, expectError %v", err, tt.expectError)
				} else if !tt.expectError && got != tt.expected {
					t.Errorf("Get[float64]() = %v, want %v", got, tt.expected)
				}
			}
		})
	}
}
