// Package controllers 定义 HTTP 控制器
package controllers

import (
	"github.com/lite-lake/litecore-go/common"
	"github.com/lite-lake/litecore-go/samples/messageboard/internal/dtos"
	"github.com/lite-lake/litecore-go/samples/messageboard/internal/services"
	"github.com/lite-lake/litecore-go/util/logger"

	"github.com/gin-gonic/gin"
)

// IAdminAuthController 管理员登录控制器接口
type IAdminAuthController interface {
	common.IBaseController
}

type adminAuthControllerImpl struct {
	AuthService services.IAuthService `inject:""`
	Logger      logger.ILogger        `inject:""`
}

// NewAdminAuthController 创建控制器实例
func NewAdminAuthController() IAdminAuthController {
	return &adminAuthControllerImpl{}
}

func (c *adminAuthControllerImpl) ControllerName() string {
	return "adminAuthControllerImpl"
}

func (c *adminAuthControllerImpl) GetRouter() string {
	return "/api/admin/login [POST]"
}

func (c *adminAuthControllerImpl) Handle(ctx *gin.Context) {
	var req dtos.LoginRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		if c.Logger != nil {
			c.Logger.Error("管理员登录失败：参数绑定失败", "error", err)
		}
		ctx.JSON(common.HTTPStatusBadRequest, dtos.ErrBadRequest)
		return
	}

	if c.Logger != nil {
		c.Logger.Debug("开始管理员登录")
	}

	token, err := c.AuthService.Login(req.Password)
	if err != nil {
		if c.Logger != nil {
			c.Logger.Warn("管理员登录失败：密码错误")
		}
		ctx.JSON(common.HTTPStatusUnauthorized, dtos.ErrorResponse(common.HTTPStatusUnauthorized, "管理员密码错误"))
		return
	}

	if c.Logger != nil {
		c.Logger.Info("管理员登录成功", "token", token)
	}

	ctx.JSON(common.HTTPStatusOK, dtos.SuccessWithData(dtos.LoginResponse{
		Token: token,
	}))
}

var _ IAdminAuthController = (*adminAuthControllerImpl)(nil)
