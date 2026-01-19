// Package controllers 定义 HTTP 控制器
package controllers

import (
	"com.litelake.litecore/common"
	"com.litelake.litecore/samples/messageboard/internal/dtos"
	"com.litelake.litecore/samples/messageboard/internal/services"
	"strconv"

	"github.com/gin-gonic/gin"
)

// IMessageDeleteController 删除留言控制器接口
type IMessageDeleteController interface {
	common.BaseController
}

type MessageDeleteController struct {
	MessageService services.IMessageService `inject:""`
}

// NewMessageDeleteController 创建控制器实例
func NewMessageDeleteController() IMessageDeleteController {
	return &MessageDeleteController{}
}

func (c *MessageDeleteController) ControllerName() string {
	return "MessageDeleteController"
}

func (c *MessageDeleteController) GetRouter() string {
	return "/api/admin/messages/:id/delete [POST]"
}

func (c *MessageDeleteController) Handle(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		ctx.JSON(400, dtos.ErrorResponse(400, "无效的留言 ID"))
		return
	}

	if err := c.MessageService.DeleteMessage(uint(id)); err != nil {
		ctx.JSON(400, dtos.ErrorResponse(400, err.Error()))
		return
	}

	ctx.JSON(200, dtos.SuccessWithMessage("删除成功"))
}

var _ IMessageDeleteController = (*MessageDeleteController)(nil)
