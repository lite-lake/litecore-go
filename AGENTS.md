# AGENTS.md

Guidelines for agentic coding tools in this repository.

## Project Overview

- **Language**: Go 1.25+, Module: `com.litelake.litecore`
- **Framework**: Gin, GORM, Zap
- **Architecture**: 7-tier layered dependency injection (Config → Entity → Manager → Repository → Service → Controller/Middleware)

## Essential Commands

### Build/Test/Lint
```bash
# Build
go build -o litecore ./...

# Test all tests
go test ./...
go test -cover ./...

# Test specific package
go test ./util/jwt

# Run single test
go test ./util/jwt -run TestGenerateHS256Token
go test -v ./util/jwt -run TestGenerateHS256Token/valid_StandardClaims

# Run benchmarks
go test -bench=. ./util/hash
go test -bench=BenchmarkMD5 ./util/hash

# Format/vet
go fmt ./...
go vet ./...
go mod tidy
```

## Code Style

### Imports
```go
import (
	"crypto"       // stdlib first
	"errors"
	"time"

	"github.com/gin-gonic/gin"  // third-party second
	"github.com/stretchr/testify/assert"

	"com.litelake.litecore/common"  // local modules last
)
```

### Naming
- **Interfaces**: `ILiteUtilJWT`, `IDatabaseManager` (I prefix)
- **Private structs**: `jwtEngine`, `hashEngine` (lowercase)
- **Public structs**: `StandardClaims`, `ServerConfig` (PascalCase)
- **Functions**: PascalCase exported, camelCase private
- **Enums**: `iota` with comments

```go
const (
	// HS256 HMAC使用SHA-256
	HS256 JWTAlgorithm = "HS256"
	// HS384 HMAC使用SHA-384
	HS384 JWTAlgorithm = "HS384"
)
```

### Comments (Chinese)
- Use Chinese for all user-facing docs and comments
- Exported functions must have comments

```go
// GenerateHS256Token 使用HMAC SHA-256算法生成JWT
func (j *jwtEngine) GenerateHS256Token(...) (string, error) {
	// 创建头部
	header := newJWTHeader(algorithm)
}
```

### Error Handling
```go
if err != nil {
	return "", fmt.Errorf("encode claims failed: %w", err)
}
```

### Structs/Interfaces
- Interface first, then implementation
- Singleton default instances
- `inject:""` tags for DI

```go
type ILiteUtilJWT interface { ... }
type jwtEngine struct{}
var defaultJWT = newJWTEngine()
var JWT = defaultJWT

type UserServiceImpl struct {
	Config    BaseConfigProvider `inject:""`
	DBManager DatabaseManager   `inject:""`
}
```

### Testing
- Table-driven tests with `t.Run()` subtests
- Helpers at top, benchmarks with `Benchmark` prefix
- Use `testify/assert` or stdlib

```go
func TestGenerateHS256Token(t *testing.T) {
	tests := []struct {
		name    string
		claims  ILiteUtilJWTClaims
		wantErr bool
	}{
		{"valid StandardClaims", &StandardClaims{}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) { ... })
	}
}
```

### Formatting
- Use tabs (Go standard)
- `gofmt` to format
- Max line: 120 chars (soft)

### Generics
```go
func HashGeneric[T HashAlgorithm](data string, algorithm T) []byte {
	hasher := algorithm.Hash()
	hasher.Write([]byte(data))
	return hasher.Sum(nil)
}
```

## Architecture

### Dependency Rules (7 tiers)
- Config (no deps) → Entity (no deps)
- Manager → Config, other Managers
- Repository → Config, Manager, Entity
- Service → Config, Manager, Repository, other Services
- Controller → Config, Manager, Service
- Middleware → Config, Manager, Service

### DI Pattern
```go
// 创建容器
configContainer := container.NewConfigContainer()
entityContainer := container.NewEntityContainer()
managerContainer := container.NewManagerContainer(configContainer)
repositoryContainer := container.NewRepositoryContainer(configContainer, managerContainer, entityContainer)
serviceContainer := container.NewServiceContainer(configContainer, managerContainer, repositoryContainer)
controllerContainer := container.NewControllerContainer(configContainer, managerContainer, serviceContainer)
middlewareContainer := container.NewMiddlewareContainer(configContainer, managerContainer, serviceContainer)

// 注册配置
configProvider, _ := config.NewConfigProvider("yaml", "config.yaml")
container.RegisterConfig[common.BaseConfigProvider](configContainer, configProvider)

// 注册管理器
dbMgr := databasemgr.NewDatabaseManager()
container.RegisterManager[databasemgr.DatabaseManager](managerContainer, dbMgr)

// 注册实体
entityContainer.Register(&entity.User{})

// 注册仓储
userRepo := repository.NewUserRepository()
container.RegisterRepository[repository.IUserRepository](repositoryContainer, userRepo)

// 注册服务
userService := service.NewUserService()
container.RegisterService[service.IUserService](serviceContainer, userService)

// 注册控制器
userController := controller.NewUserController()
container.RegisterController[controller.IUserController](controllerContainer, userController)

// 注册中间件
authMiddleware := middleware.NewAuthMiddleware()
container.RegisterMiddleware[middleware.IAuthMiddleware](middlewareContainer, authMiddleware)

// 创建引擎，传入容器
engine := server.NewEngine(
    configContainer,
    entityContainer,
    managerContainer,
    repositoryContainer,
    serviceContainer,
    controllerContainer,
    middlewareContainer,
)

// 启动引擎
if err := engine.Run(); err != nil {
    panic(err)
}
```

## When Completing Tasks

1. `go test ./...` - verify no regressions
2. `go fmt ./...` - format code
3. `go vet ./...` - check issues
4. Verify package boundaries
5. Add tests and documentation
