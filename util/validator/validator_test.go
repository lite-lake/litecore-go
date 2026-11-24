package validator

import (
	"bytes"
	"encoding/json"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/assert"
)

func init() {
	// 设置 Gin 为测试模式
	gin.SetMode(gin.TestMode)
}

// TestNewDefaultValidator 测试创建默认验证器
func TestNewDefaultValidator(t *testing.T) {
	v := NewDefaultValidator()

	assert.NotNil(t, v)
	assert.NotNil(t, v.engine)
}

// TestDefaultValidator_Validate_Success 测试成功验证
func TestDefaultValidator_Validate_Success(t *testing.T) {
	v := NewDefaultValidator()

	// 创建测试上下文
	type TestRequest struct {
		Name  string `json:"name" validate:"required"`
		Email string `json:"email" validate:"required,email"`
	}

	reqBody := TestRequest{
		Name:  "John Doe",
		Email: "john@example.com",
	}

	bodyBytes, _ := json.Marshal(reqBody)

	// 创建 Gin 上下文
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("POST", "/test", bytes.NewBuffer(bodyBytes))
	c.Request.Header.Set("Content-Type", "application/json")

	// 验证
	var req TestRequest
	err := v.Validate(c, &req)

	assert.NoError(t, err)
	assert.Equal(t, "John Doe", req.Name)
	assert.Equal(t, "john@example.com", req.Email)
}

