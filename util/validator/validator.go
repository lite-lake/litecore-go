package validator

import (
	"reflect"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

// Validator 验证器接口
type Validator interface {
	Validate(ctx *gin.Context, obj interface{}) error
}

// DefaultValidator 默认验证器（基于 go-playground/validator）
type DefaultValidator struct {
	engine *validator.Validate
}

// NewDefaultValidator 创建默认验证器
func NewDefaultValidator() *DefaultValidator {
	v := validator.New()

	// 注册自定义标签名函数，使用 json 标签作为字段名
	v.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
		if name == "-" {
			return ""
		}
		return name
	})

	return &DefaultValidator{engine: v}
}

// RegisterValidation 注册自定义验证函数
// tag: 验证标签名(如 "complexPassword")
// fn: 验证函数,签名为 func(fl validator.FieldLevel) bool
func (v *DefaultValidator) RegisterValidation(tag string, fn validator.Func) error {
	return v.engine.RegisterValidation(tag, fn)
}

// Validate 验证请求
func (v *DefaultValidator) Validate(ctx *gin.Context, obj interface{}) error {
	if err := ctx.ShouldBindJSON(obj); err != nil {
		return err
	}

	if err := v.engine.Struct(obj); err != nil {
		return v.formatValidationError(err)
	}

	return nil
}

// formatValidationError 格式化验证错误
func (v *DefaultValidator) formatValidationError(err error) error {
	if validationErrors, ok := err.(validator.ValidationErrors); ok {
		var errMsgs []string
		for _, e := range validationErrors {
			field := e.Field()
			tag := e.Tag()
			param := e.Param()

			switch tag {
			case "required":
				errMsgs = append(errMsgs, field+" is required")
			case "min":
				errMsgs = append(errMsgs, field+" must be at least "+param+" characters")
			case "max":
				errMsgs = append(errMsgs, field+" must be at most "+param+" characters")
			case "email":
				errMsgs = append(errMsgs, field+" must be a valid email")
			case "complexPassword":
				errMsgs = append(errMsgs, field+
					" must contain: at least 12 characters, uppercase, lowercase, number and special character")
			default:
				errMsgs = append(errMsgs, field+" validation failed on "+tag)
			}
		}
		return &ValidationError{
			Message: strings.Join(errMsgs, "; "),
			Errors:  validationErrors,
		}
	}
	return err
}

// ValidationError 验证错误
type ValidationError struct {
	Message string
	Errors  validator.ValidationErrors
}

func (ve *ValidationError) Error() string {
	return ve.Message
}

// BindAndValidate 绑定并验证请求（泛型辅助函数）
func BindAndValidate[T any](ctx *gin.Context, validator Validator) (*T, error) {
	var req T
	if err := validator.Validate(ctx, &req); err != nil {
		return nil, err
	}
	return &req, nil
}
