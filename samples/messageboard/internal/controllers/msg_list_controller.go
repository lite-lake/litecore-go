// Package controllers 定义 HTTP 控制器
package controllers

import (
	"com.litelake.litecore/common"
	"com.litelake.litecore/samples/messageboard/internal/dtos"
	"com.litelake.litecore/samples/messageboard/internal/services"

	"github.com/gin-gonic/gin"
)

// IMsgListController 获取留言控制器接口
type IMsgListController interface {
	common.BaseController
}

type msgListControllerImpl struct {
	MessageService services.IMessageService `inject:""`
}

// NewMsgListController 创建控制器实例
func NewMsgListController() IMsgListController {
	return &msgListControllerImpl{}
}

func (c *msgListControllerImpl) ControllerName() string {
	return "msgListControllerImpl"
}

func (c *msgListControllerImpl) GetRouter() string {
	return "/api/messages [GET]"
}

func (c *msgListControllerImpl) Handle(ctx *gin.Context) {
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

var _ IMsgListController = (*msgListControllerImpl)(nil)
