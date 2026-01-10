package drivers

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"com.litelake.litecore/common"
)

// NoneDatabaseManager 空数据库管理器
// 在不需要数据库功能或数据库初始化失败时使用，提供空实现以避免条件判断
type NoneDatabaseManager struct {
	*BaseManager
}

// NewNoneDatabaseManager 创建空数据库管理器
func NewNoneDatabaseManager() *NoneDatabaseManager {
	return &NoneDatabaseManager{
		BaseManager: NewBaseManager("none-database"),
	}
}

// DB 返回 nil
func (m *NoneDatabaseManager) DB() *sql.DB {
	return nil
}

// Driver 返回 "none"
func (m *NoneDatabaseManager) Driver() string {
	return "none"
}

// Ping 返回错误
func (m *NoneDatabaseManager) Ping(ctx context.Context) error {
	return fmt.Errorf("database not available (none driver)")
}

// BeginTx 返回错误
func (m *NoneDatabaseManager) BeginTx(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error) {
	return nil, fmt.Errorf("database not available (none driver)")
}

// Stats 返回空的统计信息
func (m *NoneDatabaseManager) Stats() sql.DBStats {
	return sql.DBStats{}
}

// Close 空实现，无需关闭连接
func (m *NoneDatabaseManager) Close() error {
	return nil
}

// Health 返回错误
func (m *NoneDatabaseManager) Health() error {
	return errors.New("database not available (none driver)")
}

// Shutdown 空实现，无需清理资源
func (m *NoneDatabaseManager) Shutdown(ctx context.Context) error {
	return nil
}

// 编译时检查：确保 NoneDatabaseManager 实现了 common.Manager 接口
var _ common.Manager = (*NoneDatabaseManager)(nil)