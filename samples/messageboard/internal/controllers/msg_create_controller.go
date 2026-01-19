// Package controllers 定义 HTTP 控制器
package controllers

import (
	"com.litelake.litecore/common"
	"com.litelake.litecore/samples/messageboard/internal/dtos"
	"com.litelake.litecore/samples/messageboard/internal/services"

	"github.com/gin-gonic/gin"
)

// IMsgCreateController 创建留言控制器接口
type IMsgCreateController interface {
	common.BaseController
}

type msgCreateControllerImpl struct {
	MessageService services.IMessageService `inject:""`
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
		ctx.JSON(400, dtos.ErrorResponse(400, err.Error()))
		return
	}

	message, err := c.MessageService.CreateMessage(req.Nickname, req.Content)
	if err != nil {
		ctx.JSON(400, dtos.ErrorResponse(400, err.Error()))
		return
	}

	ctx.JSON(200, dtos.SuccessResponse("留言提交成功，等待审核", gin.H{
		"id": message.ID,
	}))
}

var _ IMsgCreateController = (*msgCreateControllerImpl)(nil)
