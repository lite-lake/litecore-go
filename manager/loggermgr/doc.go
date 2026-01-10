// Package loggermgr 提供日志管理功能，支持多种输出方式和日志级别。
//
// 核心特性：
//   - 多输出支持：支持控制台、文件和观测日志（OpenTelemetry）三种输出方式
//   - 灵活配置：通过配置文件或代码配置日志级别、输出格式和日志轮转策略
//   - 性能优化：基于 zap 高性能日志库，支持结构化日志和并发写入
//   - 日志轮转：支持按大小、时间和备份数量自动轮转日志文件
//   - 零成本降级：配置失败时自动降级到空日志器，避免影响程序运行
//
// 基本用法：
//
//	// 配置日志（启用控制台输出）
//	cfg := map[string]any{
//	    "console_enabled": true,
//	    "console_config": map[string]any{
//	        "level": "info",
//	    },
//	}
//
//	// 构建日志管理器
//	mgr := loggermgr.Build(cfg, nil)
//
//	// 获取 Logger 实例
//	logger := mgr.(loggermgr.LoggerManager).Logger("my-app")
//
//	// 记录日志
//	logger.Info("Application started", "version", "1.0.0")
//	logger.Error("Failed to connect", "error", err)
//
//	// 关闭日志管理器
//	_ = mgr.(loggermgr.LoggerManager).Shutdown(context.Background())
//
// 日志级别：
//
//	包支持五个日志级别，按优先级从低到高：
//	- DebugLevel：调试信息，用于开发阶段
//	- InfoLevel：一般信息，记录程序运行状态
//	- WarnLevel：警告信息，表示潜在问题
//	- ErrorLevel：错误信息，记录程序错误
//	- FatalLevel：致命错误，记录后退出程序
//
// 配置选项：
//
//	控制台输出配置：
//	  console_enabled: 是否启用控制台输出
//	  console_config.level: 控制台日志级别（debug/info/warn/error/fatal）
//
//	文件输出配置：
//	  file_enabled: 是否启用文件输出
//	  file_config.level: 文件日志级别
//	  file_config.path: 日志文件路径
//	  file_config.rotation.max_size: 单个日志文件最大大小（MB），如 100MB
//	  file_config.rotation.max_age: 日志文件保留天数，如 30d
//	  file_config.rotation.max_backups: 保留的旧日志文件最大数量
//	  file_config.rotation.compress: 是否压缩旧日志文件
//
//	观测日志配置：
//	  telemetry_enabled: 是否启用观测日志
//	  telemetry_config.level: 观测日志级别
//
// 文件日志轮转：
//
//	文件日志支持自动轮转，当日志文件达到指定大小时会自动创建新文件，
//	旧文件会按照配置的时间保留，超过保留时间的日志文件会被自动删除。
//	轮转配置支持多种格式：
//	- 大小：100MB、1GB、100（默认单位为 MB）
//	- 时间：30d、48h（默认单位为天）
//	- 备份数：10（整数）
//
// 错误处理：
//
//	日志管理器采用零成本降级策略：
//	- 配置解析失败时，返回空日志管理器（NoneLoggerManager）
//	- 日志初始化失败时，自动降级到控制台输出
//	- 所有日志操作都不会抛出异常，保证程序稳定运行
//
// 性能考虑：
//
//	- Logger 实例会被缓存，重复获取相同名称的 Logger 会返回同一实例
//	- 日志写入采用异步方式，避免阻塞主程序
//	- 日志级别检查在调用前完成，避免不必要的字符串格式化
//	- 支持并发写入，线程安全
package loggermgr