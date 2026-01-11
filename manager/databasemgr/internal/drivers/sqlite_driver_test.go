package drivers

import (
	"os"
	"testing"

	"gorm.io/gorm"

	"com.litelake.litecore/manager/databasemgr/internal/config"
)

// TestNewSQLiteManager_InvalidConfig 测试无效配置
func TestNewSQLiteManager_InvalidConfig(t *testing.T) {
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
			name: "空 DSN",
			config: &config.DatabaseConfig{
				Driver:      "sqlite",
				SQLiteConfig: &config.SQLiteConfig{},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewSQLiteManager(tt.config)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewSQLiteManager() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestNewSQLiteManager_MemoryDB 测试内存数据库
func TestNewSQLiteManager_MemoryDB(t *testing.T) {
	cfg := &config.DatabaseConfig{
		Driver: "sqlite",
		SQLiteConfig: &config.SQLiteConfig{
			DSN: ":memory:",
		},
	}

	mgr, err := NewSQLiteManager(cfg)
	if err != nil {
		t.Fatalf("NewSQLiteManager() error = %v", err)
	}

	if mgr == nil {
		t.Fatal("NewSQLiteManager() returned nil manager")
	}

	if mgr.ManagerName() != "sqlite-database" {
		t.Errorf("ManagerName() = %v, want 'sqlite-database'", mgr.ManagerName())
	}

	if mgr.Driver() != "sqlite" {
		t.Errorf("Driver() = %v, want 'sqlite'", mgr.Driver())
	}

	// 清理
	_ = mgr.Close()
}

// TestNewSQLiteManager_FileDB 测试文件数据库
func TestNewSQLiteManager_FileDB(t *testing.T) {
	tmpFile := "/tmp/test_sqlite_manager.db"
	defer os.Remove(tmpFile)

	cfg := &config.DatabaseConfig{
		Driver: "sqlite",
		SQLiteConfig: &config.SQLiteConfig{
			DSN: tmpFile,
		},
	}

	mgr, err := NewSQLiteManager(cfg)
	if err != nil {
		t.Fatalf("NewSQLiteManager() error = %v", err)
	}

	if mgr == nil {
		t.Fatal("NewSQLiteManager() returned nil manager")
	}

	// 验证文件已创建
	if _, err := os.Stat(tmpFile); os.IsNotExist(err) {
		t.Error("Database file was not created")
	}

	// 清理
	_ = mgr.Close()
	_ = os.Remove(tmpFile)
}

// TestNewSQLiteManager_SharedCache 测试共享缓存模式
func TestNewSQLiteManager_SharedCache(t *testing.T) {
	cfg := &config.DatabaseConfig{
		Driver: "sqlite",
		SQLiteConfig: &config.SQLiteConfig{
			DSN: "file::memory:?cache=shared",
		},
	}

	mgr, err := NewSQLiteManager(cfg)
	if err != nil {
		t.Fatalf("NewSQLiteManager() error = %v", err)
	}

	if mgr == nil {
		t.Fatal("NewSQLiteManager() returned nil manager")
	}

	_ = mgr.Close()
}

// TestNewSQLiteManager_WithPoolConfig 测试带连接池配置
func TestNewSQLiteManager_WithPoolConfig(t *testing.T) {
	cfg := &config.DatabaseConfig{
		Driver: "sqlite",
		SQLiteConfig: &config.SQLiteConfig{
			DSN: ":memory:",
			PoolConfig: &config.PoolConfig{
				MaxOpenConns:    1,
				MaxIdleConns:    1,
				ConnMaxLifetime: 0,
				ConnMaxIdleTime: 0,
			},
		},
	}

	mgr, err := NewSQLiteManager(cfg)
	if err != nil {
		t.Fatalf("NewSQLiteManager() error = %v", err)
	}

	if mgr == nil {
		t.Fatal("NewSQLiteManager() returned nil")
	}

	_ = mgr.Close()
}

// TestSQLiteManager_DatabaseOperations 测试数据库操作
func TestSQLiteManager_DatabaseOperations(t *testing.T) {
	cfg := &config.DatabaseConfig{
		Driver: "sqlite",
		SQLiteConfig: &config.SQLiteConfig{
			DSN: ":memory:",
		},
	}

	mgr, err := NewSQLiteManager(cfg)
	if err != nil {
		t.Fatalf("NewSQLiteManager() error = %v", err)
	}
	defer mgr.Close()

	// 创建测试表
	type TestModel struct {
		ID   uint
		Name string
	}

	// 测试 AutoMigrate
	if err := mgr.AutoMigrate(&TestModel{}); err != nil {
		t.Errorf("AutoMigrate() error = %v", err)
	}

	// 测试创建记录
	db := mgr.DB()
	result := db.Create(&TestModel{Name: "test"})
	if result.Error != nil {
		t.Errorf("Create() error = %v", result.Error)
	}

	// 测试查询
	var models []TestModel
	result = db.Find(&models)
	if result.Error != nil {
		t.Errorf("Find() error = %v", result.Error)
	}

	if len(models) != 1 {
		t.Errorf("Expected 1 record, got %d", len(models))
	}
}

// TestSQLiteManager_Transaction 测试事务
func TestSQLiteManager_Transaction(t *testing.T) {
	cfg := &config.DatabaseConfig{
		Driver: "sqlite",
		SQLiteConfig: &config.SQLiteConfig{
			DSN: ":memory:",
		},
	}

	mgr, err := NewSQLiteManager(cfg)
	if err != nil {
		t.Fatalf("NewSQLiteManager() error = %v", err)
	}
	defer mgr.Close()

	type TestModel struct {
		ID   uint
		Name string
	}

	// 创建表
	if err := mgr.AutoMigrate(&TestModel{}); err != nil {
		t.Fatalf("AutoMigrate() error = %v", err)
	}

	// 测试成功的事务
	err = mgr.Transaction(func(tx *gorm.DB) error {
		return tx.Create(&TestModel{Name: "test"}).Error
	})
	if err != nil {
		t.Errorf("Transaction() error = %v", err)
	}

	// 验证记录已创建
	var count int64
	mgr.DB().Model(&TestModel{}).Count(&count)
	if count != 1 {
		t.Errorf("Expected 1 record, got %d", count)
	}
}

// TestSQLiteManager_Lifecycle 测试生命周期方法
func TestSQLiteManager_Lifecycle(t *testing.T) {
	cfg := &config.DatabaseConfig{
		Driver: "sqlite",
		SQLiteConfig: &config.SQLiteConfig{
			DSN: ":memory:",
		},
	}

	mgr, err := NewSQLiteManager(cfg)
	if err != nil {
		t.Fatalf("NewSQLiteManager() error = %v", err)
	}

	// 测试 OnStart
	if err := mgr.OnStart(); err != nil {
		t.Errorf("OnStart() error = %v", err)
	}

	// 测试 Health
	if err := mgr.Health(); err != nil {
		t.Errorf("Health() error = %v", err)
	}

	// 测试 Stats
	stats := mgr.Stats()
	if stats.MaxOpenConnections == 0 {
		// SQLite 的默认配置
		t.Log("MaxOpenConnections is 0")
	}

	// 测试 OnStop
	if err := mgr.OnStop(); err != nil {
		t.Errorf("OnStop() error = %v", err)
	}
}

// TestSQLiteManager_ImplementsDatabaseManager 测试实现接口
func TestSQLiteManager_ImplementsDatabaseManager(t *testing.T) {
	cfg := &config.DatabaseConfig{
		Driver: "sqlite",
		SQLiteConfig: &config.SQLiteConfig{
			DSN: ":memory:",
		},
	}

	mgr, err := NewSQLiteManager(cfg)
	if err != nil {
		t.Fatalf("NewSQLiteManager() error = %v", err)
	}
	defer mgr.Close()

	// 验证基本方法
	_ = mgr.ManagerName()
	_ = mgr.Driver()
	_ = mgr.DB()
	_ = mgr.Stats()
}

// TestSQLiteManager_Exec 测试执行原生 SQL
func TestSQLiteManager_Exec(t *testing.T) {
	cfg := &config.DatabaseConfig{
		Driver: "sqlite",
		SQLiteConfig: &config.SQLiteConfig{
			DSN: ":memory:",
		},
	}

	mgr, err := NewSQLiteManager(cfg)
	if err != nil {
		t.Fatalf("NewSQLiteManager() error = %v", err)
	}
	defer mgr.Close()

	// 测试创建表
	result := mgr.Exec("CREATE TABLE IF NOT EXISTS test (id INTEGER, name TEXT)")
	if result.Error != nil {
		t.Errorf("Exec() error = %v", result.Error)
	}

	// 测试插入
	result = mgr.Exec("INSERT INTO test (id, name) VALUES (1, 'test')")
	if result.Error != nil {
		t.Errorf("Exec() INSERT error = %v", result.Error)
	}

	// 测试查询
	rows, err := mgr.Raw("SELECT * FROM test").Rows()
	if err != nil {
		t.Errorf("Raw() error = %v", err)
	}
	defer rows.Close()

	hasRow := rows.Next()
	if !hasRow {
		t.Error("Expected at least one row")
	}
}

// TestSQLiteManager_ConcurrentAccess 测试并发访问
func TestSQLiteManager_ConcurrentAccess(t *testing.T) {
	cfg := &config.DatabaseConfig{
		Driver: "sqlite",
		SQLiteConfig: &config.SQLiteConfig{
			DSN: ":memory:",
		},
	}

	mgr, err := NewSQLiteManager(cfg)
	if err != nil {
		t.Fatalf("NewSQLiteManager() error = %v", err)
	}
	defer mgr.Close()

	done := make(chan bool, 10)

	// 并发读取
	for i := 0; i < 10; i++ {
		go func() {
			_ = mgr.DB()
			_ = mgr.Driver()
			_ = mgr.Stats()
			_ = mgr.Health()
			done <- true
		}()
	}

	// 等待所有 goroutine 完成
	for i := 0; i < 10; i++ {
		<-done
	}
}

// TestSQLiteManager_DSNVariants 测试不同的 DSN 格式
func TestSQLiteManager_DSNVariants(t *testing.T) {
	tests := []struct {
		name    string
		dsn     string
		wantErr bool
	}{
		{
			name:    "内存数据库",
			dsn:     ":memory:",
			wantErr: false,
		},
		{
			name:    "共享缓存",
			dsn:     "file::memory:?cache=shared",
			wantErr: false,
		},
		{
			name:    "读写模式",
			dsn:     "file::memory:?mode=memory",
			wantErr: false,
		},
		{
			name:    "文件路径",
			dsn:     "/tmp/test.db",
			wantErr: false,
		},
		{
			name:    "相对路径",
			dsn:     "./test.db",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &config.DatabaseConfig{
				Driver: "sqlite",
				SQLiteConfig: &config.SQLiteConfig{
					DSN: tt.dsn,
				},
			}

			mgr, err := NewSQLiteManager(cfg)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewSQLiteManager() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err == nil && mgr != nil {
				_ = mgr.Close()
			}

			// 清理可能创建的文件
			if tt.dsn != "" && tt.dsn[0] != ':' {
				os.Remove(tt.dsn)
			}
		})
	}
}
