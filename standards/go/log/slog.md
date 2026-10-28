---
language: go
category: dependency
applies_to: []
---

# Logging

Go-specific logging patterns using standard library `log/slog`.

Based on [log/slog](https://pkg.go.dev/log/slog) (Go 1.21+).

## Quick Reference

- **log/slog** - structured logging (Go 1.21+)
- **JSON handler** - production format
- **Text handler** - development format
- **Context-aware** - pass logger in context
- **No sensitive data** - never log passwords, tokens

## Basic Setup

### JSON Handler (Production)

```go
import (
    "log/slog"
    "os"
)

// ✅ JSON logging for production
func NewLogger() *slog.Logger {
    handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
        Level: slog.LevelInfo,
        AddSource: true, // Include file:line
    })
    return slog.New(handler)
}

// Usage
logger := NewLogger()
logger.Info("server started", "port", 8080, "env", "production")
// Output: {"time":"2026-10-15T10:30:00Z","level":"INFO","source":{"file":"main.go","line":42},"msg":"server started","port":8080,"env":"production"}
```

### Text Handler (Development)

```go
// ✅ Human-readable for development
func NewDevLogger() *slog.Logger {
    handler := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
        Level: slog.LevelDebug,
    })
    return slog.New(handler)
}

// Output: time=2026-10-15T10:30:00Z level=INFO msg="server started" port=8080
```

## Log Levels

```go
// ✅ Use appropriate levels
logger.Debug("detailed info", "query", query)      // Development only
logger.Info("normal operation", "user_id", userID)  // General info
logger.Warn("unexpected but handled", "error", err) // Warnings
logger.Error("operation failed", "error", err)      // Errors

// ✅ Check level before expensive operations
if logger.Enabled(context.Background(), slog.LevelDebug) {
    logger.Debug("expensive debug info", "data", expensiveOperation())
}
```

## Structured Logging

```go
// ✅ Key-value pairs
logger.Info("user login",
    "user_id", userID,
    "ip", clientIP,
    "duration_ms", duration.Milliseconds(),
)

// ✅ Group related fields
logger.Info("request completed",
    slog.Group("request",
        slog.String("method", "GET"),
        slog.String("path", "/users/123"),
        slog.Int("status", 200),
    ),
    slog.Group("timing",
        slog.Duration("duration", duration),
        slog.Int64("db_ms", dbTime.Milliseconds()),
    ),
)
```

## Context-Aware Logging

```go
// ✅ Store logger in context
func WithLogger(ctx context.Context, logger *slog.Logger) context.Context {
    return context.WithValue(ctx, loggerKey, logger)
}

func LoggerFromContext(ctx context.Context) *slog.Logger {
    if logger, ok := ctx.Value(loggerKey).(*slog.Logger); ok {
        return logger
    }
    return slog.Default()
}

// ✅ Add request ID to logger
func RequestIDMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        requestID := uuid.New().String()
        
        logger := slog.Default().With("request_id", requestID)
        ctx := WithLogger(c.Request.Context(), logger)
        c.Request = c.Request.WithContext(ctx)
        
        c.Next()
    }
}

// Usage in handler
func GetUser(c *gin.Context) {
    logger := LoggerFromContext(c.Request.Context())
    logger.Info("fetching user", "user_id", c.Param("id"))
}
```

## Gin Integration

```go
// ✅ Gin logging middleware
func LoggingMiddleware(logger *slog.Logger) gin.HandlerFunc {
    return func(c *gin.Context) {
        start := time.Now()
        path := c.Request.URL.Path
        query := c.Request.URL.RawQuery

        c.Next()

        duration := time.Since(start)
        
        logger.Info("request completed",
            slog.String("method", c.Request.Method),
            slog.String("path", path),
            slog.String("query", query),
            slog.Int("status", c.Writer.Status()),
            slog.Duration("duration", duration),
            slog.String("ip", c.ClientIP()),
            slog.String("user_agent", c.Request.UserAgent()),
        )
    }
}

// Setup
router := gin.New()
logger := NewLogger()
router.Use(LoggingMiddleware(logger))
```

## Sensitive Data

```go
// ❌ Never log sensitive data
logger.Info("user login", "password", password) // WRONG!
logger.Info("payment", "credit_card", cardNumber) // WRONG!
logger.Info("auth", "token", jwtToken) // WRONG!

// ✅ Log safe identifiers only
logger.Info("user login", "user_id", userID)
logger.Info("payment", "amount", amount, "currency", "USD")
logger.Info("auth", "user_id", userID)

// ✅ Mask if you must log
func maskEmail(email string) string {
    parts := strings.Split(email, "@")
    if len(parts) != 2 {
        return "***"
    }
    return parts[0][:1] + "***@" + parts[1]
}

logger.Info("user registered", "email", maskEmail(email))
// Output: email=a***@example.com
```

## Error Logging

```go
// ✅ Log error with context
if err != nil {
    logger.Error("failed to get user",
        "error", err,
        "user_id", userID,
    )
    return err
}

// ✅ Log error with stack trace (using custom attribute)
type errorWithStack struct {
    error
    stack string
}

func logErrorWithStack(logger *slog.Logger, msg string, err error) {
    logger.Error(msg,
        "error", err.Error(),
        "stack", string(debug.Stack()),
    )
}
```

## Custom Handler

```go
// ✅ Add custom fields to all logs
type ContextHandler struct {
    handler slog.Handler
    attrs   []slog.Attr
}

func NewContextHandler(handler slog.Handler, attrs []slog.Attr) *ContextHandler {
    return &ContextHandler{
        handler: handler,
        attrs:   attrs,
    }
}

func (h *ContextHandler) Enabled(ctx context.Context, level slog.Level) bool {
    return h.handler.Enabled(ctx, level)
}

func (h *ContextHandler) Handle(ctx context.Context, r slog.Record) error {
    for _, attr := range h.attrs {
        r.AddAttrs(attr)
    }
    return h.handler.Handle(ctx, r)
}

func (h *ContextHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
    return NewContextHandler(h.handler, append(h.attrs, attrs...))
}

func (h *ContextHandler) WithGroup(name string) slog.Handler {
    return NewContextHandler(h.handler.WithGroup(name), h.attrs)
}

// Usage: add service name to all logs
baseHandler := slog.NewJSONHandler(os.Stdout, nil)
handler := NewContextHandler(baseHandler, []slog.Attr{
    slog.String("service", "user-api"),
    slog.String("version", "1.0.0"),
})
logger := slog.New(handler)
```

## Performance

```go
// ✅ Use With() for repeated attributes
userLogger := logger.With("user_id", userID)
userLogger.Info("login successful")
userLogger.Info("profile updated")
// Both logs include user_id

// ✅ Lazy evaluation for expensive operations
logger.LogAttrs(context.Background(), slog.LevelDebug,
    "debug info",
    slog.Any("data", func() interface{} {
        // Only called if debug level enabled
        return expensiveOperation()
    }()),
)
```

## Best Practices

```go
// ✅ One logger per application
var logger = NewLogger()

// ✅ Use slog.Default() for libraries
slog.Info("library operation")

// ✅ Include context
logger.Info("operation",
    "user_id", userID,
    "action", "update",
    "resource", resourceID,
)

// ❌ Don't use fmt.Println
fmt.Println("user logged in") // WRONG

// ❌ Don't use log package
log.Println("user logged in") // WRONG (unless wrapping slog)
```

---

**References**:
- [log/slog Package](https://pkg.go.dev/log/slog)
- [Go 1.21 slog Guide](https://go.dev/blog/slog)
