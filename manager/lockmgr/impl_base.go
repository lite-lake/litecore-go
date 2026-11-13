package lockmgr

import (
	"context"
	"fmt"
	"github.com/lite-lake/litecore-go/manager/cachemgr"
	"github.com/lite-lake/litecore-go/manager/loggermgr"
	"github.com/lite-lake/litecore-go/manager/telemetrymgr"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"
)

// lockManagerBaseImpl 锁管理器基类实现
// 提供可观测性（日志、指标、链路追踪）和工具函数
type lockManagerBaseImpl struct {
	// loggerMgr 日志管理器，用于记录日志
	loggerMgr loggermgr.ILoggerManager
	// telemetryMgr 遥测管理器，用于指标和链路追踪
	telemetryMgr telemetrymgr.ITelemetryManager
	// cacheMgr 缓存管理器，用于 Redis 实现的底层支持
	cacheMgr cachemgr.ICacheManager
	// tracer 链路追踪器，用于记录操作链路
	tracer trace.Tracer
	// meter 指标记录器，用于记录性能指标
	meter metric.Meter
	// lockAcquireCounter 锁获取计数器
	lockAcquireCounter metric.Int64Counter
	// lockReleaseCounter 锁释放计数器
	lockReleaseCounter metric.Int64Counter
	// lockAcquireFailedCounter 锁获取失败计数器
	lockAcquireFailedCounter metric.Int64Counter
	// operationDuration 操作耗时直方图
	operationDuration metric.Float64Histogram
}

// newLockManagerBaseImpl 创建基类
// 参数：
//   - loggerMgr: 日志管理器
//   - telemetryMgr: 遥测管理器
//   - cacheMgr: 缓存管理器
func newLockManagerBaseImpl(
	loggerMgr loggermgr.ILoggerManager,
	telemetryMgr telemetrymgr.ITelemetryManager,
	cacheMgr cachemgr.ICacheManager,
) *lockManagerBaseImpl {
	return &lockManagerBaseImpl{
		loggerMgr:    loggerMgr,
		telemetryMgr: telemetryMgr,
		cacheMgr:     cacheMgr,
	}
}

// initObservability 初始化可观测性组件
// 在依赖注入完成后调用，用于初始化链路追踪器和指标收集器
func (b *lockManagerBaseImpl) initObservability() {
	if b.telemetryMgr == nil {
		return
	}

	b.tracer = b.telemetryMgr.Tracer("lockmgr")
	b.meter = b.telemetryMgr.Meter("lockmgr")

	b.lockAcquireCounter, _ = b.meter.Int64Counter(
		"lock.acquire",
		metric.WithDescription("锁获取次数"),
		metric.WithUnit("{lock}"),
	)

	b.lockReleaseCounter, _ = b.meter.Int64Counter(
		"lock.release",
		metric.WithDescription("锁释放次数"),
		metric.WithUnit("{lock}"),
	)

	b.lockAcquireFailedCounter, _ = b.meter.Int64Counter(
		"lock.acquire_failed",
		metric.WithDescription("锁获取失败次数"),
		metric.WithUnit("{lock}"),
	)

	b.operationDuration, _ = b.meter.Float64Histogram(
		"lock.operation.duration",
		metric.WithDescription("锁操作耗时（秒）"),
		metric.WithUnit("s"),
	)
}

// recordOperation 记录操作并执行函数
// 封装了链路追踪、指标记录、日志记录等功能
func (b *lockManagerBaseImpl) recordOperation(
	ctx context.Context,
	driver string,
	operation string,
	key string,
	fn func() error,
) error {
	if b.tracer == nil && b.loggerMgr == nil && b.operationDuration == nil {
		return fn()
	}

	var span trace.Span
	if b.tracer != nil {
		ctx, span = b.tracer.Start(ctx, "lock."+operation,
			trace.WithAttributes(
				attribute.String("lock.key", sanitizeKey(key)),
				attribute.String("lock.driver", driver),
			),
		)
		defer span.End()
	}

	start := time.Now()
	err := fn()
	duration := time.Since(start).Seconds()

	if b.operationDuration != nil {
		b.operationDuration.Record(ctx, duration,
			metric.WithAttributes(
				attribute.String("operation", operation),
				attribute.String("status", getStatus(err)),
			),
		)
	}

	if b.loggerMgr != nil {
		logger := b.loggerMgr.Ins()
		if err != nil {
			logger.Error("lock operation failed",
				"operation", operation,
				"key", sanitizeKey(key),
				"error", err.Error(),
				"duration", duration,
			)
			if span != nil {
				span.RecordError(err)
				span.SetStatus(codes.Error, err.Error())
			}
		} else {
			logger.Debug("lock operation success",
				"operation", operation,
				"key", sanitizeKey(key),
				"duration", duration,
			)
		}
	}

	return err
}

// recordLockAcquire 记录锁获取事件
func (b *lockManagerBaseImpl) recordLockAcquire(ctx context.Context, driver string, success bool) {
	if b.meter == nil {
		return
	}

	attrs := metric.WithAttributes(
		attribute.String("lock.driver", driver),
	)

	if success {
		if b.lockAcquireCounter != nil {
			b.lockAcquireCounter.Add(ctx, 1, attrs)
		}
	} else {
		if b.lockAcquireFailedCounter != nil {
			b.lockAcquireFailedCounter.Add(ctx, 1, attrs)
		}
	}
}

// recordLockRelease 记录锁释放事件
func (b *lockManagerBaseImpl) recordLockRelease(ctx context.Context, driver string) {
	if b.meter == nil {
		return
	}

	attrs := metric.WithAttributes(
		attribute.String("lock.driver", driver),
	)

	if b.lockReleaseCounter != nil {
		b.lockReleaseCounter.Add(ctx, 1, attrs)
	}
}

// sanitizeKey 对锁键进行脱敏处理
func sanitizeKey(key string) string {
	if len(key) <= 10 {
		return key
	}
	return key[:5] + "***"
}

// getStatus 根据错误返回状态字符串
func getStatus(err error) string {
	if err != nil {
		return "error"
	}
	return "success"
}

// ValidateContext 验证上下文是否有效
func ValidateContext(ctx context.Context) error {
	if ctx == nil {
		return fmt.Errorf("context cannot be nil")
	}
	return nil
}

// ValidateKey 验证锁键是否有效
func ValidateKey(key string) error {
	if key == "" {
		return fmt.Errorf("lock key cannot be empty")
	}
	return nil
}
