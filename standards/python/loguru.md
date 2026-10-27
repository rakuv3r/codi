# Logging

Python-specific logging patterns using loguru.

Based on [loguru](https://loguru.readthedocs.io/).

## Quick Reference

- **loguru** - modern, simple logging (recommended)
- **Auto JSON** - serialize=True for production
- **Context** - bind() for request ID, user ID
- **No setup** - works out of the box
- **Colored output** - automatic in development

## loguru Setup

### Basic Configuration

```python
from loguru import logger
import sys

# ✅ Remove default handler
logger.remove()

# ✅ Production (JSON)
logger.add(
    sys.stdout,
    format="{time:YYYY-MM-DD HH:mm:ss} | {level} | {message}",
    level="INFO",
    serialize=True  # JSON output
)

# ✅ Development (colored console)
logger.add(
    sys.stdout,
    format="<green>{time:YYYY-MM-DD HH:mm:ss}</green> | <level>{level: <8}</level> | <cyan>{name}</cyan>:<cyan>{function}</cyan> | {message}",
    level="DEBUG",
    colorize=True
)

# ✅ File output with rotation
logger.add(
    "logs/app.log",
    rotation="500 MB",
    retention="10 days",
    compression="zip",
    level="INFO",
    serialize=True
)
```

### Environment-Based Config

```python
import os

logger.remove()

if os.getenv("ENVIRONMENT") == "production":
    # JSON to stdout
    logger.add(sys.stdout, level="INFO", serialize=True)
else:
    # Colored console
    logger.add(
        sys.stdout,
        level="DEBUG",
        colorize=True,
        format="<green>{time:HH:mm:ss}</green> | <level>{level: <8}</level> | {message}"
    )
```

## Basic Usage

```python
from loguru import logger

# ✅ Simple logging
logger.debug("Debug info")
logger.info("User logged in")
logger.warning("Deprecated API used")
logger.error("Failed to process")
logger.critical("Service down")

# ✅ Structured data
logger.info("User login", user_id=123, ip="192.168.1.1", action="login")
# JSON: {"text": "User login", "record": {"extra": {"user_id": 123, "ip": "192.168.1.1", "action": "login"}}}

# ✅ With context binding
logger = logger.bind(request_id="abc123")
logger.info("Processing request")  # Includes request_id
logger.info("Request completed")   # Includes request_id
```

## FastAPI Integration

```python
from loguru import logger
import uuid
import time

# ✅ Logging middleware
@app.middleware("http")
async def logging_middleware(request: Request, call_next):
    request_id = str(uuid.uuid4())
    
    # Bind request context
    with logger.contextualize(
        request_id=request_id,
        method=request.method,
        path=request.url.path,
    ):
        logger.info("Request started")
        
        start_time = time.time()
        response = await call_next(request)
        duration = time.time() - start_time
        
        logger.info(
            "Request completed",
            status_code=response.status_code,
            duration_ms=int(duration * 1000)
        )
    
    return response

# ✅ Use in handlers
@app.get("/users/{user_id}")
async def get_user(user_id: int):
    logger.info("Fetching user", user_id=user_id)
    return {"user_id": user_id}
```

## Exception Logging

```python
# ✅ Auto exception logging
@logger.catch
async def risky_operation():
    result = 1 / 0  # Exception automatically logged with traceback
    return result

# ✅ Manual exception logging
try:
    result = await process_order(order_id)
except Exception as e:
    logger.exception("Failed to process order", order_id=order_id)
    raise

# ✅ Catch and re-raise
try:
    risky_operation()
except ValueError:
    logger.error("Invalid value")
    raise
```

## Sensitive Data Protection

```python
# ❌ Never log sensitive data
logger.info("User login", password=password)  # WRONG!
logger.info("Payment", credit_card=card)      # WRONG!

# ✅ Log safe identifiers only
logger.info("User login", user_id=user_id)
logger.info("Payment processed", amount=amount, currency="USD")

# ✅ Mask sensitive data
def mask_email(email: str) -> str:
    local, domain = email.split("@")
    return f"{local[0]}***@{domain}"

logger.info("User registered", email=mask_email(email))
```

## Performance

```python
# ✅ Lazy evaluation (automatic with f-strings)
logger.debug(f"Expensive: {expensive_operation()}")
# Only evaluated if debug enabled

# ✅ Conditional logging
if logger.level("DEBUG").no >= logger._core.min_level:
    logger.debug("Debug info", data=compute_expensive_data())

# ✅ Use opt() for performance-critical paths
logger.opt(lazy=True).debug("Expensive: {data}", data=lambda: expensive_operation())
```

## Custom Formatting

```python
# ✅ Custom format per handler
logger.add(
    sys.stdout,
    format="{time:YYYY-MM-DD HH:mm:ss} | {level: <8} | {name}:{function}:{line} | {message}",
    level="INFO"
)

# ✅ Custom filter
def production_filter(record):
    return record["level"].no >= logger.level("INFO").no

logger.add(sys.stdout, filter=production_filter, serialize=True)

# ✅ Custom serializer
def custom_serializer(record):
    subset = {
        "time": record["time"].isoformat(),
        "level": record["level"].name,
        "message": record["message"],
        **record["extra"]
    }
    return json.dumps(subset)

logger.add(sys.stdout, serialize=custom_serializer)
```

## Intercept Standard Library

```python
import logging
from loguru import logger

# ✅ Intercept stdlib logging
class InterceptHandler(logging.Handler):
    def emit(self, record):
        # Get corresponding loguru level
        try:
            level = logger.level(record.levelname).name
        except ValueError:
            level = record.levelno

        # Find caller
        frame, depth = logging.currentframe(), 2
        while frame.f_code.co_filename == logging.__file__:
            frame = frame.f_back
            depth += 1

        logger.opt(depth=depth, exception=record.exc_info).log(level, record.getMessage())

# Configure
logging.basicConfig(handlers=[InterceptHandler()], level=0)

# Now stdlib logging goes through loguru
logging.getLogger("uvicorn").handlers = [InterceptHandler()]
logging.getLogger("fastapi").handlers = [InterceptHandler()]
```

## Uvicorn Integration

```python
# ✅ Configure uvicorn to use loguru
import uvicorn

if __name__ == "__main__":
    # Intercept uvicorn logging
    logging.getLogger("uvicorn").handlers = [InterceptHandler()]
    logging.getLogger("uvicorn.access").handlers = [InterceptHandler()]
    
    uvicorn.run(
        "app.main:app",
        host="0.0.0.0",
        port=8000,
        log_config=None  # Disable uvicorn default logging
    )
```

## Testing

```python
from loguru import logger
import pytest

# ✅ Capture logs in tests
def test_logging(caplog):
    # Add handler to capture
    logger.add(caplog.handler, format="{message}")
    
    logger.info("test message", user_id=123)
    
    assert "test message" in caplog.text

# ✅ Test with context
def test_with_context():
    logs = []
    
    def sink(message):
        logs.append(message)
    
    logger.add(sink, format="{message}")
    
    with logger.contextualize(request_id="test123"):
        logger.info("test")
    
    assert any("test123" in str(log) for log in logs)
```

## Context Management

```python
# ✅ Add context for duration
with logger.contextualize(user_id=123, action="update"):
    logger.info("Starting operation")
    do_work()
    logger.info("Completed")
    # Both logs include user_id and action

# ✅ Bind creates new logger
request_logger = logger.bind(request_id="abc123")
request_logger.info("Processing")  # Includes request_id

# ✅ Patch context
logger.patch(lambda record: record["extra"].update(service="api"))
logger.info("Service log")  # Includes service=api
```

## Multiple Sinks

```python
# ✅ Different levels per sink
logger.add(sys.stdout, level="INFO", colorize=True)
logger.add("error.log", level="ERROR", rotation="1 day")
logger.add("debug.log", level="DEBUG", rotation="500 MB", retention="3 days")

# ✅ Filter by module
logger.add(
    "api.log",
    filter=lambda record: "api" in record["name"],
    level="INFO"
)
```

## Best Practices

```python
# ✅ Import logger directly
from loguru import logger

# ✅ Use contextualize for request scope
with logger.contextualize(request_id=request_id):
    logger.info("Processing request")

# ✅ Use @logger.catch for async functions
@logger.catch
async def async_operation():
    await something()

# ✅ Include relevant context
logger.info("Operation", user_id=user_id, action="update", resource_id=id)

# ❌ Don't log in hot loops
for item in items:  # 10,000 items
    logger.debug("Processing", item=item)  # WRONG: too verbose

# ✅ Log summary
logger.info("Processing items", count=len(items))
# Process
logger.info("Completed", success=success_count, failed=fail_count)
```

## Migration from stdlib

```python
# Before (stdlib)
import logging
logger = logging.getLogger(__name__)
logger.info("Message", extra={"user_id": 123})

# After (loguru)
from loguru import logger
logger.info("Message", user_id=123)

# Benefits:
# - No setup needed
# - Automatic JSON serialization
# - Better exception handling
# - Colored output
# - Simpler API
```

---

**References**:
- [loguru Documentation](https://loguru.readthedocs.io/)
- [loguru GitHub](https://github.com/Delgan/loguru)
