# Go Generics Standards

## Overview

Generics (type parameters) were added in Go 1.18, enabling type-safe generic code without code duplication or `interface{}`. Use generics for algorithms and data structures that work across types, but don't overuse them.

## When to Use Generics

### ✅ Good Use Cases

```go
// Generic data structures
type Stack[T any] struct {
    items []T
}

// Generic algorithms
func Map[T, U any](slice []T, fn func(T) U) []U

// Type-safe containers
type Cache[K comparable, V any] struct {
    data map[K]V
}

// Functions working with slices/maps/channels of any type
func Filter[T any](slice []T, fn func(T) bool) []T
```

### ❌ Avoid Generics When

```go
// Only 1-2 uses - just duplicate the code
// Methods need different logic - not truly generic
// Interface already sufficient - io.Reader better than Reader[T]
// Performance-critical path - generics have compilation overhead
```

**Rule**: Use generics to eliminate repetitive boilerplate, not for premature abstraction.

## Type Parameter Syntax

### Function

```go
func Min[T cmp.Ordered](a, b T) T {
    if a < b {
        return a
    }
    return b
}

// Multiple type parameters
func Map[T, U any](slice []T, fn func(T) U) []U {
    result := make([]U, len(slice))
    for i, v := range slice {
        result[i] = fn(v)
    }
    return result
}
```

### Type

```go
type Pair[T, U any] struct {
    First  T
    Second U
}

type Node[T any] struct {
    Value T
    Next  *Node[T]
}

// Usage:
var p Pair[int, string]
p.First = 42
p.Second = "hello"
```

### Method

```go
type Stack[T any] struct {
    items []T
}

func (s *Stack[T]) Push(item T) {
    s.items = append(s.items, item)
}

func (s *Stack[T]) Pop() (T, bool) {
    if len(s.items) == 0 {
        var zero T
        return zero, false
    }
    item := s.items[len(s.items)-1]
    s.items = s.items[:len(s.items)-1]
    return item, true
}
```

## Constraints

### Built-in Constraints

**any** - Any type:

```go
func Print[T any](v T) {
    fmt.Println(v)  // ✅ Can assign, pass, print
    // v + 1        // ❌ Cannot do arithmetic
    // v < 0        // ❌ Cannot compare with <
}
```

**comparable** - Types supporting `==` and `!=`:

```go
func Index[T comparable](slice []T, target T) int {
    for i, v := range slice {
        if v == target {  // ✅ Can use == and !=
            return i
        }
        // if v < target  // ❌ Cannot use < or >
    }
    return -1
}

// Required for map keys
type Set[T comparable] map[T]struct{}
```

**cmp.Ordered** (Go 1.21+) - Types supporting ordering:

```go
import "cmp"

func Max[T cmp.Ordered](a, b T) T {
    if a > b {  // ✅ Can use <, >, <=, >=
        return a
    }
    return b
}

// Includes: all integers, floats, strings
```

### Type Unions

```go
type Number interface {
    int | int64 | float64
}

func Sum[T Number](values []T) T {
    var total T
    for _, v := range values {
        total += v  // Works for all numeric types
    }
    return total
}

// Only operations valid for ALL types in union are allowed
type IntOrString interface {
    int | string
}

func Add[T IntOrString](a, b T) T {
    return a + b  // ✅ Both support +
}

func Subtract[T IntOrString](a, b T) T {
    return a - b  // ❌ string doesn't support -
}
```

### Underlying Type Operator `~`

**Without `~`** - Exact type match only:

```go
type exactInt interface { int }

type MyInt int

func process[T exactInt](v T) {}

process(42)          // ✅ int
process(MyInt(42))   // ❌ MyInt is not int
```

**With `~`** - Underlying type match:

```go
type intLike interface { ~int }

func process[T intLike](v T) T {
    return v + 1
}

type MyInt int

process(42)          // ✅
process(MyInt(42))   // ✅ Underlying type is int
```

