# Testify Testing Framework Standards

## Overview

Testify is the most popular testing toolkit for Go, providing rich assertions, mocking capabilities, and test suites. It reduces boilerplate and provides clearer test failures than standard library alone.

## Core Packages

- **assert** - Assertions that mark test as failed but continue execution
- **require** - Assertions that stop test execution immediately on failure
- **mock** - Mocking framework for interfaces
- **suite** - Test suite with lifecycle hooks

## Assert vs Require (Critical Choice)

### assert - Continue on Failure

```go
func TestUser(t *testing.T) {
    user := getUser(1)
    
    assert.NotNil(t, user)           // Fails but continues
    assert.Equal(t, "Alice", user.Name)  // Still executes (may panic if user is nil)
    assert.Equal(t, 30, user.Age)        // Still executes
}
```

**Use when**: Multiple independent checks provide useful diagnostic information.

### require - Stop on Failure

```go
func TestUser(t *testing.T) {
    user := getUser(1)
    
    require.NotNil(t, user)          // Fails and stops test
    assert.Equal(t, "Alice", user.Name)  // Won't execute if user is nil
    assert.Equal(t, 30, user.Age)
}
```

**Use when**: Failed condition makes subsequent checks meaningless or unsafe.

### Rule of Thumb

```go
// ✅ Correct pattern
require.NoError(t, err)           // Guard: stop if error
require.NotNil(t, result)         // Guard: stop if nil
assert.Equal(t, expected, result.Field)  // Safe check

// ❌ Wrong pattern
assert.NoError(t, err)            // Continues even with error
result.DoSomething()              // May panic if result is nil
```

## Common Assertions

### Equality

```go
assert.Equal(t, expected, actual)
assert.NotEqual(t, notExpected, actual)
assert.EqualValues(t, 42, int64(42))     // Different types, same value
assert.Exactly(t, expected, actual)       // Strict type + value equality
```

### Comparison

```go
assert.Greater(t, 10, 5)
assert.GreaterOrEqual(t, 10, 10)
assert.Less(t, 5, 10)
assert.LessOrEqual(t, 5, 5)
```

### Nil and Zero

```go
assert.Nil(t, obj)
assert.NotNil(t, obj)
assert.Zero(t, value)              // Zero value of type
assert.NotZero(t, value)
assert.Empty(t, collection)        // len() == 0
assert.NotEmpty(t, collection)
```

### Booleans

```go
assert.True(t, condition)
assert.False(t, condition)
```

### Errors

```go
assert.NoError(t, err)
assert.Error(t, err)
assert.ErrorIs(t, err, ErrNotFound)
assert.ErrorAs(t, err, &target)
assert.ErrorContains(t, err, "not found")
assert.EqualError(t, err, "exact message")
```

### Collections

```go
assert.Len(t, collection, 3)
assert.Contains(t, slice, element)
assert.NotContains(t, slice, element)
assert.ElementsMatch(t, expected, actual)  // Same elements, any order
assert.Subset(t, list, subset)
```

### Types

```go
assert.IsType(t, &MyStruct{}, obj)
assert.Implements(t, (*MyInterface)(nil), obj)
assert.IsTypeof(t, "string", actual)
```

### Panics

```go
assert.Panics(t, func() {
    divideByZero()
})

assert.NotPanics(t, func() {
    safeOperation()
})
```

### JSON

```go
expected := `{"name":"Alice","age":30}`
assert.JSONEq(t, expected, actual)  // Ignores whitespace, order
```

### Eventually (Async)

```go
assert.Eventually(t, func() bool {
    return service.IsReady()
}, 5*time.Second, 100*time.Millisecond, "service should become ready")
```

### HTTP

```go
assert.HTTPSuccess(t, handler, "GET", "/users", nil)
assert.HTTPStatusCode(t, handler, "POST", "/users", body, 201)
assert.HTTPBodyContains(t, handler, "GET", "/users", nil, "Alice")
```

### Custom Messages

```go
assert.Equal(t, 42, result, "user ID should match")
assert.Equalf(t, 42, result, "user %s should have ID %d", name, 42)
```

## Mock Framework

### Defining Mocks

