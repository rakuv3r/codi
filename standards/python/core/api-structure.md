---
language: python
category: core
applies_to: [api]
---

# Python API Project Structure

## Overview

This guide covers structure and organization patterns for Python HTTP API projects using FastAPI.

---

## Directory Layout

### `/app` - Application Directory

Use `app/` for application code. The `src/` layout is for libraries, not applications.

**Organization principle**: Organize by feature/domain (users/, orders/), NOT by technical layer (routes/, services/, repositories/).

**Why feature-based?**
- High cohesion: all code for one feature together
- Easy maintenance: change one feature = modify one directory
- Clear boundaries: module dependencies are explicit
- Consistent with Go template philosophy

---

## Recommended Structure

```
myapi/
├── app/                        # Application directory
│   ├── users/                  # User feature module
│   │   ├── __init__.py
│   │   ├── routes.py           # FastAPI routes
│   │   ├── schemas.py          # Pydantic models
│   │   ├── service.py          # Business logic
│   │   ├── repository.py       # Data access
│   │   └── models.py           # SQLAlchemy models (optional)
│   ├── orders/                 # Order feature module
│   │   ├── routes.py
│   │   ├── schemas.py
│   │   └── service.py
│   ├── core/                   # Core configuration
│   │   ├── __init__.py
│   │   ├── config.py           # Settings
│   │   └── database.py         # Database connection (optional)
│   └── main.py                 # FastAPI application entry
├── tests/
│   ├── test_users.py
│   └── conftest.py
├── pyproject.toml              # Project configuration
├── .python-version             # Python version (3.12)
├── Makefile
└── README.md
```

---

## Feature Module Organization

### Three-Layer Pattern

**routes.py** → **service.py** → **repository.py**

```
app/users/
  routes.py        # HTTP handling
  schemas.py       # Request/response models
  service.py       # Business logic
  repository.py    # Database operations
```

**When to skip service layer**: Simple CRUD operations can go directly from routes to repository.

---

## File Responsibilities

### routes.py - HTTP Routes

```python
from fastapi import APIRouter, Depends, HTTPException
from .schemas import UserCreate, UserResponse
from .service import UserService

router = APIRouter(prefix="/users", tags=["users"])

@router.get("/{user_id}", response_model=UserResponse)
async def get_user(
    user_id: int,
    service: UserService = Depends(),
) -> UserResponse:
    """Get user by ID"""
    user = await service.get_user(user_id)
    if not user:
        raise HTTPException(status_code=404, detail="User not found")
    return user

@router.post("/", response_model=UserResponse, status_code=201)
async def create_user(
    user_data: UserCreate,
    service: UserService = Depends(),
) -> UserResponse:
    """Create new user"""
    return await service.create_user(user_data)
```

**Responsibilities**:
- Parse HTTP request
- Validate input (via Pydantic)
- Call service layer
- Return HTTP response
- Handle HTTP-specific errors

---

### schemas.py - Pydantic Models

```python
from pydantic import BaseModel, EmailStr, ConfigDict

class UserBase(BaseModel):
    """Base user fields"""
    email: EmailStr
    name: str

class UserCreate(UserBase):
    """Create user request"""
    password: str

class UserUpdate(BaseModel):
    """Update user request (all fields optional)"""
    email: EmailStr | None = None
    name: str | None = None

class UserResponse(UserBase):
    """User response"""
    id: int
    is_active: bool = True
    
    model_config = ConfigDict(from_attributes=True)
```

**Responsibilities**:
- Define request/response structure
- Validation rules
- Serialization/deserialization

---

### service.py - Business Logic

```python
from fastapi import Depends, HTTPException
from .repository import UserRepository
from .schemas import UserCreate, UserUpdate, UserResponse

class UserService:
    def __init__(self, repo: UserRepository = Depends()):
        self.repo = repo
    
    async def get_user(self, user_id: int) -> UserResponse | None:
        """Get user by ID"""
        return await self.repo.get(user_id)
    
    async def create_user(self, data: UserCreate) -> UserResponse:
        """Create new user with business logic"""
        # Check if email exists
        existing = await self.repo.get_by_email(data.email)
        if existing:
            raise HTTPException(400, "Email already exists")
        
        # Hash password (business logic)
        hashed_password = hash_password(data.password)
        
        # Create user
        return await self.repo.create(data, hashed_password)
    
    async def update_user(
        self, 
        user_id: int, 
        data: UserUpdate
    ) -> UserResponse:
        """Update user"""
        user = await self.repo.get(user_id)
        if not user:
            raise HTTPException(404, "User not found")
        
        return await self.repo.update(user_id, data)
```

**Responsibilities**:
- Business rules
- Orchestrate multiple repositories
- Domain logic
- Return domain objects (not HTTP responses)

---

### repository.py - Data Access

```python
from sqlalchemy import select
from sqlalchemy.ext.asyncio import AsyncSession
from fastapi import Depends
from .models import User
from .schemas import UserCreate

class UserRepository:
    def __init__(self, db: AsyncSession = Depends(get_db)):
        self.db = db
    
    async def get(self, user_id: int) -> User | None:
        """Get user by ID"""
        result = await self.db.execute(
            select(User).where(User.id == user_id)
        )
        return result.scalar_one_or_none()
    
    async def get_by_email(self, email: str) -> User | None:
        """Get user by email"""
        result = await self.db.execute(
            select(User).where(User.email == email)
        )
        return result.scalar_one_or_none()
    
    async def create(self, data: UserCreate, hashed_password: str) -> User:
        """Create new user"""
        user = User(
            email=data.email,
            name=data.name,
            hashed_password=hashed_password,
        )
        self.db.add(user)
        await self.db.commit()
        await self.db.refresh(user)
        return user
    
    async def update(self, user_id: int, data: dict) -> User:
        """Update user"""
        user = await self.get(user_id)
        for key, value in data.items():
            if value is not None:
                setattr(user, key, value)
        await self.db.commit()
        await self.db.refresh(user)
        return user
```

