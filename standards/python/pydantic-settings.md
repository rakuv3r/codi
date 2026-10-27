# Configuration

Python-specific configuration management patterns.

Based on [pydantic-settings](https://docs.pydantic.dev/latest/concepts/pydantic_settings/) and environment variables.

## Quick Reference

- **pydantic-settings** - type-safe configuration (recommended)
- **Environment variables** - 12-factor app standard
- **python-dotenv** - .env file support
- **Validation** - automatic with Pydantic
- **No hardcoded values** - all config from external sources

## pydantic-settings (Recommended)

### Basic Setup

```python
from pydantic_settings import BaseSettings, SettingsConfigDict
from pydantic import Field

# ✅ Type-safe configuration
class Settings(BaseSettings):
    model_config = SettingsConfigDict(
        env_file=".env",
        env_file_encoding="utf-8",
        case_sensitive=False
    )
    
    # Server
    server_port: int = 8000
    server_host: str = "0.0.0.0"
    
    # Database
    database_url: str
    db_pool_size: int = 20
    
    # JWT
    jwt_secret: str
    jwt_access_expiration: int = 900  # 15 minutes
    jwt_refresh_expiration: int = 604800  # 7 days
    
    # Environment
    environment: str = "development"

# ✅ Load settings
settings = Settings()
```

### Nested Configuration

```python
from pydantic import BaseModel

# ✅ Nested settings
class DatabaseSettings(BaseModel):
    url: str
    pool_size: int = 20
    pool_timeout: int = 30
    echo: bool = False

class JWTSettings(BaseModel):
    secret: str
    algorithm: str = "RS256"
    access_expiration: int = 900
    refresh_expiration: int = 604800

class Settings(BaseSettings):
    model_config = SettingsConfigDict(
        env_file=".env",
        env_nested_delimiter="__"  # DATABASE__URL
    )
    
    database: DatabaseSettings
    jwt: JWTSettings
    log_level: str = "INFO"

# Environment variables:
# DATABASE__URL=postgres://...
# DATABASE__POOL_SIZE=30
# JWT__SECRET=supersecret
```

### Field Validation

```python
from pydantic import field_validator, model_validator

class Settings(BaseSettings):
    server_port: int = 8000
    database_url: str
    jwt_secret: str
    environment: str = "development"
    
    @field_validator("server_port")
    @classmethod
    def validate_port(cls, v: int) -> int:
        if not 1 <= v <= 65535:
            raise ValueError("Port must be between 1 and 65535")
        return v
    
    @field_validator("environment")
    @classmethod
    def validate_environment(cls, v: str) -> str:
        allowed = ["development", "staging", "production"]
        if v not in allowed:
            raise ValueError(f"Environment must be one of {allowed}")
        return v
    
    @model_validator(mode="after")
    def validate_production(self) -> "Settings":
        if self.environment == "production":
            if not self.jwt_secret or len(self.jwt_secret) < 32:
                raise ValueError("JWT secret must be at least 32 characters in production")
        return self
```

## Environment Variables

### .env File

```bash
# .env (never commit!)
DATABASE_URL=postgresql+asyncpg://user:pass@localhost/mydb
JWT_SECRET=supersecretkeythatisatleast32characterslong
ENVIRONMENT=development
SERVER_PORT=8000

# Nested with delimiter
DATABASE__URL=postgresql+asyncpg://...
DATABASE__POOL_SIZE=30
JWT__SECRET=supersecret
JWT__ACCESS_EXPIRATION=900
```

### python-dotenv

```python
from dotenv import load_dotenv
import os

# ✅ Load .env file
load_dotenv()  # Automatically finds .env

# ✅ Override with specific file
load_dotenv(".env.production", override=True)

# Access
db_url = os.getenv("DATABASE_URL")
```

## Secrets Management

### AWS Secrets Manager

```python
import boto3
import json
from pydantic_settings import BaseSettings

class Settings(BaseSettings):
    database_url: str
    jwt_secret: str
    
    @classmethod
    def from_aws_secrets(cls, secret_name: str, region: str = "us-east-1"):
        client = boto3.client("secretsmanager", region_name=region)
        response = client.get_secret_value(SecretId=secret_name)
        secrets = json.loads(response["SecretString"])
        return cls(**secrets)

# Usage
settings = Settings.from_aws_secrets("myapp/production")
```

### Environment Variables (Production)

```python
# ✅ Production: use environment variables
class Settings(BaseSettings):
    model_config = SettingsConfigDict(
        env_file=".env" if os.getenv("ENVIRONMENT") != "production" else None,
        case_sensitive=False
    )
    
    database_url: str
    jwt_secret: str
```

## Dependency Injection (FastAPI)

```python
from functools import lru_cache
from fastapi import Depends

# ✅ Cached settings
@lru_cache
def get_settings() -> Settings:
    return Settings()

# ✅ Use as dependency
@app.get("/config")
async def get_config(settings: Settings = Depends(get_settings)):
    return {
        "environment": settings.environment,
        "log_level": settings.log_level
    }

# ✅ Override in tests
def get_test_settings():
    return Settings(
        database_url="sqlite+aiosqlite:///:memory:",
        jwt_secret="test-secret",
        environment="test"
    )

app.dependency_overrides[get_settings] = get_test_settings
```

## Multiple Environments

### Environment-Specific Files

```python
import os

# ✅ Load based on environment
env = os.getenv("ENVIRONMENT", "development")
env_file = f".env.{env}"

class Settings(BaseSettings):
    model_config = SettingsConfigDict(
        env_file=env_file,
        env_file_encoding="utf-8"
    )
    
    database_url: str
    jwt_secret: str
```

```bash
# Project structure
.env.development
.env.staging
.env.production
```

### Inheritance

```python
# ✅ Base + environment-specific
class BaseSettings(BaseSettings):
    server_port: int = 8000
    log_level: str = "INFO"

class DevelopmentSettings(BaseSettings):
    model_config = SettingsConfigDict(env_file=".env.development")
    debug: bool = True
    log_level: str = "DEBUG"

class ProductionSettings(BaseSettings):
    model_config = SettingsConfigDict(env_file=".env.production")
    debug: bool = False
    log_level: str = "INFO"

def get_settings() -> BaseSettings:
    env = os.getenv("ENVIRONMENT", "development")
    if env == "production":
        return ProductionSettings()
    return DevelopmentSettings()
```

## Feature Flags

```python
class FeatureFlags(BaseModel):
    enable_new_ui: bool = False
    enable_rate_limiting: bool = True
    max_upload_size_mb: int = 10

class Settings(BaseSettings):
    features: FeatureFlags = Field(default_factory=FeatureFlags)

# Usage
settings = Settings()
if settings.features.enable_new_ui:
    # Use new UI
    pass
```

## Type Conversion

```python
from pydantic import HttpUrl, PostgresDsn, Field
from typing import Literal

class Settings(BaseSettings):
    # ✅ URL validation
    database_url: PostgresDsn
    api_url: HttpUrl
    
    # ✅ List from comma-separated string
    allowed_hosts: list[str] = Field(default_factory=list)
    
    # ✅ Literal types
    log_level: Literal["DEBUG", "INFO", "WARNING", "ERROR"] = "INFO"
    
    # ✅ Path
    from pathlib import Path
    log_file: Path = Path("app.log")

# Environment:
# DATABASE_URL=postgresql://...
# API_URL=https://api.example.com
# ALLOWED_HOSTS=localhost,127.0.0.1,example.com
# LOG_LEVEL=INFO
```

## Secrets in Files

```python
from pydantic import SecretStr

class Settings(BaseSettings):
    # ✅ SecretStr hides value in logs
    jwt_secret: SecretStr
    database_password: SecretStr
    
    def get_jwt_secret(self) -> str:
        return self.jwt_secret.get_secret_value()

settings = Settings()
print(settings.jwt_secret)  # **********
print(settings.get_jwt_secret())  # actual-secret-value
```

## Complete Example

```python
from pydantic_settings import BaseSettings, SettingsConfigDict
from pydantic import Field, PostgresDsn, field_validator, SecretStr
from typing import Literal
import os

class DatabaseSettings(BaseSettings):
    url: PostgresDsn
    pool_size: int = 20
    pool_max_overflow: int = 10
    pool_timeout: int = 30
    echo: bool = False

class JWTSettings(BaseSettings):
    secret: SecretStr
    algorithm: str = "RS256"
    access_expiration_minutes: int = 15
    refresh_expiration_days: int = 7

class Settings(BaseSettings):
    model_config = SettingsConfigDict(
        env_file=".env",
        env_file_encoding="utf-8",
        env_nested_delimiter="__",
        case_sensitive=False
    )
    
    # Server
    environment: Literal["development", "staging", "production"] = "development"
    server_host: str = "0.0.0.0"
    server_port: int = 8000
    debug: bool = False
    
    # Database
    database: DatabaseSettings
    
    # JWT
    jwt: JWTSettings
    
    # Logging
    log_level: Literal["DEBUG", "INFO", "WARNING", "ERROR"] = "INFO"
    
    # CORS
    allowed_origins: list[str] = Field(default_factory=lambda: ["http://localhost:3000"])
    
    @field_validator("debug", mode="after")
    @classmethod
    def set_debug_from_env(cls, v: bool, info) -> bool:
        if info.data.get("environment") == "production":
            return False
        return v

# Usage
settings = Settings()

# FastAPI
from functools import lru_cache

@lru_cache
def get_settings() -> Settings:
    return Settings()
```

## Best Practices

```python
# ✅ Use pydantic-settings
from pydantic_settings import BaseSettings

# ✅ Validate on startup
try:
    settings = Settings()
except ValidationError as e:
    print(f"Configuration error: {e}")
    sys.exit(1)

# ✅ Use dependency injection
@app.get("/status")
async def status(settings: Settings = Depends(get_settings)):
    return {"environment": settings.environment}

# ❌ Don't hardcode values
DATABASE_URL = "postgresql://localhost/mydb"  # WRONG

# ✅ All config from environment
database_url: str  # Required from environment

# ❌ Don't commit secrets
# .env should be in .gitignore

# ✅ Use .env.example
# .env.example (committed)
DATABASE_URL=postgresql://user:pass@localhost/dbname
JWT_SECRET=your-secret-here
```

---

**References**:
- [pydantic-settings](https://docs.pydantic.dev/latest/concepts/pydantic_settings/)
- [python-dotenv](https://github.com/theskumar/python-dotenv)
- [12-Factor App Config](https://12factor.net/config)
