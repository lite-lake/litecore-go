package databasemgr

import (
	"context"
	"database/sql"

	"com.litelake.litecore/manager/databasemgr/internal/drivers"
)

// DatabaseManagerAdapter 数据库管理器适配器
// 将内部驱动适配到 DatabaseManager 接口
type DatabaseManagerAdapter struct {
	driver interface {
		ManagerName() string
		Health() error
		OnStart() error
		OnStop() error
		DB() *sql.DB
		Driver() string
		Ping(ctx context.Context) error
		BeginTx(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error)
		Stats() sql.DBStats
		Close() error
	}
}

// NewDatabaseManagerAdapter 创建数据库管理器适配器
func NewDatabaseManagerAdapter(driver interface {
	ManagerName() string
	Health() error
	OnStart() error
	OnStop() error
	DB() *sql.DB
	Driver() string
	Ping(ctx context.Context) error
	BeginTx(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error)
	Stats() sql.DBStats
	Close() error
}) *DatabaseManagerAdapter {
	return &DatabaseManagerAdapter{
		driver: driver,
	}
}

// DB 获取数据库连接
func (a *DatabaseManagerAdapter) DB() *sql.DB {
	return a.driver.DB()
}

// Driver 获取数据库驱动类型
func (a *DatabaseManagerAdapter) Driver() string {
	return a.driver.Driver()
}

// Ping 检查数据库连接
func (a *DatabaseManagerAdapter) Ping(ctx context.Context) error {
	return a.driver.Ping(ctx)
}

// BeginTx 开始事务
func (a *DatabaseManagerAdapter) BeginTx(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error) {
	return a.driver.BeginTx(ctx, opts)
}

// Stats 获取数据库连接池统计信息
func (a *DatabaseManagerAdapter) Stats() sql.DBStats {
	return a.driver.Stats()
}

// Close 关闭数据库连接
func (a *DatabaseManagerAdapter) Close() error {
	return a.driver.Close()
}

// ManagerName 返回管理器名称
func (a *DatabaseManagerAdapter) ManagerName() string {
	return a.driver.ManagerName()
}

// Health 检查管理器健康状态
func (a *DatabaseManagerAdapter) Health() error {
	return a.driver.Health()
}

// OnStart 在服务器启动时触发
func (a *DatabaseManagerAdapter) OnStart() error {
	return a.driver.OnStart()
}

// OnStop 在服务器停止时触发
func (a *DatabaseManagerAdapter) OnStop() error {
	return a.driver.OnStop()
}

// 编译时检查：确保 DatabaseManagerAdapter 实现了 DatabaseManager 接口
var _ DatabaseManager = (*DatabaseManagerAdapter)(nil)

// NewNoneDatabaseManagerAdapter 创建空数据库管理器适配器
func NewNoneDatabaseManagerAdapter() *DatabaseManagerAdapter {
	return NewDatabaseManagerAdapter(drivers.NewNoneDatabaseManager())
}

// NewSQLiteDatabaseManagerAdapter 创建 SQLite 数据库管理器适配器
func NewSQLiteDatabaseManagerAdapter(driver *drivers.SQLiteManager) *DatabaseManagerAdapter {
	return NewDatabaseManagerAdapter(driver)
}

// NewMySQLDatabaseManagerAdapter 创建 MySQL 数据库管理器适配器
func NewMySQLDatabaseManagerAdapter(driver *drivers.MySQLManager) *DatabaseManagerAdapter {
	return NewDatabaseManagerAdapter(driver)
}

// NewPostgreSQLDatabaseManagerAdapter 创建 PostgreSQL 数据库管理器适配器
func NewPostgreSQLDatabaseManagerAdapter(driver *drivers.PostgreSQLManager) *DatabaseManagerAdapter {
	return NewDatabaseManagerAdapter(driver)
}