// Package controllers 定义 HTTP 控制器
package controllers

import (
	"com.litelake.litecore/common"
	"com.litelake.litecore/samples/messageboard/internal/dtos"
	"com.litelake.litecore/samples/messageboard/internal/services"

	"github.com/gin-gonic/gin"
)

// GetMessagesController 获取已审核留言列表控制器
type GetMessagesController struct {
	MessageService *services.MessageService `inject:""`
}

// NewGetMessagesController 创建控制器实例
func NewGetMessagesController() *GetMessagesController {
	return &GetMessagesController{}
}

// ControllerName 实现 BaseController 接口
func (c *GetMessagesController) ControllerName() string {
	return "GetMessagesController"
}

// GetRouter 实现 BaseController 接口
func (c *GetMessagesController) GetRouter() string {
	return "/api/messages [GET]"
}

// Handle 实现 BaseController 接口
func (c *GetMessagesController) Handle(ctx *gin.Context) {
	// 获取已审核通过的留言
	messages, err := c.MessageService.GetApprovedMessages()
	if err != nil {
		ctx.JSON(500, dtos.ErrInternalServer)
		return
	}

	// 转换为响应格式
	var responseList []dtos.MessageResponse
	for _, msg := range messages {
		responseList = append(responseList, dtos.ToMessageResponse(
			msg.ID,
			msg.Nickname,
			msg.Content,
			msg.Status,
			msg.CreatedAt,
		))
	}

	// 返回响应（用户端不返回 status 字段）
	for i := range responseList {
		responseList[i].Status = ""
	}

	ctx.JSON(200, dtos.SuccessWithData(responseList))
}

var _ common.BaseController = (*GetMessagesController)(nil)
