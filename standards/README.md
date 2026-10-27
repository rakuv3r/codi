# Coding Standards

Language-agnostic coding standards and library-specific best practices.

## Architecture

```
standards/
├── universal/          # Language-agnostic principles
├── go/
│   ├── core/          # Go language features
│   └── [import-path]/ # Library best practices (github.com/gin-gonic/gin.md)
└── python/
    ├── core/          # Python language features
    └── [package].md   # Library best practices (fastapi.md)
```

## Supported Languages

- [Go](#go)
- [Python](#python)

---

## Go

### Core Language
- [Naming Conventions](go/core/naming.md)
- [Error Handling](go/core/error-handling.md)
- [Testing](go/core/testing.md)
- [Concurrency](go/core/concurrency.md)
- [Context](go/core/context.md)
- [Interfaces](go/core/interfaces.md)
- [Imports](go/core/imports.md)
- [Formatting](go/core/formatting.md)
- [Comments](go/core/comments.md)
- [Performance](go/core/performance.md)
- [Logging (log/slog)](go/core/logging.md)

### Libraries & Frameworks

#### Web Frameworks
- [github.com/gin-gonic/gin](go/github.com/gin-gonic/gin.md) - HTTP web framework

#### Databases
- [gorm.io/gorm](go/gorm.io/gorm.md) - ORM library

#### Security
- [github.com/golang-jwt/jwt/v5](go/github.com/golang-jwt/jwt/v5.md) - JWT implementation

#### Configuration
- [github.com/spf13/viper](go/github.com/spf13/viper.md) - Configuration management

[See all Go standards →](go/README.md)

---

## Python

### Core Language
- [Naming Conventions](python/core/naming.md)
- [Error Handling](python/core/error-handling.md)
- [Testing](python/core/testing.md)
- [Type Hints](python/core/type-hints.md)
- [Imports](python/core/imports.md)
- [Async Programming](python/core/async.md)
- [Formatting](python/core/formatting.md)
- [Comments](python/core/comments.md)

### Libraries & Frameworks

#### Web Frameworks
- [fastapi](python/fastapi.md) - Modern async web framework

#### Data Validation
- [pydantic](python/pydantic.md) - Data validation using Python type hints

#### Databases
- [sqlalchemy](python/sqlalchemy.md) - SQL toolkit and ORM (SQLAlchemy 2.0)

#### Security
- [passlib](python/passlib.md) - Password hashing (Argon2)

#### Logging
- [loguru](python/loguru.md) - Modern logging library

#### Configuration
- [pydantic-settings](python/pydantic-settings.md) - Settings management

#### Testing
- [pytest](python/pytest.md) - Testing framework

[See all Python standards →](python/README.md)

---

## Universal Principles

Language-agnostic design and architecture principles:

- [API Design](universal/api-design.md) - RESTful API design, HTTP methods, error responses
- [Security Principles](universal/security-principles.md) - Authentication, authorization, password hashing
- [Database Principles](universal/database-principles.md) - Connection pooling, transactions, indexing
- [Design Principles](universal/design-principles.md) - DRY, KISS, YAGNI, SOLID
- [Documentation](universal/documentation.md) - README templates, 盘古之白, changelog

[See all universal standards →](universal/README.md)

---

## Adding New Standards

### For a new library

1. Create file following import path:
   ```
   # Go
   golang/github.com/labstack/echo.md
   
   # Python
   python/django.md
   ```

2. Document best practices for that library

3. No configuration needed - auto-detected by `codi`

### For a new language

1. Create directory structure:
   ```
   standards/
   └── rust/
       ├── core/
       │   ├── naming.md
       │   └── ...
       └── [crate-path]/
   ```

2. Add language to README

---

## How It Works

When you run `codi adopt` or `codi init`, it:

1. Scans `go.mod` / `pyproject.toml`
2. Maps dependencies to standards files (zero-config)
3. Generates `.claude/CLAUDE.md` with relevant standards
4. Auto-syncs on every commit (via Git hook)

**Example:**

```go
// go.mod
require github.com/gin-gonic/gin v1.9.0
```

↓

```markdown
# .claude/CLAUDE.md
Standards:
- go/core/naming.md
- go/github.com/gin-gonic/gin.md
```

Zero mapping. Path is the standard.
