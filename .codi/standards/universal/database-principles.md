---
language: universal
category: universal
applies_to: []
---

# Database Patterns

Based on [Postgres Best Practices](https://supabase.com/blog/postgres-best-practices-for-ai-agents), [Transaction Isolation Guide](https://codemia.io/courses/system_design_fundamentals/transaction_isolation_level), and [Database Optimization](https://blog.easecloud.io/cloud-infrastructure/optimization-for-slow-queries-and-indexing-issues/).

## Quick Reference

- **Connection pooling** - reuse connections, prevent exhaustion
- **Read Committed** - default isolation level for most cases
- **N+1 prevention** - eager load related data
- **Index selectively** - index queries you actually run
- **Migration safety** - zero-downtime migrations

## Core Rules

### 1. Connection Pooling

```
# ✅ Connection pool configuration
Min connections: 5
Max connections: 20 (rule of thumb: 2-3x CPU cores)
Connection timeout: 5s
Idle timeout: 300s (5 minutes)
Max lifetime: 1800s (30 minutes)

# ✅ Always release connections
try:
    conn = pool.get_connection()
    # Use connection
finally:
    conn.close()  # Return to pool

# ❌ Connection leak
conn = pool.get_connection()
result = conn.query()
# Forgot to close - connection never returned to pool
```

**Best practices**:
- Set `max_connections` < database server limit
- Enable `pool_pre_ping` to detect stale connections
- Set `pool_recycle` < database timeout
- Monitor pool utilization

### 2. Transaction Isolation Levels

Based on [Transaction Isolation Guide](https://codemia.io/courses/system_design_fundamentals/transaction_isolation_level).

```
Read Uncommitted - dirty reads possible (rarely used)
  ❌ Can see uncommitted changes from other transactions

Read Committed - default for most databases
  ✅ Only see committed data
  ❌ Non-repeatable reads possible
  Use: Most OLTP applications

Repeatable Read - consistent snapshot
  ✅ Same query returns same results in transaction
  ❌ Phantom reads possible in some databases
  Use: Reports, analytics

Serializable - strongest isolation
  ✅ Full isolation, as if transactions run sequentially
  ❌ Performance impact, possible deadlocks
  Use: Financial transactions, inventory updates
```

**Choosing isolation level**:
```
# ✅ Default: Read Committed
Most web APIs, CRUD operations

# ✅ Repeatable Read
Reports that require consistency
Multi-step calculations

# ✅ Serializable
Money transfers
Inventory management
Booking systems (prevent double-booking)
```

### 3. N+1 Query Prevention

Based on [N+1 Prevention Guide](https://logicandlegacy.hashnode.dev/advanced-python-sqlite-indices-n1-prevention-the-repository-pattern-2026).

```
# ❌ N+1 query problem
users = db.query("SELECT * FROM users LIMIT 100")
for user in users:
    posts = db.query("SELECT * FROM posts WHERE user_id = ?", user.id)
    # 1 query for users + 100 queries for posts = 101 queries

# ✅ Eager loading (single JOIN)
results = db.query("""
    SELECT users.*, posts.*
    FROM users
    LEFT JOIN posts ON posts.user_id = users.id
    WHERE users.id IN (...)
""")
# 1 query total

# ✅ Separate queries with IN clause
users = db.query("SELECT * FROM users LIMIT 100")
user_ids = [u.id for u in users]
posts = db.query("SELECT * FROM posts WHERE user_id IN (?)", user_ids)
# 2 queries total (much better than 101)
```

**Detection**:
- Monitor query count per request
- Use database query logs
- Profile with tools like `django-debug-toolbar`

### 4. Index Design

Based on [Database Index Review](https://www.momentslog.com/development/database-index-review-checklist-how-to-fix-slow-queries-without-creating-new-production-risk).

```
# ✅ Single-column index
CREATE INDEX idx_users_email ON users(email);

# ✅ Composite index (order matters!)
CREATE INDEX idx_orders_user_status ON orders(user_id, status, created_at);
# Good for: WHERE user_id = ? AND status = ?
# Good for: WHERE user_id = ?
# Bad for: WHERE status = ? (doesn't use index)

# ✅ Partial index (for sparse data)
CREATE INDEX idx_active_users ON users(email) WHERE status = 'active';
# Only indexes active users, smaller and faster

# ✅ Covering index (includes all needed columns)
CREATE INDEX idx_users_lookup ON users(email) INCLUDE (name, created_at);
# Query doesn't need to access table
```

**Index guidelines**:
- Index columns in WHERE, JOIN, ORDER BY
- Index reference columns (user_id, order_id)
- Don't index low-cardinality columns (gender, boolean)
- Composite index: most selective column first
- Limit: ~5-10 indexes per table

### 5. Query Optimization

```
# ❌ SELECT * (fetches unnecessary data)
SELECT * FROM users WHERE id = 1;

# ✅ Select specific columns
SELECT id, name, email FROM users WHERE id = 1;

# ❌ No index on WHERE clause
SELECT * FROM orders WHERE status = 'pending';
# Requires full table scan

# ✅ Indexed column
CREATE INDEX idx_orders_status ON orders(status);
SELECT * FROM orders WHERE status = 'pending';

# ✅ Use EXPLAIN to analyze
EXPLAIN ANALYZE SELECT * FROM orders WHERE status = 'pending';
# Check: Index Scan vs Sequential Scan
```

### 6. Migration Safety

Based on [Zero-Downtime Migrations](https://www.bluetickconsultants.com/rails-database-indexing/).

```
# ✅ Add column with default (safe)
ALTER TABLE users ADD COLUMN status VARCHAR(20) DEFAULT 'active';

# ❌ Add NOT NULL column without default (breaks inserts)
ALTER TABLE users ADD COLUMN status VARCHAR(20) NOT NULL;

# ✅ Multi-step approach
Step 1: Add nullable column
ALTER TABLE users ADD COLUMN status VARCHAR(20);

Step 2: Backfill data in batches
UPDATE users SET status = 'active' WHERE status IS NULL LIMIT 1000;

Step 3: Add NOT NULL constraint
ALTER TABLE users ALTER COLUMN status SET NOT NULL;

# ✅ Create index concurrently (PostgreSQL)
CREATE INDEX CONCURRENTLY idx_users_email ON users(email);
# Doesn't lock table for writes
```

**Migration guidelines**:
- Test on production-like data volume
- Avoid DDL in transactions (PostgreSQL)
- Use `CONCURRENTLY` for large tables
- Have rollback plan
- Monitor replication lag

### 7. Transaction Best Practices

```
# ✅ Keep transactions short
BEGIN;
UPDATE accounts SET balance = balance - 100 WHERE id = 1;
UPDATE accounts SET balance = balance + 100 WHERE id = 2;
COMMIT;

# ❌ Long-running transaction
BEGIN;
UPDATE accounts ...;
# Do NOT perform network calls, file I/O inside transaction
COMMIT;

# ✅ Handle deadlocks
for attempt in range(3):
    try:
        with transaction:
            # Critical operations
        break
    except DeadlockDetected:
        if attempt == 2:
            raise
        sleep(0.1 * 2 ** attempt)  # Exponential backoff
```

### 8. Schema Design

```
# ✅ Normalize for consistency
users:
  id, name, email

orders:
  id, user_id, total, created_at

# ✅ Denormalize for performance (when needed)
orders:
  id, user_id, user_name, total, created_at
# Duplicate user_name to avoid JOIN

# ✅ When to denormalize:
- Read-heavy workload
- Rare updates
- JOIN is expensive
- Always keep primary source of truth

# ✅ Maintain referential integrity in application
- Validate user_id exists before insert
- Handle cascading deletes in code
- Use database indexes for performance
```

## Common Mistakes

### Mistake 1: Connection Pool Exhaustion

```
# ❌ Pool too small, connections leaked
Max connections: 5
# Under load: pool exhausted, requests timeout

# ✅ Proper sizing + leak prevention
Max connections: 20
Always use try/finally or context managers
```

### Mistake 2: Wrong Isolation Level

```
# ❌ Using Serializable for everything
Serializable isolation for all transactions
# Result: Poor performance, many deadlocks

# ✅ Choose appropriate level
Read Committed: most operations
Repeatable Read: reports
Serializable: critical operations only
```

### Mistake 3: Missing Indexes

```
# ❌ No index on reference column
orders(user_id) - no index
# Slow: SELECT * FROM orders WHERE user_id = ?

# ✅ Index reference columns
CREATE INDEX idx_orders_user_id ON orders(user_id);
```

### Mistake 4: Over-indexing

```
# ❌ Too many indexes
10+ indexes on same table
# Result: Slow writes, wasted storage

# ✅ Index what you query
Analyze actual query patterns
Remove unused indexes
```

---

**References**:
- [Postgres Best Practices for AI Agents](https://supabase.com/blog/postgres-best-practices-for-ai-agents)
- [Transaction Isolation Levels Guide](https://codemia.io/courses/system_design_fundamentals/transaction_isolation_level)
- [Database Optimization Guide](https://blog.easecloud.io/cloud-infrastructure/optimization-for-slow-queries-and-indexing-issues/)
- [N+1 Prevention Pattern](https://logicandlegacy.hashnode.dev/advanced-python-sqlite-indices-n1-prevention-the-repository-pattern-2026)
