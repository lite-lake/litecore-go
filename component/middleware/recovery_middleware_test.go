package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"

	"com.litelake.litecore/component/manager/loggermgr"
)

type mockLogger struct {
	debugMsgs []string
	infoMsgs  []string
	warnMsgs  []string
	errorMsgs []string
	fatalMsgs []string
}

func (m *mockLogger) Debug(msg string, args ...any) {
	m.debugMsgs = append(m.debugMsgs, msg)
}

func (m *mockLogger) Info(msg string, args ...any) {
	m.infoMsgs = append(m.infoMsgs, msg)
}

func (m *mockLogger) Warn(msg string, args ...any) {
	m.warnMsgs = append(m.warnMsgs, msg)
}

func (m *mockLogger) Error(msg string, args ...any) {
	m.errorMsgs = append(m.errorMsgs, msg)
}

func (m *mockLogger) Fatal(msg string, args ...any) {
	m.fatalMsgs = append(m.fatalMsgs, msg)
}

func (m *mockLogger) With(args ...any) loggermgr.ILogger {
	return m
}

func (m *mockLogger) SetLevel(level loggermgr.LogLevel) {
}

type mockLoggerManager struct {
	loggers map[string]loggermgr.ILogger
}

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

func (m *mockLoggerManager) Logger(name string) loggermgr.ILogger {
	if m.loggers == nil {
		m.loggers = make(map[string]loggermgr.ILogger)
	}
	logger, ok := m.loggers[name]
	if !ok {
		logger = &mockLogger{}
		m.loggers[name] = logger
	}
	return logger
}

func (m *mockLoggerManager) SetGlobalLevel(level loggermgr.LogLevel) {
}

func (m *mockLoggerManager) Shutdown(ctx context.Context) error {
	return nil
}

type fatalMockLogger struct {
	*mockLogger
}

func TestNewRecoveryMiddleware(t *testing.T) {
	tests := []struct {
		name string
	}{
		{
			name: "创建Recovery中间件",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			middleware := NewRecoveryMiddleware()
			assert.NotNil(t, middleware)
			assert.IsType(t, &RecoveryMiddleware{}, middleware)
		})
	}
}

func TestRecoveryMiddleware_MiddlewareName(t *testing.T) {
	tests := []struct {
		name     string
		expected string
	}{
		{
			name:     "返回中间件名称",
			expected: "RecoveryMiddleware",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			middleware := NewRecoveryMiddleware().(*RecoveryMiddleware)
			assert.Equal(t, tt.expected, middleware.MiddlewareName())
		})
	}
}

func TestRecoveryMiddleware_Order(t *testing.T) {
	tests := []struct {
		name     string
		expected int
	}{
		{
			name:     "返回执行顺序",
			expected: 10,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			middleware := NewRecoveryMiddleware().(*RecoveryMiddleware)
			assert.Equal(t, tt.expected, middleware.Order())
		})
	}
}

func TestRecoveryMiddleware_Wrapper(t *testing.T) {
	tests := []struct {
		name           string
		panicValue     interface{}
		hasPanic       bool
		expectedStatus int
		withLogger     bool
		withRequestID  bool
	}{
		{
			name:           "正常请求",
			panicValue:     nil,
			hasPanic:       false,
			expectedStatus: http.StatusOK,
			withLogger:     true,
		},
		{
			name:           "字符串panic恢复",
			panicValue:     "test panic",
			hasPanic:       true,
			expectedStatus: http.StatusInternalServerError,
			withLogger:     true,
		},
		{
			name:           "错误panic恢复",
			panicValue:     assert.AnError,
			hasPanic:       true,
			expectedStatus: http.StatusInternalServerError,
			withLogger:     true,
		},
		{
			name:           "无日志管理器的panic恢复",
			panicValue:     "test panic",
			hasPanic:       true,
			expectedStatus: http.StatusInternalServerError,
			withLogger:     false,
		},
		{
			name:           "带有请求ID的panic恢复",
			panicValue:     "test panic",
			hasPanic:       true,
			expectedStatus: http.StatusInternalServerError,
			withLogger:     true,
			withRequestID:  true,
		},
		{
			name:           "整数panic恢复",
			panicValue:     12345,
			hasPanic:       true,
			expectedStatus: http.StatusInternalServerError,
			withLogger:     true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gin.SetMode(gin.TestMode)
			router := gin.New()

			middleware := NewRecoveryMiddleware().(*RecoveryMiddleware)
			if tt.withLogger {
				middleware.LoggerManager = &mockLoggerManager{}
			}
			router.Use(middleware.Wrapper())

			router.GET("/test", func(c *gin.Context) {
				if tt.hasPanic {
					panic(tt.panicValue)
				}
				c.JSON(http.StatusOK, gin.H{"message": "ok"})
			})

			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			if tt.withRequestID {
				req.Header.Set("X-Request-ID", "test-request-id")
			}
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.hasPanic && tt.withLogger {
				body := w.Body.String()
				assert.Contains(t, body, "error")
				assert.Contains(t, body, "INTERNAL_SERVER_ERROR")
			}
		})
	}
}

func TestRecoveryMiddleware_OnStart(t *testing.T) {
	tests := []struct {
		name string
	}{
		{
			name: "服务器启动回调",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			middleware := NewRecoveryMiddleware().(*RecoveryMiddleware)
			err := middleware.OnStart()
			assert.NoError(t, err)
		})
	}
}

func TestRecoveryMiddleware_OnStop(t *testing.T) {
	tests := []struct {
		name string
	}{
		{
			name: "服务器停止回调",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			middleware := NewRecoveryMiddleware().(*RecoveryMiddleware)
			err := middleware.OnStop()
			assert.NoError(t, err)
		})
	}
}

func TestRecoveryMiddleware_PanicWithContext(t *testing.T) {
	tests := []struct {
		name           string
		handler        func(*gin.Context)
		expectedStatus int
	}{
		{
			name: "panic后返回错误响应",
			handler: func(c *gin.Context) {
				panic("unexpected error")
			},
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name: "panic后被捕获且请求终止",
			handler: func(c *gin.Context) {
				panic("test")
			},
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name: "正常处理流程",
			handler: func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{"status": "success"})
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "panic在嵌套调用中",
			handler: func(c *gin.Context) {
				nestedFunc := func() {
					panic("nested panic")
				}
				nestedFunc()
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gin.SetMode(gin.TestMode)
			router := gin.New()

			middleware := NewRecoveryMiddleware().(*RecoveryMiddleware)
			middleware.LoggerManager = &mockLoggerManager{}
			router.Use(middleware.Wrapper())

			router.GET("/test", tt.handler)

			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedStatus == http.StatusInternalServerError {
				body := w.Body.String()
				assert.Contains(t, body, "内部服务器错误")
			}
		})
	}
}
