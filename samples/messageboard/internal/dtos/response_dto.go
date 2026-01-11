// Package dtos 定义统一响应格式
package dtos

// CommonResponse 统一响应结构
type CommonResponse struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// SuccessResponse 成功响应
func SuccessResponse(message string, data interface{}) CommonResponse {
	return CommonResponse{
		Code:    200,
		Message: message,
		Data:    data,
	}
}

// SuccessWithData 成功响应（仅数据）
func SuccessWithData(data interface{}) CommonResponse {
	return CommonResponse{
		Code:    200,
		Message: "success",
		Data:    data,
	}
}

// SuccessWithMessage 成功响应（仅消息）
func SuccessWithMessage(message string) CommonResponse {
	return CommonResponse{
		Code:    200,
		Message: message,
	}
}

// ErrorResponse 错误响应
func ErrorResponse(code int, message string) CommonResponse {
	return CommonResponse{
		Code:    code,
		Message: message,
	}
}

// 常用错误响应
var (
	ErrBadRequest     = ErrorResponse(400, "请求参数错误")
	ErrUnauthorized   = ErrorResponse(401, "未授权")
	ErrForbidden      = ErrorResponse(403, "禁止访问")
	ErrNotFound       = ErrorResponse(404, "资源不存在")
	ErrInternalServer = ErrorResponse(500, "服务器内部错误")
)
