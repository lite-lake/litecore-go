// Package controllers 定义 HTTP 控制器
package controllers

import (
	"github.com/lite-lake/litecore-go/common"
	"github.com/lite-lake/litecore-go/component/manager/loggermgr"
	"github.com/lite-lake/litecore-go/samples/messageboard/internal/dtos"
	"github.com/lite-lake/litecore-go/samples/messageboard/internal/services"

	"github.com/gin-gonic/gin"
)

// IMsgListController 获取留言控制器接口
type IMsgListController interface {
	common.IBaseController
}

type msgListControllerImpl struct {
	MessageService services.IMessageService `inject:""`
	LoggerMgr      loggermgr.ILoggerManager `inject:""`
	logger         loggermgr.ILogger
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

func (c *msgListControllerImpl) Logger() loggermgr.ILogger {
	return c.logger
}

func (c *msgListControllerImpl) SetLoggerManager(mgr loggermgr.ILoggerManager) {
	c.LoggerMgr = mgr
	c.initLogger()
}

func (c *msgListControllerImpl) initLogger() {
	if c.LoggerMgr != nil {
		c.logger = c.LoggerMgr.Logger("MsgListController")
	}
}

func (c *msgListControllerImpl) Handle(ctx *gin.Context) {
	if c.logger != nil {
		c.logger.Debug("开始获取已审核留言列表")
	}

	messages, err := c.MessageService.GetApprovedMessages()
	if err != nil {
		if c.logger != nil {
			c.logger.Error("获取已审核留言失败", "error", err)
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

	if c.logger != nil {
		c.logger.Info("获取已审核留言成功", "count", len(responseList))
	}

	ctx.JSON(common.HTTPStatusOK, dtos.SuccessWithData(responseList))
}

var _ IMsgListController = (*msgListControllerImpl)(nil)
