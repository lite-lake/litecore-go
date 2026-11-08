package server

import (
	"fmt"
	"strings"

	"github.com/lite-lake/litecore-go/common"
	"github.com/lite-lake/litecore-go/container"
	"github.com/lite-lake/litecore-go/logger"
	"github.com/lite-lake/litecore-go/manager/cachemgr"
	"github.com/lite-lake/litecore-go/manager/configmgr"
	"github.com/lite-lake/litecore-go/manager/databasemgr"
	"github.com/lite-lake/litecore-go/manager/limitermgr"
	"github.com/lite-lake/litecore-go/manager/lockmgr"
	"github.com/lite-lake/litecore-go/manager/loggermgr"
	"github.com/lite-lake/litecore-go/manager/mqmgr"
	"github.com/lite-lake/litecore-go/manager/notificationmgr"
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

	logPhaseStart(tempLogger, PhaseConfig, "Starting to initialize builtin components")
	logStartup(tempLogger, PhaseConfig, "Config file: "+cfg.FilePath)
	logStartup(tempLogger, PhaseConfig, "Config driver: "+cfg.Driver)

	// 1. 初始化配置管理器（必须最先初始化，其他管理器依赖它）
	configManager, err := configmgr.Build(cfg.Driver, cfg.FilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to create config manager: %w", err)
	}
	if err := container.RegisterManager[configmgr.IConfigManager](cntr, configManager); err != nil {
		return nil, fmt.Errorf("failed to register config manager: %w", err)
	}
	logStartup(tempLogger, PhaseManagers, "Initialization complete: ConfigManager")

	// 2. 初始化遥测管理器（依赖配置管理器）
	telemetryMgr, err := telemetrymgr.BuildWithConfigProvider(configManager)
	if err != nil {
		return nil, fmt.Errorf("failed to create telemetry manager: %w", err)
	}
	if err := container.RegisterManager[telemetrymgr.ITelemetryManager](cntr, telemetryMgr); err != nil {
		return nil, fmt.Errorf("failed to register telemetry manager: %w", err)
	}
	logStartup(tempLogger, PhaseManagers, "Initialization complete: TelemetryManager")

	// 3. 初始化日志管理器（依赖配置管理器和遥测管理器）
	loggerManager, err := loggermgr.BuildWithConfigProvider(configManager, telemetryMgr)
	if err != nil {
		return nil, fmt.Errorf("failed to create logger manager: %w", err)
	}
	if err := container.RegisterManager[loggermgr.ILoggerManager](cntr, loggerManager); err != nil {
		return nil, fmt.Errorf("failed to register logger manager: %w", err)
	}
	logStartup(tempLogger, PhaseManagers, "Initialization complete: LoggerManager")

	// 4. 初始化数据库管理器（依赖配置管理器）
	databaseMgr, err := databasemgr.BuildWithConfigProvider(configManager, loggerManager, telemetryMgr)
	if err != nil {
		return nil, fmt.Errorf("failed to create database manager: %w", err)
	}
	if err := container.RegisterManager[databasemgr.IDatabaseManager](cntr, databaseMgr); err != nil {
		return nil, fmt.Errorf("failed to register database manager: %w", err)
	}
	logStartup(tempLogger, PhaseManagers, "Initialization complete: DatabaseManager")

	// 5. 初始化缓存管理器（依赖配置管理器）
	cacheMgr, err := cachemgr.BuildWithConfigProvider(configManager, loggerManager, telemetryMgr)
	if err != nil {
		return nil, fmt.Errorf("failed to create cache manager: %w", err)
	}
	if err := container.RegisterManager[cachemgr.ICacheManager](cntr, cacheMgr); err != nil {
		return nil, fmt.Errorf("failed to register cache manager: %w", err)
	}
	logStartup(tempLogger, PhaseManagers, "Initialization complete: CacheManager")

	// 6. 初始化锁管理器（依赖配置管理器）
	lockMgr, err := lockmgr.BuildWithConfigProvider(configManager, loggerManager, telemetryMgr, cacheMgr)
	if err != nil {
		return nil, fmt.Errorf("failed to create lock manager: %w", err)
	}
	if err := container.RegisterManager[lockmgr.ILockManager](cntr, lockMgr); err != nil {
		return nil, fmt.Errorf("failed to register lock manager: %w", err)
	}
	logStartup(tempLogger, PhaseManagers, "Initialization complete: LockManager")

	// 7. 初始化限流管理器（依赖配置管理器）
	limiterMgr, err := limitermgr.BuildWithConfigProvider(configManager, loggerManager, telemetryMgr, cacheMgr)
	if err != nil {
		return nil, fmt.Errorf("failed to create limiter manager: %w", err)
	}
	if err := container.RegisterManager[limitermgr.ILimiterManager](cntr, limiterMgr); err != nil {
		return nil, fmt.Errorf("failed to register limiter manager: %w", err)
	}
	logStartup(tempLogger, PhaseManagers, "Initialization complete: LimiterManager")

	// 7.5 一致性校验（拦「配置说谎」）：lock/limiter=redis 时，其 redis_config 与 cache.redis_config
	// 的 host/port/db 三元组必须一致——lock/limiter 驱动实际复用全局 cacheMgr 连接，redis_config 不独立生效
	if err := validateRedisConfigConsistency(configManager); err != nil {
		return nil, err
	}

	// 8. 初始化消息队列管理器（依赖配置管理器、日志管理器、遥测管理器）
	mqMgr, err := mqmgr.BuildWithConfigProvider(configManager, loggerManager, telemetryMgr)
	if err != nil {
		return nil, fmt.Errorf("failed to create mq manager: %w", err)
	}
	if err := container.RegisterManager[mqmgr.IMQManager](cntr, mqMgr); err != nil {
		return nil, fmt.Errorf("failed to register mq manager: %w", err)
	}
	logStartup(tempLogger, PhaseManagers, "Initialization complete: MQManager")

	// 9. 初始化定时任务管理器（依赖配置管理器、日志管理器）
	schedulerMgr, err := schedulermgr.BuildWithConfigProvider(configManager, loggerManager)
	if err != nil {
		return nil, fmt.Errorf("failed to create scheduler manager: %w", err)
	}
	if err := container.RegisterManager[schedulermgr.ISchedulerManager](cntr, schedulerMgr); err != nil {
		return nil, fmt.Errorf("failed to register scheduler manager: %w", err)
	}
	logStartup(tempLogger, PhaseManagers, "Initialization complete: SchedulerManager")

	// 10. 初始化服务状态通知管理器（依赖配置管理器、日志管理器）
	notificationMgr, err := notificationmgr.BuildWithConfigProvider(configManager, loggerManager)
	if err != nil {
		return nil, fmt.Errorf("failed to create notification manager: %w", err)
	}
	if err := container.RegisterManager[notificationmgr.INotificationManager](cntr, notificationMgr); err != nil {
		return nil, fmt.Errorf("failed to register notification manager: %w", err)
	}
	logStartup(tempLogger, PhaseManagers, "Initialization complete: NotificationManager")

	logPhaseEnd(tempLogger, PhaseManagers, "Managers initialization complete", logger.F("count", 10))

	return cntr, nil
}

