package cachebust

import (
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	// ImmutableMaxAge 不可变缓存时长：1 年
	ImmutableMaxAge = "31536000"

	// StaticPrefix 静态资源 URL 前缀
	StaticPrefix = "/static"

	// HeaderImmutable 生产环境版本化静态资源的不可变缓存头
	HeaderImmutable = "public, max-age=31536000, immutable"

	// HeaderNoCache 每次回源校验（开发环境静态资源与 302 响应）
	HeaderNoCache = "no-cache"

	// HeaderHTMLNoCache HTML 页面统一缓存头，保证每次打开都回源校验、第一时间拿到最新版本化 URL
	HeaderHTMLNoCache = "no-cache"
)

var (
	once    sync.Once
	version string

	// rewriteStaticRe 匹配模板源码中的本站 /static/ 资源引用（<script src> / <link href>）
	rewriteStaticRe = regexp.MustCompile(`(src|href)="/static/([^"?#]+)"`)
)

// Version 返回当前静态资源版本号（进程内固定一次）。
// 优先级：LITE_BUILD_ID > ./static 目录最大文件 mtime > 进程启动时间戳。
func Version() string {
	once.Do(func() {
		version = computeVersion()
	})
	return version
}

// VersionPrefix 返回版本化前缀，如 "/static/20260725_163000_abc1234"。
func VersionPrefix() string {
	return StaticPrefix + "/" + Version()
}

// RewriteHTML 将模板源码中的 /static/ 资源引用重写为版本化路径。
// 在模板解析期调用一次，版本号烘焙进已解析模板，请求期零开销。
// 幂等：已含当前版本前缀（/static/{version}/...）或已带 query 的路径不重复处理；
// 仅重写 <script src> / <link href> 引用的本站 /static/ 资源，外链 CDN、内联、非 /static/ 引用保持不变。
func RewriteHTML(src []byte) []byte {
	prefix := StaticPrefix + "/" + Version()
	return rewriteStaticRe.ReplaceAllFunc(src, func(match []byte) []byte {
		sub := rewriteStaticRe.FindSubmatch(match)
		path := string(sub[2])
		// 跳过已含当前版本前缀的路径，避免重复版本段
		if strings.HasPrefix(path, Version()+"/") {
			return match
		}
		return []byte(string(sub[1]) + `="` + prefix + "/" + path + `"`)
	})
}

// IsVersionedPath 判断 filepath 是否已是当前版本路径。
// 入参为路由参数，形如 "/css/x.css" 或 "/20260725_163000_abc1234/css/x.css"。
func IsVersionedPath(filepath string) bool {
	return strings.HasPrefix(filepath, "/"+Version()+"/")
}

// StripVersion 剥离版本前缀，返回真实文件路径，如 "/css/x.css"。
// 兼容两种入参形式："/{version}/..."（路由参数）与 "/static/{version}/..."（完整路径）。
func StripVersion(filepath string) string {
	versionPrefix := "/" + Version()
	// 先匹配完整路径形式（含 /static 前缀），再匹配路由参数形式
	for _, prefix := range []string{StaticPrefix + versionPrefix, versionPrefix} {
		if strings.HasPrefix(filepath, prefix) {
			return strings.TrimPrefix(filepath, prefix)
		}
	}
	return filepath
}

// IsProd 判断是否为生产构建（存在 LITE_BUILD_ID）。
func IsProd() bool {
	return os.Getenv("LITE_BUILD_ID") != ""
}

// CacheControlForStatic 返回静态资源缓存头。
// 生产返回 immutable 长缓存；开发返回 no-cache，防止 dev 场景被 immutable 锁死。
func CacheControlForStatic() string {
	if IsProd() {
		return HeaderImmutable
	}
	return HeaderNoCache
}

// computeVersion 计算版本号。
func computeVersion() string {
	if id := os.Getenv("LITE_BUILD_ID"); id != "" {
		return id
	}
	if mtime := latestFileMtime("./static"); mtime > 0 {
		return strconv.FormatInt(mtime, 10)
	}
	return strconv.FormatInt(time.Now().Unix(), 10)
}

// latestFileMtime 递归遍历目录，返回最大文件的 mtime（unix 时间戳）。
// 目录不存在或为空时返回 0。
func latestFileMtime(root string) int64 {
	entries, err := os.ReadDir(root)
	if err != nil {
		return 0
	}
	var latest int64
	for _, entry := range entries {
		info, err := entry.Info()
		if err != nil {
			continue
		}
		path := filepath.Join(root, entry.Name())
		if info.IsDir() {
			if mtime := latestFileMtime(path); mtime > latest {
				latest = mtime
			}
			continue
		}
		if info.ModTime().Unix() > latest {
			latest = info.ModTime().Unix()
		}
	}
	return latest
}
