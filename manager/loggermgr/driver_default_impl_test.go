package loggermgr

import (
	"testing"

	"github.com/lite-lake/litecore-go/logger"
	"github.com/stretchr/testify/assert"
)

func TestDriverDefaultLoggerManager(t *testing.T) {
	t.Run("创建默认日志管理器", func(t *testing.T) {
		mgr := NewDriverDefaultLoggerManager()
		assert.NotNil(t, mgr)
		assert.Equal(t, "LoggerDefaultManager", mgr.ManagerName())
	})

	t.Run("获取日志实例", func(t *testing.T) {
		mgr := NewDriverDefaultLoggerManager()
		log := mgr.Ins()
		assert.NotNil(t, log)
	})

	t.Run("健康检查", func(t *testing.T) {
		mgr := NewDriverDefaultLoggerManager()
		err := mgr.Health()
		assert.NoError(t, err)
	})

	t.Run("生命周期", func(t *testing.T) {
		mgr := NewDriverDefaultLoggerManager()
		assert.NoError(t, mgr.OnStart())
		assert.NoError(t, mgr.OnStop())
	})

	t.Run("日志输出", func(t *testing.T) {
		mgr := NewDriverDefaultLoggerManager()
		log := mgr.Ins()

		assert.NotPanics(t, func() {
			log.Debug("debug message", "key", "value")
			log.Info("info message", "key", "value")
			log.Warn("warn message", "key", "value")
			log.Error("error message", "key", "value")
		})
	})

	t.Run("With 方法", func(t *testing.T) {
		mgr := NewDriverDefaultLoggerManager()
		log := mgr.Ins()
		logWithCtx := log.With("service", "test-service", "version", "1.0.0")
		assert.NotNil(t, logWithCtx)

		assert.NotPanics(t, func() {
			logWithCtx.Info("message with context")
		})
	})

	t.Run("SetLevel 方法", func(t *testing.T) {
		mgr := NewDriverDefaultLoggerManager()
		log := mgr.Ins()

		assert.NotPanics(t, func() {
			log.SetLevel(logger.DebugLevel)
			log.SetLevel(logger.InfoLevel)
			log.SetLevel(logger.WarnLevel)
			log.SetLevel(logger.ErrorLevel)
		})
	})
}
