package migration

import (
	"gorm.io/gorm"
)

// Migrator 迁移管理器
type Migrator struct {
	db *gorm.DB
}

// NewMigrator 创建迁移管理器
func NewMigrator(db *gorm.DB) *Migrator {
	return &Migrator{db: db}
}

// AutoMigrate 自动迁移
func (m *Migrator) AutoMigrate(models ...interface{}) error {
	return m.db.AutoMigrate(models...)
}

// CreateTables 创建表
func (m *Migrator) CreateTables(models ...interface{}) error {
	return m.db.Migrator().CreateTable(models...)
}

// DropTables 删除表
func (m *Migrator) DropTables(models ...interface{}) error {
	return m.db.Migrator().DropTable(models...)
}

// RenameTable 重命名表
func (m *Migrator) RenameTable(oldName, newName string) error {
	return m.db.Migrator().RenameTable(oldName, newName)
}

// AddColumn 添加列
func (m *Migrator) AddColumn(model interface{}, field string) error {
	return m.db.Migrator().AddColumn(model, field)
}

// DropColumn 删除列
func (m *Migrator) DropColumn(model interface{}, field string) error {
	return m.db.Migrator().DropColumn(model, field)
}

// AlterColumn 修改列
func (m *Migrator) AlterColumn(model interface{}, field string) error {
	return m.db.Migrator().AlterColumn(model, field)
}

// HasColumn 检查列是否存在
func (m *Migrator) HasColumn(model interface{}, field string) bool {
	return m.db.Migrator().HasColumn(model, field)
}

// CreateIndex 创建索引
func (m *Migrator) CreateIndex(model interface{}, name string) error {
	return m.db.Migrator().CreateIndex(model, name)
}

// DropIndex 删除索引
func (m *Migrator) DropIndex(model interface{}, name string) error {
	return m.db.Migrator().DropIndex(model, name)
}

// HasIndex 检查索引是否存在
func (m *Migrator) HasIndex(model interface{}, name string) bool {
	return m.db.Migrator().HasIndex(model, name)
}

// GetIndexes 获取所有索引
func (m *Migrator) GetIndexes(model interface{}) ([]gorm.Index, error) {
	return m.db.Migrator().GetIndexes(model)
}

// GetColumns 获取所有列
func (m *Migrator) GetColumns(model interface{}) ([]gorm.ColumnType, error) {
	return m.db.Migrator().ColumnTypes(model)
}
