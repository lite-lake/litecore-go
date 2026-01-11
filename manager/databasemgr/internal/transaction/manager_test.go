package transaction

import (
	"database/sql"
	"testing"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// setupTestDB 创建测试数据库
func setupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to create test database: %v", err)
	}
	return db
}

// TestNewManager 测试创建事务管理器
func TestNewManager(t *testing.T) {
	db := setupTestDB(t)
	m := NewManager(db)

	if m == nil {
		t.Fatal("NewManager() returned nil")
	}

	if m.db != db {
		t.Error("Manager db is not the same as the provided db")
	}
}

// TestManager_Transaction 测试事务执行
func TestManager_Transaction(t *testing.T) {
	db := setupTestDB(t)
	m := NewManager(db)

	type TestModel struct {
		ID   uint
		Name string
	}

	// 创建表
	if err := db.AutoMigrate(&TestModel{}); err != nil {
		t.Fatalf("AutoMigrate() error = %v", err)
	}

	// 测试成功的事务
	err := m.Transaction(func(tx *gorm.DB) error {
		return tx.Create(&TestModel{Name: "test"}).Error
	})
	if err != nil {
		t.Errorf("Transaction() error = %v", err)
	}

	// 验证记录已创建
	var count int64
	db.Model(&TestModel{}).Count(&count)
	if count != 1 {
		t.Errorf("Expected 1 record, got %d", count)
	}
}

// TestManager_Transaction_Rollback 测试事务回滚
func TestManager_Transaction_Rollback(t *testing.T) {
	db := setupTestDB(t)
	m := NewManager(db)

	type TestModel struct {
		ID   uint
		Name string
	}

	// 创建表
	if err := db.AutoMigrate(&TestModel{}); err != nil {
		t.Fatalf("AutoMigrate() error = %v", err)
	}

	// 测试失败的事务（回滚）
	err := m.Transaction(func(tx *gorm.DB) error {
		tx.Create(&TestModel{Name: "should-rollback"})
		return gorm.ErrInvalidTransaction
	})
	if err == nil {
		t.Error("Transaction() should return error when transaction fails")
	}

	// 验证记录已回滚
	var count int64
	db.Model(&TestModel{}).Count(&count)
	if count != 0 {
		t.Errorf("Expected 0 records after rollback, got %d", count)
	}
}

// TestManager_Begin 测试手动开启事务
func TestManager_Begin(t *testing.T) {
	db := setupTestDB(t)
	m := NewManager(db)

	type TestModel struct {
		ID   uint
		Name string
	}

	// 创建表
	if err := db.AutoMigrate(&TestModel{}); err != nil {
		t.Fatalf("AutoMigrate() error = %v", err)
	}

	// 开启事务
	tx := m.Begin()
	if tx == nil {
		t.Fatal("Begin() returned nil")
	}

	// 在事务中插入数据
	if err := tx.Create(&TestModel{Name: "test"}).Error; err != nil {
		t.Errorf("Create() in transaction error = %v", err)
	}

	// 提交事务
	if err := tx.Commit().Error; err != nil {
		t.Errorf("Commit() error = %v", err)
	}

	// 验证记录已提交
	var count int64
	db.Model(&TestModel{}).Count(&count)
	if count != 1 {
		t.Errorf("Expected 1 record, got %d", count)
	}
}

// TestManager_Begin_Rollback 测试手动回滚事务
func TestManager_Begin_Rollback(t *testing.T) {
	db := setupTestDB(t)
	m := NewManager(db)

	type TestModel struct {
		ID   uint
		Name string
	}

	// 创建表
	if err := db.AutoMigrate(&TestModel{}); err != nil {
		t.Fatalf("AutoMigrate() error = %v", err)
	}

	// 开启事务
	tx := m.Begin()
	if tx == nil {
		t.Fatal("Begin() returned nil")
	}

	// 在事务中插入数据
	if err := tx.Create(&TestModel{Name: "should-rollback"}).Error; err != nil {
		t.Errorf("Create() in transaction error = %v", err)
	}

	// 回滚事务
	if err := tx.Rollback().Error; err != nil {
		t.Errorf("Rollback() error = %v", err)
	}

	// 验证记录已回滚
	var count int64
	db.Model(&TestModel{}).Count(&count)
	if count != 0 {
		t.Errorf("Expected 0 records after rollback, got %d", count)
	}
}

// TestManager_BeginTx 测试带选项开启事务
func TestManager_BeginTx(t *testing.T) {
	db := setupTestDB(t)
	m := NewManager(db)

	// 测试不同的隔离级别
	options := []*sql.TxOptions{
		nil,
		{Isolation: sql.LevelReadCommitted},
		{Isolation: sql.LevelSerializable},
		{Isolation: sql.LevelDefault, ReadOnly: true},
	}

	for i, opts := range options {
		t.Run(string(rune('a'+i)), func(t *testing.T) {
			tx := m.BeginTx(opts)
			if tx == nil {
				t.Error("BeginTx() returned nil")
			}
			if tx != nil {
				_ = tx.Rollback()
			}
		})
	}
}

