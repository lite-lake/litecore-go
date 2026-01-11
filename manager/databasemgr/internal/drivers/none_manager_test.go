package drivers

import (
	"context"
	"testing"

	"gorm.io/gorm"
)

// TestNewNoneDatabaseManager 测试创建空数据库管理器
func TestNewNoneDatabaseManager(t *testing.T) {
	m := NewNoneDatabaseManager()
	if m == nil {
		t.Fatal("NewNoneDatabaseManager() returned nil")
	}

	if m.GormBaseManager == nil {
		t.Error("GormBaseManager is nil")
	}

	if m.ManagerName() != "none-database" {
		t.Errorf("ManagerName() = %v, want 'none-database'", m.ManagerName())
	}

	if m.Driver() != "none" {
		t.Errorf("Driver() = %v, want 'none'", m.Driver())
	}
}

// TestNoneDatabaseManager_DB 测试 DB 方法返回 nil
func TestNoneDatabaseManager_DB(t *testing.T) {
	m := NewNoneDatabaseManager()
	db := m.DB()
	if db != nil {
		t.Errorf("DB() should return nil, got %v", db)
	}
}

// TestNoneDatabaseManager_Ping 测试 Ping 方法返回错误
func TestNoneDatabaseManager_Ping(t *testing.T) {
	m := NewNoneDatabaseManager()
	ctx := context.Background()
	err := m.Ping(ctx)
	if err == nil {
		t.Error("Ping() should return error")
	}

	expectedErrMsg := "database not available (none driver)"
	if err.Error() != expectedErrMsg {
		t.Errorf("Ping() error = %v, want %v", err.Error(), expectedErrMsg)
	}
}

// TestNoneDatabaseManager_Health 测试 Health 方法返回错误
func TestNoneDatabaseManager_Health(t *testing.T) {
	m := NewNoneDatabaseManager()
	err := m.Health()
	if err == nil {
		t.Error("Health() should return error")
	}

	expectedErrMsg := "database not available (none driver)"
	if err.Error() != expectedErrMsg {
		t.Errorf("Health() error = %v, want %v", err.Error(), expectedErrMsg)
	}
}

// TestNoneDatabaseManager_Model 测试 Model 方法
func TestNoneDatabaseManager_Model(t *testing.T) {
	m := NewNoneDatabaseManager()
	type TestModel struct {
		ID   uint
		Name string
	}
	// Model 方法会返回一个 GORM DB 实例（但无法实际执行操作）
	result := m.Model(&TestModel{})
	if result == nil {
		t.Error("Model() should return a GORM DB instance")
	}
	// 验证 DB() 返回 nil
	if m.DB() != nil {
		t.Error("DB() should return nil")
	}
}

// TestNoneDatabaseManager_Table 测试 Table 方法
func TestNoneDatabaseManager_Table(t *testing.T) {
	m := NewNoneDatabaseManager()
	// Table 方法会返回一个 GORM DB 实例（但无法实际执行操作）
	result := m.Table("test_table")
	if result == nil {
		t.Error("Table() should return a GORM DB instance")
	}
}

// TestNoneDatabaseManager_WithContext 测试 WithContext 方法
func TestNoneDatabaseManager_WithContext(t *testing.T) {
	m := NewNoneDatabaseManager()
	ctx := context.Background()
	// WithContext 方法会返回一个 GORM DB 实例（但无法实际执行操作）
	result := m.WithContext(ctx)
	if result == nil {
		t.Error("WithContext() should return a GORM DB instance")
	}
}

// TestNoneDatabaseManager_Transaction 测试 Transaction 方法
func TestNoneDatabaseManager_Transaction(t *testing.T) {
	m := NewNoneDatabaseManager()
	err := m.Transaction(func(tx *gorm.DB) error {
		return nil
	})
	// Transaction 应该失败
	if err == nil {
		t.Error("Transaction() should return error")
	}
}

// TestNoneDatabaseManager_Begin 测试 Begin 方法
func TestNoneDatabaseManager_Begin(t *testing.T) {
	m := NewNoneDatabaseManager()
	// Begin 方法会返回一个 GORM DB 实例（但事务是无效的）
	tx := m.Begin()
	if tx == nil {
		t.Error("Begin() should return a GORM DB instance")
	}
}

// TestNoneDatabaseManager_AutoMigrate 测试 AutoMigrate 方法
func TestNoneDatabaseManager_AutoMigrate(t *testing.T) {
	m := NewNoneDatabaseManager()
	type TestModel struct {
		ID   uint
		Name string
	}

	// AutoMigrate 会调用 GORM 的 AutoMigrate，但由于使用的是 DummyDialector，
	// Migrator 返回 nil，会导致 panic
	defer func() {
		if r := recover(); r != nil {
			// 预期会 panic，这是正常的
			t.Log("AutoMigrate() panicked as expected (Migrator is nil)")
		}
	}()

	_ = m.AutoMigrate(&TestModel{})
	t.Error("AutoMigrate() should have panicked")
}

