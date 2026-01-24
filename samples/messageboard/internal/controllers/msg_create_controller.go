// Package controllers 定义 HTTP 控制器
package controllers

import (
	"github.com/gin-gonic/gin"
	"github.com/lite-lake/litecore-go/common"
	"github.com/lite-lake/litecore-go/manager/loggermgr"
	"github.com/lite-lake/litecore-go/samples/messageboard/internal/dtos"
	"github.com/lite-lake/litecore-go/samples/messageboard/internal/services"
)

// IMsgCreateController 创建留言控制器接口
type IMsgCreateController interface {
	common.IBaseController
}

type msgCreateControllerImpl struct {
	MessageService services.IMessageService `inject:""` // 留言服务
	LoggerMgr      loggermgr.ILoggerManager `inject:""` // 日志管理器
}

// NewMsgCreateController 创建控制器实例
func NewMsgCreateController() IMsgCreateController {
	return &msgCreateControllerImpl{}
}

// ControllerName 返回控制器名称
func (c *msgCreateControllerImpl) ControllerName() string {
	return "msgCreateControllerImpl"
}

// GetRouter 返回路由信息
func (c *msgCreateControllerImpl) GetRouter() string {
	return "/api/messages [POST]"
}

// Handle 处理创建留言请求
func (c *msgCreateControllerImpl) Handle(ctx *gin.Context) {
	var req dtos.CreateMessageRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		c.LoggerMgr.Ins().Error("Failed to create message: parameter binding error", "error", err)
		ctx.JSON(common.HTTPStatusBadRequest, dtos.ErrorResponse(common.HTTPStatusBadRequest, err.Error()))
		return
	}

	c.LoggerMgr.Ins().Debug("Starting to create message", "nickname", req.Nickname, "content_length", len(req.Content))

	message, err := c.MessageService.CreateMessage(req.Nickname, req.Content)
	if err != nil {
		c.LoggerMgr.Ins().Error("Failed to create message", "nickname", req.Nickname, "error", err)
		ctx.JSON(common.HTTPStatusBadRequest, dtos.ErrorResponse(common.HTTPStatusBadRequest, err.Error()))
		return
	}

	c.LoggerMgr.Ins().Info("Message created successfully", "id", message.ID, "nickname", message.Nickname)

	ctx.JSON(common.HTTPStatusOK, dtos.SuccessResponse("留言提交成功，等待审核", gin.H{
		"id": message.ID,
	}))
}

var _ IMsgCreateController = (*msgCreateControllerImpl)(nil)
