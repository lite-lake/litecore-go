package databasemgr

import (
	"context"
	"database/sql"
)

// DatabaseManager 数据库管理器接口
type DatabaseManager interface {
	// ManagerName 返回管理器名称
	ManagerName() string

	// Health 检查管理器健康状态
	Health() error

	// OnStart 在服务器启动时触发
	OnStart() error

	// OnStop 在服务器停止时触发
	OnStop() error

	// DB 获取数据库连接
	DB() *sql.DB

	// Driver 获取数据库驱动类型
	Driver() string

	// Ping 检查数据库连接
	Ping(ctx context.Context) error

	// BeginTx 开始事务
	BeginTx(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error)

	// Stats 获取数据库连接池统计信息
	Stats() sql.DBStats

	// Close 关闭数据库连接
	Close() error
}