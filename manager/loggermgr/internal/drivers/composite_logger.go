package drivers

import (
	"context"
	"fmt"
	"os"
	"sync"

	"com.litelake.litecore/common"
	"com.litelake.litecore/manager/loggermgr/internal/config"
	"com.litelake.litecore/manager/loggermgr/internal/loglevel"
	"com.litelake.litecore/manager/telemetrymgr"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// ZapLogger 基于 zap 的日志输出器
type ZapLogger struct {
	name   string
	logger *zap.Logger
	sync   *zap.SugaredLogger
	level  zapcore.Level
	mu     sync.RWMutex
}

// NewZapLogger 创建基于 zap 的日志输出器
func NewZapLogger(name string, cfg *config.LoggerConfig, telemetryMgr telemetrymgr.TelemetryManager) (*ZapLogger, error) {
	// 验证配置
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid logger config: %w", err)
	}

	// 创建各个核心
	var cores []zapcore.Core

	// 1. 观测日志核心
	if cfg.TelemetryEnabled {
		otelLevel := zapcore.InfoLevel
		if cfg.TelemetryConfig != nil && cfg.TelemetryConfig.Level != "" {
			otelLevel = LogLevelToZap(loglevel.ParseLogLevel(cfg.TelemetryConfig.Level))
		}
		cores = append(cores, NewOTELCore(otelLevel, telemetryMgr))
	}

	// 2. 控制台核心
	if cfg.ConsoleEnabled {
		consoleWriter := NewConsoleWriter(cfg.ConsoleConfig)
		cores = append(cores, consoleWriter.GetLogger().Core())
	}

	// 3. 文件核心
	if cfg.FileEnabled {
		fileWriter, err := NewFileWriter(cfg.FileConfig)
		if err != nil {
			// 文件日志初始化失败，降级到控制台输出
			if !cfg.ConsoleEnabled {
				consoleWriter := NewConsoleWriter(cfg.ConsoleConfig)
				cores = append(cores, consoleWriter.GetLogger().Core())
			}
		} else {
			cores = append(cores, fileWriter.GetLogger().Core())
		}
	}

	// 组合核心
	core := zapcore.NewTee(cores...)

	// 创建 logger
	logger := zap.New(core,
		zap.AddCallerSkip(1),
		zap.AddStacktrace(zapcore.ErrorLevel),
		zap.Fields(zap.String("logger", name)),
	)

	// 确定全局最低级别
	minLevel := zapcore.InfoLevel
	if cfg.ConsoleEnabled && cfg.ConsoleConfig != nil {
		level := LogLevelToZap(loglevel.ParseLogLevel(cfg.ConsoleConfig.Level))
		if level < minLevel {
			minLevel = level
		}
	}

	return &ZapLogger{
		name:   name,
		logger: logger,
		sync:   logger.Sugar(),
		level:  minLevel,
	}, nil
}

// Debug 记录调试级别日志
func (l *ZapLogger) Debug(msg string, args ...any) {
	l.mu.RLock()
	defer l.mu.RUnlock()

	if l.isEnabled(zapcore.DebugLevel) {
		fields := argsToFields(args...)
		l.logger.Debug(msg, fields...)
	}
}

// Info 记录信息级别日志
func (l *ZapLogger) Info(msg string, args ...any) {
	l.mu.RLock()
	defer l.mu.RUnlock()

	if l.isEnabled(zapcore.InfoLevel) {
		fields := argsToFields(args...)
		l.logger.Info(msg, fields...)
	}
}

// Warn 记录警告级别日志
func (l *ZapLogger) Warn(msg string, args ...any) {
	l.mu.RLock()
	defer l.mu.RUnlock()

	if l.isEnabled(zapcore.WarnLevel) {
		fields := argsToFields(args...)
		l.logger.Warn(msg, fields...)
	}
}

// Error 记录错误级别日志
func (l *ZapLogger) Error(msg string, args ...any) {
	l.mu.RLock()
	defer l.mu.RUnlock()

	if l.isEnabled(zapcore.ErrorLevel) {
		fields := argsToFields(args...)
		l.logger.Error(msg, fields...)
	}
}

