package schedulermgr

import (
	"github.com/lite-lake/litecore-go/common"
)

// ISchedulerManager 定时任务管理器接口
type ISchedulerManager interface {
	common.IBaseManager

	// ValidateScheduler 验证定时器配置是否正确
	// 在程序加载时调用，配置错误直接 panic
	// scheduler: 待验证的定时器实例
	// 返回: 验证错误（调用方负责 panic）
	ValidateScheduler(scheduler common.IBaseScheduler) error

	// RegisterScheduler 注册定时器
	// 在 SchedulerManager.OnStart() 时由容器调用
	// scheduler: 待注册的定时器实例
	// 返回: 注册错误
	RegisterScheduler(scheduler common.IBaseScheduler) error

	// UnregisterScheduler 注销定时器
	// 在 SchedulerManager.OnStop() 时由容器调用
	// scheduler: 待注销的定时器实例
	// 返回: 注销错误
	UnregisterScheduler(scheduler common.IBaseScheduler) error
}