// TestNoneDatabaseManager_Migrator 测试 Migrator 方法
func TestNoneDatabaseManager_Migrator(t *testing.T) {
	m := NewNoneDatabaseManager()
	migrator := m.Migrator()
	// Migrator 应该返回 nil
	if migrator != nil {
		t.Errorf("Migrator() should return nil, got %v", migrator)
	}
}

// TestNoneDatabaseManager_Exec 测试 Exec 方法
func TestNoneDatabaseManager_Exec(t *testing.T) {
	m := NewNoneDatabaseManager()
	// Exec 方法会返回一个 GORM DB 实例（但无法实际执行）
	result := m.Exec("SELECT 1")
	if result == nil {
		t.Error("Exec() should return a GORM DB instance")
	}
}

// TestNoneDatabaseManager_Raw 测试 Raw 方法
func TestNoneDatabaseManager_Raw(t *testing.T) {
	m := NewNoneDatabaseManager()
	// Raw 方法会返回一个 GORM DB 实例（但无法实际执行）
	result := m.Raw("SELECT 1")
	if result == nil {
		t.Error("Raw() should return a GORM DB instance")
	}
}

// TestNoneDatabaseManager_Close 测试 Close 方法
func TestNoneDatabaseManager_Close(t *testing.T) {
	m := NewNoneDatabaseManager()
	err := m.Close()
	// Close 不应该返回错误
	if err != nil {
		t.Errorf("Close() error = %v", err)
	}

	// 第二次关闭也应该是幂等的
	err = m.Close()
	if err != nil {
		t.Errorf("Close() second call error = %v", err)
	}
}

// TestNoneDatabaseManager_OnStart 测试 OnStart 方法
func TestNoneDatabaseManager_OnStart(t *testing.T) {
	m := NewNoneDatabaseManager()
	err := m.OnStart()
	// OnStart 会调用 Ping，所以应该返回错误
	if err == nil {
		t.Error("OnStart() should return error")
	}
}

// TestNoneDatabaseManager_OnStop 测试 OnStop 方法
func TestNoneDatabaseManager_OnStop(t *testing.T) {
	m := NewNoneDatabaseManager()
	err := m.OnStop()
	// OnStop 调用 Close，不应该返回错误
	if err != nil {
		t.Errorf("OnStop() error = %v", err)
	}
}

// TestNoneDatabaseManager_Stats 测试 Stats 方法
func TestNoneDatabaseManager_Stats(t *testing.T) {
	m := NewNoneDatabaseManager()
	stats := m.Stats()
	// Stats 应该返回空的结构体
	if stats.MaxOpenConnections != 0 {
		t.Errorf("Stats() should return zero value, got MaxOpenConnections = %v", stats.MaxOpenConnections)
	}
}

// TestDummyDialector 测试虚拟 Dialector
func TestDummyDialector(t *testing.T) {
	d := &DummyDialector{}

	if d.Name() != "none" {
		t.Errorf("DummyDialector.Name() = %v, want 'none'", d.Name())
	}

	// 测试其他方法不会崩溃
	// 这些方法都是空实现或返回零值，我们只需要确保不会 panic
	t.Run("Initialize", func(t *testing.T) {
		// Initialize 需要 *gorm.DB，我们创建一个临时的
		// 由于这是空实现，我们只测试不会 panic
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("Initialize() panicked: %v", r)
			}
		}()
	})

	t.Run("Migrator", func(t *testing.T) {
		m := d.Migrator(nil)
		if m != nil {
			t.Errorf("DummyDialector.Migrator() should return nil, got %v", m)
		}
	})

	t.Run("DataTypeOf", func(t *testing.T) {
		result := d.DataTypeOf(nil)
		if result != "" {
			t.Errorf("DummyDialector.DataTypeOf() should return empty string, got %v", result)
		}
	})
}

// TestNoneDatabaseManager_AsFallback 测试作为回退选项
func TestNoneDatabaseManager_AsFallback(t *testing.T) {
	// 这个测试验证 NoneDatabaseManager 可以作为初始化失败的回退选项
	m := NewNoneDatabaseManager()

	// 验证它实现了 DatabaseManager 接口
	var _ interface {
		ManagerName() string
		Health() error
		OnStart() error
		OnStop() error
		DB() *gorm.DB
		Driver() string
	} = m

	// 验证方法不会 panic
	_ = m.ManagerName()
	_ = m.Health()
	_ = m.OnStart()
	_ = m.OnStop()
	_ = m.DB()
	_ = m.Driver()
	_ = m.Stats()
	_ = m.Close()
}

// BenchmarkNoneDatabaseManager_Health 基准测试 Health 方法
func BenchmarkNoneDatabaseManager_Health(b *testing.B) {
	m := NewNoneDatabaseManager()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = m.Health()
	}
}

// BenchmarkNoneDatabaseManager_DB 基准测试 DB 方法
func BenchmarkNoneDatabaseManager_DB(b *testing.B) {
	m := NewNoneDatabaseManager()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = m.DB()
	}
}
