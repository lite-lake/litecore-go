package loggermgr

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"go.opentelemetry.io/otel/log"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"

	"github.com/lite-lake/litecore-go/component/manager/telemetrymgr"
)

// zapLoggerManagerImpl Zap 日志管理器实现
type zapLoggerManagerImpl struct {
	*loggerManagerBaseImpl
	config       *LoggerConfig
	telemetryMgr telemetrymgr.ITelemetryManager
	mu           sync.RWMutex
	loggers      map[string]*zapLoggerImpl
}

// NewLoggerManagerZapImpl 创建 Zap 日志管理器实现
func NewLoggerManagerZapImpl(cfg *LoggerConfig) (ILoggerManager, error) {
	if cfg.Driver != "zap" {
		return nil, fmt.Errorf("invalid driver for zap manager: %s", cfg.Driver)
	}

	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid logger config: %w", err)
	}

	mgr := &zapLoggerManagerImpl{
		loggerManagerBaseImpl: newLoggerManagerBaseImpl("zap-logger"),
		config:                cfg,
		telemetryMgr:          nil, // 通过依赖注入设置
		loggers:               make(map[string]*zapLoggerImpl),
	}

	return mgr, nil
}

// setTelemetryMgr 设置 TelemetryManager（依赖注入）
func (m *zapLoggerManagerImpl) setTelemetryMgr(telemetryMgr telemetrymgr.ITelemetryManager) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.telemetryMgr = telemetryMgr
}

// Logger 获取指定名称的 Logger 实例
func (m *zapLoggerManagerImpl) Logger(name string) ILogger {
	m.mu.Lock()
	defer m.mu.Unlock()

	// 如果已存在，直接返回
	if logger, ok := m.loggers[name]; ok {
		return logger
	}

	// 创建新的 zap logger
	logger, err := newZapLogger(name, m.config, m.telemetryMgr)
	if err != nil {
		// 如果创建失败，返回一个默认的 logger
		defaultCfg := &LoggerConfig{
			Driver: "zap",
			ZapConfig: &ZapConfig{
				ConsoleEnabled: true,
				ConsoleConfig:  &LogLevelConfig{Level: "info"},
			},
		}
		logger, _ = newZapLogger(name, defaultCfg, nil)
	}

	m.loggers[name] = logger
	return logger
}

// SetGlobalLevel 设置全局日志级别
func (m *zapLoggerManagerImpl) SetGlobalLevel(level LogLevel) {
	m.setGlobalLevel(LogLevelToZap(level))

	// 更新所有现有 logger 的级别
	m.mu.Lock()
	defer m.mu.Unlock()
	zapLevel := LogLevelToZap(level)
	for _, logger := range m.loggers {
		logger.setLevel(zapLevel)
	}
}

// Shutdown 关闭日志管理器，刷新所有待处理的日志
func (m *zapLoggerManagerImpl) Shutdown(ctx context.Context) error {
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
		m.loggers = make(map[string]*zapLoggerImpl)
	})

	return shutdownErr
}

// zapLoggerImpl Zap 日志输出器实现
type zapLoggerImpl struct {
	name   string
	logger *zap.Logger
	sync   *zap.SugaredLogger
	level  zapcore.Level
	mu     sync.RWMutex
}

