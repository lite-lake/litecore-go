package drivers

import (
	"context"
	"database/sql"
	"fmt"
	"sync"
	"time"

	_ "github.com/go-sql-driver/mysql"

	"com.litelake.litecore/common"
	"com.litelake.litecore/manager/databasemgr/internal/config"
)

// MySQLManager MySQL 数据库管理器
type MySQLManager struct {
	*BaseManager
	config        *config.DatabaseConfig
	db            *sql.DB
	driver        string
	mu            sync.RWMutex
	shutdownFuncs []func(context.Context) error
	shutdownOnce  sync.Once
}

// NewMySQLManager 创建 MySQL 数据库管理器
func NewMySQLManager(cfg *config.DatabaseConfig) (*MySQLManager, error) {
	// 验证配置
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid database config: %w", err)
	}

	if cfg.MySQLConfig == nil {
		return nil, fmt.Errorf("mysql_config is required")
	}

	if cfg.MySQLConfig.DSN == "" {
		return nil, fmt.Errorf("mysql DSN is required")
	}

	mgr := &MySQLManager{
		BaseManager:   NewBaseManager("mysql-database"),
		config:        cfg,
		driver:        "mysql",
		shutdownFuncs: make([]func(context.Context) error, 0),
	}

	// 打开数据库连接
	db, err := sql.Open("mysql", cfg.MySQLConfig.DSN)
	if err != nil {
		return nil, fmt.Errorf("failed to open mysql database: %w", err)
	}
	mgr.db = db

	// 配置连接池
	poolConfig := cfg.MySQLConfig.PoolConfig
	if poolConfig != nil {
		if err := mgr.configurePool(poolConfig); err != nil {
			db.Close()
			return nil, fmt.Errorf("failed to configure connection pool: %w", err)
		}
	}

	// 添加关闭函数
	mgr.shutdownFuncs = append(mgr.shutdownFuncs, func(ctx context.Context) error {
		return mgr.db.Close()
	})

	return mgr, nil
}

// configurePool 配置连接池
func (m *MySQLManager) configurePool(poolConfig *config.PoolConfig) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if poolConfig.MaxOpenConns > 0 {
		m.db.SetMaxOpenConns(poolConfig.MaxOpenConns)
	}

	if poolConfig.MaxIdleConns > 0 {
		m.db.SetMaxIdleConns(poolConfig.MaxIdleConns)
	}

	if poolConfig.ConnMaxLifetime > 0 {
		m.db.SetConnMaxLifetime(poolConfig.ConnMaxLifetime)
	}

	if poolConfig.ConnMaxIdleTime > 0 {
		m.db.SetConnMaxIdleTime(poolConfig.ConnMaxIdleTime)
	}

	return nil
}

// DB 获取数据库连接
func (m *MySQLManager) DB() *sql.DB {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.db
}

// Driver 获取数据库驱动类型
func (m *MySQLManager) Driver() string {
	return m.driver
}

// Ping 检查数据库连接
func (m *MySQLManager) Ping(ctx context.Context) error {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.db == nil {
		return fmt.Errorf("database connection is nil")
	}

	if err := ValidateContext(ctx); err != nil {
		return err
	}

	return m.db.PingContext(ctx)
}

// BeginTx 开始事务
func (m *MySQLManager) BeginTx(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.db == nil {
		return nil, fmt.Errorf("database connection is nil")
	}

	if err := ValidateContext(ctx); err != nil {
		return nil, err
	}

	return m.db.BeginTx(ctx, opts)
}

// Stats 获取数据库连接池统计信息
func (m *MySQLManager) Stats() sql.DBStats {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.db == nil {
		return sql.DBStats{}
	}

	return m.db.Stats()
}

// Close 关闭数据库连接
func (m *MySQLManager) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.db == nil {
		return nil
	}

	err := m.db.Close()
	m.db = nil
	return err
}

// Health 检查管理器健康状态
func (m *MySQLManager) Health() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	return m.Ping(ctx)
}

// OnStart 在服务器启动时触发
func (m *MySQLManager) OnStart() error {
	// 验证数据库连接
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := m.Ping(ctx); err != nil {
		return fmt.Errorf("failed to ping database: %w", err)
	}

	return nil
}

// OnStop 在服务器停止时触发
func (m *MySQLManager) OnStop() error {
	return m.Close()
}

// Shutdown 关闭管理器
func (m *MySQLManager) Shutdown(ctx context.Context) error {
	var shutdownErr error

	m.shutdownOnce.Do(func() {
		m.mu.Lock()
		defer m.mu.Unlock()

		// 按相反顺序执行 shutdown 函数
		for i := len(m.shutdownFuncs) - 1; i >= 0; i-- {
			if err := m.shutdownFuncs[i](ctx); err != nil {
				if shutdownErr == nil {
					shutdownErr = err
				} else {
					shutdownErr = fmt.Errorf("%v; %w", shutdownErr, err)
				}
			}
		}

		// 将 db 设置为 nil
		m.db = nil
		m.shutdownFuncs = make([]func(context.Context) error, 0)
	})

	return shutdownErr
}

// 编译时检查：确保 MySQLManager 实现了 common.Manager 接口
var _ common.Manager = (*MySQLManager)(nil)