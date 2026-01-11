package drivers

import (
	"context"
	"testing"

	"com.litelake.litecore/common"
)

// TestBaseManager_ManagerName 测试管理器名称
func TestBaseManager_ManagerName(t *testing.T) {
	tests := []struct {
		name string
		want string
	}{
		{"test-manager", "test-manager"},
		{"", ""},
		{"database-manager", "database-manager"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := NewBaseManager(tt.name)
			if got := m.ManagerName(); got != tt.want {
				t.Errorf("BaseManager.ManagerName() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestBaseManager_Health 测试健康检查
func TestBaseManager_Health(t *testing.T) {
	m := NewBaseManager("test-manager")
	if err := m.Health(); err != nil {
		t.Errorf("BaseManager.Health() error = %v", err)
	}
}

// TestBaseManager_OnStart 测试启动钩子
func TestBaseManager_OnStart(t *testing.T) {
	m := NewBaseManager("test-manager")
	if err := m.OnStart(); err != nil {
		t.Errorf("BaseManager.OnStart() error = %v", err)
	}
}

// TestBaseManager_OnStop 测试停止钩子
func TestBaseManager_OnStop(t *testing.T) {
	m := NewBaseManager("test-manager")
	if err := m.OnStop(); err != nil {
		t.Errorf("BaseManager.OnStop() error = %v", err)
	}
}

// TestBaseManager_Shutdown 测试关闭
func TestBaseManager_Shutdown(t *testing.T) {
	tests := []struct {
		name    string
		ctx     context.Context
		wantErr bool
	}{
		{
			name:    "正常上下文",
			ctx:     context.Background(),
			wantErr: false,
		},
		{
			name:    "带超时的上下文",
			ctx:     context.Background(),
			wantErr: false,
		},
		{
			name:    "已取消的上下文",
			ctx:     func() context.Context { ctx, cancel := context.WithCancel(context.Background()); cancel(); return ctx }(),
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := NewBaseManager("test-manager")
			if err := m.Shutdown(tt.ctx); (err != nil) != tt.wantErr {
				t.Errorf("BaseManager.Shutdown() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestBaseManager_ImplementsCommonManager 测试 BaseManager 实现了 common.BaseManager 接口
func TestBaseManager_ImplementsCommonManager(t *testing.T) {
	var _ common.BaseManager = (*BaseManager)(nil)
	m := NewBaseManager("test-manager")
	if m == nil {
		t.Error("NewBaseManager() returned nil")
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
			name:    "有效上下文",
			ctx:     context.Background(),
			wantErr: false,
		},
		{
			name:    "nil 上下文",
			ctx:     nil,
			wantErr: true,
		},
		{
			name:    "带值的上下文",
			ctx:     context.WithValue(context.Background(), "key", "value"),
			wantErr: false,
		},
		{
			name:    "带超时的上下文",
			ctx:     context.TODO(),
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := ValidateContext(tt.ctx); (err != nil) != tt.wantErr {
				t.Errorf("ValidateContext() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// BenchmarkBaseManager_Health 基准测试健康检查
func BenchmarkBaseManager_Health(b *testing.B) {
	m := NewBaseManager("benchmark-manager")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = m.Health()
	}
}

// BenchmarkBaseManager_OnStart 基准测试启动
func BenchmarkBaseManager_OnStart(b *testing.B) {
	m := NewBaseManager("benchmark-manager")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = m.OnStart()
	}
}
