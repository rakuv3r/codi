# Database

Python-specific database patterns and implementations.

Based on [SQLAlchemy 2.0](https://docs.sqlalchemy.org/), [asyncpg](https://magicstack.github.io/asyncpg/), and [Alembic](https://alembic.sqlalchemy.org/).

## Quick Reference

- **SQLAlchemy 2.0** - ORM and SQL toolkit
- **asyncpg** - high-performance PostgreSQL async driver
- **Alembic** - database migrations
- **Connection pool** - 20 max connections (2-3x CPU cores)
- **Async sessions** - use AsyncSession for async operations

## SQLAlchemy Setup

### Async Engine and Session

```python
from sqlalchemy.ext.asyncio import create_async_engine, AsyncSession, async_sessionmaker
from sqlalchemy.orm import DeclarativeBase

# ✅ Create async engine
engine = create_async_engine(
    "postgresql+asyncpg://user:pass@localhost/dbname",
    echo=True,  # Log SQL (development only)
    pool_size=20,
    max_overflow=0,
    pool_pre_ping=True,  # Verify connections
    pool_recycle=3600,   # Recycle after 1 hour
)

# ✅ Session factory
async_session_maker = async_sessionmaker(
    engine,
    class_=AsyncSession,
    expire_on_commit=False
)

# ✅ Base model
class Base(DeclarativeBase):
    pass
```

### Models

```python
from sqlalchemy import Column, Integer, String, DateTime, func
from sqlalchemy.orm import Mapped, mapped_column
from datetime import datetime

# ✅ SQLAlchemy 2.0 style models
class User(Base):
    __tablename__ = "users"
    
    id: Mapped[int] = mapped_column(primary_key=True)
    name: Mapped[str] = mapped_column(String(100))
    email: Mapped[str] = mapped_column(String(100), unique=True, index=True)
    password: Mapped[str] = mapped_column(String(255))
    created_at: Mapped[datetime] = mapped_column(DateTime, server_default=func.now())
    updated_at: Mapped[datetime] = mapped_column(DateTime, server_default=func.now(), onupdate=func.now())
```

### Dependency Injection

```python
from typing import AsyncGenerator

# ✅ Database session dependency
async def get_db() -> AsyncGenerator[AsyncSession, None]:
    async with async_session_maker() as session:
        try:
            yield session
            await session.commit()
        except Exception:
            await session.rollback()
            raise
        finally:
            await session.close()

# Usage in FastAPI
@app.get("/users/{user_id}")
async def get_user(user_id: int, db: AsyncSession = Depends(get_db)):
    result = await db.execute(select(User).where(User.id == user_id))
    user = result.scalar_one_or_none()
    if not user:
        raise HTTPException(status_code=404, detail="User not found")
    return user
```

## CRUD Operations

```python
from sqlalchemy import select, update, delete
from sqlalchemy.ext.asyncio import AsyncSession

# ✅ Create
async def create_user(db: AsyncSession, name: str, email: str, password: str) -> User:
    user = User(name=name, email=email, password=password)
    db.add(user)
    await db.flush()  # Get ID without committing
    await db.refresh(user)
    return user

# ✅ Read
async def get_user(db: AsyncSession, user_id: int) -> User | None:
    result = await db.execute(select(User).where(User.id == user_id))
    return result.scalar_one_or_none()

# ✅ Update
async def update_user(db: AsyncSession, user_id: int, name: str) -> User | None:
    stmt = update(User).where(User.id == user_id).values(name=name).returning(User)
    result = await db.execute(stmt)
    await db.commit()
    return result.scalar_one_or_none()

# ✅ Delete
async def delete_user(db: AsyncSession, user_id: int) -> bool:
    stmt = delete(User).where(User.id == user_id)
    result = await db.execute(stmt)
    await db.commit()
    return result.rowcount > 0
```

## N+1 Query Prevention

```python
from sqlalchemy.orm import selectinload, joinedload

# ❌ N+1 problem
users = await db.execute(select(User))
for user in users.scalars():
    posts = await db.execute(select(Post).where(Post.user_id == user.id))  # N queries

# ✅ Eager loading with selectinload
stmt = select(User).options(selectinload(User.posts))
result = await db.execute(stmt)
users = result.scalars().all()  # 1 or 2 queries

# ✅ Eager loading with joinedload (LEFT JOIN)
stmt = select(User).options(joinedload(User.profile))
result = await db.execute(stmt)
users = result.unique().scalars().all()  # 1 query
```

## Pagination

```python
# ✅ Cursor pagination
async def get_users_cursor(db: AsyncSession, cursor: int, limit: int = 20) -> list[User]:
    stmt = select(User).where(User.id > cursor).order_by(User.id).limit(limit)
    result = await db.execute(stmt)
    return result.scalars().all()

# ✅ Offset pagination (small datasets only)
async def get_users_page(db: AsyncSession, page: int, page_size: int = 20) -> tuple[list[User], int]:
    # Get total count
    count_stmt = select(func.count()).select_from(User)
    total = await db.scalar(count_stmt)
    
    # Get page
    stmt = select(User).offset((page - 1) * page_size).limit(page_size)
    result = await db.execute(stmt)
    users = result.scalars().all()
    
    return users, total
```

## Transactions

```python
# ✅ Transaction with context manager
async def transfer_money(db: AsyncSession, from_id: int, to_id: int, amount: float):
    async with db.begin():
        # Deduct from sender
        await db.execute(
            update(Account)
            .where(Account.id == from_id)
            .values(balance=Account.balance - amount)
        )
        
        # Add to receiver
        await db.execute(
            update(Account)
            .where(Account.id == to_id)
            .values(balance=Account.balance + amount)
        )
    # Auto-commit if no exception

# ✅ Manual transaction control
async def manual_transaction(db: AsyncSession):
    await db.begin()
    try:
        # Operations
        await db.commit()
    except Exception:
        await db.rollback()
        raise
```

## Raw SQL

```python
from sqlalchemy import text

# ✅ Raw SQL with parameters (prevent SQL injection)
async def raw_query(db: AsyncSession, email: str) -> list:
    stmt = text("SELECT * FROM users WHERE email = :email")
    result = await db.execute(stmt, {"email": email})
    return result.fetchall()

# ✅ Raw SQL with models
async def raw_query_model(db: AsyncSession) -> list[User]:
    stmt = text("SELECT * FROM users WHERE created_at > :date")
    result = await db.execute(stmt, {"date": "2026-01-01"})
    return [User(**row._mapping) for row in result]
```

## Alembic Migrations

### Setup

```bash
# Install
pip install alembic

# Initialize
alembic init alembic

# Create migration
alembic revision --autogenerate -m "create users table"

# Apply migrations
alembic upgrade head

# Rollback
alembic downgrade -1
```

### alembic.ini

```ini
[alembic]
script_location = alembic
sqlalchemy.url = postgresql+asyncpg://user:pass@localhost/dbname
```

### env.py (Async)

```python
# alembic/env.py
from sqlalchemy.ext.asyncio import create_async_engine
from alembic import context
from app.models import Base

config = context.config
target_metadata = Base.metadata

def run_migrations_offline():
    url = config.get_main_option("sqlalchemy.url")
    context.configure(
        url=url,
        target_metadata=target_metadata,
        literal_binds=True,
        dialect_opts={"paramstyle": "named"},
    )

    with context.begin_transaction():
        context.run_migrations()

async def run_migrations_online():
    connectable = create_async_engine(
        config.get_main_option("sqlalchemy.url"),
        poolclass=pool.NullPool,
    )

    async with connectable.connect() as connection:
        await connection.run_sync(do_run_migrations)

def do_run_migrations(connection):
    context.configure(connection=connection, target_metadata=target_metadata)
    with context.begin_transaction():
        context.run_migrations()

if context.is_offline_mode():
    run_migrations_offline()
else:
    import asyncio
    asyncio.run(run_migrations_online())
```

### Migration File

```python
# alembic/versions/xxx_create_users_table.py
from alembic import op
import sqlalchemy as sa

def upgrade():
    op.create_table(
        'users',
        sa.Column('id', sa.Integer(), nullable=False),
        sa.Column('name', sa.String(100), nullable=False),
        sa.Column('email', sa.String(100), nullable=False),
        sa.Column('password', sa.String(255), nullable=False),
        sa.Column('created_at', sa.DateTime(), server_default=sa.text('now()')),
        sa.PrimaryKeyConstraint('id')
    )
    op.create_index('ix_users_email', 'users', ['email'], unique=True)

def downgrade():
    op.drop_index('ix_users_email', table_name='users')
    op.drop_table('users')
```

## Connection Pool Management

```python
# ✅ Check pool status
async def get_pool_status():
    pool = engine.pool
    return {
        "size": pool.size(),
        "checked_in": pool.checkedin(),
        "checked_out": pool.checkedout(),
        "overflow": pool.overflow(),
    }

# ✅ Dispose connections (for testing)
await engine.dispose()
```

## Best Practices

```python
# ✅ Always use context managers
async with async_session_maker() as session:
    # Operations
    await session.commit()

# ✅ Use parameters (prevent SQL injection)
stmt = select(User).where(User.email == email)  # ✅ Safe
stmt = text(f"SELECT * FROM users WHERE email = '{email}'")  # ❌ Dangerous

# ✅ Handle exceptions
try:
    result = await db.execute(stmt)
except IntegrityError:
    raise HTTPException(status_code=409, detail="Email already exists")

# ✅ Use indexes
class User(Base):
    __tablename__ = "users"
    
    email: Mapped[str] = mapped_column(String(100), index=True)  # ✅ Indexed

# ✅ Expire on commit = False for async
async_session_maker = async_sessionmaker(
    engine,
    expire_on_commit=False  # Important for async
)
```

---

**References**:
- [SQLAlchemy 2.0 Documentation](https://docs.sqlalchemy.org/)
- [SQLAlchemy Async](https://docs.sqlalchemy.org/en/20/orm/extensions/asyncio.html)
- [asyncpg](https://magicstack.github.io/asyncpg/)
- [Alembic](https://alembic.sqlalchemy.org/)
