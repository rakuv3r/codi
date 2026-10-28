---
language: go
category: core
applies_to: []
---

# Comments

Based on [Google Go Style Guide - Commentary](https://google.github.io/styleguide/go/decisions#commentary).

## Quick Reference

- **Comments say "why"** - code says "what", comments say "why"
- **Good naming beats comments** - prefer clear naming
- **Exported symbols must have comments** - all exported packages, functions, types
- **Start with name** - comment starts with declared name
- **Complete sentences** - capitalize first letter, end with period

## Core Rules

### 1. Package Comments

```go
// ✅ Package comment: concise description
// Package storage provides database client for data persistence.
// Supports batch insertion, filtered queries, and data retrieval.
package storage

// ❌ Doesn't start with Package
// Storage client for database
package storage

// ❌ After imports
package storage

import "context"

// Storage package
```

**Package comment requirements**:
- Start with `Package <name>`
- First sentence is summary (shown on godoc homepage)
- Complete sentences
- Before any imports

### 2. Function Comments

```go
// ✅ Exported function: must comment, start with function name
// SaveConnections saves a batch of connections to the database.
// It groups connections by type and processes each group separately.
// Returns an error if any batch fails.
func (c *Client) SaveConnections(ctx context.Context, connections []*Connection) error

// ❌ Doesn't start with function name
// Saves connections to the database
func SaveConnections(...)

// ❌ Repeats function signature
// SaveConnections saves connections []*Connection to database and returns error
func SaveConnections(ctx context.Context, connections []*Connection) error
```

**Document special behavior**:
```go
// ✅ Explain return value
// QueryConnections queries connections with optional filters.
// Returns an empty slice (not nil) if no connections match.
func QueryConnections(ctx context.Context, query *Query) ([]*Connection, error)

// ✅ Document concurrency safety
// Get retrieves a value from the cache.
// This method is safe for concurrent use.
func (c *Cache) Get(key string) (any, bool)

// ✅ Document side effects
// Close closes the database connection and releases resources.
// The client cannot be used after Close is called.
func (c *Client) Close(ctx context.Context) error
```

### 3. Type Comments

```go
// ✅ Struct comment
// Client is a database client for managing application data.
// Provides save, query, and data retrieval methods.
type Client struct {
    driver         Driver
    logger         *slog.Logger
    batchSize      int
    queryLimit     int
}

// ✅ Interface comment
// Storage defines the interface for connection data persistence.
// Implementations must be safe for concurrent use.
type Storage interface {
    Save(ctx context.Context, conn *Connection) error
    Query(ctx context.Context, filter *Filter) ([]*Connection, error)
}
```

### 4. Field Comments

```go
// ✅ Preferred: above-line comments (extracted by godoc)
type Config struct {
    // URI is the database connection URI (e.g., postgres://localhost:5432)
    URI string
    
    // BatchSize is the maximum number of records per transaction.
    // Defaults to 1000.
    BatchSize int
    
    // QueryLimit limits the number of records returned in queries
    // to prevent timeouts in large datasets.
    QueryLimit int
}

// ✅ Optional: inline comments (brief only, not extracted by godoc)
type Batch struct {
    AgentID  string  `json:"agent_id"`   // Agent unique identifier
    NodeName string  `json:"node_name"`  // Node name
}
```

### 5. Inline Comments

```go
// ✅ Explain why
func processData(data []byte) error {
    // Retry to work around intermittent DNS issues in Kubernetes.
    // See issue #123.
    for i := 0; i < 3; i++ {
        if err := tryProcess(data); err == nil {
            return nil
        }
        time.Sleep(time.Second)
    }
    return ErrProcessFailed
}

// ❌ Comment repeats code
// Loop through users
for _, user := range users {
    // Print user name
    fmt.Println(user.Name)
}
```

### 6. TODO and FIXME

```go
// ✅ Format: TODO(username): description
// TODO(alice): Add connection pooling to improve performance

// ✅ With issue number
// TODO(bob): Implement retry logic. See issue #456

// ✅ FIXME marks known issues
// FIXME(charlie): This panics when input is empty, needs validation

// ❌ No owner
// TODO: fix this

// ❌ Unclear description
// FIXME: bad code
```

**When to use**:
- **TODO**: Planned features, optimizations
- **FIXME**: Known bugs, technical debt
- **HACK**: Temporary solution
- **NOTE**: Important notes

## When Not to Comment

### Don't Comment Obvious Content

```go
// ❌ Repeats variable name
// user ID
userID := "123"

// ❌ Repeats function signature
// GetUser gets a user by ID and returns user and error
func GetUser(id string) (*User, error)

// ❌ Repeats code logic
// Loop through users
for _, user := range users {
    fmt.Println(user.Name)
}
```

### Use Good Naming Instead

```go
// ❌ Comment explains bad naming
// d is the duration in seconds
d := 300

// ✅ Use clear naming
timeout := 5 * time.Minute

// ❌ Comment explains complex expression
// Check if user is active and has permission
if u.s == 1 && u.p == true {
    // ...
}

// ✅ Extract into well-named function
if user.IsActive() && user.HasPermission() {
    // ...
}
```

### Don't Comment Out Code

```go
// ❌ Don't commit commented code
func Process(data *Data) error {
    // Old implementation
    // if data.Type == "old" {
    //     return processOld(data)
    // }
    
    return processNew(data)
}

// ✅ Delete old code (Git preserves history)
func Process(data *Data) error {
    return processNew(data)
}
```

## Common Mistakes

### Mistake 1: Comments Out of Sync

```go
// ❌ Outdated comment
// GetUser retrieves a user by email
func GetUser(id string) (*User, error)  // Now queries by ID

// ✅ Keep in sync
// GetUser retrieves a user by ID
func GetUser(id string) (*User, error)
```

### Mistake 2: Only Say "What"

```go
// ❌ Only says what
// Set timeout to 30 seconds
timeout := 30 * time.Second

// ✅ Explain why
// Use longer timeout for large file uploads
timeout := 30 * time.Second
```

### Mistake 3: Missing Comment on Exported Symbol

```go
// ❌ Exported function without comment
func SaveConnections(ctx context.Context, connections []*Connection) error

// ✅ Exported function with comment
// SaveConnections saves a batch of connections to the database.
func SaveConnections(ctx context.Context, connections []*Connection) error
```

### Mistake 4: Wrong Format

```go
// ❌ Doesn't start with function name
// Saves connections to the database
func SaveConnections(...)

// ✅ Starts with function name
// SaveConnections saves connections to the database.
func SaveConnections(...)
```

---

**Reference**: [Effective Go - Commentary](https://go.dev/doc/effective_go#commentary)
