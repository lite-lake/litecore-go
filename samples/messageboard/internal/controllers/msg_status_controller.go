// Package controllers 定义 HTTP 控制器
package controllers

import (
	"github.com/gin-gonic/gin"
	"github.com/lite-lake/litecore-go/common"
	"github.com/lite-lake/litecore-go/manager/loggermgr"
	"github.com/lite-lake/litecore-go/samples/messageboard/internal/dtos"
	"github.com/lite-lake/litecore-go/samples/messageboard/internal/services"
)

// IMsgStatusController 更新留言状态控制器接口（管理员专用）
type IMsgStatusController interface {
	common.IBaseController
}

type msgStatusControllerImpl struct {
	MessageService services.IMessageService `inject:""` // 留言服务
	LoggerMgr      loggermgr.ILoggerManager `inject:""` // 日志管理器
}

// NewMsgStatusController 创建控制器实例
func NewMsgStatusController() IMsgStatusController {
	return &msgStatusControllerImpl{}
}

// ControllerName 返回控制器名称
func (c *msgStatusControllerImpl) ControllerName() string {
	return "msgStatusControllerImpl"
}

// GetRouter 返回路由信息
func (c *msgStatusControllerImpl) GetRouter() string {
	return "/api/admin/messages/:id/status [POST]"
}

// Handle 处理更新留言状态请求（管理员专用）
func (c *msgStatusControllerImpl) Handle(ctx *gin.Context) {
	id := ctx.Param("id")

	var req dtos.UpdateStatusRequest
	if err := ctx.ShouldBind(&req); err != nil {
		c.LoggerMgr.Ins().Error("Failed to update message status: parameter binding failed", "id", id, "error", err)
		ctx.JSON(common.HTTPStatusBadRequest, dtos.ErrorResponse(common.HTTPStatusBadRequest, err.Error()))
		return
	}

	c.LoggerMgr.Ins().Debug("Starting to update message status", "id", id, "status", req.Status)

	if err := c.MessageService.UpdateMessageStatus(id, req.Status); err != nil {
		c.LoggerMgr.Ins().Error("Failed to update message status", "id", id, "status", req.Status, "error", err)
		ctx.JSON(common.HTTPStatusBadRequest, dtos.ErrorResponse(common.HTTPStatusBadRequest, err.Error()))
		return
	}

	c.LoggerMgr.Ins().Info("Message status updated successfully", "id", id, "status", req.Status)

	ctx.JSON(common.HTTPStatusOK, dtos.SuccessWithMessage("状态更新成功"))
}

var _ IMsgStatusController = (*msgStatusControllerImpl)(nil)
