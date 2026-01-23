// Package controllers 定义 HTTP 控制器
package controllers

import (
	"github.com/gin-gonic/gin"
	"github.com/lite-lake/litecore-go/common"
	"github.com/lite-lake/litecore-go/manager/loggermgr"
	"github.com/lite-lake/litecore-go/samples/messageboard/internal/dtos"
	"github.com/lite-lake/litecore-go/samples/messageboard/internal/services"
)

// IMsgAllController 获取所有留言控制器接口（管理员专用）
type IMsgAllController interface {
	common.IBaseController
}

type msgAllControllerImpl struct {
	MessageService services.IMessageService `inject:""` // 留言服务
	LoggerMgr      loggermgr.ILoggerManager `inject:""` // 日志管理器
}

// NewMsgAllController 创建控制器实例
func NewMsgAllController() IMsgAllController {
	return &msgAllControllerImpl{}
}

// ControllerName 返回控制器名称
func (c *msgAllControllerImpl) ControllerName() string {
	return "msgAllControllerImpl"
}

// GetRouter 返回路由信息
func (c *msgAllControllerImpl) GetRouter() string {
	return "/api/admin/messages [GET]"
}

// Handle 处理获取所有留言列表请求（管理员专用）
func (c *msgAllControllerImpl) Handle(ctx *gin.Context) {
	c.LoggerMgr.Ins().Debug("开始获取所有留言列表")

	messages, err := c.MessageService.GetAllMessages()
	if err != nil {
		c.LoggerMgr.Ins().Error("获取所有留言失败", "error", err)
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

	c.LoggerMgr.Ins().Info("获取所有留言成功", "count", len(responseList))

	ctx.JSON(common.HTTPStatusOK, dtos.SuccessWithData(responseList))
}

var _ IMsgAllController = (*msgAllControllerImpl)(nil)
