package loggermgr

import (
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func TestNewGinConsoleEncoder(t *testing.T) {
	t.Run("有效配置创建编码器", func(t *testing.T) {
		cfg := zapcore.EncoderConfig{
			MessageKey:   "msg",
			LevelKey:     "level",
			TimeKey:      "time",
			EncodeLevel:  zapcore.CapitalLevelEncoder,
			EncodeTime:   zapcore.ISO8601TimeEncoder,
			EncodeCaller: zapcore.ShortCallerEncoder,
		}

		encoder := NewGinConsoleEncoder(cfg, true, "2006-01-02 15:04:05")
		assert.NotNil(t, encoder)

		ginEncoder, ok := encoder.(*ginConsoleEncoder)
		assert.True(t, ok)
		assert.True(t, ginEncoder.color)
		assert.Equal(t, "2006-01-02 15:04:05", ginEncoder.timeFormat)
	})

	t.Run("默认时间格式", func(t *testing.T) {
		cfg := zapcore.EncoderConfig{}
		encoder := NewGinConsoleEncoder(cfg, false, "")
		assert.NotNil(t, encoder)

		ginEncoder, ok := encoder.(*ginConsoleEncoder)
		assert.True(t, ok)
		assert.Equal(t, "2006-01-02 15:04:05.000", ginEncoder.timeFormat)
	})

	t.Run("禁用颜色", func(t *testing.T) {
		cfg := zapcore.EncoderConfig{}
		encoder := NewGinConsoleEncoder(cfg, false, "")
		assert.NotNil(t, encoder)

		ginEncoder, ok := encoder.(*ginConsoleEncoder)
		assert.True(t, ok)
		assert.False(t, ginEncoder.color)
	})
}

func TestGinConsoleEncoder_Clone(t *testing.T) {
	t.Run("克隆编码器", func(t *testing.T) {
		cfg := zapcore.EncoderConfig{
			MessageKey: "msg",
		}
		encoder := NewGinConsoleEncoder(cfg, true, "2006-01-02 15:04:05")

		cloned := encoder.Clone()
		assert.NotNil(t, cloned)

		ginEncoder, ok := encoder.(*ginConsoleEncoder)
		assert.True(t, ok)

		clonedGinEncoder, ok := cloned.(*ginConsoleEncoder)
		assert.True(t, ok)

		assert.Equal(t, ginEncoder.EncoderConfig, clonedGinEncoder.EncoderConfig)
		assert.Equal(t, ginEncoder.color, clonedGinEncoder.color)
		assert.Equal(t, ginEncoder.timeFormat, clonedGinEncoder.timeFormat)

		assert.NotSame(t, encoder, cloned)
	})
}

func TestGinConsoleEncoder_EncodeEntry(t *testing.T) {
	t.Run("DEBUG级别日志", func(t *testing.T) {
		cfg := zapcore.EncoderConfig{}
		encoder := NewGinConsoleEncoder(cfg, false, "2006-01-02 15:04:05.000")

		entry := zapcore.Entry{
			Time:    time.Date(2024, 1, 15, 10, 30, 45, 123, time.UTC),
			Level:   zapcore.DebugLevel,
			Message: "debug message",
		}

		buf, err := encoder.EncodeEntry(entry, nil)
		assert.NoError(t, err)
		assert.NotNil(t, buf)

		output := buf.String()
		assert.Contains(t, output, "2024-01-15 10:30:45.000")
		assert.Contains(t, output, "DEBUG")
		assert.Contains(t, output, "debug message")
	})

	t.Run("INFO级别日志", func(t *testing.T) {
		cfg := zapcore.EncoderConfig{}
		encoder := NewGinConsoleEncoder(cfg, false, "2006-01-02 15:04:05.000")

		entry := zapcore.Entry{
			Time:    time.Date(2024, 1, 15, 10, 30, 45, 123, time.UTC),
			Level:   zapcore.InfoLevel,
			Message: "info message",
		}

		buf, err := encoder.EncodeEntry(entry, nil)
		assert.NoError(t, err)
		assert.NotNil(t, buf)

		output := buf.String()
		assert.Contains(t, output, "INFO ")
		assert.Contains(t, output, "info message")
	})

	t.Run("WARN级别日志", func(t *testing.T) {
		cfg := zapcore.EncoderConfig{}
		encoder := NewGinConsoleEncoder(cfg, false, "2006-01-02 15:04:05.000")

		entry := zapcore.Entry{
			Time:    time.Date(2024, 1, 15, 10, 30, 45, 123, time.UTC),
			Level:   zapcore.WarnLevel,
			Message: "warn message",
		}

		buf, err := encoder.EncodeEntry(entry, nil)
		assert.NoError(t, err)
		assert.NotNil(t, buf)

		output := buf.String()
		assert.Contains(t, output, "WARN ")
		assert.Contains(t, output, "warn message")
	})

	t.Run("ERROR级别日志", func(t *testing.T) {
		cfg := zapcore.EncoderConfig{}
		encoder := NewGinConsoleEncoder(cfg, false, "2006-01-02 15:04:05.000")

		entry := zapcore.Entry{
			Time:    time.Date(2024, 1, 15, 10, 30, 45, 123, time.UTC),
			Level:   zapcore.ErrorLevel,
			Message: "error message",
		}

		buf, err := encoder.EncodeEntry(entry, nil)
		assert.NoError(t, err)
		assert.NotNil(t, buf)

		output := buf.String()
		assert.Contains(t, output, "ERROR")
		assert.Contains(t, output, "error message")
	})

	t.Run("FATAL级别日志", func(t *testing.T) {
		cfg := zapcore.EncoderConfig{}
		encoder := NewGinConsoleEncoder(cfg, false, "2006-01-02 15:04:05.000")

		entry := zapcore.Entry{
			Time:    time.Date(2024, 1, 15, 10, 30, 45, 123, time.UTC),
			Level:   zapcore.FatalLevel,
			Message: "fatal message",
		}

		buf, err := encoder.EncodeEntry(entry, nil)
		assert.NoError(t, err)
		assert.NotNil(t, buf)

		output := buf.String()
		assert.Contains(t, output, "FATAL")
		assert.Contains(t, output, "fatal message")
	})

	t.Run("时间格式验证", func(t *testing.T) {
		cfg := zapcore.EncoderConfig{}
		encoder := NewGinConsoleEncoder(cfg, false, "2006-01-02 15:04:05.000")

		entry := zapcore.Entry{
			Time:    time.Date(2024, 1, 15, 10, 30, 45, 123, time.UTC),
			Level:   zapcore.InfoLevel,
			Message: "test",
		}

		buf, err := encoder.EncodeEntry(entry, nil)
		assert.NoError(t, err)

		output := buf.String()
		assert.Contains(t, output, "2024-01-15 10:30:45.000 | ")
		assert.Equal(t, 23, len("2024-01-15 10:30:45.000"))
	})

	t.Run("分隔符验证", func(t *testing.T) {
		cfg := zapcore.EncoderConfig{}
		encoder := NewGinConsoleEncoder(cfg, false, "2006-01-02 15:04:05.000")

		entry := zapcore.Entry{
			Time:    time.Date(2024, 1, 15, 10, 30, 45, 123, time.UTC),
			Level:   zapcore.InfoLevel,
			Message: "test message",
		}

		buf, err := encoder.EncodeEntry(entry, nil)
		assert.NoError(t, err)

		output := buf.String()
		parts := strings.Split(output, " | ")
		assert.Equal(t, 3, len(parts))
	})
}

func TestGinConsoleEncoder_FieldFormatting(t *testing.T) {
	t.Run("字符串字段", func(t *testing.T) {
		cfg := zapcore.EncoderConfig{}
		encoder := NewGinConsoleEncoder(cfg, false, "2006-01-02 15:04:05.000")

		entry := zapcore.Entry{
			Time:    time.Date(2024, 1, 15, 10, 30, 45, 123, time.UTC),
			Level:   zapcore.InfoLevel,
			Message: "test",
		}

		fields := []zapcore.Field{
			zap.String("name", "test value"),
		}

		buf, err := encoder.EncodeEntry(entry, fields)
		assert.NoError(t, err)

		output := buf.String()
		assert.Contains(t, output, `name="test value"`)
	})

	t.Run("整数字段", func(t *testing.T) {
		cfg := zapcore.EncoderConfig{}
		encoder := NewGinConsoleEncoder(cfg, false, "2006-01-02 15:04:05.000")

		entry := zapcore.Entry{
			Time:    time.Date(2024, 1, 15, 10, 30, 45, 123, time.UTC),
			Level:   zapcore.InfoLevel,
			Message: "test",
		}

		fields := []zapcore.Field{
			zap.Int("count", 42),
		}

		buf, err := encoder.EncodeEntry(entry, fields)
		assert.NoError(t, err)

		output := buf.String()
		assert.Contains(t, output, "count=42")
	})

	t.Run("Int64字段", func(t *testing.T) {
		cfg := zapcore.EncoderConfig{}
		encoder := NewGinConsoleEncoder(cfg, false, "2006-01-02 15:04:05.000")

		entry := zapcore.Entry{
			Time:    time.Date(2024, 1, 15, 10, 30, 45, 123, time.UTC),
			Level:   zapcore.InfoLevel,
			Message: "test",
		}

		fields := []zapcore.Field{
			zap.Int64("timestamp", 123456789012345),
		}

		buf, err := encoder.EncodeEntry(entry, fields)
		assert.NoError(t, err)

		output := buf.String()
		assert.Contains(t, output, "timestamp=123456789012345")
	})

	t.Run("Uint64字段", func(t *testing.T) {
		cfg := zapcore.EncoderConfig{}
		encoder := NewGinConsoleEncoder(cfg, false, "2006-01-02 15:04:05.000")

		entry := zapcore.Entry{
			Time:    time.Date(2024, 1, 15, 10, 30, 45, 123, time.UTC),
			Level:   zapcore.InfoLevel,
			Message: "test",
		}

		fields := []zapcore.Field{
			zap.Uint64("id", 123456789),
		}

		buf, err := encoder.EncodeEntry(entry, fields)
		assert.NoError(t, err)

		output := buf.String()
		assert.Contains(t, output, "id=123456789")
	})

	t.Run("Float64字段", func(t *testing.T) {
		cfg := zapcore.EncoderConfig{}
		encoder := NewGinConsoleEncoder(cfg, false, "2006-01-02 15:04:05.000")

		entry := zapcore.Entry{
			Time:    time.Date(2024, 1, 15, 10, 30, 45, 123, time.UTC),
			Level:   zapcore.InfoLevel,
			Message: "test",
		}

		fields := []zapcore.Field{
			zap.Float64("price", 123.456),
		}

		buf, err := encoder.EncodeEntry(entry, fields)
		assert.NoError(t, err)

		output := buf.String()
		assert.Contains(t, output, "price=")
	})

	t.Run("Float32字段", func(t *testing.T) {
		cfg := zapcore.EncoderConfig{}
		encoder := NewGinConsoleEncoder(cfg, false, "2006-01-02 15:04:05.000")

		entry := zapcore.Entry{
			Time:    time.Date(2024, 1, 15, 10, 30, 45, 123, time.UTC),
			Level:   zapcore.InfoLevel,
			Message: "test",
		}

		fields := []zapcore.Field{
			zap.Float32("ratio", 0.75),
		}

		buf, err := encoder.EncodeEntry(entry, fields)
		assert.NoError(t, err)

		output := buf.String()
		assert.Contains(t, output, "ratio=")
	})

	t.Run("布尔字段true", func(t *testing.T) {
		cfg := zapcore.EncoderConfig{}
		encoder := NewGinConsoleEncoder(cfg, false, "2006-01-02 15:04:05.000")

		entry := zapcore.Entry{
			Time:    time.Date(2024, 1, 15, 10, 30, 45, 123, time.UTC),
			Level:   zapcore.InfoLevel,
			Message: "test",
		}

		fields := []zapcore.Field{
			zap.Bool("enabled", true),
		}

		buf, err := encoder.EncodeEntry(entry, fields)
		assert.NoError(t, err)

		output := buf.String()
		assert.Contains(t, output, "enabled=true")
	})

	t.Run("布尔字段false", func(t *testing.T) {
		cfg := zapcore.EncoderConfig{}
		encoder := NewGinConsoleEncoder(cfg, false, "2006-01-02 15:04:05.000")

		entry := zapcore.Entry{
			Time:    time.Date(2024, 1, 15, 10, 30, 45, 123, time.UTC),
			Level:   zapcore.InfoLevel,
			Message: "test",
		}

		fields := []zapcore.Field{
			zap.Bool("enabled", false),
		}

		buf, err := encoder.EncodeEntry(entry, fields)
		assert.NoError(t, err)

		output := buf.String()
		assert.Contains(t, output, "enabled=false")
	})

	t.Run("多个字段", func(t *testing.T) {
		cfg := zapcore.EncoderConfig{}
		encoder := NewGinConsoleEncoder(cfg, false, "2006-01-02 15:04:05.000")

		entry := zapcore.Entry{
			Time:    time.Date(2024, 1, 15, 10, 30, 45, 123, time.UTC),
			Level:   zapcore.InfoLevel,
			Message: "test",
		}

		fields := []zapcore.Field{
			zap.String("name", "test"),
			zap.Int("age", 25),
			zap.Bool("active", true),
		}

		buf, err := encoder.EncodeEntry(entry, fields)
		assert.NoError(t, err)

		output := buf.String()
		assert.Contains(t, output, `name="test"`)
		assert.Contains(t, output, "age=25")
		assert.Contains(t, output, "active=true")
	})
}

func TestGinConsoleEncoder_ColorSupport(t *testing.T) {
	t.Run("启用颜色DEBUG", func(t *testing.T) {
		cfg := zapcore.EncoderConfig{}
		encoder := NewGinConsoleEncoder(cfg, true, "2006-01-02 15:04:05.000")

		entry := zapcore.Entry{
			Time:    time.Date(2024, 1, 15, 10, 30, 45, 123, time.UTC),
			Level:   zapcore.DebugLevel,
			Message: "test",
		}

		buf, err := encoder.EncodeEntry(entry, nil)
		assert.NoError(t, err)

		output := buf.String()
		assert.Contains(t, output, "\033[90m")
		assert.Contains(t, output, "\033[0m")
	})

	t.Run("启用颜色INFO", func(t *testing.T) {
		cfg := zapcore.EncoderConfig{}
		encoder := NewGinConsoleEncoder(cfg, true, "2006-01-02 15:04:05.000")

		entry := zapcore.Entry{
			Time:    time.Date(2024, 1, 15, 10, 30, 45, 123, time.UTC),
			Level:   zapcore.InfoLevel,
			Message: "test",
		}

		buf, err := encoder.EncodeEntry(entry, nil)
		assert.NoError(t, err)

		output := buf.String()
		assert.Contains(t, output, "\033[32m")
		assert.Contains(t, output, "\033[0m")
	})

	t.Run("启用颜色WARN", func(t *testing.T) {
		cfg := zapcore.EncoderConfig{}
		encoder := NewGinConsoleEncoder(cfg, true, "2006-01-02 15:04:05.000")

		entry := zapcore.Entry{
			Time:    time.Date(2024, 1, 15, 10, 30, 45, 123, time.UTC),
			Level:   zapcore.WarnLevel,
			Message: "test",
		}

		buf, err := encoder.EncodeEntry(entry, nil)
		assert.NoError(t, err)

		output := buf.String()
		assert.Contains(t, output, "\033[33m")
		assert.Contains(t, output, "\033[0m")
	})

	t.Run("启用颜色ERROR", func(t *testing.T) {
		cfg := zapcore.EncoderConfig{}
		encoder := NewGinConsoleEncoder(cfg, true, "2006-01-02 15:04:05.000")

		entry := zapcore.Entry{
			Time:    time.Date(2024, 1, 15, 10, 30, 45, 123, time.UTC),
			Level:   zapcore.ErrorLevel,
			Message: "test",
		}

		buf, err := encoder.EncodeEntry(entry, nil)
		assert.NoError(t, err)

		output := buf.String()
		assert.Contains(t, output, "\033[31m")
		assert.Contains(t, output, "\033[0m")
	})

	t.Run("启用颜色FATAL", func(t *testing.T) {
		cfg := zapcore.EncoderConfig{}
		encoder := NewGinConsoleEncoder(cfg, true, "2006-01-02 15:04:05.000")

		entry := zapcore.Entry{
			Time:    time.Date(2024, 1, 15, 10, 30, 45, 123, time.UTC),
			Level:   zapcore.FatalLevel,
			Message: "test",
		}

		buf, err := encoder.EncodeEntry(entry, nil)
		assert.NoError(t, err)

		output := buf.String()
		assert.Contains(t, output, "\033[31m\033[1m")
		assert.Contains(t, output, "\033[0m")
	})

	t.Run("禁用颜色", func(t *testing.T) {
		cfg := zapcore.EncoderConfig{}
		encoder := NewGinConsoleEncoder(cfg, false, "2006-01-02 15:04:05.000")

		entry := zapcore.Entry{
			Time:    time.Date(2024, 1, 15, 10, 30, 45, 123, time.UTC),
			Level:   zapcore.InfoLevel,
			Message: "test",
		}

		buf, err := encoder.EncodeEntry(entry, nil)
		assert.NoError(t, err)

		output := buf.String()
		assert.NotContains(t, output, "\033[")
	})
}

func TestGinConsoleEncoder_EdgeCases(t *testing.T) {
	t.Run("空消息", func(t *testing.T) {
		cfg := zapcore.EncoderConfig{}
		encoder := NewGinConsoleEncoder(cfg, false, "2006-01-02 15:04:05.000")

		entry := zapcore.Entry{
			Time:    time.Date(2024, 1, 15, 10, 30, 45, 123, time.UTC),
			Level:   zapcore.InfoLevel,
			Message: "",
		}

		buf, err := encoder.EncodeEntry(entry, nil)
		assert.NoError(t, err)
		assert.NotNil(t, buf)

		output := buf.String()
		assert.Contains(t, output, "2024-01-15 10:30:45.000 | INFO  | ")
	})

	t.Run("空字段列表", func(t *testing.T) {
		cfg := zapcore.EncoderConfig{}
		encoder := NewGinConsoleEncoder(cfg, false, "2006-01-02 15:04:05.000")

		entry := zapcore.Entry{
			Time:    time.Date(2024, 1, 15, 10, 30, 45, 123, time.UTC),
			Level:   zapcore.InfoLevel,
			Message: "test message",
		}

		buf, err := encoder.EncodeEntry(entry, []zapcore.Field{})
		assert.NoError(t, err)
		assert.NotNil(t, buf)

		output := buf.String()
		assert.Contains(t, output, "test message")
		assert.NotContains(t, output, "=")
	})

	t.Run("包含特殊字符的消息", func(t *testing.T) {
		cfg := zapcore.EncoderConfig{}
		encoder := NewGinConsoleEncoder(cfg, false, "2006-01-02 15:04:05.000")

		entry := zapcore.Entry{
			Time:    time.Date(2024, 1, 15, 10, 30, 45, 123, time.UTC),
			Level:   zapcore.InfoLevel,
			Message: "message with \n and \t special chars",
		}

		buf, err := encoder.EncodeEntry(entry, nil)
		assert.NoError(t, err)
		assert.NotNil(t, buf)

		output := buf.String()
		assert.Contains(t, output, "message with \n and \t special chars")
	})

	t.Run("包含特殊字符的字段", func(t *testing.T) {
		cfg := zapcore.EncoderConfig{}
		encoder := NewGinConsoleEncoder(cfg, false, "2006-01-02 15:04:05.000")

		entry := zapcore.Entry{
			Time:    time.Date(2024, 1, 15, 10, 30, 45, 123, time.UTC),
			Level:   zapcore.InfoLevel,
			Message: "test",
		}

		fields := []zapcore.Field{
			zap.String("special", "value with \"quotes\" and spaces"),
		}

		buf, err := encoder.EncodeEntry(entry, fields)
		assert.NoError(t, err)

		output := buf.String()
		assert.Contains(t, output, `special="value with \"quotes\" and spaces"`)
	})

	t.Run("Unicode字符", func(t *testing.T) {
		cfg := zapcore.EncoderConfig{}
		encoder := NewGinConsoleEncoder(cfg, false, "2006-01-02 15:04:05.000")

		entry := zapcore.Entry{
			Time:    time.Date(2024, 1, 15, 10, 30, 45, 123, time.UTC),
			Level:   zapcore.InfoLevel,
			Message: "测试消息 🎉",
		}

		fields := []zapcore.Field{
			zap.String("name", "张三"),
		}

		buf, err := encoder.EncodeEntry(entry, fields)
		assert.NoError(t, err)

		output := buf.String()
		assert.Contains(t, output, "测试消息 🎉")
		assert.Contains(t, output, `name="张三"`)
	})

	t.Run("Panic级别", func(t *testing.T) {
		cfg := zapcore.EncoderConfig{}
		encoder := NewGinConsoleEncoder(cfg, false, "2006-01-02 15:04:05.000")

		entry := zapcore.Entry{
			Time:    time.Date(2024, 1, 15, 10, 30, 45, 123, time.UTC),
			Level:   zapcore.PanicLevel,
			Message: "panic message",
		}

		buf, err := encoder.EncodeEntry(entry, nil)
		assert.NoError(t, err)

		output := buf.String()
		assert.Contains(t, output, "FATAL")
	})

	t.Run("DPanic级别", func(t *testing.T) {
		cfg := zapcore.EncoderConfig{}
		encoder := NewGinConsoleEncoder(cfg, false, "2006-01-02 15:04:05.000")

		entry := zapcore.Entry{
			Time:    time.Date(2024, 1, 15, 10, 30, 45, 123, time.UTC),
			Level:   zapcore.DPanicLevel,
			Message: "dpanic message",
		}

		buf, err := encoder.EncodeEntry(entry, nil)
		assert.NoError(t, err)

		output := buf.String()
		assert.Contains(t, output, "FATAL")
	})

	t.Run("长消息", func(t *testing.T) {
		cfg := zapcore.EncoderConfig{}
		encoder := NewGinConsoleEncoder(cfg, false, "2006-01-02 15:04:05.000")

		longMessage := strings.Repeat("a", 1000)

		entry := zapcore.Entry{
			Time:    time.Date(2024, 1, 15, 10, 30, 45, 123, time.UTC),
			Level:   zapcore.InfoLevel,
			Message: longMessage,
		}

		buf, err := encoder.EncodeEntry(entry, nil)
		assert.NoError(t, err)

		output := buf.String()
		assert.Contains(t, output, longMessage)
		messagePart := strings.Split(output, " | ")[2]
		assert.Equal(t, 1001, len(messagePart))
	})
}

func TestGinConsoleEncoder_SpecialFieldTypes(t *testing.T) {
	t.Run("Duration字段", func(t *testing.T) {
		cfg := zapcore.EncoderConfig{}
		encoder := NewGinConsoleEncoder(cfg, false, "2006-01-02 15:04:05.000")

		entry := zapcore.Entry{
			Time:    time.Date(2024, 1, 15, 10, 30, 45, 123, time.UTC),
			Level:   zapcore.InfoLevel,
			Message: "test",
		}

		fields := []zapcore.Field{
			zap.Duration("duration", 5*time.Second),
		}

		buf, err := encoder.EncodeEntry(entry, fields)
		assert.NoError(t, err)

		output := buf.String()
		assert.Contains(t, output, "duration=")
	})

	t.Run("Time字段", func(t *testing.T) {
		cfg := zapcore.EncoderConfig{}
		encoder := NewGinConsoleEncoder(cfg, false, "2006-01-02 15:04:05.000")

		entry := zapcore.Entry{
			Time:    time.Date(2024, 1, 15, 10, 30, 45, 123, time.UTC),
			Level:   zapcore.InfoLevel,
			Message: "test",
		}

		fields := []zapcore.Field{
			zap.Time("timestamp", time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)),
		}

		buf, err := encoder.EncodeEntry(entry, fields)
		assert.NoError(t, err)

		output := buf.String()
		assert.Contains(t, output, "timestamp=")
	})

	t.Run("Error字段", func(t *testing.T) {
		cfg := zapcore.EncoderConfig{}
		encoder := NewGinConsoleEncoder(cfg, false, "2006-01-02 15:04:05.000")

		entry := zapcore.Entry{
			Time:    time.Date(2024, 1, 15, 10, 30, 45, 123, time.UTC),
			Level:   zapcore.ErrorLevel,
			Message: "test",
		}

		testErr := assert.AnError
		fields := []zapcore.Field{
			zap.NamedError("error", testErr),
		}

		buf, err := encoder.EncodeEntry(entry, fields)
		assert.NoError(t, err)

		output := buf.String()
		assert.Contains(t, output, "error=")
	})

	t.Run("Complex64字段", func(t *testing.T) {
		cfg := zapcore.EncoderConfig{}
		encoder := NewGinConsoleEncoder(cfg, false, "2006-01-02 15:04:05.000")

		entry := zapcore.Entry{
			Time:    time.Date(2024, 1, 15, 10, 30, 45, 123, time.UTC),
			Level:   zapcore.InfoLevel,
			Message: "test",
		}

		fields := []zapcore.Field{
			zap.Complex64("complex", 1+2i),
		}

		buf, err := encoder.EncodeEntry(entry, fields)
		assert.NoError(t, err)

		output := buf.String()
		assert.Contains(t, output, "complex=")
	})

	t.Run("Reflect字段", func(t *testing.T) {
		cfg := zapcore.EncoderConfig{}
		encoder := NewGinConsoleEncoder(cfg, false, "2006-01-02 15:04:05.000")

		entry := zapcore.Entry{
			Time:    time.Date(2024, 1, 15, 10, 30, 45, 123, time.UTC),
			Level:   zapcore.InfoLevel,
			Message: "test",
		}

		fields := []zapcore.Field{
			zap.Reflect("obj", map[string]int{"a": 1, "b": 2}),
		}

		buf, err := encoder.EncodeEntry(entry, fields)
		assert.NoError(t, err)

		output := buf.String()
		assert.Contains(t, output, "obj=")
	})

	t.Run("Stringer字段", func(t *testing.T) {
		cfg := zapcore.EncoderConfig{}
		encoder := NewGinConsoleEncoder(cfg, false, "2006-01-02 15:04:05.000")

		entry := zapcore.Entry{
			Time:    time.Date(2024, 1, 15, 10, 30, 45, 123, time.UTC),
			Level:   zapcore.InfoLevel,
			Message: "test",
		}

		fields := []zapcore.Field{
			zap.Stringer("stringer", time.Duration(5*time.Second)),
		}

		buf, err := encoder.EncodeEntry(entry, fields)
		assert.NoError(t, err)

		output := buf.String()
		assert.Contains(t, output, "stringer=")
	})
}

