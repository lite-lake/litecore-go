package databasemgr

import (
	"context"
	"strings"
	"testing"
	"time"

	"gorm.io/gorm"
)

// TestNewObservabilityPlugin 测试创建可观测性插件
func TestNewObservabilityPlugin(t *testing.T) {
	plugin := newObservabilityPlugin()

	if plugin == nil {
		t.Fatal("newObservabilityPlugin() returned nil")
	}

	if plugin.slowQueryThreshold != 1*time.Second {
		t.Errorf("default slowQueryThreshold = %v, want 1s", plugin.slowQueryThreshold)
	}

	if plugin.logSQL != false {
		t.Errorf("default logSQL = %v, want false", plugin.logSQL)
	}

	if plugin.sampleRate != 1.0 {
		t.Errorf("default sampleRate = %v, want 1.0", plugin.sampleRate)
	}
}

// TestObservabilityPlugin_Name 测试插件名称
func TestObservabilityPlugin_Name(t *testing.T) {
	plugin := newObservabilityPlugin()

	if plugin.Name() != "observability" {
		t.Errorf("Name() = %v, want 'observability'", plugin.Name())
	}
}

// TestObservabilityPlugin_SetConfig 测试设置配置
func TestObservabilityPlugin_SetConfig(t *testing.T) {
	plugin := newObservabilityPlugin()

	plugin.SetConfig(2*time.Second, true, 0.5)

	if plugin.slowQueryThreshold != 2*time.Second {
		t.Errorf("slowQueryThreshold = %v, want 2s", plugin.slowQueryThreshold)
	}

	if plugin.logSQL != true {
		t.Errorf("logSQL = %v, want true", plugin.logSQL)
	}

	if plugin.sampleRate != 0.5 {
		t.Errorf("sampleRate = %v, want 0.5", plugin.sampleRate)
	}
}

// TestObservabilityPlugin_Initialize 测试插件初始化
func TestObservabilityPlugin_Initialize(t *testing.T) {
	plugin := newObservabilityPlugin()
	db, _ := gorm.Open(&dummyDialector{}, &gorm.Config{})

	err := plugin.Initialize(db)
	if err != nil {
		t.Errorf("Initialize() error = %v", err)
	}
}

// TestSanitizeSQL 测试 SQL 脱敏
func TestSanitizeSQL(t *testing.T) {
	tests := []struct {
		name        string
		sql         string
		contains    []string
		notContains []string
	}{
		{
			name:        "empty SQL",
			sql:         "",
			contains:    []string{""},
			notContains: []string{},
		},
		{
			name:     "simple SELECT",
			sql:      "SELECT * FROM users WHERE id = 1",
			contains: []string{"SELECT", "users"},
		},
		{
			name:        "password in single quotes",
			sql:         "SELECT * FROM users WHERE password = 'secret123'",
			contains:    []string{"SELECT", "users"},
			notContains: []string{"secret123"},
		},
		{
			name:        "password in double quotes",
			sql:         `SELECT * FROM users WHERE password = "secret123"`,
			contains:    []string{"SELECT", "users"},
			notContains: []string{"secret123"},
		},
		{
			name:        "pwd field",
			sql:         "UPDATE users SET pwd = 'mypass' WHERE id = 1",
			contains:    []string{"UPDATE", "users"},
			notContains: []string{"mypass"},
		},
		{
			name:        "token field",
			sql:         "INSERT INTO tokens (token) VALUES ('abc123xyz')",
			contains:    []string{"INSERT", "tokens"},
			notContains: []string{}, // token in VALUES context, not in key=value format
		},
		{
			name:        "secret field",
			sql:         "SELECT * FROM secrets WHERE secret = 'mysecret'",
			contains:    []string{"SELECT", "secrets"},
			notContains: []string{"mysecret"},
		},
		{
			name:        "api_key field",
			sql:         "SELECT * FROM api_keys WHERE api_key = 'key123'",
			contains:    []string{"SELECT", "api_keys"},
			notContains: []string{"key123"},
		},
		{
			name:     "long SQL truncation",
			sql:      strings.Repeat("SELECT * FROM users WHERE name = '", 100) + "test" + strings.Repeat("'", 100),
			contains: []string{"..."},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := sanitizeSQL(tt.sql)

			// 检查应该包含的内容
			for _, c := range tt.contains {
				if !strings.Contains(result, c) && c != "" {
					t.Errorf("sanitizeSQL() should contain %q", c)
				}
			}

			// 检查不应该包含的内容
			for _, nc := range tt.notContains {
				if strings.Contains(result, nc) {
					t.Errorf("sanitizeSQL() should not contain %q", nc)
				}
			}
		})
	}
}

// TestValidateContext 测试上下文验证
func TestValidateContext(t *testing.T) {
	tests := []struct {
		name    string
		ctx     context.Context
		wantErr bool
	}{
		{
			name:    "nil context",
			ctx:     nil,
			wantErr: true,
		},
		{
			name:    "valid context",
			ctx:     context.Background(),
			wantErr: false,
		},
		{
			name:    "context with timeout",
			ctx:     context.TODO(),
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

// TestValidateDSN 测试 DSN 验证
func TestValidateDSN(t *testing.T) {
	tests := []struct {
		name    string
		dsn     string
		wantErr bool
	}{
		{
			name:    "empty DSN",
			dsn:     "",
			wantErr: true,
		},
		{
			name:    "valid SQLite DSN",
			dsn:     ":memory:",
			wantErr: false,
		},
		{
			name:    "valid MySQL DSN",
			dsn:     "root:password@tcp(localhost:3306)/test",
			wantErr: false,
		},
		{
			name:    "valid PostgreSQL DSN",
			dsn:     "host=localhost port=5432",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateDSN(tt.dsn)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateDSN() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestGetStatus 测试状态获取
func TestGetStatus(t *testing.T) {
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
			err:  gorm.ErrInvalidTransaction,
			want: "error",
		},
		{
			name: "custom error",
			err:  gorm.ErrRecordNotFound,
			want: "error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getStatus(tt.err); got != tt.want {
				t.Errorf("getStatus() = %v, want %v", got, tt.want)
			}
		})
	}
}

// BenchmarkSanitizeSQL 基准测试 SQL 脱敏
func BenchmarkSanitizeSQL(b *testing.B) {
	sql := "SELECT * FROM users WHERE password = 'secret123' AND token = 'abc123xyz' AND api_key = 'key456'"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sanitizeSQL(sql)
	}
}
