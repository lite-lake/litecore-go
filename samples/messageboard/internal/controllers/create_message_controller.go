// Package controllers 定义 HTTP 控制器
package controllers

import (
	"com.litelake.litecore/samples/messageboard/internal/dtos"
	"com.litelake.litecore/samples/messageboard/internal/services"

	"github.com/gin-gonic/gin"
)

// CreateMessageController 创建留言控制器
type CreateMessageController struct {
	MessageService services.IMessageService `inject:""`
}

// NewCreateMessageController 创建控制器实例
func NewCreateMessageController() *CreateMessageController {
	return &CreateMessageController{}
}

// ControllerName 实现 BaseController 接口
func (c *CreateMessageController) ControllerName() string {
	return "CreateMessageController"
}

// GetRouter 实现 BaseController 接口
func (c *CreateMessageController) GetRouter() string {
	return "/api/messages [POST]"
}

// Handle 实现 BaseController 接口
func (c *CreateMessageController) Handle(ctx *gin.Context) {
	// 绑定请求参数
	var req dtos.CreateMessageRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(400, dtos.ErrorResponse(400, err.Error()))
		return
	}

	// 创建留言
	message, err := c.MessageService.CreateMessage(req.Nickname, req.Content)
	if err != nil {
		ctx.JSON(400, dtos.ErrorResponse(400, err.Error()))
		return
	}

	// 返回响应
	ctx.JSON(200, dtos.SuccessResponse("留言提交成功，等待审核", gin.H{
		"id": message.ID,
	}))
}

var _ ICreateMessageController = (*CreateMessageController)(nil)
