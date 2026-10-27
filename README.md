# codi

**Quality infrastructure for AI coding agents.**

[中文文档](README.zh-CN.md)

---

## Philosophy: The Three Boundaries of Engineering

Every engineering problem exists within three boundaries:

```
         Theoretical Limit (Best Practices)
                    ▲
                   ╱│╲
                  ╱ │ ╲
                 ╱  │  ╲
                ╱   │   ╲
               ╱  Ladder ╲
              ╱   (CR+AI)  ╲
             ╱      │       ╲
            ╱       │        ╲
           ╱        │         ╲
          ╱─────────┼──────────╲
         Baseline (make lint/test)
```

### 1. Baseline (Lower Bound)
**What tools guarantee automatically.**

- `make lint` - golangci-lint, ruff, type checkers
- `make test` - Automated test suites
- Git hooks - Pre-commit validation

**Purpose:** Catch trivial errors before human review.

### 2. Theoretical Limit (Upper Bound)
**What's possible given current knowledge.**

- Google Go Style Guide + production patterns
- PEP 8 + framework best practices
- Security baselines, performance patterns

**Purpose:** Know what good code looks like.

### 3. The Ladder (Climbing Path)
**How you get from baseline to limit.**

```
Code → CR (find gaps) → AI fixes (apply standards) → CR (verify) → Repeat
```

**codi's role:**
- Provides the **limit** (curated standards)
- Lowers the **ladder cost** (AI auto-applies standards)
- Ensures the **baseline** (project templates with tooling)

**The difference between professionals and amateurs:**
- Amateurs don't know the limit exists
- Beginners have a low baseline
- Professionals know both boundaries and climb the ladder systematically

---

## What is codi?

A unified code quality system for AI-assisted development.

**Core Features:**
- **Standards Library** - Curated coding standards (Go, Python)
- **Project Templates** - Pre-configured tooling (lint, test, build)
- **AI Integration** - Auto-generates configuration for Claude Code, Codex, Cursor
- **Code Review Workflows** - AI automated checks + human review guides

**Why it matters:**
- AI writes code faster than ever
- Quality standards prevent AI from writing bad code at scale
- codi makes standards machine-readable for AI agents

---

## Installation

### From Source

```bash
git clone https://github.com/rakuv3r/codi.git
cd codi
make build
sudo cp bin/codi /usr/local/bin/
```

Or without sudo:
```bash
cp bin/codi ~/.local/bin/  # Make sure ~/.local/bin is in your PATH
```

### From Release (Coming Soon)

```bash
curl -fsSL https://raw.githubusercontent.com/rakuv3r/codi/main/install.sh | bash
```

---

## Quick Start

### Adopt Existing Project

```bash
cd your-project
codi adopt
```

**What it does:**
1. Detects language (Go/Python)
2. Analyzes dependencies
3. Generates AI tool config files (`.claude/CLAUDE.md`, `.codex/AGENTS.md`)
4. Creates `.codi/config.yaml` with project metadata
5. Adds quality tools (Makefile, linters) if missing

**Result:** AI agents now see your standards automatically.

---

### Create New Project

```bash
codi init my-api
cd my-api
make setup && make build
```

**Interactive prompts:**
- Language? (Go, Python)
- Type? (CLI, API)
- Modules? (Gin, Cobra, FastAPI, etc.)
- AI tools? (Claude, Codex, Cursor)

**Generated structure:**
```
my-api/
├── .codi/
│   └── config.yaml          # Project metadata
├── .claude/
│   ├── CLAUDE.md            # Standards for Claude Code
│   ├── review-checklist.md  # AI automated checks
│   └── review-manual.md     # Human review guide
├── Makefile                 # lint, test, build, run
├── .golangci.yml            # Linter config
└── cmd/main.go              # Scaffold code
```

---

## The Ladder: Code Review Workflow

### Step 1: Baseline Check (Automated)
```bash
make lint  # Format, types, common issues
make test  # Unit tests pass
```

### Step 2: AI Review (review-checklist.md)
**AI runs 70% of checks automatically:**
- Build/test quality
- Error handling patterns
- Security basics (SQL injection, XSS)
- Performance anti-patterns

### Step 3: Human Review (review-manual.md)
**Focus on what AI cannot judge:**
- Architecture decisions
- Business logic correctness
- API design
- Scalability considerations

### Step 4: AI Fixes Issues
**Paste CR feedback → AI reads standards → AI fixes code**

**Repeat 2-4 until code reaches the limit.**

---

## Standards System

### Universal Standards
- Git workflow (branching, commits)
- Documentation (README, API docs, ADR)
- Security baseline (secrets, dependencies)

