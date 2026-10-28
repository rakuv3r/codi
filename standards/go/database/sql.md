---
language: go
category: dependency
applies_to: []
---

# database/sql Standards

## Overview

The `database/sql` package provides a generic interface for SQL databases in Go. It manages connection pooling, prepared statements, and transactions. Understanding database/sql is essential for working with PostgreSQL, MySQL, SQLite, and any SQL database in Go.

## Core Principles

1. **Single DB instance** - Create once, reuse throughout application
2. **Always configure connection pool** - Don't use defaults
3. **Parameterized queries only** - Prevent SQL injection
4. **Always close resources** - Use defer for rows, statements, transactions
5. **Check rows.Err()** - After iteration completes
6. **Use context** - For timeouts and cancellation

## Connection Setup

### Creating Database Connection

```go
import (
    "database/sql"
    _ "github.com/lib/pq"  // PostgreSQL driver
)

func initDB() (*sql.DB, error) {
    db, err := sql.Open("postgres", "postgres://user:pass@localhost/dbname?sslmode=disable")
    if err != nil {
        return nil, err
    }
    
    // Verify connection
    if err := db.Ping(); err != nil {
        return nil, err
    }
    
    // Configure connection pool
    db.SetMaxOpenConns(25)                     // Max open connections
    db.SetMaxIdleConns(5)                      // Max idle connections
    db.SetConnMaxLifetime(5 * time.Minute)     // Max connection reuse time
    db.SetConnMaxIdleTime(1 * time.Minute)     // Max idle time
    
    return db, nil
}
```

**Why configure pool**: Default settings can cause connection exhaustion or excessive idle connections.

### Connection Pool Guidelines

```go
// Rule of thumb:
// MaxOpenConns < database max_connections
// MaxIdleConns ≈ expected concurrent queries
// ConnMaxLifetime matches database timeout policy

// Example for high-traffic API:
db.SetMaxOpenConns(100)
db.SetMaxIdleConns(20)
db.SetConnMaxLifetime(5 * time.Minute)
db.SetConnMaxIdleTime(1 * time.Minute)

// Example for low-traffic service:
db.SetMaxOpenConns(10)
db.SetMaxIdleConns(2)
db.SetConnMaxLifetime(10 * time.Minute)
db.SetConnMaxIdleTime(2 * time.Minute)
```

## SQL Injection Prevention

### ✅ Correct: Parameterized Queries

```go
// Correct: placeholder prevents injection
rows, err := db.Query("SELECT * FROM users WHERE id = ?", userID)

// PostgreSQL uses $1, $2, etc.
rows, err := db.Query("SELECT * FROM users WHERE name = $1 AND age > $2", name, minAge)

// Named parameters (some drivers)
rows, err := db.Query("SELECT * FROM users WHERE id = :id", sql.Named("id", userID))
```

### ❌ Wrong: String Concatenation

```go
// NEVER do this - SQL injection vulnerability
query := fmt.Sprintf("SELECT * FROM users WHERE id = %d", userID)
rows, err := db.Query(query)

// NEVER do this either
query := "SELECT * FROM users WHERE name = '" + name + "'"
rows, err := db.Query(query)
```

**Rule**: Always use placeholders (`?` or `$1`) for user input. Never concatenate strings.

## Querying Data

### Single Row Query

```go
func getUser(db *sql.DB, id int) (*User, error) {
    var user User
    
    err := db.QueryRow("SELECT id, name, email FROM users WHERE id = ?", id).
        Scan(&user.ID, &user.Name, &user.Email)
    
    if err != nil {
        if err == sql.ErrNoRows {
            return nil, ErrUserNotFound  // Not found is not a fatal error
        }
        return nil, fmt.Errorf("query user: %w", err)
    }
    
    return &user, nil
}
```

### Multiple Rows Query

