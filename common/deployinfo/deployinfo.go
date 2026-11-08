// Package deployinfo 提供部署环境元信息的统一访问接口。
//
// 部署工具（yamlops）在部署时会自动向容器注入 4 个 DEPLOY_ 环境变量：
//   - DEPLOY_ENV_NAME: 部署环境名称（如 prod、dev）
//   - DEPLOY_ZONE_NAME: 部署网区名称
//   - DEPLOY_SERVER_NAME: 部署主机名称
//   - DEPLOY_SERVICE_NAME: 部署服务名称
//
// 本包在进程启动时一次性读取这些变量，提供不可变的全局访问器。
// 本地开发未设置这些变量时，DeployEnv 默认为 "local"，其余为空字符串。
//
// 注意：部署环境（DeployEnv）与应用环境（app.env 配置项）是不同的概念。
// 部署环境来自 DEPLOY_* 运行时环境变量，由部署工具注入；
// 应用环境来自 config.yaml 中的 app.env 字段，由配置文件决定。
//
// 基本用法：
//
//	info := deployinfo.Get()
//	fmt.Println(info.DeployEnv)     // "prod" 或 "local"
//	fmt.Println(info.IsSet())       // true 表示至少有一个 DEPLOY_ 变量被设置
package deployinfo

import (
	"os"
	"sync"
)

// DeployInfo 部署环境元信息（不可变）
type DeployInfo struct {
	DeployEnv     string // DEPLOY_ENV_NAME，未设置时默认 "local"
	DeployZone    string // DEPLOY_ZONE_NAME
	DeployServer  string // DEPLOY_SERVER_NAME
	DeployService string // DEPLOY_SERVICE_NAME
	LiteBuildID   string // LITE_BUILD_ID，CI/CD 构建时注入的构建标识
}

var (
	once     sync.Once
	instance *DeployInfo
)

// Get 返回全局唯一的 DeployInfo 实例（惰性初始化）
func Get() *DeployInfo {
	once.Do(func() {
		instance = &DeployInfo{
			DeployEnv:     getEnvOrDefault("DEPLOY_ENV_NAME", "local"),
			DeployZone:    os.Getenv("DEPLOY_ZONE_NAME"),
			DeployServer:  os.Getenv("DEPLOY_SERVER_NAME"),
			DeployService: os.Getenv("DEPLOY_SERVICE_NAME"),
			LiteBuildID:   os.Getenv("LITE_BUILD_ID"),
		}
	})
	return instance
}

// IsSet 判断是否至少有一个 DEPLOY_ 环境变量被设置（即非空且非默认值）
func (d *DeployInfo) IsSet() bool {
	return d.DeployEnv != "local" || d.DeployZone != "" || d.DeployServer != "" || d.DeployService != ""
}

// getEnvOrDefault 获取环境变量值，未设置时返回默认值
func getEnvOrDefault(key, defaultVal string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultVal
}

// resetForTesting 重置全局单例，仅供测试使用
func resetForTesting() {
	once = sync.Once{}
	instance = nil
}
