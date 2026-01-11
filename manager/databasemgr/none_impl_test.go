package databasemgr

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"gorm.io/gorm"
)

// TestNewDatabaseManagerNoneImpl 测试创建 None 驱动
func TestNewDatabaseManagerNoneImpl(t *testing.T) {
	mgr := NewDatabaseManagerNoneImpl()

	if mgr == nil {
		t.Fatal("NewDatabaseManagerNoneImpl() returned nil")
	}

	if mgr.ManagerName() != "none" {
		t.Errorf("ManagerName() = %v, want 'none'", mgr.ManagerName())
	}

	if mgr.Driver() != "none" {
		t.Errorf("Driver() = %v, want 'none'", mgr.Driver())
	}
}

// TestNoneImpl_Health 测试健康检查
func TestNoneImpl_Health(t *testing.T) {
	mgr := NewDatabaseManagerNoneImpl()

	err := mgr.Health()
	if err == nil {
		t.Error("Health() should return error for none driver")
	}

	expectedErr := "database not available (none driver)"
	if err.Error() != expectedErr {
		t.Errorf("Health() error = %v, want %v", err.Error(), expectedErr)
	}
}

// TestNoneImpl_OnStart 测试 OnStart
func TestNoneImpl_OnStart(t *testing.T) {
	mgr := NewDatabaseManagerNoneImpl()

	err := mgr.OnStart()
	if err != nil {
		t.Errorf("OnStart() should return nil, got %v", err)
	}
}

// TestNoneImpl_OnStop 测试 OnStop
func TestNoneImpl_OnStop(t *testing.T) {
	mgr := NewDatabaseManagerNoneImpl()

	err := mgr.OnStop()
	if err != nil {
		t.Errorf("OnStop() should return nil, got %v", err)
	}
}

// TestNoneImpl_DB 测试 DB 方法
func TestNoneImpl_DB(t *testing.T) {
	mgr := NewDatabaseManagerNoneImpl()

	db := mgr.DB()
	if db != nil {
		t.Errorf("DB() should return nil, got %v", db)
	}
}

// TestNoneImpl_Model 测试 Model 方法
func TestNoneImpl_Model(t *testing.T) {
	mgr := NewDatabaseManagerNoneImpl()

	result := mgr.Model("test")
	if result != nil {
		t.Errorf("Model() should return nil, got %v", result)
	}
}

// TestNoneImpl_Table 测试 Table 方法
func TestNoneImpl_Table(t *testing.T) {
	mgr := NewDatabaseManagerNoneImpl()

	result := mgr.Table("test_table")
	if result != nil {
		t.Errorf("Table() should return nil, got %v", result)
	}
}

// TestNoneImpl_WithContext 测试 WithContext 方法
func TestNoneImpl_WithContext(t *testing.T) {
	mgr := NewDatabaseManagerNoneImpl()

	ctx := context.Background()
	result := mgr.WithContext(ctx)
	if result != nil {
		t.Errorf("WithContext() should return nil, got %v", result)
	}
}

// TestNoneImpl_Transaction 测试 Transaction 方法
func TestNoneImpl_Transaction(t *testing.T) {
	mgr := NewDatabaseManagerNoneImpl()

	err := mgr.Transaction(func(db *gorm.DB) error {
		return nil
	})
	if err == nil {
		t.Error("Transaction() should return error for none driver")
	}

	expectedErr := "database not available (none driver)"
	if err.Error() != expectedErr {
		t.Errorf("Transaction() error = %v, want %v", err.Error(), expectedErr)
	}
}

// TestNoneImpl_Begin 测试 Begin 方法
func TestNoneImpl_Begin(t *testing.T) {
	mgr := NewDatabaseManagerNoneImpl()

	result := mgr.Begin()
	if result != nil {
		t.Errorf("Begin() should return nil, got %v", result)
	}
}

// TestNoneImpl_AutoMigrate 测试 AutoMigrate 方法
func TestNoneImpl_AutoMigrate(t *testing.T) {
	mgr := NewDatabaseManagerNoneImpl()

	type TestModel struct {
		ID   uint
		Name string
	}

	err := mgr.AutoMigrate(&TestModel{})
	if err == nil {
		t.Error("AutoMigrate() should return error for none driver")
	}

	expectedErr := "database not available (none driver)"
	if err.Error() != expectedErr {
		t.Errorf("AutoMigrate() error = %v, want %v", err.Error(), expectedErr)
	}
}

// TestNoneImpl_Migrator 测试 Migrator 方法
func TestNoneImpl_Migrator(t *testing.T) {
	mgr := NewDatabaseManagerNoneImpl()

	migrator := mgr.Migrator()
	if migrator != nil {
		t.Errorf("Migrator() should return nil, got %v", migrator)
	}
}

