package controller

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/lite-lake/litecore-go/common"
)

// IMetricsController 指标控制器接口
type IMetricsController interface {
	common.IBaseController
}

type MetricsController struct {
	ManagerContainer common.IBaseManager `inject:""`
	ServiceContainer common.IBaseService `inject:""`
}

func NewMetricsController() IMetricsController {
	return &MetricsController{}
}

func (c *MetricsController) ControllerName() string {
	return "MetricsController"
}

func (c *MetricsController) GetRouter() string {
	return "/metrics [GET]"
}

func (c *MetricsController) Handle(ctx *gin.Context) {
	metrics := map[string]interface{}{
		"server":  "litecore-go",
		"status":  "running",
		"version": "1.0.0",
	}

	if c.ManagerContainer != nil {
		managers := []common.IBaseManager{c.ManagerContainer}
		metrics["managers"] = len(managers)
	}

	if c.ServiceContainer != nil {
		services := []common.IBaseService{c.ServiceContainer}
		metrics["services"] = len(services)
	}

	ctx.JSON(http.StatusOK, metrics)
}

var _ common.IBaseController = (*MetricsController)(nil)
