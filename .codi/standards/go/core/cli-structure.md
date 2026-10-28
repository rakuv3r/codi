---
language: go
category: core
applies_to: [cli]
---

# Go CLI Project Structure

## Overview

This guide covers structure and organization patterns for Go command-line interface (CLI) applications.

---

## Directory Layout

### `/cmd` - Command Definitions

The `cmd/` directory contains Cobra command definitions and structure.

**Purpose**:
- Define command hierarchy (root, subcommands)
- Parse flags and arguments
- Call internal/ functions with parsed input

### `/internal` - Command Logic

The `internal/` directory contains the actual implementation of commands.

**Go compiler enforced**: Code in `internal/` cannot be imported by external projects.

**Organization principle**: Organize by feature/function (serve/, config/, output/), matching command structure.

**Why separate cmd/ and internal/?**
- Testability: internal/ can be tested without Cobra overhead
- Reusability: logic can be called from different commands
- Clean separation: command structure vs implementation
- Validated by: GitHub CLI, kubectl, Docker CLI all follow this pattern

---

## Recommended Structure

```
mycli/
├── main.go                      # Application entry point
├── cmd/
│   ├── root.go                  # Root command definition
│   ├── serve.go                 # Subcommand definition
│   └── version.go               # Version subcommand
├── internal/
│   ├── serve/                   # Serve command logic
│   │   ├── service.go           # Core functionality
│   │   └── service_test.go
│   ├── config/                  # Configuration
│   │   └── config.go
│   └── output/                  # Output formatting
│       └── formatter.go
├── go.mod
├── go.sum
├── Makefile
└── README.md
```

---

## Key Principles

### Separation of Concerns

