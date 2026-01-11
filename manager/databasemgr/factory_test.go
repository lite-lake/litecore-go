package databasemgr

import (
	"testing"

	"com.litelake.litecore/common"
	"com.litelake.litecore/manager/databasemgr/internal/config"
)

// TestNewFactory 测试创建工厂
func TestNewFactory(t *testing.T) {
	f := NewFactory()
	if f == nil {
		t.Fatal("NewFactory() returned nil")
	}
}

// TestFactory_Build_InvalidConfig 测试无效配置
func TestFactory_Build_InvalidConfig(t *testing.T) {
	f := NewFactory()

	tests := []struct {
		name   string
		driver string
		cfg    map[string]any
	}{
		{
			name:   "nil 配置",
			driver: "",
			cfg:    nil,
		},
		{
			name:   "空配置",
			driver: "",
			cfg:    map[string]any{},
		},
		{
			name:   "无效驱动",
			driver: "invalid",
			cfg:    map[string]any{},
		},
		{
			name: "SQLite 缺少 DSN",
			driver: "sqlite",
			cfg: map[string]any{
				"sqlite_config": map[string]any{},
			},
		},
		{
			name: "MySQL 缺少 DSN",
			driver: "mysql",
			cfg: map[string]any{
				"mysql_config": map[string]any{},
			},
		},
		{
			name: "PostgreSQL 缺少 DSN",
			driver: "postgresql",
			cfg: map[string]any{
				"postgresql_config": map[string]any{},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mgr := f.Build(tt.driver, tt.cfg)
			// 无效配置应该返回 NoneDatabaseManager
			if mgr == nil {
				t.Error("Build() should return NoneDatabaseManager for invalid config")
			}
			if mgr.ManagerName() != "none-database" {
				t.Errorf("Build() should return NoneDatabaseManager, got %s", mgr.ManagerName())
			}
		})
	}
}

// TestFactory_Build_NoneDriver 测试 none 驱动
func TestFactory_Build_NoneDriver(t *testing.T) {
	f := NewFactory()

	mgr := f.Build("none", nil)
	if mgr == nil {
		t.Fatal("Build() returned nil")
	}

	if mgr.ManagerName() != "none-database" {
		t.Errorf("ManagerName() = %v, want 'none-database'", mgr.ManagerName())
	}
}

// TestFactory_Build_SQLite 测试 SQLite 驱动
func TestFactory_Build_SQLite(t *testing.T) {
	f := NewFactory()

	cfg := map[string]any{
		"driver": "sqlite",
		"sqlite_config": map[string]any{
			"dsn": ":memory:",
		},
	}

	mgr := f.Build("", cfg)
	if mgr == nil {
		t.Fatal("Build() returned nil")
	}

	if mgr.ManagerName() != "sqlite-database" {
		t.Errorf("ManagerName() = %v, want 'sqlite-database'", mgr.ManagerName())
	}

	// 验证实现了 common.Manager 接口
	var _ common.Manager = mgr
}

// TestFactory_Build_SQLite_WithPoolConfig 测试 SQLite 带连接池配置
func TestFactory_Build_SQLite_WithPoolConfig(t *testing.T) {
	f := NewFactory()

	cfg := map[string]any{
		"driver": "sqlite",
		"sqlite_config": map[string]any{
			"dsn": ":memory:",
			"pool_config": map[string]any{
				"max_open_conns": 1,
				"max_idle_conns": 1,
			},
		},
	}

	mgr := f.Build("", cfg)
	if mgr == nil {
		t.Fatal("Build() returned nil")
	}

	if mgr.ManagerName() != "sqlite-database" {
		t.Errorf("ManagerName() = %v, want 'sqlite-database'", mgr.ManagerName())
	}
}

// TestFactory_Build_DriverOverride 测试驱动覆盖
func TestFactory_Build_DriverOverride(t *testing.T) {
	f := NewFactory()

	// 通过参数指定驱动，配置中提供 sqlite_config
	cfg := map[string]any{
		"sqlite_config": map[string]any{
			"dsn": ":memory:",
		},
	}

	mgr := f.Build("sqlite", cfg)
	if mgr == nil {
		t.Fatal("Build() returned nil")
	}

	if mgr.ManagerName() != "sqlite-database" {
		t.Errorf("ManagerName() = %v, want 'sqlite-database'", mgr.ManagerName())
	}
}

