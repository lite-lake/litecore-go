package schedulers

import (
	"github.com/lite-lake/litecore-go/common"
	"github.com/lite-lake/litecore-go/logger"
	"github.com/lite-lake/litecore-go/manager/loggermgr"
	"github.com/lite-lake/litecore-go/samples/messageboard/internal/services"
)

type ICleanupScheduler interface {
	common.IBaseScheduler
}

type cleanupSchedulerImpl struct {
	MessageService services.IMessageService `inject:""`
	LoggerMgr      loggermgr.ILoggerManager `inject:""`
	logger         logger.ILogger
}

func NewCleanupScheduler() ICleanupScheduler {
	return &cleanupSchedulerImpl{}
}

func (s *cleanupSchedulerImpl) SchedulerName() string {
	return "cleanupScheduler"
}

func (s *cleanupSchedulerImpl) GetRule() string {
	return "0 0 2 * * *"
}

func (s *cleanupSchedulerImpl) GetTimezone() string {
	return "Asia/Shanghai"
}

func (s *cleanupSchedulerImpl) OnTick(tickID int64) error {
	s.initLogger()
	s.logger.Info("开始清理任务", "tick_id", tickID)

	stats, err := s.MessageService.GetStatistics()
	if err != nil {
		s.logger.Error("获取统计信息失败", "error", err)
		return err
	}

	s.logger.Info("模拟清理任务完成", "tick_id", tickID, "total_count", stats["total"])
	return nil
}

func (s *cleanupSchedulerImpl) OnStart() error {
	s.initLogger()
	s.logger.Info("清理定时器已启动")
	return nil
}

func (s *cleanupSchedulerImpl) OnStop() error {
	s.initLogger()
	s.logger.Info("清理定时器已停止")
	return nil
}

func (s *cleanupSchedulerImpl) initLogger() {
	if s.logger == nil && s.LoggerMgr != nil {
		s.logger = s.LoggerMgr.Ins()
	}
}

var _ ICleanupScheduler = (*cleanupSchedulerImpl)(nil)
var _ common.IBaseScheduler = (*cleanupSchedulerImpl)(nil)
