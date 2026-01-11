package drivers

import (
	"context"
	"errors"
	"fmt"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"gorm.io/gorm/schema"
)

// NoneDatabaseManager 空数据库管理器
// 在不需要数据库功能或数据库初始化失败时使用，提供空实现以避免条件判断
type NoneDatabaseManager struct {
	*GormBaseManager
}

// NewNoneDatabaseManager 创建空数据库管理器
func NewNoneDatabaseManager() *NoneDatabaseManager {
	// 创建一个空的 GORM DB 实例（使用Dialector但不打开连接）
	db, _ := gorm.Open(&DummyDialector{}, &gorm.Config{})

	return &NoneDatabaseManager{
		GormBaseManager: NewGormBaseManager("none-database", "none", db),
	}
}

// DB 返回 nil（覆盖基类方法）
func (m *NoneDatabaseManager) DB() *gorm.DB {
	return nil
}

// Ping 返回错误（覆盖基类方法）
func (m *NoneDatabaseManager) Ping(ctx context.Context) error {
	return fmt.Errorf("database not available (none driver)")
}

// Health 返回错误（覆盖基类方法）
func (m *NoneDatabaseManager) Health() error {
	return errors.New("database not available (none driver)")
}

// DummyDialector 虚拟的 GORM Dialector，用于 NoneDatabaseManager
type DummyDialector struct{}

func (d *DummyDialector) Name() string {
	return "none"
}

func (d *DummyDialector) Initialize(db *gorm.DB) error {
	return nil
}

func (d *DummyDialector) Migrator(db *gorm.DB) gorm.Migrator {
	return nil
}

func (d *DummyDialector) DataTypeOf(field *schema.Field) string {
	return ""
}

func (d *DummyDialector) DefaultValueOf(field *schema.Field) clause.Expression {
	return clause.Expr{}
}

func (d DummyDialector) BindVarTo(writer clause.Writer, stmt *gorm.Statement, v interface{}) {
	return
}

func (d DummyDialector) QuoteTo(writer clause.Writer, str string) {
	return
}

func (d DummyDialector) Explain(sql string, vars ...interface{}) string {
	return ""
}
