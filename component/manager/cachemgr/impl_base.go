package cachemgr

import (
	"context"
	"fmt"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"

	"com.litelake.litecore/component/manager/loggermgr"
	"com.litelake.litecore/component/manager/telemetrymgr"
)

// cacheManagerBaseImpl 提供可观测性和工具函数
type cacheManagerBaseImpl struct {
	loggerMgr         loggermgr.ILoggerManager       `inject:""`
	telemetryMgr      telemetrymgr.ITelemetryManager `inject:""`
	logger            loggermgr.Logger
	tracer            trace.Tracer
	meter             metric.Meter
	cacheHitCounter   metric.Int64Counter
	cacheMissCounter  metric.Int64Counter
	operationDuration metric.Float64Histogram
}

// newICacheManagerBaseImpl 创建基类
func newICacheManagerBaseImpl() *cacheManagerBaseImpl {
	return &cacheManagerBaseImpl{}
}

// initObservability 初始化可观测性组件（在依赖注入后调用）
func (b *cacheManagerBaseImpl) initObservability() {
	// 初始化 logger
	if b.loggerMgr != nil {
		b.logger = b.loggerMgr.Logger("cachemgr")
	}

	// 初始化 telemetry
	if b.telemetryMgr != nil {
		b.tracer = b.telemetryMgr.Tracer("cachemgr")
		b.meter = b.telemetryMgr.Meter("cachemgr")

		// 创建指标
		b.cacheHitCounter, _ = b.meter.Int64Counter(
			"cache.hit",
			metric.WithDescription("Cache hit count"),
			metric.WithUnit("{hit}"),
		)
		b.cacheMissCounter, _ = b.meter.Int64Counter(
			"cache.miss",
			metric.WithDescription("Cache miss count"),
			metric.WithUnit("{miss}"),
		)
		b.operationDuration, _ = b.meter.Float64Histogram(
			"cache.operation.duration",
			metric.WithDescription("Cache operation duration in seconds"),
			metric.WithUnit("s"),
		)
	}
}

// recordOperation 记录操作（带链路追踪、指标、日志）
func (b *cacheManagerBaseImpl) recordOperation(
	ctx context.Context,
	driver string,
	operation string,
	key string,
	fn func() error,
) error {
	// 如果没有可观测性配置，直接执行操作
	if b.tracer == nil && b.logger == nil && b.operationDuration == nil {
		return fn()
	}

	var span trace.Span
	if b.tracer != nil {
		ctx, span = b.tracer.Start(ctx, "cache."+operation,
			trace.WithAttributes(
				attribute.String("cache.key", sanitizeKey(key)),
				attribute.String("cache.driver", driver),
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

	if b.logger != nil {
		if err != nil {
			b.logger.Error("cache operation failed",
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
			b.logger.Debug("cache operation success",
				"operation", operation,
				"key", sanitizeKey(key),
				"duration", duration,
			)
		}
	}

	return err
}

// recordCacheHit 记录缓存命中
func (b *cacheManagerBaseImpl) recordCacheHit(ctx context.Context, driver string, hit bool) {
	if b.meter == nil {
		return
	}

	attrs := metric.WithAttributes(
		attribute.String("cache.driver", driver),
	)

	if hit {
		if b.cacheHitCounter != nil {
			b.cacheHitCounter.Add(ctx, 1, attrs)
		}
	} else {
		if b.cacheMissCounter != nil {
			b.cacheMissCounter.Add(ctx, 1, attrs)
		}
	}
}

// sanitizeKey 对缓存键进行脱敏处理，避免敏感信息泄露
func sanitizeKey(key string) string {
	if len(key) <= 10 {
		return key
	}
	// 保留前5个字符，其余用***代替
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

// ValidateKey 验证缓存键是否有效
func ValidateKey(key string) error {
	if key == "" {
		return fmt.Errorf("cache key cannot be empty")
	}
	return nil
}
