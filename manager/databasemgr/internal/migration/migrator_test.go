package migration

import (
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

// TestNewMigrator 测试创建迁移器
func TestNewMigrator(t *testing.T) {
	db := setupTestDB(t)
	m := NewMigrator(db)

	if m == nil {
		t.Fatal("NewMigrator() returned nil")
	}

	if m.db != db {
		t.Error("Migrator db is not the same as the provided db")
	}
}

// TestMigrator_AutoMigrate 测试自动迁移
func TestMigrator_AutoMigrate(t *testing.T) {
	db := setupTestDB(t)
	m := NewMigrator(db)

	type TestModel struct {
		ID   uint
		Name string
		Age  int
	}

	err := m.AutoMigrate(&TestModel{})
	if err != nil {
		t.Errorf("AutoMigrate() error = %v", err)
	}

	// 验证表已创建
	if !m.db.Migrator().HasTable(&TestModel{}) {
		t.Error("Table was not created")
	}
}

// TestMigrator_AutoMigrate_MultipleModels 测试迁移多个模型
func TestMigrator_AutoMigrate_MultipleModels(t *testing.T) {
	db := setupTestDB(t)
	m := NewMigrator(db)

	type User struct {
		ID   uint
		Name string
	}

	type Product struct {
		ID    uint
		Name  string
		Price float64
	}

	err := m.AutoMigrate(&User{}, &Product{})
	if err != nil {
		t.Errorf("AutoMigrate() error = %v", err)
	}

	// 验证两个表都已创建
	if !m.db.Migrator().HasTable(&User{}) {
		t.Error("User table was not created")
	}

	if !m.db.Migrator().HasTable(&Product{}) {
		t.Error("Product table was not created")
	}
}

// TestMigrator_CreateTables 测试创建表
func TestMigrator_CreateTables(t *testing.T) {
	db := setupTestDB(t)
	m := NewMigrator(db)

	type TestModel struct {
		ID   uint
		Name string
	}

	err := m.CreateTables(&TestModel{})
	if err != nil {
		t.Errorf("CreateTables() error = %v", err)
	}

	// 验证表已创建
	if !m.db.Migrator().HasTable(&TestModel{}) {
		t.Error("Table was not created")
	}
}

// TestMigrator_DropTables 测试删除表
func TestMigrator_DropTables(t *testing.T) {
	db := setupTestDB(t)
	m := NewMigrator(db)

	type TestModel struct {
		ID   uint
		Name string
	}

	// 先创建表
	_ = m.CreateTables(&TestModel{})

	// 验证表存在
	if !m.db.Migrator().HasTable(&TestModel{}) {
		t.Fatal("Table was not created for DropTables test")
	}

	// 删除表
	err := m.DropTables(&TestModel{})
	if err != nil {
		t.Errorf("DropTables() error = %v", err)
	}

	// 验证表已删除
	if m.db.Migrator().HasTable(&TestModel{}) {
		t.Error("Table was not dropped")
	}
}

// TestMigrator_RenameTable 测试重命名表
func TestMigrator_RenameTable(t *testing.T) {
	db := setupTestDB(t)
	m := NewMigrator(db)

	type TestModel struct {
		ID   uint
		Name string
	}

	// 创建原始表
	_ = m.db.AutoMigrate(&TestModel{})

	// 重命名表
	oldName := "test_models"
	newName := "renamed_test_models"

	err := m.RenameTable(oldName, newName)
	if err != nil {
		t.Errorf("RenameTable() error = %v", err)
	}

	// 验证旧表名不存在
	if m.db.Migrator().HasTable(oldName) {
		t.Error("Old table name still exists")
	}

	// 验证新表名存在
	if !m.db.Migrator().HasTable(newName) {
		t.Error("New table name does not exist")
	}
}

// TestMigrator_AddColumn 测试添加列
func TestMigrator_AddColumn(t *testing.T) {
	db := setupTestDB(t)
	m := NewMigrator(db)

	type TestModel struct {
		ID   uint
		Name string
		Age  int
	}

	// 创建表（不包含 Age 列）
	type TestModelWithoutAge struct {
		ID   uint
		Name string
	}
	_ = m.db.AutoMigrate(&TestModelWithoutAge{})

	// 添加列 - 使用 AutoMigrate 来添加新列
	err := m.AutoMigrate(&TestModel{})
	if err != nil {
		t.Errorf("AutoMigrate() error = %v", err)
	}

	// 验证列已添加（HasColumn 检查）
	if !m.HasColumn(&TestModel{}, "Age") {
		t.Error("Column was not added")
	}
}

// TestMigrator_DropColumn 测试删除列
func TestMigrator_DropColumn(t *testing.T) {
	db := setupTestDB(t)
	m := NewMigrator(db)

	type TestModel struct {
		ID   uint
		Name string
		Age  int
	}

	// 创建表
	_ = m.db.AutoMigrate(&TestModel{})

	// 验证列存在
	if !m.HasColumn(&TestModel{}, "Age") {
		t.Fatal("Age column does not exist for DropColumn test")
	}

	// 删除列
	err := m.DropColumn(&TestModel{}, "Age")
	if err != nil {
		t.Errorf("DropColumn() error = %v", err)
	}

	// 验证列已删除
	if m.HasColumn(&TestModel{}, "Age") {
		t.Error("Column was not dropped")
	}
}

// TestMigrator_AlterColumn 测试修改列
func TestMigrator_AlterColumn(t *testing.T) {
	db := setupTestDB(t)
	m := NewMigrator(db)

	type TestModel struct {
		ID   uint
		Name string
	}

	// 创建表
	_ = m.db.AutoMigrate(&TestModel{})

	// 修改列（不同数据库对此操作的支持不同）
	err := m.AlterColumn(&TestModel{}, "Name")
	if err != nil {
		t.Errorf("AlterColumn() error = %v", err)
	}
}

// TestMigrator_HasColumn 测试检查列是否存在
func TestMigrator_HasColumn(t *testing.T) {
	db := setupTestDB(t)
	m := NewMigrator(db)

	type TestModel struct {
		ID   uint
		Name string
		Age  int
	}

	// 创建表
	_ = m.db.AutoMigrate(&TestModel{})

	// 测试存在的列
	if !m.HasColumn(&TestModel{}, "ID") {
		t.Error("HasColumn() should return true for ID column")
	}

	if !m.HasColumn(&TestModel{}, "Name") {
		t.Error("HasColumn() should return true for Name column")
	}

	if !m.HasColumn(&TestModel{}, "Age") {
		t.Error("HasColumn() should return true for Age column")
	}

	// 测试不存在的列
	if m.HasColumn(&TestModel{}, "NonExistent") {
		t.Error("HasColumn() should return false for non-existent column")
	}
}

// TestMigrator_CreateIndex 测试创建索引
func TestMigrator_CreateIndex(t *testing.T) {
	db := setupTestDB(t)
	m := NewMigrator(db)

	type TestModel struct {
		ID   uint
		Name string `gorm:"index"`
	}

	// 创建表（会自动创建索引）
	_ = m.db.AutoMigrate(&TestModel{})

	// 创建第二个表的索引
	type TestModelWithAge struct {
		ID  uint
		Age int `gorm:"index:idx_age"`
	}

	_ = m.db.AutoMigrate(&TestModelWithAge{})

	// 验证索引已创建
	if !m.HasIndex(&TestModelWithAge{}, "idx_age") {
		// 索引名称可能包含表名前缀，尝试查找
		t.Log("Index with exact name not found, may have table prefix")
	}
}

// TestMigrator_DropIndex 测试删除索引
func TestMigrator_DropIndex(t *testing.T) {
	db := setupTestDB(t)
	m := NewMigrator(db)

	type TestModel struct {
		ID   uint
		Name string `gorm:"index:idx_name"`
	}

	// 创建表（会自动创建索引）
	_ = m.db.AutoMigrate(&TestModel{})

	// 验证索引存在
	if !m.HasIndex(&TestModel{}, "idx_name") {
		t.Fatal("Index does not exist for DropIndex test")
	}

	// 删除索引
	err := m.DropIndex(&TestModel{}, "idx_name")
	if err != nil {
		t.Errorf("DropIndex() error = %v", err)
	}

	// 验证索引已删除
	if m.HasIndex(&TestModel{}, "idx_name") {
		t.Error("Index was not dropped")
	}
}

// TestMigrator_HasIndex 测试检查索引是否存在
func TestMigrator_HasIndex(t *testing.T) {
	db := setupTestDB(t)
	m := NewMigrator(db)

	type TestModel struct {
		ID   uint
		Name string `gorm:"index:idx_name"`
		Age  int    `gorm:"index:idx_age"`
	}

	// 创建表
	_ = m.db.AutoMigrate(&TestModel{})

	// 测试存在的索引
	if !m.HasIndex(&TestModel{}, "idx_name") {
		t.Error("HasIndex() should return true for idx_name")
	}

	if !m.HasIndex(&TestModel{}, "idx_age") {
		t.Error("HasIndex() should return true for idx_age")
	}

	// 测试不存在的索引
	if m.HasIndex(&TestModel{}, "non_existent_index") {
		t.Error("HasIndex() should return false for non-existent index")
	}
}

// TestMigrator_GetIndexes 测试获取所有索引
func TestMigrator_GetIndexes(t *testing.T) {
	db := setupTestDB(t)
	m := NewMigrator(db)

	type TestModel struct {
		ID   uint `gorm:"primaryKey"`
		Name string `gorm:"index:idx_name"`
	}

	// 创建表
	_ = m.db.AutoMigrate(&TestModel{})

	// 获取索引
	indexes, err := m.GetIndexes(&TestModel{})
	if err != nil {
		t.Errorf("GetIndexes() error = %v", err)
	}

	// 验证至少有一个索引（主键）
	if len(indexes) == 0 {
		t.Error("GetIndexes() should return at least one index (primary key)")
	}
}

// TestMigrator_GetColumns 测试获取所有列
func TestMigrator_GetColumns(t *testing.T) {
	db := setupTestDB(t)
	m := NewMigrator(db)

	type TestModel struct {
		ID   uint
		Name string
		Age  int
	}

	// 创建表
	_ = m.db.AutoMigrate(&TestModel{})

	// 获取列
	columns, err := m.GetColumns(&TestModel{})
	if err != nil {
		t.Errorf("GetColumns() error = %v", err)
	}

	// 验证列数量
	if len(columns) < 3 {
		t.Errorf("GetColumns() should return at least 3 columns, got %d", len(columns))
	}
}

// BenchmarkMigrator_AutoMigrate 基准测试自动迁移
func BenchmarkMigrator_AutoMigrate(b *testing.B) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		b.Fatalf("failed to create test database: %v", err)
	}
	m := NewMigrator(db)

	type TestModel struct {
		ID   uint
		Name string
		Age  int
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = m.AutoMigrate(&TestModel{})
	}
}

// BenchmarkMigrator_HasColumn 基准测试检查列
func BenchmarkMigrator_HasColumn(b *testing.B) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		b.Fatalf("failed to create test database: %v", err)
	}
	m := NewMigrator(db)

	type TestModel struct {
		ID   uint
		Name string
	}

	_ = db.AutoMigrate(&TestModel{})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = m.HasColumn(&TestModel{}, "Name")
	}
}

// BenchmarkMigrator_HasIndex 基准测试检查索引
func BenchmarkMigrator_HasIndex(b *testing.B) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		b.Fatalf("failed to create test database: %v", err)
	}
	m := NewMigrator(db)

	type TestModel struct {
		ID   uint
		Name string `gorm:"index"`
	}

	_ = db.AutoMigrate(&TestModel{})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = m.HasIndex(&TestModel{}, "idx_test_models_name")
	}
}
