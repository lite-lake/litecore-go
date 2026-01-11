// Package controllers 定义 HTTP 控制器
package controllers

import (
	"com.litelake.litecore/common"
	"com.litelake.litecore/samples/messageboard/internal/dtos"
	"com.litelake.litecore/samples/messageboard/internal/services"

	"github.com/gin-gonic/gin"
)

// GetAllMessagesController 获取所有留言控制器（管理端）
type GetAllMessagesController struct {
	MessageService *services.MessageService `inject:""`
}

// NewGetAllMessagesController 创建控制器实例
func NewGetAllMessagesController() *GetAllMessagesController {
	return &GetAllMessagesController{}
}

// ControllerName 实现 BaseController 接口
func (c *GetAllMessagesController) ControllerName() string {
	return "GetAllMessagesController"
}

// GetRouter 实现 BaseController 接口
func (c *GetAllMessagesController) GetRouter() string {
	return "/api/admin/messages [GET]"
}

// Handle 实现 BaseController 接口
func (c *GetAllMessagesController) Handle(ctx *gin.Context) {
	// 获取所有留言
	messages, err := c.MessageService.GetAllMessages()
	if err != nil {
		ctx.JSON(500, dtos.ErrInternalServer)
		return
	}

	// 转换为响应格式
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

var _ common.BaseController = (*GetAllMessagesController)(nil)
