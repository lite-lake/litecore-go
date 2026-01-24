# AGENTS.md

此仓库中 AI 编码工具的指南。

## 项目概述

- **语言**: Go 1.25+, 模块: `github.com/lite-lake/litecore-go`
- **框架**: Gin, GORM, Zap
 - **架构**: 5 层分层依赖注入 (内置管理器层 → Entity → Repository → Service → 交互层)
 - **交互层**: Controller/Middleware/Listener/Scheduler 统一处理 HTTP 请求、MQ 消息和定时任务
- **内置组件**: 位于 `server/builtin/manager/` 的管理器组件，自动初始化并注入

## 基本命令

### 构建/测试/检查
```bash
go build -o litecore ./...
go test ./...                     # 测试所有
go test -cover ./...              # 生成覆盖率
go test ./util/jwt                # 测试指定包
go test ./util/jwt -run TestGenerateHS256Token
go test -v ./util/jwt -run TestGenerateHS256Token/valid_StandardClaims
go test -bench=. ./util/hash       # 基准测试
go test -bench=BenchmarkMD5 ./util/hash
go fmt ./...
go vet ./...
go mod tidy
```

## 代码风格

### 导入顺序（标准库 → 第三方库 → 本地模块）
```go
import (
	"crypto"       // 标准库优先
	"errors"
	"time"

	"github.com/gin-gonic/gin"  // 第三方库其次
	"github.com/stretchr/testify/assert"

	"github.com/lite-lake/litecore-go/common"  // 本地模块最后
)
```

### 命名
- **接口**: `I*` 前缀 (例如: `ILiteUtilJWT`, `IDatabaseManager`)
- **私有结构体**: 小写 (例如: `jwtEngine`, `hashEngine`)
- **公共结构体**: 大驼峰 (例如: `StandardClaims`, `ServerConfig`)
- **函数**: 导出用大驼峰，私有用小驼峰
- **枚举**: `iota` 配合中文注释

### 注释（中文）
- 所有注释必须用中文
- 导出函数需要 godoc 注释

### 错误处理
```go
if err != nil {
	return "", fmt.Errorf("操作失败: %w", err)
}
```

### 依赖注入模式
```go
type UserServiceImpl struct {
	Config    configmgr.IConfigManager    `inject:""`
	DBManager databasemgr.IDatabaseManager `inject:""`
}
```

### 测试
- 使用 `t.Run()` 的表驱动测试，用中文
- 基准测试使用 `b.ResetTimer()`

### 格式化
- 使用 Tab，每行最多 120 字符
- 修改后运行 `go fmt ./...`

## 架构

### 依赖规则
- Entity（无依赖）
- Manager → Config + 其他 Manager
- Repository → Config + Manager + Entity
- Service → Config + Manager + Repository + 其他 Service
- 交互层 (Controller/Middleware/Listener/Scheduler) → Config + Manager + Service

### 依赖注入设置
```go
entityContainer := container.NewEntityContainer()
repositoryContainer := container.NewRepositoryContainer(entityContainer)
serviceContainer := container.NewServiceContainer(repositoryContainer)
controllerContainer := container.NewControllerContainer(serviceContainer)
middlewareContainer := container.NewMiddlewareContainer(serviceContainer)
listenerContainer := container.NewListenerContainer(serviceContainer)
schedulerContainer := container.NewSchedulerContainer(serviceContainer)

// Manager 由引擎通过 server.BuiltinConfig 自动初始化
engine := server.NewEngine(
    &server.BuiltinConfig{
        Driver:   "yaml",
        FilePath: "configs/config.yaml",
    },
    entityContainer,
    repositoryContainer,
    serviceContainer,
    controllerContainer,
    middlewareContainer,
    listenerContainer,
    schedulerContainer,
)
engine.Run()
```

## 完成任务时
1. `go test ./...` - 验证无回归
2. `go fmt ./...` - 格式化代码
3. `go vet ./...` - 检查问题
4. 验证包边界
5. 添加测试和文档

## 日志使用规范

### 禁止使用
- ❌ 标准库 log.Fatal/Print/Printf/Println
- ❌ fmt.Printf/fmt.Println（仅限开发调试）
- ❌ println/print

### 推荐使用
- ✅ 依赖注入 ILoggerManager
- ✅ 使用结构化日志：logger.Info("msg", "key", value)
- ✅ 使用 With 添加上下文：logger.With("user_id", id).Info("...")

### 各层日志级别
- Debug: 开发调试信息
- Info: 正常业务流程（请求开始/完成、资源创建）
- Warn: 降级处理、慢查询、重试
- Error: 业务错误、操作失败（需人工关注）
- Fatal: 致命错误，需要立即终止

### 敏感信息处理
- 密码、token、密钥等必须脱敏
- 使用内置过滤规则或自定义脱敏函数

### 业务层日志实现
```go
type MyService struct {
    LoggerMgr loggermgr.ILoggerManager `inject:""`
}

func (s *MyService) SomeMethod() {
    s.LoggerMgr.Ins().Info("操作开始", "param", value)
}
```

### 日志格式

#### 格式类型
- **gin**: Gin 风格，竖线分隔符，适合控制台输出（默认格式）
- **json**: JSON 格式，适合日志分析和监控
- **default**: 默认 ConsoleEncoder 格式

#### Gin 格式特点
- 统一格式：`{时间} | {级别} | {消息} | {字段1}={值1} {字段2}={值2} ...`
- 时间固定宽度 23 字符：`2006-01-02 15:04:05.000`
- 级别固定宽度 5 字符，右对齐，带颜色
- 字段格式：`key=value`，字符串值用引号包裹

#### 配置方法
在配置文件中通过 `console_config` 设置控制台日志格式：

```yaml
logger:
  driver: "zap"
  zap_config:
    console_enabled: true
    console_config:
      level: "info"                               # 日志级别：debug, info, warn, error, fatal
      format: "gin"                                # 格式：gin | json | default
      color: true                                  # 是否启用颜色
      time_format: "2006-01-02 15:04:05.000"     # 时间格式
```

#### 颜色配置
- **color**: 控制是否启用彩色输出（默认 true，由终端自动检测）
- 支持在配置文件中手动关闭颜色：`color: false`

#### 日志级别颜色
| 级别 | ANSI 颜色 | 说明 |
|------|-----------|------|
| DEBUG | 灰色 | 开发调试信息 |
| INFO | 绿色 | 正常业务流程 |
| WARN | 黄色 | 降级处理、慢查询 |
| ERROR | 红色 | 业务错误、操作失败 |
| FATAL | 红色+粗体 | 致命错误 |

#### HTTP 状态码颜色
| 状态码范围 | 颜色 | 说明 |
|-----------|------|------|
| 2xx | 绿色 | 成功 |
| 3xx | 黄色 | 重定向 |
| 4xx | 橙色 | 客户端错误 |
| 5xx | 红色 | 服务器错误 |

#### 格式示例
```go
// Gin 格式输出示例
2026-01-24 15:04:05.123 | INFO  | 开始依赖注入 | count=23
2026-01-24 15:04:05.456 | WARN  | 慢查询检测 | duration=1.2s
2026-01-24 15:04:05.789 | ERROR | 数据库连接失败 | error="connection refused"

// 请求日志（Gin 格式）
2026-01-24 15:04:05.123 | 200   | 1.234ms | 127.0.0.1 | GET | /api/messages
2026-01-24 15:04:05.456 | 404   | 0.456ms | 192.168.1.1 | POST | /api/unknown
```
