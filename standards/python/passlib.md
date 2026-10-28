---
language: python
category: dependency
applies_to: [api]
---

# Security

Python-specific security implementations.

Based on [passlib](https://passlib.readthedocs.io/), [PyJWT](https://pyjwt.readthedocs.io/), and [cryptography](https://cryptography.io/).

## Quick Reference

- **Argon2** - password hashing (passlib or argon2-cffi)
- **PyJWT** - JWT implementation
- **cryptography** - RSA key generation
- **httpx** - async HTTP client with timeouts
- **slowapi** - rate limiting for FastAPI

## Password Hashing

### Argon2 with passlib

```python
from passlib.hash import argon2

# ✅ Hash password with Argon2id
def hash_password(password: str) -> str:
    return argon2.using(
        type="ID",           # Argon2id
        memory_cost=65536,   # 64 MB
        time_cost=3,         # 3 iterations
        parallelism=4,       # 4 threads
        salt_len=16,         # 16 bytes salt
        hash_len=32          # 32 bytes output
    ).hash(password)

# ✅ Verify password
def verify_password(password: str, hashed: str) -> bool:
    try:
        return argon2.verify(password, hashed)
    except Exception:
        return False

# ✅ Check if rehash needed (upgrade old hashes)
def needs_rehash(hashed: str) -> bool:
    return argon2.using(
        type="ID",
        memory_cost=65536,
        time_cost=3,
        parallelism=4
    ).needs_update(hashed)
```

### Usage

```python
# ✅ On signup
async def create_user(user: UserCreate, db: AsyncSession):
    hashed_password = hash_password(user.password)
    db_user = User(
        name=user.name,
        email=user.email,
        password=hashed_password
    )
    db.add(db_user)
    await db.commit()
    return db_user

# ✅ On login with rehash
async def login(credentials: LoginRequest, db: AsyncSession):
    user = await get_user_by_email(db, credentials.email)
    if not user:
        raise HTTPException(status_code=401, detail="Invalid credentials")
    
    if not verify_password(credentials.password, user.password):
        raise HTTPException(status_code=401, detail="Invalid credentials")
    
    # Upgrade old hash if needed
    if needs_rehash(user.password):
        user.password = hash_password(credentials.password)
        await db.commit()
    
    return user
```

### Alternative: argon2-cffi

```python
from argon2 import PasswordHasher
from argon2.exceptions import VerifyMismatchError

# ✅ Argon2 with argon2-cffi
ph = PasswordHasher(
    time_cost=3,
    memory_cost=65536,  # 64 MB
    parallelism=4,
    hash_len=32,
    salt_len=16
)

def hash_password(password: str) -> str:
    return ph.hash(password)

def verify_password(password: str, hashed: str) -> bool:
    try:
        ph.verify(hashed, password)
        return True
    except VerifyMismatchError:
        return False
```

## JWT Authentication

### RS256 Implementation

```python
import jwt
from datetime import datetime, timedelta, timezone
from cryptography.hazmat.primitives import serialization
from cryptography.hazmat.primitives.asymmetric import rsa
from cryptography.hazmat.backends import default_backend
import uuid

class TokenService:
    def __init__(self, private_key: str, public_key: str, issuer: str, audience: str):
        self.private_key = serialization.load_pem_private_key(
            private_key.encode(),
            password=None,
            backend=default_backend()
        )
        self.public_key = serialization.load_pem_public_key(
            public_key.encode(),
            backend=default_backend()
        )
        self.issuer = issuer
        self.audience = audience
    
    # ✅ Generate access token (15 minutes)
    def generate_access_token(self, user_id: str) -> str:
        now = datetime.now(timezone.utc)
        payload = {
            "iss": self.issuer,
            "aud": self.audience,
            "sub": user_id,
            "exp": now + timedelta(minutes=15),
            "iat": now,
            "nbf": now,
            "jti": str(uuid.uuid4())  # Unique token ID
        }
        return jwt.encode(payload, self.private_key, algorithm="RS256")
    
    # ✅ Generate refresh token (7 days)
    def generate_refresh_token(self, user_id: str) -> str:
        now = datetime.now(timezone.utc)
        payload = {
            "iss": self.issuer,
            "aud": self.audience,
            "sub": user_id,
            "exp": now + timedelta(days=7),
            "iat": now,
            "nbf": now,
            "jti": str(uuid.uuid4())
        }
        return jwt.encode(payload, self.private_key, algorithm="RS256")
    
    # ✅ Validate token
    def validate_token(self, token: str) -> dict:
        try:
            payload = jwt.decode(
                token,
                self.public_key,
                algorithms=["RS256"],
                issuer=self.issuer,
                audience=self.audience,
                options={"require": ["exp", "iss", "aud", "sub"]}
            )
            return payload
        except jwt.ExpiredSignatureError:
            raise ValueError("Token expired")
        except jwt.InvalidTokenError:
            raise ValueError("Invalid token")
```

### Generate RSA Keys

```python
# ✅ Generate RSA key pair (run once)
def generate_rsa_keys():
    private_key = rsa.generate_private_key(
        public_exponent=65537,
        key_size=2048,
        backend=default_backend()
    )
    
    private_pem = private_key.private_bytes(
        encoding=serialization.Encoding.PEM,
        format=serialization.PrivateFormat.PKCS8,
        encryption_algorithm=serialization.NoEncryption()
    )
    
    public_key = private_key.public_key()
    public_pem = public_key.public_bytes(
        encoding=serialization.Encoding.PEM,
        format=serialization.PublicFormat.SubjectPublicKeyInfo
    )
    
    return private_pem.decode(), public_pem.decode()

# Save to files
private_key, public_key = generate_rsa_keys()
with open("private_key.pem", "w") as f:
    f.write(private_key)
with open("public_key.pem", "w") as f:
    f.write(public_key)
```

### FastAPI Dependency

```python
from fastapi import Depends, HTTPException, status
from fastapi.security import HTTPBearer, HTTPAuthorizationCredentials

security = HTTPBearer()

async def get_current_user(
    credentials: HTTPAuthorizationCredentials = Depends(security),
    token_service: TokenService = Depends(get_token_service)
) -> str:
    try:
        payload = token_service.validate_token(credentials.credentials)
        return payload["sub"]  # user_id
    except ValueError as e:
        raise HTTPException(
            status_code=status.HTTP_401_UNAUTHORIZED,
            detail=str(e)
        )

# Usage
@app.get("/me")
async def get_me(user_id: str = Depends(get_current_user)):
    return {"user_id": user_id}
```

## Timeout Configuration

### httpx with Timeouts

```python
import httpx

# ✅ Async HTTP client with timeouts
async def make_request(url: str) -> dict:
    timeout = httpx.Timeout(
        connect=5.0,   # Connection timeout
        read=30.0,     # Read timeout
        write=30.0,    # Write timeout
        pool=5.0       # Pool timeout
    )
    
    async with httpx.AsyncClient(timeout=timeout) as client:
        response = await client.get(url)
        response.raise_for_status()
        return response.json()

# ✅ Context timeout
import asyncio

async def fetch_with_timeout(url: str):
    try:
        async with asyncio.timeout(10):  # Python 3.11+
            return await make_request(url)
    except asyncio.TimeoutError:
        raise HTTPException(status_code=504, detail="Request timeout")
```

## Rate Limiting

### slowapi for FastAPI

```python
from slowapi import Limiter, _rate_limit_exceeded_handler
from slowapi.util import get_remote_address
from slowapi.errors import RateLimitExceeded

# ✅ Setup rate limiter
limiter = Limiter(key_func=get_remote_address)
app = FastAPI()
app.state.limiter = limiter
app.add_exception_handler(RateLimitExceeded, _rate_limit_exceeded_handler)

# ✅ Apply rate limit
@app.get("/api/data")
@limiter.limit("100/minute")
async def get_data(request: Request):
    return {"data": "..."}

# ✅ Different limits for different endpoints
@app.post("/api/login")
@limiter.limit("5/15minutes")
async def login(request: Request, credentials: LoginRequest):
    return {"token": "..."}

# ✅ Per-user rate limit
def get_user_id(request: Request) -> str:
    # Extract from token
    return request.state.user_id

user_limiter = Limiter(key_func=get_user_id)

@app.get("/api/user/data")
@user_limiter.limit("1000/hour")
async def get_user_data(request: Request, user_id: str = Depends(get_current_user)):
    return {"data": "..."}
```

## Input Validation

```python
from pydantic import BaseModel, Field, field_validator

# ✅ Pydantic validation
class UserCreate(BaseModel):
    name: str = Field(min_length=1, max_length=100)
    email: str = Field(pattern=r"^[\w\.-]+@[\w\.-]+\.\w+$")
    age: int = Field(ge=0, le=150)

    @field_validator("email")
    @classmethod
    def validate_email(cls, v: str) -> str:
        # Additional validation
        if "+" in v:
            raise ValueError("Email cannot contain +")
        return v.lower()
```

## CORS Configuration

```python
from fastapi.middleware.cors import CORSMiddleware

# ✅ Strict CORS
app.add_middleware(
    CORSMiddleware,
    allow_origins=["https://app.example.com"],  # Specific domain
    allow_credentials=True,
    allow_methods=["GET", "POST", "PUT", "DELETE"],
    allow_headers=["Authorization", "Content-Type"],
    max_age=86400
)

# ❌ Dangerous CORS
app.add_middleware(
    CORSMiddleware,
    allow_origins=["*"],  # WRONG: allows all domains
)
```

---

**References**:
- [passlib Documentation](https://passlib.readthedocs.io/)
- [PyJWT Documentation](https://pyjwt.readthedocs.io/)
- [cryptography](https://cryptography.io/)
- [slowapi](https://github.com/laurents/slowapi)
