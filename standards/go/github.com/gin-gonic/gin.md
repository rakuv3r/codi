---
language: go
category: dependency
applies_to: [api]
---

# Gin Web Framework Standards

## Overview

Gin is the most popular Go web framework (75k+ stars), known for fast routing and rich middleware ecosystem. It provides HTTP request/response handling, validation, JSON rendering, and graceful error handling.

## Core Principles

1. **ShouldBind over Bind** - Control error responses yourself
2. **c.Error() for error collection** - Centralized error handling  
3. **Context for request scope** - c.Set/c.Get for middleware communication
4. **Graceful shutdown required** - 5-10 second timeout
5. **Middleware order matters** - Logger → Recovery → Custom

## Basic Setup

```go
func main() {
    r := gin.Default()  // Includes Logger() + Recovery()
    
    // Or without defaults:
    // r := gin.New()
    // r.Use(gin.Logger())
    // r.Use(gin.Recovery())
    
    r.GET("/ping", func(c *gin.Context) {
        c.JSON(200, gin.H{"message": "pong"})
    })
    
    r.Run(":8080")
}
```

## Request Binding and Validation

### ShouldBind vs Bind (Critical)

**Bind** - Auto-abort on error:

```go
var json LoginRequest
if err := c.BindJSON(&json); err != nil {
    // Won't reach here - Gin already returned 400
}
```

**ShouldBind** - Manual control (Recommended):

```go
var json LoginRequest
if err := c.ShouldBindJSON(&json); err != nil {
    c.JSON(400, gin.H{
        "error": err.Error(),
        "code":  "INVALID_INPUT",
    })
    return
}
```

### Struct Tags and Validation

```go
type CreateUserRequest struct {
    Name     string `json:"name" binding:"required,min=2,max=50"`
    Email    string `json:"email" binding:"required,email"`
    Age      int    `json:"age" binding:"gte=0,lte=130"`
    Password string `json:"password" binding:"required,min=8"`
}
```

### Custom Validators

```go
import (
    "github.com/gin-gonic/gin/binding"
    "github.com/go-playground/validator/v10"
)

var usernameValidator validator.Func = func(fl validator.FieldLevel) bool {
    username := fl.Field().String()
    matched, _ := regexp.MatchString(`^[a-zA-Z0-9_]{3,20}$`, username)
    return matched
}

// Register
if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
    v.RegisterValidation("username", usernameValidator)
}
```

## Error Handling (Centralized)

```go
// Error collection
func getUser(c *gin.Context) {
    user, err := userRepo.Get(c.Param("id"))
    if err != nil {
        c.Error(err)  // Add to error list
        return
    }
    c.JSON(200, user)
}

// Error handler middleware
func ErrorHandler() gin.HandlerFunc {
    return func(c *gin.Context) {
        c.Next()  // Execute handlers first
        
        if len(c.Errors) > 0 {
            err := c.Errors.Last().Err
            
            code := 500
            message := err.Error()
            
            switch e := err.(type) {
            case *AppError:
                code = e.StatusCode
                message = e.Message
            }
            
            log.Error(err)
            c.JSON(code, gin.H{"success": false, "error": message})
        }
    }
}

// Register before routes
r.Use(ErrorHandler())
```

## Middleware

```go
func MyMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        // Before
        start := time.Now()
        
        c.Next()  // Process request
        
        // After
        latency := time.Since(start)
        log.Printf("Latency: %v", latency)
    }
}

// Early termination
func AuthRequired() gin.HandlerFunc {
    return func(c *gin.Context) {
        token := c.GetHeader("Authorization")
        if !isValid(token) {
            c.AbortWithStatusJSON(401, gin.H{"error": "unauthorized"})
            return  // Stop execution
        }
        c.Next()
    }
}
```

## Context Usage

```go
// Middleware sets value
func AuthMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        userID := extractUserID(c)
        c.Set("user_id", userID)
        c.Next()
    }
}

// Handler retrieves value
func getProfile(c *gin.Context) {
    userID := c.GetInt("user_id")
    // Use userID
}
```

## Graceful Shutdown

```go
func main() {
    r := gin.Default()
    
    srv := &http.Server{Addr: ":8080", Handler: r}
    
    go func() {
        if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
            log.Fatalf("listen: %s\n", err)
        }
    }()
    
    quit := make(chan os.Signal, 1)
    signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
    <-quit
    
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()
    
    if err := srv.Shutdown(ctx); err != nil {
        log.Fatal("Server forced to shutdown:", err)
    }
}
```

## Testing

```go
func TestGetUser(t *testing.T) {
    gin.SetMode(gin.TestMode)
    
    r := gin.New()
    r.GET("/users/:id", getUser)
    
    req := httptest.NewRequest("GET", "/users/1", nil)
    w := httptest.NewRecorder()
    
    r.ServeHTTP(w, req)
    
    assert.Equal(t, 200, w.Code)
}
```

## Best Practices

1. **ShouldBind over Bind** - Custom error responses
2. **Centralize errors** - c.Error() + middleware
3. **Graceful shutdown** - 5-10 second timeout
4. **Release mode** - Production deployments
5. **Handler factory pattern** - Dependency injection

## References

- GitHub: https://github.com/gin-gonic/gin
- Docs: https://gin-gonic.com
