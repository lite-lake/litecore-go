package loggermgr

import (
	"github.com/lite-lake/litecore-go/common"
	"github.com/lite-lake/litecore-go/util/logger"
)

// ILogger 日志接口（类型别名，指向 util/logger）
type ILogger = common.ILogger

// ILoggerManager 日志管理器接口（类型别名，指向 util/logger）
type ILoggerManager = logger.ILoggerManager

// LogLevel 日志级别类型（类型别名，指向 util/logger）
type LogLevel = common.LogLevel

// ParseLogLevel 从字符串解析日志级别
var ParseLogLevel = common.ParseLogLevel

// IsValidLogLevel 检查日志级别字符串是否有效
var IsValidLogLevel = common.IsValidLogLevel
