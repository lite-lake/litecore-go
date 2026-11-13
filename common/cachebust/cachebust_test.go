package cachebust

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

// resetForTesting 重置全局版本号，仅供测试使用
func resetForTesting() {
	once = sync.Once{}
	version = ""
}

func TestVersion_设置LiteBuildID时返回该值(t *testing.T) {
	t.Setenv("LITE_BUILD_ID", "20260803_120000_abc1234")
	resetForTesting()

	assert.Equal(t, "20260803_120000_abc1234", Version())
}

func TestVersion_未设置时返回非空(t *testing.T) {
	t.Setenv("LITE_BUILD_ID", "")
	resetForTesting()

	v := Version()
	assert.NotEmpty(t, v)
	// 版本号仅含字母数字与下划线（URL 安全）
	assert.Regexp(t, `^[A-Za-z0-9_]+$`, v)
}

func TestVersion_进程内多次调用一致(t *testing.T) {
	t.Setenv("LITE_BUILD_ID", "20260803_120000_abc1234")
	resetForTesting()

	v1 := Version()
	v2 := Version()
	assert.Equal(t, v1, v2)
}

func TestIsVersionedPath(t *testing.T) {
	t.Setenv("LITE_BUILD_ID", "20260803_120000_abc1234")
	resetForTesting()

	tests := []struct {
		name     string
		filepath string
		want     bool
	}{
		{"当前版本路径", "/20260803_120000_abc1234/css/x.css", true},
		{"当前版本JSON路径", "/20260803_120000_abc1234/locales/zhs.json", true},
		{"未版本化路径", "/css/x.css", false},
		{"过期版本路径", "/20260715_100000_abc1234/css/x.css", false},
		{"空路径", "", false},
		{"版本根路径无尾斜杠", "/20260803_120000_abc1234", false},
		{"其他资源", "/locales/zhs.json", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, IsVersionedPath(tt.filepath))
		})
	}
}

func TestStripVersion(t *testing.T) {
	t.Setenv("LITE_BUILD_ID", "20260803_120000_abc1234")
	resetForTesting()

	tests := []struct {
		name     string
		filepath string
		want     string
	}{
		{"路由参数形式", "/20260803_120000_abc1234/css/x.css", "/css/x.css"},
		{"完整路径形式", "/static/20260803_120000_abc1234/css/x.css", "/css/x.css"},
		{"locales JSON", "/20260803_120000_abc1234/locales/zhs.json", "/locales/zhs.json"},
		{"未版本化路径原样返回", "/css/x.css", "/css/x.css"},
		{"空路径", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, StripVersion(tt.filepath))
		})
	}
}

func TestCacheControlForStatic(t *testing.T) {
	t.Run("生产返回immutable", func(t *testing.T) {
		t.Setenv("LITE_BUILD_ID", "20260803_120000_abc1234")
		assert.Equal(t, HeaderImmutable, CacheControlForStatic())
	})

	t.Run("开发返回no-cache", func(t *testing.T) {
		t.Setenv("LITE_BUILD_ID", "")
		assert.Equal(t, HeaderNoCache, CacheControlForStatic())
	})
}

func TestIsProd(t *testing.T) {
	t.Run("存在LITE_BUILD_ID", func(t *testing.T) {
		t.Setenv("LITE_BUILD_ID", "20260803_120000_abc1234")
		assert.True(t, IsProd())
	})

	t.Run("无LITE_BUILD_ID", func(t *testing.T) {
		t.Setenv("LITE_BUILD_ID", "")
		assert.False(t, IsProd())
	})
}

func TestVersionPrefix(t *testing.T) {
	t.Setenv("LITE_BUILD_ID", "20260803_120000_abc1234")
	resetForTesting()

	assert.Equal(t, "/static/20260803_120000_abc1234", VersionPrefix())
}

func TestRewriteHTML(t *testing.T) {
	t.Setenv("LITE_BUILD_ID", "20260803_120000_abc1234")
	resetForTesting()

	tests := []struct {
		name string
		src  string
		want string
	}{
		{"script src 重写为版本化路径",
			`<script src="/static/js/a.js"></script>`,
			`<script src="/static/20260803_120000_abc1234/js/a.js"></script>`},
		{"link href 重写为版本化路径",
			`<link href="/static/css/a.css" rel="stylesheet">`,
			`<link href="/static/20260803_120000_abc1234/css/a.css" rel="stylesheet">`},
		{"已含当前版本前缀不变（幂等）",
			`<script src="/static/20260803_120000_abc1234/js/a.js"></script>`,
			`<script src="/static/20260803_120000_abc1234/js/a.js"></script>`},
		{"已带 query 不变（跳过）",
			`<link href="/static/css/a.css?v=1">`,
			`<link href="/static/css/a.css?v=1">`},
		{"外链 CDN 不变",
			`<script src="https://cdn.example.com/js/a.js"></script>`,
			`<script src="https://cdn.example.com/js/a.js"></script>`},
		{"内联脚本中的字符串不变",
			`<script>var x = "/static/js/a.js";</script>`,
			`<script>var x = "/static/js/a.js";</script>`},
		{"非 static 引用不变",
			`<link href="/css/a.css">`,
			`<link href="/css/a.css">`},
		{"相对路径引用不变",
			`<img src="./img/a.png">`,
			`<img src="./img/a.png">`},
		{"partials define 块同样被重写",
			`{{define "head"}}<link href="/static/css/style.css">{{end}}`,
			`{{define "head"}}<link href="/static/20260803_120000_abc1234/css/style.css">{{end}}`},
		{"多资源混用全部重写",
			`<script src="/static/js/a.js"></script><link href="/static/css/a.css"><script src="/static/js/b.js"></script>`,
			`<script src="/static/20260803_120000_abc1234/js/a.js"></script><link href="/static/20260803_120000_abc1234/css/a.css"><script src="/static/20260803_120000_abc1234/js/b.js"></script>`},
		{"href 与 src 混用均重写",
			`<link href="/static/css/x.css"><script src="/static/js/y.js"></script>`,
			`<link href="/static/20260803_120000_abc1234/css/x.css"><script src="/static/20260803_120000_abc1234/js/y.js"></script>`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, string(RewriteHTML([]byte(tt.src))))
		})
	}
}