**cmd/** defines commands, **internal/** implements logic

```
cmd/serve.go:
- Command structure (flags, arguments)
- Help text
- Calls internal/serve functions

internal/serve/:
- Actual implementation
- Business logic
- Testable independently
```

**Validated by**: GitHub CLI (46k⭐), kubectl, Docker CLI all follow this pattern.

---

## Entry Point

### Minimal main.go

Keep the entry point simple:

```go
package main

import (
    "os"
    "github.com/user/mycli/cmd"
)

func main() {
    if err := cmd.Execute(); err != nil {
        os.Exit(1)
    }
}
```

**3-5 lines maximum**. All logic goes in cmd/ or internal/.

---

## Command Organization

### Root Command (cmd/root.go)

```go
package cmd

import (
    "fmt"
    "os"
    "github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
    Use:   "mycli",
    Short: "A brief description",
    Long:  `A longer description of what the CLI does`,
}

func Execute() error {
    return rootCmd.Execute()
}

func init() {
    // Global flags (available to all subcommands)
    rootCmd.PersistentFlags().StringP("config", "c", "", "config file path")
    rootCmd.PersistentFlags().BoolP("verbose", "v", false, "verbose output")
}
```

### Subcommand (cmd/serve.go)

```go
package cmd

import (
    "github.com/spf13/cobra"
    "github.com/user/mycli/internal/serve"
)

var serveCmd = &cobra.Command{
    Use:   "serve",
    Short: "Start the server",
    Long:  `Start the HTTP server on specified port`,
    RunE:  runServe,
}

var (
    port int
)

func init() {
    rootCmd.AddCommand(serveCmd)
    
    // Command-specific flags
    serveCmd.Flags().IntVarP(&port, "port", "p", 8080, "port to listen on")
}

func runServe(cmd *cobra.Command, args []string) error {
    // Parse flags and call internal logic
    cfg := &serve.Config{
        Port: port,
    }
    
    return serve.Run(cfg)
}
```

**cmd/ only handles**:
- Flag parsing
- Argument validation
- Calling internal functions
- Error wrapping

---

## Internal Logic Organization

### Feature-based Modules

Organize `internal/` by feature/function, not technical layer:

```
internal/
  serve/           # Server functionality
    service.go
    routes.go
  config/          # Configuration handling
    loader.go
    validate.go
  output/          # Output formatting
    table.go
    json.go
```

Same principle as API projects: organize by what it does, not what it is.

### Implementation (internal/serve/service.go)

```go
package serve

import (
    "fmt"
    "net/http"
)

type Config struct {
    Port int
}

func Run(cfg *Config) error {
    addr := fmt.Sprintf(":%d", cfg.Port)
    
    mux := http.NewServeMux()
    mux.HandleFunc("/", handleRoot)
    
    fmt.Printf("Starting server on %s\n", addr)
    return http.ListenAndServe(addr, mux)
}

func handleRoot(w http.ResponseWriter, r *http.Request) {
    fmt.Fprintf(w, "Hello from mycli\n")
}
```

**Keep internal packages focused**: one concern per package.

---

## Configuration Management

### Viper for Configuration

```go
package config

import (
    "github.com/spf13/viper"
)

type Config struct {
    Port     int    `mapstructure:"port"`
    LogLevel string `mapstructure:"log_level"`
}

func Load(path string) (*Config, error) {
    // Config file
    if path != "" {
        viper.SetConfigFile(path)
    } else {
        viper.SetConfigName("config")
        viper.AddConfigPath(".")
        viper.AddConfigPath("$HOME/.mycli")
    }
    
    // Environment variables
    viper.AutomaticEnv()
    viper.SetEnvPrefix("MYCLI")
    
    // Read config (optional - allow running without config file)
    if err := viper.ReadInConfig(); err != nil {
        if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
            return nil, err
        }
    }
    
    var cfg Config
    if err := viper.Unmarshal(&cfg); err != nil {
        return nil, err
    }
    
    return &cfg, nil
}
```

**Priority**: flags > environment variables > config file > defaults

---

## Output Formatting

### Consistent Output Patterns

```go
package output

import (
    "fmt"
    "os"
    "text/tabwriter"
)

type Formatter struct {
    writer *tabwriter.Writer
}

func NewFormatter() *Formatter {
    return &Formatter{
        writer: tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0),
    }
}

func (f *Formatter) Table(headers []string, rows [][]string) {
    // Print headers
    for i, h := range headers {
        fmt.Fprint(f.writer, h)
        if i < len(headers)-1 {
            fmt.Fprint(f.writer, "\t")
        }
    }
    fmt.Fprintln(f.writer)
    
    // Print rows
    for _, row := range rows {
        for i, cell := range row {
            fmt.Fprint(f.writer, cell)
            if i < len(row)-1 {
                fmt.Fprint(f.writer, "\t")
            }
        }
        fmt.Fprintln(f.writer)
    }
    
    f.writer.Flush()
}
```

**Use tabwriter** for aligned output, makes CLI feel professional.

---

## Error Handling

### CLI-specific Error Handling

```go
package cmd

import (
    "fmt"
)

func runServe(cmd *cobra.Command, args []string) error {
    cfg, err := config.Load(configFile)
    if err != nil {
        return fmt.Errorf("failed to load config: %w", err)
    }
    
    if err := serve.Run(cfg); err != nil {
        return fmt.Errorf("server error: %w", err)
    }
    
    return nil
}
```

**Error wrapping**: provide context at each layer.

### Exit Codes

```go
func main() {
    if err := cmd.Execute(); err != nil {
        fmt.Fprintln(os.Stderr, err)
        os.Exit(1)
    }
}
```

Exit with code 1 on error, 0 on success.

---

## Testing

### Command Tests

```go
package cmd

import (
    "bytes"
    "testing"
    "github.com/stretchr/testify/assert"
)

func TestServeCommand(t *testing.T) {
    cmd := rootCmd
    b := bytes.NewBufferString("")
    cmd.SetOut(b)
    cmd.SetArgs([]string{"serve", "--help"})
    
    err := cmd.Execute()
    
    assert.NoError(t, err)
    assert.Contains(t, b.String(), "Start the server")
}
```

### Logic Tests

Test internal packages independently (easier to test, no Cobra overhead):

```go
package serve

import (
    "testing"
    "github.com/stretchr/testify/assert"
)

func TestRun(t *testing.T) {
    cfg := &Config{Port: 8080}
    
    // Test logic without actually starting server
    // Use interfaces and mocks for external dependencies
}
```

---

## Common Patterns

### Progress Indicators

```go
import "github.com/schollz/progressbar/v3"

func processItems(items []string) error {
    bar := progressbar.Default(int64(len(items)))
    
    for _, item := range items {
        // Process item
        bar.Add(1)
    }
    
    return nil
}
```

### Interactive Prompts

```go
import "github.com/AlecAivazis/survey/v2"

func confirmAction() (bool, error) {
    confirm := false
    prompt := &survey.Confirm{
        Message: "Do you want to continue?",
    }
    err := survey.AskOne(prompt, &confirm)
    return confirm, err
}
```

### Colored Output

```go
import "github.com/fatih/color"

func printSuccess(msg string) {
    color.Green("✓ %s", msg)
}

func printError(msg string) {
    color.Red("✗ %s", msg)
}
```

---

## Project Variations

### Simple CLI (no external dependencies)

```
mycli/
  main.go
  cmd/
    root.go
    list.go
    create.go
  internal/
    output/       # Output formatters
    config/
```

### Complex CLI (with state/database)

```
mycli/
  main.go
  cmd/
    root.go
    server.go
    migrate.go
  internal/
    server/       # Server logic
    database/     # DB operations
    migration/    # Migration logic
    config/
```

### CLI + Library (exportable package)

```
mycli/
  main.go
  cmd/            # CLI commands
  pkg/            # Public API (can be imported)
    client/
  internal/       # Private implementation
```

---

## Reference Projects

- **GitHub CLI (gh)** - 46k⭐ - Excellent example of clean CLI architecture
- **kubectl** - Kubernetes CLI - Complex command hierarchy
- **Terraform CLI** - State management patterns
- **Docker CLI** - Subcommand organization and UX

---

## Summary

✅ **DO**:
- Keep main.go minimal (3-5 lines)
- Separate command definition (cmd/) from logic (internal/)
- Organize internal/ by feature/function
- Use Cobra for command structure
- Use Viper for configuration
- Test internal packages independently
- Provide helpful error messages and usage text

❌ **DON'T**:
- Put business logic in cmd/ package
- Mix command parsing with execution logic
- Leave internal/ empty
- Skip help text and examples
- Hardcode configuration values
- Ignore exit codes

---

**Remember**: CLI apps should feel fast, responsive, and user-friendly. Clean separation between command structure and logic makes your CLI maintainable and testable.