// TestManager_NestedTransaction 测试嵌套事务
func TestManager_NestedTransaction(t *testing.T) {
	db := setupTestDB(t)
	m := NewManager(db)

	type TestModel struct {
		ID   uint
		Name string
	}

	// 创建表
	if err := db.AutoMigrate(&TestModel{}); err != nil {
		t.Fatalf("AutoMigrate() error = %v", err)
	}

	// 测试嵌套事务
	err := m.NestedTransaction(func(tx *gorm.DB) error {
		// 外层事务
		if err := tx.Create(&TestModel{Name: "outer"}).Error; err != nil {
			return err
		}

		// 内层事务（保存点）
		return tx.Transaction(func(tx2 *gorm.DB) error {
			if err := tx2.Create(&TestModel{Name: "inner"}).Error; err != nil {
				return err
			}
			return nil
		})
	})

	if err != nil {
		t.Errorf("NestedTransaction() error = %v", err)
	}

	// 验证两条记录都已创建
	var count int64
	db.Model(&TestModel{}).Count(&count)
	if count != 2 {
		t.Errorf("Expected 2 records, got %d", count)
	}
}

// TestManager_NestedTransaction_Rollback 测试嵌套事务回滚
func TestManager_NestedTransaction_Rollback(t *testing.T) {
	db := setupTestDB(t)
	m := NewManager(db)

	type TestModel struct {
		ID   uint
		Name string
	}

	// 创建表
	if err := db.AutoMigrate(&TestModel{}); err != nil {
		t.Fatalf("AutoMigrate() error = %v", err)
	}

	// 测试内层事务回滚
	err := m.NestedTransaction(func(tx *gorm.DB) error {
		// 外层事务
		if err := tx.Create(&TestModel{Name: "outer"}).Error; err != nil {
			return err
		}

		// 内层事务（失败）
		return tx.Transaction(func(tx2 *gorm.DB) error {
			tx2.Create(&TestModel{Name: "inner-rollback"})
			return gorm.ErrInvalidTransaction
		})
	})

	if err == nil {
		t.Error("NestedTransaction() should return error when inner transaction fails")
	}

	// 验证外层事务也回滚了
	var count int64
	db.Model(&TestModel{}).Count(&count)
	if count != 0 {
		t.Errorf("Expected 0 records after nested transaction rollback, got %d", count)
	}
}

// TestManager_Transaction_MultipleOperations 测试事务中的多个操作
func TestManager_Transaction_MultipleOperations(t *testing.T) {
	db := setupTestDB(t)
	m := NewManager(db)

	type TestModel struct {
		ID   uint
		Name string
	}

	// 创建表
	if err := db.AutoMigrate(&TestModel{}); err != nil {
		t.Fatalf("AutoMigrate() error = %v", err)
	}

	// 在事务中执行多个操作
	err := m.Transaction(func(tx *gorm.DB) error {
		// 创建
		if err := tx.Create(&TestModel{Name: "test1"}).Error; err != nil {
			return err
		}

		// 更新
		if err := tx.Model(&TestModel{}).Where("name = ?", "test1").Update("name", "updated").Error; err != nil {
			return err
		}

		// 查询
		var model TestModel
		if err := tx.Where("name = ?", "updated").First(&model).Error; err != nil {
			return err
		}

		// 删除
		if err := tx.Delete(&model).Error; err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		t.Errorf("Transaction() with multiple operations error = %v", err)
	}

	// 验证所有操作都已提交
	var count int64
	db.Model(&TestModel{}).Count(&count)
	if count != 0 {
		t.Errorf("Expected 0 records after delete, got %d", count)
	}
}

// TestManager_Transaction_Panic 测试事务中的 panic
func TestManager_Transaction_Panic(t *testing.T) {
	db := setupTestDB(t)
	m := NewManager(db)

	type TestModel struct {
		ID   uint
		Name string
	}

	// 创建表
	if err := db.AutoMigrate(&TestModel{}); err != nil {
		t.Fatalf("AutoMigrate() error = %v", err)
	}

	// 测试 panic 会导致事务回滚
	defer func() {
		if r := recover(); r != nil {
			// panic 被捕获，验证事务已回滚
			var count int64
			db.Model(&TestModel{}).Count(&count)
			if count != 0 {
				t.Errorf("Expected 0 records after panic, got %d", count)
			}
		}
	}()

	_ = m.Transaction(func(tx *gorm.DB) error {
		tx.Create(&TestModel{Name: "panic-test"})
		panic("intentional panic")
	})
}

// BenchmarkManager_Transaction 基准测试事务
func BenchmarkManager_Transaction(b *testing.B) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		b.Fatalf("failed to create test database: %v", err)
	}
	m := NewManager(db)

	type TestModel struct {
		ID   uint
		Name string
	}

	_ = db.AutoMigrate(&TestModel{})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = m.Transaction(func(tx *gorm.DB) error {
			return tx.Create(&TestModel{Name: "benchmark"}).Error
		})
	}
}

// BenchmarkManager_Begin 基准测试手动事务
func BenchmarkManager_Begin(b *testing.B) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		b.Fatalf("failed to create test database: %v", err)
	}
	m := NewManager(db)

	type TestModel struct {
		ID   uint
		Name string
	}

	_ = db.AutoMigrate(&TestModel{})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tx := m.Begin()
		_ = tx.Create(&TestModel{Name: "benchmark"}).Error
		_ = tx.Commit().Error
	}
}
