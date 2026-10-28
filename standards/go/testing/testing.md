---
language: go
category: dependency
applies_to: []
---

# Testing

Based on [Google Go Best Practices - Testing](https://google.github.io/styleguide/go/best-practices#tests).

## Quick Reference

- **Table-driven tests** - use struct slice to organize test cases
- **Subtests** - use `t.Run()` to isolate each case
- **Test helpers** - mark with `t.Helper()`
- **Failure handling** - `t.Fatal` for setup failures, `t.Error` for test failures
- **Mock naming** - describe behavior (`AlwaysSucceeds`) or type (`StubClient`)

## Core Rules

### 1. Table-Driven Tests

```go
// ✅ Standard structure
func TestParseQuery(t *testing.T) {
    tests := []struct {
        name    string
        input   string
        want    Query
        wantErr bool
    }{
        {
            name:  "valid query",
            input: "SELECT * FROM users",
            want:  Query{Type: "SELECT", Table: "users"},
        },
        {
            name:    "invalid syntax",
            input:   "SELCT",
            wantErr: true,
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got, err := ParseQuery(tt.input)
            if (err != nil) != tt.wantErr {
                t.Errorf("ParseQuery() error = %v, wantErr %v", err, tt.wantErr)
                return
            }
            if !reflect.DeepEqual(got, tt.want) {
                t.Errorf("ParseQuery() = %v, want %v", got, tt.want)
            }
        })
    }
}

// ❌ Bad: duplicate code
func TestParseQueryValid(t *testing.T) { ... }
func TestParseQueryInvalid(t *testing.T) { ... }
```

### 2. Subtests

```go
// ✅ Use t.Run for isolation
func TestDatabase(t *testing.T) {
    db := setupTestDB(t)
    
    t.Run("Insert", func(t *testing.T) {
        err := db.Insert(...)
        if err != nil {
            t.Fatal(err)  // Subtest failure doesn't affect others
        }
    })
    
    t.Run("Query", func(t *testing.T) {
        rows, err := db.Query(...)
        if err != nil {
            t.Fatal(err)
        }
    })
}

// ✅ Benefits:
// - Run individually: go test -run TestDatabase/Insert
// - Failure isolation
// - Shared setup
```

### 3. Test Helpers

```go
// ✅ Mark with t.Helper()
func setupTestDB(t *testing.T) *DB {
    t.Helper()  // Report caller's line on failure
    
    db, err := Connect("test-dsn")
    if err != nil {
        t.Fatalf("setup failed: %v", err)  // Use Fatal for setup failures
    }
    
    t.Cleanup(func() {
        db.Close()  // Auto cleanup
    })
    
    return db
}

// ✅ Usage
func TestQuery(t *testing.T) {
    db := setupTestDB(t)  // Failure reported at this line
    // ...
}

// ❌ Bad: helper accepts *testing.T but doesn't mark Helper
func setupDB(t *testing.T) *DB {
    // Missing t.Helper()
    db, err := Connect("test-dsn")
    if err != nil {
        t.Fatalf("setup failed: %v", err)  // Failure reported here, not at caller
    }
    return db
}
```

### 4. Error Handling

```go
// ✅ Use t.Fatal for setup failures (abort test)
func TestWithDB(t *testing.T) {
    db, err := setupDB()
    if err != nil {
        t.Fatalf("setup failed: %v", err)  // Can't continue, abort
    }
    // ...
}

// ✅ Use t.Error for test failures (continue other checks)
func TestValidation(t *testing.T) {
    result := Validate(input)
    
    if result.ErrorCount != 2 {
        t.Errorf("ErrorCount = %d, want 2", result.ErrorCount)
        // Continue checking other fields
    }
    
    if len(result.Errors) == 0 {
        t.Error("Errors is empty, want non-empty")
    }
}

// ❌ Bad: call t.Fatal in goroutine
func TestConcurrent(t *testing.T) {
    go func() {
        if err := doWork(); err != nil {
            t.Fatal(err)  // Wrong: not in main test goroutine
        }
    }()
}

// ✅ Correct: use t.Error in goroutine
func TestConcurrent(t *testing.T) {
    var wg sync.WaitGroup
    wg.Add(1)
    
    go func() {
        defer wg.Done()
        if err := doWork(); err != nil {
            t.Error(err)  // Correct
        }
    }()
    
    wg.Wait()
}
```

### 5. Mock Naming

```go
// ✅ Single mock: use Stub
type Stub struct{}

func (s *Stub) Call() error { return nil }

// ✅ Multiple behaviors: describe behavior
type AlwaysSucceeds struct{}
type AlwaysFailsWith5XX struct{}
type ReturnsPartialResults struct{}

// ✅ Multiple types: include type name
type StubUserStore struct{}
type StubOrderStore struct{}

// ❌ Bad: Mock prefix (doesn't describe behavior)
type MockClient struct{}  // Better: StubClient or behavior description
```

## Production Patterns

### HTTP Handler Testing

```go
func TestSaveConnectionsHandler(t *testing.T) {
    tests := []struct {
        name       string
        input      []Connection
        storeErr   error
        wantCode   int
        wantRetcode int
    }{
        {
            name:       "success",
            input:      []Connection{{Src: "a", Dst: "b"}},
            wantCode:   200,
            wantRetcode: 0,
        },
        {
            name:       "storage error",
            storeErr:   errors.New("db error"),
            wantCode:   500,
            wantRetcode: 1,
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            store := &mockStore{err: tt.storeErr}
            logger := slog.New(slog.NewTextHandler(io.Discard, nil))
            handler := NewSaveConnectionsHandler(store, logger)
            
            w := httptest.NewRecorder()
            c, _ := gin.CreateTestContext(w)
            
            body, _ := json.Marshal(tt.input)
            c.Request = httptest.NewRequest("POST", "/", bytes.NewBuffer(body))
            
            handler(c)
            
            if w.Code != tt.wantCode {
                t.Errorf("code = %d, want %d", w.Code, tt.wantCode)
            }
        })
    }
}
```

### Concurrency Testing

```go
func TestConcurrentAccess(t *testing.T) {
    cache := NewCache()
    
    var wg sync.WaitGroup
    for i := 0; i < 100; i++ {
        wg.Add(1)
        go func(id int) {
            defer wg.Done()
            cache.Set(id, id*2)
            got := cache.Get(id)
            if got != id*2 {
                t.Errorf("id %d: got %d, want %d", id, got, id*2)
            }
        }(i)
    }
    
    wg.Wait()
}
```

### Integration Tests with Real Dependencies

```go
// ✅ Integration tests use real database (testcontainers)
func TestDatabaseIntegration(t *testing.T) {
    if testing.Short() {
        t.Skip("skipping integration test")
    }
    
    ctx := context.Background()
    container, err := testcontainers.GenericContainer(ctx, ...)
    if err != nil {
        t.Fatal(err)
    }
    defer container.Terminate(ctx)
    
    client := NewClient(container.URI())
    // Real operations test
}

// Run: go test -short  # Skip integration tests
// Run: go test         # Include integration tests
```

---

**Full details**: [Google Go Best Practices - Testing](https://google.github.io/styleguide/go/best-practices#tests)
