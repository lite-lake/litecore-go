# AGENTS.md

AI 编码工具指南。

> **实例业务开发**：开发业务系统请从 `samples/messageboard/` 目录参考完整示例。

## 项目概述

- **语言**: Go 1.25+, 模块: `github.com/lite-lake/litecore-go`
- **框架**: Gin, GORM, Zap
- **架构**: 5 层分层依赖注入 (Manager → Entity → Repository → Service → 交互层)

## 基本命令

```bash
go build -o litecore ./...            # 构建
go test ./...                         # 测试所有
go test ./util/jwt                    # 测试指定包
go test -run TestHashEngine_MD5 ./... # 运行单个测试
go test -cover ./...                  # 生成覆盖率
go test -bench=. ./util/hash          # 基准测试
go fmt ./...                          # 格式化代码
go vet ./...                          # 静态检查
go mod tidy                           # 整理依赖
```

## 架构层次

采用 5 层分层架构，依赖方向：Manager → Entity → Repository → Service → 交互层

| 层次 | 组件 | 目录 | 接口 | 职责 |
|------|------|------|------|------|
| 内置管理器层 | Config | `manager/configmgr/` | `IConfigManager` | 配置管理 |
| | Database | `manager/databasemgr/` | `IDatabaseManager` | 数据库连接 |
| | Cache | `manager/cachemgr/` | `ICacheManager` | 缓存服务 |
| | Logger | `manager/loggermgr/` | `ILoggerManager` | 日志服务 |
| | Lock | `manager/lockmgr/` | `ILockManager` | 分布式锁 |
| | Limiter | `manager/limitermgr/` | `ILimiterManager` | 限流服务 |
| | MQ | `manager/mqmgr/` | `IMQManager` | 消息队列 |
| | Telemetry | `manager/telemetrymgr/` | `ITelemetryManager` | 遥测追踪 |
| | Scheduler | `manager/schedulermgr/` | `ISchedulerManager` | 定时调度 |
| 实体层 | Entity | `internal/entities/` | `IBaseEntity` | 数据模型，无依赖 |
| 仓储层 | Repository | `internal/repositories/` | `IBaseRepository` | 数据访问，依赖 Manager + Entity |
| 服务层 | Service | `internal/services/` | `IBaseService` | 业务逻辑，依赖 Manager + Repository |
| 交互层 | Controller | `internal/controllers/` | `IBaseController` | HTTP 请求处理 |
| | Middleware | `internal/middlewares/` | `IBaseMiddleware` | 请求拦截处理 |
| | Listener | `internal/listeners/` | `IBaseListener` | 消息队列消费 |
| | Scheduler | `internal/schedulers/` | `IBaseScheduler` | 定时任务执行 |

## 代码风格

### 导入顺序

```go
import (
	"crypto"
	"errors"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"

	"github.com/lite-lake/litecore-go/common"
	"github.com/lite-lake/litecore-go/container"
)
```

### 命名规范

| 类型 | 规范 | 示例 |
|------|------|------|
| 接口 | `I` 前缀 + PascalCase | `IUserService`, `IDatabaseManager` |
| 私有结构体 | camelCase | `hashEngine`, `serverConfig` |
| 公共结构体 | PascalCase | `ServerConfig`, `Engine` |
| 实现结构体 | camelCase + `Impl` 后缀 | `userServiceImpl`, `userRepositoryImpl` |
| 工厂函数 | `New` 前缀 | `NewUserService()`, `NewEngine()` |
| 常量 | PascalCase 或全大写 | `FormatHexFull`, `BcryptDefaultCost` |
| 错误类型 | `XxxError` 后缀 | `DependencyNotFoundError` |

### 接口与依赖注入

