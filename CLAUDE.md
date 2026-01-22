# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

**LiteCore-Go** is a layered application framework for Go that provides dependency injection containers, lifecycle management, and infrastructure managers. Built on top of Gin, GORM, and Zap.

**Module**: `github.com/lite-lake/litecore-go`
**Go Version**: 1.25+
**Architecture**: 5-tier layered dependency injection (Entity → Repository → Service → Controller/Middleware)
**Built-in Components**: Manager components at `server/builtin/manager/`, auto-initialized and injected

## Essential Commands

### Build, Test, and Lint

```bash
# Build all packages
go build -o litecore ./...

# Run all tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Test specific package
go test ./util/jwt

# Run single test
go test ./util/jwt -run TestGenerateHS256Token
go test -v ./util/jwt -run TestGenerateHS256Token/valid_StandardClaims

# Run benchmarks
go test -bench=. ./util/hash
go test -bench=BenchmarkMD5 ./util/hash

# Format and vet
go fmt ./...
go vet ./...
go mod tidy
```

## Architecture Overview

### 5-Tier Layered Architecture

The framework enforces strict layer boundaries with unidirectional dependencies:

```
┌─────────────────────────────────────────────────────────────┐
│  Controller Layer   (BaseController)                        │
│  Middleware Layer   (BaseMiddleware)                        │
├─────────────────────────────────────────────────────────────┤
│  Service Layer      (BaseService)                           │
├─────────────────────────────────────────────────────────────┤
│  Repository Layer   (BaseRepository)                        │
├─────────────────────────────────────────────────────────────┤
│  Entity Layer       (BaseEntity)                            │
│  Manager Layer      (BaseManager)                           │
│  Location: server/builtin/manager/                          │
└─────────────────────────────────────────────────────────────┘
```

**Dependency Rules**:
- **Entity** - No dependencies
- **Manager** → Config, other Managers (same-layer dependencies)
- **Repository** → Config, Manager, Entity
- **Service** → Config, Manager, Repository, other Services (same-layer)
- **Controller** → Config, Manager, Service
- **Middleware** → Config, Manager, Service

### Dependency Injection Container

The container system (`container/`) manages component lifecycle and enforces layer boundaries.

**Key Pattern - Registration by Interface Type**:
```go
// Register instances by interface type using RegisterByType
serviceContainer.RegisterByType(
    reflect.TypeOf((*UserService)(nil)).Elem(),
    &UserServiceImpl{},
)
```

**Two-Phase Injection**:
1. **Registration Phase** (`RegisterByType`) - Add instances to container, no injection
2. **Injection Phase** (`InjectAll`) - Resolve and inject dependencies with topological sort

**Dependency Declaration**:
```go
type UserServiceImpl struct {
    Config    configmgr.IConfigManager    `inject:""`
    DBMgr     databasemgr.IDatabaseManager `inject:""`
    UserRepo  IUserRepository              `inject:""`
    OrderSvc  IOrderService               `inject:""`  // Same-layer dependency
    Cache     cachemgr.ICacheManager      `inject:"optional"` // Optional dependency
}
```

### Manager Pattern

Managers (`manager/*/`) provide infrastructure capabilities (database, cache, logging, telemetry).

**Standard Manager Structure**:
- `interface.go` - Core interface (extends `BaseManager`)
- `config.go` - Configuration structures and parsing
- `impl_base.go` - Base implementation with observability
- `{driver}_impl.go` - Driver-specific implementations
- `factory.go` - Factory functions for DI

**Configuration Path Convention**:
```
{manager}.driver           # Driver type
{manager}.{driver}_config  # Driver config
```

Examples:
- `database.driver` + `database.mysql_config`
- `cache.driver` + `cache.redis_config`
- `telemetry.driver` + `telemetry.otel_config`

### Server Engine

The `server` package provides the HTTP server lifecycle management:

**Lifecycle Flow**:
1. `NewEngine()` - Create engine with containers
2. `Initialize()` - Auto-inject dependencies, setup Gin
3. `Start()` - Start managers, repositories, services, HTTP server
4. `Stop()` - Graceful shutdown

