// Package controllers 定义 HTTP 控制器
package controllers

import (
	"github.com/gin-gonic/gin"

	"github.com/lite-lake/litecore-go/common"
	"github.com/lite-lake/litecore-go/manager/loggermgr"
	"github.com/lite-lake/litecore-go/samples/messageboard/internal/dtos"
	"github.com/lite-lake/litecore-go/samples/messageboard/internal/services"
)

// IMsgListController 获取留言控制器接口
type IMsgListController interface {
	common.IBaseController
}

type msgListControllerImpl struct {
	MessageService services.IMessageService `inject:""` // 留言服务
	LoggerMgr      loggermgr.ILoggerManager `inject:""` // 日志管理器
}

// NewMsgListController 创建控制器实例
func NewMsgListController() IMsgListController {
	return &msgListControllerImpl{}
}

// ControllerName 返回控制器名称
func (c *msgListControllerImpl) ControllerName() string {
	return "msgListControllerImpl"
}

// GetRouter 返回路由信息
func (c *msgListControllerImpl) GetRouter() string {
	return "/api/messages [GET]"
}

// Handle 处理获取已审核留言列表请求（隐藏状态信息）
func (c *msgListControllerImpl) Handle(ctx *gin.Context) {
	c.LoggerMgr.Ins().Debug("Starting to get approved messages")

	messages, err := c.MessageService.GetApprovedMessages()
	if err != nil {
		c.LoggerMgr.Ins().Error("Failed to get approved messages", "error", err)
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

	c.LoggerMgr.Ins().Info("Successfully retrieved approved messages", "count", len(responseList))

	ctx.JSON(common.HTTPStatusOK, dtos.SuccessWithData(responseList))
}

var _ IMsgListController = (*msgListControllerImpl)(nil)
