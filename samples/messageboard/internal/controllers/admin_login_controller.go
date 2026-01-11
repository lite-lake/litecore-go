// Package controllers 定义 HTTP 控制器
package controllers

import (
	"com.litelake.litecore/common"
	"com.litelake.litecore/samples/messageboard/internal/dtos"
	"com.litelake.litecore/samples/messageboard/internal/services"

	"github.com/gin-gonic/gin"
)

// AdminLoginController 管理员登录控制器
type AdminLoginController struct {
	AuthService *services.AuthService `inject:""`
}

// NewAdminLoginController 创建控制器实例
func NewAdminLoginController() *AdminLoginController {
	return &AdminLoginController{}
}

// ControllerName 实现 BaseController 接口
func (c *AdminLoginController) ControllerName() string {
	return "AdminLoginController"
}

// GetRouter 实现 BaseController 接口
func (c *AdminLoginController) GetRouter() string {
	return "/api/admin/login [POST]"
}

// Handle 实现 BaseController 接口
func (c *AdminLoginController) Handle(ctx *gin.Context) {
	// 绑定请求参数
	var req dtos.LoginRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(400, dtos.ErrBadRequest)
		return
	}

	// 验证密码并创建会话
	token, err := c.AuthService.Login(req.Password)
	if err != nil {
		ctx.JSON(401, dtos.ErrorResponse(401, "管理员密码错误"))
		return
	}

	// 返回令牌
	ctx.JSON(200, dtos.SuccessWithData(dtos.LoginResponse{
		Token: token,
	}))
}

var _ common.BaseController = (*AdminLoginController)(nil)
