// Package loggermgr 提供日志管理功能
//
// 日志管理器支持多种驱动：
//   - zap: 基于 Uber Zap 的高性能日志实现，支持控制台、文件和 OTEL 输出
//   - none: 空实现，用于测试或禁用日志的场景
//
// 使用示例：
//
//	// 使用工厂函数创建
//	mgr, err := loggermgr.BuildWithConfigProvider(provider)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	mgr.OnStart()
//	defer mgr.Shutdown(ctx)
//
//	// 获取 Logger 实例
//	logger := mgr.Logger("my-app")
//	logger.Info("Application started", "version", "1.0.0")
//
// 配置示例（YAML）：
//
//	logger:
//	  driver: zap
//	  zap_config:
//	    telemetry_enabled: false
//	    console_enabled: true
//	    console_config:
//	      level: info
//	    file_enabled: true
//	    file_config:
//	      path: ./logs/app.log
//	      level: info
//	      rotation:
//	        max_size: 100
//	        max_age: 30
//	        max_backups: 10
//	        compress: true
package loggermgr
