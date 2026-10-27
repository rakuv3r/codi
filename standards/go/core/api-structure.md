# Go API Project Structure

## Overview

This guide covers structure and organization patterns for Go HTTP API projects (REST APIs, microservices, web backends).

---

## Directory Layout

### `/cmd` - Application Entry Points

The `cmd/` directory contains application entry points. For APIs, typically `cmd/server/main.go`.

**Purpose**:
- One subdirectory per executable
- Keep `main.go` minimal (3-10 lines)
- Only responsible for wiring up dependencies and starting the app

### `/internal` - Private Application Code

The `internal/` directory contains your application's core logic.

**Go compiler enforced**: Code in `internal/` cannot be imported by external projects. This enforces encapsulation.

**Organization principle**: Organize by feature/domain (user/, order/), NOT by technical layer (handler/, service/, repository/).

**Why feature-based?**
- High cohesion: all code for one feature together
- Easy maintenance: change one feature = modify one directory
- Clear boundaries: module dependencies are explicit
- Validated by: Kubernetes, Terraform, Docker all use this pattern

---

## Recommended Structure

```
myapi/
├── cmd/
│   └── server/
│       └── main.go              # Application entry point
├── internal/
│   ├── user/                    # Feature module
│   │   ├── handler.go           # HTTP handlers
│   │   ├── handler_test.go      # Handler tests
│   │   ├── repository.go        # Data access
│   │   ├── repository_test.go   # Repository tests (with mocks)
│   │   └── user.go              # Domain types (optional)
│   ├── order/                   # Another feature module
│   │   ├── handler.go
│   │   └── repository.go
│   ├── config/                  # Configuration
│   │   ├── config.go            # App config
│   │   └── database.go          # DB connection pool
│   └── middleware/              # HTTP middleware
│       ├── logger.go            # Request logging
│       ├── cors.go              # CORS handling
│       └── recovery.go          # Panic recovery
├── go.mod
├── go.sum
├── Makefile
└── README.md
```

---

## Feature Module Organization

### Two-Layer Pattern (Recommended)

Keep it simple: handler → repository

```
internal/user/
  handler.go       # HTTP handling + simple logic
  repository.go    # Database operations
  user.go          # Domain types (optional)
```

**handler.go** responsibilities:
- Parse HTTP request
- Validate input
- Call repository
- Return HTTP response
- Handle errors

**repository.go** responsibilities:
- Database queries
- CRUD operations
- Data mapping
- No business logic

### When to Add Service Layer

Only add `service.go` if you have:
- Complex business rules spanning multiple operations
- Need to orchestrate multiple repositories
- Business logic that needs reuse across handlers

```
internal/user/
  handler.go       # HTTP only
  service.go       # Business logic
  repository.go    # Database only
```

**Go community prefers simple layering**. Most projects don't need service layer (see Kubernetes, Docker, Terraform).

---

## HTTP Handling Patterns

### Handler Function vs Method

#### Option 1: Package-level functions (simpler)
```go
package user

func GetUser(c *gin.Context) {
    id := c.Param("id")
    // ...
}

func CreateUser(c *gin.Context) {
    // ...
}
```

#### Option 2: Handler struct with dependencies (recommended)
```go
package user

type Handler struct {
    repo *Repository
}

func NewHandler(repo *Repository) *Handler {
    return &Handler{repo: repo}
}

func (h *Handler) GetUser(c *gin.Context) {
    id := c.Param("id")
    user, err := h.repo.Get(id)
    // ...
}
```

**Use Option 2** for dependency injection and testability.

---

## Routing Registration

### In main.go

```go
func main() {
    // Initialize dependencies
    db := config.InitDB()
    
    // Create repositories
    userRepo := user.NewRepository(db)
    
    // Create handlers
    userHandler := user.NewHandler(userRepo)
    
    // Setup router
    r := gin.Default()
    
    // Apply middleware
    r.Use(middleware.Logger())
    r.Use(middleware.Recovery())
    
    // Register routes
    api := r.Group("/api/v1")
    {
        users := api.Group("/users")
        {
            users.GET("/:id", userHandler.GetUser)
            users.POST("", userHandler.CreateUser)
            users.PATCH("/:id", userHandler.UpdateUser)
            users.DELETE("/:id", userHandler.DeleteUser)
        }
    }
    
    r.Run(":8080")
}
```

**Keep routing in main.go** for small to medium APIs. For large APIs, consider a dedicated router package.

---

## Configuration Management

### config/config.go

```go
package config

import (
    "os"
    "strconv"
)

type Config struct {
    Port     string
    DBHost   string
    DBPort   int
    DBUser   string
    DBPass   string
    DBName   string
}

func Load() *Config {
    return &Config{
        Port:   getEnv("PORT", "8080"),
        DBHost: getEnv("DB_HOST", "localhost"),
        DBPort: getEnvInt("DB_PORT", 5432),
        DBUser: getEnv("DB_USER", "postgres"),
        DBPass: getEnv("DB_PASSWORD", ""),
        DBName: getEnv("DB_NAME", "mydb"),
    }
}

func getEnv(key, defaultVal string) string {
    if val := os.Getenv(key); val != "" {
        return val
    }
    return defaultVal
}

func getEnvInt(key string, defaultVal int) int {
    if val := os.Getenv(key); val != "" {
        if intVal, err := strconv.Atoi(val); err == nil {
            return intVal
        }
    }
    return defaultVal
}
```

