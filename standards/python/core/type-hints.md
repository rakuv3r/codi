---
language: python
category: core
applies_to: []
---

# Type Hints

Based on [typing best practices](https://typing.python.org/en/latest/reference/best_practices.html) and Python 3.10+ features.

## Quick Reference

- **Function signatures** - parameter and return types
- **Variable annotations** - complex types need annotations
- **Python 3.10+** - use `|` instead of `Union`
- **Generics** - `list[int]` instead of `List[int]`
- **Optional** - `str | None` instead of `Optional[str]`

## Core Rules

### 1. Function Type Hints

```python
# ✅ Complete type hints
def get_user(user_id: int) -> User:
    pass

def process_data(data: list[dict[str, any]]) -> None:
    pass

async def fetch_users(limit: int = 10) -> list[User]:
    pass

# ✅ Multiple return values
def parse_name(full_name: str) -> tuple[str, str]:
    first, last = full_name.split()
    return first, last

# ❌ Missing types
def get_user(user_id):  # Wrong
    pass
```

### 2. Optional Types (Python 3.10+)

```python
# ✅ Use | None (Python 3.10+)
def get_user(user_id: int) -> User | None:
    pass

def find_email(name: str) -> str | None:
    pass

# ❌ Old-style Optional (Python 3.9-)
from typing import Optional

def get_user(user_id: int) -> Optional[User]:  # Not recommended
    pass
```

### 3. Union Types

```python
# ✅ Use | (Python 3.10+)
def process(value: int | str | float) -> str:
    return str(value)

def parse(data: dict | list) -> any:
    pass

# ❌ Old-style Union
from typing import Union

def process(value: Union[int, str, float]) -> str:  # Not recommended
    pass
```

### 4. Generic Containers (Python 3.9+)

```python
# ✅ Built-in generics (Python 3.9+)
def sum_numbers(numbers: list[int]) -> int:
    return sum(numbers)

def get_mapping() -> dict[str, int]:
    return {"a": 1, "b": 2}

def process_items(items: set[str]) -> list[str]:
    return sorted(items)

# ❌ Old-style typing
from typing import List, Dict, Set

def sum_numbers(numbers: List[int]) -> int:  # Not recommended
    pass
```

### 5. Complex Types

```python
from collections.abc import Callable, Iterable, Mapping

# ✅ Callable
def apply_func(func: Callable[[int], str], value: int) -> str:
    return func(value)

# ✅ Iterable
def process_items(items: Iterable[str]) -> list[str]:
    return [item.upper() for item in items]

# ✅ Mapping
def merge_dicts(d1: Mapping[str, int], d2: Mapping[str, int]) -> dict[str, int]:
    return {**d1, **d2}

# ✅ Nested generics
def group_by_type(items: list[dict[str, any]]) -> dict[str, list[dict[str, any]]]:
    pass
```

### 6. TypeAlias (Python 3.10+)

```python
from typing import TypeAlias

# ✅ Type aliases
UserId: TypeAlias = int
UserDict: TypeAlias = dict[str, any]
Headers: TypeAlias = dict[str, str]

def get_user(user_id: UserId) -> UserDict:
    pass

def make_request(headers: Headers) -> str:
    pass

# ✅ Complex type alias
JsonValue: TypeAlias = str | int | float | bool | None | list["JsonValue"] | dict[str, "JsonValue"]

def parse_json(data: str) -> JsonValue:
    pass
```

### 7. Protocol (Structural Subtyping)

```python
from typing import Protocol

# ✅ Define Protocol
class Drawable(Protocol):
    def draw(self) -> None: ...

class Circle:
    def draw(self) -> None:
        print("Drawing circle")

class Square:
    def draw(self) -> None:
        print("Drawing square")

def render(obj: Drawable) -> None:  # Both Circle and Square satisfy
    obj.draw()

# ✅ Protocol for duck typing
class SupportsWrite(Protocol):
    def write(self, data: str) -> int: ...

def log_to(file: SupportsWrite, message: str) -> None:
    file.write(message)
```

### 8. Type Variables

```python
from typing import TypeVar

# ✅ Generic functions
T = TypeVar("T")

def first(items: list[T]) -> T | None:
    return items[0] if items else None

def identity(value: T) -> T:
    return value

# ✅ Bounded type variable
from numbers import Number

N = TypeVar("N", bound=Number)

def add(a: N, b: N) -> N:
    return a + b  # type: ignore
```

## FastAPI Type Hints

```python
from fastapi import Depends, Query
from typing import Annotated

# ✅ Query parameter types
@app.get("/search")
async def search(
    q: Annotated[str, Query(min_length=3, max_length=50)],
    limit: Annotated[int, Query(ge=1, le=100)] = 10
) -> list[dict]:
    pass

# ✅ Dependency injection types
async def get_current_user(
    token: Annotated[str, Depends(oauth2_scheme)]
) -> User:
    pass

@app.get("/me")
async def read_users_me(
    current_user: Annotated[User, Depends(get_current_user)]
) -> User:
    return current_user
```

## Pydantic Models

```python
from pydantic import BaseModel, Field

# ✅ Strict types
class UserCreate(BaseModel):
    name: str
    email: str
    age: int = Field(ge=0, le=150)
    tags: list[str] = []
    metadata: dict[str, any] | None = None

class UserResponse(BaseModel):
    id: int
    name: str
    email: str
    created_at: datetime
    
    class Config:
        from_attributes = True
```

## Common Patterns

### Typed Dictionaries

```python
from typing import TypedDict

# ✅ TypedDict
class UserDict(TypedDict):
    id: int
    name: str
    email: str

def create_user() -> UserDict:
    return {"id": 1, "name": "Alice", "email": "alice@example.com"}

# ✅ Partial optional fields
class UserDict(TypedDict, total=False):
    id: int  # Required
    name: str  # Required
    age: int  # Optional
```

### Literal Types

```python
from typing import Literal

# ✅ Constrain values
def set_mode(mode: Literal["read", "write", "append"]) -> None:
    pass

set_mode("read")  # OK
set_mode("delete")  # Type error

Status = Literal["pending", "active", "completed"]

def update_status(status: Status) -> None:
    pass
```

### Using Any

```python
from typing import Any

# ⚠️ Avoid Any when possible
def process(data: Any) -> Any:  # Loses type safety
    pass

# ✅ Use specific types
def process(data: dict[str, str | int]) -> list[str]:
    pass

# ✅ Reasonable use of Any
def serialize_to_json(obj: Any) -> str:  # JSON can be any type
    return json.dumps(obj)
```

## Type Checking Tools

```bash
# mypy
mypy app/

# pyright (faster)
pyright app/

# Configure mypy
# mypy.ini
[mypy]
python_version = 3.11
warn_return_any = True
warn_unused_configs = True
disallow_untyped_defs = True
```

## Common Mistakes

### Mistake 1: Missing Return Type

```python
# ❌ Missing return type
def get_user(user_id: int):
    return User(id=user_id)

# ✅ Add return type
def get_user(user_id: int) -> User:
    return User(id=user_id)
```

### Mistake 2: Old-Style Syntax

```python
# ❌ Python 3.9- syntax
from typing import List, Dict, Optional

def process(items: List[str]) -> Optional[Dict[str, int]]:
    pass

# ✅ Python 3.10+ syntax
def process(items: list[str]) -> dict[str, int] | None:
    pass
```

### Mistake 3: Overusing Any

```python
# ❌ Abusing Any
def process_data(data: Any) -> Any:
    pass

# ✅ Specific types
from collections.abc import Mapping

def process_data(data: Mapping[str, int | str]) -> list[tuple[str, int]]:
    pass
```

---

**Full documentation**:
- [typing — Support for type hints](https://docs.python.org/3/library/typing.html)
- [PEP 484 – Type Hints](https://peps.python.org/pep-0484/)
- [Python Type Hints in 2026](https://codegym.cc/groups/posts/python-type-hints-2026-guide)
