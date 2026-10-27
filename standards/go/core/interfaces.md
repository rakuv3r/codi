# Interfaces

Based on Go interface design best practices.

## Quick Reference

- **Accept interfaces, return concrete types** - function parameters use interfaces, return values use concrete types
- **Small interfaces over large** - single-method interfaces are best
- **Consumer defines interfaces** - define in the using package, not the implementing package
- **Avoid premature abstraction** - define interfaces when you have a second implementation
- **Implicit implementation** - no need to explicitly declare implementation

## Core Rules

### 1. Small Interface Principle

```go
// ✅ Single-method interface (best)
type Reader interface {
    Read(p []byte) (n int, err error)
}

type Writer interface {
    Write(p []byte) (n int, err error)
}

type Closer interface {
    Close() error
}

// ✅ Compose to build larger interfaces
type ReadWriteCloser interface {
    Reader
    Writer
    Closer
}

// ❌ Large interfaces are hard to implement
type FileOperations interface {
    Read(p []byte) (n int, err error)
    Write(p []byte) (n int, err error)
    Close() error
    Seek(offset int64, whence int) (int64, error)
    Stat() (FileInfo, error)
}
```

### 2. Consumer Defines Interface

```go
// ❌ Define in implementation package
// package storage
type Storage interface {
    Save(data *Data) error
    Get(id string) (*Data, error)
}

type PostgresStorage struct{}
func (p *PostgresStorage) Save(data *Data) error { ... }

// ✅ Define in using package
// package service
type dataStore interface {  // Private, declares only what's needed
    Save(data *Data) error
}

type Service struct {
    store dataStore
}

// package storage
type PostgresStorage struct{}
func (p *PostgresStorage) Save(data *Data) error { ... }
// Implicitly implements dataStore
```

**Why**:
- Implementation package doesn't depend on using package (decoupled)
- Minimal interface (only declares what's needed)
- Flexible (different users define different interfaces)

### 3. Accept Interfaces, Return Concrete Types

```go
// ✅ Parameters use interfaces, return concrete types
func NewService(store Storage, logger Logger) *Service {
    return &Service{
        store:  store,
        logger: logger,
    }
}

// ❌ Return interface (limits flexibility)
func NewService(store Storage, logger Logger) ServiceInterface {
    return &Service{...}
}
```

**Why return concrete types**:
- Callers can use all methods
- Callers can define their own interfaces
- Maintains flexibility

### 4. Interface Naming

```go
// ✅ Single-method: -er suffix
type Reader interface { ... }
type Writer interface { ... }
type Formatter interface { ... }
type Validator interface { ... }

// ✅ Multi-method: descriptive noun
type FileSystem interface { ... }
type ConnectionPool interface { ... }

// ❌ I prefix (C#/Java style)
type IReader interface { ... }

// ❌ Generic names
type Manager interface { ... }
type Handler interface { ... }
type Service interface { ... }  // Too vague

// ✅ Specific names
type ConnectionManager interface { ... }
type RequestHandler interface { ... }
type UserService interface { ... }
```

### 5. Compile-Time Verification

```go
// ✅ Verify implementation at compile time
type Storage interface {
    Save(data *Data) error
    Get(id string) (*Data, error)
}

type PostgresStorage struct{}

var _ Storage = (*PostgresStorage)(nil)  // Compile-time check

func (p *PostgresStorage) Save(data *Data) error { ... }
func (p *PostgresStorage) Get(id string) (*Data, error) { ... }

// Verify mocks in tests
type MockStorage struct { ... }

var _ Storage = (*MockStorage)(nil)
```

### 6. Avoid Premature Abstraction

```go
// ❌ Only one implementation, no need for interface
type UserRepository interface {
    GetUser(id string) (*User, error)
}

type PostgresUserRepository struct{}

// ✅ Use concrete type directly
type UserRepository struct {
    db *sql.DB
}

func (r *UserRepository) GetUser(id string) (*User, error) { ... }

// For testing, define interface in test package
type userGetter interface {
    GetUser(id string) (*User, error)
}
```

**When to define interfaces**:
1. Have a second implementation
2. For testing (define in test package)
3. Standard library requires (e.g., io.Reader)

### 7. Empty Interface (any)

```go
// ✅ Use any (Go 1.18+)
func Print(value any) {
    fmt.Println(value)
}

// ❌ Use interface{} (old style)
func Print(value interface{}) { ... }

// ⚠️ any loses type safety, use sparingly
func Process(data any) error {
    str, ok := data.(string)  // Needs type assertion
    if !ok {
        return errors.New("expected string")
    }
    // ...
}

// ✅ Prefer concrete types
func Process(data string) error {
    // Type safe
}
```

## Production Patterns

### Service Layer Interface

```go
// ✅ Define interface in service package
package service

type userStore interface {
    GetUser(ctx context.Context, id string) (*User, error)
    SaveUser(ctx context.Context, user *User) error
}

type UserService struct {
    store  userStore
    logger *slog.Logger
}

func NewUserService(store userStore, logger *slog.Logger) *UserService {
    return &UserService{
        store:  store,
        logger: logger,
    }
}

func (s *UserService) GetUser(ctx context.Context, id string) (*User, error) {
    return s.store.GetUser(ctx, id)
}

// Use mock in tests
type mockStore struct {
    GetUserFunc func(ctx context.Context, id string) (*User, error)
}

var _ userStore = (*mockStore)(nil)

func (m *mockStore) GetUser(ctx context.Context, id string) (*User, error) {
    if m.GetUserFunc != nil {
        return m.GetUserFunc(ctx, id)
    }
    return nil, nil
}
```

## Common Mistakes

### Mistake 1: Returning Interface

```go
// ❌ Limits flexibility
func NewService() ServiceInterface {
    return &Service{}
}

// ✅ Return concrete type
func NewService() *Service {
    return &Service{}
}
```

### Mistake 2: Interface Too Large

```go
// ❌ Too many methods
type Database interface {
    Connect() error
    Disconnect() error
    Query(sql string) (*Rows, error)
    Execute(sql string) error
    BeginTransaction() (*Transaction, error)
    Commit() error
    Rollback() error
}

// ✅ Split into small interfaces
type Connector interface {
    Connect() error
    Disconnect() error
}

type Querier interface {
    Query(sql string) (*Rows, error)
}

type Executor interface {
    Execute(sql string) error
}
```

### Mistake 3: Define in Implementation Package

```go
// ❌ Interface and implementation together
package storage

type Storage interface { ... }
type PostgresStorage struct { ... }

// ✅ Interface where it's used
package service

type storage interface { ... }  // Only declares what's needed
```

---

**Reference**: [Effective Go - Interfaces](https://go.dev/doc/effective_go#interfaces)