func TestGinConsoleEncoder_InterfaceMethods(t *testing.T) {
	t.Run("ConsoleSeparator", func(t *testing.T) {
		cfg := zapcore.EncoderConfig{}
		encoder := NewGinConsoleEncoder(cfg, false, "")

		ginEncoder := encoder.(*ginConsoleEncoder)
		separator := ginEncoder.ConsoleSeparator()

		assert.Equal(t, " | ", separator)
	})

	t.Run("空实现方法不panic", func(t *testing.T) {
		cfg := zapcore.EncoderConfig{}
		encoder := NewGinConsoleEncoder(cfg, false, "")
		ginEncoder := encoder.(*ginConsoleEncoder)

		ginEncoder.AddArray("key", nil)
		ginEncoder.AddObject("key", nil)
		ginEncoder.AddBinary("key", []byte("test"))
		ginEncoder.AddByteString("key", []byte("test"))
		ginEncoder.AddBool("key", true)
		ginEncoder.AddComplex128("key", 1+2i)
		ginEncoder.AddComplex64("key", 1+2i)
		ginEncoder.AddDuration("key", time.Second)
		ginEncoder.AddFloat64("key", 1.23)
		ginEncoder.AddFloat32("key", 1.23)
		ginEncoder.AddInt("key", 123)
		ginEncoder.AddInt64("key", 123)
		ginEncoder.AddInt32("key", 123)
		ginEncoder.AddInt16("key", 123)
		ginEncoder.AddInt8("key", 123)
		ginEncoder.AddString("key", "value")
		ginEncoder.AddTime("key", time.Now())
		ginEncoder.AddUint("key", 123)
		ginEncoder.AddUint64("key", 123)
		ginEncoder.AddUint32("key", 123)
		ginEncoder.AddUint16("key", 123)
		ginEncoder.AddUint8("key", 123)
		ginEncoder.AddUintptr("key", uintptr(123))
		ginEncoder.AddReflected("key", nil)
		ginEncoder.OpenNamespace("key")
	})
}

