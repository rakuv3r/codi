---
language: go
category: core
applies_to: []
---

# Logging

Logging principles and best practices (library-agnostic).

## Why Logging Matters

- **Debugging** - trace execution flow and identify issues
- **Monitoring** - track application health and performance
- **Auditing** - record user actions and system events
- **Alerting** - detect and respond to problems

## Log Levels

Use appropriate levels to control verbosity:

```
DEBUG   - detailed information for diagnosing problems
INFO    - general informational messages
WARN    - warning messages for potentially harmful situations
ERROR   - error events that might still allow the app to continue
FATAL   - severe errors that cause the application to abort
```

**Guidelines**:
- Production: INFO and above
- Development: DEBUG and above
- Never use DEBUG in production (performance impact)

## Structured Logging

**Principle**: Log data as key-value pairs, not strings.

```
# ❌ Unstructured
log("User 123 logged in from 192.168.1.1")

# ✅ Structured
log.Info("user login", 
    "user_id", 123,
    "ip", "192.168.1.1")
```

**Benefits**:
- Searchable
- Parseable
- Aggregatable
- Machine-readable

## Context Propagation

**Principle**: Include request context in all logs.

```
# ✅ Add request ID to all logs in a request
log.Info("processing started", "request_id", "abc123")
log.Info("fetching data", "request_id", "abc123")
log.Info("processing completed", "request_id", "abc123")
```

**Benefits**:
- Trace requests across services
- Debug specific requests
- Measure request duration

## Sensitive Data

**Never log**:
- ❌ Passwords
- ❌ API keys / tokens
- ❌ Credit card numbers
- ❌ Personal information (unless required)

```
# ❌ Wrong
log.Info("user login", "password", password)

# ✅ Right
log.Info("user login", "user_id", userID)
```

## Performance

**Principles**:
- Use lazy evaluation for expensive operations
- Don't log in hot loops (aggregate instead)
- Buffer logs in high-throughput scenarios

```
# ❌ Log every iteration
for item in items:  # 10,000 items
    log.Debug("processing item", item)

# ✅ Log summary
log.Info("processing items", "count", len(items))
# ... process ...
log.Info("completed", "success", successCount, "failed", failCount)
```

## Error Logging

**Principles**:
- Always log errors with context
- Include stack traces for unexpected errors
- Don't log and return (choose one)

```
# ❌ Log and return (duplicate)
if err != nil:
    log.Error("failed to save", err)
    return err

# ✅ Return error (let caller decide whether to log)
if err != nil:
    return fmt.Errorf("save user: %w", err)
```

## Output Format

**Development**: Human-readable (text)
```
2026-10-15 10:30:00 INFO user login user_id=123 ip=192.168.1.1
```

**Production**: Machine-readable (JSON)
```json
{"time":"2026-10-15T10:30:00Z","level":"INFO","msg":"user login","user_id":123,"ip":"192.168.1.1"}
```

## Choosing a Logging Library

**Criteria**:
- Structured logging support
- Performance (minimal overhead)
- Context propagation
- Multiple output formats
- Active maintenance

**For specific library usage, see**: [log/slog](../log/slog.md)
