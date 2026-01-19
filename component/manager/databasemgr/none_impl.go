package databasemgr

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"gorm.io/gorm/schema"
)

// databaseManagerNoneImpl 空数据库管理器实现
// 在不需要数据库功能或数据库初始化失败时使用，提供空实现以避免条件判断
type databaseManagerNoneImpl struct {
	*databaseManagerBaseImpl
}

// NewDatabaseManagerNoneImpl 创建空数据库管理器
func NewDatabaseManagerNoneImpl() IDatabaseManager {
	// 创建一个空的 GORM DB 实例（使用 Dialector 但不打开连接）
	db, _ := gorm.Open(&dummyDialector{}, &gorm.Config{})

	baseImpl := newIDatabaseManagerBaseImpl("databaseManagerNoneImpl", "none", db)
	// 不初始化可观测性（none 驱动不需要）

	return &databaseManagerNoneImpl{
		databaseManagerBaseImpl: baseImpl,
	}
}

// ManagerName 返回管理器名称
func (n *databaseManagerNoneImpl) ManagerName() string {
	return n.name
}

// Health 返回错误（覆盖基类方法）
func (n *databaseManagerNoneImpl) Health() error {
	return errors.New("database not available (none driver)")
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