**Rule**: Use `~` for numeric/string types unless you specifically need exact type.

### Combining Constraints

**Type union + methods**:

```go
type Stringer interface {
    String() string
}

type LabeledNumber interface {
    ~int | ~float64
    Stringer
}

func Print[T LabeledNumber](v T) {
    fmt.Println(v + 1)      // Numeric operation
    fmt.Println(v.String()) // Method call
}
```

**Multiple type parameters**:

```go
func Transform[S ~[]E, E any, R any](slice S, fn func(E) R) []R {
    result := make([]R, len(slice))
    for i, v := range slice {
        result[i] = fn(v)
    }
    return result
}
```

## Type Inference

### Automatic Inference

```go
func Min[T cmp.Ordered](a, b T) T { ... }

// ✅ Compiler infers T = int
result := Min(3, 5)

// ❌ Don't specify explicitly unless necessary
result := Min[int](3, 5)
```

### When Inference Fails

Type parameters only in return value cannot be inferred:

```go
// ❌ Cannot infer T
func MakeSlice[T any](size int) []T {
    return make([]T, size)
}

// Must specify explicitly
s := MakeSlice[int](10)  // ✅
```

### Constraint Inference

```go
func Scale[S ~[]E, E constraints.Integer](s S, c E) S {
    r := make(S, len(s))
    for i, v := range s {
        r[i] = v * c
    }
    return r
}

type MySlice []int

var s MySlice = []int{1, 2, 3}

// Infers: S = MySlice, E = int
// Returns MySlice, not []int
result := Scale(s, 2)
```

## Common Patterns

### Generic Slice Operations

```go
func Map[T, U any](slice []T, fn func(T) U) []U {
    result := make([]U, len(slice))
    for i, v := range slice {
        result[i] = fn(v)
    }
    return result
}

func Filter[T any](slice []T, fn func(T) bool) []T {
    result := []T{}
    for _, v := range slice {
        if fn(v) {
            result = append(result, v)
        }
    }
    return result
}

func Reduce[T, U any](slice []T, init U, fn func(U, T) U) U {
    acc := init
    for _, v := range slice {
        acc = fn(acc, v)
    }
    return acc
}

func Contains[T comparable](slice []T, target T) bool {
    for _, v := range slice {
        if v == target {
            return true
        }
    }
    return false
}
```

### Generic Data Structures

**Stack**:

```go
type Stack[T any] struct {
    items []T
}

func (s *Stack[T]) Push(item T) {
    s.items = append(s.items, item)
}

func (s *Stack[T]) Pop() (T, bool) {
    if len(s.items) == 0 {
        var zero T
        return zero, false
    }
    item := s.items[len(s.items)-1]
    s.items = s.items[:len(s.items)-1]
    return item, true
}

func (s *Stack[T]) IsEmpty() bool {
    return len(s.items) == 0
}
```

**Option Type**:

```go
type Option[T any] struct {
    value T
    valid bool
}

func Some[T any](v T) Option[T] {
    return Option[T]{value: v, valid: true}
}

func None[T any]() Option[T] {
    return Option[T]{valid: false}
}

func (o Option[T]) Unwrap() T {
    if !o.valid {
        panic("unwrap on None")
    }
    return o.value
}

func (o Option[T]) UnwrapOr(defaultValue T) T {
    if !o.valid {
        return defaultValue
    }
    return o.value
}

func (o Option[T]) IsSome() bool {
    return o.valid
}
```

**Result Type**:

```go
type Result[T any] struct {
    value T
    err   error
}

func Ok[T any](v T) Result[T] {
    return Result[T]{value: v}
}

func Err[T any](err error) Result[T] {
    return Result[T]{err: err}
}

func (r Result[T]) Unwrap() (T, error) {
    return r.value, r.err
}

func (r Result[T]) IsOk() bool {
    return r.err == nil
}
```

### Generic Constraints

