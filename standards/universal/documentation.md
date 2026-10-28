---
language: universal
category: universal
applies_to: []
---

# Documentation

Documentation standards for clear, maintainable project documentation.

## Quick Reference

- **盘古之白** - space between Chinese and English/numbers
- **README template** - consistent structure
- **API documentation** - OpenAPI/Swagger
- **Keep current** - outdated docs worse than none

## Core Rules

### 1. 盘古之白 (Pangu Spacing)

Chinese and English/numbers must be separated by spaces for readability.

```
# ❌ No spacing
在GitHub上查看项目，使用Python3.11开发。

# ✅ With spacing (盘古之白)
在 GitHub 上查看项目，使用 Python 3.11 开发。

# ❌ No spacing
这个API支持100个请求/分钟的速率限制。

# ✅ With spacing
这个 API 支持 100 个请求 / 分钟的速率限制。
```

**Rules**:
- Space between Chinese and English: `在 GitHub 上`
- Space between Chinese and numbers: `使用 Python 3.11`
- Space between Chinese and symbols: `支持 100 个`
- No space before Chinese punctuation: `项目，使用` not `项目 ，使用`

**Exceptions**:
- Proper nouns: `GitHub` not `Git Hub`
- Version numbers: `Python 3.11` not `Python3.11`
- Units: `100 MB` not `100MB`

### 2. Punctuation

```
# ✅ Chinese text - Chinese punctuation
错误：配置文件不存在。
这是一个示例，展示如何使用。

# ✅ English text - English punctuation
Error: config file not found.
This is an example, showing how to use it.

# ❌ Mixed punctuation
错误: 配置文件不存在.  # Wrong: English colon and period
Error：config file not found。  # Wrong: Chinese colon and period
```

**Guidelines**:
- Chinese: `，。！？：；`
- English: `, . ! ? : ;`
- Code/URLs: always English punctuation

### 3. README Template

```markdown
# Project Name

[One-line description]

## Features

- Feature 1
- Feature 2
- Feature 3

## Quick Start

### Prerequisites

- Go 1.21+ / Python 3.11+
- PostgreSQL 15+
- Redis 7+

### Installation

\`\`\`bash
git clone https://github.com/user/project
cd project
make install
\`\`\`

### Configuration

\`\`\`bash
cp .env.example .env
# Edit .env with your settings
\`\`\`

### Running

\`\`\`bash
make run
# Or
go run cmd/server/main.go
\`\`\`

## API Documentation

API docs available at: http://localhost:8080/docs

## Development

### Project Structure

\`\`\`
project/
├── cmd/           # Main applications
├── internal/      # Private code
├── pkg/           # Public libraries
├── api/           # API definitions
└── docs/          # Documentation
\`\`\`

### Running Tests

\`\`\`bash
make test
\`\`\`

### Code Standards

See [CONTRIBUTING.md](CONTRIBUTING.md) for coding standards.

## Deployment

[Deployment instructions]

## License

[License type]
```

### 4. API Documentation

**Use OpenAPI/Swagger**:

```yaml
# openapi.yaml
openapi: 3.0.0
info:
  title: User API
  version: 1.0.0

paths:
  /users:
    get:
      summary: List users
      parameters:
        - name: page
          in: query
          schema:
            type: integer
            default: 1
      responses:
        '200':
          description: Success
          content:
            application/json:
              schema:
                type: object
                properties:
                  data:
                    type: array
                    items:
                      $ref: '#/components/schemas/User'
  
  /users/{id}:
    get:
      summary: Get user by ID
      parameters:
        - name: id
          in: path
          required: true
          schema:
            type: integer
      responses:
        '200':
          description: Success
        '404':
          description: User not found

components:
  schemas:
    User:
      type: object
      properties:
        id:
          type: integer
        name:
          type: string
        email:
          type: string
```

### 5. Code Comments

```
# ✅ Explain why, not what
# Add 1 cent to avoid floating point rounding issues
# that caused transaction failures (issue #123)
adjusted_amount = amount + 0.01

# ❌ Repeat code
# Add 0.01 to amount
adjusted_amount = amount + 0.01
```

**When to comment**:
- Non-obvious business logic
- Workarounds for bugs
- Performance optimizations
- Security considerations

**When NOT to comment**:
- Obvious operations: `user_count += 1`
- Repeating function names
- Explaining what code does (use clear naming instead)

### 6. Changelog

```markdown
# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/).

## [Unreleased]

### Added
- New feature X

### Changed
- Updated dependency Y to v2.0

### Fixed
- Bug fix for issue #123

## [1.2.0] - 2026-10-15

### Added
- User authentication with JWT
- Rate limiting middleware

### Changed
- Updated Go to 1.21

### Deprecated
- Old auth endpoint `/auth/login` (use `/v2/auth/login`)

### Removed
- Deprecated `/v1/users` endpoint

### Fixed
- Fixed memory leak in connection pool

### Security
- Updated dependencies to fix CVE-2026-1234
```

### 7. Architecture Documentation

**ADR (Architecture Decision Record)**:

```markdown
# ADR 001: Use PostgreSQL for Primary Database

## Status

Accepted

## Context

We need to choose a database for storing user data and orders.

Requirements:
- ACID transactions
- JSON support
- High performance for read-heavy workload
- Open source

## Decision

Use PostgreSQL 15+ as primary database.

## Consequences

### Positive
- Strong ACID guarantees
- Excellent JSON support (JSONB)
- Mature ecosystem
- Good performance

### Negative
- Vertical scaling limits
- Complex replication setup

### Neutral
- Team needs to learn PostgreSQL-specific features
```

## Common Mistakes

### Mistake 1: Outdated Documentation

```
# ❌ Documentation says one thing, code does another
# README: "Run with: python app.py"
# Actual: Project switched to FastAPI, run with: uvicorn main:app

# ✅ Keep documentation in sync
Update README when changing run commands
Use CI to validate documentation examples
```

### Mistake 2: No 盘古之白

```
# ❌ Hard to read
在GitHub上使用Python3.11开发的项目。

# ✅ Readable
在 GitHub 上使用 Python 3.11 开发的项目。
```

### Mistake 3: Missing Context

```
# ❌ No context
# Fix bug
if amount > 0:
    adjusted_amount = amount + 0.01

# ✅ With context
# Add 1 cent to avoid floating point rounding that caused
# payment failures in production (issue #123, 2026-09-15)
if amount > 0:
    adjusted_amount = amount + 0.01
```

---

**Tools**:
- [pangu.js](https://github.com/vinta/pangu.js) - auto spacing for web
- [autocorrect](https://github.com/huacnlee/autocorrect) - CLI tool for spacing
- [Swagger UI](https://swagger.io/tools/swagger-ui/) - API documentation

**References**:
- [盘古之白](https://github.com/vinta/pangu.js)
- [Keep a Changelog](https://keepachangelog.com/)
- [ADR Template](https://github.com/joelparkerhenderson/architecture-decision-record)
