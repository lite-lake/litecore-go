package loggermgr

import (
	"fmt"
	"time"

	"go.uber.org/zap/buffer"
	"go.uber.org/zap/zapcore"
)

const (
	colorReset  = "\033[0m"
	colorGray   = "\033[90m"
	colorGreen  = "\033[32m"
	colorYellow = "\033[33m"
	colorRed    = "\033[31m"
	colorBold   = "\033[1m"
)

var (
	bufPool = buffer.NewPool()
)

type ginConsoleEncoder struct {
	zapcore.EncoderConfig
	color      bool
	timeFormat string
}

func NewGinConsoleEncoder(cfg zapcore.EncoderConfig, color bool, timeFormat string) zapcore.Encoder {
	if timeFormat == "" {
		timeFormat = "2006-01-02 15:04:05.000"
	}
	return &ginConsoleEncoder{
		EncoderConfig: cfg,
		color:         color,
		timeFormat:    timeFormat,
	}
}

func (e *ginConsoleEncoder) Clone() zapcore.Encoder {
	return &ginConsoleEncoder{
		EncoderConfig: e.EncoderConfig,
		color:         e.color,
		timeFormat:    e.timeFormat,
	}
}

func (e *ginConsoleEncoder) EncodeEntry(entry zapcore.Entry, fields []zapcore.Field) (*buffer.Buffer, error) {
	buf := bufPool.Get()

	buf.WriteString(entry.Time.Format(e.timeFormat))
	buf.WriteString(" | ")

	levelStr := e.formatLevel(entry.Level)
	if e.color {
		buf.WriteString(e.levelColor(entry.Level))
		buf.WriteString(levelStr)
		buf.WriteString(colorReset)
	} else {
		buf.WriteString(levelStr)
	}

	buf.WriteString(" | ")
	buf.WriteString(entry.Message)

	for _, field := range fields {
		buf.WriteString(" ")
		buf.WriteString(field.Key)
		buf.WriteString("=")
		buf.WriteString(e.formatField(field))
	}

	buf.WriteString("\n")
	return buf, nil
}

func (e *ginConsoleEncoder) formatLevel(level zapcore.Level) string {
	switch level {
	case zapcore.DebugLevel:
		return "DEBUG"
	case zapcore.InfoLevel:
		return "INFO "
	case zapcore.WarnLevel:
		return "WARN "
	case zapcore.ErrorLevel:
		return "ERROR"
	case zapcore.FatalLevel, zapcore.PanicLevel, zapcore.DPanicLevel:
		return "FATAL"
	default:
		return level.String()
	}
}

func (e *ginConsoleEncoder) levelColor(level zapcore.Level) string {
	switch level {
	case zapcore.DebugLevel:
		return colorGray
	case zapcore.InfoLevel:
		return colorGreen
	case zapcore.WarnLevel:
		return colorYellow
	case zapcore.ErrorLevel:
		return colorRed
	case zapcore.FatalLevel, zapcore.PanicLevel, zapcore.DPanicLevel:
		return colorRed + colorBold
	default:
		return colorReset
	}
}

func (e *ginConsoleEncoder) formatField(field zapcore.Field) string {
	switch field.Type {
	case zapcore.StringType:
		return fmt.Sprintf("%q", field.String)
	case zapcore.Int64Type:
		return fmt.Sprintf("%d", field.Integer)
	case zapcore.Int32Type:
		return fmt.Sprintf("%d", int32(field.Integer))
	case zapcore.Int16Type:
		return fmt.Sprintf("%d", int16(field.Integer))
	case zapcore.Int8Type:
		return fmt.Sprintf("%d", int8(field.Integer))
	case zapcore.Uint64Type:
		return fmt.Sprintf("%d", uint64(field.Integer))
	case zapcore.Uint32Type:
		return fmt.Sprintf("%d", uint32(field.Integer))
	case zapcore.Uint16Type:
		return fmt.Sprintf("%d", uint16(field.Integer))
	case zapcore.Uint8Type:
		return fmt.Sprintf("%d", uint8(field.Integer))
	case zapcore.UintptrType:
		return fmt.Sprintf("%d", field.Integer)
	case zapcore.Float64Type:
		return fmt.Sprintf("%v", field.Interface)
	case zapcore.Float32Type:
		return fmt.Sprintf("%v", field.Interface)
	case zapcore.BoolType:
		return fmt.Sprintf("%t", field.Integer == 1)
	case zapcore.DurationType:
		return field.String
	case zapcore.TimeType:
		return field.String
	case zapcore.TimeFullType:
		return field.String
	case zapcore.ErrorType:
		return fmt.Sprintf("%q", field.Interface)
	case zapcore.Complex128Type:
		return fmt.Sprintf("%v", field.Interface)
	case zapcore.Complex64Type:
		return fmt.Sprintf("%v", field.Interface)
	case zapcore.ArrayMarshalerType:
		return fmt.Sprintf("%v", field.Interface)
	case zapcore.ObjectMarshalerType:
		return fmt.Sprintf("%v", field.Interface)
	case zapcore.ReflectType:
		return fmt.Sprintf("%v", field.Interface)
	case zapcore.SkipType:
		return ""
	case zapcore.NamespaceType:
		return ""
	case zapcore.StringerType:
		return fmt.Sprintf("%v", field.Interface)
	default:
		return fmt.Sprintf("%v", field.Interface)
	}
}

func (e *ginConsoleEncoder) ConsoleSeparator() string {
	return " | "
}

func (e *ginConsoleEncoder) AddArray(key string, marshaler zapcore.ArrayMarshaler) error {
	return nil
}

func (e *ginConsoleEncoder) AddObject(key string, marshaler zapcore.ObjectMarshaler) error {
	return nil
}

func (e *ginConsoleEncoder) AddBinary(key string, value []byte) {
}

func (e *ginConsoleEncoder) AddByteString(key string, value []byte) {
}

func (e *ginConsoleEncoder) AddBool(key string, value bool) {
}

func (e *ginConsoleEncoder) AddComplex128(key string, value complex128) {
}

func (e *ginConsoleEncoder) AddComplex64(key string, value complex64) {
}

func (e *ginConsoleEncoder) AddDuration(key string, value time.Duration) {
}

func (e *ginConsoleEncoder) AddFloat64(key string, value float64) {
}

func (e *ginConsoleEncoder) AddFloat32(key string, value float32) {
}

func (e *ginConsoleEncoder) AddInt(key string, value int) {
}

func (e *ginConsoleEncoder) AddInt64(key string, value int64) {
}

func (e *ginConsoleEncoder) AddInt32(key string, value int32) {
}

func (e *ginConsoleEncoder) AddInt16(key string, value int16) {
}

func (e *ginConsoleEncoder) AddInt8(key string, value int8) {
}

func (e *ginConsoleEncoder) AddString(key, value string) {
}

func (e *ginConsoleEncoder) AddTime(key string, value time.Time) {
}

func (e *ginConsoleEncoder) AddUint(key string, value uint) {
}

func (e *ginConsoleEncoder) AddUint64(key string, value uint64) {
}

func (e *ginConsoleEncoder) AddUint32(key string, value uint32) {
}

func (e *ginConsoleEncoder) AddUint16(key string, value uint16) {
}

func (e *ginConsoleEncoder) AddUint8(key string, value uint8) {
}

func (e *ginConsoleEncoder) AddUintptr(key string, value uintptr) {
}

func (e *ginConsoleEncoder) AddReflected(key string, value interface{}) error {
	return nil
}

func (e *ginConsoleEncoder) OpenNamespace(key string) {
}
