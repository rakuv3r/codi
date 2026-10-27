# Context

Based on [Google Go Best Practices - Context](https://google.github.io/styleguide/go/best-practices#contexts).

## Quick Reference

- **First parameter** - Context is always the first function parameter
- **Don't store** - Never store Context in structs
- **Never pass nil** - Use `context.Background()` or `context.TODO()`
- **Cancellation is cooperative** - Check `ctx.Done()` and return quickly
- **Value for metadata only** - Not a way to pass optional parameters
- **Always defer cancel** - Avoid resource leaks

## Core Rules

### 1. Context as First Parameter

```go
// ✅ Context is always first
func DoWork(ctx context.Context, param1 string, param2 int) error

func (s *Service) Process(ctx context.Context, data *Data) error

// ❌ Context not first
func DoWork(param1 string, ctx context.Context) error
```

### 2. Don't Store Context in Structs

```go
// ❌ Don't store Context
type Service struct {
    ctx    context.Context  // Wrong!
    client *http.Client
}

// ✅ Context as method parameter
type Service struct {
    client *http.Client
}

func (s *Service) Process(ctx context.Context, data *Data) error {
    // Use passed ctx
}
```

**Why**:
- Unclear Context lifetime
- Hard to control cancellation scope
- Breaks cancellation propagation

### 3. Creating Context

```go
// ✅ Use Background as root
ctx := context.Background()

// ✅ Use TODO when unsure
ctx := context.TODO()

// ✅ Manual cancellation
ctx, cancel := context.WithCancel(parentCtx)
defer cancel()

// ✅ Timeout auto-cancellation
ctx, cancel := context.WithTimeout(parentCtx, 5*time.Second)
defer cancel()

// ✅ Specific deadline
ctx, cancel := context.WithDeadline(parentCtx, time.Now().Add(5*time.Second))
defer cancel()
```

### 4. Always defer cancel

```go
// ❌ Forgot cancel leads to resource leak
func ProcessData() error {
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    // Missing defer cancel()
    return doWork(ctx)
}

// ✅ Always defer
func ProcessData() error {
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()  // Ensure cleanup
    return doWork(ctx)
}
```

### 5. Check Cancellation

```go
// ✅ Check in loops
func ProcessItems(ctx context.Context, items []Item) error {
    for _, item := range items {
        if err := ctx.Err(); err != nil {
            return err  // context.Canceled or context.DeadlineExceeded
        }
        
        if err := processItem(item); err != nil {
            return err
        }
    }
    return nil
}

// ✅ Listen in select
func Worker(ctx context.Context) {
    for {
        select {
        case <-ctx.Done():
            return  // Context cancelled
        case work := <-workChan:
            processWork(work)
        }
    }
}
```

### 6. Handle Cancellation Errors

```go
// ✅ Distinguish different errors
err := doWork(ctx)
if err != nil {
    switch {
    case errors.Is(err, context.Canceled):
        return fmt.Errorf("operation canceled: %w", err)
    case errors.Is(err, context.DeadlineExceeded):
        return fmt.Errorf("operation timeout: %w", err)
    default:
        return fmt.Errorf("operation failed: %w", err)
    }
}

// ✅ Or return ctx.Err() directly
func DoWork(ctx context.Context) error {
    select {
    case <-ctx.Done():
        return ctx.Err()  // Preserve Canceled or DeadlineExceeded
    case result := <-workChan:
        return process(result)
    }
}
```

### 7. Context.Value Usage

```go
// ✅ Define type-safe keys
type contextKey int

const (
    userIDKey contextKey = iota
    requestIDKey
)

func WithUserID(ctx context.Context, userID string) context.Context {
    return context.WithValue(ctx, userIDKey, userID)
}

func GetUserID(ctx context.Context) (string, bool) {
    userID, ok := ctx.Value(userIDKey).(string)
    return userID, ok
}

// ❌ String keys (collision risk)
func WithUserID(ctx context.Context, userID string) context.Context {
    return context.WithValue(ctx, "user_id", userID)  // May collide
}
```

**Use only for request-scoped metadata**:
- Request IDs
- Auth tokens
- Tracing info

**Don't use for**:
- Optional parameters
- Business logic data
- Configuration

```go
// ❌ Don't pass parameters via Context
func ProcessData(ctx context.Context) error {
    timeout := ctx.Value("timeout").(time.Duration)  // Wrong!
}

// ✅ Use structs or parameters
type Options struct {
    Timeout time.Duration
    Retries int
}

func ProcessData(ctx context.Context, opts Options) error {
    // ...
}
```

## Production Patterns

### HTTP Handler

```go
// ✅ Use request's Context
func HandleAPI(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()
    
    // Add request ID
    requestID := generateRequestID()
    ctx = WithRequestID(ctx, requestID)
    
    result, err := processRequest(ctx, r)
    if err != nil {
        http.Error(w, err.Error(), 500)
        return
    }
    
    json.NewEncoder(w).Encode(result)
}
```

### Database Transactions

```go
// ✅ Use Context to control timeout
func UpdateUser(ctx context.Context, db *sql.DB, user *User) error {
    tx, err := db.BeginTx(ctx, nil)
    if err != nil {
        return err
    }
    defer tx.Rollback()
    
    _, err = tx.ExecContext(ctx, "UPDATE users SET name = ? WHERE id = ?", user.Name, user.ID)
    if err != nil {
        return err
    }
    
    return tx.Commit()
}
```

### Concurrent Task Coordination

```go
// ✅ Use Context to coordinate goroutines
func ProcessConcurrently(ctx context.Context, items []Item) error {
    ctx, cancel := context.WithCancel(ctx)
    defer cancel()
    
    errCh := make(chan error, len(items))
    
    for _, item := range items {
        go func(item Item) {
            if err := processItem(ctx, item); err != nil {
                errCh <- err
                cancel()  // One fails, cancel all
            } else {
                errCh <- nil
            }
        }(item)
    }
    
    for range items {
        if err := <-errCh; err != nil {
            return err
        }
    }
    
    return nil
}
```

### Graceful Shutdown

```go
// ✅ Use Context for graceful shutdown
func (s *Server) Start(ctx context.Context) error {
    go func() {
        <-ctx.Done()
        
        // 5 second grace period
        shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
        defer cancel()
        
        if err := s.httpServer.Shutdown(shutdownCtx); err != nil {
            log.Printf("server shutdown error: %v", err)
        }
    }()
    
    if err := s.httpServer.ListenAndServe(); err != http.ErrServerClosed {
        return err
    }
    
    return nil
}

// Usage in main
func main() {
    ctx, cancel := context.WithCancel(context.Background())
    defer cancel()
    
    server := &Server{httpServer: &http.Server{Addr: ":8080"}}
    
    // Listen for signals
    sigCh := make(chan os.Signal, 1)
    signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
    
    go func() {
        <-sigCh
        cancel()  // Signal received, cancel context
    }()
    
    if err := server.Start(ctx); err != nil {
        log.Fatal(err)
    }
}
```

## Context Shadowing Trap

```go
// ❌ Shadowing: creates new variable
func Handler(ctx context.Context, req *Request) {
    if *shortenDeadlines {
        ctx, cancel := context.WithTimeout(ctx, 3*time.Second)  // Shadows!
        defer cancel()
    }
    // BUG: ctx here is the original context
}

// ✅ Stomping: reassigns variable
func Handler(ctx context.Context, req *Request) {
    if *shortenDeadlines {
        var cancel func()
        ctx, cancel = context.WithTimeout(ctx, 3*time.Second)  // Stomps
        defer cancel()
    }
    // ctx correctly refers to timeout-wrapped context
}
```

---

**Full details**: [Google Go Best Practices - Context](https://google.github.io/styleguide/go/best-practices#contexts)