// TestDefaultValidator_Validate_RequiredField 测试必填字段验证
func TestDefaultValidator_Validate_RequiredField(t *testing.T) {
	v := NewDefaultValidator()

	type TestRequest struct {
		Name string `json:"name" validate:"required"`
	}

	tests := []struct {
		name    string
		reqBody string
		wantErr bool
		errMsg  string
	}{
		{
			name:    "Missing required field",
			reqBody: `{"name":""}`,
			wantErr: true,
			errMsg:  "name is required",
		},
		{
			name:    "Valid required field",
			reqBody: `{"name":"John"}`,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest("POST", "/test", bytes.NewBufferString(tt.reqBody))
			c.Request.Header.Set("Content-Type", "application/json")

			var req TestRequest
			err := v.Validate(c, &req)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestDefaultValidator_Validate_Email 测试邮箱验证
func TestDefaultValidator_Validate_Email(t *testing.T) {
	v := NewDefaultValidator()

	type TestRequest struct {
		Email string `json:"email" validate:"required,email"`
	}

	tests := []struct {
		name    string
		reqBody string
		wantErr bool
		errMsg  string
	}{
		{
			name:    "Invalid email format",
			reqBody: `{"email":"notanemail"}`,
			wantErr: true,
			errMsg:  "email must be a valid email",
		},
		{
			name:    "Empty email",
			reqBody: `{"email":""}`,
			wantErr: true,
			errMsg:  "email is required",
		},
		{
			name:    "Valid email",
			reqBody: `{"email":"test@example.com"}`,
			wantErr: false,
		},
		{
			name:    "Valid email with subdomain",
			reqBody: `{"email":"user@mail.example.com"}`,
			wantErr: false,
		},
		{
			name:    "Valid email with numbers",
			reqBody: `{"email":"user123@example.com"}`,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest("POST", "/test", bytes.NewBufferString(tt.reqBody))
			c.Request.Header.Set("Content-Type", "application/json")

			var req TestRequest
			err := v.Validate(c, &req)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestDefaultValidator_Validate_MinLength 测试最小长度验证
func TestDefaultValidator_Validate_MinLength(t *testing.T) {
	v := NewDefaultValidator()

	type TestRequest struct {
		Password string `json:"password" validate:"required,min=8"`
	}

	tests := []struct {
		name    string
		reqBody string
		wantErr bool
		errMsg  string
	}{
		{
			name:    "Password too short",
			reqBody: `{"password":"1234567"}`,
			wantErr: true,
			errMsg:  "password must be at least 8 characters",
		},
		{
			name:    "Password exactly min length",
			reqBody: `{"password":"12345678"}`,
			wantErr: false,
		},
		{
			name:    "Password longer than min",
			reqBody: `{"password":"123456789"}`,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest("POST", "/test", bytes.NewBufferString(tt.reqBody))
			c.Request.Header.Set("Content-Type", "application/json")

			var req TestRequest
			err := v.Validate(c, &req)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestDefaultValidator_Validate_MaxLength 测试最大长度验证
func TestDefaultValidator_Validate_MaxLength(t *testing.T) {
	v := NewDefaultValidator()

	type TestRequest struct {
		Name string `json:"name" validate:"required,max=10"`
	}

	tests := []struct {
		name    string
		reqBody string
		wantErr bool
		errMsg  string
	}{
		{
			name:    "Name too long",
			reqBody: `{"name":"John Doe Smith"}`,
			wantErr: true,
			errMsg:  "name must be at most 10 characters",
		},
		{
			name:    "Name exactly max length",
			reqBody: `{"name":"John Doe S"}`,
			wantErr: false,
		},
		{
			name:    "Name shorter than max",
			reqBody: `{"name":"John"}`,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest("POST", "/test", bytes.NewBufferString(tt.reqBody))
			c.Request.Header.Set("Content-Type", "application/json")

			var req TestRequest
			err := v.Validate(c, &req)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestDefaultValidator_Validate_MultipleErrors 测试多个验证错误
func TestDefaultValidator_Validate_MultipleErrors(t *testing.T) {
	v := NewDefaultValidator()

	type TestRequest struct {
		Name     string `json:"name" validate:"required,min=3"`
		Email    string `json:"email" validate:"required,email"`
		Password string `json:"password" validate:"required,min=8"`
	}

	reqBody := `{"name":"Jo","email":"invalid","password":"short"}`

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("POST", "/test", bytes.NewBufferString(reqBody))
	c.Request.Header.Set("Content-Type", "application/json")

	var req TestRequest
	err := v.Validate(c, &req)

	assert.Error(t, err)
	errMsg := err.Error()

	// 检查所有错误消息都包含在内
	assert.Contains(t, errMsg, "name")
	assert.Contains(t, errMsg, "email")
	assert.Contains(t, errMsg, "password")
}

// TestDefaultValidator_Validate_InvalidJSON 测试无效的 JSON
func TestDefaultValidator_Validate_InvalidJSON(t *testing.T) {
	v := NewDefaultValidator()

	type TestRequest struct {
		Name string `json:"name"`
	}

	reqBody := `{invalid json}`

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("POST", "/test", bytes.NewBufferString(reqBody))
	c.Request.Header.Set("Content-Type", "application/json")

	var req TestRequest
	err := v.Validate(c, &req)

	assert.Error(t, err)
}

// TestDefaultValidator_Validate_EmptyBody 测试空请求体
func TestDefaultValidator_Validate_EmptyBody(t *testing.T) {
	v := NewDefaultValidator()

	type TestRequest struct {
		Name string `json:"name" validate:"required"`
	}

	reqBody := `{}`

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("POST", "/test", bytes.NewBufferString(reqBody))
	c.Request.Header.Set("Content-Type", "application/json")

	var req TestRequest
	err := v.Validate(c, &req)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "required")
}

// TestDefaultValidator_Validate_JSONTagName 测试使用 json 标签作为字段名
func TestDefaultValidator_Validate_JSONTagName(t *testing.T) {
	v := NewDefaultValidator()

	type TestRequest struct {
		UserName  string `json:"user_name" validate:"required"`
		UserEmail string `json:"user_email" validate:"required,email"`
	}

	reqBody := `{"user_name":"","user_email":"invalid"}`

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("POST", "/test", bytes.NewBufferString(reqBody))
	c.Request.Header.Set("Content-Type", "application/json")

	var req TestRequest
	err := v.Validate(c, &req)

	assert.Error(t, err)
	errMsg := err.Error()

	// 应该使用 json 标签名称（user_name, user_email）而不是结构体字段名（UserName, UserEmail）
	assert.Contains(t, errMsg, "user_name")
	assert.Contains(t, errMsg, "user_email")
	assert.NotContains(t, errMsg, "UserName")
	assert.NotContains(t, errMsg, "UserEmail")
}

// TestDefaultValidator_Validate_OmitEmptyTag 测试 omitempty 标签
func TestDefaultValidator_Validate_OmitEmptyTag(t *testing.T) {
	v := NewDefaultValidator()

	type TestRequest struct {
		Name  string `json:"name" validate:"required"`
		Email string `json:"email,omitempty" validate:"omitempty,email"`
	}

	// 只提供必填字段
	reqBody := `{"name":"John"}`

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("POST", "/test", bytes.NewBufferString(reqBody))
	c.Request.Header.Set("Content-Type", "application/json")

	var req TestRequest
	err := v.Validate(c, &req)

	// Email 是可选的，所以应该验证成功
	assert.NoError(t, err)
	assert.Equal(t, "John", req.Name)
	assert.Equal(t, "", req.Email)
}

// TestValidationError_Error 测试 ValidationError 的 Error 方法
func TestValidationError_Error(t *testing.T) {
	ve := &ValidationError{
		Message: "validation failed",
		Errors:  nil,
	}

	assert.Equal(t, "validation failed", ve.Error())
}

// TestValidationError_CustomTag 测试自定义验证标签
func TestDefaultValidator_Validate_CustomTag(t *testing.T) {
	v := NewDefaultValidator()

	type TestRequest struct {
		Age int `json:"age" validate:"gte=18,lte=100"`
	}

	tests := []struct {
		name    string
		reqBody string
		wantErr bool
	}{
		{
			name:    "Age too young",
			reqBody: `{"age":17}`,
			wantErr: true,
		},
		{
			name:    "Valid age - minimum",
			reqBody: `{"age":18}`,
			wantErr: false,
		},
		{
			name:    "Valid age - middle",
			reqBody: `{"age":50}`,
			wantErr: false,
		},
		{
			name:    "Valid age - maximum",
			reqBody: `{"age":100}`,
			wantErr: false,
		},
		{
			name:    "Age too old",
			reqBody: `{"age":101}`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest("POST", "/test", bytes.NewBufferString(tt.reqBody))
			c.Request.Header.Set("Content-Type", "application/json")

			var req TestRequest
			err := v.Validate(c, &req)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestBindAndValidate 测试泛型辅助函数
func TestBindAndValidate(t *testing.T) {
	v := NewDefaultValidator()

	type TestRequest struct {
		Name  string `json:"name" validate:"required"`
		Email string `json:"email" validate:"required,email"`
	}

	reqBody := `{"name":"John","email":"john@example.com"}`

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("POST", "/test", bytes.NewBufferString(reqBody))
	c.Request.Header.Set("Content-Type", "application/json")

	// 使用泛型函数
	req, err := BindAndValidate[TestRequest](c, v)

	assert.NoError(t, err)
	assert.NotNil(t, req)
	assert.Equal(t, "John", req.Name)
	assert.Equal(t, "john@example.com", req.Email)
}

// TestBindAndValidate_Error 测试 BindAndValidate 错误情况
func TestBindAndValidate_Error(t *testing.T) {
	v := NewDefaultValidator()

	type TestRequest struct {
		Name  string `json:"name" validate:"required"`
		Email string `json:"email" validate:"required,email"`
	}

	reqBody := `{"name":"","email":"invalid"}`

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("POST", "/test", bytes.NewBufferString(reqBody))
	c.Request.Header.Set("Content-Type", "application/json")

	// 使用泛型函数
	req, err := BindAndValidate[TestRequest](c, v)

	assert.Error(t, err)
	assert.Nil(t, req)
}

// TestDefaultValidator_NestedStruct 测试嵌套结构体验证
func TestDefaultValidator_NestedStruct(t *testing.T) {
	v := NewDefaultValidator()

	type Address struct {
		Street string `json:"street" validate:"required"`
		City   string `json:"city" validate:"required"`
	}

	type TestRequest struct {
		Name    string  `json:"name" validate:"required"`
		Address Address `json:"address" validate:"required"`
	}

	reqBody := `{"name":"John","address":{"street":"","city":""}}`

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("POST", "/test", bytes.NewBufferString(reqBody))
	c.Request.Header.Set("Content-Type", "application/json")

	var req TestRequest
	err := v.Validate(c, &req)

	assert.Error(t, err)
	errMsg := err.Error()
	assert.Contains(t, errMsg, "street")
	assert.Contains(t, errMsg, "city")
}

// TestDefaultValidator_SliceValidation 测试切片验证
func TestDefaultValidator_SliceValidation(t *testing.T) {
	v := NewDefaultValidator()

	type TestRequest struct {
		Tags []string `json:"tags" validate:"required,min=1"`
	}

	tests := []struct {
		name    string
		reqBody string
		wantErr bool
	}{
		{
			name:    "Empty tags array",
			reqBody: `{"tags":[]}`,
			wantErr: true,
		},
		{
			name:    "Tags with one element",
			reqBody: `{"tags":["tag1"]}`,
			wantErr: false,
		},
		{
			name:    "Tags with multiple elements",
			reqBody: `{"tags":["tag1","tag2","tag3"]}`,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest("POST", "/test", bytes.NewBufferString(tt.reqBody))
			c.Request.Header.Set("Content-Type", "application/json")

			var req TestRequest
			err := v.Validate(c, &req)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestDefaultValidator_NumericValidation 测试数值验证
func TestDefaultValidator_NumericValidation(t *testing.T) {
	v := NewDefaultValidator()

	type TestRequest struct {
		Price float64 `json:"price" validate:"required,gt=0"`
		Qty   int     `json:"qty" validate:"required,gte=1"`
	}

	tests := []struct {
		name    string
		reqBody string
		wantErr bool
		errMsg  string
	}{
		{
			name:    "Zero price",
			reqBody: `{"price":0,"qty":1}`,
			wantErr: true,
			errMsg:  "price",
		},
		{
			name:    "Negative price",
			reqBody: `{"price":-10,"qty":1}`,
			wantErr: true,
			errMsg:  "price",
		},
		{
			name:    "Zero quantity",
			reqBody: `{"price":10,"qty":0}`,
			wantErr: true,
			errMsg:  "qty",
		},
		{
			name:    "Valid values",
			reqBody: `{"price":99.99,"qty":5}`,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest("POST", "/test", bytes.NewBufferString(tt.reqBody))
			c.Request.Header.Set("Content-Type", "application/json")

			var req TestRequest
			err := v.Validate(c, &req)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestDefaultValidator_ComplexScenario 测试复杂真实场景
func TestDefaultValidator_ComplexScenario(t *testing.T) {
	v := NewDefaultValidator()

	type CreateUserRequest struct {
		Name     string `json:"name" validate:"required,min=2,max=50"`
		Email    string `json:"email" validate:"required,email"`
		Password string `json:"password" validate:"required,min=8,max=100"`
		Age      int    `json:"age" validate:"required,gte=18,lte=120"`
	}

	tests := []struct {
		name    string
		reqBody string
		wantErr bool
	}{
		{
			name:    "Valid request",
			reqBody: `{"name":"John Doe","email":"john@example.com","password":"password123","age":25}`,
			wantErr: false,
		},
		{
			name:    "Name too short",
			reqBody: `{"name":"J","email":"john@example.com","password":"password123","age":25}`,
			wantErr: true,
		},
		{
			name:    "Invalid email",
			reqBody: `{"name":"John Doe","email":"invalid","password":"password123","age":25}`,
			wantErr: true,
		},
		{
			name:    "Password too short",
			reqBody: `{"name":"John Doe","email":"john@example.com","password":"pass","age":25}`,
			wantErr: true,
		},
		{
			name:    "Age too young",
			reqBody: `{"name":"John Doe","email":"john@example.com","password":"password123","age":15}`,
			wantErr: true,
		},
		{
			name:    "Multiple errors",
			reqBody: `{"name":"J","email":"invalid","password":"pass","age":15}`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest("POST", "/test", bytes.NewBufferString(tt.reqBody))
			c.Request.Header.Set("Content-Type", "application/json")

			var req CreateUserRequest
			err := v.Validate(c, &req)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, "John Doe", req.Name)
				assert.Equal(t, "john@example.com", req.Email)
				assert.Equal(t, "password123", req.Password)
				assert.Equal(t, 25, req.Age)
			}
		})
	}
}

// TestDefaultValidator_Interface 测试 Validator 接口实现
func TestDefaultValidator_Interface(t *testing.T) {
	v := NewDefaultValidator()

	// 确保实现了接口
	var _ Validator = v

	assert.NotNil(t, v)
}

// TestFormatValidationError 测试格式化验证错误
func TestFormatValidationError(t *testing.T) {
	v := NewDefaultValidator()

	// 测试不同的验证标签
	tests := []struct {
		name       string
		structData interface{}
		reqBody    string
		wantErr    bool
		checkMsg   string
	}{
		{
			name: "Required field error",
			structData: &struct {
				Name string `json:"name" validate:"required"`
			}{},
			reqBody:  `{"name":""}`,
			wantErr:  true,
			checkMsg: "is required",
		},
		{
			name: "Min length error",
			structData: &struct {
				Pass string `json:"pass" validate:"min=5"`
			}{},
			reqBody:  `{"pass":"123"}`,
			wantErr:  true,
			checkMsg: "must be at least",
		},
		{
			name: "Max length error",
			structData: &struct {
				Code string `json:"code" validate:"max=3"`
			}{},
			reqBody:  `{"code":"1234"}`,
			wantErr:  true,
			checkMsg: "must be at most",
		},
		{
			name: "Email error",
			structData: &struct {
				Email string `json:"email" validate:"email"`
			}{},
			reqBody:  `{"email":"notanemail"}`,
			wantErr:  true,
			checkMsg: "must be a valid email",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest("POST", "/test", bytes.NewBufferString(tt.reqBody))
			c.Request.Header.Set("Content-Type", "application/json")

			err := v.Validate(c, tt.structData)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.checkMsg)
			}
		})
	}
}

// TestRegisterValidation 测试 RegisterValidation 函数
func TestRegisterValidation(t *testing.T) {
	v := NewDefaultValidator()

	// 注册自定义验证函数 - 验证字符串只包含小写字母
	err := v.RegisterValidation("lowercase", func(fl validator.FieldLevel) bool {
		value := fl.Field().String()
		for _, c := range value {
			if c < 'a' || c > 'z' {
				return false
			}
		}
		return true
	})
	assert.NoError(t, err)

	// 测试自定义验证器
	type TestRequest struct {
		Username string `json:"username" validate:"required,lowercase"`
	}

	tests := []struct {
		name    string
		reqBody string
		wantErr bool
	}{
		{
			name:    "Valid lowercase",
			reqBody: `{"username":"john"}`,
			wantErr: false,
		},
		{
			name:    "Contains uppercase",
			reqBody: `{"username":"John"}`,
			wantErr: true,
		},
		{
			name:    "Contains numbers",
			reqBody: `{"username":"john123"}`,
			wantErr: true,
		},
		{
			name:    "Empty string",
			reqBody: `{"username":""}`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest("POST", "/test", bytes.NewBufferString(tt.reqBody))
			c.Request.Header.Set("Content-Type", "application/json")

			var req TestRequest
			err := v.Validate(c, &req)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestRegisterValidationMultiple 测试注册多个自定义验证器
func TestRegisterValidationMultiple(t *testing.T) {
	v := NewDefaultValidator()

	// 注册多个自定义验证器
	err := v.RegisterValidation("even", func(fl validator.FieldLevel) bool {
		value := fl.Field().Int()
		return value%2 == 0
	})
	assert.NoError(t, err)

	err = v.RegisterValidation("positive", func(fl validator.FieldLevel) bool {
		value := fl.Field().Int()
		return value > 0
	})
	assert.NoError(t, err)

	// 测试多个验证器
	type TestRequest struct {
		Number int `json:"number" validate:"required,even,positive"`
	}

	tests := []struct {
		name    string
		reqBody string
		wantErr bool
	}{
		{
			name:    "Valid even and positive",
			reqBody: `{"number":4}`,
			wantErr: false,
		},
		{
			name:    "Odd number",
			reqBody: `{"number":3}`,
			wantErr: true,
		},
		{
			name:    "Negative even",
			reqBody: `{"number":-2}`,
			wantErr: true,
		},
		{
			name:    "Zero",
			reqBody: `{"number":0}`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest("POST", "/test", bytes.NewBufferString(tt.reqBody))
			c.Request.Header.Set("Content-Type", "application/json")

			var req TestRequest
			err := v.Validate(c, &req)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
