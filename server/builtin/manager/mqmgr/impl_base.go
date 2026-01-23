package mqmgr

import (
	"context"
	"fmt"
	"time"

	"github.com/lite-lake/litecore-go/logger"
	"github.com/lite-lake/litecore-go/server/builtin/manager/telemetrymgr"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"
)

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

func newMqManagerBaseImpl() *mqManagerBaseImpl {
	return &mqManagerBaseImpl{}
}

func (b *mqManagerBaseImpl) initObservability() {
	if b.telemetryMgr == nil {
		return
	}

	b.tracer = b.telemetryMgr.Tracer("mqmgr")
	b.meter = b.telemetryMgr.Meter("mqmgr")

	b.publishCounter, _ = b.meter.Int64Counter(
		"mq.publish",
		metric.WithDescription("消息发布次数"),
		metric.WithUnit("{message}"),
	)

	b.consumeCounter, _ = b.meter.Int64Counter(
		"mq.consume",
		metric.WithDescription("消息消费次数"),
		metric.WithUnit("{message}"),
	)

	b.ackCounter, _ = b.meter.Int64Counter(
		"mq.ack",
		metric.WithDescription("消息确认次数"),
		metric.WithUnit("{message}"),
	)

	b.nackCounter, _ = b.meter.Int64Counter(
		"mq.nack",
		metric.WithDescription("消息拒绝次数"),
		metric.WithUnit("{message}"),
	)

	b.operationDuration, _ = b.meter.Float64Histogram(
		"mq.operation.duration",
		metric.WithDescription("消息队列操作耗时（秒）"),
		metric.WithUnit("s"),
	)
}

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

func ValidateContext(ctx context.Context) error {
	if ctx == nil {
		return fmt.Errorf("context cannot be nil")
	}
	return nil
}

func ValidateQueue(queue string) error {
	if queue == "" {
		return fmt.Errorf("queue name cannot be empty")
	}
	return nil
}

func getStatus(err error) string {
	if err != nil {
		return "error"
	}
	return "success"
}
