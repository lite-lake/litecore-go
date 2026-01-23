package loggermgr

// Package loggermgr 提供日志管理器功能，支持多种日志驱动。
//
// 核心特性：
//   - 支持三种驱动：zap（高性能）、default（简单）、none（空实现）
//   - 支持 OpenTelemetry 可观测性集成
//   - 支持多输出：控制台、文件、观测日志
//   - 支持日志轮转和文件压缩
//   - 支持自定义日志级别和格式
//   - 线程安全，支持并发使用
//
// 基本用法：
//
//	cfg := &loggermgr.Config{
//	    Driver: "zap",
//	    ZapConfig: &loggermgr.DriverZapConfig{
//	        ConsoleEnabled: true,
//	        ConsoleConfig:  &loggermgr.LogLevelConfig{Level: "info"},
//	    },
//	}
//	mgr, err := loggermgr.NewLoggerManager(cfg, telemetryMgr)
//	if err != nil {
//	    panic(err)
//	}
//	log := mgr.Ins()
//	log.Info("应用启动", "port", 8080)
//
// 配置示例：
//
//	driver: zap
//	zap_config:
//	  console_enabled: true
//	  console_config:
//	    level: info
//	  file_enabled: true
//	  file_config:
//	    level: debug
//	    path: ./logs/app.log
//	    rotation:
//	      max_size: 100
//	      max_age: 30
//	      max_backups: 10
//	      compress: true
//	  telemetry_enabled: true
//	  telemetry_config:
//	    level: info
//
// 日志级别：
//
//	Loggermgr 支持 5 个日志级别：debug、info、warn、error、fatal。
//	默认级别为 info，可通过 SetLevel 方法动态调整。
//
//	日志接口：
//
//	ILoggerManager 接口继承自 common.IBaseManager，提供日志实例获取功能。
//	ILogger 接口提供 Debug、Info、Warn、Error、Fatal 五个级别的方法，
//	以及 With 和 SetLevel 方法用于上下文和级别控制。
