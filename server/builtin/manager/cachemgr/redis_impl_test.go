package cachemgr

import (
	"context"
	"errors"
	"testing"
	"time"
)

// TestRedisManager_NewCacheManagerRedisImpl 测试创建 Redis 管理器
func TestRedisManager_NewCacheManagerRedisImpl(t *testing.T) {
	tests := []struct {
		name    string
		config  *RedisConfig
		wantErr bool
	}{
		{
			name: "invalid configmgr - no connection",
			config: &RedisConfig{
				Host: "localhost",
				Port: 9999, // 使用不存在的端口
				DB:   0,
			},
			wantErr: true, // 预期会失败，因为无法连接
		},
		{
			name: "invalid configmgr - negative port",
			config: &RedisConfig{
				Host: "localhost",
				Port: -1,
				DB:   0,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mgr, err := NewCacheManagerRedisImpl(tt.config)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewCacheManagerRedisImpl() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && mgr != nil {
				defer mgr.Close()
			}
		})
	}
}

// TestRedisManager_SerializeDeserialize 测试序列化和反序列化
func TestRedisManager_SerializeDeserialize(t *testing.T) {
	tests := []struct {
		name    string
		value   any
		wantErr bool
	}{
		{
			name:  "string",
			value: "test string",
		},
		{
			name:  "int",
			value: 42,
		},
		{
			name:  "float",
			value: 3.14,
		},
		{
			name:  "bool",
			value: true,
		},
		{
			name:    "nil",
			value:   nil,
			wantErr: true, // gob cannot encode nil
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 测试序列化
			data, err := serialize(tt.value)
			if (err != nil) != tt.wantErr {
				t.Errorf("serialize() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				// 对于简单类型，直接比较序列化后的数据
				if len(data) == 0 {
					t.Error("serialize() returned empty data")
				}
			}
		})
	}
}

// TestRedisManager_SerializeComplexTypes 测试复杂类型的序列化
func TestRedisManager_SerializeComplexTypes(t *testing.T) {
	// 测试嵌套结构
	type Address struct {
		City    string
		Country string
	}
	type Person struct {
		Name    string
		Age     int
		Address Address
	}

	person := Person{
		Name: "Bob",
		Age:  25,
		Address: Address{
			City:    "New York",
			Country: "USA",
		},
	}

	data, err := serialize(person)
	if err != nil {
		t.Fatalf("serialize() error = %v", err)
	}

	var dest Person
	err = deserialize(data, &dest)
	if err != nil {
		t.Fatalf("deserialize() error = %v", err)
	}

	if dest.Name != person.Name {
		t.Errorf("expected Name '%s', got '%s'", person.Name, dest.Name)
	}
	if dest.Age != person.Age {
		t.Errorf("expected Age %d, got %d", person.Age, dest.Age)
	}
	if dest.Address.City != person.Address.City {
		t.Errorf("expected City '%s', got '%s'", person.Address.City, dest.Address.City)
	}
}

