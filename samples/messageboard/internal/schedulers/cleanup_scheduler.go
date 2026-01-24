package schedulers

import (
	"github.com/lite-lake/litecore-go/common"
	"github.com/lite-lake/litecore-go/logger"
	"github.com/lite-lake/litecore-go/manager/loggermgr"
	"github.com/lite-lake/litecore-go/samples/messageboard/internal/services"
)

// ICleanupScheduler 清理定时器接口
type ICleanupScheduler interface {
	common.IBaseScheduler
}

type cleanupSchedulerImpl struct {
	MessageService services.IMessageService `inject:""` // 留言服务
	LoggerMgr      loggermgr.ILoggerManager `inject:""` // 日志管理器
	logger         logger.ILogger
}

// NewCleanupScheduler 创建清理定时器实例
func NewCleanupScheduler() ICleanupScheduler {
	return &cleanupSchedulerImpl{}
}

// SchedulerName 返回调度器名称
func (s *cleanupSchedulerImpl) SchedulerName() string {
	return "cleanupScheduler"
}

// GetRule 返回 cron 表达式
func (s *cleanupSchedulerImpl) GetRule() string {
	return "0 0 2 * * *"
}

// GetTimezone 返回时区
func (s *cleanupSchedulerImpl) GetTimezone() string {
	return "Asia/Shanghai"
}

// OnTick 定时任务执行回调
func (s *cleanupSchedulerImpl) OnTick(tickID int64) error {
	s.initLogger()
	s.logger.Info("Starting cleanup task", "tick_id", tickID)

	stats, err := s.MessageService.GetStatistics()
	if err != nil {
		s.logger.Error("Failed to get statistics", "error", err)
		return err
	}

	s.logger.Info("Mock cleanup task completed", "tick_id", tickID, "total_count", stats["total"])
	return nil
}

// OnStart 调度器启动回调
func (s *cleanupSchedulerImpl) OnStart() error {
	s.initLogger()
	s.logger.Info("Cleanup scheduler started")
	return nil
}

// OnStop 调度器停止回调
func (s *cleanupSchedulerImpl) OnStop() error {
	s.initLogger()
	s.logger.Info("Cleanup scheduler stopped")
	return nil
}

// initLogger 初始化日志记录器
func (s *cleanupSchedulerImpl) initLogger() {
	if s.logger == nil && s.LoggerMgr != nil {
		s.logger = s.LoggerMgr.Ins()
	}
}

var _ ICleanupScheduler = (*cleanupSchedulerImpl)(nil)
var _ common.IBaseScheduler = (*cleanupSchedulerImpl)(nil)
