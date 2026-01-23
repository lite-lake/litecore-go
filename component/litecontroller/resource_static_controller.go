package litecontroller

import (
	"github.com/lite-lake/litecore-go/server/builtin/manager/loggermgr"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/lite-lake/litecore-go/common"
)

// ResourceStaticConfig 静态文件配置
type ResourceStaticConfig struct {
	URLPath  string // URL路径前缀，如 /static
	FilePath string // 文件系统路径，如 ./static
}

// ResourceStaticController 静态文件控制器
// 用于处理静态文件服务
type ResourceStaticController struct {
	config    *ResourceStaticConfig
	LoggerMgr loggermgr.ILoggerManager `inject:""`
}

// NewResourceStaticController 创建静态文件控制器
func NewResourceStaticController(urlPath, filePath string) *ResourceStaticController {
	return &ResourceStaticController{
		config: &ResourceStaticConfig{
			URLPath:  urlPath,
			FilePath: filePath,
		},
	}
}

func (c *ResourceStaticController) ControllerName() string {
	return "ResourceStaticController"
}

func (c *ResourceStaticController) GetRouter() string {
	return c.config.URLPath + "/*filepath [GET]"
}

func (c *ResourceStaticController) Handle(ctx *gin.Context) {
	ctx.FileFromFS("/"+ctx.Param("filepath"), http.Dir(c.config.FilePath))
}

// GetConfig 获取静态文件配置
func (c *ResourceStaticController) GetConfig() *ResourceStaticConfig {
	return c.config
}

var _ common.IBaseController = (*ResourceStaticController)(nil)
