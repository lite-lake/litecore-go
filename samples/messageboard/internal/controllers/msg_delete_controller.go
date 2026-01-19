// Package controllers 定义 HTTP 控制器
package controllers

import (
	"com.litelake.litecore/common"
	"com.litelake.litecore/samples/messageboard/internal/dtos"
	"com.litelake.litecore/samples/messageboard/internal/services"
	"strconv"

	"github.com/gin-gonic/gin"
)

// IMsgDeleteController 删除留言控制器接口
type IMsgDeleteController interface {
	common.IBaseController
}

type msgDeleteControllerImpl struct {
	MessageService services.IMessageService `inject:""`
}

// NewMsgDeleteController 创建控制器实例
func NewMsgDeleteController() IMsgDeleteController {
	return &msgDeleteControllerImpl{}
}

func (c *msgDeleteControllerImpl) ControllerName() string {
	return "msgDeleteControllerImpl"
}

func (c *msgDeleteControllerImpl) GetRouter() string {
	return "/api/admin/messages/:id/delete [POST]"
}

func (c *msgDeleteControllerImpl) Handle(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		ctx.JSON(common.HTTPStatusBadRequest, dtos.ErrorResponse(common.HTTPStatusBadRequest, "无效的留言 ID"))
		return
	}

	if err := c.MessageService.DeleteMessage(uint(id)); err != nil {
		ctx.JSON(common.HTTPStatusBadRequest, dtos.ErrorResponse(common.HTTPStatusBadRequest, err.Error()))
		return
	}

	ctx.JSON(common.HTTPStatusOK, dtos.SuccessWithMessage("删除成功"))
}

var _ IMsgDeleteController = (*msgDeleteControllerImpl)(nil)
