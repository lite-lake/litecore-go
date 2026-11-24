package request

import (
	"errors"

	"github.com/gin-gonic/gin"
)

// ValidatorInterface 验证器接口
// 应用层需要实现此接口以提供具体的验证逻辑
type ValidatorInterface interface {
	Validate(ctx *gin.Context, obj interface{}) error
}

// defaultValidator 默认验证器实例（包级变量，可测试时替换）
// 应用层应该在使用前设置此变量
var defaultValidator ValidatorInterface

// SetDefaultValidator 设置默认验证器
// 应用层应该在初始化时调用此方法设置验证器
func SetDefaultValidator(validator ValidatorInterface) {
	defaultValidator = validator
}

// GetDefaultValidator 获取默认验证器
func GetDefaultValidator() ValidatorInterface {
	return defaultValidator
}

// BindRequest 绑定并验证请求（泛型辅助函数）
// 这是一个便捷方法，用于在 Controller 中快速绑定和验证请求参数
//
// 使用示例：
//
//	req, err := request.BindRequest[dtos.ArticleCreateRequest](ctx)
//	if err != nil {
//	    ctx.Error(apperrors.BadRequest("invalid request").Wrap(err))
//	    return
//	}
func BindRequest[T any](ctx *gin.Context) (*T, error) {
	if defaultValidator == nil {
		return nil, errors.New("validator not set, please call SetDefaultValidator first")
	}
	var req T
	if err := defaultValidator.Validate(ctx, &req); err != nil {
		return nil, err
	}
	return &req, nil
}
