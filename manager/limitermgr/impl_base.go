package limitermgr

import (
	"context"
	"fmt"
	"time"

	"github.com/lite-lake/litecore-go/manager/cachemgr"
	"github.com/lite-lake/litecore-go/manager/loggermgr"
	"github.com/lite-lake/litecore-go/manager/telemetrymgr"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"
)

// limiterManagerBaseImpl 限流管理器基类实现
// 提供可观测性（日志、指标、链路追踪）和工具函数
type limiterManagerBaseImpl struct {
	// loggerMgr 日志管理器，用于记录日志
	loggerMgr loggermgr.ILoggerManager
	// telemetryMgr 遥测管理器，用于指标和链路追踪
	telemetryMgr telemetrymgr.ITelemetryManager
	// cacheMgr 缓存管理器，用于 Redis 实现
	cacheMgr cachemgr.ICacheManager
	// tracer 链路追踪器，用于记录操作链路
	tracer trace.Tracer
	// meter 指标记录器，用于记录性能指标
	meter metric.Meter
	// allowedCounter 允许通过计数器
	allowedCounter metric.Int64Counter
	// rejectedCounter 拒绝计数器
	rejectedCounter metric.Int64Counter
	// operationDuration 操作耗时直方图
	operationDuration metric.Float64Histogram
}

// newILimiterManagerBaseImpl 创建基类
// 参数：
//   - loggerMgr: 日志管理器
//   - telemetryMgr: 遥测管理器
//   - cacheMgr: 缓存管理器（Redis 实现需要，memory 可传 nil）
func newILimiterManagerBaseImpl(
	loggerMgr loggermgr.ILoggerManager,
	telemetryMgr telemetrymgr.ITelemetryManager,
	cacheMgr cachemgr.ICacheManager,
) *limiterManagerBaseImpl {
	return &limiterManagerBaseImpl{
		loggerMgr:    loggerMgr,
		telemetryMgr: telemetryMgr,
		cacheMgr:     cacheMgr,
	}
}

// initObservability 初始化可观测性组件
// 在依赖注入完成后调用，用于初始化链路追踪器和指标收集器
func (b *limiterManagerBaseImpl) initObservability() {
	if b.telemetryMgr == nil {
		return
	}

	// 初始化 tracer，用于链路追踪
	b.tracer = b.telemetryMgr.Tracer("limitermgr")

	// 初始化 meter，用于指标记录
	b.meter = b.telemetryMgr.Meter("limitermgr")

	// 创建允许通过计数器指标
	b.allowedCounter, _ = b.meter.Int64Counter(
		"limiter.allowed",
		metric.WithDescription("限流允许通过次数"),
		metric.WithUnit("{request}"),
	)

	// 创建拒绝计数器指标
	b.rejectedCounter, _ = b.meter.Int64Counter(
		"limiter.rejected",
		metric.WithDescription("限流拒绝次数"),
		metric.WithUnit("{request}"),
	)

	// 创建操作耗时直方图指标
	b.operationDuration, _ = b.meter.Float64Histogram(
		"limiter.operation.duration",
		metric.WithDescription("限流操作耗时（秒）"),
		metric.WithUnit("s"),
	)
}

// recordOperation 记录操作并执行函数
// 封装了链路追踪、指标记录、日志记录等功能
// 参数：
//   - ctx: 上下文
//   - driver: 限流驱动类型（memory、redis、none）
//   - operation: 操作名称（allow、get_remaining 等）
//   - key: 限流键（用于日志脱敏）
//   - fn: 要执行的操作函数
func (b *limiterManagerBaseImpl) recordOperation(
	ctx context.Context,
	driver string,
	operation string,
	key string,
	fn func() error,
) error {
	// 如果没有配置任何可观测性组件，直接执行操作
	if b.tracer == nil && b.loggerMgr == nil && b.operationDuration == nil {
		return fn()
	}

	var span trace.Span
	// 创建链路追踪 span
	if b.tracer != nil {
		ctx, span = b.tracer.Start(ctx, "limiter."+operation,
			trace.WithAttributes(
				attribute.String("limiter.key", sanitizeKey(key)),
				attribute.String("limiter.driver", driver),
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
	if b.loggerMgr != nil {
		logger := b.loggerMgr.Ins()
		if err != nil {
			// 操作失败，记录错误日志
			logger.Error("limiter operation failed",
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
			logger.Debug("limiter operation success",
				"operation", operation,
				"key", sanitizeKey(key),
				"duration", duration,
			)
		}
	}

	return err
}

// recordAllowance 记录允许通过或拒绝指标
// 参数：
//   - ctx: 上下文
//   - driver: 限流驱动类型
//   - allowed: 是否允许通过
func (b *limiterManagerBaseImpl) recordAllowance(ctx context.Context, driver string, allowed bool) {
	if b.meter == nil {
		return
	}

	// 设置指标属性
	attrs := metric.WithAttributes(
		attribute.String("limiter.driver", driver),
	)

	// 根据允许情况记录对应的计数器
	if allowed {
		if b.allowedCounter != nil {
			b.allowedCounter.Add(ctx, 1, attrs)
		}
	} else {
		if b.rejectedCounter != nil {
			b.rejectedCounter.Add(ctx, 1, attrs)
		}
	}
}

// sanitizeKey 对限流键进行脱敏处理
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

// ValidateKey 验证限流键是否有效
// 确保键不为空字符串
func ValidateKey(key string) error {
	if key == "" {
		return fmt.Errorf("limiter key cannot be empty")
	}
	return nil
}

// ValidateLimit 验证限流限制是否有效
func ValidateLimit(limit int) error {
	if limit <= 0 {
		return fmt.Errorf("limit must be greater than 0")
	}
	return nil
}

// ValidateWindow 验证时间窗口是否有效
func ValidateWindow(window time.Duration) error {
	if window <= 0 {
		return fmt.Errorf("window must be greater than 0")
	}
	return nil
}
