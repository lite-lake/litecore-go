// Package controllers 定义 HTTP 控制器
package controllers

import (
	"com.litelake.litecore/common"
	"com.litelake.litecore/samples/messageboard/internal/dtos"
	"com.litelake.litecore/samples/messageboard/internal/services"
	"strconv"

	"github.com/gin-gonic/gin"
)

// UpdateStatusController 更新留言状态控制器
type UpdateStatusController struct {
	MessageService *services.MessageService `inject:""`
}

// NewUpdateStatusController 创建控制器实例
func NewUpdateStatusController() *UpdateStatusController {
	return &UpdateStatusController{}
}

// ControllerName 实现 BaseController 接口
func (c *UpdateStatusController) ControllerName() string {
	return "UpdateStatusController"
}

// GetRouter 实现 BaseController 接口
func (c *UpdateStatusController) GetRouter() string {
	return "/api/admin/messages/:id/status [POST]"
}

// Handle 实现 BaseController 接口
func (c *UpdateStatusController) Handle(ctx *gin.Context) {
	// 获取留言 ID
	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		ctx.JSON(400, dtos.ErrorResponse(400, "无效的留言 ID"))
		return
	}

	// 绑定请求参数
	var req dtos.UpdateStatusRequest
	if err := ctx.ShouldBind(&req); err != nil {
		ctx.JSON(400, dtos.ErrorResponse(400, err.Error()))
		return
	}

	// 更新状态
	if err := c.MessageService.UpdateMessageStatus(uint(id), req.Status); err != nil {
		ctx.JSON(400, dtos.ErrorResponse(400, err.Error()))
		return
	}

	ctx.JSON(200, dtos.SuccessWithMessage("状态更新成功"))
}

var _ common.BaseController = (*UpdateStatusController)(nil)
