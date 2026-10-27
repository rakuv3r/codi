# Formatting

Based on [PEP 8](https://peps.python.org/pep-0008/) and [Black](https://black.readthedocs.io/) code style.

## Quick Reference

- **Black** - opinionated code formatter
- **4 spaces** - indentation (not tabs)
- **88 characters** - line length (Black default)
- **ruff** - fast linter and formatter
- **Format on save** - configure editor

## Core Rules

### 1. Use Black

```bash
# Format code
black app/

# Check without changes
black --check app/

# Configure in pyproject.toml
[tool.black]
line-length = 88
target-version = ['py311']
```

### 2. Line Length

```python
# ✅ 88 characters (Black default)
def long_function_name(
    var_one: str, var_two: str, var_three: str, var_four: str
) -> str:
    return f"{var_one} {var_two} {var_three} {var_four}"

# ✅ Break long lines
result = some_function_that_takes_arguments(
    argument_one,
    argument_two,
    argument_three,
)
```

### 3. Indentation

```python
# ✅ 4 spaces per level
def function():
    if condition:
        do_something()

# ❌ Tabs or inconsistent spacing
def function():
	if condition:  # Tab
	  do_something()  # 2 spaces
```

### 4. Blank Lines

```python
# ✅ Two blank lines between top-level definitions
class MyClass:
    pass


def my_function():
    pass


# ✅ One blank line between methods
class MyClass:
    def method_one(self):
        pass

    def method_two(self):
        pass

# ✅ Logical sections within functions
def complex_function():
    # Setup
    x = 1
    y = 2
    
    # Process
    result = x + y
    
    # Return
    return result
```

### 5. Imports Formatting

```python
# ✅ One import per line
import os
import sys
import json

# ✅ Grouped and sorted (use isort)
import asyncio
import os

import fastapi
import pydantic

from app.models import User
```

### 6. String Quotes

```python
# ✅ Consistent quotes (Black uses double by default)
name = "Alice"
message = "Hello, world!"

# ✅ Use single for string containing doubles
text = 'He said "Hello"'

# ✅ Triple quotes for docstrings
def function():
    """
    This is a docstring.
    """
    pass
```

### 7. Trailing Commas

```python
# ✅ Trailing comma in multi-line (Black adds automatically)
items = [
    "item1",
    "item2",
    "item3",  # Trailing comma
]

function_call(
    arg1,
    arg2,
    arg3,  # Trailing comma
)

# ✅ Single line: no trailing comma
items = ["item1", "item2", "item3"]
```

## Tool Configuration

### pyproject.toml

```toml
[tool.black]
line-length = 88
target-version = ['py311']
include = '\.pyi?$'

[tool.isort]
profile = "black"
line_length = 88
multi_line_output = 3

[tool.ruff]
line-length = 88
target-version = "py311"

[tool.ruff.lint]
select = [
    "E",  # pycodestyle errors
    "W",  # pycodestyle warnings
    "F",  # pyflakes
    "I",  # isort
]
```

### VS Code Settings

```json
{
  "[python]": {
    "editor.formatOnSave": true,
    "editor.defaultFormatter": "ms-python.black-formatter",
    "editor.codeActionsOnSave": {
      "source.organizeImports": true
    }
  },
  "python.linting.enabled": true,
  "python.linting.ruffEnabled": true
}
```

## Pre-commit Hook

```yaml
# .pre-commit-config.yaml
repos:
  - repo: https://github.com/psf/black
    rev: 24.1.1
    hooks:
      - id: black

  - repo: https://github.com/pycqa/isort
    rev: 5.13.2
    hooks:
      - id: isort

  - repo: https://github.com/astral-sh/ruff-pre-commit
    rev: v0.2.0
    hooks:
      - id: ruff
        args: [--fix]
```

## Common Mistakes

### Mistake 1: Manual Formatting

```python
# ❌ Manual alignment (hard to maintain)
result = my_function(arg1,
                     arg2,
                     arg3)

# ✅ Let Black handle it
result = my_function(
    arg1,
    arg2,
    arg3,
)
```

### Mistake 2: Inconsistent Style

```python
# ❌ Mixed quotes and spacing
name='Alice'
email =  "alice@example.com"

# ✅ Consistent (Black enforces)
name = "Alice"
email = "alice@example.com"
```

### Mistake 3: Not Using Formatter

```python
# ❌ Fighting the formatter
# fmt: off
result = my_badly_formatted_code( arg1,arg2,arg3 )
# fmt: on

# ✅ Let formatter work
result = my_properly_formatted_code(arg1, arg2, arg3)
```

---

**Tools**:
- [Black](https://black.readthedocs.io/)
- [isort](https://pycqa.github.io/isort/)
- [ruff](https://docs.astral.sh/ruff/)
