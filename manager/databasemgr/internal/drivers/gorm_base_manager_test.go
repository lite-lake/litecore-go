package drivers

import (
	"context"
	"testing"
	"time"

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

// TestNewGormBaseManager 测试创建 GORM 基础管理器
func TestNewGormBaseManager(t *testing.T) {
	db := setupTestDB(t)
	m := NewGormBaseManager("test-manager", "sqlite", db)

	if m == nil {
		t.Fatal("NewGormBaseManager() returned nil")
	}

	if m.ManagerName() != "test-manager" {
		t.Errorf("ManagerName() = %v, want test-manager", m.ManagerName())
	}

	if m.Driver() != "sqlite" {
		t.Errorf("Driver() = %v, want sqlite", m.Driver())
	}

	if m.DB() == nil {
		t.Error("DB() returned nil")
	}
}

// TestGormBaseManager_DB 测试获取数据库实例
func TestGormBaseManager_DB(t *testing.T) {
	db := setupTestDB(t)
	m := NewGormBaseManager("test", "sqlite", db)

	got := m.DB()
	if got == nil {
		t.Error("DB() returned nil")
	}

	if got != db {
		t.Error("DB() returned different instance")
	}
}

// TestGormBaseManager_Driver 测试获取驱动类型
func TestGormBaseManager_Driver(t *testing.T) {
	tests := []struct {
		name   string
		driver string
	}{
		{"sqlite", "sqlite"},
		{"mysql", "mysql"},
		{"postgresql", "postgresql"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := setupTestDB(t)
			m := NewGormBaseManager("test", tt.driver, db)
			if got := m.Driver(); got != tt.driver {
				t.Errorf("Driver() = %v, want %v", got, tt.driver)
			}
		})
	}
}

// TestGormBaseManager_Ping 测试数据库连接
func TestGormBaseManager_Ping(t *testing.T) {
	db := setupTestDB(t)
	m := NewGormBaseManager("test", "sqlite", db)

	ctx := context.Background()
	if err := m.Ping(ctx); err != nil {
		t.Errorf("Ping() error = %v", err)
	}
}

// TestGormBaseManager_Ping_Timeout 测试超时上下文
func TestGormBaseManager_Ping_Timeout(t *testing.T) {
	db := setupTestDB(t)
	m := NewGormBaseManager("test", "sqlite", db)

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	if err := m.Ping(ctx); err != nil {
		t.Errorf("Ping() with timeout error = %v", err)
	}
}

// TestGormBaseManager_Ping_Cancelled 测试已取消的上下文
func TestGormBaseManager_Ping_Cancelled(t *testing.T) {
	db := setupTestDB(t)
	m := NewGormBaseManager("test", "sqlite", db)

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // 立即取消

	if err := m.Ping(ctx); err == nil {
		t.Error("Ping() with cancelled context should return error")
	}
}

// TestGormBaseManager_Health 测试健康检查
func TestGormBaseManager_Health(t *testing.T) {
	db := setupTestDB(t)
	m := NewGormBaseManager("test", "sqlite", db)

	if err := m.Health(); err != nil {
		t.Errorf("Health() error = %v", err)
	}
}

// TestGormBaseManager_Stats 测试获取连接池统计
func TestGormBaseManager_Stats(t *testing.T) {
	db := setupTestDB(t)
	m := NewGormBaseManager("test", "sqlite", db)

	stats := m.Stats()
	if stats.MaxOpenConnections == 0 {
		// SQLite 默认配置可能为 0，这是正常的
		t.Log("MaxOpenConnections is 0, this may be expected for SQLite")
	}
}

// TestGormBaseManager_Close 测试关闭数据库连接
func TestGormBaseManager_Close(t *testing.T) {
	db := setupTestDB(t)
	m := NewGormBaseManager("test", "sqlite", db)

	// 第一次关闭应该成功
	if err := m.Close(); err != nil {
		t.Errorf("Close() error = %v", err)
	}

	// 第二次关闭也应该成功（幂等）
	if err := m.Close(); err != nil {
		t.Errorf("Close() second call error = %v", err)
	}

	// 关闭后 DB() 应该返回 nil 或无效实例
	db2 := m.DB()
	if db2 != nil {
		// 检查是否真的关闭了
		sqlDB, err2 := db2.DB()
		if err2 == nil && sqlDB != nil {
			// 尝试 Ping 应该失败
			ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
			defer cancel()
			if sqlDB.PingContext(ctx) == nil {
				t.Error("Database should be closed but still responds")
			}
		}
	}
}

// TestGormBaseManager_OnStart 测试启动钩子
func TestGormBaseManager_OnStart(t *testing.T) {
	db := setupTestDB(t)
	m := NewGormBaseManager("test", "sqlite", db)

	if err := m.OnStart(); err != nil {
		t.Errorf("OnStart() error = %v", err)
	}
}

// TestGormBaseManager_OnStop 测试停止钩子
func TestGormBaseManager_OnStop(t *testing.T) {
	db := setupTestDB(t)
	m := NewGormBaseManager("test", "sqlite", db)

	if err := m.OnStop(); err != nil {
		t.Errorf("OnStop() error = %v", err)
	}
}

