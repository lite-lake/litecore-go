package configmgr

import (
	"os"
	"regexp"
)

var envVarPattern = regexp.MustCompile(`\$\{([A-Za-z_][A-Za-z0-9_]*)\}`)

// expandEnvVars 递归展开配置数据中的 ${VAR_NAME} 环境变量引用
func expandEnvVars(val any) any {
	switch v := val.(type) {
	case string:
		return expandString(v)
	case map[string]any:
		for k, v2 := range v {
			v[k] = expandEnvVars(v2)
		}
		return v
	case []any:
		for i, item := range v {
			v[i] = expandEnvVars(item)
		}
		return v
	default:
		return val
	}
}

// expandString 替换字符串中的 ${VAR_NAME} 为环境变量值
// 如果环境变量不存在，保留原始占位符
func expandString(s string) string {
	return envVarPattern.ReplaceAllStringFunc(s, func(match string) string {
		varName := match[2 : len(match)-1]
		if val, ok := os.LookupEnv(varName); ok {
			return val
		}
		return match
	})
}
