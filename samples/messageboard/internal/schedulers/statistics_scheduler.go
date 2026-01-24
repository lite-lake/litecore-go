package schedulers

import (
	"github.com/lite-lake/litecore-go/common"
	"github.com/lite-lake/litecore-go/logger"
	"github.com/lite-lake/litecore-go/manager/loggermgr"
	"github.com/lite-lake/litecore-go/samples/messageboard/internal/services"
)

type IStatisticsScheduler interface {
	common.IBaseScheduler
}

type statisticsSchedulerImpl struct {
	MessageService services.IMessageService `inject:""`
	LoggerMgr      loggermgr.ILoggerManager `inject:""`
	logger         logger.ILogger
}

func NewStatisticsScheduler() IStatisticsScheduler {
	return &statisticsSchedulerImpl{}
}

func (s *statisticsSchedulerImpl) SchedulerName() string {
	return "statisticsScheduler"
}

func (s *statisticsSchedulerImpl) GetRule() string {
	return "0 0 * * * *"
}

func (s *statisticsSchedulerImpl) GetTimezone() string {
	return "Asia/Shanghai"
}

func (s *statisticsSchedulerImpl) OnTick(tickID int64) error {
	s.initLogger()
	s.logger.Info("开始统计任务", "tick_id", tickID)

	stats, err := s.MessageService.GetStatistics()
	if err != nil {
		s.logger.Error("获取统计信息失败", "error", err)
		return err
	}

	s.logger.Info("统计任务完成", "tick_id", tickID, "pending", stats["pending"], "approved", stats["approved"], "rejected", stats["rejected"], "total", stats["total"])
	return nil
}

func (s *statisticsSchedulerImpl) OnStart() error {
	s.initLogger()
	s.logger.Info("统计定时器已启动")
	return nil
}

func (s *statisticsSchedulerImpl) OnStop() error {
	s.initLogger()
	s.logger.Info("统计定时器已停止")
	return nil
}

func (s *statisticsSchedulerImpl) initLogger() {
	if s.logger == nil && s.LoggerMgr != nil {
		s.logger = s.LoggerMgr.Ins()
	}
}

var _ IStatisticsScheduler = (*statisticsSchedulerImpl)(nil)
var _ common.IBaseScheduler = (*statisticsSchedulerImpl)(nil)
