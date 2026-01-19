package cachemgr

import (
	"context"
	"testing"
	"time"
)

// TestValidateContext 测试上下文验证函数
func TestValidateContext(t *testing.T) {
	tests := []struct {
		name    string
		ctx     context.Context
		wantErr bool
	}{
		{
			name:    "valid context - Background",
			ctx:     context.Background(),
			wantErr: false,
		},
		{
			name:    "valid context - TODO",
			ctx:     context.TODO(),
			wantErr: false,
		},
		{
			name:    "valid context - WithValue",
			ctx:     context.WithValue(context.Background(), "key", "value"),
			wantErr: false,
		},
		{
			name: "valid context - WithCancel",
			ctx: func() context.Context {
				ctx, cancel := context.WithCancel(context.Background())
				defer cancel()
				return ctx
			}(),
			wantErr: false,
		},
		{
			name: "valid context - WithTimeout",
			ctx: func() context.Context {
				ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				defer cancel()
				return ctx
			}(),
			wantErr: false,
		},
		{
			name: "valid context - WithDeadline",
			ctx: func() context.Context {
				ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(5*time.Second))
				defer cancel()
				return ctx
			}(),
			wantErr: false,
		},
		{
			name:    "nil context",
			ctx:     nil,
			wantErr: true,
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

// TestValidateKey 测试键验证函数
func TestValidateKey(t *testing.T) {
	tests := []struct {
		name    string
		key     string
		wantErr bool
	}{
		{
			name:    "valid simple key",
			key:     "user:123",
			wantErr: false,
		},
		{
			name:    "valid key with special chars",
			key:     "cache::user::session::abc123",
			wantErr: false,
		},
		{
			name:    "valid key with numbers",
			key:     "key_12345",
			wantErr: false,
		},
		{
			name:    "valid key with dashes",
			key:     "my-cache-key",
			wantErr: false,
		},
		{
			name:    "valid key with dots",
			key:     "cache.key.value",
			wantErr: false,
		},
		{
			name:    "valid key with underscores",
			key:     "my_cache_key",
			wantErr: false,
		},
		{
			name:    "valid key with mixed separators",
			key:     "user:123_profile:settings_theme",
			wantErr: false,
		},
		{
			name:    "valid long key",
			key:     "this_is_a_very_long_cache_key_that_contains_lot_of_information",
			wantErr: false,
		},
		{
			name:    "valid single character key",
			key:     "a",
			wantErr: false,
		},
		{
			name:    "valid numeric key as string",
			key:     "12345",
			wantErr: false,
		},
		{
			name:    "empty key",
			key:     "",
			wantErr: true,
		},
		{
			name:    "key with only spaces",
			key:     "   ",
			wantErr: false, // 空格在技术上是有效字符
		},
		{
			name:    "key with unicode characters",
			key:     "用户:123",
			wantErr: false,
		},
		{
			name:    "key with emoji",
			key:     "cache:🔥",
			wantErr: false,
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

// TestSanitizeKey 测试键脱敏函数
func TestSanitizeKey(t *testing.T) {
	tests := []struct {
		name     string
		key      string
		expected string
	}{
		{
			name:     "short key - less than 10 chars",
			key:      "short",
			expected: "short",
		},
		{
			name:     "exactly 10 characters",
			key:      "0123456789",
			expected: "0123456789",
		},
		{
			name:     "exactly 5 characters",
			key:      "abcde",
			expected: "abcde",
		},
		{
			name:     "11 characters",
			key:      "01234567890",
			expected: "01234***",
		},
		{
			name:     "long key - 20 characters",
			key:      "this_is_a_test_key_1",
			expected: "this_***",
		},
		{
			name:     "very long key - 50 characters",
			key:      "this_is_a_very_long_cache_key_that_should_be_hidden",
			expected: "this_***",
		},
		{
			name:     "key with special chars",
			key:      "user:12345:profile:settings",
			expected: "user:***",
		},
		{
			name:     "key with emoji",
			key:      "cache:🔥🔥🔥🔥🔥🔥🔥🔥🔥🔥",
			expected: "cache***",
		},
		{
			name:     "key with spaces",
			key:      "cache key with spaces",
			expected: "cache***", // sanitizeKey 不会在冒号前加空格
		},
		{
			name:     "single character",
			key:      "a",
			expected: "a",
		},
		{
			name:     "empty string",
			key:      "",
			expected: "",
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

// TestGetStatus 测试状态获取函数
func TestGetStatus(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want string
	}{
		{
			name: "no error returns success",
			err:  nil,
			want: "success",
		},
		{
			name: "error returns error",
			err:  context.Canceled,
			want: "error",
		},
		{
			name: "context deadline exceeded",
			err:  context.DeadlineExceeded,
			want: "error",
		},
		{
			name: "generic error",
			err:  &testError{},
			want: "error",
		},
		{
			name: "custom error",
			err:  &testError{},
			want: "error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getStatus(tt.err); got != tt.want {
				t.Errorf("getStatus() = %s, want %s", got, tt.want)
			}
		})
	}
}

// testError 自定义错误类型用于测试
type testError struct{}

func (e *testError) Error() string {
	return "test error"
}

// TestNewCacheManagerBaseImpl 测试基础实现创建
func TestNewCacheManagerBaseImpl(t *testing.T) {
	base := newCacheManagerBaseImpl()

	if base == nil {
		t.Fatal("newCacheManagerBaseImpl() returned nil")
	}

	if base.loggerMgr != nil {
		t.Error("expected loggerMgr to be nil initially")
	}

	if base.telemetryMgr != nil {
		t.Error("expected telemetryMgr to be nil initially")
	}

	if base.logger != nil {
		t.Error("expected logger to be nil before initialization")
	}

	if base.tracer != nil {
		t.Error("expected tracer to be nil before initialization")
	}

	if base.meter != nil {
		t.Error("expected meter to be nil before initialization")
	}
}

// TestCacheManagerBaseImpl_InitObservability 测试初始化可观测性
func TestCacheManagerBaseImpl_InitObservability(t *testing.T) {
	base := newCacheManagerBaseImpl()

	// 调用初始化（没有依赖注入的情况下）
	base.initObservability()

	// 验证没有 panic，且字段保持为 nil
	if base.logger != nil {
		t.Error("expected logger to remain nil without loggerMgr")
	}

	if base.tracer != nil {
		t.Error("expected tracer to remain nil without telemetryMgr")
	}

	if base.meter != nil {
		t.Error("expected meter to remain nil without telemetryMgr")
	}

	if base.cacheHitCounter != nil {
		t.Error("expected cacheHitCounter to remain nil without telemetryMgr")
	}

	if base.cacheMissCounter != nil {
		t.Error("expected cacheMissCounter to remain nil without telemetryMgr")
	}

	if base.operationDuration != nil {
		t.Error("expected operationDuration to remain nil without telemetryMgr")
	}
}

// TestRecordOperation 测试记录操作
func TestRecordOperation(t *testing.T) {
	base := newCacheManagerBaseImpl()

	tests := []struct {
		name      string
		driver    string
		operation string
		key       string
		fn        func() error
		wantErr   bool
	}{
		{
			name:      "successful operation",
			driver:    "memory",
			operation: "get",
			key:       "test_key",
			fn:        func() error { return nil },
			wantErr:   false,
		},
		{
			name:      "failed operation",
			driver:    "memory",
			operation: "set",
			key:       "test_key",
			fn:        func() error { return &testError{} },
			wantErr:   true,
		},
		{
			name:      "operation with empty key",
			driver:    "redis",
			operation: "delete",
			key:       "",
			fn:        func() error { return nil },
			wantErr:   false,
		},
		{
			name:      "operation with long key",
			driver:    "memory",
			operation: "get",
			key:       "this_is_a_very_long_cache_key_that_should_be_sanitized_in_logs",
			fn:        func() error { return nil },
			wantErr:   false,
		},
	}

	ctx := context.Background()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := base.recordOperation(ctx, tt.driver, tt.operation, tt.key, tt.fn)
			if (err != nil) != tt.wantErr {
				t.Errorf("recordOperation() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestRecordOperationWithNilContext 测试使用 nil 上下文记录操作
func TestRecordOperationWithNilContext(t *testing.T) {
	base := newCacheManagerBaseImpl()

	err := base.recordOperation(nil, "memory", "get", "key", func() error {
		return nil
	})

	// 如果操作函数不验证上下文，应该成功
	if err != nil {
		t.Logf("recordOperation with nil context returned error: %v", err)
	}
}

// TestRecordCacheHit 测试记录缓存命中
func TestRecordCacheHit(t *testing.T) {
	base := newCacheManagerBaseImpl()
	base.initObservability()

	ctx := context.Background()

	// 测试没有 meter 的情况
	base.recordCacheHit(ctx, "memory", true)
	base.recordCacheHit(ctx, "memory", false)
	base.recordCacheHit(ctx, "redis", true)
	base.recordCacheHit(ctx, "redis", false)

	// 这些调用不应该 panic
}

// TestCacheManagerBaseImplConcurrent 测试并发安全性
func TestCacheManagerBaseImplConcurrent(t *testing.T) {
	base := newCacheManagerBaseImpl()
	base.initObservability()

	ctx := context.Background()
	done := make(chan bool)

	// 并发调用 recordOperation
	for i := 0; i < 100; i++ {
		go func(id int) {
			err := base.recordOperation(ctx, "memory", "get", "test_key", func() error {
				return nil
			})
			if err != nil {
				t.Errorf("concurrent operation %d failed: %v", id, err)
			}
			done <- true
		}(i)
	}

	// 等待所有 goroutine 完成
	for i := 0; i < 100; i++ {
		<-done
	}
}

// BenchmarkSanitizeKey 性能测试 - 键脱敏
func BenchmarkSanitizeKey(b *testing.B) {
	keys := []string{
		"short",
		"medium_length_key",
		"this_is_a_very_long_cache_key_that_should_be_sanitized_for_logging",
		"user:12345:profile:settings:theme:dark:language:en",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, key := range keys {
			sanitizeKey(key)
		}
	}
}

// BenchmarkValidateContext 性能测试 - 上下文验证
func BenchmarkValidateContext(b *testing.B) {
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ValidateContext(ctx)
	}
}

// BenchmarkValidateKey 性能测试 - 键验证
func BenchmarkValidateKey(b *testing.B) {
	key := "user:12345:profile:settings"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ValidateKey(key)
	}
}
