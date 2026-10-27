# Async Programming

Async/await patterns and best practices for Python asynchronous code.

## Quick Reference

- **async def** - define async functions
- **await** - call async functions
- **asyncio.gather** - run tasks concurrently
- **asyncio.create_task** - schedule coroutine
- **async with** - async context managers
- **async for** - async iterators

## Core Rules

### 1. Async Function Definition

```python
# ✅ Async function
async def fetch_user(user_id: int) -> User:
    async with httpx.AsyncClient() as client:
        response = await client.get(f"/users/{user_id}")
        return User(**response.json())

# ✅ Must await async function
async def main():
    user = await fetch_user(1)  # Correct

# ❌ Forgot await
async def main():
    user = fetch_user(1)  # Wrong: returns coroutine, not User
```

### 2. Running Async Code

```python
import asyncio

# ✅ Run from sync code
def main():
    result = asyncio.run(async_main())

async def async_main():
    user = await fetch_user(1)
    return user

# ✅ In Jupyter/IPython
await fetch_user(1)  # Can await directly

# ❌ Don't mix sync and async incorrectly
def sync_function():
    user = await fetch_user(1)  # Error: await outside async
```

### 3. Concurrent Execution

```python
# ✅ Use asyncio.gather for concurrent tasks
async def fetch_multiple_users(user_ids: list[int]) -> list[User]:
    tasks = [fetch_user(user_id) for user_id in user_ids]
    return await asyncio.gather(*tasks)

# ✅ Use asyncio.create_task
async def process_in_background():
    task1 = asyncio.create_task(fetch_user(1))
    task2 = asyncio.create_task(fetch_user(2))
    
    # Do other work
    await other_work()
    
    # Wait for tasks
    user1 = await task1
    user2 = await task2

# ❌ Sequential (slow)
async def fetch_multiple_users(user_ids: list[int]) -> list[User]:
    users = []
    for user_id in user_ids:
        user = await fetch_user(user_id)  # Waits for each
        users.append(user)
    return users
```

### 4. Error Handling

```python
# ✅ Try/except with async
async def safe_fetch_user(user_id: int) -> User | None:
    try:
        return await fetch_user(user_id)
    except httpx.HTTPError as e:
        logger.error(f"Failed to fetch user {user_id}: {e}")
        return None

# ✅ Handle gather errors
async def fetch_multiple_with_errors(user_ids: list[int]) -> list[User | None]:
    tasks = [safe_fetch_user(user_id) for user_id in user_ids]
    return await asyncio.gather(*tasks, return_exceptions=False)

# ✅ Partial failure handling
async def fetch_with_fallback(user_ids: list[int]) -> list[User]:
    results = await asyncio.gather(
        *[fetch_user(uid) for uid in user_ids],
        return_exceptions=True
    )
    return [r for r in results if isinstance(r, User)]
```

### 5. Async Context Managers

```python
# ✅ Async with
async def make_request():
    async with httpx.AsyncClient() as client:
        response = await client.get("/api/data")
        return response.json()

# ✅ Multiple async context managers
async def process_data():
    async with get_db() as db, get_redis() as redis:
        data = await db.fetch()
        await redis.cache(data)

# ✅ Custom async context manager
class AsyncDatabase:
    async def __aenter__(self):
        self.conn = await connect_db()
        return self.conn
    
    async def __aexit__(self, exc_type, exc_val, exc_tb):
        await self.conn.close()

async def use_db():
    async with AsyncDatabase() as db:
        await db.query("SELECT * FROM users")
```

### 6. Async Iterators

```python
# ✅ Async for loop
async def process_stream():
    async for item in fetch_stream():
        await process_item(item)

# ✅ Async generator
async def fetch_pages(url: str):
    page = 1
    while True:
        response = await fetch_page(url, page)
        if not response:
            break
        yield response
        page += 1

async def main():
    async for page in fetch_pages("/api/data"):
        process_page(page)

# ✅ Async comprehension
async def get_all_data():
    return [item async for item in fetch_stream()]
```

### 7. Timeouts

```python
import asyncio

# ✅ Timeout for single operation
async def fetch_with_timeout(url: str) -> dict:
    try:
        async with asyncio.timeout(5.0):  # Python 3.11+
            return await fetch_data(url)
    except asyncio.TimeoutError:
        logger.error(f"Timeout fetching {url}")
        raise

# ✅ Timeout for Python 3.10 and earlier
async def fetch_with_timeout_old(url: str) -> dict:
    try:
        return await asyncio.wait_for(fetch_data(url), timeout=5.0)
    except asyncio.TimeoutError:
        logger.error(f"Timeout fetching {url}")
        raise
```