## Code Style and Conventions

### Imports Order
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

### Naming Conventions
- **Interfaces**: `I*` prefix (e.g., `IConfigManager`, `IDatabaseManager`, `IUserService`)
- **Private structs**: lowercase (e.g., `messageService`, `messageRepository`)
- **Public structs**: PascalCase (e.g., `ServerConfig`, `User`)
- **Functions**: PascalCase exported, camelCase private
- **Factory functions**: `Build()`, `BuildWithConfigProvider()`, `NewXxx()`

### Comments and Documentation
- Use **Chinese** for all user-facing documentation and code comments
- Package documentation in `doc.go`
- Function comments must explain purpose and parameters
- Exported functions must have comments

### Error Handling
```go
if err != nil {
    return "", fmt.Errorf("operation failed: %w", err)
}
```

### Dependency Injection Tags
```go
type MyService struct {
    Config     configmgr.IConfigManager    `inject:""`
    DBMgr      databasemgr.IDatabaseManager `inject:""`
    Optional   cachemgr.ICacheManager      `inject:"optional"`
}
```

### Testing Patterns
- Table-driven tests with `t.Run()` subtests
- Use `testify/assert` for assertions
- Benchmark functions prefixed with `Benchmark`

```go
func TestGenerateToken(t *testing.T) {
    tests := []struct {
        name    string
        input   string
        wantErr bool
    }{
        {"valid input", "data", false},
        {"empty input", "", true},
    }
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // test logic
        })
    }
}
```

## Manager Implementation SOP

When creating or modifying managers, they should be placed at `server/builtin/manager/`:

1. **Flat structure** - No subdirectories within manager package
2. **File organization**:
   - `interface.go` - Core interface (extends `common.IBaseManager`)
   - `config.go` - Config structures and parsing
   - `impl_base.go` - Base implementation with observability
   - `{driver}_impl.go` - Driver-specific implementations
   - `factory.go` - Factory functions for DI
3. **DI tags** - Use `inject:""` for dependencies
4. **Config paths** - Follow `{manager}.driver` convention
5. **Auto-initialization** - Add to `server/builtin/builtin.go:Initialize()`

## Common Development Patterns

### Adding a New Feature
1. Create Entity in `entities/`
2. Create Repository interface and impl in `repositories/`
3. Create Service interface and impl in `services/`
4. Create Controller in `controllers/`
5. Register all in container using `RegisterByType()`
6. Dependencies auto-injected on `InjectAll()`

### Creating a Manager
1. Define interface extending `BaseManager`
2. Implement with `impl_base.go` for observability
3. Create driver implementations
4. Provide `Build()` and `BuildWithConfigProvider()` factory functions
5. Follow config path convention

### Sample Application
See `samples/messageboard/` for a complete working example demonstrating:
- Full layer architecture
- Container registration
- Manager usage
- GORM integration
- Custom routes and middleware

## Configuration

All configuration uses YAML format. Manager components follow this pattern:

```yaml
# Manager configs follow pattern:
database:
  driver: mysql
  mysql_config:
    dsn: "user:pass@tcp(localhost:3306)/db"
    pool_config:
      max_open_conns: 100

cache:
  driver: redis
  redis_config:
    host: localhost
    port: 6379

logger:
  driver: zap
  zap_config:
    level: "info"
    format: "json"

telemetry:
  driver: otel
  otel_config:
    endpoint: localhost:4317
```

Built-in managers: `configmgr`, `loggermgr`, `databasemgr`, `cachemgr`, `telemetrymgr` at `server/builtin/manager/`

## Important Architecture Constraints

1. **No circular dependencies** - Container detects and reports cycles
2. **Layer boundaries enforced** - Upper layers cannot depend on lower layers
3. **Interface-based DI** - Register by interface type, not concrete type
4. **Two-phase injection** - Register first, inject later
5. **Manager lifecycle** - All managers implement `OnStart()/OnStop()`

## Testing Strategy

- Unit tests in `*_test.go` files alongside source
- Use table-driven tests for multiple scenarios
- Mock interfaces using `testify/mock`
- Integration tests in samples
- Benchmark critical paths
