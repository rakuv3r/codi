---
language: go
category: dependency
applies_to: [cli]
---

# Cobra CLI Framework Standards

## Overview

Cobra is the standard CLI framework for Go, used by Kubernetes, Docker, GitHub CLI, and hundreds of other projects. It provides command structure, flag management, automatic help generation, and shell completion.

## Core Principles

1. **RunE over Run** - Always return errors for testability
2. **Constructor pattern** - Return commands from functions, avoid package-level variables
3. **Local flags by default** - Only use PersistentFlags for truly global concerns
4. **Args validators** - Use built-in validators instead of manual checks
5. **Testable output** - Write to `cmd.OutOrStdout()`, not `fmt.Println`

## Command Structure

### Constructor Pattern (Recommended)

```go
func NewRootCmd() *cobra.Command {
    cmd := &cobra.Command{
        Use:   "myapp",
        Short: "A brief description",
        Long:  "A longer description...",
        RunE:  runRoot,
    }
    
    // Add flags
    cmd.Flags().StringP("config", "c", "", "config file")
    
    // Add subcommands
    cmd.AddCommand(NewServeCmd())
    cmd.AddCommand(NewVersionCmd())
    
    return cmd
}
```

**Why**: Avoids hidden dependencies from `init()`, makes commands testable, explicit control flow.

### RunE vs Run (Critical)

```go
// ✅ Correct: Use RunE
RunE: func(cmd *cobra.Command, args []string) error {
    if err := doWork(); err != nil {
        return fmt.Errorf("failed to do work: %w", err)
    }
    return nil
}

// ❌ Wrong: Use Run
Run: func(cmd *cobra.Command, args []string) {
    if err := doWork(); err != nil {
        os.Exit(1)  // Breaks testing, can't intercept
    }
}
```

**Rule**: Always use `RunE`. It returns errors that Cobra handles centrally with proper exit codes.

## Argument Validation

Use built-in validators:

```go
cmd := &cobra.Command{
    Use:  "delete [id]",
    Args: cobra.ExactArgs(1),  // Requires exactly 1 argument
    RunE: runDelete,
}

// Available validators:
Args: cobra.NoArgs                    // No arguments allowed
Args: cobra.ExactArgs(2)              // Exactly N arguments
Args: cobra.MinimumNArgs(1)           // At least N arguments
Args: cobra.MaximumNArgs(3)           // At most N arguments
Args: cobra.RangeArgs(1, 3)           // Between min and max
Args: cobra.MatchAll(                 // Combine validators
    cobra.ExactArgs(1),
    cobra.OnlyValidArgs,
)
```

**Don't** manually check `len(args)` - use validators to get proper usage errors.

## Flags

### Local vs Persistent

```go
// Local flag - only for this command
cmd.Flags().StringP("output", "o", "json", "output format")

// Persistent flag - for this command and all descendants
rootCmd.PersistentFlags().BoolP("verbose", "v", false, "verbose output")
```

**Rule**: Most flags should be local. Only use `PersistentFlags` for cross-cutting concerns like `--config`, `--verbose`, `--output`.

### Flag Types

```go
// String
cmd.Flags().String("name", "", "user name")
cmd.Flags().StringP("name", "n", "", "user name")  // With shorthand

// Integer
cmd.Flags().Int("port", 8080, "server port")
cmd.Flags().IntP("port", "p", 8080, "server port")

// Boolean
cmd.Flags().Bool("force", false, "force operation")
cmd.Flags().BoolP("force", "f", false, "force operation")

// String slice (repeatable: --tag a --tag b)
cmd.Flags().StringSlice("tags", []string{}, "tags")

// Duration
cmd.Flags().Duration("timeout", 30*time.Second, "timeout")

// Bind to variable
var port int
cmd.Flags().IntVarP(&port, "port", "p", 8080, "server port")
```

### Reading Flag Values

```go
RunE: func(cmd *cobra.Command, args []string) error {
    // Get flag value
    port, err := cmd.Flags().GetInt("port")
    if err != nil {
        return err
    }
    
    // Or if bound to variable, use directly
    fmt.Println(port)  // Uses bound variable
    
    return nil
}
```

### Required Flags

```go
cmd.Flags().String("username", "", "username (required)")
cmd.MarkFlagRequired("username")
```

### Flag Validation

```go
PreRunE: func(cmd *cobra.Command, args []string) error {
    // Check mutually exclusive flags
    json, _ := cmd.Flags().GetBool("json")
    yaml, _ := cmd.Flags().GetBool("yaml")
    if json && yaml {
        return errors.New("--json and --yaml are mutually exclusive")
    }
    
    // Validate enum values
    env, _ := cmd.Flags().GetString("env")
    if env != "dev" && env != "prod" {
        return fmt.Errorf("invalid env: %s (must be dev or prod)", env)
    }
    
    return nil
}
```

## Configuration Precedence (with Viper)

```go
import "github.com/spf13/viper"

// Bind flags to viper
viper.BindPFlag("port", cmd.Flags().Lookup("port"))

// Set up environment variables
viper.SetEnvPrefix("myapp")
viper.AutomaticEnv()

// Read config file
viper.SetConfigName("config")
viper.AddConfigPath(".")
viper.ReadInConfig()

// Priority: flag > env var > config file > default
port := viper.GetInt("port")
```