**Use environment variables** for configuration. Follow [12-factor app](https://12factor.net/config) principles.

---

## Database Access Patterns

### Repository Pattern

```go
package user

import (
    "database/sql"
    "errors"
)

type Repository struct {
    db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
    return &Repository{db: db}
}

func (r *Repository) Get(id string) (*User, error) {
    var user User
    err := r.db.QueryRow(
        "SELECT id, name, email FROM users WHERE id = $1",
        id,
    ).Scan(&user.ID, &user.Name, &user.Email)
    
    if err == sql.ErrNoRows {
        return nil, errors.New("user not found")
    }
    if err != nil {
        return nil, err
    }
    
    return &user, nil
}

func (r *Repository) Create(user *User) error {
    _, err := r.db.Exec(
        "INSERT INTO users (name, email) VALUES ($1, $2)",
        user.Name, user.Email,
    )
    return err
}

func (r *Repository) Update(id string, user *User) error {
    _, err := r.db.Exec(
        "UPDATE users SET name = $1, email = $2 WHERE id = $3",
        user.Name, user.Email, id,
    )
    return err
}

func (r *Repository) Delete(id string) error {
    _, err := r.db.Exec("DELETE FROM users WHERE id = $1", id)
    return err
}
```

**Use prepared statements** (via QueryRow/Exec parameters) for SQL injection protection.

---

## Error Handling

### Handler-level error handling

```go
func (h *Handler) GetUser(c *gin.Context) {
    id := c.Param("id")
    
    user, err := h.repo.Get(id)
    if err != nil {
        if errors.Is(err, ErrNotFound) {
            c.JSON(404, gin.H{"error": "User not found"})
            return
        }
        c.JSON(500, gin.H{"error": "Internal server error"})
        return
    }
    
    c.JSON(200, user)
}
```

**Don't leak internal errors** to clients. Map to appropriate HTTP status codes.

### Custom error types

```go
package user

import "errors"

var (
    ErrNotFound      = errors.New("user not found")
    ErrDuplicate     = errors.New("user already exists")
    ErrInvalidInput  = errors.New("invalid input")
)
```

---

## Testing

### Handler tests

```go
package user

import (
    "net/http/httptest"
    "testing"
    "github.com/gin-gonic/gin"
    "github.com/stretchr/testify/assert"
)

func TestGetUser(t *testing.T) {
    // Setup
    gin.SetMode(gin.TestMode)
    repo := NewMockRepository() // Use mock
    handler := NewHandler(repo)
    
    // Create test request
    w := httptest.NewRecorder()
    c, _ := gin.CreateTestContext(w)
    c.Params = gin.Params{{Key: "id", Value: "123"}}
    
    // Execute
    handler.GetUser(c)
    
    // Assert
    assert.Equal(t, 200, w.Code)
}
```

### Repository tests with sqlmock

```go
package user

import (
    "testing"
    "github.com/DATA-DOG/go-sqlmock"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)

func TestGetUser(t *testing.T) {
    // Setup mock database
    db, mock, err := sqlmock.New()
    require.NoError(t, err)
    defer db.Close()
    
    repo := NewRepository(db)
    
    // Define expected query and result
    rows := sqlmock.NewRows([]string{"id", "name", "email"}).
        AddRow(1, "John Doe", "john@example.com")
    
    mock.ExpectQuery("SELECT (.+) FROM users WHERE id").
        WithArgs("1").
        WillReturnRows(rows)
    
    // Execute
    user, err := repo.Get("1")
    
    // Assert
    assert.NoError(t, err)
    assert.Equal(t, "John Doe", user.Name)
    assert.NoError(t, mock.ExpectationsWereMet())
}
```

---

## Middleware

### Common middleware patterns

```go
package middleware

import (
    "log"
    "time"
    "github.com/gin-gonic/gin"
)

// Logger middleware
func Logger() gin.HandlerFunc {
    return func(c *gin.Context) {
        start := time.Now()
        path := c.Request.URL.Path
        method := c.Request.Method
        
        c.Next()
        
        latency := time.Since(start)
        status := c.Writer.Status()
        
        log.Printf("[%d] %s %s %v", status, method, path, latency)
    }
}

// CORS middleware
func CORS() gin.HandlerFunc {
    return func(c *gin.Context) {
        c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
        c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PATCH, DELETE")
        c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
        
        if c.Request.Method == "OPTIONS" {
            c.AbortWithStatus(204)
            return
        }
        
        c.Next()
    }
}

// Recovery middleware
func Recovery() gin.HandlerFunc {
    return func(c *gin.Context) {
        defer func() {
            if err := recover(); err != nil {
                log.Printf("Panic recovered: %v", err)
                c.JSON(500, gin.H{"error": "Internal server error"})
            }
        }()
        c.Next()
    }
}
```

---

## API Versioning

### URL-based versioning (recommended)

```go
v1 := r.Group("/api/v1")
{
    v1.GET("/users/:id", userHandler.GetUser)
}

v2 := r.Group("/api/v2")
{
    v2.GET("/users/:id", userHandlerV2.GetUser)
}
```

Keep old versions in separate packages if they differ significantly.

---

## Summary

✅ **DO**:
- Organize by feature modules (internal/user/, internal/order/)
- Keep layering simple (handler → repository)
- Use dependency injection via constructor functions
- Test handlers and repositories separately with mocks
- Use environment variables for configuration
- Handle errors at handler level, don't leak internals
- Use prepared statements for SQL

❌ **DON'T**:
- Add service layer by default (only when needed)
- Organize by technical layers (handler/, service/, repository/)
- Leak internal errors to API clients
- Hardcode configuration values
- Skip writing tests
- Use string concatenation for SQL queries

---

**Remember**: Go's strength is simplicity. Start with the two-layer pattern and add complexity only when you have a concrete need.
