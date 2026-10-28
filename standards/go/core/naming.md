---
language: go
category: core
applies_to: []
---

# Naming

Based on [Google Go Style Guide - Naming](https://google.github.io/styleguide/go/decisions#naming).

## Quick Reference

- **Package**: `lowercase`, multi-word stays unbroken (e.g., `tabwriter`)
- **Exported**: `MixedCaps` (PascalCase)
- **Unexported**: `mixedCaps` (camelCase)
- **Acronyms**: Consistent case (`HTTP` or `http`, not `Http`)
- **Getters**: Omit `Get` prefix (use `Owner()` not `GetOwner()`)
- **Constants**: `MixedCaps` (not `SCREAMING_SNAKE_CASE`)

## Common Patterns

| Type | ✅ Good | ❌ Bad |
|------|---------|--------|
| Package | `user`, `oauth2` | `user_service`, `userService` |
| Interface | `Reader`, `FileStore` | `IReader`, `ReaderInterface` |
| Struct | `HTTPClient` | `HttpClient`, `HTTP_Client` |
| Method | `Owner()`, `SetOwner()` | `GetOwner()`, `get_owner()` |
| Constant | `MaxRetries` | `MAX_RETRIES`, `kMaxRetries` |
| Receiver | `func (c *Client)` | `func (this *Client)`, `func (_ *Client)` |

## Core Rules

### 1. Package Name as Context

```go
// ✅ Good: package provides context
package widget
func New() *Widget  // widget.New()

// ❌ Bad: stuttering
package widget
func NewWidget() *Widget  // widget.NewWidget() redundant
```

### 2. Interface Naming

```go
// ✅ Single-method: verb + er
type Reader interface {
    Read(p []byte) (n int, err error)
}

// ✅ Multi-method: describe responsibility
type FileStore interface {
    Open(name string) (File, error)
    Create(name string) (File, error)
}

// ❌ Bad: I prefix
type IReader interface { ... }
```

### 3. Constants Use MixedCaps

```go
// ✅ Good
const MaxPacketSize = 512
const DefaultTimeout = 30 * time.Second

// ❌ Bad
const MAX_PACKET_SIZE = 512      // Go doesn't use ALL_CAPS
const kMaxBufferSize = 1024      // Don't use k prefix
```

### 4. Acronym Rules

```go
// ✅ Acronyms use consistent case
type HTTPSConnection  // HTTPS all caps
type urlParser        // url all lowercase
var userID int        // ID all caps

// ❌ Bad: mixed case
type HttpsConnection  // Wrong: Https
type UrlParser        // Wrong: Url (should be urlParser if unexported)
```

### 5. Variable Name Length ∝ Scope

```go
// ✅ Small scope (1-7 lines) → short name
for i := 0; i < 10; i++ {
    sum += i  // i is clear
}

// ✅ Large scope (25+ lines) → long descriptive name
type ConfigurationManager struct {
    httpClient *http.Client
    maxRetries int
}

// ❌ Bad: small scope but long name
for index := range items {  // i is enough
```

### 6. Receiver Names

```go
// ✅ 1-2 letters, type abbreviation, stay consistent
func (c *Client) Connect() { ... }
func (c *Client) Close() { ... }

func (ri *ResearchInfo) Process() { ... }

// ❌ Bad
func (this *Client) Connect() { ... }  // Don't use this/self
func (c *Client) Close() { ... }
func (client *Client) Send() { ... }   // Inconsistent
```

### 7. Omit Type and Context

```go
// ✅ Type is obvious
var count int           // Not numUsers, usersInt
var users []User        // Not userSlice

// ✅ Context is obvious
package report
type Data struct { ... }  // Not ReportData (package provides context)

// ✅ Local variables omit context
id := req.GetUserID()     // Not userID (context is clear)
```

## Special Cases

### Test Functions

```go
// ✅ Good
func TestUserCreate(t *testing.T) { ... }
func TestHTTPClientTimeout(t *testing.T) { ... }

// ❌ Bad
func Test_User_Create(t *testing.T) { ... }  // No underscores
func testUserCreate(t *testing.T) { ... }    // Must be exported
```

### Flag Naming

```go
// ✅ Flags use snake_case, variables use camelCase
var pollInterval = flag.Duration("poll_interval", 10*time.Second, "")
```

### Must Functions

```go
// ✅ Functions that panic on failure use Must prefix
func MustParse(s string) *Config {
    c, err := Parse(s)
    if err != nil {
        panic(err)
    }
    return c
}
```

---

**Full details**: [Google Go Style Guide - Naming](https://google.github.io/styleguide/go/decisions#naming)
