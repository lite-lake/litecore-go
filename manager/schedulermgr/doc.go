package schedulermgr

// Package schedulermgr 提供定时任务管理器功能，支持 Crontab 表达式调度。
//
// 核心特性：
//   - 支持标准 6 段式 Crontab 表达式（秒 分 时 日 月 周）
//   - 支持时区配置，每个定时器可独立设置时区
//   - 完全并发执行，每次触发启动独立协程
//   - 支持配置验证，启动时检查所有定时器配置
//   - 支持 panic 捕获，防止定时器崩溃
//   - 线程安全，支持并发使用
//
// 基本用法：
//
//	cfg := &schedulermgr.CronConfig{
//	    ValidateOnStartup: true,
//	}
//	mgr := schedulermgr.NewSchedulerManagerCronImpl(cfg)
//
//	scheduler := &mySchedulerImpl{}
//	if err := mgr.ValidateScheduler(scheduler); err != nil {
//	    panic(err)
//	}
//	if err := mgr.RegisterScheduler(scheduler); err != nil {
//	    panic(err)
//	}
//	scheduler.OnStart()
//
// Crontab 表达式格式：
//
//	┌─────────────── 秒 (0-59)
//	│ ┌───────────── 分 (0-59)
//	│ │ ┌─────────── 时 (0-23)
//	│ │ │ ┌───────── 日 (1-31)
//	│ │ │ │ ┌─────── 月 (1-12)
//	│ │ │ │ │ ┌───── 周 (0-6, 0=周日)
//	│ │ │ │ │ │
//	* * * * * *
//
// 支持的特殊字符：
//   * : 任意值
//   , : 多个值（如 0,30 表示第 0 秒和第 30 秒）
//   - : 范围（如 0-29 表示第 0 到 29 秒）
//   / : 步长（如 */10 表示每 10 秒）
//   ? : 不指定（仅用于日和周）
//
// 常用表达式示例：
//   "0 * * * * *"        : 每分钟的第 0 秒执行
//   "*/5 * * * * *"      : 每 5 秒执行一次
//   "0 */5 * * * *"      : 每 5 分钟执行一次
//   "0 0 * * * *"        : 每小时执行
//   "0 0 0 * * *"        : 每天凌晨执行
//   "0 0 0 * * 0"        : 每周日凌晨执行
//   "0 0 2 * * *"        : 每天凌晨 2 点执行
//
// 时区支持：
//   返回空字符串时使用服务器本地时间
//   支持标准时区名称，如 "Asia/Shanghai", "UTC"
//
// 并发执行：
//   每次触发定时任务时，启动独立的 goroutine 执行 OnTick()
//   不等待前一次执行完成，允许并发执行