func TestGinConsoleEncoder_FormatLevel(t *testing.T) {
	cfg := zapcore.EncoderConfig{}
	encoder := NewGinConsoleEncoder(cfg, false, "")
	ginEncoder := encoder.(*ginConsoleEncoder)

	tests := []struct {
		name     string
		level    zapcore.Level
		expected string
	}{
		{"DebugLevel", zapcore.DebugLevel, "DEBUG"},
		{"InfoLevel", zapcore.InfoLevel, "INFO "},
		{"WarnLevel", zapcore.WarnLevel, "WARN "},
		{"ErrorLevel", zapcore.ErrorLevel, "ERROR"},
		{"FatalLevel", zapcore.FatalLevel, "FATAL"},
		{"PanicLevel", zapcore.PanicLevel, "FATAL"},
		{"DPanicLevel", zapcore.DPanicLevel, "FATAL"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ginEncoder.formatLevel(tt.level)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGinConsoleEncoder_LevelColor(t *testing.T) {
	cfg := zapcore.EncoderConfig{}
	encoder := NewGinConsoleEncoder(cfg, true, "")
	ginEncoder := encoder.(*ginConsoleEncoder)

	tests := []struct {
		name     string
		level    zapcore.Level
		expected string
	}{
		{"DebugLevel", zapcore.DebugLevel, "\033[90m"},
		{"InfoLevel", zapcore.InfoLevel, "\033[32m"},
		{"WarnLevel", zapcore.WarnLevel, "\033[33m"},
		{"ErrorLevel", zapcore.ErrorLevel, "\033[31m"},
		{"FatalLevel", zapcore.FatalLevel, "\033[31m\033[1m"},
		{"PanicLevel", zapcore.PanicLevel, "\033[31m\033[1m"},
		{"DPanicLevel", zapcore.DPanicLevel, "\033[31m\033[1m"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ginEncoder.levelColor(tt.level)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGinConsoleEncoder_FormatField(t *testing.T) {
	cfg := zapcore.EncoderConfig{}
	encoder := NewGinConsoleEncoder(cfg, false, "")
	ginEncoder := encoder.(*ginConsoleEncoder)

	t.Run("SkipType返回空字符串", func(t *testing.T) {
		field := zapcore.Field{Type: zapcore.SkipType}
		result := ginEncoder.formatField(field)
		assert.Equal(t, "", result)
	})

	t.Run("NamespaceType返回空字符串", func(t *testing.T) {
		field := zapcore.Field{Type: zapcore.NamespaceType}
		result := ginEncoder.formatField(field)
		assert.Equal(t, "", result)
	})
}