**Responsibilities**:
- Database queries
- CRUD operations
- Data mapping
- No business logic

---

### models.py - SQLAlchemy Models (Optional)

```python
from sqlalchemy import String, Boolean
from sqlalchemy.orm import Mapped, mapped_column
from app.core.database import Base

class User(Base):
    __tablename__ = "users"
    
    id: Mapped[int] = mapped_column(primary_key=True)
    email: Mapped[str] = mapped_column(String(255), unique=True, index=True)
    name: Mapped[str] = mapped_column(String(100))
    hashed_password: Mapped[str] = mapped_column(String(255))
    is_active: Mapped[bool] = mapped_column(Boolean, default=True)
```

**Only include if using an ORM**. For projects without a database, skip this file.

---

## Application Entry Point

### main.py

```python
from fastapi import FastAPI
from .users.routes import router as users_router
from .orders.routes import router as orders_router
from .core.config import settings

app = FastAPI(
    title=settings.APP_NAME,
    version="1.0.0",
)

# Register routers
app.include_router(users_router, prefix="/api/v1")
app.include_router(orders_router, prefix="/api/v1")

@app.get("/health")
async def health():
    """Health check endpoint"""
    return {"status": "ok"}
```

---

## Configuration Management

### core/config.py

```python
from pydantic_settings import BaseSettings, SettingsConfigDict

class Settings(BaseSettings):
    APP_NAME: str = "My API"
    DATABASE_URL: str
    SECRET_KEY: str
    DEBUG: bool = False
    
    model_config = SettingsConfigDict(
        env_file=".env",
        case_sensitive=True,
    )

settings = Settings()
```

**Use environment variables** for configuration. Never hardcode secrets.

---

## Database Setup (Optional)

### core/database.py

```python
from sqlalchemy.ext.asyncio import create_async_engine, AsyncSession, async_sessionmaker
from sqlalchemy.orm import declarative_base
from .config import settings

engine = create_async_engine(
    settings.DATABASE_URL,
    echo=settings.DEBUG,
)

async_session_maker = async_sessionmaker(
    engine,
    class_=AsyncSession,
    expire_on_commit=False,
)

Base = declarative_base()

async def get_db() -> AsyncSession:
    """Dependency for database session"""
    async with async_session_maker() as session:
        yield session
```

---

## API Versioning

### Simple Approach (Recommended)

Use route prefixes in main.py:

```python
# main.py
app.include_router(users_router, prefix="/api/v1")

# For v2, create new module if significantly different
from .usersv2.routes import router as users_router_v2
app.include_router(users_router_v2, prefix="/api/v2")
```

**Directory structure**:
```
app/
  users/          # V1 implementation
  usersv2/        # V2 (if different)
```

### Complex Approach (Large Projects)

Organize by version directory:

```
app/
  api/
    v1/
      users.py    # V1 routes
    v2/
      users.py    # V2 routes
  users/          # Shared logic
    service.py
    repository.py
```

**Use simple approach until you need the complex one**.

---

## Error Handling

### Standard error responses

```python
from fastapi import HTTPException

# routes.py
@router.get("/{user_id}")
async def get_user(user_id: int, service: UserService = Depends()):
    user = await service.get_user(user_id)
    if not user:
        raise HTTPException(status_code=404, detail="User not found")
    return user
```

### Custom exception handler

```python
# main.py
from fastapi import Request
from fastapi.responses import JSONResponse

@app.exception_handler(ValueError)
async def value_error_handler(request: Request, exc: ValueError):
    return JSONResponse(
        status_code=400,
        content={"detail": str(exc)},
    )
```

---

## Testing

### Unit tests

```python
# tests/test_users.py
import pytest
from app.users.service import UserService
from app.users.schemas import UserCreate

@pytest.mark.asyncio
async def test_create_user(mock_repo):
    service = UserService(repo=mock_repo)
    data = UserCreate(email="test@example.com", name="Test", password="pass")
    
    user = await service.create_user(data)
    
    assert user.email == "test@example.com"
```

### Integration tests

```python
from fastapi.testclient import TestClient
from app.main import app

client = TestClient(app)

def test_get_user():
    response = client.get("/api/v1/users/1")
    assert response.status_code == 200
    assert "email" in response.json()
```

---

## When to Adjust Structure

### Shared Schemas

When multiple modules need the same schemas:

```
app/
  schemas/          # Global schemas
    common.py
  users/
    routes.py
    service.py
```

### Complex Dependencies

For projects like infra-graph with cross-cutting concerns:

```
app/
  api/v1/          # Thin route layer
    audit.py
    sync.py
  services/        # Business modules
    aliyun/
    kubernetes/
  schemas/         # Shared schemas
```

**Start simple, evolve as needed**.

---

## Summary

✅ **DO**:
- Organize by feature modules (users/, orders/)
- Use three-layer pattern (routes → service → repository)
- Use Pydantic for validation
- Use environment variables for config
- Write tests for service and repository layers
- Use dependency injection (Depends())

❌ **DON'T**:
- Organize by technical layers (routes/, services/, repositories/)
- Use src/ layout for applications (only for libraries)
- Skip service layer when you have complex business logic
- Hardcode configuration
- Mix HTTP concerns with business logic

---

**Remember**: Start with this simple structure and evolve based on your project's needs. Complex patterns emerge naturally as requirements grow.
