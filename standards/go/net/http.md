# net/http Standards

## Overview

The `net/http` package is Go's standard library for HTTP servers and clients. It provides production-ready HTTP/1.1 and HTTP/2 support with excellent performance. Understanding net/http is essential for building web services and understanding frameworks built on top of it (Gin, Echo, Fiber).

## Server Patterns

### Basic Server Structure

```go
func main() {
    mux := http.NewServeMux()
    mux.HandleFunc("/", handleHome)
    mux.HandleFunc("/users/{id}", handleUser)  // Go 1.22+ path params
    
    server := &http.Server{
        Addr:         ":8080",
        Handler:      mux,
        ReadTimeout:  5 * time.Second,
        WriteTimeout: 10 * time.Second,
        IdleTimeout:  120 * time.Second,
    }
    
    log.Fatal(server.ListenAndServe())
}
```

### Handler Patterns

**Handler Function**:

```go
func handleUser(w http.ResponseWriter, r *http.Request) {
    id := r.PathValue("id")  // Go 1.22+
    
    user, err := getUser(id)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(user)
}
```

**Handler Maker Pattern** (Recommended):

```go
func handleUser(logger *log.Logger, db *sql.DB) http.HandlerFunc {
    // Initialization happens once
    stmt, err := db.Prepare("SELECT * FROM users WHERE id = ?")
    if err != nil {
        log.Fatal(err)
    }
    
    return func(w http.ResponseWriter, r *http.Request) {
        // Handler has access to logger, db, stmt via closure
        id := r.PathValue("id")
        var user User
        err := stmt.QueryRow(id).Scan(&user.ID, &user.Name)
        if err != nil {
            logger.Printf("error: %v", err)
            http.Error(w, "Not found", http.StatusNotFound)
            return
        }
        json.NewEncoder(w).Encode(user)
    }
}

// Usage:
mux.HandleFunc("/users/{id}", handleUser(logger, db))
```

**Why**: Dependency injection, initialization outside hot path, testability.

### Lazy Initialization (Performance)

```go
func handleTemplate(files ...string) http.HandlerFunc {
    var (
        init sync.Once
        tpl  *template.Template
        err  error
    )
    
    return func(w http.ResponseWriter, r *http.Request) {
        init.Do(func() {
            tpl, err = template.ParseFiles(files...)
        })
        
        if err != nil {
            http.Error(w, err.Error(), http.StatusInternalServerError)
            return
        }
        
        tpl.Execute(w, data)
    }
}
```

**Why**: Improves startup time, parses templates only on first request.

## Middleware Pattern

```go
func middleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // Before
        start := time.Now()
        
        next.ServeHTTP(w, r)
        
        // After
        log.Printf("%s %s %v", r.Method, r.URL.Path, time.Since(start))
    })
}
```

### Common Middleware Examples

**Logging**:

```go
func logging(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        log.Printf("%s %s %s", r.RemoteAddr, r.Method, r.URL)
        next.ServeHTTP(w, r)
    })
}
```

**Authentication**:

```go
func requireAuth(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        token := r.Header.Get("Authorization")
        if !isValid(token) {
            http.Error(w, "Unauthorized", http.StatusUnauthorized)
            return  // Don't call next
        }
        next.ServeHTTP(w, r)
    })
}
```

**CORS**:

```go
func cors(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("Access-Control-Allow-Origin", "*")
        w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE")
        w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
        
        if r.Method == http.MethodOptions {
            w.WriteHeader(http.StatusOK)
            return
        }
        
        next.ServeHTTP(w, r)
    })
}
```

### Middleware Chain

```go
// Apply multiple middleware
handler := logging(cors(requireAuth(mux)))
server := &http.Server{
    Addr:    ":8080",
    Handler: handler,
}
```

**Or** create a helper:

```go
func chain(h http.Handler, middleware ...func(http.Handler) http.Handler) http.Handler {
    for i := len(middleware) - 1; i >= 0; i-- {
        h = middleware[i](h)
    }
    return h
}

// Usage:
handler := chain(mux, logging, cors, requireAuth)
```

## Request Handling

### Reading Request

