package drivers

import (
	"context"
	"database/sql"
	"fmt"
	"sync"
	"time"

	"gorm.io/gorm"
)

// GormBaseManager GORM 基础管理器
type GormBaseManager struct {
	name   string
	driver string
	db     *gorm.DB
	sqlDB  *sql.DB // 用于连接池管理
	mu     sync.RWMutex
}

// NewGormBaseManager 创建 GORM 基础管理器
func NewGormBaseManager(name, driver string, db *gorm.DB) *GormBaseManager {
	sqlDB, _ := db.DB()
	return &GormBaseManager{
		name:   name,
		driver: driver,
		db:     db,
		sqlDB:  sqlDB,
	}
}

// ManagerName 返回管理器名称
func (m *GormBaseManager) ManagerName() string {
	return m.name
}

// DB 获取 GORM 数据库实例
func (m *GormBaseManager) DB() *gorm.DB {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.db
}

// Driver 获取驱动类型
func (m *GormBaseManager) Driver() string {
	return m.driver
}

// Ping 检查数据库连接
func (m *GormBaseManager) Ping(ctx context.Context) error {
	sqlDB, err := m.db.DB()
	if err != nil {
		return err
	}
	return sqlDB.PingContext(ctx)
}

// Health 检查健康状态
func (m *GormBaseManager) Health() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return m.Ping(ctx)
}

// Stats 获取连接池统计信息
func (m *GormBaseManager) Stats() sql.DBStats {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.sqlDB == nil {
		return sql.DBStats{}
	}

	return m.sqlDB.Stats()
}

// Close 关闭数据库连接
func (m *GormBaseManager) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.sqlDB == nil {
		return nil
	}

	err := m.sqlDB.Close()
	m.sqlDB = nil
	return err
}

// OnStart 启动时的初始化
func (m *GormBaseManager) OnStart() error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := m.Ping(ctx); err != nil {
		return fmt.Errorf("failed to ping database: %w", err)
	}

	return nil
}

// OnStop 停止时的清理
func (m *GormBaseManager) OnStop() error {
	return m.Close()
}

// ========== GORM 便捷方法 ==========

// Model 指定模型进行操作
func (m *GormBaseManager) Model(value any) *gorm.DB {
	return m.db.Model(value)
}

// Table 指定表名进行操作
func (m *GormBaseManager) Table(name string) *gorm.DB {
	return m.db.Table(name)
}

// WithContext 设置上下文
func (m *GormBaseManager) WithContext(ctx context.Context) *gorm.DB {
	return m.db.WithContext(ctx)
}

// Transaction 执行事务
func (m *GormBaseManager) Transaction(fn func(*gorm.DB) error, opts ...*sql.TxOptions) error {
	return m.db.Transaction(fn, opts...)
}

// Begin 开启事务
func (m *GormBaseManager) Begin(opts ...*sql.TxOptions) *gorm.DB {
	return m.db.Begin(opts...)
}

// AutoMigrate 自动迁移
func (m *GormBaseManager) AutoMigrate(models ...any) error {
	return m.db.AutoMigrate(models...)
}

// Migrator 获取迁移器
func (m *GormBaseManager) Migrator() gorm.Migrator {
	return m.db.Migrator()
}

// Exec 执行原生 SQL
func (m *GormBaseManager) Exec(sql string, values ...any) *gorm.DB {
	return m.db.Exec(sql, values...)
}

// Raw 执行原生查询
func (m *GormBaseManager) Raw(sql string, values ...any) *gorm.DB {
	return m.db.Raw(sql, values...)
}
