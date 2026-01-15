// Package controllers 定义 HTTP 控制器
package controllers

import (
	"com.litelake.litecore/common"
	"com.litelake.litecore/samples/messageboard/internal/dtos"
	"com.litelake.litecore/samples/messageboard/internal/services"
	"strconv"

	"github.com/gin-gonic/gin"
)

// IUpdateStatusController 更新留言状态控制器接口
type IUpdateStatusController interface {
	common.BaseController
}

type UpdateStatusController struct {
	MessageService services.IMessageService `inject:""`
}

// NewUpdateStatusController 创建控制器实例
func NewUpdateStatusController() IUpdateStatusController {
	return &UpdateStatusController{}
}

func (c *UpdateStatusController) ControllerName() string {
	return "UpdateStatusController"
}

func (c *UpdateStatusController) GetRouter() string {
	return "/api/admin/messages/:id/status [POST]"
}

func (c *UpdateStatusController) Handle(ctx *gin.Context) {
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

var _ IUpdateStatusController = (*UpdateStatusController)(nil)
