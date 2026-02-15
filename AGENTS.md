# AGENTS.md

此仓库中 AI 编码工具的指南。

> **业务开发**：开发业务系统请从 [docs/user-guides/INDEX.md](docs/user-guides/INDEX.md) 开始。

## 项目概述

- **语言**: Go 1.25+, 模块: `github.com/lite-lake/litecore-go`
- **框架**: Gin, GORM, Zap
- **架构**: 5 层分层依赖注入 (Manager → Entity → Repository → Service → 交互层)

## 基本命令

```bash
go build -o litecore ./...
go test ./...                     # 测试所有
go test -cover ./...              # 生成覆盖率
go test ./util/jwt                # 测试指定包
go test -bench=. ./util/hash      # 基准测试
go fmt ./...
go vet ./...
go mod tidy
```

## 代码风格

### 导入顺序

```go
import (
	"crypto"       // 标准库

	"github.com/gin-gonic/gin"  // 第三方库

	"github.com/lite-lake/litecore-go/common"  // 本地模块
)
```

### 命名

| 类型 | 规范 | 示例 |
|------|------|------|
| 接口 | `I` 前缀 | `IDatabaseManager` |
| 私有结构体 | 小写 | `jwtEngine` |
| 公共结构体 | 大驼峰 | `ServerConfig` |
| 工厂函数 | `Build()`, `NewXxx()` | `NewMessageService()` |

### 格式化

- 使用 Tab，每行最多 120 字符
- 修改后运行 `go fmt ./...`
- 注释必须用中文

## 完成任务时

1. `go test ./...` - 验证无回归
2. `go fmt ./...` - 格式化代码
3. `go vet ./...` - 检查问题
