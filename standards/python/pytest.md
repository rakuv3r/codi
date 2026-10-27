# Testing

Based on pytest best practices.

## Quick Reference

- **pytest** - Python standard testing framework
- **Fixtures** - test resource management
- **Parametrize** - parameterized tests
- **Mock** - mock external dependencies
- **Coverage** - pytest-cov for test coverage

## Core Rules

### 1. Test Function Naming

```python
# ✅ test_ prefix + descriptive name
def test_user_creation():
    user = User(name="Alice")
    assert user.name == "Alice"

def test_get_user_returns_none_when_not_found():
    result = get_user(999)
    assert result is None

def test_email_validation_raises_error_on_invalid_format():
    with pytest.raises(ValidationError):
        validate_email("invalid")

# ❌ Vague names
def test_1():  # Wrong
    pass

def test_user():  # Too vague
    pass
```

### 2. Fixtures

```python
import pytest

# ✅ Use fixtures to share test resources
@pytest.fixture
def db():
    """Create test database."""
    db = create_test_db()
    yield db
    db.close()

@pytest.fixture
def user(db):
    """Create test user."""
    return db.create_user(name="Alice", email="alice@example.com")

def test_user_creation(user):
    assert user.name == "Alice"
    assert user.email == "alice@example.com"

# ✅ Fixture scopes
@pytest.fixture(scope="session")  # Once per test session
def app():
    return create_app()

@pytest.fixture(scope="module")  # Once per module
def db_connection():
    return connect_db()

@pytest.fixture(scope="function")  # Once per function (default)
def temp_file():
    f = create_temp_file()
    yield f
    f.cleanup()
```

### 3. Parametrize

```python
# ✅ Use parametrize for multiple inputs
@pytest.mark.parametrize("email,expected", [
    ("alice@example.com", True),
    ("invalid", False),
    ("@example.com", False),
    ("alice@", False),
])
def test_email_validation(email, expected):
    result = is_valid_email(email)
    assert result == expected

# ✅ Multiple parameters
@pytest.mark.parametrize("age,is_adult", [
    (17, False),
    (18, True),
    (65, True),
])
def test_age_validation(age, is_adult):
    assert is_adult_age(age) == is_adult

# ✅ Parameterized fixtures
@pytest.fixture(params=["sqlite", "postgres", "mysql"])
def db(request):
    return create_db(request.param)

def test_query(db):
    # Test runs for each database
    result = db.query("SELECT 1")
    assert result
```

### 4. Mock and Patch

```python
from unittest.mock import Mock, patch, AsyncMock

# ✅ Mock external services
@patch("app.services.external_api.call")
def test_external_service(mock_call):
    mock_call.return_value = {"status": "success"}
    
    result = process_with_external_service()
    assert result["status"] == "success"
    mock_call.assert_called_once()

# ✅ Async mock
@pytest.mark.asyncio
@patch("app.services.async_client.fetch", new_callable=AsyncMock)
async def test_async_service(mock_fetch):
    mock_fetch.return_value = {"data": "test"}
    
    result = await fetch_data()
    assert result["data"] == "test"

# ✅ Mock fixture
@pytest.fixture
def mock_db():
    db = Mock()
    db.get_user.return_value = User(id=1, name="Alice")
    return db

def test_with_mock_db(mock_db):
    user = mock_db.get_user(1)
    assert user.name == "Alice"
```

### 5. Exception Testing

```python
# ✅ Test exception raised
def test_division_by_zero():
    with pytest.raises(ZeroDivisionError):
        divide(1, 0)

# ✅ Check exception message
def test_validation_error_message():
    with pytest.raises(ValidationError, match="Invalid email"):
        validate_email("invalid")

# ✅ Capture exception object
def test_custom_exception():
    with pytest.raises(UserNotFoundError) as exc_info:
        get_user(999)
    
    assert exc_info.value.user_id == 999
```

### 6. Async Tests

```python
import pytest

# ✅ Async test function
@pytest.mark.asyncio
async def test_async_function():
    result = await fetch_data()
    assert result is not None

# ✅ Async fixture
@pytest.fixture
async def async_db():
    db = await create_async_db()
    yield db
    await db.close()

@pytest.mark.asyncio
async def test_with_async_fixture(async_db):
    user = await async_db.get_user(1)
    assert user is not None
```

### 7. Database Testing

```python
# ✅ Use transaction rollback
@pytest.fixture
async def db_session():
    async with get_async_session() as session:
        async with session.begin():
            yield session
            await session.rollback()

@pytest.mark.asyncio
async def test_create_user(db_session):
    user = User(name="Alice")
    db_session.add(user)
    await db_session.flush()
    
    assert user.id is not None
    # Transaction rolls back, no impact on real database

# ✅ Use test database
@pytest.fixture(scope="session")
def test_db():
    # Create test database
    create_test_database()
    yield
    # Drop test database
    drop_test_database()
```

## Test Organization

```python
# ✅ Use classes to organize related tests
class TestUserService:
    def test_get_user(self, db):
        user = UserService(db).get_user(1)
        assert user is not None
    
    def test_create_user(self, db):
        user = UserService(db).create_user(name="Alice")
        assert user.id is not None
    
    def test_delete_user(self, db):
        result = UserService(db).delete_user(1)
        assert result is True

# ✅ Use markers to categorize
@pytest.mark.slow
def test_slow_operation():
    pass

@pytest.mark.integration
def test_database_integration():
    pass

# Run specific markers:
# pytest -m "not slow"
# pytest -m integration
```

## Test Coverage

```bash
# Run tests with coverage
pytest --cov=app --cov-report=html

# Fail if coverage below 80%
pytest --cov=app --cov-fail-under=80

# Show missing lines
pytest --cov=app --cov-report=term-missing
```

## Common Mistakes

### Mistake 1: Incorrect Fixture Scope

```python
# ❌ Create database every time (slow)
@pytest.fixture
def db():
    return create_database()

# ✅ Session-level sharing
@pytest.fixture(scope="session")
def db():
    db = create_database()
    yield db
    db.cleanup()
```

### Mistake 2: Not Cleaning Up

```python
# ❌ No cleanup
@pytest.fixture
def temp_file():
    return open("temp.txt", "w")

# ✅ Use yield for cleanup
@pytest.fixture
def temp_file():
    f = open("temp.txt", "w")
    yield f
    f.close()
    os.remove("temp.txt")
```

### Mistake 3: Test Dependencies

```python
# ❌ Tests depend on each other
def test_create_user():
    global created_user_id
    response = create_user(name="Alice")
    created_user_id = response["id"]

def test_get_user():
    user = get_user(created_user_id)  # Depends on previous test

# ✅ Independent tests
@pytest.fixture
def created_user():
    return create_user(name="Alice")

def test_get_user(created_user):
    user = get_user(created_user["id"])
    assert user is not None
```

---

**Full documentation**:
- [pytest Documentation](https://docs.pytest.org/)
- [pytest-asyncio](https://pytest-asyncio.readthedocs.io/)
- [pytest-cov](https://pytest-cov.readthedocs.io/)
