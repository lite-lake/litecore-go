package request

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

// mockValidator 模拟验证器实现
type mockValidator struct {
	validateFunc func(ctx *gin.Context, obj interface{}) error
}

func (m *mockValidator) Validate(ctx *gin.Context, obj interface{}) error {
	if m.validateFunc != nil {
		return m.validateFunc(ctx, obj)
	}
	return nil
}

// testRequest 测试请求结构体
type testRequest struct {
	Name string `json:"name"`
	Age  int    `json:"age"`
}

func TestSetDefaultValidator(t *testing.T) {
	tests := []struct {
		name      string
		validator ValidatorInterface
		want      ValidatorInterface
	}{
		{
			name:      "设置成功",
			validator: &mockValidator{},
			want:      &mockValidator{},
		},
		{
			name:      "设置nil",
			validator: nil,
			want:      nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			SetDefaultValidator(tt.validator)
			got := GetDefaultValidator()
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestGetDefaultValidator(t *testing.T) {
	tests := []struct {
		name  string
		setup func()
		want  ValidatorInterface
	}{
		{
			name:  "获取已设置的验证器",
			setup: func() { SetDefaultValidator(&mockValidator{}) },
			want:  &mockValidator{},
		},
		{
			name:  "获取未设置的验证器返回nil",
			setup: func() { SetDefaultValidator(nil) },
			want:  nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			got := GetDefaultValidator()
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestBindRequest(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name      string
		setup     func()
		request   func() *http.Request
		want      *testRequest
		wantErr   bool
		errString string
	}{
		{
			name: "成功绑定请求",
			setup: func() {
				SetDefaultValidator(&mockValidator{
					validateFunc: func(ctx *gin.Context, obj interface{}) error {
						// 模拟从JSON绑定
						if req, ok := obj.(*testRequest); ok {
							req.Name = "test"
							req.Age = 25
						}
						return nil
					},
				})
			},
			request: func() *http.Request {
				req := httptest.NewRequest("POST", "/test", nil)
				req.Header.Set("Content-Type", "application/json")
				return req
			},
			want: &testRequest{
				Name: "test",
				Age:  25,
			},
			wantErr: false,
		},
		{
			name: "验证器未设置返回错误",
			setup: func() {
				SetDefaultValidator(nil)
			},
			request: func() *http.Request {
				req := httptest.NewRequest("POST", "/test", nil)
				return req
			},
			want:      nil,
			wantErr:   true,
			errString: "validator not set",
		},
		{
			name: "验证失败返回错误",
			setup: func() {
				SetDefaultValidator(&mockValidator{
					validateFunc: func(ctx *gin.Context, obj interface{}) error {
						return errors.New("validation failed")
					},
				})
			},
			request: func() *http.Request {
				req := httptest.NewRequest("POST", "/test", nil)
				return req
			},
			want:      nil,
			wantErr:   true,
			errString: "validation failed",
		},
		{
			name: "验证器返回空错误",
			setup: func() {
				SetDefaultValidator(&mockValidator{
					validateFunc: func(ctx *gin.Context, obj interface{}) error {
						return nil
					},
				})
			},
			request: func() *http.Request {
				req := httptest.NewRequest("GET", "/test", nil)
				return req
			},
			want:    &testRequest{},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			w := httptest.NewRecorder()
			ctx, _ := gin.CreateTestContext(w)
			ctx.Request = tt.request()

			got, err := BindRequest[testRequest](ctx)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errString != "" {
					assert.Contains(t, err.Error(), tt.errString)
				}
				assert.Nil(t, got)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, got)
				if tt.want != nil {
					assert.Equal(t, tt.want.Name, got.Name)
					assert.Equal(t, tt.want.Age, got.Age)
				}
			}
		})
	}
}

func TestBindRequest_EdgeCases(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name      string
		request   func() *http.Request
		setup     func()
		wantErr   bool
		errString string
	}{
		{
			name: "nil context处理",
			request: func() *http.Request {
				return httptest.NewRequest("POST", "/test", nil)
			},
			setup: func() {
				SetDefaultValidator(&mockValidator{
					validateFunc: func(ctx *gin.Context, obj interface{}) error {
						if ctx == nil {
							return errors.New("context is nil")
						}
						return nil
					},
				})
			},
			wantErr: false,
		},
		{
			name: "空请求体",
			request: func() *http.Request {
				return httptest.NewRequest("POST", "/test", nil)
			},
			setup: func() {
				SetDefaultValidator(&mockValidator{
					validateFunc: func(ctx *gin.Context, obj interface{}) error {
						return nil
					},
				})
			},
			wantErr: false,
		},
		{
			name: "大请求体",
			request: func() *http.Request {
				return httptest.NewRequest("POST", "/test", nil)
			},
			setup: func() {
				SetDefaultValidator(&mockValidator{
					validateFunc: func(ctx *gin.Context, obj interface{}) error {
						return nil
					},
				})
			},
			wantErr: false,
		},
		{
			name: "多次调用BindRequest",
			request: func() *http.Request {
				return httptest.NewRequest("POST", "/test", nil)
			},
			setup: func() {
				SetDefaultValidator(&mockValidator{
					validateFunc: func(ctx *gin.Context, obj interface{}) error {
						return nil
					},
				})
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			w := httptest.NewRecorder()
			ctx, _ := gin.CreateTestContext(w)
			ctx.Request = tt.request()

			// 多次调用测试
			for i := 0; i < 3; i++ {
				got, err := BindRequest[testRequest](ctx)
				if tt.wantErr {
					assert.Error(t, err)
					if tt.errString != "" {
						assert.Contains(t, err.Error(), tt.errString)
					}
					assert.Nil(t, got)
				} else {
					assert.NoError(t, err)
					assert.NotNil(t, got)
				}
			}
		})
	}
}

func TestBindRequest_DifferentTypes(t *testing.T) {
	gin.SetMode(gin.TestMode)

	type simpleRequest struct {
		ID int
	}
	type complexRequest struct {
		ID       int
		Name     string
		Email    string
		IsActive bool
	}

	tests := []struct {
		name    string
		setup   func()
		request func() *http.Request
	}{
		{
			name: "绑定简单结构体",
			setup: func() {
				SetDefaultValidator(&mockValidator{
					validateFunc: func(ctx *gin.Context, obj interface{}) error {
						if req, ok := obj.(*simpleRequest); ok {
							req.ID = 123
						}
						return nil
					},
				})
			},
			request: func() *http.Request {
				return httptest.NewRequest("POST", "/test", nil)
			},
		},
		{
			name: "绑定复杂结构体",
			setup: func() {
				SetDefaultValidator(&mockValidator{
					validateFunc: func(ctx *gin.Context, obj interface{}) error {
						if req, ok := obj.(*complexRequest); ok {
							req.ID = 456
							req.Name = "test"
							req.Email = "test@example.com"
							req.IsActive = true
						}
						return nil
					},
				})
			},
			request: func() *http.Request {
				return httptest.NewRequest("POST", "/test", nil)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			w := httptest.NewRecorder()
			ctx, _ := gin.CreateTestContext(w)
			ctx.Request = tt.request()

			switch tt.name {
			case "绑定简单结构体":
				got, err := BindRequest[simpleRequest](ctx)
				assert.NoError(t, err)
				assert.NotNil(t, got)
				assert.Equal(t, 123, got.ID)
			case "绑定复杂结构体":
				got, err := BindRequest[complexRequest](ctx)
				assert.NoError(t, err)
				assert.NotNil(t, got)
				assert.Equal(t, 456, got.ID)
				assert.Equal(t, "test", got.Name)
				assert.Equal(t, "test@example.com", got.Email)
				assert.True(t, got.IsActive)
			}
		})
	}
}

// BenchmarkBindRequest 性能基准测试
func BenchmarkBindRequest(b *testing.B) {
	gin.SetMode(gin.TestMode)
	SetDefaultValidator(&mockValidator{
		validateFunc: func(ctx *gin.Context, obj interface{}) error {
			if req, ok := obj.(*testRequest); ok {
				req.Name = "benchmark"
				req.Age = 30
			}
			return nil
		},
	})

	req := httptest.NewRequest("POST", "/test", nil)
	req.Header.Set("Content-Type", "application/json")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		ctx, _ := gin.CreateTestContext(w)
		ctx.Request = req
		_, _ = BindRequest[testRequest](ctx)
	}
}

// BenchmarkGetDefaultValidator 获取验证器基准测试
func BenchmarkGetDefaultValidator(b *testing.B) {
	SetDefaultValidator(&mockValidator{})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = GetDefaultValidator()
	}
}

// BenchmarkSetDefaultValidator 设置验证器基准测试
func BenchmarkSetDefaultValidator(b *testing.B) {
	validator := &mockValidator{}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		SetDefaultValidator(validator)
	}
}

// ExampleValidatorInterface 验证器接口示例实现
type ExampleValidatorInterface struct{}

func (e *ExampleValidatorInterface) Validate(ctx *gin.Context, obj interface{}) error {
	return nil
}

func TestValidatorInterface_Implementation(t *testing.T) {
	var v ValidatorInterface = &ExampleValidatorInterface{}
	assert.NotNil(t, v)
	assert.Implements(t, (*ValidatorInterface)(nil), v)
}
