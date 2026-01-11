package server

import (
	"testing"
	"time"
)

func TestDefaultServerConfig(t *testing.T) {
	config := DefaultServerConfig()

	tests := []struct {
		name     string
		expected interface{}
		actual   interface{}
	}{
		{"Host", "0.0.0.0", config.Host},
		{"Port", 8080, config.Port},
		{"Mode", "release", config.Mode},
		{"ReadTimeout", 10 * time.Second, config.ReadTimeout},
		{"WriteTimeout", 10 * time.Second, config.WriteTimeout},
		{"IdleTimeout", 60 * time.Second, config.IdleTimeout},
		{"EnableMetrics", true, config.EnableMetrics},
		{"EnableHealth", true, config.EnableHealth},
		{"EnablePprof", false, config.EnablePprof},
		{"EnableRecovery", true, config.EnableRecovery},
		{"ShutdownTimeout", 30 * time.Second, config.ShutdownTimeout},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.expected != tt.actual {
				t.Errorf("%s: expected %v, got %v", tt.name, tt.expected, tt.actual)
			}
		})
	}
}

func TestServerConfig_Address(t *testing.T) {
	tests := []struct {
		name     string
		host     string
		port     int
		expected string
	}{
		{"default config", "0.0.0.0", 8080, "0.0.0.0:8080"},
		{"localhost", "localhost", 3000, "localhost:3000"},
		{"custom host", "192.168.1.1", 9000, "192.168.1.1:9000"},
		{"high port", "127.0.0.1", 65535, "127.0.0.1:65535"},
		{"zero port", "0.0.0.0", 0, "0.0.0.0:0"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := &ServerConfig{
				Host: tt.host,
				Port: tt.port,
			}
			if got := config.Address(); got != tt.expected {
				t.Errorf("Address() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestToString(t *testing.T) {
	tests := []struct {
		name     string
		input    int
		expected string
	}{
		{"zero", 0, "0"},
		{"single digit", 5, "5"},
		{"double digits", 42, "42"},
		{"triple digits", 808, "808"},
		{"large number", 65535, "65535"},
		{"very large", 2147483647, "2147483647"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := toString(tt.input); got != tt.expected {
				t.Errorf("toString(%d) = %v, want %v", tt.input, got, tt.expected)
			}
		})
	}
}

func TestServerConfig_CustomConfig(t *testing.T) {
	config := &ServerConfig{
		Host:            "127.0.0.1",
		Port:            9090,
		Mode:            "debug",
		ReadTimeout:     30 * time.Second,
		WriteTimeout:    30 * time.Second,
		IdleTimeout:     120 * time.Second,
		EnableMetrics:   false,
		EnableHealth:    false,
		EnablePprof:     true,
		EnableRecovery:  false,
		ShutdownTimeout: 60 * time.Second,
	}

	tests := []struct {
		name     string
		expected interface{}
		actual   interface{}
	}{
		{"Host", "127.0.0.1", config.Host},
		{"Port", 9090, config.Port},
		{"Mode", "debug", config.Mode},
		{"ReadTimeout", 30 * time.Second, config.ReadTimeout},
		{"EnableMetrics", false, config.EnableMetrics},
		{"EnableHealth", false, config.EnableHealth},
		{"EnablePprof", true, config.EnablePprof},
		{"ShutdownTimeout", 60 * time.Second, config.ShutdownTimeout},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.expected != tt.actual {
				t.Errorf("%s: expected %v, got %v", tt.name, tt.expected, tt.actual)
			}
		})
	}

	expectedAddr := "127.0.0.1:9090"
	if got := config.Address(); got != expectedAddr {
		t.Errorf("Address() = %v, want %v", got, expectedAddr)
	}
}
