// Package controllers 定义 HTTP 控制器
package controllers

import (
	"com.litelake.litecore/common"
	"com.litelake.litecore/samples/messageboard/internal/dtos"
	"com.litelake.litecore/samples/messageboard/internal/services"
	"strconv"

	"github.com/gin-gonic/gin"
)

// DeleteMessageController 删除留言控制器
type DeleteMessageController struct {
	MessageService *services.MessageService `inject:""`
}

// NewDeleteMessageController 创建控制器实例
func NewDeleteMessageController() *DeleteMessageController {
	return &DeleteMessageController{}
}

// ControllerName 实现 BaseController 接口
func (c *DeleteMessageController) ControllerName() string {
	return "DeleteMessageController"
}

// GetRouter 实现 BaseController 接口
func (c *DeleteMessageController) GetRouter() string {
	return "/api/admin/messages/:id/delete [POST]"
}

// Handle 实现 BaseController 接口
func (c *DeleteMessageController) Handle(ctx *gin.Context) {
	// 获取留言 ID
	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		ctx.JSON(400, dtos.ErrorResponse(400, "无效的留言 ID"))
		return
	}

	// 删除留言
	if err := c.MessageService.DeleteMessage(uint(id)); err != nil {
		ctx.JSON(400, dtos.ErrorResponse(400, err.Error()))
		return
	}

	ctx.JSON(200, dtos.SuccessWithMessage("删除成功"))
}

var _ common.BaseController = (*DeleteMessageController)(nil)
