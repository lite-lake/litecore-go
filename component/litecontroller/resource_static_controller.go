package litecontroller

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/lite-lake/litecore-go/common"
	"github.com/lite-lake/litecore-go/common/cachebust"
	"github.com/lite-lake/litecore-go/manager/loggermgr"
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

// Handle 处理静态资源请求。
// 版本化路径（/static/{version}/...）剥离版本前缀后服务文件，生产返回 immutable 长缓存；
// 未版本化（或过期版本）请求返回 302 收敛到当前版本，保证部署后浏览器必然拉取最新资源。
func (c *ResourceStaticController) Handle(ctx *gin.Context) {
	filepath := ctx.Param("filepath")

	if cachebust.IsVersionedPath(filepath) {
		// 版本化请求：剥离版本前缀后服务文件
		ctx.Header("Cache-Control", cachebust.CacheControlForStatic())
		ctx.FileFromFS(cachebust.StripVersion(filepath), http.Dir(c.config.FilePath))
		return
	}

	// 未版本化（或过期版本）：302 收敛到当前版本，302 不可缓存
	ctx.Header("Cache-Control", cachebust.HeaderNoCache)
	http.Redirect(ctx.Writer, ctx.Request, cachebust.VersionPrefix()+filepath, http.StatusFound)
}

// GetConfig 获取静态文件配置
func (c *ResourceStaticController) GetConfig() *ResourceStaticConfig {
	return c.config
}

var _ common.IBaseController = (*ResourceStaticController)(nil)
