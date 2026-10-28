---
language: go
category: core
applies_to: []
---

# Context

Context management principles and best practices (library-agnostic).

## Why Context Matters

- **Cancellation** - stop operations when no longer needed
- **Timeouts** - prevent operations from running too long
- **Request-scoped data** - pass metadata without globals
- **Resource cleanup** - ensure proper cleanup on cancellation

## Context Lifecycle

**Principle**: Context flows from caller to callee.

```
HTTP Request
  → Handler receives context
    → Service receives context
      → Repository receives context
        → Database query uses context
```

## Cancellation

**Principle**: Respect context cancellation immediately.

```
# ✅ Check cancellation in loops
for item in items:
    if context.is_cancelled():
        return CancelledError
    process(item)

# ❌ Ignore cancellation
for item in items:
    process(item)  # Keeps running even if cancelled
```

## Timeouts

**Principle**: Set timeouts at appropriate levels.

```
# ✅ Per-operation timeout
context_with_timeout = context.with_timeout(5 seconds)
result = database.query(context_with_timeout, sql)

# ❌ No timeout
result = database.query(context, sql)  # Can hang forever
```

**Where to set timeouts**:
- HTTP requests: per-request
- Database queries: per-query
- External API calls: per-call
- Long operations: based on SLA

## Request-Scoped Data

**Principle**: Use context for request-scoped metadata only.

**Good uses**:
- ✅ Request ID
- ✅ User ID
- ✅ Trace ID
- ✅ Authentication token

**Bad uses**:
- ❌ Business logic data
- ❌ Configuration
- ❌ Optional parameters

```
# ✅ Request metadata
context.set("request_id", "abc123")
context.set("user_id", 456)

# ❌ Business logic
context.set("discount_rate", 0.1)  # Should be function parameter
```

## Context Propagation

**Principle**: Always pass context as first parameter.

```
# ✅ Context as first parameter
function save_user(context, user):
    db.insert(context, user)

# ❌ Context buried or missing
function save_user(user, context):  # Wrong position
function save_user(user):           # Missing
```

## Context Creation

**Principle**: Don't create root context in libraries.

```
# ❌ Library creates root context
function fetch_data():
    context = new_context()  # Wrong
    return api.call(context)

# ✅ Library receives context
function fetch_data(context):
    return api.call(context)
```

## Immutability

**Principle**: Context is immutable - modifications create new context.

```
# ✅ Create child context
parent_context = request.context
child_context = parent_context.with_value("key", "value")

# Original parent_context unchanged
```

## Context vs Globals

**Why context is better than globals**:
- ✅ Explicit dependency
- ✅ Cancellation support
- ✅ Request isolation
- ✅ Testable

```
# ❌ Global state
global current_user_id
function save_order():
    user_id = current_user_id  # Fragile

# ✅ Context
function save_order(context):
    user_id = context.get("user_id")  # Explicit
```

## Testing with Context

**Principle**: Create test contexts with necessary values.

```
# ✅ Test context
function test_save_user():
    context = new_test_context()
    context.set("user_id", 123)
    
    result = service.save_user(context, user_data)
    assert result.success
```

## Common Patterns

### 1. HTTP Request Context
```
Request arrives
  → Create root context
  → Add request ID
  → Add user ID (from auth)
  → Pass to handlers
```

### 2. Background Job Context
```
Job starts
  → Create root context with timeout
  → Add job ID
  → Pass to workers
```

### 3. Graceful Shutdown
```
Shutdown signal received
  → Cancel root context
  → All operations receive cancellation
  → Wait for cleanup
```

## Anti-Patterns

### 1. Storing Business Data
```
# ❌ Don't store business logic in context
context.set("product", product_obj)

# ✅ Pass as parameter
function process_order(context, product):
    ...
```

### 2. Ignoring Cancellation
```
# ❌ Never check cancellation
while True:
    process_data()

# ✅ Check regularly
while not context.is_cancelled():
    process_data()
```

### 3. Context in Struct
```
# ❌ Don't store context in struct
class Service:
    context: Context  # Wrong

# ✅ Pass context to methods
class Service:
    function do_work(context):
        ...
```

---

**For specific library usage, see**: [context](../context.md)
