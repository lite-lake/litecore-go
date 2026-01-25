// Package controllers 定义 HTTP 控制器
package controllers

import (
	"github.com/gin-gonic/gin"
	"github.com/lite-lake/litecore-go/common"
	"github.com/lite-lake/litecore-go/manager/loggermgr"
	"github.com/lite-lake/litecore-go/samples/messageboard/internal/dtos"
	"github.com/lite-lake/litecore-go/samples/messageboard/internal/services"
)

// IMsgDeleteController 删除留言控制器接口（管理员专用）
type IMsgDeleteController interface {
	common.IBaseController
}

type msgDeleteControllerImpl struct {
	MessageService services.IMessageService `inject:""` // 留言服务
	LoggerMgr      loggermgr.ILoggerManager `inject:""` // 日志管理器
}

// NewMsgDeleteController 创建控制器实例
func NewMsgDeleteController() IMsgDeleteController {
	return &msgDeleteControllerImpl{}
}

// ControllerName 返回控制器名称
func (c *msgDeleteControllerImpl) ControllerName() string {
	return "msgDeleteControllerImpl"
}

// GetRouter 返回路由信息
func (c *msgDeleteControllerImpl) GetRouter() string {
	return "/api/admin/messages/:id/delete [POST]"
}

// Handle 处理删除留言请求（管理员专用）
func (c *msgDeleteControllerImpl) Handle(ctx *gin.Context) {
	id := ctx.Param("id")

	c.LoggerMgr.Ins().Debug("Starting to delete message", "id", id)

	if err := c.MessageService.DeleteMessage(id); err != nil {
		c.LoggerMgr.Ins().Error("Failed to delete message", "id", id, "error", err)
		ctx.JSON(common.HTTPStatusBadRequest, dtos.ErrorResponse(common.HTTPStatusBadRequest, err.Error()))
		return
	}

	c.LoggerMgr.Ins().Info("Message deleted successfully", "id", id)

	ctx.JSON(common.HTTPStatusOK, dtos.SuccessWithMessage("删除成功"))
}

var _ IMsgDeleteController = (*msgDeleteControllerImpl)(nil)
