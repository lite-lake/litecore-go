// Package controllers 定义 HTTP 控制器
package controllers

import (
	"github.com/gin-gonic/gin"
	"github.com/lite-lake/litecore-go/common"
	"github.com/lite-lake/litecore-go/samples/messageboard/internal/dtos"
	"github.com/lite-lake/litecore-go/samples/messageboard/internal/services"
)

// IMsgListController 获取留言控制器接口
type IMsgListController interface {
	common.IBaseController
}

type msgListControllerImpl struct {
	MessageService services.IMessageService `inject:""`
	Logger         common.ILogger           `inject:""`
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
	if c.Logger != nil {
		c.Logger.Debug("开始获取已审核留言列表")
	}

	messages, err := c.MessageService.GetApprovedMessages()
	if err != nil {
		if c.Logger != nil {
			c.Logger.Error("获取已审核留言失败", "error", err)
		}
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

	for i := range responseList {
		responseList[i].Status = ""
	}

	if c.Logger != nil {
		c.Logger.Info("获取已审核留言成功", "count", len(responseList))
	}

	ctx.JSON(common.HTTPStatusOK, dtos.SuccessWithData(responseList))
}

var _ IMsgListController = (*msgListControllerImpl)(nil)
