package mqmgr

import (
	"context"
	"fmt"
	"github.com/lite-lake/litecore-go/manager/loggermgr"
	"github.com/lite-lake/litecore-go/manager/telemetrymgr"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"
)

// mqManagerBaseImpl 消息队列管理器基础实现
// 提供可观测性（日志、指标、链路追踪）和工具函数
type mqManagerBaseImpl struct {
	// loggerMgr 日志管理器，用于记录日志
	loggerMgr loggermgr.ILoggerManager
	// telemetryMgr 遥测管理器，用于指标和链路追踪
	telemetryMgr telemetrymgr.ITelemetryManager
	// tracer 链路追踪器，用于记录操作链路
	tracer trace.Tracer
	// meter 指标记录器，用于记录性能指标
	meter metric.Meter
	// publishCounter 消息发布计数器
	publishCounter metric.Int64Counter
	// consumeCounter 消息消费计数器
	consumeCounter metric.Int64Counter
	// ackCounter 消息确认计数器
	ackCounter metric.Int64Counter
	// nackCounter 消息拒绝计数器
	nackCounter metric.Int64Counter
	// operationDuration 操作耗时直方图
	operationDuration metric.Float64Histogram
}

// newMqManagerBaseImpl 创建消息队列管理器基础实现
// 参数：
//   - loggerMgr: 日志管理器
//   - telemetryMgr: 遥测管理器
func newMqManagerBaseImpl(
	loggerMgr loggermgr.ILoggerManager,
	telemetryMgr telemetrymgr.ITelemetryManager,
) *mqManagerBaseImpl {
	return &mqManagerBaseImpl{
		loggerMgr:    loggerMgr,
		telemetryMgr: telemetryMgr,
	}
}

// initObservability 初始化可观测性组件
// 在依赖注入完成后调用，用于初始化链路追踪器和指标收集器
func (b *mqManagerBaseImpl) initObservability() {
	if b.telemetryMgr == nil {
		return
	}

	// 初始化 tracer，用于链路追踪
	b.tracer = b.telemetryMgr.Tracer("mqmgr")

	// 初始化 meter，用于指标记录
	b.meter = b.telemetryMgr.Meter("mqmgr")

	// 创建消息发布计数器指标
	b.publishCounter, _ = b.meter.Int64Counter(
		"mq.publish",
		metric.WithDescription("消息发布数量"),
		metric.WithUnit("{message}"),
	)

	// 创建消息消费计数器指标
	b.consumeCounter, _ = b.meter.Int64Counter(
		"mq.consume",
		metric.WithDescription("消息消费数量"),
		metric.WithUnit("{message}"),
	)

	// 创建消息确认计数器指标
	b.ackCounter, _ = b.meter.Int64Counter(
		"mq.ack",
		metric.WithDescription("消息确认数量"),
		metric.WithUnit("{message}"),
	)

	// 创建消息拒绝计数器指标
	b.nackCounter, _ = b.meter.Int64Counter(
		"mq.nack",
		metric.WithDescription("消息拒绝数量"),
		metric.WithUnit("{message}"),
	)

	// 创建操作耗时直方图指标
	b.operationDuration, _ = b.meter.Float64Histogram(
		"mq.operation.duration",
		metric.WithDescription("消息队列操作耗时（秒）"),
		metric.WithUnit("s"),
	)
}

// recordOperation 记录操作并执行
// 封装了链路追踪、指标记录、日志记录等功能
// 参数：
//   - ctx: 上下文
//   - driver: 消息队列驱动类型（rabbitmq、memory）
//   - operation: 操作名称（publish、consume、ack 等）
//   - queue: 队列名称
//   - fn: 要执行的操作函数
func (b *mqManagerBaseImpl) recordOperation(
	ctx context.Context,
	driver string,
	operation string,
	queue string,
	fn func() error,
) error {
	// 如果没有配置任何可观测性组件，直接执行操作
	if b.tracer == nil && b.loggerMgr == nil && b.operationDuration == nil {
		return fn()
	}

	var span trace.Span
	// 创建链路追踪 span
	if b.tracer != nil {
		ctx, span = b.tracer.Start(ctx, "mq."+operation,
			trace.WithAttributes(
				attribute.String("mq.queue", queue),
				attribute.String("mq.driver", driver),
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
			logger.Error("mq operation failed",
				"operation", operation,
				"queue", queue,
				"error", err.Error(),
				"duration", duration,
			)
			if span != nil {
				span.RecordError(err)
				span.SetStatus(codes.Error, err.Error())
			}
		} else {
			// 操作成功，记录调试日志
			logger.Debug("mq operation success",
				"operation", operation,
				"queue", queue,
				"duration", duration,
			)
		}
	}

	return err
}

// recordPublish 记录消息发布指标
// 参数：
//   - ctx: 上下文
//   - driver: 消息队列驱动类型
func (b *mqManagerBaseImpl) recordPublish(ctx context.Context, driver string) {
	if b.meter == nil {
		return
	}

	if b.publishCounter != nil {
		b.publishCounter.Add(ctx, 1,
			metric.WithAttributes(
				attribute.String("mq.driver", driver),
			),
		)
	}
}

// recordConsume 记录消息消费指标
// 参数：
//   - ctx: 上下文
//   - driver: 消息队列驱动类型
func (b *mqManagerBaseImpl) recordConsume(ctx context.Context, driver string) {
	if b.meter == nil {
		return
	}

	if b.consumeCounter != nil {
		b.consumeCounter.Add(ctx, 1,
			metric.WithAttributes(
				attribute.String("mq.driver", driver),
			),
		)
	}
}

// recordAck 记录消息确认指标
// 参数：
//   - ctx: 上下文
//   - driver: 消息队列驱动类型
func (b *mqManagerBaseImpl) recordAck(ctx context.Context, driver string) {
	if b.meter == nil {
		return
	}

	if b.ackCounter != nil {
		b.ackCounter.Add(ctx, 1,
			metric.WithAttributes(
				attribute.String("mq.driver", driver),
			),
		)
	}
}

// recordNack 记录消息拒绝指标
// 参数：
//   - ctx: 上下文
//   - driver: 消息队列驱动类型
func (b *mqManagerBaseImpl) recordNack(ctx context.Context, driver string) {
	if b.meter == nil {
		return
	}

	if b.nackCounter != nil {
		b.nackCounter.Add(ctx, 1,
			metric.WithAttributes(
				attribute.String("mq.driver", driver),
			),
		)
	}
}

// ValidateContext 验证上下文是否有效
// 确保传入的 context 不为 nil
func ValidateContext(ctx context.Context) error {
	if ctx == nil {
		return fmt.Errorf("context cannot be nil")
	}
	return nil
}

// ValidateQueue 验证队列名是否有效
// 确保队列名不为空字符串
func ValidateQueue(queue string) error {
	if queue == "" {
		return fmt.Errorf("queue name cannot be empty")
	}
	return nil
}

// getStatus 根据错误返回状态字符串
// 用于指标记录和日志分类
func getStatus(err error) string {
	if err != nil {
		return "error"
	}
	return "success"
}
