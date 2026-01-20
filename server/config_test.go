package server

import (
	"strconv"
	"testing"
	"time"
)

// TestDefaultServerConfig 测试默认服务器配置
func TestDefaultServerConfig(t *testing.T) {
	t.Run("默认配置_正确值", func(t *testing.T) {
		config := defaultServerConfig()

		if config.Host != "0.0.0.0" {
			t.Errorf("期望 Host = '0.0.0.0', 实际 = '%s'", config.Host)
		}

		if config.Port != 8080 {
			t.Errorf("期望 Port = 8080, 实际 = %d", config.Port)
		}

		if config.Mode != "release" {
			t.Errorf("期望 Mode = 'release', 实际 = '%s'", config.Mode)
		}

		if config.ReadTimeout != 10*time.Second {
			t.Errorf("期望 ReadTimeout = 10s, 实际 = %v", config.ReadTimeout)
		}

		if config.WriteTimeout != 10*time.Second {
			t.Errorf("期望 WriteTimeout = 10s, 实际 = %v", config.WriteTimeout)
		}

		if config.IdleTimeout != 60*time.Second {
			t.Errorf("期望 IdleTimeout = 60s, 实际 = %v", config.IdleTimeout)
		}

		if config.ShutdownTimeout != 30*time.Second {
			t.Errorf("期望 ShutdownTimeout = 30s, 实际 = %v", config.ShutdownTimeout)
		}
	})
}

// TestServerConfigAddress 测试 Address 方法
func TestServerConfigAddress(t *testing.T) {
	tests := []struct {
		name     string
		host     string
		port     int
		expected string
	}{
		{
			name:     "默认地址",
			host:     "0.0.0.0",
			port:     8080,
			expected: "0.0.0.0:8080",
		},
		{
			name:     "自定义地址",
			host:     "127.0.0.1",
			port:     3000,
			expected: "127.0.0.1:3000",
		},
		{
			name:     "localhost",
			host:     "localhost",
			port:     443,
			expected: "localhost:443",
		},
		{
			name:     "IPv6 地址",
			host:     "::1",
			port:     8080,
			expected: "::1:8080",
		},
		{
			name:     "最小端口",
			host:     "0.0.0.0",
			port:     1,
			expected: "0.0.0.0:1",
		},
		{
			name:     "最大端口",
			host:     "0.0.0.0",
			port:     65535,
			expected: "0.0.0.0:65535",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := &serverConfig{
				Host: tt.host,
				Port: tt.port,
			}

			result := config.Address()
			if result != tt.expected {
				t.Errorf("期望 Address() = '%s', 实际 = '%s'", tt.expected, result)
			}
		})
	}
}

// TestServerConfigPortConversion 测试端口转换
func TestServerConfigPortConversion(t *testing.T) {
	tests := []struct {
		name  string
		port  int
		valid bool
	}{
		{
			name:  "有效端口_80",
			port:  80,
			valid: true,
		},
		{
			name:  "有效端口_443",
			port:  443,
			valid: true,
		},
		{
			name:  "有效端口_3000",
			port:  3000,
			valid: true,
		},
		{
			name:  "边界值_0",
			port:  0,
			valid: false,
		},
		{
			name:  "边界值_65536",
			port:  65536,
			valid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := &serverConfig{
				Port: tt.port,
			}

			addr := config.Address()
			portStr := strconv.Itoa(tt.port)

			if tt.valid {
				if addr == "" {
					t.Errorf("有效端口 %d 应该生成有效地址", tt.port)
				}
				if len(addr) >= len(portStr) && addr[len(addr)-len(portStr):] != portStr {
					t.Errorf("地址应该包含端口 %d", tt.port)
				}
			} else {
				if addr == "" {
					t.Logf("无效端口 %d 生成地址: %s", tt.port, addr)
				}
			}
		})
	}
}

// TestServerConfigTimeoutValues 测试超时配置值
func TestServerConfigTimeoutValues(t *testing.T) {
	tests := []struct {
		name             string
		config           *serverConfig
		expectZero       bool
		expectedTimeouts map[string]time.Duration
	}{
		{
			name:       "零超时值",
			config:     &serverConfig{},
			expectZero: true,
		},
		{
			name: "自定义超时值",
			config: &serverConfig{
				ReadTimeout:     5 * time.Second,
				WriteTimeout:    5 * time.Second,
				IdleTimeout:     30 * time.Second,
				ShutdownTimeout: 10 * time.Second,
			},
			expectZero: false,
			expectedTimeouts: map[string]time.Duration{
				"ReadTimeout":     5 * time.Second,
				"WriteTimeout":    5 * time.Second,
				"IdleTimeout":     30 * time.Second,
				"ShutdownTimeout": 10 * time.Second,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.expectZero {
				if tt.config.ReadTimeout != 0 {
					t.Errorf("期望 ReadTimeout = 0, 实际 = %v", tt.config.ReadTimeout)
				}
				if tt.config.WriteTimeout != 0 {
					t.Errorf("期望 WriteTimeout = 0, 实际 = %v", tt.config.WriteTimeout)
				}
				if tt.config.IdleTimeout != 0 {
					t.Errorf("期望 IdleTimeout = 0, 实际 = %v", tt.config.IdleTimeout)
				}
				if tt.config.ShutdownTimeout != 0 {
					t.Errorf("期望 ShutdownTimeout = 0, 实际 = %v", tt.config.ShutdownTimeout)
				}
			} else {
				for name, expected := range tt.expectedTimeouts {
					var actual time.Duration
					switch name {
					case "ReadTimeout":
						actual = tt.config.ReadTimeout
					case "WriteTimeout":
						actual = tt.config.WriteTimeout
					case "IdleTimeout":
						actual = tt.config.IdleTimeout
					case "ShutdownTimeout":
						actual = tt.config.ShutdownTimeout
					}
					if actual != expected {
						t.Errorf("期望 %s = %v, 实际 = %v", name, expected, actual)
					}
				}
			}
		})
	}
}
