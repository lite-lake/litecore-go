package loggermgr

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/lite-lake/litecore-go/logger"
	"github.com/lite-lake/litecore-go/server/builtin/manager/telemetrymgr"
	"go.opentelemetry.io/otel/log"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

type driverZapLoggerManager struct {
	ins          logger.ILogger
	level        zapcore.Level
	telemetryMgr telemetrymgr.ITelemetryManager
	mu           sync.RWMutex
}

func NewDriverZapLoggerManager(cfg *DriverZapConfig, telemetryMgr telemetrymgr.ITelemetryManager) (ILoggerManager, error) {
	if cfg == nil {
		return nil, fmt.Errorf("DriverZapConfig cannot be nil")
	}

	if cfg.TelemetryEnabled && telemetryMgr == nil {
		return nil, fmt.Errorf("telemetry_manager is required when telemetry_enabled is true")
	}

	if !cfg.ConsoleEnabled && !cfg.FileEnabled && !cfg.TelemetryEnabled {
		return nil, fmt.Errorf("at least one output must be enabled (console, file or telemetry)")
	}

	var cores []zapcore.Core

	if cfg.TelemetryEnabled {
		otelLevel := zapcore.InfoLevel
		if cfg.TelemetryConfig != nil && cfg.TelemetryConfig.Level != "" {
			otelLevel = parseLogLevel(cfg.TelemetryConfig.Level)
		}
		cores = append(cores, buildOTELCore(otelLevel, telemetryMgr))
	}

	if cfg.ConsoleEnabled {
		consoleCore, err := buildConsoleCore(cfg.ConsoleConfig)
		if err != nil {
			return nil, fmt.Errorf("failed to build console core: %w", err)
		}
		cores = append(cores, consoleCore)
	}

	if cfg.FileEnabled {
		if cfg.FileConfig == nil {
			return nil, fmt.Errorf("file_config is required when file logging is enabled")
		}
		fileCore, err := buildFileCore(cfg.FileConfig)
		if err != nil {
			return nil, fmt.Errorf("failed to build file core: %w", err)
		}
		cores = append(cores, fileCore)
	}

	core := zapcore.NewTee(cores...)

	zapLoggerInstance := zap.New(core,
		zap.AddCallerSkip(1),
		zap.AddStacktrace(zapcore.ErrorLevel),
		zap.Fields(zap.String("logger", "zap")),
	)

	minLevel := zapcore.InfoLevel
	if cfg.ConsoleEnabled && cfg.ConsoleConfig != nil {
		level := parseLogLevel(cfg.ConsoleConfig.Level)
		if level < minLevel {
			minLevel = level
		}
	}
	if cfg.FileEnabled && cfg.FileConfig != nil {
		level := parseLogLevel(cfg.FileConfig.Level)
		if level < minLevel {
			minLevel = level
		}
	}

	return &driverZapLoggerManager{
		ins:   &zapLoggerImpl{logger: zapLoggerInstance, level: minLevel},
		level: minLevel,
	}, nil
}

func (d *driverZapLoggerManager) ManagerName() string {
	return "LoggerZapManager"
}

func (d *driverZapLoggerManager) Health() error {
	return nil
}

func (d *driverZapLoggerManager) OnStart() error {
	return nil
}

func (d *driverZapLoggerManager) OnStop() error {
	if zl, ok := d.ins.(*zapLoggerImpl); ok {
		_ = zl.sync()
	}
	return nil
}

func (d *driverZapLoggerManager) Ins() logger.ILogger {
	return d.ins
}

type zapLoggerImpl struct {
	logger *zap.Logger
	level  zapcore.Level
	mu     sync.RWMutex
}

func (l *zapLoggerImpl) Debug(msg string, args ...any) {
	l.mu.RLock()
	defer l.mu.RUnlock()

	if zapcore.DebugLevel >= l.level {
		fields := argsToFields(args...)
		l.logger.Debug(msg, fields...)
	}
}