// TestFactory_BuildWithConfig_SQLite 测试使用配置结构体构建 SQLite
func TestFactory_BuildWithConfig_SQLite(t *testing.T) {
	f := NewFactory()

	cfg := &config.DatabaseConfig{
		Driver: "sqlite",
		SQLiteConfig: &config.SQLiteConfig{
			DSN: ":memory:",
		},
	}

	mgr, err := f.BuildWithConfig(cfg)
	if err != nil {
		t.Fatalf("BuildWithConfig() error = %v", err)
	}

	if mgr == nil {
		t.Fatal("BuildWithConfig() returned nil")
	}

	if mgr.ManagerName() != "sqlite-database" {
		t.Errorf("ManagerName() = %v, want 'sqlite-database'", mgr.ManagerName())
	}

	_ = mgr.Close()
}

// TestFactory_BuildWithConfig_InvalidConfig 测试无效配置结构体
func TestFactory_BuildWithConfig_InvalidConfig(t *testing.T) {
	f := NewFactory()

	tests := []struct {
		name    string
		config  *config.DatabaseConfig
		wantErr bool
	}{
		{
			name:    "nil 配置",
			config:  nil,
			wantErr: true,
		},
		{
			name: "空驱动",
			config: &config.DatabaseConfig{
				Driver: "",
			},
			wantErr: true,
		},
		{
			name: "SQLite 配置为空",
			config: &config.DatabaseConfig{
				Driver: "sqlite",
			},
			wantErr: true,
		},
		{
			name: "无效驱动",
			config: &config.DatabaseConfig{
				Driver: "invalid",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mgr, err := f.BuildWithConfig(tt.config)
			if (err != nil) != tt.wantErr {
				t.Errorf("BuildWithConfig() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err == nil && mgr != nil {
				_ = mgr.Close()
			}
		})
	}
}

// TestFactory_BuildWithConfig_AllDrivers 测试所有驱动类型
func TestFactory_BuildWithConfig_AllDrivers(t *testing.T) {
	f := NewFactory()

	tests := []struct {
		name          string
		config        *config.DatabaseConfig
		expectSuccess bool
	}{
		{
			name: "SQLite",
			config: &config.DatabaseConfig{
				Driver: "sqlite",
				SQLiteConfig: &config.SQLiteConfig{
					DSN: ":memory:",
				},
			},
			expectSuccess: true,
		},
		{
			name: "none",
			config: &config.DatabaseConfig{
				Driver: "none",
			},
			expectSuccess: true,
		},
		{
			name: "MySQL（需要真实连接）",
			config: &config.DatabaseConfig{
				Driver: "mysql",
				MySQLConfig: &config.MySQLConfig{
					DSN: "root:password@tcp(localhost:3306)/test",
				},
			},
			expectSuccess: false,
		},
		{
			name: "PostgreSQL（需要真实连接）",
			config: &config.DatabaseConfig{
				Driver: "postgresql",
				PostgreSQLConfig: &config.PostgreSQLConfig{
					DSN: "host=localhost port=5432",
				},
			},
			expectSuccess: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mgr, err := f.BuildWithConfig(tt.config)
			if tt.expectSuccess {
				if err != nil {
					t.Fatalf("BuildWithConfig() error = %v", err)
				}
				if mgr == nil {
					t.Fatal("BuildWithConfig() returned nil")
				}
				_ = mgr.Close()
			} else {
				if err == nil {
					_ = mgr.Close()
					t.Skip("Database is available, skipping error test")
				}
			}
		})
	}
}

// TestFactory_ImplementsManagerInterface 测试工厂实现了 Manager 接口
func TestFactory_ImplementsManagerInterface(t *testing.T) {
	f := NewFactory()

	cfg := map[string]any{
		"driver": "sqlite",
		"sqlite_config": map[string]any{
			"dsn": ":memory:",
		},
	}

	mgr := f.Build("", cfg)
	if mgr == nil {
		t.Fatal("Build() returned nil")
	}

	// 验证返回值实现了 common.Manager 接口
	var _ common.Manager = mgr

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

// BenchmarkFactory_Build 基准测试 Build 方法
func BenchmarkFactory_Build(b *testing.B) {
	f := NewFactory()

	cfg := map[string]any{
		"driver": "sqlite",
		"sqlite_config": map[string]any{
			"dsn": ":memory:",
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		mgr := f.Build("", cfg)
		if mgr != nil {
			if dbMgr, ok := mgr.(DatabaseManager); ok {
				_ = dbMgr.Close()
			}
		}
	}
}

// BenchmarkFactory_BuildWithConfig 基准测试 BuildWithConfig 方法
func BenchmarkFactory_BuildWithConfig(b *testing.B) {
	f := NewFactory()

	cfg := &config.DatabaseConfig{
		Driver: "sqlite",
		SQLiteConfig: &config.SQLiteConfig{
			DSN: ":memory:",
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		mgr, _ := f.BuildWithConfig(cfg)
		if mgr != nil {
			_ = mgr.Close()
		}
	}
}
