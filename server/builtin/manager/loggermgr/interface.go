package loggermgr

import (
	"github.com/lite-lake/litecore-go/common"
	"github.com/lite-lake/litecore-go/logger"
)

// ILoggerManager 日志管理器接口
type ILoggerManager interface {
	common.IBaseManager

	// GetLogger 获取日志实例
	GetLogger() logger.ILogger
}
