// Package controllers 定义 HTTP 控制器
package controllers

import (
	"com.litelake.litecore/common"
	"com.litelake.litecore/samples/messageboard/internal/dtos"
	"com.litelake.litecore/samples/messageboard/internal/services"
	"strconv"

	"github.com/gin-gonic/gin"
)

// IMsgStatusController 更新留言状态控制器接口
type IMsgStatusController interface {
	common.BaseController
}

type msgStatusControllerImpl struct {
	MessageService services.IMessageService `inject:""`
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
		ctx.JSON(400, dtos.ErrorResponse(400, "无效的留言 ID"))
		return
	}

	var req dtos.UpdateStatusRequest
	if err := ctx.ShouldBind(&req); err != nil {
		ctx.JSON(400, dtos.ErrorResponse(400, err.Error()))
		return
	}

	if err := c.MessageService.UpdateMessageStatus(uint(id), req.Status); err != nil {
		ctx.JSON(400, dtos.ErrorResponse(400, err.Error()))
		return
	}

	ctx.JSON(200, dtos.SuccessWithMessage("状态更新成功"))
}

var _ IMsgStatusController = (*msgStatusControllerImpl)(nil)