```go
func handleCreate(w http.ResponseWriter, r *http.Request) {
    // Parse JSON body
    var req CreateUserRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }
    defer r.Body.Close()
    
    // Query parameters
    page := r.URL.Query().Get("page")
    
    // Path parameters (Go 1.22+)
    id := r.PathValue("id")
    
    // Headers
    token := r.Header.Get("Authorization")
    
    // Form data
    r.ParseForm()
    name := r.FormValue("name")
    
    // Files
    file, header, err := r.FormFile("upload")
    if err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }
    defer file.Close()
}
```

### Writing Response

```go
// JSON
func respondJSON(w http.ResponseWriter, status int, data interface{}) {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(status)
    json.NewEncoder(w).Encode(data)
}

// Error
func respondError(w http.ResponseWriter, code int, message string) {
    respondJSON(w, code, map[string]string{"error": message})
}

// Plain text
func respondText(w http.ResponseWriter, status int, text string) {
    w.Header().Set("Content-Type", "text/plain; charset=utf-8")
    w.WriteHeader(status)
    fmt.Fprint(w, text)
}
```

**Important**: Call `WriteHeader()` before `Write()`. Headers cannot be modified after writing body.

## Context Usage

```go
func handleWithTimeout(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()
    
    // Add timeout
    ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
    defer cancel()
    
    // Pass to database/external service
    result, err := queryDB(ctx, id)
    if err != nil {
        if errors.Is(err, context.DeadlineExceeded) {
            http.Error(w, "Request timeout", http.StatusGatewayTimeout)
            return
        }
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    
    json.NewEncoder(w).Encode(result)
}
```

### Storing Values in Context

```go
type contextKey string

const userIDKey contextKey = "user_id"

// Middleware sets value
func auth(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        userID := extractUserID(r)
        ctx := context.WithValue(r.Context(), userIDKey, userID)
        next.ServeHTTP(w, r.WithContext(ctx))
    })
}

// Handler retrieves value
func handleProfile(w http.ResponseWriter, r *http.Request) {
    userID := r.Context().Value(userIDKey).(int)
    // Use userID
}
```

## Graceful Shutdown

```go
func main() {
    mux := http.NewServeMux()
    mux.HandleFunc("/", handleHome)
    
    server := &http.Server{
        Addr:    ":8080",
        Handler: mux,
    }
    
    // Start server in goroutine
    go func() {
        if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
            log.Fatalf("ListenAndServe: %v", err)
        }
    }()
    
    // Wait for interrupt signal
    quit := make(chan os.Signal, 1)
    signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
    <-quit
    
    log.Println("Shutting down server...")
    
    // Graceful shutdown with timeout
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()
    
    if err := server.Shutdown(ctx); err != nil {
        log.Fatalf("Server forced to shutdown: %v", err)
    }
    
    log.Println("Server exiting")
}
```

**Flow**:
1. Stop accepting new connections
2. Wait for active requests to complete (up to timeout)
3. Close idle connections
4. Exit

## Server Configuration

```go
server := &http.Server{
    Addr:    ":8080",
    Handler: mux,
    
    // Timeouts
    ReadTimeout:       5 * time.Second,   // Time to read request
    ReadHeaderTimeout: 2 * time.Second,   // Time to read headers
    WriteTimeout:      10 * time.Second,  // Time to write response
    IdleTimeout:       120 * time.Second, // Keep-alive timeout
    
    // Limits
    MaxHeaderBytes: 1 << 20, // 1 MB
    
    // TLS (optional)
    TLSConfig: &tls.Config{
        MinVersion: tls.VersionTLS12,
    },
    
    // Error logging
    ErrorLog: log.New(os.Stderr, "http: ", log.LstdFlags),
}
```

**Why configure timeouts**: Prevents slow clients from exhausting server resources.

## Testing

### httptest.ResponseRecorder

```go
func TestHandleUser(t *testing.T) {
    req := httptest.NewRequest("GET", "/users/1", nil)
    w := httptest.NewRecorder()
    
    handleUser(w, req)
    
    resp := w.Result()
    defer resp.Body.Close()
    
    assert.Equal(t, http.StatusOK, resp.StatusCode)
    assert.Equal(t, "application/json", resp.Header.Get("Content-Type"))
    
    var user User
    json.NewDecoder(resp.Body).Decode(&user)
    assert.Equal(t, "Alice", user.Name)
}
```

### httptest.Server

