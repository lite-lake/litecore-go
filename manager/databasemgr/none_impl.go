package databasemgr

import (
	"context"
	"database/sql"
	"fmt"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"gorm.io/gorm/schema"

	"github.com/lite-lake/litecore-go/manager/loggermgr"
	"github.com/lite-lake/litecore-go/manager/telemetrymgr"
)

// databaseManagerNoneImpl 空数据库管理器实现
//
// 仅在配置显式声明 database.driver: "none" 时由 Build 构造，表示应用不依赖数据库，
// 提供空实现以避免业务代码到处判空。它是显式配置的结果，不是初始化失败后的退化路径：
// mysql/postgresql/sqlite 初始化失败会直接返回错误使进程启动失败，不会回退到本实现。
// 若未来引入「数据库初始化失败后回退 none」的逻辑，必须重新评估 Health() 与就绪语义。
type databaseManagerNoneImpl struct {
	*databaseManagerBaseImpl
}

// NewDatabaseManagerNoneImpl 创建空数据库管理器
//
// 该实现代表「显式声明无数据库依赖」，因此 Health() 返回健康（nil），
// 就绪探针据此判定应用声明的依赖集合为空、可以接收流量。
func NewDatabaseManagerNoneImpl(
	loggerMgr loggermgr.ILoggerManager,
	telemetryMgr telemetrymgr.ITelemetryManager,
) IDatabaseManager {
	// 创建一个空的 GORM DB 实例（使用 Dialector 但不打开连接）
	db, _ := gorm.Open(&dummyDialector{}, &gorm.Config{})

	baseImpl := newIDatabaseManagerBaseImpl(loggerMgr, telemetryMgr, "databaseManagerNoneImpl", "none", db)
	// 不初始化可观测性（none 驱动不需要）

	return &databaseManagerNoneImpl{
		databaseManagerBaseImpl: baseImpl,
	}
}

// ManagerName 返回管理器名称
func (n *databaseManagerNoneImpl) ManagerName() string {
	return n.name
}

// Health 就绪健康检查（覆盖基类方法）。
//
// none 表示应用在配置中显式声明「无数据库依赖」，因此就绪检查应视为健康。
// 这不会掩盖真实数据库故障：factory 只在 database.driver 显式为 "none" 时构造本实现；
// mysql/postgresql/sqlite 初始化失败会直接返回错误使进程启动失败，不会退化到 none。
// 注意：Ping / Transaction / AutoMigrate / DB 等方法的「数据库不可用」语义保持不变。
func (n *databaseManagerNoneImpl) Health() error {
	return nil
}

// OnStart 启动时初始化
func (n *databaseManagerNoneImpl) OnStart() error {
	// none 驱动不需要初始化
	return nil
}

// OnStop 停止时清理
func (n *databaseManagerNoneImpl) OnStop() error {
	// none 驱动不需要清理
	return nil
}

// ========== GORM 核心 ==========

// DB 返回 nil（覆盖基类方法）
func (n *databaseManagerNoneImpl) DB() *gorm.DB {
	return nil
}

// Model 返回空 DB
func (n *databaseManagerNoneImpl) Model(value any) *gorm.DB {
	return nil
}

// Table 返回空 DB
func (n *databaseManagerNoneImpl) Table(name string) *gorm.DB {
	return nil
}

// WithContext 返回空 DB
func (n *databaseManagerNoneImpl) WithContext(ctx context.Context) *gorm.DB {
	return nil
}

// ========== 事务管理 ==========

// Transaction 返回错误
func (n *databaseManagerNoneImpl) Transaction(fn func(*gorm.DB) error, opts ...*sql.TxOptions) error {
	return fmt.Errorf("database not available (none driver)")
}

// Begin 返回空 DB
func (n *databaseManagerNoneImpl) Begin(opts ...*sql.TxOptions) *gorm.DB {
	return nil
}

// ========== 迁移管理 ==========

// AutoMigrate 返回错误
func (n *databaseManagerNoneImpl) AutoMigrate(models ...any) error {
	return fmt.Errorf("database not available (none driver)")
}

// Migrator 返回 nil
func (n *databaseManagerNoneImpl) Migrator() gorm.Migrator {
	return nil
}

// ========== 连接管理 ==========

// Driver 获取驱动类型
func (n *databaseManagerNoneImpl) Driver() string {
	return n.driver
}

// Ping 返回错误（覆盖基类方法）
func (n *databaseManagerNoneImpl) Ping(ctx context.Context) error {
	return fmt.Errorf("database not available (none driver)")
}

// Stats 返回空统计
func (n *databaseManagerNoneImpl) Stats() sql.DBStats {
	return sql.DBStats{}
}

// Close 返回 nil
func (n *databaseManagerNoneImpl) Close() error {
	return nil
}

// ========== 原生 SQL ==========

// Exec 返回空 DB
func (n *databaseManagerNoneImpl) Exec(sql string, values ...any) *gorm.DB {
	return nil
}

// Raw 返回空 DB
func (n *databaseManagerNoneImpl) Raw(sql string, values ...any) *gorm.DB {
	return nil
}

// ========== dummyDialector 虚拟的 GORM Dialector ==========

type dummyDialector struct{}

func (d *dummyDialector) Name() string {
	return "none"
}

func (d *dummyDialector) Initialize(db *gorm.DB) error {
	return nil
}

func (d *dummyDialector) Migrator(db *gorm.DB) gorm.Migrator {
	return nil
}

func (d *dummyDialector) DataTypeOf(field *schema.Field) string {
	return ""
}

func (d *dummyDialector) DefaultValueOf(field *schema.Field) clause.Expression {
	return clause.Expr{}
}

func (d dummyDialector) BindVarTo(writer clause.Writer, stmt *gorm.Statement, v interface{}) {
	return
}

func (d dummyDialector) QuoteTo(writer clause.Writer, str string) {
	return
}

func (d dummyDialector) Explain(sql string, vars ...interface{}) string {
	return ""
}
