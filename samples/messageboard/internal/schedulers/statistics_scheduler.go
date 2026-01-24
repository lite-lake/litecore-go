package schedulers

import (
	"github.com/lite-lake/litecore-go/common"
	"github.com/lite-lake/litecore-go/manager/loggermgr"
	"github.com/lite-lake/litecore-go/samples/messageboard/internal/services"
)

// IStatisticsScheduler 统计定时器接口
type IStatisticsScheduler interface {
	common.IBaseScheduler
}

type statisticsSchedulerImpl struct {
	MessageService services.IMessageService `inject:""` // 留言服务
	LoggerMgr      loggermgr.ILoggerManager `inject:""` // 日志管理器
}

// NewStatisticsScheduler 创建统计定时器实例
func NewStatisticsScheduler() IStatisticsScheduler {
	return &statisticsSchedulerImpl{}
}

// SchedulerName 返回调度器名称
func (s *statisticsSchedulerImpl) SchedulerName() string {
	return "statisticsScheduler"
}

// GetRule 返回 cron 表达式
func (s *statisticsSchedulerImpl) GetRule() string {
	return "0 0 * * * *"
}

// GetTimezone 返回时区
func (s *statisticsSchedulerImpl) GetTimezone() string {
	return "Asia/Shanghai"
}

// OnTick 定时任务执行回调
func (s *statisticsSchedulerImpl) OnTick(tickID int64) error {
	s.LoggerMgr.Ins().Info("Starting statistics task", "tick_id", tickID)

	stats, err := s.MessageService.GetStatistics()
	if err != nil {
		s.LoggerMgr.Ins().Error("Failed to get statistics", "error", err)
		return err
	}

	s.LoggerMgr.Ins().Info("Statistics task completed", "tick_id", tickID, "pending", stats["pending"], "approved", stats["approved"], "rejected", stats["rejected"], "total", stats["total"])
	return nil
}

// OnStart 调度器启动回调
func (s *statisticsSchedulerImpl) OnStart() error {
	s.LoggerMgr.Ins().Info("Statistics scheduler started")
	return nil
}

// OnStop 调度器停止回调
func (s *statisticsSchedulerImpl) OnStop() error {
	s.LoggerMgr.Ins().Info("Statistics scheduler stopped")
	return nil
}

var _ IStatisticsScheduler = (*statisticsSchedulerImpl)(nil)
var _ common.IBaseScheduler = (*statisticsSchedulerImpl)(nil)