func (l *zapLoggerImpl) Info(msg string, args ...any) {
	l.mu.RLock()
	defer l.mu.RUnlock()

	if zapcore.InfoLevel >= l.level {
		fields := argsToFields(args...)
		l.logger.Info(msg, fields...)
	}
}

func (l *zapLoggerImpl) Warn(msg string, args ...any) {
	l.mu.RLock()
	defer l.mu.RUnlock()

	if zapcore.WarnLevel >= l.level {
		fields := argsToFields(args...)
		l.logger.Warn(msg, fields...)
	}
}

func (l *zapLoggerImpl) Error(msg string, args ...any) {
	l.mu.RLock()
	defer l.mu.RUnlock()

	if zapcore.ErrorLevel >= l.level {
		fields := argsToFields(args...)
		l.logger.Error(msg, fields...)
	}
}

func (l *zapLoggerImpl) Fatal(msg string, args ...any) {
	l.mu.RLock()
	defer l.mu.RUnlock()

	if zapcore.FatalLevel >= l.level {
		fields := argsToFields(args...)
		l.logger.Fatal(msg, fields...)
	}
	os.Exit(1)
}

func (l *zapLoggerImpl) With(args ...any) logger.ILogger {
	l.mu.RLock()
	defer l.mu.RUnlock()

	fields := argsToFields(args...)
	newLogger := l.logger.With(fields...)

	return &zapLoggerImpl{
		logger: newLogger,
		level:  l.level,
	}
}

func (l *zapLoggerImpl) SetLevel(level logger.LogLevel) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.level = logger.LogLevelToZap(level)
}

func (l *zapLoggerImpl) sync() error {
	return l.logger.Sync()
}

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

var (
	supportsColor bool
	colorChecked  bool
)

func buildConsoleCore(cfg *LogLevelConfig) (zapcore.Core, error) {
	level := parseLogLevel(cfg.Level)

	if !colorChecked {
		supportsColor = detectColorSupport()
		colorChecked = true
	}

	encoderConfig := zapcore.EncoderConfig{
		TimeKey:          "time",
		LevelKey:         "level",
		NameKey:          "logger",
		CallerKey:        "caller",
		MessageKey:       "msg",
		StacktraceKey:    "stacktrace",
		LineEnding:       zapcore.DefaultLineEnding,
		EncodeLevel:      customLevelEncoder,
		EncodeTime:       customTimeEncoder,
		EncodeDuration:   zapcore.StringDurationEncoder,
		EncodeCaller:     zapcore.ShortCallerEncoder,
		ConsoleSeparator: " | ",
	}

	encoder := zapcore.NewConsoleEncoder(encoderConfig)

	stdoutCore := zapcore.NewCore(encoder, zapcore.AddSync(os.Stdout), level)
	stderrCore := zapcore.NewCore(encoder, zapcore.AddSync(os.Stderr), zapcore.ErrorLevel)

	return zapcore.NewTee(stdoutCore, stderrCore), nil
}

func buildFileCore(cfg *FileLogConfig) (zapcore.Core, error) {
	if cfg == nil {
		return nil, fmt.Errorf("file config is required")
	}

	level := parseLogLevel(cfg.Level)

	path := cfg.Path
	if path == "" {
		path = "./logs/app.log"
	}

	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create log directory: %w", err)
	}

	lumberjackLogger := &lumberjack.Logger{
		Filename:   path,
		MaxSize:    100,
		MaxAge:     30,
		MaxBackups: 10,
		Compress:   true,
	}

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

	encoderConfig := zapcore.EncoderConfig{
		TimeKey:          "time",
		LevelKey:         "level",
		NameKey:          "logger",
		CallerKey:        "caller",
		MessageKey:       "msg",
		StacktraceKey:    "stacktrace",
		LineEnding:       zapcore.DefaultLineEnding,
		EncodeLevel:      zapcore.CapitalLevelEncoder,
		EncodeTime:       customTimeEncoder,
		EncodeDuration:   zapcore.StringDurationEncoder,
		EncodeCaller:     zapcore.ShortCallerEncoder,
		ConsoleSeparator: " | ",
	}

	encoder := zapcore.NewConsoleEncoder(encoderConfig)

	return zapcore.NewCore(encoder, zapcore.AddSync(lumberjackLogger), level), nil
}

