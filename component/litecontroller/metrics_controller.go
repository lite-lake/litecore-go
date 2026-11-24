package litecontroller

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/lite-lake/litecore-go/common"
	"github.com/lite-lake/litecore-go/common/metrics"
)

// IMetricsController 指标控制器接口（GET /metrics）。
type IMetricsController interface {
	common.IBaseController
}

// MetricsController Prometheus 文本指标出口。
//
// 说明：本控制器不持有任何 `inject:""` 依赖——指标数据来自进程级注册表
// （common/metrics），注册到容器后可直接输出，不会触发 DI 精确类型解析失败。
type MetricsController struct{}

// NewMetricsController 创建指标控制器。
func NewMetricsController() IMetricsController {
	return &MetricsController{}
}

// ControllerName 返回控制器名称。
func (c *MetricsController) ControllerName() string {
	return "MetricsController"
}

// GetRouter 返回路由定义（根路径，非 /api/**）。
func (c *MetricsController) GetRouter() string {
	return "/metrics [GET]"
}

// Handle 输出 Prometheus 文本格式指标。
func (c *MetricsController) Handle(ctx *gin.Context) {
	ctx.Data(http.StatusOK, "text/plain; version=0.0.4; charset=utf-8", []byte(metrics.Render()))
}

var _ common.IBaseController = (*MetricsController)(nil)
