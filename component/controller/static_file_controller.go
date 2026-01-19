package controller

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"com.litelake.litecore/common"
)

// StaticFileConfig 静态文件配置
type StaticFileConfig struct {
	URLPath  string // URL路径前缀，如 /static
	FilePath string // 文件系统路径，如 ./static
}

// StaticFileController 静态文件控制器
// 用于处理静态文件服务
type StaticFileController struct {
	config *StaticFileConfig
}

// NewStaticFileController 创建静态文件控制器
func NewStaticFileController(urlPath, filePath string) *StaticFileController {
	return &StaticFileController{
		config: &StaticFileConfig{
			URLPath:  urlPath,
			FilePath: filePath,
		},
	}
}

func (c *StaticFileController) ControllerName() string {
	return "StaticFileController"
}

func (c *StaticFileController) GetRouter() string {
	return c.config.URLPath + "/*filepath [GET]"
}

func (c *StaticFileController) Handle(ctx *gin.Context) {
	ctx.FileFromFS(ctx.Request.URL.Path, http.Dir(c.config.FilePath))
}

// GetConfig 获取静态文件配置
func (c *StaticFileController) GetConfig() *StaticFileConfig {
	return c.config
}

var _ common.BaseController = (*StaticFileController)(nil)