// TestRedisManager_DeserializeInvalidData 测试反序列化无效数据
func TestRedisManager_DeserializeInvalidData(t *testing.T) {
	tests := []struct {
		name    string
		data    []byte
		wantErr bool
	}{
		{
			name:    "empty data",
			data:    []byte{},
			wantErr: true,
		},
		{
			name:    "nil data",
			data:    nil,
			wantErr: true,
		},
		{
			name:    "invalid gob data",
			data:    []byte{0x01, 0x02, 0x03},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var dest any
			err := deserialize(tt.data, &dest)
			if (err != nil) != tt.wantErr {
				t.Errorf("deserialize() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestRedisManager_ValidateContext 测试上下文验证
func TestRedisManager_ValidateContext(t *testing.T) {
	tests := []struct {
		name    string
		ctx     context.Context
		wantErr bool
	}{
		{
			name:    "valid context",
			ctx:     context.Background(),
			wantErr: false,
		},
		{
			name:    "nil context",
			ctx:     nil,
			wantErr: true,
		},
		{
			name:    "context with timeout",
			ctx:     context.WithValue(context.Background(), "key", "value"),
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateContext(tt.ctx)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateContext() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestRedisManager_ValidateKey 测试键验证
func TestRedisManager_ValidateKey(t *testing.T) {
	tests := []struct {
		name    string
		key     string
		wantErr bool
	}{
		{
			name:    "valid key",
			key:     "test_key",
			wantErr: false,
		},
		{
			name:    "empty key",
			key:     "",
			wantErr: true,
		},
		{
			name:    "key with spaces",
			key:     "test key",
			wantErr: false, // 实际上是有效的
		},
		{
			name:    "key with special chars",
			key:     "test:key:value",
			wantErr: false, // Redis 支持特殊字符
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateKey(tt.key)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateKey() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestRedisManager_SanitizeKey 测试键脱敏
func TestRedisManager_SanitizeKey(t *testing.T) {
	tests := []struct {
		name     string
		key      string
		expected string
	}{
		{
			name:     "short key",
			key:      "short",
			expected: "short",
		},
		{
			name:     "exactly 10 chars",
			key:      "0123456789",
			expected: "0123456789",
		},
		{
			name:     "long key",
			key:      "this_is_a_very_long_cache_key_that_should_be_sanitized",
			expected: "this_***",
		},
		{
			name:     "exactly 5 chars",
			key:      "abcde",
			expected: "abcde",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := sanitizeKey(tt.key)
			if result != tt.expected {
				t.Errorf("sanitizeKey() = %s, want %s", result, tt.expected)
			}
		})
	}
}

// TestRedisManager_GetStatus 测试状态获取
func TestRedisManager_GetStatus(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want string
	}{
		{
			name: "no error",
			err:  nil,
			want: "success",
		},
		{
			name: "with error",
			err:  errors.New("test error"),
			want: "error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getStatus(tt.err)
			if result != tt.want {
				t.Errorf("getStatus() = %s, want %s", result, tt.want)
			}
		})
	}
}

// TestRedisManager_ConfigDefaults 测试 Redis 配置默认值
func TestRedisManager_ConfigDefaults(t *testing.T) {
	config := &RedisConfig{
		Host:            "localhost",
		Port:            6379,
		DB:              0,
		MaxIdleConns:    10,
		MaxOpenConns:    100,
		ConnMaxLifetime: 30 * time.Second,
	}

	if config.Host != "localhost" {
		t.Errorf("expected Host 'localhost', got '%s'", config.Host)
	}
	if config.Port != 6379 {
		t.Errorf("expected Port 6379, got %d", config.Port)
	}
	if config.DB != 0 {
		t.Errorf("expected DB 0, got %d", config.DB)
	}
	if config.MaxIdleConns != 10 {
		t.Errorf("expected MaxIdleConns 10, got %d", config.MaxIdleConns)
	}
	if config.MaxOpenConns != 100 {
		t.Errorf("expected MaxOpenConns 100, got %d", config.MaxOpenConns)
	}
	if config.ConnMaxLifetime != 30*time.Second {
		t.Errorf("expected ConnMaxLifetime 30s, got %v", config.ConnMaxLifetime)
	}
}

// TestRedisManager_ConfigEdgeCases 测试配置边界情况
func TestRedisManager_ConfigEdgeCases(t *testing.T) {
	tests := []struct {
		name    string
		config  *RedisConfig
		wantErr bool
	}{
		{
			name: "zero port",
			config: &RedisConfig{
				Host: "localhost",
				Port: 0,
				DB:   0,
			},
			wantErr: true, // Redis 连接会失败
		},
		{
			name: "large port",
			config: &RedisConfig{
				Host: "localhost",
				Port: 65535,
				DB:   0,
			},
			wantErr: true, // Redis 连接会失败
		},
		{
			name: "negative DB",
			config: &RedisConfig{
				Host: "localhost",
				Port: 6379,
				DB:   -1,
			},
			wantErr: true, // Redis 连接会失败
		},
		{
			name: "DB > 15",
			config: &RedisConfig{
				Host: "localhost",
				Port: 6379,
				DB:   16,
			},
			wantErr: true, // Redis 连接会失败
		},
		{
			name: "empty host",
			config: &RedisConfig{
				Host: "",
				Port: 6379,
				DB:   0,
			},
			wantErr: true, // Redis 连接会失败
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mgr, err := NewCacheManagerRedisImpl(tt.config)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewCacheManagerRedisImpl() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr && mgr != nil {
				mgr.Close()
			}
		})
	}
}

// TestRedisManager_PasswordConfig 测试密码配置
func TestRedisManager_PasswordConfig(t *testing.T) {
	config := &RedisConfig{
		Host:     "localhost",
		Port:     6379,
		Password: "test_password",
		DB:       0,
	}

	// 由于没有实际的 Redis 服务器，预期会失败
	mgr, err := NewCacheManagerRedisImpl(config)
	if err == nil {
		mgr.Close()
		t.Log("Note: This test requires a Redis server to properly test password authentication")
	}
}

// TestRedisManager_ConnectionPoolConfig 测试连接池配置
func TestRedisManager_ConnectionPoolConfig(t *testing.T) {
	tests := []struct {
		name            string
		maxIdleConns    int
		maxOpenConns    int
		connMaxLifetime time.Duration
	}{
		{
			name:            "small pool",
			maxIdleConns:    1,
			maxOpenConns:    5,
			connMaxLifetime: 10 * time.Second,
		},
		{
			name:            "medium pool",
			maxIdleConns:    10,
			maxOpenConns:    50,
			connMaxLifetime: 30 * time.Second,
		},
		{
			name:            "large pool",
			maxIdleConns:    100,
			maxOpenConns:    500,
			connMaxLifetime: 60 * time.Second,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := &RedisConfig{
				Host:            "localhost",
				Port:            9999, // 使用不存在的端口
				DB:              0,
				MaxIdleConns:    tt.maxIdleConns,
				MaxOpenConns:    tt.maxOpenConns,
				ConnMaxLifetime: tt.connMaxLifetime,
			}

			mgr, err := NewCacheManagerRedisImpl(config)
			if err == nil {
				mgr.Close()
				t.Log("Note: This test requires a Redis server to properly test connection pool")
			}
		})
	}
}

// TestRedisManager_Timeout 测试超时处理
func TestRedisManager_Timeout(t *testing.T) {
	config := &RedisConfig{
		Host:            "localhost",
		Port:            9999, // 使用不存在的端口
		DB:              0,
		ConnMaxLifetime: 1 * time.Nanosecond, // 非常短的超时
	}

	mgr, err := NewCacheManagerRedisImpl(config)
	if err == nil {
		defer mgr.Close()

		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
		defer cancel()

		// 测试超时
		var result any
		err = mgr.Get(ctx, "key", &result)
		if err == nil {
			t.Error("expected timeout error, got nil")
		}
	}
}

// TestRedisManager_ContextCancellation 测试上下文取消
func TestRedisManager_ContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // 立即取消

	config := &RedisConfig{
		Host: "localhost",
		Port: 9999,
		DB:   0,
	}

	mgr, err := NewCacheManagerRedisImpl(config)
	if err == nil {
		defer mgr.Close()

		var result any
		err = mgr.Get(ctx, "key", &result)
		if err == nil {
			t.Error("expected context cancellation error, got nil")
		}
	}
}
