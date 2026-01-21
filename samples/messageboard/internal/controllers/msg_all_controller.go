// Package controllers 定义 HTTP 控制器
package controllers

import (
	"github.com/lite-lake/litecore-go/common"
	"github.com/lite-lake/litecore-go/component/manager/loggermgr"
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
	LoggerMgr      loggermgr.ILoggerManager `inject:""`
	logger         loggermgr.ILogger
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

func (c *msgAllControllerImpl) Logger() loggermgr.ILogger {
	return c.logger
}

func (c *msgAllControllerImpl) SetLoggerManager(mgr loggermgr.ILoggerManager) {
	c.LoggerMgr = mgr
	c.initLogger()
}

func (c *msgAllControllerImpl) initLogger() {
	if c.LoggerMgr != nil {
		c.logger = c.LoggerMgr.Logger("MsgAllController")
	}
}

func (c *msgAllControllerImpl) Handle(ctx *gin.Context) {
	if c.logger != nil {
		c.logger.Debug("开始获取所有留言列表")
	}

	messages, err := c.MessageService.GetAllMessages()
	if err != nil {
		if c.logger != nil {
			c.logger.Error("获取所有留言失败", "error", err)
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

	if c.logger != nil {
		c.logger.Info("获取所有留言成功", "count", len(responseList))
	}

	ctx.JSON(common.HTTPStatusOK, dtos.SuccessWithData(responseList))
}

var _ IMsgAllController = (*msgAllControllerImpl)(nil)
