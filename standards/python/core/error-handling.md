---
language: python
category: core
applies_to: []
---

# Error Handling

Based on [PEP 8](https://peps.python.org/pep-0008/) and [Google Python Style Guide](https://google.github.io/styleguide/pyguide.html).

## Quick Reference

- **Specific exceptions** - catch specific exceptions, not bare `except:`
- **Exception chaining** - use `raise X from Y` to preserve original error
- **Minimal try blocks** - only include code that can fail
- **Custom exceptions** - inherit from `Exception`, name ends with `Error`
- **with statement** - use `with` for resource management

## Core Rules

### 1. Catch Specific Exceptions

```python
# ✅ Catch specific exception
try:
    value = data["key"]
except KeyError:
    return default_value

# ❌ Bare except (catches all, including KeyboardInterrupt)
try:
    value = data["key"]
except:  # Wrong: catches SystemExit, KeyboardInterrupt
    return default_value

# ❌ Too broad
try:
    value = data["key"]
except Exception:  # Too broad
    return default_value
```

### 2. Minimal Try Blocks

```python
# ✅ Only include code that can fail
try:
    value = collection[key]
except KeyError:
    return key_not_found(key)
else:
    return handle_value(value)

# ❌ Try block too large
try:
    return handle_value(collection[key])  # May catch KeyError from handle_value
except KeyError:
    return key_not_found(key)
```

### 3. Exception Chaining

```python
# ✅ Use raise from to preserve original exception
try:
    data = json.loads(response.text)
except json.JSONDecodeError as e:
    raise ValidationError(f"Invalid JSON response") from e

# ✅ Use from None to hide original (use sparingly)
try:
    user_id = int(value)
except ValueError:
    raise ValidationError(f"Invalid user ID: {value}") from None

# ❌ Lose original exception context
try:
    data = json.loads(response.text)
except json.JSONDecodeError:
    raise ValidationError("Invalid JSON")  # Original error lost
```

### 4. Custom Exceptions

```python
# ✅ Inherit from Exception, Error suffix
class ValidationError(Exception):
    """Raised when data validation fails."""
    pass

class DatabaseConnectionError(Exception):
    """Raised when database connection fails."""
    def __init__(self, host: str, port: int):
        self.host = host
        self.port = port
        super().__init__(f"Failed to connect to {host}:{port}")

# ✅ Exception hierarchy
class ServiceError(Exception):
    """Base exception for service errors."""
    pass

class ServiceUnavailableError(ServiceError):
    """Service is temporarily unavailable."""
    pass

class ServiceTimeoutError(ServiceError):
    """Service request timed out."""
    pass

# ❌ Inherit from BaseException (wrong)
class MyError(BaseException):  # Wrong: use Exception
    pass
```

### 5. Use with Statement (Context Manager)

```python
# ✅ with automatically cleans up resources
with open("file.txt") as f:
    data = f.read()
# File automatically closed

# ✅ Multiple resources
with open("input.txt") as f_in, open("output.txt", "w") as f_out:
    f_out.write(f_in.read())

# ❌ Manual try/finally (not recommended)
f = open("file.txt")
try:
    data = f.read()
finally:
    f.close()
```

### 6. Don't Use assert for Validation

```python
# ❌ Don't use assert to validate input
def process_user(user_id: int):
    assert user_id > 0  # assert can be disabled (python -O)
    # ...

# ✅ Use raise to validate
def process_user(user_id: int):
    if user_id <= 0:
        raise ValueError(f"Invalid user_id: {user_id}")
    # ...

# ✅ assert only for tests
def test_user_creation():
    user = User.create(name="Alice")
    assert user.name == "Alice"  # OK in tests
```

### 7. else and finally Clauses

```python
# ✅ else: executed when no exception
try:
    data = load_data()
except FileNotFoundError:
    data = create_default_data()
else:
    # Only executed if load_data() succeeds
    validate_data(data)
finally:
    # Always executed
    cleanup()

# ✅ Clear success path
try:
    file = open("config.json")
except FileNotFoundError:
    use_defaults()
else:
    config = json.load(file)
    file.close()
```

## FastAPI Error Handling

### HTTP Exceptions

```python
from fastapi import HTTPException, status

# ✅ Use HTTPException
@app.get("/users/{user_id}")
async def get_user(user_id: int):
    user = await db.get_user(user_id)
    if not user:
        raise HTTPException(
            status_code=status.HTTP_404_NOT_FOUND,
            detail=f"User {user_id} not found"
        )
    return user

# ✅ Custom exception + exception handler
class UserNotFoundError(Exception):
    def __init__(self, user_id: int):
        self.user_id = user_id

@app.exception_handler(UserNotFoundError)
async def user_not_found_handler(request: Request, exc: UserNotFoundError):
    return JSONResponse(
        status_code=404,
        content={"detail": f"User {exc.user_id} not found"}
    )
```

### Dependency Errors

```python
# ✅ Raise HTTPException in dependencies
async def get_current_user(token: str = Depends(oauth2_scheme)) -> User:
    try:
        payload = jwt.decode(token, SECRET_KEY)
        user_id = payload.get("sub")
        if user_id is None:
            raise HTTPException(status_code=401, detail="Invalid token")
        return await get_user(user_id)
    except JWTError:
        raise HTTPException(status_code=401, detail="Could not validate credentials")
```

## Pydantic Validation Errors

```python
from pydantic import BaseModel, field_validator, ValidationError

# ✅ Use field_validator
class UserCreate(BaseModel):
    email: str
    age: int

    @field_validator("email")
    @classmethod
    def validate_email(cls, v: str) -> str:
        if "@" not in v:
            raise ValueError("Invalid email format")
        return v

    @field_validator("age")
    @classmethod
    def validate_age(cls, v: int) -> int:
        if v < 0 or v > 150:
            raise ValueError("Age must be between 0 and 150")
        return v

# ✅ Catch ValidationError
try:
    user = UserCreate(email="invalid", age=200)
except ValidationError as e:
    print(e.errors())
```

## Common Mistakes

### Mistake 1: Swallowing Exceptions

```python
# ❌ Silently ignore errors
try:
    risky_operation()
except Exception:
    pass  # Error swallowed

# ✅ At least log
try:
    risky_operation()
except Exception as e:
    logger.error(f"Operation failed: {e}")
    raise
```

### Mistake 2: Over-catching

```python
# ❌ Catch too much
def process_data(data: dict):
    try:
        validate(data)
        transform(data)
        save(data)
        send_notification()
    except Exception:  # Catches all errors
        return None

# ✅ Handle separately
def process_data(data: dict):
    validate(data)  # Let it raise
    transform(data)
    
    try:
        save(data)
    except DatabaseError as e:
        raise ProcessingError("Failed to save") from e
    
    send_notification()  # OK if it fails
```

### Mistake 3: Return in finally

```python
# ❌ Return in finally (hides exception)
def bad_example():
    try:
        raise ValueError("error")
    finally:
        return "success"  # Exception hidden

# ✅ Don't return in finally
def good_example():
    try:
        raise ValueError("error")
    finally:
        cleanup()  # Only cleanup
```

---

**Full details**:
- [PEP 8 - Programming Recommendations](https://peps.python.org/pep-0008/#programming-recommendations)
- [Exception Chaining Guide](https://krython.com/tutorial/python/exception-chaining-from-and-cause/)
