// Package controllers 定义 HTTP 控制器
package controllers

import (
	"github.com/lite-lake/litecore-go/common"
	"github.com/lite-lake/litecore-go/samples/messageboard/internal/dtos"
	"github.com/lite-lake/litecore-go/samples/messageboard/internal/services"
	"github.com/lite-lake/litecore-go/server/builtin/manager/loggermgr"
	"strconv"

	"github.com/gin-gonic/gin"
)

// IMsgStatusController 更新留言状态控制器接口
type IMsgStatusController interface {
	common.IBaseController
}

type msgStatusControllerImpl struct {
	MessageService services.IMessageService `inject:""`
	LoggerMgr      loggermgr.ILoggerManager `inject:""`
}

// NewMsgStatusController 创建控制器实例
func NewMsgStatusController() IMsgStatusController {
	return &msgStatusControllerImpl{}
}

func (c *msgStatusControllerImpl) ControllerName() string {
	return "msgStatusControllerImpl"
}

func (c *msgStatusControllerImpl) GetRouter() string {
	return "/api/admin/messages/:id/status [POST]"
}

func (c *msgStatusControllerImpl) Handle(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.LoggerMgr.Ins().Error("更新留言状态失败：无效的留言 ID", "id_str", idStr, "error", err)
		ctx.JSON(common.HTTPStatusBadRequest, dtos.ErrorResponse(common.HTTPStatusBadRequest, "无效的留言 ID"))
		return
	}

	var req dtos.UpdateStatusRequest
	if err := ctx.ShouldBind(&req); err != nil {
		c.LoggerMgr.Ins().Error("更新留言状态失败：参数绑定失败", "id", id, "error", err)
		ctx.JSON(common.HTTPStatusBadRequest, dtos.ErrorResponse(common.HTTPStatusBadRequest, err.Error()))
		return
	}

	c.LoggerMgr.Ins().Debug("开始更新留言状态", "id", id, "status", req.Status)

	if err := c.MessageService.UpdateMessageStatus(uint(id), req.Status); err != nil {
		c.LoggerMgr.Ins().Error("更新留言状态失败", "id", id, "status", req.Status, "error", err)
		ctx.JSON(common.HTTPStatusBadRequest, dtos.ErrorResponse(common.HTTPStatusBadRequest, err.Error()))
		return
	}

	c.LoggerMgr.Ins().Info("更新留言状态成功", "id", id, "status", req.Status)

	ctx.JSON(common.HTTPStatusOK, dtos.SuccessWithMessage("状态更新成功"))
}

var _ IMsgStatusController = (*msgStatusControllerImpl)(nil)
