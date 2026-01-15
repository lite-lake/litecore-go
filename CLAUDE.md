# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

**LiteCore-Go** is a layered application framework for Go that provides dependency injection containers, lifecycle management, and infrastructure managers. Built on top of Gin, GORM, and Zap.

**Module**: `com.litelake.litecore`
**Go Version**: 1.25+
**Architecture**: 7-tier layered dependency injection

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

### 7-Tier Layered Architecture

The framework enforces strict layer boundaries with单向依赖 (unidirectional dependencies):

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
├─────────────────────────────────────────────────────────────┤
│  Config Layer       (BaseConfigProvider)                    │
└─────────────────────────────────────────────────────────────┘
```

**Dependency Rules**:
- **Config** - No dependencies
- **Entity** - No dependencies
- **Manager** → Config, other Managers (同层依赖/same-layer dependencies)
- **Repository** → Config, Manager, Entity
- **Service** → Config, Manager, Repository, other Services (同层依赖)
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
    Config    common.BaseConfigProvider `inject:""`
    DBMgr     databasemgr.DatabaseManager `inject:""`
    UserRepo  UserRepository  `inject:""`
    OrderSvc  OrderService    `inject:""`  // Same-layer dependency
    Cache     CacheManager    `inject:"optional"` // Optional dependency
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

    "com.litelake.litecore/common"  // local modules last
)
```

### Naming Conventions
- **Interfaces**: `XxxManager` (no I prefix for managers), `XxxService` (no I prefix)
- **Private structs**: `xxxManagerImpl`, `xxxServiceImpl` (lowercase impl)
- **Public structs**: `ServerConfig`, `User` (PascalCase)
- **Functions**: PascalCase exported, camelCase private
- **Factory functions**: `Build()`, `BuildWithConfigProvider()`, `NewXxxImpl()`

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
    Config     common.BaseConfigProvider `inject:""`
    DBMgr      databasemgr.DatabaseManager `inject:""`
    Optional   CacheManager `inject:"optional"`
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

When creating or modifying managers, follow the pattern in `docs/SOP-manager-refactoring.md`:

1. **Flat structure** - No `internal/` subdirectories
2. **File organization**:
   - `interface.go` - Core interface
   - `config.go` - Config structures
   - `impl_base.go` - Base with observability
   - `{driver}_impl.go` - Driver implementations
   - `factory.go` - Build functions
3. **DI tags** - Use `inject:""` for dependencies
4. **Config paths** - Follow `{manager}.driver` convention

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

All configuration uses YAML format with ConfigProvider:

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

telemetry:
  driver: otel
  otel_config:
    endpoint: localhost:4317
```

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
