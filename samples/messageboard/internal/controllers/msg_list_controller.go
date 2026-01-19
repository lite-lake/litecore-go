// Package controllers 定义 HTTP 控制器
package controllers

import (
	"com.litelake.litecore/common"
	"com.litelake.litecore/samples/messageboard/internal/dtos"
	"com.litelake.litecore/samples/messageboard/internal/services"

	"github.com/gin-gonic/gin"
)

// IMessageListController 获取留言控制器接口
type IMessageListController interface {
	common.BaseController
}

type MessageListController struct {
	MessageService services.IMessageService `inject:""`
}

// NewMessageListController 创建控制器实例
func NewMessageListController() IMessageListController {
	return &MessageListController{}
}

func (c *MessageListController) ControllerName() string {
	return "MessageListController"
}

func (c *MessageListController) GetRouter() string {
	return "/api/messages [GET]"
}

func (c *MessageListController) Handle(ctx *gin.Context) {
	messages, err := c.MessageService.GetApprovedMessages()
	if err != nil {
		ctx.JSON(500, dtos.ErrInternalServer)
		return
	}

	responseList := make([]dtos.MessageResponse, 0, len(messages))
	for _, msg := range messages {
		responseList = append(responseList, dtos.ToMessageResponse(
			msg.ID,
			msg.Nickname,
			msg.Content,
			msg.Status,
			msg.CreatedAt,
		))
	}

	for i := range responseList {
		responseList[i].Status = ""
	}

	ctx.JSON(200, dtos.SuccessWithData(responseList))
}

var _ IMessageListController = (*MessageListController)(nil)