## FastAPI Async Patterns

### Async Dependencies

```python
# ✅ Async dependency
async def get_db() -> AsyncGenerator[AsyncSession, None]:
    async with get_async_session() as session:
        yield session

@app.get("/users/{user_id}")
async def get_user(
    user_id: int,
    db: AsyncSession = Depends(get_db)
) -> User:
    result = await db.execute(select(User).where(User.id == user_id))
    return result.scalar_one_or_none()
```

### Background Tasks

```python
from fastapi import BackgroundTasks

# ✅ Async background task
async def send_email(email: str, message: str):
    await email_service.send(email, message)

@app.post("/users")
async def create_user(
    user: UserCreate,
    background_tasks: BackgroundTasks
):
    new_user = await db.create_user(user)
    background_tasks.add_task(send_email, new_user.email, "Welcome!")
    return new_user
```

## Common Patterns

### Retry with Exponential Backoff

```python
async def retry_with_backoff(
    func: Callable,
    max_retries: int = 3,
    base_delay: float = 1.0
):
    for attempt in range(max_retries):
        try:
            return await func()
        except Exception as e:
            if attempt == max_retries - 1:
                raise
            delay = base_delay * (2 ** attempt)
            await asyncio.sleep(delay)
```

### Rate Limiting

```python
import asyncio
from collections import deque
from time import time

class RateLimiter:
    def __init__(self, rate: int, period: float = 1.0):
        self.rate = rate
        self.period = period
        self.timestamps = deque()
    
    async def acquire(self):
        while len(self.timestamps) >= self.rate:
            if time() - self.timestamps[0] > self.period:
                self.timestamps.popleft()
            else:
                await asyncio.sleep(0.01)
        self.timestamps.append(time())

limiter = RateLimiter(rate=10, period=1.0)

async def make_request(url: str):
    await limiter.acquire()
    return await fetch(url)
```

### Connection Pooling

```python
# ✅ Reuse client session
class APIClient:
    def __init__(self):
        self._client: httpx.AsyncClient | None = None
    
    async def __aenter__(self):
        self._client = httpx.AsyncClient()
        return self
    
    async def __aexit__(self, *args):
        if self._client:
            await self._client.aclose()
    
    async def get(self, url: str) -> dict:
        if not self._client:
            raise RuntimeError("Client not initialized")
        response = await self._client.get(url)
        return response.json()

async with APIClient() as client:
    user1 = await client.get("/users/1")
    user2 = await client.get("/users/2")  # Reuses connection
```

## Common Mistakes

### Mistake 1: Mixing Sync and Async

```python
# ❌ Blocking call in async function
async def fetch_data():
    response = requests.get("/api/data")  # Blocks event loop!
    return response.json()

# ✅ Use async library
async def fetch_data():
    async with httpx.AsyncClient() as client:
        response = await client.get("/api/data")
        return response.json()
```

### Mistake 2: Not Awaiting Coroutines

```python
# ❌ Forgot await
async def main():
    user = fetch_user(1)  # Returns coroutine, not User
    print(user.name)  # AttributeError

# ✅ Await the coroutine
async def main():
    user = await fetch_user(1)
    print(user.name)
```

### Mistake 3: Creating Too Many Tasks

```python
# ❌ Unlimited concurrent tasks
async def process_all(items: list):
    tasks = [process_item(item) for item in items]  # 10,000 tasks!
    await asyncio.gather(*tasks)

# ✅ Limit concurrency with semaphore
async def process_all(items: list, max_concurrent: int = 10):
    semaphore = asyncio.Semaphore(max_concurrent)
    
    async def process_with_limit(item):
        async with semaphore:
            return await process_item(item)
    
    tasks = [process_with_limit(item) for item in items]
    return await asyncio.gather(*tasks)
```

### Mistake 4: Not Handling Cancellation

```python
# ❌ Doesn't handle cancellation
async def long_running_task():
    for i in range(1000):
        await process(i)  # Can't be cancelled mid-loop

# ✅ Check cancellation
async def long_running_task():
    try:
        for i in range(1000):
            if i % 10 == 0:
                await asyncio.sleep(0)  # Cancellation point
            await process(i)
    except asyncio.CancelledError:
        logger.info("Task cancelled, cleaning up")
        await cleanup()
        raise
```

---

**Full documentation**: [asyncio — Asynchronous I/O](https://docs.python.org/3/library/asyncio.html)
