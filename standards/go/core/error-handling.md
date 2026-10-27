# Error Handling

Based on [Google Go Best Practices - Errors](https://google.github.io/styleguide/go/best-practices#errors).

## Quick Reference

- **Errors are values** - created and handled intentionally, not exceptions
- **Propagate up** - return errors and let callers decide how to handle
- **Add context** - wrap with `fmt.Errorf("context: %w", err)`
- **Check errors** - use `errors.Is()` and `errors.As()`, not string matching
- **Don't double-log** - either handle or return, not both

## Core Rules

### 1. Error Naming

```go
// ✅ Sentinel errors: Err prefix
var (
    ErrNotFound    = errors.New("not found")
    ErrDuplicate   = errors.New("duplicate entry")
    ErrTimeout     = errors.New("operation timeout")
)

// ✅ Error types: Error suffix
type ValidationError struct {
    Field string
    Value interface{}
}

func (e *ValidationError) Error() string {
    return fmt.Sprintf("invalid %s: %v", e.Field, e.Value)
}

// ❌ Bad
var NotFoundError = errors.New("...")  // Inconsistent
type ErrValidation struct { ... }      // Not Error suffix
```

### 2. Add Context

```go
// ✅ Wrap with %w to preserve original error
func LoadUser(id int) (*User, error) {
    user, err := db.Query(id)
    if err != nil {
        return nil, fmt.Errorf("load user %d: %w", id, err)
    }
    return user, nil
}

// ✅ Clear call chain
// Output: load user 123: query failed: connection refused

// ❌ Bad: lose original error
return nil, fmt.Errorf("load user failed")  // Context lost

// ❌ Bad: use %v instead of %w
return nil, fmt.Errorf("load user: %v", err)  // Can't unwrap
```

### 3. Check Errors

```go
// ✅ Use errors.Is for sentinel errors
if errors.Is(err, ErrNotFound) {
    // handle not found
}

// ✅ Use errors.As for error types
var ve *ValidationError
if errors.As(err, &ve) {
    fmt.Printf("validation failed on field: %s\n", ve.Field)
}

// ❌ Bad: string matching
if strings.Contains(err.Error(), "not found") {  // Fragile
```

### 4. Don't Double-Log

```go
// ✅ Choose one: handle or return
func ProcessFile(name string) error {
    data, err := os.ReadFile(name)
    if err != nil {
        return fmt.Errorf("process file: %w", err)  // Return, let caller log
    }
    // ...
}

// ❌ Bad: both log and return
func ProcessFile(name string) error {
    data, err := os.ReadFile(name)
    if err != nil {
        log.Printf("read file failed: %v", err)  // Log
        return err                                // And return
    }
    // Result: same error logged multiple times
}
```

### 5. Error Handling Position

```go
// ✅ Handle errors first, happy path unindented
func Save(data []byte) error {
    if err := validate(data); err != nil {
        return fmt.Errorf("validate: %w", err)
    }
    if err := write(data); err != nil {
        return fmt.Errorf("write: %w", err)
    }
    return nil  // Happy path at the end
}

// ❌ Bad: happy path nested
func Save(data []byte) error {
    if err := validate(data); err == nil {  // Inverted check
        if err := write(data); err == nil {
            return nil
        } else {
            return err
        }
    } else {
        return err
    }
}
```

## Panic Usage

```go
// ✅ Panic only on unrecoverable errors
func MustConnect(dsn string) *DB {
    db, err := Connect(dsn)
    if err != nil {
        panic(fmt.Sprintf("database connection failed: %v", err))
    }
    return db
}

// ✅ Use Must functions only in main/init
func main() {
    db := MustConnect(os.Getenv("DSN"))
    defer db.Close()
    // ...
}

// ❌ Bad: panic in business logic
func ProcessOrder(order *Order) {
    if order.Amount < 0 {
        panic("negative amount")  // Should return error
    }
}
```

## Error Variable Scope

```go
// ✅ Limit err scope
if err := file.Close(); err != nil {
    return fmt.Errorf("close: %w", err)
}

// ❌ Bad: err pollutes outer scope
err := file.Close()
if err != nil {
    return fmt.Errorf("close: %w", err)
}
// err still visible
```

---

**Full details**: [Google Go Best Practices - Errors](https://google.github.io/styleguide/go/best-practices#errors)