// newZapLogger 创建基于 zap 的日志输出器
func newZapLogger(name string, cfg *LoggerConfig, telemetryMgr telemetrymgr.ITelemetryManager) (*zapLoggerImpl, error) {
	// 验证配置
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid logger config: %w", err)
	}

	// 创建各个核心
	var cores []zapcore.Core

	// 1. 观测日志核心
	if cfg.ZapConfig.TelemetryEnabled {
		otelLevel := zapcore.InfoLevel
		if cfg.ZapConfig.TelemetryConfig != nil && cfg.ZapConfig.TelemetryConfig.Level != "" {
			otelLevel = LogLevelToZap(ParseLogLevel(cfg.ZapConfig.TelemetryConfig.Level))
		}
		cores = append(cores, newOTELCore(otelLevel, telemetryMgr))
	}

	// 2. 控制台核心
	if cfg.ZapConfig.ConsoleEnabled {
		consoleWriter := newConsoleWriter(cfg.ZapConfig.ConsoleConfig)
		cores = append(cores, consoleWriter.GetLogger().Core())
	}

	// 3. 文件核心
	if cfg.ZapConfig.FileEnabled {
		fileWriter, err := newFileWriter(cfg.ZapConfig.FileConfig)
		if err != nil {
			// 文件日志初始化失败，降级到控制台输出
			if !cfg.ZapConfig.ConsoleEnabled {
				consoleWriter := newConsoleWriter(cfg.ZapConfig.ConsoleConfig)
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
	if cfg.ZapConfig.ConsoleEnabled && cfg.ZapConfig.ConsoleConfig != nil {
		level := LogLevelToZap(ParseLogLevel(cfg.ZapConfig.ConsoleConfig.Level))
		if level < minLevel {
			minLevel = level
		}
	}

	return &zapLoggerImpl{
		name:   name,
		logger: logger,
		sync:   logger.Sugar(),
		level:  minLevel,
	}, nil
}

// Debug 记录调试级别日志
func (l *zapLoggerImpl) Debug(msg string, args ...any) {
	l.mu.RLock()
	defer l.mu.RUnlock()

	if l.isEnabled(zapcore.DebugLevel) {
		fields := argsToFields(args...)
		l.logger.Debug(msg, fields...)
	}
}

// Info 记录信息级别日志
func (l *zapLoggerImpl) Info(msg string, args ...any) {
	l.mu.RLock()
	defer l.mu.RUnlock()

	if l.isEnabled(zapcore.InfoLevel) {
		fields := argsToFields(args...)
		l.logger.Info(msg, fields...)
	}
}

// Warn 记录警告级别日志
func (l *zapLoggerImpl) Warn(msg string, args ...any) {
	l.mu.RLock()
	defer l.mu.RUnlock()

	if l.isEnabled(zapcore.WarnLevel) {
		fields := argsToFields(args...)
		l.logger.Warn(msg, fields...)
	}
}

// Error 记录错误级别日志
func (l *zapLoggerImpl) Error(msg string, args ...any) {
	l.mu.RLock()
	defer l.mu.RUnlock()

	if l.isEnabled(zapcore.ErrorLevel) {
		fields := argsToFields(args...)
		l.logger.Error(msg, fields...)
	}
}

// Fatal 记录致命错误级别日志，然后退出程序
func (l *zapLoggerImpl) Fatal(msg string, args ...any) {
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
func (l *zapLoggerImpl) With(args ...any) ILogger {
	l.mu.RLock()
	defer l.mu.RUnlock()

	fields := argsToFields(args...)
	newLogger := l.logger.With(fields...)

	return &zapLoggerImpl{
		name:   l.name,
		logger: newLogger,
		sync:   newLogger.Sugar(),
		level:  l.level,
	}
}

// SetLevel 设置日志级别
func (l *zapLoggerImpl) SetLevel(level LogLevel) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.level = LogLevelToZap(level)
}

// setLevel 设置 zap 日志级别（内部方法）
func (l *zapLoggerImpl) setLevel(level zapcore.Level) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.level = level
}

// Sync 同步日志
func (l *zapLoggerImpl) Sync() error {
	return l.logger.Sync()
}

// isEnabled 检查是否启用指定级别的日志
func (l *zapLoggerImpl) isEnabled(level zapcore.Level) bool {
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

// ========== Console Writer ==========

var (
	// 全局终端颜色支持检测，缓存检测结果
	supportsColor bool
	colorChecked  bool
)

// consoleWriter 控制台日志输出器（基于 zap）
type consoleWriter struct {
	logger *zap.Logger
	level  zapcore.Level
	sync   *zap.SugaredLogger
	mu     sync.RWMutex
}

// newConsoleWriter 创建控制台日志输出器
func newConsoleWriter(cfg *LogLevelConfig) *consoleWriter {
	level := zapcore.InfoLevel
	if cfg != nil && cfg.Level != "" {
		level = LogLevelToZap(ParseLogLevel(cfg.Level))
	}

	// 延迟检测颜色支持
	if !colorChecked {
		supportsColor = detectColorSupport()
		colorChecked = true
	}

	// 创建控制台编码器配置
	encoderConfig := zapcore.EncoderConfig{
		TimeKey:          "time",
		LevelKey:         "level",
		NameKey:          "logger",
		CallerKey:        "caller",
		MessageKey:       "msg",
		StacktraceKey:    "stacktrace",
		LineEnding:       zapcore.DefaultLineEnding,
		EncodeLevel:      customLevelEncoder,
		EncodeTime:       CustomTimeEncoder,
		EncodeDuration:   zapcore.StringDurationEncoder,
		EncodeCaller:     zapcore.ShortCallerEncoder,
		ConsoleSeparator: " | ",
	}

	// 创建控制台编码器
	encoder := zapcore.NewConsoleEncoder(encoderConfig)

	// 创建核心（输出到 stdout/stderr）
	stdoutCore := zapcore.NewCore(encoder, zapcore.AddSync(os.Stdout), level)
	stderrCore := zapcore.NewCore(encoder, zapcore.AddSync(os.Stderr), zapcore.ErrorLevel)

	// 组合核心
	core := zapcore.NewTee(stdoutCore, stderrCore)

	// 创建 logger
	logger := zap.New(core, zap.AddCallerSkip(1), zap.AddStacktrace(zapcore.ErrorLevel))

	return &consoleWriter{
		logger: logger,
		level:  level,
		sync:   logger.Sugar(),
	}
}

// detectColorSupport 检测终端是否支持颜色
func detectColorSupport() bool {
	// 检查 NO_COLOR 环境变量（标准）
	if os.Getenv("NO_COLOR") != "" {
		return false
	}

	// 检查 TERM 环境变量
	term := os.Getenv("TERM")
	if term == "" || term == "dumb" {
		return false
	}

	// 检查是否在 Windows 终端中
	// Windows 10+ 的终端支持 ANSI 颜色
	// 通过检查是否在 CI 环境中
	if os.Getenv("CI") != "" {
		return false
	}

	// 检查输出是否为终端
	if fileInfo, _ := os.Stdout.Stat(); (fileInfo.Mode() & os.ModeCharDevice) == 0 {
		return false
	}

	return true
}

// customLevelEncoder 自定义级别编码器（带颜色）
func customLevelEncoder(level zapcore.Level, enc zapcore.PrimitiveArrayEncoder) {
	// 延迟检测颜色支持
	if !colorChecked {
		supportsColor = detectColorSupport()
		colorChecked = true
	}

	const (
		colorReset  = "\033[0m"
		colorRed    = "\033[31m"
		colorYellow = "\033[33m"
		colorGreen  = "\033[32m"
		colorGray   = "\033[90m"
		colorBold   = "\033[1m"
	)

	var levelStr string
	var color string

	switch level {
	case zapcore.DebugLevel:
		levelStr = "DEBUG"
		color = colorGray
	case zapcore.InfoLevel:
		levelStr = "INFO "
		color = colorGreen
	case zapcore.WarnLevel:
		levelStr = "WARN "
		color = colorYellow
	case zapcore.ErrorLevel:
		levelStr = "ERROR"
		color = colorRed
	case zapcore.FatalLevel, zapcore.PanicLevel, zapcore.DPanicLevel:
		levelStr = "FATAL"
		color = colorRed + colorBold
	default:
		levelStr = level.String()
		color = colorReset
	}

	if supportsColor {
		enc.AppendString(color + levelStr + colorReset)
	} else {
		enc.AppendString(levelStr)
	}
}

// GetLogger 获取底层 zap.Logger
func (w *consoleWriter) GetLogger() *zap.Logger {
	w.mu.RLock()
	defer w.mu.RUnlock()
	return w.logger
}

// ========== File Writer ==========

// fileWriter 文件日志输出器（基于 zap）
type fileWriter struct {
	logger     *zap.Logger
	level      zapcore.Level
	sync       *zap.SugaredLogger
	lumberjack *lumberjack.Logger
	mu         sync.RWMutex
	closed     bool
}

// newFileWriter 创建文件日志输出器
func newFileWriter(cfg *FileLogConfig) (*fileWriter, error) {
	if cfg == nil {
		return nil, fmt.Errorf("file log config is required")
	}

	level := zapcore.InfoLevel
	if cfg.Level != "" {
		level = LogLevelToZap(ParseLogLevel(cfg.Level))
	}

	// 如果路径为空，使用默认路径
	path := cfg.Path
	if path == "" {
		path = "./logs/app.log"
	}

	// 确保目录存在
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create log directory: %w", err)
	}

	// 配置日志轮转
	lumberjackLogger := &lumberjack.Logger{
		Filename:   path,
		MaxSize:    100,  // MB
		MaxAge:     30,   // days
		MaxBackups: 10,   // number of backups
		Compress:   true, // compress old files
	}

	// 应用自定义轮转配置
	if cfg.Rotation != nil {
		if cfg.Rotation.MaxSize > 0 {
			lumberjackLogger.MaxSize = cfg.Rotation.MaxSize
		}
		if cfg.Rotation.MaxAge > 0 {
			lumberjackLogger.MaxAge = cfg.Rotation.MaxAge
		}
		if cfg.Rotation.MaxBackups > 0 {
			lumberjackLogger.MaxBackups = cfg.Rotation.MaxBackups
		}
		lumberjackLogger.Compress = cfg.Rotation.Compress
	}

	// 创建文件编码器配置
	encoderConfig := zapcore.EncoderConfig{
		TimeKey:          "time",
		LevelKey:         "level",
		NameKey:          "logger",
		CallerKey:        "caller",
		MessageKey:       "msg",
		StacktraceKey:    "stacktrace",
		LineEnding:       zapcore.DefaultLineEnding,
		EncodeLevel:      zapcore.CapitalLevelEncoder,
		EncodeTime:       CustomTimeEncoder,
		EncodeDuration:   zapcore.StringDurationEncoder,
		EncodeCaller:     zapcore.ShortCallerEncoder,
		ConsoleSeparator: " | ",
	}

	// 创建文件编码器
	encoder := zapcore.NewConsoleEncoder(encoderConfig)

	// 创建核心（输出到文件）
	writeSyncer := zapcore.AddSync(lumberjackLogger)
	core := zapcore.NewCore(encoder, writeSyncer, level)

	// 创建 logger
	logger := zap.New(core, zap.AddCallerSkip(1), zap.AddStacktrace(zapcore.ErrorLevel))

	return &fileWriter{
		logger:     logger,
		level:      level,
		sync:       logger.Sugar(),
		lumberjack: lumberjackLogger,
		closed:     false,
	}, nil
}

// GetLogger 获取底层 zap.Logger
func (w *fileWriter) GetLogger() *zap.Logger {
	w.mu.RLock()
	defer w.mu.RUnlock()
	return w.logger
}

// ========== OTEL Core ==========

// otelCore 是自定义的 zapcore.Core，用于将日志输出到 OpenTelemetry
type otelCore struct {
	level           zapcore.Level
	telemetryMgr    telemetrymgr.ITelemetryManager
	telemetryLogger log.Logger
	fields          []zapcore.Field
	mu              sync.RWMutex
}

// newOTELCore 创建 OTEL 核心
func newOTELCore(level zapcore.Level, telemetryMgr telemetrymgr.ITelemetryManager) *otelCore {
	var telemetryLogger log.Logger
	if telemetryMgr != nil {
		telemetryLogger = telemetryMgr.Logger("loggermgr")
	}

	return &otelCore{
		level:           level,
		telemetryMgr:    telemetryMgr,
		telemetryLogger: telemetryLogger,
		fields:          make([]zapcore.Field, 0),
	}
}

// With 添加字段
func (c *otelCore) With(fields []zapcore.Field) zapcore.Core {
	c.mu.RLock()
	defer c.mu.RUnlock()

	newFields := make([]zapcore.Field, 0, len(c.fields)+len(fields))
	newFields = append(newFields, c.fields...)
	newFields = append(newFields, fields...)

	return &otelCore{
		level:           c.level,
		telemetryMgr:    c.telemetryMgr,
		telemetryLogger: c.telemetryLogger,
		fields:          newFields,
	}
}

// Check 检查日志级别是否启用
func (c *otelCore) Check(ent zapcore.Entry, ce *zapcore.CheckedEntry) *zapcore.CheckedEntry {
	if c.Enabled(ent.Level) {
		return ce.AddCore(ent, c)
	}
	return ce
}

// Write 写入日志
func (c *otelCore) Write(ent zapcore.Entry, fields []zapcore.Field) error {
	c.mu.RLock()
	defer c.mu.RUnlock()

	// 如果没有 TelemetryLogger，直接返回
	if c.telemetryLogger == nil {
		return nil
	}

	ctx := context.Background()

	// 构建日志记录
	record := log.Record{}
	record.SetTimestamp(ent.Time)
	record.SetSeverity(otelSeverityMap[ent.Level])
	record.SetSeverityText(otelSeverityTextMap[ent.Level])
	record.SetBody(log.StringValue(ent.Message))

	// 添加 With 添加的字段
	if len(c.fields) > 0 {
		attrs := make([]log.KeyValue, 0, len(c.fields))
		for _, field := range c.fields {
			if kv := fieldToKV(field); kv != nil {
				attrs = append(attrs, *kv)
			}
		}
		if len(attrs) > 0 {
			record.AddAttributes(attrs...)
		}
	}

	// 添加当前日志的字段
	if len(fields) > 0 {
		attrs := make([]log.KeyValue, 0, len(fields))
		for _, field := range fields {
			if kv := fieldToKV(field); kv != nil {
				attrs = append(attrs, *kv)
			}
		}
		if len(attrs) > 0 {
			record.AddAttributes(attrs...)
		}
	}

	c.telemetryLogger.Emit(ctx, record)
	return nil
}

// Sync 同步日志（OTEL 不需要同步）
func (c *otelCore) Sync() error {
	return nil
}

// Enabled 检查日志级别是否启用
func (c *otelCore) Enabled(level zapcore.Level) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return level >= c.level
}

