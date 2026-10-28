---
language: go
category: core
applies_to: []
---

# Concurrency

Based on [Google Go Best Practices - Concurrency](https://google.github.io/styleguide/go/best-practices#concurrency).

## Quick Reference

- **Share memory by communicating** - prefer channels over shared variables
- **Know goroutine lifecycle** - when it starts and exits
- **Avoid leaks** - ensure all goroutines can exit
- **Use Context for cancellation** - graceful shutdown
- **Limit concurrency** - use worker pool or semaphore
- **Don't call t.Fatal from goroutines** - only from main test goroutine

## Core Rules

### 1. Goroutine Lifecycle

```go
// ✅ Clear exit condition
func StartWorker(ctx context.Context) {
    go func() {
        for {
            select {
            case <-ctx.Done():
                return  // Exit on context cancel
            case work := <-workQueue:
                processWork(work)
            }
        }
    }()
}

// ❌ Never exits (goroutine leak)
func StartWorker() {
    go func() {
        for {
            work := <-workQueue
            processWork(work)
        }
    }()  // No way to stop
}
```

### 2. WaitGroup for Completion

```go
// ✅ Wait for all goroutines
func ProcessBatch(items []Item) error {
    var wg sync.WaitGroup
    errCh := make(chan error, len(items))
    
    for _, item := range items {
        wg.Add(1)
        go func(item Item) {  // Pass as parameter to avoid closure issues
            defer wg.Done()
            if err := process(item); err != nil {
                errCh <- err
            }
        }(item)
    }
    
    wg.Wait()
    close(errCh)
    
    for err := range errCh {
        if err != nil {
            return err
        }
    }
    return nil
}

// ❌ Closure captures loop variable
for _, item := range items {
    go func() {
        process(item)  // Wrong: all goroutines see last item
    }()
}
```

### 3. Limit Concurrency

```go
// ✅ Use semaphore
func ProcessWithLimit(items []Item, maxConcurrency int) {
    sem := make(chan struct{}, maxConcurrency)
    var wg sync.WaitGroup
    
    for _, item := range items {
        wg.Add(1)
        go func(item Item) {
            defer wg.Done()
            
            sem <- struct{}{}        // Acquire
            defer func() { <-sem }() // Release
            
            process(item)
        }(item)
    }
    
    wg.Wait()
}

// ✅ Use worker pool
func ProcessWithPool(items []Item, numWorkers int) {
    jobs := make(chan Item, len(items))
    var wg sync.WaitGroup
    
    for i := 0; i < numWorkers; i++ {
        wg.Add(1)
        go func() {
            defer wg.Done()
            for item := range jobs {
                process(item)
            }
        }()
    }
    
    for _, item := range items {
        jobs <- item
    }
    close(jobs)
    wg.Wait()
}
```

### 4. Channel Direction

```go
// ✅ Specify channel direction
func Producer(out chan<- int) {  // Send-only
    defer close(out)
    for i := 0; i < 10; i++ {
        out <- i
    }
}

func Consumer(in <-chan int) {  // Receive-only
    for value := range in {
        process(value)
    }
}

// ✅ Compiler checks
func Consumer(in <-chan int) {
    close(in)  // Compile error: cannot close receive channel
}
```

### 5. Select Multiplexing

```go
// ✅ Standard pattern
func Worker(ctx context.Context, jobs <-chan Job, results chan<- Result) {
    for {
        select {
        case <-ctx.Done():
            return  // Context cancelled
            
        case job, ok := <-jobs:
            if !ok {
                return  // jobs closed
            }
            result := processJob(job)
            
            // Nested select to avoid blocking
            select {
            case results <- result:
            case <-ctx.Done():
                return
            }
        }
    }
}

// ✅ Non-blocking check
select {
case msg := <-messages:
    handleMessage(msg)
default:
    // Channel empty, continue other work
}
```

### 6. Mutex for Shared State

```go
// ✅ Basic usage
type Counter struct {
    mu    sync.Mutex
    value int
}

func (c *Counter) Inc() {
    c.mu.Lock()
    defer c.mu.Unlock()
    c.value++
}

// ✅ RWMutex (read-heavy workload)
type Cache struct {
    mu    sync.RWMutex
    items map[string]Item
}

func (c *Cache) Get(key string) (Item, bool) {
    c.mu.RLock()
    defer c.mu.RUnlock()
    return c.items[key], true
}

func (c *Cache) Set(key string, item Item) {
    c.mu.Lock()
    defer c.mu.Unlock()
    c.items[key] = item
}

// ❌ Don't copy Mutex
func (c Counter) Inc() {  // Value receiver, copies Mutex
    c.mu.Lock()  // Wrong
}
```

### 7. Concurrency in Tests

```go
// ❌ Wrong: call t.Fatal from goroutine
func TestConcurrent(t *testing.T) {
    go func() {
        if err := doWork(); err != nil {
            t.Fatal(err)  // Wrong: not in main test goroutine
        }
    }()
}

// ✅ Correct: use t.Error
func TestConcurrent(t *testing.T) {
    var wg sync.WaitGroup
    wg.Add(1)
    
    go func() {
        defer wg.Done()
        if err := doWork(); err != nil {
            t.Error(err)  // Correct
        }
    }()
    
    wg.Wait()
}

// ✅ Or use errgroup
func TestConcurrent(t *testing.T) {
    g, ctx := errgroup.WithContext(context.Background())
    
    g.Go(func() error {
        return doWork()
    })
    
    if err := g.Wait(); err != nil {
        t.Fatal(err)  // Main goroutine, can use Fatal
    }
}
```

## Production Patterns

### Batch Sender Pattern

```go
// ✅ Batch sending + periodic flush
func (a *Agent) batchSender(ctx context.Context, connCh <-chan *Connection) error {
    ticker := time.NewTicker(a.flushInterval)
    defer ticker.Stop()
    
    batch := make([]*Connection, 0, a.batchSize)
    
    for {
        select {
        case <-ctx.Done():
            // Send remaining data
            if len(batch) > 0 {
                a.sendBatch(context.Background(), batch)
            }
            return ctx.Err()
            
        case conn := <-connCh:
            batch = append(batch, conn)
            if len(batch) >= a.batchSize {
                if err := a.sendBatch(ctx, batch); err != nil {
                    return err
                }
                batch = batch[:0]
            }
            
        case <-ticker.C:
            if len(batch) > 0 {
                if err := a.sendBatch(ctx, batch); err != nil {
                    return err
                }
                batch = batch[:0]
            }
        }
    }
}
```

### Worker Pool

```go
// ✅ Complete worker pool
type WorkerPool struct {
    workers int
    jobs    chan Job
    results chan Result
    wg      sync.WaitGroup
}

func NewWorkerPool(workers int) *WorkerPool {
    return &WorkerPool{
        workers: workers,
        jobs:    make(chan Job, workers*2),
        results: make(chan Result, workers*2),
    }
}

func (p *WorkerPool) Start(ctx context.Context) {
    for i := 0; i < p.workers; i++ {
        p.wg.Add(1)
        go p.worker(ctx)
    }
}

func (p *WorkerPool) worker(ctx context.Context) {
    defer p.wg.Done()
    
    for {
        select {
        case <-ctx.Done():
            return
        case job, ok := <-p.jobs:
            if !ok {
                return
            }
            result := processJob(job)
            
            select {
            case p.results <- result:
            case <-ctx.Done():
                return
            }
        }
    }
}

func (p *WorkerPool) Submit(job Job) {
    p.jobs <- job
}

func (p *WorkerPool) Stop() {
    close(p.jobs)
    p.wg.Wait()
    close(p.results)
}
```

## Common Mistakes

### Mistake 1: Not Closing Channel

```go
// ❌ Receiver blocks forever
func Producer() <-chan int {
    out := make(chan int)
    go func() {
        for i := 0; i < 10; i++ {
            out <- i
        }
        // Forgot close(out)
    }()
    return out
}

for value := range Producer() {  // Never ends
    process(value)
}
```

### Mistake 2: Concurrent Map Access

```go
// ❌ Data race
var cache = make(map[string]string)

func Set(key, value string) {
    cache[key] = value  // Race
}

// ✅ Use sync.Map or Mutex
var cache sync.Map

func Set(key, value string) {
    cache.Store(key, value)
}
```

---

**Full details**: [Google Go Best Practices - Concurrency](https://google.github.io/styleguide/go/best-practices#concurrency)
