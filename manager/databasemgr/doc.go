package databasemgr

// Package databasemgr 提供数据库管理功能，支持 MySQL、PostgreSQL 和 SQLite，基于 GORM 实现。
//
// 核心特性：
//   - 多数据库支持：MySQL、PostgreSQL、SQLite
//   - 依赖注入支持：通过 Container 自动注入配置和依赖
//   - 连接池管理：支持连接池配置和状态监控
//   - 事务支持：完整的事务管理和自动迁移功能
//   - 生命周期管理：集成服务启停接口，支持健康检查
//   - 可观测性：集成 OpenTelemetry，支持链路追踪、指标和日志
//
// 基本用法：
//
//	// 创建数据库管理器
//	mgr := databasemgr.NewManager("primary")
//
//	// 通过依赖注入容器初始化（推荐）
//	container.Register("config", configProvider)
//	container.Register("logger.default", loggermgr.NewManager("default"))
//	container.Register("telemetry.default", telemetrymgr.NewManager("default"))
//	container.Register("database.primary", mgr)
//	container.InjectAll()
//	mgr.OnStart()
//	defer mgr.OnStop()
//
//	// 使用 GORM 操作数据库
//	db := mgr.DB()
//	db.Create(&User{Name: "test"})
//
// 可观测性：
//
//	Manager 支持可选的日志和遥测功能，通过依赖注入接入：
//
//	container.Register("logger.default", loggermgr.NewManager("default"))
//	container.Register("telemetry.default", telemetrymgr.NewManager("default"))
//	container.Register("database.default", databasemgr.NewManager("default"))
//	container.InjectAll()
//
//	配置可观测性选项（在数据库配置中）：
//	  - SlowQueryThreshold：慢查询阈值
//	  - LogSQL：是否记录完整 SQL（生产环境建议关闭）
//	  - SampleRate：采样率（0.0-1.0）
//
//	自动采集的指标：
//	  - db.query.duration：查询耗时直方图
//	  - db.query.count：查询计数器
//	  - db.query.error_count：查询错误计数器
//	  - db.query.slow_count：慢查询计数器
//	  - db.connection.pool：连接池状态
//
// 配置选项：
//
//	通过依赖注入自动从配置中读取（配置键：database.{manager_name}）：
//
//	database.primary:
//	  driver: mysql
//	  mysql_config:
//	    dsn: user:password@tcp(localhost:3306)/dbname?charset=utf8mb4
//	    pool_config:
//	      max_open_conns: 20
//	      max_idle_conns: 10
//	      conn_max_lifetime: 30s
//	      conn_max_idle_time: 5m
//	  observability_config:
//	    slow_query_threshold: 1s
//	    log_sql: false
//	    sample_rate: 1.0
//
//	连接池配置（PoolConfig）：
//	  - MaxOpenConns：最大打开连接数
//	  - MaxIdleConns：最大空闲连接数
//	  - ConnMaxLifetime：连接最大存活时间
//	  - ConnMaxIdleTime：连接最大空闲时间
//
//	可观测性配置（ObservabilityConfig）：
//	  - SlowQueryThreshold：慢查询阈值（默认 1s）
//	  - LogSQL：是否记录完整 SQL（默认 false）
//	  - SampleRate：采样率（默认 1.0）
//
//	各驱动的 DSN 格式：
//	  - SQLite：file:./data.db?cache=shared&mode=rwc
//	  - MySQL：user:password@tcp(localhost:3306)/dbname?charset=utf8mb4&parseTime=True&loc=Local
//	  - PostgreSQL：host=localhost port=5432 user=postgres password=password dbname=dbname sslmode=disable
//
// 支持的驱动类型：
//
//   - "mysql": MySQL 数据库
//   - "postgresql": PostgreSQL 数据库
//   - "sqlite": SQLite 数据库
//   - "none": 无数据库（空管理器）
//
// 错误处理：
//
//	OnStart 方法会验证配置并返回详细的错误信息。
//	配置解析失败或驱动初始化失败时，会返回 NoneDatabaseManager（空管理器）。
//
// 线程安全：
//
//	DatabaseManager 的所有方法都是线程安全的，可以在多个 goroutine 中并发使用。
