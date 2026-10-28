---
language: python
category: core
applies_to: []
---

# Naming

Based on [PEP 8](https://peps.python.org/pep-0008/) and [Google Python Style Guide](https://google.github.io/styleguide/pyguide.html).

## Quick Reference

- **Modules/Packages**: `lower_with_under` (lowercase with underscores)
- **Classes**: `CapWords` (PascalCase)
- **Functions/Variables**: `lower_with_under` (lowercase with underscores)
- **Constants**: `CAPS_WITH_UNDER` (all caps with underscores)
- **Private**: `_` prefix (single underscore)

## Core Rules

### 1. Modules and Packages

```python
# ✅ Module names: lowercase with underscores
user_service.py
database_client.py
http_handler.py

# ✅ Package names: lowercase, avoid underscores
mypackage/
    __init__.py
    userservice/
    database/

# ❌ Don't use dashes or uppercase
user-service.py  # Wrong: can't use -
UserService.py   # Wrong: don't capitalize
```

### 2. Class Names

```python
# ✅ Classes: CapWords (PascalCase)
class UserService:
    pass

class HTTPClient:  # Acronyms all caps
    pass

class DatabaseConnection:
    pass

# ✅ Exceptions: Error suffix
class ValidationError(Exception):
    pass

class DatabaseConnectionError(Exception):
    pass

# ❌ Don't use underscores
class User_Service:  # Wrong
    pass
```

### 3. Functions and Methods

```python
# ✅ Functions: lowercase with underscores
def get_user(user_id: int) -> User:
    pass

def process_data(data: list[dict]) -> None:
    pass

# ✅ Private methods: single underscore prefix
class Service:
    def _internal_method(self) -> None:
        pass

# ✅ Instance methods first parameter: self
class User:
    def save(self) -> None:
        pass

# ✅ Class methods first parameter: cls
class User:
    @classmethod
    def from_dict(cls, data: dict) -> "User":
        pass

# ❌ Don't use camelCase
def getUserData():  # Wrong
    pass
```

### 4. Variables

```python
# ✅ Variables: lowercase with underscores
user_count = 10
max_retries = 3
connection_pool = []

# ✅ Private variables: single underscore prefix
class Config:
    _api_key = "secret"

# ❌ Don't use camelCase
userCount = 10  # Wrong
```

### 5. Constants

```python
# ✅ Constants: all caps with underscores
MAX_CONNECTIONS = 100
DEFAULT_TIMEOUT = 30
API_BASE_URL = "https://api.example.com"

# ✅ Private constants
_INTERNAL_BUFFER_SIZE = 1024

# ❌ Don't use lowercase
max_connections = 100  # Wrong: not constant style
```

### 6. Underscore Meanings

```python
# ✅ Single leading _ : internal use (weak private)
class Service:
    def _internal_helper(self):
        pass

# ✅ Single trailing _ : avoid keyword conflicts
class_ = type("MyClass", (), {})
type_ = "string"

# ⚠️ Double leading __ : name mangling (not recommended, use single _)
class Example:
    def __private(self):  # Becomes _Example__private
        pass

# ✅ Double leading and trailing __name__ : magic methods (Python reserved)
def __init__(self):
    pass
```

### 7. Acronym Rules

```python
# ✅ Acronyms all caps
class HTTPServer:
    pass

class URLParser:
    pass

class XMLProcessor:
    pass

# ❌ Mixed case
class HttpServer:  # Wrong
    pass
```

### 8. Names to Avoid

```python
# ❌ Single character l, O, I (confusing)
l = 1  # Looks like 1
O = 0  # Looks like 0
I = 1  # Looks like l

# ✅ Exception: loop counters
for i in range(10):
    pass

# ❌ Don't encode type in name
user_dict = {}  # Wrong: type is obvious
user_list = []  # Wrong

# ✅ Descriptive names
users = {}
active_users = []
```

## Special Cases

### FastAPI Path Functions

```python
# ✅ Path functions: verb + noun
@app.get("/users/{user_id}")
async def get_user(user_id: int) -> User:
    pass

@app.post("/users")
async def create_user(user: UserCreate) -> User:
    pass

# ✅ Dependencies: descriptive names
def get_db() -> Generator:
    pass

def get_current_user(token: str = Depends(oauth2_scheme)) -> User:
    pass
```

### Pydantic Models

```python
# ✅ Models: CapWords
class UserBase(BaseModel):
    name: str
    email: str

class UserCreate(UserBase):
    password: str

class UserResponse(UserBase):
    id: int
    created_at: datetime

# ✅ Fields: lowercase with underscores
class Config(BaseModel):
    api_key: str
    max_retries: int
    base_url: str
```

### Test Functions

```python
# ✅ Test functions: test_ prefix + description
def test_user_creation():
    pass

def test_get_user_returns_404_when_not_found():
    pass

def test_validate_email_raises_error_on_invalid_format():
    pass

# ❌ Don't use vague names
def test_1():  # Wrong
    pass
```

## Common Mistakes

### Mistake 1: Mixed Style

```python
# ❌ JavaScript/Java style
def getUserById(userId):
    pass

class userService:
    pass

# ✅ Python style
def get_user_by_id(user_id):
    pass

class UserService:
    pass
```

### Mistake 2: Over-abbreviation

```python
# ❌ Unclear abbreviations
def proc_usr_dat(uid):
    pass

# ✅ Clear names
def process_user_data(user_id):
    pass
```

### Mistake 3: Type in Name

```python
# ❌ Redundant type information
user_dict: dict = {}
id_list: list = []

# ✅ Type expressed via type hints
users: dict[int, User] = {}
user_ids: list[int] = []
```

---

**Full details**: [PEP 8 - Naming Conventions](https://peps.python.org/pep-0008/#naming-conventions)