func detectColorSupport() bool {
	if os.Getenv("NO_COLOR") != "" {
		return false
	}

	term := os.Getenv("TERM")
	if term == "" || term == "dumb" {
		return false
	}

	if os.Getenv("CI") != "" {
		return false
	}

	if fileInfo, _ := os.Stdout.Stat(); (fileInfo.Mode() & os.ModeCharDevice) == 0 {
		return false
	}

	return true
}

func customLevelEncoder(level zapcore.Level, enc zapcore.PrimitiveArrayEncoder) {
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

func customTimeEncoder(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
	enc.AppendString(t.Format("2006-01-02 15:04:05.000"))
}

func parseLogLevel(level string) zapcore.Level {
	if level == "" {
		return zapcore.InfoLevel
	}
	return logger.LogLevelToZap(logger.ParseLogLevel(level))
}

type otelCore struct {
	level           zapcore.Level
	telemetryMgr    telemetrymgr.ITelemetryManager
	telemetryLogger log.Logger
	fields          []zapcore.Field
	mu              sync.RWMutex
}

func buildOTELCore(level zapcore.Level, telemetryMgr telemetrymgr.ITelemetryManager) zapcore.Core {
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

func (c *otelCore) Check(ent zapcore.Entry, ce *zapcore.CheckedEntry) *zapcore.CheckedEntry {
	if c.Enabled(ent.Level) {
		return ce.AddCore(ent, c)
	}
	return ce
}

func (c *otelCore) Write(ent zapcore.Entry, fields []zapcore.Field) error {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.telemetryLogger == nil {
		return nil
	}

	ctx := context.Background()

	record := log.Record{}
	record.SetTimestamp(ent.Time)
	record.SetSeverity(otelSeverityMap[ent.Level])
	record.SetSeverityText(otelSeverityTextMap[ent.Level])
	record.SetBody(log.StringValue(ent.Message))

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

func (c *otelCore) Sync() error {
	return nil
}

func (c *otelCore) Enabled(level zapcore.Level) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return level >= c.level
}

var otelSeverityMap = map[zapcore.Level]log.Severity{
	zapcore.DebugLevel:  log.SeverityTrace,
	zapcore.InfoLevel:   log.SeverityInfo,
	zapcore.WarnLevel:   log.SeverityWarn,
	zapcore.ErrorLevel:  log.SeverityError,
	zapcore.DPanicLevel: log.SeverityFatal1,
	zapcore.PanicLevel:  log.SeverityFatal2,
	zapcore.FatalLevel:  log.SeverityFatal3,
}

var otelSeverityTextMap = map[zapcore.Level]string{
	zapcore.DebugLevel:  "TRACE",
	zapcore.InfoLevel:   "INFO",
	zapcore.WarnLevel:   "WARN",
	zapcore.ErrorLevel:  "ERROR",
	zapcore.DPanicLevel: "DPANIC",
	zapcore.PanicLevel:  "PANIC",
	zapcore.FatalLevel:  "FATAL",
}

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
		if val, ok := field.Interface.(float64); ok {
			return &log.KeyValue{Key: key, Value: log.Float64Value(val)}
		}
		return &log.KeyValue{Key: key, Value: log.StringValue(fmt.Sprint(field.Interface))}
	case zapcore.BoolType:
		return &log.KeyValue{Key: key, Value: log.BoolValue(field.Integer == 1)}
	default:
		return &log.KeyValue{Key: key, Value: log.StringValue(field.String)}
	}
}

var _ ILoggerManager = (*driverZapLoggerManager)(nil)
var _ logger.ILogger = (*zapLoggerImpl)(nil)
