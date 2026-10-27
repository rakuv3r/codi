# Imports

Based on [Google Go Style Guide - Imports](https://google.github.io/styleguide/go/decisions#imports).

## Quick Reference

- **Use goimports** - automatically manage imports
- **Group order** - standard library, third-party, local packages, side-effect imports
- **Alphabetical within groups** - automatic sorting
- **Avoid dot imports** - pollutes namespace
- **Meaningful aliases** - use when conflicts or too long

## Core Rules

### 1. Import Grouping

```go
// ✅ Standard grouping (4 groups, separated by blank lines)
import (
    // Standard library
    "context"
    "fmt"
    "time"
    
    // Third-party libraries
    "github.com/gin-gonic/gin"
    "github.com/lib/pq"
    
    // Local packages
    "myproject/internal/config"
    "myproject/internal/domain"
    
    // Side-effect imports (last, with comments)
    _ "github.com/lib/pq"  // PostgreSQL driver
)

// ❌ No grouping
import (
    "fmt"
    "github.com/gin-gonic/gin"
    "context"
    "myproject/internal/config"
)
```

### 2. Alphabetical Within Groups

```go
// ✅ Auto-sorted (goimports)
import (
    "context"
    "errors"
    "fmt"
    "log"
    "time"
)

// ❌ Random order
import (
    "time"
    "context"
    "log"
)
```

### 3. Import Aliases

```go
// ✅ Use aliases for conflicts
import (
    "crypto/rand"
    mathrand "math/rand"
)

// ✅ Simplify long package names
import (
    promclient "github.com/prometheus/client_golang/prometheus"
)

// ✅ Versioned packages
import (
    corev1 "k8s.io/api/core/v1"
    metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ✅ Protobuf packages: remove underscores + pb suffix
import (
    foosvcpb "path/to/package/foo_service_go_proto"
)

// ❌ Meaningless aliases
import (
    pkg1 "github.com/user/package"
    f "fmt"  // Hard to search
)
```

### 4. Alias Naming Convention

```go
// ✅ Descriptive aliases
import (
    userdb "github.com/user/database"
    configdb "github.com/config/database"
)

// ✅ Common variable name collision: pkg suffix
import (
    "net/url"
    
    urlpkg "some/other/url"
)

// ❌ Don't arbitrarily rename standard library
import (
    f "fmt"
    c "context"
)
```

### 5. Side-Effect Imports (Blank Imports)

```go
// ✅ For driver registration, init initialization
import (
    "database/sql"
    
    "github.com/gin-gonic/gin"
    
    "myproject/internal/config"
    
    _ "github.com/lib/pq"           // PostgreSQL driver
    _ "go.uber.org/automaxprocs"    // Auto GOMAXPROCS
)

// ⚠️ Use only in main packages or tests
// Avoid side-effect imports in library packages
```

### 6. Dot Imports

```go
// ❌ Don't use dot imports (pollutes namespace)
import (
    . "fmt"
    . "strings"
)

func Example() {
    Println("Hello")  // Unclear origin
}

// ✅ Use normal imports
import (
    "fmt"
    "strings"
)

func Example() {
    fmt.Println("Hello")
    strings.ToUpper("test")
}

// ✅ Exception: test packages can use dot imports
package mypackage_test

import (
    . "mypackage"
    "testing"
)
```

### 7. Avoid Circular Dependencies

```go
// ❌ Circular dependency
package user
import "myproject/order"  // user → order

package order
import "myproject/user"   // order → user → compile error

// ✅ Solution 1: Extract common package
package domain
type User struct { ... }
type Order struct { ... }

package user
import "myproject/domain"

package order
import "myproject/domain"

// ✅ Solution 2: Use interface to decouple
package order
type UserGetter interface {
    GetUser(id string) (*User, error)
}

package user
type Service struct{}
func (s *Service) GetUser(id string) (*User, error) { ... }

// Inject in main
orderService := order.NewService(userService)
```

## Production Patterns

### Configure goimports

```bash
# Specify local package prefix
goimports -local myproject -w .

# Makefile
.PHONY: fmt
fmt:
	goimports -local myproject -w .
	gofmt -s -w .
```

### Project Example

```go
// ✅ Standard grouping
package storage

import (
    "context"
    "fmt"
    "strings"
    
    "log/slog"
    
    "github.com/lib/pq"
    
    "myproject/internal/config"
    "myproject/internal/domain"
)
```

## Common Mistakes

### Mistake 1: Unused Imports

```go
// ❌ Imported but unused
import (
    "fmt"
    "time"  // Compile error
)

// ✅ goimports auto-removes
```

### Mistake 2: Unnecessary Aliases

```go
// ❌ Too many aliases
import (
    f "fmt"
    s "strings"
    t "time"
)

// ✅ Only when necessary
import (
    "fmt"
    "strings"
    "time"
)
```

### Mistake 3: Mixed Grouping

```go
// ❌ Standard library and third-party mixed
import (
    "fmt"
    "github.com/gin-gonic/gin"
    "context"
)

// ✅ Correct grouping
import (
    "context"
    "fmt"
    
    "github.com/gin-gonic/gin"
)
```

---

**Tool**: [goimports](https://pkg.go.dev/golang.org/x/tools/cmd/goimports)
