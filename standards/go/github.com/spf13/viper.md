---
language: go
category: dependency
applies_to: []
---

# Configuration

Go-specific configuration management patterns.

Based on [viper](https://github.com/spf13/viper) and environment variables.

## Quick Reference

- **Environment variables** - 12-factor app standard
- **viper** - configuration management library
- **Validation** - validate config on startup
- **Secrets** - use environment variables or secret management
- **No hardcoded values** - all config from external sources

## Environment Variables

### Using Standard Library

```go
import (
    "os"
    "strconv"
)

// ✅ Simple config struct
type Config struct {
    ServerPort     int
    DatabaseURL    string
    LogLevel       string
    JWTSecret      string
    Environment    string
}

// ✅ Load from environment
func LoadConfig() (*Config, error) {
    config := &Config{
        ServerPort:  getEnvAsInt("SERVER_PORT", 8080),
        DatabaseURL: getEnv("DATABASE_URL", ""),
        LogLevel:    getEnv("LOG_LEVEL", "info"),
        JWTSecret:   getEnv("JWT_SECRET", ""),
        Environment: getEnv("ENVIRONMENT", "development"),
    }

    // Validate required fields
    if config.DatabaseURL == "" {
        return nil, errors.New("DATABASE_URL is required")
    }
    if config.JWTSecret == "" && config.Environment == "production" {
        return nil, errors.New("JWT_SECRET is required in production")
    }

    return config, nil
}

func getEnv(key, defaultValue string) string {
    if value := os.Getenv(key); value != "" {
        return value
    }
    return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
    if value := os.Getenv(key); value != "" {
        if intValue, err := strconv.Atoi(value); err == nil {
            return intValue
        }
    }
    return defaultValue
}
```

### .env File Support

```go
// ✅ Load .env file (development only)
import "github.com/joho/godotenv"

func init() {
    if err := godotenv.Load(); err != nil {
        // OK if .env doesn't exist in production
        log.Println("No .env file found")
    }
}
```

## Viper

### Setup

```go
import (
    "github.com/spf13/viper"
)

// ✅ Viper configuration
func LoadConfigWithViper() (*Config, error) {
    viper.SetConfigName("config")        // config.yaml
    viper.SetConfigType("yaml")
    viper.AddConfigPath(".")             // Current directory
    viper.AddConfigPath("./config")      // ./config directory
    viper.AddConfigPath("$HOME/.myapp")  // Home directory

    // Environment variables take precedence
    viper.AutomaticEnv()
    viper.SetEnvPrefix("APP") // APP_SERVER_PORT

    // Set defaults
    viper.SetDefault("server.port", 8080)
    viper.SetDefault("log.level", "info")

    // Read config file
    if err := viper.ReadInConfig(); err != nil {
        if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
            return nil, err
        }
        // Config file not found, use defaults + env vars
    }

    // Unmarshal to struct
    var config Config
    if err := viper.Unmarshal(&config); err != nil {
        return nil, err
    }

    return &config, nil
}
```

### Configuration Struct

```go
// ✅ Nested config with tags
type Config struct {
    Server   ServerConfig   `mapstructure:"server"`
    Database DatabaseConfig `mapstructure:"database"`
    JWT      JWTConfig      `mapstructure:"jwt"`
    Log      LogConfig      `mapstructure:"log"`
}

type ServerConfig struct {
    Port            int    `mapstructure:"port"`
    Host            string `mapstructure:"host"`
    ReadTimeout     int    `mapstructure:"read_timeout"`
    WriteTimeout    int    `mapstructure:"write_timeout"`
    ShutdownTimeout int    `mapstructure:"shutdown_timeout"`
}

type DatabaseConfig struct {
    URL             string `mapstructure:"url"`
    MaxOpenConns    int    `mapstructure:"max_open_conns"`
    MaxIdleConns    int    `mapstructure:"max_idle_conns"`
    ConnMaxLifetime int    `mapstructure:"conn_max_lifetime"`
}

type JWTConfig struct {
    Secret            string `mapstructure:"secret"`
    AccessExpiration  int    `mapstructure:"access_expiration"`
    RefreshExpiration int    `mapstructure:"refresh_expiration"`
}

type LogConfig struct {
    Level  string `mapstructure:"level"`
    Format string `mapstructure:"format"` // json or text
}
```

### YAML Configuration

```yaml
# config.yaml
server:
  port: 8080
  host: 0.0.0.0
  read_timeout: 30
  write_timeout: 30
  shutdown_timeout: 10

database:
  url: ${DATABASE_URL} # Environment variable interpolation
  max_open_conns: 20
  max_idle_conns: 5
  conn_max_lifetime: 1800

jwt:
  secret: ${JWT_SECRET}
  access_expiration: 900    # 15 minutes
  refresh_expiration: 604800 # 7 days

log:
  level: info
  format: json
```

## Validation

```go
// ✅ Validate configuration
func (c *Config) Validate() error {
    if c.Server.Port < 1 || c.Server.Port > 65535 {
        return errors.New("invalid server port")
    }
    
    if c.Database.URL == "" {
        return errors.New("database URL is required")
    }
    
    if c.JWT.Secret == "" {
        return errors.New("JWT secret is required")
    }
    
    validLevels := map[string]bool{"debug": true, "info": true, "warn": true, "error": true}
    if !validLevels[c.Log.Level] {
        return errors.New("invalid log level")
    }
    
    return nil
}

// Load and validate
config, err := LoadConfig()
if err != nil {
    log.Fatal(err)
}

if err := config.Validate(); err != nil {
    log.Fatal("invalid configuration:", err)
}
```

## Secrets Management

### Environment Variables (Development)

```bash
# .env (never commit!)
DATABASE_URL=postgres://user:pass@localhost:5432/mydb
JWT_SECRET=supersecretkey
API_KEY=abc123
```

### AWS Secrets Manager

```go
import (
    "github.com/aws/aws-sdk-go-v2/service/secretsmanager"
)

// ✅ Load secrets from AWS
func LoadSecrets(ctx context.Context, secretName string) (map[string]string, error) {
    cfg, err := config.LoadDefaultConfig(ctx)
    if err != nil {
        return nil, err
    }

    client := secretsmanager.NewFromConfig(cfg)
    result, err := client.GetSecretValue(ctx, &secretsmanager.GetSecretValueInput{
        SecretId: aws.String(secretName),
    })
    if err != nil {
        return nil, err
    }

    var secrets map[string]string
    if err := json.Unmarshal([]byte(*result.SecretString), &secrets); err != nil {
        return nil, err
    }

    return secrets, nil
}
```

### HashiCorp Vault

```go
import "github.com/hashicorp/vault/api"

// ✅ Load secrets from Vault
func LoadVaultSecrets(vaultAddr, token, path string) (map[string]interface{}, error) {
    config := api.DefaultConfig()
    config.Address = vaultAddr

    client, err := api.NewClient(config)
    if err != nil {
        return nil, err
    }

    client.SetToken(token)

    secret, err := client.Logical().Read(path)
    if err != nil {
        return nil, err
    }

    return secret.Data, nil
}
```

## Configuration Per Environment

```go
// ✅ Environment-specific config
func LoadConfig() (*Config, error) {
    env := os.Getenv("ENVIRONMENT")
    if env == "" {
        env = "development"
    }

    viper.SetConfigName(fmt.Sprintf("config.%s", env))
    viper.SetConfigType("yaml")
    viper.AddConfigPath("./config")

    if err := viper.ReadInConfig(); err != nil {
        return nil, err
    }

    var config Config
    if err := viper.Unmarshal(&config); err != nil {
        return nil, err
    }

    return &config, nil
}
```

```
config/
├── config.development.yaml
├── config.staging.yaml
└── config.production.yaml
```

## Feature Flags

```go
// ✅ Simple feature flags
type FeatureFlags struct {
    EnableNewUI      bool `mapstructure:"enable_new_ui"`
    EnableRateLimiting bool `mapstructure:"enable_rate_limiting"`
    MaxUploadSize    int  `mapstructure:"max_upload_size"`
}

// Usage
if config.Features.EnableNewUI {
    // Use new UI
}
```

## Best Practices

```go
// ✅ Global config singleton
var globalConfig *Config

func InitConfig() error {
    config, err := LoadConfig()
    if err != nil {
        return err
    }
    
    if err := config.Validate(); err != nil {
        return err
    }
    
    globalConfig = config
    return nil
}

func GetConfig() *Config {
    return globalConfig
}

// ✅ Use in main
func main() {
    if err := InitConfig(); err != nil {
        log.Fatal("failed to load config:", err)
    }
    
    config := GetConfig()
    // Use config
}

// ❌ Don't hardcode values
const DatabaseURL = "postgres://localhost/mydb" // WRONG

// ✅ All config from environment
config.DatabaseURL = os.Getenv("DATABASE_URL")

// ❌ Don't commit secrets
JWT_SECRET=supersecret // WRONG: in config.yaml

// ✅ Secrets from environment
viper.BindEnv("jwt.secret", "JWT_SECRET")
```

## Example: Complete Setup

```go
package config

import (
    "errors"
    "os"

    "github.com/spf13/viper"
)

type Config struct {
    Server   ServerConfig
    Database DatabaseConfig
    JWT      JWTConfig
    Log      LogConfig
}

func Load() (*Config, error) {
    env := os.Getenv("ENVIRONMENT")
    if env == "" {
        env = "development"
    }

    viper.SetConfigName(fmt.Sprintf("config.%s", env))
    viper.SetConfigType("yaml")
    viper.AddConfigPath("./config")
    viper.AutomaticEnv()

    if err := viper.ReadInConfig(); err != nil {
        if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
            return nil, err
        }
    }

    var config Config
    if err := viper.Unmarshal(&config); err != nil {
        return nil, err
    }

    if err := config.Validate(); err != nil {
        return nil, err
    }

    return &config, nil
}

func (c *Config) Validate() error {
    if c.Database.URL == "" {
        return errors.New("database URL is required")
    }
    if c.JWT.Secret == "" {
        return errors.New("JWT secret is required")
    }
    return nil
}
```

---

**References**:
- [viper](https://github.com/spf13/viper)
- [godotenv](https://github.com/joho/godotenv)
- [12-Factor App Config](https://12factor.net/config)