// OTEL 日志级别映射
var otelSeverityMap = map[zapcore.Level]log.Severity{
	zapcore.DebugLevel:  log.SeverityTrace,
	zapcore.InfoLevel:   log.SeverityInfo,
	zapcore.WarnLevel:   log.SeverityWarn,
	zapcore.ErrorLevel:  log.SeverityError,
	zapcore.DPanicLevel: log.SeverityFatal1,
	zapcore.PanicLevel:  log.SeverityFatal2,
	zapcore.FatalLevel:  log.SeverityFatal3,
}

// OTEL 日志级别文本映射
var otelSeverityTextMap = map[zapcore.Level]string{
	zapcore.DebugLevel:  "TRACE",
	zapcore.InfoLevel:   "INFO",
	zapcore.WarnLevel:   "WARN",
	zapcore.ErrorLevel:  "ERROR",
	zapcore.DPanicLevel: "DPANIC",
	zapcore.PanicLevel:  "PANIC",
	zapcore.FatalLevel:  "FATAL",
}

// fieldToKV 将 zap 字段转换为 OTEL KeyValue
func fieldToKV(field zapcore.Field) *log.KeyValue {
	key := field.Key
	switch field.Type {
	case zapcore.StringType:
		return &log.KeyValue{Key: key, Value: log.StringValue(field.String)}
	case zapcore.Int64Type:
		return &log.KeyValue{Key: key, Value: log.Int64Value(field.Integer)}
	case zapcore.Int32Type:
		return &log.KeyValue{Key: key, Value: log.IntValue(int(field.Integer))}
	case zapcore.Float64Type:
		return &log.KeyValue{Key: key, Value: log.Float64Value(field.Interface.(float64))}
	case zapcore.BoolType:
		return &log.KeyValue{Key: key, Value: log.BoolValue(field.Integer == 1)}
	default:
		// 其他类型转换为字符串
		return &log.KeyValue{Key: key, Value: log.StringValue(field.String)}
	}
}
