# AGENTS.md

Guidelines for agentic coding tools in this repository.

## Project Overview

- **Language**: Go 1.25+, Module: `github.com/lite-lake/litecore-go`
- **Framework**: Gin, GORM, Zap
- **Architecture**: 5-tier layered dependency injection (Entity → Repository → Service → Controller/Middleware)
- **Built-in Components**: Manager components located at `server/builtin/manager/`, auto-initialized and injected

## Essential Commands

### Build/Test/Lint
```bash
go build -o litecore ./...
go test ./...                     # Test all
go test -cover ./...              # With coverage
go test ./util/jwt                # Specific package
go test ./util/jwt -run TestGenerateHS256Token
go test -v ./util/jwt -run TestGenerateHS256Token/valid_StandardClaims
go test -bench=. ./util/hash       # Benchmarks
go test -bench=BenchmarkMD5 ./util/hash
go fmt ./...
go vet ./...
go mod tidy
```

## Code Style

### Imports (stdlib → third-party → local)
```go
import (
	"crypto"       // stdlib first
	"errors"
	"time"

	"github.com/gin-gonic/gin"  // third-party second
	"github.com/stretchr/testify/assert"

	"github.com/lite-lake/litecore-go/common"  // local modules last
)
```

### Naming
- **Interfaces**: `I*` prefix (e.g., `ILiteUtilJWT`, `IDatabaseManager`)
- **Private structs**: lowercase (e.g., `jwtEngine`, `hashEngine`)
- **Public structs**: PascalCase (e.g., `StandardClaims`, `ServerConfig`)
- **Functions**: PascalCase exported, camelCase private
- **Enums**: `iota` with Chinese comments

### Comments (Chinese)
- All comments must be in Chinese
- Exported functions need godoc comments

### Error Handling
```go
if err != nil {
	return "", fmt.Errorf("operation failed: %w", err)
}
```

### DI Pattern
```go
type UserServiceImpl struct {
	Config    configmgr.IConfigManager    `inject:""`
	DBManager databasemgr.IDatabaseManager `inject:""`
}
```

### Testing
- Table-driven tests with `t.Run()` in Chinese
- Benchmarks with `b.ResetTimer()`

### Formatting
- Tabs, max 120 chars/line
- Run `go fmt ./...` after changes

## Architecture

### Dependency Rules
- Entity (no deps)
- Manager → Config + other Managers
- Repository → Config + Manager + Entity
- Service → Config + Manager + Repository + other Services
- Controller → Config + Manager + Service
- Middleware → Config + Manager + Service

### DI Setup
```go
entityContainer := container.NewEntityContainer()
repositoryContainer := container.NewRepositoryContainer(entityContainer)
serviceContainer := container.NewServiceContainer(repositoryContainer)
controllerContainer := container.NewControllerContainer(serviceContainer)
middlewareContainer := container.NewMiddlewareContainer(serviceContainer)

// Managers are auto-initialized by the engine via server.BuiltinConfig
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
)
engine.Run()
```

## When Completing Tasks
1. `go test ./...` - verify no regressions
2. `go fmt ./...` - format code
3. `go vet ./...` - check issues
4. Verify package boundaries
5. Add tests and documentation

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
    logger     loggermgr.ILogger
}

func (s *MyService) initLogger() {
    if s.LoggerMgr != nil {
        s.logger = s.LoggerMgr.Logger("MyService")
    }
}

func (s *MyService) SomeMethod() {
    s.initLogger()
    s.logger.Info("操作开始", "param", value)
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
