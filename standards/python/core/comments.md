# Comments

Based on [PEP 257](https://peps.python.org/pep-0257/) docstring conventions.

## Quick Reference

- **Docstrings** - triple-quoted strings for documentation
- **Google style** - structured docstring format
- **Comments explain why** - code explains what
- **Keep current** - outdated comments are worse than none
- **Avoid obvious comments** - don't repeat code

## Core Rules

### 1. Module Docstrings

```python
# ✅ Module docstring
"""User authentication and authorization module.

Provides functions for user login, token generation, and permission checks.
"""

import os
from typing import Optional

# ❌ No module docstring or too brief
# auth.py
import os
```

### 2. Function Docstrings

```python
# ✅ Complete function docstring (Google style)
def get_user(user_id: int, db: Session) -> User | None:
    """Retrieve a user by ID from the database.

    Args:
        user_id: The unique identifier for the user.
        db: Database session for querying.

    Returns:
        User object if found, None otherwise.

    Raises:
        DatabaseError: If database connection fails.
    """
    pass

# ✅ Simple one-liner for obvious functions
def validate_email(email: str) -> bool:
    """Check if email format is valid."""
    return "@" in email

# ❌ No docstring
def process_data(data):
    pass
```

### 3. Class Docstrings

```python
# ✅ Class docstring
class UserService:
    """Service for managing user operations.

    Provides methods for creating, updating, and retrieving users.
    All methods are safe for concurrent use.

    Attributes:
        db: Database session.
        logger: Logger instance.
    """

    def __init__(self, db: Session, logger: Logger):
        self.db = db
        self.logger = logger

    def create_user(self, user: UserCreate) -> User:
        """Create a new user.

        Args:
            user: User data for creation.

        Returns:
            Created user object.

        Raises:
            ValidationError: If user data is invalid.
            DuplicateError: If email already exists.
        """
        pass
```

### 4. Inline Comments

```python
# ✅ Explain why, not what
def process_payment(amount: float) -> bool:
    # Add 1 cent to avoid floating point rounding issues
    # that caused transaction failures in production (issue #123)
    adjusted_amount = amount + 0.01
    
    return charge_card(adjusted_amount)

# ❌ Comment repeats code
def process_payment(amount: float) -> bool:
    # Add 0.01 to amount
    adjusted_amount = amount + 0.01  # Add cents
    
    # Call charge_card function
    return charge_card(adjusted_amount)
```

### 5. TODO Comments

```python
# ✅ TODO with context
# TODO(alice): Add rate limiting after user feedback.
# See issue #456 for details.
def api_endpoint():
    pass

# TODO: Refactor to use async database calls (Python 3.11+)
def fetch_users():
    pass

# ❌ Vague TODO
# TODO: fix this
def broken_function():
    pass
```

## Docstring Formats

### Google Style (Recommended)

```python
def complex_function(
    name: str,
    age: int,
    tags: list[str] | None = None
) -> dict[str, any]:
    """Process user data and return summary.

    Performs validation and data transformation on user input.
    This is the recommended format for FastAPI projects.

    Args:
        name: User's full name.
        age: User's age in years.
        tags: Optional list of user tags.

    Returns:
        Dictionary containing processed user data with keys:
            - validated_name: Cleaned name string
            - age_group: Age category
            - tag_count: Number of tags

    Raises:
        ValueError: If name is empty or age is negative.
        ValidationError: If data format is invalid.

    Example:
        >>> result = complex_function("Alice", 25, ["python", "fastapi"])
        >>> print(result["age_group"])
        'adult'
    """
    pass
```

### NumPy Style (Alternative)

```python
def numpy_style_function(x: int, y: int) -> int:
    """Add two numbers together.

    Parameters
    ----------
    x : int
        First number
    y : int
        Second number

    Returns
    -------
    int
        Sum of x and y

    Examples
    --------
    >>> numpy_style_function(2, 3)
    5
    """
    return x + y
```

## FastAPI Documentation

```python
from fastapi import FastAPI, Query
from typing import Annotated

app = FastAPI()

@app.get("/users/{user_id}")
async def get_user(
    user_id: Annotated[int, "User ID"],
    include_posts: Annotated[bool, Query(description="Include user's posts")] = False
) -> dict:
    """Retrieve user information.

    Fetches user data from database and optionally includes their posts.

    Args:
        user_id: Unique identifier for the user.
        include_posts: Whether to include user's posts in response.

    Returns:
        User data dictionary with optional posts array.

    Raises:
        HTTPException: 404 if user not found.
    """
    pass
```

## Pydantic Model Documentation

```python
from pydantic import BaseModel, Field

class User(BaseModel):
    """User model for API requests and responses.

    Represents a user in the system with validation for all fields.
    """

    id: int = Field(description="Unique user identifier")
    name: str = Field(min_length=1, max_length=100, description="User's full name")
    email: str = Field(pattern=r"^[\w\.-]+@[\w\.-]+\.\w+$", description="Email address")
    age: int | None = Field(default=None, ge=0, le=150, description="User's age in years")

    model_config = {
        "json_schema_extra": {
            "examples": [
                {
                    "id": 1,
                    "name": "Alice Smith",
                    "email": "alice@example.com",
                    "age": 30
                }
            ]
        }
    }
```

## When Not to Comment

### Don't Comment Obvious Code

```python
# ❌ Obvious comment
# Increment counter by 1
counter += 1

# ❌ Repeats variable name
# User name
user_name = "Alice"

# ✅ No comment needed
counter += 1
user_name = "Alice"
```

### Use Good Naming Instead

```python
# ❌ Comment explains bad name
# d is duration in seconds
d = 300

# ✅ Clear naming, no comment needed
timeout_seconds = 300
```

### Don't Comment Commented Code

```python
# ❌ Don't commit commented code
def process_data(data):
    # Old implementation
    # result = old_process(data)
    # if result:
    #     return result
    
    return new_process(data)

# ✅ Delete old code (Git preserves history)
def process_data(data):
    return new_process(data)
```

## Common Mistakes

### Mistake 1: Outdated Comments

```python
# ❌ Comment doesn't match code
# Returns user name
def get_user(user_id: int) -> User:  # Returns User object!
    pass

# ✅ Keep comments in sync
# Returns user object with all fields
def get_user(user_id: int) -> User:
    pass
```

### Mistake 2: Over-commenting

```python
# ❌ Too many obvious comments
def calculate_total(items):
    # Initialize total to zero
    total = 0
    
    # Loop through each item
    for item in items:
        # Add item price to total
        total += item.price
    
    # Return the total
    return total

# ✅ Only comment non-obvious parts
def calculate_total(items):
    total = 0
    for item in items:
        total += item.price
    return total
```

### Mistake 3: Missing Docstrings

```python
# ❌ Public function without docstring
def process_payment(amount: float, user_id: int) -> bool:
    pass

# ✅ Public function with docstring
def process_payment(amount: float, user_id: int) -> bool:
    """Process payment for user.

    Args:
        amount: Payment amount in USD.
        user_id: User identifier.

    Returns:
        True if payment succeeded, False otherwise.
    """
    pass
```

---

**Full details**:
- [PEP 257 – Docstring Conventions](https://peps.python.org/pep-0257/)
- [Google Python Style Guide - Comments](https://google.github.io/styleguide/pyguide.html#38-comments-and-docstrings)
