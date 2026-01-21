package loggermgr

import (
	"github.com/lite-lake/litecore-go/util/logger"
)

// ILogger 日志接口（类型别名，指向 util/logger）
type ILogger = logger.ILogger

// ILoggerManager 日志管理器接口（类型别名，指向 util/logger）
type ILoggerManager = logger.ILoggerManager

// LogLevel 日志级别类型（类型别名，指向 util/logger）
type LogLevel = logger.LogLevel

// ParseLogLevel 从字符串解析日志级别
var ParseLogLevel = logger.ParseLogLevel

// IsValidLogLevel 检查日志级别字符串是否有效
var IsValidLogLevel = logger.IsValidLogLevel
