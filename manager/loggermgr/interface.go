package loggermgr

import (
	"github.com/lite-lake/litecore-go/common"
	"github.com/lite-lake/litecore-go/logger"
)

// ILoggerManager 日志管理器接口
type ILoggerManager interface {
	common.IBaseManager

	// Ins 获取日志实例
	Ins() logger.ILogger
}