// Fatal 记录致命错误级别日志，然后退出程序
func (l *ZapLogger) Fatal(msg string, args ...any) {
	l.mu.RLock()
	defer l.mu.RUnlock()

	if l.isEnabled(zapcore.FatalLevel) {
		fields := argsToFields(args...)
		l.logger.Fatal(msg, fields...)
	}
	// 注意：直接使用 os.Exit 可能会导致资源未正确释放
	// 建议在调用 Fatal 前确保所有资源已清理
	os.Exit(1)
}

// With 返回一个带有额外字段的新 Logger
func (l *ZapLogger) With(args ...any) *ZapLogger {
	l.mu.RLock()
	defer l.mu.RUnlock()

	fields := argsToFields(args...)
	newLogger := l.logger.With(fields...)

	return &ZapLogger{
		name:   l.name,
		logger: newLogger,
		sync:   newLogger.Sugar(),
		level:  l.level,
	}
}

// SetLevel 设置日志级别
func (l *ZapLogger) SetLevel(level zapcore.Level) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.level = level
}

// Sync 同步日志
func (l *ZapLogger) Sync() error {
	return l.logger.Sync()
}

// isEnabled 检查是否启用指定级别的日志
func (l *ZapLogger) isEnabled(level zapcore.Level) bool {
	return level >= l.level
}

// argsToFields 将键值对参数转换为 zap.Field
func argsToFields(args ...any) []zap.Field {
	fields := make([]zap.Field, 0, len(args)/2)
	for i := 0; i < len(args); i += 2 {
		if i+1 < len(args) {
			key := fmt.Sprint(args[i])
			value := args[i+1]
			fields = append(fields, zap.Any(key, value))
		}
	}
	return fields
}

// ZapLoggerManager zap 日志管理器
type ZapLoggerManager struct {
	*BaseManager
	config       *config.LoggerConfig
	telemetryMgr telemetrymgr.TelemetryManager
	globalLevel  zapcore.Level
	mu           sync.RWMutex
	loggers      map[string]*ZapLogger
	shutdownOnce sync.Once
}

// NewZapLoggerManager 创建 zap 日志管理器
func NewZapLoggerManager(cfg *config.LoggerConfig, telemetryMgr telemetrymgr.TelemetryManager) (*ZapLoggerManager, error) {
	mgr := &ZapLoggerManager{
		BaseManager:  NewBaseManager("zap-logger"),
		config:       cfg,
		telemetryMgr: telemetryMgr,
		globalLevel:  zapcore.InfoLevel,
		loggers:      make(map[string]*ZapLogger),
	}

	// 验证配置
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid logger config: %w", err)
	}

	return mgr, nil
}

// GetLogger 获取指定名称的 Logger 实例（内部方法）
func (m *ZapLoggerManager) GetLogger(name string) *ZapLogger {
	m.mu.Lock()
	defer m.mu.Unlock()

	// 如果已存在，直接返回
	if logger, ok := m.loggers[name]; ok {
		return logger
	}

	// 创建新的 zap logger
	logger, err := NewZapLogger(name, m.config, m.telemetryMgr)
	if err != nil {
		// 如果创建失败，返回一个默认的 logger
		defaultCfg := &config.LoggerConfig{
			ConsoleEnabled: true,
			ConsoleConfig:  &config.LogLevelConfig{Level: "info"},
		}
		logger, _ = NewZapLogger(name, defaultCfg, nil)
	}

	m.loggers[name] = logger
	return logger
}

// SetGlobalLevel 设置全局日志级别
func (m *ZapLoggerManager) SetGlobalLevel(level zapcore.Level) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.globalLevel = level
}

// Shutdown 关闭日志管理器，刷新所有待处理的日志
func (m *ZapLoggerManager) Shutdown(ctx context.Context) error {
	var shutdownErr error

	m.shutdownOnce.Do(func() {
		m.mu.Lock()
		defer m.mu.Unlock()

		// 同步所有日志器
		for _, logger := range m.loggers {
			if err := logger.Sync(); err != nil {
				shutdownErr = fmt.Errorf("failed to sync logger: %w", err)
			}
		}

		// 清空日志器映射
		m.loggers = make(map[string]*ZapLogger)
	})

	return shutdownErr
}

// OnStart 在服务器启动时触发
func (m *ZapLoggerManager) OnStart() error {
	// 日志文件目录创建由 FileWriter 自动处理
	return nil
}

// ensure ZapLoggerManager implements common.Manager interface
var _ common.Manager = (*ZapLoggerManager)(nil)
