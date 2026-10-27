# Security Principles

Based on [OWASP API Security Top 10](https://owasp.org/API-Security/editions/2023/en/0x11-t10/), [OWASP Password Storage](https://cheatsheetseries.owasp.org/cheatsheets/Password_Storage_Cheat_Sheet.html), and [OAuth 2.1 Best Practices](https://daily.dev/blog/dev-guide-api-security-oauth-2-1-jwt-vulnerabilities).

## Quick Reference

- **HTTPS mandatory** - all APIs must use HTTPS
- **Authentication required** - default deny, exceptions need review
- **Strong password hashing** - use modern algorithms (Argon2id/bcrypt)
- **Short token lifetime** - limit session duration
- **Timeout configuration** - prevent resource exhaustion
- **Input validation** - validate all user input

## Core Principles

### 1. HTTPS Enforcement

**Principle**: All API traffic must use HTTPS.

```
# ✅ HTTPS only
https://api.example.com/users

# ❌ No HTTP
http://api.example.com/users
```

**Requirements**:
- Force HTTPS redirect
- Use HSTS header
- Valid TLS certificates
- TLS 1.2+ only

### 2. Authentication & Authorization

**Default authentication required**:
- ✅ Deny unauthenticated requests by default
- ✅ Explicitly mark public endpoints
- ✅ Check resource ownership (prevent BOLA)

**Authorization checks**:
```
# ❌ Only check authentication
GET /users/123/orders
# User can access other users' orders

# ✅ Check resource ownership
GET /users/123/orders
if current_user.id != 123:
    return 403 Forbidden
```

### 3. Token-Based Authentication

**Principles**:
- ✅ Use asymmetric signing algorithms (RS256/ES256)
- ❌ Never use `none` algorithm
- ✅ Short access token lifetime (minutes, not hours)
- ✅ Long refresh token lifetime (days/weeks)
- ✅ Store refresh tokens securely (HttpOnly cookies)
- ✅ Validate issuer, audience, expiration
- ✅ Use unique token IDs to prevent replay attacks

**Implementation details** - see language-specific standards.

### 4. Password Hashing

Based on [OWASP Password Storage](https://cheatsheetseries.owasp.org/cheatsheets/Password_Storage_Cheat_Sheet.html).

**Principles**:
- ✅ Use modern key derivation functions (Argon2id, bcrypt, scrypt)
- ❌ Never use fast hashes (MD5, SHA1, SHA256 alone)
- ✅ Use random salt (unique per password)
- ✅ Target 250-500ms per hash operation
- ✅ Configure work factor to resist brute-force
- ✅ Upgrade old hashes on user login

**Algorithm selection**:
- **Argon2id** - OWASP recommended (memory-hard, resistant to GPU/ASIC)
- **bcrypt** - Widely supported, battle-tested
- **scrypt** - Alternative to Argon2

**Implementation details** - see language-specific standards.

### 5. Timeout Configuration

**Principles**:
- ✅ Set timeouts for all network operations
- ✅ Connection timeout (establish connection)
- ✅ Read timeout (receive response)
- ✅ Write timeout (send request)
- ✅ Overall request timeout (max duration)
- ✅ Database query timeout
- ✅ Connection pool acquisition timeout

**Guidelines**:
- Connection: short (seconds)
- Read/Write: moderate (tens of seconds)
- Overall request: based on business requirements
- Database query: prevent long-running queries

**Specific timeout values** - see language-specific standards.

### 6. Rate Limiting

**Principles**:
- ✅ Implement rate limiting at multiple levels
- ✅ IP-based limits (prevent DDoS)
- ✅ User-based limits (prevent abuse)
- ✅ Endpoint-specific limits (sensitive operations)
- ✅ Return `429 Too Many Requests` with `Retry-After` header

**Sensitive operations need stricter limits**:
- Login attempts
- Password reset requests
- Account creation
- Payment transactions

**Specific rate limits** - see language-specific standards.

### 7. Input Validation

**Principles**:
- ✅ Whitelist validation (allow known-good)
- ✅ Length limits on all inputs
- ✅ Type validation
- ✅ SQL injection prevention (parameterized queries)
- ✅ XSS prevention (escape output, CSP headers)

### 8. Sensitive Data Protection

**Principles**:
- ✅ Never log passwords, tokens, or secrets
- ✅ Mask sensitive data in logs and UI
- ✅ Encrypt sensitive fields at rest
- ✅ Use key management services (KMS/Vault)

### 9. CORS Configuration

**Principles**:
- ✅ Specify exact allowed origins (not `*`)
- ✅ Limit allowed methods
- ✅ Limit allowed headers
- ✅ Set appropriate cache duration

### 10. Security Headers

**Required headers**:
- `Strict-Transport-Security` - force HTTPS
- `X-Content-Type-Options: nosniff` - prevent MIME sniffing
- `X-Frame-Options: DENY` - prevent clickjacking
- `Content-Security-Policy` - prevent XSS
- `X-XSS-Protection` - enable XSS filter

## OWASP API Security Top 10 (2023)

Based on [OWASP API Security Top 10](https://offensive360.com/blog/api-security-best-practices-guide/).

```
1. Broken Object Level Authorization (BOLA)
   Principle: Always check resource ownership

2. Broken Authentication
   Principle: Use OAuth 2.1 + strong tokens

3. Broken Object Property Level Authorization
   Principle: Validate modifiable fields

4. Unrestricted Resource Consumption
   Principle: Rate limiting + timeouts

5. Broken Function Level Authorization
   Principle: Check user permissions for each operation

6. Unrestricted Access to Sensitive Business Flows
   Principle: Protect business logic flows

7. Server Side Request Forgery (SSRF)
   Principle: Validate URLs, whitelist domains

8. Security Misconfiguration
   Principle: Default deny, least privilege

9. Improper Inventory Management
   Principle: API versioning, deprecation notices

10. Unsafe Consumption of APIs
    Principle: Validate third-party API responses
```

## Common Mistakes

### Mistake 1: Weak Password Hashing

```
# ❌ Fast hashes (unsafe)
hash = md5(password)
hash = sha256(password)

# ✅ Key derivation functions
hash = argon2id(password, salt)
hash = bcrypt(password, salt, cost_factor)
```

### Mistake 2: Token Algorithm Issues

```
# ❌ Allow none algorithm
# ❌ Use symmetric algorithm with shared secret
# ✅ Use asymmetric algorithm (RS256/ES256)
# ✅ Validate algorithm in token header
```

### Mistake 3: Not Checking Resource Ownership

```
# ❌ BOLA vulnerability
GET /orders/123
# Only checks authentication, not ownership

# ✅ Check ownership
GET /orders/123
if order.user_id != current_user.id:
    return 403 Forbidden
```

---

**References**:
- [OWASP API Security Top 10](https://owasp.org/API-Security/editions/2023/en/0x11-t10/)
- [OWASP Password Storage Cheat Sheet](https://cheatsheetseries.owasp.org/cheatsheets/Password_Storage_Cheat_Sheet.html)
- [OAuth 2.1 and JWT Best Practices](https://daily.dev/blog/dev-guide-api-security-oauth-2-1-jwt-vulnerabilities)
