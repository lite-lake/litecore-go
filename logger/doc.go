// Package logger 提供统一的日志记录接口和默认实现。
//
// 核心特性：
//   - 统一的日志接口：定义了 Debug、Info、Warn、Error、Fatal 五个级别
//   - 灵活的日志级别控制：支持运行时动态调整日志级别
//   - 上下文字段传递：通过 With 方法添加额外字段到日志中
//   - 序列化支持：LogLevel 类型实现了 TextMarshaler 和 TextUnmarshaler 接口
//   - Zap 集成：提供与 uber-go/zap 框架的级别转换函数
//
// 基本用法：
//
//	// 创建日志记录器
//	logger := logger.NewDefaultLogger("MyService")
//
//	// 设置日志级别
//	logger.SetLevel(logger.DebugLevel)
//
//	// 记录各级别日志
//	logger.Debug("调试信息", "request_id", "12345")
//	logger.Info("操作成功", "user_id", 100)
//	logger.Warn("查询较慢", "duration", "500ms")
//	logger.Error("连接失败", "host", "example.com")
//
//	// 使用 With 添加上下文字段
//	logger.With("service", "payment").Info("处理订单", "order_id", "1001")
//
// 日志级别：
//
// 日志级别从低到高依次为：Debug、Info、Warn、Error、Fatal。
// 设置某个级别后，只会记录该级别及以上的日志。
//
//	// 只记录 Error 及以上级别的日志
//	logger.SetLevel(logger.ErrorLevel)
//	logger.Debug("这条不会被记录")
//	logger.Error("这条会被记录")
//
// 与 Zap 集成：
//
//	// 将 LogLevel 转换为 zapcore.Level
//	zapLevel := logger.LogLevelToZap(logger.InfoLevel)
//
//	// 将 zapcore.Level 转换为 LogLevel
//	logLevel := logger.ZapToLogLevel(zapcore.DebugLevel)
package logger
