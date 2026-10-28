---
language: go
category: dependency
applies_to: []
---

# Database

Go-specific database patterns and implementations.

Based on [database/sql](https://pkg.go.dev/database/sql), [GORM](https://gorm.io/), and [sqlx](https://github.com/jmoiron/sqlx).

## Quick Reference

- **database/sql** - standard library database interface
- **Connection pool** - 20 max connections (2-3x CPU cores)
- **GORM** - ORM for rapid development
- **sqlx** - extensions to database/sql
- **pgx** - high-performance PostgreSQL driver

## Connection Pooling

### database/sql Configuration

```go
import (
    "database/sql"
    "time"

    _ "github.com/lib/pq" // PostgreSQL driver
)

// ✅ Configure connection pool
func NewDB(connString string) (*sql.DB, error) {
    db, err := sql.Open("postgres", connString)
    if err != nil {
        return nil, err
    }

    // Connection pool settings
    db.SetMaxOpenConns(20)                 // Maximum open connections
    db.SetMaxIdleConns(5)                  // Minimum idle connections
    db.SetConnMaxLifetime(30 * time.Minute) // Connection max lifetime
    db.SetConnMaxIdleTime(5 * time.Minute)  // Idle connection timeout

    // Verify connection
    if err := db.Ping(); err != nil {
        return nil, err
    }

    return db, nil
}
```

### Usage with Context

```go
// ✅ Always use context
func GetUser(ctx context.Context, db *sql.DB, userID int) (*User, error) {
    query := "SELECT id, name, email FROM users WHERE id = $1"
    
    var user User
    err := db.QueryRowContext(ctx, query, userID).Scan(&user.ID, &user.Name, &user.Email)
    if err == sql.ErrNoRows {
        return nil, nil
    }
    if err != nil {
        return nil, err
    }
    
    return &user, nil
}
```

## Transactions

```go
// ✅ Transaction pattern
func TransferMoney(ctx context.Context, db *sql.DB, fromID, toID int, amount float64) error {
    tx, err := db.BeginTx(ctx, &sql.TxOptions{
        Isolation: sql.LevelReadCommitted,
    })
    if err != nil {
        return err
    }
    defer tx.Rollback() // Rollback if not committed

    // Deduct from sender
    _, err = tx.ExecContext(ctx,
        "UPDATE accounts SET balance = balance - $1 WHERE id = $2",
        amount, fromID)
    if err != nil {
        return err
    }

    // Add to receiver
    _, err = tx.ExecContext(ctx,
        "UPDATE accounts SET balance = balance + $1 WHERE id = $2",
        amount, toID)
    if err != nil {
        return err
    }

    return tx.Commit()
}
```

## GORM

### Setup

```go
import (
    "gorm.io/driver/postgres"
    "gorm.io/gorm"
    "gorm.io/gorm/logger"
)

// ✅ GORM configuration
func NewGormDB(dsn string) (*gorm.DB, error) {
    db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
        Logger: logger.Default.LogMode(logger.Info),
        NowFunc: func() time.Time {
            return time.Now().UTC()
        },
    })
    if err != nil {
        return nil, err
    }

    sqlDB, err := db.DB()
    if err != nil {
        return nil, err
    }

    // Connection pool
    sqlDB.SetMaxOpenConns(20)
    sqlDB.SetMaxIdleConns(5)
    sqlDB.SetConnMaxLifetime(30 * time.Minute)

    return db, nil
}
```

### Models

```go
// ✅ GORM model
type User struct {
    ID        uint           `gorm:"primaryKey"`
    Name      string         `gorm:"size:100;not null"`
    Email     string         `gorm:"uniqueIndex;size:100;not null"`
    Password  string         `gorm:"size:255;not null"`
    CreatedAt time.Time
    UpdatedAt time.Time
    DeletedAt gorm.DeletedAt `gorm:"index"`
}
```

### CRUD Operations

```go
// ✅ Create
func CreateUser(db *gorm.DB, user *User) error {
    return db.Create(user).Error
}

// ✅ Read
func GetUser(db *gorm.DB, id uint) (*User, error) {
    var user User
    err := db.First(&user, id).Error
    if err == gorm.ErrRecordNotFound {
        return nil, nil
    }
    return &user, err
}

// ✅ Update
func UpdateUser(db *gorm.DB, user *User) error {
    return db.Save(user).Error
}

// ✅ Delete (soft delete)
func DeleteUser(db *gorm.DB, id uint) error {
    return db.Delete(&User{}, id).Error
}
```

### Eager Loading (N+1 Prevention)

```go
// ❌ N+1 problem
users, _ := db.Find(&[]User{})
for _, user := range users {
    db.Model(&user).Association("Posts").Find(&user.Posts) // N queries
}

// ✅ Eager loading
var users []User
db.Preload("Posts").Find(&users) // 1 or 2 queries
```

### Pagination

```go
// ✅ Cursor pagination
func GetUsers(db *gorm.DB, cursor uint, limit int) ([]User, error) {
    var users []User
    err := db.Where("id > ?", cursor).
        Order("id ASC").
        Limit(limit).
        Find(&users).Error
    return users, err
}

// ✅ Offset pagination (small datasets only)
func GetUsersPage(db *gorm.DB, page, pageSize int) ([]User, int64, error) {
    var users []User
    var total int64

    db.Model(&User{}).Count(&total)
    
    err := db.Offset((page - 1) * pageSize).
        Limit(pageSize).
        Find(&users).Error
    
    return users, total, err
}
```

## sqlx

```go
import "github.com/jmoiron/sqlx"

// ✅ sqlx.DB setup
db, err := sqlx.Connect("postgres", connString)
db.SetMaxOpenConns(20)
db.SetMaxIdleConns(5)

// ✅ Struct scanning
type User struct {
    ID    int    `db:"id"`
    Name  string `db:"name"`
    Email string `db:"email"`
}

// ✅ Get single row
var user User
err := db.Get(&user, "SELECT * FROM users WHERE id = $1", userID)

// ✅ Select multiple rows
var users []User
err := db.Select(&users, "SELECT * FROM users WHERE status = $1", "active")

// ✅ Named queries
query := `INSERT INTO users (name, email) VALUES (:name, :email) RETURNING id`
rows, err := db.NamedQuery(query, map[string]interface{}{
    "name":  "Alice",
    "email": "alice@example.com",
})
```

## Migrations

### golang-migrate

```bash
# Install
go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest

# Create migration
migrate create -ext sql -dir migrations -seq create_users_table

# Run migrations
migrate -path migrations -database "postgres://..." up

# Rollback
migrate -path migrations -database "postgres://..." down 1
```

```sql
-- migrations/000001_create_users_table.up.sql
CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    email VARCHAR(100) NOT NULL UNIQUE,
    password VARCHAR(255) NOT NULL,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX idx_users_email ON users(email);

-- migrations/000001_create_users_table.down.sql
DROP TABLE IF EXISTS users;
```

### Programmatic Migrations

```go
import (
    "github.com/golang-migrate/migrate/v4"
    _ "github.com/golang-migrate/migrate/v4/database/postgres"
    _ "github.com/golang-migrate/migrate/v4/source/file"
)

func RunMigrations(dbURL string) error {
    m, err := migrate.New(
        "file://migrations",
        dbURL,
    )
    if err != nil {
        return err
    }
    
    if err := m.Up(); err != nil && err != migrate.ErrNoChange {
        return err
    }
    
    return nil
}
```

## Best Practices

```go
// ✅ Use context for cancellation
ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
defer cancel()
rows, err := db.QueryContext(ctx, query)

// ✅ Always use placeholders (prevent SQL injection)
query := "SELECT * FROM users WHERE email = $1"
db.QueryContext(ctx, query, email) // ✅ Safe

// ❌ Never concatenate SQL
query := fmt.Sprintf("SELECT * FROM users WHERE email = '%s'", email) // ❌ Dangerous

// ✅ Handle sql.ErrNoRows
user, err := GetUser(ctx, db, userID)
if err == sql.ErrNoRows {
    return nil, errors.New("user not found")
}

// ✅ Close rows
rows, err := db.Query(query)
if err != nil {
    return err
}
defer rows.Close() // Always close
```

---

**References**:
- [database/sql Documentation](https://pkg.go.dev/database/sql)
- [GORM Guide](https://gorm.io/docs/)
- [sqlx](https://github.com/jmoiron/sqlx)
- [golang-migrate](https://github.com/golang-migrate/migrate)
