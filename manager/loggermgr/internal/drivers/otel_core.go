package drivers

import (
	"context"
	"sync"

	"com.litelake.litecore/manager/loggermgr/internal/loglevel"
	"com.litelake.litecore/manager/telemetrymgr"
	"go.opentelemetry.io/otel/log"
	"go.uber.org/zap/zapcore"
)

// OTELCore 是自定义的 zapcore.Core，用于将日志输出到 OpenTelemetry
type OTELCore struct {
	level           zapcore.Level
	telemetryMgr    telemetrymgr.TelemetryManager
	telemetryLogger log.Logger
	fields          []zapcore.Field
	mu              sync.RWMutex
}

// NewOTELCore 创建 OTEL 核心
func NewOTELCore(level zapcore.Level, telemetryMgr telemetrymgr.TelemetryManager) *OTELCore {
	var telemetryLogger log.Logger
	if telemetryMgr != nil {
		telemetryLogger = telemetryMgr.Logger("loggermgr")
	}

	return &OTELCore{
		level:           level,
		telemetryMgr:    telemetryMgr,
		telemetryLogger: telemetryLogger,
		fields:          make([]zapcore.Field, 0),
	}
}

// With 添加字段
func (c *OTELCore) With(fields []zapcore.Field) zapcore.Core {
	c.mu.RLock()
	defer c.mu.RUnlock()

	newFields := make([]zapcore.Field, 0, len(c.fields)+len(fields))
	newFields = append(newFields, c.fields...)
	newFields = append(newFields, fields...)

	return &OTELCore{
		level:           c.level,
		telemetryMgr:    c.telemetryMgr,
		telemetryLogger: c.telemetryLogger,
		fields:          newFields,
	}
}

// Check 检查日志级别是否启用
func (c *OTELCore) Check(ent zapcore.Entry, ce *zapcore.CheckedEntry) *zapcore.CheckedEntry {
	if c.Enabled(ent.Level) {
		return ce.AddCore(ent, c)
	}
	return ce
}

// Write 写入日志
func (c *OTELCore) Write(ent zapcore.Entry, fields []zapcore.Field) error {
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
func (c *OTELCore) Sync() error {
	return nil
}

// Enabled 检查日志级别是否启用
func (c *OTELCore) Enabled(level zapcore.Level) bool {
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

// ZapToLogLevel 将 zapcore.Level 转换为 LogLevel
func ZapToLogLevel(level zapcore.Level) loglevel.LogLevel {
	switch level {
	case zapcore.DebugLevel:
		return loglevel.DebugLevel
	case zapcore.InfoLevel:
		return loglevel.InfoLevel
	case zapcore.WarnLevel:
		return loglevel.WarnLevel
	case zapcore.ErrorLevel:
		return loglevel.ErrorLevel
	case zapcore.FatalLevel:
		return loglevel.FatalLevel
	default:
		return loglevel.InfoLevel
	}
}

// LogLevelToZap 将 LogLevel 转换为 zapcore.Level
func LogLevelToZap(level loglevel.LogLevel) zapcore.Level {
	switch level {
	case loglevel.DebugLevel:
		return zapcore.DebugLevel
	case loglevel.InfoLevel:
		return zapcore.InfoLevel
	case loglevel.WarnLevel:
		return zapcore.WarnLevel
	case loglevel.ErrorLevel:
		return zapcore.ErrorLevel
	case loglevel.FatalLevel:
		return zapcore.FatalLevel
	default:
		return zapcore.InfoLevel
	}
}

// ensure OTELCore implements zapcore.Core interface
var _ zapcore.Core = (*OTELCore)(nil)
