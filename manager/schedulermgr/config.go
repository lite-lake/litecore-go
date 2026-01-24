package schedulermgr

// CronConfig Crontab 定时器配置
type CronConfig struct {
	ValidateOnStartup bool `yaml:"validate_on_startup"` // 启动时是否检查所有 Scheduler 配置
}