// TestGormBaseManager_Model 测试 Model 方法
func TestGormBaseManager_Model(t *testing.T) {
	type TestModel struct {
		ID   uint
		Name string
	}

	db := setupTestDB(t)
	m := NewGormBaseManager("test", "sqlite", db)

	result := m.Model(&TestModel{})
	if result == nil {
		t.Error("Model() returned nil")
	}
}

// TestGormBaseManager_Table 测试 Table 方法
func TestGormBaseManager_Table(t *testing.T) {
	db := setupTestDB(t)
	m := NewGormBaseManager("test", "sqlite", db)

	result := m.Table("test_table")
	if result == nil {
		t.Error("Table() returned nil")
	}
}

// TestGormBaseManager_WithContext 测试 WithContext 方法
func TestGormBaseManager_WithContext(t *testing.T) {
	db := setupTestDB(t)
	m := NewGormBaseManager("test", "sqlite", db)

	ctx := context.Background()
	result := m.WithContext(ctx)
	if result == nil {
		t.Error("WithContext() returned nil")
	}
}

// TestGormBaseManager_Transaction 测试事务
func TestGormBaseManager_Transaction(t *testing.T) {
	db := setupTestDB(t)
	m := NewGormBaseManager("test", "sqlite", db)

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

	// 测试失败的事务（回滚）
	err = m.Transaction(func(tx *gorm.DB) error {
		tx.Create(&TestModel{Name: "should-rollback"})
		return gorm.ErrInvalidTransaction
	})
	if err == nil {
		t.Error("Transaction() should return error when transaction fails")
	}

	// 验证回滚
	var count int64
	m.DB().Model(&TestModel{}).Count(&count)
	if count != 1 {
		t.Errorf("Expected 1 record, got %d", count)
	}
}

// TestGormBaseManager_Begin 测试手动开启事务
func TestGormBaseManager_Begin(t *testing.T) {
	db := setupTestDB(t)
	m := NewGormBaseManager("test", "sqlite", db)

	tx := m.Begin()
	if tx == nil {
		t.Error("Begin() returned nil")
	}

	// 回滚事务
	tx.Rollback()
}

// TestGormBaseManager_AutoMigrate 测试自动迁移
func TestGormBaseManager_AutoMigrate(t *testing.T) {
	db := setupTestDB(t)
	m := NewGormBaseManager("test", "sqlite", db)

	type TestModel struct {
		ID   uint
		Name string
	}

	if err := m.AutoMigrate(&TestModel{}); err != nil {
		t.Errorf("AutoMigrate() error = %v", err)
	}
}

// TestGormBaseManager_Migrator 测试获取迁移器
func TestGormBaseManager_Migrator(t *testing.T) {
	db := setupTestDB(t)
	m := NewGormBaseManager("test", "sqlite", db)

	migrator := m.Migrator()
	if migrator == nil {
		t.Error("Migrator() returned nil")
	}
}

// TestGormBaseManager_Exec 测试执行原生 SQL
func TestGormBaseManager_Exec(t *testing.T) {
	db := setupTestDB(t)
	m := NewGormBaseManager("test", "sqlite", db)

	result := m.Exec("CREATE TABLE IF NOT EXISTS test (id INTEGER)")
	if result.Error != nil {
		t.Errorf("Exec() error = %v", result.Error)
	}

	if result.RowsAffected == 0 {
		// 表已存在，RowsAffected 可能是 0，这是正常的
		t.Log("Table already exists, RowsAffected is 0")
	}
}

// TestGormBaseManager_Raw 测试执行原生查询
func TestGormBaseManager_Raw(t *testing.T) {
	db := setupTestDB(t)
	m := NewGormBaseManager("test", "sqlite", db)

	result := m.Raw("SELECT 1")
	if result.Error != nil {
		t.Errorf("Raw() error = %v", result.Error)
	}
}

// TestGormBaseManager_ConcurrentAccess 测试并发访问
func TestGormBaseManager_ConcurrentAccess(t *testing.T) {
	db := setupTestDB(t)
	m := NewGormBaseManager("test", "sqlite", db)

	done := make(chan bool)

	// 并发读取
	for i := 0; i < 10; i++ {
		go func() {
			_ = m.DB()
			_ = m.Driver()
			_ = m.Stats()
			done <- true
		}()
	}

	// 等待所有 goroutine 完成
	for i := 0; i < 10; i++ {
		<-done
	}
}

// BenchmarkGormBaseManager_DB 基准测试 DB 方法
func BenchmarkGormBaseManager_DB(b *testing.B) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		b.Fatalf("failed to create test database: %v", err)
	}
	m := NewGormBaseManager("test", "sqlite", db)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = m.DB()
	}
}

// BenchmarkGormBaseManager_Stats 基准测试 Stats 方法
func BenchmarkGormBaseManager_Stats(b *testing.B) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		b.Fatalf("failed to create test database: %v", err)
	}
	m := NewGormBaseManager("test", "sqlite", db)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = m.Stats()
	}
}

// BenchmarkGormBaseManager_Health 基准测试 Health 方法
func BenchmarkGormBaseManager_Health(b *testing.B) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		b.Fatalf("failed to create test database: %v", err)
	}
	m := NewGormBaseManager("test", "sqlite", db)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = m.Health()
	}
}
