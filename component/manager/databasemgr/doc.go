package databasemgr

// Package databasemgr 提供统一的数据库管理功能,基于 GORM 支持 MySQL、PostgreSQL 和 SQLite。
//
// 核心特性：
//   - 多数据库支持：MySQL、PostgreSQL、SQLite 和 None(空实现)
//   - 连接池管理：统一的连接池配置和统计
//   - 可观测性集成：内置日志、追踪和指标收集
//   - 事务管理：支持事务操作和自动回滚
//   - 自动迁移：基于 GORM 的数据库迁移能力
//   - 配置驱动：支持通过配置提供者创建实例
//
// 基本用法：
//
//	// 使用工厂函数创建
//	cfg := map[string]any{
//	    "dsn": "root:password@tcp(localhost:3306)/mydb?charset=utf8mb4&parseTime=True&loc=Local",
//	}
//	dbMgr, err := databasemgr.Build("mysql", cfg)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	defer dbMgr.Close()
//
//	// 使用 GORM 进行数据库操作
//	type User struct {
//	    ID   uint
//	    Name string
//	}
//	dbMgr.AutoMigrate(&User{})
//
//	var user User
//	dbMgr.DB().First(&user, 1)
//
// 配置选项：
//
// 不同数据库驱动的配置格式：
//
//	MySQL:
//	  cfg := &databasemgr.MySQLConfig{
//	      DSN: "user:password@tcp(localhost:3306)/dbname?charset=utf8mb4&parseTime=True&loc=Local",
//	      PoolConfig: &databasemgr.PoolConfig{
//	          MaxOpenConns:    100,
//	          MaxIdleConns:    10,
//	          ConnMaxLifetime: 3600 * time.Second,
//	          ConnMaxIdleTime: 600 * time.Second,
//	      },
//	  }
//
//	PostgreSQL:
//	  cfg := &databasemgr.PostgreSQLConfig{
//	      DSN: "host=localhost port=5432 user=postgres password=password dbname=dbname sslmode=disable",
//	      PoolConfig: &databasemgr.PoolConfig{...},
//	  }
//
//	SQLite:
//	  cfg := &databasemgr.SQLiteConfig{
//	      DSN: "file:./cache.db?cache=shared&mode=rwc",
//	      PoolConfig: &databasemgr.PoolConfig{
//	          MaxOpenConns: 1, // SQLite 通常设置为 1
//	          MaxIdleConns: 1,
//	      },
//	  }
//
// 可观测性配置：
//
//	databasemgr.ObservabilityConfig{
//	    SlowQueryThreshold: 1 * time.Second, // 慢查询阈值
//	    LogSQL:             false,            // 是否记录完整 SQL
//	    SampleRate:         1.0,              // 采样率 (0.0-1.0)
//	}
//
// 错误处理：
//
// 所有数据库操作都应检查错误。Health() 方法可用于健康检查:
//
//	if err := dbMgr.Health(); err != nil {
//	    log.Error("database health check failed", err)
//	}
//
// 性能考虑：
//
//   - 使用 Stats() 获取连接池统计信息,监控连接使用情况
//   - 合理配置连接池参数,避免连接过多或过少
//   - 使用 WithContext() 设置超时,防止长时间阻塞
//   - 对于高并发场景,建议调整 SampleRate 减少可观测性开销