```go
func TestIntegration(t *testing.T) {
    mux := http.NewServeMux()
    mux.HandleFunc("/ping", func(w http.ResponseWriter, r *http.Request) {
        w.Write([]byte("pong"))
    })
    
    server := httptest.NewServer(mux)
    defer server.Close()
    
    resp, err := http.Get(server.URL + "/ping")
    require.NoError(t, err)
    defer resp.Body.Close()
    
    body, _ := io.ReadAll(resp.Body)
    assert.Equal(t, "pong", string(body))
}
```

## HTTP Client

### Basic Client

```go
resp, err := http.Get("https://api.example.com/users")
if err != nil {
    return err
}
defer resp.Body.Close()

if resp.StatusCode != http.StatusOK {
    return fmt.Errorf("unexpected status: %d", resp.StatusCode)
}

var users []User
if err := json.NewDecoder(resp.Body).Decode(&users); err != nil {
    return err
}
```

### Custom Client

```go
client := &http.Client{
    Timeout: 10 * time.Second,
    Transport: &http.Transport{
        MaxIdleConns:        100,
        MaxIdleConnsPerHost: 10,
        IdleConnTimeout:     90 * time.Second,
    },
}

resp, err := client.Get(url)
```

### POST with JSON

```go
func createUser(user *User) error {
    body, err := json.Marshal(user)
    if err != nil {
        return err
    }
    
    req, err := http.NewRequest("POST", "https://api.example.com/users", bytes.NewReader(body))
    if err != nil {
        return err
    }
    
    req.Header.Set("Content-Type", "application/json")
    req.Header.Set("Authorization", "Bearer "+token)
    
    resp, err := client.Do(req)
    if err != nil {
        return err
    }
    defer resp.Body.Close()
    
    if resp.StatusCode != http.StatusCreated {
        body, _ := io.ReadAll(resp.Body)
        return fmt.Errorf("failed: %s", body)
    }
    
    return nil
}
```

### With Context

```go
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()

req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
if err != nil {
    return err
}

resp, err := client.Do(req)
if err != nil {
    if errors.Is(err, context.DeadlineExceeded) {
        return errors.New("request timeout")
    }
    return err
}
defer resp.Body.Close()
```

## Helper Functions

### JSON Encode/Decode

```go
func encode[T any](w http.ResponseWriter, status int, v T) error {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(status)
    return json.NewEncoder(w).Encode(v)
}

func decode[T any](r *http.Request) (T, error) {
    var v T
    if err := json.NewDecoder(r.Body).Decode(&v); err != nil {
        return v, err
    }
    return v, nil
}

// Usage:
user, err := decode[User](r)
if err != nil {
    http.Error(w, err.Error(), http.StatusBadRequest)
    return
}

encode(w, http.StatusOK, user)
```

## Anti-Patterns

### ❌ Don't use global ServeMux

```go
// Wrong: uses default ServeMux (global state)
http.HandleFunc("/", handler)
http.ListenAndServe(":8080", nil)

// Right: explicit ServeMux
mux := http.NewServeMux()
mux.HandleFunc("/", handler)
http.ListenAndServe(":8080", mux)
```

### ❌ Don't ignore errors from Write

```go
// Wrong
w.Write([]byte("response"))

// Right
if _, err := w.Write([]byte("response")); err != nil {
    log.Printf("write error: %v", err)
}
```

### ❌ Don't call WriteHeader multiple times

```go
// Wrong
w.WriteHeader(http.StatusOK)
w.WriteHeader(http.StatusCreated)  // Ignored, logs warning

// Right: call once
w.WriteHeader(http.StatusCreated)
```

### ❌ Don't forget to set Content-Type

```go
// Wrong: browser may misinterpret
json.NewEncoder(w).Encode(data)

// Right
w.Header().Set("Content-Type", "application/json")
json.NewEncoder(w).Encode(data)
```

### ❌ Don't configure server without timeouts

```go
// Wrong: vulnerable to slow clients
server := &http.Server{Addr: ":8080", Handler: mux}

// Right
server := &http.Server{
    Addr:         ":8080",
    Handler:      mux,
    ReadTimeout:  5 * time.Second,
    WriteTimeout: 10 * time.Second,
    IdleTimeout:  120 * time.Second,
}
```

## References

- Official docs: https://pkg.go.dev/net/http
- HTTP/2: Enabled automatically with TLS
- Used by: All Go web frameworks (Gin, Echo, Fiber build on this)