// TestNoneImpl_Ping 测试 Ping 方法
func TestNoneImpl_Ping(t *testing.T) {
	mgr := NewDatabaseManagerNoneImpl()

	ctx := context.Background()
	err := mgr.Ping(ctx)
	if err == nil {
		t.Error("Ping() should return error for none driver")
	}

	expectedErr := "database not available (none driver)"
	if err.Error() != expectedErr {
		t.Errorf("Ping() error = %v, want %v", err.Error(), expectedErr)
	}
}

// TestNoneImpl_Stats 测试 Stats 方法
func TestNoneImpl_Stats(t *testing.T) {
	mgr := NewDatabaseManagerNoneImpl()

	stats := mgr.Stats()
	if stats != (sql.DBStats{}) {
		t.Errorf("Stats() should return empty DBStats, got %v", stats)
	}
}

// TestNoneImpl_Close 测试 Close 方法
func TestNoneImpl_Close(t *testing.T) {
	mgr := NewDatabaseManagerNoneImpl()

	err := mgr.Close()
	if err != nil {
		t.Errorf("Close() should return nil, got %v", err)
	}

	// 多次调用 Close 应该不会出错
	err = mgr.Close()
	if err != nil {
		t.Errorf("Close() called multiple times should return nil, got %v", err)
	}
}

// TestNoneImpl_Exec 测试 Exec 方法
func TestNoneImpl_Exec(t *testing.T) {
	mgr := NewDatabaseManagerNoneImpl()

	result := mgr.Exec("SELECT 1")
	if result != nil {
		t.Errorf("Exec() should return nil, got %v", result)
	}
}

// TestNoneImpl_Raw 测试 Raw 方法
func TestNoneImpl_Raw(t *testing.T) {
	mgr := NewDatabaseManagerNoneImpl()

	result := mgr.Raw("SELECT 1")
	if result != nil {
		t.Errorf("Raw() should return nil, got %v", result)
	}
}

// TestNoneImpl_InterfaceCompliance 测试接口实现
func TestNoneImpl_InterfaceCompliance(t *testing.T) {
	mgr := NewDatabaseManagerNoneImpl()

	// 验证实现了 DatabaseManager 接口
	var _ DatabaseManager = mgr

	// 验证所有方法都可以调用（不应该 panic）
	_ = mgr.ManagerName()
	_ = mgr.Health()
	_ = mgr.OnStart()
	_ = mgr.OnStop()
	_ = mgr.DB()
	_ = mgr.Model(nil)
	_ = mgr.Table("")
	_ = mgr.WithContext(nil)
	_ = mgr.Transaction(nil)
	_ = mgr.Begin()
	_ = mgr.AutoMigrate()
	_ = mgr.Migrator()
	_ = mgr.Driver()
	_ = mgr.Ping(nil)
	_ = mgr.Stats()
	_ = mgr.Close()
	_ = mgr.Exec("")
	_ = mgr.Raw("")
}

// TestNoneImpl_ErrorConsistency 测试错误消息一致性
func TestNoneImpl_ErrorConsistency(t *testing.T) {
	mgr := NewDatabaseManagerNoneImpl()

	expectedErr := errors.New("database not available (none driver)")

	tests := []struct {
		name string
		fn   func() error
	}{
		{"Health", func() error { return mgr.Health() }},
		{"Ping", func() error { return mgr.Ping(context.Background()) }},
		{"Transaction", func() error { return mgr.Transaction(nil) }},
		{"AutoMigrate", func() error { return mgr.AutoMigrate() }},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.fn()
			if err == nil {
				t.Error("expected error, got nil")
				return
			}
			if err.Error() != expectedErr.Error() {
				t.Errorf("error = %v, want %v", err.Error(), expectedErr.Error())
			}
		})
	}
}

// TestNoneImpl_NilReturns 测试所有返回 *gorm.DB 的方法都返回 nil
func TestNoneImpl_NilReturns(t *testing.T) {
	mgr := NewDatabaseManagerNoneImpl()

	tests := []struct {
		name string
		got  *gorm.DB
	}{
		{"DB", mgr.DB()},
		{"Model", mgr.Model(nil)},
		{"Table", mgr.Table("")},
		{"WithContext", mgr.WithContext(nil)},
		{"Begin", mgr.Begin()},
		{"Exec", mgr.Exec("")},
		{"Raw", mgr.Raw("")},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.got != nil {
				t.Errorf("%s() should return nil", tt.name)
			}
		})
	}
}

// BenchmarkNoneImpl_Operations 基准测试 None 驱动操作
func BenchmarkNoneImpl_Operations(b *testing.B) {
	mgr := NewDatabaseManagerNoneImpl()

	b.Run("Health", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = mgr.Health()
		}
	})

	b.Run("ManagerName", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = mgr.ManagerName()
		}
	})

	b.Run("Driver", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = mgr.Driver()
		}
	})
}
