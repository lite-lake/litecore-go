# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

**LiteCore-Go** is a layered application framework for Go that provides dependency injection containers, lifecycle management, and infrastructure managers. Built on top of Gin, GORM, and Zap.

**Module**: `github.com/lite-lake/litecore-go`
**Go Version**: 1.25+
**Architecture**: 5-tier layered dependency injection (Entity → Repository → Service → Controller/Middleware)
**Built-in Components**:
- **Manager Components**: `manager/*/` (configmgr, loggermgr, databasemgr, cachemgr, telemetrymgr, limitermgr, lockmgr, mqmgr)
- **Component Layer**: `component/litecontroller`, `component/litemiddleware`, `component/liteservice`

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
│  Controller Layer   (component/litecontroller)             │
│  Middleware Layer   (component/litemiddleware)               │
├─────────────────────────────────────────────────────────────┤
│  Service Layer      (component/liteservice)                 │
├─────────────────────────────────────────────────────────────┤
│  Repository Layer   (BaseRepository)                        │
├─────────────────────────────────────────────────────────────┤
│  Entity Layer       (BaseEntity)                            │
│  Manager Layer      (manager/*)                            │
│  Managers: configmgr, loggermgr, databasemgr, cachemgr,      │
│           telemetrymgr, limitermgr, lockmgr, mqmgr         │
└─────────────────────────────────────────────────────────────┘
```

**Dependency Rules**:
- **Entity** - No dependencies
- **Manager** → Config, other Managers (same-layer dependencies)
- **Repository** → Config, Manager, Entity
- **Service** → Config, Manager, Repository, other Services (same-layer)
- **Controller** → Config, Manager, Service
- **Middleware** → Config, Manager, Service

**Manager Components** (`manager/*/`):
- `configmgr` - Configuration management (YAML/JSON)
- `loggermgr` - Structured logging (Zap with Gin/JSON/default formats)
- `databasemgr` - Database management (GORM: MySQL/PostgreSQL/SQLite)
- `cachemgr` - Cache management (Ristretto/Redis/None)
- `telemetrymgr` - OpenTelemetry (Traces/Metrics/Logs)
- `limitermgr` - Rate limiting (Memory/Redis)
- `lockmgr` - Distributed lock (Memory/Redis)
- `mqmgr` - Message queue (RabbitMQ/Memory)

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
}
```

### Manager Pattern

Managers (`manager/*/`) provide infrastructure capabilities (database, cache, logging, telemetry, rate limiting, locking, messaging).

**Available Managers**:
- `configmgr` - Configuration loader (YAML/JSON) with path query support
- `loggermgr` - Structured logging with Gin/JSON/default formats, color support
- `databasemgr` - GORM database (MySQL/PostgreSQL/SQLite)
- `cachemgr` - High-performance cache (Ristretto for memory, Redis for distributed, None)
- `telemetrymgr` - OpenTelemetry integration (Traces, Metrics, Logs)
- `limitermgr` - Rate limiting (sliding window, Memory/Redis)
- `lockmgr` - Distributed lock (blocking/non-blocking, Memory/Redis)
- `mqmgr` - Message queue (RabbitMQ, Memory queue)

**Standard Manager Structure**:
- `interface.go` - Core interface (extends `common.IBaseManager`)
- `config.go` - Configuration structures and parsing
- `impl_base.go` - Base implementation with observability
- `{driver}_impl.go` - Driver-specific implementations
- `factory.go` - Factory functions for DI

**Configuration Path Convention**:
```
{manager}.driver           # Driver type (e.g., mysql, redis, rabbitmq)
{manager}.{driver}_config  # Driver config
```

Examples:
- `database.driver` + `database.mysql_config`
- `cache.driver` + `cache.redis_config` / `cache.memory_config`
- `logger.driver` + `logger.zap_config`
- `limiter.driver` + `limiter.redis_config` / `limiter.memory_config`
- `lock.driver` + `lock.redis_config` / `lock.memory_config`
- `mq.driver` + `mq.rabbitmq_config`

### Server Engine

The `server` package provides the HTTP server lifecycle management:

**Lifecycle Flow**:
1. `Initialize()` - Auto-initialize managers (config → telemetry → logger → database → cache → lock → limiter → mq), register to container
2. `NewEngine()` - Create engine with containers
3. `Start()` - Start managers, repositories, services, HTTP server with startup logs
4. `Stop()` - Graceful shutdown

**Startup Logs**:
The framework logs each startup phase:
- Config file and driver info
- Manager initialization status
- Component counts (controllers, middlewares, services)
- Dependency injection results

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

### Logging Best Practices
- **Inject ILoggerManager** in business layers: `LoggerMgr loggermgr.ILoggerManager \`inject:""\``
- **Initialize logger**: Use `s.logger = s.LoggerMgr.Ins()` after DI
- **Structured logging**: `s.logger.Info("msg", "key", value)`
- **Context-aware**: `s.logger.With("user_id", id).Info("...")`
- **Log levels**:
  - Debug: Development debug information
  - Info: Normal business flow (request start/complete, resource creation)
  - Warn: Degradation, slow query, retry
  - Error: Business error, operation failure (requires attention)
  - Fatal: Fatal error, needs immediate termination

**Log Formats** (configurable via `logger.zap_config.console_config.format`):
- `gin` - Gin style with pipe separators, console-friendly (default)
- `json` - JSON format for log analysis and monitoring
- `default` - Default ConsoleEncoder format

**Gin Format Example**:
```
2026-01-24 15:04:05.123 | INFO  | 开始依赖注入 | count=23
2026-01-24 15:04:05.456 | WARN  | 慢查询检测 | duration=1.2s
2026-01-24 15:04:05.789 | ERROR | 数据库连接失败 | error="connection refused"
```

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
    CacheMgr   cachemgr.ICacheManager      `inject:""`
    LimiterMgr limitermgr.ILimiterManager `inject:""`
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

When creating or modifying managers, they should be placed at `manager/*/`:

1. **Flat structure** - No subdirectories within manager package
2. **File organization**:
   - `interface.go` - Core interface (extends `common.IBaseManager`)
   - `config.go` - Config structures and parsing
   - `impl_base.go` - Base implementation with observability
   - `{driver}_impl.go` - Driver-specific implementations
   - `factory.go` - Factory functions for DI
3. **DI tags** - Use `inject:""` for dependencies
4. **Config paths** - Follow `{manager}.driver` convention
5. **Auto-initialization** - Add to `server/builtin.go:Initialize()` in the proper order

**Manager Initialization Order** (dependencies matter):
1. configmgr (must be first)
2. telemetrymgr
3. loggermgr
4. databasemgr
5. cachemgr
6. lockmgr
7. limitermgr
8. mqmgr

## Common Development Patterns

### Adding a New Feature
1. Create Entity in `entities/`
2. Create Repository interface and impl in `repositories/`
3. Create Service interface and impl in `services/`
4. Create Controller in `controllers/`
5. Register all in container using `RegisterByType()`
6. Dependencies auto-injected on `InjectAll()`

### Creating a Manager
1. Define interface extending `common.IBaseManager`
2. Implement with `impl_base.go` for observability
3. Create driver implementations (memory, redis, etc.)
4. Provide `Build()` and `BuildWithConfigProvider()` factory functions
5. Follow config path convention (`{manager}.driver`, `{manager}.{driver}_config`)
6. Add to `server/builtin.go:Initialize()` in proper order

### Using Built-in Components
Controllers, middlewares, and services are available in `component/`:
- `component/litecontroller` - Health, Metrics, Pprof, Resource controllers
- `component/litemiddleware` - CORS, Recovery, RequestLogger, SecurityHeaders, RateLimiter, Telemetry
- `component/liteservice` - HTMLTemplateService

```go
// Register built-in middleware with default config
cors := litemiddleware.NewCorsMiddleware(nil)
recovery := litemiddleware.NewRecoveryMiddleware(nil)
limiter := litemiddleware.NewRateLimiterMiddleware(nil)
middlewareContainer.RegisterMiddleware(cors)
middlewareContainer.RegisterMiddleware(recovery)
middlewareContainer.RegisterMiddleware(limiter)

// Custom middleware config
allowOrigins := []string{"https://example.com"}
customCors := litemiddleware.NewCorsMiddleware(&litemiddleware.CorsConfig{
    AllowOrigins: &allowOrigins,
})
```

### Sample Application
See `samples/messageboard/` for a complete working example demonstrating:
- Full layer architecture
- Container registration
- Manager usage (database, cache, limiter, lock, mq)
- GORM integration with Ristretto cache
- Built-in middleware (CORS, RateLimiter, Telemetry)
- Custom routes and middleware

## Configuration

All configuration uses YAML format. Manager components follow this pattern:

```yaml
# Manager configs follow pattern:
database:
  driver: mysql
  mysql_config:
    host: "localhost"
    port: 3306
    database: "mydb"
    username: "root"
    password: "password"

cache:
  driver: memory  # memory | redis | none
  memory_config:
    max_size: 100        # MB
    max_age: 720h        # 30 days
    compress: false
  # redis_config:
  #   host: localhost
  #   port: 6379

logger:
  driver: zap
  zap_config:
    console_enabled: true
    console_config:
      level: "info"
      format: "gin"      # gin | json | default
      color: true
      time_format: "2006-01-02 15:04:05.000"
    file_enabled: false
    file_config:
      level: "info"
      path: "./logs/app.log"
      rotation:
        max_size: 100
        max_age: 30
        max_backups: 10
        compress: true

limiter:
  driver: memory  # memory | redis
  memory_config:
    max_backups: 1000

lock:
  driver: redis  # memory | redis
  redis_config:
    host: localhost
    port: 6379
    db: 0

mq:
  driver: rabbitmq  # rabbitmq | memory
  rabbitmq_config:
    url: "amqp://guest:guest@localhost:5672/"
    durable: true

telemetry:
  driver: otel
  otel_config:
    endpoint: "http://localhost:4318"
    enabled_traces: true
    enabled_metrics: true
    enabled_logs: true
```

**Built-in managers** (`manager/*/`):
- `configmgr` - Config management
- `loggermgr` - Logging with Gin/JSON/default formats
- `databasemgr` - Database (MySQL/PostgreSQL/SQLite)
- `cachemgr` - Cache (Ristretto/Redis/None)
- `telemetrymgr` - OpenTelemetry
- `limitermgr` - Rate limiting
- `lockmgr` - Distributed lock
- `mqmgr` - Message queue

**Built-in components** (`component/*/`):
- `litecontroller` - Health, Metrics, Pprof controllers
- `litemiddleware` - CORS, Recovery, RequestLogger, SecurityHeaders, RateLimiter, Telemetry
- `liteservice` - HTMLTemplateService

## Important Architecture Constraints

1. **No circular dependencies** - Container detects and reports cycles
2. **Layer boundaries enforced** - Upper layers cannot depend on lower layers
3. **Interface-based DI** - Register by interface type, not concrete type
4. **Two-phase injection** - Register first, inject later
5. **Manager lifecycle** - All managers implement `OnStart()/OnStop()`
6. **Manager initialization order** - config → telemetry → logger → database → cache → lock → limiter → mq
7. **Middleware execution order** - Recovery (0) → RequestLogger (50) → CORS (100) → SecurityHeaders (150) → RateLimiter (200) → Telemetry (250)
8. **Component paths** - Managers in `manager/*/`, Components in `component/litecontroller`, `component/litemiddleware`, `component/liteservice`

## Testing Strategy

- Unit tests in `*_test.go` files alongside source
- Use table-driven tests for multiple scenarios
- Mock interfaces using `testify/mock`
- Integration tests in samples
- Benchmark critical paths

## Related Documentation

- **AGENTS.md** - Development guidelines for AI agents (coding standards, logging, architecture)
- **manager/README.md** - Manager component documentation (detailed API and usage)
- **component/README.md** - Built-in components documentation
- **component/litemiddleware/README.md** - Middleware configuration guide
