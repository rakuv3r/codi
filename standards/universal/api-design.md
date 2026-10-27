# API Design

Based on [Microsoft Azure API Design](https://learn.microsoft.com/azure/architecture/microservices/design/api-design), [Google API Design Guide](https://cloud.google.com/apis/design), and [RFC 9457](https://www.rfc-editor.org/rfc/rfc9457.html).

## Quick Reference

- **Resource-oriented** - use nouns, not verbs
- **Standard HTTP methods** - GET/POST/PUT/PATCH/DELETE
- **RFC 9457** - standardized error responses
- **Cursor pagination** - preferred for large datasets
- **Idempotency** - PUT/DELETE idempotent, POST uses idempotency key

## Core Rules

### 1. URL Design (Resource-Oriented)

```
# ✅ Resource-oriented design
GET    /users                # Collection
GET    /users/{id}           # Single resource
POST   /users                # Create
PUT    /users/{id}           # Full replacement
PATCH  /users/{id}           # Partial update
DELETE /users/{id}           # Delete

# ✅ Nested resources (max 2 levels)
GET    /users/{id}/posts
POST   /users/{id}/posts
GET    /users/{id}/posts/{post_id}

# ❌ Avoid deep nesting (hard to maintain)
GET    /users/{id}/posts/{post_id}/comments/{comment_id}/likes

# ✅ Flatten instead
GET    /comments/{id}/likes
```

**Naming rules**:
- Plural nouns: `/users` not `/user`
- Lowercase: `/users` not `/Users`
- Hyphens for separation: `/user-profiles` not `/user_profiles`
- No verbs: `/users` not `/getUsers`

### 2. HTTP Method Semantics

Based on [RFC 9110](https://www.rfc-editor.org/rfc/rfc9110.html) HTTP semantics.

```
GET     - Retrieve resource (safe, idempotent, cacheable)
POST    - Create resource or trigger action (non-idempotent)
PUT     - Full replacement of resource (idempotent)
PATCH   - Partial update of resource (usually idempotent)
DELETE  - Delete resource (idempotent)
HEAD    - Get response headers only (same semantics as GET)
OPTIONS - Get supported methods
```

**Idempotency**:
- GET, PUT, DELETE - naturally idempotent
- POST - non-idempotent, requires idempotency key
- PATCH - should be designed as idempotent

```
# ✅ POST idempotency implementation
POST /orders
Header: Idempotency-Key: uuid-550e8400-e29b-41d4-a716-446655440000

# Same key returns same result (201 or 200)
```

### 3. HTTP Status Codes

```
2xx - Success
  200 OK              - GET/PUT/PATCH success
  201 Created         - POST creation success (return Location header)
  202 Accepted        - Async processing accepted
  204 No Content      - DELETE success or update with no return

3xx - Redirection
  301 Moved Permanently  - Resource permanently moved
  304 Not Modified       - Cache valid

4xx - Client errors
  400 Bad Request           - Malformed request
  401 Unauthorized          - Not authenticated (missing/invalid credentials)
  403 Forbidden             - Authenticated but no permission
  404 Not Found             - Resource not found
  405 Method Not Allowed    - HTTP method not supported
  409 Conflict              - Resource conflict (e.g., unique constraint)
  422 Unprocessable Entity  - Validation failed
  429 Too Many Requests     - Rate limit exceeded

5xx - Server errors
  500 Internal Server Error - Server internal error
  502 Bad Gateway           - Upstream service error
  503 Service Unavailable   - Service temporarily unavailable
  504 Gateway Timeout       - Upstream service timeout
```

### 4. Error Responses (RFC 9457)

Use [RFC 9457 Problem Details](https://www.rfc-editor.org/rfc/rfc9457.html) standard format.

```json
// ✅ RFC 9457 standard error response
{
  "type": "https://api.example.com/errors/validation-error",
  "title": "Validation Failed",
  "status": 422,
  "detail": "Email format is invalid",
  "instance": "/users/123",
  "invalid_fields": [
    {
      "field": "email",
      "reason": "must be a valid email address"
    }
  ]
}

// ✅ Simplified version (RFC 9457 compatible)
{
  "type": "/errors/not-found",
  "title": "User Not Found",
  "status": 404,
  "detail": "User with ID 999 does not exist"
}
```

**Fields**:
- `type`: Error type URI (resolvable documentation)
- `title`: Human-readable error summary
- `status`: HTTP status code
- `detail`: Specific error description
- `instance`: API endpoint where error occurred

### 5. Success Response Format

```json
// ✅ Single resource
{
  "id": 1,
  "name": "Alice",
  "email": "alice@example.com",
  "created_at": "2026-10-15T10:30:00Z"
}

// ✅ Collection with cursor pagination
{
  "data": [
    {"id": 1, "name": "Alice"},
    {"id": 2, "name": "Bob"}
  ],
  "pagination": {
    "cursor": "eyJpZCI6MiwibmFtZSI6IkJvYiJ9",
    "limit": 20,
    "has_more": true,
    "next_url": "/users?cursor=eyJ...&limit=20"
  }
}

// ✅ Offset pagination (simple scenarios)
{
  "data": [...],
  "pagination": {
    "page": 1,
    "page_size": 20,
    "total": 100,
    "total_pages": 5
  }
}
```

### 6. Pagination Strategies

**Cursor pagination (recommended for large datasets)**:
```
# ✅ Initial request
GET /users?limit=20

# ✅ Subsequent requests
GET /users?cursor=eyJpZCI6MjB9&limit=20

Advantages:
- Stable performance (unaffected by data growth)
- Avoids data duplication/skipping
- Suitable for real-time data
```

**Offset pagination (suitable for small datasets)**:
```
# ✅ Simple pagination
GET /users?page=1&page_size=20

Disadvantages:
- Poor performance with large offsets
- Data insertion/deletion causes duplication
- Recommended limit: < 10,000 records
```

### 7. Query Parameters

```
# ✅ Filtering
GET /users?status=active&age_gte=18

# ✅ Sorting (multiple fields)
GET /users?sort=created_at:desc,name:asc

# ✅ Field selection (reduce response size)
GET /users?fields=id,name,email

# ✅ Search
GET /users?q=alice

# ✅ Embed related resources
GET /users?expand=posts,profile
```

### 8. Versioning

```
# ✅ URL path versioning (recommended)
GET /v1/users
GET /v2/users

# ✅ Custom header versioning
GET /users
Header: API-Version: 2024-10-15

# ❌ Query parameter (not recommended)
GET /users?version=2
```

**Versioning strategy**:
- Major version change: breaking changes (`/v1` → `/v2`)
- Minor version change: backward compatible (date-based)
- Maintain N-1 versions simultaneously
- Announce deprecation in advance (Deprecation header)

### 9. Rate Limiting

```
# ✅ Response headers
HTTP/1.1 200 OK
X-RateLimit-Limit: 1000
X-RateLimit-Remaining: 999
X-RateLimit-Reset: 1697371200

# ✅ Rate limit exceeded response
HTTP/1.1 429 Too Many Requests
Retry-After: 60

{
  "type": "/errors/rate-limit-exceeded",
  "title": "Rate Limit Exceeded",
  "status": 429,
  "detail": "API rate limit of 1000 requests per hour exceeded"
}
```

## Common Mistakes

### Mistake 1: Non-RESTful Design

```
# ❌ RPC style
POST /getUser
POST /createUser

# ✅ REST style
GET  /users/{id}
POST /users
```

### Mistake 2: Non-standard Error Responses

```json
// ❌ Custom format
{"error": "user not found", "code": 404}

// ✅ RFC 9457 standard
{
  "type": "/errors/not-found",
  "title": "User Not Found",
  "status": 404,
  "detail": "User with ID 999 does not exist"
}
```

### Mistake 3: Pagination Performance Issues

```
# ❌ Large offset (slow)
GET /users?offset=100000&limit=20

# ✅ Cursor pagination (fast)
GET /users?cursor=eyJpZCI6MTAwMDAwfQ&limit=20
```

---

**References**:
- [Microsoft Azure API Design Best Practices](https://learn.microsoft.com/azure/architecture/microservices/design/api-design)
- [Google Cloud API Design Guide](https://cloud.google.com/apis/design)
- [RFC 9457 - Problem Details for HTTP APIs](https://www.rfc-editor.org/rfc/rfc9457.html)
- [REST API Design Complete Guide (2026)](https://www.digitalapplied.com/blog/rest-api-design-2026-engineering-reference-best-practices)