```go
func listUsers(db *sql.DB, minAge int) ([]User, error) {
    rows, err := db.Query("SELECT id, name, email, age FROM users WHERE age >= ?", minAge)
    if err != nil {
        return nil, fmt.Errorf("query users: %w", err)
    }
    defer rows.Close()  // Critical: always close
    
    var users []User
    for rows.Next() {
        var user User
        if err := rows.Scan(&user.ID, &user.Name, &user.Email, &user.Age); err != nil {
            return nil, fmt.Errorf("scan user: %w", err)
        }
        users = append(users, user)
    }
    
    // Check for errors during iteration
    if err = rows.Err(); err != nil {
        return nil, fmt.Errorf("iterate users: %w", err)
    }
    
    return users, nil
}
```

**Critical**: Always call `defer rows.Close()` and check `rows.Err()`.

### With Context

```go
func getUser(ctx context.Context, db *sql.DB, id int) (*User, error) {
    var user User
    
    err := db.QueryRowContext(ctx, "SELECT id, name FROM users WHERE id = ?", id).
        Scan(&user.ID, &user.Name)
    
    if err != nil {
        if errors.Is(err, context.DeadlineExceeded) {
            return nil, errors.New("query timeout")
        }
        return nil, err
    }
    
    return &user, nil
}
```

## Handling NULL Values

### Using sql.Null Types

```go
type User struct {
    ID    int
    Name  string
    Email sql.NullString  // Can be NULL
    Age   sql.NullInt64
    Bio   sql.NullString
}

func getUser(db *sql.DB, id int) (*User, error) {
    var user User
    
    err := db.QueryRow("SELECT id, name, email, age, bio FROM users WHERE id = ?", id).
        Scan(&user.ID, &user.Name, &user.Email, &user.Age, &user.Bio)
    
    if err != nil {
        return nil, err
    }
    
    // Use NULL values
    email := "no email"
    if user.Email.Valid {
        email = user.Email.String
    }
    
    return &user, nil
}
```

**Available NULL types**:
- `sql.NullBool`
- `sql.NullFloat64`
- `sql.NullInt32`
- `sql.NullInt64`
- `sql.NullString`
- `sql.NullTime`

### Using Pointers (Simple Alternative)

```go
type User struct {
    ID    int
    Name  string
    Email *string  // nil if NULL
    Age   *int
}

func getUser(db *sql.DB, id int) (*User, error) {
    var user User
    
    err := db.QueryRow("SELECT id, name, email, age FROM users WHERE id = ?", id).
        Scan(&user.ID, &user.Name, &user.Email, &user.Age)
    
    if err != nil {
        return nil, err
    }
    
    // Check for nil
    if user.Email != nil {
        fmt.Println(*user.Email)
    }
    
    return &user, nil
}
```

## Executing Statements

### INSERT, UPDATE, DELETE

```go
func createUser(db *sql.DB, user *User) error {
    result, err := db.Exec(
        "INSERT INTO users (name, email, age) VALUES (?, ?, ?)",
        user.Name, user.Email, user.Age,
    )
    if err != nil {
        return fmt.Errorf("insert user: %w", err)
    }
    
    // Get auto-increment ID (if applicable)
    id, err := result.LastInsertId()
    if err != nil {
        return fmt.Errorf("get insert id: %w", err)
    }
    user.ID = int(id)
    
    return nil
}

func updateUser(db *sql.DB, user *User) error {
    result, err := db.Exec(
        "UPDATE users SET name = ?, email = ? WHERE id = ?",
        user.Name, user.Email, user.ID,
    )
    if err != nil {
        return fmt.Errorf("update user: %w", err)
    }
    
    // Check rows affected
    rowsAffected, err := result.RowsAffected()
    if err != nil {
        return fmt.Errorf("get rows affected: %w", err)
    }
    
    if rowsAffected == 0 {
        return ErrUserNotFound
    }
    
    return nil
}

func deleteUser(db *sql.DB, id int) error {
    _, err := db.Exec("DELETE FROM users WHERE id = ?", id)
    return err
}
```

## Transactions

### Standard Pattern with defer

