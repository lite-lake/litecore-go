// Package databasemgr 提供数据库管理功能，支持 MySQL、PostgreSQL 和 SQLite 三种数据库驱动。
//
// 核心特性：
//   - 多驱动支持：支持 MySQL、PostgreSQL 和 SQLite 三种数据库
//   - 连接池管理：自动管理数据库连接池，支持连接池参数配置
//   - 健康检查：提供数据库连接健康检查功能
//   - 事务支持：支持事务操作，保证数据一致性
//   - 零成本降级：配置失败时自动降级到空数据库管理器，避免影响程序运行
//   - 线程安全：所有操作都是线程安全的，支持并发访问
//
// 基本用法：
//
//	// 使用工厂创建数据库管理器
//	factory := databasemgr.NewFactory()
//
//	// 配置 SQLite 数据库
//	cfg := map[string]any{
//	    "driver": "sqlite",
//	    "sqlite_config": map[string]any{
//	        "dsn": "file:./data.db?cache=shared&mode=rwc",
//	    },
//	}
//
//	// 构建管理器
//	mgr := factory.Build("sqlite", cfg)
//	dbMgr := mgr.(databasemgr.DatabaseManager)
//
//	// 使用数据库
//	db := dbMgr.DB()
//	rows, err := db.QueryContext(ctx, "SELECT * FROM users")
//	if err != nil {
//	    return err
//	}
//	defer rows.Close()
//
//	// 使用事务
//	tx, err := dbMgr.BeginTx(ctx, nil)
//	if err != nil {
//	    return err
//	}
//	defer tx.Rollback()
//	_, err = tx.Exec("INSERT INTO users (name) VALUES (?)", "John")
//	if err != nil {
//	    return err
//	}
//	err = tx.Commit()
//	if err != nil {
//	    return err
//	}
//
//	// 关闭连接
//	_ = dbMgr.Close()
//
// 数据库驱动：
//
//	包支持四种数据库驱动：
//	- mysql：使用 go-sql-driver/mysql 驱动，适合生产环境和高并发场景
//	- postgresql：使用 lib/pq 驱动，适合需要高级功能的企业级应用
//	- sqlite：使用 mattn/go-sqlite3 驱动，适合嵌入式应用和测试环境
//	- none：空驱动，用于降级场景
//
// 配置格式：
//
//	通用配置：
//	  driver: 数据库驱动类型（mysql/postgresql/sqlite/none）
//
//	MySQL 配置：
//	  mysql_config.dsn: 数据源字符串，如 "root:password@tcp(localhost:3306)/dbname?charset=utf8mb4"
//	  mysql_config.pool_config: 连接池配置（可选）
//
//	PostgreSQL 配置：
//	  postgresql_config.dsn: 数据源字符串，如 "host=localhost port=5432 user=postgres password=password dbname=dbname sslmode=disable"
//	  postgresql_config.pool_config: 连接池配置（可选）
//
//	SQLite 配置：
//	  sqlite_config.dsn: 数据源字符串，如 "file:./data.db?cache=shared&mode=rwc"
//	  sqlite_config.pool_config: 连接池配置（可选）
//
// 连接池配置选项：
//
//	max_open_conns: 最大打开连接数，默认 10（SQLite 建议设置为 1）
//	max_idle_conns: 最大空闲连接数，默认 5
//	conn_max_lifetime: 连接最大存活时间，默认 30s
//	conn_max_idle_time: 连接最大空闲时间，默认 5m
//
// 错误处理：
//
//	数据库管理器采用零成本降级策略：
//	- 配置解析失败时，自动返回 none 驱动的管理器
//	- 数据库连接创建失败时，自动降级到 none 管理器
//	- 可通过 Driver() 方法检查是否为 none 驱动
//	- 所有数据库操作返回原始错误，由调用方处理
//
//	示例：
//	  if dbMgr.Driver() == "none" {
//	      return errors.New("database not available")
//	  }
package databasemgr