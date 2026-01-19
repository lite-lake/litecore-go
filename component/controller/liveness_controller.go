package controller

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"com.litelake.litecore/common"
)

// ILivenessController 存活检查控制器接口
type ILivenessController interface {
	common.BaseController
}

type LivenessController struct {
	ManagerContainer common.BaseManager `inject:""`
}

func NewLivenessController() ILivenessController {
	return &LivenessController{}
}

func (c *LivenessController) ControllerName() string {
	return "LivenessController"
}

func (c *LivenessController) GetRouter() string {
	return "/live [GET]"
}

func (c *LivenessController) Handle(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, gin.H{
		"status":    "alive",
		"timestamp": time.Now().Format(time.RFC3339),
	})
}

var _ common.BaseController = (*LivenessController)(nil)
