---
language: go
category: core
applies_to: []
---

# Formatting

Based on [Google Go Style Guide - Formatting](https://google.github.io/styleguide/go/guide#formatting).

## Quick Reference

- **Use gofmt** - all code must be formatted
- **Use goimports** - automatically organize imports
- **No strict line length** - prefer refactoring over splitting
- **Tab indentation** - not spaces
- **Format on save** - configure editor

## Core Rules

### 1. Mandatory gofmt

```bash
# Format single file
gofmt -w file.go

# Format entire project
gofmt -s -w .

# gofmt -s simplifies code
# Before: var x []int = []int{1, 2, 3}
# After:  var x = []int{1, 2, 3}
```

### 2. Use goimports

```bash
# Organize imports
goimports -w .

# Specify local package prefix
goimports -local myproject -w .
```

**Automatic features**:
- Add missing imports
- Remove unused imports
- Standard grouping

### 3. Editor Configuration

**VS Code**:
```json
{
  "[go]": {
    "editor.formatOnSave": true,
    "editor.codeActionsOnSave": {
      "source.organizeImports": true
    }
  },
  "go.formatTool": "goimports"
}
```

### 4. Line Length

**No strict limit, but recommended**:
- 80-100 characters: ideal
- 120 characters: maximum
- When too long: prefer refactoring, don't force splitting

```go
// ✅ Appropriate
func ProcessData(data *Data) error {
    return validate(data)
}

// ❌ Too long, wrap
func CreateConnection(
    ctx context.Context,
    sourcePod, sourceNamespace string,
    destPod, destNamespace string,
    protocol Protocol,
    port int,
) error {
```

### 5. Brace Position

```go
// ✅ Opening brace on same line
func Example() {
    if condition {
        doSomething()
    }
}

// ❌ C-style (Go doesn't allow)
func Example()
{  // Compile error
```

### 6. Struct Initialization

```go
// ✅ Short struct single line
user := User{Name: "Alice", Age: 30}

// ✅ Long struct multi-line, trailing comma
user := User{
    Name:      "Alice",
    Age:       30,
    Email:     "alice@example.com",
    CreatedAt: time.Now(),  // ← comma
}
```

### 7. Code Grouping

```go
// ✅ Group related declarations
const (
    StatusPending = "pending"
    StatusActive  = "active"
)

const (
    MaxRetries     = 3
    DefaultTimeout = 30 * time.Second
)

// ✅ Group struct fields
type Server struct {
    // Connection config
    host string
    port int
    
    // Timeout config
    readTimeout  time.Duration
    writeTimeout time.Duration
    
    // Dependencies
    db     *sql.DB
    logger *slog.Logger
}
```

### 8. Blank Lines

```go
// ✅ Add blank lines between logical blocks
func ProcessUser(user *User) error {
    // Validate
    if err := validateUser(user); err != nil {
        return err
    }
    
    // Transform
    data := transformUser(user)
    
    // Save
    if err := saveUser(data); err != nil {
        return err
    }
    
    return nil
}

// ❌ Too many blank lines
func ProcessUser(user *User) error {


    if err := validateUser(user); err != nil {
        return err
    }


    data := transformUser(user)
}

// ❌ Missing blank lines (hard to read)
func ProcessUser(user *User) error {
    if err := validateUser(user); err != nil {
        return err
    }
    data := transformUser(user)
    if err := saveUser(data); err != nil {
        return err
    }
    return nil
}
```

### 9. Tab Indentation

```go
// ✅ Use tabs (gofmt automatic)
func Example() {
	if true {
		doSomething()
	}
}

// ❌ Don't use spaces manually
func Example() {
    if true {  // Spaces, not tabs
        doSomething()
    }
}
```

### 10. Alignment

```go
// ✅ gofmt auto-aligns
type Config struct {
	Host     string
	Port     int
	Username string
	Password string
}

var (
	name    = "Alice"
	age     = 30
	country = "USA"
)

// ❌ Don't over-align
const (
	StatusActive              = "active"
	StatusPending             = "pending"
	StatusVeryLongStatusName  = "very_long"
)
```

## Common Mistakes

### Mistake 1: Not Using gofmt

```go
// ❌ Manual formatting
func Example( ) {
if x==1{
return true
}
}

// ✅ After gofmt
func Example() {
	if x == 1 {
		return true
	}
	return false
}
```

### Mistake 2: Too Long Without Wrapping

```go
// ❌ One line too long
func CreateConnection(ctx context.Context, sourcePod, sourceNamespace, destPod, destNamespace string, protocol Protocol, port int) error {

// ✅ Wrap
func CreateConnection(
    ctx context.Context,
    sourcePod, sourceNamespace string,
    destPod, destNamespace string,
    protocol Protocol,
    port int,
) error {
```

### Mistake 3: Unorganized Imports

```go
// ❌ Messy
import (
"fmt"
  "time"
    "context"
)

// ✅ After goimports
import (
	"context"
	"fmt"
	"time"
)
```

---

**Tools**:
- [gofmt](https://pkg.go.dev/cmd/gofmt)
- [goimports](https://pkg.go.dev/golang.org/x/tools/cmd/goimports)
