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

// IMsgDeleteController 删除留言控制器接口
type IMsgDeleteController interface {
	common.IBaseController
}

type msgDeleteControllerImpl struct {
	MessageService services.IMessageService `inject:""`
	LoggerMgr      loggermgr.ILoggerManager `inject:""`
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
		c.LoggerMgr.Ins().Error("删除留言失败：无效的留言 ID", "id_str", idStr, "error", err)
		ctx.JSON(common.HTTPStatusBadRequest, dtos.ErrorResponse(common.HTTPStatusBadRequest, "无效的留言 ID"))
		return
	}

	c.LoggerMgr.Ins().Debug("开始删除留言", "id", id)

	if err := c.MessageService.DeleteMessage(uint(id)); err != nil {
		c.LoggerMgr.Ins().Error("删除留言失败", "id", id, "error", err)
		ctx.JSON(common.HTTPStatusBadRequest, dtos.ErrorResponse(common.HTTPStatusBadRequest, err.Error()))
		return
	}

	c.LoggerMgr.Ins().Info("删除留言成功", "id", id)

	ctx.JSON(common.HTTPStatusOK, dtos.SuccessWithMessage("删除成功"))
}

var _ IMsgDeleteController = (*msgDeleteControllerImpl)(nil)