```go
type MockUserRepo struct {
    mock.Mock
}

func (m *MockUserRepo) GetUser(id int) (*User, error) {
    args := m.Called(id)
    
    // Handle nil return
    if args.Get(0) == nil {
        return nil, args.Error(1)
    }
    
    return args.Get(0).(*User), args.Error(1)
}

func (m *MockUserRepo) SaveUser(user *User) error {
    args := m.Called(user)
    return args.Error(0)
}
```

### Setting Expectations

```go
func TestService(t *testing.T) {
    mockRepo := new(MockUserRepo)
    service := NewService(mockRepo)
    
    // Expect specific call with specific args
    mockRepo.On("GetUser", 1).Return(&User{ID: 1, Name: "Alice"}, nil)
    mockRepo.On("GetUser", 999).Return(nil, ErrNotFound)
    
    // Expect with any arguments
    mockRepo.On("GetUser", mock.Anything).Return(&User{}, nil)
    
    // Expect with argument matcher
    mockRepo.On("GetUser", mock.MatchedBy(func(id int) bool {
        return id > 0
    })).Return(&User{}, nil)
    
    // Run test
    user, err := service.GetUser(1)
    
    // Verify
    require.NoError(t, err)
    assert.Equal(t, "Alice", user.Name)
    
    // Verify all expectations were met
    mockRepo.AssertExpectations(t)
}
```

### Return Value Extraction

```go
args := m.Called(id)

// Type-specific extractors
value := args.Int(0)
str := args.String(0)
flag := args.Bool(0)
err := args.Error(1)

// Generic extractor (needs type assertion)
obj := args.Get(0).(*MyStruct)
```

### Verifying Calls

```go
// Verify specific call was made
mockRepo.AssertCalled(t, "GetUser", 1)

// Verify call was NOT made
mockRepo.AssertNotCalled(t, "DeleteUser", mock.Anything)

// Verify call count
mockRepo.AssertNumberOfCalls(t, "GetUser", 3)

// Verify all expectations
mockRepo.AssertExpectations(t)
```

### Multiple Return Values

```go
mockRepo.On("GetUser", 1).Return(&User{ID: 1}, nil)
mockRepo.On("GetUser", 2).Return(nil, ErrNotFound)

// Different behavior on successive calls
mockRepo.On("GetUser", 1).Return(&User{ID: 1}, nil).Once()
mockRepo.On("GetUser", 1).Return(&User{ID: 1, Name: "Updated"}, nil)
```

## Test Suites

### Basic Suite

```go
type UserTestSuite struct {
    suite.Suite
    db     *sql.DB
    repo   *UserRepo
}

// Run suite
func TestUserTestSuite(t *testing.T) {
    suite.Run(t, new(UserTestSuite))
}
```

### Lifecycle Hooks

```go
// Suite-level (once per suite)
func (s *UserTestSuite) SetupSuite() {
    s.db = connectTestDB()
}

func (s *UserTestSuite) TearDownSuite() {
    s.db.Close()
}

// Test-level (before/after each test)
func (s *UserTestSuite) SetupTest() {
    s.db.Exec("TRUNCATE TABLE users")
    s.repo = NewUserRepo(s.db)
}

func (s *UserTestSuite) TearDownTest() {
    // Cleanup after each test
}

// SubTest-level (before/after each subtest)
func (s *UserTestSuite) SetupSubTest() {
    // Setup for subtests
}

func (s *UserTestSuite) TearDownSubTest() {
    // Cleanup after subtests
}

// Named hooks (more control)
func (s *UserTestSuite) BeforeTest(suiteName, testName string) {
    log.Printf("Starting %s.%s", suiteName, testName)
}

func (s *UserTestSuite) AfterTest(suiteName, testName string) {
    log.Printf("Finished %s.%s", suiteName, testName)
}
```

### Suite Methods

```go
func (s *UserTestSuite) TestGetUser() {
    user, err := s.repo.GetUser(1)
    
    // Use s.Assert() or s.Require()
    s.NoError(err)
    s.Equal("Alice", user.Name)
}

func (s *UserTestSuite) TestCreateUser() {
    err := s.repo.CreateUser(&User{Name: "Bob"})
    s.NoError(err)
}
```

