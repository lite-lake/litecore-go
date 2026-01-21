# AGENTS.md

Guidelines for agentic coding tools in this repository.

## Project Overview

- **Language**: Go 1.25+, Module: `github.com/lite-lake/litecore-go`
- **Framework**: Gin, GORM, Zap
- **Architecture**: 5-tier layered dependency injection (Entity → Repository → Service → Controller/Middleware)
- **Built-in Components**: Config and Manager are server built-in components, auto-initialized and injected

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
	Config    BaseConfigProvider `inject:""`
	DBManager DatabaseManager   `inject:""`
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
- Config (no deps) → Entity (no deps)
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

// Config and Manager are auto-initialized by the engine
// Register and inject all layers, then create engine
engine := server.NewEngine(entityContainer, repositoryContainer, serviceContainer, controllerContainer, middlewareContainer)
engine.Run()
```

## When Completing Tasks
1. `go test ./...` - verify no regressions
2. `go fmt ./...` - format code
3. `go vet ./...` - check issues
4. Verify package boundaries
5. Add tests and documentation
