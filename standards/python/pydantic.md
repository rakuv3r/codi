# Pydantic Models

Pydantic V2 patterns and best practices for data validation.

Based on [Pydantic V2 Documentation](https://docs.pydantic.dev/latest/).

## Quick Reference

- **BaseModel** - base class for models
- **Field** - field configuration and validation
- **field_validator** - custom validation
- **model_validator** - model-level validation
- **ConfigDict** - model configuration
- **from_attributes** - ORM mode (formerly orm_mode)

## Core Rules

### 1. Basic Models

```python
from pydantic import BaseModel, Field
from datetime import datetime

# ✅ Simple model
class User(BaseModel):
    id: int
    name: str
    email: str
    age: int | None = None
    created_at: datetime = Field(default_factory=datetime.now)

# Usage
user = User(id=1, name="Alice", email="alice@example.com")
print(user.model_dump())  # V2: model_dump() instead of dict()
print(user.model_dump_json())  # V2: model_dump_json() instead of json()
```

### 2. Field Configuration

```python
from pydantic import BaseModel, Field

# ✅ Field constraints
class UserCreate(BaseModel):
    name: str = Field(min_length=1, max_length=100)
    email: str = Field(pattern=r"^[\w\.-]+@[\w\.-]+\.\w+$")
    age: int = Field(ge=0, le=150)
    bio: str | None = Field(default=None, max_length=500)
    tags: list[str] = Field(default_factory=list, max_length=10)

# ✅ Field metadata
class Product(BaseModel):
    name: str = Field(description="Product name", examples=["iPhone 15"])
    price: float = Field(gt=0, description="Price in USD")
    stock: int = Field(ge=0, description="Available quantity")
```

### 3. Field Validators

```python
from pydantic import BaseModel, field_validator

# ✅ Single field validator
class User(BaseModel):
    email: str
    age: int

    @field_validator("email")
    @classmethod
    def validate_email(cls, v: str) -> str:
        if "@" not in v:
            raise ValueError("Invalid email format")
        return v.lower()  # Normalize

    @field_validator("age")
    @classmethod
    def validate_age(cls, v: int) -> int:
        if v < 0 or v > 150:
            raise ValueError("Age must be between 0 and 150")
        return v

# ✅ Validator for multiple fields
class User(BaseModel):
    first_name: str
    last_name: str

    @field_validator("first_name", "last_name")
    @classmethod
    def validate_name(cls, v: str) -> str:
        if not v.strip():
            raise ValueError("Name cannot be empty")
        return v.strip().title()
```

### 4. Model Validators

```python
from pydantic import BaseModel, model_validator

# ✅ Cross-field validation
class DateRange(BaseModel):
    start_date: datetime
    end_date: datetime

    @model_validator(mode="after")
    def validate_date_range(self) -> "DateRange":
        if self.end_date < self.start_date:
            raise ValueError("end_date must be after start_date")
        return self

# ✅ Before validation
class User(BaseModel):
    username: str
    email: str

    @model_validator(mode="before")
    @classmethod
    def normalize_data(cls, data: dict) -> dict:
        # Normalize before field validation
        if isinstance(data, dict):
            data["username"] = data.get("username", "").lower()
            data["email"] = data.get("email", "").lower()
        return data
```

### 5. Model Configuration

```python
from pydantic import BaseModel, ConfigDict

# ✅ Model config (V2)
class User(BaseModel):
    model_config = ConfigDict(
        from_attributes=True,  # ORM mode (was orm_mode)
        str_strip_whitespace=True,
        validate_assignment=True,
        populate_by_name=True,  # Allow field aliases
    )
    
    id: int
    name: str
    email: str

# Usage with ORM
from sqlalchemy.orm import Session

db_user = session.query(DBUser).first()
user = User.from_orm(db_user)  # V1
user = User.model_validate(db_user)  # V2
```

### 6. Inheritance

```python
# ✅ Model inheritance
class UserBase(BaseModel):
    name: str
    email: str

class UserCreate(UserBase):
    password: str

class UserUpdate(UserBase):
    name: str | None = None
    email: str | None = None

class UserResponse(UserBase):
    id: int
    created_at: datetime
    
    model_config = ConfigDict(from_attributes=True)
```

### 7. Nested Models

```python
# ✅ Nested models
class Address(BaseModel):
    street: str
    city: str
    country: str
    postal_code: str

class User(BaseModel):
    name: str
    email: str
    address: Address
    tags: list[str] = []
    metadata: dict[str, any] = {}

# Usage
user = User(
    name="Alice",
    email="alice@example.com",
    address={
        "street": "123 Main St",
        "city": "New York",
        "country": "USA",
        "postal_code": "10001"
    }
)
```

## FastAPI Integration

### Request/Response Models

```python
from fastapi import FastAPI, HTTPException
from pydantic import BaseModel, Field

app = FastAPI()

# ✅ Separate models for create/update/response
class UserCreate(BaseModel):
    name: str = Field(min_length=1, max_length=100)
    email: str
    password: str = Field(min_length=8)

class UserUpdate(BaseModel):
    name: str | None = None
    email: str | None = None

class UserResponse(BaseModel):
    id: int
    name: str
    email: str
    created_at: datetime
    
    model_config = ConfigDict(from_attributes=True)

@app.post("/users", response_model=UserResponse)
async def create_user(user: UserCreate) -> UserResponse:
    # Pydantic automatically validates
    db_user = await db.create_user(user)
    return UserResponse.model_validate(db_user)

@app.patch("/users/{user_id}", response_model=UserResponse)
async def update_user(user_id: int, user: UserUpdate) -> UserResponse:
    db_user = await db.update_user(user_id, user.model_dump(exclude_unset=True))
    return UserResponse.model_validate(db_user)
```

### Validation Errors

```python
from pydantic import ValidationError

# ✅ Handle validation errors
try:
    user = UserCreate(name="", email="invalid", password="short")
except ValidationError as e:
    print(e.errors())
    # [
    #   {'loc': ('name',), 'msg': 'String should have at least 1 character', ...},
    #   {'loc': ('email',), 'msg': 'value is not a valid email address', ...},
    #   {'loc': ('password',), 'msg': 'String should have at least 8 characters', ...}
    # ]

# FastAPI automatically returns 422 with error details
@app.post("/users")
async def create_user(user: UserCreate):
    # If validation fails, FastAPI returns:
    # {
    #   "detail": [
    #     {"loc": ["body", "email"], "msg": "value is not a valid email", ...}
    #   ]
    # }
    pass
```

## Advanced Patterns

### Computed Fields

```python
from pydantic import computed_field

class User(BaseModel):
    first_name: str
    last_name: str

    @computed_field
    @property
    def full_name(self) -> str:
        return f"{self.first_name} {self.last_name}"

user = User(first_name="Alice", last_name="Smith")
print(user.full_name)  # "Alice Smith"
print(user.model_dump())  # Includes full_name
```

### Generic Models

```python
from typing import Generic, TypeVar
from pydantic import BaseModel

T = TypeVar("T")

class Response(BaseModel, Generic[T]):
    data: T
    message: str
    status: int

class User(BaseModel):
    id: int
    name: str

# Usage
response = Response[User](
    data=User(id=1, name="Alice"),
    message="Success",
    status=200
)
```

### Custom Types

```python
from pydantic import BaseModel, field_validator
from typing import Annotated

# ✅ Custom validation with Annotated
def validate_username(v: str) -> str:
    if len(v) < 3:
        raise ValueError("Username too short")
    if not v.isalnum():
        raise ValueError("Username must be alphanumeric")
    return v

Username = Annotated[str, field_validator(validate_username)]

class User(BaseModel):
    username: Username
    email: str
```

## Common Mistakes

### Mistake 1: Using V1 Syntax in V2

```python
# ❌ V1 syntax in V2
class User(BaseModel):
    class Config:  # V1
        orm_mode = True

user.dict()  # V1
user.json()  # V1

# ✅ V2 syntax
class User(BaseModel):
    model_config = ConfigDict(  # V2
        from_attributes=True
    )

user.model_dump()  # V2
user.model_dump_json()  # V2
```

### Mistake 2: Mutable Defaults

```python
# ❌ Mutable default
class User(BaseModel):
    tags: list[str] = []  # Shared across instances!

# ✅ Use default_factory
class User(BaseModel):
    tags: list[str] = Field(default_factory=list)
```

### Mistake 3: Not Using exclude_unset

```python
# ❌ Includes all fields (even unset)
class UserUpdate(BaseModel):
    name: str | None = None
    email: str | None = None

update_data = UserUpdate(name="Alice").model_dump()
# {'name': 'Alice', 'email': None}  # Sets email to None!

# ✅ Exclude unset fields
update_data = UserUpdate(name="Alice").model_dump(exclude_unset=True)
# {'name': 'Alice'}  # Only includes provided fields
```

---

**Full documentation**: [Pydantic V2](https://docs.pydantic.dev/latest/)
