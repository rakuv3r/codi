# Performance

Based on Go performance optimization best practices.

## Quick Reference

- **Correctness first** - premature optimization is the root of all evil
- **Measure then optimize** - use benchmarks to find bottlenecks
- **Pre-allocate capacity** - slice, map allocation upfront
- **strings.Builder** - for string concatenation
- **sync.Pool** - reuse objects
- **atomic** - avoid lock contention

## Core Principles

1. **Correctness first** - readable, maintainable code beats micro-optimizations
2. **Measure then optimize** - use profiling to find real bottlenecks
3. **Optimize hot paths** - focus on the 20% consuming most CPU/memory
4. **Balance readability** - performance gain must be significant to sacrifice readability

## Core Rules

### 1. Pre-allocate Capacity

```go
// ❌ No pre-allocation, multiple expansions
var result []int
for i := 0; i < 1000; i++ {
    result = append(result, i)
}

// ✅ Pre-allocate capacity
result := make([]int, 0, 1000)
for i := 0; i < 1000; i++ {
    result = append(result, i)
}

// ✅ Map pre-allocation
m := make(map[string]Item, len(items))
for _, item := range items {
    m[item.ID] = item
}
```

### 2. String Concatenation

```go
// ❌ Using + for concatenation (multiple allocations)
var result string
for _, s := range strings {
    result += s  // Creates new string each time
}

// ✅ Use strings.Builder
var builder strings.Builder
for _, s := range strings {
    builder.WriteString(s)
}
result := builder.String()

// ✅ Pre-allocate Builder capacity
var builder strings.Builder
builder.Grow(expectedSize)
```

### 3. Avoid Unnecessary Allocations

```go
// ❌ Repeated allocation in loop
for _, item := range items {
    buf := make([]byte, 1024)  // Allocate each iteration
    process(item, buf)
}

// ✅ Allocate outside loop
buf := make([]byte, 1024)
for _, item := range items {
    process(item, buf)
}

// ❌ Small struct with pointer
type Point struct {
    X, Y int
}
func NewPoint(x, y int) *Point {  // Heap allocation
    return &Point{X: x, Y: y}
}

// ✅ Small struct return value
func NewPoint(x, y int) Point {  // Stack allocation
    return Point{X: x, Y: y}
}
```

### 4. sync.Pool for Object Reuse

```go
// ✅ Use sync.Pool to reuse frequently created objects
var bufferPool = sync.Pool{
    New: func() any {
        return new(bytes.Buffer)
    },
}

func ProcessData(data []byte) []byte {
    buf := bufferPool.Get().(*bytes.Buffer)
    defer func() {
        buf.Reset()
        bufferPool.Put(buf)
    }()
    
    buf.Write(data)
    // Process
    return buf.Bytes()
}
```

### 5. Choose Right Data Structure

```go
// ❌ Frequent lookups with slice (O(n))
func Contains(items []string, target string) bool {
    for _, item := range items {
        if item == target {
            return true
        }
    }
    return false
}

// ✅ Frequent lookups with map (O(1))
type Set map[string]struct{}

func (s Set) Contains(key string) bool {
    _, ok := s[key]
    return ok
}

// ✅ Empty struct{} as set value (saves memory)
seen := make(map[string]struct{})
seen["key"] = struct{}{}
```

### 6. Avoid Lock Contention

```go
// ❌ High-frequency operation with Mutex
type Counter struct {
    mu    sync.Mutex
    value int
}

func (c *Counter) Inc() {
    c.mu.Lock()
    c.value++
    c.mu.Unlock()
}

// ✅ High-frequency operation with atomic
type Counter struct {
    value int64
}

func (c *Counter) Inc() {
    atomic.AddInt64(&c.value, 1)
}

// ✅ Read-heavy workload with RWMutex
type Cache struct {
    mu    sync.RWMutex
    items map[string]Item
}

func (c *Cache) Get(key string) (Item, bool) {
    c.mu.RLock()
    defer c.mu.RUnlock()
    return c.items[key], true
}
```

### 7. Limit Concurrency

```go
// ❌ Unlimited goroutine creation
for _, item := range items {
    go process(item)  // May create millions of goroutines
}

// ✅ Use semaphore to limit
semaphore := make(chan struct{}, 100)
var wg sync.WaitGroup

for _, item := range items {
    wg.Add(1)
    semaphore <- struct{}{}
    
    go func(item Item) {
        defer wg.Done()
        defer func() { <-semaphore }()
        process(item)
    }(item)
}

wg.Wait()
```

### 8. Avoid []byte and string Conversions

```go
// ❌ Repeated conversions
func ProcessData(data []byte) error {
    str := string(data)  // Allocation
    if strings.HasPrefix(str, "prefix") {
        // ...
    }
    if strings.Contains(str, "text") {  // Another conversion
        // ...
    }
}

// ✅ Use bytes package
func ProcessData(data []byte) error {
    if bytes.HasPrefix(data, []byte("prefix")) {
        // ...
    }
    if bytes.Contains(data, []byte("text")) {
        // ...
    }
}
```

## Performance Testing

### Benchmark

```go
// ✅ Write benchmarks
func BenchmarkProcess(b *testing.B) {
    data := generateTestData()
    
    b.ResetTimer()  // Reset timer (exclude setup)
    
    for i := 0; i < b.N; i++ {
        process(data)
    }
}

// Run
// go test -bench=. -benchmem
```

### Profiling

```bash
# CPU profile
go test -cpuprofile=cpu.prof -bench=.
go tool pprof cpu.prof

# Memory profile
go test -memprofile=mem.prof -bench=.
go tool pprof mem.prof

# Visualize
go tool pprof -http=:8080 cpu.prof
```

## Common Mistakes

### Mistake 1: Premature Optimization

```go
// ❌ Sacrifice readability for performance (no benchmark proof)
func obscureOptimization() {
    // Hard to understand code
}

// ✅ Write clear code first
func clearCode() {
    // Readability first
}

// Then benchmark to find bottlenecks
```

### Mistake 2: Over Pre-allocation

```go
// ❌ Pre-allocate large capacity without knowing size
result := make([]int, 0, 1000000)  // Actually use only 10

// ✅ Reasonable pre-allocation or dynamic growth
result := make([]int, 0)  // Small data, grow dynamically
```

---

**Tools**:
- [pprof](https://pkg.go.dev/net/http/pprof)
- [benchmark](https://pkg.go.dev/testing#hdr-Benchmarks)