### Go Standards
- Based on [Google Go Style Guide](https://google.github.io/styleguide/go/)
- Extensions: HTTP handler patterns, client-go usage, Neo4j batching

### Python Standards
- Based on PEP 8, [Google Python Style Guide](https://google.github.io/styleguide/pyguide.html)
- Extensions: Django ORM patterns, FastAPI validation, async best practices

### Dependency-Specific Standards
**Auto-loaded based on your imports:**
- `github.com/gin-gonic/gin` → Gin middleware patterns
- `github.com/spf13/cobra` → CLI design patterns
- `sqlalchemy` → ORM transaction management
- `fastapi` → Dependency injection, validation

**Zero-mapping:** Dependency path = Standard path
- `github.com/gin-gonic/gin` → `standards/go/github.com/gin-gonic/gin.md`

---

## CLI Reference

```bash
codi adopt              # Integrate into existing project
codi init <name>        # Create new project
codi sync               # Update standards after dependency changes
codi status             # Show project configuration
codi view [language]    # Browse standards interactively
```

---

## Contributing

We welcome contributions! Here's how to extend codi:

### Add Support for a New Language

**1. Create language structure:**
```bash
mkdir -p standards/{language}/core
mkdir -p templates/{language}/base/{cli,api}
mkdir -p templates/{language}/module
```

**2. Add core standards:**
```bash
# standards/rust/core/
- formatting.md
- error-handling.md
- testing.md
```

**3. Implement analyzer:**
```go
// internal/analyzer/rust.go
type RustAnalyzer struct{}

func (a *RustAnalyzer) Analyze(projectPath string) (*DependencyList, error) {
    // Parse Cargo.toml
    // Return dependencies
}
```

**4. Add templates:**
```bash
# templates/rust/base/cli/
- Cargo.toml.tmpl
- main.rs.tmpl
- Makefile.tmpl
```

**5. Update detector:**
```go
// internal/analyzer/detector.go
case "Cargo.toml":
    return LanguageRust, nil
```

---

### Add a Library Standard

**Example: Add `axum` (Rust web framework)**

**1. Create standard file:**
```bash
# standards/rust/crates.io/axum.md
```

**2. Write standard:**
````markdown
# axum

## Router Organization
```rust
// Group related routes with Router::nest
let api = Router::new()
    .route("/users", get(list_users).post(create_user))
    .route("/users/:id", get(get_user));

let app = Router::new()
    .nest("/api/v1", api)
    .layer(TraceLayer::new_for_http());
```

## Error Handling
- Use `Result<T, AppError>` for fallible handlers
- Implement `IntoResponse` for custom error types
- Log errors with tracing before converting to HTTP responses
````

**3. Test:**
```bash
# In a Rust project with axum
codi adopt
# Check .claude/CLAUDE.md includes axum standards
```

**That's it!** Zero-mapping means codi auto-discovers it.

---

### Add a Code Review Check

**Edit `.claude/review-checklist.md`:**

```markdown
## Security

### Input Validation
- [ ] All user inputs validated
- [ ] SQL queries use parameterized statements
- [ ] File paths sanitized (no `../` traversal)
- [ ] API rate limiting implemented  ← NEW CHECK
```

**AI will now check this automatically.**

---

### Project Structure

```
codi/
├── standards/           # The "limit" - best practices
│   ├── universal/       # Language-agnostic
│   ├── go/
│   │   ├── core/        # Language fundamentals
│   │   └── github.com/  # Library-specific (zero-mapped)
│   └── python/
│       ├── core/
│       └── pypi/         # Package standards
│
├── templates/           # Project scaffolds
│   ├── go/
│   │   ├── base/        # cli, api
│   │   └── module/      # gin, cobra, testify
│   └── python/
│       ├── base/
│       └── module/      # fastapi, pytest
│
├── internal/
│   ├── analyzer/        # Detect language, parse deps
│   ├── generator/       # Scaffold projects
│   ├── standard/        # Load and serve standards
│   └── command/         # CLI commands
│
└── cmd/codi/main.go
```

---

### Validation: The Simplicity Test

**Can someone add support for a new library in under 10 minutes?**

**Yes:**
1. Create `standards/{lang}/{path}.md` (5 min to write)
2. Run `codi adopt` in test project (1 min)
3. Verify AI sees the standard (1 min)

**If it takes longer, the design failed.**

---

## Design Principles

### 1. Reference Authority, Don't Reinvent
- Google Style Guides
- PEP 8
- Language-specific idioms

**Add only:**
- Real-world patterns not in guides
- Dependency-specific usage

### 2. Machine-Readable Standards
- Standards are for AI agents, not just humans
- Markdown format: human-readable, AI-parseable
- Code examples in every standard

### 3. Zero-Mapping
- Dependency path = Standard path
- No config files to maintain
- Drop in a `.md` file, it works

### 4. Composable Quality
```
Baseline (tools) + Standards (limit) + CR (ladder) = Quality
```

Each layer is independent but synergistic.

---

## Roadmap

- [ ] Rust support
- [ ] TypeScript support
- [ ] Java support
- [ ] LSP integration (real-time standard hints in IDE)
- [ ] CI integration (auto-CR in GitHub Actions)
- [ ] Standard versioning (pin to specific standard versions)

---

## License

MIT License - see [LICENSE](LICENSE)

---

## Acknowledgments

**Inspired by:**
- Google Style Guides
- The Pragmatic Programmer
- 吴军《态度》 - The Three Boundaries framework

**Built for:**
- Claude Code
- GitHub Copilot
- OpenAI Codex
- Cursor
- Any AI coding assistant that reads project context
