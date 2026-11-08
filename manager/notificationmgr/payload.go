package notificationmgr

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/lite-lake/litecore-go/common/deployinfo"
)

// eventInfo 事件元信息
type eventInfo struct {
	Icon  string
	Label string
}

// 事件映射表
var eventMap = map[string]eventInfo{
	"starting":     {Icon: "\U0001f680", Label: "服务启动中"},
	"started":      {Icon: "\u2705", Label: "启动成功"},
	"start_failed": {Icon: "\u274c", Label: "启动失败"},
	"stopping":     {Icon: "\U0001f6d1", Label: "正在停止"},
	"stopped":      {Icon: "\U0001f4f4", Label: "已停止"},
}

// buildTextContent 构建企业微信 TEXT 格式消息内容
func buildTextContent(event, appName, appEnv, hostname string, details map[string]string) string {
	info, ok := eventMap[event]
	if !ok {
		info = eventInfo{Icon: "\u2139\ufe0f", Label: event}
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("[%s %s] %s", info.Icon, info.Label, appName))

	// 构建版本（LITE_BUILD_ID 环境变量，由 CI/CD 在镜像构建时注入）
	if buildID := os.Getenv("LITE_BUILD_ID"); buildID != "" {
		sb.WriteString(fmt.Sprintf(" [%s]", buildID))
	}

	sb.WriteString(fmt.Sprintf("\n主机: %s", hostname))

	if appEnv != "" {
		sb.WriteString(fmt.Sprintf("\n应用环境: %s", appEnv))
	}

	for k, v := range details {
		if v != "" {
			sb.WriteString(fmt.Sprintf("\n%s: %s", k, v))
		}
	}

	// 追加部署环境信息（DEPLOY_ 环境变量）
	di := deployinfo.Get()
	if di.IsSet() {
		if di.DeployEnv != "" {
			sb.WriteString(fmt.Sprintf("\n部署环境: %s", di.DeployEnv))
		}
		if di.DeployZone != "" {
			sb.WriteString(fmt.Sprintf("\n部署网区: %s", di.DeployZone))
		}
		if di.DeployServer != "" {
			sb.WriteString(fmt.Sprintf("\n部署主机: %s", di.DeployServer))
		}
		if di.DeployService != "" {
			sb.WriteString(fmt.Sprintf("\n部署服务: %s", di.DeployService))
		}
	}

	sb.WriteString(fmt.Sprintf("\n时间: %s", time.Now().Format("2006-01-02 15:04:05 -0700")))
	return sb.String()
}

// buildWeComTextPayload 构建企业微信群机器人 TEXT 消息体
func buildWeComTextPayload(content string) map[string]interface{} {
	return map[string]interface{}{
		"msgtype": "text",
		"text": map[string]string{
			"content": content,
		},
	}
}
