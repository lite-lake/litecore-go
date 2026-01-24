// Package controllers 定义 HTTP 控制器
package controllers

import (
	"github.com/gin-gonic/gin"
	"github.com/lite-lake/litecore-go/common"
	"github.com/lite-lake/litecore-go/manager/loggermgr"
	"github.com/lite-lake/litecore-go/samples/messageboard/internal/dtos"
	"github.com/lite-lake/litecore-go/samples/messageboard/internal/services"
)

// IAdminAuthController 管理员登录控制器接口
type IAdminAuthController interface {
	common.IBaseController
}

type adminAuthControllerImpl struct {
	AuthService services.IAuthService    `inject:""` // 认证服务
	LoggerMgr   loggermgr.ILoggerManager `inject:""` // 日志管理器
}

// NewAdminAuthController 创建控制器实例
func NewAdminAuthController() IAdminAuthController {
	return &adminAuthControllerImpl{}
}

// ControllerName 返回控制器名称
func (c *adminAuthControllerImpl) ControllerName() string {
	return "adminAuthControllerImpl"
}

// GetRouter 返回路由信息
func (c *adminAuthControllerImpl) GetRouter() string {
	return "/api/admin/login [POST]"
}

// Handle 处理管理员登录请求
func (c *adminAuthControllerImpl) Handle(ctx *gin.Context) {
	var req dtos.LoginRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		c.LoggerMgr.Ins().Error("Admin login failed: parameter binding error", "error", err)
		ctx.JSON(common.HTTPStatusBadRequest, dtos.ErrBadRequest)
		return
	}

	c.LoggerMgr.Ins().Debug("Starting admin login")

	token, err := c.AuthService.Login(req.Password)
	if err != nil {
		c.LoggerMgr.Ins().Warn("Admin login failed: incorrect password")
		ctx.JSON(common.HTTPStatusUnauthorized, dtos.ErrorResponse(common.HTTPStatusUnauthorized, "管理员密码错误"))
		return
	}

	c.LoggerMgr.Ins().Info("Admin login successful", "token", token)

	ctx.JSON(common.HTTPStatusOK, dtos.SuccessWithData(dtos.LoginResponse{
		Token: token,
	}))
}

var _ IAdminAuthController = (*adminAuthControllerImpl)(nil)
