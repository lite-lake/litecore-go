package litecontroller

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/lite-lake/litecore-go/common"
	"github.com/lite-lake/litecore-go/manager/loggermgr"
)

// RobotsConfig robots.txt 配置
type RobotsConfig struct {
	// Content robots.txt 响应内容
	Content string
	// URLPath 路由路径
	URLPath string
}

// DefaultRobotsContent 默认拒绝所有搜索引擎爬虫的 robots.txt 内容
const DefaultRobotsContent = "User-agent: *\nDisallow: /\n"

// defaultRobotsURLPath robots.txt 的默认路由路径
const defaultRobotsURLPath = "/robots.txt"

// ResourceRobotsController robots.txt 控制器
// 用于响应搜索引擎爬虫的 robots.txt 请求，默认拒绝所有爬虫
type ResourceRobotsController struct {
	config    *RobotsConfig
	LoggerMgr loggermgr.ILoggerManager `inject:""`
}

// NewResourceRobotsController 创建 robots.txt 控制器
// 默认内容为拒绝所有搜索引擎爬虫，路由路径为 /robots.txt
func NewResourceRobotsController() *ResourceRobotsController {
	return &ResourceRobotsController{
		config: &RobotsConfig{
			Content: DefaultRobotsContent,
			URLPath: defaultRobotsURLPath,
		},
	}
}

// NewResourceRobotsControllerWithContent 创建 robots.txt 控制器，使用自定义内容
// 当 content 为空时回退到默认内容（拒绝所有爬虫）
func NewResourceRobotsControllerWithContent(content string) *ResourceRobotsController {
	if content == "" {
		content = DefaultRobotsContent
	}
	return &ResourceRobotsController{
		config: &RobotsConfig{
			Content: content,
			URLPath: defaultRobotsURLPath,
		},
	}
}

// ControllerName 返回控制器名称
func (c *ResourceRobotsController) ControllerName() string {
	return "ResourceRobotsController"
}

// GetRouter 返回路由定义
func (c *ResourceRobotsController) GetRouter() string {
	return c.config.URLPath + " [GET]"
}

// Handle 处理 robots.txt 请求，返回纯文本内容
func (c *ResourceRobotsController) Handle(ctx *gin.Context) {
	ctx.Header("Content-Type", "text/plain; charset=utf-8")
	ctx.Header("Cache-Control", "public, max-age=3600")
	ctx.String(http.StatusOK, c.config.Content)
}

// GetConfig 获取 robots.txt 配置
func (c *ResourceRobotsController) GetConfig() *RobotsConfig {
	return c.config
}

var _ common.IBaseController = (*ResourceRobotsController)(nil)
