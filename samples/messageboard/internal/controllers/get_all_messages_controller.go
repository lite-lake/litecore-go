// Package controllers 定义 HTTP 控制器
package controllers

import (
	"com.litelake.litecore/common"
	"com.litelake.litecore/samples/messageboard/internal/dtos"
	"com.litelake.litecore/samples/messageboard/internal/services"

	"github.com/gin-gonic/gin"
)

// IGetAllMessagesController 获取所有留言控制器接口
type IGetAllMessagesController interface {
	common.BaseController
}

type GetAllMessagesController struct {
	MessageService services.IMessageService `inject:""`
}

// NewGetAllMessagesController 创建控制器实例
func NewGetAllMessagesController() IGetAllMessagesController {
	return &GetAllMessagesController{}
}

func (c *GetAllMessagesController) ControllerName() string {
	return "GetAllMessagesController"
}

func (c *GetAllMessagesController) GetRouter() string {
	return "/api/admin/messages [GET]"
}

func (c *GetAllMessagesController) Handle(ctx *gin.Context) {
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

var _ IGetAllMessagesController = (*GetAllMessagesController)(nil)
