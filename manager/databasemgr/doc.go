package databasemgr

// Package databasemgr 提供数据库管理功能，支持 MySQL、PostgreSQL 和 SQLite，基于 GORM 实现。
//
// 核心特性：
//   - 多数据库支持：MySQL、PostgreSQL、SQLite
//   - 工厂模式创建：通过 Factory 统一创建不同驱动的管理器
//   - 连接池管理：支持连接池配置和状态监控
//   - 事务支持：完整的事务管理和自动迁移功能
//   - 生命周期管理：集成服务启停接口，支持健康检查
//
// 基本用法：
//
//	factory := NewFactory()
//	cfg := &config.DatabaseConfig{
//	    Driver: "sqlite",
//	    SQLiteConfig: &config.SQLiteConfig{
//	        DSN: ":memory:",
//	    },
//	}
//	mgr, err := factory.BuildWithConfig(cfg)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	defer mgr.Close()
//
//	// 使用 GORM 操作数据库
//	db := mgr.DB()
//	db.Create(&User{Name: "test"})
//
// 配置选项：
//
//	连接池配置（PoolConfig）：
//	  - MaxOpenConns：最大打开连接数
//	  - MaxIdleConns：最大空闲连接数
//	  - ConnMaxLifetime：连接最大存活时间
//	  - ConnMaxIdleTime：连接最大空闲时间
//
//	各驱动的 DSN 格式：
//	  - SQLite：file:./data.db?cache=shared&mode=rwc
//	  - MySQL：user:password@tcp(localhost:3306)/dbname?charset=utf8mb4&parseTime=True&loc=Local
//	  - PostgreSQL：host=localhost port=5432 user=postgres password=password dbname=dbname sslmode=disable
//
// 错误处理：
//
//	BuildWithConfig 方法会验证配置并返回详细的错误信息。
//	Build 方法在配置错误时会返回 NoneDatabaseManager（空管理器），不会返回错误。
//	建议使用 BuildWithConfig 以获得更好的错误处理能力。
//
// 线程安全：
//
//	DatabaseManager 的所有方法都是线程安全的，可以在多个 goroutine 中并发使用。