```go
// Number constraint
type Number interface {
    ~int | ~int8 | ~int16 | ~int32 | ~int64 |
    ~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 |
    ~float32 | ~float64
}

func Sum[T Number](values []T) T {
    var total T
    for _, v := range values {
        total += v
    }
    return total
}

// Signed number constraint
type Signed interface {
    ~int | ~int8 | ~int16 | ~int32 | ~int64 |
    ~float32 | ~float64
}

func Abs[T Signed](n T) T {
    if n < 0 {
        return -n
    }
    return n
}
```

## Standard Library Generics (Go 1.21+)

### cmp Package

```go
import "cmp"

// Ordered constraint
func Max[T cmp.Ordered](a, b T) T

// Compare function
result := cmp.Compare(5, 3)  // Returns 1
result := cmp.Compare(3, 5)  // Returns -1
result := cmp.Compare(3, 3)  // Returns 0

// Or function (first non-zero value)
result := cmp.Or(0, "", 42)  // Returns 42
```

### slices Package

```go
import "slices"

// Sort
slices.Sort(numbers)
slices.SortFunc(users, func(a, b User) int {
    return cmp.Compare(a.Age, b.Age)
})

// Search
index, found := slices.BinarySearch(sorted, target)

// Contains
if slices.Contains(slice, value) { }

// Min/Max
min := slices.Min(numbers)
max := slices.Max(numbers)

// Clone
copy := slices.Clone(original)
```

### maps Package

```go
import "maps"

// Clone
copy := maps.Clone(original)

// Equal
if maps.Equal(map1, map2) { }

// Keys/Values
keys := maps.Keys(m)
values := maps.Values(m)
```

## Performance Considerations

### Compile-Time Monomorphization

```go
func Add[T int | float64](a, b T) T {
    return a + b
}

// Compiler generates separate functions:
// func Add_int(a, b int) int
// func Add_float64(a, b float64) float64
```

- Each type instantiation creates separate code
- No runtime reflection overhead
- Binary size increases with type variations

### vs interface{}

```go
// Generics: type-safe, no boxing
func Max[T cmp.Ordered](a, b T) T

// interface{}: runtime type assertion, boxing overhead
func Max(a, b interface{}) interface{}
```

**Rule**: Generics for type safety and performance, interfaces for dynamic dispatch.

## Anti-Patterns

### ❌ Overusing Generics

```go
// Wrong: unnecessary generic
func PrintInt[T int](v T) {
    fmt.Println(v)
}

// Right: just use int
func PrintInt(v int) {
    fmt.Println(v)
}
```

### ❌ Interface Would Be Better

```go
// Wrong: generic for behavior
type Reader[T any] interface {
    Read(T) error
}

// Right: use interface{}
type Reader interface {
    Read([]byte) (int, error)
}
```

### ❌ Forgetting `~` for Named Types

```go
// Wrong: rejects type Celsius float64
type floatExact interface { float64 }

// Right: accepts underlying float64
type floatLike interface { ~float64 }
```

### ❌ Using comparable for Ordering

```go
// Wrong: comparable only supports ==, !=
func Max[T comparable](a, b T) T {
    if a > b {  // ❌ Compile error
        return a
    }
    return b
}

// Right: use cmp.Ordered for ordering
func Max[T cmp.Ordered](a, b T) T {
    if a > b {
        return a
    }
    return b
}
```

## Best Practices

1. **Minimize constraints** - Only require operations actually used
2. **Prefer inference** - Let compiler infer types when possible
3. **Use `~` for named types** - Unless you need exact type match
4. **cmp.Ordered vs comparable** - Ordered for `</>`, comparable for `==`
5. **Interface over generics** - When behavior polymorphism is needed
6. **Type parameters in parameters** - For inference to work

## References

- Go 1.18 release: https://go.dev/blog/intro-generics
- Type parameters proposal: https://go.dev/design/43651-type-parameters
- Standard library: https://pkg.go.dev/cmp, https://pkg.go.dev/slices, https://pkg.go.dev/maps
