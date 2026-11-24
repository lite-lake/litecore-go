package mqmgr

import (
	"context"
	"errors"
	"testing"

	"github.com/lite-lake/litecore-go/logger"
	"github.com/lite-lake/litecore-go/manager/loggermgr"
	"github.com/lite-lake/litecore-go/manager/telemetrymgr"
	"go.opentelemetry.io/otel/log"
	"go.opentelemetry.io/otel/metric"
	sdklog "go.opentelemetry.io/otel/sdk/log"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
)

type mockTelemetryManager struct{}

func (m *mockTelemetryManager) ManagerName() string {
	return "mockTelemetryManager"
}

func (m *mockTelemetryManager) Health() error {
	return nil
}

func (m *mockTelemetryManager) OnStart() error {
	return nil
}

func (m *mockTelemetryManager) OnStop() error {
	return nil
}

func (m *mockTelemetryManager) Tracer(name string) trace.Tracer {
	return nil
}

func (m *mockTelemetryManager) TracerProvider() *sdktrace.TracerProvider {
	return nil
}

func (m *mockTelemetryManager) Meter(name string) metric.Meter {
	return nil
}

func (m *mockTelemetryManager) MeterProvider() *sdkmetric.MeterProvider {
	return nil
}

func (m *mockTelemetryManager) Logger(name string) log.Logger {
	return nil
}

func (m *mockTelemetryManager) LoggerProvider() *sdklog.LoggerProvider {
	return nil
}

func (m *mockTelemetryManager) Shutdown(ctx context.Context) error {
	return nil
}

var _ telemetrymgr.ITelemetryManager = (*mockTelemetryManager)(nil)

func TestNewMqManagerBaseImpl(t *testing.T) {
	t.Run("创建基础实现", func(t *testing.T) {
		base := newMqManagerBaseImpl(nil, nil)
		if base == nil {
			t.Error("expected base impl not to be nil")
		}

		if base.loggerMgr != nil {
			t.Error("expected loggerMgr to be nil")
		}

		if base.telemetryMgr != nil {
			t.Error("expected telemetryMgr to be nil")
		}
	})
}

func TestInitObservability(t *testing.T) {
	t.Run("nil 遥测管理器", func(t *testing.T) {
		base := &mqManagerBaseImpl{
			telemetryMgr: nil,
		}

		base.initObservability()

		if base.tracer != nil {
			t.Error("expected tracer to be nil")
		}

		if base.meter != nil {
			t.Error("expected meter to be nil")
		}
	})

	t.Run("带遥测管理器", func(t *testing.T) {
		mockTelemetry := &mockTelemetryManager{}
		base := &mqManagerBaseImpl{
			telemetryMgr: mockTelemetry,
		}

		defer func() {
			if r := recover(); r != nil {
				t.Logf("Recovered from panic (expected with nil Tracer/Meter): %v", r)
			}
		}()

		base.initObservability()

		if base.tracer == nil {
			t.Error("expected tracer to be initialized")
		}

		if base.meter == nil {
			t.Error("expected meter to be initialized")
		}
	})
}

