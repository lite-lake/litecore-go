// Package controllers 定义 HTTP 控制器
package controllers

import (
	"github.com/lite-lake/litecore-go/common"
	"github.com/lite-lake/litecore-go/samples/messageboard/internal/dtos"
	"github.com/lite-lake/litecore-go/samples/messageboard/internal/services"

	"github.com/gin-gonic/gin"
)

// IMsgAllController 获取所有留言控制器接口
type IMsgAllController interface {
	common.IBaseController
}

type msgAllControllerImpl struct {
	MessageService services.IMessageService `inject:""`
}

// NewMsgAllController 创建控制器实例
func NewMsgAllController() IMsgAllController {
	return &msgAllControllerImpl{}
}

func (c *msgAllControllerImpl) ControllerName() string {
	return "msgAllControllerImpl"
}

func (c *msgAllControllerImpl) GetRouter() string {
	return "/api/admin/messages [GET]"
}

func (c *msgAllControllerImpl) Handle(ctx *gin.Context) {
	messages, err := c.MessageService.GetAllMessages()
	if err != nil {
		ctx.JSON(common.HTTPStatusInternalServerError, dtos.ErrInternalServer)
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

	ctx.JSON(common.HTTPStatusOK, dtos.SuccessWithData(responseList))
}

var _ IMsgAllController = (*msgAllControllerImpl)(nil)