```go
func transferMoney(ctx context.Context, db *sql.DB, fromID, toID int, amount float64) error {
    tx, err := db.BeginTx(ctx, nil)
    if err != nil {
        return fmt.Errorf("begin transaction: %w", err)
    }
    
    // Ensure rollback on error
    defer func() {
        if p := recover(); p != nil {
            tx.Rollback()
            panic(p)  // Re-throw panic
        } else if err != nil {
            tx.Rollback()
        }
    }()
    
    // Deduct from sender
    _, err = tx.ExecContext(ctx, "UPDATE accounts SET balance = balance - ? WHERE id = ?", amount, fromID)
    if err != nil {
        return fmt.Errorf("deduct from account: %w", err)
    }
    
    // Add to receiver
    _, err = tx.ExecContext(ctx, "UPDATE accounts SET balance = balance + ? WHERE id = ?", amount, toID)
    if err != nil {
        return fmt.Errorf("add to account: %w", err)
    }
    
    // Commit transaction
    if err = tx.Commit(); err != nil {
        return fmt.Errorf("commit transaction: %w", err)
    }
    
    return nil
}
```

### Transactor Pattern (Advanced)

```go
type Transactor interface {
    WithinTransaction(context.Context, func(context.Context) error) error
}

type transactor struct {
    db *sql.DB
}

func (t *transactor) WithinTransaction(ctx context.Context, fn func(context.Context) error) error {
    tx, err := t.db.BeginTx(ctx, nil)
    if err != nil {
        return fmt.Errorf("begin tx: %w", err)
    }
    
    // Store transaction in context
    txCtx := context.WithValue(ctx, txKey{}, tx)
    
    // Execute function
    if err := fn(txCtx); err != nil {
        tx.Rollback()  // Ignore rollback error
        return err
    }
    
    // Commit
    if err := tx.Commit(); err != nil {
        return fmt.Errorf("commit tx: %w", err)
    }
    
    return nil
}

// Usage:
err := transactor.WithinTransaction(ctx, func(ctx context.Context) error {
    // All operations use context with transaction
    _, err := repo.UpdateBalance(ctx, accountID, newBalance)
    return err
})
```

### Transaction Isolation Levels

```go
tx, err := db.BeginTx(ctx, &sql.TxOptions{
    Isolation: sql.LevelSerializable,
    ReadOnly:  false,
})

// Available levels:
// sql.LevelDefault         - Driver default
// sql.LevelReadUncommitted
// sql.LevelReadCommitted
// sql.LevelRepeatableRead
// sql.LevelSerializable
```

## Prepared Statements

### When to Use

Use prepared statements for queries executed multiple times with different parameters.

```go
func batchInsert(db *sql.DB, users []User) error {
    stmt, err := db.Prepare("INSERT INTO users (name, email) VALUES (?, ?)")
    if err != nil {
        return err
    }
    defer stmt.Close()  // Critical: always close
    
    for _, user := range users {
        _, err := stmt.Exec(user.Name, user.Email)
        if err != nil {
            return fmt.Errorf("insert user %s: %w", user.Name, err)
        }
    }
    
    return nil
}
```

### In Transactions

```go
func batchInsertTx(ctx context.Context, tx *sql.Tx, users []User) error {
    stmt, err := tx.PrepareContext(ctx, "INSERT INTO users (name, email) VALUES (?, ?)")
    if err != nil {
        return err
    }
    defer stmt.Close()
    
    for _, user := range users {
        _, err := stmt.ExecContext(ctx, user.Name, user.Email)
        if err != nil {
            return err
        }
    }
    
    return nil
}
```

**Note**: Prepared statement ties to its connection. In connection pool, may not provide performance benefit for one-off queries.

## Error Handling

### Common Errors

```go
// No rows found (not an error in many cases)
if err == sql.ErrNoRows {
    return nil, ErrNotFound
}

// Connection issues
if errors.Is(err, driver.ErrBadConn) {
    // Connection broken, will auto-retry
}

// Transaction already completed
if errors.Is(err, sql.ErrTxDone) {
    // Transaction already committed or rolled back
}

// Context cancellation
if errors.Is(err, context.Canceled) {
    // Request cancelled
}

if errors.Is(err, context.DeadlineExceeded) {
    // Request timeout
}
```

