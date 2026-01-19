// Package controllers 定义 HTTP 控制器
package controllers

import (
	"com.litelake.litecore/common"
	"com.litelake.litecore/samples/messageboard/internal/dtos"
	"com.litelake.litecore/samples/messageboard/internal/services"

	"github.com/gin-gonic/gin"
)

// IMessageAllController 获取所有留言控制器接口
type IMessageAllController interface {
	common.BaseController
}

type MessageAllController struct {
	MessageService services.IMessageService `inject:""`
}

// NewMessageAllController 创建控制器实例
func NewMessageAllController() IMessageAllController {
	return &MessageAllController{}
}

func (c *MessageAllController) ControllerName() string {
	return "MessageAllController"
}

func (c *MessageAllController) GetRouter() string {
	return "/api/admin/messages [GET]"
}

func (c *MessageAllController) Handle(ctx *gin.Context) {
	messages, err := c.MessageService.GetAllMessages()
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

	ctx.JSON(200, dtos.SuccessWithData(responseList))
}

var _ IMessageAllController = (*MessageAllController)(nil)
