// Package controllers 定义 HTTP 控制器
package controllers

import (
	"com.litelake.litecore/common"
	"com.litelake.litecore/samples/messageboard/internal/dtos"
	"com.litelake.litecore/samples/messageboard/internal/services"

	"github.com/gin-gonic/gin"
)

// IAdminLoginController 管理员登录控制器接口
type IAdminLoginController interface {
	common.BaseController
}

type AdminLoginController struct {
	AuthService services.IAuthService `inject:""`
}

// NewAdminLoginController 创建控制器实例
func NewAdminLoginController() IAdminLoginController {
	return &AdminLoginController{}
}

func (c *AdminLoginController) ControllerName() string {
	return "AdminLoginController"
}

func (c *AdminLoginController) GetRouter() string {
	return "/api/admin/login [POST]"
}

func (c *AdminLoginController) Handle(ctx *gin.Context) {
	var req dtos.LoginRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(400, dtos.ErrBadRequest)
		return
	}

	token, err := c.AuthService.Login(req.Password)
	if err != nil {
		ctx.JSON(401, dtos.ErrorResponse(401, "管理员密码错误"))
		return
	}

	ctx.JSON(200, dtos.SuccessWithData(dtos.LoginResponse{
		Token: token,
	}))
}

var _ IAdminLoginController = (*AdminLoginController)(nil)