### Wrapping Errors

```go
if err != nil {
    return fmt.Errorf("failed to get user %d: %w", id, err)
}
```

## Testing Patterns

### With Test Database

```go
func setupTestDB(t *testing.T) *sql.DB {
    db, err := sql.Open("postgres", "postgres://localhost/test_db")
    require.NoError(t, err)
    
    // Run migrations
    _, err = db.Exec(`CREATE TABLE IF NOT EXISTS users (
        id SERIAL PRIMARY KEY,
        name TEXT NOT NULL,
        email TEXT
    )`)
    require.NoError(t, err)
    
    // Clean up
    t.Cleanup(func() {
        db.Exec("DROP TABLE users")
        db.Close()
    })
    
    return db
}

func TestCreateUser(t *testing.T) {
    db := setupTestDB(t)
    
    user := &User{Name: "Alice", Email: "alice@example.com"}
    err := createUser(db, user)
    
    require.NoError(t, err)
    assert.Greater(t, user.ID, 0)
}
```

### With Mocks

Use interface abstraction:

```go
type DB interface {
    QueryContext(context.Context, string, ...interface{}) (*sql.Rows, error)
    ExecContext(context.Context, string, ...interface{}) (sql.Result, error)
    // ... other methods
}

type UserRepo struct {
    db DB
}

// In tests, mock the DB interface
```

## Common Pitfalls

### ❌ Not closing rows

```go
// Wrong: resource leak
rows, _ := db.Query("SELECT * FROM users")
for rows.Next() {
    // ...
}

// Right
rows, _ := db.Query("SELECT * FROM users")
defer rows.Close()
for rows.Next() {
    // ...
}
```

### ❌ Not checking rows.Err()

```go
// Wrong: missing iteration errors
for rows.Next() {
    rows.Scan(&user)
}

// Right
for rows.Next() {
    rows.Scan(&user)
}
if err := rows.Err(); err != nil {
    return err
}
```

### ❌ Creating multiple DB instances

```go
// Wrong: each call creates new pool
func getUser(id int) {
    db, _ := sql.Open("postgres", connString)
    defer db.Close()
    // ...
}

// Right: single global instance
var db *sql.DB

func init() {
    db, _ = sql.Open("postgres", connString)
}

func getUser(id int) {
    // Use global db
}
```

### ❌ Not using parameterized queries

```go
// Wrong: SQL injection
query := fmt.Sprintf("SELECT * FROM users WHERE name = '%s'", name)
db.Query(query)

// Right
db.Query("SELECT * FROM users WHERE name = ?", name)
```

### ❌ Forgetting defer in transaction

```go
// Wrong: no rollback on error
tx, _ := db.Begin()
tx.Exec("UPDATE ...")
tx.Commit()

// Right
tx, _ := db.Begin()
defer func() {
    if err != nil {
        tx.Rollback()
    }
}()
tx.Exec("UPDATE ...")
err = tx.Commit()
```

## Best Practices Summary

1. **Create DB once** - Single instance per database, reuse globally
2. **Configure pool** - Set MaxOpenConns, MaxIdleConns, timeouts
3. **Use placeholders** - Never concatenate SQL strings
4. **Always defer Close** - rows, statements, transactions
5. **Check rows.Err()** - After iteration completes
6. **Use context** - For timeout and cancellation
7. **Handle NULL** - Use sql.Null* types or pointers
8. **Wrap errors** - Include context in error messages
9. **Test with real DB** - Use test database for integration tests

## References

- Official docs: https://pkg.go.dev/database/sql
- Tutorial: https://go.dev/doc/database/
- Drivers: https://github.com/golang/go/wiki/SQLDrivers
- Popular libraries: sqlx, pgx (PostgreSQL-specific)
