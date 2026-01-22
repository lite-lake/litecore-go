// Package controllers 定义 HTTP 控制器
package controllers

import (
	"github.com/gin-gonic/gin"
	"github.com/lite-lake/litecore-go/common"
	"github.com/lite-lake/litecore-go/samples/messageboard/internal/dtos"
	"github.com/lite-lake/litecore-go/samples/messageboard/internal/services"
	"github.com/lite-lake/litecore-go/server/builtin/manager/loggermgr"
)

// IMsgCreateController 创建留言控制器接口
type IMsgCreateController interface {
	common.IBaseController
}

type msgCreateControllerImpl struct {
	MessageService services.IMessageService `inject:""`
	LoggerMgr      loggermgr.ILoggerManager `inject:""`
}

// NewMsgCreateController 创建控制器实例
func NewMsgCreateController() IMsgCreateController {
	return &msgCreateControllerImpl{}
}

func (c *msgCreateControllerImpl) ControllerName() string {
	return "msgCreateControllerImpl"
}

func (c *msgCreateControllerImpl) GetRouter() string {
	return "/api/messages [POST]"
}

func (c *msgCreateControllerImpl) Handle(ctx *gin.Context) {
	var req dtos.CreateMessageRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		c.LoggerMgr.Ins().Error("创建留言失败：参数绑定失败", "error", err)
		ctx.JSON(common.HTTPStatusBadRequest, dtos.ErrorResponse(common.HTTPStatusBadRequest, err.Error()))
		return
	}

	c.LoggerMgr.Ins().Debug("开始创建留言", "nickname", req.Nickname, "content_length", len(req.Content))

	message, err := c.MessageService.CreateMessage(req.Nickname, req.Content)
	if err != nil {
		c.LoggerMgr.Ins().Error("创建留言失败", "nickname", req.Nickname, "error", err)
		ctx.JSON(common.HTTPStatusBadRequest, dtos.ErrorResponse(common.HTTPStatusBadRequest, err.Error()))
		return
	}

	c.LoggerMgr.Ins().Info("创建留言成功", "id", message.ID, "nickname", message.Nickname)

	ctx.JSON(common.HTTPStatusOK, dtos.SuccessResponse("留言提交成功，等待审核", gin.H{
		"id": message.ID,
	}))
}

var _ IMsgCreateController = (*msgCreateControllerImpl)(nil)
