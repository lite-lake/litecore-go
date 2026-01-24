package mqmgr

import (
	"context"
	"fmt"
	"github.com/lite-lake/litecore-go/manager/telemetrymgr"
	"time"

	"github.com/lite-lake/litecore-go/logger"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"
)

// mqManagerBaseImpl 消息队列管理器基础实现
type mqManagerBaseImpl struct {
	Logger            logger.ILogger                 `inject:""`
	telemetryMgr      telemetrymgr.ITelemetryManager `inject:""`
	tracer            trace.Tracer
	meter             metric.Meter
	publishCounter    metric.Int64Counter
	consumeCounter    metric.Int64Counter
	ackCounter        metric.Int64Counter
	nackCounter       metric.Int64Counter
	operationDuration metric.Float64Histogram
}

// newMqManagerBaseImpl 创建消息队列管理器基础实现
func newMqManagerBaseImpl() *mqManagerBaseImpl {
	return &mqManagerBaseImpl{}
}

// initObservability 初始化可观测性组件
func (b *mqManagerBaseImpl) initObservability() {
	if b.telemetryMgr == nil {
		return
	}

	b.tracer = b.telemetryMgr.Tracer("mqmgr")
	b.meter = b.telemetryMgr.Meter("mqmgr")

	b.publishCounter, _ = b.meter.Int64Counter(
		"mq.publish",
		metric.WithDescription("Number of messages published"),
		metric.WithUnit("{message}"),
	)

	b.consumeCounter, _ = b.meter.Int64Counter(
		"mq.consume",
		metric.WithDescription("Number of messages consumed"),
		metric.WithUnit("{message}"),
	)

	b.ackCounter, _ = b.meter.Int64Counter(
		"mq.ack",
		metric.WithDescription("Number of messages acknowledged"),
		metric.WithUnit("{message}"),
	)

	b.nackCounter, _ = b.meter.Int64Counter(
		"mq.nack",
		metric.WithDescription("Number of messages rejected"),
		metric.WithUnit("{message}"),
	)

	b.operationDuration, _ = b.meter.Float64Histogram(
		"mq.operation.duration",
		metric.WithDescription("Duration of message queue operations in seconds"),
		metric.WithUnit("s"),
	)
}

// recordOperation 记录操作并执行
func (b *mqManagerBaseImpl) recordOperation(
	ctx context.Context,
	driver string,
	operation string,
	queue string,
	fn func() error,
) error {
	if b.tracer == nil && b.Logger == nil && b.operationDuration == nil {
		return fn()
	}

	var span trace.Span
	if b.tracer != nil {
		ctx, span = b.tracer.Start(ctx, "mq."+operation,
			trace.WithAttributes(
				attribute.String("mq.queue", queue),
				attribute.String("mq.driver", driver),
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

	if b.Logger != nil {
		if err != nil {
			b.Logger.Error("mq operation failed",
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
			b.Logger.Debug("mq operation success",
				"operation", operation,
				"queue", queue,
				"duration", duration,
			)
		}
	}

	return err
}

// recordPublish 记录消息发布
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

// recordConsume 记录消息消费
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

// recordAck 记录消息确认
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

// recordNack 记录消息拒绝
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
func ValidateContext(ctx context.Context) error {
	if ctx == nil {
		return fmt.Errorf("context cannot be nil")
	}
	return nil
}

// ValidateQueue 验证队列名是否有效
func ValidateQueue(queue string) error {
	if queue == "" {
		return fmt.Errorf("queue name cannot be empty")
	}
	return nil
}

// getStatus 获取操作状态
func getStatus(err error) string {
	if err != nil {
		return "error"
	}
	return "success"
}
