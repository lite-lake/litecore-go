package common

// IBaseScheduler 基础定时器接口
// 所有 Scheduler 类必须继承此接口并实现相关方法
// 用于定义定时器的基础行为和契约
type IBaseScheduler interface {
	// SchedulerName 返回定时器名称
	// 格式：xxxScheduler（小驼峰）
	// 示例："cleanupScheduler"
	SchedulerName() string

	// GetRule 返回 Crontab 定时规则
	// 使用标准 6 段式格式：秒 分 时 日 月 周
	// 示例："0 */5 * * * *" 表示每 5 分钟执行一次
	//      "0 0 2 * * *" 表示每天凌晨 2 点执行
	//      "0 0 * * * 1" 表示每周一凌晨执行
	GetRule() string

	// GetTimezone 返回定时器使用的时区
	// 返回空字符串时使用服务器本地时间
	// 支持标准时区名称，如 "Asia/Shanghai", "UTC", "America/New_York"
	// 默认值：空字符串（服务器本地时间）
	GetTimezone() string

	// OnTick 定时触发时调用
	// tickID: 计划执行时间的 Unix 时间戳（秒级），可用于去重或日志追踪
	// 返回: 执行错误（返回 error 不会触发重试，仅记录日志）
	OnTick(tickID int64) error

	// OnStart 在服务器启动时触发
	// 用于初始化定时器状态、连接资源等
	OnStart() error

	// OnStop 在服务器停止时触发
	// 用于清理资源、保存状态等
	OnStop() error
}