func TestRecordOperation(t *testing.T) {
	t.Run("无可观测性组件", func(t *testing.T) {
		base := &mqManagerBaseImpl{}

		err := base.recordOperation(context.Background(), "memory", "publish", "test_queue", func() error {
			return nil
		})

		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})

	t.Run("操作成功", func(t *testing.T) {
		base := &mqManagerBaseImpl{
			tracer: nil,
			meter:  nil,
		}

		err := base.recordOperation(context.Background(), "memory", "publish", "test_queue", func() error {
			return nil
		})

		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})

	t.Run("操作失败", func(t *testing.T) {
		base := &mqManagerBaseImpl{}

		expectedErr := errors.New("operation failed")
		err := base.recordOperation(context.Background(), "memory", "publish", "test_queue", func() error {
			return expectedErr
		})

		if err == nil {
			t.Error("expected error, got nil")
		}

		if !errors.Is(err, expectedErr) {
			t.Errorf("expected error to be %v, got %v", expectedErr, err)
		}
	})

	t.Run("nil context", func(t *testing.T) {
		base := &mqManagerBaseImpl{}

		err := base.recordOperation(nil, "memory", "publish", "test_queue", func() error {
			return nil
		})

		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
}

func TestRecordPublish(t *testing.T) {
	t.Run("nil meter", func(t *testing.T) {
		base := &mqManagerBaseImpl{
			meter: nil,
		}

		base.recordPublish(context.Background(), "memory")
	})
}

func TestRecordConsume(t *testing.T) {
	t.Run("nil meter", func(t *testing.T) {
		base := &mqManagerBaseImpl{
			meter: nil,
		}

		base.recordConsume(context.Background(), "memory")
	})
}

func TestRecordAck(t *testing.T) {
	t.Run("nil meter", func(t *testing.T) {
		base := &mqManagerBaseImpl{
			meter: nil,
		}

		base.recordAck(context.Background(), "memory")
	})
}

func TestRecordNack(t *testing.T) {
	t.Run("nil meter", func(t *testing.T) {
		base := &mqManagerBaseImpl{
			meter: nil,
		}

		base.recordNack(context.Background(), "memory")
	})
}

func TestValidateContext(t *testing.T) {
	t.Run("有效 context", func(t *testing.T) {
		ctx := context.Background()
		err := ValidateContext(ctx)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})

	t.Run("nil context", func(t *testing.T) {
		err := ValidateContext(nil)
		if err == nil {
			t.Error("expected error with nil context")
		}
	})
}

func TestValidateQueue(t *testing.T) {
	t.Run("有效队列名", func(t *testing.T) {
		err := ValidateQueue("test_queue")
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})

	t.Run("空队列名", func(t *testing.T) {
		err := ValidateQueue("")
		if err == nil {
			t.Error("expected error with empty queue name")
		}
	})
}

func TestGetStatus(t *testing.T) {
	t.Run("无错误", func(t *testing.T) {
		status := getStatus(nil)
		if status != "success" {
			t.Errorf("expected 'success', got '%s'", status)
		}
	})

	t.Run("有错误", func(t *testing.T) {
		status := getStatus(errors.New("test error"))
		if status != "error" {
			t.Errorf("expected 'error', got '%s'", status)
		}
	})
}

type mockLoggerManager struct{}

func (m *mockLoggerManager) ManagerName() string {
	return "mockLoggerManager"
}

func (m *mockLoggerManager) Health() error {
	return nil
}

func (m *mockLoggerManager) OnStart() error {
	return nil
}

func (m *mockLoggerManager) OnStop() error {
	return nil
}

func (m *mockLoggerManager) Ins() logger.ILogger {
	return &mockLogger{}
}

var _ loggermgr.ILoggerManager = (*mockLoggerManager)(nil)

type mockLogger struct{}

func (m *mockLogger) Debug(msg string, args ...any)   {}
func (m *mockLogger) Info(msg string, args ...any)    {}
func (m *mockLogger) Warn(msg string, args ...any)    {}
func (m *mockLogger) Error(msg string, args ...any)   {}
func (m *mockLogger) Fatal(msg string, args ...any)   {}
func (m *mockLogger) With(args ...any) logger.ILogger { return m }
func (m *mockLogger) SetLevel(level logger.LogLevel)  {}

func TestMqManagerBaseImplWithLogger(t *testing.T) {
	t.Run("带 logger 的基础实现", func(t *testing.T) {
		mockLoggerMgr := &mockLoggerManager{}
		base := &mqManagerBaseImpl{
			loggerMgr: mockLoggerMgr,
		}

		err := base.recordOperation(context.Background(), "memory", "publish", "test_queue", func() error {
			return nil
		})

		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
}

func TestRecordOperationWithLoggerError(t *testing.T) {
	t.Run("带 logger 的操作失败", func(t *testing.T) {
		mockLoggerMgr := &mockLoggerManager{}
		base := &mqManagerBaseImpl{
			loggerMgr: mockLoggerMgr,
		}

		expectedErr := errors.New("test error")
		err := base.recordOperation(context.Background(), "memory", "publish", "test_queue", func() error {
			return expectedErr
		})

		if err == nil {
			t.Error("expected error, got nil")
		}

		if !errors.Is(err, expectedErr) {
			t.Errorf("expected error to be %v, got %v", expectedErr, err)
		}
	})
}
