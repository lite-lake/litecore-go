package drivers

import (
	"context"
	"fmt"
	"sync"

	"com.litelake.litecore/manager/loggermgr/internal/config"
	"com.litelake.litecore/manager/loggermgr/internal/loglevel"
	"com.litelake.litecore/manager/telemetrymgr"
)

// ZapDriver zap 日志驱动
// 实现 Driver 接口，封装 ZapLoggerManager
type ZapDriver struct {
	config      *config.LoggerConfig
	telemetryMgr telemetrymgr.TelemetryManager
	manager     *ZapLoggerManager
	mu          sync.RWMutex
	started     bool
}

// NewZapDriver 创建 zap 日志驱动
// cfg: 日志配置
// otelTracerProvider: OTEL TracerProvider（可选），用于集成观测
func NewZapDriver(cfg *config.LoggerConfig, otelTracerProvider interface{}) (*ZapDriver, error) {
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid logger config: %w", err)
	}

	// 将 OTEL TracerProvider 转换为 TelemetryManager（如果有）
	var telemetryMgr telemetrymgr.TelemetryManager
	if tp, ok := otelTracerProvider.(interface{}); ok {
		// 这里假设 TelemetryManager 已经实现了相关接口
		// 实际使用时需要根据具体情况调整
		_ = tp // 暂时忽略，避免 unused 错误
		telemetryMgr = nil
	}

	return &ZapDriver{
		config:      cfg,
		telemetryMgr: telemetryMgr,
		started:     false,
	}, nil
}

// Start 启动日志驱动
// 实现 Driver 接口
func (d *ZapDriver) Start() error {
	d.mu.Lock()
	defer d.mu.Unlock()

	if d.started {
		return nil
	}

	// 创建 ZapLoggerManager
	manager, err := NewZapLoggerManager(d.config, d.telemetryMgr)
	if err != nil {
		return fmt.Errorf("create zap logger manager failed: %w", err)
	}

	d.manager = manager
	d.started = true

	return nil
}

// Shutdown 关闭日志驱动
// 实现 Driver 接口
func (d *ZapDriver) Shutdown(ctx context.Context) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	if !d.started || d.manager == nil {
		return nil
	}

	if err := d.manager.Shutdown(ctx); err != nil {
		return fmt.Errorf("shutdown zap logger manager failed: %w", err)
	}

	d.started = false
	d.manager = nil

	return nil
}

// Health 检查驱动健康状态
// 实现 Driver 接口
func (d *ZapDriver) Health() error {
	d.mu.RLock()
	defer d.mu.RUnlock()

	if !d.started {
		return fmt.Errorf("driver not started")
	}

	if d.manager == nil {
		return fmt.Errorf("manager not initialized")
	}

	return d.manager.Health()
}

// GetLogger 获取指定名称的 Logger 实例
// 实现 Driver 接口
func (d *ZapDriver) GetLogger(name string) Logger {
	d.mu.RLock()
	defer d.mu.RUnlock()

	if !d.started || d.manager == nil {
		// 如果驱动未启动，返回 NoneLogger
		return NewNoneLogger()
	}

	zapLogger := d.manager.GetLogger(name)
	return zapLogger
}

// SetLevel 设置日志级别
// 实现 Driver 接口
func (d *ZapDriver) SetLevel(level loglevel.LogLevel) {
	d.mu.RLock()
	defer d.mu.RUnlock()

	if !d.started || d.manager == nil {
		return
	}

	zapLevel := loglevel.LogLevelToZap(level)
	d.manager.SetGlobalLevel(zapLevel)
}

// 确保 ZapDriver 实现 Driver 接口
var _ Driver = (*ZapDriver)(nil)