```go
// IUserService 用户服务接口
// 所有 Service 类必须继承此接口并实现 ServiceName 方法
// 系统通过此接口判断是否符合标准服务定义
type IUserService interface {
	common.IBaseService
	// GetUserByID 根据 ID 获取用户
	GetUserByID(id string) (*User, error)
}

var _ IUserService = (*userServiceImpl)(nil)

type userServiceImpl struct {
	Config     configmgr.IConfigManager     `inject:""` // 配置管理器
	DBManager  databasemgr.IDatabaseManager `inject:""` // 数据库管理器
	LoggerMgr  loggermgr.ILoggerManager     `inject:""` // 日志管理器
	UserRepo   IUserRepository              `inject:""` // 用户仓储
	OrderSvc   IOrderService                `inject:""` // 订单服务
}
```

### 格式化

- 使用 Tab 缩进，每行最多 120 字符
- 修改后运行 `go fmt ./...`
- 注释使用中文

### 错误处理

```go
type DependencyNotFoundError struct {
	InstanceName  string
	FieldName     string
	FieldType     reflect.Type
	ContainerType string
	Message       string
}

func (e *DependencyNotFoundError) Error() string {
	if e.Message != "" {
		return fmt.Sprintf("dependency not found for %s.%s: need type %s from %s container: %s",
			e.InstanceName, e.FieldName, e.FieldType, e.ContainerType, e.Message)
	}
	return fmt.Sprintf("dependency not found for %s.%s: need type %s from %s container",
		e.InstanceName, e.FieldName, e.FieldType, e.ContainerType)
}

// Controller 响应
ctx.JSON(common.HTTPStatusNotFound, dtos.ErrorResponse(common.HTTPStatusNotFound, "资源不存在"))
ctx.JSON(common.HTTPStatusOK, dtos.SuccessResponse("操作成功", user))
```

### 测试规范

```go
func TestHashEngine_MD5(t *testing.T) {
	tests := []struct {
		name     string
		data     string
		expected string
	}{
		{"空字符串", "", "d41d8cd98f00b204e9800998ecf8427e"},
		{"简单字符串", "hello", "5d41402abc4b2a76b9719d911017c592"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if result := Hash.MD5String(tt.data); result != tt.expected {
				t.Errorf("MD5String() = %v, want %v", result, tt.expected)
			}
		})
	}
}
```

推荐使用 `github.com/stretchr/testify/assert` 进行断言：

```go
import "github.com/stretchr/testify/assert"

func TestSomething(t *testing.T) {
	assert.NoError(t, err)
	assert.Equal(t, expected, actual)
	assert.NotNil(t, result)
}
```

## 实体基类

```go
type User struct {
	common.BaseEntityWithTimestamps // ID (CUID2) + CreatedAt + UpdatedAt
	Name string `gorm:"type:varchar(100);not null" json:"name"`
}

func (u *User) EntityName() string { return "User" }
func (u *User) TableName() string  { return "users" }
func (u *User) GetId() string      { return u.ID }

var _ common.IBaseEntity = (*User)(nil)
```

实体可用的基类：
- `common.BaseEntityOnlyID` - 仅包含 ID
- `common.BaseEntityWithCreatedAt` - ID + CreatedAt
- `common.BaseEntityWithTimestamps` - ID + CreatedAt + UpdatedAt

## Controller 路由

```go
// GetRouter 返回路由信息
// 格式: "/path [METHOD]"，支持逗号分隔多个路由
func (c *userControllerImpl) GetRouter() string {
	return "/api/users [POST]"
}

// Handle 处理请求
func (c *userControllerImpl) Handle(ctx *gin.Context) {
	// 处理逻辑
}
```

路由格式说明：
- 单路由：`"/api/users [POST]"`
- 多路由：`"/api/users [POST],/api/users/:id [GET]"`

## 完成任务时

1. `go test ./...` - 验证无回归
2. `go fmt ./...` - 格式化代码
3. `go vet ./...` - 静态检查
4. 添加新组件后运行 `go run ./cmd/generate` 重新生成容器代码

## 常用工具包

| 包 | 用途 |
|---|---|
| `util/hash` | MD5, SHA256, SHA512, HMAC, Bcrypt |
| `util/jwt` | JWT 生成和解析 |
| `util/id` | CUID2 唯一标识符 |
| `util/rand` | 随机字符串 |
| `util/validator` | 参数校验 |
| `util/json` | JSON 序列化 |
