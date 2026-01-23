package cachemgr

import (
	"context"
	"fmt"
	"github.com/lite-lake/litecore-go/logger"
	"github.com/lite-lake/litecore-go/manager/telemetrymgr"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"
)

// cacheManagerBaseImpl 缓存管理器基类实现
// 提供可观测性（日志、指标、链路追踪）和工具函数
type cacheManagerBaseImpl struct {
	// Logger 日志记录器，通过依赖注入获取
	Logger logger.ILogger `inject:""`
	// telemetryMgr 遥测管理器，用于指标和链路追踪
	telemetryMgr telemetrymgr.ITelemetryManager `inject:""`
	// tracer 链路追踪器，用于记录操作链路
	tracer trace.Tracer
	// meter 指标记录器，用于记录性能指标
	meter metric.Meter
	// cacheHitCounter 缓存命中计数器
	cacheHitCounter metric.Int64Counter
	// cacheMissCounter 缓存未命中计数器
	cacheMissCounter metric.Int64Counter
	// operationDuration 操作耗时直方图
	operationDuration metric.Float64Histogram
}

// newICacheManagerBaseImpl 创建基类
func newICacheManagerBaseImpl() *cacheManagerBaseImpl {
	return &cacheManagerBaseImpl{}
}

// initObservability 初始化可观测性组件
// 在依赖注入完成后调用，用于初始化链路追踪器和指标收集器
func (b *cacheManagerBaseImpl) initObservability() {
	if b.telemetryMgr == nil {
		return
	}

	// 初始化 tracer，用于链路追踪
	b.tracer = b.telemetryMgr.Tracer("cachemgr")

	// 初始化 meter，用于指标记录
	b.meter = b.telemetryMgr.Meter("cachemgr")

	// 创建缓存命中计数器指标
	b.cacheHitCounter, _ = b.meter.Int64Counter(
		"cache.hit",
		metric.WithDescription("缓存命中次数"),
		metric.WithUnit("{hit}"),
	)

	// 创建缓存未命中计数器指标
	b.cacheMissCounter, _ = b.meter.Int64Counter(
		"cache.miss",
		metric.WithDescription("缓存未命中次数"),
		metric.WithUnit("{miss}"),
	)

	// 创建操作耗时直方图指标
	b.operationDuration, _ = b.meter.Float64Histogram(
		"cache.operation.duration",
		metric.WithDescription("缓存操作耗时（秒）"),
		metric.WithUnit("s"),
	)
}

// recordOperation 记录操作并执行函数
// 封装了链路追踪、指标记录、日志记录等功能
// 参数：
//   - ctx: 上下文
//   - driver: 缓存驱动类型（memory、redis、none）
//   - operation: 操作名称（get、set、delete 等）
//   - key: 缓存键（用于日志脱敏）
//   - fn: 要执行的操作函数
func (b *cacheManagerBaseImpl) recordOperation(
	ctx context.Context,
	driver string,
	operation string,
	key string,
	fn func() error,
) error {
	// 如果没有配置任何可观测性组件，直接执行操作
	if b.tracer == nil && b.Logger == nil && b.operationDuration == nil {
		return fn()
	}

	var span trace.Span
	// 创建链路追踪 span
	if b.tracer != nil {
		ctx, span = b.tracer.Start(ctx, "cache."+operation,
			trace.WithAttributes(
				attribute.String("cache.key", sanitizeKey(key)),
				attribute.String("cache.driver", driver),
			),
		)
		defer span.End()
	}

	// 记录操作开始时间并执行函数
	start := time.Now()
	err := fn()
	duration := time.Since(start).Seconds()

	// 记录操作耗时指标
	if b.operationDuration != nil {
		b.operationDuration.Record(ctx, duration,
			metric.WithAttributes(
				attribute.String("operation", operation),
				attribute.String("status", getStatus(err)),
			),
		)
	}

	// 记录日志
	if b.Logger != nil {
		if err != nil {
			// 操作失败，记录错误日志
			b.Logger.Error("cache operation failed",
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
			// 操作成功，记录调试日志
			b.Logger.Debug("cache operation success",
				"operation", operation,
				"key", sanitizeKey(key),
				"duration", duration,
			)
		}
	}

	return err
}

// recordCacheHit 记录缓存命中或未命中指标
// 参数：
//   - ctx: 上下文
//   - driver: 缓存驱动类型
//   - hit: 是否命中缓存
func (b *cacheManagerBaseImpl) recordCacheHit(ctx context.Context, driver string, hit bool) {
	if b.meter == nil {
		return
	}

	// 设置指标属性
	attrs := metric.WithAttributes(
		attribute.String("cache.driver", driver),
	)

	// 根据命中情况记录对应的计数器
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

// sanitizeKey 对缓存键进行脱敏处理
// 避免在日志中暴露敏感信息，只保留前5个字符
func sanitizeKey(key string) string {
	if len(key) <= 10 {
		return key
	}
	// 保留前5个字符，其余用***代替
	return key[:5] + "***"
}

// getStatus 根据错误返回状态字符串
// 用于指标记录和日志分类
func getStatus(err error) string {
	if err != nil {
		return "error"
	}
	return "success"
}

// ValidateContext 验证上下文是否有效
// 确保传入的 context 不为 nil
func ValidateContext(ctx context.Context) error {
	if ctx == nil {
		return fmt.Errorf("context cannot be nil")
	}
	return nil
}

// ValidateKey 验证缓存键是否有效
// 确保键不为空字符串
func ValidateKey(key string) error {
	if key == "" {
		return fmt.Errorf("cache key cannot be empty")
	}
	return nil
}
