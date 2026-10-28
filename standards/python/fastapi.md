---
language: python
category: dependency
applies_to: [api]
---

# FastAPI Framework Patterns

FastAPI-specific patterns and best practices for async HTTP services.

Based on [FastAPI Best Practices](https://github.com/zhanymkanov/fastapi-best-practices) and [official documentation](https://fastapi.tiangolo.com/).

## Path Operation Patterns

### Function Naming

```python
# ✅ Verb + noun
@app.get("/users/{user_id}")
async def get_user(user_id: int) -> User:
    pass

@app.post("/users")
async def create_user(user: UserCreate) -> User:
    pass

@app.put("/users/{user_id}")
async def update_user(user_id: int, user: UserUpdate) -> User:
    pass

@app.delete("/users/{user_id}")
async def delete_user(user_id: int) -> None:
    pass

# ❌ Generic names
@app.get("/users/{user_id}")
async def handler(user_id: int):  # Wrong
    pass
```

### Dependency Injection

```python
from fastapi import Depends

# ✅ Database dependency
async def get_db() -> AsyncGenerator[AsyncSession, None]:
    async with get_async_session() as session:
        yield session

@app.get("/users/{user_id}")
async def get_user(
    user_id: int,
    db: AsyncSession = Depends(get_db)
) -> User:
    return await db.get(User, user_id)

# ✅ Authentication dependency
async def get_current_user(
    token: str = Depends(oauth2_scheme),
    db: AsyncSession = Depends(get_db)
) -> User:
    user = await verify_token(token, db)
    if not user:
        raise HTTPException(status_code=401, detail="Invalid token")
    return user

@app.get("/me")
async def get_current_user_info(
    current_user: User = Depends(get_current_user)
) -> User:
    return current_user
```

### Dependency Overrides (Testing)

```python
# ✅ Override dependencies in tests
def override_get_db():
    return MockDB()

def override_get_current_user():
    return User(id=1, name="TestUser")

app.dependency_overrides[get_db] = override_get_db
app.dependency_overrides[get_current_user] = override_get_current_user

def test_protected_endpoint():
    response = client.get("/me")
    assert response.status_code == 200
```

## Error Handling

### HTTPException

```python
from fastapi import HTTPException, status

# ✅ Use HTTPException with status codes
@app.get("/users/{user_id}")
async def get_user(user_id: int, db: AsyncSession = Depends(get_db)):
    user = await db.get(User, user_id)
    if not user:
        raise HTTPException(
            status_code=status.HTTP_404_NOT_FOUND,
            detail=f"User {user_id} not found"
        )
    return user

# ✅ Custom detail
@app.post("/users")
async def create_user(user: UserCreate, db: AsyncSession = Depends(get_db)):
    existing = await db.get_user_by_email(user.email)
    if existing:
        raise HTTPException(
            status_code=status.HTTP_409_CONFLICT,
            detail="Email already registered"
        )
    return await db.create_user(user)
```

### Custom Exception Handlers

```python
from fastapi import Request
from fastapi.responses import JSONResponse

# ✅ Custom exception
class UserNotFoundError(Exception):
    def __init__(self, user_id: int):
        self.user_id = user_id

# ✅ Exception handler
@app.exception_handler(UserNotFoundError)
async def user_not_found_handler(request: Request, exc: UserNotFoundError):
    return JSONResponse(
        status_code=404,
        content={"detail": f"User {exc.user_id} not found"}
    )

# ✅ Generic error handler
@app.exception_handler(Exception)
async def generic_exception_handler(request: Request, exc: Exception):
    logger.error(f"Unhandled exception: {exc}", exc_info=True)
    return JSONResponse(
        status_code=500,
        content={"detail": "Internal server error"}
    )
```

## Testing Patterns

### TestClient (Sync)

```python
from fastapi.testclient import TestClient

@pytest.fixture
def client():
    return TestClient(app)

def test_get_user(client):
    response = client.get("/users/1")
    assert response.status_code == 200
    assert response.json()["name"] == "Alice"

def test_create_user(client):
    response = client.post(
        "/users",
        json={"name": "Bob", "email": "bob@example.com"}
    )
    assert response.status_code == 201
```

### AsyncClient (Async)

```python
from httpx import AsyncClient, ASGITransport
import pytest

@pytest.fixture
async def async_client():
    async with AsyncClient(
        transport=ASGITransport(app=app),
        base_url="http://test"
    ) as client:
        yield client

@pytest.mark.asyncio
async def test_get_user(async_client):
    response = await async_client.get("/users/1")
    assert response.status_code == 200

@pytest.mark.asyncio
async def test_create_user(async_client):
    response = await async_client.post(
        "/users",
        json={"name": "Alice", "email": "alice@example.com"}
    )
    assert response.status_code == 201
```

### Testing with Database

```python
@pytest.fixture
async def db_session():
    async with get_async_session() as session:
        async with session.begin():
            yield session
            await session.rollback()

@pytest.fixture
async def async_client(db_session):
    async def override_get_db():
        yield db_session
    
    app.dependency_overrides[get_db] = override_get_db
    
    async with AsyncClient(
        transport=ASGITransport(app=app),
        base_url="http://test"
    ) as client:
        yield client
    
    app.dependency_overrides.clear()

@pytest.mark.asyncio
async def test_create_and_get_user(async_client):
    # Create
    response = await async_client.post(
        "/users",
        json={"name": "Alice", "email": "alice@example.com"}
    )
    assert response.status_code == 201
    user_id = response.json()["id"]
    
    # Get
    response = await async_client.get(f"/users/{user_id}")
    assert response.status_code == 200
    assert response.json()["name"] == "Alice"
```

## Request/Response Models

### Input Validation

```python
from pydantic import BaseModel, field_validator

# ✅ Use Pydantic models for validation
class UserCreate(BaseModel):
    name: str
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

@app.post("/users")
async def create_user(user: UserCreate) -> User:
    # Pydantic automatically validates
    return await db.create_user(user)
```

### Response Models

```python
# ✅ Separate response models
class UserResponse(BaseModel):
    id: int
    name: str
    email: str
    created_at: datetime

    class Config:
        from_attributes = True  # Pydantic V2

@app.get("/users/{user_id}", response_model=UserResponse)
async def get_user(user_id: int, db: AsyncSession = Depends(get_db)):
    user = await db.get(User, user_id)
    if not user:
        raise HTTPException(status_code=404)
    return user  # Automatically serialized to UserResponse
```

## Middleware Patterns

### Request ID

```python
from starlette.middleware.base import BaseHTTPMiddleware
import uuid

class RequestIDMiddleware(BaseHTTPMiddleware):
    async def dispatch(self, request: Request, call_next):
        request_id = request.headers.get("X-Request-ID", str(uuid.uuid4()))
        request.state.request_id = request_id
        
        response = await call_next(request)
        response.headers["X-Request-ID"] = request_id
        return response

app.add_middleware(RequestIDMiddleware)
```

### Logging

```python
import time

class LoggingMiddleware(BaseHTTPMiddleware):
    async def dispatch(self, request: Request, call_next):
        start_time = time.time()
        
        response = await call_next(request)
        
        duration = time.time() - start_time
        logger.info(
            "request",
            method=request.method,
            path=request.url.path,
            status=response.status_code,
            duration=duration,
            request_id=getattr(request.state, "request_id", None)
        )
        
        return response

app.add_middleware(LoggingMiddleware)
```

## Background Tasks

```python
from fastapi import BackgroundTasks

# ✅ Simple background task
def send_email(email: str, message: str):
    # Send email
    pass

@app.post("/users")
async def create_user(
    user: UserCreate,
    background_tasks: BackgroundTasks
):
    new_user = await db.create_user(user)
    background_tasks.add_task(send_email, new_user.email, "Welcome!")
    return new_user

# ✅ Async background task
async def process_data(data: dict):
    # Async processing
    pass

@app.post("/process")
async def trigger_processing(
    data: dict,
    background_tasks: BackgroundTasks
):
    background_tasks.add_task(process_data, data)
    return {"status": "processing"}
```

## Common Mistakes

### Mistake 1: Not Using async/await

```python
# ❌ Blocking sync call in async function
@app.get("/users/{user_id}")
async def get_user(user_id: int):
    user = db.query(User).filter_by(id=user_id).first()  # Blocks!
    return user

# ✅ Use async database operations
@app.get("/users/{user_id}")
async def get_user(user_id: int, db: AsyncSession = Depends(get_db)):
    user = await db.get(User, user_id)
    return user
```

### Mistake 2: Mutable Default Arguments

```python
# ❌ Mutable default
@app.get("/search")
async def search(tags: list[str] = []):  # Wrong!
    pass

# ✅ Use None with default_factory
from typing import Annotated
from fastapi import Query

@app.get("/search")
async def search(tags: Annotated[list[str], Query()] = []):  # Correct
    pass
```

### Mistake 3: Not Handling Database Sessions

```python
# ❌ Session not closed
@app.get("/users")
async def get_users():
    session = get_session()
    users = await session.execute(select(User))
    return users  # Session never closed!

# ✅ Use dependency with yield
async def get_db():
    async with get_async_session() as session:
        yield session  # Auto-closed

@app.get("/users")
async def get_users(db: AsyncSession = Depends(get_db)):
    result = await db.execute(select(User))
    return result.scalars().all()
```

---

**Reference**:
- [FastAPI Documentation](https://fastapi.tiangolo.com/)
- [FastAPI Best Practices](https://github.com/zhanymkanov/fastapi-best-practices)
