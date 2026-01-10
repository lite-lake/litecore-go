package databasemgr

import (
	"context"
	"gorm.io/gorm"
)

// DatabaseManager 数据库管理器接口（完全基于 GORM）
type DatabaseManager interface {
	// ========== 生命周期管理 ==========
	// ManagerName 返回管理器名称
	ManagerName() string

	// Health 检查管理器健康状态
	Health() error

	// OnStart 在服务器启动时触发
	OnStart() error

	// OnStop 在服务器停止时触发
	OnStop() error

	// ========== GORM 核心 ==========
	// DB 获取 GORM 数据库实例
	DB() *gorm.DB

	// Model 指定模型进行操作
	Model(value any) *gorm.DB

	// Table 指定表名进行操作
	Table(name string) *gorm.DB

	// WithContext 设置上下文
	WithContext(ctx context.Context) *gorm.DB

	// ========== 事务管理 ==========
	// Transaction 执行事务
	Transaction(fn func(*gorm.DB) error, opts ...*interface{}) error

	// Begin 开启事务
	Begin(opts ...*interface{}) *gorm.DB

	// ========== 迁移管理 ==========
	// AutoMigrate 自动迁移
	AutoMigrate(models ...any) error

	// Migrator 获取迁移器
	Migrator() gorm.Migrator

	// ========== 连接管理 ==========
	// Driver 获取数据库驱动类型
	Driver() string

	// Ping 检查数据库连接
	Ping(ctx context.Context) error

	// Stats 获取连接池统计信息
	Stats() gorm.SQLStats

	// Close 关闭数据库连接
	Close() error

	// ========== 原生 SQL ==========
	// Exec 执行原生 SQL
	Exec(sql string, values ...any) *gorm.DB

	// Raw 执行原生查询
	Raw(sql string, values ...any) *gorm.DB
}
