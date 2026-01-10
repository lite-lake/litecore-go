package drivers

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
	"time"

	"com.litelake.litecore/manager/loggermgr/internal/config"
	"com.litelake.litecore/manager/loggermgr/internal/loglevel"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

// FileWriter 文件日志输出器（基于 zap）
type FileWriter struct {
	logger        *zap.Logger
	level         zapcore.Level
	sync          *zap.SugaredLogger
	lumberjack    *lumberjack.Logger
	mu            sync.RWMutex
	closed        bool
	writeErrors   uint64 // 写入错误计数（原子操作）
	lastWriteError time.Time
	onWriteError  func(error) // 写入错误回调
}

// NewFileWriter 创建文件日志输出器
func NewFileWriter(cfg *config.FileLogConfig) (*FileWriter, error) {
	if cfg == nil {
		return nil, fmt.Errorf("file log config is required")
	}

	level := zapcore.InfoLevel
	if cfg.Level != "" {
		level = LogLevelToZap(loglevel.ParseLogLevel(cfg.Level))
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

	return &FileWriter{
		logger:     logger,
		level:      level,
		sync:       logger.Sugar(),
		lumberjack: lumberjackLogger,
		closed:     false,
	}, nil
}

// SetWriteErrorCallback 设置写入错误回调函数
func (w *FileWriter) SetWriteErrorCallback(fn func(error)) {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.onWriteError = fn
}

// Debug 记录调试级别日志
func (w *FileWriter) Debug(msg string, fields ...zap.Field) {
	w.mu.RLock()
	if w.closed {
		w.mu.RUnlock()
		return
	}
	w.mu.RUnlock()

	w.logger.Debug(msg, fields...)
}

// Info 记录信息级别日志
func (w *FileWriter) Info(msg string, fields ...zap.Field) {
	w.mu.RLock()
	if w.closed {
		w.mu.RUnlock()
		return
	}
	w.mu.RUnlock()

	w.logger.Info(msg, fields...)
}

// Warn 记录警告级别日志
func (w *FileWriter) Warn(msg string, fields ...zap.Field) {
	w.mu.RLock()
	if w.closed {
		w.mu.RUnlock()
		return
	}
	w.mu.RUnlock()

	w.logger.Warn(msg, fields...)
}

// Error 记录错误级别日志
func (w *FileWriter) Error(msg string, fields ...zap.Field) {
	w.mu.RLock()
	if w.closed {
		w.mu.RUnlock()
		return
	}
	w.mu.RUnlock()

	w.logger.Error(msg, fields...)
}

// Fatal 记录致命错误级别日志
func (w *FileWriter) Fatal(msg string, fields ...zap.Field) {
	w.mu.RLock()
	if w.closed {
		w.mu.RUnlock()
		return
	}
	w.mu.RUnlock()

	w.logger.Fatal(msg, fields...)
}

// Debugf 记录调试级别日志（格式化）
func (w *FileWriter) Debugf(template string, args ...any) {
	w.mu.RLock()
	if w.closed {
		w.mu.RUnlock()
		return
	}
	w.mu.RUnlock()

	w.sync.Debugf(template, args...)
}

// Infof 记录信息级别日志（格式化）
func (w *FileWriter) Infof(template string, args ...any) {
	w.mu.RLock()
	if w.closed {
		w.mu.RUnlock()
		return
	}
	w.mu.RUnlock()

	w.sync.Infof(template, args...)
}

// Warnf 记录警告级别日志（格式化）
func (w *FileWriter) Warnf(template string, args ...any) {
	w.mu.RLock()
	if w.closed {
		w.mu.RUnlock()
		return
	}
	w.mu.RUnlock()

	w.sync.Warnf(template, args...)
}

// Errorf 记录错误级别日志（格式化）
func (w *FileWriter) Errorf(template string, args ...any) {
	w.mu.RLock()
	if w.closed {
		w.mu.RUnlock()
		return
	}
	w.mu.RUnlock()

	w.sync.Errorf(template, args...)
}

// Fatalf 记录致命错误级别日志（格式化）
func (w *FileWriter) Fatalf(template string, args ...any) {
	w.mu.RLock()
	if w.closed {
		w.mu.RUnlock()
		return
	}
	w.mu.RUnlock()

	w.sync.Fatalf(template, args...)
}

// SetLevel 设置日志级别
func (w *FileWriter) SetLevel(level zapcore.Level) {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.level = level
}

// IsEnabled 检查是否启用指定级别的日志
func (w *FileWriter) IsEnabled(level zapcore.Level) bool {
	w.mu.RLock()
	defer w.mu.RUnlock()
	return level >= w.level
}

// With 返回一个带有额外字段的新 Logger
func (w *FileWriter) With(fields ...zap.Field) *FileWriter {
	w.mu.RLock()
	defer w.mu.RUnlock()

	return &FileWriter{
		logger:     w.logger.With(fields...),
		level:      w.level,
		sync:       w.logger.With(fields...).Sugar(),
		lumberjack: w.lumberjack,
	}
}

// Close 关闭日志输出器
func (w *FileWriter) Close() error {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.closed {
		return nil
	}

	w.closed = true
	return w.lumberjack.Close()
}

// Sync 同步日志
func (w *FileWriter) Sync() error {
	w.mu.RLock()
	defer w.mu.RUnlock()

	if w.closed {
		return nil
	}

	return w.logger.Sync()
}

// GetWriteErrorCount 获取写入错误计数
func (w *FileWriter) GetWriteErrorCount() uint64 {
	return atomic.LoadUint64(&w.writeErrors)
}

// GetLastWriteError 获取最后一次写入错误的时间
func (w *FileWriter) GetLastWriteError() time.Time {
	w.mu.RLock()
	defer w.mu.RUnlock()
	return w.lastWriteError
}

// Rotate 手动触发日志轮转
func (w *FileWriter) Rotate() error {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.closed {
		return fmt.Errorf("file writer is closed")
	}

	return w.lumberjack.Rotate()
}

// GetLogger 获取底层 zap.Logger
func (w *FileWriter) GetLogger() *zap.Logger {
	w.mu.RLock()
	defer w.mu.RUnlock()
	return w.logger
}
