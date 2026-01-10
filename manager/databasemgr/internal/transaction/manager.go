package transaction

import (
	"gorm.io/gorm"
)

// Manager 事务管理器
type Manager struct {
	db *gorm.DB
}

// NewManager 创建事务管理器
func NewManager(db *gorm.DB) *Manager {
	return &Manager{db: db}
}

// Transaction 执行事务
func (m *Manager) Transaction(fn func(*gorm.DB) error) error {
	return m.db.Transaction(fn)
}

// Begin 开启事务
func (m *Manager) Begin() *gorm.DB {
	return m.db.Begin()
}

// BeginTx 开启事务（带选项）
func (m *Manager) BeginTx(opts ...*interface{}) *gorm.DB {
	return m.db.Begin(opts...)
}

// NestedTransaction 嵌套事务
func (m *Manager) NestedTransaction(fn func(*gorm.DB) error) error {
	return m.db.Transaction(func(tx *gorm.DB) error {
		// 外层事务
		return tx.Transaction(func(tx2 *gorm.DB) error {
			// 内层事务（保存点）
			return fn(tx2)
		})
	})
}