// redisConfigTriple 从解析后的 Redis 配置中取一致性比对所需的三元组
type redisConfigTriple struct {
	Host string
	Port int
	DB   int
}

// validateRedisConfigConsistency 校验 lock/limiter 与 cache 的 redis_config 一致性（同实例同 db，拦「配置说谎」）。
// lock/limiter 驱动实际复用全局 cacheMgr 连接，其 redis_config 不独立生效；
// 配置声明不一致即「配置说谎」，在引擎装配层（同时拥有配置与三段构建结果的唯一位置）fail-fast。
// 比对口径：两侧均经各管理器同一套解析器取可信解析值（缺省键回落默认值，不误报），密码不参与比较。
func validateRedisConfigConsistency(configManager configmgr.IConfigManager) error {
	cacheDriver, err := configManager.Get("cache.driver")
	if err != nil {
		return fmt.Errorf("redis 配置一致性校验失败：无法读取 cache.driver: %w", err)
	}
	cacheDriverStr, err := common.GetString(cacheDriver)
	if err != nil {
		return fmt.Errorf("redis 配置一致性校验失败：cache.driver: %w", err)
	}
	if strings.ToLower(strings.TrimSpace(cacheDriverStr)) != "redis" {
		// cache 非 redis 时无从比对；lock/limiter=redis 而 cache 非 redis 的契约违反已由工厂层拦截
		return nil
	}

	cacheRedisAny, err := configManager.Get("cache.redis_config")
	if err != nil {
		return fmt.Errorf("redis 配置一致性校验失败：无法读取 cache.redis_config: %w", err)
	}
	cacheRedisMap, err := common.GetMap(cacheRedisAny)
	if err != nil {
		return fmt.Errorf("redis 配置一致性校验失败：cache.redis_config: %w", err)
	}
	cacheParsed, err := cachemgr.ParseCacheConfigFromMap(map[string]any{"redis_config": cacheRedisMap})
	if err != nil {
		return fmt.Errorf("redis 配置一致性校验失败：cache.redis_config: %w", err)
	}
	cacheTriple := redisConfigTriple{
		Host: cacheParsed.RedisConfig.Host,
		Port: cacheParsed.RedisConfig.Port,
		DB:   cacheParsed.RedisConfig.DB,
	}

	for _, dep := range []struct {
		name    string
		driverK string
		cfgKey  string
	}{
		{"lock", "lock.driver", "lock.redis_config"},
		{"limiter", "limiter.driver", "limiter.redis_config"},
	} {
		depDriver, err := configManager.Get(dep.driverK)
		if err != nil {
			return fmt.Errorf("redis 配置一致性校验失败：无法读取 %s: %w", dep.driverK, err)
		}
		depDriverStr, err := common.GetString(depDriver)
		if err != nil {
			return fmt.Errorf("redis 配置一致性校验失败：%s: %w", dep.driverK, err)
		}
		if strings.ToLower(strings.TrimSpace(depDriverStr)) != "redis" {
			continue
		}

		depCfgAny, err := configManager.Get(dep.cfgKey)
		if err != nil {
			return fmt.Errorf("redis 配置一致性校验失败：无法读取 %s: %w", dep.cfgKey, err)
		}
		depCfgMap, err := common.GetMap(depCfgAny)
		if err != nil {
			return fmt.Errorf("redis 配置一致性校验失败：%s: %w", dep.cfgKey, err)
		}

		var depTriple redisConfigTriple
		switch dep.name {
		case "lock":
			lockParsed, err := lockmgr.ParseLockConfigFromMap(map[string]any{"redis_config": depCfgMap})
			if err != nil {
				return fmt.Errorf("redis 配置一致性校验失败：%s: %w", dep.cfgKey, err)
			}
			depTriple = redisConfigTriple{Host: lockParsed.RedisConfig.Host, Port: lockParsed.RedisConfig.Port, DB: lockParsed.RedisConfig.DB}
		case "limiter":
			limiterParsed, err := limitermgr.ParseLimiterConfigFromMap(map[string]any{"redis_config": depCfgMap})
			if err != nil {
				return fmt.Errorf("redis 配置一致性校验失败：%s: %w", dep.cfgKey, err)
			}
			depTriple = redisConfigTriple{Host: limiterParsed.RedisConfig.Host, Port: limiterParsed.RedisConfig.Port, DB: limiterParsed.RedisConfig.DB}
		}

		if depTriple.Host != cacheTriple.Host || depTriple.Port != cacheTriple.Port || depTriple.DB != cacheTriple.DB {
			return fmt.Errorf(
				"redis 配置一致性校验失败：%s.redis_config 与 cache.redis_config 不一致（同实例同 db 契约）："+
					"host: %s=%q cache=%q；port: %s=%d cache=%d；db: %s=%d cache=%d；"+
					"（%s 驱动实际复用全局 cacheMgr 连接，redis_config 不独立生效，声明不一致即配置说谎；"+
					"修复指引：将两侧 host/port/db 修改为一致，合法取值以 docs/ops/middlewares/middleware-redis.md 分配表为准）",
				dep.name,
				dep.name, depTriple.Host, cacheTriple.Host,
				dep.name, depTriple.Port, cacheTriple.Port,
				dep.name, depTriple.DB, cacheTriple.DB,
				dep.name)
		}
	}
	return nil
}