### Table-Driven Tests in Suites

```go
func (s *UserTestSuite) TestValidation() {
    tests := []struct{
        name    string
        user    *User
        wantErr bool
    }{
        {"valid", &User{Name: "Alice", Email: "alice@example.com"}, false},
        {"no name", &User{Email: "alice@example.com"}, true},
        {"no email", &User{Name: "Alice"}, true},
    }
    
    for _, tt := range tests {
        s.Run(tt.name, func() {
            // SetupSubTest runs here if defined
            err := s.repo.Validate(tt.user)
            if tt.wantErr {
                s.Error(err)
            } else {
                s.NoError(err)
            }
            // TearDownSubTest runs here if defined
        })
    }
}
```

## Assertion Instances (Reduce Boilerplate)

```go
func TestSomething(t *testing.T) {
    assert := assert.New(t)
    require := require.New(t)
    
    // No need to pass t to each assertion
    require.NoError(err)
    assert.Equal(42, result)
    assert.True(condition)
}
```

## Testing Patterns

### Testing with Mocks

```go
func TestUserService_GetUser(t *testing.T) {
    mockRepo := new(MockUserRepo)
    mockCache := new(MockCache)
    
    service := NewUserService(mockRepo, mockCache)
    
    // Setup expectations
    mockCache.On("Get", "user:1").Return(nil, ErrCacheMiss)
    mockRepo.On("GetUser", 1).Return(&User{ID: 1, Name: "Alice"}, nil)
    mockCache.On("Set", "user:1", mock.Anything).Return(nil)
    
    // Execute
    user, err := service.GetUser(1)
    
    // Verify
    require.NoError(t, err)
    assert.Equal(t, "Alice", user.Name)
    
    // Verify all mocks
    mockRepo.AssertExpectations(t)
    mockCache.AssertExpectations(t)
}
```

### Testing Errors

```go
func TestGetUser_NotFound(t *testing.T) {
    mockRepo := new(MockUserRepo)
    service := NewUserService(mockRepo)
    
    mockRepo.On("GetUser", 999).Return(nil, ErrNotFound)
    
    user, err := service.GetUser(999)
    
    assert.Nil(t, user)
    assert.ErrorIs(t, err, ErrNotFound)
}
```

### Testing with Real Dependencies

```go
func TestIntegration(t *testing.T) {
    if testing.Short() {
        t.Skip("skipping integration test")
    }
    
    db := setupTestDB(t)
    defer db.Close()
    
    repo := NewUserRepo(db)
    
    // Test with real database
    err := repo.CreateUser(&User{Name: "Alice"})
    require.NoError(t, err)
    
    user, err := repo.GetUser(1)
    require.NoError(t, err)
    assert.Equal(t, "Alice", user.Name)
}
```

## Anti-Patterns

### ❌ Don't use assert for guards

```go
// Wrong: continues and panics
assert.NotNil(t, user)
assert.Equal(t, "Alice", user.Name)  // Panics if user is nil

// Right: stops on failure
require.NotNil(t, user)
assert.Equal(t, "Alice", user.Name)
```

### ❌ Don't forget AssertExpectations

```go
// Wrong: mock expectations never verified
mockRepo.On("GetUser", 1).Return(user, nil)
service.GetUser(1)
// Missing: mockRepo.AssertExpectations(t)

// Right
mockRepo.AssertExpectations(t)
```

### ❌ Don't ignore mock.Anything overuse

```go
// Wrong: too permissive
mockRepo.On("GetUser", mock.Anything).Return(user, nil)

// Right: specific expectations
mockRepo.On("GetUser", 1).Return(user1, nil)
mockRepo.On("GetUser", 2).Return(user2, nil)
```

### ❌ Don't mix assertion styles

```go
// Wrong: inconsistent
if user == nil {
    t.Fatal("user is nil")
}
assert.Equal(t, "Alice", user.Name)

// Right: consistent
require.NotNil(t, user)
assert.Equal(t, "Alice", user.Name)
```

## References

- GitHub: https://github.com/stretchr/testify
- Package docs: https://pkg.go.dev/github.com/stretchr/testify
- Most popular Go testing toolkit (40k+ stars)
