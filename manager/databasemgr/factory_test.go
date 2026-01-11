package databasemgr

import (
	"errors"
	"testing"

	"com.litelake.litecore/common"
)

// MockConfigProvider 用于测试的模拟配置提供者
type MockConfigProvider struct {
	data map[string]any
	err  error
}

func (m *MockConfigProvider) Get(key string) (any, error) {
	if m == nil {
		return nil, errors.New("config provider is nil")
	}
	if m.err != nil {
		return nil, m.err
	}
	if val, ok := m.data[key]; ok {
		return val, nil
	}
	return nil, errors.New("key not found")
}

func (m *MockConfigProvider) Has(key string) bool {
	if m == nil || m.err != nil {
		return false
	}
	_, ok := m.data[key]
	return ok
}

func (m *MockConfigProvider) ConfigProviderName() string {
	return "mock"
}

// TestBuild_NoneDriver 测试 none 驱动
func TestBuild_NoneDriver(t *testing.T) {
	mgr, err := Build("none", nil)
	if err != nil {
		t.Fatalf("Build() error = %v", err)
	}

	if mgr == nil {
		t.Fatal("Build() returned nil")
	}

	if mgr.ManagerName() != "none" {
		t.Errorf("ManagerName() = %v, want 'none'", mgr.ManagerName())
	}
}

// TestBuild_SQLite 测试 SQLite 驱动
func TestBuild_SQLite(t *testing.T) {
	cfg := map[string]any{
		"dsn": ":memory:",
	}

	mgr, err := Build("sqlite", cfg)
	if err != nil {
		t.Fatalf("Build() error = %v", err)
	}

	if mgr == nil {
		t.Fatal("Build() returned nil")
	}

	if mgr.ManagerName() != "sqlite" {
		t.Errorf("ManagerName() = %v, want 'sqlite'", mgr.ManagerName())
	}

	// 验证实现了 common.BaseManager 接口
	var _ common.BaseManager = mgr

	// 清理
	_ = mgr.Close()
}

// TestBuild_SQLite_WithPoolConfig 测试 SQLite 带连接池配置
func TestBuild_SQLite_WithPoolConfig(t *testing.T) {
	cfg := map[string]any{
		"dsn": ":memory:",
		"pool_config": map[string]any{
			"max_open_conns": 1,
			"max_idle_conns": 1,
		},
	}

	mgr, err := Build("sqlite", cfg)
	if err != nil {
		t.Fatalf("Build() error = %v", err)
	}

	if mgr == nil {
		t.Fatal("Build() returned nil")
	}

	if mgr.ManagerName() != "sqlite" {
		t.Errorf("ManagerName() = %v, want 'sqlite'", mgr.ManagerName())
	}

	_ = mgr.Close()
}

// TestBuild_InvalidDriver 测试无效驱动
func TestBuild_InvalidDriver(t *testing.T) {
	_, err := Build("invalid", nil)
	if err == nil {
		t.Error("Build() should return error for invalid driver")
	}
}

// TestBuild_SQLite_MissingDSN 测试 SQLite 缺少 DSN
func TestBuild_SQLite_MissingDSN(t *testing.T) {
	cfg := map[string]any{}

	_, err := Build("sqlite", cfg)
	if err == nil {
		t.Error("Build() should return error for missing DSN")
	}
}

// TestBuild_ImplementsManagerInterface 测试实现了 Manager 接口
func TestBuild_ImplementsManagerInterface(t *testing.T) {
	cfg := map[string]any{
		"dsn": ":memory:",
	}

	mgr, err := Build("sqlite", cfg)
	if err != nil {
		t.Fatalf("Build() error = %v", err)
	}

	if mgr == nil {
		t.Fatal("Build() returned nil")
	}

	// 验证返回值实现了 common.BaseManager 接口
	var _ common.BaseManager = mgr

	// 测试接口方法
	_ = mgr.ManagerName()
	_ = mgr.Health()
	_ = mgr.OnStart()
	_ = mgr.OnStop()

	// 清理
	if dbMgr, ok := mgr.(DatabaseManager); ok {
		_ = dbMgr.Close()
	}
}

// BenchmarkBuild 基准测试 Build 方法
func BenchmarkBuild(b *testing.B) {
	cfg := map[string]any{
		"dsn": ":memory:",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		mgr, _ := Build("sqlite", cfg)
		if mgr != nil {
			if dbMgr, ok := mgr.(DatabaseManager); ok {
				_ = dbMgr.Close()
			}
		}
	}
}

// TestBuildWithConfigProvider_NoneDriver 测试使用 ConfigProvider 构建 none 驱动
func TestBuildWithConfigProvider_NoneDriver(t *testing.T) {
	provider := &MockConfigProvider{
		data: map[string]any{
			"database.driver": "none",
		},
	}

	mgr, err := BuildWithConfigProvider(provider)
	if err != nil {
		t.Fatalf("BuildWithConfigProvider() error = %v", err)
	}

	if mgr == nil {
		t.Fatal("BuildWithConfigProvider() returned nil")
	}

	if mgr.ManagerName() != "none" {
		t.Errorf("ManagerName() = %v, want 'none'", mgr.ManagerName())
	}
}

// TestBuildWithConfigProvider_SQLite 测试使用 ConfigProvider 构建 SQLite
func TestBuildWithConfigProvider_SQLite(t *testing.T) {
	provider := &MockConfigProvider{
		data: map[string]any{
			"database.driver": "sqlite",
			"database.sqlite_config": map[string]any{
				"dsn": ":memory:",
			},
		},
	}

	mgr, err := BuildWithConfigProvider(provider)
	if err != nil {
		t.Fatalf("BuildWithConfigProvider() error = %v", err)
	}

	if mgr == nil {
		t.Fatal("BuildWithConfigProvider() returned nil")
	}

	if mgr.ManagerName() != "sqlite" {
		t.Errorf("ManagerName() = %v, want 'sqlite'", mgr.ManagerName())
	}

	_ = mgr.Close()
}

// TestBuildWithConfigProvider_NilProvider 测试 nil provider
func TestBuildWithConfigProvider_NilProvider(t *testing.T) {
	_, err := BuildWithConfigProvider(nil)
	if err == nil {
		t.Error("BuildWithConfigProvider() should return error for nil provider")
	}
}

// TestBuildWithConfigProvider_InvalidDriver 测试无效驱动
func TestBuildWithConfigProvider_InvalidDriver(t *testing.T) {
	provider := &MockConfigProvider{
		data: map[string]any{
			"database.driver": "invalid",
		},
	}

	_, err := BuildWithConfigProvider(provider)
	if err == nil {
		t.Error("BuildWithConfigProvider() should return error for invalid driver")
	}
}

// TestBuildWithConfigProvider_MissingConfig 测试缺少配置
func TestBuildWithConfigProvider_MissingConfig(t *testing.T) {
	provider := &MockConfigProvider{
		data: map[string]any{
			"database.driver": "sqlite",
			// 缺少 sqlite_config
		},
	}

	_, err := BuildWithConfigProvider(provider)
	if err == nil {
		t.Error("BuildWithConfigProvider() should return error for missing config")
	}
}

// BenchmarkBuildWithConfigProvider 基准测试 BuildWithConfigProvider 方法
func BenchmarkBuildWithConfigProvider(b *testing.B) {
	provider := &MockConfigProvider{
		data: map[string]any{
			"database.driver": "sqlite",
			"database.sqlite_config": map[string]any{
				"dsn": ":memory:",
			},
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		mgr, _ := BuildWithConfigProvider(provider)
		if mgr != nil {
			_ = mgr.Close()
		}
	}
}