**Check if flag was explicitly set**:

```go
if cmd.Flags().Changed("port") {
    // User explicitly set --port
} else {
    // Using default or value from env/config
}
```

## Hook Chain

Execution order:

```
PersistentPreRunE (root)
  ↓
PersistentPreRunE (parent)
  ↓
PreRunE (current)
  ↓
RunE (current)
  ↓
PostRunE (current)
  ↓
PersistentPostRunE (parent)
  ↓
PersistentPostRunE (root)
```

**Usage**:

```go
PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
    // Setup: auth, logging, DB connection
    return initDatabase()
}

PreRunE: func(cmd *cobra.Command, args []string) error {
    // Validation specific to this command
    return validateInput()
}

RunE: func(cmd *cobra.Command, args []string) error {
    // Main logic
    return executeCommand()
}

PostRunE: func(cmd *cobra.Command, args []string) error {
    // Cleanup
    return cleanup()
}

PersistentPostRunE: func(cmd *cobra.Command, args []string) error {
    // Cleanup: close DB, send telemetry
    return closeDatabase()
}
```

## Testable Output

```go
// ✅ Correct: Write to command output
RunE: func(cmd *cobra.Command, args []string) error {
    fmt.Fprintln(cmd.OutOrStdout(), "Success")
    fmt.Fprintf(cmd.ErrOrStderr(), "Warning: %s\n", msg)
    return nil
}

// ❌ Wrong: Write to stdout directly
RunE: func(cmd *cobra.Command, args []string) error {
    fmt.Println("Success")  // Can't capture in tests
    return nil
}
```

## Testing

```go
func TestServeCmd(t *testing.T) {
    cmd := NewServeCmd()
    
    // Capture output
    buf := new(bytes.Buffer)
    cmd.SetOut(buf)
    cmd.SetErr(buf)
    
    // Set arguments
    cmd.SetArgs([]string{"--port", "8080"})
    
    // Execute
    err := cmd.Execute()
    
    // Assert
    assert.NoError(t, err)
    assert.Contains(t, buf.String(), "started")
}
```

## Error Handling

```go
func main() {
    rootCmd := NewRootCmd()
    
    // Silence automatic error output
    rootCmd.SilenceErrors = true
    rootCmd.SilenceUsage = true
    
    // Execute and handle errors
    if err := rootCmd.Execute(); err != nil {
        fmt.Fprintf(os.Stderr, "Error: %v\n", err)
        os.Exit(1)
    }
}
```

**Why**: Centralized error formatting, consistent output, proper exit codes.

## Version Command

```go
var (
    Version   = "dev"
    BuildTime = "unknown"
)

func NewVersionCmd() *cobra.Command {
    return &cobra.Command{
        Use:   "version",
        Short: "Print version information",
        Run: func(cmd *cobra.Command, args []string) {
            fmt.Fprintf(cmd.OutOrStdout(), "%s (built %s)\n", Version, BuildTime)
        },
    }
}

// Build with:
// go build -ldflags "-X main.Version=1.0.0 -X main.BuildTime=$(date)"
```

## Common Patterns

### Subcommands Organization

```go
// cmd/root.go
func NewRootCmd() *cobra.Command {
    cmd := &cobra.Command{Use: "myapp"}
    cmd.AddCommand(NewServeCmd())
    cmd.AddCommand(NewMigrateCmd())
    return cmd
}

// cmd/serve.go
func NewServeCmd() *cobra.Command {
    return &cobra.Command{
        Use:   "serve",
        Short: "Start the server",
        RunE:  runServe,
    }
}

// cmd/migrate.go
func NewMigrateCmd() *cobra.Command {
    cmd := &cobra.Command{
        Use:   "migrate",
        Short: "Database migrations",
    }
    cmd.AddCommand(NewMigrateUpCmd())
    cmd.AddCommand(NewMigrateDownCmd())
    return cmd
}
```

### Dependency Injection

```go
func NewServeCmd(db *sql.DB, logger *log.Logger) *cobra.Command {
    return &cobra.Command{
        Use:  "serve",
        RunE: func(cmd *cobra.Command, args []string) error {
            return runServer(db, logger)
        },
    }
}

// In main:
db := setupDatabase()
logger := setupLogger()
rootCmd.AddCommand(NewServeCmd(db, logger))
```

## Anti-Patterns

### ❌ Don't use package-level variables

```go
// Wrong
var rootCmd = &cobra.Command{...}

func init() {
    rootCmd.AddCommand(serveCmd)
}

// Right
func NewRootCmd() *cobra.Command {
    cmd := &cobra.Command{...}
    cmd.AddCommand(NewServeCmd())
    return cmd
}
```

### ❌ Don't use Run (use RunE)

### ❌ Don't overuse PersistentFlags

```go
// Wrong: clutters help output
rootCmd.PersistentFlags().Int("port", 8080, "server port")

// Right: local to serve command
serveCmd.Flags().Int("port", 8080, "server port")
```

### ❌ Don't write to stdout directly (use cmd.OutOrStdout())

### ❌ Don't manually check len(args) (use Args validators)

## References

- Official docs: https://cobra.dev
- GitHub: https://github.com/spf13/cobra
- Used by: kubectl, docker, gh, hugo, etcd
