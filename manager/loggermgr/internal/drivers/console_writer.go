package drivers

import (
	"fmt"
	"os"
	"sync"
	"time"

	"com.litelake.litecore/manager/loggermgr/internal/config"
	"com.litelake.litecore/manager/loggermgr/internal/loglevel"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	// 全局终端颜色支持检测，缓存检测结果
	supportsColor bool
	colorChecked  bool
)

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

// ConsoleEncoderConfig 控制台编码器配置
type ConsoleEncoderConfig struct {
	EncodeLevel      zapcore.LevelEncoder
	EncodeTime       zapcore.TimeEncoder
	EncodeDuration   zapcore.DurationEncoder
	EncodeCaller     zapcore.CallerEncoder
	ConsoleSeparator string
}

// DefaultConsoleEncoderConfig 默认控制台编码器配置
func DefaultConsoleEncoderConfig() ConsoleEncoderConfig {
	return ConsoleEncoderConfig{
		EncodeLevel:      customLevelEncoder,
		EncodeTime:       CustomTimeEncoder,
		EncodeDuration:   zapcore.StringDurationEncoder,
		EncodeCaller:     zapcore.ShortCallerEncoder,
		ConsoleSeparator: " | ",
	}
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

// CustomTimeEncoder 自定义时间编码器
func CustomTimeEncoder(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
	enc.AppendString(t.Format("2006-01-02 15:04:05.000"))
}

// ConsoleConfig 控制台配置（用于兼容接口）
type ConsoleConfig struct {
	Level string `yaml:"level"` // 日志级别
}

// ParseConsoleConfigFromMap 从 ConfigMap 解析控制台配置
func ParseConsoleConfigFromMap(cfg map[string]any) (*ConsoleConfig, error) {
	consoleConfig := &ConsoleConfig{
		Level: "info",
	}

	if cfg == nil {
		return consoleConfig, nil
	}

	if level, ok := cfg["level"].(string); ok {
		consoleConfig.Level = level
	}

	return consoleConfig, nil
}

// Validate 验证控制台配置
func (c *ConsoleConfig) Validate() error {
	if !loglevel.IsValidLogLevel(c.Level) {
		return fmt.Errorf("invalid console log level: %s", c.Level)
	}
	return nil
}

// ConsoleWriter 控制台日志输出器（基于 zap）
type ConsoleWriter struct {
	logger *zap.Logger
	level  zapcore.Level
	sync   *zap.SugaredLogger
	mu     sync.RWMutex
}

// NewConsoleWriter 创建控制台日志输出器
func NewConsoleWriter(cfg *config.LogLevelConfig) *ConsoleWriter {
	level := zapcore.InfoLevel
	if cfg != nil && cfg.Level != "" {
		level = LogLevelToZap(loglevel.ParseLogLevel(cfg.Level))
	}

	// 延迟检测颜色支持
	if !colorChecked {
		supportsColor = detectColorSupport()
		colorChecked = true
	}

	// 创建控制台编码器配置
	encoderConfig := DefaultConsoleEncoderConfig()
	zapConfig := zapcore.EncoderConfig{
		TimeKey:          "time",
		LevelKey:         "level",
		NameKey:          "logger",
		CallerKey:        "caller",
		MessageKey:       "msg",
		StacktraceKey:    "stacktrace",
		LineEnding:       zapcore.DefaultLineEnding,
		EncodeLevel:      encoderConfig.EncodeLevel,
		EncodeTime:       encoderConfig.EncodeTime,
		EncodeDuration:   encoderConfig.EncodeDuration,
		EncodeCaller:     encoderConfig.EncodeCaller,
		ConsoleSeparator: encoderConfig.ConsoleSeparator,
	}

	// 创建控制台编码器
	encoder := zapcore.NewConsoleEncoder(zapConfig)

	// 创建核心（输出到 stdout/stderr）
	stdoutCore := zapcore.NewCore(encoder, zapcore.AddSync(os.Stdout), level)
	stderrCore := zapcore.NewCore(encoder, zapcore.AddSync(os.Stderr), zapcore.ErrorLevel)

	// 组合核心
	core := zapcore.NewTee(stdoutCore, stderrCore)

	// 创建 logger
	logger := zap.New(core, zap.AddCallerSkip(1), zap.AddStacktrace(zapcore.ErrorLevel))

	return &ConsoleWriter{
		logger: logger,
		level:  level,
		sync:   logger.Sugar(),
	}
}

// Debug 记录调试级别日志
func (w *ConsoleWriter) Debug(msg string, fields ...zap.Field) {
	w.mu.RLock()
	defer w.mu.RUnlock()
	w.logger.Debug(msg, fields...)
}

// Info 记录信息级别日志
func (w *ConsoleWriter) Info(msg string, fields ...zap.Field) {
	w.mu.RLock()
	defer w.mu.RUnlock()
	w.logger.Info(msg, fields...)
}

// Warn 记录警告级别日志
func (w *ConsoleWriter) Warn(msg string, fields ...zap.Field) {
	w.mu.RLock()
	defer w.mu.RUnlock()
	w.logger.Warn(msg, fields...)
}

// Error 记录错误级别日志
func (w *ConsoleWriter) Error(msg string, fields ...zap.Field) {
	w.mu.RLock()
	defer w.mu.RUnlock()
	w.logger.Error(msg, fields...)
}

// Fatal 记录致命错误级别日志
func (w *ConsoleWriter) Fatal(msg string, fields ...zap.Field) {
	w.mu.RLock()
	defer w.mu.RUnlock()
	w.logger.Fatal(msg, fields...)
}

// Debugf 记录调试级别日志（格式化）
func (w *ConsoleWriter) Debugf(template string, args ...any) {
	w.mu.RLock()
	defer w.mu.RUnlock()
	w.sync.Debugf(template, args...)
}

// Infof 记录信息级别日志（格式化）
func (w *ConsoleWriter) Infof(template string, args ...any) {
	w.mu.RLock()
	defer w.mu.RUnlock()
	w.sync.Infof(template, args...)
}

// Warnf 记录警告级别日志（格式化）
func (w *ConsoleWriter) Warnf(template string, args ...any) {
	w.mu.RLock()
	defer w.mu.RUnlock()
	w.sync.Warnf(template, args...)
}

// Errorf 记录错误级别日志（格式化）
func (w *ConsoleWriter) Errorf(template string, args ...any) {
	w.mu.RLock()
	defer w.mu.RUnlock()
	w.sync.Errorf(template, args...)
}

// Fatalf 记录致命错误级别日志（格式化）
func (w *ConsoleWriter) Fatalf(template string, args ...any) {
	w.mu.RLock()
	defer w.mu.RUnlock()
	w.sync.Fatalf(template, args...)
}

// SetLevel 设置日志级别
func (w *ConsoleWriter) SetLevel(level zapcore.Level) {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.level = level
}

// IsEnabled 检查是否启用指定级别的日志
func (w *ConsoleWriter) IsEnabled(level zapcore.Level) bool {
	w.mu.RLock()
	defer w.mu.RUnlock()
	return level >= w.level
}

// With 返回一个带有额外字段的新 Logger
func (w *ConsoleWriter) With(fields ...zap.Field) *ConsoleWriter {
	w.mu.RLock()
	defer w.mu.RUnlock()

	return &ConsoleWriter{
		logger: w.logger.With(fields...),
		level:  w.level,
		sync:   w.logger.With(fields...).Sugar(),
	}
}

// Sync 同步日志
func (w *ConsoleWriter) Sync() error {
	return w.logger.Sync()
}

// GetLogger 获取底层 zap.Logger
func (w *ConsoleWriter) GetLogger() *zap.Logger {
	w.mu.RLock()
	defer w.mu.RUnlock()
	return w.logger
}
