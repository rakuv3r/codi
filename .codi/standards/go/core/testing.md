---
language: go
category: core
applies_to: []
---

# Testing

Testing principles and best practices (library-agnostic).

## Why Testing Matters

- **Confidence** - verify code works as expected
- **Regression prevention** - catch bugs early
- **Documentation** - tests show how code should be used
- **Refactoring safety** - change code without fear

## Types of Tests

### Unit Tests
- Test single function/method in isolation
- Fast (milliseconds)
- No external dependencies
- High coverage

### Integration Tests
- Test components working together
- Slower (seconds)
- May use real database/services
- Focus on critical paths

### End-to-End Tests
- Test entire system from user perspective
- Slowest (minutes)
- Use real environment
- Focus on key user flows

**Principle**: Test pyramid - many unit tests, fewer integration tests, minimal E2E tests.

## Test Structure

Use AAA pattern:

```
# Arrange - setup test data
# Act - execute the function
# Assert - verify the result
```

**Example**:
```
test "calculate total":
    # Arrange
    cart = Cart([Item(price=10), Item(price=20)])
    
    # Act
    total = cart.calculate_total()
    
    # Assert
    assert total == 30
```

## Test Naming

**Principle**: Test names should describe what is being tested.

```
# ❌ Vague
test_user()
test_case1()

# ✅ Descriptive
test_create_user_with_valid_email()
test_create_user_returns_error_for_duplicate_email()
test_calculate_discount_for_premium_users()
```

**Format**: `test_<function>_<scenario>_<expected_result>`

## What to Test

**Do test**:
- ✅ Business logic
- ✅ Edge cases (null, empty, max values)
- ✅ Error handling
- ✅ Public APIs

**Don't test**:
- ❌ Third-party libraries
- ❌ Language built-ins
- ❌ Trivial getters/setters
- ❌ Private implementation details

## Test Independence

**Principle**: Tests should not depend on each other.

```
# ❌ Dependent tests
test1: create_user()
test2: login_user()  # Depends on test1

# ✅ Independent tests
test1: 
    user = create_user()
    assert user.id

test2:
    user = create_user()  # Create own data
    login(user)
```

## Mocking

**When to mock**:
- External services (APIs, databases)
- Slow operations
- Non-deterministic behavior (time, random)

**When NOT to mock**:
- Simple functions
- Your own code (test real implementation)
- Value objects

## Test Data

**Principles**:
- Use realistic data
- Make data obvious (avoid magic numbers)
- Factory pattern for complex objects

```
# ❌ Magic numbers
user = User("test", "test@test.com", 1, 2, 3)

# ✅ Clear intent
user = User(
    name="John Doe",
    email="john@example.com",
    age=30,
    role="admin"
)
```

## Assertions

**Principles**:
- One logical assertion per test
- Use specific assertions
- Include failure messages

```
# ❌ Generic
assert result == True

# ✅ Specific
assert user.is_active == True, f"Expected user {user.id} to be active"
```

## Performance

**Principles**:
- Tests should be fast (< 100ms per unit test)
- Parallelize when possible
- Use in-memory databases for integration tests

## Coverage

**Guideline**: Aim for 80%+ coverage, but don't chase 100%.

**Focus on**:
- Critical business logic
- Complex algorithms
- Error paths

**Don't obsess over**:
- Trivial code
- Framework code
- Configuration files

## Test Organization

**Principle**: Mirror your source code structure.

```
project/
├── src/
│   ├── user/
│   │   └── service.go
│   └── payment/
│       └── service.go
└── test/
    ├── user/
    │   └── service_test.go
    └── payment/
        └── service_test.go
```

## Common Anti-Patterns

### 1. Testing Implementation Details
```
# ❌ Tests internal state
assert user._cached_permissions == []

# ✅ Tests behavior
assert user.can_access("admin") == False
```

### 2. Flaky Tests
- Tests that randomly fail
- Usually due to: timing, shared state, external dependencies
- Fix immediately - don't ignore

### 3. Slow Tests
- Unit tests should run in seconds
- If slow, probably not a unit test

---

**For specific library usage, see**: [testing/testing](../testing/testing.md)
