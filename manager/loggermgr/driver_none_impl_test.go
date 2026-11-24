package loggermgr

import (
	"testing"

	"github.com/lite-lake/litecore-go/logger"
	"github.com/stretchr/testify/assert"
)

func TestDriverNoneLoggerManager(t *testing.T) {
	t.Run("创建空日志管理器", func(t *testing.T) {
		mgr := NewDriverNoneLoggerManager()
		assert.NotNil(t, mgr)
		assert.Equal(t, "LoggerNoneManager", mgr.ManagerName())
	})

	t.Run("获取日志实例", func(t *testing.T) {
		mgr := NewDriverNoneLoggerManager()
		log := mgr.Ins()
		assert.NotNil(t, log)
	})

	t.Run("健康检查", func(t *testing.T) {
		mgr := NewDriverNoneLoggerManager()
		err := mgr.Health()
		assert.NoError(t, err)
	})

	t.Run("生命周期", func(t *testing.T) {
		mgr := NewDriverNoneLoggerManager()
		assert.NoError(t, mgr.OnStart())
		assert.NoError(t, mgr.OnStop())
	})

	t.Run("日志输出", func(t *testing.T) {
		mgr := NewDriverNoneLoggerManager()
		log := mgr.Ins()

		assert.NotPanics(t, func() {
			log.Debug("debug message", "key", "value")
			log.Info("info message", "key", "value")
			log.Warn("warn message", "key", "value")
			log.Error("error message", "key", "value")
		})
	})

	t.Run("With 方法", func(t *testing.T) {
		mgr := NewDriverNoneLoggerManager()
		log := mgr.Ins()
		logWithCtx := log.With("service", "test-service", "version", "1.0.0")
		assert.NotNil(t, logWithCtx)

		assert.NotPanics(t, func() {
			logWithCtx.Info("message with context")
		})
	})

	t.Run("SetLevel 方法", func(t *testing.T) {
		mgr := NewDriverNoneLoggerManager()
		log := mgr.Ins()

		assert.NotPanics(t, func() {
			log.SetLevel(logger.DebugLevel)
			log.SetLevel(logger.InfoLevel)
			log.SetLevel(logger.WarnLevel)
			log.SetLevel(logger.ErrorLevel)
		})
	})
}

func TestNoneLogger(t *testing.T) {
	t.Run("空日志不输出任何内容", func(t *testing.T) {
		nl := &noneLogger{}

		assert.NotPanics(t, func() {
			nl.Debug("debug message")
			nl.Info("info message")
			nl.Warn("warn message")
			nl.Error("error message")
			nl.Fatal("fatal message")
		})
	})

	t.Run("With 方法返回自身", func(t *testing.T) {
		nl := &noneLogger{}
		result := nl.With("key", "value")
		assert.Equal(t, nl, result)
	})

	t.Run("SetLevel 方法不执行任何操作", func(t *testing.T) {
		nl := &noneLogger{}
		assert.NotPanics(t, func() {
			nl.SetLevel(logger.DebugLevel)
			nl.SetLevel(logger.InfoLevel)
			nl.SetLevel(logger.WarnLevel)
			nl.SetLevel(logger.ErrorLevel)
		})
	})
}
