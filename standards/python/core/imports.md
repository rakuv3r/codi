# Imports

Based on [PEP 8](https://peps.python.org/pep-0008/) import conventions.

## Quick Reference

- **Standard library, third-party, local** - three groups separated by blank lines
- **Alphabetical within groups** - sorted order
- **Absolute imports** - prefer over relative
- **One import per line** - for clarity
- **No wildcard imports** - avoid `from module import *`

## Core Rules

### 1. Import Grouping

```python
# ✅ Standard grouping (3 groups, blank line separated)
import os
import sys
from pathlib import Path

import numpy as np
import requests
from fastapi import FastAPI

from app.config import settings
from app.models import User

# ❌ No grouping
import os
import requests
from app.models import User
import sys
```

### 2. Import Order

```python
# ✅ Alphabetical within each group
import asyncio
import json
import os

import fastapi
import pydantic
import sqlalchemy

from app.config import settings
from app.database import get_db
from app.models import User

# ❌ Random order
import os
import json
import asyncio
```

### 3. Import Style

```python
# ✅ One import per line (for readability)
import os
import sys
import json

# ✅ Multiple items from same module
from typing import Any, Callable, Protocol

# ❌ Multiple modules on one line
import os, sys, json  # Wrong
```

### 4. Absolute vs Relative Imports

```python
# ✅ Prefer absolute imports
from app.models.user import User
from app.services.auth import authenticate

# ✅ Relative imports OK within same package
# In app/services/email.py
from .auth import authenticate  # Same level
from ..models.user import User  # Parent level

# ❌ Confusing relative imports
from ...utils import helper  # Too many levels
```

### 5. Avoid Wildcard Imports

```python
# ❌ Wildcard import (unclear what's imported)
from typing import *
from app.models import *

# ✅ Explicit imports
from typing import Any, Protocol
from app.models import User, Post
```

### 6. Import Aliases

```python
# ✅ Standard aliases
import numpy as np
import pandas as pd
import matplotlib.pyplot as plt

# ✅ Resolve name conflicts
from app.models import User as AppUser
from third_party import User as ThirdPartyUser

# ✅ Shorten long names
from app.services.authentication import AuthenticationService as AuthService

# ❌ Non-standard aliases
import numpy as n  # Wrong: use np
import requests as r  # Wrong: no standard alias
```

### 7. Conditional Imports

```python
# ✅ Runtime conditional import
from typing import TYPE_CHECKING

if TYPE_CHECKING:
    from app.models import User  # Only for type checking

def get_user() -> "User":  # String annotation to avoid runtime import
    from app.models import User  # Import at runtime
    return User()

# ✅ Optional dependency
try:
    import uvloop
except ImportError:
    uvloop = None

if uvloop:
    uvloop.install()
```

## FastAPI Imports

```python
# ✅ Organized FastAPI imports
from fastapi import (
    Depends,
    FastAPI,
    HTTPException,
    status,
)
from fastapi.responses import JSONResponse
from pydantic import BaseModel, Field

from app.database import get_db
from app.models import User
```

## Common Patterns

### Import from __init__.py

```python
# app/models/__init__.py
from .user import User
from .post import Post

__all__ = ["User", "Post"]

# Other files can import
from app.models import User, Post
```

### Lazy Imports

```python
# ✅ Import heavy module only when needed
def process_large_data():
    import pandas as pd  # Only import when function called
    return pd.DataFrame(...)

# ✅ Plugin-style imports
def load_plugin(name: str):
    if name == "json":
        import json
        return json
    elif name == "yaml":
        import yaml
        return yaml
```

## Common Mistakes

### Mistake 1: Circular Imports

```python
# ❌ Circular import
# user.py
from app.post import Post

class User:
    posts: list[Post]

# post.py
from app.user import User

class Post:
    author: User

# ✅ Use TYPE_CHECKING
# user.py
from typing import TYPE_CHECKING

if TYPE_CHECKING:
    from app.post import Post

class User:
    posts: list["Post"]

# post.py
from typing import TYPE_CHECKING

if TYPE_CHECKING:
    from app.user import User

class Post:
    author: "User"
```

### Mistake 2: Unused Imports

```python
# ❌ Import but not use
import os
import sys  # Not used

def main():
    print(os.getcwd())

# ✅ Remove unused
import os

def main():
    print(os.getcwd())
```

### Mistake 3: Import Inside Loop

```python
# ❌ Import inside loop
def process_items(items):
    for item in items:
        import json  # Imported every iteration!
        data = json.loads(item)

# ✅ Import at top
import json

def process_items(items):
    for item in items:
        data = json.loads(item)
```

## Tools

### isort

```bash
# Sort imports automatically
isort app/

# Configure in pyproject.toml
[tool.isort]
profile = "black"
line_length = 88
multi_line_output = 3
```

### ruff

```bash
# Check and fix imports
ruff check --select I --fix app/
```

---

**Full details**: [PEP 8 - Imports](https://peps.python.org/pep-0008/#imports)
