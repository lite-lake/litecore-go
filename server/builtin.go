package server

import (
	"fmt"

	"github.com/lite-lake/litecore-go/container"
	"github.com/lite-lake/litecore-go/logger"
	"github.com/lite-lake/litecore-go/manager/cachemgr"
	"github.com/lite-lake/litecore-go/manager/configmgr"
	"github.com/lite-lake/litecore-go/manager/databasemgr"
	"github.com/lite-lake/litecore-go/manager/limitermgr"
	"github.com/lite-lake/litecore-go/manager/lockmgr"
	"github.com/lite-lake/litecore-go/manager/loggermgr"
	"github.com/lite-lake/litecore-go/manager/mqmgr"
	"github.com/lite-lake/litecore-go/manager/schedulermgr"
	"github.com/lite-lake/litecore-go/manager/telemetrymgr"
)

// BuiltinConfig 内置管理器配置结构体
type BuiltinConfig struct {
	Driver   string // 配置驱动类型（如：yaml、json 等）
	FilePath string // 配置文件路径
}

// Validate 验证配置参数是否有效
func (c *BuiltinConfig) Validate() error {
	if c.Driver == "" {
		return fmt.Errorf("configmgr driver cannot be empty")
	}
	if c.FilePath == "" {
		return fmt.Errorf("configmgr file path cannot be empty")
	}
	return nil
}

// logStartup 记录启动日志
func logStartup(log logger.ILogger, phase StartupPhase, msg string, fields ...logger.Field) {
	log.Info(msg, fields...)
}

// logPhaseStart 记录阶段开始
func logPhaseStart(log logger.ILogger, phase StartupPhase, msg string, fields ...logger.Field) {
	log.Info(msg, fields...)
}

// logPhaseEnd 记录阶段结束（带耗时）
func logPhaseEnd(log logger.ILogger, phase StartupPhase, msg string, extraFields ...logger.Field) {
	fields := append(extraFields, logger.F("phase", phase.String()))
	log.Info(msg, fields...)
}

// Initialize 初始化所有内置管理器并注册到容器中
// 初始化顺序：config -> telemetry -> logger -> database -> cache -> lock -> limiter -> mq
func Initialize(cfg *BuiltinConfig) (*container.ManagerContainer, error) {

	cntr := container.NewManagerContainer()

	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configmgr: %w", err)
	}

	tempLogger := logger.NewDefaultLogger("Builtin")

	logPhaseStart(tempLogger, PhaseConfig, "开始初始化内置组件")
	logStartup(tempLogger, PhaseConfig, "配置文件: "+cfg.FilePath)
	logStartup(tempLogger, PhaseConfig, "配置驱动: "+cfg.Driver)

	// 1. 初始化配置管理器（必须最先初始化，其他管理器依赖它）
	configManager, err := configmgr.Build(cfg.Driver, cfg.FilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to create config manager: %w", err)
	}
	if err := container.RegisterManager[configmgr.IConfigManager](cntr, configManager); err != nil {
		return nil, fmt.Errorf("failed to register config manager: %w", err)
	}
	logStartup(tempLogger, PhaseManagers, "初始化完成: ConfigManager")

	// 2. 初始化遥测管理器（依赖配置管理器）
	telemetryMgr, err := telemetrymgr.BuildWithConfigProvider(configManager)
	if err != nil {
		return nil, fmt.Errorf("failed to create telemetry manager: %w", err)
	}
	if err := container.RegisterManager[telemetrymgr.ITelemetryManager](cntr, telemetryMgr); err != nil {
		return nil, fmt.Errorf("failed to register telemetry manager: %w", err)
	}
	logStartup(tempLogger, PhaseManagers, "初始化完成: TelemetryManager")

	// 3. 初始化日志管理器（依赖配置管理器和遥测管理器）
	loggerManager, err := loggermgr.BuildWithConfigProvider(configManager, telemetryMgr)
	if err != nil {
		return nil, fmt.Errorf("failed to create logger manager: %w", err)
	}
	if err := container.RegisterManager[loggermgr.ILoggerManager](cntr, loggerManager); err != nil {
		return nil, fmt.Errorf("failed to register logger manager: %w", err)
	}
	logStartup(tempLogger, PhaseManagers, "初始化完成: LoggerManager")

	// 4. 初始化数据库管理器（依赖配置管理器）
	databaseMgr, err := databasemgr.BuildWithConfigProvider(configManager)
	if err != nil {
		return nil, fmt.Errorf("failed to create database manager: %w", err)
	}
	if err := container.RegisterManager[databasemgr.IDatabaseManager](cntr, databaseMgr); err != nil {
		return nil, fmt.Errorf("failed to register database manager: %w", err)
	}
	logStartup(tempLogger, PhaseManagers, "初始化完成: DatabaseManager")

	// 5. 初始化缓存管理器（依赖配置管理器）
	cacheMgr, err := cachemgr.BuildWithConfigProvider(configManager)
	if err != nil {
		return nil, fmt.Errorf("failed to create cache manager: %w", err)
	}
	if err := container.RegisterManager[cachemgr.ICacheManager](cntr, cacheMgr); err != nil {
		return nil, fmt.Errorf("failed to register cache manager: %w", err)
	}
	logStartup(tempLogger, PhaseManagers, "初始化完成: CacheManager")

	// 6. 初始化锁管理器（依赖配置管理器）
	lockMgr, err := lockmgr.BuildWithConfigProvider(configManager)
	if err != nil {
		return nil, fmt.Errorf("failed to create lock manager: %w", err)
	}
	if err := container.RegisterManager[lockmgr.ILockManager](cntr, lockMgr); err != nil {
		return nil, fmt.Errorf("failed to register lock manager: %w", err)
	}
	logStartup(tempLogger, PhaseManagers, "初始化完成: LockManager")

	// 7. 初始化限流管理器（依赖配置管理器）
	limiterMgr, err := limitermgr.BuildWithConfigProvider(configManager)
	if err != nil {
		return nil, fmt.Errorf("failed to create limiter manager: %w", err)
	}
	if err := container.RegisterManager[limitermgr.ILimiterManager](cntr, limiterMgr); err != nil {
		return nil, fmt.Errorf("failed to register limiter manager: %w", err)
	}
	logStartup(tempLogger, PhaseManagers, "初始化完成: LimiterManager")

	// 8. 初始化消息队列管理器（依赖配置管理器）
	mqMgr, err := mqmgr.BuildWithConfigProvider(configManager)
	if err != nil {
		return nil, fmt.Errorf("failed to create mq manager: %w", err)
	}
	if err := container.RegisterManager[mqmgr.IMQManager](cntr, mqMgr); err != nil {
		return nil, fmt.Errorf("failed to register mq manager: %w", err)
	}
	logStartup(tempLogger, PhaseManagers, "初始化完成: MQManager")

	// 9. 初始化定时任务管理器（依赖配置管理器）
	schedulerMgr, err := schedulermgr.BuildWithConfigProvider(configManager)
	if err != nil {
		return nil, fmt.Errorf("failed to create scheduler manager: %w", err)
	}
	if err := container.RegisterManager[schedulermgr.ISchedulerManager](cntr, schedulerMgr); err != nil {
		return nil, fmt.Errorf("failed to register scheduler manager: %w", err)
	}
	logStartup(tempLogger, PhaseManagers, "初始化完成: SchedulerManager")

	logPhaseEnd(tempLogger, PhaseManagers, "管理器初始化完成", logger.F("count", 9))

	return cntr, nil
}
